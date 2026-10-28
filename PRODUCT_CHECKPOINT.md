# Familiar — product checkpoint

Saved 2026-10-28. This is a status snapshot, not a replacement delivery plan.
Target public source repository: `ZanzyTHEbar/familiar`. Final serial integration
and public publication are authorized. Familiar source uses **Apache-2.0**,
explicitly selected by the project owner; dependencies/models keep their own terms.
This checkpoint is updated with actual integration/gate results before publication.

## Product at 10,000 feet

Familiar is a separate local companion around the Go search library, intended to
grow into a policy-bounded assistant. Today's usable CLI is a **read-only offline
doctor**, plus synthetic retrieval fixtures/baseline tooling—not an agent loop.
There is no user-workspace mutation authority, broker, SQLite event log, embedding
ingestion, personal corpus access or automatic model/runtime provisioning.

| Slice | Honest current status |
| --- | --- |
| F3 doctor | Prior local build/native-linkage and no-filesystem/network acceptance passed; schema 1, model endpoints deliberately not-configured. Independent F3 review exists. |
| F4 fixtures/baseline | Synthetic 10-file/13-task corpus and native retrieval baseline passed prior gates; not end-to-end agent or model quality/fit. |
| R1 inference | Unmerged generation-first source compiled/vetted with F3/F4; immutable reviewed candidates, bounded local HTTP and owned lazy lifecycle. **No real runtime/model/containment acceptance.** |
| R2 exact evidence | **Safety-disabled, experimental and unmerged.** Public Open refuses before argument/filesystem/state/native work. Retained expansion prototype needs a later reviewed published RootOnly dependency and privacy acceptance; its compilation is pending that stable pin. |
| Later slices | No R3 ingestion/ranking overhaul, R4 replay, R6 agent loop, P2 decision-routing evaluation, LAN activation or write capabilities are enabled. |

Generation preference is reviewed/pinned **LFM3-1.7B Q4_K_M**. Embedding preference
is **Liquid-family first**: LFM2.5-Embedding-350M Q8_0 (1024d CLS/L2/cosine).
LFM2.5-ColBERT-350M is a conditional per-token/MaxSim retrieval/reranking candidate,
not a single-vector encoder or stock cross-encoder. Gemma is explicit fallback
only if suitable Liquid integration cannot qualify. None is locally model-qualified;
published hashes are metadata, and file size is not a RAM/VRAM fit measurement.
See [R1 source boundary](inference/README.md) and the consolidated `DESIGN.md`
when task branches are serially integrated. There is no automatic update/fallback.

## What is deliberately blocked

- No preprovisioned compatible model/runtime/containment launcher or accepted
  license/telemetry evidence was supplied. The launcher is a required, unimplemented
  deployment contract, not security proved by a flag or nonempty evidence string.
- R2's old native path can read ambient Git/config/ignore state outside grants.
  RootOnly integration/privacy acceptance must precede re-enablement. Historical
  functional tests did not establish this confinement and are not rewritten to
  make safety denial look like functional success.
- The committed search module pin remains `fe092dd03d398570b5d93cf0c030f573bc90257b`.
  No dirty/unpublished dependency is adopted while the upstream owner integrates.
- Publication requires reachable-history private-content review and actual merged
  build/acceptance. Historical nonsecret host paths are accepted publication
  metadata; normal history is preserved, never silently rewritten.

## Next bounded work

Wait for the coordinator's exact-session serial-integration resume after upstream
work finishes. Reconcile the stable dependency and overlapping task README/gate
edits; retain R2 safety denial until its separate gates pass. Validate the merged
state before any public creation/push. No source history is rewritten here.
Separately, the next inference slice is a small Liquid dense embedding-role path
and real reference conformance on explicitly approved preprovisioned artifacts.
ColBERT needs judged ranking failures plus exact tokenizer/mask/projection proof.

Authority remains the upstream
[delivery plan](https://github.com/ZanzyTHEbar/virtual-vectorfs/blob/main/docs/LOCAL_AGENT_PLAN.md)
and [policy](https://github.com/ZanzyTHEbar/virtual-vectorfs/blob/main/docs/LOCAL_AGENT_POLICY.md).
No plan transfer or completion marking is implied. The local module identity
`familiar.local/familiar` is intentional; a public repository is not yet a published
`go install` contract. Model license approval is distinct from the source license.
