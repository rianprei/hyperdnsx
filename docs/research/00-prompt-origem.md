# Prompt principal para o OpenCode — HyperDNSx

Copie o conteúdo abaixo inteiro para o OpenCode. Ele é deliberadamente operacional: exige inspeção, implementação, testes, benchmarks e relatório; não apenas um plano.

---

## MISSÃO

Você é o engenheiro principal responsável por construir **HyperDNSx**, uma evolução própria, confiável, altamente performática e multiplataforma do `projectdiscovery/dnsx`.

**Objetivo concreto:** preservar e superar a funcionalidade do dnsx upstream e integrar de forma consciente toda contribuição útil encontrada nos forks diretos, sem degradar correção, performance, compatibilidade, segurança ou portabilidade. “Melhor do mundo” é a meta de engenharia; não escreva essa alegação como fato. Só alegue superioridade onde houver testes e benchmarks reproduzíveis.

### Princípios inegociáveis

- AUDIT → FORK INVENTORY → GAP MATRIX → DESIGN → IMPLEMENT → TEST → BENCHMARK → REVIEW → DOCUMENT → COMMIT.
- Execução direta: não pare após produzir checklist ou plano. Implemente o que couber no escopo, resolva problemas encontrados e continue até que um bloqueio externo real impeça.
- TDD; testes antes/durante da implementação quando viável; nenhuma regressão aceita.
- Leia o código e a configuração reais. Não presuma nomes de funções, flags, versões, dependências, estrutura ou resultados.
- Preserve arquitetura útil e alterações locais do usuário. Nunca apague trabalho existente nem execute reset/clean/force push destrutivo.
- Não gere falsos testes, benchmarks ou resultados. Quando não puder executar uma verificação, registre `NOT RUN`, motivo e comando que o usuário pode executar.
- Não faça scans DNS contra alvos públicos durante a fase de pesquisa/benchmark. Usar fixtures locais determinísticos ou infraestrutura explicitamente controlada/autorizada.
- O binário do produto é `hyperdnsx`; não substituir silenciosamente o `dnsx` instalado no sistema.

## FASE 0 — ESTABELECER O ESTADO DO WORKSPACE

1. Execute `pwd`, `git status --short --branch`, `git log -5 --oneline`, `git remote -v`, `go version`, `go env GOOS GOARCH CGO_ENABLED GOPATH GOMOD`.
2. Identifique se o diretório atual é HyperDNSx, um clone de `projectdiscovery/dnsx`, ou outro projeto. Se houver alterações locais, catalogue-as e preserve-as.
3. Se já houver um projeto HyperDNSx, trabalhe nele. Se for clone upstream, não renomeie destrutivamente nem sobrescreva alterações; crie um branch específico apenas depois de confirmar a árvore de trabalho e registre o commit-base. Se o diretório for outro projeto, crie uma nova pasta `hyperdnsx/` em local seguro e clone o upstream para uma pasta temporária separada. Não duplique o repositório dentro de si mesmo.
4. Fonte oficial: `https://github.com/projectdiscovery/dnsx`. Verifique a release estável e o branch padrão atual; não misture release e `dev`. Prefira a revisão mais atual que passe nos testes e fixe o SHA exato em `docs/research/upstream-baseline.md`.
5. O upstream consultado em 16/10/2026 mostrava release `v1.3.1` (31/08/2026), `go.mod` daquela tag em Go 1.25 e `go.mod` do `dev` em Go 1.26, enquanto o README dizia Go 1.21. Confirme tudo antes de usar: versão e requisitos podem mudar.

## FASE 1 — INVENTÁRIO COMPLETO DOS FORKS DIRETOS

Esta fase é obrigatória. **Não confunda busca por nome com fork real.** Use o endpoint paginado da API oficial, com todas as páginas, e grave dados em `docs/research/forks.tsv` ou `forks.json`.

Fonte canônica: `https://github.com/projectdiscovery/dnsx/forks`.

Se `gh` estiver instalado/autenticado, confira `gh auth status` (não mostre tokens). Use, por exemplo:

