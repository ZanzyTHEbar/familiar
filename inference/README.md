# R1 opt-in source boundary (not qualified)

This generation-first package is not imported by the CLI or doctor. Constructors,
profile access and `Status` are filesystem/network-free. No model/runtime is
enabled, installed or fetched. `doctor.Inspect()` and schema 1 are unchanged.
Canonical plan/policy remain in virtual-vectorfs; [R1_GATES.md](../R1_GATES.md)
records this slice's gates, not R1 completion.

## Narrow API and ownership

- `GenerationCandidate()` and Liquid-first `EmbeddingCandidate()` return immutable
  value profiles; `LateInteractionCandidate()` is conditional, and
  `EmbeddingFallbackCandidate()` preserves Gemma only as fallback. These are
  **published metadata**, not local verification or working embedding/ranking APIs.
  LFM3+ is a reviewed preference; there are no moving releases or auto-updates.
- `Generator`, `Decider` and `Embedder` are separate ports. Only `Owned.Generate`
  is implemented. d1's future decision endpoint is `/v1/systemone`; no router
  evaluation or model routing is implemented. The caller retains deterministic
  routing, grants, four-call/ten-minute task limits, depth/fanout one and dispatch.
- `NewOwned(LocalConfig)` accepts only explicitly provisioned absolute paths,
  expected local runtime/launcher hashes, an operator-chosen nonzero loopback
  port, license approval and reviewed containment/fit evidence. There is one
  logical CPU generation owner per Go process, no multi-model scheduler or
  generic endpoint/configuration layer. Close it before constructing another.
- `Acquire(ctx)` coalesces loads and returns a residency-only lease; it exposes
  no endpoint, process or tool authority. Each waiter can cancel independently;
  the last cancels loading. Always release leases. Generation admission permits
  one active and four waiting requests; cancellation removes waiting calls.
  Failed loads have a 30-second cooldown; a later explicit call may retry. There
  is no automatic generation replay, including after ambiguous transport failure.
- State is `cold/loading/ready/idle/unloading/failed` (then `closed` on shutdown).
  Only the last lease starts the fixed five-minute idle TTL. `EvictIdle` permits
  earlier pressure eviction without touching active work. Acquisition waits for
  reaping during unload. `Close(ctx)` cancels calls/loads and waits for leases and
  owned teardown; a timeout is not proof of stopped resources.

Every load checks `/proc/meminfo` for measured peak reservation plus 512 MiB
headroom, then hashes only the configured model/runtime/launcher. Model size and
SHA-256 must match the candidate. This is not a directory scan. Runtime hashes
must be supplied from the operator's reviewed build record: the v0.8.0 source
revision alone does not identify a binary or its dynamically loaded libraries.
An available-memory snapshot is not an OS reservation; containment must reserve
and cap the process tree before native loading. Missing telemetry blocks loading.

The fixed runtime arguments disable GPU layers, KV/op offload, automatic fit,
context shifting, UI, built-in agent/MCP proxy and prompt caches; set 8192 context,
1024 output maximum, one native slot and two CPU threads; and enable the embedded
Jinja template with thinking disabled. Per-instance API auth uses fresh random
bytes, passed only in a sanitized child environment, not command-line arguments,
URLs or logs. No inherited runtime flags, proxies, credentials or search paths.
The HTTP client has no proxies/redirects/DNS or general destination support and
dials only the configured numeric `127.0.0.1` address. Plain HTTP exists **only for
this owned loopback process**. No LAN/external-server path or fallback exists.

### Explicit containment prerequisite, not a supplied sandbox

No containment launcher is provisioned by this change. To avoid treating
`--offline` or the Go heap as a security/memory boundary, loading requires an
explicitly pinned, preprovisioned launcher with this fixed argv protocol:

```text
launcher --memory-max-bytes <PeakBytes> --deny-outbound -- <runtime> <fixed-server-args...>
```

The reviewed launcher must apply process-tree memory accounting/reservation and
hard limits (including native allocations and no swap), deny all new outbound
connections (including loopback), allow only replies on the owned loopback listener,
drop any setup privilege, retain the supplied API key, and **exec** the runtime
without detached children. It must keep the runtime reachable only on owned
loopback. Evidence must establish actual RAM/VRAM headroom, local immutable
artifact/build/native identities and CPU-only behavior. No implementation or
capability is inferred from a filename or a nonempty evidence reference. These
are operator-provisioned prerequisites and actual qualification gates; arbitrary
`/bin/sh` wrappers or unreviewed scripts do not satisfy them. No privilege changes,
namespace/firewall/cgroup setup or license acceptance are performed here.

