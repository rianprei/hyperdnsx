# HyperDNSx — Release Candidate Report

> Release candidate based on commit `7887b00` (Fase 12 remediation), built on Go 1.27.2 android/arm64.
> Date: 17 October 2026.

---

## 1. Commit Information

| Field | Value |
|-------|-------|
| Commit base | `4c84e50` (initial HyperDNSx baseline) |
| Commit final | `7887b00` (Fase 12 remediation + Fase 13 release candidate) |
| Branch | `main` |
| Go version | `go1.27.2 android/arm64` |
| CGO_ENABLED | `1` |

## 2. Security Fixes Confirmed

| CVE | Package | Version Before | Version After | Status |
|-----|---------|----------------|---------------|--------|
| GO-2026-6617 | golang.org/x/net | v0.55.0 | v0.60.0 | Fixed |
| GO-2026-6613 | net/http (stdlib) | go1.27.1 | go1.27.2 | Fixed via Go toolchain upgrade |
| GO-2026-6612 | golang.org/x/net | v0.55.0 | v0.60.0 | Fixed |
| GO-2026-6611 | golang.org/x/net | v0.55.0 | v0.60.0 | Fixed |
| GO-2026-6610 | golang.org/x/net | v0.55.0 | v0.60.0 | Fixed |
| GO-2026-6608 | net/textproto (stdlib) | go1.27.1 | go1.27.2 | Fixed via Go toolchain upgrade |
| GO-2026-6607 | crypto/tls (stdlib) | go1.27.1 | go1.27.2 | Fixed via Go toolchain upgrade |
| GO-2026-6605 | net/http (stdlib) | go1.27.1 | go1.27.2 | Fixed via Go toolchain upgrade |
| GO-2026-6603 | golang.org/x/net | v0.55.0 | v0.60.0 | Fixed |
| GO-2026-5970 | golang.org/x/text | v0.37.0 | v0.42.0 | Fixed |
| GO-2026-5932 | golang.org/x/crypto | v0.57.0 | N/A | **Accepted risk** — no fix available, not directly called by project code |
| GO-2026-5320 | github.com/yuin/goldmark | v1.7.4 | v1.7.17 | Fixed |

### Accepted Risk: GO-2026-5932

The `golang.org/x/crypto/openpgp` package is unmaintained and has known security issues. However:
- `govulncheck` reports that the project code **does not appear to call** this vulnerability directly
- The dependency is transitive (via `github.com/projectdiscovery/utils/update`)
- No fix is available (`Fixed in: N/A`)
- This is documented as an accepted risk in `docs/research/release-audit.md`

## 3. Test Results

| Test Suite | Result | Notes |
|------------|--------|-------|
| `go test ./... -count=1` | PASS | `libs/dnsx` OK, `internal/runner` OK |
| `go vet ./...` | PASS | Clean |
| `go build ./...` | PASS | Clean |
| `go mod verify` | PASS | All modules verified |
| `git diff --check` | PASS | Clean |
| `govulncheck ./...` | 1 remaining | GO-2026-5932 (openpgp, no fix available, not directly called) |
| `go-licenses check --include_tests ./...` | PASS | All licenses permissive |
| `go-licenses report ./...` | 107 deps | All permissive (MIT, BSD, Apache-2.0, ISC, MPL-2.0) |

### Tests NOT_RUN (environment limitation)

| Test | Reason |
|------|--------|
| Rate limit global/per-resolver (real traffic) | Requires a real UDP DNS server that responds; `retryabledns` cannot reach local UDP servers in Termux/Android |
| RRSet differences (real traffic) | Same reason |
| Wildcard multi-level (real traffic) | Same reason |

These are tested via configuration verification instead (resolver mode configured correctly, resolver count verified, etc.).

## 4. Platforms Tested

| Platform | Build | Execution | Notes |
|----------|-------|-----------|-------|
| android/arm64 (Termux) | BUILD-VERIFIED | TESTED-NATIVE | All tests pass on device |
| linux/amd64 | BUILD-VERIFIED | NOT TESTED | Not available in this environment |

No proot, chroot, glibc rootfs, Docker, QEMU, or emulation used.

## 5. Artifacts

| Artifact | Path | Description |
|----------|------|-------------|
| Binary (Termux android/arm64) | Built via `go build -o hyperdnsx ./cmd/dnsx/` | ELF 64-bit LSB pie, ARM aarch64, Android 24 |
| Vulnerability scan | `docs/research/vulncheck-after-fix.txt` | Post-fix govulncheck output |
| License report | `docs/research/licenses.csv` | 107 dependencies, all permissive |
| License check | `docs/research/licenses-check.txt` | Passed |
| License errors | `docs/research/licenses-errors.log` | 19 W1017 asm warnings (expected) |
| License audit | `docs/research/license-audit.md` | Full audit documentation |
| Release audit | `docs/research/release-audit.md` | Full audit with PASS/FAIL/NOT_RUN/BLOCKED |
| Benchmark results | `benchmarks/results/bench-20261016.txt` | Local DNS server benchmarks |
| Changelog | `CHANGELOG.md` | All changes since upstream |

## 6. Checksums

SHA-256 of the built binary:
```
sha256sum hyperdnsx
```

(To be generated at release time.)

## 7. Reproduction Instructions

```bash
# Clone the repository
git clone <repo-url>
cd hyperdnsx

# Checkout the release candidate commit
git checkout 7887b00

# Build natively
go build -o hyperdnsx ./cmd/dnsx/

# Verify build
go vet ./...
go test ./... -count=1
go mod verify

# Run the binary
./hyperdnsx -h
./hyperdnsx -version
```

## 8. Decision: READY_WITH_DOCUMENTED_LIMITATIONS

The release candidate is ready with the following documented limitations:

1. **GO-2026-5932 (golang.org/x/crypto openpgp)**: No fix available, not directly called by project code. Documented as accepted risk.
2. **Rate limit / RRSet differences / wildcard multi-level tests**: Cannot be executed in Termux/Android because `retryabledns` cannot reach local UDP DNS servers. Configuration is verified instead.
3. **WAN testing**: Not performed (per project rules, no public targets used).
4. **go test -race**: Not supported on android/arm64.

None of these constitute a critical security vulnerability or functional regression. The release candidate is approved with documented limitations.
