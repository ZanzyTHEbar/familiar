# R2 local execution evidence — 2026-10-28

## Current safety-disabled prototype (final publication)

Public `retrieval.Open` now returns `ErrUnavailable` before argument inspection,
filesystem access, state creation or native open. The former implementation is
uncalled internal prototype code, not an enabled feature. No flag or environment
fallback is provided. This deliberately changes the acceptance contract: original
functional assertions remain unchanged and cannot be claimed to pass through the
denied API. No-side-effect denial regression qualification and any future approved
private native proof remain pending. Preparation ran no builds/tests; final
branch compilation/vet and original functional acceptance results follow.

The module/gate pin is published `88108866c3cac44d7a66e1a9532d35b1f2271ecf`
(`v0.0.0-20261028155634-88108866c3ca`) retrieved through normal Go modules,
checksum `h1:HCWqs3HAJU7TKxrHcZfgiLXaVR5rqdXiKG852L50Un4=` with exact
Origin commit verified. No workspace replacement; offline vector build and vet
both passed on this branch using installed canonical native libraries, CPU 2,
Go jobs 1. Canonical production remains Ambient; no FFF adoption/privacy approval.
Original `TestScopedEvidenceAcceptance` ran ONCE and FAILED (exit 1) at line 97
with `retrieval: experimental API unavailable until reviewed RootOnly dependency and privacy acceptance`.
The unchanged test SHA-256 is
`a085b4e3c94c2163b7add9716f6a6747fa23aa8dad6c63aee00710310e9a682b`.
This is genuine unresolved failed functional acceptance, not passing a new contract.
The branch is committed/published separately as experimental, excluded from main.
`1e2e882...` below remains historical only. Future reviewed RootOnly integration
and privacy acceptance must precede any re-enablement; the guard stays immediate.
No dirty upstream source, retired FFF copy or unpublished pin is adopted here.
Host-specific archive/report paths in this public summary use logical private
archive labels; measurements, failed checks, source pins and hashes are unchanged.

## Historical outcome and ownership (before safety denial)

Implemented locally; original declared F3/F4/R2 gates passed. **Independent review
subsequently found a native ambient-read scope BLOCKER: NOT integration-ready,
not accepted, integrated, committed or published.** Original measurements below
are historical evidence, not proof of native read confinement. Owner/session:
`ses_e7dbabeedffeMJzMaIV9tJFUEO`; sole finite worker, no fanout.
Worktree `r2-scoped-evidence`,
branch `r2-scoped-evidence`, unchanged HEAD/main
`d7cd8999643bdbb74603e6afa25dcf226f380a00`.

Private durable R2 record archive (0700; record files 0600; not distributed).
Contains ownership, results/checkpoint, expanded command logs, hashes, provenance,
final diff/status, baseline and final scratch. No personal data or credentials.

Canonical plan/policy and `git show review-f3:F3_REVIEW_OC2.md` were read without
modification. R2_GATES.md preceded production code; production code preceded the
single real acceptance test. No mocked/unit-helper suite or CLI framework added.

## Exact gate invocations and outcomes

### Preview-fix continuation (historical, still BLOCKED)

Only the reviewed display defect was fixed: `Candidate.Lead` copies the hit,
limits Preview to the existing 512-byte upstream preview budget, retreats to a
UTF-8 boundary and marks PreviewTruncated when bytes are omitted. It clones only
the displayed prefix; private candidate ID/source/hash/range binding is unchanged.
The existing real acceptance journey now includes a 786,447-byte UTF-8 title-only
record (empty Text/Context), verifies its exact bounded prefix/truncation and
valid UTF-8, and expands the original title's final rune beyond the preview with
its unchanged hash and exact byte range. No new unit/matrix suite.

Fresh serial validation: `timeout 660s bash -x <private-dir>/preview-review/validate.sh`
exited 0, using SCRATCH `<private-dir>/preview-review/scratch`, offline resolver,
same exact committed dependency/native libraries and existing bounded Go/native
settings. Expanded commands and output in `preview-review/{validation,f3,f4,r2}.log`:

