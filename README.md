# Familiar

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

## Experimental retrieval API

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

## License

Familiar source uses [Apache-2.0](LICENSE), explicitly selected by the project
owner. Dependencies, native libraries and model artifacts retain their own terms;
this does not relicense them or accept model terms. No models/native binaries
are bundled.
