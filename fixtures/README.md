# Synthetic retrieval fixtures

This directory contains ten invented UTF-8 text files and thirteen judged tasks.
Corpus text and manifest are dedicated to the public domain under
[CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/).
They contain no extracted personal data, models or runtime binaries.

[`v1/manifest.json`](v1/manifest.json) describes each input's path, byte size,
SHA-256 and gold evidence. A gold range is an exact UTF-8 half-open byte window
`[start,end)` whose bytes match its quote. Tasks include prompts, required answer
criteria and fixed path/grep/keyword probes with relevant-file judgments. These
are fixture data, not a production citation API or demonstrated assistant answers.
The obsolete draft and identical/revised receipts are intentionally distinct.

## Verify without native dependencies

With **Python 3.9+**, run from the repository root:

```bash
python3 scripts/fixture_baseline.py --corpus "$PWD/fixtures/v1" --verify-only
```

This uses only the Python standard library. It validates input inventory, paths,
UTF-8, byte sizes/hashes, evidence ranges/quotes and task references. It does not
launch search, run a model or materialize a scratch corpus.

## Measure native retrieval

Full measurement additionally needs Linux `/proc`, the compatible native search
dependencies and the [build environment](../README.md#build). In the same shell
as that build, compile the helper:

```bash
go build -mod=readonly -p=1 -tags=vectors -o "$BUILD_DIR/familiar-baseline" ./cmd/familiar-baseline
```

Choose a new empty scratch directory outside the corpus; the loader creates two
independent verified input copies and index/state directories there:

```bash
umask 077
FIXTURE_RUN=$(mktemp -d "$BUILD_ROOT/familiar-fixtures.XXXXXX")
export OMP_NUM_THREADS=2 OPENBLAS_NUM_THREADS=2 RAYON_NUM_THREADS=2
python3 scripts/fixture_baseline.py --corpus "$PWD/fixtures/v1" \
  --binary "$BUILD_DIR/familiar-baseline" --scratch "$FIXTURE_RUN"
```

Use only disposable synthetic inputs, **not a personal workspace**. The helper
writes native indexes to scratch state. Explicit root/state arguments are not an
OS sandbox or proof that native dependencies cannot read ambient Git/configuration
outside the input tree. Do not rely on it for private-workspace confinement.

The loader checks that committed and copied inputs remain unchanged, requires
all judged relevant files in the returned top five and no hits for absent-input
probes, and compares the two runs' judgments. Each probe has one unmeasured warmup
and twenty measured rounds. It enforces p95 retrieval latency at most 500 ms,
per-run wall time at most 30 seconds and observed memory at most 512 MiB.
Reports and raw samples are written beneath the selected scratch directory.
Failures exit nonzero; do not alter the judgments to fit results.

Queries are fixed fixture probes, not model-generated search plans. The helper's
vector dimension of 1 is an unused index-schema placeholder: it supplies no
embeddings or semantic queries. RSS is sampled through `/proc`, not an exact
kernel-enforced aggregate peak, and VRAM is unmeasured. Retrieval hits do not
measure end-to-end task success, model fit, refusal behavior or safe file actions.
No corpus prompt is executed as an instruction to move or delete files.
