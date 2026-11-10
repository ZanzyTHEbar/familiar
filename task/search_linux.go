package task

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"sort"
	"strings"
	"unicode/utf8"
)

type sourceLine struct {
	Line    int    `json:"line"`
	Text    string `json:"text"`
	Clipped bool   `json:"clipped,omitempty"`
}
type searchSource struct {
	Path    string            `json:"path"`
	Matches []sourceLine      `json:"matches"`
	Read    map[string]string `json:"read"`
	score   int
}

func clipText(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n]
}

// A transparent lexical ordering, not semantic relevance or evidence confidence.
func sourceRank(p, text string) int {
	score := 0
	switch strings.ToLower(path.Ext(p)) {
	case ".go", ".rs", ".c", ".h", ".cpp", ".cc", ".py", ".js", ".ts", ".tsx", ".jsx", ".kt", ".java", ".swift", ".rb", ".sh":
		score = 20
	case ".md", ".txt", ".rst":
		score = -100
	}
	t := strings.TrimSpace(text)
	for _, prefix := range []string{"func ", "fn ", "pub fn ", "def ", "class ", "function ", "export function ", "async def "} {
		if strings.HasPrefix(t, prefix) {
			score += 40
			break
		}
	}
	return score
}

func (w *workspace) search(ctx context.Context, p, q string) (string, error) {
	if q == "" || len(q) > 512 || strings.ContainsAny(q, "\n\r\x00") {
		return "", errors.New("query must be 1..512 bytes of single-line literal text, without shell quoting or regex syntax")
	}
	result := struct {
		Kind    string         `json:"kind"`
		Sources []searchSource `json:"sources"`
		Files   int            `json:"files_scanned"`
		Matches int            `json:"total_matches"`
		Skipped int            `json:"skipped"`
		Limited bool           `json:"scan_limited"`
		Omitted int            `json:"omitted_source_files"`
		Note    string         `json:"note"`
	}{Kind: "search_leads", Sources: []searchSource{}, Note: "Literal/case-sensitive. Definitions and source files ranked before docs; ties by path. Read exact returned arguments for current evidence. Hidden/symlink/mount/binary paths excluded; ignore rules NOT consulted. Skips/limits/omissions mean incomplete results."}
	var candidates []searchSource
	entries := 0
	bytesLeft := int64(32 << 20)
	// The caller owns this already scope-checked file handle.
	scan := func(file *os.File, name string) {
		st, err := file.Stat()
		if err != nil || !st.Mode().IsRegular() || st.Size() > 1<<20 {
			result.Skipped++
			return
		}
		result.Files++
		capBytes := min(int64(1<<20), bytesLeft)
		data, err := io.ReadAll(io.LimitReader(file, capBytes+1))
		bytesLeft -= int64(len(data))
		if err != nil || int64(len(data)) > capBytes || !utf8.Valid(data) || strings.ContainsRune(string(data), 0) {
			result.Skipped++
			return
		}
		source := searchSource{Path: name, score: -1000}
		bestLine := 0
		for i, text := range strings.Split(string(data), "\n") {
			if !strings.Contains(text, q) {
				continue
			}
			result.Matches++
			rank := sourceRank(name, text)
			if rank > source.score {
				source.score = rank
				bestLine = i + 1
				// Keep the best declaration even after earlier incidental hits.
				if len(source.Matches) == 3 {
					source.Matches = source.Matches[:2]
				}
			}
			if len(source.Matches) < 3 {
				source.Matches = append(source.Matches, sourceLine{Line: i + 1, Text: clipText(text, 240), Clipped: len(text) > 240})
			}
		}
		if bestLine > 0 {
			// Include surrounding declarations, not just isolated uses.
			source.Read = map[string]string{"path": name, "start": fmt.Sprint(max(1, bestLine-32))}
			candidates = append(candidates, source)
		}
	}
	var walk func(string, int) error
	walk = func(dir string, depth int) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if depth > 32 {
			result.Skipped++
			result.Limited = true
			return nil
		}
		f, err := w.open(dir)
		if err != nil {
			result.Skipped++
			return nil
		}
		defer f.Close()
		st, err := f.Stat()
		if err != nil {
			result.Skipped++
			return nil
		}
		if st.Mode().IsRegular() {
			scan(f, dir)
			return nil
		}
		for {
			list, e := f.ReadDir(128)
			if e != nil && e != io.EOF {
				result.Skipped++
				return nil
			}
			for _, entry := range list {
				if err := ctx.Err(); err != nil {
					return err
				}
				entries++
				if entries > 10000 || result.Files >= 2000 || bytesLeft <= 0 {
					result.Limited = true
					return nil
				}
				if strings.HasPrefix(entry.Name(), ".") || entry.Type()&os.ModeSymlink != 0 {
					result.Skipped++
					continue
				}
				name := entry.Name()
				if dir != "." {
					name = dir + "/" + name
				}
				if entry.IsDir() {
					if err := walk(name, depth+1); err != nil {
						return err
					}
					if entries > 10000 || result.Files >= 2000 || bytesLeft <= 0 {
						result.Limited = true
						return nil
					}
					continue
				}
				file, err := w.open(name)
				if err != nil {
					result.Skipped++
					continue
				}
				scan(file, name)
				file.Close()
			}
			if e == io.EOF {
				return nil
			}
		}
	}
	f, err := w.open(p)
	if err != nil {
		return "", err
	}
	f.Close()
	if err := walk(p, 0); err != nil {
		return "", err
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score != candidates[j].score {
			return candidates[i].score > candidates[j].score
		}
		return candidates[i].Path < candidates[j].Path
	})
	result.Sources = append(result.Sources, candidates[:min(3, len(candidates))]...)
	for {
		result.Omitted = len(candidates) - len(result.Sources)
		encoded, _ := json.Marshal(result)
		if len(encoded) <= 4096 {
			return string(encoded), nil
		}
		if len(result.Sources) == 0 {
			return "", errors.New("search metadata exceeds excerpt budget")
		}
		result.Sources = result.Sources[:len(result.Sources)-1]
	}
}
