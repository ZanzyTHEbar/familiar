package inference

import (
	"context"
	"errors"
	"time"
)

var (
	ErrBusy     = errors.New("inference: queue full")
	ErrClosed   = errors.New("inference: owner closed")
	ErrProtocol = errors.New("inference: invalid or unsupported runtime response")
	ErrLimit    = errors.New("inference: budget exceeded")
)

// Envelope carries references, not authority. The caller owns task-wide limits
// (four calls, ten minutes, depth/fanout one), grants and tool dispatch. This
// layer checks per-call limits including admission and loading time.
type Envelope struct {
	Version                                      int
	TaskID, RequestID, AttemptID, PolicySnapshot string
	ProfileFingerprint                           string
	Deadline                                     time.Time
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Tool is a deliberately small schema: a closed object of required strings.
// No arbitrary JSON Schema dialect, optional/nested arguments or effects.
// These are read-only proposals; the inference layer never dispatches them.
type Tool struct {
	Name, Description string
	RequiredStrings   []string
}

type GenerateRequest struct {
	Envelope
	Messages        []Message
	Tools           []Tool
	MaxOutputTokens int
}

type ToolCall struct {
	ID, Name  string
	Arguments map[string]string
}

// Event has a monotonic sequence and immutable local provenance. Tool fragments
// are untrusted until the completed event contains a validated ToolCall.
type Event struct {
	Sequence                             uint64
	Kind                                 string // started, text_delta, tool_call_delta, usage, completed, failed, canceled
	TaskID, RequestID, AttemptID         string
	ProfileFingerprint, LocalFingerprint string
	Verification                         string // configured-not-verified or local-artifact-bytes-verified
	Text                                 string
	Tool                                 *ToolCall
	InputTokens, OutputTokens            int
	Err                                  error
}

// Emit is synchronous and must return promptly and honor ctx; backpressure is
// intentional. No detached event goroutines or unbounded buffering are used.
type Emit func(ctx context.Context, event Event) error

type Generator interface {
	Generate(context.Context, GenerateRequest, Emit) error
}

// Separate future ports; no decision/embedding HTTP adapter is implemented in
// this generation-first slice. Decisions use /v1/systemone, NOT chat completion.
type DecideRequest struct {
	Envelope
	Question string
	Choices  []string
}
type Decision struct {
	Choice      string
	Probability float64
}
type Decider interface {
	Decide(context.Context, DecideRequest) (Decision, error)
}
type EmbedRequest struct {
	Envelope
	Text, Title string
	Query       bool
}
type Embedding struct {
	ProfileFingerprint string
	Vector             []float32
}
type Embedder interface {
	Embed(context.Context, EmbedRequest) (Embedding, error)
}
