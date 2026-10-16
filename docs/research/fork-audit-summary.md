# HyperDNSx — Fork audit summary

> Summary of the direct-fork inventory and audit for `projectdiscovery/dnsx`.
> Based on `docs/research/forks.tsv` and `docs/research/fork-compare.jsonl`
> (356 unique forks, all validated against the GitHub API).

## Inventory

| Item | Value |
|------|-------|
| Direct forks catalogued | 356 |
| Forks compared (metadata + divergence) | 356 |
| Forks identical to upstream snapshot | 2 |
| Forks behind upstream (no functional delta) | ~313 |
| Forks inaccessible (HTTP 404) | 26 |
| Forks with no common ancestor (compare failed) | 1 (X-Team-Robots/dnsx) |
| Forks with functional delta examined | 16 |

## Audit method

1. Fetched all fork metadata via the paginated GitHub API
   (`repos/projectdiscovery/dnsx/forks?per_page=100`).
2. Deduplicated by owner/repository name.
3. For each fork, compared the fork's default branch against upstream
   `dev` (snapshot `eab63bea`) using the GitHub Compare API, recording
   `ahead`, `behind`, `status`, and total commits.
4. Forks inaccessible via the API (HTTP 404) are recorded as
   `ERROR_404`; no functional conclusion is invented for them.
5. Forks with `ahead > 0`, `diverged`, or `compare_failed` were
   deep-diff inspected (`git fetch` + `git diff`) to confirm whether the
   delta is functional, documentation-only, or cosmetic.

## Results by category

| Category | Forks | Action |
|----------|-------|--------|
| Identical to upstream | 2 | No functional delta; not adopted |
| Behind / outdated mirror | ~313 | No functional delta; not adopted |
| Inaccessible (404) | 26 | Listed as `UNVERIFIED` |
| Compare failed (no ancestor) | 1 | Listed as `UNVERIFIED` |
| Functional delta examined | 16 | See `docs/research/fork-feature-matrix.md` |

## Features adopted from forks

| Feature | Fork | HyperDNSx status |
|---------|------|------------------|
| Importable Go library | Kadrick/dnsx | Done (`libs/dnsx`, unit tested) |
| Multi-sink output (single writer lifecycle) | scournomad/dnsx | Planned (improved implementation) |
| CSV/TSV output formats | lucasolslund/dnsx | Planned |
| Publicsuffix base-domain extraction | Pitrat-wav/dnsx, assakafpix/dnsx | Noted for wildcard hardening |

## Features explicitly NOT adopted

| Feature | Fork | Reason |
|---------|------|--------|
| Hardcoded China resolver defaults | eexp/dnsx | No geographic resolver hardcoded as universal policy |
| Per-result file open + `defer` in worker | scournomad/dnsx | File-descriptor leak risk; stable writer lifecycle used instead |

## Forks recorded as UNVERIFIED

The following forks could not be fetched or compared. No functional delta
is assumed or claimed for them:

- `ERROR_404` (26 forks): sakirjd-dev-infinite/dnsx, whitehathackerpr/dnsx,
  salifnd-sketch/dnsx, SolariSystems/dnsx, Outs1d3r-Net/dnsx, ZankoyDll/dnsx,
  126789t/dnsx, eexp/dnsx (note: eexp also has a reachable diverged entry;
  both recorded), Hey-code-star/dnsx, mahzenb/projectdiscovery_dnsx,
  repomahzen/projectdiscovery_dnsx, axinvolve/dnsx, productionby16/dnsx,
  thanhthatthathat944/dnsx, camdanl/dnsx, Smoking-k/dnsx, limkokholefork/dnsx,
  vaniilla/dnsx, CPT-Jack-A-Castle/dnsx, bijaysenihang/dnsx, FLyG33K/dnsx,
  awsassets/dnsx, 5l1v3r1/dnsx,
  Bae1VAR552/bae1.bae.va.var52.BraeVincent.dnsx, hthanh6500/dnsx,
  hthanh6503/dnsx, thanhhoanghh/dnsx
- `compare_failed` (1 fork): X-Team-Robots/dnsx (no common ancestor with
  upstream `dev`)

## License / attribution note

No code was copied from GPL-3.0 projects (puredns, shuffledns, MassDNS).
Only publicly described, high-level ideas were reimplemented independently.
All fork-derived ideas are attributed above with commit/file evidence.
