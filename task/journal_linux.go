package task

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"familiar.local/familiar/inference"
	"golang.org/x/sys/unix"
)

type selection struct {
	Workspace     string `json:"workspace"`
	Endpoint      string `json:"endpoint"`
	OwnedRuntime  string `json:"owned_runtime,omitempty"`
	CreateFiles   bool   `json:"create_files,omitempty"`
	Model         string `json:"model"`
	PromptVersion string `json:"prompt_version"`
	PromptHash    string `json:"prompt_hash"`
	ToolVersion   string `json:"tool_version"`
	Thinking      string `json:"thinking"`
	Llama         bool   `json:"llama"`
	Cache         bool   `json:"cache"`
	ContextTokens int    `json:"context_tokens,omitempty"`
	OutputTokens  int    `json:"output_tokens"`
	ContextBytes  int    `json:"context_bytes"`
}

type event struct {
	Version   int                         `json:"version"`
	Sequence  int                         `json:"sequence"`
	At        time.Time                   `json:"at"`
	Kind      string                      `json:"kind"`
	Task      string                      `json:"task,omitempty"`
	Text      string                      `json:"text,omitempty"`
	Selection *selection                  `json:"selection,omitempty"`
	Message   *inference.Message          `json:"message,omitempty"`
	Call      *inference.ConversationCall `json:"call,omitempty"`
}

type journal struct {
	f      *os.File
	events []event
	size   int64
	failed error
}

