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
T23 → T24 → T25 → T26 → T27 → T28 → T29
```

### Phase 6: Auditoria

```
T30 → T31 → T32 → T33 → T34
```

### Phase 7: Autorização e papéis

```
T35 → T36 → T37 → T38 → T39
```

### Phase 8: Identidade: base

```
T40 → T41 → T42 → T43 → T44 → T45 → T46 → T47
```

### Phase 9: Identidade: casos de uso

```
T48 → T49 → T50 → T51 → T52
```

### Phase 10: HTTP: contrato de identidade e middlewares

```
T53 → T54 → T55
```

### Phase 11: HTTP: handlers e fiação

```
T56 → T57 → T58 → T59 → T60 → T61 → T62 → T63
```

### Phase 12: Guardrails e documentação

```
T64 → T65 → T66 → T67
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

- [ ] `pnpm install --frozen-lockfile` passa com o lockfile atualizado
- [ ] Script `pnpm test` existe
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

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

- [ ] `pnpm test` executa `*.test.ts` e `*.test.tsx` em jsdom
- [ ] O teste de fumaça (`web/lib/smoke.test.ts`) passa
- [ ] Lint e `tsc` continuam passando
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

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

- [ ] `pnpm test` roda antes de `pnpm build` no job `web`
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

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

- [ ] Overflow de soma e subtração devolve `ErrOverflow` sem dar a volta
- [ ] Nenhuma função aceita ou devolve float32 ou float64 (verificado por teste que inspeciona a API do pacote)
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 5 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

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

- [ ] 123456 vira `R$ 1.234,56` e -123456 vira `-R$ 1.234,56` com espaço comum
- [ ] `Parse` rejeita mais de duas casas decimais e caracteres inválidos
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

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

- [ ] Marshal emite literal inteiro; unmarshal rejeita string, decimal, expoente e null
- [ ] Valor acima de 9007199254740991 em módulo devolve erro mapeável para 422 `amount_out_of_range`
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

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

- [ ] `Allocate(10000,[1,1,1])` devolve 3334, 3333, 3333
- [ ] `Percent(10000,400)` devolve 400 e `Percent(1,5000)` devolve 1
- [ ] Teste de propriedade: a soma das partes é sempre o total
- [ ] Cobertura de instruções do pacote `money` maior ou igual a 95%
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 7 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

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

- [ ] Passa no esquema atual
- [ ] Um esquema de fixture com coluna proibida faz o teste falhar (teste negativo)
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

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

- [ ] Mesmos vetores do Go passam
- [ ] Número que não é inteiro seguro lança `RangeError`
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 5 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: unit  
**Gate**: build

**Commit**: `feat(web): adiciona utilitario de dinheiro`

---

### Phase 5: Contrato OpenAPI (tarefas)

### T23: Criar o esqueleto do contrato OpenAPI 3.0.3 com `/healthz`, esquema de erro, `Cents` e parâmetros de paginação

**What**: Criar o esqueleto do contrato OpenAPI 3.0.3 com `/healthz`, esquema de erro, `Cents` e parâmetros de paginação.  
**Where**: `api/openapi/openapi.yaml`  
**Depends on**: None  
**Reuses**: Handler `/healthz` existente  
**Requirement**: API-01 (ACs 1); API-02 (ACs 1, 5)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Documento válido em OpenAPI 3.0.3
- [ ] `/healthz` descrito com as respostas 200 e 503
- [ ] Esquemas `Problem`, `Cents`, `Cursor` e `Limit` definidos
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `feat(api): adiciona esqueleto do contrato openapi`

---

### T24: Escolher e fixar o gerador de servidor Gin/modelos e o linter de OpenAPI após checar a documentação vigente, e configurar a geração

**What**: Escolher e fixar o gerador de servidor Gin/modelos e o linter de OpenAPI após checar a documentação vigente, e configurar a geração.  
**Where**: `api/oapi-codegen.yaml`  
**Depends on**: T23  
**Reuses**: Contrato da tarefa anterior  
**Requirement**: API-01 (ACs 2, 4)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Versões fixadas com a fonte consultada registrada no commit
- [ ] Comando de geração reproduzível documentado
- [ ] Se alguma ferramenta estiver descontinuada, a escolha vai a novo ADR antes de seguir
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `chore(api): configura geracao de codigo openapi`

---

### T25: Criar o gatilho `go generate` e versionar o código gerado (interface do servidor e modelos)

