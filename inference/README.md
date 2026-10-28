# Local generation API

Experimental API; model and containment qualification are not established.

## API and ownership

- `GenerationCandidate()` and Liquid-first `EmbeddingCandidate()` return immutable
  value profiles; `LateInteractionCandidate()` is conditional, and
  `EmbeddingFallbackCandidate()` preserves Gemma only as fallback. These are
  **published metadata**, not local verification or working embedding/ranking APIs.
  LFM3+ is a reviewed preference; there are no moving releases or auto-updates.
- `Generator`, `Decider` and `Embedder` are separate interfaces.
  Only `Owned.Generate` is implemented; no model router is supplied.
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
from source alone. A separately provisioned, qualified containment launcher is required.

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
needs source-handle/grant/budget validation and call-ID deduplication by the
caller. No arbitrary generated text, schema or tool can execute an effect here.

## Artifact licenses

The [Apache-2.0 source license](../LICENSE) does not cover separately obtained
model/runtime artifacts. The listed Liquid artifacts declare the
[LFM Open License v1.0](https://huggingface.co/LiquidAI/LFM3-1.7B-GGUF/raw/b3f2ae5d73cd3c91bf40e8ba752c648aa05dffdf/LICENSE),
including a $10M annual-revenue threshold for a separate commercial license and
attribution/redistribution conditions. The EmbeddingGemma2 candidate declares
Apache-2.0. Review the full terms for the exact artifacts before provisioning or
distributing them; this summary does not replace those licenses. No terms are
accepted on your behalf, and no model weights or runtime binaries are bundled.
