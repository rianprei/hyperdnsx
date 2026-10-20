# Changelog

## [Unreleased] - HyperDNSx

### Added
- **Resolver dispatch modes**: `-resolver-mode` with `first-success` (default), `all`, `quorum`, `compare` for flexible DNS query dispatching
- **Per-resolver provenance**: JSONL output now includes `resolver`, `protocol`, `latency`, `attempts` fields for each answer
- **Per-resolver rate limiting**: `-resolver-rate-limit` flag to limit DNS requests per second per resolver
- **TTL-bounded answer cache**: `-cache-enabled` and `-cache-ttl` flags for caching DNS answers with TTL-based expiration
- **Dry-run mode**: `-dry-run` flag to show planned queries and estimates without sending any DNS traffic
- **Scope enforcement**: `-scope-file` flag for allowed DNS suffixes; rejects out-of-scope candidates including suffix-spoofing (e.g., `example.com.attacker.invalid` is rejected for scope `example.com`)
- **DNSSEC/DANE record collection**: `-dnskey`, `-ds`, `-rrsig`, `-nsec`, `-nsec3`, `-tlsa`, `-sshfp` flags for collecting DNSSEC/DANE records (collection only, no cryptographic validation)
- **Unknown record types**: `unknown` field in JSONL output for DNS record types not parsed by the underlying library (e.g., DNSSEC/DANE records)

### Changed
- Go toolchain upgraded from 1.27.1 to 1.27.2 (addresses multiple CVEs in net/http, crypto/tls, mime/multipart)
- `golang.org/x/net` upgraded from v0.55.0 to v0.60.0 (fixes 5 CVEs)
- `golang.org/x/text` upgraded from v0.37.0 to v0.42.0 (fixes GO-2026-5970)
- `github.com/yuin/goldmark` upgraded from v1.7.4 to v1.7.17 (fixes GO-2026-5320)
- `golang.org/x/crypto` upgraded from v0.52.0 to v0.57.0

### Deprecated
- None

### Removed
- None

### Fixed
- None

### Security
- Addressed 9 standard library CVEs via Go 1.27.2 upgrade (GO-2026-6617, GO-2026-6613, GO-2026-6612, GO-2026-6611, GO-2026-6610, GO-2026-6608, GO-2026-6607, GO-2026-6605, GO-2026-6603)
- Addressed GO-2026-5970 (golang.org/x/text) via upgrade to v0.42.0
- Addressed GO-2026-5320 (github.com/yuin/goldmark) via upgrade to v1.7.17
- Remaining: GO-2026-5932 (golang.org/x/crypto openpgp, no fix available, not directly called by project code) - documented as accepted risk
