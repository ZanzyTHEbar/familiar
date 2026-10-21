# F3 independent review — oc2

**Verdict: PASS for the authorized F3 slice; no blocking findings. No integration performed.**

- Reviewer: Workmux `oc2`, interactive independent reviewer, branch `review-f3`.
- Implementation: `9f80f78c0cda3c4324da241da2d5ae58ae0e722f` (`f3-doctor`).
- Target: `main`, `317f050b8d55078a20360d12dbc0565d854b8bae`.
- Review date: 2026-10-21. Product code and implementation evidence were not edited.

## Scope and correctness

Reviewed the actual committed diff and all eight added files: `README.md`,
`GATES.md`, `F3_EVIDENCE.md`, `go.mod`, `go.sum`, `cmd/familiar/main.go`,
`doctor/doctor.go`, and `scripts/check-f3.sh`. Checked tracked, untracked and
ignored state separately; `.workmux/` is ignored harness state, not product code.
No applicable repository `AGENTS.md` was found. Global instructions and the
authoritative canonical plan/policy/search-improvement documents and dirty
companion handoff were read without modification. The user's Familiar approval
supersedes the old handoff's location blocker; no plan transfer is implied.

The module is local-only `familiar.local/familiar`. The CLI emits schema 1,
scope F3, explicit capability reasons and aggregate blocked status. Missing
generation/embedding configuration remains `not-configured`, not fake readiness.
There is no endpoint configuration/probe interface in this slice; model/hardware
qualification remains R1. No agent loop, Goja, inference, download, index, watch,
or write authority was added. Native readiness means linkage only, not ABI or
operational qualification.

Verified `search.Open` in the **committed pin**, not the dirty canonical source:
`fe092dd03d398570b5d93cf0c030f573bc90257b:search/engine.go:50-53`
returns the pre-canceled context before normalization or any engine/native open.
The approved scratch replacement uses only `git archive` of that pin. Committed
module files contain no replacement, vendoring or host-specific dependency path.
The pin is not available as a complete cached module archive here; the documented
clean-export override remains necessary for this offline reproduction. This is
not evidence of standalone offline module-cache availability on another host.

## Independently rerun parent-boundary gates

Run from `/mnt/common/projects/tools/familiar__worktrees/review-f3`:

```sh
umask 077
SCRATCH=$(mktemp -d /mnt/common/familiar-review-oc2/gates.XXXXXX)
SOURCE=$(realpath ../../../go-pkgs/virtual-vectorfs)
NATIVE="$SOURCE/build/native"
export SCRATCH SOURCE NATIVE
timeout 300s bash -x scripts/check-f3.sh > /mnt/common/familiar-review-oc2/gates.log 2>&1
```

Actual scratch: `/mnt/common/familiar-review-oc2/gates.35fegB`.
Outer gate exit: **0**. Full expanded commands/output: `gates.log` above;
the immutable implementation's `scripts/check-f3.sh` defines all assertions.
Every Go build/vet used `GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local
GOMAXPROCS=2 GOMEMLIMIT=256MiB`, `-mod=readonly -p=2 -tags=vectors`,
isolated cache/temp paths, and native thread limits of two.

| Gate | Independent result |
| --- | --- |
| `timeout 180s go build ... -o "$SCRATCH/bin/familiar" ./cmd/familiar` | Exit 0 |
| `timeout 60s go build ... ./...` | Exit 0 |
| `timeout 60s go vet ... ./...` | Exit 0, no findings |
| `test -z "$(gofmt -l cmd doctor)"`; `git diff --check` | Exit 0 |
| `git diff --check main HEAD` | Exit 0; committed additions checked too |
| `timeout 10s "$SCRATCH/bin/familiar" doctor --json --offline` | Expected exit 1; schema 1/F3; platform/import/linkage ready, both model endpoints not-configured |
| Empty corpus/workspace and separate state manifests | Byte-identical before/after; roots remain empty, scratch HOME absent |
| Loader with `--inhibit-cache --inhibit-rpath ''`, system-only library path and empty `LD_LIBRARY_PATH` | Exit 127: missing `libfffbridge.so` before main; not a JSON diagnostic |

Only installed canonical `build/native` FFF/FAISS libraries were reused; their
three hashes match `F3_EVIDENCE.md`. ELF NEEDED/RUNPATH and a live syscall trace
confirm those libraries, not other project artifacts, were used. No downloads,
installs, native builds, `make`, user database access or personal scans occurred.

Additional real CLI audit: bounded `strace -f -qq` with
`-e trace=%file,%network,write,writev,pwrite64,ftruncate,fsync,fdatasync,getdents64,inotify_init,inotify_init1,inotify_add_watch,fanotify_init`.
Exact argv, nine-variable isolated environment, working directory and exit code
are in `/mnt/common/familiar-review-oc2/strace-command.json`; trace is
`doctor.strace` there. Observed only native/system/runtime/executable reads and
one JSON write to stdout: no writable file opens, file mutations, directory
enumeration, sockets, or filesystem watches. Exit 1 and JSON match the gate run.
This is observed behavior, not an OS sandbox or general native qualification.

## Durable evidence and preservation

Raw logs/manifests/trace/build metadata remain in
`/mnt/common/familiar-review-oc2`; `SHA256SUMS` fingerprints them.
Key SHA-256 values:

```text
gates.log:      770a2004fdb74fd6a6cb14716e5329186e34dc9bac9c93b8b12ed41b1058ed18
doctor.json:    f8970facd5517591667cb1131884b142723e7c57dd45d39671aaf88e8afdc687
before/after:   c0938b4fda353d63f939380263948d2b2a883652cd7cab8fed2500264401b615
source before/after: cccc47dbea6ac9ae1e04ee32353c3925f4a98bed2573c8daa0830f02d901635a
pinned archive: 917c15a34ae240d1d18b63fac6860fd27fbfdd57b35256297dbe43204d27d51d
```

The saved implementation doctor JSON, corpus/state manifests, archive and native
fingerprints were independently checked against its evidence record. Fresh
review snapshots preserve canonical Familiar's clean main HEAD/status, all 12
dirty/untracked search files, and the dirty companion handoff byte-for-byte;
search/handoff fingerprints also match the implementation's original baseline.
Search's tracked code diff is only the pre-existing F2 expansion work; its
untracked inventory contains no Familiar application files. No application code
was added to virtual-vectorfs. No source was copied from dirty F2.

## Limits and next step

No F3 blocker found. This review does not qualify R1 models, F4 corpus quality,
native ABI/cancellation, recovery, other platforms, or F integration. R2 remains
blocked on coordinator-approved F2 integration. Canonical main is clean and has
no existing tracked policy/CI files or configured remote; the independent-review
and parent-boundary rules come from the user and authoritative delivery plan.
Only this review document is intended for a review-branch commit. No merge,
push, PR, plan update or cleanup is performed. Next: coordinator considers this
evidence and requests integration separately; no integration is authorized here.
