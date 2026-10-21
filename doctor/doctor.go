// Package doctor reports the read-only F3 build boundary. It never opens an
// index, inspects directories, starts watchers, or contacts model endpoints.
package doctor

import (
	"context"
	"errors"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/ZanzyTHEbar/virtual-vectorfs/search"
)

// Status describes capability readiness, not the success of collecting a report.
type Status string

const (
	Ready         Status = "ready"
	Blocked       Status = "blocked"
	NotConfigured Status = "not-configured"
)

// Check is one capability observation. Reason states the observation's limits.
type Check struct {
	ID     string `json:"id"`
	Status Status `json:"status"`
	Reason string `json:"reason"`
}

// Report is the version-1 readiness wire format. Status is blocked if any
// capability is blocked or not configured; ready never means agent readiness.
type Report struct {
	SchemaVersion     int     `json:"schema_version"`
	Scope             string  `json:"scope"`
	Status            Status  `json:"status"`
	GoVersion         string  `json:"go_version"`
	SearchVersion     string  `json:"search_version"`
	SearchReplacement string  `json:"search_replacement,omitempty"`
	Checks            []Check `json:"checks"`
}

// Inspect collects build and platform observations without filesystem or network
// access. Model configuration/probes are deliberately absent until R1; no
// environment variables or personal configuration files are read.
func Inspect() Report {
	r := Report{SchemaVersion: 1, Scope: "F3", Status: Ready, GoVersion: runtime.Version()}
	platform := Check{"platform", Ready, runtime.GOOS + "/" + runtime.GOARCH + "; filesystem recovery and hardware/model fit unqualified"}
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		platform.Status = Blocked
		platform.Reason += "; F3 currently qualified only on Linux amd64"
	}
	r.Checks = append(r.Checks, platform)

	cgo, vectors := false, false
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "CGO_ENABLED":
				cgo = setting.Value == "1"
			case "-tags":
				for _, tag := range strings.Split(setting.Value, ",") {
					vectors = vectors || tag == "vectors"
				}
			}
		}
		for _, dep := range info.Deps {
			if dep.Path == "github.com/ZanzyTHEbar/virtual-vectorfs" {
				r.SearchVersion = dep.Version
				if dep.Replace != nil {
					r.SearchReplacement = "explicit override; source provenance must be verified by build gates"
				}
			}
		}
	}

	// The pinned API checks cancellation before normalization, native open, or
	// index/watch setup. Exercise the real exported import without opening roots.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := search.Open(ctx, search.Options{})
	importCheck := Check{"search-import", Ready, "committed search API returned pre-canceled context; no engine/index/watch opened; F2 unavailable"}
	if !errors.Is(err, context.Canceled) || r.SearchVersion == "" {
		importCheck.Status = Blocked
		importCheck.Reason = "search import or build provenance could not be confirmed"
	}
	r.Checks = append(r.Checks, importCheck)
	native := Check{"native-linkage", Ready, "CGO/vector build started with linked FFF/FAISS libraries; no native operation, ABI/version, or cancellation qualification"}
	if !cgo || !vectors {
		native.Status = Blocked
		native.Reason = "requires CGO_ENABLED=1 and -tags=vectors build metadata; native readiness not established"
	}
	r.Checks = append(r.Checks, native,
		Check{"generation-endpoint", NotConfigured, "F3 has no model configuration or endpoint probes; no contact or inference; R1 required"},
		Check{"embedding-endpoint", NotConfigured, "F3 has no model configuration or endpoint probes; no contact or inference; R1 required"},
	)
	for _, check := range r.Checks {
		if check.Status != Ready {
			r.Status = Blocked
		}
	}
	return r
}
