# HyperDNSx — CLI reference

All flags are compatible with upstream `dnsx` unless marked **HyperDNSx**.
The default resolver mode is `first-success` (upstream-compatible).

## Input

| Flag | Alias | Default | Description |
|------|-------|---------|-------------|
| `-list` | `-l` | `""` | List of subdomains/hosts to resolve (file, comma-separated, or stdin) |
| `-domain` | `-d` | `""` | List of domains to brute-force (file, comma-separated, or stdin) |
| `-wordlist` | `-w` | `""` | List of words to brute-force (file, comma-separated, or stdin) |

Validation: `-l` cannot be combined with `-d`/`-w`; `-w` requires `-d`; `-d`
requires `-w`; stdin can be set for only one of these flags.

## Query

| Flag | Alias | Default | Description |
|------|-------|---------|-------------|
| `-a` | | `false` | Query A record (default when no explicit type) |
| `-aaaa` | | `false` | Query AAAA record |
| `-cname` | | `false` | Query CNAME record |
| `-ns` | | `false` | Query NS record |
| `-txt` | | `false` | Query TXT record |
| `-srv` | | `false` | Query SRV record |
| `-ptr` | | `false` | Query PTR record |
| `-mx` | | `false` | Query MX record |
| `-soa` | | `false` | Query SOA record |
| `-any` | | `false` | Query ANY record |
| `-axfr` | | `false` | Query AXFR (zone transfer) |
| `-caa` | | `false` | Query CAA record |
| `-recon` | `-all` | `false` | Query all record types (a,aaaa,cname,ns,txt,srv,ptr,mx,soa,axfr,caa) |
| `-exclude-type` | `-e` | (none) | DNS query type to exclude |
| `-dnskey` | | `false`**H** | Query DNSKEY record (DNSSEC collection only) |
| `-ds` | | `false`**H** | Query DS record (DNSSEC collection only) |
| `-rrsig` | | `false`**H** | Query RRSIG record (DNSSEC collection only) |
| `-nsec` | | `false`**H** | Query NSEC record (DNSSEC collection only) |
| `-nsec3` | | `false`**H** | Query NSEC3 record (DNSSEC collection only) |
| `-tlsa` | | `false`**H** | Query TLSA record (DANE collection only) |
| `-sshfp` | | `false`**H** | Query SSHFP record (collection only) |

**H** = HyperDNSx addition. DNSSEC/DANE flags are **collection only**; no
cryptographic validation is claimed.

## Filter

| Flag | Alias | Default | Description |
|------|-------|---------|-------------|
| `-resp` | `-re` | `false` | Display DNS response |
| `-resp-only` | `-ro` | `false` | Display DNS response only |
| `-rcode` | `-rc` | `""` | Filter result by DNS status code |
| `-response-type-filter` | `-rtf` | `""` | Return entries with no records for the specified types |

`-resp` and `-resp-only` are mutually exclusive.

## Probe

| Flag | Alias | Default | Description |
|------|-------|---------|-------------|
| `-cdn` | | `false` | Display CDN name |
| `-asn` | | `false` | Display host ASN information |

## Rate-limit

| Flag | Alias | Default | Description |
|------|-------|---------|-------------|
| `-threads` | `-t` | `100` | Number of concurrent threads |
| `-rate-limit` | `-rl` | `-1` | DNS requests/second (global; `-1` = disabled) |
| `-resolver-rate-limit` | | `-1`**H** | Per-resolver requests/second limit (`-1` = disabled) |

**H** = HyperDNSx addition.

## Output

| Flag | Alias | Default | Description |
|------|-------|---------|-------------|
| `-output` | `-o` | `""` | File to write output |
| `-json` | `-j` | `false` | Write output in JSONL format |
| `-omit-raw` | `-or` | `false` | Omit raw DNS response from JSONL |
| `-output-template` | `-ot` | `""` | Custom output template |

`-output-template` cannot be used with `-json` or `-raw`.

JSONL includes `resolver`, `protocol`, `latency`, `attempts`, and `unknown`
(fields added by HyperDNSx).

## Debug

