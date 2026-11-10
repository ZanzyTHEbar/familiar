package retrieval_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"familiar.local/familiar/retrieval"
	"github.com/ZanzyTHEbar/virtual-vectorfs/search"
)

// One acceptance journey through the public API and real native engine. All
// inputs are synthetic; the required fresh scratch is retained for inspection.
func TestScopedEvidenceAcceptance(t *testing.T) {
	base := os.Getenv("R2_SCRATCH")
	if base == "" {
		t.Fatal("set R2_SCRATCH to a new nonexistent scratch path outside corpus roots")
	}
	mkdir := func(p string) {
		t.Helper()
		if err := os.Mkdir(p, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	write := func(p, text string) {
		t.Helper()
		if err := os.WriteFile(p, []byte(text), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	mkdir(base)
	aRoot, bRoot, state := filepath.Join(base, "allowed"), filepath.Join(base, "ungranted"), filepath.Join(base, "state")
	for _, p := range []string{aRoot, bRoot, state} {
		mkdir(p)
	}
	aPath, bPath := filepath.Join(aRoot, "brief.txt"), filepath.Join(bRoot, "brief.txt")
	text := "needle header\n" + strings.Repeat("prefix ", 180) + "café needle evidence.\n"
	write(aPath, text)
	write(bPath, strings.Repeat("needle outsidersentinel ", 30))
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	open := func(g retrieval.Grant, records []retrieval.Record) *retrieval.Adapter {
		t.Helper()
		a, err := retrieval.Open(ctx, g, records, state)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := a.Close(); err != nil {
				t.Error(err)
			}
		})
		return a
	}
	find := func(a *retrieval.Adapter, query string) retrieval.Candidate {
		t.Helper()
		c, err := a.Search(ctx, query, 1)
		if err != nil || len(c) != 1 {
			t.Fatalf("search %q: %v, %v", query, c, err)
		}
		return c[0]
	}
	refuse := func(a *retrieval.Adapter, c retrieval.Candidate, field string, start, limit int) {
		t.Helper()
		x, err := a.Expand(ctx, c, field, start, limit)
		if err == nil || x.Text != "" {
			t.Fatalf("expected refusal with no evidence: %+v, %v", x, err)
		}
	}
	// Snapshots include file bytes, size, mode and mtime, not only query results.
	manifest := func() string {
		t.Helper()
		var m strings.Builder
		for _, p := range []string{aPath, bPath} {
			b, err := os.ReadFile(p)
			if err != nil {
				t.Fatal(err)
			}
			st, err := os.Lstat(p)
			if err != nil {
				t.Fatal(err)
			}
			fmt.Fprintf(&m, "%s %o %d %d %x\n", p, st.Mode(), st.Size(), st.ModTime().UnixNano(), sha256.Sum256(b))
		}
		return m.String()
	}
	before := manifest()
	write(filepath.Join(base, "inputs-before.txt"), before)
	sources := []string{retrieval.FileSource, "notes"}
	records := []retrieval.Record{{ID: "note-1", Source: "notes", Path: aPath, Title: "Snapshot", Text: "recordtoken — caller snapshot", Context: "independent context"}}
	a := open(retrieval.Grant{Root: aRoot, Sources: sources}, records)
	b := open(retrieval.Grant{Root: bRoot, Sources: []string{retrieval.FileSource}}, nil)
	live, foreign := find(a, "needle"), find(b, "needle")
	if live.Lead().Path != "brief.txt" || foreign.Lead().Path != "brief.txt" || live.Lead().ContentHash == foreign.Lead().ContentHash {
		t.Fatal("separate roots/versions not represented")
	}
	if hits, err := a.Search(ctx, "outsidersentinel", 20); err != nil || len(hits) != 0 {
		t.Fatalf("ungranted root influenced retrieval: %v, %v", hits, err)
	}
	refuse(a, foreign, "text", 0, 20)
	refuse(b, live, "text", 0, 20)
	refuse(a, retrieval.Candidate{}, "text", 0, 20)
	t.Log("PASS: separate scoped indexes; no foreign or unbound candidate authority")

	start := strings.Index(text, "café")
	quote := "café needle evidence.\n"
	lead := live.Lead()
	if !lead.PreviewTruncated || lead.PreviewStart != 0 || strings.Contains(lead.Preview, quote) || start <= len(lead.Preview) {
		t.Fatalf("expected preview distinct from later exact range: %+v", lead)
	}
	x, err := a.Expand(ctx, live, "text", start, len(quote))
	if err != nil || x.Text != quote || x.StartByte != int64(start) || x.EndByte != int64(start+len(quote)) || x.TotalBytes != int64(len(text)) || x.Root != aRoot || x.Origin != search.OriginLiveFile || x.Source != retrieval.FileSource || x.ContentHash != lead.ContentHash || x.Field != "text" {
		t.Fatalf("exact live evidence: %+v, %v", x, err)
	}
	short, err := a.Expand(ctx, live, "text", start, 4)
	if err != nil || short.Text != "caf" || short.EndByte != int64(start+3) {
		t.Fatalf("UTF-8 byte-bound excerpt: %+v, %v", short, err)
	}
	refuse(a, live, "text", start+4, 10) // continuation byte, not rune start
	refuse(a, live, "text", len(text), 10)
	refuse(a, live, "text", 0, retrieval.MaxExcerptBytes+1)
	refuse(a, live, "text", -1, 10)
	refuse(a, live, "text", 0, 0)
	refuse(a, live, "title", 0, 10)
	t.Log("PASS: hash/source-bound exact range beyond preview; UTF-8 and byte limits")

	note := find(a, "recordtoken")
	sources[0], sources[1] = "ungranted", "ungranted"                           // cannot revoke or widen a live instance
	records[0].Text = "changed caller input"                                    // cannot rewrite indexed snapshot
	lead.Source, lead.ContentHash, lead.Path = "ungranted", "wrong-hash", bPath // display copy is not authority
	if _, err := a.Expand(ctx, live, "text", start, len(quote)); err != nil {
		t.Fatal("mutating display/grant copy affected private binding:", err)
	}
	n, err := a.Expand(ctx, note, "text", 12, 6)
	if err != nil || n.Text != "— ca" || n.Origin != search.OriginIndexedRecord || n.Root != "" || n.Path != aPath || n.StartByte != 12 || n.EndByte != 18 {
		t.Fatalf("caller record evidence must not read its live-file label: %+v, %v", n, err)
	}
	contextRange, err := a.Expand(ctx, note, "context", 0, 11)
	if err != nil || contextRange.Text != "independent" || contextRange.Field != "context" {
		t.Fatalf("independent record field: %+v, %v", contextRange, err)
	}
	refuse(a, note, "path", 0, 10)
	if denied, err := retrieval.Open(ctx, retrieval.Grant{Sources: []string{"notes"}}, []retrieval.Record{{ID: "secret", Source: "ungranted", Text: "secret"}}, state); err == nil {
		_ = denied.Close()
		t.Fatal("ungranted source admitted")
	}
	titlePrefix := "titleonlytoken "
	title := titlePrefix + strings.Repeat("界", 1<<18)
	r := open(retrieval.Grant{Sources: []string{"notes"}}, []retrieval.Record{
		{ID: "note-1", Source: "notes", Path: aPath, Text: "recordtoken new version"},
		{ID: "title-only", Source: "notes", Title: title},
	})
	titleCandidate := find(r, "titleonlytoken")
	titleLead := titleCandidate.Lead()
	expectedPreview := titlePrefix + strings.Repeat("界", (retrieval.MaxPreviewBytes-len(titlePrefix))/len("界"))
	if len(titleLead.Preview) > retrieval.MaxPreviewBytes || !utf8.ValidString(titleLead.Preview) || !titleLead.PreviewTruncated || titleLead.Preview != expectedPreview {
		t.Fatal("large UTF-8 title-only display preview is not correctly bounded")
	}
	titleOffset := len(title) - len("界")
	titleEvidence, err := r.Expand(ctx, titleCandidate, "title", titleOffset, len("界"))
	if err != nil || titleEvidence.Text != "界" || titleEvidence.StartByte != int64(titleOffset) || titleEvidence.EndByte != int64(len(title)) || titleEvidence.TotalBytes != int64(len(title)) || titleEvidence.ContentHash != titleLead.ContentHash || titleEvidence.Origin != search.OriginIndexedRecord {
		t.Fatalf("bounding title preview changed private citation binding: %+v, %v", titleEvidence, err)
	}
	t.Log("PASS: large title-only UTF-8 preview bounded to 512 bytes; original title/hash/range still expandable")
	newNote := find(r, "recordtoken")
	if newNote.Lead().ContentHash == note.Lead().ContentHash {
		t.Fatal("changed record version has same hash")
	}
	refuse(r, note, "text", 0, 10)
	refuse(a, newNote, "text", 0, 10)
	if hits, err := r.Search(ctx, "needle", 20); err != nil || len(hits) != 0 {
		t.Fatalf("record path conferred file authority: %v, %v", hits, err)
	}
	t.Log("PASS: immutable source grants/snapshots; caller paths confer no live-file authority")

	after := manifest()
	write(filepath.Join(base, "inputs-after.txt"), after)
	if before != after {
		t.Fatal("read-only API changed input bytes or metadata")
	}
	t.Log("PASS: input bytes/modes/size/mtime identical before intentional mutations")
	write(aPath, "changed version with needle")
	refuse(a, live, "text", start, len(quote))
	if err := os.Rename(aPath, filepath.Join(aRoot, "renamed.txt")); err != nil {
		t.Fatal(err)
	}
	refuse(a, live, "text", 0, 10)
	// An escaping symlink at the old path cannot restore authority, even if the
	// external file contains exactly the originally indexed bytes/hash.
	write(bPath, text)
	if err := os.Symlink(bPath, aPath); err != nil {
		t.Fatal(err)
	}
	refuse(a, live, "text", 0, 10)
	if err := os.Remove(aPath); err != nil {
		t.Fatal(err)
	}
	refuse(a, live, "text", 0, 10)
	t.Log("PASS: changed, renamed, missing and escaping-symlink files cannot be cited current")
	if err := os.Rename(aRoot, aRoot+"-original"); err != nil {
		t.Fatal(err)
	}
	mkdir(aRoot)
	write(aPath, text)
	refuse(a, live, "text", 0, 10)
	if hits, err := a.Search(ctx, "needle", 1); err == nil || len(hits) != 0 {
		t.Fatalf("replaced root accepted: %v, %v", hits, err)
	}
	t.Log("PASS: replaced root rejected even with identical file bytes; all adapters close on exit")
}
