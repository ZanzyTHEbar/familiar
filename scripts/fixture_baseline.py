#!/usr/bin/env python3
"""Offline F4 fixture loader and retrieval judge. No model/task-success scoring."""
import argparse
import hashlib
import json
import math
import os
from pathlib import Path, PurePosixPath
import platform
import subprocess
import time


def require(condition, message):
    if not condition:
        raise ValueError(message)


def sha(data):
    return hashlib.sha256(data).hexdigest()


def snapshot(root):
    result = []
    for p in [root, *sorted(root.rglob('*'))]:
        st = p.lstat()
        require(not p.is_symlink() and (p.is_dir() or p.is_file()), f'unsafe input: {p}')
        result.append((str(p.relative_to(root)), st.st_mode, st.st_size,
                       st.st_mtime_ns, sha(p.read_bytes()) if p.is_file() else None))
    return result


def load(root):
    require(root.is_absolute() and root.resolve() == root, 'corpus must be absolute and nonsymlink')
    before = snapshot(root)
    raw = (root / 'manifest.json').read_bytes()
    require(len(raw) <= 1 << 20, 'manifest too large')
    m = json.loads(raw)
    require(m['schema_version'] == 1 and m['corpus_version'] == 'synthetic-garden-v1', 'unsupported schema/corpus')
    require(m['license'] == 'CC0-1.0', 'fixture license missing')
    require(1 <= len(m['files']) <= 100, 'invalid corpus size')
    contents = {}
    for f in m['files']:
        p = PurePosixPath(f['path'])
        require(not p.is_absolute() and '..' not in p.parts and str(p) == f['path']
                and str(p) != '.' and '\\' not in str(p), 'unsafe fixture path')
        require(f['path'] not in contents, 'duplicate fixture path')
        data = (root / 'files' / f['path']).read_bytes()
        data.decode('utf-8')
        require(b'\0' not in data and len(data) <= 64 << 10, 'not bounded UTF-8 text')
        require(len(data) == f['bytes'] and sha(data) == f['sha256'], f'input hash/size: {p}')
        g = f['gold']
        require(type(g['start']) is int and type(g['end']) is int
                and 0 <= g['start'] < g['end'] <= len(data), f'invalid gold range: {p}')
        require(data[g['start']:g['end']].decode('utf-8') == g['quote'], f'gold quote mismatch: {p}')
        contents[f['path']] = data
    require(sum(map(len, contents.values())) <= 1 << 20, 'corpus byte cap')
    actual = {str(p.relative_to(root)) for p in root.rglob('*') if p.is_file()}
    require(actual == {'manifest.json'} | {'files/' + p for p in contents}, 'unexpected/missing input')
    ids, categories = set(), {}
    require(1 <= len(m['tasks']) <= 100, 'invalid task count')
    for t in m['tasks']:
        require(t['id'] and t['id'] not in ids, 'missing/duplicate task ID')
        ids.add(t['id'])
        categories[t['category']] = categories.get(t['category'], 0) + 1
        require(t['prompt'].strip() and t['required'] and all(s.strip() for s in t['required']), 'unjudged task')
        require(len(t['gold']) == len(set(t['gold'])) and set(t['gold']) <= contents.keys(), 'invalid gold references')
        p = t['probe']
        require(p['lane'] in ('path', 'grep', 'keyword') and p['query'].strip(), 'invalid retrieval probe')
        require(len(p['relevant']) <= 5 and len(p['relevant']) == len(set(p['relevant']))
                and set(p['relevant']) <= set(t['gold']), 'invalid retrieval judgment')
        if t['category'] != 'negative':
            require(t['gold'] and p['relevant'], 'positive task lacks evidence')
        if t['category'] == 'cross_document':
            require(len(t['gold']) >= 2, 'cross-document task needs multiple sources')
    require(set(categories) == {'exact_lookup', 'sourced_answer', 'cross_document', 'duplicates', 'organization', 'negative'}
            and min(categories.values()) >= 2, 'missing representative task categories')
    require(snapshot(root) == before, 'corpus changed during verification')
    return m, contents, sha(raw), before


