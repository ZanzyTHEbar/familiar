#!/usr/bin/env bash
set -euo pipefail
: "${SCRATCH:?run check-f3.sh in fresh scratch first}"
: "${SOURCE:?canonical search repository}"
: "${NATIVE:?installed native prefix}"
umask 077
export GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local GOMAXPROCS=2 GOMEMLIMIT=256MiB
export CGO_ENABLED=1 GOCACHE="$SCRATCH/cache" TMPDIR="$SCRATCH/tmp"
export OMP_NUM_THREADS=2 OPENBLAS_NUM_THREADS=2 RAYON_NUM_THREADS=2
export PKG_CONFIG_PATH="$NATIVE/lib/pkgconfig" CGO_CFLAGS="-I$NATIVE/include"
export CGO_LDFLAGS="-L$NATIVE/lib -Wl,-rpath,$NATIVE/lib" LD_LIBRARY_PATH="$NATIVE/lib"
export GOWORK="$SCRATCH/go.work"
test -f "$GOWORK"
timeout 60s go build -mod=readonly -p=1 -tags=vectors -o "$SCRATCH/bin/familiar-baseline" ./cmd/familiar-baseline
timeout 60s go vet -mod=readonly -p=1 -tags=vectors ./...
test -z "$(gofmt -l cmd doctor)"
git diff --check
mkdir -m 700 "$SCRATCH/f4"
{
  uname -srmo
  go version -m "$SCRATCH/bin/familiar-baseline"
  pkg-config --modversion fffbridge
  sha256sum "$NATIVE/lib/libfffbridge.so" "$NATIVE/lib/libfaiss_c.so" \
    "$NATIVE/lib/libfaiss.so" "$SCRATCH/bin/familiar-baseline"
} >"$SCRATCH/f4-provenance.txt"
python3 scripts/fixture_baseline.py --corpus "$PWD/fixtures/v1" --verify-only
python3 scripts/fixture_baseline.py --corpus "$PWD/fixtures/v1" \
  --binary "$SCRATCH/bin/familiar-baseline" --scratch "$SCRATCH/f4" >"$SCRATCH/f4-report.stdout"
# Real loader rejection boundaries, using disposable copies, not mocked search.
python3 - "$PWD" "$SCRATCH" <<'PY'
import json, pathlib, shutil, subprocess, sys
root, scratch = map(pathlib.Path, sys.argv[1:])
for kind in ('bytes', 'gold', 'unexpected'):
    dest = scratch / ('invalid-' + kind)
    shutil.copytree(root / 'fixtures/v1', dest)
    if kind == 'bytes':
        with (dest / 'files/cedar/brief.txt').open('ab') as f:
            f.write(b'changed')
    elif kind == 'gold':
        p = dest / 'manifest.json'
        m = json.loads(p.read_text())
        m['files'][0]['gold']['start'] = 1
        p.write_text(json.dumps(m))
    else:
        (dest / 'files/unlisted.txt').write_text('not approved')
    run = subprocess.run(['python3', str(root / 'scripts/fixture_baseline.py'),
                          '--corpus', str(dest), '--verify-only'], capture_output=True, text=True)
    (scratch / (kind + '-rejection.stderr')).write_text(run.stderr)
    assert run.returncode != 0, kind
    assert {'bytes': 'input hash/size', 'gold': 'gold quote mismatch',
            'unexpected': 'unexpected/missing input'}[kind] in run.stderr, run.stderr
    print('Loader rejection PASS:', kind)
print('F4 PASS: two real native retrieval runs; reproducible judgments, unchanged inputs')
PY
