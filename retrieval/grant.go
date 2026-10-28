package retrieval

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf8"

	"github.com/ZanzyTHEbar/virtual-vectorfs/search"
)

const (
	FileSource = "fff-file"
	// MaxPreviewBytes matches the upstream ordinary hit-preview byte budget.
	MaxPreviewBytes = 512
	MaxExcerptBytes = 4096
	MaxCandidates   = 20
	MaxInputBytes   = 8 << 20
	MaxRecords      = 1000
)

// Grant authorizes a whole live root and/or caller-record sources, not paths
// found in results. Root requires FileSource and vice versa. An empty Root is
// record-only. Grant no secrets or subtrees requiring exclusions: this API has
// no deny rules. Open copies Sources; later changes cannot widen authority.
type Grant struct {
	Root    string
	Sources []string
}

// Record is an immutable caller-owned snapshot, never a live file. Path is only
// a label. Fields must be UTF-8; no vectors, arbitrary metadata or supplied hashes.
type Record struct {
	ID, Source, Path, Title, Text, Context string
}

func prepare(grant Grant, records []Record) (map[string]bool, []search.Document, os.FileInfo, error) {
	sources := make(map[string]bool, len(grant.Sources))
	if len(grant.Sources) == 0 || len(grant.Sources) > 64 {
		return nil, nil, nil, errors.New("grant requires 1..64 explicit sources")
	}
	for _, source := range grant.Sources {
		if source == "" || len(source) > 256 || !utf8.ValidString(source) || strings.ContainsRune(source, 0) || sources[source] {
			return nil, nil, nil, errors.New("invalid or duplicate source grant")
		}
		sources[source] = true
	}
	if sources[FileSource] != (grant.Root != "") {
		return nil, nil, nil, errors.New("live root and fff-file source must be granted together")
	}
	var rootInfo os.FileInfo
	if grant.Root != "" {
		var err error
		rootInfo, err = canonicalDirectory(grant.Root, false)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("root grant: %w", err)
		}
		if grant.Root == string(filepath.Separator) {
			return nil, nil, nil, errors.New("filesystem-wide root is not supported")
		}
	}
	if len(records) > MaxRecords {
		return nil, nil, nil, errors.New("too many caller records")
	}
	docs := make([]search.Document, 0, len(records))
	ids := make(map[string]bool, len(records))
	total := 0
	for _, r := range records {
		if !sources[r.Source] || r.Source == FileSource || r.ID == "" || strings.HasPrefix(r.ID, "file:") || ids[r.ID] {
			return nil, nil, nil, errors.New("record has ungranted/reserved source or invalid/duplicate ID")
		}
		ids[r.ID] = true
		for _, value := range []string{r.ID, r.Source, r.Path, r.Title, r.Text, r.Context} {
			if len(value) > MaxInputBytes-total || !utf8.ValidString(value) || strings.ContainsRune(value, 0) {
				return nil, nil, nil, errors.New("record input exceeds byte budget or is not UTF-8 text")
			}
			total += len(value)
		}
		docs = append(docs, search.Document{ID: r.ID, Source: r.Source, Path: r.Path, Title: r.Title, Text: r.Text, Context: r.Context})
	}
	return sources, docs, rootInfo, nil
}

func canonicalDirectory(path string, private bool) (os.FileInfo, error) {
	if !filepath.IsAbs(path) || filepath.Clean(path) != path {
		return nil, errors.New("directory must be an absolute clean path")
	}
	real, err := filepath.EvalSymlinks(path)
	if err != nil {
		return nil, err
	}
	if real != path {
		return nil, errors.New("directory ancestry must not contain symlinks")
	}
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("not a directory")
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != uint32(os.Geteuid()) {
		return nil, errors.New("directory must be owned by the caller")
	}
	if private && info.Mode().Perm() != 0o700 {
		return nil, errors.New("state parent must have mode 0700")
	}
	return info, nil
}

func within(root, candidate string) bool {
	rel, err := filepath.Rel(root, candidate)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
