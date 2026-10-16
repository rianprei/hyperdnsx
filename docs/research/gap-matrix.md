# Gap Matrix — HyperDNSx vs Upstream e Forks

> Fonte: `docs/research/upstream-baseline.md` + `docs/research/forks.tsv` + `docs/research/fork-compare.jsonl` + dossiê (`00-dossie-origem.md`).
> Data: 16/10/2026.
> Colunas: Recurso | Baseline Upstream | Forks com delta | Concorrentes | Decisão HyperDNSx | Prioridade

## 1. Multi-resolver / provenance

| Recurso | Baseline Upstream | Forks com delta | Concorrentes | Decisão HyperDNSx | Prioridade |
|---------|-------------------|-----------------|--------------|-------------------|------------|
| Modo `first-success` | Sim (único modo) | Nenhum fork altera | puredns valida com resolvers confiáveis | Manter como default (compatibilidade) | P1 |
| Modo `all` | Não | Nenhum | puredns coleta múltiplas respostas | Implementar: consulta todos os resolvers, preserva provenance | P1 |
| Modo `quorum` | Não | Nenhum | — | Implementar: consenso configurável, não esconde divergência | P1 |
| Modo `compare` | Não | Nenhum | — | Implementar: mostra divergência por RRset/RCODE/TTL/resolver | P1 |
| Provenance por resposta (resolver, protocolo, latência, tentativa, RCODE, TTL) | Não | Nenhum | puredns registra por que aceitou/filtrou | Implementar em JSONL e texto | P1 |
| `--unique-only` (dedupe apresentação, não apaga provenance) | Não | Nenhum | — | Implementar | P1 |

Referência: issue upstream #884 (consultar todos os resolvedores) fechada como "not planned".

## 2. Resolver health / profiling

| Recurso | Baseline Upstream | Forks com delta | Concorrentes | Decisão HyperDNSx | Prioridade |
|---------|-------------------|-----------------|--------------|-------------------|------------|
| `-health-check` | Sim (diagnóstico pontual) | Nenhum | — | Manter; expandir para perfil contínuo | P1 |
| Perfil por resolver (RTT, timeouts, RCODE, erro TLS, inconsistência) | Não | Nenhum | — | Implementar; exportar resultados separado dos alvos | P1 |
| Retry/backoff com jitter | Retry fixo (-retry) | Nenhum | — | Implementar backoff exponencial com jitter | P1 |
| Limite global e por endpoint | Global apenas (-rate-limit) | Nenhum | — | Implementar limite por resolvedor | P1 |
| Circuit breaker temporário | Não | Nenhum | — | Implementar para resolvers que falham | P1 |
| Evitar retry storms | Não | Nenhum | — | Implementar | P1 |
| Nenhum resolver geográfico hardcoded como default | Nenhum hardcoded (usa sistema) | eexp/dnsx hardcoda resolvers da China (114.114.114.114, 180.76.76.76, etc.) | — | Não hardcodear; perfis via arquivo de configuração opt-in | P1 |

## 3. RR type coverage

| Recurso | Baseline Upstream | Forks com delta | Concorrentes | Decisão HyperDNSx | Prioridade |
|---------|-------------------|-----------------|--------------|-------------------|------------|
| A, AAAA, CNAME, PTR, NS, MX, TXT, SRV, SOA, CAA, ANY, AXFR | Sim | Nenhum adiciona tipos | puredns/shuffledns focam A/AAAA | Manter todos | P0 |
| HTTPS / SVCB (RFC 9460) | Não | Nenhum | — | Implementar (--type suporte numérico + nomeado) | P1 |
| NAPTR, TLSA, SSHFP, DS, DNSKEY, RRSIG, NSEC/NSEC3, DNAME, CERT | Não | Nenhum | — | Implementar gradualmente; tipos não parseados via RFC 3597 | P1 |
| Tipo numérico desconhecido (RFC 3597) | Não | Nenhum | — | Implementar representação genérica/raw | P1 |
| IANA registry versionado/auditável | Não | Nenhum | — | Implementar tabela de tipos versionada | P1 |

## 4. DNSSEC

| Recurso | Baseline Upstream | Forks com delta | Concorrentes | Decisão HyperDNSx | Prioridade |
|---------|-------------------|-----------------|--------------|-------------------|------------|
| Consultar tipos DNSSEC (DNSKEY, DS, RRSIG, NSEC/NSEC3) | Não (não na lista de flags) | Nenhum | — | Implementar coleta | P1 |
| Observar flags AD/CD | Não | Nenhum | — | Implementar | P1 |
| Validação criptográfica da cadeia (secure/insecure/bogus/indeterminate) | Não | Nenhum | — | Implementar só com biblioteca madura + teste de cadeia; inicialmente NOT_VALIDATED | P1 |
| Trust anchors + expiração de assinatura | Não | Nenhum | — | Implementar se validação completa entrar em escopo | P1 |
| Não chamar consulta de registro de "validação" | N/A | Nenhum | — | Regra de produto: separar claramente | P1 |

