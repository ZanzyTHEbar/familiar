# Familiar — product checkpoint

Saved 2026-10-28. This is a status snapshot, not a replacement delivery plan.
Public source target: [ZanzyTHEbar/familiar](https://github.com/ZanzyTHEbar/familiar).
Familiar source uses **Apache-2.0**, explicitly selected by the project owner;
dependencies/models keep their own terms. Serial main integration includes R1,
the Liquid-first design and independent F3 review (original `2ffab22` preserved).
R2 is deliberately preserved on its separate experimental branch, NOT main.

## Product at 10,000 feet

Familiar is a separate local companion around the Go search library, intended to
grow into a policy-bounded assistant. Today's usable CLI is a **read-only offline
doctor**, plus synthetic retrieval fixtures/baseline tooling—not an agent loop.
There is no user-workspace mutation authority, broker, SQLite event log, embedding
ingestion, personal corpus access or automatic model/runtime provisioning.
The goal is a safe local companion that can produce grounded, cited answers
before any agentic action authority. Agent loop, write tools, durable event/replay
system, TUI and LAN service remain unimplemented/not enabled.

| Slice | Honest current status |
| --- | --- |
| F3 doctor | Prior local build/native-linkage and no-filesystem/network acceptance passed; schema 1, model endpoints deliberately not-configured. Independent F3 review exists. |
| F4 fixtures/baseline | Synthetic 10-file/13-task corpus and native retrieval baseline passed prior gates; not end-to-end agent or model quality/fit. |
| R1 inference | Integrated generation-first source; immutable reviewed candidates, bounded local HTTP and owned lazy lifecycle. Prior source build/vet/F3/F4 passed; final merged checks recorded below. **No real runtime/model/containment acceptance.** |
| R2 exact evidence | **Safety-disabled, experimental; separate `r2-scoped-evidence` branch only.** Open refuses before argument/filesystem/state/native work. Branch build/vet with vectors passed against published vvfs 88108866. Original unchanged functional acceptance genuinely FAILED at denied Open (exit 1). Not accepted, not merged. |
| Later slices | No R3 ingestion/ranking overhaul, R4 replay, R6 agent loop, P2 decision-routing evaluation, LAN activation or write capabilities are enabled. |

Generation preference is reviewed/pinned **LFM3-1.7B Q4_K_M**. Embedding preference
is **Liquid-family first**: LFM2.5-Embedding-350M Q8_0 (1024d CLS/L2/cosine).
LFM2.5-ColBERT-350M is a conditional per-token/MaxSim retrieval/reranking candidate,
not a single-vector encoder or stock cross-encoder. Gemma is explicit fallback
only if suitable Liquid integration cannot qualify. None is locally model-qualified;
published hashes are metadata, and file size is not a RAM/VRAM fit measurement.
See [R1 source boundary](inference/README.md) and consolidated [DESIGN.md](DESIGN.md).
Liquid dense embedding is selected but **no embedding adapter was implemented**.
There is no automatic update/fallback.

## What is deliberately blocked

- No preprovisioned compatible model/runtime/containment launcher or accepted
  license/telemetry evidence was supplied. The launcher is a required, unimplemented
  deployment contract, not security proved by a flag or nonempty evidence string.
- R2's retained native path is not qualified against ambient Git/config/ignore
  reads outside grants. RootOnly integration/privacy acceptance must precede re-enablement. Historical
  functional tests did not establish this confinement and are not rewritten to
  make safety denial look like functional success.
- The committed search module pin remains `fe092dd03d398570b5d93cf0c030f573bc90257b`.
  Main does not need R2's API upgrade. Upstream search published
  `88108866c3cac44d7a66e1a9532d35b1f2271ecf`; only the separate R2 branch pins
  its authenticated normal module pseudo-version `v0.0.0-20261028155634-88108866c3ca`.
  Canonical production remains Ambient: FFF adoption and privacy qualification
  are still pending. No dirty/unpublished dependency or native rebuild/adoption.
- Publication requires reachable-history private-content review and actual merged
  build/acceptance. Historical nonsecret host paths are accepted publication
  metadata; normal history is preserved, never silently rewritten.

## Final integration checks

Final merged main validation is pending at this saved checkpoint: full Go build,
vet and existing tests with vectors, plus F3 doctor and F4 fixture/baseline gates,
serial CPU 2 / Go jobs 1 with isolated scratch and approved installed native
libraries. Results will be recorded before push. No model provisioning, runtime
launch, endpoint probing, native build or LAN is part of these checks. R2's known
failed functional journey is not part of main and is not hidden by a skip.

## Single best next product milestone

**A safe local read-only cited answer using a real model.** Qualify the approved
preprovisioned runtime/model/containment, then scoped evidence/privacy and exact
citations before wiring the bounded read-only answer flow. Liquid dense embedding
integration/reference conformance is still work, not delivered by model selection.
Keep R2 denied until its separate functional/privacy gates honestly pass. ColBERT
requires judged ranking failures and exact tokenizer/mask/projection evidence.
No extra features or broader action authority are implied by source publication.

Authority remains the upstream
[delivery plan](https://github.com/ZanzyTHEbar/virtual-vectorfs/blob/main/docs/LOCAL_AGENT_PLAN.md)
and [policy](https://github.com/ZanzyTHEbar/virtual-vectorfs/blob/main/docs/LOCAL_AGENT_POLICY.md).
No plan transfer or completion marking is implied. The local module identity
`familiar.local/familiar` is intentional; a public repository is not yet a published
`go install` contract. Model license approval is distinct from the source license.
