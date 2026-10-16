# HyperDNSx — Architecture

## Design principles

- **Preserve what works:** upstream `dnsx` CLI, flags, JSONL, templates,
  stdin/stdout, wildcard, resume, and transports are kept unchanged unless a
  proven gap justifies a change.
- **Small, testable interfaces:** no empty "enterprise" abstractions or
  factories without a real consumer.
- **TDD:** tests before/during implementation; no regressions accepted.
- **No unverified claims:** every performance, coverage, compatibility, or
  security statement has reproducible evidence.

## Responsibility separation (per FASE 3 of the mission)

```
1. CLI/config            validate arguments, aliases, config file, schema
2. Input pipeline        file/stdin/domains/wordlists/CIDR/ASN; normalization
3. Scheduler             global/per-resolver/per-domain rate limits; bounded
                         concurrency; cancellation; dry-run; max queries
4. Resolver pool         transport, health/latency, retries+jitter, failover,
                         breaker, profile, provenance
5. Query engine          DNS types, RRset interpretation, RCODE/flags/TTL,
                         resolver/protocol/attempts, raw option
6. Validation layer      cross-resolver consistency, wildcard, poisoned
                         results; explanation recorded per decision
7. Cache/state           TTL-aware cache, negative caching, versioned and
                         idempotent checkpoint/resume
8. Output sinks          console/stderr separated; JSONL/JSON/CSV/TSV;
                         stable-writer lifecycle; I/O errors propagated
9. Public API            importable Go package, stable structs, result/progress
                         callbacks; REST/MCP wrappers separate from core
```

## Resolver dispatch modes

| Mode | Behavior | Default | Provenance |
|------|----------|---------|------------|
| `first-success` | Query resolvers until the first valid answer (upstream behavior) | **Yes** | Yes (single answer) |
| `all` | Query every configured resolver; preserve each answer with provenance | No | Yes (per resolver) |
| `quorum` | Classify answers by content consensus; requires configurable vote minimum | No | Yes (vote counts) |
| `compare` | Show every answer without picking a winner; list divergences | No | Yes (all answers) |

Consensus in `quorum` ignores TTL differences alone (per DNSSEC guidance:
different TTLs are not content divergence). When no signature reaches the
quorum, `consensus_reached=false` and every answer is still preserved.

## Data flow

```
Input (file/stdin/domain+wordlist/CIDR/ASN)
    │
    ▼
Input pipeline ─────► Hybrid map (dedup, bounded memory)
    │
    ▼
Scheduler (rate limits, concurrency, cancellation)
    │
    ▼
Resolver pool (transport, retries+jitter, health, per-resolver rate limit)
    │
    ▼
Query engine ───────────────────────► Cache (TTL-aware, per resolver+type)
    │                                       ▲
    │                                       │
    ▼                                       │
Validation layer (cross-resolver, wildcard) │
    │                                       │
    ▼                                       │
Output sinks (console/stderr separated)     │
    ▲                                       │
    └───────────────────────────────────────┘
```

## Trust boundaries

- User input (hosts, domains, wordlist, resolvers, scope file) is validated
  at entry (allowlist patterns, path traversal rejection, scope suffix check).
- Resolver answers are treated as untrusted: no raw resp in JSON by default,
  no credential logging, TLS validation not disabled.
- Cache keys include `host/type/resolver` so one resolver's answer is never
  attributed to another.

## Platform

- Native build: `GOOS=android GOARCH=arm64` (Termux), `GOOS=linux GOARCH=amd64/arm64`.
- No external binary required by default (no MassDNS).
- No proot/chroot/glibc rootfs/Docker/QEMU used to fake Termux compatibility.

## Security-relevant design notes

- Secrets (PDCP API key, resolver credentials) are never logged or included in
  error messages.
- `-dry-run` performs no network traffic, no update check, no authentication.
- `-scope-file` rejects out-of-scope candidates including suffix-spoofing
  (`example.com.attacker.invalid` is rejected for scope `example.com`).
- Output files are opened once per lifecycle, buffered, and flushed/closed
  exactly once (no per-result open/close, no fd leaks).