## 5. Wildcard e validação cruzada

| Recurso | Baseline Upstream | Forks com delta | Concorrentes | Decisão HyperDNSx | Prioridade |
|---------|-------------------|-----------------|--------------|-------------------|------------|
| `-auto-wildcard` / `-aw` | Sim (v1.3.0+) | Nenhum | puredns detecção orientada a precisão | Preservar e testar antes de reimplementar | P0 |
| `-wildcard-domain` / `-wd` | Sim | Nenhum | — | Preservar | P0 |
| Múltiplas raízes registráveis e níveis | Básico | Nenhum | puredns multinível | Melhorar cobertura | P1 |
| Distinguir wildcard global vs subárvore vs CDN/load-balanced | Não | Nenhum | puredns | Implementar | P1 |
| Comparar respostas repetidas e entre resolvers confiáveis | Não | Nenhum | puredns validação posterior | Implementar | P1 |
| Marcar divergentes como `inconsistent` (não descartar silenciosamente) | Não | Nenhum | — | Implementar | P1 |
| Evidência da classificação (probes, RRsets, TTL, resolver, motivo, confiança) | Não | Nenhum | — | Implementar | P1 |
| Budget de consultas de sondagem | Não | Nenhum | — | Implementar | P1 |

Referência: issue upstream #924 (auto wildcard).

## 6. Cache / state

| Recurso | Baseline Upstream | Forks com delta | Concorrentes | Decisão HyperDNSx | Prioridade |
|---------|-------------------|-----------------|--------------|-------------------|------------|
| Cache positivo TTL-aware por RRset | Não | Nenhum | puredns persistência separada | Implementar | P1 |
| Cache negativo (RFC 2308) | Não | Nenhum | — | Implementar (não TTL arbitrário) | P1 |
| Cache on-disk opcional com versão/schema | Não | Nenhum | — | Implementar | P2 |
| Checkpoint/resume versionado e idempotente | `resume.cfg` simples, sem versão | Nenhum | — | Implementar versão de sessão | P1 |
| Não duplicar resultados na retomada | Não (risco) | Nenhum | — | Implementar | P1 |
| Limite de memória + streaming para listas grandes | Hybrid map (memória/disco) | Nenhum | — | Manter e aperfeiçoar | P1 |

## 7. Output / integração

| Recurso | Baseline Upstream | Forks com delta | Concorrentes | Decisão HyperDNSx | Prioridade |
|---------|-------------------|-----------------|--------------|-------------------|------------|
| console / JSONL / JSON / CSV / TSV | console, JSONL (-json), template, -o | scournomad/dnsx adiciona `-json-file/-jf` (JSON em arquivo + legível no terminal) | shuffledns NDJSON | Implementar multi-sink simultâneo (console + JSONL + CSV + TSV) | P1 |
| Writers de ciclo de vida estável (abrir uma vez, flush/close garantidos) | Abre uma vez, bufio, defer close | scournomad abre/fecha por resultado (suspeita de vazar descriptors) | — | Corrigir: writers abertos uma vez, flush/checkpoint, close único no ciclo de vida | P0 |
| Escrita simultânea stdout + múltiplos arquivos | stdout + 1 arquivo (-o) | scournomad: console + jsonl | — | Implementar múltiplos sinks em pipeline único | P1 |
| JSONL versionado com campos de provenance | JSONL sem campos de provenance | Nenhum | — | Implementar (resolver, transport, RCODE, TTL, latency, attempts, validation, wildcard, raw opcional) | P1 |
| Logs só em stderr; stdout limpo para pipes | Parcial (stats em stderr) | Nenhum | — | Impor rigorosamente | P0 |
| Sem mistura stdout/stderr em modo -silent | Parcial | Nenhum | — | Impor | P0 |
| API Go importável com structs estáveis | Não (CLI acoplada) | Nenhum | — | Implementar (FASE 3) | P2 |
| REST/MCP wrapper separado do core | Não | Nenhum | — | Opcional, fora do core | P2 |

## 8. Correção e I/O (P0)

| Recurso | Baseline Upstream | Forks com delta | Concorrentes | Decisão HyperDNSx | Prioridade |
|---------|-------------------|-----------------|--------------|-------------------|------------|
| NXDOMAIN vs NODATA vs NOERROR sem answer vs REFUSED vs SERVFAIL vs timeout vs truncation vs erro local | Não (todos caem em "no response"/dropped) | Nenhum | — | Distinguir todos em saída estruturada | P0 |
| Erros de escrita reportados | Não (escrita ignora erro com `_ =`) | scournomad ignora erros de flush | — | Corrigir: propagar erros de I/O | P0 |
| Flush/checkpoint periódico | Não | scournomad adianta Close/Flush dentro do worker (risco de acumular descriptors) | — | Implementar flush periódico + close seguro | P0 |
| Arquivo não corrompido após interrupção | Não garantido | Não | — | Implementar checkpoint transacional | P0 |
| Nenhuma chamada de rede inesperada em testes | N/A | Nenhum | — | Regra de produto | P0 |
| Nenhum segredo em logs | N/A | Nenhum | — | Regra de produto | P0 |
| Não desabilitar validação TLS por padrão | Não desabilita | Nenhum | — | Manter | P0 |