**What**: Criar o gatilho `go generate` e versionar o código gerado (interface do servidor e modelos).  
**Where**: `api/internal/platform/apigen/generate.go`  
**Depends on**: T24  
**Reuses**: `oapi-codegen.yaml`  
**Requirement**: API-01 (ACs 2)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Gerar duas vezes produz zero diferença
- [ ] O router atual passa a implementar a interface gerada para `/healthz`
- [ ] Testes existentes continuam passando
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(api): gera interface do servidor a partir do openapi`

---

### T26: Criar o validador de respostas contra o contrato para uso nos testes

**What**: Criar o validador de respostas contra o contrato para uso nos testes.  
**Where**: `api/internal/platform/testutil/contract.go`  
**Depends on**: T25  
**Reuses**: Contrato e código gerado  
**Requirement**: API-01 (ACs 5)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Resposta conforme passa; status, cabeçalho ou corpo divergente falha o teste
- [ ] `/healthz` é validado por ele
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `test(platform): valida respostas contra o openapi`

---

### T27: Criar o teste que compara as rotas registradas no Gin com as operações do contrato

**What**: Criar o teste que compara as rotas registradas no Gin com as operações do contrato.  
**Where**: `api/internal/httpapi/routes_test.go`  
**Depends on**: T26  
**Reuses**: `gin.Engine.Routes()`  
**Requirement**: API-01 (ACs 6)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Rota sem operação no contrato falha o teste
- [ ] Operação sem rota também falha (teste negativo com fixture)
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `test(httpapi): garante paridade entre rotas e contrato`

---

### T28: Gerar os tipos TypeScript do contrato e versionar `web/lib/api/schema.d.ts`

**What**: Gerar os tipos TypeScript do contrato e versionar `web/lib/api/schema.d.ts`.  
**Where**: `web/package.json`  
**Depends on**: T27, T14  
**Reuses**: Contrato e `web/`  
**Requirement**: API-01 (ACs 3)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Script `gen:api` gera o arquivo sem diferença na segunda execução
- [ ] `tsc` passa com o arquivo gerado
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `feat(web): gera tipos do contrato openapi`

---

### T29: Validar o contrato no CI: lint e falha se a geração de Go ou TypeScript produzir diferença

**What**: Validar o contrato no CI: lint e falha se a geração de Go ou TypeScript produzir diferença.  
**Where**: `.github/workflows/ci.yml`  
**Depends on**: T28, T13, T16  
**Reuses**: Jobs `api` e `web`  
**Requirement**: API-01 (ACs 2, 3, 4)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Alterar o YAML sem regenerar faz o CI falhar
- [ ] Contrato inválido faz o CI falhar
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `ci: valida contrato openapi e codigo gerado`

---

### Phase 6: Auditoria (tarefas)

### T30: Criar a migração `audit_log` com índices, triggers de imutabilidade (UPDATE, DELETE e TRUNCATE) e concessões ao papel `tj_app`; incluir o arquivo `down`

**What**: Criar a migração `audit_log` com índices, triggers de imutabilidade (UPDATE, DELETE e TRUNCATE) e concessões ao papel `tj_app`; incluir o arquivo `down`.  
**Where**: `api/migrations/000002_audit_log.up.sql`  
**Depends on**: T12, T11  
**Reuses**: Migração 000001 e o padrão de `tj_app`  
**Requirement**: AUD-02 (ACs 1, 2)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Com o papel dono, UPDATE, DELETE e TRUNCATE falham por trigger
- [ ] Com `tj_app`, só INSERT e SELECT funcionam
- [ ] Migração `down` desfaz sem erro
- [ ] Falha com mensagem clara se o papel `tj_app` não existir
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(audit): cria tabela de auditoria imutavel`

---

### T31: Criar `WithTx` e `TxFrom`, a unidade de trabalho que carrega a transação no contexto

**What**: Criar `WithTx` e `TxFrom`, a unidade de trabalho que carrega a transação no contexto.  
**Where**: `api/internal/platform/database/tx.go`  
**Depends on**: T30, T6  
**Reuses**: `database.Open`  
**Requirement**: AUD-01 (ACs 1, 6)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Erro no callback reverte a transação
- [ ] `TxFrom` sem transação devolve `ok=false`
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(platform): adiciona unidade de trabalho transacional`

---

### T32: Criar o tipo `Entry`, sua validação e a redação de chaves sensíveis

**What**: Criar o tipo `Entry`, sua validação e a redação de chaves sensíveis.  
**Where**: `api/internal/platform/audit/entry.go`  
**Depends on**: T31  
**Reuses**: Lista de chaves sensíveis do `logx`  
**Requirement**: AUD-01 (ACs 3, 4, 5)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Ação de cancelamento ou ajuste sem motivo é rejeitada
- [ ] Chaves sensíveis em before e after viram `[redacted]`
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(audit): adiciona entrada de auditoria com validacao e redacao`

