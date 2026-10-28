# Familiar lazy LFM3 inference and policy-bounded routing

Research date: **2026-10-28**. Status: **recommendation, not implementation or qualification**.
This document persists the sole designer's findings; it does not amend canonical plans or authorize runtime activity.
Public-preparation update: the user broadened embedding/retrieval preference to
the Liquid family, not just LFM3. The earlier Gemma-primary recommendation is
superseded below. Historical metadata/research is not local artifact verification.

## Recommendation and boundaries

Use the existing Go harness, immutable pinned profiles, one scheduler/lazy process manager, and narrow `Generate`, `Decide`, and `Embed` ports.
Reuse one llama.cpp HTTP adapter for owned loopback processes and separately allowlisted LAN servers, while keeping their lifecycle guarantees distinct.
Prefer reviewed LFM3 releases for generation; “LFM3+” means reviewed preference, never automatic upgrades or fetching.
Prefer suitable official **Liquid-family embedding and retrieval/reranking models**,
newest suitable reviewed release first. Use Gemma only if no suitable compatible
Liquid integration can be qualified; absence of an LFM3-specific embedder is not
a fallback criterion. Late-interaction retrieval remains conditional on judged need.
Start with deterministic routing. Enable a P2 decision model only if measured quality, latency, and memory gates beat that baseline.
Keep offline operation the default; LAN requires separate authorization. No agent framework, plugin system, Python runtime, or in-process inference CGO binding is needed.

| Alternative | Assessment |
| --- | --- |
| Direct llama.cpp sidecar | Pinned process separation; qualified containment is still required. |
| Operator-owned compatible server | Shares the qualified adapter, not owned lifecycle, isolation, or capacity guarantees. |
| Python Transformers d1 reference | Useful if typed extraction later becomes necessary; adds runtime burden and is not the first adapter. |
| Wrappers | Consider only after a specific compatibility proof; generic “OpenAI-compatible” claims are insufficient. |

## Repository and authority boundary

