# TJ Platform

Plataforma de gestão institucional da Torcida Jovem do Campo Mourão Futsal. Monorepo com API em Go, frontend em Next.js e documentação. Contexto de negócio: `docs/00-CONTEXTO-PROJETO.md`, `docs/01-VISAO-PRODUTO.md`, `docs/02-MODULOS-SISTEMA.md`, `docs/finance/FIN-001-VISAO-GERAL.md`.

## Antes de agir

1. Leia `.specs/STATE.md` (decisões `AD-NNN` ativas e snapshot de retomada).
2. Decisões de arquitetura completas: `docs/adr/`. Fronteiras entre módulos: `docs/architecture/domain-boundaries.md`.
3. Nada funcional (banco, API, telas, autenticação) é implementado sem spec aprovada em `.specs/features/`.

## Fluxo (Spec Driven, skill `tlc-spec-driven`)

Specify → Design → Tasks → Implement → Validate. Specs em `.specs/features/<modulo>-<recorte>/`. Documentos de produto em `docs/` são convertidos gradualmente para specs. Skills usadas e quando acioná-las: `docs/development/ai-environment.md`. As skills não são versionadas.

## Arquitetura (resumo; fonte: ADRs)

- Monólito modular em Go: `identity`, `financeiro`, `estoque`, `loja`, `associados`, `eventos`, `acesso`, `transparencia`, `comunicacao`, sobre o núcleo `platform`. Só o módulo dono altera seus dados; comunicação por interface publicada; uma transação por caso de uso; sem dependência circular.
- Dinheiro: inteiro em centavos (`BIGINT`, sufixo `_cents`). Nunca `float`.
- Auditoria append-only na mesma transação; sem exclusão definitiva de dado financeiro (cancelamento, desativação, ajuste extraordinário).
- Autenticação: sessão no servidor com cookie `httpOnly` e proteção CSRF. Sem JWT. Autorização RBAC por permissões, verificada no caso de uso. O associado autenticado é diferente do visitante do portal público.
- Banco: PostgreSQL 16 com `golang-migrate`. Proibido `AutoMigrate`; mudança de schema só por migration revisada.

## Comandos

```bash
# Banco local (papéis: tj_owner migra, tj_app é usado pela API; para recriar o banco: docker compose down -v)
cp .env.example .env
docker compose up -d db && docker compose --profile tools run --rm migrate

# API (api/)
go vet ./... && go test ./... && go build ./...
go test -tags=integration ./...   # exige Docker (testcontainers)
DATABASE_URL="postgres://tj_app:tj_app_dev@localhost:5432/tj?sslmode=disable" APP_ENV=development COOKIE_SECURE=false go run ./cmd/api

# Web (web/)
pnpm install && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm build

# Script de versão
bash .github/scripts/test-next-version.sh
```

## Git e entrega

- `main` (produção) e `develop` (integração, padrão) são protegidas: sem push direto. Trabalhe em `feature/<assunto>` a partir da `develop` e abra PR para `develop` (squash).
- Commits e título de PR em Conventional Commits (`feat:`, `fix:`, `docs:`, `chore:`, `ci:`); o título do PR vira o commit e define a versão.
- Releases: workflow "Cut release" cria `release/vX.Y.Z`; back-merge `main` → `develop` sempre com **merge commit**. Detalhes em `docs/CONTRIBUTING.md`.
- `git push`, PR, merge e deploy exigem autorização explícita; aprovar uma spec autoriza apenas implementação e commits locais.

## Segurança

- Repositório público: nunca versionar segredos, `.env` ou dados pessoais. Ver `SECURITY.md`.
- Código proprietário, todos os direitos reservados (`LICENSE`).
- Revise com `security-best-practices` mudanças em autenticação, sessões, dinheiro, uploads e entrada de usuário.

## Convenções

- Documentação em português (Brasil). Nomes de módulos como acima; colunas de dinheiro com sufixo `_cents`.
- Toda funcionalidade considera usuários, regras de negócio, permissões, auditoria, histórico e integração entre módulos.