---

### T33: Criar o `Recorder` que grava a entrada na transação do contexto

**What**: Criar o `Recorder` que grava a entrada na transação do contexto.  
**Where**: `api/internal/platform/audit/recorder.go`  
**Depends on**: T32  
**Reuses**: `database.TxFrom` e `Entry`  
**Requirement**: AUD-01 (ACs 1, 3, 6)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Sem transação no contexto devolve erro e não grava
- [ ] O registro contém request_id e ator vindos do contexto
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(audit): adiciona gravacao de auditoria na transacao`

---

### T34: Criar o teste de atomicidade: falha na auditoria reverte a alteração e sucesso grava as duas na mesma transação

**What**: Criar o teste de atomicidade: falha na auditoria reverte a alteração e sucesso grava as duas na mesma transação.  
**Where**: `api/internal/platform/audit/atomicity_test.go`  
**Depends on**: T33  
**Reuses**: `Recorder` e `WithTx`  
**Requirement**: AUD-01 (ACs 1, 2)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Falha simulada na auditoria deixa o dado alterado intacto
- [ ] Sucesso grava dado e auditoria juntos
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `test(audit): prova atomicidade entre dado e auditoria`

---

### Phase 7: Autorização e papéis (tarefas)

### T35: Criar a migração de identidade: users, roles, permissions, role_permissions, user_roles, sessions e login_attempts, com `down` e concessões

**What**: Criar a migração de identidade: users, roles, permissions, role_permissions, user_roles, sessions e login_attempts, com `down` e concessões.  
**Where**: `api/migrations/000003_identity.up.sql`  
**Depends on**: T12  
**Reuses**: Migração 000002 como padrão  
**Requirement**: RBAC-01 (ACs 2, 3); IDN-02 (ACs 2); IDN-05 (ACs 5, 6)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Coluna `users.must_change_password` (boolean, padrão falso)
- [ ] Índice único por `lower(email)`
- [ ] `sessions.token_hash` único
- [ ] `tj_app` com SELECT, INSERT, UPDATE e DELETE nas tabelas de identidade
- [ ] Migração `down` desfaz sem erro
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 5 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): cria tabelas de identidade e sessoes`

---

### T36: Criar o tipo `Permission` com validação de formato

**What**: Criar o tipo `Permission` com validação de formato.  
**Where**: `api/internal/platform/authz/permission.go`  
**Depends on**: T35  
**Reuses**: Nenhum (pacote novo)  
**Requirement**: RBAC-01 (ACs 1)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Aceita `financeiro:lancamento:read` e rejeita formas inválidas
- [ ] Regex `^[a-z_]+:[a-z_]+:[a-z_]+$` aplicada
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(authz): adiciona tipo de permissao`

---

### T37: Criar `Principal` e `Require`

**What**: Criar `Principal` e `Require`.  
**Where**: `api/internal/platform/authz/authz.go`  
**Depends on**: T36  
**Reuses**: `Permission`  
**Requirement**: RBAC-02 (ACs 2, 3)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Sem a permissão devolve erro que vira 403 `forbidden`
- [ ] Erro não revela se o recurso existe
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(authz): adiciona principal e verificacao de permissao`

---

### T38: Criar a matriz papel-permissão em código com o teste de invariante do Conselho Fiscal

