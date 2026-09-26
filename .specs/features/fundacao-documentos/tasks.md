# Fundação Documentos Tasks

## Execution Protocol (MANDATORY -- do not skip)

Implement these tasks with the `tlc-spec-driven` skill: **activate it by name and follow its Execute flow and Critical Rules.** Do not search for skill files by filesystem path. The skill is the source of truth for the full flow (per-task cycle, sub-agent delegation, adequacy review, Verifier, discrimination sensor).

**If the skill cannot be activated, STOP and tell the user - do not proceed without it.**

---

**Design**: `.specs/features/fundacao-documentos/design.md`
**Status**: Approved (2026-09-27); execução em lotes pequenos, por fase

**Pré-requisito**: `fundacao-core` concluída (usa `WithTx`, `audit.Recorder`, `authz.Require`, `testutil` e as migrações da core).

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

### Phase 1: Storage e configuração

```
T1 → T2 → T3 → T4
```

### Phase 2: Documentos

```
T5 → T6
```

---

## Task Breakdown

### Phase 1: Storage e configuração (tarefas)

### T1: Adicionar à configuração as variáveis de storage S3 e validar a validade da URL assinada

**What**: Adicionar à configuração as variáveis de storage S3 e validar a validade da URL assinada.  
**Where**: `api/internal/config/config.go`  
**Depends on**: None  
**Reuses**: `config.Load` (da fundacao-core)  
**Requirement**: DOC-03 (ACs 1, 2, 3)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Variáveis de endpoint, região, bucket, chaves e estilo de path lidas e validadas, sem imprimir segredos
- [ ] Validade da URL menor ou igual a zero é erro de configuração
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: unit  
**Gate**: quick

**Commit**: `feat(config): adiciona configuracao de storage s3`

---

### T2: Criar a interface `Storage` e seus erros

**What**: Criar a interface `Storage` e seus erros.  
**Where**: `api/internal/platform/storage/storage.go`  
**Depends on**: T1  
**Reuses**: Nenhum (pacote novo)  
**Requirement**: DOC-01 (ACs 6, 7)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Interface expõe Put, PresignGet e DeleteCreated, sem operação de sobrescrever ou apagar documentos
- [ ] Erros tipados para não encontrado e falha de armazenamento
- [ ] Gate check passes: `cd api && go vet ./... && go test ./...`

**Tests**: none  
**Gate**: quick

**Commit**: `feat(storage): adiciona interface de object storage`

---

### T3: Escolher o emulador S3 local após checar licença e manutenção atuais e criar o helper de testes

**What**: Escolher o emulador S3 local após checar licença e manutenção atuais e criar o helper de testes.  
**Where**: `api/internal/platform/testutil/s3.go`  
**Depends on**: T2  
**Reuses**: `testutil/postgres.go` como modelo  
**Requirement**: DOC-02 (ACs 4, 5)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] Escolha e fonte consultada registradas no commit; se o candidato for inadequado, escolher outro
- [ ] Helper sobe o emulador e cria um bucket privado
- [ ] Teste de fumaça grava e lê um objeto
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 2 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `test(platform): adiciona helper de emulador s3`

---

### T4: Implementar `Storage` com o SDK S3 (endpoint configurável, sem provedor fixo), incluindo URL assinada

**What**: Implementar `Storage` com o SDK S3 (endpoint configurável, sem provedor fixo), incluindo URL assinada.  
**Where**: `api/internal/platform/storage/s3.go`  
**Depends on**: T3  
**Reuses**: `storage.go` e o helper de emulador  
**Requirement**: DOC-02 (ACs 1, 4, 5); DOC-01 (ACs 1)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] SDK e versão conferidos na documentação vigente
- [ ] URL assinada com validade de 1 segundo é recusada após expirar
- [ ] Objeto pedido sem assinatura é recusado
- [ ] Só usa o protocolo S3, sem recursos de um provedor específico
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 4 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(storage): implementa storage s3`

---

### Phase 2: Documentos (tarefas)

### T5: Criar a migração `documents` com `down` e concessões (sem UPDATE e DELETE)

**What**: Criar a migração `documents` com `down` e concessões (sem UPDATE e DELETE).  
**Where**: `api/migrations/000005_documents.up.sql`  
**Depends on**: None  
**Reuses**: Migração 000002 como padrão  
**Requirement**: DOC-01 (ACs 1, 5, 6)

**Tools**:

- MCP: NONE
- Skill: NONE

**Done when**:

- [ ] `supersedes_id` único impede bifurcação
- [ ] Com `tj_app`, UPDATE e DELETE falham
- [ ] Migração `down` desfaz sem erro
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 3 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(documents): cria tabela de documentos`

---

### T6: Criar o serviço de documentos: guardar, versionar e gerar URL de acesso, com permissão e auditoria

**What**: Criar o serviço de documentos: guardar, versionar e gerar URL de acesso, com permissão e auditoria.  
**Where**: `api/internal/platform/documents/service.go`  
**Depends on**: T5  
**Reuses**: `storage`, `audit`, `authz`, `WithTx`  
**Requirement**: DOC-01 (ACs 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11); DOC-02 (ACs 1, 2, 3)

**Tools**:

- MCP: NONE
- Skill: security-best-practices

**Done when**:

- [ ] Extensão fora de `.pdf`, `.jpg`, `.jpeg`, `.png` e `.webp` é recusada antes de ler o conteúdo
- [ ] Tipo detectado pelo conteúdo deve coincidir com a extensão
- [ ] Limite de 10 MiB e tipos por conteúdo aplicados antes de gravar metadados
- [ ] Versão nova referencia a anterior e não altera o objeto antigo
- [ ] Falha ao gravar metadados remove só o objeto recém-criado
- [ ] `document.create` e `document.access` auditados na transação
- [ ] Sem permissão não escreve nem gera URL
- [ ] Cada AC listado em Requirement tem ao menos um teste (mínimo de 14 testes; nenhuma exclusão silenciosa)
- [ ] Gate check passes: `cd api && go vet ./... && go test -tags=integration ./...`

**Tests**: integration  
**Gate**: full

**Commit**: `feat(documents): adiciona servico de documentos versionados`

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

## Diagram-Definition Cross-Check

Dentro de cada fase, cada tarefa depende apenas da anterior (execução sequencial); dependências entre fases apontam sempre para trás.

| Task | Depends On (task body) | Diagram Shows | Status |
| ---- | ---------------------- | ------------- | ------ |
| T1 | - | - | ✅ Match |
| T2 | T1 | T1 | ✅ Match |
| T3 | T2 | T2 | ✅ Match |
| T4 | T3 | T3 | ✅ Match |
| T5 | - | - | ✅ Match |
| T6 | T5 | T5 | ✅ Match |

## Test Co-location Validation

| Task | Code Layer Created/Modified | Matrix Requires | Task Says | Status |
| ---- | --------------------------- | --------------- | --------- | ------ |
| T1: cfg_s3 | Lógica pura, middleware ou CLI | unit | unit | ✅ OK |
| T2: st_if | Contrato, config, CI, Docker, docs ou tipos sem lógica | none | none | ✅ OK |
| T3: tu_s3 | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T4: st_s3 | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T5: d_mig | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |
| T6: d_svc | Caso de uso, repositório, SQL, handler ou storage | integration | integration | ✅ OK |

## Requirement Coverage

| Requirement ID | Tasks |
| -------------- | ----- |
| DOC-01 | T2, T4, T5, T6 |
| DOC-02 | T3, T4, T6 |
| DOC-03 | T1 |

**Coverage:** 3 total, 3 mapped to tasks, 0 unmapped
