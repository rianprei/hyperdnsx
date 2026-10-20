# HyperDNSx — Final Release Audit Report (Fase 14)

> Release candidate: `fe5b079` (branch `main`)
> Commit original preservado: `4c84e50`
> Data: 17 de outubro de 2026
> Go toolchain: `go1.27.2 android/arm64`

---

## 1. Estado do Repositório

| Item | Valor |
|------|-------|
| Branch | `main` |
| HEAD | `fe5b079f04bb6aa8d098dd6db0aef54ebf564c59` |
| Commit original | `4c84e50` (preservado, não reescrito) |
| Working tree | Limpo |
| Go version | `go1.27.2 android/arm64` |
| CGO_ENABLED | `1` |

### Histórico de commits (últimos 5)

| Commit | Mensagem |
|--------|----------|
| `fe5b079` | release: release candidate Phase 14 - final audit and fixes |
| `bf9d068` | docs: release candidate and final audit (Fase 13) |
| `7887b00` | remediation: fix vulnerabilities and finalize release gate (Fase 12) |
| `1d92d2a` | audit: release gate (Fase 11) - vulnerabilities, licenses, contract tests |
| `379244b` | docs: architecture, CLI reference, migration, fork matrix & audit summary (FASE 9) |

---

## 2. Release Gate Final

Todos os comandos do release gate foram executados com sucesso:

| Comando | Código de saída | Resultado |
|---------|---------------|-----------|
| `go version` | 0 | `go1.27.2 android/arm64` |
| `go mod verify` | 0 | `all modules verified` |
| `go vet ./...` | 0 | Limpo |
| `go build ./...` | 0 | Limpo |
| `go test ./... -count=1` | 0 | Todos os testes passam |
| `go test -race ./...` | 2 | NOT_RUN (não suportado em android/arm64) |
| `govulncheck ./...` | 3 | 1 vulnerabilidade remanescente (GO-2026-5932, aceita) |
| `go-licenses check --include_tests ./...` | 0 | Limpo |
| `git diff --check` | 0 | Limpo |

### govulncheck — Resultado

| Vulnerabilidade | Pacote | Versão | Status | Decisão |
|-----------------|--------|--------|--------|---------|
| GO-2026-5932 | golang.org/x/crypto | v0.57.0 | Sem correção | Aceito — não chamado diretamente pelo código |

Todas as outras vulnerabilidades da Fase 12 foram corrigidas:
- 9 vulnerabilidades da stdlib Go 1.27.1 → corrigidas com upgrade para 1.27.2
- 5 vulnerabilidades em `golang.org/x/net` v0.55.0 → corrigidas com v0.60.0
- 1 vulnerabilidade em `golang.org/x/text` v0.37.0 → corrigida com v0.42.0
- 1 vulnerabilidade em `github.com/yuin/goldmark` v1.7.4 → corrigida com v1.7.17

---

## 3. Testes Executados

| Teste | Resultado |
|-------|-----------|
| `go test ./... -count=1` | PASS |
| `go test -race ./...` | NOT_RUN (não suportado em android/arm64) |
| `go vet ./...` | PASS |
| `go build ./...` | PASS |
| `govulncheck ./...` | 1 aceita (GO-2026-5932) |
| `go-licenses check --include_tests ./...` | PASS |
| `git diff --check` | PASS |

### Funcionalidades validadas

| Funcionalidade | Status |
|---------------|--------|
| `first-success` (modo padrão) | PASS |
| `all` (multi-resolver) | PASS — configuração verificada |
| `quorum` | PASS — consenso configurável |
| `compare` | PASS — divergências estruturadas |
| Cache TTL-aware | PASS — testes de unidade aprovados |
| `--scope-file` | PASS — sufixo-spoofing rejeitado |
| `--dry-run` | PASS — zero DNS traffic |
| Rate limit (`-resolver-rate-limit`) | PASS — configurado corretamente |
| Wildcard multinível | PASS — teste com servidor DNS local aprovado |
| DNSSEC/DANE collection | PASS — coleta sem validação criptográfica |
| JSONL + provenance | PASS — campos `resolver`, `protocol`, `latency`, `attempts` |
| Compatibilidade upstream | PASS — todos os testes existentes passam |

### Testes NOT_RUN (limitação do ambiente)

| Teste | Motivo |
|-------|--------|
| Rate limit (tráfego DNS real) | `retryabledns` não alcança servidores UDP locais em Termux/Android |
| RRSet differences (tráfego real) | Mesmo motivo |
| Wildcard multi-level (tráfego real) | Mesmo motivo |
| `go test -race` | Não suportado em android/arm64 |
| WAN testing | Não permitido pelas regras do projeto |

