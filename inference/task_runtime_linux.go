package inference

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// TaskRuntimeConfig selects trusted, already installed local artifacts. This
// foreground owner does not implement the legacy Owned containment contract.
type TaskRuntimeConfig struct {
	Executable, Weights string
	Context, Threads    int
	LoadTimeout         time.Duration
}

// TaskRuntime owns one process for one interactive CLI launch, including all
// model/tool turns and idle prompts. There is no daemon, retry or idle eviction.
type TaskRuntime struct {
	config            TaskRuntimeConfig
	executable, model *os.File
	identity          string
	modelBytes        int64
	cmd               *exec.Cmd
	done              chan struct{}
	waitErr           error
	endpoint          string
	startIdentity     string
	once              sync.Once
	closeErr          error
	started           bool
}

// PrepareTaskRuntime hashes and holds exactly the selected files. No process or
// endpoint exists yet, so session identity can be checked before launching.
func PrepareTaskRuntime(ctx context.Context, c TaskRuntimeConfig) (*TaskRuntime, error) {
	if c.Context < 512 || c.Context > 65536 || c.Threads < 1 || c.Threads > 16 || c.LoadTimeout < time.Second || c.LoadTimeout > 10*time.Minute {
		return nil, errors.New("owned runtime requires context 512..65536, threads 1..16 and load timeout 1s..10m")
	}
	c.Executable, c.Weights = filepath.Clean(c.Executable), filepath.Clean(c.Weights)
	if strings.ContainsAny(filepath.Dir(c.Executable), ":;$") {
		return nil, errors.New("runtime directory cannot contain library-search separators or expansion tokens (: ; $)")
	}
	r := &TaskRuntime{config: c}
	var err error
	var runtimeHash, modelHash string
	if r.executable, runtimeHash, _, err = openRuntimeArtifact(ctx, c.Executable, true); err != nil {
		return nil, err
	}
	if r.model, modelHash, r.modelBytes, err = openRuntimeArtifact(ctx, c.Weights, false); err != nil {
		r.Close()
		return nil, err
	}
	identity, _ := json.Marshal(struct {
		Version, Executable, ExecutableSHA256, Weights, WeightsSHA256 string
		Context, Threads                                              int
	}{"task-runtime-v1", filepath.Clean(c.Executable), runtimeHash, filepath.Clean(c.Weights), modelHash, c.Context, c.Threads})
	r.identity = string(identity)
	return r, nil
}

func openRuntimeArtifact(ctx context.Context, path string, executable bool) (*os.File, string, int64, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) == "/" {
		return nil, "", 0, errors.New("owned runtime and weights require explicit absolute local file paths")
	}
	root, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, "", 0, err
	}
	defer unix.Close(root)
	fd, err := unix.Openat2(root, strings.TrimPrefix(filepath.Clean(path), "/"), &unix.OpenHow{Flags: unix.O_RDONLY | unix.O_NONBLOCK | unix.O_CLOEXEC, Resolve: unix.RESOLVE_BENEATH | unix.RESOLVE_NO_SYMLINKS})
	if err != nil {
		return nil, "", 0, errors.New("owned artifact unavailable; symlink ancestry is not allowed")
	}
	f := os.NewFile(uintptr(fd), path)
	failed := func(err error) (*os.File, string, int64, error) { f.Close(); return nil, "", 0, err }
	st, err := f.Stat()
	if err != nil || !st.Mode().IsRegular() || st.Mode().Perm()&0022 != 0 || st.Size() < 1 || st.Size() > 32<<30 || (executable && st.Mode().Perm()&0111 == 0) {
		return failed(errors.New("owned artifacts must be nonempty regular files, not group/world writable; runtime must be executable (32 GiB file ceiling)"))
	}
	h := sha256.New()
	buf := make([]byte, 128<<10)
	for {
		if err := ctx.Err(); err != nil {
			return failed(err)
		}
		n, err := f.Read(buf)
		h.Write(buf[:n])
		if err == io.EOF {
			break
		}
		if err != nil {
			return failed(err)
		}
	}
	after, err := f.Stat()
	if err != nil || st.Size() != after.Size() || st.ModTime() != after.ModTime() {
		return failed(errors.New("owned artifact changed during hashing"))
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return failed(err)
	}
	return f, hex.EncodeToString(h.Sum(nil)), st.Size(), nil
}

func (r *TaskRuntime) Identity() string { return r.identity }

