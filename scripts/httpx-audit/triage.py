#!/usr/bin/env python3
"""Triage the diverged (ahead>0) httpx forks by their unique commit subjects.

For each candidate fork, fetch the compare BASE...fork:branch and record the
fork's divergent commit messages + changed file list. High-signal, cheap first
pass before real deep-diff. Output: docs/research/httpx/fork-triage.jsonl
"""
import json
import os
import subprocess

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
RES = os.path.join(ROOT, "docs", "research", "httpx")
BASE_REPO = "projectdiscovery/httpx"
with open(os.path.join(RES, "upstream-snapshot.txt")) as fh:
    SNAP = fh.read().strip()

# candidates = diverged with ahead>0
cands = []
with open(os.path.join(RES, "fork-compare.jsonl")) as fh:
    for line in fh:
        d = json.loads(line)
        if d.get("ahead", 0) > 0:
            cands.append((d["ahead"], d["fork"], d["branch"]))
cands.sort(reverse=True)

out = open(os.path.join(RES, "fork-triage.jsonl"), "w", buffering=1)
for ahead, full, branch in cands:
    owner = full.split("/", 1)[0]
    head = f"{owner}:{branch}"
    rec = {"fork": full, "branch": branch, "ahead": ahead}
    try:
        cp = subprocess.run(
            ["gh", "api", f"repos/{BASE_REPO}/compare/{SNAP}...{head}"],
            stdin=subprocess.DEVNULL, capture_output=True, text=True, timeout=60)
        if cp.returncode == 0 and cp.stdout.strip():
            d = json.loads(cp.stdout)
            msgs = [c["commit"]["message"].splitlines()[0]
                    for c in d.get("commits", [])]
            files = sorted({f["filename"] for f in d.get("files", [])})
            rec["commits"] = msgs
            rec["files"] = files
        else:
            rec["error"] = cp.stderr.strip()[:160]
    except subprocess.TimeoutExpired:
        rec["error"] = "timeout"
    out.write(json.dumps(rec) + "\n")
out.close()
print(f"triaged {len(cands)} candidates")