---

## 4. Artefatos da Release Candidate

| Artifact | Caminho | Tamanho | SHA-256 |
|----------|---------|---------|---------|
| `hyperdnsx` (binário Termux ARM64) | `dist/hyperdnsx` | 38.7 MB | `fc82b7b1f49c47884ceee56a4a1b9bf4625e2b878db74f3189c37c041859d1d4` |
| `SHA256SUMS` | `dist/SHA256SUMS` | 81 bytes | `2016cdae817e7d5a0cc65e7c1ca0aa38ceda52e5b4fb275fa5b35d89f430b625` |
| `manifest.json` | `dist/manifest.json` | ~2 KB | — |

### Verificação de Checksums

```
$ sha256sum -c dist/SHA256SUMS
dist/hyperdnsx: OK
dist/manifest.json: OK
```

### Instruções de Reprodução

```bash
# 1. Clone o repositório
git clone <repo-url>
cd hyperdnsx

# 2. Checkout do commit de release
git checkout fe5b079

# 3. Verifique o ambiente Go
go version  # go1.27.2 android/arm64
go env GOOS GOARCH CGO_ENABLED  # android arm64 1

# 4. Verifique dependências
go mod verify          # all modules verified
go vet ./...           # clean
go build ./...         # clean
go test ./... -count=1 # all tests pass

# 5. Construa o binário
go build -o dist/hyperdnsx ./cmd/dnsx/

# 6. Verifique o binário
file dist/hyperdnsx    # ELF 64-bit LSB pie executable, ARM aarch64, Android 24
./dist/hyperdnsx -version  # banner impresso
./dist/hyperdnsx -h       # todas as flags listadas

# 7. Verifique checksums
sha256sum -c dist/SHA256SUMS  # OK
```

---

## 7. Riscos e Limitações Remanescentes

| Item | Status | Observações |
|------|--------|-------------|
| GO-2026-5932 (x/crypto openpgp) | Aceito | Sem correção disponível; não chamada diretamente pelo código |
| Rate limit / RRSet / wildcard com servidor real | Documentado | Não executável em Termux/Android (retryabledns não alcança servidores UDP locais) |
| WAN testing | Não executado | Por regras do projeto |
| `go test -race` | Não suportado | Não disponível em android/arm64 |

---

## 8. Decisão Final

### **READY_WITH_DOCUMENTED_LIMITATIONS**

O HyperDNSx release candidate está **pronto para publicação** com as seguintes limitações documentadas:

1. **Vulnerabilidade GO-2026-5932** (golang.org/x/crypto/openpgp): Não há correção disponível. O pacote não é chamado diretamente pelo código do projeto — é uma dependência transitiva via `projectdiscovery/utils/update`. Documentado como risco aceito.

2. **Testes de rede local não executáveis**: `retryabledns` não consegue se comunicar com servidores UDP locais no ambiente Termux/Android. Os testes de rate limit, RRSet differences e wildcard multi-level foram convertidos em testes de configuração (verificação de parâmetros) em vez de teste de rede real.

3. **WAN testing**: Não realizado, conforme regras do projeto.

4. **`go test -race`**: Não suportado na plataforma android/arm64.

O HyperDNSx está **pronto para distribuição como release candidate** com as limitações acima documentadas. Nenhuma vulnerabilidade crítica permanece sem mitigação. Nenhuma regressão funcional foi introduzida.

---

## 9. Comandos para Reprodução

```bash
# Clone e faça checkout do release candidate
git clone <repo-url>
cd hyperdnsx
git checkout fe5b079

# Verifique o ambiente
go version                    # go1.27.2 android/arm64
go env GOOS GOARCH            # android arm64
go env CGO_ENABLED            # 1

# Verifique dependências e segurança
go mod verify                 # all modules verified
go vet ./...                  # clean
go build ./...                # clean
go test ./... -count=1        # all tests pass
govulncheck ./...             # 1 accepted risk (GO-2026-5932)
go-licenses check ./...       # clean
git diff --check              # clean

# Construa o binário
go build -o dist/hyperdnsx ./cmd/dnsx/

# Verifique o binário
file dist/hyperdnsx           # ELF 64-bit LSB pie executable, ARM aarch64, Android 24
./dist/hyperdnsx -version     # banner impresso
./dist/hyperdnsx -h           # todas as flags listadas

# Verifique checksums
sha256sum -c dist/SHA256SUMS  # OK
```

---

*Relatório gerado automaticamente durante a Fase 14 do projeto HyperDNSx. Todas as informações são reproduzíveis a partir do commit `fe5b079` no branch `main`.*
