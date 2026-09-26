# Fundação Core Tasks

## Execution Protocol (MANDATORY -- do not skip)

Implement these tasks with the `tlc-spec-driven` skill: **activate it by name and follow its Execute flow and Critical Rules.** Do not search for skill files by filesystem path. The skill is the source of truth for the full flow (per-task cycle, sub-agent delegation, adequacy review, Verifier, discrimination sensor).

**If the skill cannot be activated, STOP and tell the user - do not proceed without it.**

---

**Design**: `.specs/features/fundacao-core/design.md`
**Status**: In Progress (execução em lotes pequenos, por fase)

Regras deste plano:

- Um commit atômico por tarefa, no padrão Conventional Commits. `git push` e PR só com autorização explícita.
- Cada tarefa inclui os testes da sua camada; "testado em outra tarefa" não é justificativa.
- Os ACs citados em `Requirement` são os que a tarefa cobre, e cada um precisa de ao menos um teste.
- Testes de integração exigem Docker e a tag `integration`.
- Ferramentas e versões só são fixadas depois de checar a documentação vigente; a fonte consultada vai na mensagem do commit.

---

## Test Coverage Matrix

> Generated from codebase, project guidelines, and spec - confirm before Execute. Guidelines found: `CLAUDE.md`, `docs/CONTRIBUTING.md`, `docs/architecture/domain-boundaries.md`, `.github/workflows/ci.yml`, `api/.golangci.yml`. Não há meta de cobertura documentada; aplicam-se os padrões fortes (todo AC e todo caso de borda listado tem teste).

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| ---------- | ------------------ | -------------------- | ---------------- | ----------- |
| Lógica pura (`platform/money`, `authz`, `audit/entry`, `identity/domain`, `password`) | unit | Todas as ramificações; 1:1 com os ACs; todo caso de borda listado; `money` com 95% de cobertura | `api/internal/**/*_test.go` | `cd api && go test ./...` |
| Casos de uso (`identity/app`, `platform/documents`) | integration | Todos os ACs, incluindo permissão negada, transação e auditoria, contra PostgreSQL real | `api/internal/**/*_test.go` com tag `integration` | `cd api && go test -tags=integration ./...` |
| Repositório e SQL (`identity/infra`, `platform/database`, `audit`, migrações) | integration | Caminhos principais de consulta, restrições, triggers, concessões de papel e erros | `api/internal/**/*_test.go` com tag `integration` | `cd api && go test -tags=integration ./...` |
| Middlewares HTTP (`platform/httpx`) | unit | Todos os ramos de sucesso e falha de cada middleware | `api/internal/platform/httpx/*_test.go` | `cd api && go test ./...` |
| Handlers e router | integration | Todas as rotas no escopo: caminho feliz, bordas e erros, com resposta validada contra o contrato | `api/internal/**/http/*_test.go`, `api/internal/httpapi/*_test.go` | `cd api && go test -tags=integration ./...` |
| Storage S3 e documentos | integration | Emulador S3 real: guardar, versionar, URL assinada, expiração, acesso sem assinatura | `api/internal/platform/{storage,documents}/*_test.go` | `cd api && go test -tags=integration ./...` |
| CLI (`cmd/*`) | unit | Leitura de argumentos e ambiente; a lógica é testada na camada de aplicação | `api/cmd/**/*_test.go` | `cd api && go test ./...` |
| Front, utilitários (`web/lib`) | unit | 1:1 com os vetores de dinheiro; borda de inteiro seguro | `web/lib/**/*.test.ts` | `cd web && pnpm test` |
| Contrato OpenAPI, migrações de dados, config de CI, Docker, documentação | none | - (gate de build, geração sem diferença e lint) | - | gate de build |
| Interface/tipos sem lógica (`storage.go`, `main.go`) e configuração | none | - (gate de build; a configuração com lógica é coberta por unit) | - | gate de build |

## Gate Check Commands

> Generated from codebase - confirm before Execute.

| Gate Level | When to Use | Command |
| ---------- | ----------- | ------- |
| Quick | Após tarefas apenas com testes de unidade | `cd api && go vet ./... && go test ./...` |
| Full | Após tarefas com testes de integração (exige Docker) | `cd api && go vet ./... && go test -tags=integration ./...` |
| Build | Após fim de fase, ou tarefas de config, CI, contrato e front | `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build` |

Tarefas somente de front usam a metade `web` do gate Build; tarefas somente de API usam a metade `api`. `golangci-lint` roda no CI (não está instalado localmente).

---

## Execution Plan

Phases are ordered and run sequentially - each phase completes before the next begins, and tasks within a phase execute in order.

### Phase 1: Plataforma HTTP e configuração

```
T1 → T2 → T3 → T4 → T5
```

### Phase 2: Infraestrutura de testes de integração (Go)

```
T6 → T7 → T8 → T9 → T10 → T11 → T12 → T13
```

### Phase 3: Infraestrutura de testes do front

```
T14 → T15 → T16
```

### Phase 4: Dinheiro

```
T17 → T18 → T19 → T20 → T21 → T22
```

### Phase 5: Contrato OpenAPI

```
T23 → T24 → T25 → T26 → T27 → T28 → T29 → T30
```

### Phase 6: Auditoria

```
T31 → T32 → T33 → T34 → T35
```

### Phase 7: Autorização e papéis

```
T36 → T37 → T38 → T39 → T40 → T41 → T42
```

### Phase 8: Identidade: base

```
T43 → T44 → T45 → T46 → T47 → T48 → T49 → T50 → T51 → T52 → T53 → T54 → T55
```

### Phase 9: Identidade: casos de uso

```
T56 → T57 → T58 → T59 → T60 → T61 → T62 → T63 → T64 → T65 → T66 → T67 → T68
```

### Phase 10: HTTP: contrato de identidade e middlewares

```
T69 → T70 → T71 → T72 → T73
```

### Phase 11: HTTP: handlers e fiação

```
T74 → T75 → T76 → T77 → T78 → T79 → T80 → T81 → T82
```

### Phase 12: Guardrails e documentação

```
T83 → T84 → T85 → T86
```

---

## Task Breakdown

### Phase 1: Plataforma HTTP e configuração (tarefas)

### T1: Estender o carregamento de configuração com APP_ENV, LOG_LEVEL, ALLOWED_ORIGINS, COOKIE_SECURE, COOKIE_DOMAIN, SESSION_IDLE_MINUTES (padrão 60) e SESSION_ABSOLUTE_HOURS (padrão 8), validando na partida

**What**: Estender o carregamento de configuração com APP_ENV, LOG_LEVEL, ALLOWED_ORIGINS, COOKIE_SECURE, COOKIE_DOMAIN, SESSION_IDLE_MINUTES (padrão 60) e SESSION_ABSOLUTE_HOURS (padrão 8), validando na partida.  
**Where**: `api/internal/config/config.go`  
**Depends on**: None  
**Reuses**: `config.Load` atual e seus testes  
**Requirement**: PLT-01 (ACs 1, 2)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Variável ausente ou inválida devolve erro que nomeia a variável e não contém valores secretos
- [x] `COOKIE_SECURE=false` com `APP_ENV` diferente de `development` é recusado
- [x] Padrões: 60 minutos de ociosidade e 8 horas de teto absoluto
- [x] Testes existentes de `config` continuam passando
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(config): valida configuracao de sessao, origem e ambiente`

---

### T2: Criar o logger `slog` em JSON com redação das chaves password, password_hash, token, session_token, csrf_token, cookie e authorization

**What**: Criar o logger `slog` em JSON com redação das chaves password, password_hash, token, session_token, csrf_token, cookie e authorization.  
**Where**: `api/internal/platform/logx/logx.go`  
**Depends on**: T1  
**Reuses**: Biblioteca padrão `log/slog`  
**Requirement**: PLT-01 (ACs 4)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Chave sensível vira `[redacted]` também em grupos aninhados
- [x] Nível configurável por `LOG_LEVEL`
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(platform): adiciona logger com redacao de campos sensiveis`

---

### T3: Criar o middleware de identificador de requisição

**What**: Criar o middleware de identificador de requisição.  
**Where**: `api/internal/platform/httpx/requestid.go`  
**Depends on**: T2  
**Reuses**: `gin` já usado no router  
**Requirement**: API-02 (ACs 2)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Cabeçalho válido é reaproveitado; inválido ou ausente gera UUID v4
- [x] O id é devolvido no cabeçalho `X-Request-Id` e fica disponível no contexto
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(platform): adiciona middleware de request id`

---

### T4: Criar o escritor de erros `application/problem+json` e o mapeamento de erros de validação

**What**: Criar o escritor de erros `application/problem+json` e o mapeamento de erros de validação.  
**Where**: `api/internal/platform/httpx/problem.go`  
**Depends on**: T3  
**Reuses**: Middleware `requestid` da tarefa anterior  
**Requirement**: API-02 (ACs 1, 3, 4)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Corpo com type, title, status, detail, code e request_id
- [x] JSON malformado produz 400 `invalid_json`; validação produz 422 `validation_failed` com `errors[]`
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(platform): adiciona respostas problem+json`

---

### T5: Criar os middlewares de recuperação de panic, log de acesso (sem query string) e limite de corpo de 1 MiB

**What**: Criar os middlewares de recuperação de panic, log de acesso (sem query string) e limite de corpo de 1 MiB.  
**Where**: `api/internal/platform/httpx/middleware.go`  
**Depends on**: T4  
**Reuses**: `logx` e `problem`  
**Requirement**: API-02 (ACs 6, 7); PLT-01 (ACs 3)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Panic vira 500 `internal_error` sem detalhes internos e com request_id
- [x] Corpo acima de 1 MiB vira 413 `payload_too_large`
- [x] Log de acesso é uma linha JSON com método, caminho sem query, status, duração, request_id e user id quando houver
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(platform): adiciona recover, log de acesso e limite de corpo`

---

### Phase 2: Infraestrutura de testes de integração (Go) (tarefas)

### T6: Mover `internal/database` para `internal/platform/database` e configurar o logger do GORM com consultas parametrizadas

**What**: Mover `internal/database` para `internal/platform/database` e configurar o logger do GORM com consultas parametrizadas.  
**Where**: `api/internal/platform/database/database.go`  
**Depends on**: T2  
**Reuses**: `database.Open` e `database.Ping` atuais; `logx`  
**Requirement**: PLT-01 (ACs 7)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] `Open` recebe o logger e nunca registra valores de parâmetros
- [x] Importações em `cmd/api/main.go` atualizadas e o build continua passando
- [x] `git mv` preserva o histórico do arquivo
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `refactor(platform): move pacote database e parametriza log do GORM`

---

### T7: Criar o script de papéis do ambiente local: `tj_owner` (dono, DDL) e `tj_app` (aplicação, sem DDL)

**What**: Criar o script de papéis do ambiente local: `tj_owner` (dono, DDL) e `tj_app` (aplicação, sem DDL).  
**Where**: `docker/postgres/init/01-roles.sql`  
**Depends on**: T6  
**Reuses**: `docker-compose.yml` atual  
**Requirement**: PLT-01 (ACs 6); TST-01 (ACs 4)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Script idempotente cria os dois papéis com senhas de desenvolvimento
- [x] Papel `tj_app` não tem DDL nem superusuário
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `chore(docker): cria papeis dono e aplicacao no postgres local`

---

### T8: Montar o script de papéis no serviço `db` e usar a conexão do dono no serviço `migrate`

**What**: Montar o script de papéis no serviço `db` e usar a conexão do dono no serviço `migrate`.  
**Where**: `docker-compose.yml`  
**Depends on**: T7  
**Reuses**: Serviços `db` e `migrate` atuais  
**Requirement**: PLT-01 (ACs 5); TST-01 (ACs 4)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] `docker compose config` valida
- [x] `docker compose up -d db` seguido de `migrate` conclui com o papel dono
- [x] Migração 000001 aplicada continua funcionando
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `chore(docker): separa papeis dono e aplicacao no compose`

---

### T9: Atualizar o `.env.example` com `DATABASE_URL` do papel `tj_app`, `MIGRATE_DATABASE_URL` do dono e as novas variáveis

**What**: Atualizar o `.env.example` com `DATABASE_URL` do papel `tj_app`, `MIGRATE_DATABASE_URL` do dono e as novas variáveis.  
**Where**: `.env.example`  
**Depends on**: T8  
**Reuses**: `.env.example` atual  
**Requirement**: PLT-01 (ACs 1, 6)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Nenhum segredo real no arquivo
- [x] Todas as variáveis lidas por `config.Load` estão presentes ou documentadas
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `chore: atualiza env de exemplo com papeis de banco e sessao`

---

### T10: Criar o helper que sobe PostgreSQL 16 via testcontainers uma vez por pacote e cria os papéis

**What**: Criar o helper que sobe PostgreSQL 16 via testcontainers uma vez por pacote e cria os papéis.  
**Where**: `api/internal/platform/testutil/postgres.go`  
**Depends on**: T9  
**Reuses**: Script `01-roles.sql`  
**Requirement**: TST-01 (ACs 1, 2, 6)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Arquivo com a tag `integration`; sem a tag, `go test ./...` não exige Docker
- [x] Sem Docker e com a tag, o teste falha citando Docker
- [x] Teste de fumaça conecta e confirma que os papéis existem
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `test(platform): adiciona helper de postgres com testcontainers`

---

### T11: Criar a função `Migrate` (golang-migrate como biblioteca) para testes e para a ferramenta de migração

**What**: Criar a função `Migrate` (golang-migrate como biblioteca) para testes e para a ferramenta de migração.  
**Where**: `api/internal/platform/database/migrate.go`  
**Depends on**: T10  
**Reuses**: Diretório `api/migrations`  
**Requirement**: PLT-01 (ACs 5); TST-01 (ACs 2)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Aplica a migração 000001 e o `pgcrypto` fica disponível
- [x] Reexecutar é idempotente
- [x] API não chama `Migrate` na partida
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(platform): adiciona execucao de migracoes como biblioteca`

