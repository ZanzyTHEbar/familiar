package retrieval

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"familiar.local/familiar/internal/scopedwire"
	"github.com/ZanzyTHEbar/virtual-vectorfs/search"
)

type scopedClient struct {
	mu         sync.Mutex
	cmd        *exec.Cmd
	in         io.WriteCloser
	out        io.ReadCloser
	encode     *json.Encoder
	decode     *json.Decoder
	done       chan error
	closed     bool
	cancelled  bool
	parent     *os.Root
	stateName  string
	probeName  string
	diagnostic []byte // written before done is signalled; read only after Close
}

func openScoped(ctx context.Context, grant Grant, records []Record, stateParent string) (*Adapter, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	sources, docs, rootInfo, err := prepare(grant, records)
	if err != nil {
		return nil, err
	}
	if _, err = canonicalDirectory(stateParent, true); err != nil {
		return nil, err
	}
	if grant.Root != "" && (within(grant.Root, stateParent) || within(stateParent, grant.Root)) {
		return nil, errors.New("state and live root must be disjoint")
	}
	bundle := os.Getenv("FAMILIAR_SCOPED_RUNTIME")
	if bundle == "" {
		return nil, fmt.Errorf("%w: FAMILIAR_SCOPED_RUNTIME must name an installed bundle", ErrUnavailable)
	}
	if err = checkRuntimeDirectory(bundle); err != nil {
		return nil, fmt.Errorf("%w: runtime: %v", ErrUnavailable, err)
	}
	if grant.Root != "" && (within(grant.Root, bundle) || within(bundle, grant.Root)) {
		return nil, errors.New("runtime and workspace must be disjoint")
	}
	parent, err := os.OpenRoot(stateParent)
	if err != nil {
		return nil, err
	}
	var nonce [16]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		parent.Close()
		return nil, err
	}
	name := "scope-" + hex.EncodeToString(nonce[:])
	if err = parent.Mkdir(name, 0o700); err != nil {
		parent.Close()
		return nil, err
	}
	c := &scopedClient{parent: parent, stateName: name}
	ok := false
	defer func() {
		if !ok {
			_ = c.Close()
		}
	}()
	state := filepath.Join(stateParent, name)
	root := grant.Root
	if root == "" {
		root = filepath.Join(state, "empty-root")
		if err = parent.Mkdir(filepath.Join(name, "empty-root"), 0o700); err != nil {
			return nil, err
		}
	}
	// The outside sentinel is never granted to the worker. Startup must observe
	// EACCES for this existing file before any engine is opened.
	probeName := name + ".denied"
	probe, err := parent.OpenFile(probeName, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	c.probeName = probeName
	if _, err = probe.Write([]byte("outside scoped grant\n")); err != nil {
		probe.Close()
		return nil, err
	}
	if err = probe.Close(); err != nil {
		return nil, err
	}
	rootFile, err := os.Open(root)
	if err != nil {
		return nil, err
	}
	defer rootFile.Close()
	ri, err := rootFile.Stat()
	if err != nil {
		return nil, err
	}
	if rootInfo != nil && !os.SameFile(rootInfo, ri) {
		return nil, errors.New("root changed during startup")
	}
	stateFile, err := parent.Open(name)
	if err != nil {
		return nil, err
	}
	defer stateFile.Close()
	c.cmd = exec.Command(filepath.Join(bundle, "launcher"), bundle, root, state, filepath.Join(stateParent, probeName))
	// No LD_PRELOAD, HOME, Git configuration or inherited environment enters the launcher.
	c.cmd.Env = []string{"LANG=C", "PATH=/nonexistent"}
	c.cmd.Dir = state
	c.cmd.ExtraFiles = []*os.File{rootFile, stateFile}
	c.in, err = c.cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	c.out, err = c.cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	// Never inherit a caller's regular-file descriptor, including stderr.
	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err = c.cmd.Start(); err != nil {
		stderr.Close()
		return nil, fmt.Errorf("%w: launcher: %v", ErrUnavailable, err)
	}
	c.done = make(chan error, 1)
	go func() {
		c.diagnostic, _ = io.ReadAll(io.LimitReader(stderr, 16384))
		_, _ = io.Copy(io.Discard, stderr)
		c.done <- c.cmd.Wait()
	}()
	c.encode, c.decode = json.NewEncoder(c.in), json.NewDecoder(c.out)
	if _, err = c.call(ctx, scopedwire.Request{Op: "open", Documents: docs}); err != nil {
		closeErr := c.Close()
		return nil, fmt.Errorf("%w: scoped startup: %w (%v): %s", ErrUnavailable, err, closeErr, c.diagnostic)
	}
	if err = parent.Remove(probeName); err != nil {
		return nil, err
	}
	c.probeName = ""
	a := &Adapter{engine: c, root: grant.Root, rootInfo: rootInfo, sources: sources}
	if err = a.checkRoot(); err != nil {
		return nil, err
	}
	ok = true
	return a, nil
}

