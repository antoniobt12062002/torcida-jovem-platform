# Como contribuir

## Branches

- `main`: producao. `develop`: integracao (branch padrao). `release/vX.Y.Z`: versao em homologacao (staging).
- Trabalhe em `feature/<assunto>` a partir da `develop` e abra PR para `develop` (squash).
- Correcao urgente em producao: `hotfix/<assunto>` a partir da `main`, PR para `main`, depois back-merge na `develop`.

## Politica de merge

Enquanto o repositorio tiver um unico mantenedor/desenvolvedor, o bypass administrativo esta autorizado para o merge de PRs, desde que: o CI esteja verde; a validacao tecnica pelo processo SDD tenha sido feita; as decisoes arquiteturais tenham sido revisadas antes do merge; e o historico fique registrado no PR. Todo uso do bypass e registrado em comentario no PR, informando que ocorreu porque o repositorio possui um unico mantenedor. A politica sera reavaliada quando novos colaboradores ingressarem.

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

## Ambiente local

Requisitos: Go (versão do `api/go.mod`), Node com pnpm, Docker. Versões de ferramentas em [tooling](development/tooling.md).

```bash
cp .env.example .env                                              # nada aqui é segredo; o .env é ignorado pelo git
docker compose up -d db                                           # PostgreSQL 16 com os papéis tj_owner e tj_app
docker compose --profile tools run --rm migrate                   # migrações, aplicadas por tj_owner
```

Para recriar o banco do zero: `docker compose down -v`. A API e os comandos leem **variáveis de ambiente**, não o arquivo `.env`: exporte as variáveis (ou use `set -a; source .env; set +a`) antes de rodar.

| Variável | Uso |
|---|---|
| `DATABASE_URL` | Conexão da aplicação, com o papel `tj_app` (sem DDL); obrigatória |
| `MIGRATE_DATABASE_URL` | Conexão do dono, só para migrações |
| `APP_ENV` | `development`, `staging` ou `production`; ausente vale `production` |
| `ALLOWED_ORIGINS`, `COOKIE_SECURE` | Origens do navegador aceitas; `COOKIE_SECURE=false` só em `development` |
| `AUTH_HASH_KEY` | Chave do HMAC dos e-mails, com 32 bytes ou mais; obrigatória fora de `development` (`openssl rand -base64 48`) |
| `EMAIL_PROVIDER` | `log` em desenvolvimento: registra só o domínio do destinatário e o assunto, nunca o corpo, o link ou o token; `disabled` não envia nada |
| `APP_BASE_URL` | Base dos links enviados por e-mail |

A lista completa, com comentários, está em `.env.example`.

## Rodando a API

```bash
cd api
export DATABASE_URL="postgres://tj_app:tj_app_dev@localhost:5432/tj?sslmode=disable"
APP_ENV=development COOKIE_SECURE=false go run ./cmd/api
curl -s localhost:8080/healthz                                    # {"status":"ok"}
```

Na partida a API sincroniza os papéis e as permissões com o código (idempotente, auditado como `rbac.sync`). Ela não roda migrações.

## Primeiro administrador

Depois das migrações, crie o primeiro administrador. A senha vai **só** pela variável `BOOTSTRAP_ADMIN_PASSWORD` (o comando não tem opção de senha, para não aparecer na lista de processos nem no histórico do shell) e nunca é impressa:

```bash
cd api
export DATABASE_URL="postgres://tj_app:tj_app_dev@localhost:5432/tj?sslmode=disable"
BOOTSTRAP_ADMIN_PASSWORD='uma-senha-com-10-ou-mais-caracteres'   go run ./cmd/bootstrap-admin --email admin@local.test --name "Administrador Local"
# --role PRESIDENTE cria um presidente em vez do ADMIN_SISTEMA (padrão)
```

O comando só cria o usuário enquanto não existe nenhum vínculo administrativo ativo; nas execuções seguintes sai com código diferente de zero e não cria nada. O primeiro login exige a troca da senha (`POST /api/v1/auth/password`). Em produção, rode-o uma única vez, com a senha vinda do gerenciador de segredos do ambiente.

## Testes

Os testes que usam banco ficam atrás da tag `integration` e exigem Docker (o PostgreSQL 16 sobe por testcontainers, sem depender do banco local):

```bash
cd api
go vet ./... && go test ./...                    # unidade, sem Docker
go test -tags=integration ./...                  # inclui integração (use -p 2 se o Docker ficar lento)
cd ../web
pnpm install && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build
```

Se um teste de integração falhar com "failed to create Docker provider" ou "falha ao iniciar via Docker", é o Docker (comum no Docker Desktop do Windows), não o código: rode de novo. O teste `api/internal/architecture_test.go` verifica as [fronteiras entre módulos](architecture/domain-boundaries.md).

## Contrato da API

O contrato OpenAPI em `api/openapi/` é a fonte das rotas. Depois de editar os arquivos `.yaml`:

```bash
(cd web && pnpm lint:api && pnpm gen:api)   # lint e tipos TypeScript
(cd api && go generate ./...)                # interfaces e modelos Go
```

A API não inicia se uma rota e o contrato divergirem, e o CI falha se o código gerado não corresponder ao contrato.

## Segredos

O repositorio e publico: nunca comite segredos. Use GitHub Secrets por ambiente (`staging`, `production`).