---

### T12: Criar `NewTestDB(t)`: clona um banco-modelo já migrado por teste e devolve o pool com o papel `tj_app`

**What**: Criar `NewTestDB(t)`: clona um banco-modelo já migrado por teste e devolve o pool com o papel `tj_app`.  
**Where**: `api/internal/platform/testutil/db.go`  
**Depends on**: T11  
**Reuses**: `testutil/postgres.go` e `database.Migrate`  
**Requirement**: TST-01 (ACs 2, 3, 4)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Teste prova que dados de um teste não aparecem em outro
- [x] O pool usa `tj_app`, não o dono
- [x] Banco do teste é removido ao final
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `test(platform): adiciona banco isolado por teste`

---

### T13: Executar `go test -tags=integration ./...` no job `api` do CI

**What**: Executar `go test -tags=integration ./...` no job `api` do CI.  
**Where**: `.github/workflows/ci.yml`  
**Depends on**: T12  
**Reuses**: Job `api` atual  
**Requirement**: TST-01 (ACs 5)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Job `api` roda os testes de integração
- [x] O CI fica verde com os testes existentes
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `ci: roda testes de integracao no job api`

---

### Phase 3: Infraestrutura de testes do front (tarefas)

### T14: Adicionar Vitest, jsdom e Testing Library às dependências de desenvolvimento e o script `test`

**What**: Adicionar Vitest, jsdom e Testing Library às dependências de desenvolvimento e o script `test`.  
**Where**: `web/package.json`  
**Depends on**: None  
**Reuses**: `web/package.json` atual e `pnpm-lock.yaml`  
**Requirement**: TST-03 (ACs 1)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] `pnpm install --frozen-lockfile` passa com o lockfile atualizado
- [x] Script `pnpm test` existe
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `chore(web): adiciona dependencias de teste`

---

### T15: Configurar o Vitest com jsdom e incluir um teste de fumaça

**What**: Configurar o Vitest com jsdom e incluir um teste de fumaça.  
**Where**: `web/vitest.config.ts`  
**Depends on**: T14  
**Reuses**: Configuração do Next em `web/`  
**Requirement**: TST-03 (ACs 1)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] `pnpm test` executa `*.test.ts` e `*.test.tsx` em jsdom
- [x] O teste de fumaça (`web/lib/smoke.test.ts`) passa
- [x] Lint e `tsc` continuam passando
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: unit  
**Gate**: build

**Commit**: `test(web): configura vitest com jsdom`

---

### T16: Executar `pnpm test` no job `web` do CI, antes do build

**What**: Executar `pnpm test` no job `web` do CI, antes do build.  
**Where**: `.github/workflows/ci.yml`  
**Depends on**: T15  
**Reuses**: Job `web` atual  
**Requirement**: TST-03 (ACs 2)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] `pnpm test` roda antes de `pnpm build` no job `web`
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `ci: roda testes do front antes do build`

---

### Phase 4: Dinheiro (tarefas)

### T17: Criar o tipo `Cents` com Add, Sub, Neg, Cmp e `ErrOverflow`

**What**: Criar o tipo `Cents` com Add, Sub, Neg, Cmp e `ErrOverflow`.  
**Where**: `api/internal/platform/money/money.go`  
**Depends on**: None  
**Reuses**: Nenhum (pacote novo)  
**Requirement**: MNY-01 (ACs 1, 2, 3, 4, 5)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Overflow de soma e subtração devolve `ErrOverflow` sem dar a volta
- [x] Nenhuma função aceita ou devolve float32 ou float64 (verificado por teste que inspeciona a API do pacote)
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 5 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(money): adiciona tipo Cents com aritmetica exata`

---

### T18: Criar `Format` e `Parse` em pt-BR

**What**: Criar `Format` e `Parse` em pt-BR.  
**Where**: `api/internal/platform/money/format.go`  
**Depends on**: T17  
**Reuses**: `money.go`  
**Requirement**: MNY-02 (ACs 4, 5, 6)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] 123456 vira `R$ 1.234,56` e -123456 vira `-R$ 1.234,56` com espaço comum
- [x] `Parse` rejeita mais de duas casas decimais e caracteres inválidos
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(money): adiciona formatacao e parse pt-BR`

---

### T19: Criar a serialização JSON estrita e a checagem de faixa segura

**What**: Criar a serialização JSON estrita e a checagem de faixa segura.  
**Where**: `api/internal/platform/money/json.go`  
**Depends on**: T18  
**Reuses**: `money.go`  
**Requirement**: MNY-02 (ACs 1, 2, 3)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Marshal emite literal inteiro; unmarshal rejeita string, decimal, expoente e null
- [x] Valor acima de 9007199254740991 em módulo devolve erro mapeável para 422 `amount_out_of_range`
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(money): adiciona serializacao JSON estrita`

---

### T20: Criar `Allocate` e `Percent`

**What**: Criar `Allocate` e `Percent`.  
**Where**: `api/internal/platform/money/allocate.go`  
**Depends on**: T19  
**Reuses**: `money.go`  
**Requirement**: MNY-03 (ACs 1, 2, 3, 4, 5, 6); MNY-01 (ACs 3)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] `Allocate(10000,[1,1,1])` devolve 3334, 3333, 3333
- [x] `Percent(10000,400)` devolve 400 e `Percent(1,5000)` devolve 1
- [x] Teste de propriedade: a soma das partes é sempre o total
- [x] Cobertura de instruções do pacote `money` maior ou igual a 95%
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 7 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(money): adiciona rateio e percentual`

---

### T21: Criar o teste que percorre o esquema e barra colunas `real`, `double precision` e `numeric`, e `_cents` que não seja `bigint`

**What**: Criar o teste que percorre o esquema e barra colunas `real`, `double precision` e `numeric`, e `_cents` que não seja `bigint`.  
**Where**: `api/internal/platform/database/schema_guard_test.go`  
**Depends on**: T20, T12  
**Reuses**: `testutil.NewTestDB`  
**Requirement**: MNY-02 (ACs 7, 8)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Passa no esquema atual
- [x] Um esquema de fixture com coluna proibida faz o teste falhar (teste negativo)
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `test(database): adiciona guarda de esquema para dinheiro`

---

### T22: Criar o formatador e o parser de dinheiro em TypeScript com os mesmos vetores do Go

**What**: Criar o formatador e o parser de dinheiro em TypeScript com os mesmos vetores do Go.  
**Where**: `web/lib/money.ts`  
**Depends on**: T21, T15  
**Reuses**: Vitest da fase anterior  
**Requirement**: MNY-02 (ACs 4, 5, 6, 9); TST-03 (ACs 3)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Mesmos vetores do Go passam
- [x] Número que não é inteiro seguro lança `RangeError`
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 5 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: unit  
**Gate**: build

**Commit**: `feat(web): adiciona utilitario de dinheiro`

---

### Phase 5: Contrato OpenAPI (tarefas)

### T23: Criar o contrato compartilhado `common.yaml`: esquemas `Problem`, `Cents`, `Limit` e `Cursor` e os esquemas de segurança de cookie de sessão e `X-CSRF-Token`

**What**: Criar o contrato compartilhado `common.yaml`: esquemas `Problem`, `Cents`, `Limit` e `Cursor` e os esquemas de segurança de cookie de sessão e `X-CSRF-Token`.  
**Where**: `api/openapi/common.yaml`  
**Depends on**: None  
**Reuses**: Convenções de `problem+json`, paginação e dinheiro da spec  
**Requirement**: API-01 (ACs 1); API-02 (ACs 1, 5)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] OpenAPI 3.0.3 válido, contendo só componentes de comunicação e nenhuma regra de negócio
- [x] `Cents` é `integer` `int64` e mapeia para `money.Cents` por `x-go-type`
- [x] `Problem` tem type, title, status, detail, code, request_id e `errors[]` opcional
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `feat(api): adiciona contrato compartilhado common`

---

### T24: Fazer o teste rápido do `oapi-codegen` com `application/problem+json`, cookie de sessão, `X-CSRF-Token` e `$ref` externo para `common.yaml`, e só então fixar `oapi-codegen` (como tool) e o runtime no `go.mod`

**What**: Fazer o teste rápido do `oapi-codegen` com `application/problem+json`, cookie de sessão, `X-CSRF-Token` e `$ref` externo para `common.yaml`, e só então fixar `oapi-codegen` (como tool) e o runtime no `go.mod`.  
**Where**: `api/go.mod`  
**Depends on**: T23  
**Reuses**: Ferramentas e versões da seção Contrato OpenAPI do design  
**Requirement**: API-01 (ACs 2, 4)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Os quatro pontos do teste rápido foram executados e o resultado foi reportado ao mantenedor
- [x] Se houve incompatibilidade, a execução parou e a solução voltou ao mantenedor antes de fixar qualquer coisa
- [x] Versões fixadas (`oapi-codegen` v2.8.0, runtime v1.7.0, `kin-openapi` v0.149.0) com a fonte registrada em `docs/development/tooling.md`
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `build(api): fixa oapi-codegen como tool e o runtime`

---

### T25: Criar o contrato do módulo `platform` com `GET /healthz` (200 `ok` e 503 `degraded`), reutilizando `common.yaml`

**What**: Criar o contrato do módulo `platform` com `GET /healthz` (200 `ok` e 503 `degraded`), reutilizando `common.yaml`.  
**Where**: `api/openapi/platform.yaml`  
**Depends on**: T24  
**Reuses**: Handler `/healthz` existente e `common.yaml`  
**Requirement**: API-01 (ACs 1); API-02 (ACs 1)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Documento OpenAPI 3.0.3 válido que referencia `common.yaml`
- [x] `/healthz` descrito com as respostas 200 e 503 exatamente como o handler atual
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `feat(api): adiciona contrato do modulo platform`

---

### T26: Criar o gatilho `go generate` do `platform`, a configuração do `oapi-codegen` e versionar o código gerado; o handler de `/healthz` passa a implementar a interface gerada

