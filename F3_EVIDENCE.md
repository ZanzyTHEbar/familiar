# F3 execution evidence — 2026-10-21

Public copy: original host locations are represented by logical private-archive
labels/environment variables below. Original measured results, pins and artifact
hashes are unchanged; exact local path mappings/raw logs are not distributed.

## Ownership and source

- Canonical Familiar repository: local `main` checkout.
- Worktree: `f3-doctor`.
- Worker branch: `f3-doctor`; integration target: `main`.
- Starting Familiar commit: `317f050` (repository initialization).
- Gates commit: `7e3ab8c` (`docs(f3): define offline doctor gates`). Gates were
  written before product implementation in this session; this commit records
  them before the product commit, not before the local implementation work.
- Module: `familiar.local/familiar`, executable `familiar`; local-only identity,
  not a public hosting identity. No upstream plan transfer or update.
- Search commit: **`fe092dd03d398570b5d93cf0c030f573bc90257b`**.
  Requirement: `v0.0.0-20261011132509-fe092dd03d39` (UTC commit timestamp verified).
  No F2 transfer; R2 remains gated on coordinator-approved F2 integration.

The exact search revision's `.info` is cached, but its module archive/go.mod are
not. The approved offline mechanism is `git archive` of that commit into an
owner-controlled scratch directory, with a scratch-only `go.work` containing
Familiar and a version-specific replace to the clean export. No repository
replace, vendoring, host-specific module path, dirty source copy, or dependency
publication. Archive SHA-256:

```text
917c15a34ae240d1d18b63fac6860fd27fbfdd57b35256297dbe43204d27d51d
```

`go.mod` indirect requirements and `go.sum` were generated using offline
`go mod tidy -modfile="$TIDY_SCRATCH/familiar.mod"` with
`GOWORK=off` and a temporary version-specific replacement to that same export.
Only dependency requirements/checksums were retained, not the replacement.
No search archive checksum is fabricated in `go.sum`: its provenance for these
gates is the explicit committed export, not a downloaded module zip.

## Environment and native boundary

Linux amd64; `go version go1.26.8 linux/amd64`. Installed native prefix:
the explicitly selected `$NATIVE` location in the private command record.
The script explicitly sets `PKG_CONFIG_PATH`, `CGO_CFLAGS`, `CGO_LDFLAGS`
(including rpath), and `LD_LIBRARY_PATH` to that prefix. `pkg-config
--modversion fffbridge` returned **0.1.0**, not a FFF engine-version claim.
No native install/build/download or Rust command was run.

Native shared-library SHA-256 at validation:

```text
libfffbridge.so  1ee92c96fb4307ad57331fc4cea7f5d09385c6e73ed5337a776dceda84c9bf8d
libfaiss_c.so    d3061d1fa475f937efb1e1f9b5694f44e4abd307761d914be25ea15b8a93ed0e
libfaiss.so      a5448f049afdf22d876067ae1b11d77528e468511c14697a0eb3638aabeedb29
```

All Go validation uses `GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
GOMAXPROCS=2 GOMEMLIMIT=256MiB`, `-p=2` for builds/vet, explicit timeouts,
scratch build cache/temp outputs, and two-thread native environment limits.
The Go memory limit is not a hard process-tree/native memory ceiling.
`findmnt -T "$SCRATCH_PARENT"` reported `virtiofs`;
`stat -f -c %T "$SCRATCH_PARENT"` returned `fuseblk`. These are tool-specific
observations, not evidence of a mount change or broker qualification.

## Actual CHECK / EXPECT results

Exact runnable definitions: [GATES.md](GATES.md), `scripts/check-f3.sh`.
Final full run after module metadata completion:

```sh
: "${SCRATCH:?original fresh private F3 scratch}"
: "${SOURCE:?approved clean search checkout}"
: "${NATIVE:?installed compatible native prefix}"
timeout 300s bash scripts/check-f3.sh
```

| CHECK | EXPECT / actual |
| --- | --- |
| Clean committed export + explicit scratch GOWORK override | Passed; no canonical dirty dependency |
| Vector-enabled CLI build, 180 s timeout | Exit 0 |
| Vector-enabled `go build ./...`, 60 s timeout | Exit 0 |
| Vector-enabled `go vet ./...`, 60 s timeout | Exit 0, no findings |
| `gofmt -l cmd doctor`, `git diff --check` | Exit 0, no findings |
| Doctor from new empty scratch workspace; separate empty state; absent HOME | Exit 1; schema 1; scope F3; aggregate blocked; platform/search-import/native-linkage ready; generation/embedding endpoints not-configured |
| Before/after paths, types, modes, sizes, mtimes, file hashes | Identical; both roots remain empty, HOME absent |
| Isolated loader launch with cache/RPATH disabled, no native loader path | Exit 127 before main; `error while loading shared libraries: libfffbridge.so: cannot open shared object file: No such file or directory` |

Raw final evidence remains in the private F3 final archive:
`doctor.json`, `doctor.stderr` (empty), `before.json`, `after.json`,
`loader.stdout` (empty), `loader.stderr`, clean source export, `go.work`, binary
and isolated build cache. Before/after manifest SHA-256 is identical:
`6c6d54f12988b023221dd269862cf61de024c5d34066ad334ecf6fec500a5e75`.
Doctor JSON SHA-256:
`f8970facd5517591667cb1131884b142723e7c57dd45d39671aaf88e8afdc687`.
The private path mappings are execution evidence, not dependency configuration.

The initial private-scratch attempt stopped at the first build:
`module lookup disabled by GOPROXY=off`. Merely adding both modules to `use`
did not resolve the pinned requirement's graph offline. The documented fix was
an explicit **version-specific scratch workspace replacement**, not a stub,
network fetch, or dirty-tree replacement. The revised full gate then passed in
the private tidy/retry scratch and was repeated in the final scratch above.

Before/after SHA-256 comparison verified all **12** pre-existing dirty/untracked
canonical search files and the dirty companion handoff document byte-identical.
The initial `source-before.json` manifest is retained in that private archive.
No canonical code/docs, app data, or personal directories were modified.

## Independent review and limits

A separate read-only reviewer (`general`, session
`ses_e9fce5470ffe401LXEPg7CeB7h`) reported **no blocking findings** on the F3
implementation. They verified the pinned pre-cancellation contract and compared
the saved before/after manifests independently. They did not independently
rerun build/vet or the loader launch. Their final metadata/evidence follow-up
also found no blockers: checksums matched offline tidy output; final JSON,
manifest and archive fingerprints matched this record. Build/vet exit results
and loader exit 127 remain parent-reported. Integration still requires
coordinator review and fresh parent-boundary gate re-verification.
No merge/push/PR/cleanup is performed.

No helper unit tests, mocks, or test matrix were added. One real CLI acceptance
scenario is embedded in the gates script. No full search/race/Rust suite or
benchmark was run. No native operations were qualified beyond process linkage
and the exported search API's pre-canceled return. Hardware/model fit, filesystem
recovery and native cancellation remain unqualified. Models/endpoints are
deliberately not configured or contacted; F3 does not complete R1 or F integration.

Next planned product slice: **F4**, then separately gated **R1**. R2 must not
start until coordinator-approved F2 integration. Event replay, offloads, Goja,
`llm.ask`, writes, and model downloads remain outside this slice.