**What**: Criar a matriz papel-permissão em código com o teste de invariante do Conselho Fiscal.  
**Where**: `api/internal/identity/app/roles_matrix.go`  
**Depends on**: T37  
**Reuses**: `Permission`  
**Requirement**: RBAC-01 (ACs 3, 4)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Seis papéis definidos, com a matriz provisória da spec
- [ ] CONSELHO_FISCAL não tem nenhuma permissão com ação `create`, `update`, `delete` ou `cancel`
- [ ] Adicionar uma permissão proibida ao Conselho faz o teste falhar (teste negativo)
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(identity): define matriz de papeis e permissoes`

---

### T39: Criar o repositório de papéis com sincronização idempotente e cálculo de permissões efetivas

**What**: Criar o repositório de papéis com sincronização idempotente e cálculo de permissões efetivas.  
**Where**: `api/internal/identity/infra/role_repository.go`  
**Depends on**: T38  
**Reuses**: `roles_matrix.go`  
**Requirement**: RBAC-01 (ACs 2, 5, 6)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Sincronizar duas vezes deixa o banco idêntico
- [ ] Permissão removida do código fica inativa e não é apagada
- [ ] Permissões efetivas são a união dos papéis do usuário
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): sincroniza papeis e calcula permissoes efetivas`

---

### Phase 8: Identidade: base (tarefas)

### T40: Criar hash e verificação de senha com argon2id

**What**: Criar hash e verificação de senha com argon2id.  
**Where**: `api/internal/platform/password/argon2.go`  
**Depends on**: None  
**Reuses**: `golang.org/x/crypto/argon2`  
**Requirement**: IDN-04 (ACs 9); IDN-01 (ACs 3)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] Parâmetros iniciais conferidos com a recomendação vigente do OWASP e custo medido
- [ ] Hash inclui parâmetros e sal; verificação em tempo constante
- [ ] Hash nunca aparece em log
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(password): adiciona hash argon2id`

---

### T41: Criar a lista embutida de senhas comuns ou comprometidas e a verificação sem diferenciar maiúsculas

**What**: Criar a lista embutida de senhas comuns ou comprometidas e a verificação sem diferenciar maiúsculas.  
**Where**: `api/internal/platform/password/denylist.go`  
**Depends on**: T40  
**Reuses**: Diretiva `go:embed` sobre `denylist.txt`  
**Requirement**: IDN-05 (ACs 4)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] Fonte e licença da lista registradas no commit; se não houver fonte adequada, gera-se lista própria de senhas comuns
- [ ] `Contains` ignora maiúsculas e minúsculas
- [ ] Arquivo `denylist.txt` embutido no binário, sem consulta externa
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(password): adiciona lista de senhas comprometidas`

---

### T42: Criar a entidade de usuário, a normalização de e-mail e a política de senha por papel (10 para papéis administrativos, 8 só para ASSOCIADO, máximo 128 pontos de código)

**What**: Criar a entidade de usuário, a normalização de e-mail e a política de senha por papel (10 para papéis administrativos, 8 só para ASSOCIADO, máximo 128 pontos de código).  
**Where**: `api/internal/identity/domain/user.go`  
**Depends on**: T41  
**Reuses**: Nenhum (pacote novo)  
**Requirement**: IDN-01 (ACs 3); IDN-04 (ACs 2); IDN-05 (ACs 1, 2, 3, 7)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] E-mail é normalizado em minúsculas
- [ ] Usuário com papel diferente de ASSOCIADO exige 10 caracteres; só ASSOCIADO exige 8; acima de 128 é rejeitado
- [ ] Comprimento medido em pontos de código Unicode e sem regras de composição
- [ ] Conjunto vazio de papéis é inválido
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 6 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(identity): adiciona entidade de usuario e politica de senha`

---

### T43: Criar o repositório de usuários

**What**: Criar o repositório de usuários.  
**Where**: `api/internal/identity/infra/user_repository.go`  
**Depends on**: T42  
**Reuses**: `user.go`  
**Requirement**: IDN-04 (ACs 1, 2, 3)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Criação simultânea do mesmo e-mail deixa exatamente uma vencer (teste concorrente)
- [ ] Consulta por e-mail ignora maiúsculas
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona repositorio de usuarios`

---

### T44: Criar o repositório de sessões (hash do token, revogação, expiração)

**What**: Criar o repositório de sessões (hash do token, revogação, expiração).  
**Where**: `api/internal/identity/infra/session_repository.go`  
**Depends on**: T43  
**Reuses**: `user_repository.go`  
**Requirement**: IDN-02 (ACs 2, 5, 6)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] Só o hash SHA-256 é gravado
- [ ] Revogar todas as sessões de um usuário funciona em uma transação
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona repositorio de sessoes`

---

### T45: Criar o repositório de tentativas de login

**What**: Criar o repositório de tentativas de login.  
**Where**: `api/internal/identity/infra/attempt_repository.go`  
**Depends on**: T44  
**Reuses**: `session_repository.go`  
**Requirement**: IDN-02 (ACs 8)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Conta falhas por hash do e-mail na janela de 15 minutos
- [ ] Sucesso zera a contagem
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 1 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona repositorio de tentativas de login`

