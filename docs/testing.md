# Fluxo de testes

Como o GHP organiza e roda testes: unit, integration, e2e e cobertura — tudo via Go nativo, sem build tags.

## Onde cada teste mora

A separação unit vs integration/e2e não usa build tag nem Makefile: Go já resolve tudo com `go test ./src/...`. A separação rápido/lento é feita com o flag nativo `-short` (`testing.Short()`):

| Tipo | Onde fica | Como roda | Por quê |
| --- | --- | --- | --- |
| Unit | `*_test.go` junto do código (ex.: `src/cmd/ghp/main_test.go`) | `go test -short ./src/...` | Mesmo pacote do código testado — acessa funções não-exportadas direto, sem precisar exportar nada só pra testar. Roda em modo curto, sempre. |
| Integration | `src/test/integration/*_test.go` | `go test ./src/...` | Pacote externo — exercita múltiplos pacotes juntos ou chama `go build` de verdade. Pula em modo `-short`. |
| E2E | `src/test/e2e/*_test.go` | `go test ./src/...` (ou `go test ./src/test/e2e/...` com `GHP_BINARY` setado) | Roda contra o binário `ghp` já compilado. Pula em modo `-short`. |

Não existe (e não deve existir) uma pasta `src/test/unit/` — colocar o teste unitário longe do código que ele testa vai contra a convenção idiomática do Go.

O padrão em todo teste de integration/e2e é:

```go
if testing.Short() {
	t.Skip("skipped in short mode (go test -short)")
}
```

Assim `go test -short ./src/...` roda só o rápido e `go test ./src/...` roda tudo, sem precisar de tag nem Makefile.

## Rodando localmente

Sem Makefile — comandos Go puros (os mesmos que o CI roda):

```bash
gofmt -l ./src                    # sem saída = formatado
go vet ./src/...                  # análise estática
go test -short ./src/... -race    # só testes rápidos (pula integration/e2e)
go test ./src/test/integration/... -race
go test ./src/test/e2e/...        # precisa do binário compilado (ver abaixo)
go build -o bin/ghp ./src/cmd/ghp # builda o binário usado pelos testes e2e
```

`go test ./src/... -race` roda tudo, incluindo integration/e2e.

Os testes de e2e vão chamar o binário `ghp` compilado (via `GHP_BINARY`), não o código-fonte diretamente — por isso `go build -o bin/ghp ./src/cmd/ghp` antes. Hoje eles são placeholders que pulam sempre (GHP-14/15).

## Cobertura

- Mínimo exigido: **90%**, calculado sobre `./src/...` inteiro (`-coverpkg=./src/...` garante que a cobertura conta mesmo quando o teste mora num pacote externo como `src/test/integration`).
- Comando local:
  `go test ./src/... -coverprofile=coverage.out -covermode=atomic -coverpkg=./src/...`
  e depois checar com `go tool cover -func=coverage.out | tail -1` — é o mesmo cálculo que o job `coverage` do `ci.yml` faz.
- O relatório também sobe pro Codecov (`codecov.yml` define os mesmos 90% como target de `project` e `patch`).

```bash
go tool cover -func=coverage.out   # ver cobertura por função
go tool cover -html=coverage.out   # ver cobertura por linha, no navegador
```

## No CI (`.github/workflows/ci.yml`)

Cada tipo de teste roda num job separado, todos exigidos pelo job `gate`:

- `unit-tests` — `go test -short ./src/... -race` (integration/e2e pulam via `-short`)
- `integration-tests` — `go test ./src/test/integration/... -race`
- `e2e-tests` — builda o binário, depois `go test ./src/test/e2e/...` com `GHP_BINARY`
- `coverage` — roda `go test ./src/...` com `-coverprofile`, checa o mínimo de 90%, sobe pro Codecov

## Convenções ao escrever um teste

- **Table-driven test** é o padrão pra múltiplos casos na mesma função (ver `src/cmd/ghp/main_test.go` → `TestRun`) — um `[]struct{...}` com `name`, entradas e saídas esperadas, iterado com `t.Run(tt.name, ...)`.
- Testar contra `io.Writer`/entrada explícita, nunca contra `os.Stdout`/`os.Args` globais direto — é o que torna a função testável sem hacks (ver `run(args []string, stdout io.Writer) int` em `src/cmd/ghp/main.go`).
- `t.Helper()` em toda função auxiliar de teste, pra apontar o erro na linha certa.
- Teste de integration/e2e **nunca roda em modo `-short`**: começa com o guard `if testing.Short() { t.Skip(...) }`.
- Testes de integration/e2e ainda **não existem de verdade** — os arquivos em `src/test/integration/` e `src/test/e2e/` hoje só têm um `TestPlaceholder` com `t.Skip(...)`, apontando pra issue do Linear que vai substituir o skip por teste real (GHP-13, GHP-14, GHP-15).

## Relacionado

- [`git-workflow.md`](./git-workflow.md) — todo PR precisa do job `gate` (que inclui os quatro tipos de teste acima) verde antes do merge.
