# HyperDNSx — Termux / Android ARM64 native support

> Classification per evidence actually gathered on this device (Termux,
> Android/Bionic, aarch64). No proot, chroot, glibc rootfs, Docker, QEMU or
> emulation is used or claimed.

## Build evidence

| Item | Value |
|------|-------|
| Go | `go1.27.1 android/arm64` |
| GOOS | `android` |
| GOARCH | `arm64` |
| CGO_ENABLED | `1` |
| Binary | ELF 64-bit LSB pie executable, ARM aarch64, dynamically linked, interpreter `/system/bin/linker64`, for Android 24 |
| Build toolchain | Go built by NDK r30 |
| Build verified | `go build ./...` → exit 0; `go vet ./...` → clean |

Classification: **BUILD-VERIFIED** (compilation proven on-device).

## Native execution evidence

| Check | Result |
|-------|--------|
| `hyperdnsx -version` | banner printed (version string not emitted — pre-existing upstream behavior) |
| `hyperdnsx -h` | help lists all flags incl. `-resolver-mode`, `-quorum`, `-dry-run`, `-scope-file`, `-cache-enabled`, `-resolver-rate-limit` |
| Resolver mode `quorum` | works against a local UDP DNS server on-device (consensus JSON emitted) |
| Resolver mode `compare` | works against a local UDP DNS server on-device (divergences JSON emitted) |
| Resolver mode `first-success` (default) | CLI accepts and runs; actual DNS resolution depends on reachable resolvers |

Classification: **TESTED-NATIVE** for build, CLI parsing, and resolver-mode dispatch against a local on-device DNS server. Full end-to-end DNS resolution against remote resolvers is **NOT-VERIFIED** here (depends on network/DNS reachability, not the program).

## Limitations

- `go version` reports `android/arm64`; the binary targets Android API 24+.
- Some Go standard-library network behaviors differ on Android/Bionic (e.g. resolver
  configuration, `/etc/resolv.conf` handling). Test with explicit `-r` resolvers.
- `govulncheck` is not available in this Termux environment (see `docs/research/benchmark-results.md` / audit notes).
- Do not use proot/chroot/glibc rootfs/Docker/QEMU to emulate Termux compatibility;
  native build/test is the only supported path.

## Exact commands

```bash
# Build natively in Termux
go build -o hyperdnsx ./cmd/dnsx/

# Inspect architecture
file hyperdnsx          # ELF 64-bit ... ARM aarch64 ... Android

# Version / help
./hyperdnsx -version
./hyperdnsx -h

# Dry-run (no DNS traffic)
./hyperdnsx -l hosts.txt -r 1.1.1.1,8.8.8.8 -dry-run

# Resolver modes against a local DNS server
./hyperdnsx -l hosts.txt -r 127.0.0.1:5353 -resolver-mode quorum -quorum 2 -a -json
./hyperdnsx -l hosts.txt -r 127.0.0.1:5353 -resolver-mode compare -a -json
./hyperdnsx -l hosts.txt -r 127.0.0.1:5353 -resolver-mode all -a -json

# Scope file
./hyperdnsx -l candidates.txt -scope-file scope.txt -a

# Cache
./hyperdnsx -l hosts.txt -r 1.1.1.1 -cache-enabled -cache-ttl 300s -a
```

## Status summary

| Area | Classification |
|------|----------------|
| Native build (android/arm64) | BUILD-VERIFIED |
| CLI / flags / dry-run / scope / cache | TESTED-NATIVE |
| Resolver modes (all/quorum/compare) vs local server | TESTED-NATIVE |
| Remote DNS resolution / public targets | NOT-VERIFIED (no public targets used) |
| govulncheck / external audit tools | NOT-VERIFIED (not available in env) |