```bash
gh api --paginate 'repos/projectdiscovery/dnsx/forks?per_page=100&sort=stargazers' \
  --jq '.[] | [.full_name, .html_url, (.stargazers_count|tostring), (.forks_count|tostring), (.pushed_at // ""), (.default_branch // ""), (.archived|tostring), (.parent.full_name // ""), (.description // "")] | @tsv' \
  > docs/research/forks.tsv
```

Valide que:
- todas as páginas foram lidas;
- nomes de repositório não estão duplicados;
- cada registro é um fork do repositório esperado (confirme `parent.full_name`/`source.full_name`); exclua wrappers, actions, ferramentas homônimas, mirrors e projetos não relacionados da lista de forks diretos, mas mantenha-os em uma categoria separada se forem relevantes;
- o número final bate com a API no momento da consulta. Registre a data/hora e a contagem, pois ela muda.

### Comparação sistemática

Para **cada fork direto**, recolha: owner/repo, URL, default branch, stars, pushed_at, archived, SHA do upstream, SHA/branch do fork, ahead/behind, último commit, arquivos alterados e indicação de possível feature.

1. Compare `projectdiscovery/dnsx` (branch/snapshot fixado) com o branch padrão do fork. Use GitHub Compare API quando houver credencial/rate budget; prefira a API para metadados compactos e diffs de arquivos suspeitos. Use `git fetch`/`git diff` quando necessário.
2. Limite concorrência, respeite rate limits, faça retry/backoff e persista checkpoints. Se rate limit ou indisponibilidade bloquear parte do inventário, não esconda: registre quais repositórios não foram avaliados e continue no restante.
3. Não descarregue todos os forks indiscriminadamente se a API puder fornecer diffs. Para diferenças com mudanças reais, busque somente arquivos relevantes e confirme a implementação.
4. Procure alterações na CLI, resolvedores padrão, protocolo/transporte, wildcard, validação, saída, cache, desempenho, tracing, tratamento de erros, Android/Termux e build.
5. Teste a hipótese da feature e a qualidade do patch: código, testes, versão Go, dependências, documentação, licença, última atividade e regressões óbvias.
6. Marque cada fork como exatamente uma categoria principal:
   - `NO_FUNCTIONAL_DELTA`;
   - `OUTDATED_MIRROR`;
   - `BUGFIX`;
   - `UNIQUE_FEATURE`;
   - `PLATFORM_PORT`;
   - `BROKEN_OR_RISKY`;
   - `UNRELATED_OR_NOT_DIRECT`;
   - `UNVERIFIED_RATE_LIMIT`.
7. Gere `docs/research/fork-inventory.md` com todos os forks diretos e uma linha por repo. Não omita forks sem alterações: são importantes para documentar cobertura. Gere `docs/research/fork-deltas.md` com análise técnica de cada fork que tenha delta real. Inclua commits e paths como evidência.

### Alvos já descobertos nesta pesquisa (confirme do zero)

- `eexp/dnsx`: module path separado, resolvedores padrão voltados à China e dependência `github.com/eexp/cdncheck`; snapshot muito atrás do upstream.
- `scournomad/dnsx`: opção `-json-file/-jf` para gravar JSONL em arquivo mantendo saída legível; patch observado abre arquivo por resultado com `defer` dentro do worker — suspeita de acumular descriptors, confirmar com teste.
- `geeknik/dnsx`: candidato de triagem; README muito próximo do upstream, não declarar feature exclusiva sem diff.
- `enlightment-droid/dnsx` e `esgeeks/dnsx`: cópias históricas a verificar.
- `gprime31/dnsx`, `FDlucifer/dnsx`, `0xNoobie1100/dnsx`, `jdawg24/dnsx`, `0xflotus/dnsx`, `pedroevaristo960/dnsx`, `rvrsh3ll/dnsx`, `BlackNurse/dnsx`, `1234-ad/dnsx`, `gcarmonatellez15-spec/dnsx`, `tetadsf23/dnsx`, `Veirt/dnsx`: candidatos que apareceram em busca do GitHub; não presuma que todos têm deltas.

O upstream já incorporou algumas features originalmente solicitadas via issues. Por exemplo, auto-wildcard aparece no ciclo v1.3.0 e output templates também; não atribua essas features como exclusivas de fork quando já estão no upstream.

