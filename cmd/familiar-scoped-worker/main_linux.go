// The worker is private to the launcher; library consumers use retrieval.Open.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"familiar.local/familiar/internal/scopedwire"
	"github.com/ZanzyTHEbar/virtual-vectorfs/search"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 4 || os.Getenv("FAMILIAR_SCOPED_CHILD") != "1" {
		return errors.New("private worker requires scoped launcher")
	}
	root, state, probe := os.Args[1], os.Args[2], os.Args[3]
	// An existing outside file must be denied by the actual inherited policy.
	f, err := os.Open(probe)
	if err == nil {
		f.Close()
		return errors.New("outside read was not confined")
	}
	if !errors.Is(err, syscall.EACCES) {
		return fmt.Errorf("cannot verify boundary: %w", err)
	}
	decode, encode := json.NewDecoder(os.Stdin), json.NewEncoder(os.Stdout)
	var engine *search.Engine
	defer func() {
		if engine != nil {
			_ = engine.Close()
		}
	}()
	for {
		var req scopedwire.Request
		if err = decode.Decode(&req); err == io.EOF {
			if engine != nil {
				return engine.Close()
			}
			return nil
		} else if err != nil {
			return err
		}
		var response scopedwire.Response
		switch req.Op {
		case "open":
			if engine != nil {
				err = errors.New("already opened")
				break
			}
			engine, err = search.Open(context.Background(), search.Options{Root: root, IndexPath: filepath.Join(state, "index"), Storage: search.StorageMemory, VectorDimensions: 1, MaxFileBytes: 8 << 20, Workers: 2, QueueCapacity: 32, WriteBatchBytes: 1 << 20})
			if err == nil {
				_, err = engine.SyncFiles(context.Background())
			}
			if err == nil {
				err = engine.Upsert(context.Background(), req.Documents...)
			}
		case "search":
			if engine == nil {
				err = errors.New("not opened")
				break
			}
			response.Hits, err = engine.Search(context.Background(), req.Query)
		case "expand":
			if engine == nil {
				err = errors.New("not opened")
				break
			}
			response.Excerpt, err = engine.Expand(context.Background(), req.Expand)
		default:
			err = errors.New("unknown worker operation")
		}
		if err != nil {
			response = scopedwire.Response{Error: err.Error()}
		}
		if err = encode.Encode(response); err != nil {
			return err
		}
	}
}
