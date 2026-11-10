#!/usr/bin/env python3
"""One real local-model acceptance journey; no model stubs or fixture workspace.

FAMILIAR_BIN, FAMILIAR_ENDPOINT, FAMILIAR_MODEL, FAMILIAR_WORKSPACE and
FAMILIAR_CHECK_DIR are explicit operator selections. The model must support
OpenAI-compatible structured calls. The workspace must be this Familiar checkout.
This script does not start/stop the operator's model server.
"""
import json
import os
from pathlib import Path
import signal
import subprocess
import time

binary = os.environ["FAMILIAR_BIN"]
workspace = Path(os.environ["FAMILIAR_WORKSPACE"]).resolve(strict=True)
assert (workspace / "doctor/doctor.go").is_file(), "select real Familiar checkout"
artifacts = Path(os.environ["FAMILIAR_CHECK_DIR"])
artifacts.mkdir(mode=0o700, parents=True, exist_ok=False)
state = artifacts / "sessions"
common = [binary, "task", "--state", str(state), "--workspace", str(workspace),
          "--read-workspace", "--endpoint", os.environ["FAMILIAR_ENDPOINT"],
          "--model", os.environ["FAMILIAR_MODEL"]]
if os.environ.get("FAMILIAR_LLAMA") == "1":
    common += ["--llama"]
    if os.environ.get("FAMILIAR_THINKING"):
        common += ["--thinking", os.environ["FAMILIAR_THINKING"]]
    if os.environ.get("FAMILIAR_CACHE_PROMPT") == "1":
        common += ["--cache-prompt"]


def run(args, text, name):
    result = subprocess.run(args, input=text, text=True, capture_output=True, timeout=360)
    (artifacts / (name + ".stdout")).write_text(result.stdout)
    (artifacts / (name + ".stderr")).write_text(result.stderr)
    assert result.returncode == 0, (name, result.returncode, result.stderr)
    return result


run(common, "Where is the Go function Inspect defined? Use search for the literal "
    "symbol Inspect and read its implementation. Explain the generation-endpoint "
    "readiness status using source paths and line numbers. Do not run shell commands.\n/quit\n", "first")
journals = list(state.glob("*.jsonl"))
assert len(journals) == 1
journal = journals[0]
events = [json.loads(line) for line in journal.read_text().splitlines()]
assert events[-1]["kind"] == "task_completed", "model must actually finish"
calls = [e for e in events if e["kind"] == "tool_intent"]
assert {"search", "read"} <= {e["call"]["function"]["name"] for e in calls}
assert any(e["call"]["function"]["name"] == "read" and
           json.loads(e["call"]["function"]["arguments"])["path"] == "doctor/doctor.go"
           for e in calls), "read actual implementation, not just documentation/test references"
for call in calls:
    results = [e for e in events if e["kind"] == "tool_result"
               and e["message"]["tool_call_id"] == call["call"]["id"]]
    assert len(results) == 1 and results[0]["sequence"] > call["sequence"]
    assert not results[0]["message"]["content"].startswith("ERROR:"), results
answer = [e["message"]["content"] for e in events if e["kind"] == "message"
          and e["message"]["role"] == "assistant"][-1]
assert "doctor/doctor.go" in answer and "not-configured" in answer and any(c.isdigit() for c in answer), answer
before = journal.read_bytes()
replayed = run([binary, "task", "--state", str(state), "--session", journal.stem,
               "--replay"], "", "replay")
assert "[tool result" in replayed.stdout and "[assistant completed]" in replayed.stdout
assert journal.read_bytes() == before, "replay must not append or dispatch"
run(common + ["--session", journal.stem],
    "Briefly summarize your previous finding, using the existing conversation.\n/quit\n",
    "continued")
after = [json.loads(line) for line in journal.read_text().splitlines()]
assert after[:len(events)] == events, "append-only history must remain unchanged"
assert len([e for e in after if e["kind"] == "task_completed"]) == 2
for call in calls:
    assert sum(e["kind"] == "tool_intent" and e["call"]["id"] == call["call"]["id"]
               for e in after) == 1, "old tool intents must not be replayed"
# Cancel a fresh turn only after an actual streamed observation is durable.
# This uses the same real conversation/runtime, not a mocked cancellation path.
with (artifacts / "canceled.stdout").open("w") as stdout, \
        (artifacts / "canceled.stderr").open("w") as stderr:
    child = subprocess.Popen(common + ["--session", journal.stem], stdin=subprocess.PIPE,
                             stdout=stdout, stderr=stderr, text=True)
    try:
        child.stdin.write("Explain your previous source finding in more detail.\n")
        child.stdin.flush()
        deadline = time.monotonic() + 90
        while time.monotonic() < deadline:
            # Observe only newline-terminated records while the CLI is appending.
            current = [json.loads(line) for line in journal.read_bytes().split(b"\n")[:-1]]
            if any(e["kind"] in ("text_delta", "reasoning_delta", "tool_call_delta")
                   for e in current[len(after):]):
                child.send_signal(signal.SIGINT)
                break
            assert child.poll() is None, "model turn ended before streaming cancellation"
            time.sleep(0.05)
        else:
            raise AssertionError("no real streamed observation before cancellation deadline")
        assert child.wait(timeout=10) == 2
        canceled = [json.loads(line) for line in journal.read_text().splitlines()]
        assert canceled[-1]["kind"] == "task_canceled"
    finally:
        if child.poll() is None:
            child.kill()
            child.wait(timeout=10)
        child.stdin.close()
print("PASS: real local model search/read/result/answer, read-only replay, restart continuation and streaming cancellation")
print("Session:", journal.stem)
print("Journal:", journal)