Artifact leaf symlinks, special files, group/other-writable files and executables
larger than 512 MiB are refused. Hash verification precedes path-based execution/loading: deployment must keep
these files and trusted path ancestry immutable between verification and use.
This is not protection from a same-UID adversary replacing artifacts or inspecting
the child environment. Linux parent-thread death signaling is defensive, not a
qualified crash/descendant containment guarantee. Do not infer such guarantees
from source alone. No launcher was found or searched for beyond supplied paths
(none were supplied); qualification remains blocked.

## HTTP and tool contract

Requests carry version 1, bounded task/request/attempt IDs, policy snapshot
reference, exact candidate fingerprint and deadline. Events add the local
fingerprint (configured runtime/launcher hashes, provisioning record, fixed options)
and monotonic sequence. Verification is explicitly `configured-not-verified`
until a lease is acquired, then `local-artifact-bytes-verified`; neither proves
model behavior or build/native provenance. There is one attempted terminal event even on failure or
cancellation. A sink that rejects delivery cannot be promised a delivered event;
sinks must honor their context and return promptly. No unbounded event goroutines.

Queue, load, template/token preflight and inference share a two-minute call
deadline. After closing canceled/failed generation HTTP, the owner conservatively
terminates that owned process rather than assume native cancellation succeeded;
SIGTERM gets two seconds before SIGKILL. Reaping must finish before a new load;
OS teardown may outlast the request deadline and is not claimed as model latency.
No other actual HTTP user exists. Passive readiness leases do not authorize work.

Local readiness is load-triggered, authenticated `/v1/models` (4 KiB responses,
one-second attempts at most four/second under the bounded load deadline), not
default doctor probing, discovery or LAN capability checks. A port collision
cannot cause adoption of an unauthenticated existing server. The selected model
alias must match the profile. This does not prove backend/template compatibility.

`/apply-template` receives the same messages/tools/template options as generation;
`/tokenize` counts the rendered prompt with special tokens before sending chat
completion. Reject input + requested output + one-token headroom above 8192.
Qualification must confirm the template/tokenizer path matches actual prompt
usage. No truncation, context shifting or assumed characters-per-token estimate.
Only text system/user/assistant messages ending in a user message are accepted;
tool-result conversations/agent iteration are deliberately deferred.

Bounds: 32 messages, 32 KiB message content, 64 KiB encoded request, 48 KiB rendered
prompt, 256 KiB JSON preflight response, 16 KiB response headers, 64 KiB SSE line,
1 MiB aggregate stream, 4096 frames, JSON depth 16, 256 KiB text output, one tool
call with 8 KiB arguments. JSON duplicate keys are refused. Completion requires
matching response/model IDs, valid finish reason, bounded usage and `[DONE]`;
short reads, length stops, malformed frames or nonempty normalized
`reasoning_content` fields are failures; raw text is not a reasoning detector.
Stream text/tool fragments are untrusted partial observations, not final answers.

Allowed tool schemas are deliberately limited to at most four named read-only
tools, each a closed object with at most eight required string fields. No generic
JSON Schema engine or dependency. The assembled single normalized call must have
a bounded identifier, an allowed name, exact required keys and string values;
null/extra/duplicate/nested arguments fail. Raw LFM XML-like text is never parsed
as a tool call. Only a completed event can carry the validated proposal; it still
needs source-handle/grant/budget validation and call-ID deduplication by the future
harness. No arbitrary generated text, schema or tool can execute an effect here.

## Reviewed candidate evidence — accessed 2026-10-28

