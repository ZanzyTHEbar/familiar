package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"unicode/utf8"
)

const maxRequestBytes = 64 << 10

func (p *localProcess) request(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var b []byte
	if body != nil {
		var err error
		b, err = json.Marshal(body)
		if err != nil {
			return nil, ErrProtocol
		}
		if len(b) > maxRequestBytes {
			return nil, ErrLimit
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, p.url+path, bytes.NewReader(b))
	if err != nil {
		return nil, ErrProtocol
	}
	req.Header.Set("Authorization", "Bearer "+p.key)
	req.Header.Set("Content-Type", "application/json")
	res, err := p.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("inference: owned loopback transport failed")
	}
	if res.StatusCode != http.StatusOK {
		res.Body.Close() // never put response/prompt content in diagnostics
		return nil, errors.New("inference: owned runtime rejected request")
	}
	return res, nil
}

func (p *localProcess) postOrGet(ctx context.Context, method, path string, body, out any, limit int64) error {
	res, err := p.request(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	media, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		return ErrProtocol
	}
	b, err := io.ReadAll(io.LimitReader(res.Body, limit+1))
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return ErrProtocol
	}
	if int64(len(b)) > limit {
		return ErrLimit
	}
	return decodeJSON(b, out)
}

// Validate depth and duplicate keys before decoding untrusted runtime JSON.
// Byte/frame/call deadlines are bounded by the caller, independent of depth.
func decodeJSON(b []byte, out any) error {
	if !utf8.Valid(b) {
		return ErrProtocol
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.UseNumber()
	if err := walkJSON(d, 0); err != nil {
		return ErrProtocol
	}
	if _, err := d.Token(); err != io.EOF {
		return ErrProtocol
	}
	if err := json.Unmarshal(b, out); err != nil {
		return ErrProtocol
	}
	return nil
}

func walkJSON(d *json.Decoder, depth int) error {
	if depth > 16 {
		return ErrLimit
	}
	t, err := d.Token()
	if err != nil {
		return err
	}
	delim, composite := t.(json.Delim)
	if !composite {
		return nil
	}
	switch delim {
	case '{':
		seen := make(map[string]bool)
		for d.More() {
			key, err := d.Token()
			if err != nil {
				return err
			}
			s, ok := key.(string)
			if !ok || seen[s] {
				return ErrProtocol
			}
			seen[s] = true
			if err := walkJSON(d, depth+1); err != nil {
				return err
			}
		}
		end, err := d.Token()
		if err != nil || end != json.Delim('}') {
			return ErrProtocol
		}
	case '[':
		for d.More() {
			if err := walkJSON(d, depth+1); err != nil {
				return err
			}
		}
		end, err := d.Token()
		if err != nil || end != json.Delim(']') {
			return ErrProtocol
		}
	default:
		return ErrProtocol
	}
	return nil
}