---

### T46: Criar o caso de uso de login: verificação de senha, bloqueio por tentativas, erro uniforme e rotação de sessão

**What**: Criar o caso de uso de login: verificação de senha, bloqueio por tentativas, erro uniforme e rotação de sessão.  
**Where**: `api/internal/identity/app/authenticate.go`  
**Depends on**: T45  
**Reuses**: Repositórios da fase, `password`, `audit`  
**Requirement**: IDN-02 (ACs 1, 2, 3, 4, 7, 8, 10)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] Nenhuma sequência de falhas bloqueia a conta de forma permanente
- [ ] Usuário inexistente, senha errada e inativo devolvem o mesmo erro
- [ ] Quinta falha em 15 minutos bloqueia por 15 minutos
- [ ] Login bem-sucedido emite novo token
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 7 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona caso de uso de login`

---

### T47: Criar o serviço de sessão: validação (ociosidade de 60 minutos e teto de 12 horas), logout e atualização limitada de `last_seen_at`

**What**: Criar o serviço de sessão: validação (ociosidade de 60 minutos e teto de 12 horas), logout e atualização limitada de `last_seen_at`.  
**Where**: `api/internal/identity/app/session_service.go`  
**Depends on**: T46  
**Reuses**: Repositórios de sessão e de papéis  
**Requirement**: IDN-02 (ACs 5, 6, 9); RBAC-02 (ACs 5)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] Sessão ociosa ou antiga é inválida
- [ ] Usuário inativo invalida todas as suas sessões
- [ ] `last_seen_at` atualiza no máximo uma vez por minuto
- [ ] Devolve Principal e token CSRF
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 4 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona servico de sessao`

---

### Phase 9: Identidade: casos de uso (tarefas)

### T48: Criar o caso de uso de criação de usuário

**What**: Criar o caso de uso de criação de usuário.  
**Where**: `api/internal/identity/app/create_user.go`  
**Depends on**: None  
**Reuses**: `WithTx`, `audit.Recorder`, `authz.Require`  
**Requirement**: IDN-04 (ACs 1, 2); RBAC-02 (ACs 3); IDN-05 (ACs 1, 2, 3, 4)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Aplica a política de senha por papel e a lista de comprometidas
- [ ] Exige `identity:user:create` antes de qualquer leitura ou escrita
- [ ] Grava `user.create` na mesma transação
- [ ] E-mail repetido devolve erro mapeável para 409 `email_taken`
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 7 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona criacao de usuario`

---

### T49: Criar o caso de uso de desativação e reativação de usuário

**What**: Criar o caso de uso de desativação e reativação de usuário.  
**Where**: `api/internal/identity/app/deactivate_user.go`  
**Depends on**: T48  
**Reuses**: `WithTx` e `audit.Recorder`  
**Requirement**: IDN-04 (ACs 3, 4); RBAC-02 (ACs 3, 5)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Desativar revoga todas as sessões na mesma transação
- [ ] Desativar o último ADMIN ativo devolve erro mapeável para 409 `last_admin`
- [ ] Grava `user.deactivate`
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 4 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona desativacao de usuario`

---

### T50: Criar o caso de uso de troca de senha

**What**: Criar o caso de uso de troca de senha.  
**Where**: `api/internal/identity/app/change_password.go`  
**Depends on**: T49  
**Reuses**: `password` e `audit.Recorder`  
**Requirement**: IDN-04 (ACs 5, 6, 9); IDN-05 (ACs 1, 2, 3, 4, 6)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] Aplica a política de senha e limpa `must_change_password` ao trocar
- [ ] Senha atual errada devolve erro mapeável para 403 `invalid_current_password`
- [ ] Revoga as outras sessões do usuário
- [ ] Auditoria `user.password_change` não contém valor de senha
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 8 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona troca de senha`

---

### T51: Criar o caso de uso de atribuição de papéis

**What**: Criar o caso de uso de atribuição de papéis.  
**Where**: `api/internal/identity/app/assign_roles.go`  
**Depends on**: T50  
**Reuses**: Repositório de papéis e `audit.Recorder`  
**Requirement**: IDN-04 (ACs 7); RBAC-01 (ACs 5); RBAC-02 (ACs 3); IDN-05 (ACs 5)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Acrescentar papel administrativo a quem só era ASSOCIADO marca `must_change_password`
- [ ] Conjunto vazio de papéis devolve 422 `validation_failed`
- [ ] Exige `identity:role:assign`
- [ ] Substitui o conjunto e grava papéis anteriores e novos em `user.roles_set`
- [ ] Novas permissões valem na requisição seguinte
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 4 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona atribuicao de papeis`

