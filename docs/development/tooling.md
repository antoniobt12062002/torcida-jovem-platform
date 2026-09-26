# Ferramentas e versões adotadas

Registro das ferramentas fixadas durante a execução das specs, com a fonte consultada. Nenhuma versão é fixada sem checar a documentação vigente; cada linha nova entra junto do commit que a adota.

## Backend (Go)

| Ferramenta | Versão | Onde | Fonte consultada | Data | Observações |
|---|---|---|---|---|---|
| `testcontainers-go` e `modules/postgres` | v0.44.0 | `api/go.mod` | [pkg.go.dev](https://pkg.go.dev/github.com/testcontainers/testcontainers-go/modules/postgres) | 2026-09-26 | Usa `postgres.Run`; `RunContainer` está obsoleto. Só em arquivos com a tag `integration` |
| `golang-migrate/migrate/v4` | v4.20.1 | `api/go.mod` | [pkg.go.dev](https://pkg.go.dev/github.com/golang-migrate/migrate/v4) | 2026-09-26 | Como biblioteca (`iofs` + driver `pgx/v5`). API conferida no código do módulo |
| Imagem `migrate/migrate` | v4.20.1 | `docker-compose.yml` | Docker Hub (a tag existe) | 2026-09-26 | Mesma versão da biblioteca |
| Imagem `postgres` | 16-alpine | `docker-compose.yml`, `testutil` | ADR-002 | 2026-09-26 | PostgreSQL 16; troque nos testes com `TJ_TEST_POSTGRES_IMAGE` |
| `golangci-lint` | `latest` (CI) | `.github/workflows/ci.yml` | golangci-lint-action | 2026-09-26 | Roda com `--build-tags=integration` para analisar os helpers de integração |

## Pendentes de escolha (sempre com consulta à documentação vigente)

- Gerador de servidor OpenAPI para Gin, gerador de tipos TypeScript e linter de OpenAPI (tarefa T24 da `fundacao-core`).
- Lista de senhas comprometidas: fonte e licença (tarefa T41).
- SDK de S3 e emulador local de S3 (`fundacao-documentos`, tarefas T3 e T4).

## Limitações conhecidas

- O `testcontainers-go` volta ao socket padrão do Docker quando `DOCKER_HOST` é inválido; por isso o cenário "Docker parado" não é simulado nos testes. A falha ao iniciar o contêiner é coberta com uma imagem inexistente.
- O script `docker/postgres/init/01-roles.sql` só roda em volume novo; para recriar o banco local use `docker compose down -v`.