## 9. Segurança e UX operacional

| Recurso | Baseline Upstream | Forks com delta | Concorrentes | Decisão HyperDNSx | Prioridade |
|---------|-------------------|-----------------|--------------|-------------------|------------|
| `--dry-run` (mostrar consultas antes de executar) | Não | Nenhum | — | Implementar | P1 |
| `--max-inputs`, `--max-queries`, `--scope-file`/allowlist | Não | Nenhum | — | Implementar | P1 |
| Rate limit conservador + feedback REFUSED/SERVFAIL/throttling | Rate limit global, sem feedback específico | Nenhum | — | Implementar | P1 |
| Redação opcional de credenciais/parâmetros sensíveis de endpoints | Não | Nenhum | — | Implementar | P1 |
| Telemetria/cloud auth opt-in, documentada, testável offline | PDCP update check automático | Nenhum | — | Tornar opt-in; documentar | P1 |

## 10. Plataforma / build

| Recurso | Baseline Upstream | Forks com delta | Concorrentes | Decisão HyperDNSx | Prioridade |
|---------|-------------------|-----------------|--------------|-------------------|------------|
| Build nativo Termux android/arm64 | Compila (go build OK neste env) | enlightment-droid (nome sugere Android, não comprovado) | — | Testar nativamente; marcar TESTED-NATIVE | P0 |
| Nenhum binário externo obrigatório (MassDNS) | Nenhum | Nenhum | shuffledns requer MassDNS externo | Manter sem dependência externa obrigatória | P0 |
| Nenhum PRoot/chroot/Docker/QEMU para fingir Termux nativo | N/A | Nenhum | — | Regra de produto | P0 |

## 11. Licença e dependências

| Recurso | Baseline Upstream | Forks com delta | Concorrentes | Decisão HyperDNSx | Prioridade |
|---------|-------------------|-----------------|--------------|-------------------|------------|
| Licença MIT preservada | Sim (MIT) | Nenhum | puredns/shuffledns/MassDNS GPL-3.0 | Manter MIT; não copiar código GPL | P0 |
| Copyright/notices preservados | Sim | Nenhum | — | Manter | P0 |
| Auditoria de licenças transitivas | Não | Nenhum | — | Implementar | P0 |
| Dependências fixas em go.mod/go.sum | Sim | Nenhum | — | Manter | P0 |
| govulncheck / advisories | Não | Nenhum | — | Implementar | P1 |

## 12. Categorização dos forks verificados (resumo)

| Categoria | Forks | Observação |
|-----------|-------|------------|
| `NO_FUNCTIONAL_DELTA` | Taeyong2366, shotarochauhan, whitehathackerpr, ourora-hub, geeknik, Veirt, ..., (maioria dos 182) | Idênticos ao upstream ou sem delta funcional verificável |
| `OUTDATED_MIRROR` | RakhithJK (behind 3), diogosilva-adyta (behind 3), esgeeks, ... | Defasados sem feature útil |
| `BUGFIX` | Nenhum confirmado | Nenhum fork com correção documentada |
| `UNIQUE_FEATURE` | scournomad/dnsx (`-json-file/-jf`) | Saída multi-destino, mas com padrão de arquivo por resultado suspeito |
| `PLATFORM_PORT` | Nenhum confirmado | enlightment-droid não comprovado como port Android |
| `BROKEN_OR_RISKY` | scournomad/dnsx (arquivo aberto por resultado com defer no worker — risco de vazar file descriptors) | Confirmar com teste antes de usar |
| `UNRELATED_OR_NOT_DIRECT` | dnsx-action, wrappers, mirrors | Excluídos do inventário de forks diretos |
| `UNVERIFIED_RATE_LIMIT` | ~174 forks não comparados (356 total - 182 comparados) | Rate limit da API bloqueou parte; registrar quais não foram avaliados |

Forks citados no dossiê com delta: `eexp/dnsx` (resolvers China + cdncheck, muito defasado), `scournomad/dnsx` (`-json-file`).

---

## Decisão de escopo imediato (P0 → P1)

1. **P0 (correção/paridade):** lifecycle de output (writers estáveis, flush/close, erro de I/O), distinguir NXDOMAIN/NODATA/REFUSED/SERVFAIL/timeout/truncation, resume versionado, wildcard preservado e testado, build nativo Termux comprovado.
2. **P1 (diferencial):** `resolver-mode=all|quorum|compare` com provenance; resolver health/profiling; RR types HTTPS/SVCB + DNSSEC coleta; cache TTL-aware; output multi-sink (CSV/TSV + JSONL versionado); dry-run/scope/max-queries.

*Este gap-matrix é derivado do código real e dos forks inventariados. Forks não verificados (rate limit) estão listados como `UNVERIFIED_RATE_LIMIT` e não foram usados como base de decisão.*
