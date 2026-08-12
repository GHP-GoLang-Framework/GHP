# Fluxo de Git e Pull Requests

Como contribuir com o GHP: branch, commit, PR e review.

## Regra geral

- `main` é protegida — ninguém dá push direto nela.
- Toda mudança nasce como uma branch a partir da `main` e volta como Pull Request.
- Todo PR precisa passar no CI **e** ser aprovado por [@castrogusttavo](https://github.com/castrogusttavo) antes do merge.

## Passo a passo

1. Clone e sincronize a `main`:

   ```bash
   git clone https://github.com/GHP-GoLang-Framework/GHP.git
   cd GHP
   git checkout main
   git pull --ff-only origin main
   ```

2. Crie uma branch a partir da `main`:

   ```bash
   git checkout -b feat/go-for-tag
   ```

   Nome sugerido: `<tipo>/<descrição-curta>`, usando os mesmos tipos do commit (veja abaixo) — `feat/`, `fix/`, `docs/`, `refactor/`, `test/`, `chore/`, `ci/`, `build/`.

3. Faça commits na branch seguindo [Conventional Commits](https://www.conventionalcommits.org/pt-br/) (enforçado localmente por commitlint + Husky):

   ```
   feat(parser): reconhece a tag <go:for expression>
   fix(codegen): corrige //line directive fora de ordem
   ```

   Tipos aceitos: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`.

4. Rode localmente antes de abrir o PR (é o que o CI vai rodar de qualquer forma):

   ```bash
   make fmt vet test-unit test-integration
   ```

5. Suba a branch e abra o PR contra a `main`:

   ```bash
   git push -u origin feat/go-for-tag
   gh pr create --base main --title "feat(parser): reconhece a tag <go:for expression>"
   ```

   O **título do PR também precisa ser um Conventional Commit** — é validado automaticamente pelo CI.

6. Aguarde o CI (job `gate`: lint, testes unitários, integração, e2e, cobertura mínima de 90%, build, verificação de vulnerabilidades e de segredos) e a aprovação de [@castrogusttavo](https://github.com/castrogusttavo).

7. Depois de aprovado e verde, o PR é mergeado por **squash** — vira um único commit na `main`, com o título do PR como mensagem.

## O que acontece depois do merge

- O merge na `main` dispara o CI de novo e, se verde, publica uma imagem Docker `edge` no GitHub Container Registry — build contínuo, não é uma release.
- Releases de verdade saem de tags de versão (`v1.2.3`), criadas manualmente por [@castrogusttavo](https://github.com/castrogusttavo) — publicam a imagem versionada e o GitHub Release.

## O que não vai funcionar

- Push direto na `main` — bloqueado por proteção de branch.
- PR sem o CI verde.
- PR sem aprovação.
- Mensagens de commit fora do Conventional Commits — o hook `commit-msg` bloqueia localmente antes mesmo do push chegar no CI.
