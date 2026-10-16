# Upstream Baseline — projectdiscovery/dnsx

> Fonte: código real lido em `~/hyperdnsx` (cmd/dnsx/dnsx.go, internal/runner/options.go, internal/runner/runner.go, internal/runner/output_template.go, internal/runner/healthcheck.go, internal/runner/resume.go, internal/runner/wildcard.go).
> Data: 16/10/2026.
> Esta base descreve o que o upstream **realmente faz**, não o que o README promete.

## 1. Identidade do módulo

| Campo | Valor |
|-------|-------|
| Módulo Go | `github.com/projectdiscovery/dnsx` |
| Versão Go declarada (go.mod) | `go 1.25.0` |
| Go testado neste ambiente | `go1.27.1 android/arm64` (Termux nativo, CGO_ENABLED=1) |
| Binário do produto (HyperDNSx) | `hyperdnsx` (não substituir silenciosamente o `dnsx` instalado) |
| Biblioteca DNS | `github.com/projectdiscovery/retryabledns v1.0.116` |
| Parser DNS | `github.com/miekg/dns v1.1.72` |
| Configuração | `github.com/projectdiscovery/goconfig v0.0.1` |
| Flags/CLI | `github.com/projectdiscovery/goflags v0.1.76` |
| Rate limiting | `github.com/projectdiscovery/ratelimit v0.0.88` |
| ASN/CDN | `github.com/projectdiscovery/asnmap v1.1.1`, `github.com/projectdiscovery/cdncheck v1.2.50` |
| Estatísticas | `github.com/projectdiscovery/clistats v0.1.6` |
| Templates de saída | `github.com/valyala/fasttemplate v1.2.2` |
| Armazenamento híbrido (memória/disco) | `github.com/projectdiscovery/hmap v0.0.102` |

**Build verificado neste ambiente:** `go build ./...` → exit 0 (compila nativamente em Termux android/arm64).
**Vet verificado:** `go vet ./...` → exit 0.

---

## 2. Flags e defaults (extraídos de `internal/runner/options.go`)

### Grupo: Input

| Flag | Alias | Default | Semântica |
|------|-------|---------|-----------|
| `-list` | `-l` | `""` | Lista de subdomínios/hosts a resolver (arquivo, vírgula ou stdin) |
| `-domain` | `-d` | `""` | Lista de domínios para brute-force (arquivo, vírgula ou stdin) |
| `-wordlist` | `-w` | `""` | Lista de palavras para brute-force (arquivo, vírgula ou stdin) |

Regras de validação: `-l` não pode ser usado com `-d` ou `-w`; `-w` exige `-d`; `-d` exige `-w`; stdin só em um destes flags.

### Grupo: Query

| Flag | Alias | Default | Semântica |
|------|-------|---------|-----------|
| `-a` | — | `false` | Query registro A (default quando nenhum tipo explícito) |
| `-aaaa` | — | `false` | Query AAAA |
| `-cname` | — | `false` | Query CNAME |
| `-ns` | — | `false` | Query NS |
| `-txt` | — | `false` | Query TXT |
| `-srv` | — | `false` | Query SRV |
| `-ptr` | — | `false` | Query PTR |
| `-mx` | — | `false` | Query MX |
| `-soa` | — | `false` | Query SOA |
| `-any` | — | `false` | Query ANY |
| `-axfr` | — | `false` | Query AXFR (transferência de zona) |
| `-caa` | — | `false` | Query CAA |
| `-recon` | `-all` | `false` | Query todos os tipos (a,aaaa,cname,ns,txt,srv,ptr,mx,soa,axfr,caa); ativa `-resp` |
| `-exclude-type` | `-e` | vazio | Tipo a excluir (a,aaaa,cname,ns,txt,srv,ptr,mx,soa,axfr,caa,any) |

Nota: se nenhum tipo for explícito, o default é `A`. O tipo `ANY` não é suportado pela `retryabledns` para filtragem combinada.

### Grupo: Filter

| Flag | Alias | Default | Semântica |
|------|-------|---------|-----------|
| `-resp` | `-re` | `false` | Exibir resposta DNS |
| `-resp-only` | `-ro` | `false` | Exibir apenas resposta DNS (sem o host) |
| `-rcode` | `-rc` | `""` | Filtrar por código de status DNS (noerror,formerr,servfail,nxdomain,...) |
| `-response-type-filter` | `-rtf` | `""` | Retornar entradas sem registros para os tipos especificados (a,cname,...) |

`resp` e `resp-only` são mutuamente exclusivos.

### Grupo: Probe

| Flag | Alias | Default | Semântica |
|------|-------|---------|-----------|
| `-cdn` | — | `false` | Exibir nome CDN |
| `-asn` | — | `false` | Exibir informação ASN do host |

### Grupo: Rate-limit

| Flag | Alias | Default | Semântica |
|------|-------|---------|-----------|
| `-threads` | `-t` | `100` | Número de threads concorrentes |
| `-rate-limit` | `-rl` | `-1` | Requisições DNS por segundo (`-1` = desabilitado/ilimitado) |

