# Familiar

Familiar is intended to be a local-first assistant for working with your files
and ongoing tasks: grounded answers, local models, and explicit permission before
accessing data or taking action. This version offers an **experimental, reactive
task CLI**, not a persistent or proactive assistant. Answers can be incorrect or
incomplete; inspect the cited source and any proposed file content yourself.

## What this version provides

- **Read-only diagnostics:** `familiar doctor` reports platform, build and native
  library linkage information. It does not search your files or run a model.
- **Synthetic retrieval tooling:** a small invented corpus, a Python verifier and
  a native search measurement helper. See the [fixture guide](fixtures/README.md).
- **Interactive tasks:** `familiar task` talks to an explicitly selected local
  model, searches and reads bounded workspace excerpts, and keeps a session
  journal. Optional new-file proposals require separate interactive review.
- **Generation APIs:** structured local conversations and a foreground runtime
  owner used by the task CLI, plus a separate legacy generation API requiring an
  external containment launcher. See the [inference API guide](inference/README.md).

There is no headless assistant service, background monitoring, durable assistant
memory system, automatic file organization or general shell tool. Session history
is not a persistent cognitive runtime. Embedding and decision profiles/interfaces
are present, but their adapters are not implemented. No models or runtimes are
downloaded automatically.

The scoped-evidence API (R2) requires a separately built [Linux scoped-worker
bundle](#linux-scoped-search), Landlock ABI 5 or newer, and an explicit
workspace/source grant. Missing or unsupported enforcement fails closed, without
an in-process fallback. Task search remains a separate bounded literal search,
not this API or semantic search. Compiling alone does not qualify containment.

## Build

See also [Linux scoped search](#linux-scoped-search) for the separate worker bundle.

The current target is **Linux amd64**. You need:

- A Go toolchain compatible with this module's Go 1.25.0 directive and its pinned
  dependencies, plus a C/C++ toolchain for CGO and `pkg-config`.
- Installed compatible **FFF bridge and FAISS** headers and shared libraries,
  including `fffbridge.pc`, `libfffbridge.so`, `libfaiss_c.so` and `libfaiss.so`.
  Familiar does not build or install them.
- The [virtual-vectorfs](https://github.com/ZanzyTHEbar/virtual-vectorfs) module
  pinned in [go.mod](go.mod) to `v0.0.0-20261028155634-88108866c3ca`, plus its
  transitive Go dependencies in the local module cache for an offline build.
- For task operations, Linux `openat2` support and a compatible local model/runtime
  that supports structured tool calls. The task CLI has no weaker filesystem
  fallback. Model weights and inference runtimes are separate from the search
  libraries above.

The Go module name is `familiar.local/familiar`; build from a source checkout,
not with `go install` using the repository URL. The following Bash recipe uses
normal pinned module resolution, with downloads disabled and no workspace/source
replacement. Set `NATIVE` to the compatible installed native prefix and
`BUILD_ROOT` to an existing absolute writable directory outside the source tree.
Run from Familiar's root:

```bash
set -euo pipefail
: "${NATIVE:?set the installed native prefix}"
: "${BUILD_ROOT:?set an external build parent directory}"
umask 077
BUILD_DIR=$(mktemp -d "$BUILD_ROOT/familiar.XXXXXX")
export GOWORK=off GOPROXY=off GOTOOLCHAIN=local CGO_ENABLED=1 GOMAXPROCS=2
export PKG_CONFIG_PATH="$NATIVE/lib/pkgconfig"
export CGO_CFLAGS="-I$NATIVE/include"
export CGO_LDFLAGS="-L$NATIVE/lib -Wl,-rpath,$NATIVE/lib"
export LD_LIBRARY_PATH="$NATIVE/lib"
go build -mod=readonly -p=1 -tags=vectors -o "$BUILD_DIR/familiar" ./cmd/familiar
```

An uncached dependency causes this offline recipe to fail; it does not fetch a
replacement. Install native dependencies and prepare the Go dependency cache
separately. A successful link is not proof of native ABI compatibility or model
fit. Missing shared libraries can prevent the executable from reaching `main`,
producing a loader error rather than a JSON diagnostic. Keep the selected native
library prefix available when running the binary.

## Run diagnostics

In the same shell, after building:

```bash
"$BUILD_DIR/familiar" doctor --help
"$BUILD_DIR/familiar" doctor --json --offline
```

**The diagnostic command normally exits 1**, because generation and embedding
endpoints always report `not-configured` in this version. This is not an
installation detector for models elsewhere on your machine. Omit `--json` for
human-readable output. `--offline=false` does not enable networking.

Exit codes are 0 for ready checks or help, 1 for blocked readiness, and 2 for
usage/output errors. JSON uses `schema_version: 1`, an aggregate `status`, build
provenance and `checks`. Consumers should key on check IDs and statuses rather
than human-readable reason text. Statuses are `ready`, `blocked` and
`not-configured`; any non-ready check makes the aggregate blocked. The same
report is available through [`doctor.Inspect()`](doctor/doctor.go).

Doctor does not inspect your working directory, home or model directories,
create indexes/state, start watchers, contact endpoints or perform inference.
Its native-linkage check only reports that the CGO/vector binary reached `main`;
it does not establish search operation, cancellation, recovery or hardware/model
readiness. These diagnostic properties are not a sandbox guarantee for the
separate retrieval helper or inference library.

## Run an experimental task

**Read the privacy warning before granting access.** `--read-workspace` grants
reads of visible regular files, **including files ignored by Git**. Hidden paths,
symlinks and submount paths are excluded by the task tools, but this is not a
whole-process sandbox or outbound-network containment. The selected model runtime
can have its own filesystem, network and retention behavior. Use only a workspace
whose visible contents you are willing to disclose to it.

The journal stores **sensitive plaintext**: workspace/model selections, prompts,
source excerpts, model output/reasoning, tool calls/results and file proposals.
It is retained until you remove it. The final state directory must be owned by you
with mode 0700; session files must be owned 0600 regular files. Prefer an explicit
absolute `--state` outside the workspace. Otherwise state defaults under
`XDG_STATE_HOME/familiar/sessions`, or your home directory's `.local/state`.
These file modes do not protect against another process running as your user.

### Connect to a running local model

Select an existing compatible server and model yourself. Set `WORKSPACE` to an
absolute nonsymlink workspace directory, `STATE_DIR` to an absolute private state
directory outside it, and `MODEL_ID` to the server's model ID. For a selected server
on loopback port 8080, for example:

```bash
"$BUILD_DIR/familiar" task --workspace "$WORKSPACE" --read-workspace \
  --state "$STATE_DIR" --endpoint http://127.0.0.1:8080/v1 --model "$MODEL_ID"
```

For llama.cpp, add `--llama`; `--thinking runtime|on|off` controls the requested
reasoning mode and `--cache-prompt` requests prefix reuse, not a measured cache hit.
The client uses no environment proxies or redirects. External-endpoint mode does
not send API credentials and does not start, stop or sandbox that server.

### Let the CLI own a local runtime

Alternatively, select trusted, already installed artifacts with `LLAMA_SERVER`
(absolute executable path) and `MODEL_GGUF` (absolute weights path). Do not combine
this mode with `--endpoint`:

```bash
"$BUILD_DIR/familiar" task --workspace "$WORKSPACE" --read-workspace \
  --state "$STATE_DIR" --runtime "$LLAMA_SERVER" --weights "$MODEL_GGUF" \
  --model "$MODEL_ID"
```

`--model` is still required as a selection label; the owned server uses a fresh
private alias. This mode hashes the selected artifacts, starts one foreground
CPU-only llama.cpp child and keeps it resident until this CLI exits. It needs
compatible runtime arguments, `/proc` access and shared libraries available in
the executable's directory. There is no download, daemon, idle eviction, hard
memory/CPU quota or enforced outbound-network isolation. `--runtime-context`
(default 16384), `--runtime-threads` (default 2) and `--load-timeout` (default 2m)
are configuration/admission limits, not containment guarantees.

### Tasks, history and optional file creation

Enter one task per input line; `/quit` exits. Ctrl-C cancels the active task and
exits. Output remains provisional until `[assistant completed]`, which is not a
guarantee that the answer is correct. Each task has a ten-minute deadline, at most
twelve model turns and eight tool calls. Search results and excerpts may be
incomplete or clipped; they are current reads, not immutable citation snapshots.

Add `--session` with the printed session ID to the same launch command to continue
with a fresh read grant and matching workspace/model/runtime/tool options. Only
completed historical tasks are reused; incomplete tasks are not automatically
retried. Whole oldest completed tasks may be omitted after context preflight
overflow, without deleting their journal records. To display history without
model contact, workspace access or tool dispatch:

```bash
"$BUILD_DIR/familiar" task --state "$STATE_DIR" --session "$SESSION_ID" --replay
```

Replay is not action recovery or a retry. Torn/invalid journals are refused and
preserved; inspect them and start a new session rather than assuming completion.
Task exits are 0 for help/success and 2 for errors or cancellation.

Writes are off by default. Adding **`--create-files`** only enables proposals for
new UTF-8 text files of at most 4096 bytes. Each exact path and complete content
must be reviewed and approved using a fresh one-use challenge with stdin/stdout
on the same interactive terminal; piped input cannot approve. The tool does not
overwrite, delete or create parent directories. After exclusive creation starts,
cancellation does not guarantee rollback. An uncertain result can leave an empty,
partial or moved file: inspect it manually and do not blindly retry. Race,
interruption and recovery guarantees are not established for this experimental
creation path. Do not use it for critical data.

## Linux scoped search

`retrieval.Open(ctx, grant, records, stateParent)` uses a fresh worker per grant.
Linux amd64, Landlock ABI **5+**, `openat2` and `close_range` are required. Other
platforms report unsupported, not macOS support. No in-process fallback exists.

### Build and library integration

Use the existing offline vector-enabled CGO environment: installed FFF/FAISS
headers/libraries/pkg-config, compatible Go, a C compiler with static libc
development files, and `ldd`. Set native flags as for the build above, then:

```sh
export GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local GOWORK=off
export GOMAXPROCS=2 GOMEMLIMIT=512MiB CGO_ENABLED=1
python3 scripts/build-scoped-runtime.py /absolute/new/scoped-runtime
export FAMILIAR_SCOPED_RUNTIME=/absolute/new/scoped-runtime
```

Library consumers set that variable before Open; no consumer `main` registration
is needed. Install **launcher, worker, lib files and runtime.files together** as a
trusted, integrity-preserved bundle. It must be disjoint from the workspace.
Bundle paths must not contain `:`, `;` or `$`: the loader interprets path-list
separators and expands `$ORIGIN`, `$LIB`, `$PLATFORM` and their braced forms.
Packaging, Open and direct launcher invocation reject all dollar characters,
not just known tokens. Spaces remain supported. Source and destination bundle
paths must obey this rule; shell quoting does not prevent loader expansion.
The caller-owned canonical state parent must be 0700 and disjoint from the
workspace. Each adapter creates a random 0700 state directory and removes only
its owned state on Close. Memory indexing still writes private durable evidence.
Records-only adapters use an empty root in their own state, not caller paths.

The bundle is target-host/ABI-specific, not a universal Linux executable. The
packager copies the trusted worker's exact `ldd` dependency closure into `lib`,
records source paths/hashes in `provenance.json`, and lists each runtime file.
The `@interpreter` entry grants only the worker's ELF interpreter. Other entries
must be regular bundled files, resolved beneath the bundle without symlinks.
No broad system/user directories, HOME, `/etc`, `/proc` or `/sys` are granted.
Repackage/revalidate after runtime changes; missing/denied resources fail closed.
Treat the manifest as trusted security configuration; never add user data to
make a failing library work. Nothing is fetched or installed automatically.

Packaging defaults remain private (0700 directories/executables, 0600 data).
A root-built private bundle is therefore **not** automatically usable by other
users. For shared installation, an administrator can copy the reviewed runtime
to a **new**, canonical destination below existing root-owned, non-group/world-
writable parents. Every ancestor must be searchable by intended users; do not
make the private build directory public or recursively change its permissions.
For example, as the administrator (not an application startup step):

```sh
SRC=/absolute/private/build/scoped-runtime
DEST=/opt/familiar/scoped-runtime-VERSION
# DEST must not exist; its parent must already satisfy the trust/access rules.
mkdir -m 0755 -- "$DEST"
chown root:root -- "$DEST"
install -d -o root -g root -m 0755 -- "$DEST/lib"
install -o root -g root -m 0755 -- "$SRC/launcher" "$SRC/worker" "$DEST/"
install -o root -g root -m 0644 -- "$SRC/runtime.files" "$DEST/"
install -o root -g root -m 0644 -- "$SRC"/lib/* "$DEST/lib/"
```

Publish/select that versioned destination only after verifying root ownership,
0755 directories/executables, 0644 manifest/libraries, and the installed hashes;
no runtime file or directory may be writable by untrusted users. Do not overwrite
a live bundle in place. Consumers set `FAMILIAR_SCOPED_RUNTIME` to the installed
destination; each still supplies its own private caller-owned 0700 state parent.
Keep required provenance, notices, source and relink materials with the approved
distribution; review generated build metadata for private paths before sharing.
This procedure changes access modes, not host compatibility or legal clearance.

### Enforcement and lifetime

The static C launcher is single-threaded: it accepts pipe IPC on descriptors 0–2
and root/state directory descriptors 3–4, verifies their identities, handles all
filesystem rights through ABI 5 (including truncation and device ioctl), grants
workspace reads, ordinary state file/directory operations, and file-specific
runtime reads/execution. It sets `no_new_privs`, successfully restricts itself,
then closes **all** descriptors above 2 before exec. Go runtime threads, cgo/native
initialization and FFF start only afterward and inherit enforcement. There is no
unsafe post-start thread retrofit. Loader/Git/HOME configuration is not inherited.

Before opening FFF the actual worker must observe EACCES opening an existing
outside sentinel. Unexpected success or any different error aborts startup.
The existing ordinary FFF provider runs **inside** this kernel boundary. Its
ambient Git/config/ignore attempts do not grant permission to read unrelated
contents. This does not adopt the experimental RootOnly provider or change the
search dependency. Native errors never trigger an ambient retry.

The parent retains immutable candidate/source/hash bindings and root identity
checks. Only search/expansion and copied caller records cross private JSON pipes.
IPC is serialized; cancellation during an operation kills the worker and renders
that adapter unusable. Close closes IPC and waits/reaps native shutdown, killing
and reaping a stuck worker after five seconds before deleting owned state.
Parent death closes IPC; abrupt termination can leave private state for explicit
recovery, rather than sweeping unknown directories.

### Grant policy and limitations

- The entire workspace, **including hidden/ignored content**, is authorized for
  reads. Current FFF filtering is retained, not RootOnly's ignores-off policy.
  In the non-Git fixture, hidden and `.ignore`-excluded files produce no hits.
  Ignores are not security exclusions. Do not grant roots containing secrets
  that must remain unreadable; there are no subtree deny rules.
- Workspace content writes, truncation, creation, deletion and renames are denied.
  Landlock does not restrict every metadata operation (such as stat/chmod/utime).
  This is not isolation of all metadata access/mutation or untrusted native code.
- Escaping symlinks to unrelated content are denied at open time, not merely
  filtered from results. Symlinks to granted content/runtime files are not
  globally forbidden; FFF's no-follow scanner setting is a separate behavior.
- Reachable mounts and hardlinked content in the workspace are included. Origin
  and other aliases are not excluded. Runtime file exceptions are file/inode
  specific. Host mutation and hostile same-user processes are not isolated.
- No network, general process/IPC, host, model or whole-application isolation is
  claimed. The parent is outside the policy; kernel, installed runtime, manifest,
  launcher and consumer are trusted. Private state is a separate read/write grant.

### Validation and redistribution

With the installed bundle environment, run the unchanged R2 journey and the
small real-worker boundary check against fresh paths:

```sh
R2_SCRATCH=/absolute/new/r2 go test -mod=readonly -count=1 -p=1 \
  -tags=vectors ./retrieval -run '^TestScopedEvidenceAcceptance$' -v -timeout=60s
python3 scripts/check-scoped-boundary.py "$FAMILIAR_SCOPED_RUNTIME" /absolute/new/boundary
```

The boundary check uses the real engine: permitted reads/hash-bound expansion,
an outside symlink denied before engine startup, and startup refusal when the
required denial check instead succeeds on a granted file. It deliberately passes
an unrelated descriptor to the launcher. Inspect a `strace -f` receipt for
restrict-self, close-range and exec ordering before Go/native thread creation.
Finite checks do not qualify every deployment or a malicious engine.

The launcher is original code using Linux UAPI definitions (GPL-2.0 WITH
Linux-syscall-note). The packager records installed binary provenance, not
upstream authenticity or license clearance. Before redistributing bundles,
retain FFF/FAISS MIT notices/source identity, Go dependency notices and each
system library's applicable notices/source offer (including glibc LGPL and GCC
runtime exceptions). Static glibc linkage needs appropriate LGPL relinking
materials/source availability: retain launcher source, toolchain/build command
and relinkable objects. Local hashes are not reproducible-build attestation.

## License

Familiar source is licensed under the [Apache License 2.0](LICENSE).
The synthetic corpus text and manifest are dedicated to the public domain under
[CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/).
Dependencies, native libraries and model artifacts retain their own licenses.
The pinned search dependency is MIT-licensed; retain its
[copyright and permission notice](https://github.com/ZanzyTHEbar/virtual-vectorfs/blob/88108866c3cac44d7a66e1a9532d35b1f2271ecf/LICENSE)
when redistributing it or bundled binaries.
No model weights, runtime binaries or third-party source are bundled here;
using Familiar does not accept their terms or relicense them.
