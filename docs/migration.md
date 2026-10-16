# HyperDNSx — Migration guide

This document describes every difference between HyperDNSx and upstream
`projectdiscovery/dnsx`. Nothing is removed silently; all upstream features
are preserved.

## Preserved (upstream-compatible)

| Item | Status |
|------|--------|
| CLI flags `-l/-d/-w`, `-a/-aaaa/-cname/-ns/-txt/-srv/-ptr/-mx/-soa/-any/-axfr/-caa` | Preserved |
| `-recon`/`-all`, `-exclude-type` | Preserved |
| `-resp`/`-resp-only`, `-rcode`, `-response-type-filter` | Preserved |
| `-cdn`, `-asn` | Preserved |
| `-threads`/`-t`, `-rate-limit`/`-rl` | Preserved |
| `-output`/`-o`, `-json`/`-j`, `-omit-raw`/`-or`, `-output-template`/`-ot` | Preserved |
| `-health-check`, `-silent`, `-verbose`, `-debug`/`-raw`, `-stats`, `-version`, `-no-color` | Preserved |
| `-retry`, `-hostsfile`/`-hf`, `-trace`, `-trace-max-recursion`, `-resume`, `-stream`, `-timeout` | Preserved |
| `-auth`, `-resolver`/`-r`, `-wildcard-threshold`/`-wt`, `-auto-wildcard`, `-wildcard-domain`/`-wd`, `-proxy` | Preserved |
| JSONL output, templates, stdin/stdout piping | Preserved |
| Wildcard (`-auto-wildcard`, `-wildcard-domain`) | Preserved |
| Resume (`-resume`, `resume.cfg`) | Preserved |
| Transports UDP/TCP/DoH/DoT | Preserved |
| MIT license, copyright and notices | Preserved |

## Added (HyperDNSx only)

| Flag | Default | Description |
|------|---------|-------------|
| `-resolver-mode` | `first-success` | `first-success\|all\|quorum\|compare` |
| `-quorum` | `0` | Minimum votes for consensus (`0` = majority of resolvers) |
| `-resolver-rate-limit` | `-1` | Per-resolver requests/second (`-1` = disabled) |
| `-cache-enabled` | `false` | Enable TTL-bounded answer cache |
| `-cache-ttl` | `0s` | Default cache TTL when response TTL is zero |
| `-dry-run` | `false` | Show planned queries/estimates; no DNS traffic |
| `-scope-file` | `""` | Allowed DNS suffixes; rejects out-of-scope candidates |

## JSONL changes

| Field | Upstream | HyperDNSx |
|-------|----------|-----------|
| `resolver` | (via DNSData) | Explicit, per-answer |
| `protocol` | (via DNSData) | Explicit, per-answer |
| `latency` | (via DNSData) | Explicit, per-answer |
| `attempts` | (via DNSData) | Explicit, per-answer |
| `unknown` | — | Records of types not parsed by the library (e.g. DNSSEC/DANE) |

## Behavior changes (explicit)

- **Default resolver mode is `first-success`** — identical to upstream, so
  existing pipelines produce the same output.
- **`all`/`quorum`/`compare` emit one answer per resolver** with provenance.
  In `first-success` mode there is exactly one answer (unchanged).
- **`quorum` requires `consensus_reached=true`** only when some signature
  reaches the vote minimum; otherwise every answer is still preserved with
  `consensus_reached=false`.
- **`compare` never picks a winner** and lists divergences independent of
  response arrival order.
- **Scope enforcement**: when `-scope-file` is set, candidates outside the
  allowed DNS suffixes are rejected and logged (debug). Suffix-spoofing
  (`example.com.attacker.invalid`) is rejected for scope `example.com`.
- **Dry-run**: `-dry-run` exits after printing the plan; no DNS traffic,
  update check, or authentication occurs.
- **Cache**: when `-cache-enabled`, answers are stored per
  `host/type/resolver` with their TTL and served from cache on repeat queries
  within the same process.

## Nothing removed

No upstream flag, output field, transport, or behavior has been removed. Every
HyperDNSx addition is opt-in and defaults to the upstream behavior when not
specified.
