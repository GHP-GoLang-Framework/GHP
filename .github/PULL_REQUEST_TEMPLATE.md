## Resumo

<!-- O que mudou e por quê, em 1-3 bullets -->

-

## Issue relacionada

<!-- Referencie a issue do Linear, se houver -->
Refs: GHP-

## Como testar

<!-- Passos pra quem for revisar reproduzir/validar a mudança -->

- [ ]

## Checklist

- [ ] `gofmt -l ./src` sem saída e `go vet ./src/...` limpos; `go test -short ./src/... -race` e `go test ./src/test/integration/... -race` passam localmente
- [ ] Testes cobrindo o caso novo (o CI exige cobertura mínima de 90%)
- [ ] Título do PR segue [Conventional Commits](https://www.conventionalcommits.org/pt-br/) (`tipo(escopo): descrição`)
- [ ] Documentação atualizada, se necessário (`docs/`, `README.md`)

## Impacto

- Breaking change (sintaxe `.ghp` ou API do `runtime`)? sim/não
- Precisa de ação manual depois do merge (migração, nova config, secret)? sim/não
