package inference

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	Cold          = "cold"
	Loading       = "loading"
	Ready         = "ready"
	Idle          = "idle"
	Unloading     = "unloading"
	Failed        = "failed"
	idleTTL       = 5 * time.Minute
	retryCooldown = 30 * time.Second
	callLimit     = 2 * time.Minute
)

// FilePin is an explicitly provisioned absolute file, never a download URL.
type FilePin struct{ Path, SHA256 string }

// LocalConfig is copied at construction. Launcher must be an already provisioned,
// reviewed containment executable accepting the protocol in inference/README.md.
// It applies process-tree memory/egress controls, then execs the
// runtime without forking. Evidence identifies that review/measurement record.
// The launcher must deny new outbound connections (including loopback) while
// permitting replies on the owned loopback listener, enforce PeakBytes,
// disable swap and preserve the supplied per-instance API key. No shell is used.
// Familiar does NOT manufacture or qualify a sandbox from these assertions.
type LocalConfig struct {
	Runtime, Launcher FilePin
	ModelPath         string
	Port              uint16
	Evidence          string
	LicenseApproved   bool
	PeakBytes         int64
}

// One logical owner per process prevents independent adapters from bypassing
// generation capacity or coalescing. Close it before constructing another.
var ownerClaim atomic.Bool

type Owned struct {
	config           LocalConfig
	profile          Profile
	localFingerprint string
	life             context.Context
	cancel           context.CancelFunc
	mu               sync.Mutex
	state, reason    string
	changed          chan struct{}
	closed           bool
	waiters, leases  int
	loadCancel       context.CancelFunc
	process          *localProcess
	idle             *time.Timer
	idleUntil        time.Time
	retryAt          time.Time
	queue            chan struct{}
	run              chan struct{}
}

