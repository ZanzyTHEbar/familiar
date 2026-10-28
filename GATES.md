# F3 gates (defined before product implementation)

Owner: `f3-doctor`; target: `main`. Integration requires independent review.
The authoritative F/R/W/P plan remains
[`LOCAL_AGENT_PLAN.md`](https://github.com/ZanzyTHEbar/virtual-vectorfs/blob/main/docs/LOCAL_AGENT_PLAN.md);
this is an execution record, not a transferred plan or a second backlog.

## Exact CHECK / EXPECT

Run from this worktree. No installs, downloads, inference, or native builds.
`scripts/check-f3.sh` is the executable gate definition; its CLI contract is
`familiar doctor --json --offline`. Doctor has no workspace/state flags: it
does not inspect any directory, including its working directory or HOME.
The acceptance scenario runs from a newly created empty workspace with a separate
empty state directory selected by `XDG_STATE_HOME`; all outputs stay outside both.

```sh
umask 077
export SCRATCH=$(mktemp -d "${SCRATCH_PARENT:?select private scratch parent}/familiar-f3.XXXXXX")
: "${NATIVE:?set installed compatible native prefix}"
timeout 300s bash scripts/check-f3.sh
```

CHECKs executed by that script:

1. Materialize published `v0.0.0-20261028155634-88108866c3ca` through normal
   authenticated Go module retrieval before the offline gate. EXPECT: exact
   Origin `88108866c3cac44d7a66e1a9532d35b1f2271ecf`, no workspace override,
   local replacement or dirty source transfer. `GOWORK=off`; this experimental
   branch's API pin is not FFF/native adoption or privacy acceptance.
2. With explicit native prefix, CGO include/link/rpath and loader path,
   `timeout 180s go build -mod=readonly -p=1 -tags=vectors -o "$SCRATCH/bin/familiar" ./cmd/familiar`
   and `timeout 60s go build -mod=readonly -p=1 -tags=vectors ./...`.
   EXPECT: both exit 0. All Go commands use `GOPROXY=off GOSUMDB=off
   GOTOOLCHAIN=local GOMAXPROCS=2 GOMEMLIMIT=256MiB`, isolated `GOCACHE` and
   `TMPDIR`; native thread counts are bounded to two.
3. `timeout 60s go vet -mod=readonly -p=1 -tags=vectors ./...`;
   `test -z "$(gofmt -l cmd doctor retrieval)"`; `git diff --check`.
   EXPECT: exit 0, no findings. No search suite, Rust build, or benchmarks.
4. One real CLI acceptance scenario: snapshot empty workspace/state, execute
   `timeout 10s "$SCRATCH/bin/familiar" doctor --json --offline` from workspace
   with `HOME` pointing at an absent scratch path and `XDG_STATE_HOME` at state.
   EXPECT: exit 1 (blocked), schema version 1, platform/search/native-linkage
   ready, generation/embedding endpoints `not-configured`, overall blocked.
   Recompare manifests (paths, types, modes, size, mtime, hashes) and absent HOME.
   EXPECT: no changed files, workspace/state remain empty, HOME remains absent.
5. Launch through Linux's ELF loader with RPATH and loader cache disabled,
   `LD_LIBRARY_PATH` empty, and only system library directories available.
   EXPECT: nonzero exit with an actual missing FFF/FAISS shared-library error
   before main, not a doctor JSON diagnosis. Libraries are never moved/removed.

## Results

Passed on 2026-10-21 with the documented clean-export override. Vector builds,
vet, formatting and diff checks exited 0. The CLI acceptance scenario exited 1
with the expected explicit model blockers and identical before/after manifests.
The missing-native launch exited 127 with a real missing-`libfffbridge.so` loader
error. See [F3_EVIDENCE.md](F3_EVIDENCE.md) for paths, fingerprints, independent
review and the initial failed dependency-resolution attempt.

Successful linkage is not native ABI, cancellation, hardware/model fit,
filesystem recovery, R1 readiness, or F-integration qualification. R2 is now
safety-disabled and unmerged pending a reviewed published RootOnly dependency;
its expansion prototype does not compile against this older stable pin. These
historical F3 results are not a passing build claim for the current R2 worktree.
Re-run gates at the integration boundary;
the canonical plan has not been marked complete or transferred.
