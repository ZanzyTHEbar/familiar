# F4 execution evidence — 2026-10-21

Public copy: original host locations use logical private-archive labels or
environment variables. Measurements, failures, pins and original artifact hashes
are unchanged; raw logs and exact local path mappings are not distributed.

## Ownership and boundary

Worker/worktree: `f4`.
Target: local Familiar `main`, starting at independently reviewed F3
`9f80f78c0cda3c4324da241da2d5ae58ae0e722f`.
The canonical F/R/W/P plan and F1 policy were read in the search repository;
the dirty original companion handoff was read-only reference. No contradictory
F4 scope was found. No plan transfer, upstream edits or completion marking.

`F4_GATES.md` was written before implementation/measurement. Deliverable: versioned
synthetic UTF-8 inputs, judged tasks/gold byte ranges, offline verifier and real
retrieval measurement utility. No models, inference, downloads, agent loop,
R2 adapter, expansion API, write tool, personal-file inventory, or native builds.
Doctor and module dependency files are unchanged. No unit tests, mocks or search
test-suite expansion. The existing real F3 acceptance check is rerun; F4 has one
real scratch acceptance scenario, repeated for measurement reproducibility.

## Exact final checks

```sh
umask 077
: "${SCRATCH:?original newly allocated empty private directory}"
: "${SOURCE:?approved clean search checkout}"
: "${NATIVE:?installed compatible native prefix}"
timeout 300s bash scripts/check-f3.sh
timeout 180s bash scripts/check-f4.sh
```

Both script commands exited **0**. They explicitly set offline Go resolution,
local toolchain, CGO/vector flags, installed native prefix, bounded Go/native
threads, isolated build cache/temp directory, and timeouts. Search dependency is
only `git archive fe092dd03d398570b5d93cf0c030f573bc90257b`, using F3's approved
scratch-only, version-specific GOWORK override. No dirty F2 transfer or committed
replacement. Full runnable build/native environment is in the scripts.

| Check | Actual result |
| --- | --- |
| `timeout 180s go build -mod=readonly -p=2 -tags=vectors -o "$SCRATCH/bin/familiar" ./cmd/familiar` | Exit 0 |
| `timeout 60s go build -mod=readonly -p=2 -tags=vectors ./...` | Exit 0 |
| `timeout 60s go build -mod=readonly -p=2 -tags=vectors -o "$SCRATCH/bin/familiar-baseline" ./cmd/familiar-baseline` | Exit 0 |
| `timeout 60s go vet -mod=readonly -p=2 -tags=vectors ./...` (both scripts) | Exit 0, no findings |
| `test -z "$(gofmt -l cmd doctor)"`; `git diff --check` | Exit 0 |
| `bash -n scripts/check-f4.sh`; Python `ast.parse` of loader | Passed |
| Existing F3 real CLI scenario | Doctor exit 1, expected schema/blockers; empty roots unchanged, HOME absent |
| Existing F3 isolated native loader scenario | Exit 127, actual missing `libfffbridge.so` error before main |
| `python3 scripts/fixture_baseline.py --corpus "$PWD/fixtures/v1" --verify-only` | 10 inputs, 13 judged tasks, all input/gold checks passed |
| Same loader with `--binary "$SCRATCH/bin/familiar-baseline" --scratch "$SCRATCH/f4"` | Two fresh native-search runs passed, same judgments |
| Disposable fixture copies: changed bytes, changed gold offset, unexpected file | Each rejected nonzero with its expected specific verification reason, before search |

Each run indexed all 10 files, skipped none, and reported no indexing errors.
One warmup per probe, then 20 rounds × 13 probes = 260 measured retrieval calls
per run. All nonempty relevance judgments achieved recall@5 1.0 on every round;
both absent-input probes returned no hits on every round. Safety-refusal prompts
are **not** executed or scored as agent success: F4 only measures retrieval.

