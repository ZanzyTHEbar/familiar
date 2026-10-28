# Familiar

F3: a minimal, read-only Linux doctor. F4: synthetic evaluation fixtures and an
offline retrieval baseline. R2 is a safety-disabled experimental prototype, not
a working retrieval API. No agent runtime.

Local module identity: **`familiar.local/familiar`**; executable: **`familiar`**.
This reversible local-only name is **not a public hosting identity**.
The module name is intentionally local; publishing source is not a `go install`
contract. Public repository/licensing preparation does not qualify models or R2.

```sh
# Requires the pinned search dependency and installed compatible FFF/FAISS.
go build -mod=readonly -p=2 -tags=vectors -o familiar ./cmd/familiar
./familiar doctor --json --offline
```

See [GATES.md](GATES.md) for the exact offline dependency export, native linking
environment, bounded build/vet commands, and real CLI acceptance scenario. A Go
import does not install native libraries. Missing dynamically linked libraries
can stop the process **before main**, with a loader error instead of JSON.

## API and readiness

`doctor.Inspect() doctor.Report` is the reusable API; the CLI only presents it.
JSON has integer `schema_version: 1`, `scope: "F3"`, aggregate `status`, build
provenance and `checks` with stable IDs, status and human-readable reasons.
Consumers should key on IDs/statuses, not reason text. Statuses are `ready`,
`blocked`, and `not-configured`; any non-ready check makes the aggregate blocked.
Exit codes: **0** all reported capabilities ready; **1** blocked readiness;
**2** usage/output error. `doctor --help` exits 0.

F3 always reports generation/embedding endpoints `not-configured` and exits 1.
It has no model configuration/probe interface, even if models exist elsewhere.
This is the permitted F3 boundary, **not R1 model/hardware qualification**.
No endpoint contact, inference, download, or personal/model directory inventory
occurs. `--offline` is an explicit assertion of the unconditional behavior;
`--offline=false` does not enable networking.

Native readiness means only a CGO/vector binary reached main with its shared
libraries linked. It does not establish library version/ABI compatibility,
native cancellation, model fit, latency, or filesystem recovery. The real search
import is exercised only with a pre-canceled context; the pinned implementation
returns before it opens anything. No engine/index/watch is initialized.

Doctor does not inspect its working directory, HOME, environment configuration,
or workspace/state roots and has no root flags or defaults. It creates no
directories, state, indexes, or watchers. Empty scratch roots are acceptance
fixtures, not adopted workspace identities or approved read/write grants.
There is no agent loop, embedding ingestion, SQLite/replay, Goja, broker, or
mutation capability.

## Authority and dependency boundary