func randomID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func openJournal(dir, id string, create bool) (*journal, error) {
	if !filepath.IsAbs(dir) {
		return nil, errors.New("state directory must be absolute")
	}
	if !regexp.MustCompile(`^[a-f0-9]{32}$`).MatchString(id) {
		return nil, errors.New("session ID must be 32 lowercase hex characters")
	}
	// Open directory without following symlinks in any component. State is local
	// sensitive plaintext; refuse shared or foreign-owned directories/files.
	base, err := unix.Open("/", unix.O_RDONLY|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	defer unix.Close(base)
	if create {
		// Create components relative to already-open directories, never through a
		// symlink. Existing ancestors need not be private; the final directory must.
		current, err := unix.Dup(base)
		if err != nil {
			return nil, err
		}
		for _, part := range strings.Split(strings.TrimPrefix(filepath.Clean(dir), "/"), "/") {
			if err = unix.Mkdirat(current, part, 0700); err != nil && err != unix.EEXIST {
				unix.Close(current)
				return nil, err
			}
			next, e := unix.Openat(current, part, unix.O_RDONLY|unix.O_DIRECTORY|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0)
			unix.Close(current)
			if e != nil {
				return nil, errors.New("state ancestry must be directories, not symlinks")
			}
			current = next
		}
		unix.Close(current)
	}
	dfd, err := unix.Openat2(base, filepath.Clean(dir)[1:], &unix.OpenHow{Flags: unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC, Resolve: unix.RESOLVE_BENEATH | unix.RESOLVE_NO_SYMLINKS})
	if err != nil {
		return nil, errors.New("state directory unavailable; symlinks not allowed")
	}
	defer unix.Close(dfd)
	var st unix.Stat_t
	if unix.Fstat(dfd, &st) != nil || st.Uid != uint32(os.Geteuid()) || st.Mode&0777 != 0700 {
		return nil, errors.New("state directory must be owned by this user with mode 0700")
	}
	flags := unix.O_RDWR | unix.O_APPEND | unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_NONBLOCK
	if create {
		flags |= unix.O_CREAT | unix.O_EXCL
	}
	fd, err := unix.Openat(dfd, id+".jsonl", flags, 0600)
	if err != nil {
		return nil, errors.New("cannot open session (missing, existing or denied)")
	}
	f := os.NewFile(uintptr(fd), id)
	fail := func(err error) (*journal, error) { f.Close(); return nil, err }
	if unix.Fstat(fd, &st) != nil || st.Uid != uint32(os.Geteuid()) || st.Mode&0777 != 0600 || st.Mode&unix.S_IFMT != unix.S_IFREG || st.Nlink != 1 {
		return fail(errors.New("journal must be a private 0600 regular file with one link"))
	}
	if err := unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err != nil {
		return fail(errors.New("session is already open; concurrent access refused"))
	}
	if st.Size > 16<<20 {
		return fail(errors.New("journal exceeds 16 MiB limit; retain it and start a new session"))
	}
	j := &journal{f: f, size: st.Size}
	r := bufio.NewReader(f)
	for {
		line, err := r.ReadBytes('\n')
		if len(line) == 0 && err == io.EOF {
			break
		}
		if err != nil {
			return fail(errors.New("journal has interrupted/torn tail; preserved unchanged, no continuation; inspect retained file and start a new session"))
		}
		var e event
		if json.Unmarshal(line, &e) != nil || e.Version != 1 || e.Sequence != len(j.events)+1 {
			return fail(errors.New("invalid journal; preserved unchanged"))
		}
		j.events = append(j.events, e)
	}
	if create {
		if err := unix.Fsync(dfd); err != nil {
			return fail(err)
		}
	} else if len(j.events) == 0 || j.events[0].Kind != "session" || j.events[0].Selection == nil {
		return fail(errors.New("journal has no session header"))
	}
	return j, nil
}

func (j *journal) append(e event) (err error) {
	if j.failed != nil {
		return j.failed
	}
	defer func() {
		if err != nil {
			j.failed = err
		}
	}()
	e.Version = 1
	e.Sequence = len(j.events) + 1
	e.At = time.Now().UTC()
	b, err := json.Marshal(e)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	if j.size+int64(len(b)) > 16<<20 {
		return errors.New("journal limit reached; no further work dispatched")
	}
	n, err := j.f.Write(b)
	if err != nil {
		return err
	}
	if n != len(b) {
		return io.ErrShortWrite
	}
	if err := j.f.Sync(); err != nil {
		return err
	}
	j.size += int64(n)
	j.events = append(j.events, e)
	return nil
}

func (j *journal) transcript(out io.Writer) {
	pending := map[string]bool{}
	tasks := map[string]bool{}
	partials := map[string]string{}
	for _, e := range j.events {
		if e.Kind == "text_delta" {
			partials[e.Task] += e.Text
		}
		if e.Kind == "message" && e.Message != nil && e.Message.Role == "assistant" {
			delete(partials, e.Task)
		}
		if e.Kind == "message" && e.Message != nil && e.Message.Role == "user" {
			tasks[e.Task] = true
		}
		if e.Kind == "task_completed" || e.Kind == "task_failed" || e.Kind == "task_canceled" {
			delete(tasks, e.Task)
		}
		switch e.Kind {
		case "session":
			fmt.Fprintf(out, "Session workspace %q; model %q; endpoint %q; prompt %s %s\n", e.Selection.Workspace, e.Selection.Model, e.Selection.Endpoint, e.Selection.PromptVersion, e.Selection.PromptHash)
		case "message":
			if e.Message != nil {
				fmt.Fprintf(out, "[%s completed] %s\n", e.Message.Role, display(e.Message.Content))
				for _, c := range e.Message.ToolCalls {
					fmt.Fprintf(out, "[proposal %q] %q %s\n", c.ID, c.Function.Name, display(c.Function.Arguments))
				}
			}
		case "tool_intent":
			if e.Call != nil {
				pending[e.Task+"/"+e.Call.ID] = true
				fmt.Fprintf(out, "[tool intent %q] %q %s\n", e.Call.ID, e.Call.Function.Name, display(e.Call.Function.Arguments))
			}
		case "action_proposal", "action_decision":
			if e.Call != nil {
				fmt.Fprintf(out, "[%s %q] %q %s %s\n", e.Kind, e.Call.ID, e.Call.Function.Name, display(e.Call.Function.Arguments), display(e.Text))
			}
		case "tool_result":
			if e.Message != nil {
				delete(pending, e.Task+"/"+e.Message.ToolCallID)
				fmt.Fprintf(out, "[tool result %q] %s\n", e.Message.ToolCallID, display(e.Message.Content))
			}
		case "task_failed", "task_canceled", "context_boundary", "context_selection", "context_usage", "runtime_start", "runtime_ready", "runtime_failed", "runtime_stopped":
			fmt.Fprintf(out, "[%s] %s\n", e.Kind, display(e.Text))
		}
	}
	for id := range pending {
		fmt.Fprintf(out, "[uncertain outcome %q] intent without durable result; never replayed\n", id)
	}
	for id := range tasks {
		fmt.Fprintf(out, "[interrupted task %q] no durable terminal event; not assumed completed and never replayed\n", id)
	}
	for _, text := range partials {
		fmt.Fprintf(out, "[partial, not completed] %s\n", display(text))
	}
}

func (j *journal) completedTasks() map[string]bool {
	completed := map[string]bool{}
	for _, e := range j.events {
		if e.Kind == "task_completed" {
			completed[e.Task] = true
		}
	}
	return completed
}

type historyTask struct {
	ID       string
	Messages []inference.Message
}

// Projection uses only complete tasks, without altering durable messages. Past
// reasoning is not needed by the next user turn; active tool reasoning is kept.
func (j *journal) historyTasks() []historyTask {
	completed := j.completedTasks()
	var tasks []historyTask
	for _, e := range j.events {
		if (e.Kind == "message" || e.Kind == "tool_result") && e.Message != nil && completed[e.Task] {
			if len(tasks) == 0 || tasks[len(tasks)-1].ID != e.Task {
				tasks = append(tasks, historyTask{ID: e.Task})
			}
			m := *e.Message
			m.ReasoningContent = ""
			tasks[len(tasks)-1].Messages = append(tasks[len(tasks)-1].Messages, m)
		}
	}
	return tasks
}