def memory(pid):
    """Sample the live process tree; never label this an exact kernel peak."""
    pending, seen, rss, hwm = [pid], set(), 0, 0
    while pending:
        child = pending.pop()
        if child in seen:
            continue
        seen.add(child)
        try:
            status = Path(f'/proc/{child}/status').read_text()
            fields = {line.split(':', 1)[0]: line.split(':', 1)[1].strip() for line in status.splitlines()}
            rss += int(fields.get('VmRSS', '0 kB').split()[0])
            hwm = max(hwm, int(fields.get('VmHWM', '0 kB').split()[0]))
            for task in Path(f'/proc/{child}/task').iterdir():
                pending.extend(int(s) for s in (task / 'children').read_text().split())
        except FileNotFoundError:
            pass  # Exit between samples; limitation recorded in the report.
    return rss, hwm, len(seen) - 1


def measure(binary, contents, m, dest):
    dest.mkdir(mode=0o700)
    inputs, state = dest / 'inputs', dest / 'state'
    inputs.mkdir(mode=0o700)
    state.mkdir(mode=0o700)
    for name, data in contents.items():
        p = inputs / name
        p.parent.mkdir(mode=0o700, parents=True, exist_ok=True)
        with p.open('xb') as out:
            out.write(data)
        p.chmod(0o600)
    before = snapshot(inputs)
    (dest / 'inputs-before.json').write_text(json.dumps(before, indent=2) + '\n')
    probes = [{'id': t['id'], 'lane': t['probe']['lane'], 'query': t['probe']['query']} for t in m['tasks']]
    probe_file = dest / 'probes.json'
    probe_file.write_text(json.dumps(probes) + '\n')
    start = time.monotonic()
    peak, hwm, descendants, count, max_gap, previous = 0, 0, 0, 0, 0, start
    with probe_file.open('rb') as stdin, (dest / 'raw.json').open('wb') as stdout, (dest / 'stderr').open('wb') as stderr:
        proc = subprocess.Popen([str(binary), '--root', str(inputs), '--state', str(state)],
                                stdin=stdin, stdout=stdout, stderr=stderr, cwd=dest)
        try:
            while proc.poll() is None:
                now = time.monotonic()
                max_gap = max(max_gap, now - previous)
                previous = now
                rss, high, children = memory(proc.pid)
                peak, hwm, descendants = max(peak, rss), max(hwm, high), max(descendants, children)
                count += 1
                require(now - start <= 30, 'baseline exceeded 30s wall deadline')
                time.sleep(0.005)
        finally:
            if proc.poll() is None:
                proc.kill()
                proc.wait()
    wall = time.monotonic() - start
    require(proc.returncode == 0, f'baseline exit {proc.returncode}: {(dest / "stderr").read_text()}')
    after = snapshot(inputs)
    (dest / 'inputs-after.json').write_text(json.dumps(after, indent=2) + '\n')
    require(after == before, 'search changed synthetic inputs')
    r = json.loads((dest / 'raw.json').read_text())
    require(r['schema_version'] == 1 and r['build']['indexed'] == len(contents)
            and r['build']['errors'] == 0 and r['build']['skipped'] == 0, 'incomplete corpus indexing')
    require(len(r['samples']) == len(probes) * 20, 'incomplete measurements')
    by_id = {t['id']: t for t in m['tasks']}
    judgments, seen, latencies, recalls = {}, set(), [], []
    for s in r['samples']:
        key = (s['id'], s['round'])
        require(s['id'] in by_id and 1 <= s['round'] <= 20 and key not in seen, 'invalid measurement identity')
        seen.add(key)
        paths = s['paths']
        require(len(paths) <= 5 and len(paths) == len(set(paths)) and set(paths) <= contents.keys(), 'invalid returned paths')
        relevant = set(by_id[s['id']]['probe']['relevant'])
        recall = len(relevant & set(paths)) / len(relevant) if relevant else float(not paths)
        require(recall == 1, f'retrieval judgment failed: {s["id"]}, got {paths}, need {sorted(relevant)}')
        if relevant:
            recalls.append(recall)
        require(type(s['elapsed_ns']) is int and s['elapsed_ns'] >= 0, 'invalid duration')
        latencies.append(s['elapsed_ns'] / 1e6)
        judgments[s['id']] = {'required': sorted(relevant), 'all_rounds_passed': True}
    p95 = sorted(latencies)[math.ceil(len(latencies) * .95) - 1]
    per_probe = {p['id']: sorted(s['elapsed_ns'] / 1e6 for s in r['samples'] if s['id'] == p['id']) for p in probes}
    require(p95 <= 500 and all(v[18] <= 500 for v in per_probe.values()), 'retrieval p95 exceeded 500ms')
    require(wall <= 30 and 0 < peak <= 512 * 1024 and hwm <= 512 * 1024, 'wall/RAM criterion failed or unavailable')
    require(descendants == 0, 'unexpected subprocess: baseline is designed as one process')
    return {'schema_version': 1, 'corpus_version': m['corpus_version'], 'status': 'passed',
            'judgments': judgments, 'mean_recall_at_5': sum(recalls) / len(recalls),
            'absent_controls_passed': [t['id'] for t in m['tasks'] if not t['probe']['relevant']],
            'warm_p95_ms': p95, 'per_probe_p95_ms': {k: v[18] for k, v in per_probe.items()},
            'wall_seconds': wall, 'index_ms': r['index_ns'] / 1e6,
            'sampled_peak_tree_rss_kib': peak, 'observed_process_hwm_kib': hwm,
            'memory_samples': count, 'max_sampling_gap_ms': max_gap * 1000,
            'observed_descendants': descendants, 'vram': 'not measured; CPU-only retrieval, no model qualification',
            'memory_method': 'Linux /proc tree RSS; 5ms target, sampled not exact peak; short-lived processes may be missed',
            'end_to_end_task_success': 'not measured; requires R6',
            'go_version': r['go_version'], 'platform': r['platform'], 'kernel': platform.release()}


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--corpus', required=True, type=Path)
    parser.add_argument('--verify-only', action='store_true')
    parser.add_argument('--binary', type=Path)
    parser.add_argument('--scratch', type=Path)
    args = parser.parse_args()
    os.umask(0o077)
    m, contents, manifest_hash, before = load(args.corpus)
    if args.verify_only:
        print(f'Fixture verification PASS: {m["corpus_version"]}, {len(contents)} files, {len(m["tasks"])} judged tasks, sha256 {manifest_hash}')
        return
    require(args.binary and args.binary.is_absolute() and args.scratch and args.scratch.is_absolute(), 'explicit binary/scratch required')
    require(args.scratch.resolve() == args.scratch and args.scratch.is_dir()
            and not list(args.scratch.iterdir()), 'scratch must be empty and nonsymlink')
    require(not args.scratch.is_relative_to(args.corpus), 'scratch must be outside corpus')
    (args.scratch / 'corpus-before.json').write_text(json.dumps(before, indent=2) + '\n')
    runs = [measure(args.binary, contents, m, args.scratch / f'run-{i}') for i in (1, 2)]
    require(runs[0]['judgments'] == runs[1]['judgments'], 'nonreproducible quality')
    after = snapshot(args.corpus)
    (args.scratch / 'corpus-after.json').write_text(json.dumps(after, indent=2) + '\n')
    require(after == before, 'committed corpus changed')
    report = {'schema_version': 1, 'manifest_sha256': manifest_hash, 'runs': runs}
    (args.scratch / 'report.json').write_text(json.dumps(report, indent=2) + '\n')
    print(json.dumps(report, indent=2))


if __name__ == '__main__':
    main()