The authoritative [delivery plan](https://github.com/ZanzyTHEbar/virtual-vectorfs/blob/main/docs/LOCAL_AGENT_PLAN.md)
and [policy](https://github.com/ZanzyTHEbar/virtual-vectorfs/blob/main/docs/LOCAL_AGENT_POLICY.md) remain
in the separate search repository. No formal plan transfer occurred.
This experimental branch pins published search commit
`88108866c3cac44d7a66e1a9532d35b1f2271ecf`
(`v0.0.0-20261028155634-88108866c3ca`) for API compilation, through normal Go
module retrieval with no local/workspace replacement. The earlier unpublished
`1e2e882...` pin is historical evidence only. Canonical production remains Ambient;
this pin does not authorize FFF adoption or establish RootOnly/privacy acceptance.
Main keeps its prior dependency and does not include R2. The prototype stays disabled.
No source or docs in search are changed by Familiar. Independent review and
parent-boundary gate re-verification are required before integration.
F4's corpus, judgments and measurement utility are described in
[fixtures/README.md](fixtures/README.md), with predeclared criteria and exact
commands in [F4_GATES.md](F4_GATES.md). Actual end-to-end task success waits for
R6; separately gated R1 remains unqualified. This repository is not a second
F/R/W/P backlog.

## R2 scoped retrieval and exact evidence

**SAFETY-DISABLED — experimental, not integration-ready.** Public
`retrieval.Open` returns `(nil, retrieval.ErrUnavailable)` immediately, before
argument inspection, filesystem access/state creation or `search.Open`. There is
no bypass flag or environment fallback. No working retrieval feature is advertised.
The uncalled internal implementation is retained only as an unmerged prototype.

The older native FFF path starts even for record-only adapters, discovers ancestor
Git repositories and launches repo-wide status scans;
its walker enables ambient Git/global ignores and external configuration reads.
Separate indexes constrain returned evidence, **not all native reads** to the
root/source grants. This unresolved read-scope defect is not remedied by the
root checks, os.OpenRoot, follow_symlinks=false, or a sandbox disclaimer. Before
integration, scoped consumers need an upstream option disabling Git discovery/
status and external ignore/config reads, verified by real nested-repository
acceptance, followed by an explicit downstream repin and revalidation.

The **proposed, disabled contract** would create a fresh scope-specific index,
never open a shared index, and copy the caller's source grants. A grant
has at most one whole live `Root`, requiring source `retrieval.FileSource`, plus
optional caller-record sources. Record-only grants omit Root. StateParent must
be an existing caller-owned `0700` directory disjoint from the live root; paths
must be absolute, canonical and symlink-free. State stays outside input roots.

`adapter.Search(ctx, text, limit)` returns up to 20 private-bound candidates.
`candidate.Lead()` returns an index-time preview bounded to 512 UTF-8 bytes with
PreviewTruncated indicating omitted bytes, not a current citation or match
location. `adapter.Expand(ctx, candidate, field, byteOffset, byteLimit)` verifies
the candidate's original source/hash and returns an exact half-open UTF-8 range,
bounded to 4096 bytes. It refuses foreign/unbound candidates and changed, missing
or stale content. Fields are `text`, `title`, `context` (live files: `text` only).
Offsets are independently selected source-field bytes, never grep display ranges.

Live citations identify their root, source, hash, field, byte range and origin.
Caller-record citations explicitly say `indexed_record`: an immutable snapshot,
not a live-file claim. A record's Path is only a label and never grants file access.
There are no supplied hashes, vectors, update methods or exposed upstream engine.
Proposed admission limits: 1000 records, 8 MiB total record inputs, 8 MiB per live file,
4096 query bytes; callers must bound corpus size and operation deadlines. Reopen
with a new grant to change authority/snapshots. **Always Close the adapter** to
drain upstream watchers/workers. Fresh in-memory indexes are not persisted; private
native/evidence state is retained under StateParent for caller management.

This is application-level scoping, not an OS sandbox or secret detector. Callers
must choose a root excluding secrets/ungranted subtrees and keep ancestry/mounts
stable during calls. Root identity changes between calls are refused, but no
same-UID adversarial root-replacement race guarantee is claimed. “Current” means
verified during Expand; no citation promises freshness after subsequent edits.
See [R2_GATES.md](R2_GATES.md) and [R2_EVIDENCE.md](R2_EVIDENCE.md) for the real
historical native acceptance flow and unchanged functional assertions. Those
functional assertions FAILED at guarded Open on 2026-10-28 (exit 1); branch build
and vet passed with vectors, offline published modules and canonical installed
native libraries. Assertions have not been weakened or rewritten. This branch is
preserved separately, not accepted or merged into main. Denial regression qualification and
any approved private native proof remain pending. Independent review found the unresolved native
ambient-read blocker above; passing existing checks does not establish scoped-read
acceptance or integration readiness. No R1/R3/R4 work is included.

## License

Familiar source uses [Apache-2.0](LICENSE), explicitly selected by the project
owner. Dependencies, native libraries and model artifacts retain their own terms;
this does not relicense them or accept model terms. No models/native binaries
are bundled.
