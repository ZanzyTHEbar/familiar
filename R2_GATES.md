# R2 gates — defined before implementation

## Current experimental branch contract (2026-10-28)

R2 is **safety-disabled and experimental**, not a working public retrieval API.
Before any validation, state creation or native open, public `retrieval.Open`
must return `(nil, ErrUnavailable)` without reading caller paths, inspecting
arguments, or creating state. There is no flag/environment fallback enabling it.
The retained internal prototype is unmerged and not an approved native-read path.

The branch module pin is published `88108866c3cac44d7a66e1a9532d35b1f2271ecf`,
`v0.0.0-20261028155634-88108866c3ca`, materialized through normal authenticated
module checksum verification, not a local replacement. Branch vector build/vet
passed offline against installed canonical native libraries. This is API source
compatibility, not native adoption or privacy acceptance: canonical production is
Ambient; FFF/RootOnly integration remains blocked. Main does not merge this branch.
The older `1e2e882...` runs below are historical, not an adopted dependency.

Original functional acceptance assertions remain unchanged. They are NOT expected
to pass through this denied API. No assertions are skipped/weakened to manufacture
success. A real no-side-effect denial regression check and any separately approved
private native proof/functional requalification remain pending. Original functional
acceptance ran once in final integration and FAILED at denied Open (exit 1).
A future
enablement change must first prove scoped native reads, then rerun the unchanged
functional journey with the approved stable dependency. The original successful
retrieval measurements below do not qualify the new disabled contract.

Final branch checks: `GOWORK=off GOPROXY=off GOSUMDB=off`, `GOMAXPROCS=2`,
`GOMEMLIMIT=256MiB`, native thread limits 2, `CGO_ENABLED=1`, installed native
pkg-config/link paths; bounded `go build -mod=readonly -p=1 -tags=vectors ./...`
and `go vet -mod=readonly -p=1 -tags=vectors ./...` both exit 0. Unchanged
`go test -mod=readonly -count=1 -p=1 -tags=vectors ./retrieval
-run TestScopedEvidenceAcceptance -v -timeout 45s` exits 1 with
`retrieval: experimental API unavailable until reviewed RootOnly dependency and privacy acceptance`.
This genuine unresolved failure excludes R2 from main. No new tests, bypasses,
skips or assertion changes. `scripts/check-r2.sh` retains the failing journey.

Owner: sole finite worker `ses_e7dbabeedffeMJzMaIV9tJFUEO`, branch
`r2-scoped-evidence`, base `d7cd899`. Canonical plan/policy remain in search;
this is a local execution contract, not a transferred plan.

## Historical review continuation — BLOCKED, not integration-ready

Independent read-only review `ses_e7daee679ffeuu1oqgYQWJ9yei` identified an
unbounded title-only preview and a native ambient-read scope blocker. Only the
preview fix is authorized here. Before that fix, add this CHECK/EXPECT:
extend the existing real acceptance journey with a large UTF-8 title-only record;
EXPECT Lead.Preview <=512 bytes, valid UTF-8, accurate PreviewTruncated and exact
prefix, while title expansion beyond the preview still verifies the original
hash/range. Run F3, F4 and R2 scripts serially offline in a fresh isolated scratch,
retaining their existing build/vet/format and input-preservation assertions.

**Passing these checks does not resolve native read scope.** FFF starts even for
record-only adapters, discovers ancestor Git repositories and launches repo-wide
status scans; its walker enables ambient Git/global ignores. Separate indexes
constrain retrieved evidence but do not confine all native reads to grants.
os.OpenRoot and follow_symlinks=false do not cover those ambient reads. This is
an unresolved scope defect, not a sandbox disclaimer or an accepted assumption.
Do not integrate R2 until an upstream option disables Git discovery/status AND
external ignore/config reads for scoped consumers, passes a real nested-repository
acceptance flow, and is deliberately repinned/revalidated downstream. No upstream
fix, global environment/config mutation or weakening of checks is authorized here.
Reviewer brief and captured-report provenance are archived privately, not
distributed with the public source.