Familiar is a separate Go 1.25 module. R1 has a compiled, unqualified generation
prototype on its task branch, not a configured CLI runtime or agent loop.
Preserve `doctor.Inspect() doctor.Report`, schema version 1, and its filesystem/network-free default: generation and embedding remain `not-configured`.
Add any future qualification as an explicit opt-in command, not startup checks or implicit doctor probes.
The committed search dependency is `fe092dd03d398570b5d93cf0c030f573bc90257b`.
The historical R2 prototype proposed unpublished `1e2e882fe8808c4601ecd61ebd9e98e39404844e`;
that is not the accepted module pin. Public R2 `Open` is now safety-disabled before
argument/filesystem/native access. Its retained prototype stays unmerged pending
a reviewed published RootOnly dependency and actual privacy acceptance. Earlier
scope-patch observations are not evidence of current integration or clearance.
R2 binds candidates to its adapter and version-checks expansion: 512-byte previews, 20 candidates, and 4096-byte excerpts. Search leads are not exact evidence or authority.
R2's keyword-only hidden engine dimension-1 placeholder must never become an embedding profile or expose the engine.
R3 may use narrow authorized `MissingVectors`/`UpsertVectors` ingestion only once the scoped boundary is ready.
Canonical authority remains the upstream
[delivery plan](https://github.com/ZanzyTHEbar/virtual-vectorfs/blob/main/docs/LOCAL_AGENT_PLAN.md)
and [policy](https://github.com/ZanzyTHEbar/virtual-vectorfs/blob/main/docs/LOCAL_AGENT_POLICY.md);
this is not a second backlog or a plan transfer.

Local source checks for this persistence task: `README.md`, `go.mod`, `doctor/doctor.go`, and `F4_GATES.md`. R2 findings above are supplied designer findings, not a new writer inspection.

## Release evidence and pinned candidates

Liquid's official library dates “Introducing LFM3” to **2026-10-27**.
Released identities include `LFM3-1.7B-Base`, `LFM3-1.7B`, `VL-2B`, `Audio-2B`, `Omni-2B`, `d1-2B`, and `d1-Omni-2B`.
The text model has 1.73B parameters and documented 256K context; Familiar still enforces an **8K ceiling**.
Tool use is claimed, with warnings about complex multi-step reasoning and agentic coding. Vendor benchmarks are not Familiar qualification.
Sources (accessed 2026-10-28): [official models/library](https://www.liquid.ai/models), [collection API](https://huggingface.co/api/collections/LiquidAI/lfm3), [text documentation](https://docs.liquid.ai/lfm/models/lfm3-1.7b), [text card](https://huggingface.co/LiquidAI/LFM3-1.7B/raw/main/README.md).
The [announcement](https://www.liquid.ai/blog/lfm3) direct body fetch exceeded the designer's tool response limit; the release date was verified from the library, not that body.

### Generation: LFM3-1.7B Q4_K_M

- Repository: `LiquidAI/LFM3-1.7B-GGUF`.
- Revision: `b3f2ae5d73cd3c91bf40e8ba752c648aa05dffdf`.
- Artifact: `LFM3-1.7B-Q4_K_M.gguf`; **1099627808 bytes**.
- SHA-256: `3793e1e6e2266919d6e015287221b41100175a98eddaba92e111aed3b05c93c4`.
- Published alternatives BF16/F16, Q4_K_M, Q5_K_M, Q6_K, Q8_0, and QAD-Q4_0 are artifact availability, not proof of backend compatibility.

Source (accessed 2026-10-28): [artifact metadata](https://huggingface.co/api/models/LiquidAI/LFM3-1.7B-GGUF?blobs=true).

### Optional P2 router: d1-2B Q4_K_M

- Repository: `LiquidAI/d1-2B-GGUF`.
- Revision: `63be75e6586882defeb706ed4bda812d90256a6d`.
- Artifact: `d1-2B-Q4_K_M.gguf`; **1099629888 bytes**.
- SHA-256: `adca5789ad30ae6027e8e936749bb9f19673097492ab26fa1c5b69643ca50b4b`.
- This LFM3-derived decision model uses **`/v1/systemone`, not chat completions**. First support is text-only fixed choices with probabilities, not unrestricted Transformers typed extraction.
- GGUF card/support text saying “coming” or “pending” is stale relative to merged upstream support and published artifacts.
- Optional `d1-omni-600M` is **587M and LFM2.5, not LFM3**. Never silently substitute it; allow only an explicit router-role exception if P2 proves it better. No extra pin is needed for this unselected alternative.

Sources (accessed 2026-10-28): [decision models](https://docs.liquid.ai/lfm/models/decision-models.md), [d1-2B](https://docs.liquid.ai/lfm/models/d1-2b.md), [artifact metadata](https://huggingface.co/api/models/LiquidAI/d1-2B-GGUF?blobs=true), [pinned GGUF card](https://huggingface.co/LiquidAI/d1-2B-GGUF/raw/63be75e6586882defeb706ed4bda812d90256a6d/README.md), [smaller alternative](https://docs.liquid.ai/lfm/models/d1-omni-600m.md).

### Runtime: llama.cpp v0.8.0

- Source: `8642541b255bfb5dc39cd351a178be09013a55d6`; published **2026-10-28T12:47:51Z**, explicitly supporting LFM3 and d1.
- LFM3 support merge: `19cce6919992fdffd92e0ac152caaf1e2b8f254e`.
- d1 support merge: `d05706b4f32a545814856e98ca518292585443b9`.
- EmbeddingGemma2 support merge: `4fbc76dec51d0add466f0210855c0596589b60d4`.

Sources (accessed 2026-10-28): [release API](https://api.github.com/repos/ggml-org/llama.cpp/releases/tags/v0.8.0) (independently fetched by coordinator), [release](https://github.com/ggml-org/llama.cpp/releases/tag/v0.8.0), [LFM3 PR](https://api.github.com/repos/ggml-org/llama.cpp/pulls/31252), [d1 PR](https://api.github.com/repos/ggml-org/llama.cpp/pulls/31275), [embedding PR](https://api.github.com/repos/ggml-org/llama.cpp/pulls/30054).

### Primary embedding: Liquid LFM2.5-Embedding-350M Q8_0

- Repository: `LiquidAI/LFM2.5-Embedding-350M-GGUF`.
- Revision: `a80de9c5b941d429104f0038292a0ef5a860e486`.
- Artifact: `LFM2.5-Embedding-350M-Q8_0.gguf`; **379216640 bytes**.
- Published SHA-256: `6ec5f8e8750dbc8a0e40c431fd1b7b07a13688136b2244c5a1364b54d9032599`.
- Trained bidirectional dense bi-encoder: **1024d CLS**, no separate learned
  projection, L2 normalization and cosine/normalized-dot similarity. Not generator
  hidden-state pooling, mean pooling or a relevance probability.
- Raw prefixes `query: ` / `document: `, case preserved, reference BOS id 1,
  no added EOS, pad id 0. **512 total tokens including prefix/specials**, despite
  GGUF metadata's 128000. Refuse excess; no silent truncation or unversioned chunks.
- Reference configs/card: `f35ae2c91d687658dbf1f2b449382f0b019b9808`. The official
  GGUF manifest does not attest conversion source/build; current base reference
  configs are NOT asserted conversion provenance. Actual tokenizer/CLS/norm and
  reference equivalence remain qualification gates.
- Official GGUF/card and reviewed llama.cpp source support an embedding-role
  `/v1/embeddings` path. No different backend is required by current evidence;
  the generation-only R1 owner is not an implemented embedding adapter.

Sources accessed 2026-10-28: [official embedding docs](https://docs.liquid.ai/lfm/models/lfm25-embedding-350m),
[pinned artifact metadata](https://huggingface.co/api/models/LiquidAI/LFM2.5-Embedding-350M-GGUF/revision/a80de9c5b941d429104f0038292a0ef5a860e486?blobs=true),
[GGUF card](https://huggingface.co/LiquidAI/LFM2.5-Embedding-350M-GGUF/raw/a80de9c5b941d429104f0038292a0ef5a860e486/README.md),
[reference card](https://huggingface.co/LiquidAI/LFM2.5-Embedding-350M/raw/f35ae2c91d687658dbf1f2b449382f0b019b9808/README.md),
[pooling](https://huggingface.co/LiquidAI/LFM2.5-Embedding-350M/raw/f35ae2c91d687658dbf1f2b449382f0b019b9808/1_Pooling/config.json),
[prefixes](https://huggingface.co/LiquidAI/LFM2.5-Embedding-350M/raw/f35ae2c91d687658dbf1f2b449382f0b019b9808/config_sentence_transformers.json).

### Conditional retrieval/reranking: Liquid LFM2.5-ColBERT-350M Q8_0

- Repository: `LiquidAI/LFM2.5-ColBERT-350M-GGUF`; revision
  `bc240003aba07253e261a8aaf0d2c9683318a967`.
- Artifact: `LFM2.5-ColBERT-350M-Q8_0.gguf`; **378097248 bytes**.
- Published SHA-256: `5765e4987f1e63e6c52572198abd65fe9616416edf523c0fa95a077a3e501d47`.
- **128d per token**, no pooling; required learned 1024→128 projection (no bias,
  identity activation), per-token L2 and MaxSim: sum of each query token's maximum
  document-token dot product. Relative score, not cosine of document vectors or
  a calibrated probability; do not mix scales/models or infer access authority.
- Prefixes `[Q] ` / `[D] `, query cap/pad-to **32**; pinned punctuation skiplist.
  Reference `9772bdf797255d8693b83e84aa98e9b2d36dd0be` says document 512 but sentence
  config 511: candidate conservatively caps total input at **511**. No truncation.
  Tokenizer pad EOS id 7 vs model-config pad 0, query expansion and
  `attend_to_expansion_tokens=false` require exact tokenizer/mask conformance.
- A future raw `/embedding` per-token path plus verified client preprocessing/
  MaxSim could rerank bounded first-stage candidates. It is NOT a single-vector
  Embedder, stock `/v1/rerank` cross-encoder or generative retriever. Do not assume
  the vendor Python/token-ID demo proves masks or adopt its downloads/truncation.
- No reranking/index/R3 overhaul until judged failures justify a separate bounded
  benefit/conformance decision. Current tiny F4 retrieval passes; vendor scores
  and benchmarks are not Familiar qualification. Conversion provenance is unattested.

Sources accessed 2026-10-28: [official ColBERT docs](https://docs.liquid.ai/lfm/models/lfm25-colbert-350m),
[pinned metadata](https://huggingface.co/api/models/LiquidAI/LFM2.5-ColBERT-350M-GGUF/revision/bc240003aba07253e261a8aaf0d2c9683318a967?blobs=true),
[GGUF card](https://huggingface.co/LiquidAI/LFM2.5-ColBERT-350M-GGUF/raw/bc240003aba07253e261a8aaf0d2c9683318a967/README.md),
[reference preprocessing](https://huggingface.co/LiquidAI/LFM2.5-ColBERT-350M/raw/9772bdf797255d8693b83e84aa98e9b2d36dd0be/config_sentence_transformers.json),
[projection](https://huggingface.co/LiquidAI/LFM2.5-ColBERT-350M/raw/9772bdf797255d8693b83e84aa98e9b2d36dd0be/1_Dense/config.json),
[runtime converter](https://github.com/ggml-org/llama.cpp/blob/8642541b255bfb5dc39cd351a178be09013a55d6/conversion/lfm2.py).
The [complete library](https://docs.liquid.ai/lfm/models/complete-library.md) and
[Liquid Nanos collection](https://huggingface.co/api/collections/LiquidAI/liquid-nanos-68b98d898414dd94d4d5f99a)
identify these Liquid-family retrieval models. Generic encoders are not trained
retriever substitutes; LFM2-1.2B-RAG generates answers from supplied context,
not embeddings or relevance scores. No automatic backend/model fallback.

### Explicit fallback only: EmbeddingGemma2 Q8_0

- The first recommendation selected this fallback under the old LFM3-only scope.
  The user's broader Liquid-first correction supersedes that selection. Consider
  Gemma only if no suitable compatible Liquid integration can be qualified.
- Repository: `ggml-org/embeddinggemma-2-GGUF`; revision `58faf90940886df2d62be555c8765d00dac362c5`.
- Artifact: `embeddinggemma-2-Q8_0.gguf`; **309855456 bytes**.
- SHA-256: `2188ac1deca4b77dffefd603c2776a9d76d9d74ec01841392982ebb840b09135`.
- Conversion `.src_sha`: `914f7f89142e33e77833254d9c9b90c3cef7303b`; current Google main differs.
- Require **768 dimensions, mean pooling, the required learned projection, and L2 normalization**.
- Query prefix: `task: search result | query: …`; document prefix: `title: … | text: …`.
- Model card says 8192 tokens; converted metadata says 262144. Enforce the smaller documented limit and applicable policy limit pending R1; never silently truncate.
- Model-profile fingerprint covers artifact, tokenizer, template, runtime and conversion, pooling, projection, dimensions, normalization, prefixes, and chunking policy.
- Any model-profile change invalidates vectors, even at identical dimensions. Build a new derived namespace, validate it, then switch; never mix silently.
- Each vector separately records its source content/range hash. A source hash change invalidates only affected vectors, not the full namespace.

Sources (accessed 2026-10-28): [Google card](https://huggingface.co/google/embeddinggemma-2/raw/main/README.md), [artifact metadata](https://huggingface.co/api/models/ggml-org/embeddinggemma-2-GGUF?blobs=true), [conversion source](https://huggingface.co/ggml-org/embeddinggemma-2-GGUF/raw/58faf90940886df2d62be555c8765d00dac362c5/.src_sha).

### Licensing and evidence limits

Liquid's LFM Open License v1.0 has a $10M annual-revenue threshold for a separate commercial license, plus attribution/redistribution obligations. Accept applicable terms before provisioning or distribution.
EmbeddingGemma2 declares Apache-2.0; review the full applicable terms.
All hashes above are **published metadata only**, not local artifact verification.
Binary hash, build flags, backend/native libraries, tokenizer/template, tool parser, embedding projection, cancellation, and resource fit remain qualification gates.
Sources (accessed 2026-10-28): [license documentation](https://docs.liquid.ai/lfm/help/model-license.md) (updated 2026-10-15), [pinned Liquid license](https://huggingface.co/LiquidAI/LFM3-1.7B-GGUF/raw/b3f2ae5d73cd3c91bf40e8ba752c648aa05dffdf/LICENSE), and embedding card above.

## Lazy ownership and admission

Use `cold → loading → ready → idle → unloading → cold`, plus failed state with a reason, cooldown, and bounded retry.
No startup loading, probing, or downloading. Only preprovisioned verified paths; no runtime `-hf`. Disable download, UI, and peer-to-peer features.
Coalesce loads by exact model/runtime/device identity. Each waiter can cancel independently; cancel underlying work only when no waiters or leases remain.
Reference-count leases; never evict active work. Last release starts a proposed **5-minute idle TTL**, with earlier pressure eviction.
Serialize unload against acquisition; acquisition waits for unloading rather than racing process teardown.
Allow **one local load at a time** and reserve peak loading memory, resident weights, context, and native buffers before launch.
Admit one generation request (router shares that slot) and one embedding request only if their combined footprint fits.
Proposed queue capacity: **4**, deadline-aware; cancellation removes queued work and overload returns busy backpressure.
Use OS memory controls where available: a Go heap limit does not constrain native memory.
The current R1 source requires an explicitly preprovisioned, reviewed containment
launcher. That launcher is not supplied or qualified. Its contract denies all new
outbound connections, including loopback, permits only owned listener replies,
and caps/reserves process-tree memory with no swap. Flags/evidence strings do not
prove containment, and no model/runtime/launcher provisioning is authorized here.
For owned-process cancellation, close the HTTP request, allow a qualified grace period, then terminate only that owned process if cancellation fails; record failure.
For external/LAN servers, Familiar owns only client request leases. It cannot promise server cancellation, unloading, isolation, or capacity.

## Deterministic routes before decision models

| Route | Behavior |
| --- | --- |
| `lookup` | Explicit lookup; no model. |
| `sourced_answer` | Verified retrieval before generation; blocked until the R2 boundary is ready. |
| `proposal_only` | Read-only organization proposal; no execution authority. |
| `clarify` / `refuse` | Ask for missing scope or decline unsupported/unauthorized work. |

Routing features are bounded intent, requested output, evidence availability, residency, profile capability, remaining budgets, and grants.
Policy filters routes and destinations **before** any model sees choices; a model never authorizes access or disclosure.
The optional local router gets at most **1024 input tokens and 2 seconds total**, including admission and loading, and counts as one call against the four-call budget.
Invalid output, low confidence, timeout, or a cold miss falls back deterministically: no repeated rerouting or LAN escalation.
Never load a router to choose what to load. The baseline chooses the generator; skip routing if it would thrash or evict that generator.
Before P2 measurements, declare cold/warm comparisons on the same F4 corpus/hardware and total resource bounds.
Proposed benefit gate: **at least one additional judged task with no regressions**, or **at least 10% task-p95 improvement with no quality regression**; latency/memory ceilings still apply.
Calibrate confidence on separate examples, not the evaluation examples used to claim benefit.

## Offline and separately authorized LAN transport

Default: strict outbound denial, owned loopback only, no LAN DNS resolution or contact. With no suitable local resource, fail closed as blocked.
LAN enablement requires a separately approved **F1 amendment** identifying exact destinations and disclosure categories.
Authorize user prompts, history, filenames, previews, verified excerpts, and embedding inputs separately. A read grant is not a send grant.
No discovery, cloud fallback, remote filesystem authority, tool authority, or write authority.
Endpoint configuration fixes scheme, host/IP, port, base path, adapter, TLS CA/pin, auth-secret reference, and allowed profiles. Never place credentials in URLs.
Capability/health checks are task-triggered and coalesced: at most **two small requests within 3 seconds total**, bounded cached results, and no corpus data in probes.
Disable redirects and environment proxies. Resolve only configured names, validate every returned address, then dial the validated address while preserving TLS SNI.
Revalidate reconnects; reject unapproved DNS changes, IPv4-mapped bypasses, link-local/metadata/multicast destinations, and model-supplied URLs.
Require TLS and authentication for LAN; initially disallow plain HTTP.
Owned local servers bind explicitly to `127.0.0.1`, using per-instance auth where supported; this is not protection against same-UID processes.
LAN servers bind a chosen interface with firewall, TLS, and auth, not a copied vendor `0.0.0.0` example.
The operator must confirm no upstream relay or retention; the client cannot prove that assertion.
Log direct local artifact hashes separately from remote asserted provenance: remote model identity is an assertion, not cryptographic proof.

## Narrow ports, bounded streams, and tool dispatch

Version requests with task/request/attempt IDs, role, immutable profile, destination, policy/grant snapshot, deadline, and token/byte budgets.
Models cannot supply destination URLs or credentials.
Keep ports separate: `Generate` accepts messages and allowed read-only tool schemas; `Decide` accepts fixed choices; `Embed` accepts profiled text.
Events are `started`, `text_delta`, `tool_call_delta`, `usage`, and exactly one terminal `completed | failed | canceled`, with sequence numbers and actual profile/provenance.
Bound headers, frames, aggregate bytes, parser depth, and parsing time.
Before local harness dispatch, assemble the complete tool call and validate its name, schema, size, call ID, source handle, and remaining budget.
Qualify the LFM3 XML-like template against runtime-normalized tool calls; never execute arbitrary generated text.
Use a small static tool set; deferred-tool machinery is unnecessary.

## Provenance, cancellation, and replay

R4 records immutable ordered route/profile/provenance, queue/load/timing/token/cancellation, tool, and evidence events.
Use private bounded payload references, not raw diagnostic dumps, secrets, or exposed reasoning.
Stable task/request IDs remain separate from attempt IDs. Never automatically replay partial output or ambiguously accepted generation.
Allow at most one pre-send retry within the existing budget; do not assume server deduplication. The harness dispatches each tool-call ID once.
Record uncertain remote cancellation explicitly rather than claiming that server work stopped.
Replay never probes endpoints, calls models/tools, or writes effects. Regeneration is a new task.
R6 alone owns the agent loop; R2 precedes sourced answers. The inference layer gains no broker authority.

## Hardware profiles and acceptance gates

Known context is Linux amd64 with roughly 31 GiB RAM and a historical 4 GiB GPU; current GPU state is unknown. No scan or fit promise is implied.
Weight file bytes are not working-set estimates.

| Proposed profile | Enablement boundary |
| --- | --- |
| `device-cpu-text` | First local target: LFM3 Q4_K_M, 8K maximum context; prove GPU offload is disabled. |
| `device-gpu-text` | Measured offload only; disabled until separately qualified. |
| `lan-text` | Separate server qualification and client-footprint measurement, plus LAN authorization. |

The designer read F4's 13-task criteria; preserve them rather than relaxing them for a selected model:
- At least **90% overall** judged end-to-end success and **80% in each positive category** (with two tasks per category, effectively both).
- **100% negative-control refusals**; zero invented citations, unsupported claims, or effects. Organization remains proposal-only.
- Task **p95 ≤120 seconds**; process-tree peak **RAM ≤8 GiB**; incremental per-process **VRAM ≤2 GiB**.
- Also fit actually available capacity with **512 MiB system/VRAM headroom**. Missing reliable telemetry blocks qualification.
- CPU mode must prove offload disabled, not assume it.
- Retained F1 limits: **1 generation and 1 embedding subject to combined fit; 8K total context; 1024 output tokens/call; 4 calls/task; 2 minutes/call; 10 minutes/task; recursion depth 1; fanout 1**. Stricter role-specific limits still apply.

These are ceilings, not fit claims. Actual end-to-end task success requires R6; R1 documents alone cannot establish it.
Local authority: [F4_GATES.md](F4_GATES.md), especially future R1/R6/P2 criteria.

## Plan deltas and next implementable slice

These are deltas to the canonical plan, not a duplicate backlog:
- **F1:** offline/LAN and privacy amendment with per-destination disclosure grants.
- **R1:** pinned LFM3 generation/runtime and Liquid-first embedding profiles;
  lifecycle, template/tool, tokenizer/CLS-or-projection, cancellation and fit qualification.
- **R3:** embedding provenance, derived namespaces, and narrowly scoped ingestion after the read boundary is ready.
- **P2:** d1 against deterministic routing, including cold residency, quality, latency, and memory.

Current R1 task source has immutable profiles, a bounded generation-first HTTP
adapter and owned lazy lifecycle. Earlier build/vet/F3/F4 passed, but **no actual
inference/lifecycle/model/containment acceptance exists**. No runtime is enabled.
Next concrete slice: **bounded Liquid dense embedding-role integration and real
conformance once approved preprovisioned artifacts/runtime/containment and pinned
reference outputs/tolerances exist**. No speculative adapter or Python dependency.
Keep qualification explicit and opt-in; preserve doctor defaults. No runtime startup checks, downloads, corpus access, LAN, or agent loop in that slice.
Stage acceptance in this order:
1. Local cold load, coalescing, generation, tool parsing, cancellation, idle eviction, and offline containment.
2. Liquid dense CLS/1024d/L2/prefix/512-token reference conformance; ColBERT
   projection/mask/MaxSim only if justified, Gemma conformance only as explicit fallback.
3. Reviewed published RootOnly dependency/privacy acceptance and R2 revalidation;
   keep public R2 fail-closed and its prototype unmerged until then.
4. Sourced R4/R6 flows against F4 end-to-end criteria.
5. P2 decision-model benefit gate.
6. Separate LAN transport/privacy acceptance, never implicit fallback.

Recommended product defaults: text-only first, defer multimodal; CPU/local-only by default; LAN opt-in per destination; strict reviewed LFM3 generation preference; deterministic router unless measured benefit justifies d1 or an explicit older-small-router exception; no automatic upgrades.

## Persistence scope and remaining gates

Original persistence was source-only, based on the designer brief dated 2026-10-28.
This authorized public-preparation revision consolidates the same-date Liquid-first
research and current safety-disabled R2 status; it does not invent new qualification.
No builds/tests, inference, downloads, endpoint probes, installations, main
integration or publication occurred in this preparation stage. Model-license
acceptance is separate from Familiar source's explicitly selected Apache-2.0 license.
Remaining gates are license acceptance, local artifact/runtime verification, adapter and embedding conformance, cancellation/fit/telemetry qualification, FFF/R2 adoption, R4/R6 integration, P2 benefit, and separately authorized LAN acceptance.
