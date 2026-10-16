# HyperDNSx evolução e forks

**Dossiê técnico para orientar a criação do HyperDNSx**  
**Data de referência:** 16 de outubro de 2026  
**Projeto-base:** [`projectdiscovery/dnsx`](https://github.com/projectdiscovery/dnsx)  
**Objetivo:** desenhar uma evolução própria, auditável e mensuravelmente melhor que o dnsx oficial e os forks relevantes, preservando compatibilidade, correção DNS e funcionamento nativo em Termux/ARM64 quando possível.

## 1. Resumo executivo

O `dnsx` é uma ferramenta CLI em Go para consultas DNS em lote, resolução de hosts, brute-force de subdomínios, filtragem de wildcard, utilização de resolvedores personalizados e integração por `stdin`/`stdout`. O repositório usa `retryabledns` como cliente DNS e oferece consultas A, AAAA, CNAME, PTR, NS, MX, TXT, SRV, SOA, CAA, além de ANY e AXFR por opções específicas. Também oferece TCP/UDP/DoH/DoT por meio da camada de resolução usada pelo projeto, saída JSONL, templates, consultas PTR de faixas/ASNs, informação de CDN/ASN, tracing, retomada e controle de taxa.

**Diagnóstico:** não basta copiar a CLI e adicionar flags. Para superar de forma defensável o upstream, HyperDNSx deve priorizar quatro áreas:

1. **Precisão:** comparação de respostas entre resolvedores, classificação explícita de inconsistências, detecção de wildcard multinível, validação de respostas suspeitas e testes reproduzíveis para falsos positivos.
2. **Cobertura DNS:** inventário atualizável de RR types, melhor tratamento de tipos modernos como HTTPS/SVCB e NAPTR, e DNSSEC real como capacidade opt-in claramente distinta de simplesmente consultar DNSKEY/DS.
3. **Operação:** limites de taxa globais e por resolvedor, timeouts/retries observáveis, perfis de saúde dos resolvedores, cache TTL-aware, checkpoints idempotentes e relatórios estruturados.
4. **Engenharia:** preservar os recursos e os contratos de saída do dnsx, testar regressões, publicar benchmarks honestos e comprovar builds em Linux, Android/Termux e outras plataformas declaradas.

A melhor oportunidade identificada em issues públicas é a consulta a **todos** os resolvedores informados, com deduplicação opcional e registro da origem de cada resposta. A issue upstream que pediu esse modo foi fechada como “not planned”; isso a torna uma lacuna de produto plausível, não prova de que uma implementação seja trivial. Fonte: [issue #884](https://github.com/projectdiscovery/dnsx/issues/884).

**Limite desta pesquisa:** consultei o upstream, release notes, issues, implementações públicas de forks representativos e ferramentas concorrentes. A API/listagem completa de forks não foi disponibilizada pelo conector de leitura usado nesta sessão, e o índice de busca mistura forks reais com wrappers, ações e projetos de nome semelhante. Portanto, não afirmo ter comparado manualmente o código de cada um dos ~329–332 forks observados. O prompt separado exige que o OpenCode faça o inventário paginado completo, compare cada fork e registre explicitamente o que foi ou não verificado.

## 2. Estado do upstream: separar release estável de `dev`

Fontes:

- [Repositório oficial](https://github.com/projectdiscovery/dnsx)
- [Releases](https://github.com/projectdiscovery/dnsx/releases)
- [README do branch `dev`](https://github.com/projectdiscovery/dnsx/blob/dev/README.md)
- [`go.mod` de `dev`](https://github.com/projectdiscovery/dnsx/blob/dev/go.mod)
- [`go.mod` de `v1.3.1`](https://github.com/projectdiscovery/dnsx/blob/v1.3.1/go.mod)

Na consulta de 16/10/2026, a release marcada como latest era **v1.3.1**, publicada em 31/08/2026. O branch padrão do repositório é `dev`, cujo histórico continha alterações até 12/10/2026. Essas referências não devem ser misturadas como se fossem o mesmo snapshot.

Existe uma divergência importante nos requisitos declarados: o README ainda diz Go 1.21; o `go.mod` da tag `v1.3.1` exige Go 1.25; o `go.mod` de `dev` exige Go 1.26. HyperDNSx deve documentar requisitos extraídos do `go.mod` e do build realmente testado, não copiar cegamente a versão escrita no README.

### 2.1 Funcionalidades documentadas no upstream

| Área | Recursos observados |
|---|---|
| Entradas | `-l/-list` para hosts/subdomínios, `-d/-domain` para domínios e `-w/-wordlist` para palavras; leitura por arquivo, lista separada por vírgulas e `stdin`. |
| Consultas | A, AAAA, CNAME, PTR, NS, MX, TXT, SRV, SOA, CAA; além de ANY, AXFR e modo `-recon`. O conjunto exato de tipos incluídos em `-recon` precisa ser conferido no snapshot que será usado. |
| Resolução/brute-force | Resolver listas e gerar nomes a partir de domínios + wordlist. |
| Transporte | Resolvers configuráveis e transporte UDP/TCP/DoH/DoT por meio de `retryabledns`; versões anteriores documentam suporte a proxy SOCKS5 para DNS. |
| Wildcard | Filtro manual por domínio e detecção automática com `-auto-wildcard`/alias `-aw`, introduzida no ciclo v1.3.0. |
| Resolvedores | Lista de resolvedores via `-r`; tentativas (`-retry`), timeout e concorrência. O padrão documentado é 100 threads e 3 s de timeout; rate limit fica desabilitado por padrão no README consultado. |
| Limitação de taxa | `-rl/-rate-limit`; a nova ferramenta deve acrescentar controle global e por resolvedor sem remover o modo atual. |
| Saída | Texto legível, `-o`, JSONL (`-j/-json`), opção para omitir raw DNS (`-omit-raw/-or`) e templates (`-ot`). |
| Operação | `-resume`, `-stream`, `-trace`, `-trace-max-recursion`, `-hf/-hostsfile`, `-stats`, `-hc/-health-check`, `-timeout`, `-retry`. |
| Enriquecimento | `-cdn`, `-asn`, status de consulta (`-rcode`) e resposta/response-only (`-resp`, `-resp-only`). |
| Integração | Pipe Unix com `stdin`/`stdout`, compatibilidade com outras ferramentas ProjectDiscovery e workflows de automação. |

A release v1.3.0 adicionou detecção automática de wildcard em múltiplas raízes registráveis e templates de saída. v1.2.2 adicionou suporte a proxy SOCKS5 para DNS. Essas funcionalidades são do **upstream atual**, não devem ser atribuídas como exclusivas de forks. Referências: [releases oficiais](https://github.com/projectdiscovery/dnsx/releases), [README](https://github.com/projectdiscovery/dnsx) e [retryabledns](https://github.com/projectdiscovery/retryabledns).

### 2.2 Recursos a preservar para compatibilidade

- Sintaxe e aliases existentes de CLI.
- Semântica de `-l`, `-d`, `-w`, wildcard, `-recon`, `-r`, `-rl`, `-t`, timeout e retries.
- Entrada/saída por pipe e modo silencioso, para que pipelines existentes não quebrem.
- JSONL determinístico, templates e campos existentes.
- Formato/configuração dos resolvedores suportados por `retryabledns`.
- `-resume` e `-stream`, respeitando suas limitações documentadas.

Qualquer mudança incompatível deve ser opt-in, registrada em `docs/migration.md` e coberta por testes de contrato.

## 3. Auditoria dos forks e derivados encontrados

### 3.1 Índice canônico e contagem

- [Lista oficial de forks](https://github.com/projectdiscovery/dnsx/forks)
- [Busca GitHub por forks chamados dnsx](https://github.com/search?q=dnsx+fork%3Aonly&type=repositories)

As consultas feitas em 16/10/2026 mostraram contagens diferentes conforme a fonte/captura (aproximadamente 329–332). Isso é esperado em interfaces com caches e atividade dinâmica. Uma busca por nome não é o mesmo que um inventário do endpoint direto de forks: inclui, por exemplo, `dnsx-action`, wrappers e projetos derivados. O prompt do OpenCode contém um procedimento para consultar a API paginada oficial e produzir um inventário verificável, com dados por fork.

### 3.2 Forks diretos/derivados com delta concreto verificado

#### `eexp/dnsx`

- Link: https://github.com/eexp/dnsx
- Base identificada: `projectdiscovery/dnsx`.
- README resumido em chinês descreve a adaptação para “DNS chinês”.
- O diff verificável muda o module path para `github.com/eexp/dnsx`, substitui algumas referências internas e introduz `github.com/eexp/cdncheck`.
- O diff também acrescenta resolvedores padrão como `114.114.114.114`, `180.76.76.76`, `112.124.47.27`, `223.5.5.5` e `223.6.6.6`, além dos resolvedores já presentes.
- Último push consultado: 21/09/2024. Comparação feita contra `dev`: 3 commits à frente e 257 atrás no momento da consulta.

**Valor para HyperDNSx:** configuração de resolvedores por perfil/região pode ser útil, mas não hardcode resolvedores de um país como padrão universal. Implementar `resolver-profiles` via arquivo de configuração e mostrar origem, protocolo, saúde, latência e política de confiança de cada resolvedor.

#### `scournomad/dnsx`

- Link: https://github.com/scournomad/dnsx
- O próprio README diz que adiciona um flag para gravar JSON em arquivo, mantendo a saída legível no terminal.
- Introduz campos/opção `JSONFile` e `-json-file/-jf`.
- O patch abre o arquivo em modo append no processamento de cada resultado e adia `Close`/`Flush` dentro do worker. Como o `defer` só executa quando a função do worker termina, este padrão pode acumular descritores de arquivo conforme a quantidade de resultados. Isso deve ser confirmado por teste antes de concluir o impacto exato, mas é um padrão de implementação que não deve ser copiado.
- Último push consultado: 14/09/2023. Comparação contra `dev`: 1 commit à frente e 733 atrás.

**Valor para HyperDNSx:** saída multi-destino real (`console` + `jsonl` + `csv`) por writers persistentes, com buffering, tratamento de erro, flush/checkpoint e fechamento único no ciclo de vida da execução. Um resultado deve ser emitido uma vez por destino, sem abrir/fechar arquivo por resposta.

#### `geeknik/dnsx`

- Link: https://github.com/geeknik/dnsx
- Mantém README muito próximo do upstream.
- O push consultado é de 12/08/2026, mas o último commit reportado na consulta foi de 24/09/2025 (`Revert "non fatal"`). Push/updated timestamp não prova mudança funcional; a comparação contra `dev` indicava 71 commits atrás.

**Valor para HyperDNSx:** candidato para triagem de commits; nenhuma funcionalidade exclusiva foi comprovada nesta pesquisa.

#### `enlightment-droid/dnsx`

- Link: https://github.com/enlightment-droid/dnsx
- Código/README muito próximos do upstream; último commit consultado em 28/10/2024.

**Valor:** referência histórica, sem contribuição Android/Termux demonstrada. O nome do owner não é evidência de que seja uma portabilidade para Android.

#### `esgeeks/dnsx`

- Link: https://github.com/esgeeks/dnsx
- Fork histórico; último commit consultado em 19/04/2021.

**Valor:** histórico, sem recurso diferenciador atual identificado.

### 3.3 Lista de candidatos à inspeção automatizada

A pesquisa GitHub também retornou estes nomes como candidatos relacionados a `dnsx`; eles **não devem ser considerados forks úteis sem verificar o campo `parent`, o branch e o diff real**:

- https://github.com/gprime31/dnsx
- https://github.com/0xNoobie1100/dnsx
- https://github.com/jdawg24/dnsx
- https://github.com/FDlucifer/dnsx
- https://github.com/0xflotus/dnsx
- https://github.com/pedroevaristo960/dnsx
- https://github.com/rvrsh3ll/dnsx
- https://github.com/BlackNurse/dnsx
- https://github.com/1234-ad/dnsx
- https://github.com/gcarmonatellez15-spec/dnsx
- https://github.com/tetadsf23/dnsx
- https://github.com/Veirt/dnsx
- https://github.com/maux339-cpu/dnsx
- https://github.com/Conqwest/dnsx
- https://github.com/asekibaala/dnsx

A tarefa do agente é classificar **cada fork direto** como: idêntico/sem delta funcional, fork defasado, fork com correções, fork com recurso exclusivo, fork quebrado/inconclusivo ou projeto não relacionado. Não assumir que um commit extra significa uma feature útil.

## 4. O que aprender com ferramentas concorrentes

Não são forks do dnsx; são projetos independentes que resolvem problemas adjacentes. Eles servem como referência de comportamento/produto, não como fonte para copiar código.

### 4.1 `puredns`

Fonte: https://github.com/d3mondev/puredns

Pontos de valor documentados:

- pipeline de resolução em massa com MassDNS;
- detecção de wildcard orientada a precisão, incluindo raízes multinível;
- validação posterior usando resolvedores considerados confiáveis para reduzir respostas envenenadas;
- persistência separada de domínios válidos, raízes wildcard e saída de resolução;
- sanitização/normalização de inputs e modo quieto para pipes.

**Lição:** não basta perguntar a um resolver e aceitar a primeira resposta. HyperDNSx deve poder comparar e validar respostas com perfis de resolvedores independentes e registrar por que um resultado foi aceito, filtrado ou marcado como divergente.

**Licença:** o README atual consultado declara GPL-3.0. Não copiar código do puredns para uma base MIT sem análise de compatibilidade/licença. Reimplementar ideias de alto nível de forma independente ou obter revisão jurídica antes de importar código.

### 4.2 `shuffledns`

Fonte: https://github.com/projectdiscovery/shuffledns

É um wrapper em Go de MassDNS voltado a resolução/brute-force de subdomínios com wildcard handling, `stdin/stdout` e saída NDJSON. Atualmente requer o binário externo MassDNS.

**Lição:** um modo de bulk resolution com alto throughput e um algoritmo wildcard dedicado tem valor real. HyperDNSx pode buscar a simplicidade de um único binário, sem dependência obrigatória de um executável externo, mantendo adaptadores opcionais se o benchmark justificar.

**Licença:** GPLv3 conforme o repositório consultado. Não importar implementação sem revisão de licença.

### 4.3 `massdns`

Fonte: https://github.com/blechschmidt/massdns

O README descreve o projeto como um resolver para consultas DNS em massa e informa throughput muito alto em condições específicas. A página também lista como desafios: controlar flood/rate limiting, validar respostas entre resolvers, lidar com DNS poisoning e detectar wildcard. Números publicados por um projeto não equivalem a benchmark comparável para HyperDNSx; devem ser contextualizados e reproduzidos sob condições iguais.

### 4.4 `retryabledns`

Fonte: https://github.com/projectdiscovery/retryabledns

É a biblioteca DNS Go usada pelo dnsx e documenta resolvers do sistema ou customizados, retry, diversos tipos de consulta e protocolos UDP/TCP/DoH/DoT. HyperDNSx deve primeiro avaliar a abstração existente antes de desenvolver um cliente próprio. Se substituir ou estender essa camada, criar interfaces para transporte e testes de contrato para impedir regressões de protocolo.

## 5. Gaps de produto e design recomendado

### P0 — correção e paridade (antes de adicionar recursos)

1. Congelar baseline upstream por tag/commit e conservar a licença/avisos de copyright.
2. Criar testes de contrato para a CLI, JSONL, templates, `stdin/stdout`, aliases, wildcard, resume e modo stream.
3. Construir um ambiente de teste DNS local isolado: autoritativo fake, resolvers controlados, TTL variável, DNS truncado, atrasos, NXDOMAIN, NODATA, SERVFAIL, REFUSED, CNAME loops e wildcard multinível.
4. Corrigir lifecycle de output: writers abertos uma vez, tratamento de erros, flush e fechamento garantidos, checkpoint idempotente e arquivo não corrompido após interrupção.
5. Normalizar domínio/IDNA/IP sem alterar labels válidos; deduplicar sem perder a origem ou os dados conflitantes.
6. Produzir testes de regressão para bugs encontrados no upstream e nos forks antes de dizer que foram corrigidos.

### P1 — diferenciação mais forte

#### 5.1 Modos de consulta aos resolvedores

Implementar modos explícitos e compatíveis:

- `first-success`: compatível com o comportamento atual;
- `all`: consulta cada resolver e coleta todos os resultados;
- `quorum`: classifica respostas por consenso, com configuração de quantidade/percentual;
- `compare`: não escolhe automaticamente; mostra diferenças por resolver.

Registrar, por resposta: resolver, protocolo, latência, tentativa, RCODE, flags relevantes, TTL e RRsets recebidos. Deduplicar dados sem perder proveniência. Não afirmar que uma resposta é “verdadeira” só porque a maioria concordou: resolvers podem compartilhar upstream, políticas ou cache.

Issue de referência: https://github.com/projectdiscovery/dnsx/issues/884.

#### 5.2 Sistema de resolvedores

- Arquivos/perfis por objetivo/região/ambiente.
- Health check e benchmark separados da varredura de alvos.
- Métricas de timeout, perda, RTT, RCODE, inconsistência e taxa efetiva.
- Seleção/failover por score explicável, com opção de desativar heurísticas.
- Limite global e por resolver; backoff exponencial com jitter; circuit breaker temporário para resolvers que falham; evitar retry storms.
- Não hardcode uma lista geográfica como default universal. Não enviar consultas para endpoint externo não configurado sem informar o usuário.

#### 5.3 Cobertura de RR Types

O IANA mantém o registro oficial de tipos DNS: https://www.iana.org/assignments/dns-parameters. Não codificar uma lista manual desatualizada como única verdade. Usar uma tabela versionada/auditável e permitir `--type TYPE` para registros que a biblioteca consegue representar.

Além dos tipos atuais, avaliar pelo menos:

- SVCB e HTTPS (RFC 9460);
- NAPTR, TLSA, SSHFP, DS, DNSKEY, RRSIG, NSEC/NSEC3, DNAME, CERT e tipos adicionais de uso operacional;
- tipos desconhecidos como RFC 3597, preservando wire-format/representação textual segura quando possível;
- seleção de múltiplos tipos sem duplicar a consulta desnecessariamente.

Não prometer “todo tipo DNS” se a biblioteca não descodifica ou se os testes não cobrem aquele tipo. Para tipos não interpretados, exibir tipo numérico, raw opcional e texto genérico conforme RFC 3597.

Referências de normas: [IANA DNS Parameters](https://www.iana.org/assignments/dns-parameters), [RFC 9460 — SVCB/HTTPS](https://www.rfc-editor.org/rfc/rfc9460), [RFC 4033 — DNSSEC](https://www.rfc-editor.org/rfc/rfc4033).

#### 5.4 DNSSEC de verdade

Separar claramente três funcionalidades:

1. consultar tipos DNSSEC (`DNSKEY`, `DS`, `RRSIG`, `NSEC/NSEC3`);
2. solicitar/observar dados DNSSEC e flags AD/CD;
3. **validar criptograficamente a cadeia de confiança**.

Não chamar a etapa 1 ou 2 de “validação DNSSEC”. Se a validação completa entrar no escopo, usar biblioteca madura ou implementação cuidadosamente revisada, validar trust anchors e datas de assinatura, e testar estados secure/insecure/bogus/indeterminate. Se não for viável, entregar descoberta de registros DNSSEC com status explícito “não validado”, sem falsa segurança.

#### 5.5 Wildcard e poisoning

- Detectar wildcard em múltiplas raízes e níveis, com amostras aleatórias e controle sobre custo de sondagem.
- Distinguir wildcard global, wildcard numa subárvore, respostas CDN/load-balanced e resultados realmente específicos.
- Permitir comparar respostas repetidas e entre resolvedores de confiança diferentes.
- Guardar evidências da classificação: probes usados, RRsets, TTLs, resolver, motivo e confiança.
- Não eliminar resposta divergente silenciosamente; marcar como `inconsistent` ou `needs-validation`.

Referências: [puredns](https://github.com/d3mondev/puredns), [shuffledns](https://github.com/projectdiscovery/shuffledns), [issue de auto-wildcard upstream](https://github.com/projectdiscovery/dnsx/issues/924). A feature automática já aparece em release recente, portanto deve ser preservada e testada antes de reimplementar.

#### 5.6 Cache e resiliência

- Cache positivo respeitando TTL por RRset.
- Cache negativo segundo comportamento DNS aplicável (SOA/negative caching; conferir RFC 2308).
- Cache opcional em disco, com schema/versionamento e invalidation clara.
- Não confundir timeout, NXDOMAIN, NODATA, REFUSED, SERVFAIL e erro local.
- Retries idempotentes com jitter; contagem de tentativas e causa terminal na saída JSON.
- Checkpoints transacionais e retomada que não duplica resultados.
- Limite de memória e opção de streaming para listas grandes, sem carregar tudo na RAM.

#### 5.7 Saída e integração

- `console`, `jsonl`, `json`, `csv` e `tsv` com schemas versionados.
- Escrita simultânea para stdout e múltiplos arquivos em um pipeline único.
- Campos por resolver, protocolo, query type, RCODE, TTL, tempo, número de tentativas, estado wildcard, validação, CDN/ASN e proveniência.
- JSONL estável, encoding UTF-8, escape correto, nenhuma mistura de mensagens de log na saída `-silent`.
- Resume/checkpoints seguros para outputs grandes.
- API Go importável com structs estáveis e callbacks de resultados/progresso/erros; opcionalmente fornecer REST/MCP como wrapper separado, não acoplar o core ao protocolo MCP.

#### 5.8 Segurança e UX

- Modo `--dry-run` mostrando número de nomes/consultas que seria gerado antes de executar brute-force.
- `--max-inputs`, `--max-queries`, `--scope-file` e allowlist opcional para evitar erros de escopo.
- Rate limit conservador configurável e feedback quando resolvers retornam `REFUSED`, `SERVFAIL` ou throttling.
- Não ocultar erro TLS/DoH nem desabilitar validação de certificado por default.
- Configuração sem transmissão telemétrica por default; qualquer cloud/auth e tráfego externo deve ser explícito, documentado e testável offline.
- Saída redigida opcional para credenciais ou parâmetros sensíveis de endpoints de resolver.

### P2 — diferenciação adicional, priorizada após P0/P1

- Transport adapter para DoQ (RFC 9250) ou HTTP/3 quando estável e útil; opcional, sem degradar UDP/TCP/DoH/DoT existentes.
- Trace de cadeia CNAME e resolução iterativa/trace com logs estruturados.
- Relatórios por zona/subárvore, diferenças entre execuções e histórico de mudanças DNS.
- Exportadores para SQLite/Parquet somente se a complexidade e portabilidade em Android forem aceitáveis; evitar dependência CGo obrigatória.
- Wordlist templates com limites de cardinalidade e geração de variantes claramente separada do motor DNS.
- Biblioteca pública, bindings e integração CLI consistentes.

## 6. Benchmark: como provar superioridade

Nenhum resultado de velocidade deve ser alegado sem medição reproduzível. O plano deve comparar a mesma versão upstream e HyperDNSx no mesmo hardware, sistema, Go, resolvers, wordlist, timeout, retries e limite de taxa.

### 6.1 Métricas mínimas

- consultas DNS/segundo e nomes resolvidos/segundo;
- mediana, p95 e p99 da latência por consulta;
- CPU, RSS/memória máxima, alocação/GC e file descriptors máximos;
- taxa de timeout/erro por resolver e por tipo;
- resultados verdadeiros/filtrados/falsos positivos no fixture conhecido;
- número de queries adicionais exigidas pelo wildcard detection/consensus;
- tempo total em modo `first-success`, `all` e `quorum`;
- throughput sob limite de taxa fixo;
- impacto de cache frio/quente e modo streaming.

### 6.2 Corpus e metodologia

1. Fixture local determinístico com RRsets conhecidos, latência controlada, wildcard global/multinível, respostas variáveis, CNAME chains e failures injetadas.
2. Corpus sintético de pelo menos 100 mil nomes para teste de throughput/memória; não usar alvos públicos em massa para benchmark.
3. Testes de integração contra domínios que o operador controla, com volume baixo e autorização explícita.
4. Repetir pelo menos 10 vezes, reportar mediana e dispersão; guardar comandos, commit SHA, versão Go, OS/arch, hardware e resolver config.
5. Comparar precisão e performance separadamente: resolver mais rápido não é melhor se retorna respostas erradas ou falsos positivos.
6. Publicar dados brutos CSV/JSONL e scripts. Sem prova, marcar benchmark como `NOT MEASURED`.

## 7. Matriz de qualidade e aceitação

| Área | Gate de aceitação |
|---|---|
| Paridade upstream | todos os flags/aliases básicos catalogados; testes de contrato aprovados; sem regressão silenciosa de JSONL. |
| Tipos DNS | fixtures por tipo implementado, incluindo múltiplos RRsets, TTL e erro. Tipos não implementados são documentados como tais. |
| Transportes | testes para UDP/TCP/DoH/DoT preservados; DoQ/HTTP3 só declarados se passar integração real. |
| All-resolvers | cada resposta conserva a origem; timeout de um resolver não apaga resposta válida de outro; dedupe determinístico. |
| Wildcard | testes de wildcard, multi-root, multi-level, CDN/load-balancing e domínio sem wildcard; falsos positivos medidos. |
| DNSSEC | estados crypto verificados e trust anchors documentados; sem chamar mera consulta de registro de validação. |
| I/O | sem mistura stdout/stderr no modo silencioso; outputs concorrentes consistentes; sem vazamento de file descriptors. |
| Resume | parar e retomar em vários pontos sem perder/dobrar eventos; versões incompatíveis da sessão detectadas. |
| Limites | rate limit global/per-resolver, max queries e dry-run testados; nenhuma concorrência irrestrita por acidente. |
| Plataforma | `go test ./...`, `go test -race ./...`, `go vet ./...`, build Linux amd64/arm64 e execução real nativa em Termux/Android ARM64 se o dispositivo estiver disponível. |
| Reprodutibilidade | benchmarks e testes têm scripts, versões e commit; claims não sustentados são removidos. |
| Segurança | nenhuma chamada de rede inesperada em testes; nenhum segredo em logs; dependências/licenças verificadas. |

## 8. Licença, proveniência e dependências

O dnsx upstream é MIT. Os projetos puredns, shuffledns e MassDNS consultados declaram GPL-3.0. Não copiar trechos GPL para um produto pretendido como MIT sem análise de compatibilidade/licença. Reaproveitar ideias e comportamentos descritos publicamente não significa copiar implementação. Para cada feature transplantada de um fork, guardar o link/commit e comparar o diff; manter notices aplicáveis e atribuição devida.

Inventariar dependências transitivas e licenças; fixar versões em `go.mod`/`go.sum`; revisar advisories e compatibilidade com Go/Termux antes de upgrades grandes.

## 9. Prioridade recomendada

**Ordem mais racional para produzir superioridade real:**

1. Inventário completo dos forks + congelamento do baseline e todos os testes do upstream.
2. P0 de correção: lifecycle de arquivos, JSONL, resume, wildcard e erros de rede.
3. `resolver-mode=all|quorum|compare` com provenance de cada resposta.
4. Resolver health/profiling, rate limiting global e por servidor, retries com jitter.
5. RR type registry extensível, incluindo HTTPS/SVCB e registros DNSSEC; validar DNSSEC com implementação confiável só depois de teste de cadeia.
6. Wildcard multinível e validação de respostas no estilo de produto puredns, com evidência e métricas próprias.
7. cache e persistência com TTL, checkpoint robusto, relatórios múltiplos.
8. benchmarks e builds reais, incluindo Termux nativo.
9. recursos avançados (DoQ, UI, REST/MCP, analytics) apenas após as camadas anteriores ficarem estáveis.

## 10. Fontes consultadas

### Projetos primários

- [dnsx — código e README](https://github.com/projectdiscovery/dnsx)
- [dnsx — releases](https://github.com/projectdiscovery/dnsx/releases)
- [dnsx — issues](https://github.com/projectdiscovery/dnsx/issues)
- [dnsx — forks](https://github.com/projectdiscovery/dnsx/forks)
- [eexp/dnsx](https://github.com/eexp/dnsx)
- [scournomad/dnsx](https://github.com/scournomad/dnsx)
- [geeknik/dnsx](https://github.com/geeknik/dnsx)
- [enlightment-droid/dnsx](https://github.com/enlightment-droid/dnsx)
- [esgeeks/dnsx](https://github.com/esgeeks/dnsx)
- [retryabledns](https://github.com/projectdiscovery/retryabledns)
- [shuffledns](https://github.com/projectdiscovery/shuffledns)
- [puredns](https://github.com/d3mondev/puredns)
- [MassDNS](https://github.com/blechschmidt/massdns)

### Issues e funcionalidades

- [DNSX #884 — consultar todos os resolvedores](https://github.com/projectdiscovery/dnsx/issues/884)
- [DNSX #924 — auto wildcard](https://github.com/projectdiscovery/dnsx/issues/924)
- [Release notes](https://github.com/projectdiscovery/dnsx/releases)

### Normas e registros

- [IANA DNS Parameters](https://www.iana.org/assignments/dns-parameters)
- [RFC 8484 — DNS over HTTPS](https://www.rfc-editor.org/rfc/rfc8484)
- [RFC 7858 — DNS over TLS](https://www.rfc-editor.org/rfc/rfc7858)
- [RFC 9250 — DNS over QUIC](https://www.rfc-editor.org/rfc/rfc9250)
- [RFC 9460 — SVCB e HTTPS RRs](https://www.rfc-editor.org/rfc/rfc9460)
- [RFC 4033 — DNSSEC: introdução e requisitos](https://www.rfc-editor.org/rfc/rfc4033)
- [RFC 2308 — Negative Caching of DNS Queries](https://www.rfc-editor.org/rfc/rfc2308)

---

## Conclusão

HyperDNSx deve começar como uma evolução compatível do dnsx, mas provar seu diferencial pela qualidade das respostas e pela operacionalidade: múltiplos resolvedores com provenance, validação/wildcard explicáveis, cache/resume corretos, cobertura extensível de RR types, outputs robustos e benchmarks honestos. O inventário completo de forks deve ser gerado e validado no ambiente com acesso à API do GitHub antes de declarar que todas as contribuições relevantes foram incorporadas.
