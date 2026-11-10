package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ContextUsage is a local observation, not a cache or compaction claim.
type ContextUsage struct {
	Method        string `json:"method"`
	InputTokens   int    `json:"input_tokens,omitempty"`
	OutputReserve int    `json:"output_reserve"`
	ContextTokens int    `json:"context_tokens,omitempty"`
	RequestBytes  int    `json:"request_bytes"`
	ByteLimit     int    `json:"byte_limit"`
	Over          bool   `json:"over_limit"`
}

// ContextLimitError occurs before generation. Callers may select fewer whole
// historical tasks, but must not retry an ambiguously accepted generation.
type ContextLimitError struct{ Usage ContextUsage }

func (e *ContextLimitError) Error() string {
	u := e.Usage
	if u.InputTokens > 0 {
		return fmt.Sprintf("context budget exceeded before generation: %d input + %d output reserve + 1 headroom > %d tokens, or %d request bytes > %d; narrow the task or select a larger compatible runtime", u.InputTokens, u.OutputReserve, u.ContextTokens, u.RequestBytes, u.ByteLimit)
	}
	return fmt.Sprintf("context byte budget exceeded before generation: %d > %d bytes; tokenization not performed for this request; narrow the task or configure the byte ceiling", u.RequestBytes, u.ByteLimit)
}

func (l *Local) jsonRequest(ctx context.Context, path string, body any, target any, limit int64) error {
	var data []byte
	var err error
	method := "GET"
	if body != nil {
		method = "POST"
		data, err = json.Marshal(body)
		if err != nil {
			return err
		}
	}
	base := strings.TrimSuffix(l.endpoint, "/v1/chat/completions")
	req, err := http.NewRequestWithContext(ctx, method, base+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := l.client.Do(req)
	if err != nil {
		return errors.New("local runtime context preflight failed (connection, deadline or cancellation); no generation sent")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("local runtime context preflight %s HTTP %d; body withheld, no generation sent", path, res.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(res.Body, limit+1))
	if err != nil {
		return err
	}
	if int64(len(b)) > limit {
		return ErrLimit
	}
	if json.Unmarshal(b, target) != nil {
		return fmt.Errorf("invalid local runtime %s response; no generation sent", path)
	}
	return nil
}

func (l *Local) measure(ctx context.Context, body map[string]any) (ContextUsage, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return ContextUsage{}, err
	}
	u := ContextUsage{Method: "explicit-byte-ceiling; runtime token count unknown", OutputReserve: l.outputTokens, RequestBytes: len(data), ByteLimit: l.contextBytes}
	if !l.llama {
		// This is deliberately NOT a chars/4 token estimate. The operator controls
		// the byte ceiling and external model's token window; runtime errors remain
		// errors. Unknown provider tokenizers cannot be guaranteed by this adapter.
		u.Over = len(data) > l.contextBytes
		return u, nil
	}
	// Byte cap bounds preflight work too, independent of native token admission.
	if len(data) > l.contextBytes {
		u.Over = true
		return u, nil
	}
	var props struct {
		TotalSlots int `json:"total_slots"`
		Default    struct {
			Context int `json:"n_ctx"`
		} `json:"default_generation_settings"`
	}
	if err = l.jsonRequest(ctx, "/props", nil, &props, 256<<10); err != nil {
		return u, err
	}
	if props.TotalSlots != 1 || props.Default.Context < 1 {
		return u, errors.New("llama.cpp context discovery requires a reported single-slot n_ctx; no guessed slot/context size or generation")
	}
	limit := props.Default.Context
	if l.contextTokens > 0 {
		limit = min(limit, l.contextTokens)
	}
	var formatted struct {
		Prompt string `json:"prompt"`
	}
	if err = l.jsonRequest(ctx, "/apply-template", body, &formatted, 512<<10); err != nil {
		return u, err
	}
	if len(formatted.Prompt) == 0 || len(formatted.Prompt) > 256<<10 {
		return u, ErrLimit
	}
	var tokenized struct {
		Tokens []int `json:"tokens"`
	}
	if err = l.jsonRequest(ctx, "/tokenize", map[string]any{"content": formatted.Prompt, "add_special": true, "parse_special": true}, &tokenized, 2<<20); err != nil {
		return u, err
	}
	if len(tokenized.Tokens) == 0 {
		return u, ErrProtocol
	}
	u.Method = "llama.cpp native template + tokenizer"
	u.InputTokens, u.ContextTokens = len(tokenized.Tokens), limit
	u.Over = u.InputTokens+u.OutputReserve+1 > limit
	return u, nil
}