func checkRuntimeDirectory(path string) error {
	if strings.ContainsAny(path, ":;$") {
		return errors.New("runtime path must not contain loader separators ':' or ';', or token marker '$'")
	}
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return errors.New("runtime path must be absolute and clean")
	}
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	if real != path {
		return errors.New("runtime ancestry must not contain symlinks")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	owner, ok := info.Sys().(*syscall.Stat_t)
	if !ok || !info.IsDir() || info.Mode().Perm()&0o022 != 0 || (owner.Uid != 0 && owner.Uid != uint32(os.Geteuid())) {
		return errors.New("runtime must be owned by root or caller and not group/world writable")
	}
	return nil
}

func (c *scopedClient) call(ctx context.Context, req scopedwire.Request) (scopedwire.Response, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var response scopedwire.Response
	if err := ctx.Err(); err != nil {
		return response, err
	}
	if c.closed {
		return response, errors.New("scoped worker closed")
	}
	result := make(chan error, 1)
	go func() {
		err := c.encode.Encode(req)
		if err == nil {
			err = c.decode.Decode(&response)
		}
		result <- err
	}()
	select {
	case err := <-result:
		if err != nil {
			_ = c.cmd.Process.Kill()
			return scopedwire.Response{}, err
		}
		if response.Error != "" {
			return scopedwire.Response{}, errors.New(response.Error)
		}
		return response, nil
	case <-ctx.Done():
		c.cancelled = true
		_ = c.cmd.Process.Kill()
		_ = c.in.Close()
		_ = c.out.Close()
		<-result
		return scopedwire.Response{}, ctx.Err()
	}
}

func (c *scopedClient) Search(ctx context.Context, q search.Query) ([]search.Hit, error) {
	r, err := c.call(ctx, scopedwire.Request{Op: "search", Query: q})
	return r.Hits, err
}
func (c *scopedClient) Expand(ctx context.Context, q search.ExpandRequest) (search.Excerpt, error) {
	r, err := c.call(ctx, scopedwire.Request{Op: "expand", Expand: q})
	return r.Excerpt, err
}
func (c *scopedClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	var err error
	if c.in != nil {
		_ = c.in.Close()
	}
	if c.done == nil && c.cmd != nil {
		// Start normally closes the child pipe ends. Also release them when
		// pipe setup itself failed before Start could take ownership.
		for _, endpoint := range []any{c.cmd.Stdin, c.cmd.Stdout, c.cmd.Stderr} {
			if closer, ok := endpoint.(io.Closer); ok {
				_ = closer.Close()
			}
		}
	}
	if c.done != nil {
		select {
		case err = <-c.done:
		case <-time.After(5 * time.Second):
			_ = c.cmd.Process.Kill()
			err = <-c.done
		}
	}
	// Killing a cancelled worker is expected; the operation reports cancellation.
	var exit *exec.ExitError
	if c.cancelled && errors.As(err, &exit) && exit.Sys().(syscall.WaitStatus).Signal() == syscall.SIGKILL {
		err = nil
	}
	if c.out != nil {
		_ = c.out.Close()
	}
	if c.parent != nil {
		if c.probeName != "" {
			err = errors.Join(err, c.parent.Remove(c.probeName))
		}
		err = errors.Join(err, c.parent.RemoveAll(c.stateName), c.parent.Close())
	}
	return err
}