## FASE 2 — BASELINE FUNCIONAL E GAP MATRIX

Leia o README, `go.mod`, CLI, packages, testes, workflow, issues/release notes e o documento `HyperDNSx evolução e forks.md` se ele estiver acessível no workspace.

Monte `docs/research/upstream-baseline.md` e `docs/research/gap-matrix.md` com:

- todas as flags e aliases do upstream, defaults reais, semântica e exemplos;
- todos os tipos DNS realmente aceitos e o que `-recon` efetivamente consulta;
- transportes existentes de verdade (UDP/TCP/DoH/DoT, e proxy quando suportado);
- comportamento atual dos resolvedores: primeira resposta, fallback, seleção, retry, rate limiting e metadata de origem;
- formato exato dos outputs, templates, JSONL e resume;
- diferenças de cada fork com testes/commits comprobatórios;
- recursos adjacentes de `puredns`, `shuffledns`, `MassDNS` e `retryabledns` que agregam valor, sempre via documentação e sem copiar código licenciado inadequadamente;
- custo de cada recurso: queries extras, memória, CPU, dependências, complexidade e riscos de falsos positivos.

Fonte de issue concreta a avaliar: `https://github.com/projectdiscovery/dnsx/issues/884` pede consultar todos os resolvedores. O HyperDNSx deve tratar isso com modos explícitos `first-success`, `all`, `quorum` e `compare`, mantendo a semântica atual por padrão/compatibilidade até testes provarem o contrário.

## FASE 3 — ARQUITETURA OBJETIVO

Evite reescrever o projeto todo antes de medir. Preserve pacotes que forem bons e use uma separação clara por responsabilidade:

1. **CLI/config:** valida argumentos, aliases, arquivo de configuração, schema de config e compatibilidade.
2. **Input pipeline:** arquivo, `stdin`, domains, wordlists, templates, CIDR/ASN quando já suportados; normalização determinística e deduplicação estável.
3. **Scheduler:** limite global, por resolver e por domínio, concorrência bounded, cancellation, progresso, backpressure, max queries e dry-run.
4. **Resolver pool:** transporte, health/latency, retries com jitter, failover, breaker, perfil e provenance.
5. **Query engine:** consulta tipos, interpreta RRsets e mantém RCODE, flags, TTL, tempo, resolver/protocolo, tentativa e raw opcional.
6. **Validation layer:** consistência entre resolvers, DNSSEC (se totalmente implementado), wildcard multinível, poisoned/inconsistent result; registra a explicação de cada decisão.
7. **Cache/state:** cache TTL-aware, negative caching corretamente documentado, checkpoint/resume versionado e idempotente.
8. **Output sinks:** console/stderr separados; JSONL/JSON/CSV/TSV e arquivos simultâneos usando writers de ciclo de vida estável; erros de I/O propagados.
9. **Public API:** API Go importável, tipos estáveis, callbacks/result streams; wrappers REST/MCP separados do core se realmente necessários.

Interfaces devem ser pequenas, testáveis e justificar-se pela arquitetura real do repositório. Não crie camadas vazias, factories genéricas ou abstrações “enterprise” sem consumidor real.

## FASE 4 — FEATURES, EM ORDEM DE PRIORIDADE

### P0 — Paridade, segurança de I/O e correção

- Preserve recursos upstream e flags/aliases.
- Faça testes de contrato de CLI e JSONL.
- Corrija lifecycle de output para que cada arquivo seja aberto uma vez, escritor seja bufferizado, flush/close sejam garantidos e erro de escrita seja reportado.
- Teste `-resume`, `-stream`, `-silent`, `stdin/stdout`, cancelamento e interrupção. Retomar não pode perder nem duplicar resultados.
- Não transformar `NXDOMAIN`, `NODATA`, `REFUSED`, `SERVFAIL`, timeout, truncation e erro local em um único “not found”.
- Nunca desligue validação TLS por padrão em DoH.

### P1 — Melhorias centrais que devem diferenciar o produto

**A. Multi-resolver**