---

### T52: Criar o caso de uso de listagem de usuários com paginação por cursor

**What**: Criar o caso de uso de listagem de usuários com paginação por cursor.  
**Where**: `api/internal/identity/app/list_users.go`  
**Depends on**: T51  
**Reuses**: Convenção de cursor  
**Requirement**: IDN-04 (ACs 8); API-02 (ACs 5)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Exige `identity:user:read`
- [ ] Ordena de forma estável, respeita `limit` e devolve `next_cursor`
- [ ] Nunca devolve hash de senha
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona listagem de usuarios`

---

### Phase 10: HTTP: contrato de identidade e middlewares (tarefas)

### T53: Descrever no contrato as rotas de auth, usuários, papéis e auditoria e regenerar o código e os tipos

**What**: Descrever no contrato as rotas de auth, usuários, papéis e auditoria e regenerar o código e os tipos.  
**Where**: `api/openapi/openapi.yaml`  
**Depends on**: T25, T28, T52  
**Reuses**: Contrato base  
**Requirement**: API-01 (ACs 1, 2, 3); IDN-02 (ACs 1, 9); IDN-04 (ACs 1, 8); AUD-03 (ACs 1, 2, 3, 4)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Geração de Go e TypeScript sem diferença após regenerar
- [ ] Cada rota tem respostas de erro `Problem`
- [ ] Cookie de sessão e cabeçalho `X-CSRF-Token` descritos
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `feat(api): descreve rotas de identidade e auditoria no contrato`

---

### T54: Criar o middleware `Authn`: cookie para sessão para Principal no contexto, com 401 padronizado

**What**: Criar o middleware `Authn`: cookie para sessão para Principal no contexto, com 401 padronizado.  
**Where**: `api/internal/platform/httpx/authn.go`  
**Depends on**: T53  
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

### T55: Criar o middleware de CSRF e origem

**What**: Criar o middleware de CSRF e origem.  
**Where**: `api/internal/platform/httpx/csrf.go`  
**Depends on**: T54  
**Reuses**: `ALLOWED_ORIGINS` da config  
**Requirement**: IDN-03 (ACs 1, 2, 3, 4, 5)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] Método que altera estado sem `X-CSRF-Token` correto devolve 403 `csrf_invalid`
- [ ] Origem fora da lista devolve 403 `origin_not_allowed`, inclusive no login
- [ ] GET, HEAD e OPTIONS não exigem token
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 5 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(platform): adiciona protecao csrf e checagem de origem`

---

### Phase 11: HTTP: handlers e fiação (tarefas)

### T56: Criar os handlers de login, logout, `me` e troca de senha

**What**: Criar os handlers de login, logout, `me` e troca de senha.  
**Where**: `api/internal/identity/http/auth_handler.go`  
**Depends on**: None  
**Reuses**: Código gerado, casos de uso e `problem`  
**Requirement**: IDN-02 (ACs 1, 3, 4, 6, 8, 9); IDN-04 (ACs 5, 6)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] `me` inclui `must_change_password`
- [ ] Cookie `tj_session` com HttpOnly, Secure, SameSite=Lax e Path=/
- [ ] Erros idênticos para credenciais inválidas
- [ ] 429 traz `Retry-After`
- [ ] Respostas validadas contra o contrato
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 8 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona handlers de autenticacao`

---

### T57: Criar os handlers de criação, desativação, atribuição de papéis e listagem de usuários

**What**: Criar os handlers de criação, desativação, atribuição de papéis e listagem de usuários.  
**Where**: `api/internal/identity/http/users_handler.go`  
**Depends on**: T56  
**Reuses**: Casos de uso e código gerado  
**Requirement**: IDN-04 (ACs 1, 2, 3, 4, 7, 8, 9)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Respostas nunca contêm hash de senha nem token
- [ ] Erros mapeados: 409 `email_taken` e `last_admin`, 403 `forbidden`
- [ ] Respostas validadas contra o contrato
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 7 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona handlers de usuarios`

