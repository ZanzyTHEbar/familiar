# Familiar

Familiar is intended to be a local-first assistant for working with your files
and ongoing tasks: grounded answers, local models, and explicit permission before
accessing data or taking action. It is early software, **not yet a working
assistant**.

## What works today

- **Read-only diagnostics:** `familiar doctor` reports platform, build and native
  library linkage information. It does not search your files or run a model.
- **Synthetic retrieval tooling:** a small invented corpus, a Python verifier and
  a native search measurement helper. See the [fixture guide](fixtures/README.md).
- **Generation library source:** an opt-in Go API for managing a local generation
  process and parsing bounded responses. It is not connected to the CLI, and
  requires separately provisioned artifacts and containment. See the
  [inference API guide](inference/README.md).

There is no chat/task command, background monitoring, persistent assistant memory,
automatic file organization or user-workspace assistant search in this version.
Embedding and decision interfaces and model profiles are present, but their
adapters are not implemented. No models or runtimes are downloaded automatically.

## Build

The current target is **Linux amd64**. You need:

- A Go toolchain compatible with this module's Go 1.25.0 directive and its pinned
  dependencies, plus a C/C++ toolchain for CGO and `pkg-config`.
- Installed compatible **FFF bridge and FAISS** headers and shared libraries,
  including `fffbridge.pc`, `libfffbridge.so`, `libfaiss_c.so` and `libfaiss.so`.
  Familiar does not build or install them.
- The exact [virtual-vectorfs](https://github.com/ZanzyTHEbar/virtual-vectorfs)
  source revision `fe092dd03d398570b5d93cf0c030f573bc90257b`, and cached transitive
  Go dependencies for an offline build.

The Go module name is `familiar.local/familiar`; build from a source checkout,
not with `go install` using the repository URL. The following Bash recipe uses a
separate build directory and an explicit export of the pinned search dependency.
It does not modify either source checkout or use the dependency's working files.
Set `SOURCE` to an absolute virtual-vectorfs checkout path containing that commit,
`NATIVE` to its compatible installed native prefix, and `BUILD_ROOT` to an existing
absolute writable directory outside both source trees. Run from Familiar's root:

```bash
set -euo pipefail
: "${SOURCE:?set the absolute search checkout path}"
: "${NATIVE:?set the installed native prefix}"
: "${BUILD_ROOT:?set an external build parent directory}"
umask 077
BUILD_DIR=$(mktemp -d "$BUILD_ROOT/familiar.XXXXXX")
mkdir "$BUILD_DIR/search"
git -C "$SOURCE" archive fe092dd03d398570b5d93cf0c030f573bc90257b |
  tar -x -C "$BUILD_DIR/search"
export GOWORK="$BUILD_DIR/go.work"
export GOPROXY=off GOTOOLCHAIN=local CGO_ENABLED=1 GOMAXPROCS=2
export PKG_CONFIG_PATH="$NATIVE/lib/pkgconfig"
export CGO_CFLAGS="-I$NATIVE/include"
export CGO_LDFLAGS="-L$NATIVE/lib -Wl,-rpath,$NATIVE/lib"
export LD_LIBRARY_PATH="$NATIVE/lib"
go work init "$PWD"
go work edit -replace="github.com/ZanzyTHEbar/virtual-vectorfs@v0.0.0-20261011132509-fe092dd03d39=$BUILD_DIR/search"
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

## License

Familiar source is licensed under the [Apache License 2.0](LICENSE).
The synthetic corpus text and manifest are dedicated to the public domain under
[CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/).
Dependencies, native libraries and model artifacts retain their own licenses.
No model weights, runtime binaries or third-party source are bundled here;
using Familiar does not accept their terms or relicense them.