- `--resolver-mode first-success|all|quorum|compare` (nome final pode mudar se colidir com convenções existentes).
- `first-success` compatível com comportamento upstream.
- `all` pergunta a todos os resolvers configurados e preserva todas as respostas distintas com a origem de cada uma.
- `quorum` exige critério explícito configurável e nunca esconde divergências silenciosamente.
- `compare` mostra divergência por RRset/RCODE/TTL/resolver.
- Opção `--unique-only` só deduplica apresentação, nunca apaga provenance dos dados estruturados.

**B. Resolver health/profiling**

- `--health-check` robusto, perfil por resolver, RTT, timeouts, RCODE, erro TLS, taxa de inconsistência.
- Exportar resultados do profile; não misturar consultas de profiling com os resultados do alvo.
- Retry/backoff com jitter, timeout por query, limite global e por endpoint, evitar thundering herd.
- Perfis configuráveis; nenhum resolver geográfico hardcoded como política universal.

**C. RR type extensível**

- Consultar IANA registry atual e os tipos suportados pela biblioteca de Go efetivamente usada.
- Adicionar tipos relevantes não presentes no baseline, priorizando HTTPS/SVCB (RFC 9460), NAPTR, TLSA, SSHFP, DS, DNSKEY, RRSIG, NSEC/NSEC3 e DNAME.
- Permitir `--type TYPE` para tipo conhecido/numérico quando o stack aceitar a query.
- Para RR não parseado, preservar representação genérica/raw com indicação explícita; jamais inventar parser ou reportar dado incorreto.
- Atualizar `-recon` sem tornar o comportamento muito caro implicitamente; oferecer `--recon-profile basic|extended|dnssec|all` documentado.

**D. DNSSEC real, opcional**

- Diferenciar `query DNSSEC records`, `request/observe DNSSEC flags` e `cryptographic validation`.
- Não alegar validação só por consultar DNSKEY/DS/RRSIG ou por observar AD flag.
- Validar secure/insecure/bogus/indeterminate, trust anchor, cadeia de confiança e expiração das assinaturas por uma biblioteca madura ou implementação revisada.
- Se a validação completa exceder o escopo ou não for confiável, lançar inicialmente apenas coleta e classificação dos registros como `NOT_VALIDATED`, deixando validação criptográfica como feature explicitamente pendente.

**E. Wildcard e validação cruzada**

- Preserve `-auto-wildcard/-aw` do upstream.
- Melhore cobertura para múltiplas raízes registráveis e níveis; trate load-balancing/CDN; randomize probes; configure budget de consultas.
- Compare respostas de validadores/resolvers confiáveis com cuidado; marque respostas divergentes como `inconsistent` em vez de as descartar sem explicação.
- Guardar evidência: probes, RRsets, TTL, resolver, algoritmo e motivo da classificação.
- Medir precisão/recall no fixture de teste.

**F. Cache/state e dados estruturados**

- Cache positivo por RRset com TTL.
- Cache negativo seguindo as normas aplicáveis (ver RFC 2308; não usar TTL arbitrário como se fosse regra universal).
- Cache on-disk opcional com versão/schema e invalidação definida.
- Resume/checkpoint estável; não duplicar output após restart.
- Output multi-sink: console legível + JSONL/CSV/TSV simultaneamente, sem abrir arquivo por resultado.
- JSONL versionado com campos: query/host, type, answers, resolver, transport, RCODE, TTL, latency, attempts, validation state, wildcard decision, raw opcional e provenance.
- Logs somente em stderr; stdout deve permanecer consumível em pipes.

### P2 — Recursos avançados sob teste

- DoQ (RFC 9250) como adaptador opcional apenas se uma dependência madura e o benchmark justificarem; DoH/DoT/TCP/UDP têm prioridade de paridade.
- Trace/CNAME chain com estrutura de saída, limites e loops detectados.
- Histórico/diff entre execuções e relatório de mudanças DNS.
- SQLite ou outro store apenas se não impuser CGo obrigatório nem prejudicar o build Android; priorize formatos texto robustos.
- API Go para integração de agentes, callbacks e streaming sem invocar shell.
- REST/MCP opcional fora do core; resposta HTTP, TXT e dados externos são entrada não confiável e não podem virar instruções de agente.

## FASE 5 — COMPATIBILIDADE E SEGURANÇA OPERACIONAL

