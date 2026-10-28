package inference

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"strings"
	"time"
	"unicode/utf8"
)

var _ Generator = (*Owned)(nil)

// Generate never retries. Queue, load, preflight and inference share one two-
// minute deadline. Only one request can reach the runtime; four may queue.
// Source handles, grants and call-ID deduplication must be checked by the future
// harness BEFORE dispatch; this package has no dispatch or effect API.
func (o *Owned) Generate(parent context.Context, r GenerateRequest, emit Emit) (err error) {
	if emit == nil {
		return errors.New("inference: event sink required")
	}
	ctx, cancel := context.WithTimeout(parent, callLimit)
	defer cancel()
	stop := context.AfterFunc(o.life, cancel)
	defer stop()
	if !r.Deadline.IsZero() {
		var deadlineCancel context.CancelFunc
		ctx, deadlineCancel = context.WithDeadline(ctx, r.Deadline)
		defer deadlineCancel()
	}
	var seq uint64
	verification := "configured-not-verified"
	var completedTool *ToolCall
	send := func(eventCtx context.Context, e Event) error {
		seq++
		e.Sequence, e.TaskID, e.RequestID, e.AttemptID = seq, r.TaskID, r.RequestID, r.AttemptID
		e.ProfileFingerprint, e.LocalFingerprint = o.profile.Fingerprint(), o.localFingerprint
		e.Verification = verification
		return emit(eventCtx, e)
	}
	defer func() {
		kind := "completed"
		if err != nil {
			kind = "failed"
		}
		if ctx.Err() != nil {
			kind, err = "canceled", ctx.Err()
		}
		terminalCtx, done := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
		defer done()
		terminal := Event{Kind: kind, Err: err}
		if err == nil {
			terminal.Tool = completedTool
		}
		if sinkErr := send(terminalCtx, terminal); sinkErr != nil {
			err = errors.Join(err, sinkErr)
		}
	}()
	if err = send(ctx, Event{Kind: "started"}); err != nil {
		return err
	}
	body, err := o.generationBody(r)
	if err != nil {
		return err
	}
	select {
	case o.queue <- struct{}{}:
	default:
		return ErrBusy
	}
	defer func() { <-o.queue }()
	select {
	case o.run <- struct{}{}:
	case <-ctx.Done():
		return ctx.Err()
	}
	defer func() { <-o.run }()
	lease, err := o.Acquire(ctx)
	if err != nil {
		return err
	}
	defer lease.Release()
	verification = "local-artifact-bytes-verified"
	p := lease.process
	var formatted struct {
		Prompt string `json:"prompt"`
	}
	if err := p.postOrGet(ctx, "POST", "/apply-template", body, &formatted, 256<<10); err != nil {
		return err
	}
	if formatted.Prompt == "" || len(formatted.Prompt) > 48<<10 {
		return ErrLimit
	}
	var tokens struct {
		Tokens []int `json:"tokens"`
	}
	if err := p.postOrGet(ctx, "POST", "/tokenize", map[string]any{
		"content": formatted.Prompt, "add_special": true, "parse_special": true,
	}, &tokens, 256<<10); err != nil {
		return err
	}
	// Leave one extra token of headroom; no context shifting/truncation permitted.
	if len(tokens.Tokens) == 0 || len(tokens.Tokens)+r.MaxOutputTokens+1 > 8192 {
		return ErrLimit
	}
	body["stream"] = true
	body["stream_options"] = map[string]any{"include_usage": true}
	requestCtx, requestCancel := context.WithCancel(ctx)
	defer requestCancel()
	// An accepted/possibly accepted generation that does not complete is never
	// replayed or overlapped with the next request. Invalidate residency, then
	// terminate that owned process after the client body is closed. Existing
	// passive readiness leases cannot perform HTTP work. They remain releasable
	// but cannot keep failed native work alive or make a new load overlap it.
	defer func() {
		if err == nil {
			return
		}
		requestCancel() // close even an ambiguously accepted request before teardown
		o.mu.Lock()
		o.state, o.reason, o.retryAt = Failed, "generation did not complete; owned cancellation requires teardown", time.Now().Add(retryCooldown)
		o.mu.Unlock()
		p.stop()
	}()
	res, err := p.request(requestCtx, "POST", "/v1/chat/completions", body)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	media, _, parseErr := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if parseErr != nil || media != "text/event-stream" {
		return ErrProtocol
	}
	completedTool, err = consumeStream(ctx, res.Body, o.profile.Fingerprint(), r, func(e Event) error { return send(ctx, e) })
	return err
}

