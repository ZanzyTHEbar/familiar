#!/usr/bin/env python3
"""One real Linux worker journey; no mocked engine. Retains fresh fixtures.

Usage: check-scoped-boundary.py BUNDLE NEW_SCRATCH
"""
import json
import os
from pathlib import Path
import subprocess
import sys

os.umask(0o077)
bundle, base = (Path(p).resolve() for p in sys.argv[1:])
base.mkdir(mode=0o700)
root, state = base / 'workspace', base / 'state'
root.mkdir(mode=0o700)
state.mkdir(mode=0o700)
allowed, outside = root / 'brief.txt', base / 'outside.txt'
allowed.write_text('permittedboundarytoken exact evidence\n')
outside.write_text('unrelatedsecretmustnotberead\n')
(root / 'escape').symlink_to(outside)
(root / '.hidden').write_text('hiddenpolicytoken\n')
(root / '.ignore').write_text('ignored.txt\n')
(root / 'ignored.txt').write_text('ignorepolicytoken\n')

def start(probe):
    root_fd = os.open(root, os.O_RDONLY | os.O_DIRECTORY)
    state_fd = os.open(state, os.O_RDONLY | os.O_DIRECTORY)
    unrelated_fd = os.open(outside, os.O_RDONLY)
    assert (root_fd, state_fd, unrelated_fd) == (3, 4, 5), 'run this standalone with no extra open FDs'
    proc = subprocess.Popen([str(bundle / 'launcher'), str(bundle), str(root), str(state), str(probe)],
                            stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                            pass_fds=(root_fd, state_fd, unrelated_fd), env={'PATH': '/nonexistent'}, text=True)
    os.close(root_fd)
    os.close(state_fd)
    os.close(unrelated_fd)
    return proc

p = start(root / 'escape')  # Genuine worker must get EACCES through the symlink.
try:
    def call(request):
        p.stdin.write(json.dumps(request) + '\n')
        p.stdin.flush()
        response = json.loads(p.stdout.readline())
        assert not response['Error'], response
        return response
    call({'Op': 'open'})
    hits = call({'Op': 'search', 'Query': {'Mode': 'keyword', 'Text': 'permittedboundarytoken', 'Limit': 1}})['Hits']
    assert len(hits) == 1, hits
    hit = hits[0]
    excerpt = call({'Op': 'expand', 'Expand': {'ID': hit['id'], 'ExpectedContentHash': hit['content_hash'],
                  'ExpectedSource': hit['source'], 'Field': 'text', 'Offset': 0, 'Limit': 4096}})['Excerpt']
    assert excerpt['text'] == allowed.read_text(), excerpt
    policy = {}
    for key, token in [('outside', 'unrelatedsecretmustnotberead'), ('hidden', 'hiddenpolicytoken'), ('ignored', 'ignorepolicytoken')]:
        found = call({'Op': 'search', 'Query': {'Mode': 'keyword', 'Text': token, 'Limit': 20}})['Hits'] or []
        policy[key] = len(found)
    assert policy['outside'] == 0, policy
    p.stdin.close()
    rc = p.wait(timeout=10)
    assert rc == 0, (rc, p.stderr.read())
finally:
    if p.poll() is None:
        p.kill()
        p.wait()

# Also verify the direct outside-file denial, not only symlink resolution.
p = start(outside)
try:
    stdout, stderr = p.communicate('{"Op":"open"}\n', timeout=10)
    assert p.returncode == 0 and not json.loads(stdout)['Error'], (p.returncode, stdout, stderr)
finally:
    if p.poll() is None:
        p.kill()
        p.wait()

# The same genuine restricted worker must refuse startup when its required
# denial check is accidentally pointed at a readable, granted file.
p = start(allowed)
try:
    stdout, stderr = p.communicate('{"Op":"open"}\n', timeout=10)
    assert p.returncode != 0 and not stdout and 'outside read was not confined' in stderr, (p.returncode, stdout, stderr)
finally:
    if p.poll() is None:
        p.kill()
        p.wait()
result = {'permitted_file_and_hash_bound_expansion': True, 'outside_direct_open_EACCES_before_engine': True,
          'outside_symlink_open_EACCES_before_engine': True,
          'startup_fail_closed': True, 'observed_search_hit_counts': policy, 'owned_processes_remaining': []}
(base / 'result.json').write_text(json.dumps(result, indent=2) + '\n')
print(json.dumps(result, indent=2))
