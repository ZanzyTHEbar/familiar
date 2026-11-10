//go:build !linux

package retrieval

import "context"

func openScoped(context.Context, Grant, []Record, string) (*Adapter, error) {
	return nil, ErrUnavailable
}
