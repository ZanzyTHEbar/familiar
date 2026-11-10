package inference

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Local is an explicitly selected, externally managed local endpoint. It never
// discovers servers, reads API keys, uses proxies, follows redirects or retries.
// The operator, not Familiar, owns runtime containment and residency.
type Local struct {
	client                                    *http.Client
	endpoint, model                           string
	llama                                     bool
	thinking                                  string
	cache                                     bool
	contextTokens, outputTokens, contextBytes int
}

type EndpointConfig struct {
	Endpoint, Model, Thinking                 string
	Llama, Cache                              bool
	ContextTokens, OutputTokens, ContextBytes int
}

func NewLocal(c EndpointConfig) (*Local, error) {
	endpoint, model, thinking, llama, cache := c.Endpoint, c.Model, c.Thinking, c.Llama, c.Cache
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "http" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "/v1" && u.Path != "/v1/") {
		return nil, errors.New("endpoint must be http://127.0.0.1:PORT/v1 (or localhost / [::1]); no credentials, query or other path")
	}
	host := u.Hostname()
	if host == "localhost" {
		host = "127.0.0.1"
	}
	if host != "127.0.0.1" && host != "::1" {
		return nil, errors.New("only loopback endpoints are supported; LAN/cloud unavailable")
	}
	p, err := strconv.Atoi(u.Port())
	if err != nil || p < 1 || p > 65535 || model == "" || len(model) > 256 {
		return nil, errors.New("explicit port and model required")
	}
	if thinking != "runtime" && thinking != "on" && thinking != "off" {
		return nil, errors.New("thinking must be runtime, on or off")
	}
	if !llama && (thinking != "runtime" || cache) {
		return nil, errors.New("thinking/cache overrides require --llama; generic endpoints use runtime defaults")
	}
	if !llama && c.ContextTokens != 0 {
		return nil, errors.New("exact context token preflight requires --llama; generic endpoints use --context-bytes and their own runtime token limit")
	}
	if c.OutputTokens < 1 || c.OutputTokens > 4096 || c.ContextTokens < 0 || c.ContextTokens > 131072 || c.ContextBytes < 1024 || c.ContextBytes > 128<<10 {
		return nil, errors.New("output tokens must be 1..4096, context tokens 0..131072, context bytes 1024..131072")
	}
	address := net.JoinHostPort(host, u.Port())
	u.Host = address
	u.Path = "/v1/chat/completions"
	tr := &http.Transport{
		Proxy: nil,
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "tcp", address)
		},
		ResponseHeaderTimeout: 30 * time.Second, MaxResponseHeaderBytes: 16 << 10,
	}
	return &Local{client: &http.Client{Transport: tr, CheckRedirect: func(*http.Request, []*http.Request) error { return errors.New("redirect refused") }}, endpoint: u.String(), model: model, thinking: thinking, llama: llama, cache: cache, contextTokens: c.ContextTokens, outputTokens: c.OutputTokens, contextBytes: c.ContextBytes}, nil
}

func (l *Local) Close() { l.client.CloseIdleConnections() }

