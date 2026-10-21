# Synthetic garden corpus v1

`v1/manifest.json` is the versioned contract: 10 small invented UTF-8 text files,
13 judged tasks (two per positive category, three negative controls), input byte
lengths/SHA-256, and gold evidence. No extracted personal data or binary formats.
Corpus text and manifest are dedicated to the public domain under
[CC0 1.0](https://creativecommons.org/publicdomain/zero/1.0/).

Gold references name files whose `gold` object specifies an exact UTF-8 half-open
byte window `[start,end)` and verbatim quote; the file SHA-256 binds that window
to its version. These are offline judgments, **not a production citation API**.
Evidence windows include whole short paragraphs, including one non-ASCII name.
The archived draft is deliberately obsolete; byte-identical receipts and a
near-duplicate revision must not be conflated. Change corpus version and hashes
when changing inputs/gold, rather than silently reusing a benchmark identity.

## Offline use

Python 3.9+ (stdlib only), Linux `/proc`, installed native dependencies, and the
same cached Go build setup as F3 are required. Run from the repository root:

```sh
python3 scripts/fixture_baseline.py --corpus "$PWD/fixtures/v1" --verify-only
```

The loader verifies the complete input inventory, hashes, evidence ranges and
task references before materializing byte-identical files in an explicitly
selected empty scratch directory. It never adopts a personal/workspace root.
`cmd/familiar-baseline` is a measurement helper, intended only for those verified
scratch copies. It requires explicit root and empty external state, indexes with
the real pinned FFF/Bleve search engine, and returns schema-versioned raw samples.
It is not a second agent interface and does not independently load/judge gold.
Its required dimension of 1 is an unused index-schema placeholder, **not an
embedding profile**; no vectors or semantic queries are supplied.

Fixed retrieval probes are hand-authored baseline queries, not a claim that a
router can derive them from the prompts. Lanes: fuzzy path lookup, plain literal
grep (first five distinct files), and keyword top five. There is one unmeasured
warmup per probe followed by 20 measured rounds, with no query rewriting or
reranking. Relevant files are required, not an exhaustive ban on other hits;
absent-input controls require zero hits. Per-probe and overall p95 use the
nearest-rank statistic. Recall averages only nonempty relevance judgments;
absent controls are reported separately. Index/open time and process wall time
are separate from warm query latency. Two runs use fresh independent indexes.
Raw samples, index evidence, provenance, before/after manifests and summary JSON
stay outside the input tree. Sampling process-tree RSS is not an exact kernel
peak or an enforced resource limit; Go's soft heap limit excludes native memory.
VRAM is explicitly unmeasured; CPU-only retrieval cannot qualify model fit.
