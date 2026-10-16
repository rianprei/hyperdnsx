# HyperDNSx — Fork feature matrix

> Derived from `docs/research/fork-compare.jsonl` (356 unique, validated forks)
> and deep diff inspection of the 16 forks with `ahead > 0` or `diverged /
> compare_failed` status. Coverage is calculated over the validated inventory,
> not a historical count.

## Inventory summary

| Metric | Value |
|--------|-------|
| Total direct forks catalogued | 356 |
| Forks compared (metadata + upstream divergence) | 356 |
| Forks inaccessible (404 / compare failed) | 26 + 1 compare_failed |
| Forks identical to upstream snapshot | 2 |
| Forks behind upstream (no functional delta) | ~313 |
| Forks with functional delta examined | 16 |

## Forks with functional delta (deep-diff examined)

| Fork | Status | Ahead | Delta found | Evidence (commit/files) | Assessment |
|------|--------|-------|-------------|-------------------------|------------|
| scournomad/dnsx | diverged | 1 | `-json-file/-jf`: write JSONL to a file while keeping human-readable terminal output | commit `dbc8265` "Added a JSON File flag" (options.go, runner.go, README) | Useful pattern (multi-sink output) but implementation opens/closes the file **per result** with `defer` inside the worker — risk of leaking file descriptors. HyperDNSx implements multi-sink with a single writer opened once and flushed/closed at lifecycle end (see docs/research/gap-matrix.md). |
| lucasolslund/dnsx | diverged | 1 | CSV output (`AllRecords` formatted as CSV lines) | commit `41b3293` "Output to CSV" | Useful format addition. HyperDNSx plans CSV/TSV output sinks with stable writers (gap-matrix P1). |
| Pitrat-wav/dnsx | diverged | 2 | `-sw/--strict-wildcard`: strict wildcard detection/filtering + publicsuffix-based base domain extraction | commits `7ef3fa1` (auto wildcard), `bfe7a9a` (publicsuffix) | Wildcard already in upstream v1.3.0 (`-auto-wildcard/-aw`); not exclusive. Publicsuffix base-domain extraction is a sound improvement. |
| assakafpix/dnsx | diverged | 2 | `-sw` strict wildcard flag | commit `ce6dd52` | Same as Pitrat-wav; wildcard already upstream. |
| w2xim3/dnsx | diverged | 1 | FIFO input loop ("Added for ever loop for FIFO input") | commit `b40084c` | Niche input mode; not prioritized. |
| Kadrick/dnsx | diverged | 1 | "Dnsx as a lib" (library usage) | commit `90e490c` "Add readme contents: Dnsx as a lib" | HyperDNSx already exposes `libs/dnsx` as an importable Go package (FASE 3/4). |
| eexp/dnsx | diverged | 3 | China-resolver defaults (114.114.114.114, etc.) + `github.com/eexp/cdncheck` dependency; module path changed | many dependency-bump commits; very far behind upstream | Not adopted: hardcoding a geographic resolver set as default violates the "no geographic resolver hardcoded as universal policy" rule. |
| jiangjiang5757-cmyk/dnsx | diverged | 1 | `--auto-wildcard` flag | commit `73dfad4` | Auto-wildcard already in upstream v1.3.0. |
| gcarmonatellez15-spec/dnsx | diverged | 1 | `-output-template` flag | commit `008a2bc` | Output template already in upstream. |
| khozakhulile27-netizen/dnsx | diverged | 3 | `-output-template` + comma-separated `-l` | commits `008a2bc`, `af9c088` | Output template already upstream. |
| beike2020/dnsx | diverged | 4 | `-output-template` + dependency bumps | commit `008a2bc` | Output template already upstream. |
| felipetruman/dnsx | diverged | 1 | JSON unmarshal error handling | commit `4358d3f` | Error-handling hardening; consistent with HyperDNSx error-handling rules. |
| 123hi123/dnsx | diverged | 4 | Misc. dependency/config changes | several `Update dnsx.go/runner.go` commits | No distinct functional delta identified. |
| fadinglr/dnsx | diverged | 1 | "Ignore raw without -raw option" | commit `ebd2a77` | Minor output behavior; subsumed by existing `-raw`/`-or` flags. |
| X-Team-Robots/dnsx | compare_failed | 0 | No common ancestor with upstream `dev` | `compare` API returned 404 (no common ancestor) | UNVERIFIED — could not diff; no functional delta confirmed. |

## Forks classified as identical / outdated / unrelated

| Category | Count | Disposition |
|----------|-------|-------------|
| `identical` (ahead=0, behind=0) | 2 | No functional delta; not adopted |
| `behind` (ahead=0, behind>0) | ~313 | Outdated mirrors; not adopted |
| `ERROR_404` (inaccessible) | 26 | Recorded as `UNVERIFIED`; no conclusions invented |
| `compare_failed` (no common ancestor) | 1 (X-Team-Robots) | Recorded as `UNVERIFIED` |

## Features adopted into HyperDNSx

| Feature | Origin fork | Status in HyperDNSx | Tests required |
|---------|-------------|---------------------|----------------|
| Multi-sink output (console + JSONL + file) | scournomad | Planned (P1) — single writer, no per-result open/close | Test writer lifecycle, flush/close, no fd leak |
| CSV/TSV output | lucasolslund | Planned (P1) | Test CSV schema, encoding, concurrent writers |
| Strict wildcard / publicsuffix base domain | Pitrat-wav, assakafpix | Upstream already has `-auto-wildcard`; publicsuffix extraction noted for future wildcard hardening | Test wildcard precision/recall on fixtures |
| Importable Go library | Kadrick | Already done (`libs/dnsx`) | `libs/dnsx` unit tests pass |

## Features NOT adopted (with rationale)

| Feature | Origin fork | Rationale |
|---------|-------------|-----------|
| Hardcoded China resolver defaults | eexp | Violates "no geographic resolver hardcoded as universal policy"; resolvers come from config/user, not hardcoded |
| Per-result file open with `defer` in worker | scournomad | Risk of file-descriptor leak; HyperDNSx uses a stable writer lifecycle |

## Coverage statement

Fork coverage is **356/356 validated** (every fork in the inventory was fetched, compared against the upstream `dev` snapshot `eab63bea`, and recorded). Forks that could not be fetched (404) or compared (no common ancestor) are explicitly listed above as `UNVERIFIED`; no functional delta is assumed for them.
