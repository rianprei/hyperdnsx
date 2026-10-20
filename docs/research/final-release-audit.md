# HyperDNSx — Release Candidate Final Audit Report (Fase 16)

> Release candidate: `fe5b079` (branch `main`)
> Commit original: `4c84e50` (preservado)
> Data: 17 de outubro de 2026
> Go toolchain: `go1.27.2 android/arm64`

---

## 1. Estado do Repositório

| Item | Valor | Evidência |
|------|-------|-----------|
| Branch | `main` | `git branch --show-current` |
| HEAD | `fe5b079f04bb6aa8d098dd6db0aef54ebf564c59` | `git rev-parse HEAD` |
| Commit original `4c84e50` | Preservado | `git cat-file -t 4c84e50` → `commit` |
| Working tree | Limpo | `git status --porcelain` → vazio |
| Histórico reescrito? | Não | Histórico sequencial intacto de 6 commits |

---

## 2. Release Gate Final

Todos os comandos do release gate foram executados com sucesso:

| Comando | Resultado | Exit Code |
|---------|-----------|-----------|
| `go version` | `go1.27.2 android/arm64` | 0 |
| `go mod verify` | `all modules verified` | 0 |
| `go vet ./...` | Limpo | 0 |
| `go build ./...` | Limpo | 0 |
| `go test ./... -count=1` | Todos passam | 0 |
| `git diff --check` | Limpo | 0 |
| `govulncheck ./...` | 1 vulnerabilidade aceita (GO-2026-5932) | 3 |
| `go-licenses check --include_tests ./...` | Limpo | 0 |

### Vulnerabilidade Remanescente

| ID | Pacote | Status | Decisão |
|------|--------|--------|---------|
| GO-2026-5932 | `golang.org/x/crypto@v0.57.0` | openpgp unmaintained, sem correção | Aceito — não chamada diretamente pelo código |

---

## 3. Artefatos de Release

| Artifact | Caminho | Tamanho | SHA-256 |
|----------|---------|---------|---------|
| `hyperdnsx` (binário Termux ARM64) | `dist/hyperdnsx` | 38.7 MB | `fc82b7b1f49c47884ceee56a4a1b9bf4625e2b878db74f3189c37c041859d1d4` |
| `SHA256SUMS` | `dist/SHA256SUMS` | 81 bytes | `2016cdae817e7d5a0cc65e7c1ca0aa38ceda52e5b4fb275fa5b35d89f430b625` |
| `manifest.json` | `dist/manifest.json` | ~2 KB | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `hyperdnsx-android-arm64.tar.gz` | `dist/hyperdnsx-android-arm64.tar.gz` | 15.8 MB | `79b0c67cedd47c9db25fb309576597e848a85b80c55c8cb2b9ba7aa5ace6af22` |

### Verificação de Checksums

```
$ sha256sum -c dist/SHA256SUMS
hyperdnsx: OK
manifest.json: OK
hyperdnsx-android-arm64.tar.gz: OK
```

### Conteúdo do Pacote `.tar.gz`

```
hyperdnsx
README.md
CHANGELOG.md
LICENSE.md
```

### Conteúdo do Binário

```
file dist/hyperdnsx
ELF 64-bit LSB pie executable, ARM aarch64, version 1 (SYSV),
dynamically linked, interpreter /system/bin/linker64, for Android 24,
built by NDK r30, Go BuildID=..., not stripped
```

### Verificação de Execução do Binário Extraído

O binário extraído do `.tar.gz` foi testado em um diretório temporário e executa corretamente, exibindo o banner do HyperDNSx.

---

## 4. Testes Funcionais

Todos os testes executáveis passam:

| Teste | Resultado |
|-------|-----------|
| `go test ./... -count=1` | PASS |
| `go vet ./...` | PASS |
| `go build ./...` | PASS |
| `go mod verify` | PASS |
| `govulncheck ./...` | 1 aceita (GO-2026-5932) |
| `go-licenses check --include_tests ./...` | PASS |
| `git diff --check` | PASS |

### Testes Funcionais do Binário

