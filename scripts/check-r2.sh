#!/usr/bin/env bash
set -euo pipefail
: "${SCRATCH:?run check-f3.sh in fresh scratch first}"
: "${NATIVE:?installed native prefix}"
umask 077
export GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local GOMAXPROCS=2 GOMEMLIMIT=256MiB
export CGO_ENABLED=1 GOCACHE="$SCRATCH/cache" TMPDIR="$SCRATCH/tmp"
export OMP_NUM_THREADS=2 OPENBLAS_NUM_THREADS=2 RAYON_NUM_THREADS=2
export PKG_CONFIG_PATH="$NATIVE/lib/pkgconfig" CGO_CFLAGS="-I$NATIVE/include"
export CGO_LDFLAGS="-L$NATIVE/lib -Wl,-rpath,$NATIVE/lib" LD_LIBRARY_PATH="$NATIVE/lib"
export GOWORK=off R2_SCRATCH="$SCRATCH/r2"
timeout 60s go build -mod=readonly -p=1 -tags=vectors ./...
timeout 60s go vet -mod=readonly -p=1 -tags=vectors ./...
test -z "$(gofmt -l cmd doctor retrieval)"
git diff --check
timeout 60s go test -mod=readonly -count=1 -p=1 -tags=vectors ./retrieval \
  -run TestScopedEvidenceAcceptance -v -timeout 45s