| Command (serial, no assertion changes) | Exit |
| --- | --- |
| `timeout 300s bash -x scripts/check-f3.sh` | 0 |
| `timeout 180s bash -x scripts/check-f4.sh` | 0 |
| `timeout 120s bash -x scripts/check-r2.sh` | 0 |
| `bash -n scripts/check-f3.sh scripts/check-f4.sh scripts/check-r2.sh` | 0 |
| `git diff --check` | 0 |

Build/vet all packages and gofmt checks passed. Real R2 acceptance 0.42s, package
0.464s, including new title-only assertion and all existing range/scope/output and
input-preservation assertions. Fresh F4 recall@5 1.0 both runs; p95
0.450261/0.502922ms; wall 0.636108/0.313326s; sampled peak
37,724/37,844KiB. All original thresholds and negative-control checks passed.
F3 doctor retained expected exit 1 blockers and loader exit 127; these are
expected assertions, not readiness or native confinement claims.

Prior logs, audit artifacts and patch/status remain unchanged; prior results/
patch/status/hashes are additionally archived as `preview-review/prior-*`.
Fresh source hashes, continuation diff/status, input-preservation results and
owned-runtime observation are recorded under `preview-review/`. Reviewer brief
is `preview-review/review-findings.md`, full captured report path/hash provenance
is `preview-review/review-provenance.json`. No raw review transcript in repo.
The original supplementary audit checker remains failed and was not rerun or
rewritten. **These successful functional checks do not resolve the native
ambient-read blocker or establish integration readiness.**

### Original run (preserved historical evidence)

From this worktree, `umask 077`, with
`SOURCE=<approved-clean-search-checkout>` and
`NATIVE="$SOURCE/build/native"` exported:

| SCRATCH beneath private directory | Command | Exit |
| --- | --- | --- |
| `baseline.FOfnQj` (fresh, before edits) | `timeout 300s bash -x scripts/check-f3.sh` | 0 |
| same | `timeout 180s bash -x scripts/check-f4.sh` | 0 |
| `final.JEZMdV` (fresh, new pin, production code) | `timeout 300s bash -x scripts/check-f3.sh` | 0 |
| same, after acceptance code | `timeout 180s bash -x scripts/check-f4.sh` | 0 |
| same | `timeout 120s bash -x scripts/check-r2.sh` | 0 |

Logs: `baseline-{f3,f4}.log`, `final-{f3,f4}.log`, `r2.log`.
All build/vet commands exited 0; formatting/diff checks passed. F3 doctor exited
1 with the expected explicit model blockers; missing-native loader exited 127
with actual missing `libfffbridge.so`. Original assertions/thresholds retained.

The R2 script ran `go test -mod=readonly -count=1 -p=2 -tags=vectors
./retrieval -run TestScopedEvidenceAcceptance -v -timeout 45s` with a 60s outer
timeout: exit 0, acceptance scenario 0.17s, package 0.212s. It exercises real
Search/Expand: separate roots/competing content, cross-scope and zero-candidate
refusal, immutable grants/record snapshots, exact UTF-8 range beyond a truncated
preview, field/byte limits, record-versus-file origins, and refusal after changes,
rename/removal, escaping symlink and replaced root with identical bytes.
Wrong source/hash/path values in a display copy cannot change the private binding;
the API accepts no caller-supplied expansion hash/source/path. No grep offsets
are used. Before/after synthetic input manifests (bytes, modes, size, mtime) match
before intentional acceptance mutations. Committed/materialized F4 manifests
also match in both baseline and final runs.

New-pin F4 runs: recall@5 1.0 in both; all absent controls passed; p95
0.436061/0.426301 ms; wall 0.530848/0.308046 s; sampled tree peak
37,892/37,848 KiB. Same judgments in both runs, all predeclared thresholds passed.
Sampling is not a hard aggregate-memory bound; no model/task-success/VRAM claim.

## Dependency, native identity and limits

Search pin `1e2e882fe8808c4601ecd61ebd9e98e39404844e`, committed UTC
`2026-10-28T11:18:53Z`, yields
`v0.0.0-20261028111853-1e2e882fe880`. Only clean `git archive` plus explicit
version-specific scratch GOWORK replacement was used. No host replacement,
dirty source copy, fabricated checksum or downloaded module archive is committed.
go.sum is unchanged. Clean archive SHA-256:
`1854644bf411cd0e562315af7e06ca103b51e12e44d5bf60e8bbe614c56182a5`.

