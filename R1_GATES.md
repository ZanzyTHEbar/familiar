# R1 source slice — gates declared before implementation

Owner: `r1-lazy-local-inference`, base `d7cd899`.
Session: `ses_e7d481278ffeUawd4ZyZDxvGf0`. Date: 2026-10-28.
Authority remains the search repository's LOCAL_AGENT_PLAN / LOCAL_AGENT_POLICY
and the approved design (consolidated separately during public preparation).
No canonical policy change or qualification follows from the design update.

## Scope and enablement

Smallest slice: immutable reviewed candidate identities, narrow inference ports,
generation-first llama.cpp adapter and one owned lazy CPU generation lifecycle.
Embedding/decision ports do not imply implemented adapters or qualification.
No CLI integration, startup work, doctor changes, models, installation, LAN,
retrieval, event database, agent loop, or production enablement.

CHECK before model activity: an explicitly provisioned absolute runtime/model
and containment launcher, reviewed local SHA-256 values, license approval and
resource/containment evidence. Inspect PATH and only operator-documented paths;
never inventory personal/model directories. EXPECT: absent prerequisites block
actual acceptance, not a fallback/download. Runtime memory reservation includes
load peak, weights, context and native buffers; CPU mode must prove no offload.
Containment must deny runtime outbound traffic except owned loopback transport
and cap the process tree, not merely the Go heap. No installed runtime or explicit
model paths were supplied in this assignment; `llama-server` is absent from PATH.

## Lightweight CHECK / EXPECT (permitted before validation slot)

```sh
test -z "$(gofmt -l inference)"
git diff --check
git diff -- go.mod go.sum cmd doctor fixtures scripts
```

EXPECT: formatting/whitespace clean; existing product/dependencies/fixtures and
F3/F4 scripts unchanged. Source review verifies fixed loopback-only transport,
no proxies/redirects/downloads/environment inheritance; bounded requests, queue,
stream frames/aggregate/depth/tool schema; cancellation with no automatic replay;
lease-safe coalesced loading, idle eviction and shutdown.

## Serialized validation slot — NOT RUN until coordinator resumes this session

No background builds/tests. After virtual-vectorfs heavy validation exits:

```sh
umask 077
export SCRATCH=$(mktemp -d "${SCRATCH_PARENT:?select private scratch parent}/familiar-r1.XXXXXX")
: "${SOURCE:?set approved clean search checkout}"
: "${NATIVE:?set installed compatible native prefix}"
timeout 300s bash scripts/check-f3.sh
timeout 180s bash scripts/check-f4.sh
test -z "$(gofmt -l inference)"
git diff --check
```

EXPECT: existing offline F3 builds/vet and actual doctor isolation checks pass;
F4 real native retrieval gates retain their original criteria. No new mocked
process/model tests, no toolchain/dependency/native builds or downloads.

## Actual model acceptance — blocked without preprovisioned inputs

One explicit real offline synthetic-input acceptance must record artifact and
binary hashes separately from published metadata, build/backend/template/tool
parser identity, CPU offload disabled, bounded context refusal, cold load,
coalescing/canceling independent waiters, streaming, a normalized static read-only
tool call (never dispatch), cancellation, last-lease idle eviction and shutdown.
Measure latency/process-tree peak RAM/VRAM and available capacity: <=120 s/call,
<=8 GiB RAM, <=2 GiB incremental VRAM with 512 MiB actual headroom. Missing reliable
telemetry or containment evidence blocks qualification. Embedding qualification
requires primary Liquid 1024d CLS/L2/prefix/512-token reference conformance;
Gemma fallback separately requires 768d projection/mean-pooling/L2/prefix conformance.
F4 end-to-end success still waits for R6; P2 routing evaluation is out of scope.

**R1 is not complete without real runtime/model acceptance.** Source/build checks
are not model evidence. Request the validation slot at the source checkpoint.

## Initial source checkpoint (2026-10-28, before validation slot)

Formatting and tracked `git diff --check` passed. Protected-path diff check
(`go.mod go.sum cmd doctor fixtures scripts`) passed with no changes. New-file
whitespace was also inspected separately because ordinary Git diff omits untracked
files. No build, vet, F3/F4 script, test, model hash/load or inference has run.
Heavy validation is deliberately deferred pending the coordinator's slot.
Actual qualification is blocked on explicit preprovisioned runtime/model/launcher,
license approval, local artifact/build/native provenance, containment/fit telemetry
and a real synthetic-input acceptance. Embedding/decision adapters remain deferred.

## Resumed validation slot

The coordinator granted exclusive heavy validation to this exact session after
virtual-vectorfs exited. Only existing F3/F4 plus explicit all-package build/vet
are permitted; no model/runtime/endpoint or provisioning activity. After F3/F4,
repeat the following in their same offline/native/scratch environment:

```sh
go list -mod=readonly -tags=vectors ./...
timeout 60s go build -mod=readonly -p=2 -tags=vectors ./...
timeout 60s go vet -mod=readonly -p=2 -tags=vectors ./...
test -z "$(gofmt -l cmd doctor inference)"
git diff --check
git diff --exit-code -- go.mod go.sum cmd doctor fixtures scripts
```

EXPECT: package list includes `familiar.local/familiar/inference` even while its
source is untracked; both explicit commands exit 0. Retain F3's no-filesystem/
network doctor behavior and F4's existing input/quality/resource criteria. No
mock/helper tests or model-success claim. The launcher is an unimplemented,
review-required deployment contract: deny all new outbound connections, including
loopback, permitting only replies on the owned loopback listener. Flags/evidence
strings are not proof that containment exists or works.

## Validated-source checkpoint (2026-10-28)