---

### T58: Criar o serviço de consulta e o handler `GET /api/v1/audit-logs`

**What**: Criar o serviço de consulta e o handler `GET /api/v1/audit-logs`.  
**Where**: `api/internal/platform/audit/http.go`  
**Depends on**: T57  
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

### T59: Ligar middlewares e handlers no router com lista pública explícita e rotas negadas por padrão

**What**: Ligar middlewares e handlers no router com lista pública explícita e rotas negadas por padrão.  
**Where**: `api/internal/httpapi/router.go`  
**Depends on**: T58, T54, T55, T5  
**Reuses**: Router atual e `gin.Logger` a substituir  
**Requirement**: RBAC-02 (ACs 1, 4); API-02 (ACs 1, 2); PLT-01 (ACs 3)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Teste percorre `Routes()` e falha se uma rota não estiver na lista pública nem protegida
- [ ] Substitui o `gin.Logger` pelo log de acesso próprio
- [ ] `/healthz` continua público
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 5 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(httpapi): liga middlewares e nega rotas por padrao`

---

### T60: Criar a lógica do primeiro administrador

**What**: Criar a lógica do primeiro administrador.  
**Where**: `api/internal/identity/app/bootstrap_admin.go`  
**Depends on**: T59  
**Reuses**: `create_user.go` e `password`  
**Requirement**: IDN-01 (ACs 1, 2, 3); IDN-05 (ACs 1)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Cria ADMIN ativo e grava `user.bootstrap`
- [ ] Recusa se já existir ADMIN
- [ ] Recusa senha ausente ou fora da política de administrador (10 caracteres)
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 4 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(identity): adiciona logica do primeiro administrador`

---

### T61: Criar o comando `bootstrap-admin`, com a senha só por variável de ambiente

**What**: Criar o comando `bootstrap-admin`, com a senha só por variável de ambiente.  
**Where**: `api/cmd/bootstrap-admin/main.go`  
**Depends on**: T60  
**Reuses**: `bootstrap_admin.go`  
**Requirement**: IDN-01 (ACs 1, 4)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] Não existe flag de senha
- [ ] Sem `BOOTSTRAP_ADMIN_PASSWORD` sai com código diferente de zero
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(cmd): adiciona comando bootstrap-admin`

---

### T62: Ligar config, logger, sincronização de papéis e router no ponto de entrada da API

**What**: Ligar config, logger, sincronização de papéis e router no ponto de entrada da API.  
**Where**: `api/cmd/api/main.go`  
**Depends on**: T61, T1  
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

### T63: Criar o teste ponta a ponta do fluxo de identidade e auditoria via HTTP

**What**: Criar o teste ponta a ponta do fluxo de identidade e auditoria via HTTP.  
**Where**: `api/internal/httpapi/e2e_test.go`  
**Depends on**: T62  
**Reuses**: `testutil.NewTestDB` e o validador de contrato  
**Requirement**: IDN-02 (ACs 1, 9); IDN-03 (ACs 2, 3); IDN-04 (ACs 1, 3, 7); RBAC-02 (ACs 1, 2); AUD-03 (ACs 1)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Fluxo: bootstrap, login, criar usuário FINANCEIRO, atribuir papel, desativar, consultar auditoria
- [ ] Casos negativos: sem CSRF, origem inválida, sem permissão, sem sessão
- [ ] Todas as respostas validadas contra o contrato
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 10 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `test(httpapi): adiciona fluxo ponta a ponta de identidade`

---

### Phase 12: Guardrails e documentação (tarefas)

### T64: Criar o teste de fronteiras entre módulos, baseado em análise de imports

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

### T65: Registrar AD-010 a AD-013 em `.specs/STATE.md`, após a aprovação do design

**What**: Registrar AD-010 a AD-013 em `.specs/STATE.md`, após a aprovação do design.  
**Where**: `.specs/STATE.md`  
**Depends on**: T64  
**Reuses**: Formato dos AD existentes  
**Requirement**: API-01 (ACs 1); TST-01 (ACs 1); RBAC-02 (ACs 3); API-02 (ACs 1)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Quatro decisões novas, sem tocar na seção Handoff além de atualizá-la
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./... && go build ./... && cd ../web && pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`

**Tests**: none  
**Gate**: build

**Commit**: `docs(specs): registra decisoes da fundacao`

---

### T66: Atualizar a visão geral da arquitetura com os pacotes de `platform`, os papéis de banco e o contrato

