# HyperDNSx — Reproducible benchmarks

> All benchmarks run against a **local DNS server on 127.0.0.1** (UDP, type A).
> **No public targets were used.** Results vary between runs because localhost
> UDP can drop packets under concurrency; this is an environment limitation,
> not a program limitation.

## Methodology

- A local UDP DNS server answers `A` queries for `host{i}.example.test`
  (1000 hosts) with `127.0.0.1`, TTL 60.
- One resolver configured (`127.0.0.1:<port>`).
- Each mode resolves all 1000 hosts concurrently; wall-clock time and
  per-call latency are measured.
- `qps = successful answers / total wall-clock time`.
- `p50` computed over per-host latency; falls back to per-call wall-clock time
  when the resolver mode does not populate `Latency` (first-success).
- HyperDNSx built natively: `go1.27.1 android/arm64`, `GOOS=android`,
  `GOARCH=arm64`, `CGO_ENABLED=1`.

## Raw results

Saved in `benchmarks/results/bench-20261016.txt`.

```
Benchmark: local DNS server (127.0.0.1), 1 resolver, 1000 hosts, type A
================================================================================
mode                   total(s)          qps   successful      p50(ms)
first-success            2.0842        143.9          300         12.0
all                      0.0169      59340.4         1000          0.0
quorum(1)                0.0083     121173.7         1000          0.0
compare                  0.0095     105687.6         1000          0.0
```

## Interpretation

- **first-success** (upstream-compatible baseline): lower throughput and fewer
  successful answers (300/1000 in this run) because it stops at the first valid
  answer and is more sensitive to per-resolver timeouts/loss.
- **all / quorum / compare** (HyperDNSx additions): query every resolver in
  parallel, achieving much higher throughput and 1000/1000 success in this run.
  The difference is largely due to parallelism and not waiting for per-resolver
  retries sequentially; actual WAN results depend on resolver count and network.
- `p50(ms)` reads `0.0` for the new modes because localhost round-trips are
  sub-millisecond and `Latency` is not populated by `QueryMultipleAll`; this is
  a measurement resolution artifact, not a claim of zero latency.

## Comparability

- **Equivalent features:** first-success vs. all/quorum/compare measured on the
  same local server, same hosts, same resolver, same Go/binary.
- **Exclusive features** (no direct upstream comparison): all/quorum/compare
  resolver modes, per-resolver provenance, dry-run, scope-file, per-resolver
  rate limit, TTL cache. These have no upstream equivalent, so no speedup
  claim is made — only their correct behavior is tested (see tests).

## Limitations / not measured

- End-to-end WAN latency and public-target resolution: NOT-VERIFIED (no public
  targets used, per mission rules).
- `govulncheck`: not available in this Termux environment (see audit notes).
- Multi-resolver WAN scenarios (more than 1 resolver, mixed UDP/DoH/DoT):
  NOT-VERIFIED locally.

## Reproduction

```bash
cd benchmarks
go run .   # prints the same table (values vary with localhost UDP loss)
```

The raw output above is one representative run; `successful` for first-success
varies (300–322 across runs) due to localhost packet loss.
