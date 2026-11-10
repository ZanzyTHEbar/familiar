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
Those guarantees describe **doctor**, not `task`. The task CLI has a separate
read-only agent loop and JSONL journal/replay; it has no embedding ingestion,
SQLite, Goja, arbitrary shell or workspace mutation capability.

## License

Familiar source is licensed under the [Apache License 2.0](LICENSE), as selected
by the project owner. Dependencies, native libraries and model artifacts retain
their own licenses. This does not relicense them or accept model terms; no model
weights, runtime binaries or third-party source are bundled here.