| Measurement | Run 1 | Run 2 | Declared ceiling |
| --- | ---: | ---: | ---: |
| Overall warm p95, ms | 0.818541 | 1.072357 | 500 |
| Worst individual probe p95, ms | 1.996641 | 3.990472 | 500 |
| Process wall, s | 0.545064 | 0.403008 | 30 |
| Open/index, ms | 394.625829 | 227.330440 | reported separately |
| Sampled process-tree peak RSS, KiB | 38,296 | 37,916 | 524,288 |
| Observed process VmHWM, KiB | 38,296 | 37,916 | 524,288 |
| Memory samples | 98 | 72 | 5 ms target interval |
| Maximum sampling gap, ms | 6.349 | 6.800 | reported limitation |
| Observed descendants | 0 | 0 | no subprocess expected |

The measured process tree is the Go baseline binary and any descendants, not the
Python monitor/build tools. RSS sampling/high-water observations are not an exact
kernel-enforced aggregate peak; short-lived processes can be missed. No inference
was run; **VRAM is unmeasured, not zero or qualified**. Fixed probes are oracle
queries, not demonstrated routing. No task-success, model-fit, cancellation or
real-corpus performance claim follows from this small CPU-only synthetic baseline.
An additional `unshare -n true` probe failed with `Operation not permitted` on
this worker. Runtime network-namespace isolation was therefore not verified;
the loader/baseline have no networking or model transport code, and Go dependency
resolution was explicitly offline. This is not R1 runtime-egress qualification.

## Provenance and preserved evidence

Linux `7.0.0-38-generic`, amd64, Go `1.26.8`; bridge pkg-config `0.1.0`.
Native libraries match F3's fingerprints:

```text
libfffbridge.so  1ee92c96fb4307ad57331fc4cea7f5d09385c6e73ed5337a776dceda84c9bf8d
libfaiss_c.so    d3061d1fa475f937efb1e1f9b5694f44e4abd307761d914be25ea15b8a93ed0e
libfaiss.so      a5448f049afdf22d876067ae1b11d77528e468511c14697a0eb3638aabeedb29
```

Manifest SHA-256:
`a66f4a8f4685b919768b1a56c09426cfad8c297708c4ebcc1ccfb7f5161e473f`.
Final summary `$SCRATCH/f4/report.json` SHA-256:
`4a1f1190b1b2ec01bd48d09a26ba1a23f477aefe600d9c624af6e7184b119568`.
Raw `run-{1,2}/raw.json`, probes, stderr, external state/index/evidence, binaries,
clean export/cache, `f4-provenance.txt` (`go version -m`, native/binary hashes),
and all before/after manifests remain in the same scratch tree.

Committed-corpus before/after manifest hashes both:
`9cf4c074b2c6e51cfe5f05e93abee6621cd6b5fc95c99e1b206860b82a5a32f8`.
Materialized-input manifests match before/after for each run:
`69d997b8ea981e8312d1558fff92d29f69fbf00e14a4ffc750ed90bfd5d5de50` (run 1),
`410ec35c68127c9fe3a5067d2560edaeb95b54eb06c875f5be99895138644cc8` (run 2).
Manifests include paths, modes, size, mtime and file SHA-256; no input effects.

Before/after SHA-256 comparisons across the final full gate also verified all 12
pre-existing dirty/untracked canonical search files and the dirty companion
handoff byte-identical. Those 13-file records are
`$SCRATCH.source-before.json` and `$SCRATCH.source-after.json` (beside scratch).

An initial invocation allocated private scratch but redirected
its log inside the required-empty scratch before invoking F3: the empty-scratch
guard correctly stopped it, exit 1, before any build. No gate was weakened.
A fresh private-scratch run passed; the final full rerun above
adds saved manifests/provenance. All scratch/worktrees remain intact.

## Remaining gates / handoff

Independent **Workmux review is pending**; no integration authorization is
claimed. Coordinator must rerun gates at the local-main integration boundary.
No push, PR, merge, production changes, deletion or worktree/session cleanup.
R1 needs separate artifact/runtime authorization and measured offline model fit;
R2 remains gated on approved F2 integration. Actual end-to-end judged task success
waits for R6. Future RAM/VRAM criteria are predeclared, not measured qualification;
all write/recovery capabilities remain unavailable.