### Grupo: Update

| Flag | Alias | Default | Semântica |
|------|-------|---------|-----------|
| `-update` | `-up` | — | Atualizar dnsx para a versão mais recente |
| `-disable-update-check` | `-duc` | `false` | Desabilitar verificação automática de atualização |

### Grupo: Output

| Flag | Alias | Default | Semântica |
|------|-------|---------|-----------|
| `-output` | `-o` | `""` | Arquivo para escrita de saída |
| `-json` | `-j` | `false` | Saída JSONL (linhas JSON) |
| `-omit-raw` | `-or` | `false` | Omitir resposta DNS bruta do JSONL |
| `-output-template` | `-ot` | `""` | Template customizado (ex: `-ot '{{host}} {{a}}'`) |

`output-template` não pode ser usado com `-json` ou `-raw`.

### Grupo: Debug

| Flag | Alias | Default | Semântica |
|------|-------|---------|-----------|
| `-health-check` | `-hc` | `false` | Executar diagnóstico e sair |
| `-silent` | — | `false` | Exibir apenas resultados na saída |
| `-verbose` | `-v` | `false` | Exibir saída verbosa |
| `-debug` | `-raw` | `false` | Exibir resposta DNS bruta |
| `-stats` | — | `false` | Exibir estatísticas da varredura |
| `-version` | — | `false` | Exibir versão |
| `-no-color` | `-nc` | `false` | Desabilitar cor na saída |

### Grupo: Optimization

| Flag | Alias | Default | Semântica |
|------|-------|---------|-----------|
| `-retry` | — | `2` | Número de tentativas DNS (mínimo 1) |
| `-hostsfile` | `-hf` | `false` | Usar arquivo hosts do sistema |
| `-trace` | — | `false` | Executar trace DNS |
| `-trace-max-recursion` | — | `255` | Recursão máxima para trace DNS |
| `-resume` | — | `false` | Retomar varredura existente |
| `-stream` | — | `false` | Modo stream (desabilita wordlist, wildcard, stats, stop/resume) |
| `-timeout` | — | `3s` | Tempo máximo de espera por consulta DNS |

### Grupo: Configurations

| Flag | Alias | Default | Semântica |
|------|-------|---------|-----------|
| `-auth` | — | `true` | Configurar chave API PDCP |
| `-resolver` | `-r` | `""` | Lista de resolvedores (arquivo ou vírgula) |
| `-wildcard-threshold` | `-wt` | `5` | Limite de filtro wildcard |
| `-auto-wildcard` | — | `false` | Detectar wildcard automaticamente para filtragem |
| `-wildcard-domain` | `-wd` | `""` | Domínio para filtragem manual wildcard (mutuamente exclusivo com `-auto-wildcard`) |
| `-proxy` | — | `""` | Proxy (ex: `socks5://127.0.0.1:8080`) |

**Resolvers padrão:** nenhum hardcoded no código lido. Se `-r` não informado, `BaseResolvers` fica vazio (usa resolvers do sistema via `retryabledns`).

---

## 3. Tipos DNS suportados

Tipos implementados (mapeados para `miekg/dns`):
`A`, `AAAA`, `CNAME`, `PTR`, `SOA`, `ANY`, `TXT`, `SRV`, `MX`, `NS`, `CAA`, `AXFR`.

**O que `-recon` consulta efetivamente:** todos os tipos acima, incluindo `AXFR`. O `ANY` é excluído da filtragem quando combinado com outros tipos (não suportado pela `retryabledns`).

**Tipos NÃO presentes no baseline:** `HTTPS`, `SVCB` (RFC 9460), `NAPTR`, `TLSA`, `SSHFP`, `DS`, `DNSKEY`, `RRSIG`, `NSEC`, `NSEC3`, `DNAME`, `CERT`. Tipos desconhecidos não são preservados via RFC 3597 neste baseline.

---

## 4. Transportes

| Transporte | Suporte no baseline |
|------------|---------------------|
| UDP | Sim (via `retryabledns`) |
| TCP | Sim (via `retryabledns`, fallback quando truncado) |
| DoH (DNS over HTTPS) | Sim (via `retryabledns`) |
| DoT (DNS over TLS) | Sim (via `retryabledns`) |
| DoQ (DNS over QUIC / RFC 9250) | Não |
| Proxy SOCKS5 para DNS | Sim (via `-proxy`, documentado em releases anteriores) |

**TLS:** validação de certificado é feita pela `retryabledns`; não há flag para desabilitar TLS por padrão no código lido.

---

## 5. Comportamento de resolvedores

- `retryabledns.CheckInternalIPs = true` é definido no `New()` (evita consultas a IPs internos).
- Se `-r` for arquivo, carrega linhas do arquivo; se for string, separa por vírgula.
- Modo de consulta padrão: **primeira resposta de sucesso** (`first-success`). Não há modo `all`, `quorum` ou `compare` no baseline.
- Retry: `-retry` (default 2), timeout `-timeout` (default 3s).
- Rate limit: global apenas (`-rate-limit`), sem limite por resolvedor. Default desabilitado (`-1`).
- Não há health check periódico de resolvedores separado da varredura (exceto o flag `-health-check` que executa diagnóstico pontual).
- Não há cache TTL-aware nem checkpoint/resume versionado além do arquivo `resume.cfg` simples.

