# Design: Scaffold de /api (Go) e /web (Next.js) com CI real

Data: 2026-09-26
Status: Aprovado em conversa, aguardando revisão do documento escrito

## 1. Objetivo

Criar o esqueleto do monorepo rodando de ponta a ponta (API, front, banco local) e trocar os jobs placeholder `api` e `web` do CI por validações reais. Não inclui autenticação, entidades do financeiro nem deploy real.

## 2. Decisões

| Tema | Decisão |
|---|---|
| API | Go + Gin + GORM + Postgres 16 |
| Migrações | SQL versionado com `golang-migrate` (GORM só para consultas; sem `AutoMigrate`) |
| Front | Next.js (App Router) + TypeScript + Tailwind + shadcn/ui, gerenciado com pnpm |
| Dev local | `docker-compose.yml` com Postgres |
| Segredos | Somente `.env.example`; o repositório é público |

Motivo das migrações versionadas: no financeiro o esquema precisa ser rastreável e auditável, e `AutoMigrate` em produção é arriscado.

## 3. Estrutura

```
api/
  cmd/api/main.go        ponto de entrada com graceful shutdown
  internal/config/       leitura de DATABASE_URL e PORT
  internal/database/     conexão GORM/Postgres
  internal/http/         router Gin e handlers (GET /healthz)
  migrations/            SQL versionado (golang-migrate)
  Dockerfile             build multi-stage
  .golangci.yml
  go.mod
web/
  app/page.tsx           consulta /healthz e exibe o status da API
  .nvmrc, pnpm-lock.yaml, demais arquivos do Next.js/shadcn
docker-compose.yml
.env.example
```

## 4. Comportamento

- `GET /healthz` retorna `200 {"status":"ok"}` quando o banco responde ao ping e `503 {"status":"degraded"}` caso contrário.
- A home do front chama a API (URL via `NEXT_PUBLIC_API_URL`) e mostra "API online" ou "API indisponível".
- Teste da API: handler de `/healthz` com `httptest`, cobrindo banco ok e banco indisponível (ping injetável por interface).

## 5. CI

O `ci.yml` mantém `changes` (filtro por caminho) e `ci-gate`. Os jobs passam a ser:

- `api`: `go vet ./...`, `golangci-lint`, `go test ./...`, `go build ./...` (versão do Go lida de `api/go.mod`; cache de módulos).
- `web`: `pnpm install --frozen-lockfile`, lint, `tsc --noEmit`, `pnpm build` (versão do Node lida de `web/.nvmrc`; cache do store do pnpm).

## 6. Fora de escopo

Autenticação e permissões, entidades e regras do financeiro, deploy real (depende da hospedagem), testes de front além de lint, tipagem e build.
