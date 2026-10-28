# HyperDNSx — httpx Fork audit summary

> Summary of the direct-fork inventory and audit for
> [`projectdiscovery/httpx`](https://github.com/projectdiscovery/httpx).
> Based on `docs/research/httpx/forks.tsv` and
> `docs/research/httpx/fork-compare.jsonl` (1262 unique forks, all validated
> against the GitHub API). Mirrors the dnsx fork-audit methodology.

## Inventory

| Item | Value |
|------|-------|
| Direct forks catalogued (unique) | 1262 |
| Forks compared (metadata + divergence) | 1262 |
| Forks identical to upstream snapshot | 1 |
| Forks behind upstream (no functional delta) | 1014 |
| Forks inaccessible (HTTP 404 / errored) | 171 |
| Forks diverged (`ahead > 0`) | 76 |
| Forks with functional delta deep-diffed | 27 |

Upstream snapshot pinned at `projectdiscovery/httpx` `dev` =
`668a37ca4f46774441acba839d13cf178f8975a8` (see
`docs/research/httpx/upstream-snapshot.txt`).

## Audit method

1. Harvested all fork metadata via the paginated GitHub API
   (`repos/projectdiscovery/httpx/forks?per_page=100`), deduplicated by
   `owner/repo` → `forks.tsv`.
2. For each fork, compared its default branch against the pinned upstream
   snapshot using the GitHub Compare API
   (`compare/<snapshot>...<owner>:<branch>`), recording `status`, `ahead`,
   `behind`, and `total` → `fork-compare.jsonl`.
3. Forks returning HTTP 404 (deleted/private/renamed) are recorded as
   `ERROR_404`; no functional conclusion is invented for them.
4. Every fork with `ahead > 0` was triaged by its divergent commit subjects
   and changed-file list → `fork-triage.jsonl`.
5. Forks with a plausible functional delta were deep-diffed (`git fetch` of the
   fork branch into a full upstream clone + `git diff <merge-base>..FETCH_HEAD`)
   and the net diff saved as evidence → `fork-analysis/<slug>.diff`.

Reproducible via `scripts/httpx-audit/{compare,triage,fetch-diffs}.py`.

## Results by category

| Category | Forks | Action |
|----------|-------|--------|
| Identical to upstream | 1 | No functional delta; not adopted |
| Behind / outdated mirror | 1014 | No functional delta; not adopted |
| Inaccessible (404) | 171 | Listed as `UNVERIFIED` |
| Diverged, noise-only (CI/deps/banner/junk) | ~49 | Not adopted (see matrix) |
| Diverged, functional delta examined | 27 | See `fork-feature-matrix.md` |

## Headline features recovered from the fork ecosystem

The deep-diff of the 27 functional forks recovered the following capabilities
not present (or only partially present) in upstream `httpx`. Full evidence and
disposition in `fork-feature-matrix.md`.

- Headless / JS-runtime technology detection (React, Vite, TanStack).
- Response deduplication via simhash + LSH band index.
- Pluggable auth providers (basic / bearer / cookie / header / query).
- WAF detection and improved CDN classification.
- TLS/JA3 Chrome impersonation + HTTP/2-over-proxy CONNECT tunnel.
- Recog fingerprint-database matching and custom Wappalyzer fingerprint files.
- Enriched JSON output (redirect chain fields, body/headers-in-json, TLS
  subject org) and HTML/CSV report sinks.
- Virtual-host / custom-Host-header probing and CustomIP targets.
- CPE derivation (Wappalyzer → CPE, mappable to CVEs).
- JS-file, copyright, and custom-regex (subexpression) extraction.
- Robustness: HTTP/2 fallback control, resume concurrency fix, path-as-is.

## Features explicitly NOT adopted

| Feature / content | Fork(s) | Reason |
|-------------------|---------|--------|
| CI/workflow scaffolding | CarlosCommits (`.github/workflows/*`), testaorga, vshekhda, cjstang, Rsnehith | Repo-maintenance noise, not product features |
| Docker/K8s deploy manifests | mariandppb, realArcherL | Deployment glue, not httpx capability |
| Binary/junk uploads | Jayyeeeeeeeeeeeeeee (`PsExec.exe`), biseYpueza (movie `.md` spam), apt-grid (`.drawio`) | Junk/abandoned forks; nothing to port |
| Dependency bumps / banner removal / module-path renames | Nihility981, Explorer1092, many empty-signal diverged forks | No functional delta |

## Forks recorded as UNVERIFIED

171 forks returned HTTP 404 or an API error at audit time (deleted, made
private, or renamed). They are listed in `fork-compare.jsonl` with
`status: ERROR_404` / `ERROR`. No functional delta is assumed or claimed for
them.

## License / attribution note

Features are reimplemented from publicly visible fork source with per-fork
commit/file attribution in `fork-feature-matrix.md` and raw evidence in
`fork-analysis/`. No upstream-incompatible license was adopted; httpx is
MIT-licensed, as are the forks audited. Any port must re-verify the specific
fork's license before copying code verbatim.