// NewOwned does no filesystem/network/process work. It is opt-in, not readiness
// evidence. Runtime/build/native provenance and containment are external gates;
// actual bytes are verified lazily immediately before each launch.
func NewOwned(c LocalConfig) (*Owned, error) {
	for _, p := range []FilePin{c.Runtime, c.Launcher} {
		b, err := hex.DecodeString(p.SHA256)
		if !filepath.IsAbs(p.Path) || err != nil || len(b) != sha256.Size || hex.EncodeToString(b) != p.SHA256 {
			return nil, errors.New("inference: explicit absolute executable and lowercase SHA-256 required")
		}
	}
	if !filepath.IsAbs(c.ModelPath) || c.Port == 0 || c.Evidence == "" || !c.LicenseApproved {
		return nil, errors.New("inference: provisioned model, port, reviewed containment evidence and license approval required")
	}
	if c.PeakBytes < GenerationCandidate().Identity().Bytes || c.PeakBytes > 8<<30 {
		return nil, errors.New("inference: measured peak reservation required within 8 GiB ceiling")
	}
	if !ownerClaim.CompareAndSwap(false, true) {
		return nil, errors.New("inference: a local generation owner already exists")
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := GenerationCandidate()
	b, _ := json.Marshal(struct {
		Profile string
		Config  LocalConfig
		Options string
	}{
		p.Fingerprint(), c, "cpu;8192;1024;parallel=1;threads=2;temp=0.1;top_k=50;repeat=1.1;no-thinking;no-cache;deny-outbound;v2",
	})
	h := sha256.Sum256(b)
	return &Owned{config: c, profile: p, localFingerprint: hex.EncodeToString(h[:]), life: ctx, cancel: cancel,
		state: Cold, changed: make(chan struct{}), queue: make(chan struct{}, 5), run: make(chan struct{}, 1)}, nil
}

// Status is an in-memory snapshot only; it does not probe or load anything.
func (o *Owned) Status() (state, reason string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.state, o.reason
}

func (o *Owned) notify() { close(o.changed); o.changed = make(chan struct{}) }

// Lease pins residency, not authority or a network endpoint. Release is
// idempotent. Do not retain readiness leases beyond their caller's operation.
type Lease struct {
	once    sync.Once
	owner   *Owned
	process *localProcess
}

func (l *Lease) Release() {
	l.once.Do(func() {
		o := l.owner
		o.mu.Lock()
		defer o.mu.Unlock()
		o.leases--
		if o.leases == 0 && o.process != nil {
			if o.closed || o.state == Failed {
				o.unloadLocked()
			} else {
				o.idleLocked()
			}
		}
		o.notify()
	})
}

// Acquire coalesces exact-identity loads. Independent cancellations remove
// waiters; only the last waiter cancels loading. No retry occurs within a call.
func (o *Owned) Acquire(ctx context.Context) (*Lease, error) {
	ctx, cancel := context.WithTimeout(ctx, callLimit)
	defer cancel()
	o.mu.Lock()
	if o.closed {
		o.mu.Unlock()
		return nil, ErrClosed
	}
	if o.waiters >= 5 {
		o.mu.Unlock()
		return nil, ErrBusy
	}
	o.waiters++
	defer func() {
		o.waiters--
		if o.waiters == 0 && o.leases == 0 && o.state == Loading {
			o.loadCancel()
		}
		if o.waiters == 0 && o.leases == 0 && o.state == Ready {
			o.idleLocked()
		}
		o.notify()
		o.mu.Unlock()
	}()
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if o.closed {
			return nil, ErrClosed
		}
		switch o.state {
		case Ready, Idle:
			select {
			case <-o.process.done:
				o.process.client.CloseIdleConnections()
				o.process = nil
				o.state, o.reason, o.retryAt = Failed, "owned runtime exited", time.Now().Add(retryCooldown)
				return nil, errors.New("inference: owned runtime exited")
			default:
			}
			if o.idle != nil {
				o.idle.Stop()
				o.idle = nil
			}
			o.leases++
			o.state = Ready
			return &Lease{owner: o, process: o.process}, nil
		case Failed:
			if o.process != nil || o.leases != 0 || time.Now().Before(o.retryAt) {
				return nil, errors.New("inference: " + o.reason)
			}
			fallthrough
		case Cold:
			loadCtx, stop := context.WithTimeout(o.life, callLimit)
			o.loadCancel, o.state, o.reason = stop, Loading, ""
			go o.load(loadCtx)
		}
		changed := o.changed
		o.mu.Unlock()
		select {
		case <-ctx.Done():
		case <-changed:
		}
		o.mu.Lock()
	}
}

func (o *Owned) load(ctx context.Context) {
	p, err := o.start(ctx)
	o.mu.Lock()
	defer o.mu.Unlock()
	o.loadCancel()
	if err != nil {
		o.state, o.reason, o.retryAt = Failed, err.Error(), time.Now().Add(retryCooldown)
		if o.closed || o.waiters == 0 {
			o.state = Cold
		}
	} else {
		o.process, o.state = p, Ready
		go o.watch(p)
		if o.closed || o.waiters == 0 {
			o.unloadLocked()
		}
	}
	o.notify()
}

func (o *Owned) watch(p *localProcess) {
	<-p.done
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.process != p || (o.state != Ready && o.state != Idle) {
		return
	}
	if o.idle != nil {
		o.idle.Stop()
		o.idle = nil
	}
	o.process = nil
	o.state, o.reason, o.retryAt = Failed, "owned runtime exited", time.Now().Add(retryCooldown)
	o.notify()
}

func (o *Owned) unloadLocked() {
	if o.idle != nil {
		o.idle.Stop()
		o.idle = nil
	}
	p := o.process
	o.state = Unloading
	go func() {
		p.stop()
		o.mu.Lock()
		defer o.mu.Unlock()
		o.process = nil
		if o.reason != "" && !o.closed {
			o.state = Failed
		} else {
			o.state = Cold
		}
		o.notify()
	}()
}

