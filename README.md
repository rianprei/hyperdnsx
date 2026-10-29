# HyperDNSx

A fast, reliable, and auditable evolution of [`projectdiscovery/dnsx`](https://github.com/projectdiscovery/dnsx)
for batch DNS queries, host resolution, subdomain brute-forcing, wildcard
filtering, and custom resolver support. Built natively for Linux and
**Android/Termux ARM64**, with measurable improvements over upstream:
per-resolver provenance, resolver dispatch modes (`first-success`/`all`/
`quorum`/`compare`), TTL-bounded cache, scope enforcement, dry-run, and
per-resolver rate limiting.

> **Engineering goal:** be defensibly better than upstream where tests and
> benchmarks demonstrate it. "Better" is not claimed as fact; it is claimed
> only where reproducible evidence exists (see `docs/research/benchmark-results.md`).

## Compatibility with upstream `dnsx`

HyperDNSx preserves the upstream CLI flags, aliases, JSONL output, templates,
stdin/stdout piping, wildcard, resume, and transport behavior
(UDP/TCP/DoH/DoT). Existing `dnsx` pipelines keep working. The default mode is
`first-success`, identical to upstream.

**Differences from upstream are explicit and documented:**
- `-resolver-mode {first-success,all,quorum,compare}` — new (upstream is
  first-success only).
- `-quorum N` — minimum votes for consensus in quorum mode.
- `-resolver-rate-limit N` — per-resolver requests/second limit.
- `-cache-enabled` / `-cache-ttl` — TTL-bounded answer cache.
- `-dry-run` — show planned queries and estimates without sending DNS traffic.
- `-scope-file PATH` — allowed DNS suffixes; out-of-scope candidates rejected
  (suffix-spoofing safe).
- JSONL output now includes `resolver`, `protocol`, `latency`, `attempts`, and
  `unknown` (records of types not parsed by the underlying library, e.g.
  DNSSEC/DANE collection).
- DNSSEC/DANE record **collection** (`-dnskey`, `-ds`, `-rrsig`, `-nsec`,
  `-nsec3`, `-tlsa`, `-sshfp`) — collection only, no cryptographic validation
  claimed.

## Installation

HyperDNSx requires **Go 1.25+**. Build natively (no proot, chroot, glibc
rootfs, Docker, or QEMU):

```sh
git clone https://github.com/projectdiscovery/dnsx  # or your fork
cd dnsx
go build -o hyperdnsx ./cmd/dnsx/
./hyperdnsx -h
```

Termux/Android ARM64: `go build` compiles natively on-device
(`GOOS=android GOARCH=arm64`). See `docs/termux-android.md` for exact
commands and evidence classification.

## Quick start

Resolve a list of hosts (upstream-compatible, first-success mode):

```sh
echo "example.com" | ./hyperdnsx -silent -a
example.com [A] [93.184.216.34]
```

Query multiple record types with JSONL output:

```sh
echo "example.com" | ./hyperdnsx -silent -json -a -aaaa -cname -mx -ns
```

Use custom resolvers:

```sh
./hyperdnsx -l hosts.txt -r 1.1.1.1,8.8.8.8 -a -silent
```

Query every configured resolver and keep per-resolver provenance (issue #884):

```sh
./hyperdnsx -l hosts.txt -r 1.1.1.1,8.8.8.8,9.9.9.9 \
  -resolver-mode all -a -json -silent
```

Consensus across resolvers (quorum, majority by default):

```sh
./hyperdnsx -l hosts.txt -r 1.1.1.1,8.8.8.8 -resolver-mode quorum -quorum 2 -a -json -silent
```

Show every answer and flag divergences between resolvers:

```sh
./hyperdnsx -l hosts.txt -r 1.1.1.1,8.8.8.8 -resolver-mode compare -a -json -silent
```

Dry-run (no DNS traffic sent):

```sh
./hyperdnsx -l hosts.txt -d -w wordlist.txt -r 1.1.1.1 -dry-run
```

Enforce a scope (suffix-spoofing safe):

```sh
echo "scope.example.com" > scope.txt
./hyperdnsx -l candidates.txt -scope-file scope.txt -a -silent
```

## Features

- Batch DNS queries (A, AAAA, CNAME, PTR, NS, MX, TXT, SRV, SOA, CAA, ANY,
  AXFR) plus DNSSEC/DANE collection (DNSKEY, DS, RRSIG, NSEC, NSEC3, TLSA,
  SSHFP).
- DNS resolution and subdomain brute-forcing (`-d`/`-domain` + `-w`/`-wordlist`).
- Custom resolver input (`-r`/`-resolver`, file or comma-separated).
- Multiple resolver transports (UDP/TCP/DoH/DoT via retryabledns).
- stdin/stdout support for Unix pipelines.
- Automatic wildcard handling (`-auto-wildcard`/`-aw`, `-wildcard-domain`/`-wd`).
- Resolver dispatch modes: `first-success` (default, upstream-compatible),
  `all`, `quorum`, `compare`.
- Per-resolver provenance (resolver, protocol, latency, attempts) in JSONL.
- TTL-bounded answer cache (`-cache-enabled`, `-cache-ttl`).
- Scope enforcement (`-scope-file`), dry-run (`-dry-run`),
  per-resolver rate limit (`-resolver-rate-limit`).
- Importable Go library (`libs/dnsx`).

## Deep-research (`hyperx`)

`hyperx` is a second binary that fuses the DNS engine with the
[`httpx`](https://github.com/projectdiscovery/httpx) HTTP probing engine into a
single pass: raw targets → resolved addresses → fingerprinted HTTP endpoints,
emitted as one unified JSONL record (DNS provenance + HTTP fingerprint + TLS).
The pure-DNS `dnsx` binary is unchanged.

```bash
make build-hyperx
cat domains.txt | ./hyperx -tls-grab -json -silent
```

See [`docs/deep-research.md`](docs/deep-research.md) for flags, the unified
record schema, the architecture, and the httpx fork-feature integration roadmap.

## Research documents

| Document | Contents |
|----------|----------|
| `docs/research/upstream-baseline.md` | Functional baseline extracted from upstream code |
| `docs/research/gap-matrix.md` | Gaps vs. upstream and forks, with decisions |
| `docs/research/forks.tsv` | 356 direct forks catalogued |
| `docs/research/fork-compare.jsonl` | Comparison of every fork vs. upstream |
| `docs/research/fork-feature-matrix.md` | Functional deltas, adopted / not adopted |
| `docs/research/fork-audit-summary.md` | Audit summary and UNVERIFIED forks |
| `docs/research/benchmark-results.md` | Reproducible benchmarks (local fixtures only) |
| `docs/research/dependency-audit.md` | Dependencies, licenses, tool availability |
| `docs/termux-android.md` | Native Termux/ARM64 build & execution evidence |
| `docs/deep-research.md` | `hyperx` deep-research pipeline + fork integration roadmap |
| `docs/research/httpx/fork-audit-summary.md` | httpx fork audit (1262 forks) summary |
| `docs/research/httpx/fork-feature-matrix.md` | httpx fork functional deltas, adopted / planned |

## License

MIT. See `LICENSE.md`. Upstream `dnsx` is MIT; copyright and notices are
preserved. No GPL-3.0 code (puredns, shuffledns, MassDNS) was copied — only
publicly described ideas were reimplemented independently.

## Security

No secrets are hardcoded or logged. No telemetry or cloud authentication is
sent by default. TLS validation is not disabled. See `SECURITY.md`.
