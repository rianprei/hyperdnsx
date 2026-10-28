# HyperDNSx Deep-Research (`hyperx`)

`hyperx` fuses the HyperDNSx DNS resolution engine (`libs/dnsx`) with the
[`httpx`](https://github.com/projectdiscovery/httpx) HTTP probing engine into a
single pass: raw targets → resolved addresses → fingerprinted HTTP endpoints,
emitted as one unified JSONL record per probed endpoint. The pure-DNS `dnsx`
binary is left untouched; `hyperx` is a separate binary (`cmd/hyperx`).

## Why a second binary

`dnsx` stays a focused DNS toolkit. `hyperx` is the research pipeline that
depends on it. Existing `dnsx` pipelines keep working unchanged; `hyperx` adds
the HTTP stage on top without bloating the DNS CLI surface.

## Build

```bash
make build-hyperx       # produces ./hyperx
# or
go build -o hyperx ./cmd/hyperx
```

## Usage

```bash
# resolve + probe a list, unified JSONL with TLS data
cat domains.txt | hyperx -tls-grab -json -silent

# probe URLs directly (skip DNS stage)
cat urls.txt | hyperx -no-resolve -json

# follow redirects, impersonate a Chrome ClientHello, write to file
hyperx -l targets.txt -follow-redirects -tls-impersonate chrome -o out.jsonl -json
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-l, -list` | stdin | targets (file, comma-separated, or stdin) |
| `-schemes` | `https,http` | schemes to probe when a target carries none |
| `-follow-redirects` | off | follow HTTP redirects |
| `-tls-grab` | off | grab TLS certificate + JA3/JARM data |
| `-tls-impersonate` | — | ClientHello profile to impersonate (e.g. `chrome`) |
| `-random-agent` | off | random User-Agent per request |
| `-no-resolve` | off | skip the DNS stage, probe targets directly |
| `-r, -resolvers` | public set | DNS resolvers (file or comma-separated) |
| `-t, -threads` | 25 | concurrent workers |
| `-timeout` | 10 | per-request timeout (seconds) |
| `-retries` | 2 | retry count |
| `-o, -output` | stdout | output file |
| `-j, -json` | off | emit JSONL |
| `-silent` | off | results only, no status logs |

### Unified record

```json
{
  "input": "example.com",
  "host": "example.com",
  "a": ["172.66.147.243"],
  "aaaa": ["2606:4700:10::ac42:93f3"],
  "dns_resolver": "8.8.8.8:53",
  "url": "https://example.com",
  "scheme": "https", "port": "443",
  "status_code": 200,
  "title": "Example Domain",
  "webserver": "cloudflare",
  "content_type": "text/html",
  "content_length": 577,
  "response_time": "126ms",
  "tls": {
    "subject_cn": "example.com",
    "issuer_cn": "Cloudflare TLS Issuing ECC CA 3",
    "sans": ["example.com", "*.example.com"],
    "version": "tls13",
    "cipher": "TLS_AES_128_GCM_SHA256"
  }
}
```

## Architecture

```
inputs ──► parseTarget ──► [DNS stage: libs/dnsx.QueryMultiple] ──► base Result
                                                                      │
                                                   per scheme/port ───┤
                                                                      ▼
                                      [HTTP stage: httpx.Do] ──► enriched Result ──► emit (JSONL)
```

- `internal/research/target.go` — target parsing and scheme/port fan-out.
- `internal/research/pipeline.go` — concurrent orchestrator (worker pool,
  serialized emit), DNS stage via `QueryMultiple` (first-success).
- `internal/research/probe.go` — httpx engine wrapper; one request → one
  enriched record; TLS projection (subject/issuer/SANs/JA3/JARM).
- `internal/research/result.go` — the unified `Result` type and JSON rendering.

The DNS stage uses `QueryMultiple` (first-success) rather than
`QueryMultipleAll`: the per-resolver variant depends on
`retryabledns.QueryMultipleWithResolver`, which returns empty answers in the
Android/Termux build (see audit notes). First-success is the project's default
mode and returns the combined A/AAAA/CNAME set reliably.

## Fork-feature integration roadmap

The httpx fork audit (`docs/research/httpx/fork-feature-matrix.md`) recovered
capabilities from the 1262-fork ecosystem. The fused engine already surfaces
several for free; the rest are planned as additive modules on top of the
pipeline. Each row maps a surveyed fork feature to its integration point here.

| Capability | Origin fork(s) | Status | Integration point |
|------------|----------------|--------|--------------------|
| DNS + HTTP fusion, unified JSONL | (new) | **Done** | `internal/research` |
| TLS grab (subject/issuer/SANs/validity) | bokostmu, kiriknik | **Done** | `probe.go` `mapTLS` |
| JA3 / JARM fingerprint fields | rabb1t-dev | **Done** (surfaced when engine computes) | `TLSInfo` |
| Chrome ClientHello impersonation | rabb1t-dev | **Done** (`-tls-impersonate`) | `ProbeConfig.TLSImpersonate` |
| Random User-Agent | upstream | **Done** (`-random-agent`) | `ProbeConfig.RandomAgent` |
| Headless JS-runtime tech detect (React/Vite/TanStack) | CarlosCommits, dkarv | Planned | new `probe` option + chromedp sink |
| Response dedup (simhash + LSH) | kno-com | Planned | post-emit `internal/research/dedup` |
| Pluggable auth providers | kno-com | Planned | `ProbeConfig.Auth` |
| CDN/WAF-range classification | l0ners, atoz-project | **Done** (`-cdn`, reuses httpx cdncheck) | `probe.go` `CheckCDN` |
| Active WAF fingerprinting | l0ners | Planned | `probe.go` enrich stage |
| Recog fingerprint-DB matching | tracertea | Planned | tech-detect sink |
| Custom Wappalyzer fingerprint file | xiaoyaochen | Planned | tech-detect option |
| vhost / custom-Host probing | jsifuentes | Planned | target fan-out |
| CPE (→CVE) derivation | securifera, CarlosCommits | Planned | enrich stage |
| JS-file / email / custom-regex extraction | yabeow, nyxgeek, b3r1ch | **Done** (`-extract-js`, `-extract-emails`, `-extract-regex`) | `extract.go` |
| HTML / CSV report output | P1kAju | Planned | output sinks |
| Result storage (MongoDB) | xiaoyaochen | Planned (opt-in) | output sink |
| HTTP/2 fallback control | hetbhatt2807, innerpeace609 | Planned | `ProbeConfig` |
| SSRF guard (cert CommonName) | d11testing | Planned (hardening) | probe network policy |
| log4j/JNDI probe | sahruldotid | Planned (opt-in, off by default) | explicit scan module |

## Tests

```bash
go test -cover ./internal/research/
```

Hermetic: DNS-disabled pipeline tests run against an in-process `httptest`
server; TLS mapping and target parsing are pure unit tests. Coverage ≥ 80%.
`-race` is unavailable on android/arm64 (upstream Go limitation).
