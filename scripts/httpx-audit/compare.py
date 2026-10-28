#!/usr/bin/env python3
"""Compare every httpx fork against the pinned upstream snapshot.

Mirrors the dnsx fork-audit methodology. Robust single-process harvester:
- gh api via subprocess with stdin=DEVNULL (no loop-stdin drain)
- explicit error handling (404 / no-ancestor / transient)
- resumable: skips forks already recorded in the output
Output: docs/research/httpx/fork-compare.jsonl (one JSON object per fork)
"""
import json
import os
import subprocess
import sys

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
RES = os.path.join(ROOT, "docs", "research", "httpx")
BASE_REPO = "projectdiscovery/httpx"

with open(os.path.join(RES, "upstream-snapshot.txt")) as fh:
    SNAP = fh.read().strip()

IN = os.path.join(RES, "forks.tsv")
OUT = os.path.join(RES, "fork-compare.jsonl")
PROG = os.path.join(RES, "fork-compare.progress")

# resume: collect already-done fork names
done = set()
if os.path.exists(OUT):
    with open(OUT) as fh:
        for line in fh:
            line = line.strip()
            if not line:
                continue
            try:
                done.add(json.loads(line)["fork"])
            except Exception:
                pass

forks = []
with open(IN) as fh:
    for line in fh:
        parts = line.rstrip("\n").split("\t")
        if len(parts) < 2 or not parts[0]:
            continue
        full, branch = parts[0], parts[1]
        pushed = parts[2] if len(parts) > 2 else ""
        size = parts[3] if len(parts) > 3 else "0"
        forks.append((full, branch, pushed, size))

total = len(forks)
out = open(OUT, "a", buffering=1)  # line-buffered append
n = 0
for full, branch, pushed, size in forks:
    n += 1
    if full in done:
        continue
    owner = full.split("/", 1)[0]
    head = f"{owner}:{branch}"
    rec = {"fork": full, "branch": branch, "pushed": pushed,
           "size": int(size) if str(size).isdigit() else 0}
    try:
        cp = subprocess.run(
            ["gh", "api", f"repos/{BASE_REPO}/compare/{SNAP}...{head}",
             "--jq", "{status:.status, ahead:.ahead_by, behind:.behind_by, total:.total_commits}"],
            stdin=subprocess.DEVNULL, capture_output=True, text=True, timeout=40)
        if cp.returncode == 0 and cp.stdout.strip():
            d = json.loads(cp.stdout.strip())
            rec.update({"status": d.get("status"), "ahead": d.get("ahead", 0),
                        "behind": d.get("behind", 0), "total": d.get("total", 0)})
        else:
            err = cp.stderr.strip()
            status = "ERROR_404" if "404" in err or "Not Found" in err else "ERROR"
            rec.update({"status": status, "ahead": 0, "behind": 0, "total": 0,
                        "err": err[:160]})
    except subprocess.TimeoutExpired:
        rec.update({"status": "TIMEOUT", "ahead": 0, "behind": 0, "total": 0})
    out.write(json.dumps(rec) + "\n")
    if n % 25 == 0:
        with open(PROG, "w") as pf:
            pf.write(f"{n} / {total}\n")

out.close()
with open(PROG, "w") as pf:
    pf.write(f"DONE {n} / {total}\n")
open(os.path.join(RES, "fork-compare.done"), "w").close()
print(f"DONE {n} / {total}")
