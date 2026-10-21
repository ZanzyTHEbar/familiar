// familiar-baseline measures retrieval over verified synthetic scratch fixtures.
// It is not an agent, an evidence adapter, or a model qualification command.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/ZanzyTHEbar/virtual-vectorfs/search"
)

type probe struct {
	ID    string `json:"id"`
	Lane  string `json:"lane"`
	Query string `json:"query"`
}

type sample struct {
	ID      string   `json:"id"`
	Round   int      `json:"round"`
	Paths   []string `json:"paths"`
	Elapsed int64    `json:"elapsed_ns"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	root := flag.String("root", "", "explicit verified synthetic scratch input directory")
	state := flag.String("state", "", "explicit empty scratch state directory outside input")
	flag.Parse()
	if flag.NArg() != 0 || *root == "" || *state == "" {
		return errors.New("usage: familiar-baseline --root ABSOLUTE --state ABSOLUTE < probes.json")
	}
	for _, p := range []string{*root, *state} {
		resolved, err := filepath.EvalSymlinks(p)
		if err != nil || !filepath.IsAbs(p) || resolved != filepath.Clean(p) {
			return fmt.Errorf("require absolute nonsymlink existing directory: %q", p)
		}
	}
	rel, err := filepath.Rel(*root, *state)
	if err != nil || rel == "." || filepath.IsLocal(rel) {
		return errors.New("state must be outside input root")
	}
	entries, err := os.ReadDir(*state)
	if err != nil || len(entries) != 0 {
		return errors.New("state must be an existing empty scratch directory")
	}
	var probes []probe
	d := json.NewDecoder(io.LimitReader(os.Stdin, 1<<20))
	d.DisallowUnknownFields()
	if err := d.Decode(&probes); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return errors.New("expected one JSON probe array")
	}
	if len(probes) == 0 || len(probes) > 100 {
		return errors.New("require 1..100 probes")
	}
	ids := make(map[string]bool)
	for _, p := range probes {
		if p.ID == "" || ids[p.ID] || p.Query == "" || (p.Lane != "keyword" && p.Lane != "path" && p.Lane != "grep") {
			return fmt.Errorf("invalid probe: %+v", p)
		}
		ids[p.ID] = true
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	start := time.Now()
	engine, err := search.Open(ctx, search.Options{
		Root: *root, IndexPath: filepath.Join(*state, "index"), Storage: search.StorageDisk,
		VectorDimensions: 1, Workers: 1, QueueCapacity: 16, MaxFileBytes: 64 << 10,
	})
	if err != nil {
		return err
	}
	defer engine.Close()
	build, err := engine.SyncFiles(ctx)
	if err != nil {
		return err
	}
	indexNS := time.Since(start).Nanoseconds()
	samples := make([]sample, 0, 20*len(probes))
	for round := 0; round <= 20; round++ {
		for _, p := range probes {
			start := time.Now()
			paths, err := retrieve(ctx, engine, p)
			if err != nil {
				return fmt.Errorf("probe %s: %w", p.ID, err)
			}
			if round > 0 {
				samples = append(samples, sample{p.ID, round, paths, time.Since(start).Nanoseconds()})
			}
		}
	}
	if err := engine.Close(); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		SchemaVersion int                `json:"schema_version"`
		GoVersion     string             `json:"go_version"`
		Platform      string             `json:"platform"`
		IndexNS       int64              `json:"index_ns"`
		Build         search.BuildReport `json:"build"`
		Samples       []sample           `json:"samples"`
	}{1, runtime.Version(), runtime.GOOS + "/" + runtime.GOARCH, indexNS, build, samples})
}

func retrieve(ctx context.Context, engine *search.Engine, p probe) ([]string, error) {
	paths := make([]string, 0, 5)
	switch p.Lane {
	case "keyword":
		hits, err := engine.Search(ctx, search.Query{Mode: search.ModeKeyword, Text: p.Query, Limit: 5})
		if err != nil {
			return nil, err
		}
		for _, hit := range hits {
			if hit.Stale {
				return nil, errors.New("stale baseline input")
			}
			paths = append(paths, hit.Path)
		}
	case "path":
		page, err := engine.FindFiles(ctx, search.FileQuery{Text: p.Query, Limit: 5})
		if err != nil {
			return nil, err
		}
		for _, hit := range page.Matches {
			paths = append(paths, hit.Path)
		}
	case "grep":
		page, err := engine.Grep(ctx, search.GrepQuery{Text: p.Query, Mode: search.GrepModePlain, Limit: 5})
		if err != nil {
			return nil, err
		}
		seen := make(map[string]bool)
		for _, hit := range page.Matches {
			if !seen[hit.Path] && len(paths) < 5 {
				paths = append(paths, hit.Path)
				seen[hit.Path] = true
			}
		}
	}
	return paths, nil
}