func (l *Local) Turn(ctx context.Context, messages []Message, tools []Tool, emit func(string, string) error) (Message, error) {
	var zero Message
	body := map[string]any{"model": l.model, "messages": messages, "stream": true, "max_tokens": l.outputTokens, "temperature": 0.1, "stream_options": map[string]bool{"include_usage": true}}
	schemas := []any{}
	for _, t := range tools {
		props := map[string]any{}
		for _, key := range t.RequiredStrings {
			props[key] = map[string]string{"type": "string"}
		}
		schemas = append(schemas, map[string]any{"type": "function", "function": map[string]any{"name": t.Name, "description": t.Description, "parameters": map[string]any{"type": "object", "properties": props, "required": t.RequiredStrings, "additionalProperties": false}}})
	}
	body["tools"], body["tool_choice"] = schemas, "auto"
	if l.llama {
		body["cache_prompt"] = l.cache
		body["top_k"], body["repeat_penalty"] = 50, 1.1
		body["parallel_tool_calls"] = false
		body["reasoning_format"] = "deepseek"
		body["chat_template_kwargs"] = map[string]any{"preserve_thinking": false}
		if l.thinking != "runtime" {
			body["chat_template_kwargs"].(map[string]any)["enable_thinking"] = l.thinking == "on"
		}
	}
	budget, err := l.measure(ctx, body)
	if err != nil {
		return zero, err
	}
	observation, _ := json.Marshal(budget)
	if err := emit("context_usage", string(observation)); err != nil {
		return zero, err
	}
	if budget.Over {
		return zero, &ContextLimitError{Usage: budget}
	}
	data, err := json.Marshal(body)
	if err != nil {
		return zero, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", l.endpoint, bytes.NewReader(data))
	if err != nil {
		return zero, err
	}
	req.Header.Set("Content-Type", "application/json")
	if err := emit("generation_start", ""); err != nil {
		return zero, err
	}
	res, err := l.client.Do(req)
	if err != nil {
		return zero, errors.New("local model request failed (connection, timeout or cancellation); no retry")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return zero, fmt.Errorf("local model HTTP %d; response body withheld; check selected runtime/model/context", res.StatusCode)
	}
	media, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if media != "text/event-stream" {
		return zero, ErrProtocol
	}
	return conversationStream(io.LimitReader(res.Body, 1<<20), emit)
}

func conversationStream(r io.Reader, emit func(string, string) error) (Message, error) {
	m := Message{Role: "assistant"}
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 4096), 64<<10)
	finish := ""
	count, size := 0, 0
	for scanner.Scan() {
		line := scanner.Text()
		count++
		size += len(line)
		if count > 16384 || size > 1<<20 {
			return m, ErrLimit
		}
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data: ") {
			return m, ErrProtocol
		}
		raw := strings.TrimPrefix(line, "data: ")
		if raw == "[DONE]" {
			if finish != "stop" && finish != "tool_calls" {
				return m, errors.New("model did not complete (length/content filter/invalid finish); partial output only")
			}
			if (finish == "tool_calls") != (len(m.ToolCalls) > 0) {
				return m, ErrProtocol
			}
			if m.Content == "" && len(m.ToolCalls) == 0 {
				return m, ErrProtocol
			}
			ids := map[string]bool{}
			for _, c := range m.ToolCalls {
				if c.ID == "" || len(c.ID) > 256 || ids[c.ID] || c.Type != "function" || c.Function.Name == "" || len(c.Function.Name) > 128 || len(c.Function.Arguments) > 8192 {
					return m, ErrProtocol
				}
				ids[c.ID] = true
			}
			return m, nil
		}
		var frame struct {
			Error   json.RawMessage `json:"error"`
			Usage   json.RawMessage `json:"usage"`
			Choices []struct {
				Index int `json:"index"`
				Delta struct {
					Role      string `json:"role"`
					Content   string `json:"content"`
					Reasoning string `json:"reasoning_content"`
					Calls     []struct {
						Index    int          `json:"index"`
						ID       string       `json:"id"`
						Type     string       `json:"type"`
						Function FunctionCall `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
				Finish string `json:"finish_reason"`
			} `json:"choices"`
		}
		if json.Unmarshal([]byte(raw), &frame) != nil || len(frame.Error) > 0 || len(frame.Choices) > 1 {
			return m, ErrProtocol
		}
		if len(frame.Usage) > 0 && string(frame.Usage) != "null" {
			if err := emit("usage", string(frame.Usage)); err != nil {
				return m, err
			}
		}
		for _, c := range frame.Choices {
			if c.Index != 0 || finish != "" || (c.Delta.Role != "" && c.Delta.Role != "assistant") {
				return m, ErrProtocol
			}
			m.Content += c.Delta.Content
			m.ReasoningContent += c.Delta.Reasoning
			if len(m.Content)+len(m.ReasoningContent) > 64<<10 {
				return m, ErrLimit
			}
			if c.Delta.Content != "" {
				if err := emit("text_delta", c.Delta.Content); err != nil {
					return m, err
				}
			}
			if c.Delta.Reasoning != "" {
				if err := emit("reasoning_delta", c.Delta.Reasoning); err != nil {
					return m, err
				}
			}
			for _, d := range c.Delta.Calls {
				if d.Index < 0 || d.Index > 7 || d.Index > len(m.ToolCalls) {
					return m, ErrProtocol
				}
				if d.Index == len(m.ToolCalls) {
					m.ToolCalls = append(m.ToolCalls, ConversationCall{})
				}
				t := &m.ToolCalls[d.Index]
				t.ID += d.ID
				t.Type += d.Type
				t.Function.Name += d.Function.Name
				t.Function.Arguments += d.Function.Arguments
				if len(t.ID) > 256 || len(t.Function.Name) > 128 || len(t.Function.Arguments) > 8192 {
					return m, ErrLimit
				}
				if err := emit("tool_call_delta", fmt.Sprintf("%d:%s%s", d.Index, d.Function.Name, d.Function.Arguments)); err != nil {
					return m, err
				}
			}
			if c.Finish != "" {
				finish = c.Finish
			}
		}
	}
	return m, errors.New("model stream interrupted before [DONE]; output is partial, no tools dispatched")
}
