# Design: Fluxo de branches, versionamento e pipelines (GitHub)

Data: 2026-09-26
Status: Implementado e validado com a release v0.1.0 (ajustes da implementação incorporados abaixo)

## 1. Objetivo

Definir como o repositório `torcida-jovem-platform` é governado no GitHub: modelo de branches, proteções, versionamento e pipelines de CI/CD, para um monorepo com API em Go, front em Next.js e documentação.

## 2. Decisões

| Tema | Decisão |
|---|---|
| Estrutura | Monorepo (`/api`, `/web`, `/docs`) |
| Stack | Go (API) + React/Next.js (front) |
| Hospedagem | Ainda não definida. Sugestão: Vercel (front) + Render ou Fly.io (API + Postgres) |
| Visibilidade | Repositório público (plano gratuito não oferece proteção de branch em repo privado) |
| Versionamento | Semver calculado por Conventional Commits, disparado manualmente |
| Aprovação de produção | PO, via GitHub Environment `production` |

## 3. Modelo de branches

```
feature/* --PR (squash)--> develop --Cut release--> release/vX.Y.Z --PR (merge commit)--> main
hotfix/*  --PR-----------> main --back-merge--> develop
```

- `main`: produção. Protegida.
- `develop`: integração. Protegida e branch padrão do repositório.
- `release/vX.Y.Z`: cortada da `develop`; a mais recente é a versão em staging, onde o usuário final testa. Só recebe correções pontuais via PR.
- `hotfix/*`: sai da `main`, volta para a `main` e é mesclada na `develop`.
- Squash de feature em develop (cada PR vira um commit convencional); merge commit de release em main (preserva histórico).

### Proteções (rulesets)

| Branch | Regras |
|---|---|
| `main`, `develop` | Sem push direto, sem force push, sem deleção; PR com 1 aprovação; check obrigatório `ci-gate`; branch atualizada antes do merge |
| `release/*` | Sem force push e sem deleção. Exigir PR aqui bloquearia o `cut-release` (GITHUB_TOKEN); correções entram por PR por convenção |

CODEOWNERS pede revisão do dono e do PO. O GitHub não permite que o autor aprove o próprio PR; enquanto houver um único contribuidor, o bypass fica restrito ao administrador do repositório.

## 4. Versionamento

- Fonte da verdade: arquivo `VERSION` na raiz. Tags `vX.Y.Z` criadas somente na `main`.
- Cálculo: `feat` = minor, `fix` = patch, `!` ou `BREAKING CHANGE` = major. O workflow aceita forçar o tipo de bump.
- `CHANGELOG.md` gerado a partir dos commits da release.
- Título de PR validado no padrão Conventional Commits, pois o squash o usa como mensagem do commit.

## 5. Pipelines

| Workflow | Gatilho | Função |
|---|---|---|
| `ci.yml` | PR para `develop`, `release/*`, `main` | Jobs com filtro de caminho: docs (markdownlint + links), api (`go vet`, `go test`, build), web (lint, typecheck, build). Inclui a validação do título do PR (Conventional Commits). Job agregador `ci-gate` é o único check obrigatório, para que jobs pulados por filtro não travem o PR. Aceita `workflow_dispatch`, usado para rodar o CI na branch de back-merge |
| `cut-release.yml` | Manual, a partir da `develop` | Calcula versão, cria `release/vX.Y.Z`, atualiza `VERSION` e `CHANGELOG.md`, chama `deploy-staging.yml` |
| `deploy-staging.yml` | Push em `release/**` e `workflow_call` | Build e deploy no ambiente `staging` |
| `release.yml` | Push na `main` | Cria tag e GitHub Release, faz deploy em `production` (com aprovação do PO) e abre PR de back-merge `main` → `develop` |

PRs abertos com `GITHUB_TOKEN` (o de back-merge) não disparam `pull_request`; por isso `release.yml` dispara `ci.yml` via `workflow_dispatch` na branch, e o `ci-gate` resultante satisfaz o PR. O back-merge `main` → `develop` deve usar **merge commit** (nunca squash), senão a tag deixa de ser alcançável pela `develop` e o cálculo de versão erra.

Pushes feitos com `GITHUB_TOKEN` não disparam outros workflows; por isso `cut-release` invoca o deploy de staging explicitamente via `workflow_call`, sem PAT ou GitHub App.

### Ambientes

GitHub Environments `staging` e `production`. `production` exige aprovação do PO. Segredos ficam em GitHub Secrets por ambiente, nunca no repositório (o repo é público).

### Estado inicial

Sem código Go ou Next.js, os jobs `api` e `web` e os passos de deploy entram como placeholders que passam, com comentário do que preencher. Funcionam de verdade desde o início: docs-lint, título de PR, `ci-gate`, corte de release, tag e GitHub Release.

## 6. Arquivos a criar

```
.github/workflows/ci.yml
.github/workflows/cut-release.yml
.github/workflows/deploy-staging.yml
.github/workflows/release.yml
.github/scripts/next-version.sh
.github/rulesets/   (JSON dos rulesets aplicados via gh api)
.github/CODEOWNERS
.github/pull_request_template.md
.github/dependabot.yml
.markdownlint.json
VERSION            (0.0.0)
CHANGELOG.md
docs/CONTRIBUTING.md
```

`CODEOWNERS` aguarda o usuário do GitHub do PO; até lá a revisão dele é solicitada manualmente.

## 7. Ordem de implementação

1. Tornar o repo público; criar `develop` a partir da `main`.
2. Publicar workflows e arquivos de suporte em `main` e `develop` por commit direto (o `ci-gate` precisa existir antes de ser exigido).
3. Aplicar rulesets, criar Environments `staging` e `production`, definir `develop` como branch padrão.
4. Redirecionar o PR #1 (docs) para `develop`.
5. Teste de ponta a ponta: cortar `release/v0.1.0`, validar staging placeholder e tag/release na `main` (ou dry-run sem tag).

## 8. Fora de escopo

Escolha do provedor de hospedagem, scaffolding de `/api` e `/web`, testes reais e passos reais de deploy (dependem da definição de stack e hospedagem).

## 9. Lições da validação (v0.1.0)

- O link check do CI falha se o `README` referenciar arquivos ainda não publicados; a correção de uma release entra por PR na própria `release/*`, e o back-merge leva a correção à `develop`.
- Branches `release/*` não podem ser deletadas (ruleset); erros exigem nova versão ou correção via PR.
- Enquanto houver um único contribuidor, merges em `main`/`develop` dependem do bypass de administrador (`pull_request`); o aprovador de `production` é o dono até o PO ser cadastrado.