## Historical inherited baseline (first, before edits)

CHECK: existing `timeout 300s bash -x scripts/check-f3.sh`, then
`timeout 180s bash -x scripts/check-f4.sh`, with fresh isolated SCRATCH,
SOURCE and installed NATIVE. EXPECT both exit 0; stop on failure.
Actual: both exit 0, logs in
the private R2 archive's `baseline-{f3,f4}.log`;
scratch `baseline.FOfnQj` in that same private directory. Old F4 pending-review
text is not used as evidence of current acceptance.

## Retained prototype's declared implementation boundary

One reusable read-only adapter, one root per instance (optional for record-only
use), an explicit immutable source allowlist, fresh separate indexes per instance.
No shared global index with post-filter-only authorization. Source records are
bounded caller-owned snapshots, not file authority even if their paths name files.
No mutation API, exposed engine, models, agent loop, broker, SQLite, installs,
network transports, personal inventories or other-repository edits.

Caller explicitly grants a whole root and sources; it is responsible for excluding
secrets and subtrees before granting. R2 does not implement subtree deny rules or
an OS sandbox. Root/state paths must be absolute, canonical and symlink-free;
root identity is checked on operations. Upstream uses os.OpenRoot for indexed
content reads, not for all native ambient Git/ignore/config reads (blocker above).
Root ancestry/mounts must not be changed concurrently by a non-cooperating writer;
no same-UID adversarial race guarantee. Other ordinary file edits between calls
are allowed and must invalidate old evidence. “Current” means verified during the
call, not a lease over subsequent edits. Existing citations remain versioned facts,
not an assertion of continued freshness.

## Historical CHECK / EXPECT (not current acceptance)

1. Pin `1e2e882fe8808c4601ecd61ebd9e98e39404844e` with timestamp-derived
   pseudo-version `v0.0.0-20261028111853-1e2e882fe880`. Adjust F3/F4 executable
   gates to the declared pin; preserve prior historical evidence. Use only clean
   `git archive` plus version-specific scratch-only GOWORK override, no fabricated
   sums/committed host replacements. Retain all existing F3/F4 assertions.
2. In fresh SCRATCH, run F3 then F4 as above. EXPECT all existing checks pass
   against the new pin, unchanged fixture/materialized bytes, same judgments and
   predeclared F4 quality/latency/RSS limits. No acceptance claim based on old text.
3. Run `timeout 120s bash -x scripts/check-r2.sh` in that same exported scratch.
   Script uses offline Go/native settings and `go test -count=1 -p=2 -tags=vectors
   ./retrieval -run TestScopedEvidenceAcceptance -v -timeout 45s`, build/vet all
   packages, `gofmt -l cmd doctor retrieval` and `git diff --check`.
   EXPECT one real API acceptance flow against native Search/Expand, no mocks:
   - separately scoped roots with competing content/identical relative names do
     not affect each other's retrieval; foreign/forged candidates are refused;
   - source grants are immutable; unauthorized records rejected before indexing;
   - lead previews do not claim exact matches; a range beyond a long preview and
     UTF-8 bytes verifies exactly against current source, with bounded excerpts;
   - changed/missing/renamed files and wrong hashes/sources cannot be cited current;
   - caller record expansion explicitly says indexed_record, verifies its own
     field/range and never reads its path; unsupported fields/ranges/limits fail;
   - symlink escape/root replacement is refused; no grep display offsets used;
   - synthetic input manifests/hashes are byte-identical after read-only use,
     before intentional acceptance mutations; committed fixture manifests match
     before/after the full gates. Private index/evidence stays outside input roots.
4. Record exact commands, outcomes, native/source identities, resource bounds,
   final diff/status and checkpoint in private 0700/0600 durable records. No
   background runtime remains; Close drains upstream workers. No commit/publish.

Stop at this R2 checkpoint. Local passing gates are not independent review,
integration, canonical-plan completion, R1/R3/R4 readiness or product-wide acceptance.
