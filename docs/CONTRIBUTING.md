# Como contribuir

## Branches

- `main`: producao. `develop`: integracao (branch padrao). `release/vX.Y.Z`: versao em homologacao (staging).
- Trabalhe em `feature/<assunto>` a partir da `develop` e abra PR para `develop` (squash).
- Correcao urgente em producao: `hotfix/<assunto>` a partir da `main`, PR para `main`, depois back-merge na `develop`.

## Commits e titulo de PR

Padrao Conventional Commits: `feat: ...`, `fix: ...`, `docs: ...`, `chore: ...`. Use `feat!:` ou `BREAKING CHANGE:` para mudanca incompativel. O titulo do PR vira a mensagem do squash e define a versao.

## Cortando uma release

1. Actions > **Cut release** > Run workflow na branch `develop` (bump `auto` ou forcado).
2. O workflow cria `release/vX.Y.Z`, atualiza `VERSION`/`CHANGELOG.md` e publica em **staging**.
3. Usuario final valida em staging; correcoes entram na release via PR.
4. Abra o PR: `gh pr create --base main --head release/vX.Y.Z --title "chore(release): vX.Y.Z"` e faca merge com **merge commit**.
5. O workflow **Release** cria a tag e a GitHub Release, pede aprovacao do PO para producao e abre o PR de back-merge `main` -> `develop`.
6. Faca o merge do back-merge com **Create a merge commit** (nunca squash) **antes** da proxima release.

## Especificações e decisões

Funcionalidades nascem como spec em `.specs/features/` (fluxo Specify, Design, Tasks, Implement, Validate) e decisões de arquitetura viram ADR em `docs/adr/`. Veja o `CLAUDE.md` da raiz e `docs/development/ai-environment.md`.

## Segredos

O repositorio e publico: nunca comite segredos. Use GitHub Secrets por ambiente (`staging`, `production`).
