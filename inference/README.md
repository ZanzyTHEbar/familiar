# Local generation API

Package `familiar.local/familiar/inference` provides experimental local generation
APIs. `Local` and `TaskRuntime` support the reactive task CLI. The older `Owned`
generator is separate, is not used by that CLI, and still requires a separately
provisioned containment launcher. None is used by `familiar doctor`.
Runtime compatibility, model behavior, resource fit and containment are not
established by the source alone. Do not treat it as a ready-to-use assistant or
an included sandbox.

See [build requirements](../README.md#build) for the repository's native Go build.
No model, runtime executable or containment launcher is included or fetched.

## Profiles and implemented operations

[`Profile`](profile.go) exposes `Identity()` as a value copy and `Fingerprint()`
as a SHA-256 of the encoded identity. Candidate hashes describe expected bytes,
not evidence that an installed model works.

| Function | Candidate | Implementation |
| --- | --- | --- |
| `GenerationCandidate()` | LiquidAI LFM3-1.7B Q4_K_M | `Owned.Generate` uses this fixed profile. |
| `EmbeddingCandidate()` | LiquidAI LFM2.5-Embedding-350M Q8_0 | Metadata only; no embedding adapter. |
| `LateInteractionCandidate()` | LiquidAI LFM2.5-ColBERT-350M Q8_0 | Metadata only; no retrieval/reranking adapter. |
| `EmbeddingFallbackCandidate()` | EmbeddingGemma2 Q8_0 | Metadata only; no automatic fallback. |

[`Generator`, `Decider` and `Embedder`](ports.go) are separate interfaces. Only
`Generator` has an implementation here; `Local` separately implements
`Conversation`. Profile presence does not enable a model, semantic index or
decision router. Tool dispatch belongs to the task caller, not inference.

## Structured task conversations

`NewLocal(EndpointConfig)` selects an already running loopback HTTP endpoint with
an explicit port, `/v1` base path and model ID. It accepts `127.0.0.1`, `localhost`
(mapped to IPv4 loopback) or `[::1]`; not LAN/cloud URLs, credentials in URLs,
proxies or redirects. It does not load models, discover servers, read API keys,
retry generation or control the external server's lifecycle/egress. Generic
OpenAI-compatible behavior is not a guarantee of tool-calling compatibility.

`Local.Turn(ctx, messages, tools, emit)` returns a completed `Message` containing
text, structured tool calls and optional reasoning. Stream deltas are provisional;
completion requires a supported finish reason and `[DONE]`. The caller must
validate tool names/arguments/authority before dispatch, retain call/result IDs
and enforce task limits. `Close()` closes idle client connections; canceling an
HTTP request is not proof that an external server stopped generation.

Configuration requires `Thinking` (`runtime`, `on` or `off`), `OutputTokens`
(1..4096) and `ContextBytes` (1024..131072). Thinking/cache overrides require
`Llama: true`. Generic endpoints use only the encoded request byte ceiling, not a
token estimate or guaranteed token window. For llama.cpp, preflight uses `/props`,
`/apply-template` and `/tokenize` to check the reported single-slot context;
optional `ContextTokens` can lower that ceiling. Output reserve plus one-token
headroom is included. `ContextLimitError` occurs before generation; the task caller
may remove whole historical tasks from its projection, but must not retry an
ambiguously accepted generation. `Cache` requests prefix reuse, not a proven hit.

`PrepareTaskRuntime(ctx, TaskRuntimeConfig)` hashes and holds explicitly selected
trusted executable/weights descriptors without starting a process. `Start` starts
one CPU-only llama.cpp child, verifies its accepted loopback sockets before HTTP,
adds per-instance authentication, then returns a configured `Local`. It requires
Linux `openat2`, `/proc`, nonsymlink artifact ancestry and nonempty regular files
not writable by group/others; the executable must have an execute bit. The runtime
gets a restricted environment and a library path limited to its executable's
directory. File hashes do not verify transitive libraries, behavior or same-user
interference.

This runtime is single-use and resident for the whole CLI launch, including idle
prompts. Defaults selected by the CLI are 16384 context tokens, two CPU threads
and a two-minute startup timeout. Its available-memory check is only a snapshot;
there is no supplied sandbox, external launcher, hard quota, daemon or idle
eviction. `Close()` signals the direct child, escalates after two seconds and
waits for reaping; it does not guarantee control of detached descendants. The
process's offline flag is not enforced outbound-network isolation. See
[task usage and privacy warnings](../README.md#run-an-experimental-task).

## Legacy generator: configure an owner

`NewOwned(LocalConfig)` validates configuration without accessing files, the
network or processes. Supply:

- `Runtime` and `Launcher`: `FilePin` values with explicit absolute paths and
  lowercase SHA-256 hashes of the installed executables.
- `ModelPath`: an absolute path matching the generation candidate's exact byte
  size and SHA-256 in `GenerationCandidate().Identity()`.
- `Port`: an explicitly selected nonzero loopback port.
- `LicenseApproved`: confirmation that applicable artifact terms were accepted.
- `Evidence`: a nonempty reference to deployment containment/resource review.
- `PeakBytes`: a measured peak memory requirement, at least the model file size
  and at most 8 GiB. File size alone is not a peak memory estimate.

The configuration fields express caller assertions, not verified permission or
proof of containment. Only one owner can exist per Go process; successfully close
it before creating another. `Status()` is an in-memory state/reason snapshot.

### Launcher requirements

The separately supplied executable must accept this argument protocol:

```text
launcher --memory-max-bytes <PeakBytes> --deny-outbound -- <runtime> <fixed-server-args...>
```

It must enforce process-tree memory accounting/reservation and a hard cap without
swap, deny all new outbound connections including loopback, permit replies on the
owned loopback listener, drop setup privileges, preserve the supplied API key,
and `exec` the runtime without detached children. Familiar supplies no launcher
implementation and does not configure namespaces, cgroups or firewalls. An
executable name, flag or nonempty `Evidence` string cannot establish these controls.

Each load reads `/proc/meminfo` and requires `PeakBytes` plus 512 MiB available
RAM, then hashes only the configured artifacts. This snapshot is not an OS memory
reservation. Artifact leaf symlinks, nonregular files, group/other-writable files
and runtime/launcher executables over 512 MiB are rejected. Keep the files and
their path ancestry trusted and immutable between verification and use: execution
is path-based, not immune to replacement races or same-UID interference.

The runtime is launched with fixed CPU-only arguments, 8192-token context, at most
1024 output tokens, one native slot and two CPU threads. Thinking, prompt caches,
context shifting, UI and built-in agent/MCP features are disabled. The child gets
a restricted environment with a fresh per-instance API key. Transport uses only
its configured numeric `127.0.0.1` address, without environment proxies, redirects
or a LAN/external-server option. The runtime must support the fixed arguments and
authenticated `/v1/models`, `/apply-template`, `/tokenize` and
`/v1/chat/completions` contracts in the source; generic compatibility is insufficient.

## Legacy generator: generate and consume events

Call `Generate(ctx, GenerateRequest, Emit)` on the owner. Populate the embedded
`Envelope` with `Version: 1`, nonempty task/request/attempt/policy references,
the generation profile fingerprint and a nonzero deadline. Set
`MaxOutputTokens` between 1 and 1024.

- Supply at most 32 UTF-8 text messages, with at most 32 KiB combined content.
  Roles are `user` and `assistant`, plus an optional first `system` message.
  The last message must be `user`; structured tool calls/results and reasoning
  fields are rejected here. Use `Conversation` for structured continuation.
- Up to four unique named tools may be described. Each has a closed argument
  object with at most eight required string fields; nested/optional fields and
  arbitrary JSON Schema are not supported.
- The encoded request is capped at 64 KiB. Template/token preflight rejects
  input plus requested output plus one-token headroom exceeding 8192 tokens.
  It does not silently truncate input.

Admission, load, preflight and generation share a two-minute call deadline,
shortened by the caller's deadline. One request runs while up to four wait;
additional requests return `ErrBusy`. Calls are not automatically replayed.

`Emit` is synchronous: honor its context and return promptly. Events have ordered
sequence numbers and request/profile references. Kinds are `started`,
`text_delta`, `tool_call_delta`, `usage`, and terminal `completed`, `failed` or
`canceled`. The API attempts one terminal event, but cannot guarantee delivery if
the sink fails or blocks. Text and tool fragments are provisional, untrusted data.
Only a successful completed event may carry an assembled, schema-checked tool
proposal. **Nothing dispatches that proposal here.** The caller must enforce real
grants, source validity, task-wide budgets and call-ID deduplication. A policy
reference in a request does not enforce a policy.

Responses have bounded headers, JSON depth, stream size and frame count. Malformed
or incomplete streams, duplicate JSON keys, length stops and nonempty normalized
`reasoning_content` are rejected. Raw generated text is not a reasoning detector
or trusted evidence. Local artifact verification does not verify answer quality.

## Legacy generator: residency and shutdown

`Acquire(ctx)` lazily loads and returns a residency-only lease. Loads coalesce;
waiters cancel independently, and the last waiter can cancel an unfinished load.
Always call `Lease.Release()`; release is idempotent and grants no tool authority.
Generation acquires and releases its own lease.

The last release starts a five-minute idle timeout. `EvictIdle()` may evict sooner
but not while leases are active. Failed loads have a 30-second retry cooldown;
there is no automatic retry within a call. Canceled or incomplete accepted
generation requests trigger owned-process teardown rather than assuming native
work stopped. Teardown sends SIGTERM, then SIGKILL after two seconds if needed,
and waits for reaping; it may outlast the request deadline.

Use `Close(ctx)` to cancel loading/calls and wait for leases and teardown. A close
timeout is **not proof that resources stopped**: release outstanding leases and
call it again to finish shutdown. Parent-death signaling is defensive, not a
guarantee about detached descendants or crash containment.

## Artifact licenses

The [Apache-2.0 source license](../LICENSE) does not cover separately obtained
model/runtime artifacts. The listed Liquid artifacts declare the
[LFM Open License v1.0](https://huggingface.co/LiquidAI/LFM3-1.7B-GGUF/raw/b3f2ae5d73cd3c91bf40e8ba752c648aa05dffdf/LICENSE),
including a $10M annual-revenue threshold for a separate commercial license and
attribution/redistribution conditions. The EmbeddingGemma2 candidate declares
Apache-2.0. Review the full terms for the exact artifacts before provisioning or
distributing them; this summary does not replace those licenses. No terms are
accepted on your behalf, and no model weights or runtime binaries are bundled.
