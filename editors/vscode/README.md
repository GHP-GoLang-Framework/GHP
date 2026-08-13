# GHP — extensão VSCode

Realce de sintaxe para arquivos `.ghp`: HTML com Go real embutido.

## O que faz

- **Realce de sintaxe** — HTML normal + as tags `<go ...>`, com o Go dentro delas colorido como Go de verdade
- **Auto-fechamento** — digitar `<go:` ou `<go=` insere o `>` automaticamente
- **Snippets** — `go:if`, `go:for`, `go:switch`, `go:import` e outros expandem o bloco completo, já com a tag de fechamento
- **Indentação automática** dentro de `<go:if>`, `<go:for>` e `<go:switch>`

## Tags reconhecidas

| Tag | O que é |
| --- | --- |
| `<go:import ("fmt")>` | Imports Go da página |
| `<go ... >` | Bloco de código Go (statement), pode ser multi-linha |
| `<go= expressao >` | Renderiza o valor da expressão no HTML |
| `<go:if cond>` / `<go:else>` / `</go:if>` | Condicional |
| `<go:switch v>` / `<go:case x>` / `<go:default>` / `</go:switch>` | Switch |
| `<go:for expr>` / `</go:for>` | Laço |

A referência da sintaxe é [`docs/template.ghp`](../../docs/template.ghp) na raiz do repositório.

## Desenvolvimento

```bash
cd editors/vscode
npm install
npm test        # roda a gramática contra casos reais e valida os escopos
```

Para testar no editor: abra a pasta `editors/vscode` no VSCode e pressione `F5` — abre uma janela com a extensão carregada. Abra qualquer `.ghp` (ex.: `docs/template.ghp`) para ver o realce.

Para empacotar: `npx @vscode/vsce package`.

## Limitações conhecidas

**Operador `>` dentro de tags.** A tag fecha no primeiro `>`, então uma condição como `<go:if a > b>` é realçada incorretamente (o realce termina no `>` do operador). Isso não é um bug da extensão e sim uma ambiguidade da própria sintaxe, que precisa ser resolvida no parser — enquanto isso, `>=` e `>` em condições podem ser escritos como `<go:if b < a>` ou movidos para um bloco `<go ... >` anterior.

**Dependência da gramática Go.** O realce do conteúdo das tags depende da gramática `source.go`, que vem embutida no VSCode. Se por algum motivo ela não estiver disponível, as tags GHP deixam de ser reconhecidas por completo (o TextMate descarta a regra inteira quando um `include` não resolve — não há degradação parcial).

**Sem LSP.** Não há autocomplete de símbolos Go, "ir para definição" nem diagnósticos dentro das tags. Isso exigiria um servidor de linguagem que entenda o mapeamento `.ghp` → Go gerado — um próximo passo natural depois que o parser e o codegen existirem.
