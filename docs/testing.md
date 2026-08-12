# Fluxo de testes

Como o GHP organiza e roda testes: unit, integration, e2e e cobertura.

## Onde cada teste mora

Go não tem o conceito de "projeto de teste" configurável como o Vitest — a organização vem de duas convenções diferentes, uma pra unit e outra pra integration/e2e:

| Tipo | Onde fica | Como roda | Por quê |
| --- | --- | --- | --- |
| Unit | `*_test.go` junto do código (ex.: `cmd/ghp/main_test.go`) | `go test ./...` (padrão, sem flag) | Mesmo pacote do código testado — acessa funções não-exportadas direto, sem precisar exportar nada só pra testar. Precisa rodar sempre, sem exigir flag. |
| Integration | `test/integration/*_test.go` | `go test ./test/integration/... -tags=integration` | Pacote externo, isolado com build tag — só roda quando pedido explicitamente, nunca sem querer dentro de `go test ./...`. |
| E2E | `test/e2e/*_test.go` | `go test ./test/e2e/... -tags=e2e` | Mesma lógica do integration: pacote externo + build tag. Roda contra o binário `ghp` já compilado. |

Não existe (e não deve existir) uma pasta `test/unit/` — colocar o teste unitário longe do código que ele testa vai contra a convenção idiomática do Go.

## Rodando localmente

Via `Makefile` (mesmos comandos que o CI roda):

```bash
make test-unit         # go test ./... -race
make test-integration  # go test ./test/integration/... -tags=integration -race
make test-e2e          # builda o binário e roda test/e2e/... -tags=e2e
make test-coverage     # unit + integration com coverage, falha se < 90%
```

`test-e2e` builda o binário (`bin/ghp`) antes de rodar, porque os testes de e2e chamam o binário compilado, não o código-fonte diretamente.

## Cobertura

- Mínimo exigido: **90%**, calculado sobre unit + integration combinados (`-coverpkg=./...` garante que a cobertura conta mesmo quando o teste mora num pacote externo como `test/integration`).
- `make test-coverage` roda local e falha se ficar abaixo do mínimo — é o mesmo cálculo que o job `coverage` do `ci.yml` faz.
- O relatório também sobe pro Codecov (`codecov.yml` define os mesmos 90% como target de `project` e `patch`).

```bash
go tool cover -func=coverage.out   # ver cobertura por função
go tool cover -html=coverage.out   # ver cobertura por linha, no navegador
```

## No CI (`.github/workflows/ci.yml`)

Cada tipo de teste roda num job separado, todos exigidos pelo job `gate`:

- `unit-tests` — `go test ./... -race`
- `integration-tests` — `go test ./test/integration/... -tags=integration -race`
- `e2e-tests` — builda o binário, depois `go test ./test/e2e/... -tags=e2e`
- `coverage` — roda unit+integration com `-coverprofile`, checa o mínimo de 90%, sobe pro Codecov

## Convenções ao escrever um teste

- **Table-driven test** é o padrão pra múltiplos casos na mesma função (ver `cmd/ghp/main_test.go` → `TestRun`) — um `[]struct{...}` com `name`, entradas e saídas esperadas, iterado com `t.Run(tt.name, ...)`.
- Testar contra `io.Writer`/entrada explícita, nunca contra `os.Stdout`/`os.Args` globais direto — é o que torna a função testável sem hacks (ver `run(args []string, stdout io.Writer) int` em `cmd/ghp/main.go`).
- `t.Helper()` em toda função auxiliar de teste, pra apontar o erro na linha certa.
- Testes de integration/e2e ainda **não existem de verdade** — os arquivos em `test/integration/` e `test/e2e/` hoje só têm um `TestPlaceholder` com `t.Skip(...)`, apontando pra issue do Linear que vai substituir o skip por teste real (GHP-13, GHP-14, GHP-15).

## Relacionado

- [`git-workflow.md`](./git-workflow.md) — todo PR precisa do job `gate` (que inclui os quatro tipos de teste acima) verde antes do merge.
