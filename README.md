# GHP

[![CI](https://github.com/GHP-GoLang-Framework/GHP/actions/workflows/ci.yml/badge.svg)](https://github.com/GHP-GoLang-Framework/GHP/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/GHP-GoLang-Framework/GHP/branch/main/graph/badge.svg)](https://codecov.io/gh/GHP-GoLang-Framework/GHP)
[![Go Version](https://img.shields.io/github/go-mod/go-version/GHP-GoLang-Framework/GHP)](go.mod)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Templates estilo PHP, com Go de verdade por baixo.

O GHP compila arquivos `.ghp` — HTML com Go real embutido — direto em handlers `net/http`, sem motor de template em runtime. O que está entre as tags é Go de verdade: compila com `go build`, os erros apontam pra linha certa do `.ghp`, e qualquer pacote (padrão, externo ou do seu próprio módulo) pode ser importado.

## Status: em reescrita ativa

O GHP está sendo reescrito do zero com uma sintaxe nova, focada em DX. Hoje o repositório tem só o esqueleto: CLI ainda é um stub (`ghp dev`/`ghp build` não fazem nada de real ainda) e o parser/codegen da sintaxe abaixo ainda não existe. A sintaxe já está definida — ver [`docs/template.ghp`](docs/template.ghp) — e a implementação está em andamento.

## A sintaxe (alvo)

```html
<go:import ("fmt")>

<go
    items := []string{"café", "chá", "suco"}
>

<!doctype html>
<html lang="pt-br">
<head>
  <title><go= fmt.Sprintf("Cardápio (%d itens)", len(items)) ></title>
</head>
<body>
  <ul>
    <go:for _, item := range items>
      <li><go= item ></li>
    </go:for>
  </ul>

  <go:if len(items) == 0>
    <p>Nada no cardápio ainda.</p>
  <go:else>
    <p>Bom apetite!</p>
  </go:if>
</body>
</html>
```

| Tag | O que faz |
| --- | --- |
| `<go:import (...)>` | Importa um ou mais pacotes — padrão, externos ou do seu módulo. |
| `<go ...>` | Bloco de código Go (statement) — pode abrir escopo entre trechos de HTML. |
| `<go= expressão>` | Renderiza o valor de uma expressão no HTML, com escaping automático. |
| `<go:if>` / `<go:else>` / `</go:if>` | Condicional, com os operadores nativos do Go. |
| `<go:switch>` / `<go:case>` / `<go:default>` | Switch. |
| `<go:for>` / `</go:for>` | Laço — qualquer forma de `for`/`range` válida em Go. |

Roteamento é por arquivo: `pages/index.ghp` vira `/`, `pages/blog/[slug].ghp` vira `/blog/{slug}`.

## Instalação

Ainda não há binário publicado nem `go install` disponível — isso faz parte do trabalho em andamento. Por enquanto, buildar a partir do código-fonte:

```bash
git clone https://github.com/GHP-GoLang-Framework/GHP.git
cd GHP
go build -o bin/ghp ./src/cmd/ghp
```

Também há uma imagem Docker publicada a cada merge na `main`: `edge` é o build contínuo (a ponta do desenvolvimento), e uma tag CalVer versionada (`YYYY.MM.DD[.N]`) + `latest` são criadas automaticamente em seguida — todo merge verde já é uma release.

```bash
docker pull ghcr.io/ghp-golang-framework/ghp:latest
docker run --rm ghcr.io/ghp-golang-framework/ghp:latest help
```

## Desenvolvendo o GHP

```bash
git clone https://github.com/GHP-GoLang-Framework/GHP.git
cd GHP
npm install     # configura os git hooks (Husky + commitlint)
gofmt -l ./src  # sem saída = formatado
go vet ./src/...
go test -short ./src/... -race     # unit (pula integration/e2e)
go test ./src/... -race            # tudo, incluindo integration/e2e
go build -o bin/ghp ./src/cmd/ghp  # builda o binário
```

Cobertura (mínimo exigido: 90%): `go test ./src/... -coverprofile=coverage.out -covermode=atomic -coverpkg=./src/...` e depois `go tool cover -func=coverage.out`. Detalhes em [`docs/testing.md`](docs/testing.md).

## Documentação

- [`docs/template.ghp`](docs/template.ghp) — referência completa da sintaxe.
- [`docs/testing.md`](docs/testing.md) — como os testes (unit/integration/e2e) são organizados e rodados.
- [`docs/git-workflow.md`](docs/git-workflow.md) — fluxo de branch, commit e Pull Request.
- [`CONTRIBUTING.md`](CONTRIBUTING.md) — como contribuir.

## Licença

[MIT](LICENSE).
