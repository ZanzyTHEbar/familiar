#!/usr/bin/env python3
"""Package trusted, already installed native libraries; never fetch or install.

Run with the same offline CGO/pkg-config environment as the Familiar build.
Output must be a new absolute directory. ldd is used only on our trusted build.
"""
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys

os.umask(0o077)
if sys.platform != 'linux' or len(sys.argv) != 2:
    raise SystemExit('usage (Linux only): build-scoped-runtime.py NEW_ABSOLUTE_BUNDLE')
out = Path(sys.argv[1])
if not out.is_absolute():
    raise SystemExit('bundle must be absolute')
if ':' in str(out) or ';' in str(out) or '$' in str(out):
    raise SystemExit("bundle must not contain loader separators ':' or ';', or token marker '$'")
out.mkdir(mode=0o700)
(out / 'lib').mkdir(mode=0o700)
root = Path(__file__).resolve().parent.parent
compile_command = ['cc', '-O2', '-Wall', '-Wextra', '-Werror', '-c', '-o', str(out / 'launcher.o'), str(root / 'native/scoped-launcher/launcher.c')]
link_command = ['cc', '-static', '-o', str(out / 'launcher'), str(out / 'launcher.o')]
subprocess.run(compile_command, check=True)
subprocess.run(link_command, check=True)
shutil.copyfile(root / 'native/scoped-launcher/launcher.c', out / 'launcher.c')
shutil.copyfile(root / 'LICENSE', out / 'LICENSE')
subprocess.run(['go', 'build', '-mod=readonly', '-p=1', '-tags=vectors', '-o', str(out / 'worker'), './cmd/familiar-scoped-worker'], cwd=root, check=True)
output = subprocess.check_output(['ldd', str(out / 'worker')], text=True)
if 'not found' in output:
    raise SystemExit('unresolved native runtime: ' + output)
files = ['worker']
provenance = []
for line in output.splitlines():
    match = re.search(r'=> (/\S+) \(', line)
    if match:
        source = Path(match.group(1))
        target = out / 'lib' / source.name
        shutil.copyfile(source, target)
        files.append('lib/' + source.name)
        provenance.append({'source': str(source.resolve()), 'bundled': str(target.relative_to(out)), 'sha256': hashlib.sha256(target.read_bytes()).hexdigest()})
    else:
        match = re.match(r'\s*(/\S+) \(', line)
        if match:
            # PT_INTERP is a file-specific host runtime exception, not /lib access.
            source = Path(match.group(1)).resolve()
            files.append('@interpreter')
            provenance.append({'source': str(source), 'bundled': False,
                               'sha256': hashlib.sha256(source.read_bytes()).hexdigest()})
(out / 'runtime.files').write_text('\n'.join(sorted(set(files))) + '\n')
(out / 'provenance.json').write_text(json.dumps(provenance, indent=2) + '\n')
(out / 'build.json').write_text(json.dumps({'compile': compile_command, 'link': link_command,
    'compiler': subprocess.check_output(['cc', '--version'], text=True),
    'go': subprocess.check_output(['go', 'version'], text=True),
    'launcher_source_sha256': hashlib.sha256((out / 'launcher.c').read_bytes()).hexdigest()}, indent=2) + '\n')
print(out)
