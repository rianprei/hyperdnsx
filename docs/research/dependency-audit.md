# HyperDNSx — Dependency & license audit

> Evidence gathered on-device (Termux android/arm64, go1.27.1).

## Tool availability (honest status)

| Tool | Available? | Notes |
|------|------------|-------|
| `go vet ./...` | YES | clean (exit 0) |
| `go test ./...` | YES | all tests pass |
| `go mod verify` | YES | "all modules verified" |
| `go mod tidy -diff` | YES | no changes required (go.mod/go.sum consistent) |
| `govulncheck` | NO | not installed in this Termux environment; NOT-VERIFIED |
| `git diff --check` | YES | clean (exit 0) |

`govulncheck` could not be installed in this environment. Vulnerability
scanning is therefore NOT-VERIFIED via govulncheck. No known-vulnerability
audit was performed with that tool. Dependencies are pinned in go.mod/go.sum
and verified via `go mod verify`.

## Direct dependencies (projectdiscovery ecosystem)

| Module | Version | Role |
|--------|---------|------|
| github.com/projectdiscovery/dnsx | (this module, forked from upstream) | core DNS client |
| github.com/projectdiscovery/retryabledns | v1.0.116 | DNS resolution client |
| github.com/projectdiscovery/miekg/dns (via retryabledns) | — | DNS packet parsing |
| github.com/projectdiscovery/asnmap | v1.1.1 | ASN enrichment |
| github.com/projectdiscovery/cdncheck | v1.2.50 | CDN detection |
| github.com/projectdiscovery/clistats | v0.1.6 | statistics |
| github.com/projectdiscovery/goconfig | v0.0.1 | resume config |
| github.com/projectdiscovery/goflags | v0.1.76 | CLI flag parsing |
| github.com/projectdiscovery/gologger | v1.1.72 | logging |
| github.com/projectdiscovery/hmap | v0.0.102 | hybrid memory/disk map |
| github.com/projectdiscovery/mapcidr | v1.1.97 | CIDR/ASN input expansion |
| github.com/projectdiscovery/ratelimit | v0.0.88 | rate limiting |
| github.com/projectdiscovery/utils | v0.11.2 | file/ip/string utils |

Full transitive set is pinned in `go.mod` / `go.sum`.

## License posture

- **Upstream dnsx**: MIT. Preserved in `LICENSE.md`; copyright and notices kept.
- **puredns, shuffledns, MassDNS**: GPL-3.0 (per their repositories). No code
  was copied or ported from these projects. Only high-level, publicly described
  ideas (wildcard detection, bulk resolution, validation against trusted
  resolvers) were reimplemented independently. This avoids GPL→MIT license
  contamination.
- **Other dependencies**: licenses NOT individually audited in this session.
  NOT-VERIFIED. Before any release, run a license scan (e.g. go-licenses) and
  confirm compatibility with the MIT license.

## Security posture (documented)

- No hardcoded secrets; secrets via environment only (PDCP API key optional).
- No TLS verification disabled by default (DoH/DoT via retryabledns validates certs).
- No telemetry or cloud auth by default; PDCP update check is opt-in/disableable.
- Scope file (`-scope-file`) prevents out-of-scope resolution, including
  suffix-spoofing (`example.com.attacker.invalid` is rejected for scope
  `example.com`).
- Dry-run (`-dry-run`) performs no DNS traffic, no update check, no auth.
- Per-resolver rate limit (`-resolver-rate-limit`) limits blast radius per endpoint.

## NOT-VERIFIED items

- govulncheck vulnerability scan (tool not available).
- Per-dependency license scan (no license scanner run).
- External security audit / penetration test.
