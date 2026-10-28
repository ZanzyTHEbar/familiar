#!/usr/bin/env bash
set -euo pipefail
: "${SCRATCH:?create a new empty owner-controlled scratch directory first}"
: "${NATIVE:?set the installed native prefix}"
test -z "$(ls -A "$SCRATCH")"
umask 077
mkdir -p "$SCRATCH"/{bin,cache,tmp,workspace,state,system-libs}
# Published vvfs 88108866 is materialized through normal Go module retrieval.
# No workspace replacement or dirty/local source fallback.
export GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local GOMAXPROCS=2 GOMEMLIMIT=256MiB
export CGO_ENABLED=1 GOCACHE="$SCRATCH/cache" TMPDIR="$SCRATCH/tmp"
export OMP_NUM_THREADS=2 OPENBLAS_NUM_THREADS=2 RAYON_NUM_THREADS=2
export PKG_CONFIG_PATH="$NATIVE/lib/pkgconfig"
export CGO_CFLAGS="-I$NATIVE/include"
export CGO_LDFLAGS="-L$NATIVE/lib -Wl,-rpath,$NATIVE/lib"
export LD_LIBRARY_PATH="$NATIVE/lib"
export GOWORK=off
pkg-config --modversion fffbridge
timeout 180s go build -mod=readonly -p=1 -tags=vectors -o "$SCRATCH/bin/familiar" ./cmd/familiar
timeout 60s go build -mod=readonly -p=1 -tags=vectors ./...
timeout 60s go vet -mod=readonly -p=1 -tags=vectors ./...
test -z "$(gofmt -l cmd doctor retrieval)"
git diff --check
python3 - "$SCRATCH" <<'PY'
import hashlib, json, os, pathlib, subprocess, sys
s = pathlib.Path(sys.argv[1])
def manifest():
    result = []
    for root in (s / 'workspace', s / 'state'):
        for p in [root, *sorted(root.rglob('*'))]:
            st = p.lstat()
            result.append((str(p.relative_to(s)), st.st_mode, st.st_size,
                           st.st_mtime_ns,
                           hashlib.sha256(p.read_bytes()).hexdigest() if p.is_file() else None))
    return result
before = manifest()
(s / 'before.json').write_text(json.dumps(before, indent=2) + '\n')
env = dict(os.environ, HOME=str(s / 'absent-home'), XDG_STATE_HOME=str(s / 'state'))
run = subprocess.run(['timeout', '10s', str(s / 'bin/familiar'), 'doctor', '--json', '--offline'],
                     cwd=s / 'workspace', env=env, capture_output=True, text=True)
(s / 'doctor.json').write_text(run.stdout)
(s / 'doctor.stderr').write_text(run.stderr)
assert run.returncode == 1, (run.returncode, run.stderr)
r = json.loads(run.stdout)
assert r['schema_version'] == 1 and r['status'] == 'blocked', r
checks = {c['id']: c for c in r['checks']}
for key in ('platform', 'search-import', 'native-linkage'):
    assert checks[key]['status'] == 'ready', checks[key]
for key in ('generation-endpoint', 'embedding-endpoint'):
    assert checks[key]['status'] == 'not-configured', checks[key]
after = manifest()
(s / 'after.json').write_text(json.dumps(after, indent=2) + '\n')
assert before == after, (before, after)
assert not list((s / 'workspace').iterdir()) and not list((s / 'state').iterdir())
assert not (s / 'absent-home').exists()
print('CLI acceptance PASS: exit 1, schema 1, explicit blockers, unchanged empty roots')
PY
LOADER=$(readelf -l "$SCRATCH/bin/familiar" | sed -n 's/.*interpreter: \(.*\)]/\1/p')
set +e
LD_LIBRARY_PATH= timeout 10s "$LOADER" --inhibit-cache --inhibit-rpath '' \
  --library-path /usr/lib/x86_64-linux-gnu:/lib/x86_64-linux-gnu \
  "$SCRATCH/bin/familiar" doctor --json --offline >"$SCRATCH/loader.stdout" 2>"$SCRATCH/loader.stderr"
rc=$?
set -e
test "$rc" -ne 0
grep -E 'error while loading shared libraries: lib(fffbridge|faiss)' "$SCRATCH/loader.stderr"
printf 'Missing-native loader gate PASS: exit %s\n' "$rc"