- Alvos: Linux amd64/arm64; avaliar Windows/macOS conforme upstream; Android/Termux ARM64 como objetivo importante.
- Não usar PRoot, chroot, Docker, glibc rootfs, QEMU nem emulação para fingir compatibilidade Termux nativa. O build Android tem de ser testado com toolchain Go/Termux real.
- Se o workspace estiver num dispositivo Termux real, rode `go version`, `go env GOOS GOARCH`, teste e compile nativamente. Se não estiver, faça cross-build quando tecnicamente suportado e marque execução física como `NOT RUN`.
- Nenhum dependência de binário externo obrigatório como MassDNS no modo padrão; integração opcional só se existir benefício comprovado.
- `--dry-run`, `--max-inputs`, `--max-queries`, `--scope-file`/allowlist opcional, `--rate-limit`, limite por resolver e cancelamento previsível.
- O modo scope deve evitar expandir por acidente para domínios não autorizados; não bloquear indevidamente redes privadas necessárias para laboratório local.
- Qualquer telemetria ou autenticação cloud deve ser opt-in, documentada e coberta por teste offline. Não envie dados de alvos/consultas para serviço externo sem solicitação explícita.
- Não desabilitar verificação TLS por padrão. Não registrar tokens, secrets ou credenciais de resolver em logs.

## FASE 6 — TESTES E TDD OBRIGATÓRIO

Construa um test harness DNS local e determinístico. Não dependa de internet em unit tests.

Cobrir pelo menos:

- A/AAAA com múltiplas respostas; CNAME chain e loop;
- TXT segmentado, MX, NS, SOA, SRV, PTR, CAA e tipos novos implementados;
- NXDOMAIN versus NODATA versus NOERROR sem answer, REFUSED, SERVFAIL, timeout, truncation e fallback TCP;
- EDNS0, resposta grande, DNS compression e pacotes malformados;
- resolvers com respostas iguais/divergentes, timeout parcial e diferentes TTLs;
- modo first-success/all/quorum/compare;
- wildcard: nenhum, global, subdomínio, multinível, múltiplas raízes, CDN/load-balancing;
- DNSSEC secure/insecure/bogus/indeterminate se a validação for entregue;
- rate limit global/per resolver, cancelamento, retries/backoff e max queries;
- caching, TTL expiry, negative caching, resume, crash/restart, checkpoint e deduplicação;
- output sinks, JSONL schema, encoding, stdout/stderr, path inválido, disk full simulado, flush/close e file descriptors;
- IDNA, nomes com pontos finais, casing, espaços, CRLF, comentários e entradas duplicadas;
- fuzz tests para parser de inputs, outputs e wire packets quando viável.

Gates mínimos:

```bash
gofmt -w <somente arquivos Go alterados>
go test ./...
go test -race ./...
go vet ./...
go test -run 'Test.*' ./...
```

Execute linters que o upstream já usa. Não substitua workflows nem apague testes existentes para obter aprovação. Atualize fixtures/CI e docs no mesmo delta da feature.

## FASE 7 — BENCHMARK COMPROVÁVEL

Comparar upstream fixado e HyperDNSx com os mesmos inputs, resolvedores, timeout, retries, hardware, versão Go e rate limit. Nunca usar domínios alheios em testes de carga. Usar fixture local; para smoke tests externos, apenas domínio que o operador controla e poucas consultas.

Métricas:

- DNS queries/s e nomes resolvidos/s;
- mediana, p95/p99 de latência;
- CPU, RSS peak, alocação/GC e file descriptors;
- resultados corretos, falsos positivos, wildcard filtered e inconsistentes;
- número de consultas extras em `all`, `quorum`, wildcard e DNSSEC;
- performance cache cold/warm e streaming/non-streaming;
- tempo de startup, memória com entrada grande e throughput com limite fixo.

Rodar múltiplas repetições (mínimo 10 onde o runtime permitir), salvar comandos, hardware, OS/arch, Go, SHAs, arquivo de entrada e CSV/JSONL bruto em `benchmarks/results/`. Reportar mediana e dispersão. Se não executar, marcar `NOT RUN`. Não inventar ganho percentual.