Linux amd64, kernel `7.0.0-38-generic`, Go `1.26.8`. Installed native SHA-256:

```text
libfffbridge.so 1ee92c96fb4307ad57331fc4cea7f5d09385c6e73ed5337a776dceda84c9bf8d
libfaiss_c.so   d3061d1fa475f937efb1e1f9b5694f44e4abd307761d914be25ea15b8a93ed0e
libfaiss.so     a5448f049afdf22d876067ae1b11d77528e468511c14697a0eb3638aabeedb29
```

Go resolution offline (`GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local`), CGO/vectors,
isolated GOCACHE/TMPDIR; Go/build/native threads bounded to two, Go soft memory
target 256 MiB, per-command timeouts. This is not a hard native memory limit.
No installs, downloads, native builds, model probes or endpoint calls occurred.

Supplementary audit: `timeout 120s bash -x <private-dir>/audit.sh` **exited 1**.
Its compiled real API scenario itself passed (0.42s; process command exit 0).
The additional `test ! -s audit.network` checker failed because strace includes
20 Go SIGURG signal lines, not network calls. Checker/log retained unchanged;
read-only trace inspection found zero network syscall lines. This is an observed
no-network flow, not an egress sandbox qualification. Audit used an empty
environment except explicit runtime variables, absent HOME, CPU rlimit 30s,
descriptor rlimit 512, outer timeout 45s; measured wall 0.60s, max-child RSS
37,860 KiB (not summed aggregate peak). See audit.sh/log/network/time and
runtime-observation.json; do not describe the audit script as passing.

## Scope assumptions and checkpoint

**Correction from independent review `ses_e7daee679ffeuu1oqgYQWJ9yei`:** native
FFF starts even for record-only adapters, discovers ancestor Git repositories
(`fff-search file_picker.rs:2050-2054`), starts repo-wide status scanning
(`scan.rs:206-210`, `shared.rs:367-375`), and enables ambient Git/global ignores
(`ripgrep.rs:20-27`). Separate indexes constrain returned evidence but do not
confine all reads to the caller's root/source grants. The earlier os.OpenRoot /
follow_symlinks=false observation applies only to specific paths, not all native
reads. This is an unresolved security-scope defect, not an acceptable assumption
or a sandbox-disclaimer limitation. No native-read confinement claim is made.
The review brief and captured-report provenance are archived privately, not
distributed with the source repository.

Every adapter creates a fresh index for its explicit root/source grant, never a
shared index filtered only after retrieval. Source checks on results are defensive,
not the authorization mechanism. Caller records are immutable indexed snapshots;
their path labels confer no live-file access. Private state remains outside inputs.
All adapters were closed; no executable under the owned private prefix remained
running when checked via `/proc/*/exe`. No background runtime was launched.

This is not an OS sandbox. Caller-selected whole roots must exclude secrets and
ungranted subtrees; ancestry/mounts must remain stable during operations. Root
identity changes between calls are refused; same-UID adversarial replacement
races are not guaranteed. Upstream file reads use os.OpenRoot and FFF discovery
sets follow_symlinks=false in the exact clean pin. Evidence is current during
verification, not a lease against later edits. Corpus scale/native cancellation
and ABI generalization are not qualified by this synthetic flow.

**Stop: R2 BLOCKED, not integration-ready.** Recommended next bounded upstream
change: an option disabling Git discovery/status and external ignore/config reads
for scoped consumers, with real nested-repository acceptance, before repinning R2.
No upstream change or ambient environment/config workaround was attempted here.
No R1/R3/R4 work, plan
transfer, search edits, commit, merge, push or PR. Main/other branches untouched.
Search's final read-only status shows its pre-existing untracked `.NOTE.md`; it
was neither read nor modified. Canonical plan still labels F3/F4 planned despite
the user-supplied integrated Familiar base. Recommend coordinator reconcile
F3/F4 local implementation plus these fresh boundary checks and the existing F3
review, obtain/check F4 and R2 independent acceptance, then update canonical
states only after authorized integration. Old F4 pending-review text establishes
neither acceptance nor a current blocker by itself. F2 remains locally integrated,
not published. No F/R integration-wide completion is claimed.
