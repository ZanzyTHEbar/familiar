# F4 gates — declared before implementation

Owner/branch: `f4`; target: local `main` after independent Workmux review.
Authority: F4 in canonical `virtual-vectorfs/docs/LOCAL_AGENT_PLAN.md` and F1
policy; no plan transfer. F3 base: `9f80f78`. No R1/R2 implementation.

## CHECK / EXPECT

From this worktree, with installed native libraries and cached Go dependencies:

```sh
umask 077
export SCRATCH=$(mktemp -d "${SCRATCH_PARENT:?select private scratch parent}/familiar-f4.XXXXXX")
: "${SOURCE:?set approved clean search checkout}"
: "${NATIVE:?set installed compatible native prefix}"
timeout 300s bash scripts/check-f3.sh
timeout 180s bash scripts/check-f4.sh
```

F3 script consumes the initially empty scratch, exports only committed search
`fe092dd03d398570b5d93cf0c030f573bc90257b` and supplies an offline temporary
workspace override. F4 reuses that explicit export/environment, not dirty F2.
No downloads, installation, native builds, models, endpoints or inference.

1. Build and vet all Go packages with F3's vector/native/offline environment;
   formatting and `git diff --check` must pass.
2. Offline fixture loader validates schema/version, input SHA-256/UTF-8, exact
   inventory, normalized nonsymlink paths, unique task IDs, complete categories,
   relevant-file judgments and gold half-open UTF-8 byte ranges/quotes. Invalid
   bytes, gold ranges and unexpected inputs must fail before search starts.
3. One real acceptance flow materializes only verified synthetic inputs in fresh
   scratch, opens the pinned native search engine, syncs, and runs fixed path,
   literal grep and keyword probes, with one warmup and 20 measured rounds.
   Record per-query paths/timing and aggregate quality. No generated answers or
   proposal execution; these are retrieval leads, not R2 citations.
4. SHA-256/metadata manifests before/after must match for committed fixtures and
   materialized input tree. Index/evidence/output remain outside that tree.
   Two fresh runs must produce identical quality/judgments (not timings).
5. Measure baseline wall time and sampled Linux process-tree RSS (5 ms target
   sampling plus process high-water RSS where available), report measurement
   limitations and hardware/runtime/native provenance. No VRAM or model fit claim
   from a CPU-only baseline; unsupported VRAM measurement is explicit.

## Predeclared acceptance criteria (do not tune after measuring)

For this tiny synthetic v1 corpus only: all required relevant files in top 5 for
every positive probe; no hits for absent-token negative probes; mean recall@5
1.0; measured warm retrieval p95 <= 500 ms; complete baseline wall <= 30 s;
sampled aggregate process-tree RSS <= 512 MiB. Any exceeded criterion fails the
gate; do not weaken gold to fit retrieval. Cold build/indexing is reported
separately, not mixed into query p95. No extrapolation to personal corpora.

Future R1/R6/P2 qualification criteria, fixed before model measurements: >=90%
judged end-to-end task success overall and >=80% in each positive category;
100% negative-control refusals and zero unsupported claims, invented citations,
or effects. All required facts/ranges must support sourced/cross-document answers;
duplicates distinguish exact bytes from similar/current versions; organization
is proposal-only with explicit scope and approval requirement. Task p95 <=120 s,
process-tree peak RAM <=8 GiB and incremental per-process VRAM <=2 GiB, also <=
actually available capacity with 512 MiB system/VRAM headroom. Missing reliable
memory measurement blocks qualification, not success. F1 context/call/task limits
still apply. These are enablement ceilings, not claims that any model fits.
Actual task success waits for R6; model/runtime/artifact qualification waits for
separately authorized R1. R2 still requires approved F2 integration.

## Results

Passed locally on 2026-10-21 in retained private scratch: F3 build/vet/
doctor/loader checks and both fresh F4 retrieval runs, all three loader rejection
checks, and identical input manifests. Mean recall@5 was 1.0 in both runs;
overall warm p95 0.819/1.072 ms; sampled peak RSS 38,296/37,916 KiB. See
[F4_EVIDENCE.md](F4_EVIDENCE.md) for exact commands, provenance and limits.
This is not model/VRAM or end-to-end task qualification. Independent Workmux
review and coordinator integration-boundary re-verification remain mandatory.
