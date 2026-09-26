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

## Contrato OpenAPI (ADR-008, AD-010)

| Ferramenta | Versão | Onde | Fonte consultada | Data | Observações |
|---|---|---|---|---|---|
| `oapi-codegen` | v2.8.0 (17/07/2026) | `api/go.mod` (diretiva `tool`) | [pkg.go.dev](https://pkg.go.dev/github.com/oapi-codegen/oapi-codegen/v2) e [README](https://github.com/oapi-codegen/oapi-codegen) | 2026-09-26 | Executado com `go tool oapi-codegen`; exige Go >= 1.25. Fixada na T24 |
| `oapi-codegen/runtime` | v1.7.0 | `api/go.mod` | pkg.go.dev (proxy do Go) | 2026-09-26 | Entra no `go.mod` junto do primeiro código gerado (T26), porque o `go mod tidy` só mantém dependências importadas |
| `kin-openapi` | v0.149.0 | `api/go.mod` | [README](https://github.com/getkin/kin-openapi) | 2026-09-26 | Validação de respostas nos testes (T27). Ligar `IncludeResponseStatus` para reprovar status não documentado |
| `openapi-typescript` | 7.13.0 | `web/package.json` (T29) | [openapi-ts.dev](https://openapi-ts.dev/introduction) | 2026-09-26 | Node >= 22.12 |
| `@redocly/cli` | 2.54.3 | `web/package.json` (T29) | npm e [documentação do Redocly](https://redocly.com/docs/cli/) | 2026-09-26 | Lint dos contratos |
| `openapi-fetch` | 0.17.0 | (primeira tela real) | npm | 2026-09-26 | Não instalado agora |
| `oapi-codegen/gin-middleware` | v1.1.0 | (candidato, não adotado) | proxy do Go | 2026-09-26 | Validador de requisições a partir do contrato; ver resultado do teste rápido |

### Resultado do teste rápido (T24)

Experimento descartável com um `common.yaml` compartilhado e dois módulos com `$ref` externo, gerados com `oapi-codegen` v2.8.0 (`gin-server`, `strict-server`, `embedded-spec`) e exercitados com o Gin. Nada foi commitado além da fixação da ferramenta.

- **`application/problem+json`: passa.** O código gerado responde com `Content-Type: application/problem+json`, e o `kin-openapi` valida essas respostas sem decodificador extra.
- **Cookie de sessão: passa.** O esquema `apiKey` em cookie vai para o contrato embutido e para o requisito de segurança da operação. O código gerado **não impõe** a segurança: a imposição é do middleware `authn`.
- **`X-CSRF-Token`: passa.** Vira parâmetro de cabeçalho obrigatório; o gancho `GinServerOptions.ErrorHandler` permite responder em `problem+json`. A imposição real continua no middleware de CSRF, que roda antes.
- **`$ref` externo entre módulos: passa**, com um pacote compartilhado gerado do `common.yaml` (precisa de `gin-server`, `strict-server`, `models` e `embedded-spec`) e `import-mapping: {common.yaml: <pacote>}` nos módulos. `Cents` vira `money.Cents`, e `errors.Is(err, money.ErrOutOfRange)` funciona no `RequestErrorHandlerFunc`, que é o gancho para o mapeamento futuro em `problem+json`.

Achados adicionais:

- O código gerado **não aplica** restrições do esquema (por exemplo `minLength`); só decodifica. O `gin-middleware` v1.1.0 aplica-as a partir do contrato (testado: `minLength` e campo obrigatório resultam em erro tratável). A adoção fica para a decisão sobre validação de requisições, antes dos handlers de identidade.
- `GetSwagger` está obsoleto; usar `GetSpec`.

## Front (Next.js)

| Ferramenta | Versão | Onde | Fonte consultada | Data | Observações |
|---|---|---|---|---|---|
| `vitest` | 5.0.2 | `web/package.json` | [guia de Vitest do Next.js 16.3.6](https://nextjs.org/docs/app/guides/testing/vitest) e npm | 2026-09-26 | Requer Node ^22.12. `pnpm test` roda `vitest run` (sem watch) |
| `jsdom` | 29.1.1 | `web/package.json` | npm (campo `engines`) | 2026-09-26 | Não usa a 30.x porque ela exige Node >= 22.22.2; a 29 aceita ^22.13 |
| `@testing-library/react` e `@testing-library/dom` | 16.3.3 e 10.4.2 | `web/package.json` | Guia do Next.js e npm | 2026-09-26 | React 19 suportado pelos peers |
| `@vitejs/plugin-react` | 6.1.1 | `web/package.json` | Guia do Next.js e npm | 2026-09-26 | Peer `vite` ^8 (instalado automaticamente) |
| `vite-tsconfig-paths` | 6.1.1 | `web/package.json` | Guia do Next.js e npm | 2026-09-26 | Faz o alias `@/` funcionar nos testes |
| `@types/node` | ^22.20.4 | `web/package.json` | Peer do vitest | 2026-09-26 | Alinhado ao Node 22 do `.nvmrc` (antes ^20) |

## Pendentes de escolha (sempre com consulta à documentação vigente)

- Gerador de servidor OpenAPI para Gin, gerador de tipos TypeScript e linter de OpenAPI (tarefa T24 da `fundacao-core`).
- Lista de senhas comprometidas: fonte e licença (tarefa T42).
- SDK de S3 e emulador local de S3 (`fundacao-documentos`, tarefas T3 e T4).

## Limitações conhecidas

- O `testcontainers-go` volta ao socket padrão do Docker quando `DOCKER_HOST` é inválido; por isso o cenário "Docker parado" não é simulado nos testes. A falha ao iniciar o contêiner é coberta com uma imagem inexistente.
- O script `docker/postgres/init/01-roles.sql` só roda em volume novo; para recriar o banco local use `docker compose down -v`.