## FASE 8 — LICENÇA E DEPENDÊNCIAS

- Preserve licença MIT upstream, copyright e notices quando aplicável.
- Audite as licenças de toda nova dependência, incluindo transitivas.
- `puredns`, `shuffledns` e `MassDNS` declaram GPL-3.0 nos respectivos repositórios consultados. Não copie/porte trechos de código deles para HyperDNSx sem análise jurídica explícita; obtenha inspiração funcional e implemente de forma independente.
- Fixe dependências, execute ferramenta de vulnerabilidades disponível (por exemplo `govulncheck` se instalado) e documente exceções.
- Não faça update indiscriminado de todas as deps no mesmo commit que muda algoritmos críticos; isole alterações e diagnostique falhas.

## FASE 9 — DOCUMENTAÇÃO E ENTREGÁVEIS

Atualize docs no mesmo conjunto de commits. Produza no mínimo:

- `README.md`: proposta e diferenças mensuráveis vs upstream;
- `docs/research/upstream-baseline.md`;
- `docs/research/forks.tsv`;
- `docs/research/fork-inventory.md` (100% dos forks diretos retornados pela API, uma linha cada);
- `docs/research/fork-deltas.md` (deltas reais/unique feature e evidência);
- `docs/research/gap-matrix.md`;
- `docs/architecture.md`;
- `docs/cli-reference.md` gerada/testada contra a CLI real;
- `docs/compatibility.md` e `docs/migration.md`;
- `docs/testing.md`;
- `docs/benchmarking.md` e resultados brutos;
- `docs/security.md`;
- `docs/termux-android.md` com distinção `TESTED-NATIVE`, `CROSS-BUILT` e `NOT TESTED`.

O relatório final de execução deve trazer:
1. resumo do estado inicial e commit-base;
2. recursos do upstream preservados;
3. total de forks diretos inventariados; quantos foram comparados; quantos apresentaram deltas funcionais; lista de não verificados e motivos;
4. tabela de features implementadas/pendentes e commits;
5. testes executados e resultados literais;
6. benchmarks reais, ou `NOT RUN`;
7. compatibilidade de plataforma por evidência;
8. dependências/licenças relevantes;
9. riscos restantes e próximos passos concretos.

## FASE 10 — REVIEW FINAL E ENTREGA

Antes de finalizar:

- revise `git diff --check`, `git status`, `git diff --stat` e diff completo;
- revise segurança, race conditions, tratamento de erro, fechamento de recursos, compatibilidade CLI e impacto de memória;
- remova dead code, TODOs enganosos, fake metrics, placeholders apresentados como features, e afirmações não verificadas;
- execute os testes finais depois da última modificação;
- mantenha commits pequenos e temáticos se git identity estiver configurada. Não force push e não publique no GitHub sem pedido explícito;
- se faltarem permissões, ferramentas ou rede, continue tudo que for possível e registre o bloqueio exato; não pare só para pedir confirmação.

**Critério final:** HyperDNSx só é considerado superior onde a matriz de features e os testes demonstram isso. Deve preservar compatibilidade do que existia, incorporar apenas deltas úteis dos forks e apresentar resultados reprodutíveis de precisão, robustez, portabilidade e desempenho.

---

### Fontes primárias a consultar durante a execução

- https://github.com/projectdiscovery/dnsx
- https://github.com/projectdiscovery/dnsx/releases
- https://github.com/projectdiscovery/dnsx/forks
- https://github.com/projectdiscovery/dnsx/issues/884
- https://github.com/projectdiscovery/dnsx/issues/924
- https://github.com/projectdiscovery/retryabledns
- https://github.com/projectdiscovery/shuffledns
- https://github.com/d3mondev/puredns
- https://github.com/blechschmidt/massdns
- https://www.iana.org/assignments/dns-parameters
- https://www.rfc-editor.org/rfc/rfc8484
- https://www.rfc-editor.org/rfc/rfc7858
- https://www.rfc-editor.org/rfc/rfc9250
- https://www.rfc-editor.org/rfc/rfc9460
- https://www.rfc-editor.org/rfc/rfc4033
- https://www.rfc-editor.org/rfc/rfc2308