**What**: Criar o gatilho `go generate` do `platform`, a configuração do `oapi-codegen` e versionar o código gerado; o handler de `/healthz` passa a implementar a interface gerada.  
**Where**: `api/internal/platform/api/generate.go`  
**Depends on**: T25  
**Reuses**: `platform.yaml`, `go tool oapi-codegen` e o handler atual  
**Requirement**: API-01 (ACs 2)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Configuração em `api/openapi/codegen/platform.yaml` e código gerado commitado (pacote `platformapi`)
- [x] Gerar duas vezes produz zero diferença
- [x] O router atual atende `/healthz` pela interface gerada e o teste existente continua passando
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(api): gera interface do servidor do modulo platform`

---

### T27: Criar o validador de respostas contra o contrato de um módulo, para uso nos testes de integração

**What**: Criar o validador de respostas contra o contrato de um módulo, para uso nos testes de integração.  
**Where**: `api/internal/platform/testutil/contract.go`  
**Depends on**: T26  
**Reuses**: Contratos em `api/openapi/` e `kin-openapi`  
**Requirement**: API-01 (ACs 5)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Resposta conforme passa; status, cabeçalho ou corpo divergente falha o teste
- [x] Suporta `application/problem+json`
- [x] `/healthz` é validado por ele
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `test(platform): valida respostas contra o contrato`

---

### T28: Criar o teste que compara as rotas registradas no Gin com as operações de todos os contratos de módulo

**What**: Criar o teste que compara as rotas registradas no Gin com as operações de todos os contratos de módulo.  
**Where**: `api/internal/httpapi/routes_test.go`  
**Depends on**: T27  
**Reuses**: `gin.Engine.Routes()` e os contratos  
**Requirement**: API-01 (ACs 6)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Rota sem operação em nenhum contrato falha o teste
- [x] Operação de contrato sem rota também falha (teste negativo com fixture)
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `test(httpapi): garante paridade entre rotas e contratos`

---

### T29: Adicionar `openapi-typescript` e `@redocly/cli` ao front, os scripts `gen:api` (um arquivo por módulo em `web/lib/api/`) e `lint:api`, e versionar os tipos do `platform`

**What**: Adicionar `openapi-typescript` e `@redocly/cli` ao front, os scripts `gen:api` (um arquivo por módulo em `web/lib/api/`) e `lint:api`, e versionar os tipos do `platform`.  
**Where**: `web/package.json`  
**Depends on**: T28, T14  
**Reuses**: Contratos em `api/openapi/`  
**Requirement**: API-01 (ACs 3, 4)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] `gen:api` gera `web/lib/api/platform.d.ts` sem diferença na segunda execução
- [x] `lint:api` passa nos contratos atuais
- [x] `tsc` passa com o arquivo gerado; `openapi-fetch` não é instalado agora
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `feat(web): gera tipos e lint dos contratos openapi`

---

### T30: Criar o job `contract` no CI (Go e Node): lint, regeneração de Go e TypeScript sem diferença, filtrado por `api/openapi/**`, e incluí-lo no `ci-gate`

**What**: Criar o job `contract` no CI (Go e Node): lint, regeneração de Go e TypeScript sem diferença, filtrado por `api/openapi/**`, e incluí-lo no `ci-gate`.  
**Where**: `.github/workflows/ci.yml`  
**Depends on**: T29, T13, T16  
**Reuses**: Jobs `api` e `web` e o filtro `changes`  
**Requirement**: API-01 (ACs 2, 3, 4)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Alterar um contrato sem regenerar faz o CI falhar
- [x] Contrato inválido faz o CI falhar
- [x] `ci-gate` considera o job `contract`
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `ci: valida contratos openapi e codigo gerado`

---

### Phase 6: Auditoria (tarefas)

### T31: Criar a migração `audit_log` com ator (tipo e usuário), alvo, resultado, contexto, índices, triggers de imutabilidade (UPDATE, DELETE e TRUNCATE) e concessões ao papel `tj_app`; incluir o arquivo `down`

**What**: Criar a migração `audit_log` com ator (tipo e usuário), alvo, resultado, contexto, índices, triggers de imutabilidade (UPDATE, DELETE e TRUNCATE) e concessões ao papel `tj_app`; incluir o arquivo `down`.  
**Where**: `api/migrations/000002_audit_log.up.sql`  
**Depends on**: T12, T11  
**Reuses**: Migração 000001 e o padrão de `tj_app`  
**Requirement**: AUD-02 (ACs 1, 2)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Colunas `actor_type`, `outcome` e `context` presentes; `actor_user_id` só com `actor_type` igual a `user`, e valores fora dos permitidos são recusados pelo banco
- [x] Com o papel dono, UPDATE, DELETE e TRUNCATE falham por trigger
- [x] Com `tj_app`, só INSERT e SELECT funcionam
- [x] Migração `down` desfaz sem erro
- [x] Falha com mensagem clara se o papel `tj_app` não existir
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(audit): cria tabela de auditoria imutavel`

---

### T32: Criar `WithTx` e `TxFrom`, a unidade de trabalho que carrega a transação no contexto

**What**: Criar `WithTx` e `TxFrom`, a unidade de trabalho que carrega a transação no contexto.  
**Where**: `api/internal/platform/database/tx.go`  
**Depends on**: T31, T6  
**Reuses**: `database.Open`  
**Requirement**: AUD-01 (ACs 1, 6)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Erro no callback reverte a transação
- [x] `TxFrom` sem transação devolve `ok=false`
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(platform): adiciona unidade de trabalho transacional`

---

### T33: Criar o tipo `Entry`, o catálogo de ações, sua validação e a redação de chaves sensíveis

**What**: Criar o tipo `Entry`, o catálogo de ações, sua validação e a redação de chaves sensíveis.  
**Where**: `api/internal/platform/audit/entry.go`  
**Depends on**: T32  
**Reuses**: Lista de chaves sensíveis do `logx`  
**Requirement**: AUD-01 (ACs 3, 4, 5); AUD-04 (ACs 7)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Ação de cancelamento ou ajuste sem motivo é rejeitada
- [x] Ação fora do catálogo é rejeitada; o catálogo tem as ações de identidade e de segurança da spec
- [x] `actor_user_id` só com `actor_type` igual a `user`; `outcome` e `actor_type` inválidos são rejeitados
- [x] Chaves sensíveis em before, after e context viram `[redacted]`
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 4 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(audit): adiciona entrada de auditoria com validacao e redacao`

---

### T34: Criar o `Recorder`: `Record` grava na transação do contexto e `RecordSecurity` grava em transação própria

**What**: Criar o `Recorder`: `Record` grava na transação do contexto e `RecordSecurity` grava em transação própria.  
**Where**: `api/internal/platform/audit/recorder.go`  
**Depends on**: T33  
**Reuses**: `database.TxFrom` e `Entry`  
**Requirement**: AUD-01 (ACs 1, 3, 6); AUD-04 (ACs 6)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] `Record` sem transação no contexto devolve erro e não grava
- [x] O registro contém request_id e ator vindos do contexto
- [x] `RecordSecurity` grava em transação própria, sem depender de transação de negócio, e se a gravação falhar registra incidente operacional (log de erro com ação e request_id), sem devolver erro nem alterar a resposta; `Record` com falha reverte a operação e é mapeável para 500 `audit_failed`
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 4 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(audit): adiciona gravacao de auditoria na transacao`

---

### T35: Criar o teste de atomicidade: falha na auditoria reverte a alteração e sucesso grava as duas na mesma transação

**What**: Criar o teste de atomicidade: falha na auditoria reverte a alteração e sucesso grava as duas na mesma transação.  
**Where**: `api/internal/platform/audit/atomicity_test.go`  
**Depends on**: T34  
**Reuses**: `Recorder` e `WithTx`  
**Requirement**: AUD-01 (ACs 1, 2)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Falha simulada na auditoria deixa o dado alterado intacto
- [x] Sucesso grava dado e auditoria juntos
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `test(audit): prova atomicidade entre dado e auditoria`

---

### Phase 7: Autorização e papéis (tarefas)

### T36: Criar a migração de identidade: users, roles, permissions, role_permissions, user_roles, admin_memberships, sessions e login_attempts, com `down` e concessões

**What**: Criar a migração de identidade: users, roles, permissions, role_permissions, user_roles, admin_memberships, sessions e login_attempts, com `down` e concessões.  
**Where**: `api/migrations/000003_identity.up.sql`  
**Depends on**: T12  
**Reuses**: Migração 000002 como padrão  
**Requirement**: RBAC-01 (ACs 2); IDN-02 (ACs 2); IDN-05 (ACs 5, 6); IDN-06 (ACs 5)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] `permissions.name` só aceita o formato `módulo:recurso:ação`; `roles` e `permissions` não são semeadas pela migração (a sincronização as cria)
- [x] Coluna `users.must_change_password` (boolean, padrão falso); `users` não tem coluna de associado (o vínculo pertence a `associados`)
- [x] Índice único por `lower(email)`
- [x] `sessions.token_hash` único
- [x] `admin_memberships` com motivo obrigatório, quem concedeu, quando, e os campos de encerramento; índice único parcial garante um vínculo ativo por usuário
- [x] `tj_app` com SELECT, INSERT, UPDATE e DELETE nas tabelas de identidade, exceto `admin_memberships`, sem DELETE (o histórico é preservado)
- [x] Migração `down` desfaz sem erro
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 5 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): cria tabelas de identidade e sessoes`

---

### T37: Criar o tipo `Permission` com validação de formato e a `Definition` (permissão, descrição e marca de leitura comum)

**What**: Criar o tipo `Permission` com validação de formato e a `Definition` (permissão, descrição e marca de leitura comum).  
**Where**: `api/internal/platform/authz/permission.go`  
**Depends on**: T36  
**Reuses**: Nenhum (pacote novo)  
**Requirement**: RBAC-01 (ACs 1); RBAC-02 (ACs 7)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Aceita `financeiro:lancamento:read` e rejeita formas inválidas
- [x] Regex `^[a-z_]+:[a-z_]+:[a-z_]+$` aplicada
- [x] `Definition` com permissão inválida é rejeitada e a marca de leitura comum só é aceita em permissão de ação `read`
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(authz): adiciona tipo de permissao e sua definicao`

---

### T38: Criar `Principal`, `Require` e o `Authorizer` com o gancho de negação, e o teste que impede decisão por nome de papel

**What**: Criar `Principal`, `Require` e o `Authorizer` com o gancho de negação, e o teste que impede decisão por nome de papel.  
**Where**: `api/internal/platform/authz/authz.go`  
**Depends on**: T37  
**Reuses**: `Permission` e `Definition`  
**Requirement**: RBAC-02 (ACs 2, 3, 7, 8)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Sem a permissão devolve `ErrForbidden`, que vira 403 `forbidden`, e o erro não revela se o recurso existe
- [x] `Authorizer` chama o gancho de negação, com o principal e a permissão, para toda negação de permissão que não seja leitura comum, e não o chama quando permitido nem para leitura comum
- [x] Permissão desconhecida do catálogo é negada e passa pelo gancho
- [x] Teste de arquitetura falha se houver nome de papel literal fora da definição dos papéis ou leitura de `Principal.Roles` fora de `identity`, com fixtures negativas
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 4 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(authz): adiciona principal, verificacao e gancho de negacao`

---

### T39: Criar o gancho que grava `authz.denied` na auditoria como evento de segurança, e registrar a ação no catálogo

**What**: Criar o gancho que grava `authz.denied` na auditoria como evento de segurança, e registrar a ação no catálogo.  
**Where**: `api/internal/platform/audit/authz_hook.go`  
**Depends on**: T38  
**Reuses**: `audit.Recorder.RecordSecurity` e `authz.Authorizer`  
**Requirement**: RBAC-02 (ACs 7); AUD-04 (ACs 6)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Negação grava `authz.denied` com outcome `denied`, o usuário como ator, a permissão como alvo e o request_id
- [x] Falha ao gravar não altera a resposta nem devolve erro
- [x] O catálogo de auditoria contém `authz.denied`
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(audit): registra negacoes de autorizacao`

---

### T40: Criar a verificação de concessão sem escalada: `Covers` devolve as permissões que o ator não possui entre as necessárias

**What**: Criar a verificação de concessão sem escalada: `Covers` devolve as permissões que o ator não possui entre as necessárias.  
**Where**: `api/internal/platform/authz/escalation.go`  
**Depends on**: T39  
**Reuses**: `Principal` e `Permission`  
**Requirement**: RBAC-03 (ACs 1, 2, 4)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Ator que possui todas as permissões pedidas passa; faltando qualquer uma, a função devolve exatamente as que faltam
- [x] Conjunto vazio de permissões pedidas sempre passa
- [x] Usa apenas as permissões do `Principal` recebido, sem cache próprio
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(authz): adiciona verificacao de concessao sem escalada`

---

### T41: Criar os papéis, a contribuição modular de permissões e a agregação da matriz (`BuildMatrix`), com as contribuições da fundação e os testes de invariante

**What**: Criar os papéis, a contribuição modular de permissões e a agregação da matriz (`BuildMatrix`), com as contribuições da fundação e os testes de invariante.  
**Where**: `api/internal/identity/app/roles_matrix.go`  
**Depends on**: T40  
**Reuses**: `authz.Definition`  
**Requirement**: RBAC-01 (ACs 4, 5, 6, 9, 10)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Oito papéis definidos em `identity/domain/role.go`; a matriz provisória da spec vem das contribuições de identidade, de auditoria e das permissões institucionais reservadas ao financeiro
- [x] `BuildMatrix` recusa permissão inválida, concessão de permissão não declarada, permissão duplicada com definição diferente e ação `create`, `update`, `delete` ou `cancel` para CONSELHO_FISCAL (testes negativos)
- [x] CONSELHO_FISCAL tem `financeiro:prestacao_contas:read`, `financeiro:prestacao_contas:approve`, `financeiro:parecer:opine` e `audit:log:read`
- [x] PRESIDENTE recebe explicitamente todas as permissões do catálogo, inclusive as de uma contribuição nova de teste, sem curinga; as de ADMIN_SISTEMA são subconjunto das do PRESIDENTE
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 5 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(identity): define papeis e agrega a matriz de permissoes`

---

### T42: Criar o repositório de papéis com a sincronização idempotente (sob lock consultivo) e o cálculo das permissões efetivas

**What**: Criar o repositório de papéis com a sincronização idempotente (sob lock consultivo) e o cálculo das permissões efetivas.  
**Where**: `api/internal/identity/infra/role_repository.go`  
**Depends on**: T41, T34  
**Reuses**: `BuildMatrix`, `WithTx` e `audit.Recorder`  
**Requirement**: RBAC-01 (ACs 2, 3, 7, 8)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Sincronizar duas vezes deixa o banco idêntico e a segunda não grava auditoria
- [x] Duas sincronizações simultâneas (instâncias subindo juntas) não duplicam linhas nem o `rbac.sync`
- [x] Quando algo muda, grava `rbac.sync` na mesma transação, sem ator, com `before`, `after` (estado completo) e o diff (permissões adicionadas, desativadas e reativadas; vínculos adicionados e removidos), suficiente para reconstruir o histórico
- [x] Se a auditoria falhar, a sincronização é revertida
- [x] Permissão removida do código fica inativa e não é apagada, e não conta nas permissões efetivas
- [x] Permissões efetivas são a união dos papéis do usuário, que pode ter vários papéis ativos
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 4 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): sincroniza papeis e calcula permissoes efetivas`

---

### Phase 8: Identidade: base (tarefas)

### T43: Estender a configuração com AUTH_HASH_KEY, os parâmetros do argon2id, PASSWORD_RESET_TTL_MINUTES, EMAIL_PROVIDER, EMAIL_FROM e APP_BASE_URL, validando na partida

**What**: Estender a configuração com AUTH_HASH_KEY, os parâmetros do argon2id, PASSWORD_RESET_TTL_MINUTES, EMAIL_PROVIDER, EMAIL_FROM e APP_BASE_URL, validando na partida.  
**Where**: `api/internal/config/config.go`  
**Depends on**: None  
**Reuses**: `config.Load` e seus testes  
**Requirement**: EML-01 (ACs 3, 5); IDN-07 (ACs 2); PLT-01 (ACs 1)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] `AUTH_HASH_KEY` com no mínimo 32 bytes é obrigatória fora de `development` e nunca aparece em mensagem de erro
- [x] Parâmetros do argon2id (memória em KiB, iterações e paralelismo) com padrão 19456, 2 e 1; valores inválidos são recusados nomeando a variável
- [x] `PASSWORD_RESET_TTL_MINUTES` com padrão 30
- [x] `EMAIL_PROVIDER` aceita `log` e `disabled` (padrão `log` em `development` e `disabled` nas demais); valor desconhecido é recusado nomeando a variável
- [x] `APP_BASE_URL` (base dos links) é uma URL absoluta e exige https fora de `development`; `EMAIL_FROM` é um endereço válido quando o provedor não é `disabled`
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 4 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(config): valida configuracao de autenticacao e e-mail`

---

### T44: Acrescentar ao catálogo de auditoria as ações de recuperação e de redefinição de senha

**What**: Acrescentar ao catálogo de auditoria as ações de recuperação e de redefinição de senha.  
**Where**: `api/internal/platform/audit/entry.go`  
**Depends on**: T43  
**Reuses**: `audit.Register` e o catálogo existente  
**Requirement**: AUD-04 (ACs 5, 7)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] O catálogo contém `auth.password_reset_requested`, `auth.password_reset_completed`, `auth.password_reset_failed` e `user.password_reset`
- [x] Uma entrada `user.password_reset` com resultado `denied` é válida e as ações anteriores continuam válidas
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(audit): adiciona acoes de recuperacao de senha`

---

### T45: Criar `platform/email`: interface `Sender`, `Message` validada, remetentes `log` e `disabled` e o registrador de testes

**What**: Criar `platform/email`: interface `Sender`, `Message` validada, remetentes `log` e `disabled` e o registrador de testes.  
**Where**: `api/internal/platform/email/email.go`  
**Depends on**: T44  
**Reuses**: `logx` e a configuração  
**Requirement**: EML-01 (ACs 1, 2, 3, 4, 5)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] `Message.Validate` recusa destinatário inválido, assunto vazio e CR ou LF no destinatário ou no assunto
- [x] O remetente `log` registra só o domínio do destinatário e o assunto, nunca o corpo
- [x] O remetente `disabled` não envia nada e a fábrica registra um aviso operacional
- [x] A fábrica recusa provedor desconhecido nomeando a variável
- [x] `emailtest.Recorder` guarda as mensagens enviadas, para os testes de outros módulos
- [x] Nenhum código de módulo de negócio importa um provedor específico
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 5 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(platform): adiciona capacidade de e-mail desacoplada do provedor`

---

### T46: Criar hash e verificação de senha com argon2id, com parâmetros configuráveis

**What**: Criar hash e verificação de senha com argon2id, com parâmetros configuráveis.  
**Where**: `api/internal/platform/password/argon2.go`  
**Depends on**: T45  
**Reuses**: `golang.org/x/crypto/argon2`  
**Requirement**: IDN-04 (ACs 9); IDN-01 (ACs 3)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [x] Parâmetros iniciais conferidos com a recomendação vigente do OWASP e custo medido
- [x] Hash inclui parâmetros e sal; verificação em tempo constante
- [x] Hash nunca aparece em log
- [x] Memória, iterações e paralelismo vêm da configuração
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(password): adiciona hash argon2id`

---

### T47: Criar a lista embutida de senhas comuns ou comprometidas e a verificação sem diferenciar maiúsculas

**What**: Criar a lista embutida de senhas comuns ou comprometidas e a verificação sem diferenciar maiúsculas.  
**Where**: `api/internal/platform/password/denylist.go`  
**Depends on**: T46  
**Reuses**: Diretiva `go:embed` sobre `denylist.txt`  
**Requirement**: IDN-05 (ACs 4)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [x] Fonte e licença da lista registradas no commit; se não houver fonte adequada, gera-se lista própria de senhas comuns
- [x] `Contains` ignora maiúsculas e minúsculas
- [x] Arquivo `denylist.txt` embutido no binário, sem consulta externa
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(password): adiciona lista de senhas comprometidas`

---

### T48: Criar a entidade de usuário, a normalização de e-mail e a política de senha por papel (10 para qualquer papel diferente de ASSOCIADO, 8 só para ASSOCIADO, máximo 128 pontos de código)

**What**: Criar a entidade de usuário, a normalização de e-mail e a política de senha por papel (10 para qualquer papel diferente de ASSOCIADO, 8 só para ASSOCIADO, máximo 128 pontos de código).  
**Where**: `api/internal/identity/domain/user.go`  
**Depends on**: T47  
**Reuses**: Nenhum (pacote novo)  
**Requirement**: IDN-01 (ACs 3); IDN-04 (ACs 2); IDN-05 (ACs 1, 2, 3, 7)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] E-mail é normalizado: espaços nas pontas removidos, minúsculas e formato básico (contém `@`, até 254 caracteres)
- [x] `HashEmail` (HMAC-SHA256 com a chave) é determinístico, ignora maiúsculas e dá resultados diferentes para chaves diferentes
- [x] Usuário com papel diferente de ASSOCIADO exige 10 caracteres; só ASSOCIADO exige 8; acima de 128 é rejeitado
- [x] Comprimento medido em pontos de código Unicode e sem regras de composição
- [x] Conjunto vazio de papéis é inválido
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 6 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(identity): adiciona entidade de usuario e politica de senha`

---

### T49: Criar a entidade `AdminMembership` (por que o usuário tem acesso administrativo) e suas invariantes, separada de papel (o que ele pode fazer)

**What**: Criar a entidade `AdminMembership` (por que o usuário tem acesso administrativo) e suas invariantes, separada de papel (o que ele pode fazer).  
**Where**: `api/internal/identity/domain/admin_membership.go`  
**Depends on**: T48  
**Reuses**: `user.go`  
**Requirement**: IDN-06 (ACs 2, 3, 9)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Motivo obrigatório, com no mínimo 10 caracteres após remover espaços nas pontas; sem motivo devolve erro mapeável para 422 `reason_required`
- [x] Só pode haver um vínculo ativo por usuário
- [x] Papel diferente de ASSOCIADO exige vínculo ativo, senão devolve erro mapeável para 409 `admin_membership_required`
- [x] Encerrar registra quem, quando e por quê, e não apaga o vínculo
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(identity): adiciona vinculo administrativo e suas invariantes`

---

### T50: Criar o repositório de usuários

**What**: Criar o repositório de usuários.  
**Where**: `api/internal/identity/infra/user_repository.go`  
**Depends on**: T49  
**Reuses**: `user.go`  
**Requirement**: IDN-04 (ACs 1, 2, 3)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Criação simultânea do mesmo e-mail deixa exatamente uma vencer (teste concorrente)
- [x] Consulta por e-mail ignora maiúsculas
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona repositorio de usuarios`

---

### T51: Criar o repositório de vínculos administrativos

**What**: Criar o repositório de vínculos administrativos.  
**Where**: `api/internal/identity/infra/admin_membership_repository.go`  
**Depends on**: T50  
**Reuses**: `admin_membership.go` e `user_repository.go`  
**Requirement**: IDN-06 (ACs 3, 5)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Conceder dois vínculos ativos ao mesmo usuário ao mesmo tempo deixa exatamente um vencer (teste concorrente)
- [x] Encerrar preserva a linha e o histórico de vínculos anteriores continua consultável
- [x] O repositório não expõe operação de exclusão
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona repositorio de vinculos administrativos`

---

### T52: Criar o repositório de sessões (hash do token, revogação, expiração)

**What**: Criar o repositório de sessões (hash do token, revogação, expiração).  
**Where**: `api/internal/identity/infra/session_repository.go`  
**Depends on**: T51  
**Reuses**: `user_repository.go`  
**Requirement**: IDN-02 (ACs 2, 5, 6)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [x] Só o hash SHA-256 é gravado
- [x] Revogar todas as sessões de um usuário funciona em uma transação
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona repositorio de sessoes`

---

### T53: Criar o repositório de tentativas de login

**What**: Criar o repositório de tentativas de login.  
**Where**: `api/internal/identity/infra/attempt_repository.go`  
**Depends on**: T52  
**Reuses**: `session_repository.go`  
**Requirement**: IDN-02 (ACs 8)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Conta falhas por hash do e-mail na janela de 15 minutos
- [x] Sucesso zera a contagem
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona repositorio de tentativas de login`

---

### T54: Criar o caso de uso de login: verificação de senha, bloqueio por tentativas, erro uniforme e rotação de sessão

**What**: Criar o caso de uso de login: verificação de senha, bloqueio por tentativas, erro uniforme e rotação de sessão.  
**Where**: `api/internal/identity/app/authenticate.go`  
**Depends on**: T53  
**Reuses**: Repositórios da fase, `password`, `audit`  
**Requirement**: IDN-02 (ACs 1, 2, 3, 4, 7, 8, 10); AUD-04 (ACs 1, 2, 3, 6)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [x] Login bem-sucedido grava `auth.login` por `RecordSecurity`; falha grava `auth.login_failed` (com categoria da falha e hash do e-mail, sem senha nem e-mail em claro) e bloqueio grava `auth.login_blocked`, também por `RecordSecurity`; falha ao gravar o evento não impede a resposta do login
- [x] Nenhuma sequência de falhas bloqueia a conta de forma permanente
- [x] Usuário inexistente, senha errada e inativo devolvem o mesmo erro
- [x] Quinta falha em 15 minutos bloqueia por 15 minutos
- [x] Login bem-sucedido emite novo token e revoga a sessão apresentada no cookie, se válida; várias sessões simultâneas são permitidas
- [x] Usuário inexistente também verifica a senha contra um hash fixo, para o tempo de resposta não revelar o e-mail
- [x] Tentativas feitas durante o bloqueio não contam como falha nem o estendem
- [x] O e-mail, nas tentativas e nos eventos, aparece só como HMAC-SHA256
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 11 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona caso de uso de login`

---

### T55: Criar o serviço de sessão: validação (ociosidade de 60 minutos e teto de 8 horas), logout e atualização limitada de `last_seen_at`

**What**: Criar o serviço de sessão: validação (ociosidade de 60 minutos e teto de 8 horas), logout e atualização limitada de `last_seen_at`.  
**Where**: `api/internal/identity/app/session_service.go`  
**Depends on**: T54  
**Reuses**: Repositórios de sessão e de papéis  
**Requirement**: IDN-02 (ACs 5, 6, 9); RBAC-02 (ACs 5); AUD-04 (ACs 4)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [x] Logout grava `auth.logout`
- [x] Sessão ociosa ou antiga é inválida
- [x] Usuário inativo invalida todas as suas sessões
- [x] `last_seen_at` atualiza no máximo uma vez por minuto
- [x] Devolve Principal e token CSRF
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 5 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona servico de sessao`

---

### Phase 9: Identidade: casos de uso (tarefas)

### T56: Acrescentar a permissão `identity:user:reset_password` à contribuição de identidade, concedida ao ADMIN_SISTEMA

**What**: Acrescentar a permissão `identity:user:reset_password` à contribuição de identidade, concedida ao ADMIN_SISTEMA.  
**Where**: `api/internal/identity/app/roles_matrix.go`  
**Depends on**: None  
**Reuses**: `FoundationContributions` e os testes da matriz  
**Requirement**: IDN-08 (ACs 1)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] A permissão é declarada e concedida a ADMIN_SISTEMA, e o PRESIDENTE a recebe por ter todas
- [x] Não é leitura comum, e o Conselho Fiscal continua sem create, update, delete e cancel
- [x] Os testes de invariante da matriz continuam passando
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(identity): adiciona permissao de redefinicao de senha`

---

### T57: Criar a migração de recuperação de senha (`password_reset_tokens`, só o hash do token, `password_reset_requests` e `password_change_attempts`) com `down` e concessões

**What**: Criar a migração de recuperação de senha (`password_reset_tokens`, só o hash do token, `password_reset_requests` e `password_change_attempts`) com `down` e concessões.  
**Where**: `api/migrations/000004_password_reset.up.sql`  
**Depends on**: T56  
**Reuses**: Migração 000003 como padrão  
**Requirement**: IDN-07 (ACs 2, 3, 4); IDN-04 (ACs 12)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] `token_hash` único e nenhuma coluna guarda o token; `password_reset_requests` guarda só o HMAC do e-mail; `password_change_attempts` guarda só o usuário e o instante
- [x] `tj_app` com SELECT, INSERT e UPDATE em `password_reset_tokens`, sem DELETE
- [x] Migração `down` desfaz sem erro
- [x] Falha com mensagem clara se o papel `tj_app` não existir
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 4 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): cria tabelas de recuperacao de senha`

---

### T58: Criar o repositório de recuperação: criação de token com invalidação dos pendentes, consumo de uso único e contagem de solicitações por e-mail e das tentativas de troca de senha por usuário

**What**: Criar o repositório de recuperação: criação de token com invalidação dos pendentes, consumo de uso único e contagem de solicitações por e-mail e das tentativas de troca de senha por usuário.  
**Where**: `api/internal/identity/infra/reset_token_repository.go`  
**Depends on**: T57  
**Reuses**: `password_reset_tokens` e `password_reset_requests`  
**Requirement**: IDN-07 (ACs 2, 4, 5, 6); IDN-04 (ACs 12)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [x] As tentativas de troca de senha têm contador próprio por usuário, separado das tentativas de login
- [x] Só o SHA-256 do token é gravado
- [x] Criar um token invalida os pendentes do mesmo usuário
- [x] O consumo é de uso único, inclusive com duas confirmações simultâneas (só uma vence)
- [x] Token expirado, usado ou desconhecido devolve o mesmo erro
- [x] Conta as solicitações de um e-mail (por HMAC) na última hora
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 5 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona repositorio de tokens de recuperacao`

---

### T59: Criar o caso de uso de criação de usuário; toda conta nasce apenas com o papel ASSOCIADO, e o acesso administrativo só existe por promoção

**What**: Criar o caso de uso de criação de usuário; toda conta nasce apenas com o papel ASSOCIADO, e o acesso administrativo só existe por promoção.  
**Where**: `api/internal/identity/app/create_user.go`  
**Depends on**: T58  
**Reuses**: `WithTx`, `audit.Recorder`, `authz.Require`  
**Requirement**: IDN-04 (ACs 1, 2); RBAC-02 (ACs 3); IDN-05 (ACs 2, 3, 4, 8)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [x] Aplica a política de senha de ASSOCIADO (8) e a lista de comprometidas
- [x] O comando não aceita papéis: o usuário criado tem só ASSOCIADO, com `must_change_password` verdadeiro
- [x] Exige `identity:user:create` antes de qualquer leitura ou escrita
- [x] Grava `user.create` na mesma transação
- [x] E-mail repetido devolve erro mapeável para 409 `email_taken`
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 7 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona criacao de usuario`

---

### T60: Criar os casos de uso de desativação e de reativação de usuário

**What**: Criar os casos de uso de desativação e de reativação de usuário.  
**Where**: `api/internal/identity/app/deactivate_user.go`  
**Depends on**: T59  
**Reuses**: `WithTx`, `audit.Recorder`, `authz.Covers` e o lock do conjunto de administradores  
**Requirement**: IDN-04 (ACs 3, 4, 10); RBAC-02 (ACs 3, 5); RBAC-03 (ACs 3, 4, 5, 6)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [x] Exigem `identity:user:update` antes de qualquer leitura ou escrita
- [x] Desativar revoga todas as sessões e os tokens de recuperação pendentes do usuário na mesma transação e grava `user.deactivate`
- [x] Desativar o último usuário ativo que possui `identity:admin:grant` devolve 409 `last_admin`, sob um lock consultivo do conjunto de administradores (duas desativações simultâneas não deixam o sistema sem administrador)
- [x] Ator sem alguma permissão do alvo recebe 403 `privilege_escalation`; o próprio usuário recebe 403 `self_change_forbidden`; ambos gravam a ação da operação com resultado `denied` por `RecordSecurity`
- [x] Reativar segue as mesmas regras, grava `user.reactivate` e não restaura papéis administrativos nem o vínculo encerrado
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 9 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona desativacao e reativacao de usuario`

---

### T61: Criar o caso de uso de troca de senha do próprio usuário

**What**: Criar o caso de uso de troca de senha do próprio usuário.  
**Where**: `api/internal/identity/app/change_password.go`  
**Depends on**: T60  
**Reuses**: `password`, `audit.Recorder` e o contador `password_change_attempts`  
**Requirement**: IDN-04 (ACs 5, 6, 9, 11, 12); IDN-05 (ACs 1, 2, 3, 4, 6)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [x] Aplica a política de senha por papel e a lista de comprometidas e limpa `must_change_password` ao trocar
- [x] Senha atual errada devolve 403 `invalid_current_password` e conta no contador próprio `password_change_attempts`, separado do login
- [x] Cinco erros em 15 minutos bloqueiam a troca por 15 minutos (429 `password_change_blocked`, com `Retry-After`); tentativas bloqueadas não contam nem estendem, e o sucesso zera a contagem
- [x] Nova senha igual à atual devolve 422 `password_unchanged`
- [x] Revoga as outras sessões do usuário (a atual continua) na mesma transação e grava `user.password_change` sem valor de senha
- [x] O erro de senha atual grava `user.password_change` com resultado `failure` por `RecordSecurity`
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 10 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona troca de senha`

---

### T62: Criar o caso de uso de promoção administrativa (`PromoteToAdmin`): concede o vínculo administrativo e os papéis numa única transação

**What**: Criar o caso de uso de promoção administrativa (`PromoteToAdmin`): concede o vínculo administrativo e os papéis numa única transação.  
**Where**: `api/internal/identity/app/promote_admin.go`  
**Depends on**: T61  
**Reuses**: `AdminMembership`, repositórios, `WithTx`, `audit.Recorder`, `authz.Covers`  
**Requirement**: IDN-06 (ACs 1, 2, 3, 4); IDN-05 (ACs 5); RBAC-03 (ACs 1, 3, 4, 5); RBAC-02 (ACs 3)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [x] Exige `identity:admin:grant` antes de qualquer leitura ou escrita
- [x] Exige ao menos um papel administrativo, senão 422 `admin_role_required`
- [x] Cria o vínculo com motivo, atribui os papéis, marca `must_change_password` quando o usuário não tinha vínculo ativo, revoga todas as sessões do alvo e grava `admin.promote` com papéis anteriores e novos, tudo na mesma transação
- [x] Ator não concede papel com permissão que ele não possui (403 `privilege_escalation`) nem altera a si mesmo (403 `self_change_forbidden`)
- [x] A tentativa negada por escalada ou por autoalteração grava `role.change_denied` em transação própria
- [x] Alvo inativo devolve 409 `user_inactive` e vínculo ativo existente devolve 409 `already_admin`
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 10 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona promocao administrativa`

---

### T63: Criar o caso de uso de retirada do acesso administrativo (`RevokeAdmin`)

**What**: Criar o caso de uso de retirada do acesso administrativo (`RevokeAdmin`).  
**Where**: `api/internal/identity/app/revoke_admin.go`  
**Depends on**: T62  
**Reuses**: `AdminMembership`, repositórios, `WithTx`, `audit.Recorder`, `authz.Covers`  
**Requirement**: IDN-06 (ACs 6, 7, 8); RBAC-03 (ACs 2, 3, 4, 5); RBAC-02 (ACs 3)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [x] Exige `identity:admin:revoke` e motivo
- [x] Encerra o vínculo, remove todos os papéis diferentes de ASSOCIADO, revoga as sessões do alvo e grava `admin.revoke`, na mesma transação
- [x] Retirar o último usuário ativo com `identity:admin:grant` devolve 409 `last_admin`
- [x] Alvo sem vínculo ativo devolve 409 `not_admin`
- [x] Ator não retira acesso de quem tem permissões que ele não possui (403 `privilege_escalation`) nem o próprio (403 `self_change_forbidden`), e a tentativa negada grava `role.change_denied`
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 8 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona retirada de acesso administrativo`

---

### T64: Criar o caso de uso de atribuição de papéis para quem já tem vínculo administrativo ativo

**What**: Criar o caso de uso de atribuição de papéis para quem já tem vínculo administrativo ativo.  
**Where**: `api/internal/identity/app/assign_roles.go`  
**Depends on**: T63  
**Reuses**: Repositório de papéis e `audit.Recorder`  
**Requirement**: IDN-04 (ACs 7); RBAC-01 (ACs 7); RBAC-02 (ACs 3); IDN-06 (ACs 9); RBAC-03 (ACs 1, 2, 3, 4, 5)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [x] Exige `identity:role:assign`
- [x] Adicionar papel diferente de ASSOCIADO a quem não tem vínculo ativo devolve 409 `admin_membership_required`
- [x] O papel ASSOCIADO é sempre mantido pelo servidor; a lista recebida traz só papéis administrativos e pode ser vazia, o que deixa o vínculo dormente
- [x] Cada papel acrescentado ou removido precisa estar coberto pelas permissões do ator (403 `privilege_escalation`); o ator não altera os próprios papéis (403 `self_change_forbidden`); a tentativa negada grava `role.change_denied`
- [x] Substitui o conjunto e grava papéis anteriores e novos em `user.roles_set`
- [x] Um usuário pode ter vários papéis ativos e as novas permissões valem na requisição seguinte
- [x] Cada AC listado em Requirement tem ao menos um teste (mínimo de 9 testes; nenhuma exclusão silenciosa)
- [x] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona atribuicao de papeis`

---

### T65: Criar o caso de uso de listagem de usuários com paginação por cursor e filtros

**What**: Criar o caso de uso de listagem de usuários com paginação por cursor e filtros.  
**Where**: `api/internal/identity/app/list_users.go`  
**Depends on**: T64  
**Reuses**: Convenção de cursor  
**Requirement**: IDN-04 (ACs 8); API-02 (ACs 5)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Exige `identity:user:read`
- [ ] Cursor por (`created_at`, `id`), 50 por página por padrão e no máximo 100 (acima disso, 422 `invalid_limit`), com `next_cursor`
- [ ] Filtros `active` e `role`; cada item traz os papéis e o resumo do vínculo administrativo ativo
- [ ] Nunca devolve hash de senha, token nem tentativas
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona listagem de usuarios`

---

### T66: Criar o caso de uso de solicitação de recuperação de acesso (`RequestPasswordReset`)

**What**: Criar o caso de uso de solicitação de recuperação de acesso (`RequestPasswordReset`).  
**Where**: `api/internal/identity/app/request_password_reset.go`  
**Depends on**: T65  
**Reuses**: Repositórios, `platform/email` e `audit.Recorder`  
**Requirement**: IDN-07 (ACs 1, 2, 3, 4, 8, 9); EML-01 (ACs 4); AUD-04 (ACs 6)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] Devolve sempre o mesmo resultado, exista ou não a conta, esteja ela ativa ou não
- [ ] Conta ativa: cria o token (256 bits, base64url), grava só o hash com validade de 30 minutos e envia o e-mail com o link depois de responder; falha de envio vira incidente no log e não chega ao cliente
- [ ] Mais de 3 solicitações do mesmo e-mail em uma hora não criam token nem enviam e-mail, com o mesmo resultado
- [ ] Grava `auth.password_reset_requested` por `RecordSecurity`, com o e-mail só como HMAC e sem token nem link
- [ ] O envio é assíncrono, então o tempo de resposta não depende de a conta existir
- [ ] Nem o token nem o link aparecem em log, auditoria ou resposta
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 8 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona solicitacao de recuperacao de acesso`

---

### T67: Criar o caso de uso de redefinição de senha com token (`ResetPasswordWithToken`)

**What**: Criar o caso de uso de redefinição de senha com token (`ResetPasswordWithToken`).  
**Where**: `api/internal/identity/app/reset_password.go`  
**Depends on**: T66  
**Reuses**: `password`, repositórios, `WithTx` e `audit.Recorder`  
**Requirement**: IDN-07 (ACs 5, 6, 7, 8); IDN-05 (ACs 1, 2, 3, 4, 6)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] Token válido e senha dentro da política: troca o hash, consome o token, invalida os outros pendentes, limpa `must_change_password`, revoga todas as sessões e grava `auth.password_reset_completed` na mesma transação
- [ ] Token desconhecido, expirado ou usado devolve o mesmo erro (400 `invalid_reset_token`) e grava `auth.password_reset_failed` por `RecordSecurity`
- [ ] Senha fora da política devolve o código da política e mantém o token válido
- [ ] Se a auditoria falhar, a redefinição é revertida
- [ ] A auditoria não contém senha nem token
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 9 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona redefinicao de senha com token`

---

### T68: Criar o caso de uso de redefinição administrativa de senha (`AdminResetPassword`)

**What**: Criar o caso de uso de redefinição administrativa de senha (`AdminResetPassword`).  
**Where**: `api/internal/identity/app/admin_reset_password.go`  
**Depends on**: T67  
**Reuses**: `password`, `authz.Covers` e `WithTx`  
**Requirement**: IDN-08 (ACs 1, 2, 3, 4, 5, 6, 7); IDN-05 (ACs 8); RBAC-03 (ACs 1, 3, 6); RBAC-02 (ACs 3)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] Exige `identity:user:reset_password` antes de qualquer leitura ou escrita, e um motivo de ao menos 10 caracteres (422 `reason_required`)
- [ ] Gera uma senha temporária aleatória que cumpre a política, troca o hash, marca `must_change_password`, revoga as sessões e os tokens de recuperação pendentes do alvo e grava `user.password_reset` com o motivo, tudo na mesma transação
- [ ] A senha temporária é devolvida uma única vez ao chamador e nunca é gravada, registrada em log nem auditada; nada é enviado por e-mail
- [ ] Ator sem alguma permissão do alvo recebe 403 `privilege_escalation`; o próprio usuário recebe 403 `self_change_forbidden`; ambos gravam `user.password_reset` com resultado `denied` por `RecordSecurity`
- [ ] Alvo inativo devolve 409 `user_inactive`
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 12 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona redefinicao administrativa de senha`

---

### Phase 10: HTTP: contrato de identidade e middlewares (tarefas)

### T69: Criar o contrato do módulo `identity`: auth (login, logout, `me`, troca de senha), users (criar sem papéis, listar, desativar, papéis), roles e vínculo administrativo (`POST` e `DELETE /api/v1/users/{id}/admin-membership`), recuperação de acesso (`POST /api/v1/auth/password-reset/request` e `/confirm`) e redefinição administrativa (`POST /api/v1/users/{id}/password-reset`)

**What**: Criar o contrato do módulo `identity`: auth (login, logout, `me`, troca de senha), users (criar sem papéis, listar, desativar, papéis), roles e vínculo administrativo (`POST` e `DELETE /api/v1/users/{id}/admin-membership`), recuperação de acesso (`POST /api/v1/auth/password-reset/request` e `/confirm`) e redefinição administrativa (`POST /api/v1/users/{id}/password-reset`).  
**Where**: `api/openapi/identity.yaml`  
**Depends on**: T26, T29, T65  
**Reuses**: `common.yaml` e o padrão de `platform.yaml`  
**Requirement**: API-01 (ACs 1); IDN-02 (ACs 1, 9); IDN-04 (ACs 1, 8); IDN-06 (ACs 1, 6); IDN-07 (ACs 1, 5); IDN-08 (ACs 1)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Cada rota tem respostas de erro `Problem` e paginação por cursor nas listas
- [ ] Cookie de sessão e `X-CSRF-Token` descritos; nenhuma regra de negócio no contrato
- [ ] Lint passa
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `feat(api): adiciona contrato do modulo identity`

---

### T70: Criar o gatilho `go generate` do `identity`, sua configuração do `oapi-codegen` e regenerar os tipos TypeScript

**What**: Criar o gatilho `go generate` do `identity`, sua configuração do `oapi-codegen` e regenerar os tipos TypeScript.  
**Where**: `api/internal/identity/http/generate.go`  
**Depends on**: T69  
**Reuses**: `identity.yaml` e o gatilho do `platform`  
**Requirement**: API-01 (ACs 2, 3)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Configuração em `api/openapi/codegen/identity.yaml` e código gerado commitado (pacote `identityhttp`)
- [ ] `pnpm gen:api` gera `web/lib/api/identity.d.ts`
- [ ] Gerar duas vezes produz zero diferença
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `feat(identity): gera interface do servidor do modulo`

---

### T71: Acrescentar a consulta de auditoria (`GET /api/v1/audit-logs`) ao contrato do `platform` e regenerar o código e os tipos

**What**: Acrescentar a consulta de auditoria (`GET /api/v1/audit-logs`) ao contrato do `platform` e regenerar o código e os tipos.  
**Where**: `api/openapi/platform.yaml`  
**Depends on**: T70  
**Reuses**: `common.yaml` (`Cursor`, `Limit`, `Problem`)  
**Requirement**: AUD-03 (ACs 1, 2, 3, 4); API-01 (ACs 1, 2, 3)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Filtros `entity_type`, `entity_id`, `actor_user_id`, `action`, `outcome`, `from` e `to` e `limit` de no máximo 100; o esquema do registro tem ator, alvo, resultado, contexto e `request_id`
- [ ] Geração de Go e TypeScript sem diferença após regenerar
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `feat(api): adiciona consulta de auditoria ao contrato platform`

---

### T72: Criar o middleware `Authn`: cookie para sessão para Principal no contexto, com 401 padronizado

**What**: Criar o middleware `Authn`: cookie para sessão para Principal no contexto, com 401 padronizado.  
**Where**: `api/internal/platform/httpx/authn.go`  
**Depends on**: T71  
**Reuses**: Interface `SessionValidator`  
**Requirement**: RBAC-02 (ACs 1, 5, 6); IDN-02 (ACs 5)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] Com `must_change_password`, todas as rotas exceto logout, `me` e troca de senha devolvem 403 `password_change_required`
- [ ] Sem cookie, cookie malformado ou sessão expirada devolvem 401 com o código correto
- [ ] Não vaza detalhes internos
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 4 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(platform): adiciona middleware de autenticacao`

---

### T73: Criar o middleware de CSRF e origem

**What**: Criar o middleware de CSRF e origem.  
**Where**: `api/internal/platform/httpx/csrf.go`  
**Depends on**: T72  
**Reuses**: `ALLOWED_ORIGINS` da config  
**Requirement**: IDN-03 (ACs 1, 2, 3, 4, 5); IDN-07 (ACs 10)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] Método que altera estado sem `X-CSRF-Token` correto devolve 403 `csrf_invalid`
- [ ] Origem fora da lista devolve 403 `origin_not_allowed`, inclusive no login, na solicitação e na confirmação de recuperação
- [ ] GET, HEAD e OPTIONS não exigem token
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 6 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(platform): adiciona protecao csrf e checagem de origem`

---

### Phase 11: HTTP: handlers e fiação (tarefas)

### T74: Criar os handlers de login, logout, `me`, troca de senha e recuperação de acesso (solicitação e confirmação)

**What**: Criar os handlers de login, logout, `me`, troca de senha e recuperação de acesso (solicitação e confirmação).  
**Where**: `api/internal/identity/http/auth_handler.go`  
**Depends on**: None  
**Reuses**: Código gerado, casos de uso e `problem`  
**Requirement**: IDN-02 (ACs 1, 3, 4, 6, 8, 9); IDN-04 (ACs 5, 6); IDN-07 (ACs 1, 5, 6, 10)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] `me` inclui `must_change_password` e o vínculo administrativo ativo, quando houver
- [ ] Cookie `tj_session` com HttpOnly, Secure, SameSite=Lax e Path=/
- [ ] Erros idênticos para credenciais inválidas
- [ ] 429 traz `Retry-After`
- [ ] A solicitação de recuperação devolve 202 com o corpo fixo, exista ou não a conta
- [ ] A confirmação devolve 400 `invalid_reset_token` idêntico para token desconhecido, expirado e usado, e 422 com o código da política
- [ ] Respostas validadas contra o contrato
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 12 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona handlers de autenticacao`

---

### T75: Criar os handlers de criação, desativação, atribuição de papéis e listagem de usuários

**What**: Criar os handlers de criação, desativação, atribuição de papéis e listagem de usuários.  
**Where**: `api/internal/identity/http/users_handler.go`  
**Depends on**: T74  
**Reuses**: Casos de uso e código gerado  
**Requirement**: IDN-04 (ACs 1, 2, 3, 4, 7, 8, 9)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Respostas nunca contêm hash de senha nem token
- [ ] Erros mapeados: 409 `email_taken`, `last_admin` e `admin_membership_required`, 403 `forbidden`, `privilege_escalation` e `self_change_forbidden`
- [ ] Respostas validadas contra o contrato
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 7 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona handlers de usuarios`

---

### T76: Criar os handlers de promoção, de retirada do vínculo administrativo e de redefinição administrativa de senha

**What**: Criar os handlers de promoção, de retirada do vínculo administrativo e de redefinição administrativa de senha.  
**Where**: `api/internal/identity/http/admin_handler.go`  
**Depends on**: T75  
**Reuses**: Casos de uso `PromoteToAdmin` e `RevokeAdmin` e código gerado  
**Requirement**: IDN-06 (ACs 1, 2, 3, 4, 6, 7, 8); IDN-08 (ACs 1, 2, 3, 4); RBAC-03 (ACs 1, 2, 3)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] Erros mapeados: 422 `reason_required`, 409 `already_admin`, `user_inactive`, `not_admin` e `last_admin`, 403 `privilege_escalation` e `self_change_forbidden`
- [ ] Respostas validadas contra o contrato
- [ ] Nenhuma resposta contém hash de senha nem token
- [ ] `POST /api/v1/users/{id}/password-reset` devolve 200 com a senha temporária uma única vez e `Cache-Control: no-store`, sem registrá-la em log nem auditoria; erros mapeados: 422 `reason_required`, 403 `privilege_escalation` e `self_change_forbidden`, 409 `user_inactive`
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 14 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona handlers de vinculo administrativo`

---

### T77: Criar o serviço de consulta e o handler `GET /api/v1/audit-logs`

**What**: Criar o serviço de consulta e o handler `GET /api/v1/audit-logs`.  
**Where**: `api/internal/platform/audit/http.go`  
**Depends on**: T76  
**Reuses**: Cursor keyset por (occurred_at, id)  
**Requirement**: AUD-03 (ACs 1, 2, 3, 4)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Ordem decrescente e filtros combinados por E
- [ ] `limit` acima de 100 devolve 422 `invalid_limit`
- [ ] Sem `audit:log:read` devolve 403
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 4 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(audit): adiciona consulta de auditoria`

---

### T78: Ligar middlewares e handlers no router com lista pública explícita e rotas negadas por padrão

**What**: Ligar middlewares e handlers no router com lista pública explícita e rotas negadas por padrão.  
**Where**: `api/internal/httpapi/router.go`  
**Depends on**: T77, T72, T73, T5  
**Reuses**: Router atual e `gin.Logger` a substituir  
**Requirement**: RBAC-02 (ACs 1, 4); API-02 (ACs 1, 2); PLT-01 (ACs 3)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Teste percorre `Routes()` e falha se uma rota não estiver na lista pública nem protegida
- [ ] Substitui o `gin.Logger` pelo log de acesso próprio
- [ ] `/healthz` continua público e as únicas outras rotas públicas são o login e as duas de recuperação de acesso, com checagem de origem
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 5 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(httpapi): liga middlewares e nega rotas por padrao`

---

### T79: Criar a lógica do primeiro administrador

**What**: Criar a lógica do primeiro administrador.  
**Where**: `api/internal/identity/app/bootstrap_admin.go`  
**Depends on**: T78  
**Reuses**: `create_user.go`, `AdminMembership` e `password`  
**Requirement**: IDN-01 (ACs 1, 2, 3); IDN-05 (ACs 1)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Cria usuário ativo com o papel ADMIN_SISTEMA (padrão) ou PRESIDENTE, com vínculo administrativo de motivo `bootstrap-admin` e sem quem concedeu, e grava `user.bootstrap` sem ator, com e-mail, papel e motivo e nunca a senha
- [ ] Recusa se já existir qualquer vínculo administrativo ativo
- [ ] Recusa senha ausente ou fora da política de administrador (10 caracteres)
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 4 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona logica do primeiro administrador`

---

### T80: Criar o comando `bootstrap-admin`, com a senha só por variável de ambiente

**What**: Criar o comando `bootstrap-admin`, com a senha só por variável de ambiente.  
**Where**: `api/cmd/bootstrap-admin/main.go`  
**Depends on**: T79  
**Reuses**: `bootstrap_admin.go`  
**Requirement**: IDN-01 (ACs 1, 4)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] Não existe flag de senha
- [ ] Sem `BOOTSTRAP_ADMIN_PASSWORD` sai com código diferente de zero
- [ ] `--role` aceita só ADMIN_SISTEMA (padrão) ou PRESIDENTE
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(cmd): adiciona comando bootstrap-admin`

---

### T81: Ligar config, logger, sincronização de papéis (com auditoria do que mudou) e router no ponto de entrada da API

**What**: Ligar config, logger, sincronização de papéis (com auditoria do que mudou) e router no ponto de entrada da API.  
**Where**: `api/cmd/api/main.go`  
**Depends on**: T80, T1  
**Reuses**: `main.go` atual  
**Requirement**: PLT-01 (ACs 1, 5, 6); RBAC-01 (ACs 2)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Falha de configuração encerra o processo nomeando a variável
- [ ] Sincronização de papéis roda na partida
- [ ] Migrações não rodam na partida
- [ ] Build passa
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `feat(cmd): liga a API com config, logger e sincronizacao de papeis`

---

### T82: Criar o teste ponta a ponta do fluxo de identidade e auditoria via HTTP

**What**: Criar o teste ponta a ponta do fluxo de identidade e auditoria via HTTP.  
**Where**: `api/internal/httpapi/e2e_test.go`  
**Depends on**: T81  
**Reuses**: `testutil.NewTestDB` e o validador de contrato  
**Requirement**: IDN-02 (ACs 1, 9); IDN-03 (ACs 2, 3); IDN-04 (ACs 1, 3, 7); IDN-06 (ACs 1, 6); RBAC-03 (ACs 1); RBAC-02 (ACs 1, 2); AUD-03 (ACs 1); AUD-04 (ACs 1, 2, 3, 4, 5); IDN-07 (ACs 1, 5, 6); IDN-08 (ACs 1)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Fluxo: bootstrap, login, criar usuário, promover a TESOURARIA com motivo, trocar a senha obrigatória, recuperar o acesso por e-mail (solicitação, token, redefinição e sessões antigas revogadas), disparar a redefinição administrativa, atribuir papel, retirar o acesso, desativar, consultar auditoria
- [ ] Casos negativos: sem CSRF, origem inválida, sem permissão, sem sessão, escalada de privilégio, autoalteração e último administrador
- [ ] Eventos de segurança consultáveis: `auth.login`, `auth.login_failed`, `auth.login_blocked`, `auth.logout`, `admin.promote`, `admin.revoke`, `user.roles_set` e `role.change_denied`
- [ ] Todas as respostas validadas contra o contrato
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 22 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `test(httpapi): adiciona fluxo ponta a ponta de identidade`

---

### Phase 12: Guardrails e documentação (tarefas)

### T83: Criar o teste de fronteiras entre módulos, baseado em análise de imports

**What**: Criar o teste de fronteiras entre módulos, baseado em análise de imports.  
**Where**: `api/internal/architecture_test.go`  
**Depends on**: None  
**Reuses**: `docs/architecture/domain-boundaries.md`  
**Requirement**: TST-02 (ACs 1, 2, 3)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Falha se um módulo importa `domain`, `infra` ou `http` de outro módulo
- [ ] Falha se `platform` importa um módulo de negócio
- [ ] Fixtures violando cada regra são detectadas (testes negativos)
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `test(arch): verifica fronteiras entre modulos`

---

### T84: Registrar AD-012 a AD-014 em `.specs/STATE.md`

**What**: Registrar AD-012 a AD-014 em `.specs/STATE.md`.  
**Where**: `.specs/STATE.md`  
**Depends on**: T83  
**Reuses**: Formato dos AD existentes  
**Requirement**: API-01 (ACs 1); TST-01 (ACs 1); RBAC-02 (ACs 3); API-02 (ACs 1)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Três decisões novas (AD-012 a AD-014), sem tocar na seção Handoff além de atualizá-la
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `docs(specs): registra decisoes da fundacao`

---

### T85: Atualizar a visão geral da arquitetura com os pacotes de `platform`, os papéis de banco e o contrato

**What**: Atualizar a visão geral da arquitetura com os pacotes de `platform`, os papéis de banco e o contrato.  
**Where**: `docs/architecture/architecture-overview.md`  
**Depends on**: T84  
**Reuses**: Documento atual  
**Requirement**: API-01 (ACs 1)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Estrutura de pastas reflete o código real
- [ ] Lint de markdown e links passam
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `docs(architecture): atualiza visao geral apos a fundacao`

---

### T86: Documentar como rodar testes de integração, migrar e criar o primeiro administrador

**What**: Documentar como rodar testes de integração, migrar e criar o primeiro administrador.  
**Where**: `docs/CONTRIBUTING.md`  
**Depends on**: T85  
**Reuses**: Documento atual  
**Requirement**: TST-01 (ACs 1, 5); IDN-01 (ACs 1)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Comandos copiáveis e conferidos
- [ ] Lint de markdown e links passam
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `docs: documenta testes de integracao e bootstrap do administrador`

---

## Task Granularity Check

| Task | Scope | Status |
| ---- | ----- | ------ |
| T1 | 1 arquivo principal, com seu teste | ✅ Granular |
| T2 | 1 arquivo principal, com seu teste | ✅ Granular |
| T3 | 1 arquivo principal, com seu teste | ✅ Granular |
| T4 | 1 arquivo principal, com seu teste | ✅ Granular |
| T5 | 1 arquivo principal, com seu teste | ✅ Granular |
| T6 | 1 arquivo principal, com seu teste | ✅ Granular |
| T7 | 1 arquivo principal, com seu teste | ✅ Granular |
| T8 | 1 arquivo principal, com seu teste | ✅ Granular |
| T9 | 1 arquivo principal, com seu teste | ✅ Granular |
| T10 | 1 arquivo principal, com seu teste | ✅ Granular |
| T11 | 1 arquivo principal, com seu teste | ✅ Granular |
| T12 | 1 arquivo principal, com seu teste | ✅ Granular |
| T13 | 1 arquivo principal, com seu teste | ✅ Granular |
| T14 | 1 arquivo principal, com seu teste | ✅ Granular |
| T15 | 1 arquivo principal, com seu teste | ✅ Granular |
| T16 | 1 arquivo principal, com seu teste | ✅ Granular |
| T17 | 1 arquivo principal, com seu teste | ✅ Granular |
| T18 | 1 arquivo principal, com seu teste | ✅ Granular |
| T19 | 1 arquivo principal, com seu teste | ✅ Granular |
| T20 | 1 arquivo principal, com seu teste | ✅ Granular |
| T21 | 1 arquivo principal, com seu teste | ✅ Granular |
| T22 | 1 arquivo principal, com seu teste | ✅ Granular |
| T23 | 1 arquivo principal, com seu teste | ✅ Granular |
| T24 | 1 arquivo principal, com seu teste | ✅ Granular |
| T25 | 1 arquivo principal, com seu teste | ✅ Granular |
| T26 | 1 arquivo principal, com seu teste | ✅ Granular |
| T27 | 1 arquivo principal, com seu teste | ✅ Granular |
| T28 | 1 arquivo principal, com seu teste | ✅ Granular |
| T29 | 1 arquivo principal, com seu teste | ✅ Granular |
| T30 | 1 arquivo principal, com seu teste | ✅ Granular |
| T31 | 1 arquivo principal, com seu teste | ✅ Granular |
| T32 | 1 arquivo principal, com seu teste | ✅ Granular |
| T33 | 1 arquivo principal, com seu teste | ✅ Granular |
| T34 | 1 arquivo principal, com seu teste | ✅ Granular |
| T35 | 1 arquivo principal, com seu teste | ✅ Granular |
| T36 | 1 arquivo principal, com seu teste | ✅ Granular |
| T37 | 1 arquivo principal, com seu teste | ✅ Granular |
| T38 | 1 arquivo principal, com seu teste | ✅ Granular |
| T39 | 1 arquivo principal, com seu teste | ✅ Granular |
| T40 | 1 arquivo principal, com seu teste | ✅ Granular |
| T41 | 1 arquivo principal, com seu teste | ✅ Granular |
| T42 | 1 arquivo principal, com seu teste | ✅ Granular |
| T43 | 1 arquivo principal, com seu teste | ✅ Granular |
| T44 | 1 arquivo principal, com seu teste | ✅ Granular |
| T45 | 1 arquivo principal, com seu teste | ✅ Granular |
| T46 | 1 arquivo principal, com seu teste | ✅ Granular |
| T47 | 1 arquivo principal, com seu teste | ✅ Granular |
| T48 | 1 arquivo principal, com seu teste | ✅ Granular |
| T49 | 1 arquivo principal, com seu teste | ✅ Granular |
| T50 | 1 arquivo principal, com seu teste | ✅ Granular |
| T51 | 1 arquivo principal, com seu teste | ✅ Granular |
| T52 | 1 arquivo principal, com seu teste | ✅ Granular |
| T53 | 1 arquivo principal, com seu teste | ✅ Granular |
| T54 | 1 arquivo principal, com seu teste | ✅ Granular |
| T55 | 1 arquivo principal, com seu teste | ✅ Granular |
| T56 | 1 arquivo principal, com seu teste | ✅ Granular |
| T57 | 1 arquivo principal, com seu teste | ✅ Granular |
| T58 | 1 arquivo principal, com seu teste | ✅ Granular |
| T59 | 1 arquivo principal, com seu teste | ✅ Granular |
| T60 | 1 arquivo principal, com seu teste | ✅ Granular |
| T61 | 1 arquivo principal, com seu teste | ✅ Granular |
| T62 | 1 arquivo principal, com seu teste | ✅ Granular |
| T63 | 1 arquivo principal, com seu teste | ✅ Granular |
| T64 | 1 arquivo principal, com seu teste | ✅ Granular |
| T65 | 1 arquivo principal, com seu teste | ✅ Granular |
| T66 | 1 arquivo principal, com seu teste | ✅ Granular |
| T67 | 1 arquivo principal, com seu teste | ✅ Granular |
| T68 | 1 arquivo principal, com seu teste | ✅ Granular |
| T69 | 1 arquivo principal, com seu teste | ✅ Granular |
| T70 | 1 arquivo principal, com seu teste | ✅ Granular |
| T71 | 1 arquivo principal, com seu teste | ✅ Granular |
| T72 | 1 arquivo principal, com seu teste | ✅ Granular |
| T73 | 1 arquivo principal, com seu teste | ✅ Granular |
| T74 | 1 arquivo principal, com seu teste | ✅ Granular |
| T75 | 1 arquivo principal, com seu teste | ✅ Granular |
| T76 | 1 arquivo principal, com seu teste | ✅ Granular |
| T77 | 1 arquivo principal, com seu teste | ✅ Granular |
| T78 | 1 arquivo principal, com seu teste | ✅ Granular |
| T79 | 1 arquivo principal, com seu teste | ✅ Granular |
| T80 | 1 arquivo principal, com seu teste | ✅ Granular |
| T81 | 1 arquivo principal, com seu teste | ✅ Granular |
| T82 | 1 arquivo principal, com seu teste | ✅ Granular |
| T83 | 1 arquivo principal, com seu teste | ✅ Granular |
| T84 | 1 arquivo principal, com seu teste | ✅ Granular |
| T85 | 1 arquivo principal, com seu teste | ✅ Granular |
| T86 | 1 arquivo principal, com seu teste | ✅ Granular |

## Diagram-Definition Cross-Check

Dentro de cada fase, cada tarefa depende apenas da anterior (execução sequencial); dependências entre fases apontam sempre para trás.

| Task | Depends On (task body) | Diagram Shows | Status |
| ---- | ---------------------- | ------------- | ------ |
| T1 | - | - | ✅ Match |
| T2 | T1 | T1 | ✅ Match |
| T3 | T2 | T2 | ✅ Match |
| T4 | T3 | T3 | ✅ Match |
| T5 | T4 | T4 | ✅ Match |
| T6 | - | - | ✅ Match |
| T7 | T6 | T6 | ✅ Match |
| T8 | T7 | T7 | ✅ Match |
| T9 | T8 | T8 | ✅ Match |
| T10 | T9 | T9 | ✅ Match |
| T11 | T10 | T10 | ✅ Match |
| T12 | T11 | T11 | ✅ Match |
| T13 | T12 | T12 | ✅ Match |
| T14 | - | - | ✅ Match |
| T15 | T14 | T14 | ✅ Match |
| T16 | T15 | T15 | ✅ Match |
| T17 | - | - | ✅ Match |
| T18 | T17 | T17 | ✅ Match |
| T19 | T18 | T18 | ✅ Match |
| T20 | T19 | T19 | ✅ Match |
| T21 | T20 | T20 | ✅ Match |
| T22 | T21 | T21 | ✅ Match |
| T23 | - | - | ✅ Match |
| T24 | T23 | T23 | ✅ Match |
| T25 | T24 | T24 | ✅ Match |
| T26 | T25 | T25 | ✅ Match |
| T27 | T26 | T26 | ✅ Match |
| T28 | T27 | T27 | ✅ Match |
| T29 | T28 | T28 | ✅ Match |
| T30 | T29 | T29 | ✅ Match |
| T31 | - | - | ✅ Match |
| T32 | T31 | T31 | ✅ Match |
| T33 | T32 | T32 | ✅ Match |
| T34 | T33 | T33 | ✅ Match |
| T35 | T34 | T34 | ✅ Match |
| T36 | - | - | ✅ Match |
| T37 | T36 | T36 | ✅ Match |
| T38 | T37 | T37 | ✅ Match |
| T39 | T38 | T38 | ✅ Match |
| T40 | T39 | T39 | ✅ Match |
| T41 | T40 | T40 | ✅ Match |
| T42 | T41 | T41 | ✅ Match |
| T43 | - | - | ✅ Match |
| T44 | T43 | T43 | ✅ Match |
| T45 | T44 | T44 | ✅ Match |
| T46 | T45 | T45 | ✅ Match |
| T47 | T46 | T46 | ✅ Match |
| T48 | T47 | T47 | ✅ Match |
| T49 | T48 | T48 | ✅ Match |
| T50 | T49 | T49 | ✅ Match |
| T51 | T50 | T50 | ✅ Match |
| T52 | T51 | T51 | ✅ Match |
| T53 | T52 | T52 | ✅ Match |
| T54 | T53 | T53 | ✅ Match |
| T55 | T54 | T54 | ✅ Match |
| T56 | - | - | ✅ Match |
| T57 | T56 | T56 | ✅ Match |
| T58 | T57 | T57 | ✅ Match |
| T59 | T58 | T58 | ✅ Match |
| T60 | T59 | T59 | ✅ Match |
| T61 | T60 | T60 | ✅ Match |
| T62 | T61 | T61 | ✅ Match |
| T63 | T62 | T62 | ✅ Match |
| T64 | T63 | T63 | ✅ Match |
| T65 | T64 | T64 | ✅ Match |
| T66 | T65 | T65 | ✅ Match |
| T67 | T66 | T66 | ✅ Match |
| T68 | T67 | T67 | ✅ Match |
| T69 | - | - | ✅ Match |
| T70 | T69 | T69 | ✅ Match |
| T71 | T70 | T70 | ✅ Match |
| T72 | T71 | T71 | ✅ Match |
| T73 | T72 | T72 | ✅ Match |
| T74 | - | - | ✅ Match |
| T75 | T74 | T74 | ✅ Match |
| T76 | T75 | T75 | ✅ Match |
| T77 | T76 | T76 | ✅ Match |
| T78 | T77 | T77 | ✅ Match |
| T79 | T78 | T78 | ✅ Match |
| T80 | T79 | T79 | ✅ Match |
| T81 | T80 | T80 | ✅ Match |
| T82 | T81 | T81 | ✅ Match |
| T83 | - | - | ✅ Match |
| T84 | T83 | T83 | ✅ Match |
| T85 | T84 | T84 | ✅ Match |
| T86 | T85 | T85 | ✅ Match |

## Test Co-location Validation

| Task | Code Layer Created/Modified | Matrix Requires | Task Says | Status |
| ---- | --------------------------- | --------------- | --------- | ------ |
| T1: config | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T2: logx | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T3: reqid | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T4: problem | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T5: mw | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T6: dbopen | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T7: roles_sql | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T8: compose | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T9: envex | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T10: tu_pg | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T11: migrate | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T12: tu_db | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T13: ci_int | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T14: web_dep | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T15: web_cfg | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T16: ci_web | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T17: m_core | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T18: m_fmt | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T19: m_json | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T20: m_alloc | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T21: m_guard | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T22: m_ts | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T23: oa_common | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T24: oa_spike | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T25: oa_platform | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T26: oa_gen | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T27: oa_contract | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T28: oa_routes | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T29: oa_ts | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T30: oa_ci | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T31: a_mig | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T32: a_tx | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T33: a_entry | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T34: a_rec | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T35: a_atom | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T36: r_mig | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T37: r_perm | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T38: r_authz | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T39: r_denied | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T40: r_escal | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T41: r_matrix | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T42: r_repo | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T43: cfg_auth | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T44: a_actions | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T45: mailer | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T46: pw | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T47: pw_deny | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T48: u_dom | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T49: am_dom | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T50: u_repo | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T51: am_repo | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T52: s_repo | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T53: at_repo | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T54: authn | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T55: sess | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T56: r_reset_perm | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T57: prt_mig | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T58: prt_repo | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T59: uc_create | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T60: uc_deact | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T61: uc_pw | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T62: uc_promote | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T63: uc_revoke | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T64: uc_roles | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T65: uc_list | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T66: uc_reset_req | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T67: uc_reset_do | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T68: uc_admin_reset | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T69: oa_identity | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T70: oa_identity_gen | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T71: oa_audit | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T72: mw_authn | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T73: mw_csrf | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T74: h_auth | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T75: h_users | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T76: h_admin | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T77: h_audit | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T78: router | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T79: boot_app | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T80: boot_cli | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T81: main | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T82: e2e | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T83: arch | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T84: state | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T85: docs_arch | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T86: docs_contrib | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |

## Requirement Coverage

| Requirement ID | Tasks |
| -------------- | ----- |
| IDN-01 | T46, T48, T79, T80, T86 |
| IDN-02 | T36, T52, T53, T54, T55, T69, T72, T74, T82 |
| IDN-03 | T73, T82 |
| IDN-04 | T46, T48, T50, T57, T58, T59, T60, T61, T64, T65, T69, T74, T75, T82 |
| IDN-05 | T36, T47, T48, T59, T61, T62, T67, T68, T79 |
| IDN-06 | T36, T49, T51, T62, T63, T64, T69, T76, T82 |
| IDN-07 | T43, T57, T58, T66, T67, T69, T73, T74, T82 |
| IDN-08 | T56, T68, T69, T76, T82 |
| RBAC-01 | T36, T37, T41, T42, T64, T81 |
| RBAC-02 | T37, T38, T39, T55, T59, T60, T62, T63, T64, T68, T72, T78, T82, T84 |
| RBAC-03 | T40, T60, T62, T63, T64, T68, T76, T82 |
| AUD-01 | T32, T33, T34, T35 |
| AUD-02 | T31 |
| AUD-03 | T71, T77, T82 |
| AUD-04 | T33, T34, T39, T44, T54, T55, T66, T82 |
| MNY-01 | T17, T20 |
| MNY-02 | T18, T19, T21, T22 |
| MNY-03 | T20 |
| TST-01 | T7, T8, T10, T11, T12, T13, T84, T86 |
| TST-02 | T83 |
| TST-03 | T14, T15, T16, T22 |
| API-01 | T23, T24, T25, T26, T27, T28, T29, T30, T69, T70, T71, T84, T85 |
| API-02 | T3, T4, T5, T23, T25, T65, T78, T84 |
| PLT-01 | T1, T2, T5, T6, T7, T8, T9, T11, T43, T78, T81 |
| EML-01 | T43, T45, T66 |

**Coverage:** 25 total, 25 mapped to tasks, 0 unmapped