---

## 6. Formato de saída

### Texto (default)
- Cada linha: `host` (ou `host [TYPE] response` com `-resp`, ou `host [RCODE]` com `-rcode`).
- Com `-cdn` / `-asn`: acrescido ` [CDNName]` e/ou ` ASxxx Org Country`.
- Com `-raw`: imprime resposta DNS bruta.
- Com `-template`: aplica `fasttemplate` `{{host}} {{a}}` etc.

### JSONL (`-json`)
- Emite objeto `dnsx.ResponseData` serializado por `dnsData.JSON()`.
- Campos: host, timestamp, question type, answers (A, AAAA, CNAME, MX, NS, TXT, SRV, SOA, CAA, PTR), status code, query time, raw (opcional), CDN, ASN, AXFR, trace.
- `-omit-raw` remove o campo raw.

### Saída simultânea
- Quando `-output` é informado, escreve no arquivo **e** no stdout (sequencial, via `bufio.Writer` + `gologger.Silent`).
- stdout e stderr são separados: estatísticas vão para stderr; resultados vão para stdout.

### Limitação observada no lifecycle de output
- Arquivo de saída aberto uma vez com `os.O_APPEND\|os.O_CREATE\|os.O_WRONLY`, bufferizado com `bufio.Writer`, fechado via `defer` no `HandleOutput`.
- Não há flush/checkpoint periódico durante a execução longa; interrupção pode perder dados no buffer não-flushado.

---

## 7. Resume / checkpoint

- Arquivo: `resume.cfg` (constante `DefaultResumeFile`).
- Armazena `currentIndex` e `current` (host atual).
- `ShouldSaveResume()` sempre retorna `true`; `ShouldLoadResume()` só se `-resume` e arquivo existir.
- Na retomada, itens com índice `<= current.Index` são pulados.
- Não há versão de sessão; retomada após mudança de flags pode produzir estado inconsistente.
- Não há detecção de corrupção do arquivo de resume.

---

## 8. Wildcard

| Flag | Alias | Default |
|------|-------|---------|
| `-auto-wildcard` | — | `false` |
| `-wildcard-domain` | `-wd` | `""` |
| `-wildcard-threshold` | `-wt` | `5` |

- `-auto-wildcard` e `-wildcard-domain` são mutuamente exclusivos.
- Wildcard não suportado em modo `-stream`.
- Detecção usa resolvers A e AAAAA da `wildcard` package do `retryabledns`; marcas hosts como wildcard em `wildcards` map.
- Filtro manual por domínio: armazena DNS data no hybrid map e filtra após resolução.

---

## 9. Gaps observados no baseline (matéria-prima para HyperDNSx)

| Área | Gap | Impacto |
|------|-----|---------|
| Multi-resolver | Apenas `first-success`; não há `all`/`quorum`/`compare` | Não há provenance de resposta por resolver |
| Rate limiting | Global apenas; sem limite por resolvedor | Risco de throttling por resolver |
| Cache | Nenhum cache TTL-aware nem on-disk | Reconsulta repetida desperdiça banda |
| DNSSEC | Apenas consulta de tipos DNSSEC (se houver); nenhuma validação criptográfica | Não há distinção consulta vs validação |
| RR types | Sem HTTPS/SVCB, NAPTR, TLSA, etc. | Cobertura limitada |
| Output lifecycle | Sem flush/checkpoint periódico; buffer pode perder dados | Risco de perda em interrupção |
| Resume | Sem versionamento de sessão | Retomada insegura após mudança de flags |
| Wildcard | Detecção básica; sem comparação entre resolvers confiáveis | Falsos positivos possíveis |
| Health check | Diagnóstico pontual, não perfil contínuo | Sem score explicável por resolver |
| Termux/Android | Build nativo compila, mas não há documentação de compatibilidade | Não comprovado funcional em device |

---

## 9. Limites e validações de CLI

| Validação | Comportamento |
|-----------|---------------|
| Retries = 0 | Fatal: "retries must be at least 1" |
| `-l` com `-d` ou `-w` | Fatal |
| `-w` sem `-d` (ou vice-versa) | Fatal |
| stdin em múltiplos flags | Fatal (exceto stream) |
| `-stream` com wordlist/domains/hosts/resume/wildcard/stats | Fatal |
| `-auto-wildcard` com `-wildcard-domain` | Fatal |
| `output-template` com `-json` ou `-raw` | Fatal |
| `resp` com `resp-only` | Fatal |
| Wildcard domain inválido | Fatal: "invalid wildcard-domain value" |

---

*Este documento é o baseline funcional do upstream. O próximo passo é produzir `docs/research/gap-matrix.md` comparando com os forks inventariados em `docs/research/forks.tsv`.*
