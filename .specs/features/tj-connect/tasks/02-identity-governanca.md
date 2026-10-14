# Identity e governança — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules.

**Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/tj-connect/spec/02-identity-governanca.md`
**Design**: `.specs/features/tj-connect/design/02-identity-governanca.md`
**Status**: Draft — aguardando revisão do mantenedor. **Execute não autorizado.** Tests-first, um commit atômico por tarefa, só local (push, PR, merge e deploy exigem autorização explícita).

**Pré-requisitos**: nenhum para as fases 1 e 2. As migrations `000010` e `000011` são criadas na fase 3 e 4 e só ficam congeladas quando chegarem à `develop`; até lá, tarefas da mesma fase podem completar o mesmo arquivo de migration.

**Decisões cobertas**: D-02-1, D-02-2, D-02-3, D-02-4, D-02-5 e os ajustes obrigatórios 1 (RBAC durante o deploy) e 3 (integridade das contas de serviço).

**Regra de versão da matriz**: a partir de T10, toda tarefa que muda a matriz agregada (permissão, classe, papel, concessão ou conta de serviço) incrementa `MatrixVersion` e atualiza `testdata/rbac_matrix_version.json` no mesmo commit. Dentro da branch, a versão pode subir mais de uma vez; só a monotonicidade importa.

## Test Coverage Matrix

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| `platform/authz` (classe, `CoversExcluding`, `SeparateFrom`) | unit | todos os ramos de validação e cada erro nomeado | `api/internal/platform/authz/*_test.go` | `go test ./internal/platform/authz/...` |
| `identity/domain` (papéis, tipo de conta, conflitos) | unit | cada papel, tipo e par incompatível | `api/internal/identity/domain/*_test.go` | `go test ./internal/identity/domain/...` |
| `identity/app` (`BuildMatrix`, contribuições) | unit + snapshot | cada violação fail-closed; snapshot papel a papel com exceções aprovadas | `api/internal/identity/app/roles_matrix_test.go`, `testdata/*.json` | `go test ./internal/identity/app/...` |
| `identity/app` (casos de uso) | unit com fakes + integration | caminho aceito, cada recusa com o código e a auditoria | `api/internal/identity/app/*_test.go` | `go test ./...` e `go test -tags=integration ./...` |
| `identity/infra` (repositórios, `Sync`, barreira de versão) | integration (PostgreSQL real) | escrita, diff, auditoria, concorrência | `api/internal/identity/infra/*_test.go` | `go test -tags=integration ./internal/identity/...` |
| Migrations `000010`, `000011` | integration (PostgreSQL real, executando como `tj_app`) | up, down, grants, `CHECK`, gatilhos nos dois sentidos, `down` recusado com dados | `api/internal/platform/database/*_migration_test.go` | `go test -tags=integration ./internal/platform/database/...` |
| `platform/audit` | unit + integration | tipo de ator, contexto obrigatório, resolvedor, filtro | `api/internal/platform/audit/*_test.go` | `go test ./internal/platform/audit/...` |
| `composition` (lista de contribuições, versão da matriz) | unit | hash e versão; mesma lista nos três binários | `api/internal/composition/*_test.go` | `go test ./internal/composition/...` |
| `cmd/api`, `cmd/bootstrap-admin` (composição, comandos) | integration | partida, sincronização condicionada, `rbac-rollback` | `api/cmd/*/*_test.go` | `go test -tags=integration ./cmd/...` |
| `identity/http`, `platform/api` (handlers) | integration (httptest + PostgreSQL) | permissão, CSRF, problem+json, códigos novos | `api/internal/identity/http/*_test.go`, `api/internal/httpapi/*_test.go` | `go test -tags=integration ./internal/...` |
| Contrato OpenAPI | lint + geração | contrato válido; tipos Go e TS gerados sem diferença manual | `api/openapi/*.yaml` | Gate `Contract` |
| Guardas de código-fonte | unit (leitura de fontes) | uso restrito de `CoversExcluding` e `ActorSystem` | `api/internal/**/*_source_test.go` | `go test ./...` |
| Documentação de operação | none | revisão humana | `docs/operations/*.md` | — |

## Gate Check Commands

| Level | When | Command |
| --- | --- | --- |
| Full | toda tarefa de código | `cd api && go vet ./... && go vet -tags=integration ./... && go build ./... && go test ./... && go test -tags=integration ./...` |
| Lint | toda tarefa de código | `cd api && golangci-lint run --path-mode=abs --build-tags=integration` |
| Architecture | toda tarefa de código | `cd api && go test ./internal/... -run TestProductionCodeRespectsTheModuleBoundaries` |
| Contract | tarefas que mudam `api/openapi/*.yaml` | `cd web && pnpm lint:api && pnpm gen:api && pnpm exec tsc --noEmit` e `cd api && go generate ./... && git diff --exit-code -- internal/` (depois de gerar e versionar) |
| Docs | tarefas só de documentação | revisão humana; `python .claude/skills/tlc-spec-driven/scripts/validate_spec.py --strict` quando a spec muda |

Todos os testes de integração usam testcontainers (`platform/testutil/postgres.go`), com PostgreSQL 16 real, e exigem Docker. Nenhum teste chama serviço externo.

---

## Execution Plan

### Phase 1: Classificação e snapshot (I-1)

Matriz efetiva idêntica à atual (snapshot com zero exceções).

```
T2 → T3 → T6
T2 → T4 → T6
T2 → T5 → T6
T1 → T6
```

### Phase 2: Lista explícita do Conselho Fiscal (I-2)

```
T7
```

### Phase 3: Migration 000010, versão do RBAC, integridade e contas de serviço (I-3)

`SeparateFrom` (C1) é antecipado de I-5 para esta fase porque a reativação de conta de serviço (C4) o usa.

```
T8 → T9
T8 → T11 → T12
T10 → T11 → T15 → T18 → T19 → T20
T8 → T13 → T14
T13 → T15
T13 → T16 → T18
T17 → T19
```

### Phase 4: Autoria `service` e rastreabilidade (I-4)

```
T21 → T22 → T23
```

### Phase 5: Contrato de decisão (I-5)

```
T24
```

### Phase 6: D-02-3 + D-02-4 + D-02-5 juntas (I-6)

```
T25 → T26 → T27 → T29 → T30
T25 → T27
T25 → T28 → T29
```

### Phase 7: Verificação final e preparação da implantação (sem implantar)

```
T31 → T32
```

---

## Task Breakdown

### Phase 1: Classificação e snapshot (I-1)

#### T1: Snapshot versionado da matriz pré-classificação

**What**: gerar, **antes de qualquer outra mudança**, o retrato da matriz efetiva atual (papel → permissões) a partir das contribuições reais (`FoundationContributions`, `financeiro.Contribution`, `estoque.Contribution`) e versioná-lo. Criar `excecoes_aprovadas.json` vazio. O gerador só escreve com `UPDATE_SNAPSHOT=1`; sem a variável, o teste compara.
**Where**: `api/internal/identity/app/testdata/matriz_pre_classificacao.json`, `testdata/excecoes_aprovadas.json`, `api/internal/identity/app/matrix_snapshot_test.go`
**Depends on**: None
**Reuses**: `BuildMatrix`, `Matrix.PermissionsOf`.
**Requirement**: TRN-01 AC1, AC3.
**Impacto**: nenhum no código de produção.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] O arquivo lista os 8 papéis atuais com as permissões de hoje (`PRESIDENTE` 30, `ADMIN_SISTEMA` 8, `CONSELHO_FISCAL` 11, `TESOURARIA` 13, `ESTOQUE_LOJA` 6, `DIRETORIA` 8, `ASSOCIADO`/`EVENTOS` 0)
- [ ] O teste passa contra o código atual sem nenhuma alteração de produção
- [ ] O commit contém só o teste e os dados (prova de que o retrato é anterior às mudanças)
- [ ] Gate check passes

**Tests**: unit
**Gate**: Full
**Commit**: `test(identity): snapshot da matriz antes da classificacao`

---

#### T2: Classe e "exige decisão" em `authz.Definition`

**What**: `type Class` com as quatro classes; campos `Class` e `RequiresDecision` em `Definition`; `Validate` recusa classe desconhecida, `RequiresDecision` fora de `servico` e `CommonRead` fora de `normal`. Classe vazia ainda é aceita **nesta tarefa** (a obrigatoriedade entra em T6, quando todas as definições estiverem classificadas).
**Where**: `api/internal/platform/authz/permission.go`
**Depends on**: None
**Reuses**: `Definition.Validate` atual.
**Requirement**: CLS-01 AC1, AC8; SOD-01 AC9 (parte de validação).
**Impacto**: nenhum: definições atuais sem classe continuam válidas.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Testes de tabela cobrem cada classe válida, classe inválida, `RequiresDecision` em classe não `servico` e `CommonRead` em classe não `normal`
- [ ] Gate check passes

**Tests**: unit
**Gate**: Full
**Commit**: `feat(authz): classe de permissao e marca de decisao`

---

#### T3: Classificar as permissões de `identity` e conceder explicitamente ao `PRESIDENTE`

**What**: declarar a classe das permissões de `FoundationContributions` conforme a tabela aprovada (D-02-2) e as 6 concessões explícitas de transição ao `PRESIDENTE` (`user:create`, `user:update`, `user:reset_password`, `role:assign`, `admin:grant`, `admin:revoke`), cada uma com o comentário estruturado de justificativa.
**Where**: `api/internal/identity/app/roles_matrix.go` (`FoundationContributions`)
**Depends on**: T2
**Reuses**: `Contribution.Grants`.
**Requirement**: CLS-01 AC1, AC4; TRN-01 AC2.
**Impacto**: nenhum: a concessão automática total ainda vigora, então as concessões explícitas são redundantes até T6.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Toda definição de `identity` tem classe, igual à tabela da spec
- [ ] As 6 concessões explícitas existem com justificativa
- [ ] Snapshot de T1 continua verde
- [ ] Gate check passes

**Tests**: unit
**Gate**: Full
**Commit**: `feat(identity): classifica permissoes de identity`

---

#### T4: Classificar as permissões de `financeiro`

**What**: classe de cada definição de `financeiro.Contribution`; 8 concessões explícitas de transição ao `PRESIDENTE` (`conta:create`, `conta:update`, `conta:deactivate`, `lancamento:create`, `lancamento:update`, `lancamento:receive`, `lancamento:pay`, `lancamento:cancel`). **Mais 2 concessões transitórias** (`prestacao_contas:approve`, `parecer:opine`), marcadas "transitória até a D-02-3 (I-6, T29)", para manter o snapshot com zero exceções até lá.
**Where**: `api/internal/financeiro/module.go`
**Depends on**: T2
**Reuses**: constantes `Perm*` de `financeiro/app` (FIN-D-019).
**Requirement**: CLS-01 AC1, AC4; TRN-01 AC1–2.
**Impacto**: `financeiro`: só declarações; nenhum caso de uso muda.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Toda definição de `financeiro` tem classe, igual à tabela da spec
- [ ] 8 concessões de transição + 2 transitórias, com justificativa
- [ ] `financeiro/module_test.go` atualizado para conferir classes e concessões
- [ ] Gate check passes

**Tests**: unit
**Gate**: Full
**Commit**: `feat(financeiro): classifica permissoes do financeiro`

---

#### T5: Classificar as permissões de `estoque`

**What**: classe de cada definição de `estoque.Contribution`; 2 concessões explícitas ao `PRESIDENTE` (`movimentacao:create`, `movimentacao:adjust`).
**Where**: `api/internal/estoque/module.go`
**Depends on**: T2
**Reuses**: `estoque/app` `Perm*`.
**Requirement**: CLS-01 AC1, AC4; TRN-01 AC2.
**Impacto**: `estoque`: só declarações.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Toda definição de `estoque` tem classe; 2 concessões explícitas com justificativa
- [ ] `estoque/module_test.go` atualizado
- [ ] Gate check passes

**Tests**: unit
**Gate**: Full
**Commit**: `feat(estoque): classifica permissoes do estoque`

---

#### T6: `BuildMatrix` fail-closed e teste de snapshot

**What**: `Validate` passa a exigir classe; `BuildMatrix` aplica as regras 1–3 de C2 (classe, `servico` só a papel de serviço e vice-versa, `sensivel`/`servico` nunca a `ASSOCIADO`) e troca a concessão automática do `PRESIDENTE` de "todo o catálogo" (`roles_matrix.go:85-87`) para "todo `normal` e `propria`". Teste de snapshot compara a matriz com o retrato de T1 menos as exceções aprovadas (zero nesta fase). Ajusta os testes atuais que montam definições sem classe.
**Where**: `api/internal/identity/app/roles_matrix.go`, `roles_matrix_test.go`, `matrix_snapshot_test.go`; testes afetados em `identity/infra/role_repository_test.go:416`, `financeiro/module_test.go:15`
**Depends on**: T1, T3, T4, T5
**Reuses**: `BuildMatrix`, `belongsTo`.
**Requirement**: CLS-01 AC2–7; TRN-01 AC1, AC3, AC4; T-01, T-02, T-03, T-04.
**Impacto**: `identity`, `financeiro`, `estoque`, `httpapi`: nenhuma permissão efetiva muda (provado pelo snapshot); testes que dependiam da concessão total são ajustados sem mudar o que verificam.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Cada violação (sem classe, `servico` a humano, não `servico` a papel de serviço, `sensivel` a `ASSOCIADO`) faz `BuildMatrix` falhar com módulo e permissão na mensagem
- [ ] Snapshot verde com `excecoes_aprovadas.json` vazio: matriz efetiva idêntica papel a papel
- [ ] Permissão `sensivel` nova de teste começa sem concessão (TRN-01 AC4)
- [ ] Suítes de `identity`, `financeiro`, `estoque` e `httpapi` verdes
- [ ] Gate check passes

**Tests**: unit
**Gate**: Full
**Commit**: `feat(identity): matriz fail-closed por classe de permissao`

---

### Phase 2: Lista explícita do Conselho Fiscal (I-2)

#### T7: Lista explícita das 11 permissões do Conselho Fiscal

**What**: `conselhoFiscalAllowlist` versionada com as 11 permissões da CFS-01; `BuildMatrix` recusa concessão ao `CONSELHO_FISCAL` fora da lista, de classe `servico` ou com ação inválida; `forbiddenToConselhoFiscal` vira segunda barreira. Teste de snapshot próprio do Conselho Fiscal.
**Where**: `api/internal/identity/app/conselho_fiscal.go` (novo), `roles_matrix.go`
**Depends on**: T6
**Reuses**: `forbiddenToConselhoFiscal` (`roles_matrix.go:28`).
**Requirement**: CFS-01 AC1–6; T-18, T-19.
**Impacto**: nenhum: as 11 permissões atuais continuam.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Concessão de 12ª permissão (leitura nova, de teste) ao `CONSELHO_FISCAL` faz a partida falhar
- [ ] Snapshot do Conselho Fiscal com exatamente as 11
- [ ] Gate check passes

**Tests**: unit
**Gate**: Full
**Commit**: `feat(identity): conselho fiscal por lista explicita de permissoes`

---

### Phase 3: Migration 000010, versão do RBAC, integridade e contas de serviço (I-3)

#### T8: Migration `000010_identity_governanca` — estrutura

**What**: criar `000010` (up e down) com: `users.kind`, `roles.kind`, `permissions.class`, `permissions.requires_decision`, `service_accounts`, `service_account_events`, `service_account_contingency_reviews`, `institutional_acts` (com `CHECK` de `body`), `rbac_sync_state` (linha `version = 1`; `tj_app` sem `INSERT`/`DELETE`), `rbac_matrix_versions`, `rbac_version_rollbacks`, grants e o bloco `DO $$ ... tj_app ... $$`. `down` recusa (RAISE) se houver conta `SERVICE`, evento, revisão ou ato institucional. Gatilhos e o `CHECK` de integridade ficam para T9 e T11.
**Where**: `api/migrations/000010_identity_governanca.up.sql`, `.down.sql`, `api/internal/platform/database/identity_governanca_migration_test.go`
**Depends on**: T6
**Reuses**: padrão dos testes `*_migration_test.go`; `platform/testutil/postgres.go`.
**Requirement**: TRN-01 AC6; SVC-01 AC1 (estrutura); design `02` "Data Models".
**Impacto**: banco: só colunas com `DEFAULT` e tabelas novas; linhas atuais ficam `HUMAN`/`normal`/`false`. O binário atual continua funcionando com a migration aplicada.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Teste com PostgreSQL real: up sobre banco com dados de `000001`–`000009`; todas as contas `HUMAN`; grants de `tj_app` conferidos; `rbac_sync_state` com uma linha `version = 1`
- [ ] `down` limpo sem dados; `down` recusado com conta `SERVICE`, evento, revisão ou ato
- [ ] Guarda de schema (`schema_guard_test.go`) verde
- [ ] Nenhuma migration executada fora de testcontainers
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(identity): migration de governanca de identidade`

---

#### T9: Integridade das contas de serviço no banco (ajuste obrigatório 3)

**What**: acrescentar à `000010` o `CHECK users_service_no_password` e os gatilhos `users_kind_immutable`, `users_service_has_account` (de restrição, adiado), `service_accounts_require_service_user`, `roles_kind_immutable`, `sessions_require_human_user`, `password_reset_tokens_require_human`, `user_roles_kind_consistency`, `admin_memberships_require_human`, todos com `ERRCODE = 'check_violation'` e mensagem que nomeia a regra. Teste dedicado com SQL direto **como `tj_app`**.
**Where**: `api/migrations/000010_identity_governanca.up.sql`, `.down.sql`, `api/internal/platform/database/identity_integrity_migration_test.go`
**Depends on**: T8
**Reuses**: estrutura de T8.
**Requirement**: SVC-01 AC8, AC23, AC24; T-07, T-29.
**Impacto**: banco: nenhuma linha atual viola as regras (só há contas e papéis humanos); `UPDATE` que regrava o mesmo `kind` (ex.: `Save` do GORM) continua aceito.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Recusados, cada um com a mensagem da regra: `UPDATE users SET kind` (HUMAN→SERVICE e SERVICE→HUMAN); `UPDATE roles SET kind`; usuário `SERVICE` sem `service_accounts` no fim da transação; usuário `SERVICE` com senha; `INSERT` e `UPDATE OF user_id` em `sessions` e `password_reset_tokens` para conta de serviço; `INSERT` e `UPDATE` em `user_roles` com papel `SERVICE` a pessoa e papel `HUMAN` a conta de serviço; `INSERT` e `UPDATE` em `admin_memberships` para conta de serviço; `service_accounts` apontando para `HUMAN` e troca de `user_id`/`name`
- [ ] Aceitos: o caminho correto de cada regra e o `UPDATE` que regrava o mesmo `kind`
- [ ] `down` remove os gatilhos e o `CHECK`
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(identity): restricoes de integridade das contas de servico`

---

#### T10: Lista única de contribuições, hash e versão da matriz

**What**: pacote `composition` com `Contributions()` (a lista que hoje está repetida em `cmd/api` e `cmd/bootstrap-admin`); `Matrix.Hash()` canônico; `const MatrixVersion = 2`; teste que compara `{versão, hash}` com `testdata/rbac_matrix_version.json` e falha com "incremente `MatrixVersion`" quando o hash muda. `cmd/api` e `cmd/bootstrap-admin` passam a usar `composition.Contributions()` (refatoração sem mudança de comportamento).
**Where**: `api/internal/composition/rbac.go`, `rbac_version_test.go`, `testdata/rbac_matrix_version.json`; `api/internal/identity/app/matrix_version.go`; `api/cmd/api/main.go`, `api/cmd/bootstrap-admin/main.go`
**Depends on**: T7
**Reuses**: lista de contribuições de `cmd/api/main.go:50`.
**Requirement**: TRN-01 AC9; design `02` C10.
**Impacto**: `cmd/*`: a matriz montada é a mesma; o teste arquitetural precisa aceitar `composition` como pacote de composição (não é módulo: não tem `domain`).

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Hash estável entre execuções e independente da ordem de declaração
- [ ] Alterar uma concessão de teste sem subir a versão faz o teste falhar com a mensagem prevista
- [ ] `cmd/api` e `cmd/bootstrap-admin` usam a mesma função; teste arquitetural verde
- [ ] Gate check passes

**Tests**: unit
**Gate**: Full + Architecture
**Commit**: `feat(identity): versao e hash da matriz de rbac`

---

#### T11: Sincronização condicionada à versão e barreira no banco (ajuste obrigatório 1)

**What**: gatilhos `rbac_matrix_version_guard` (em `permissions`, `roles` e `role_permissions`, para `INSERT`, `UPDATE` e `DELETE`) e `rbac_sync_state_monotonic` na `000010`; gravação de `rbac_matrix_versions` a cada versão nova (a versão `1` vem do snapshot de T1, embutido); `RBACVersionGate` em `RoleRepository.Sync` (dentro da transação e do advisory lock: `SET LOCAL tj.rbac_matrix_version`, leitura `FOR UPDATE` de `rbac_sync_state`, regras de versão maior/igual/menor, gravação de versão, hash, versão do binário); erros `rbac_matrix_conflict` e `rbac_binary_obsolete` impedem a partida.
**Where**: `api/internal/identity/infra/rbac_version_gate.go` (novo), `role_repository.go`; `api/migrations/000010_identity_governanca.up.sql`; `api/internal/identity/infra/rbac_version_gate_test.go`
**Depends on**: T8, T10
**Reuses**: `RoleRepository.Sync`, `syncLockKey`, `database.WithTx`.
**Requirement**: TRN-01 AC10–11; T-28; design `02` C10.
**Impacto**: `identity/infra`: `Sync` ganha a verificação; binários antigos passam a ser recusados pelo banco depois que um binário novo grava a versão 2.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Binário antigo simulado (sincronização sem a variável de sessão) logo depois da migration → sincroniza
- [ ] Binário novo e depois o antigo → gatilho recusa; a matriz nova fica intacta, sem escrita parcial
- [ ] Mesma versão e hash diferente → `rbac_matrix_conflict`, nada escrito
- [ ] Versão menor → `rbac_binary_obsolete`, nada escrito
- [ ] Duas sincronizações novas simultâneas → uma aplica, a outra encontra diff vazio
- [ ] Verificação de leitura para o worker (`VerifyVersion`): espera até o prazo quando a versão gravada é menor; recusa quando é maior ou com hash diferente
- [ ] Gatilho cobre as três tabelas nas três operações (um caso por combinação); `UPDATE` que reduz a versão fora do rollback → recusado; `INSERT`/`DELETE` em `rbac_sync_state` → sem permissão
- [ ] API nova e instância antiga em execução simultânea: a antiga continua atendendo com as permissões lidas do banco; só a reinicialização dela falha
- [ ] Testes existentes que sincronizam matrizes diferentes (`role_repository_test.go`, testes de `bootstrap-admin`) passam versões crescentes explícitas; nenhum fixture escreve RBAC por SQL direto sem a variável de versão
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(identity): sincronizacao de rbac condicionada a versao da matriz`

---

#### T12: Comando auditado `rbac-rollback`

**What**: subcomando `bootstrap-admin rbac-rollback --to-version N --reason "..." --confirm <n>`: calcula, a partir de `rbac_matrix_versions`, as concessões que a reversão restauraria e retiraria; imprime a lista; exige `--confirm` igual ao número de concessões restauradas e motivo com ≥ 30 caracteres; grava `rbac_version_rollbacks` e `rbac.version_rollback` (ação de sincronização do RBAC, ator `system`) com a lista completa, e só então reduz a versão.
**Where**: `api/cmd/bootstrap-admin/rbac_rollback.go` (novo), `api/cmd/bootstrap-admin/rbac_rollback_test.go`
**Depends on**: T11
**Reuses**: `audit.Recorder`, composição atual do `bootstrap-admin`.
**Requirement**: TRN-01 AC12; AUT-01 AC5.
**Impacto**: `cmd/bootstrap-admin`: novo subcomando; o comportamento atual não muda.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Sem motivo ou com motivo curto → recusa, nada gravado
- [ ] `--confirm` ausente ou diferente da contagem → nada gravado
- [ ] Com motivo e confirmação → lista de concessões restauradas e retiradas na auditoria e em `rbac_version_rollbacks`, versão reduzida; o binário antigo simulado volta a sincronizar e grava `rbac.sync` com o diff
- [ ] Reversão simulada de I-6 lista exatamente `PRESIDENTE +prestacao_contas:approve`, `+parecer:opine` e `−conselho_fiscal:manage`
- [ ] `--to-version` maior ou igual à atual → recusa
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(identity): comando auditado de reversao da versao do rbac`

---

#### T13: Tipo de conta, `FISCAL` e papéis de serviço no domínio

**What**: `domain.User.Kind`; `RoleKind` e `Role.Kind()`; `RoleFiscal` e `SERVICO_JOBS_MANUTENCAO` em `role.go`; `CheckRolesHaveMembership` ignora papéis `SERVICE` e os recusa para pessoa; `user_repository` lê e grava `kind`; consulta de elegibilidade `FISCAL` publicada em `identity/app` (FIS-01 AC3–4). `FISCAL` sem permissões.
**Where**: `api/internal/identity/domain/role.go`, `user.go`, `admin_membership.go`; `api/internal/identity/infra/user_repository.go`; `api/internal/identity/app/fiscal_eligibility.go` (novo)
**Depends on**: T8
**Reuses**: guarda "só `role.go` soletra papéis" (`authz/rolename_guard_test.go`).
**Requirement**: FIS-01 AC1–6; SVC-01 AC1, AC10 (parte de domínio); TRN-01 AC6.
**Impacto**: `identity`: novo papel aparece na administração sem permissões; `MatrixVersion` sobe (papéis novos).

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Promover a `FISCAL`, consultar elegibilidade, revogar e ver a consulta mudar (integração)
- [ ] Papel de serviço recusado para pessoa no domínio
- [ ] Guarda de nomes de papel verde; snapshot e versão atualizados
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(identity): tipo de conta, papel fiscal e papeis de servico`

---

#### T14: Login e recuperação recusam conta de serviço

**What**: `Authenticator.check` trata `Kind != HUMAN` como usuário inativo (mesma `ErrInvalidCredentials`, mesmo `BurnCycles`); `RequestPasswordReset` responde como para qualquer endereço e não cria token.
**Where**: `api/internal/identity/app/authenticate.go`, `request_password_reset.go`
**Depends on**: T13
**Reuses**: resposta uniforme atual do login.
**Requirement**: SVC-01 AC6–7; T-06.
**Impacto**: `identity`: nenhum efeito para contas humanas.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Login com e-mail de conta de serviço → mesma resposta e mesmo trabalho que senha errada; nenhuma sessão
- [ ] Recuperação → resposta igual; nenhum token
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(identity): login e recuperacao recusam conta de servico`

---

#### T15: Sincronização de classes, tipos de papel e contas de serviço

**What**: `Contribution.ServiceAccounts`; `BuildMatrix` valida nome `svc.<modulo>-<processo>` com prefixo do módulo e "um papel de serviço por conta" (regra 5); `RoleRepository.Sync` grava `class`, `requires_decision` e `roles.kind` (auditoria por classe alterada) e chama `ServiceAccountRepository.Sync` na mesma transação e sob o mesmo lock (cria, atribui o papel, desativa as removidas com origem `SYNC`, não reativa desativada por pessoa; auditoria `service_account.sync`).
**Where**: `api/internal/identity/infra/service_account_repository.go` (novo), `role_repository.go`; `api/internal/identity/app/roles_matrix.go`; `api/internal/identity/domain/service_account.go` (novo)
**Depends on**: T11, T13
**Reuses**: `computeDiff`, `apply`, `audit.Recorder`.
**Requirement**: CLS-01 AC9; SVC-01 AC1–5, AC10; T-08.
**Impacto**: `identity/infra`: `rbac.sync` passa a incluir classes e contas; primeira sincronização grava as classes das 30 permissões.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Conta declarada é criada como `SERVICE`, com `'!servico'` e o papel; segunda sincronização não muda nada
- [ ] Conta removida do código é desativada (origem `SYNC`), nunca excluída; conta desativada por pessoa não é reativada
- [ ] Prefixo errado e papel compartilhado fazem a partida falhar
- [ ] Mudança de classe grava auditoria por permissão
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(identity): sincronizacao de classes e contas de servico`

---

#### T16: Casos de uso atuais recusam conta de serviço

**What**: `ErrServiceAccount` → 403 `forbidden`, auditado, em `PromoteToAdmin`, `AssignRoles`, `RevokeAdmin`, `AdminResetPassword`, `UserActivation`; `ListUsers` filtra `kind = 'HUMAN'`.
**Where**: `api/internal/identity/app/promote_admin.go`, `assign_roles.go`, `revoke_admin.go`, `admin_reset_password.go`, `deactivate_user.go`, `list_users.go`
**Depends on**: T13
**Reuses**: `denyRoleChange`, auditoria `role.change_denied`/`authz.denied`.
**Requirement**: SVC-01 AC9, AC14.
**Impacto**: `identity`: nenhum efeito para contas humanas; mesma mudança mecânica em seis casos de uso.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Cada caso de uso recusa conta de serviço com 403 e auditoria
- [ ] `ListUsers` não devolve contas de serviço
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(identity): casos de uso recusam conta de servico`

---

#### T17: Separação de funções por identidade (`SeparateFrom`)

**What**: `ErrSeparationOfDuties`, `ErrSelfAct` e `SeparateFrom(actorID, priorAuthors...)` comparando identidade de usuário, nunca papel. Antecipado de I-5 porque a reativação (T19) usa.
**Where**: `api/internal/platform/authz/sod.go` (novo)
**Depends on**: None
**Reuses**: pacote `authz`.
**Requirement**: SOD-01 AC1–3; INS-01 AC10; T-10, T-27.
**Impacto**: nenhum (primitiva nova).

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Ator entre os autores → `separacao_de_funcoes`; ator sobre si → `ato_sobre_si_mesmo`
- [ ] Usuário com dois papéis, sendo autor do ato anterior → recusado pela identidade (T-27)
- [ ] Gate check passes

**Tests**: unit
**Gate**: Full
**Commit**: `feat(authz): separacao de funcoes por identidade`

---

#### T18: Permissões de conta de serviço, desativação, listagem e `ServicePrincipals`

**What**: permissões `identity:service_account:read` (normal), `:deactivate` e `:reactivate` (sensíveis) com concessões explícitas a `ADMIN_SISTEMA` e `PRESIDENTE` (D-02-1); `DeactivateServiceAccount` (motivo ≥ 10, evento `MANUAL`, auditoria); `ListServiceAccounts` sem campo de credencial; `ServicePrincipals.For` (lê `is_active` a cada chamada); `Principal` de conta inativa recusado nos casos de uso.
**Where**: `api/internal/identity/app/service_accounts.go`, `service_principals.go` (novos), `roles_matrix.go`
**Depends on**: T15, T16
**Reuses**: `Matrix.PermissionsOf`, `RoleRepository`.
**Requirement**: SVC-01 AC11–13, AC15; T-09; D-02-1.
**Impacto**: matriz: `ADMIN_SISTEMA` 8 → 11 e `PRESIDENTE` +3 nesta tarefa; `MatrixVersion` sobe; snapshot registra as permissões novas na tabela "Depois do G1".

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Desativação com motivo curto → `validation_failed`; com motivo → evento e auditoria
- [ ] `ServicePrincipals.For` de conta desativada devolve `active = false`; caso de uso com esse `Principal` → `forbidden`
- [ ] Listagem sem credenciais; permissões e concessões conferidas no snapshot
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(identity): desativacao e principal de contas de servico`

---

#### T19: Reativação comum, contingência e revisão independente

**What**: `ReactivateServiceAccount` (caminho comum com `SeparateFrom`; contingência só quando, sob lock, nenhuma outra pessoa ativa tem `:reactivate`: espera `IDENTITY_CONTINGENCY_WAIT`, motivo ≥ 30, emergência, sem revisão aberta da mesma pessoa; evento `CONTINGENCY`, revisão aberta, log `warn`, auditoria `contingencia`; recusa conta desativada por `SYNC`); `CloseContingencyReview` só por outra pessoa habilitada.
**Where**: `api/internal/identity/app/service_account_reactivation.go` (novo), `api/internal/identity/infra/service_account_repository.go`
**Depends on**: T17, T18
**Reuses**: padrão de lock de `ensureAdminRemains`.
**Requirement**: SVC-01 AC16–22; T-22, T-23.
**Impacto**: `identity`: casos de uso novos; nada muda nos atuais.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Mesma pessoa desativa e tenta reativar com outra pessoa habilitada → `separacao_de_funcoes`
- [ ] Contingência antes da espera → `contingencia_indisponivel`; depois, com motivo e emergência → aceita, revisão aberta
- [ ] Revisão fechada por quem reativou → recusada; por outra pessoa → fechada
- [ ] Revisão aberta bloqueia nova contingência da mesma pessoa
- [ ] Duas contingências simultâneas para a mesma conta → uma aceita (concorrência)
- [ ] Conta desativada por `SYNC` → recusada
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(identity): reativacao de conta de servico com contingencia`

---

#### T20: Contrato e rotas de contas de serviço

**What**: rotas `GET /api/v1/service-accounts`, `POST .../{name}/deactivate`, `.../{name}/reactivate`, `.../contingency-reviews/{id}/close` em `identity.yaml` (contract-first) e handlers na cadeia autenticada; códigos `separacao_de_funcoes`, `contingencia_indisponivel`.
**Where**: `api/openapi/identity.yaml`, `api/internal/identity/http/service_accounts.go` (novo)
**Depends on**: T19
**Reuses**: handlers de `identity/http`, `httpx.WriteProblem`.
**Requirement**: SVC-01 AC11–22 (exposição HTTP); AD-010/012/013.
**Impacto**: `web`: só tipos gerados mudam; nenhuma tela neste grupo.

**Tools**: MCP: NONE · Skill: security-best-practices

**Done when**:
- [ ] Contrato válido; tipos Go e TS gerados
- [ ] Sem sessão → 401; sem CSRF → 403; sem permissão → 403; problem+json com os códigos novos
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full + Contract
**Commit**: `feat(identity): rotas de contas de servico`

---

### Phase 4: Autoria `service` e rastreabilidade (I-4)

#### T21: Migration `000011_audit_service_actor`

**What**: confirmar no banco o nome real do `CHECK` inline de `actor_type` e criar `000011` (troca das duas restrições, índice por `correlation_id`); `down` recusa se existir `actor_type = 'service'`.
**Where**: `api/migrations/000011_audit_service_actor.up.sql`, `.down.sql`, `api/internal/platform/database/audit_service_actor_migration_test.go`
**Depends on**: T8
**Reuses**: teste de migration de `000002`.
**Requirement**: AUT-01 AC1–2; TRN-01 AC7.
**Impacto**: `audit_log`: linhas antigas intocadas e válidas sob as restrições novas.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Up sobre banco com entradas `user`, `system` e `anonymous` antigas → válido; nenhuma linha alterada
- [ ] `service` sem id recusado pelo banco; `down` recusado com linha `service`
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(audit): migration do ator de servico`

---

#### T22: Ator `service`, contexto padronizado e resolvedor de tipo

**What**: `ActorService`; `Validate` exige id para `user`/`service` e `correlation_id`/`causation_id` para `service`; chaves `CtxCorrelationID`, `CtxCausationID`, `CtxOnBehalfOf`, `CtxDecisionID`, `CtxRule`; `ActorKindResolver` injetado com `WithActorKinds`, implementado por `identity/infra` e ligado em `composition`; guarda de código-fonte: `ActorSystem` só em bootstrap e sincronização do RBAC.
**Where**: `api/internal/platform/audit/entry.go`, `recorder.go`; `api/internal/identity/infra/actor_kinds.go` (novo); `api/internal/platform/audit/actor_system_source_test.go` (novo)
**Depends on**: T21, T13
**Reuses**: `Recorder.Record`, `Entry.Redacted`.
**Requirement**: AUT-01 AC1–8, AC10; T-13, T-14.
**Impacto**: todos os módulos que auditam: nenhuma mudança para `user`; `system` só nos usos atuais (bootstrap, `rbac.sync`) e nos novos de sincronização.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Cada combinação inválida de tipo de ator e conta recusada, desfazendo a transação de negócio
- [ ] `service` sem `correlation_id` ou `causation_id` recusado
- [ ] Guarda de `ActorSystem` falha com um uso de teste fora dos arquivos permitidos
- [ ] Gate check passes

**Tests**: unit + integration
**Gate**: Full
**Commit**: `feat(audit): autoria de conta de servico e contexto de causalidade`

---

#### T23: Filtro da auditoria por `correlation_id`

**What**: `Filter.CorrelationID` em `Query.Search`; parâmetro `correlation_id` em `GET /api/v1/audit-logs` no `platform.yaml` e no handler.
**Where**: `api/internal/platform/audit/query.go`, `http.go`; `api/openapi/platform.yaml`
**Depends on**: T22
**Reuses**: `audit.Query`, índice de T21.
**Requirement**: AUT-01 AC9; T-15.
**Impacto**: `web`: tipos gerados.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Fluxo de teste com três entradas do mesmo `correlation_id` devolvido em ordem de tempo
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full + Contract
**Commit**: `feat(audit): filtro por correlation_id`

---

### Phase 5: Contrato de decisão (I-5)

#### T24: Contrato de decisão e execução por regra

**What**: documentar o contrato C9 em `platform/authz/sod.go` e provar o contrato com um módulo de teste (tabela criada só no teste): decisão `APROVADA` executada uma vez por `UPDATE ... WHERE estado='APROVADA'`, recusa de decisão ausente/revogada/executada/de outro escopo, execução por regra aprovada com `regra` na auditoria e sem `decision_id`, conta de serviço sem poder criar decisão.
**Where**: `api/internal/platform/authz/sod_contract_test.go` (novo, só teste)
**Depends on**: T17, T22
**Reuses**: `SeparateFrom`, `audit.Recorder`.
**Requirement**: SOD-01 AC4–8, AC10–11; T-11, T-12, T-20, T-21.
**Impacto**: nenhum em produção.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Duas execuções simultâneas da mesma decisão → exatamente uma
- [ ] Execução por regra aprovada auditada com `regra@vN`
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `test(authz): contrato de decisao e execucao por regra`

---

### Phase 6: D-02-3 + D-02-4 + D-02-5 juntas (I-6)

#### T25: Papéis incompatíveis, permissões institucionais e `CoversExcluding`

**What**: `domain.IncompatiblePairs` (`PRESIDENTE` × `CONSELHO_FISCAL`) e `Conflicts`; `InstitutionalPermissions`; `authz.CoversExcluding`; definição de `identity:conselho_fiscal:manage` (sensível, **sem concessão nesta tarefa**); guarda de código-fonte: `CoversExcluding` só em `conselho_fiscal_acts.go` e `deactivate_user.go`, sempre com `InstitutionalPermissions`.
**Where**: `api/internal/identity/domain/role_conflicts.go`, `api/internal/platform/authz/escalation.go`, `api/internal/identity/app/conselho_fiscal.go`, `api/internal/identity/app/covers_excluding_source_test.go`
**Depends on**: T7
**Reuses**: `authz.Covers` (inalterado).
**Requirement**: INS-01 AC1, AC7, AC10–11.
**Impacto**: matriz: permissão nova sem concessão (`MatrixVersion` sobe); nenhum acesso muda.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] `Conflicts` detecta o par em qualquer ordem e com papéis extras
- [ ] `CoversExcluding` ignora só a lista recebida; `Covers` inalterado (testes atuais verdes)
- [ ] Guarda falha com uso de teste em outro arquivo ou com lista literal
- [ ] Gate check passes

**Tests**: unit
**Gate**: Full
**Commit**: `feat(identity): papeis incompativeis e permissoes institucionais`

---

#### T26: Conflito de papéis no `Principal`, na atribuição e na sincronização

**What**: `SessionService.Validate` remove as `InstitutionalPermissions` do `Principal` quando `Conflicts` não é vazio (a cada requisição); `AssignRoles` e `PromoteToAdmin` recusam papéis resultantes incompatíveis com `papeis_incompativeis` depois de reler os papéis na transação; `Sync` grava `rbac.role_conflict` com os ids; `ListUsers` marca `irregularidade: conflito_de_papeis`.
**Where**: `api/internal/identity/app/session_service.go`, `assign_roles.go`, `promote_admin.go`, `list_users.go`; `api/internal/identity/infra/role_repository.go`
**Depends on**: T25
**Reuses**: leitura sem cache de `session_service.go:66-69`.
**Requirement**: INS-01 AC8–9; T-26.
**Impacto**: usuários com os dois papéis (se existirem) perdem só os atos institucionais na requisição seguinte; os demais acessos continuam.

**Tools**: MCP: NONE · Skill: security-best-practices

**Done when**:
- [ ] Usuário existente com os dois papéis → atos institucionais negados e auditados; outros acessos mantidos; aparece no relatório da sincronização e na listagem
- [ ] Atribuição que geraria o par → `papeis_incompativeis`; nenhum registro histórico alterado
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(identity): bloqueio de atos por papeis incompativeis`

---

#### T27: Registro institucional de nomeação e destituição do Conselho Fiscal

**What**: `institutional_act_repository`; `domain.CompetentBodies`; `RegisterCFAppointment` e `RegisterCFRemoval` (regras 1–6 de C7: `Require(manage)`, dados do ato e documento `identity.ato_institucional` obrigatórios, `SeparateFrom`, ator sem `CONSELHO_FISCAL`, alvo sem `PRESIDENTE`, `CoversExcluding`, transação única com ato, vínculo, papéis, revogação de sessões e auditoria); `AssignRoles`, `PromoteToAdmin` e `RevokeAdmin` passam a recusar `CONSELHO_FISCAL` com `ato_institucional_obrigatorio`; reescrita de `promote_admin_test.go:170`.
**Where**: `api/internal/identity/app/conselho_fiscal_acts.go`, `api/internal/identity/infra/institutional_act_repository.go`, `api/internal/identity/domain/institutional_act.go` (novos); `assign_roles.go`, `promote_admin.go`, `revoke_admin.go`
**Depends on**: T25, T26
**Reuses**: `PromoteToAdmin` (vínculo e revogação de sessões), `documents.Store`, `denyRoleChange`.
**Requirement**: INS-01 AC1–4, AC12; T-25; D-02-4 limites 1–5; D-02-5 (registro, não decisão).
**Impacto**: atribuir `CONSELHO_FISCAL` passa a exigir o caso de uso institucional; o ato é registrado como ato externo, nunca como decisão de quem registra.

**Tools**: MCP: NONE · Skill: security-best-practices

**Done when**:
- [ ] Caminho comum com `CONSELHO_FISCAL` → `ato_institucional_obrigatorio`
- [ ] Sem `manage` → `forbidden`; sem documento, referência, data ou instância → `validation_failed`
- [ ] Instância fora de `CompetentBodies` (ex.: "Presidência") e data futura → `validation_failed`; o `CHECK` de `institutional_acts.body` recusa a mesma linha por SQL direto
- [ ] Auditoria e consulta mostram `registered_by` como registrador e a instância como autora da decisão
- [ ] Alvo = ator → `separacao_de_funcoes`; ator conselheiro → recusado; alvo `PRESIDENTE` → `papeis_incompativeis`
- [ ] Com tudo válido → ato, papel, vínculo, sessões revogadas e auditoria na mesma transação; falha da auditoria desfaz tudo
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(identity): registro institucional do conselho fiscal`

---

#### T28: Bloqueio técnico de conselheiro e recusa da redefinição de senha

**What**: `UserActivation.change` usa `CoversExcluding(..., InstitutionalPermissions)` só quando o alvo tem `CONSELHO_FISCAL`; `AdminResetPassword` mantém `Covers` integral (sem exceção).
**Where**: `api/internal/identity/app/deactivate_user.go`, `admin_reset_password_test.go`
**Depends on**: T25
**Reuses**: `UserActivation`, `AdminResetPassword`.
**Requirement**: INS-01 AC5–6.
**Impacto**: `PRESIDENTE` consegue bloquear e desbloquear conselheiro; não consegue redefinir a senha dele (recuperação só por e-mail).

**Tools**: MCP: NONE · Skill: security-best-practices

**Done when**:
- [ ] Desativação técnica de conselheiro por quem tem `identity:user:update` → aceita e auditada
- [ ] `AdminResetPassword` de conselheiro → `forbidden`
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(identity): bloqueio tecnico de conselheiro sem redefinicao de senha`

---

#### T29: Matriz final do G1 de identity (D-02-3 + D-02-5)

**What**: retirar as 2 concessões transitórias de T4 (`prestacao_contas:approve`, `parecer:opine`) do `PRESIDENTE`; conceder `identity:conselho_fiscal:manage` ao `PRESIDENTE` com a justificativa da D-02-5 ("registro de atos da instância competente; não decide a composição"); `excecoes_aprovadas.json` com as duas retiradas (D-02-3, 2026-10-14); `MatrixVersion` sobe; snapshot final.
**Where**: `api/internal/financeiro/module.go`, `api/internal/identity/app/roles_matrix.go`, `api/internal/identity/app/testdata/excecoes_aprovadas.json`
**Depends on**: T27, T28
**Reuses**: teste de snapshot de T6.
**Requirement**: TRN-01 AC8; INS-01; T-24; D-02-3, D-02-5.
**Impacto**: `PRESIDENTE` perde 2 permissões que **nenhum caso de uso verifica hoje** (FIN-D-014) e ganha `manage`; `jobs:job:close` continua sem concessão.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Snapshot: diferença nas 30 existentes igual exatamente às 2 exceções aprovadas
- [ ] Contagens das 30 + novas de identity: `PRESIDENTE` 33 quando somada a `jobs:job:read` do design `01` (sem ela, 32); `ADMIN_SISTEMA` 11 (13 com as de `jobs`); `CONSELHO_FISCAL` 11
- [ ] Nomeação ponta a ponta pelo `PRESIDENTE` com ato da Assembleia Geral e documento passa; sem documento, com instância não competente ou com o próprio `PRESIDENTE` como alvo, falha
- [ ] Nenhum outro caso de uso verifica `identity:conselho_fiscal:manage` (guarda de código-fonte: a constante só aparece em `conselho_fiscal_acts.go`, na contribuição e nos testes)
- [ ] Comentário da concessão com a justificativa e o limite da D-02-5; `jobs:job:close` sem papel
- [ ] Gate check passes

**Tests**: unit + integration
**Gate**: Full
**Commit**: `feat(identity): aplica d-02-3 e concede registro institucional ao presidente`

---

#### T30: Contrato e rotas do registro institucional

**What**: `POST /api/v1/conselho-fiscal/nomeacoes` e `/destituicoes` em `identity.yaml` e handlers; códigos `ato_institucional_obrigatorio`, `papeis_incompativeis`, `ato_sobre_si_mesmo`.
**Where**: `api/openapi/identity.yaml`, `api/internal/identity/http/conselho_fiscal.go` (novo)
**Depends on**: T29
**Reuses**: upload de documento existente (`owner_type = identity.ato_institucional`, formato aceito por `documents/list.go:22`).
**Requirement**: INS-01 (exposição HTTP); AD-010/012/013.
**Impacto**: `web`: só tipos gerados; nenhuma tela neste grupo.

**Tools**: MCP: NONE · Skill: security-best-practices

**Done when**:
- [ ] Contrato válido; tipos gerados
- [ ] Sessão, CSRF e permissão verificados; problem+json com os códigos novos
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full + Contract
**Commit**: `feat(identity): rotas do registro institucional do conselho fiscal`

---

### Phase 7: Verificação final e preparação da implantação (sem implantar)

#### T31: Regressão completa e revisão de segurança

**What**: rodar todas as suítes (`api` e `web`), o e2e existente e a revisão `security-best-practices` sobre autenticação, sessões, papéis, auditoria e as restrições de banco; registrar achados em `STATE.md` e corrigir o que for consequência técnica direta.
**Where**: `.specs/features/tj-connect/STATE.md`
**Depends on**: T9, T12, T20, T23, T24, T30
**Reuses**: comandos do `CLAUDE.md`.
**Requirement**: TRN-01 (regressão); Success Criteria da spec `02`.
**Impacto**: nenhum código novo, salvo correções diretas.

**Tools**: MCP: NONE · Skill: security-best-practices

**Done when**:
- [ ] `go vet`, `go test`, `go test -tags=integration`, `go build` verdes; `pnpm lint`, `tsc`, `pnpm test`, `pnpm build` verdes
- [ ] Revisão de segurança sem achado aberto de severidade alta
- [ ] Os cinco Success Criteria da spec `02` conferidos um a um
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full + Contract
**Commit**: `chore(identity): verificacao final do g1 de identity`

---

#### T32: Runbook de implantação da parte de identity (sem implantar)

**What**: documento de operação com: pré-requisitos (provedor de e-mail real; existência de pessoa com `PRESIDENTE` ativa antes de I-6; relatório de usuários com papéis em conflito); ordem (migrations `000010`/`000011` → API nova → worker → retirada das antigas); consultas **somente leitura** de verificação antes e depois (`rbac_sync_state`, contagens de concessões por papel, contas `SERVICE`); reversão com `rbac-rollback` e limites dos `down`; sinais de alerta. Não executa nada.
**Where**: `docs/operations/g1-identity-implantacao.md` (novo)
**Depends on**: T31
**Reuses**: design `02` C10, "Rollback".
**Requirement**: TRN-01 AC10–12; D-02-5 (condição de I-6).
**Impacto**: nenhum.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Cada passo tem verificação objetiva e critério de interrupção
- [ ] Nenhum segredo, dado pessoal ou dado financeiro interno no documento
- [ ] Revisado pelo mantenedor (a implantação continua exigindo autorização)
- [ ] Gate check passes

**Tests**: none
**Gate**: Docs
**Commit**: `docs(identity): runbook de implantacao do g1`

---

## Avisos do validador e justificativas

`validate_tasks.py` (2026-10-14): 0 erros, 25 avisos. Nenhum foi eliminado artificialmente. Cada um está justificado abaixo; nenhuma tarefa ficou com mais de um objetivo verificável.

| Categoria | Tarefas | Justificativa concreta |
| --- | --- | --- |
| `Where` cita o código e o seu próprio teste ou dados de teste | T1, T12, T28 | O teste é a evidência da tarefa (tests-first). T1 é só teste e dados; T12 e T28 têm um arquivo de produção cada |
| `Where` cita `up`, `down` e o teste de migration | T8, T9, T21 | Padrão do repositório (`000001`–`000009`): uma migration é o par `up`/`down` e o seu teste em `platform/database`. T9 é separada de T8 para que as restrições de integridade (ajuste 3) tenham critérios próprios |
| Mesma mudança mecânica em vários arquivos | T14 (2), T16 (6) | Uma única regra ("conta de serviço é recusada") aplicada a cada caso de uso que a exige. Dividir criaria commits intermediários em que parte dos caminhos aceita conta de serviço. Cada arquivo tem o seu caso de teste no critério de conclusão |
| Contrato e handler juntos (contract-first) | T20, T23, T30 | AD-010/012: o contrato e o handler gerado/implementado mudam no mesmo commit, senão o gate `Contract` falha ou fica uma rota sem implementação |
| Dois lados de uma mesma regra | T11 (gatilho + `Sync`), T25 (`Conflicts` + `CoversExcluding` + guarda), T26 (`Principal` + atribuição + sincronização), T29 (retirada + concessão + exceções) | Separar deixaria commits em que uma metade da regra existe sem a outra: em T11, um gatilho sem quem grave a versão; em T26, o `Principal` bloqueando sem a atribuição recusar; em T29, a D-02-3 sem a D-02-5 (ninguém conseguiria registrar nomeações) |
| Componente novo com domínio, repositório e caso de uso | T13, T15, T18, T19, T22, T27 | Cada uma entrega um caso de uso coeso de ponta a ponta. T27 inclui a recusa do caminho comum (`AssignRoles`, `PromoteToAdmin`, `RevokeAdmin`) junto com o caminho institucional, para que nenhum commit fique sem forma de nomear conselheiro |
| Refatoração mecânica de composição | T6 (testes que quebram pela mesma mudança), T7, T10 | T6 ajusta só os testes que a própria mudança quebra; T7 e T10 tocam um componente novo e o ponto em que ele é ligado |
| `Tests: none` | T32 | Documento de operação; a matriz de cobertura define "none" para documentação. A verificação é a revisão do mantenedor |

## Compatibilidade entre código e banco ao longo da sequência

- **Nenhum commit lê coluna ou tabela que a sua própria migration ainda não criou**: `kind` só é lido a partir de T13 (depois de T8); `class`, `requires_decision` e contas de serviço só são gravados a partir de T15 (depois de T8 e T11); `rbac_sync_state` só é usado a partir de T11; o ator `service` só é gravado a partir de T22 (depois de T21).
- **Fases 1 e 2 não têm migration** e não mudam a matriz efetiva (snapshot com zero exceções), então convivem com o banco atual e com instâncias antigas.
- **A matriz efetiva só muda depois da barreira de versão** (T11): T13, T15, T18, T25 e T29 sobem `MatrixVersion`.
- **T9 e T11 completam a `000010` criada em T8.** Isso só é aceitável porque a migration ainda não chegou à `develop`; quem tiver aplicado uma versão anterior localmente recria o banco (`docker compose down -v`). Os testes usam banco novo a cada execução.
- **Nenhum commit intermediário é implantado**: a implantação, quando autorizada, é do G1 completo, na ordem do runbook (T32 e tasks `01` T27).

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1 | snapshot + dados de teste | ✅ Granular |
| T2 | 1 arquivo | ✅ Granular |
| T3, T4, T5 | 1 arquivo por módulo | ✅ Granular |
| T6 | `BuildMatrix` + ajuste dos testes que quebram pela mesma mudança | ✅ Aceitável (testes quebrados são consequência mecânica) |
| T7 | 1 componente | ✅ Granular |
| T8, T9, T21 | migration + teste de migration | ✅ Aceitável (padrão do repositório) |
| T10 | pacote novo + troca mecânica nos dois `cmd` | ✅ Aceitável |
| T11 | barreira (gatilho + verificação no `Sync`) — os dois lados de uma mesma regra | ✅ Aceitável |
| T13 | domínio + repositório do mesmo conceito (`kind`) | ✅ Aceitável |
| T16 | mesma recusa mecânica em seis casos de uso | ✅ Aceitável |
| T18, T19, T27 | um caso de uso coeso com repositório | ✅ Aceitável |
| T20, T23, T30 | contrato + handler (contract-first) | ✅ Aceitável |
| T26 | uma regra aplicada nos quatro pontos em que o `Principal` e os papéis são montados | ✅ Aceitável |
| Demais | 1 componente | ✅ Granular |

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | — | ✅ Match |
| T2 | None | — | ✅ Match |
| T3, T4, T5 | T2 | T2 | ✅ Match |
| T6 | T1, T3, T4, T5 | T1, T3, T4, T5 | ✅ Match |
| T7 | T6 (fase anterior) | — | ✅ Match |
| T8 | T6 (fase anterior) | — | ✅ Match |
| T9 | T8 | T8 | ✅ Match |
| T10 | T7 (fase anterior) | — | ✅ Match |
| T11 | T8, T10 | T8, T10 | ✅ Match |
| T12 | T11 | T11 | ✅ Match |
| T13 | T8 | T8 | ✅ Match |
| T14 | T13 | T13 | ✅ Match |
| T15 | T11, T13 | T11, T13 | ✅ Match |
| T16 | T13 | T13 | ✅ Match |
| T17 | None | — | ✅ Match |
| T18 | T15, T16 | T15, T16 | ✅ Match |
| T19 | T17, T18 | T17, T18 | ✅ Match |
| T20 | T19 | T19 | ✅ Match |
| T21 | T8 (fase anterior) | — | ✅ Match |
| T22 | T21, T13 (fase anterior) | T21 | ✅ Match |
| T23 | T22 | T22 | ✅ Match |
| T24 | T17, T22 (fases anteriores) | — | ✅ Match |
| T25 | T7 (fase anterior) | — | ✅ Match |
| T26 | T25 | T25 | ✅ Match |
| T27 | T25, T26 | T25, T26 | ✅ Match |
| T28 | T25 | T25 | ✅ Match |
| T29 | T27, T28 | T27, T28 | ✅ Match |
| T30 | T29 | T29 | ✅ Match |
| T31 | T9, T12, T20, T23, T24, T30 (fases anteriores) | — | ✅ Match |
| T32 | T31 | T31 | ✅ Match |

## Test Co-location Validation

| Task | Code Layer Created/Modified | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1, T6, T7 | `identity/app` (matriz) | unit + snapshot | unit | ✅ OK |
| T2, T17, T25 | `platform/authz`, `identity/domain` | unit | unit | ✅ OK |
| T3, T4, T5 | contribuições | unit | unit | ✅ OK |
| T8, T9, T21 | migrations | integration | integration | ✅ OK |
| T10 | `composition` | unit | unit | ✅ OK |
| T11, T15 | `identity/infra` | integration | integration | ✅ OK |
| T12 | `cmd/bootstrap-admin` | integration | integration | ✅ OK |
| T13, T14, T16, T18, T19, T26, T27, T28 | `identity/app` casos de uso | unit + integration | integration | ✅ OK |
| T20, T23, T30 | handlers + contrato | integration + Contract | integration, Gate Contract | ✅ OK |
| T22 | `platform/audit` | unit + integration | unit + integration | ✅ OK |
| T24 | teste de contrato | integration | integration | ✅ OK |
| T29 | contribuições + snapshot | unit + integration | unit + integration | ✅ OK |
| T31 | verificação | integration | integration | ✅ OK |
| T32 | documentação | none | none | ✅ OK |

## Requirement Coverage

| Requisito | Tarefas |
| --- | --- |
| CLS-01 | T2, T3, T4, T5, T6, T15 |
| TRN-01 | T1, T3–T6, T8, T10, T11, T12, T13, T21, T29, T31, T32 |
| SVC-01 | T8, T9, T13, T14, T15, T16, T18, T19, T20 |
| SOD-01 | T2, T17, T24 |
| AUT-01 | T12, T21, T22, T23 |
| FIS-01 | T13 |
| CFS-01 | T7 |
| INS-01 | T17, T25, T26, T27, T28, T29, T30 |
