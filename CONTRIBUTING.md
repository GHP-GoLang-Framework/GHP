# Contribuindo com o GHP

Obrigado pelo interesse em contribuir. Este documento cobre o essencial pra começar — os detalhes de cada parte estão linkados abaixo.

## Antes de começar

- **Fluxo de git e PR**: leia [`docs/git-workflow.md`](docs/git-workflow.md). Resumo: branch a partir da `main`, PR de volta pra `main`, CI verde + aprovação obrigatória antes do merge. Sem exceção — nem o mantenedor commita direto na `main`.
- **Testes**: leia [`docs/testing.md`](docs/testing.md). Cobertura mínima de 90% (unit + integration) é obrigatória no CI.
- **Sintaxe do GHP**: [`docs/template.ghp`](docs/template.ghp) é a referência da linguagem que o parser/codegen vão implementar.

## Configurando o ambiente local

Pré-requisitos: Go (versão do [`go.mod`](go.mod)) e Node.js (só pra rodar os git hooks — o framework em si é 100% Go).

```bash
git clone https://github.com/GHP-GoLang-Framework/GHP.git
cd GHP
npm install   # instala husky e commitlint, configura os git hooks
```

Isso já deixa o `pre-commit` (gofmt + `go vet` nos arquivos alterados) e o `commit-msg` (commitlint) ativos.

## Fazendo uma mudança

```bash
git checkout main
git pull --ff-only origin main
git checkout -b feat/nome-da-mudanca

# ... edita, commita ...

gofmt -l ./src                    # sem saída = formatado
go vet ./src/...                  # análise estática
go test -short ./src/... -race    # testes unit (pula integration/e2e)
go test ./src/test/integration/... -race   # o que o CI vai rodar de qualquer forma

git push -u origin feat/nome-da-mudanca
gh pr create --base main --title "feat(escopo): descrição no imperativo"
```

## Mensagens de commit

[Conventional Commits](https://www.conventionalcommits.org/pt-br/), validado automaticamente pelo `commit-msg` hook. Formato:

```
tipo(escopo opcional): descrição curta no imperativo
```

Tipos aceitos: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`. O título do PR segue o mesmo padrão — é revalidado no CI e vira a mensagem do commit de squash no merge.

Escopo recomendado: o domínio/pacote afetado (`parser`, `codegen`, `cli`, `devserver`, `runtime`, `routes`, `docker`, `docs`...), nunca o nome do arquivo.

## Qualidade de código

- `gofmt` — formatação, corrigida automaticamente pelo `pre-commit`.
- `go vet` — roda no `pre-commit` e no CI.
- `golangci-lint` — roda no CI (job `lint`); rode `golangci-lint run ./src/...` localmente se quiser adiantar.
- Testes cobrindo o que for adicionado — ver [`docs/testing.md`](docs/testing.md) pras convenções (table-driven tests, testar contra `io.Writer`/entrada explícita em vez de globais).

## Abrindo o PR

- Base sempre `main`.
- Título em Conventional Commit.
- O job `gate` (lint, testes unit/integration/e2e, cobertura ≥90%, build, verificação de vulnerabilidades e de segredos) precisa passar.
- Precisa de aprovação de [@castrogusttavo](https://github.com/castrogusttavo) antes do merge.
- Merge é por squash — a branch pode ter quantos commits de WIP forem necessários, só o commit final na `main` importa.

## Reportando bugs ou sugerindo funcionalidades

Abra uma [issue no GitHub](https://github.com/GHP-GoLang-Framework/GHP/issues) descrevendo o problema ou a sugestão. Se for um bug, inclua um exemplo mínimo que reproduza o comportamento.

## Dúvidas

Se algo neste documento ou nos linkados estiver desatualizado ou confuso, abra uma issue ou já resolva num PR — a documentação é código como qualquer outra parte do projeto.