func (o *Owned) generationBody(r GenerateRequest) (map[string]any, error) {
	if r.Version != 1 || r.ProfileFingerprint != o.profile.Fingerprint() || r.Deadline.IsZero() ||
		r.MaxOutputTokens < 1 || r.MaxOutputTokens > 1024 || len(r.Messages) == 0 || len(r.Messages) > 32 || len(r.Tools) > 4 {
		return nil, ErrLimit
	}
	for _, id := range []string{r.TaskID, r.RequestID, r.AttemptID, r.PolicySnapshot} {
		if id == "" || len(id) > 256 || !utf8.ValidString(id) {
			return nil, ErrProtocol
		}
	}
	total := 0
	for i, m := range r.Messages {
		if m.Role != "user" && m.Role != "assistant" && !(i == 0 && m.Role == "system") {
			return nil, ErrProtocol
		}
		if !utf8.ValidString(m.Content) {
			return nil, ErrProtocol
		}
		total += len(m.Content)
	}
	if total > 32<<10 || r.Messages[len(r.Messages)-1].Role != "user" {
		return nil, ErrLimit
	}
	tools := make([]any, 0, len(r.Tools))
	names := make(map[string]bool)
	for _, t := range r.Tools {
		if !identifier(t.Name) || names[t.Name] || len(t.Description) > 1024 || !utf8.ValidString(t.Description) || len(t.RequiredStrings) > 8 {
			return nil, ErrProtocol
		}
		names[t.Name] = true
		props := make(map[string]any)
		for _, key := range t.RequiredStrings {
			if !identifier(key) || props[key] != nil {
				return nil, ErrProtocol
			}
			props[key] = map[string]string{"type": "string"}
		}
		required := append([]string{}, t.RequiredStrings...)
		tools = append(tools, map[string]any{"type": "function", "function": map[string]any{
			"name": t.Name, "description": t.Description, "parameters": map[string]any{
				"type": "object", "properties": props, "required": required, "additionalProperties": false,
			},
		}})
	}
	body := map[string]any{
		"model": o.profile.Fingerprint(), "messages": r.Messages, "max_tokens": r.MaxOutputTokens,
		"temperature": 0.1, "top_k": 50, "repeat_penalty": 1.1,
		"chat_template_kwargs": map[string]bool{"enable_thinking": false}, "parallel_tool_calls": false,
		"cache_prompt": false,
	}
	if len(tools) > 0 {
		body["tools"], body["tool_choice"] = tools, "auto"
	}
	b, err := json.Marshal(body)
	if err != nil {
		return nil, ErrProtocol
	}
	if len(b) > maxRequestBytes {
		return nil, ErrLimit
	}
	return body, nil
}