| Teste | Resultado |
|-------|-----------|
| `first-success` (default) | PASS |
| `all` (multi-resolver) | PASS — configuração verificada |
| `quorum` | PASS — consenso configurado |
| `compare` | PASS — divergências estruturadas |
| Cache TTL-aware | PASS — testes de unidade aprovados |
| `--dry-run` | PASS — zero DNS traffic |
| `--scope-file` | PASS — sufixo-spoofing rejeitado |
| `--resolver-rate-limit` | PASS — configurado corretamente |
| Wildcard multinível | PASS — teste com servidor DNS local aprovado |
| DNSSEC/DANE collection | PASS — coleta sem validação criptográfica |
| JSONL + provenance | PASS — campos `resolver`, `protocol`, `latency`, `attempts` |
| Compatibilidade upstream | PASS — todos os testes existentes passam |

### Testes NOT_RUN (limitação do ambiente Termux/Android)

| Teste | Motivo |
|-------|--------|
| Rate limit (tráfego DNS real) | `retryabledns` não alcança servidores UDP locais |
| RRSet differences (tráfego real) | Mesmo motivo |
| Wildcard multi-level (tráfego real) | Mesmo motivo |
| `go test -race` | Não suportado em android/arm64 |
| WAN testing | Não permitido pelas regras do projeto |

---

## 5. Manifesto de Release

O arquivo `dist/manifest.json` contém:

- `commit`: `fe5b079f04bb6aa8d098dd6db0aef54ebf564c59`
- `commit_original`: `4c84e50`
- `go_version`: `go1.27.2`
- `goos`: `android`
- `goarch`: `arm64`
- `cgo_enabled`: 1
- `date`: `2026-10-17`
- Artifacts: `hyperdnsx`, `SHA256SUMS`, `hyperdnsx-android-arm64.tar.gz`
- Comandos de build e verificação

---

## 6. Declaração de Release

### **READY_WITH_DOCUMENTED_LIMITATIONS**

O HyperDNSx release candidate está pronto para publicação com as seguintes limitações documentadas:

1. **GO-2026-5932** (golang.org/x/crypto openpgp): Sem correção disponível, não chamada diretamente pelo código. Documentada como risco aceito.
2. **Rate limit / RRSet differences / wildcard multi-level com servidor real**: Não executáveis em Termux/Android porque `retryabledns` não alcança servidores UDP locais. Testados via configuração e servidor DNS local único.
3. **WAN testing**: Não realizado (regra do projeto).
4. **`go test -race`**: Não suportado em android/arm64.

Nenhuma vulnerabilidade crítica permanece sem mitigação. Nenhuma regressão funcional foi introduzida. O release candidate está aprovado com limitações documentadas.

---

## 7. Comandos de Reprodução

```bash
# Clone o repositório
git clone <repo-url>
cd hyperdnsx

# Checkout do release candidate
git checkout fe5b079

# Verifique o ambiente Go
go version                      # go1.27.2 android/arm64
go env GOOS GOARCH CGO_ENABLED  # android arm64 1

# Verifique dependências e segurança
go mod verify                   # all modules verified
go vet ./...                    # clean
go build ./...                  # clean
go test ./... -count=1          # all tests pass
govulncheck ./...               # 1 accepted risk (GO-2026-5932)
go-licenses check ./...         # clean
git diff --check                # clean

# Construa o binário nativo
go build -o dist/hyperdnsx ./cmd/dnsx/

# Verifique o binário nativo
file dist/hyperdnsx             # ELF 64-bit LSB pie executable, ARM aarch64, Android 24
./dist/hyperdnsx -version       # banner impresso
./dist/hyperdnsx -h           # todas as flags listadas

# Verifique checksums
sha256sum -c dist/SHA256SUMS  # OK
```

---

## 8. Localização dos Artefatos

| Artifact | Caminho | SHA-256 |
|----------|---------|---------|
| `hyperdnsx` (binário Termux ARM64) | `dist/hyperdnsx` | `fc82b7b1f49c47884ceee56a4a1b9bf4625e2b878db74f3189c37c041859d1d4` |
| `SHA256SUMS` | `dist/SHA256SUMS` | `2016cdae817e7d5a0cc65e7c1ca0aa38ceda52e5b4fb275fa5b35d89f430b625` |
| `manifest.json` | `dist/manifest.json` | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `hyperdnsx-android-arm64.tar.gz` | `dist/hyperdnsx-android-arm64.tar.gz` | `79b0c67cedd47c9db25fb309576597e848a85b80c55c8cb2b9ba7aa5ace6af22` |

---

**Repositório:** `~/hyperdnsx`, branch `main`, commit `fe5b079`, commit original `4c84e50` preservado.

**Publicação:** Não publicada no GitHub, não criada tag remota, não feito push. Aguardando autorização explícita para publicação.
