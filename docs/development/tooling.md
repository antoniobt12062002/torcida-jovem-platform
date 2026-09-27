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
| `oapi-codegen/runtime` | v1.7.0 | `api/go.mod` | pkg.go.dev (proxy do Go) | 2026-09-26 | Entrou no `go.mod` com o contrato de identidade (T70), o primeiro com parâmetros e corpo. Antes: (o `/healthz` não os tem, então o código gerado não a importa e o `go mod tidy` a remove) |
| `kin-openapi` | v0.149.0 | `api/go.mod` | [README](https://github.com/getkin/kin-openapi) | 2026-09-26 | Fixado na T26 (contrato embutido no código gerado) e usado na validação de respostas nos testes (T27). Ligar `IncludeResponseStatus` para reprovar status não documentado |
| `openapi-typescript` | 7.13.0 | `web/package.json` | [openapi-ts.dev](https://openapi-ts.dev/introduction) e [CLI](https://openapi-ts.dev/cli) | 2026-09-26 | Node >= 22.12. Um arquivo por contrato, configurado em `web/redocly.yaml` (`apis` e `x-openapi-ts.output`); `pnpm gen:api` gera e `pnpm gen:api:check` reprova diferença. Fixado na T29 |
| `@redocly/cli` | 2.54.3 | `web/package.json` | npm e [documentação do Redocly](https://redocly.com/docs/cli/) | 2026-09-26 | `pnpm lint:api` lê o mesmo `web/redocly.yaml`; regras `info-license` e `operation-4xx-response` desligadas. Fixado na T29 |
| `openapi-fetch` | 0.17.0 | (primeira tela real) | npm | 2026-09-26 | Não instalado agora |
| `oapi-codegen/gin-middleware` | v1.1.0 | recusado | proxy do Go | 2026-09-26 | Não atende a API-02.4 (só entrega mensagem em texto); no lugar, middleware próprio em `platform/httpx` sobre o `kin-openapi` |
| `golang.org/x/crypto` (argon2) | v0.57.0 | `api/go.mod` | [pkg.go.dev](https://pkg.go.dev/golang.org/x/crypto/argon2) e [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html) | 2026-09-26 | `argon2.IDKey` (argon2id). Padrão 19 MiB, 2 iterações e paralelismo 1, o mínimo recomendado pelo OWASP para argon2id; custo medido na máquina de desenvolvimento: cerca de 17 ms por hash. Os parâmetros são configuráveis (`ARGON2_*`) e vão dentro do hash (formato PHC), então mudar o padrão não invalida hashes antigos |
| Lista de senhas comuns (`denylist.txt`) | SecLists commit `c205c36` (2025-05-08) | `api/internal/platform/password/` | [SecLists](https://github.com/danielmiessler/SecLists), arquivo `Passwords/Common-Credentials/xato-net-10-million-passwords-100000.txt` | 2026-09-26 | Licença MIT (reproduzida em `denylist.SOURCE.txt`). Filtrada para minúsculas, 8 ou mais caracteres, sem espaços nem repetidas: 38451 entradas, cerca de 355 KB. Embutida no binário; nenhuma consulta externa. Consulta externa (k-anonymity) segue como evolução em aberto |

### Geração do `common.yaml` (T70)

O `common.yaml` não tem rotas, então o `oapi-codegen` poda os componentes por não estarem em uso; a configuração `api/openapi/codegen/common.yaml` usa `output-options.skip-prune: true` e gera `gin-server`, `strict-server`, `models` e `embedded-spec` (sem eles os tipos de parâmetros compartilhados, como `CsrfToken`, não saem). O pacote gerado é `api/internal/platform/apicommon`, e cada contrato de módulo o importa por `import-mapping: {common.yaml: <pacote>}`.

### Resultado do teste rápido (T24)

Experimento descartável com um `common.yaml` compartilhado e dois módulos com `$ref` externo, gerados com `oapi-codegen` v2.8.0 (`gin-server`, `strict-server`, `embedded-spec`) e exercitados com o Gin. Nada foi commitado além da fixação da ferramenta.

- **`application/problem+json`: passa.** O código gerado responde com `Content-Type: application/problem+json`, e o `kin-openapi` valida essas respostas sem decodificador extra.
- **Cookie de sessão: passa.** O esquema `apiKey` em cookie vai para o contrato embutido e para o requisito de segurança da operação. O código gerado **não impõe** a segurança: a imposição é do middleware `authn`.
- **`X-CSRF-Token`: passa.** Vira parâmetro de cabeçalho obrigatório; o gancho `GinServerOptions.ErrorHandler` permite responder em `problem+json`. A imposição real continua no middleware de CSRF, que roda antes.
- **`$ref` externo entre módulos: passa**, com um pacote compartilhado gerado do `common.yaml` (precisa de `gin-server`, `strict-server`, `models` e `embedded-spec`) e `import-mapping: {common.yaml: <pacote>}` nos módulos. `Cents` vira `money.Cents`, e `errors.Is(err, money.ErrOutOfRange)` funciona no `RequestErrorHandlerFunc`, que é o gancho para o mapeamento futuro em `problem+json`.

Achados adicionais:

- **`gin-middleware` v1.1.0 e erros por campo (achado ao ler o código-fonte):** o `ErrorHandler` recebe só a mensagem (`func(c, message string, status int)`, primeira linha do erro do `kin-openapi`) e o middleware usa `gorilla/mux` e valida `Host` quando o contrato tem `servers`. Não dá para montar o `errors[]` (`field` e `code`) de API-02.4 sem interpretar texto. Decisão: middleware próprio, pequeno, em `platform/httpx` (`NewContractValidator`), sobre o `kin-openapi`, com erros estruturados (422 `validation_failed` com `errors[]`, 400 `invalid_json`). Ele valida só o contrato HTTP; não aplica esquemas de segurança nem regras de negócio, e deixa passar rotas sem operação no contrato. Cada módulo o liga por `GinServerOptions.Middlewares` com o próprio contrato. Formatos `uuid` e `email` são registrados no pacote, pois o `kin-openapi` só valida os que conhece.
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

## Verificação de vulnerabilidades no CI (SEC-001)

O job `api` roda `go tool govulncheck ./...` (dependência de ferramenta do `go.mod`, mesmo mecanismo do `oapi-codegen`) e o job `web` roda `pnpm audit --audit-level high`. Ambos entram no CI como etapa bloqueante: uma falha nesses passos derruba o job e, por consequência, o `ci-gate`.

**Política de severidade:**

- **Go (`govulncheck`):** falha em qualquer achado alcançável pelo código (o modo padrão já filtra só o que é de fato chamado, não toda dependência com CVE aberto). Uma vulnerabilidade num pacote importado mas nunca invocado aparece no relatório sem falhar o job.
- **Web (`pnpm audit`):** o limiar é `high` — só vulnerabilidade **crítica** ou **alta** falha o job. Achados **moderados** ou **baixos** ficam registrados na saída do CI (não bloqueiam); a decisão de agir sobre eles depende de uma política de triagem ainda não definida.

**Escopo:** cobre as dependências dos dois módulos (`api` e `web`). CodeQL fica fora desta rodada (decisão registrada em `.specs/STATE.md`): exige configuração de queries, tratamento de falsos positivos e o custo do SARIF/security dashboard; reavaliar quando os módulos de domínio (financeiro, associados, loja) justificarem o investimento.

**Toolchain do Go:** o `govulncheck` também aponta vulnerabilidades da própria standard library, então o `go` do `go.mod` (`api/go.mod`) precisa acompanhar os patches de segurança da série 1.26 — não só a versão maior. Ao subir a versão do Go, rode `go tool govulncheck ./...` localmente antes de commitar.

| Ferramenta | Versão | Onde | Fonte consultada | Data | Observações |
|---|---|---|---|---|---|
| `golang.org/x/vuln/cmd/govulncheck` | v1.8.0 | `api/go.mod` (`tool`) | [pkg.go.dev](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck) | 2026-09-27 | Rodado com `go tool`, sem instalação à parte |
| `pnpm audit` | a do `pnpm` já fixado (`web/package.json`) | `.github/workflows/ci.yml` | `pnpm audit --help` | 2026-09-27 | `--audit-level high`; sem ferramenta nova |

## Pendentes de escolha (sempre com consulta à documentação vigente)

- Gerador de servidor OpenAPI para Gin, gerador de tipos TypeScript e linter de OpenAPI (tarefa T24 da `fundacao-core`).
- SDK de S3 e emulador local de S3 (`fundacao-documentos`, tarefas T3 e T4).

## Limitações conhecidas

- O `testcontainers-go` volta ao socket padrão do Docker quando `DOCKER_HOST` é inválido; por isso o cenário "Docker parado" não é simulado nos testes. A falha ao iniciar o contêiner é coberta com uma imagem inexistente.
- O script `docker/postgres/init/01-roles.sql` só roda em volume novo; para recriar o banco local use `docker compose down -v`.
