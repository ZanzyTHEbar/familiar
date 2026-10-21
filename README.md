# Familiar

F3: a minimal, read-only Linux doctor, not an agent runtime.

Local module identity: **`familiar.local/familiar`**; executable: **`familiar`**.
This reversible local-only name is **not a public hosting identity**.
Canonical repository: `../../familiar` from this Workmux worker; target `main`.

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

The authoritative [delivery plan](../../../go-pkgs/virtual-vectorfs/docs/LOCAL_AGENT_PLAN.md)
and [policy](../../../go-pkgs/virtual-vectorfs/docs/LOCAL_AGENT_POLICY.md) remain
in the separate search repository. No formal plan transfer occurred.
Search is pinned to committed revision
`fe092dd03d398570b5d93cf0c030f573bc90257b`
(`v0.0.0-20261011132509-fe092dd03d39`). Its unpublished dirty F2 work is not
included. **R2 must wait for coordinator-approved F2 integration.**
No source or docs in search are changed by Familiar. Independent review and
parent-boundary gate re-verification are required before integration. Next
planned slice is F4, then separately gated R1; this repository is not a second
F/R/W/P backlog.
