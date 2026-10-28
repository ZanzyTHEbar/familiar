// Package retrieval retains an experimental scoped-evidence prototype.
// Open is safety-disabled until reviewed RootOnly dependency integration and
// native-read privacy acceptance. It starts no engine and creates no state.
package retrieval

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/ZanzyTHEbar/virtual-vectorfs/search"
)

// Adapter has immutable authority and no ingestion/update or mutation interface.
// File updates may be observed by upstream watchers; old candidates still bind
// their original content hash. Close must not race with client operations.
type Adapter struct {
	engine   *search.Engine
	root     string
	rootInfo os.FileInfo
	sources  map[string]bool
}

// Candidate is a lead, not a citation. Only this adapter can issue its private
// binding; a zero value or another adapter's candidate grants no authority.
type Candidate struct {
	adapter *Adapter
	hit     search.Hit
}

// Lead describes an index-time preview. PreviewStart is NOT a match position;
// Preview has no current-source guarantee. Use Expand for citable bytes.
type Lead struct {
	ID, Source, Path, ContentHash, Preview string
	PreviewStart                           int64
	PreviewTruncated, Stale                bool
}

func (c Candidate) Lead() Lead {
	h := c.hit
	// Upstream's title-only fallback can exceed its normal preview budget.
	// Bound only the display copy, preserving the candidate's private binding.
	if len(h.Snippet) > MaxPreviewBytes {
		end := MaxPreviewBytes
		for !utf8.RuneStart(h.Snippet[end]) {
			end--
		}
		h.Snippet = strings.Clone(h.Snippet[:end])
		h.SnippetTruncated = true
	}
	return Lead{h.ID, h.Source, h.Path, h.ContentHash, h.Snippet, h.ByteOffset, h.SnippetTruncated, h.Stale}
}

// Citation binds exact half-open UTF-8 byte offsets in one source field. Root is
// present only for live files. Origin indexed_record means a caller snapshot,
// not that Path exists or is current on disk. Live evidence is verified during
// Expand, not leased against later changes. No query-match claim is implied.
type Citation struct {
	Root string `json:"root,omitempty"`
	search.Excerpt
}

// ErrUnavailable identifies the intentionally safety-disabled R2 feature.
var ErrUnavailable = errors.New("retrieval: experimental API unavailable until reviewed RootOnly dependency and privacy acceptance")

// Open fails closed before argument inspection, filesystem access/state creation
// or search.Open. There is no public or environment-controlled bypass.
func Open(ctx context.Context, grant Grant, records []Record, stateParent string) (*Adapter, error) {
	return nil, ErrUnavailable
}

// openPrototype preserves the unmerged implementation for later reviewed
// integration. Nothing calls it; public Open must not delegate here until scoped
// native reads and the new dependency/API contract have been accepted.
func openPrototype(ctx context.Context, grant Grant, records []Record, stateParent string) (*Adapter, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	sources, docs, rootInfo, err := prepare(grant, records)
	if err != nil {
		return nil, err
	}
	if _, err := canonicalDirectory(stateParent, true); err != nil {
		return nil, fmt.Errorf("state parent: %w", err)
	}
	if grant.Root != "" && (within(grant.Root, stateParent) || within(stateParent, grant.Root)) {
		return nil, errors.New("state and live root must be disjoint")
	}
	state, err := os.MkdirTemp(stateParent, "scope-")
	if err != nil {
		return nil, err
	}
	root := grant.Root
	if root == "" {
		root = filepath.Join(state, "empty-root")
		if err := os.Mkdir(root, 0o700); err != nil {
			return nil, err
		}
	}
	e, err := search.Open(ctx, search.Options{
		Root: root, IndexPath: filepath.Join(state, "index"), Storage: search.StorageMemory,
		VectorDimensions: 1, MaxFileBytes: MaxInputBytes, Workers: 2,
		QueueCapacity: 32, WriteBatchBytes: 1 << 20,
	})
	if err != nil {
		return nil, err
	}
	a := &Adapter{engine: e, root: grant.Root, rootInfo: rootInfo, sources: sources}
	if grant.Root != "" {
		_, err = e.SyncFiles(ctx)
	}
	if err == nil {
		err = e.Upsert(ctx, docs...)
	}
	if err == nil {
		err = a.checkRoot()
	}
	if err != nil {
		return nil, errors.Join(err, e.Close())
	}
	return a, nil
}

// Search returns bounded keyword candidates from this scope only. It never
// accepts query filters or path strings as grants, nor returns exact citations.
func (a *Adapter) Search(ctx context.Context, text string, limit int) ([]Candidate, error) {
	if len(text) > 4096 || limit < 1 || limit > MaxCandidates {
		return nil, errors.New("query exceeds 4096 bytes or limit is outside 1..20")
	}
	if err := a.checkRoot(); err != nil {
		return nil, err
	}
	hits, err := a.engine.Search(ctx, search.Query{Mode: search.ModeKeyword, Text: text, Limit: limit})
	if err != nil {
		return nil, err
	}
	if err := a.checkRoot(); err != nil {
		return nil, err
	}
	result := make([]Candidate, 0, len(hits))
	for _, hit := range hits {
		if !a.sources[hit.Source] || (hit.Source == FileSource && a.root == "") {
			return nil, errors.New("index returned a source outside its immutable grant")
		}
		result = append(result, Candidate{adapter: a, hit: hit})
	}
	return result, nil
}

// Expand uses the candidate's private ID/source/hash, never a caller-supplied
// path or grep display offsets. Field is text, title or context (files: text
// only). offset is a source-field byte offset; limit must be 1..4096 bytes.
func (a *Adapter) Expand(ctx context.Context, c Candidate, field string, offset, limit int) (Citation, error) {
	if c.adapter != a || !a.sources[c.hit.Source] || c.hit.ContentHash == "" || c.hit.Stale {
		return Citation{}, errors.New("candidate is foreign, unbound or stale")
	}
	if offset < 0 || limit < 1 || limit > MaxExcerptBytes {
		return Citation{}, errors.New("invalid excerpt offset or byte limit")
	}
	if err := a.checkRoot(); err != nil {
		return Citation{}, err
	}
	x, err := a.engine.Expand(ctx, search.ExpandRequest{
		ID: c.hit.ID, ExpectedContentHash: c.hit.ContentHash, ExpectedSource: c.hit.Source,
		Field: field, Offset: offset, Limit: limit,
	})
	if err != nil {
		return Citation{}, err
	}
	if err := a.checkRoot(); err != nil {
		return Citation{}, err
	}
	root := ""
	if x.Origin == search.OriginLiveFile {
		if x.Source != FileSource || a.root == "" {
			return Citation{}, errors.New("live evidence lacks a root/source grant")
		}
		root = a.root
	} else if x.Origin != search.OriginIndexedRecord || x.Source == FileSource {
		return Citation{}, errors.New("unexpected evidence origin")
	}
	return Citation{Root: root, Excerpt: x}, nil
}

func (a *Adapter) checkRoot() error {
	if a.root == "" {
		return nil
	}
	info, err := canonicalDirectory(a.root, false)
	if err != nil {
		return fmt.Errorf("root grant unavailable: %w", err)
	}
	if !os.SameFile(a.rootInfo, info) {
		return errors.New("root identity changed; reopen with a new grant")
	}
	return nil
}

// Close drains all upstream native/index watchers and workers. It is idempotent.
func (a *Adapter) Close() error { return a.engine.Close() }
