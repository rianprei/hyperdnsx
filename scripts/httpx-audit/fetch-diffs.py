#!/usr/bin/env python3
"""Fetch each functional-delta fork and save its net diff vs the merge-base.

Shallow-fetches only the fork's divergent commits (depth = ahead+3), finds the
merge-base against upstream dev, and writes the fork-unique .go/.md diff as
evidence to docs/research/httpx/fork-analysis/<slug>.diff.
"""
import json
import os
import subprocess

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), "..", ".."))
RES = os.path.join(ROOT, "docs", "research", "httpx")
ANALYSIS = os.path.join(RES, "fork-analysis")
UPSTREAM = "/data/data/com.termux/files/home/httpx-upstream-audit"
os.makedirs(ANALYSIS, exist_ok=True)

# adoption-candidate set: forks with a real functional delta worth porting
CANDIDATES = {
    "CarlosCommits/httpx", "kno-com/httpx", "l0ners/httpx", "bokostmu/httpx",
    "hetbhatt2807/httpx", "xiaoyaochen/httpx", "b3r1ch/httpx", "rabb1t-dev/httpx",
    "jsifuentes/httpx", "yhy0/httpx", "yabeow/httpx", "nyxgeek/httpx",
    "tracertea/httpx", "securifera/httpx", "atoz-project/httpx", "dkarv/httpx",
    "ayanrajpoot10/httpx", "P1kAju/httpx", "chudamax/httpx", "d11testing/httpx",
    "ohmygod20260203/httpx", "kiriknik/httpx", "sahruldotid/httpx", "zp857/httpx",
    "imfht/httpx", "pelamx/httpxCustom", "innerpeace609/httpx",
}

meta = {}
with open(os.path.join(RES, "fork-compare.jsonl")) as fh:
    for line in fh:
        d = json.loads(line)
        if d["fork"] in CANDIDATES:
            meta[d["fork"]] = (d["branch"], d.get("ahead", 1))


def run(args, **kw):
    return subprocess.run(args, cwd=UPSTREAM, stdin=subprocess.DEVNULL,
                          capture_output=True, text=True, **kw)


results = []
for fork, (branch, ahead) in sorted(meta.items()):
    owner, repo = fork.split("/", 1)
    url = f"https://github.com/{fork}.git"
    slug = fork.replace("/", "_")
    # full fetch: clone already has all upstream history, so only the fork's
    # divergent objects download. No --depth (shallow breaks merge-base).
    fr = run(["git", "fetch", "--quiet", url, branch], timeout=300)
    if fr.returncode != 0:
        results.append((fork, "FETCH_FAIL", fr.stderr.strip()[:120]))
        continue
    mb = run(["git", "merge-base", "origin/dev", "FETCH_HEAD"])
    base = mb.stdout.strip()
    if not base:
        results.append((fork, "NO_MERGEBASE", ""))
        continue
    diff = run(["git", "diff", f"{base}..FETCH_HEAD", "--",
                "*.go", "*.md", "*.yaml", "*.yml"], timeout=120)
    path = os.path.join(ANALYSIS, f"{slug}.diff")
    with open(path, "w") as out:
        out.write(diff.stdout)
    loc = diff.stdout.count("\n")
    results.append((fork, "OK", f"{loc} diff lines"))

for fork, status, info in results:
    print(f"{status:14} {fork:40} {info}")