**What**: Atualizar a visão geral da arquitetura com os pacotes de `platform`, os papéis de banco e o contrato.  
**Where**: `docs/architecture/architecture-overview.md`  
**Depends on**: T65  
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

### T67: Documentar como rodar testes de integração, migrar e criar o primeiro administrador

**What**: Documentar como rodar testes de integração, migrar e criar o primeiro administrador.  
**Where**: `docs/CONTRIBUTING.md`  
**Depends on**: T66  
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
| T30 | - | - | ✅ Match |
| T31 | T30 | T30 | ✅ Match |
| T32 | T31 | T31 | ✅ Match |
| T33 | T32 | T32 | ✅ Match |
| T34 | T33 | T33 | ✅ Match |
| T35 | - | - | ✅ Match |
| T36 | T35 | T35 | ✅ Match |
| T37 | T36 | T36 | ✅ Match |
| T38 | T37 | T37 | ✅ Match |
| T39 | T38 | T38 | ✅ Match |
| T40 | - | - | ✅ Match |
| T41 | T40 | T40 | ✅ Match |
| T42 | T41 | T41 | ✅ Match |
| T43 | T42 | T42 | ✅ Match |
| T44 | T43 | T43 | ✅ Match |
| T45 | T44 | T44 | ✅ Match |
| T46 | T45 | T45 | ✅ Match |
| T47 | T46 | T46 | ✅ Match |
| T48 | - | - | ✅ Match |
| T49 | T48 | T48 | ✅ Match |
| T50 | T49 | T49 | ✅ Match |
| T51 | T50 | T50 | ✅ Match |
| T52 | T51 | T51 | ✅ Match |
| T53 | - | - | ✅ Match |
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
| T64 | - | - | ✅ Match |
| T65 | T64 | T64 | ✅ Match |
| T66 | T65 | T65 | ✅ Match |
| T67 | T66 | T66 | ✅ Match |

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
| T23: oa_spec | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T24: oa_cfg | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T25: oa_gen | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T26: oa_contract | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T27: oa_routes | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T28: oa_ts | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T29: oa_ci | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T30: a_mig | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T31: a_tx | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T32: a_entry | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T33: a_rec | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T34: a_atom | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T35: r_mig | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T36: r_perm | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T37: r_authz | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T38: r_matrix | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T39: r_repo | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T40: pw | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T41: pw_deny | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T42: u_dom | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T43: u_repo | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T44: s_repo | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T45: at_repo | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T46: authn | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T47: sess | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T48: uc_create | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T49: uc_deact | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T50: uc_pw | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T51: uc_roles | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T52: uc_list | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T53: oa_paths | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T54: mw_authn | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T55: mw_csrf | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T56: h_auth | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T57: h_users | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T58: h_audit | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T59: router | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T60: boot_app | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T61: boot_cli | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T62: main | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T63: e2e | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T64: arch | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T65: state | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T66: docs_arch | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T67: docs_contrib | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |

## Requirement Coverage

| Requirement ID | Tasks |
| -------------- | ----- |
| IDN-01 | T40, T42, T60, T61, T67 |
| IDN-02 | T35, T44, T45, T46, T47, T53, T54, T56, T63 |
| IDN-03 | T55, T63 |
| IDN-04 | T40, T42, T43, T48, T49, T50, T51, T52, T53, T56, T57, T63 |
| IDN-05 | T35, T41, T42, T48, T50, T51, T60 |
| RBAC-01 | T35, T36, T38, T39, T51, T62 |
| RBAC-02 | T37, T47, T48, T49, T51, T54, T59, T63, T65 |
| AUD-01 | T31, T32, T33, T34 |
| AUD-02 | T30 |
| AUD-03 | T53, T58, T63 |
| MNY-01 | T17, T20 |
| MNY-02 | T18, T19, T21, T22 |
| MNY-03 | T20 |
| TST-01 | T7, T8, T10, T11, T12, T13, T65, T67 |
| TST-02 | T64 |
| TST-03 | T14, T15, T16, T22 |
| API-01 | T23, T24, T25, T26, T27, T28, T29, T53, T65, T66 |
| API-02 | T3, T4, T5, T23, T52, T59, T65 |
| PLT-01 | T1, T2, T5, T6, T7, T8, T9, T11, T59, T62 |

**Coverage:** 19 total, 19 mapped to tasks, 0 unmapped
