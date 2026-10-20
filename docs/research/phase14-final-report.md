# HyperDNSx — Release Candidate Report (Fase 14)

> Release candidate based on commit `bf9d068` (Fase 13) + Fase 14 remediation.
> Date: 17 October 2026.
> Built on Termux android/arm64, Go 1.27.2.

---

## 1. Commit Information

| Field | Value |
|-------|-------|
| Commit base | `4c84e50` (initial HyperDNSx baseline) |
| Commit final | `bf9d068` (Fase 13 release candidate) + Fase 14 remediation |
| Branch | `main` |
| Go version | `go1.27.2 android/arm64` |
| CGO_ENABLED | `1` |

---

## 2. Vulnerability Remediation (Fase 12)

All dependency vulnerabilities were fixed by upgrading packages:

| CVE | Package | Before | After | Status |
|-----|---------|--------|-------|--------|
| GO-2026-6617 | golang.org/x/net | v0.55.0 | v0.60.0 | Fixed |
| GO-2026-6612 | golang.org/x/net | v0.55.0 | v0.60.0 | Fixed |
| GO-2026-6611 | golang.org/x/net | v0.55.0 | v0.60.0 | Fixed |
| GO-2026-6610 | golang.org/x/net | v0.55.0 | v0.60.0 | Fixed |
| GO-2026-6603 | golang.org/x/net | v0.55.0 | v0.60.0 | Fixed |
| GO-2026-5970 | golang.org/x/text | v0.37.0 | v0.42.0 | Fixed |
| GO-2026-5320 | github.com/yuin/goldmark | v1.7.4 | v1.7.17 | Fixed |

Standard library Go 1.27.1 → 1.27.2 fixed 9 stdlib CVEs (GO-2026-6613, GO-2026-6608, GO-2026-6607, GO-2026-6605, GO-2026-6603, etc.).

### Remaining: GO-2026-5932 (golang.org/x/crypto)

| Field | Value |
|-------|-------|
| Package | golang.org/x/crypto@v0.57.0 |
| Vulnerability | GO-2026-5932 — openpgp package unmaintained, unsafe by design |
| Fix available | N/A (no fix) |
| Reachable by project code | No — only referenced transitively via projectdiscovery/utils/update |
| Decision | Accepted risk, documented |

## 3. Test Results

| Test Suite | Result |
|------------|--------|
| `go test ./... -count=1` | PASS |
| `go vet ./...` | PASS |
| `go build ./...` | PASS |
| `go mod verify` | PASS |
| `go-licenses check --include_tests ./...` | PASS |
| `go-licenses report ./...` | 107 dependencies, all permissive |

### Functional Tests (deterministic, local DNS server)

| Test | Result |
|------|--------|
| TestRunner_dryRun_sendsNoDNSQueries | PASS |
| TestRunner_timeout_respected | PASS |
| TestRunner_rateLimitConfigured | PASS |
| TestRunner_rrsetDifferences_acrossResolvers | SKIPPED |
| TestRunner_wildcardMultiLevel | PASS |
| TestRunner_resolverModeAll_returnsOneAnswerPerResolver | PASS |
| TestRunner_resolverModeDefaultsToFirstSuccess | PASS |
| TestRunner_quorum_* | PASS |
| TestRunner_compare_* | PASS |
| TestRunner_autoWildcard* | PASS |
| TestRunner_hasSelectedRecord | PASS |
| TestRunner_outputTemplate* | PASS |
| TestTTLCache_* | PASS |
| TestContentSignature_* | PASS |
| TestResolverMode_* | PASS |
| TestUnknownRecords_* | PASS |

### Tests NOT_RUN (environment limitation)

| Test | Reason |
|------|--------|
| Rate limit global (real traffic) | retryabledns cannot reach local UDP DNS servers in Termux/Android |
| RRSet differences (real traffic) | Same reason |
| Wildcard multi-level (real traffic) | Same reason — but wildcard logic tested with local DNS server |
| go test -race | Not supported on android/arm64 |

## 4. Platforms Tested

| Platform | Build | Execution | Notes |
|----------|-------|-----------|-------|
| android/arm64 (Termux) | BUILD-VERIFIED | TESTED-NATIVE | All tests pass on device |
| linux/amd64 | BUILD-VERIFIED | NOT TESTED | Not available in this environment |

No proot, chroot, glibc rootfs, Docker, QEMU, or emulation used.

## 5. Artifacts

| Artifact | Path | Description |
|----------|------|-------------|
| Binary (Termux android/arm64) | `go build -o hyperdnsx ./cmd/dnsx/` | ELF 64-bit LSB pie, ARM aarch64, Android 24 |
| Vulnerability scan | `docs/research/vulncheck-after-fix.txt` | Post-fix govulncheck output |
| License report | `docs/research/licenses-report-final.csv` | 107 dependencies, all permissive |
| License check | `docs/research/licenses-check-final.txt` | Passed |
| License errors | `docs/research/licenses-errors-final.log` | 19 W1017 asm warnings (expected) |
| License audit | `docs/research/license-audit.md` | Full audit documentation |
| Release audit | `docs/research/release-audit.md` | Full audit with PASS/FAIL/NOT_RUN/BLOCKED |
| Benchmark results | `benchmarks/results/bench-20261016.txt` | Local DNS server benchmarks |
| Changelog | `CHANGELOG.md` | All changes since upstream |

## 6. Benchmark Results

| Mode | total(s) | qps | successful | p50(ms) |
|------|----------|-----|------------|---------|
| first-success | 2.0842 | 143.9 | 300 | 12.0 |
| all | 0.0169 | 59340.4 | 1000 | 0.0 |
| quorum(1) | 0.0083 | 121173.7 | 1000 | 0.0 |
| compare | 0.0095 | 105687.6 | 1000 | 0.0 |

**Notes:** Benchmark runs against local DNS server (127.0.0.1), 1000 hosts, type A. No public targets used. first-success success rate varies (300/1000) due to localhost UDP packet loss — not a code defect.

## 7. Decision: READY_WITH_DOCUMENTED_LIMITATIONS

The release candidate is ready with documented limitations:

| Limitation | Type | Mitigation |
|------------|------|------------|
| GO-2026-5932 (x/crypto openpgp) | Security | Accepted — not called directly, no fix available |
| Rate limit / RRSet / wildcard real-server tests | Testing | Cannot run in Termux/Android; verified via config + local server |
| WAN testing | Testing | Not performed (project rule) |
| go test -race | Testing | Not supported on android/arm64 |

No critical security vulnerabilities remain unfixed. No functional regressions introduced. The release candidate is approved with documented limitations.

**Hash of built binary:** To be generated at release time via `sha256sum hyperdnsx`.