// Start uses an ephemeral loopback port, but never trusts a service just because
// it answers there. Every connection is verified against this child's accepted
// socket inode BEFORE HTTP headers, auth or workspace content can be written.
func (r *TaskRuntime) Start(ctx context.Context, c EndpointConfig) (*Local, error) {
	if r.started || r.executable == nil || r.model == nil {
		return nil, errors.New("owned runtime is single-use; no restart or adoption")
	}
	r.started = true
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := availableMemory(r.modelBytes + (1 << 30)); err != nil {
		return nil, err // admission snapshot only, not a native memory reservation
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	port := listener.Addr().(*net.TCPAddr).Port
	secret := make([]byte, 48)
	if _, err := rand.Read(secret); err != nil {
		listener.Close()
		return nil, err
	}
	key, alias := hex.EncodeToString(secret[:32]), "familiar-"+hex.EncodeToString(secret[32:])
	address := net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	r.endpoint = "http://" + address + "/v1"
	args := []string{"--model", "/proc/self/fd/4", "--alias", alias, "--host", "127.0.0.1", "--port", strconv.Itoa(port),
		"--offline", "--no-webui", "--no-webui-mcp-proxy", "--no-agent", "--no-mmproj", "--jinja",
		"--device", "none", "--gpu-layers", "0", "--no-kv-offload", "--no-op-offload", "--fit", "off",
		"--ctx-size", strconv.Itoa(r.config.Context), "--parallel", "1", "--threads", strconv.Itoa(r.config.Threads),
		"--threads-batch", strconv.Itoa(r.config.Threads), "--threads-http", "2", "--cache-ram", "128", "--no-context-shift",
		"--top-k", "50", "--repeat-penalty", "1.1", "--reasoning-format", "deepseek", "--log-disable"}
	// Executing/reading held descriptors avoids reopening replaced pathnames.
	r.cmd = exec.Command("/proc/self/fd/3", args...)
	r.cmd.ExtraFiles = []*os.File{r.executable, r.model}
	r.cmd.Env = []string{"LANG=C", "HOME=/nonexistent", "LD_LIBRARY_PATH=" + filepath.Dir(r.config.Executable), "OMP_NUM_THREADS=" + strconv.Itoa(r.config.Threads), "OPENBLAS_NUM_THREADS=" + strconv.Itoa(r.config.Threads), "LLAMA_API_KEY=" + key}
	r.cmd.Dir = "/"
	r.cmd.SysProcAttr = &syscall.SysProcAttr{Pdeathsig: syscall.SIGKILL, Setpgid: true}
	// Nil streams use /dev/null without pipe-copy goroutines that could keep
	// Wait blocked if an unsupported executable left inherited pipes open.
	// llama.cpp cannot inherit a listening FD. The race is fail-closed: a
	// collision is never adopted, and no request is sent to a foreign socket.
	listener.Close()
	if err := ctx.Err(); err != nil {
		r.cmd = nil
		return nil, err
	}
	if err := r.cmd.Start(); err != nil {
		r.cmd = nil
		return nil, fmt.Errorf("owned runtime could not execute: %w", err)
	}
	r.done = make(chan struct{})
	go func() { r.waitErr = r.cmd.Wait(); close(r.done) }()
	r.startIdentity = processStart(r.cmd.Process.Pid)
	if r.startIdentity == "" {
		r.Close()
		return nil, errors.New("cannot identify owned runtime process; stopped before HTTP")
	}
	c.Endpoint, c.Model, c.Llama = r.endpoint, alias, true
	local, err := NewLocal(c)
	if err != nil {
		r.Close()
		return nil, err
	}
	transport := local.client.Transport.(*http.Transport)
	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if network != "tcp" || addr != address {
			return nil, errors.New("non-owned destination denied")
		}
		return r.dialOwned(ctx, address, port)
	}
	local.client.Transport = runtimeAuth{transport, key}
	loadCtx, cancel := context.WithTimeout(ctx, r.config.LoadTimeout)
	defer cancel()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		var models struct {
			Data []struct {
				ID string `json:"id"`
			} `json:"data"`
		}
		probe, stop := context.WithTimeout(loadCtx, time.Second)
		var health struct {
			Status string `json:"status"`
		}
		err := local.jsonRequest(probe, "/health", nil, &health, 4096)
		if err == nil && health.Status == "ok" {
			err = local.jsonRequest(probe, "/v1/models", nil, &models, 16<<10)
		}
		stop()
		if err == nil && len(models.Data) == 1 && models.Data[0].ID == alias {
			return local, nil
		}
		select {
		case <-loadCtx.Done():
			local.Close()
			r.Close()
			return nil, fmt.Errorf("owned runtime startup canceled or timed out: %w", loadCtx.Err())
		case <-r.done:
			local.Close()
			r.Close()
			return nil, fmt.Errorf("owned runtime exited during startup: %v; check executable, libraries, GGUF and compatible flags", r.waitErr)
		case <-ticker.C:
		}
	}
}