**Corrected user preference:** suitable official **Liquid-family embedding AND
retrieval/reranking models first**, not only the LFM3 generation. The earlier
absence of a documented LFM3 embedder is NOT a reason to select Gemma. Prefer the
newest suitable reviewed release, never autoupdate. Fresh official
[library](https://docs.liquid.ai/lfm/models/complete-library.md),
[Nanos docs](https://docs.liquid.ai/lfm/models/liquid-nanos.md),
[full collection](https://huggingface.co/api/collections/LiquidAI/liquid-nanos-68b98d898414dd94d4d5f99a)
and author-scoped API searches for
[Embedding](https://huggingface.co/api/models?author=LiquidAI&search=Embedding&limit=50),
[ColBERT](https://huggingface.co/api/models?author=LiquidAI&search=ColBERT&limit=50),
[Rerank](https://huggingface.co/api/models?author=LiquidAI&search=Rerank&limit=50),
[Retriev](https://huggingface.co/api/models?author=LiquidAI&search=Retriev&limit=50)
and [RAG](https://huggingface.co/api/models?author=LiquidAI&search=RAG&limit=50)
identify the LFM2.5 trained retrieval pair below, plus older LFM2-ColBERT. This is
bounded public discovery as of 2026-10-28, not a claim about private/future models.
Q8_0 is a reviewed candidate quantization for conformance, not a measured fit claim.

| Candidate | Reviewed revision / metadata |
| --- | --- |
| LFM3-1.7B Q4_K_M | `b3f2ae5d73cd3c91bf40e8ba752c648aa05dffdf`, 1099627808 bytes, SHA-256 `3793e1e6e2266919d6e015287221b41100175a98eddaba92e111aed3b05c93c4` |
| **Primary dense:** LiquidAI/LFM2.5-Embedding-350M-GGUF, `LFM2.5-Embedding-350M-Q8_0.gguf` | `a80de9c5b941d429104f0038292a0ef5a860e486`, 379216640 bytes, SHA-256 `6ec5f8e8750dbc8a0e40c431fd1b7b07a13688136b2244c5a1364b54d9032599` |
| **Conditional late interaction:** LiquidAI/LFM2.5-ColBERT-350M-GGUF, `LFM2.5-ColBERT-350M-Q8_0.gguf` | `bc240003aba07253e261a8aaf0d2c9683318a967`, 378097248 bytes, SHA-256 `5765e4987f1e63e6c52572198abd65fe9616416edf523c0fa95a077a3e501d47` |
| **Fallback only:** EmbeddingGemma2 Q8_0 | `58faf90940886df2d62be555c8765d00dac362c5`, 309855456 bytes, SHA-256 `2188ac1deca4b77dffefd603c2776a9d76d9d74ec01841392982ebb840b09135` |
| llama.cpp v0.8.0 source | `8642541b255bfb5dc39cd351a178be09013a55d6`, published 2026-10-28T12:47:51Z; no local binary identity |

Fresh pinned artifact metadata:
[dense](https://huggingface.co/api/models/LiquidAI/LFM2.5-Embedding-350M-GGUF/revision/a80de9c5b941d429104f0038292a0ef5a860e486?blobs=true),
[ColBERT](https://huggingface.co/api/models/LiquidAI/LFM2.5-ColBERT-350M-GGUF/revision/bc240003aba07253e261a8aaf0d2c9683318a967?blobs=true).
Both are public/non-gated; metadata declares non-causal LFM2 architecture and
128000 context. **Use the trained short retrieval limits, not that metadata cap.**
No conversion-source/build record is listed in either GGUF manifest. Pinned
current base configs below document semantics, NOT verified GGUF conversion origin;
the profile explicitly marks conversion source unattested. Resolve that provenance
and verify the actual embedded tokenizer/weights during qualification.

### Primary: trained dense bi-encoder, not generator pooling

[Official docs](https://docs.liquid.ai/lfm/models/lfm25-embedding-350m),
[pinned GGUF card](https://huggingface.co/LiquidAI/LFM2.5-Embedding-350M-GGUF/raw/a80de9c5b941d429104f0038292a0ef5a860e486/README.md)
and [base card](https://huggingface.co/LiquidAI/LFM2.5-Embedding-350M/raw/f35ae2c91d687658dbf1f2b449382f0b019b9808/README.md)
describe a bidirectional trained **dense bi-encoder**, one 1024d CLS vector per
query/document, **no separate learned projection**. L2-normalize, then dot product
is cosine: higher means more similar, nominal range [-1,1], NOT a probability,
confidence, evidence/authority score or a scale interchangeable across profiles.
Do not use arbitrary generator hidden states or mean pooling instead of CLS.

Reference `f35ae2c91d687658dbf1f2b449382f0b019b9808`:
[pooling](https://huggingface.co/LiquidAI/LFM2.5-Embedding-350M/raw/f35ae2c91d687658dbf1f2b449382f0b019b9808/1_Pooling/config.json),
[modules](https://huggingface.co/LiquidAI/LFM2.5-Embedding-350M/raw/f35ae2c91d687658dbf1f2b449382f0b019b9808/modules.json),
[prefixes](https://huggingface.co/LiquidAI/LFM2.5-Embedding-350M/raw/f35ae2c91d687658dbf1f2b449382f0b019b9808/config_sentence_transformers.json),
[length](https://huggingface.co/LiquidAI/LFM2.5-Embedding-350M/raw/f35ae2c91d687658dbf1f2b449382f0b019b9808/sentence_bert_config.json),
[tokenizer](https://huggingface.co/LiquidAI/LFM2.5-Embedding-350M/raw/f35ae2c91d687658dbf1f2b449382f0b019b9808/tokenizer_config.json).
Use raw `query: ` / `document: ` prefixes, preserve case, BOS id 1, no added EOS,
reference pad id 0; vocabulary 65536. No chat template or Gemma title formatting.
Cap **512 total tokens including prefix/specials**, refuse excess or use a separately
versioned source-range chunking policy, never silently truncate. Eleven documented
languages: en/es/de/fr/it/pt/ar/sv/no/ja/ko; suitability still needs local evidence.

### Conditional: ColBERT late-interaction retrieval/reranking

[Official docs](https://docs.liquid.ai/lfm/models/lfm25-colbert-350m),
[pinned GGUF card](https://huggingface.co/LiquidAI/LFM2.5-ColBERT-350M-GGUF/raw/bc240003aba07253e261a8aaf0d2c9683318a967/README.md),
[base card](https://huggingface.co/LiquidAI/LFM2.5-ColBERT-350M/raw/9772bdf797255d8693b83e84aa98e9b2d36dd0be/README.md)
describe **128d per-token vectors**, no pooling, a required learned 1024→128 linear
projection (no bias, identity activation), per-token L2, then
`sum_i max_j dot(query_token_i, document_token_j)` (MaxSim). Higher is a relative
match score for the same query/profile, not cosine of two document vectors or a
calibrated probability; query length/expansion/masking affect its scale.
This can rerank first-stage candidates without a persistent multi-vector index,
but is NOT a joint query/document cross-encoder or a single-vector Embedder.

Reference `9772bdf797255d8693b83e84aa98e9b2d36dd0be`:
[projection](https://huggingface.co/LiquidAI/LFM2.5-ColBERT-350M/raw/9772bdf797255d8693b83e84aa98e9b2d36dd0be/1_Dense/config.json),
[preprocessing](https://huggingface.co/LiquidAI/LFM2.5-ColBERT-350M/raw/9772bdf797255d8693b83e84aa98e9b2d36dd0be/config_sentence_transformers.json),
[tokenizer](https://huggingface.co/LiquidAI/LFM2.5-ColBERT-350M/raw/9772bdf797255d8693b83e84aa98e9b2d36dd0be/tokenizer_config.json),
[model config](https://huggingface.co/LiquidAI/LFM2.5-ColBERT-350M/raw/9772bdf797255d8693b83e84aa98e9b2d36dd0be/config.json),
[length config](https://huggingface.co/LiquidAI/LFM2.5-ColBERT-350M/raw/9772bdf797255d8693b83e84aa98e9b2d36dd0be/sentence_bert_config.json).
Raw prefixes are `[Q] ` / `[D] ` (added IDs 64400/64401, vocabulary 64402), BOS 1,
no added EOS; query cap/pad-to length **32**. Remove document punctuation-token
vectors using the pinned config's `skiplist_words` tokenized with that tokenizer.
**Unresolved:** reference `attend_to_expansion_tokens=false` and `do_query_expansion=true`;
tokenizer pad is EOS id 7 while model config says pad id 0. Documents are advertised
as 512, but sentence config says 511. Candidate conservatively caps total input at
**511**, including prefix/specials, and query at 32; no automatic truncation.
Exact GGUF pad IDs, masking, special-token accounting and retained query vectors
must match the pinned reference before any enablement.

The [official GGUF demo](https://huggingface.co/LiquidAI/LFM2.5-ColBERT-350M-GGUF/raw/bc240003aba07253e261a8aaf0d2c9683318a967/colbert-rerank.py)
uses Python/tokenizer downloads, pads queries, sends token IDs, omits an explicit
HTTP attention-mask payload and truncates documents. It is **read-only evidence**,
not adopted code or proof of reference-equivalent masks. No Python runtime,
dependency, model, new scoring API or index format is introduced here.

Enable no reranking/late-interaction route until concrete judged failures justify
a separate bounded conformance/benefit decision. Current F4 retrieval passes and
demonstrates no such need. No vendor quality/latency result is Familiar evidence.

### Runtime integration decision and remaining provenance gates

The official cards support **llama.cpp GGUF for both Liquid candidates**. No
different backend is required by the documentation. The reviewed llama.cpp
[converter](https://github.com/ggml-org/llama.cpp/blob/8642541b255bfb5dc39cd351a178be09013a55d6/conversion/lfm2.py)
explicitly lists them, sets non-causal attention and imports pooling/dense modules;
[LFM2 model tensors](https://github.com/ggml-org/llama.cpp/blob/8642541b255bfb5dc39cd351a178be09013a55d6/src/models/lfm2.cpp),
[non-causal convolution](https://github.com/ggml-org/llama.cpp/blob/8642541b255bfb5dc39cd351a178be09013a55d6/src/models/lfm-base.cpp)
and [pooling/projection](https://github.com/ggml-org/llama.cpp/blob/8642541b255bfb5dc39cd351a178be09013a55d6/src/llama-graph.cpp)
provide source support evidence, NOT a qualified local build/artifact combination.
The [server contract](https://github.com/ggml-org/llama.cpp/blob/8642541b255bfb5dc39cd351a178be09013a55d6/tools/server/README.md)
distinguishes these paths:

- Dense: a future separate embedding-role launch (`--embeddings`, CLS pooling,
  bounded context/batch, L2) plus `/v1/embeddings` for raw prefixed text.
- ColBERT: no pooling, `/embedding` with bounded token IDs → per-token tensors,
  verified projection, client-side preprocessing/masking/L2/MaxSim. `/v1/embeddings`
  requires pooled outputs; `/v1/rerank` expects a rank-pooling reranker and must not
  be assumed to implement ColBERT MaxSim. The current `Owned` is generation-only.

First resolve exact tokenizer/pooling/CLS or projection/expansion-mask conformance,
conversion provenance, and context refusal on a preprovisioned local CPU build.
If the pinned backend cannot reproduce a required semantic, name the exact API/
mask/projection gap and obtain a narrow runtime/backend decision; do not silently
swap backends, add Python, or declare Liquid unsupported just because it is older
than LFM3. No embedding, reranking or multi-model lifecycle adapter is added here.

Generic LFM2.5 fill-mask encoders are not automatically trained retrieval models.
[LFM2-1.2B-RAG at cdbd46d305e6935c544a3492feaafcc6b611970e](https://huggingface.co/LiquidAI/LFM2-1.2B-RAG/raw/cdbd46d305e6935c544a3492feaafcc6b611970e/README.md)
is a **generator answering from supplied context**, not the retriever or a scoring
model. It does not replace the unchanged LFM3 generation candidate.

### Explicit Gemma fallback and isolation

Keep Gemma only if no suitable compatible Liquid integration can be qualified;
missing local provisioning blocks all candidates, not just Liquid. No automatic
fallback is implemented. Prior 2026-10-28 fallback evidence remains
[artifact metadata](https://huggingface.co/api/models/ggml-org/embeddinggemma-2-GGUF/revision/58faf90940886df2d62be555c8765d00dac362c5?blobs=true),
[Google card](https://huggingface.co/google/embeddinggemma-2/raw/main/README.md),
[conversion source](https://huggingface.co/ggml-org/embeddinggemma-2-GGUF/raw/58faf90940886df2d62be555c8765d00dac362c5/.src_sha)
and [runtime support](https://api.github.com/repos/ggml-org/llama.cpp/pulls/30054).
Fallback semantics: 768d, mean pooling, required learned 512→768 projection, L2,
`task: search result | query: ` / `title: {title-or-none} | text: `, source
`914f7f89142e33e77833254d9c9b90c3cef7303b`, max 8192 despite larger GGUF metadata.
A candidate fingerprint alone is not a qualified vector namespace. Liquid 1024d,
Gemma 768d and ColBERT multi-vectors are incompatible; use a new derived namespace
for any model/runtime/tokenizer/template/chunking change, never mix silently.
Per-vector source-range hashes remain separate; no R3/index migration is authorized.

Generation pin evidence is unchanged:
[artifact metadata](https://huggingface.co/api/models/LiquidAI/LFM3-1.7B-GGUF/revision/b3f2ae5d73cd3c91bf40e8ba752c648aa05dffdf?blobs=true),
[model docs](https://docs.liquid.ai/lfm/models/lfm3-1.7b),
[runtime release](https://api.github.com/repos/ggml-org/llama.cpp/releases/tags/v0.8.0).

**No local artifacts were hashed or executed in this wave.** The earlier PATH
check found no `llama-server`; this correction performed no new runtime inventory
or probes, and no explicit model/runtime/launcher paths were supplied. No personal
directory inventory was performed. Liquid artifacts declare
[LFM Open License v1.0](https://huggingface.co/LiquidAI/LFM2.5-Embedding-350M-GGUF/raw/a80de9c5b941d429104f0038292a0ef5a860e486/LICENSE),
including a $10M annual revenue commercial threshold and redistribution/attribution
conditions; full review remains the operator's responsibility. Gemma's card
declares Apache-2.0. No terms were accepted. Hashes/sizes above are **published
metadata, not downloaded/local artifact verification**. Current hardware fit,
CPU/backend compatibility and reference conformance remain unknown. Actual R1
qualification and independent review are still pending.
