# HyperDNSx — Auditoria de Licenças

> Gerado com `go-licenses v2.0.1` em 16 de outubro de 2026.
> Escopo: `./...` (todos os pacotes do módulo, incluindo testes via `--include_tests`).
> Ferramenta: github.com/google/go-licenses/v2 v2.0.1.

## Resultado da execução

| Comando | Código de saída | Saída |
|---------|---------------|-------|
| `go-licenses report ./...` | 0 | `docs/research/licenses.csv` (107 dependências) |
| `go-licenses report ./...` (stderr) | 0 | `docs/research/licenses-errors.log` (19 linhas) |
| `go-licenses check --include_tests ./...` | 0 | `docs/research/licenses-check.txt` |

## Dependências por licença

| Licença | Quantidade | Observação |
|---------|-----------|------------|
| MIT | 62 | Permissiva, compatível com MIT do projeto |
| BSD-3-Clause | 25 | Permissiva, compatível |
| Apache-2.0 | 12 | Permissiva, compatível |
| ISC | 3 | Permissiva, compatível |
| BSD-2-Clause | 3 | Permissiva, compatível |
| Unlicense | 1 | Domínio público / permissiva |
| MPL-2.0 | 1 | Permissiva fraca (copyleft fraco em arquivos), compatível |
| **Total** | **107** | |

## Análise de risco de licença

- **Não há dependências GPL, LGPL ou AGPL.** Isso elimina o risco de copyleft forte que poderia forçar a divulgação do código-fonte do HyperDNSx (licenciado sob MIT).
- **Todas as licenças são permissivas.** O projeto pode ser distribuído sob MIT sem obrigação de abrir o código das dependências (exceto MPL-2.0, que exige divulgação de modificações nos arquivos cobertos por ela, mas não em todo o projeto).
- **Não há dependências com licença desconhecida.** Todas as 107 dependências tiveram a licença detectada automaticamente pelo `go-licenses`.

## Dependências com código não-Go (avisos do `go-licenses check`)

O `go-licenses check` emitiu avisos (W1017) para pacotes que contêm assembly não-Go, impedindo a inspeção de dependências transitivas dentro desses arquivos. Isso não afeta a detecção de licença, apenas a varredura de dependências adicionais:

- `golang.org/x/sys/unix` (asm_linux_arm64.s)
- `github.com/modern-go/reflect2` (vários arquivos .s)
- `github.com/klauspost/compress/zstd/internal/xxhash` (xxhash_arm64.s)
- `golang.org/x/crypto/chacha20` (chacha_arm64.s)
- `golang.org/x/sys/cpu` (cpu_arm64.s)
- `github.com/golang/snappy` (decode_arm64.s, encode_arm64.s)
- `github.com/zcalusic/sysinfo/cpuid` (cpuid_default.s)

Esses são arquivos de assembly específicos de arquitetura, comuns em bibliotecas de criptografia e compressão. Não representam risco de licença.

## Limitações

- A detecção de licença pelo `go-licenses` baseia-se em arquivos LICENSE presentes nos módulos e heurísticas. Não houve leitura manual de cada arquivo LICENSE.
- Licenças de dependências transitivas indiretas não listadas diretamente no `go.mod` podem não estar cobertas com o mesmo nível de detalhe, embora o `go-licenses report ./...` tente incluí-las.
- A auditoria não cobre licenças de ferramentas de build/CI (GitHub Actions, Docker, etc.), apenas dependências Go do módulo.

## Conclusão

A auditoria de licenças está **concluída**. Todas as 107 dependências Go têm licenças identificadas e permissivas. Não há dependências GPL/LGPL/AGPL. O projeto pode ser distribuído sob licença MIT sem conflitos de copyleft forte.
