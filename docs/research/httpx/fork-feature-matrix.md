# HyperDNSx — httpx Fork feature matrix

> Derived from `fork-compare.jsonl` (1262 unique, validated forks) and deep-diff
> inspection of the 27 forks with a plausible functional delta. Raw evidence per
> fork in `fork-analysis/<slug>.diff`. Flag/file names below were extracted from
> the actual saved diffs, not from commit messages alone.

## Inventory summary

| Metric | Value |
|--------|-------|
| Total unique direct forks | 1262 |
| Behind upstream (no delta) | 1014 |
| Inaccessible (404 / error) | 171 |
| Identical to snapshot | 1 |
| Diverged (`ahead > 0`) | 76 |
| Functional delta deep-diffed | 27 |

Upstream snapshot: `dev` @ `668a37ca`.

## Tier 1 — core deep-research capabilities (adopt)

| Fork | ahead | Feature | Evidence (flags / new files) |
|------|-------|---------|------------------------------|
| CarlosCommits/httpx | 110 | Headless JS-runtime tech detection (React, Vite, TanStack) + CPE | `-tech-detect-headless/-tdh`, `-chrome-bin`, `-chrome-settle-timeout`, `-chrome-window-size`, `-headless-browser-native-headers`, `-browser-recovery`; `runner/headless.go`, `runner/cpe.go`, `runner/wappalyzer.go` |
| kno-com/httpx | 44 | Response dedup via simhash + LSH band index; pluggable auth | `-simhash-threshold`; `common/authprovider/authx/{basic,bearer,cookies,headers,query}_auth.go`, `strategy.go` |
| l0ners/httpx | 14 | WAF detection + improved CDN strategy | `-waf`; `common/httpx/waf.go`, `common/httpx/cdn.go` |
| rabb1t-dev/httpx | 2 | TLS/JA3 Chrome impersonation, browser headers, brotli, HTTP/2-over-proxy CONNECT | `-tls-impersonate-chrome`, `-browser-headers`; `common/httpx/proxy_http2.go`, `common/httpx/encodings.go` |
| tracertea/httpx | 1 | Recog fingerprint-DB matching (http + html) | `-recog`; `common/httpx/response.go` |
| xiaoyaochen/httpx | 9 | Custom Wappalyzer fingerprint file + host-collision + result storage | `-tech-detect-file`, `-db`, `-dboutput`; `common/wap/*`, `common/db/{mongodb,storeinfo}.go` |
| bokostmu/httpx | 12 | Enriched JSON output + protocol-start control | `-body-in-json`, `-errors-in-json`, `-start-protocol`, `-dont-retry-protocol`; `common/httpx/{response,tls,option}.go` |
| jsifuentes/httpx | 3 | Virtual-host / custom-Host-header probing per IP | `-custom-hosts/-ch`; `runner/types.go` |
| securifera/httpx | 1 | CPE derivation from Wappalyzer (version-substituted) to JSON | `runner/cpe.go` |

## Tier 2 — enrichment & output (adopt)

| Fork | ahead | Feature | Evidence |
|------|-------|---------|----------|
| yabeow/httpx | 3 | Extract JS files from responses | `-output-js/-ojs` |
| nyxgeek/httpx | 3 | Copyright string search/extraction | `-copyright`; `common/httpx/copyright.go` |
| b3r1ch/httpx | 7 | Custom-regex extraction with subexpression capture | `-custom-regex`; `common/httpx/customregex.go` |
| kiriknik/httpx | 2 | SSL certificate grab into output | `-ssl-grab`; `common/httpx/tls.go` |
| P1kAju/httpx | 4 | HTML + CSV report output | `-html`, `-csv` |
| atoz-project/httpx | 1 | CustomIP targets + domain-aware CDN checks | `common/httpx/cdn.go` |
| dkarv/httpx | 1 | Browser Wappalyzer runtime detection (overlaps CarlosCommits) | `-tech-detect-runtime/-tdr`; `runner/headless.go` |
| yhy0/httpx | 3 | Extract/normalize helpers (customextract, httputilz, csp, title, filter) | `common/customextract/*`, `common/httputilz/normalize.go` |

## Tier 3 — robustness & bug fixes (adopt)

| Fork | ahead | Fix | Evidence |
|------|-------|-----|----------|
| hetbhatt2807/httpx | 11 | HTTP/2 → HTTP/1.1 fallback control + custom CheckRetry | `-dhf`; `common/httpx/retry.go` |
| innerpeace609/httpx | 1 | Strictly honor `http11` by disabling HTTP/2 transport fallback | `common/httpx/httpx.go` (12-line delta) |
| ohmygod20260203/httpx | 1 | Mutex for concurrent `lastCompletedIndex` (resume race) | `runner/resume.go` |
| chudamax/httpx | 1 | `--path-as-is` preserve raw URL path encoding | `-path-as-is` |
| imfht/httpx | 4 | Full-page screenshot + silent raw-request suppression | `-screenshot-fullpage/-ssf`, `-irreq` |
| zp857/httpx | 1 | Library option `WithDialerHistory` | `common/httpx/option.go` |

## Evaluate with caution (opt-in, documented)

| Fork | ahead | Item | Caution |
|------|-------|------|---------|
| d11testing/httpx | 2 | Fix: prevent SSRF via crafted TLS cert CommonName during recursive probe | **Security hardening** — verify whether upstream already guards this before porting; adopt as a fix, not a flag |
| sahruldotid/httpx | 6 | log4j/JNDI injection probe via headers (DNS callback) | Offensive active-probe; adopt only as explicit opt-in module (`-dns4j`), off by default, with scope enforcement |

## Not adopted (rationale)

| Item | Fork(s) | Rationale |
|------|---------|-----------|
| CI/workflow files | CarlosCommits `.github/*`, testaorga, vshekhda, cjstang, Rsnehith | Repo maintenance, not product features |
| Docker/K8s deploy | mariandppb, realArcherL | Deployment glue |
| Junk/suspicious uploads | Jayyeeeeeeeeeeeeeee (`PsExec.exe`), biseYpueza (spam `.md`), apt-grid (`.drawio`) | Abandoned/junk forks |
| Dependency bumps, banner removal, module renames | Nihility981, Explorer1092, pelamx (rename half), ByteHackr, beike2020, and other empty-signal diverged forks | No functional delta |
| ayanrajpoot10/httpx multi-value headers | ayanrajpoot10 | Net diff vs merge-base = 0 lines (already upstreamed or reverted); nothing to port |

## Coverage statement

Fork coverage is **1262/1262 validated** — every fork in the inventory was
fetched, compared against upstream `dev` snapshot `668a37ca`, and recorded.
The 27 functional-delta forks were deep-diffed with net diffs saved as
evidence. Forks that could not be fetched (404) are explicitly listed as
`UNVERIFIED`; no functional delta is assumed for them.