| Flag | Alias | Default | Description |
|------|-------|---------|-------------|
| `-health-check` | `-hc` | `false` | Run diagnostic check and exit |
| `-silent` | | `false` | Display only results in the output |
| `-verbose` | `-v` | `false` | Display verbose output |
| `-debug` | `-raw` | `false` | Display raw DNS response |
| `-stats` | | `false` | Display scan statistics |
| `-version` | | `false` | Display version |
| `-no-color` | `-nc` | `false` | Disable color in output |

## Optimization

| Flag | Alias | Default | Description |
|------|-------|---------|-------------|
| `-retry` | | `2` | Number of DNS attempts (must be ≥ 1) |
| `-hostsfile` | `-hf` | `false` | Use system host file |
| `-trace` | | `false` | Perform DNS tracing |
| `-trace-max-recursion` | | `255` | Max recursion for DNS trace |
| `-resume` | | `false` | Resume existing scan |
| `-stream` | | `false` | Stream mode (disables wordlist, wildcard, stats, stop/resume) |
| `-timeout` | | `3s` | Maximum time to wait for a DNS query |

## Configurations

| Flag | Alias | Default | Description |
|------|-------|---------|-------------|
| `-auth` | | `true` | Configure PDCP API key |
| `-resolver` | `-r` | `""` | List of resolvers (file or comma-separated) |
| `-resolver-mode` | | `""`**H** | Resolver dispatch mode: `first-success` (default), `all`, `quorum`, `compare` |
| `-quorum` | | `0`**H** | Minimum votes for consensus in quorum mode (`0` = majority of resolvers) |
| `-wildcard-threshold` | `-wt` | `5` | Wildcard filter threshold |
| `-auto-wildcard` | | `false` | Automatically detect wildcard domains for filtering |
| `-wildcard-domain` | `-wd` | `""` | Domain for manual wildcard filtering (mutually exclusive with `-auto-wildcard`) |
| `-proxy` | | `""` | Proxy (e.g. `socks5://127.0.0.1:8080`) |
| `-scope-file` | | `""`**H** | File with allowed DNS suffixes; out-of-scope candidates rejected |

**H** = HyperDNSx addition.

## HyperDNSx additions

| Flag | Alias | Default | Description |
|------|-------|---------|-------------|
| `-dry-run` | | `false` | Show planned queries and estimates; no DNS traffic, no update check, no auth |
| `-cache-enabled` | | `false` | Enable TTL-bounded answer cache |
| `-cache-ttl` | | `0s` | Default cache TTL when response TTL is zero (`0` = use response TTL only) |
| `-resolver-mode` | | `first-success` | `first-success\|all\|quorum\|compare` |
| `-quorum` | | `0` | Minimum votes for consensus (`0` = majority) |
| `-resolver-rate-limit` | | `-1` | Per-resolver requests/second (`-1` = disabled) |
| `-scope-file` | | `""` | Allowed DNS suffixes file |

## Output contract

- Console and stderr are separated: statistics go to stderr; results go to
  stdout (consumable in pipes).
- `-output` writes to a file **and** stdout.
- JSONL is deterministic, UTF-8, correctly escaped; no log messages are mixed
  into `-silent` stdout.
- Logs go to stderr only; stdout remains clean for pipes.

## Examples

```sh
# Resolve a host list (upstream-compatible)
hyperdnsx -l hosts.txt -r 1.1.1.1 -a -silent

# All resolvers with provenance (JSONL)
hyperdnsx -l hosts.txt -r 1.1.1.1,8.8.8.8 -resolver-mode all -a -json -silent

# Quorum consensus
hyperdnsx -l hosts.txt -r 1.1.1.1,8.8.8.8 -resolver-mode quorum -quorum 2 -a -json -silent

# Compare resolvers (divergences listed)
hyperdnsx -l hosts.txt -r 1.1.1.1,8.8.8.8 -resolver-mode compare -a -json -silent

# Dry-run (no traffic)
hyperdnsx -d example.com -w wordlist.txt -r 1.1.1.1 -dry-run

# Scope enforcement
hyperdnsx -l candidates.txt -scope-file scope.txt -a -silent
```