func (o *Owned) evict() {
	o.mu.Lock()
	defer o.mu.Unlock()
	// A stopped timer may already be running. Check state/refcounts again.
	if o.state == Idle && o.leases == 0 && o.waiters == 0 {
		o.unloadLocked()
		o.notify()
	}
}

func (o *Owned) idleLocked() {
	o.state = Idle
	o.idleUntil = time.Now().Add(idleTTL)
	if o.idle != nil {
		o.idle.Stop()
	}
	o.idle = time.AfterFunc(idleTTL, func() {
		o.mu.Lock()
		defer o.mu.Unlock()
		if o.state == Idle && o.leases == 0 && o.waiters == 0 && !time.Now().Before(o.idleUntil) {
			o.unloadLocked()
			o.notify()
		}
	})
}

// EvictIdle allows earlier pressure eviction, never active lease eviction.
func (o *Owned) EvictIdle() { o.evict() }

// Close cancels calls and loading, waits for leases and owned teardown. A
// deadline returns an error, NOT a claim that the runtime has stopped. Call
// Close again after releasing external readiness leases to finish retirement.
func (o *Owned) Close(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.closed = true
	o.cancel()
	if o.leases == 0 && o.process != nil && o.state != Unloading {
		o.unloadLocked()
	}
	o.notify()
	for o.state == Loading || o.state == Unloading || o.leases != 0 || o.waiters != 0 {
		ch := o.changed
		o.mu.Unlock()
		select {
		case <-ctx.Done():
		case <-ch:
		}
		o.mu.Lock()
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	// Only this owner can release its claim, even on repeated Close calls.
	if o.state != "closed" {
		o.state = "closed"
		ownerClaim.Store(false)
	}
	return nil
}

type localProcess struct {
	cmd      *exec.Cmd
	done     chan struct{}
	client   *http.Client
	url, key string
}

func (p *localProcess) stop() {
	p.client.CloseIdleConnections()
	select {
	case <-p.done:
		return
	default:
	}
	_ = p.cmd.Process.Signal(syscall.SIGTERM)
	timer := time.NewTimer(2 * time.Second)
	defer timer.Stop()
	select {
	case <-p.done:
	case <-timer.C:
		_ = p.cmd.Process.Kill()
		<-p.done // do not release residency until the owned process is reaped
	}
}

func verifyFile(ctx context.Context, pin FilePin, size int64) error {
	f, err := os.OpenFile(pin.Path, os.O_RDONLY|syscall.O_NONBLOCK|syscall.O_NOFOLLOW, 0)
	if err != nil {
		return errors.New("inference: preprovisioned artifact unavailable")
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil || !st.Mode().IsRegular() || st.Mode().Perm()&0022 != 0 || (size != 0 && st.Size() != size) || (size == 0 && st.Size() > 512<<20) {
		return errors.New("inference: incompatible artifact type or size")
	}
	h := sha256.New()
	buf := make([]byte, 128<<10)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		n, err := f.Read(buf)
		_, _ = h.Write(buf[:n])
		if err == io.EOF {
			break
		}
		if err != nil {
			return errors.New("inference: artifact verification failed")
		}
	}
	if hex.EncodeToString(h.Sum(nil)) != pin.SHA256 {
		return errors.New("inference: artifact SHA-256 mismatch")
	}
	return nil
}

func (o *Owned) start(ctx context.Context) (*localProcess, error) {
	if err := availableMemory(o.config.PeakBytes); err != nil {
		return nil, err
	}
	id := o.profile.Identity()
	for _, pin := range []FilePin{o.config.Launcher, o.config.Runtime} {
		if err := verifyFile(ctx, pin, 0); err != nil {
			return nil, err
		}
	}
	if err := verifyFile(ctx, FilePin{o.config.ModelPath, id.SHA256}, id.Bytes); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	key := hex.EncodeToString(secret)
	port := strconv.Itoa(int(o.config.Port))
	args := []string{"--memory-max-bytes", strconv.FormatInt(o.config.PeakBytes, 10), "--deny-outbound", "--",
		o.config.Runtime.Path, "--model", o.config.ModelPath, "--alias", o.profile.Fingerprint(),
		"--host", "127.0.0.1", "--port", port, "--offline", "--no-webui", "--no-ui-mcp-proxy", "--no-agent", "--no-mmproj",
		"--device", "none", "--gpu-layers", "0", "--no-kv-offload", "--no-op-offload", "--fit", "off",
		"--ctx-size", "8192", "--predict", "1024", "--parallel", "1", "--threads", "2", "--threads-batch", "2",
		"--threads-http", "2", "--cache-ram", "0", "--no-cache-prompt", "--no-context-shift", "--no-slots",
		"--jinja", "--reasoning", "off", "--chat-template-kwargs", `{"enable_thinking":false}`,
		"--timeout", "120", "--log-disable"}
	cmd := exec.Command(o.config.Launcher.Path, args...)
	cmd.Env = []string{"LANG=C", "HOME=/nonexistent", "OMP_NUM_THREADS=2", "OPENBLAS_NUM_THREADS=2", "LLAMA_API_KEY=" + key}
	cmd.Dir = "/"
	cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL}
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	client := &http.Client{Transport: &http.Transport{
		Proxy: nil, DisableKeepAlives: true, DisableCompression: true, MaxResponseHeaderBytes: 16 << 10,
		ResponseHeaderTimeout: callLimit,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			if network != "tcp" || addr != "127.0.0.1:"+port {
				return nil, errors.New("inference: non-owned destination denied")
			}
			return (&net.Dialer{Timeout: time.Second}).DialContext(ctx, "tcp4", addr)
		},
	}, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("inference: redirect denied") }}
	p := &localProcess{cmd: cmd, client: client, url: "http://127.0.0.1:" + port, key: key, done: make(chan struct{})}
	if err := cmd.Start(); err != nil {
		client.CloseIdleConnections()
		return nil, errors.New("inference: containment launcher failed to start")
	}
	go func() { _ = cmd.Wait(); close(p.done) }()
	// Load-triggered authenticated readiness only. No startup inventory/probe and
	// no adoption of an existing service at a colliding port.
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		probeCtx, cancel := context.WithTimeout(ctx, time.Second)
		var models struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		err := p.postOrGet(probeCtx, "GET", "/v1/models", nil, &models, 4096)
		cancel()
		if err == nil && len(models.Data) == 1 && models.Data[0].ID == o.profile.Fingerprint() {
			select {
			case <-p.done:
				return nil, errors.New("inference: runtime exited while loading")
			default:
				return p, nil
			}
		}
		select {
		case <-ctx.Done():
			p.stop()
			return nil, ctx.Err()
		case <-p.done:
			client.CloseIdleConnections()
			return nil, errors.New("inference: owned runtime failed to load")
		case <-ticker.C:
		}
	}
}

// Only explicit lazy loading reads this documented Linux telemetry path.
// It is an admission snapshot, not an OS reservation or model-fit measurement.
func availableMemory(peak int64) error {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return errors.New("inference: memory telemetry unavailable")
	}
	defer f.Close()
	s := bufio.NewScanner(io.LimitReader(f, 64<<10))
	for s.Scan() {
		fields := strings.Fields(s.Text())
		if len(fields) != 3 || fields[0] != "MemAvailable:" || fields[2] != "kB" {
			continue
		}
		kb, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil || kb < 0 || kb > (1<<50) {
			break
		}
		if kb*1024 < peak+(512<<20) {
			return errors.New("inference: insufficient available RAM plus 512 MiB headroom")
		}
		return nil
	}
	return errors.New("inference: memory telemetry unavailable")
}