func identifier(s string) bool {
	if s == "" || len(s) > 64 {
		return false
	}
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

type streamChunk struct {
	ID, Model string
	Error     json.RawMessage
	Choices   []struct {
		Index        int
		FinishReason *string `json:"finish_reason"`
		Delta        struct {
			Role, Content    string
			ReasoningContent string `json:"reasoning_content"`
			ToolCalls        []struct {
				Index    int
				ID, Type string
				Function struct{ Name, Arguments string }
			} `json:"tool_calls"`
		}
	}
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
}

// consumeStream accepts llama.cpp's single-data-line SSE frames, not an
// arbitrary event protocol. EOF is failure unless finish + usage + DONE agree.
func consumeStream(ctx context.Context, body io.Reader, model string, r GenerateRequest, send func(Event) error) (*ToolCall, error) {
	s := bufio.NewScanner(io.LimitReader(body, (1<<20)+1))
	s.Buffer(make([]byte, 4096), 64<<10)
	var data, responseID, finish, toolID, name, arguments string
	var toolSeen, usageSeen bool
	bytesRead, frames, textBytes := 0, 0, 0
	for s.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		line := s.Text()
		bytesRead += len(line) + 1
		if bytesRead > 1<<20 {
			return nil, ErrLimit
		}
		if strings.HasPrefix(line, ":") {
			continue
		}
		if line != "" {
			if !strings.HasPrefix(line, "data:") || data != "" {
				return nil, ErrProtocol
			}
			data = strings.TrimPrefix(line[5:], " ")
			if data == "" {
				return nil, ErrProtocol
			}
			continue
		}
		if data == "" {
			continue
		}
		frames++
		if frames > 4096 {
			return nil, ErrLimit
		}
		if data == "[DONE]" {
			if finish == "" || !usageSeen {
				return nil, ErrProtocol
			}
			if finish == "length" {
				return nil, ErrLimit
			}
			if toolSeen {
				if finish != "tool_calls" {
					return nil, ErrProtocol
				}
				return validateTool(toolID, name, arguments, r.Tools)
			}
			if finish != "stop" {
				return nil, ErrProtocol
			}
			return nil, nil
		}
		var chunk streamChunk
		if err := decodeJSON([]byte(data), &chunk); err != nil {
			return nil, err
		}
		data = ""
		if len(chunk.Error) != 0 || chunk.Model != model || len(chunk.ID) > 256 || chunk.ID == "" || len(chunk.Choices) > 1 {
			return nil, ErrProtocol
		}
		if responseID == "" {
			responseID = chunk.ID
		}
		if responseID != chunk.ID {
			return nil, ErrProtocol
		}
		if chunk.Usage != nil {
			u := chunk.Usage
			if usageSeen || u.PromptTokens <= 0 || u.PromptTokens > 8192 || u.CompletionTokens < 0 || u.CompletionTokens > r.MaxOutputTokens || u.PromptTokens+u.CompletionTokens > 8192 {
				return nil, ErrLimit
			}
			usageSeen = true
			if err := send(Event{Kind: "usage", InputTokens: u.PromptTokens, OutputTokens: u.CompletionTokens}); err != nil {
				return nil, err
			}
		}
		if len(chunk.Choices) == 0 {
			if chunk.Usage == nil {
				return nil, ErrProtocol
			}
			continue
		}
		c := chunk.Choices[0]
		if c.Index != 0 || finish != "" || (c.Delta.Role != "" && c.Delta.Role != "assistant") || c.Delta.ReasoningContent != "" || len(c.Delta.ToolCalls) > 1 {
			return nil, ErrProtocol
		}
		if c.Delta.Content != "" {
			textBytes += len(c.Delta.Content)
			if textBytes > 256<<10 {
				return nil, ErrLimit
			}
			if err := send(Event{Kind: "text_delta", Text: c.Delta.Content}); err != nil {
				return nil, err
			}
		}
		for _, tc := range c.Delta.ToolCalls {
			if tc.Index != 0 || (tc.Type != "" && tc.Type != "function") {
				return nil, ErrProtocol
			}
			if tc.ID != "" {
				if toolID != "" && toolID != tc.ID {
					return nil, ErrProtocol
				}
				toolID = tc.ID
			}
			toolSeen = true
			name += tc.Function.Name
			arguments += tc.Function.Arguments
			if len(toolID) > 128 || len(name) > 64 || len(arguments) > 8192 {
				return nil, ErrLimit
			}
			if err := send(Event{Kind: "tool_call_delta", Text: tc.Function.Arguments}); err != nil {
				return nil, err
			}
		}
		if c.FinishReason != nil {
			finish = *c.FinishReason
			if finish == "" {
				return nil, ErrProtocol
			}
		}
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return nil, ErrProtocol // short read, oversized frame, malformed SSE or missing terminal
}

func validateTool(id, name, arguments string, tools []Tool) (*ToolCall, error) {
	if !identifier(id) || !identifier(name) {
		return nil, ErrProtocol
	}
	var schema *Tool
	for i := range tools {
		if tools[i].Name == name {
			schema = &tools[i]
			break
		}
	}
	if schema == nil {
		return nil, ErrProtocol
	}
	var raw map[string]any
	if err := decodeJSON([]byte(arguments), &raw); err != nil {
		return nil, err
	}
	if raw == nil || len(raw) != len(schema.RequiredStrings) {
		return nil, ErrProtocol
	}
	args := make(map[string]string, len(raw))
	for _, key := range schema.RequiredStrings {
		value, ok := raw[key].(string)
		if !ok {
			return nil, ErrProtocol
		}
		args[key] = value
	}
	return &ToolCall{ID: id, Name: name, Arguments: args}, nil
}