Final source passed existing F3 and F4 scripts, followed by explicit vector-enabled
`go list`, all-package build and vet, formatting and protected-path/diff checks.
All commands exited 0 in the granted exclusive slot, with offline/cached Go
dependencies and installed native libraries; no native builds or provisioning.
Package listing explicitly included `familiar.local/familiar/inference`.

Final scratch is retained in a private maintainer evidence archive. Doctor retained
exit 1/schema 1/not-configured models and unchanged empty workspace/state/absent
HOME; missing-library launch retained expected exit 127. F4 verified 10 files,
13 tasks, both recall@5 1.0, warm p95 0.490372/0.445471 ms and sampled peak tree RSS
38044/38216 KiB, identical judgments and unchanged inputs. These are native
retrieval measurements, NOT model, VRAM or end-to-end agent acceptance.

Source review retained only the requested ports and generation implementation;
no speculative embedding/decision adapters, config registry or sandbox framework.
Fixed reported prompt-token overflow admission and tightened the unimplemented
launcher contract to deny new outbound connections including loopback. The Go
source footprint is 1198 lines: explicit lifecycle/launch, request/stream/schema
handling, bounded HTTP/JSON and the requested value profiles/ports. No unrelated
refactor or unit/helper/mock tests were added to reduce or disguise that footprint.

Exact environment, commands/exits, logs, source hashes and remaining blockers are
retained privately, not distributed with the source. No inference/runtime process
was started. Heavy validation is finished;
the slot is free. **R1 remains incomplete** pending real preprovisioned runtime,
model and containment qualification plus embedding conformance.

## Liquid-first correction — declared before profile changes

User broadened the preference from LFM3-only to suitable official Liquid-family
embedding AND retrieval/reranking models, newest suitable reviewed release first.
Missing an LFM3 embedder is no longer grounds to select Gemma. Generation remains
LFM3. Research date: 2026-10-28; official pinned evidence belongs in inference docs.

CHECK: `EmbeddingCandidate()` describes official LFM2.5-Embedding-350M Q8_0,
1024d CLS with no separate projection, L2/cosine, asymmetric query/document
prefixes and 512-token total input cap. Keep a named Gemma fallback, not automatic
substitution. A separate conditional ColBERT candidate must remain explicitly
multi-vector/MaxSim, with no embedding-port or cross-encoder substitution. No
adapter, ranking, ingestion, Python runtime/dependency or model execution added.
Published hashes are metadata only. Tokenizer, conversion/source correspondence,
padding/masks, pooling/projection and actual runtime output remain qualification
gates; distinguish backend support evidence from local compatibility/fit proof.

EXPECT: generation metadata, doctor/dependencies/corpus/scripts unchanged; no
production enablement. ColBERT requires judged ranking failures and a separate
bounded conformance/benefit decision before implementation; F4 retrieval success
does not justify it. Resolve the model-card document 512 vs sentence config 511
and pad/expansion-mask discrepancies; conservatively cap candidate input at 511,
query at 32 including prefix/special tokens, refuse excess rather than truncate.

After source changes run existing F3/F4 in fresh private scratch, offline and
serially, then the explicit vector all-package list/build/vet and formatting/
protected-path/whitespace checks above. EXPECT every command exit 0 and inference
included. No model probing/provisioning, native builds/install/download, license
acceptance, other-worktree edits or commits. Record the proposed design delta
privately; canonical policy and the design branch stay read-only. R1 remains
incomplete without real generation/lifecycle/offline/fit and embedding acceptance.

Liquid-first corrected source validated on 2026-10-28 in retained private scratch:
existing F3/F4, explicit vector
all-package list/build/vet, format/diff/protected-path checks all exited 0.
The inference package was explicitly listed and compiled, including untracked
profiles. Doctor isolation, absent model endpoints, missing-library control,
fixture inventory/rejections and unchanged corpus/input gates were preserved.
F4 recall@5 was 1.0 in both retrieval runs; warm p95 0.427051/0.486023 ms, sampled
tree RSS 37744/38184 KiB. These are existing native retrieval measurements, not
Liquid/Gemma/ColBERT quality, model fit, VRAM or end-to-end task evidence.

Production change in this correction is only the candidate metadata file (+33
lines, 1231 Go lines total); generation profile/adapter/lifecycle are unchanged.
Practical source review covered coalesced leases, unload/cancellation ordering,
bounded HTTP/JSON and generation cleanup. Documentation now limits its reasoning
claim to the normalized `reasoning_content` field actually checked; raw generated
text is not a reasoning detector. No new adapter/framework or tests were added.
Independent review and all real model/containment/conformance gates still remain.
Full corrected assumptions/pinned sources are in inference/README.md; exact local
command records, artifact locations and private checkpoints are not distributed.
No owned inference runtimes remain; foreground validation finished and slot is free.

## Public preparation (no new builds/tests)

Only source review, portable documentation, staging review and formatting are
authorized in this stage; the upstream owner holds the heavy slot. Prior pass
results above remain historical, not a new integrated build claim. R1 remains
unqualified, and public source licensing/integration/publication remain blocked
as recorded in PRODUCT_CHECKPOINT.md. No main merge/push or repository creation.

## Final serial integration authorization

The user selected Apache-2.0 for Familiar source and authorized normal public
publication. Model/dependency terms are unchanged. Run final merged main build,
vet, full existing Go tests and F3/F4 with vectors, CPU 2 and jobs 1. No models,
native adoption/build, inference, endpoint probing or LAN. Keep R2 OUT of main:
its guarded functional acceptance intentionally fails. Its separate prototype
branch may use published vvfs 88108866 for API compilation only; Ambient remains
canonical and no FFF privacy/adoption authorization exists. Preserve the original
functional assertions and record their real failure, not a skip or fake pass.