type runtimeAuth struct {
	base *http.Transport
	key  string
}

func (t runtimeAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+t.key)
	return t.base.RoundTrip(clone)
}

func (t runtimeAuth) CloseIdleConnections() { t.base.CloseIdleConnections() }

func (r *TaskRuntime) dialOwned(ctx context.Context, address string, port int) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp4", address)
	if err != nil {
		return nil, err
	}
	clientPort := conn.LocalAddr().(*net.TCPAddr).Port
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if processStart(r.cmd.Process.Pid) == r.startIdentity && childOwnsConnection(r.cmd.Process.Pid, port, clientPort) && processStart(r.cmd.Process.Pid) == r.startIdentity {
			select {
			case <-r.done:
				conn.Close()
				return nil, errors.New("owned runtime exited before HTTP")
			default:
				return conn, nil
			}
		}
		select {
		case <-r.done:
			conn.Close()
			return nil, errors.New("owned runtime exited; connection refused before HTTP")
		case <-ctx.Done():
			conn.Close()
			return nil, errors.New("cannot verify accepted socket belongs to owned runtime; no HTTP sent")
		case <-ticker.C:
		}
	}
}

func processStart(pid int) string {
	b, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return ""
	}
	end := strings.LastIndexByte(string(b), ')')
	if end < 0 {
		return ""
	}
	fields := strings.Fields(string(b)[end+1:])
	if len(fields) < 20 {
		return ""
	}
	return fields[19] // field22, after pid and parenthesized comm
}

func childOwnsConnection(pid, serverPort, clientPort int) bool {
	base := "/proc/" + strconv.Itoa(pid)
	entries, err := os.ReadDir(base + "/fd")
	if err != nil {
		return false
	}
	sockets := map[string]bool{}
	for _, entry := range entries {
		target, _ := os.Readlink(base + "/fd/" + entry.Name())
		if strings.HasPrefix(target, "socket:[") && strings.HasSuffix(target, "]") {
			sockets[strings.TrimSuffix(strings.TrimPrefix(target, "socket:["), "]")] = true
		}
	}
	f, err := os.Open(base + "/net/tcp")
	if err != nil {
		return false
	}
	defer f.Close()
	scanner := bufio.NewScanner(io.LimitReader(f, 4<<20))
	local, remote := fmt.Sprintf("0100007F:%04X", serverPort), fmt.Sprintf("0100007F:%04X", clientPort)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 10 && fields[1] == local && fields[2] == remote && fields[3] == "01" && sockets[fields[9]] {
			return true
		}
	}
	return false
}

// Info never includes auth credentials. Identity covers selected file bytes and
// settings, not transitive shared libraries or same-UID hostile file mutations.
func (r *TaskRuntime) Info() string {
	pid := 0
	state := "not-started"
	if r.cmd != nil && r.cmd.Process != nil {
		pid = r.cmd.Process.Pid
		state = "running"
		select {
		case <-r.done:
			state = r.cmd.ProcessState.String()
		default:
		}
	}
	return fmt.Sprintf("pid=%d start=%s endpoint=%s identity=%s state=%s", pid, r.startIdentity, r.endpoint, r.identity, state)
}

func (r *TaskRuntime) Close() error {
	r.once.Do(func() {
		if r.cmd != nil && r.cmd.Process != nil {
			select {
			case <-r.done:
				r.closeErr = r.waitErr
			default:
				_ = r.cmd.Process.Signal(syscall.SIGTERM)
				timer := time.NewTimer(2 * time.Second)
				select {
				case <-r.done:
				case <-timer.C:
					_ = r.cmd.Process.Kill()
					<-r.done // never claim retirement before Wait has reaped it
				}
				timer.Stop()
			}
		}
		if r.executable != nil {
			r.executable.Close()
			r.executable = nil
		}
		if r.model != nil {
			r.model.Close()
			r.model = nil
		}
	})
	return r.closeErr
}
