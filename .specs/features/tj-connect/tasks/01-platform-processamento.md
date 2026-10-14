# Platform e processamento — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules.

**Se a skill não puder ser ativada, PARE e avise.**

---

**Spec**: `.specs/features/tj-connect/spec/01-platform-processamento.md`
**Design**: `.specs/features/tj-connect/design/01-platform-processamento.md`
**Status**: Draft — aguardando revisão do mantenedor. **Execute não autorizado.** Tests-first, um commit atômico por tarefa, só local (push, PR, merge e deploy exigem autorização explícita).

**Pré-requisitos** (tasks `02-identity-governanca`):
- fase 1 (J-1): nenhum;
- fases 2 a 7: T10, T11, T13, T15 e T18 (versão da matriz, contas de serviço e `ServicePrincipals`) e T21–T22 (autoria `service`) das tasks `02`;
- fase 6 (T22 em diante): **toda** a fase 6 das tasks `02` (T25–T30), porque T22 verifica a matriz final de permissões.

**Ordem consolidada do G1 (59 tarefas)**: `01` T1–T2 (a qualquer momento) → `02` T1–T7 → `02` T8–T20 → `02` T21–T23 → `02` T24 → `02` T25–T30 → `01` T3–T24 → `02` T31–T32 e `01` T25–T27.

Esses pré-requisitos ficam declarados em cada tarefa como "Pré-requisito externo", porque o validador só cruza dependências dentro deste arquivo.

**Decisões cobertas**: D3, D-01-1, D-01-2 e os ajustes obrigatórios 2 (identidade da entrega de eventos) e 4 (recuperação de concessões). `jobs:job:close` continua **sem concessão**.

## Test Coverage Matrix

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| `platform/modulegraph` | unit | setas, tipos, ordenação topológica | `api/internal/platform/modulegraph/*_test.go` | `go test ./internal/platform/modulegraph/...` |
| Teste arquitetural | unit (árvores sintéticas) | cada violação de ARQ-01 detectada | `api/internal/architecture_test.go` | Gate `Architecture` |
| `platform/jobs` (tipos, registro, enfileirar) | unit + integration | cada recusa do registro; enfileirar na transação, deduplicação | `api/internal/platform/jobs/*_test.go` | `go test -tags=integration ./internal/platform/jobs/...` |
| `platform/jobs/worker` (captura, cerca, execução, recuperação) | integration + concorrência (PostgreSQL real, várias instâncias do runner) | estados, cerca, perdas de concessão sem consumir tentativas, sem efeito duplo | `api/internal/platform/jobs/worker/*_test.go` | `go test -tags=integration -race ./internal/platform/jobs/...` |
| `platform/outbox` | integration + concorrência | inbox, modos de ordem, identidade da entrega | `api/internal/platform/outbox/*_test.go` | `go test -tags=integration -race ./internal/platform/outbox/...` |
| Migration `000012` | integration (PostgreSQL real, como `tj_app`) | up, down, grants, cada `CHECK`, `down` recusado com trabalho | `api/internal/platform/database/platform_jobs_migration_test.go` | `go test -tags=integration ./internal/platform/database/...` |
| `config` (`LoadWorker`) | unit | padrões e validações | `api/internal/config/*_test.go` | `go test ./internal/config/...` |
| `cmd/worker` | integration | partida, verificação do RBAC, saída com erro, sem porta | `api/cmd/worker/*_test.go` | `go test -tags=integration ./cmd/worker/...` |
| `platform/jobs` admin + handlers | integration (httptest + PostgreSQL) | permissões, política do módulo, obrigação, auditoria | `api/internal/platform/jobs/admin_test.go`, `api/internal/httpapi/*_test.go` | `go test -tags=integration ./internal/...` |
| Contrato OpenAPI | lint + geração | contrato válido e tipos gerados | `api/openapi/platform.yaml` | Gate `Contract` |
| Ensaio de migrações | integration (PostgreSQL real) | `000001`→`000012` sobre dados sintéticos | `api/internal/platform/database/g1_rehearsal_migration_test.go` | `go test -tags=integration ./internal/platform/database/...` |
| Build de implantação | build | imagem com os dois binários | `api/Dockerfile`, `docker-compose.yml` | Gate `Build` |
| Documentação de operação | none | revisão humana | `docs/operations/*.md` | — |

## Gate Check Commands

| Level | When | Command |
| --- | --- | --- |
| Full | toda tarefa de código | `cd api && go vet ./... && go vet -tags=integration ./... && go build ./... && go test ./... && go test -tags=integration ./...` |
| Race | tarefas do worker e do outbox | `cd api && go test -tags=integration -race -count=1 ./internal/platform/jobs/... ./internal/platform/outbox/...` (exige cgo; roda no CI Linux se o ambiente local não tiver) |
| Lint | toda tarefa de código | `cd api && golangci-lint run --path-mode=abs --build-tags=integration` |
| Architecture | toda tarefa de código | `cd api && go test ./internal/... -run TestProductionCodeRespectsTheModuleBoundaries` |
| Contract | tarefas que mudam `api/openapi/*.yaml` | `cd web && pnpm lint:api && pnpm gen:api && pnpm exec tsc --noEmit` e `cd api && go generate ./...` |
| Build | tarefas de empacotamento | `cd api && go build ./cmd/api ./cmd/worker ./cmd/bootstrap-admin` e `docker build` local da imagem (sem publicar nem implantar) |
| Docs | tarefas só de documentação | revisão humana |

Testes de integração e concorrência usam testcontainers com PostgreSQL 16 real. "Relógio do banco controlado" significa ajustar `executar_em`, `locked_until`, `ultimo_batimento` e `terminado_em` para o passado; o código nunca lê o relógio da instância. Nenhum teste chama serviço externo; o "provedor" dos testes de modo externo é um servidor `httptest` local.

---

## Execution Plan

### Phase 1: Grafo de módulos e teste arquitetural (J-1)

```
T1 → T2
```

### Phase 2: Migration 000012 e registro de jobs (J-2)

```
T3 → T5
T4 → T5
T6
```

### Phase 3: Worker, concorrência e recuperação (J-3)

```
T7 → T8 → T9 → T10 → T13
T9 → T11 → T12 → T13
T13 → T14
```

### Phase 4: Outbox, inbox e identidade da entrega (J-4)

```
T15 → T16 → T17 → T18
```

### Phase 5: Agendamentos, limpeza e alertas (J-5)

```
T19 → T20
T19 → T21
```

### Phase 6: Operação (J-6)

```
T22 → T23 → T24
```

### Phase 7: Validação com PostgreSQL real e preparação da implantação (sem implantar)

```
T25 → T27
T26 → T27
```

---

## Task Breakdown

### Phase 1: Grafo de módulos e teste arquitetural (J-1)

#### T1: `platform/modulegraph`

**What**: `Kind` (`D`, `E`, `R`), `Edges` com as setas atuais (`financeiro → identity` R, `financeiro → estoque` D, `estoque → identity` R) e `Allowed`; o pacote não importa nada.
**Where**: `api/internal/platform/modulegraph/edges.go` (novo)
**Depends on**: None
**Reuses**: ADR-011.
**Requirement**: ARQ-01 AC1 (fonte das setas); OUT-01 AC5 (fonte para a validação de inscrições).
**Impacto**: nenhum.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] `Allowed` cobre seta declarada, não declarada e tipo errado
- [ ] Gate check passes

**Tests**: unit
**Gate**: Full
**Commit**: `feat(platform): grafo de modulos permitido`

---

#### T2: Teste arquitetural estendido

**What**: as sete verificações de P7 (import fora de `Edges`, tipo pelo pacote, `events` sem lógica, ciclo, módulo sem entrada, aviso de seta não usada, `cmd/api` sem `platform/jobs/worker` nem tratadores), cada uma com árvore sintética.
**Where**: `api/internal/architecture_test.go`
**Depends on**: T1
**Reuses**: `checkBoundaries`, árvores sintéticas existentes.
**Requirement**: ARQ-01 AC1–7; T-17.
**Impacto**: todos os módulos: o código atual passa (aviso esperado só para `financeiro → estoque`); `composition` (tasks `02`, T10) é aceito como pacote de composição.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Cada violação sintética faz o teste falhar com mensagem clara
- [ ] Código de produção atual passa
- [ ] Gate check passes

**Tests**: unit
**Gate**: Full + Architecture
**Commit**: `test(arch): setas permitidas e regras do worker no teste arquitetural`

---

### Phase 2: Migration 000012 e registro de jobs (J-2)

#### T3: Migration `000012_platform_jobs`

**What**: criar `000012` conforme o design (com `outbox_events` antes de `jobs`; `perdas_concessao_consecutivas`; restrição de `EM_EXECUCAO`; restrição de `jobs.entregar_evento` com `evento_id` e `consumidor`; `job_attempts.instancia_viva` (diagnóstico) e `executor_travado`; `worker_instancias.encerrado_em`; grants sem `DELETE` em `versoes_aplicadas`); `down` recusa com job `PENDENTE`, `EM_EXECUCAO` ou `FALHOU`.
**Where**: `api/migrations/000012_platform_jobs.up.sql`, `.down.sql`, `api/internal/platform/database/platform_jobs_migration_test.go`
**Depends on**: None
**Pré-requisito externo**: migrations `000010` e `000011` (tasks `02`, T8 e T21), pela numeração.
**Reuses**: padrão `*_migration_test.go`.
**Requirement**: JOB-01; LEA-01 (estrutura); REC-01 AC7–9 (colunas); design `01` "Data Models".
**Impacto**: banco: só tabelas novas.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Up/down em PostgreSQL real; grants de `tj_app` conferidos (incluindo a ausência de `DELETE` em `versoes_aplicadas`)
- [ ] Cada `CHECK` recusa a linha inválida e aceita a válida (estado × concessão; entrega sem consumidor; `instancia_viva` fora de `concessao_perdida`; payload > 64 KiB)
- [ ] `down` recusado com trabalho pendente
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(platform): migration de jobs, outbox e inbox`

---

#### T4: Tipos e registro de jobs

**What**: `Type`, `Effect`, `Mode`, `RetryPolicy`, `ObligationCheck`, `Request` e `Registry.Register` com as recusas de P1 (nome duplicado, `Idempotent=false`, `Timeout >= concessão`, efeito sem política, conta de serviço inexistente ou não `SERVICE` via interface `AccountChecker`, executor transacional com porta externa). A interface fica em `platform`; a implementação é `identity.ServicePrincipals`, ligada em `composition`.
**Where**: `api/internal/platform/jobs/job.go`, `registry.go`, `errors.go` (novos)
**Depends on**: None
**Pré-requisito externo**: tasks `02` T18 (`ServicePrincipals`) para o teste de integração da conta.
**Reuses**: `authz.Permission`.
**Requirement**: JOB-01; REC-01 AC6; SVX-01 AC1–2; DLQ-01 AC9 (política obrigatória).
**Impacto**: nenhum (pacote novo).

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Cada recusa do registro testada; registro válido aceito
- [ ] Gate check passes

**Tests**: unit
**Gate**: Full
**Commit**: `feat(platform): tipos e registro de jobs`

---

#### T5: Enfileirar na transação do caso de uso

**What**: `Enqueuer` por módulo: exige `TxFrom`; recusa prefixo de outro módulo e o tipo `jobs.entregar_evento` vindo de módulo; limite de 64 KiB; chaves proibidas recursivas; deduplicação ativa só entre `PENDENTE`/`EM_EXECUCAO` com `ON CONFLICT ... DO NOTHING` e leitura do id existente.
**Where**: `api/internal/platform/jobs/enqueue.go` (novo)
**Depends on**: T3, T4
**Reuses**: `database.TxFrom`; padrão `ErrNoTransaction` de `audit/recorder.go:46`.
**Requirement**: JOB-01 AC1–6; SVX-01 AC7 (parte de recusa); T-01, T-02.
**Impacto**: nenhum: nenhum módulo enfileira ainda.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Commit → job existe; rollback → não existe; fora de transação → erro e nada gravado
- [ ] Deduplicação devolve o id existente; depois de `FALHOU`, novo job é aceito
- [ ] `jobs.entregar_evento` por módulo → recusado
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(platform): enfileirar job na transacao do caso de uso`

---

#### T6: Contribuição `jobs` com a conta de manutenção

**What**: contribuição do pseudo-módulo `jobs` em `composition`, só com a conta `svc.jobs-manutencao` (papel `SERVICO_JOBS_MANUTENCAO`, sem permissões); `MatrixVersion` sobe.
**Where**: `api/internal/composition/jobs_contribution.go` (novo)
**Depends on**: None
**Pré-requisito externo**: tasks `02` T10, T13, T15.
**Reuses**: `Contribution.ServiceAccounts`.
**Requirement**: SVX-01 AC1–2; design `01` P5.
**Impacto**: a sincronização cria a conta `svc.jobs-manutencao` (auditada); nenhuma permissão humana muda.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Sincronização cria a conta uma vez; teste de versão e snapshot atualizados
- [ ] Gate check passes

**Tests**: unit
**Gate**: Full
**Commit**: `feat(platform): conta de servico de manutencao de jobs`

---

### Phase 3: Worker, concorrência e recuperação (J-3)

#### T7: Configuração do worker

**What**: `config.LoadWorker()` com todas as variáveis `WORKER_*` e `JOBS_RETENTION_DAYS` (padrões de P9, incluindo `WORKER_HUNG_GRACE`, `WORKER_INFRA_BACKOFF_MAX`, `WORKER_SUSPECT_LOSSES`, `WORKER_RBAC_WAIT`; sem `WORKER_MAX_LEASE_LOSSES`); valida `WORKER_DEFAULT_TIMEOUT < WORKER_LEASE`; id de instância `hostname-pid-<aleatório>`.
**Where**: `api/internal/config/worker.go` (novo)
**Depends on**: None
**Reuses**: helpers de `config.go:60`.
**Requirement**: D-01-2; OBS-01 AC6.
**Impacto**: `config`: a API não muda.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Padrões, valores inválidos e a relação tempo × concessão testados
- [ ] Gate check passes

**Tests**: unit
**Gate**: Full
**Commit**: `feat(config): configuracao do worker`

---

#### T8: Captura, concessão e cerca

**What**: captura com `FOR UPDATE SKIP LOCKED` (P2); renovação a cada `concessão/2` limitada a `Timeout + WORKER_HUNG_GRACE`; condição única de escrita cercada (`id`, `lease_token`, `estado = 'EM_EXECUCAO'`, `locked_until > now()`); toda saída de `EM_EXECUCAO` limpa a concessão.
**Where**: `api/internal/platform/jobs/worker/claim.go`, `lease.go` (novos)
**Depends on**: T7, T3, T5
**Reuses**: `database.WithTx`.
**Requirement**: LEA-01 AC1–6; T-03, T-04.
**Impacto**: nenhum.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] 5 runners × 100 jobs: nenhuma captura dupla (T-03), com `-race`
- [ ] Token antigo não conclui, não falha e não renova (T-04); concessão vencida sem recaptura também é recusada
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full + Race
**Commit**: `feat(platform): captura e cerca de jobs`

---

#### T9: Execução, tentativas, espera e classificação de erros

**What**: modo transacional (efeito, idempotência e conclusão cercada numa transação; zero linhas desfaz tudo); modo externo (sem transação durante a chamada, `TxFrom` falso na entrada do executor, chave durável antes da chamada); classificação de erros (`Permanent`, tempo, `40001`/`40P01`, rede); espera exponencial com variação; `FALHOU` com log `error`; `job_attempts` sanitizado e truncado; `recover` de panic como transitório.
**Where**: `api/internal/platform/jobs/worker/runner.go`, `backoff.go` (novos)
**Depends on**: T8
**Reuses**: `database.WithTx`, `logx`.
**Requirement**: LEA-01 AC7–9; RTY-01; REC-01 AC11; T-04b, T-04c, T-05, T-07c.
**Impacto**: nenhum.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Execução sobreposta com pausa injetada: um efeito na tabela de teste com chave única, conclusão antiga rejeitada, contexto cancelado (T-04b)
- [ ] Modo externo repetido por A e B usa a mesma chave; provedor `httptest` registra um efeito (T-04c)
- [ ] Erros transitório, permanente e de tempo: estados e esperas corretos (T-05)
- [ ] Panic recuperado; a instância continua (T-07c)
- [ ] Mensagem de erro sanitizada e ≤ 2 KiB
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full + Race
**Commit**: `feat(platform): execucao de jobs com tentativas e cerca`

---

#### T10: Conta de serviço por job

**What**: antes de executar, `ServicePrincipals.For(conta do tipo)`; conta inativa → adiamento de 60 s sem gastar tentativa, `warn` limitado a 1/min por tipo; executor recebe o `Principal`, `correlation_id`, `causation_id = job.id` e `decision_id`.
**Where**: `api/internal/platform/jobs/worker/principal.go` (novo)
**Depends on**: T9
**Pré-requisito externo**: tasks `02` T18, T22.
**Reuses**: chaves de contexto de `audit` (tasks `02`, T22).
**Requirement**: SVX-01 AC3–5; T-12, T-13.
**Impacto**: nenhum.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Executor de teste chama caso de uso auditado: entrada `service` com `causation_id` e `correlation_id` (T-13)
- [ ] Conta desativada → adiamento sem consumir tentativa; log limitado (T-12)
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(platform): execucao com principal da conta de servico`

---

#### T11: Batimento, encerramento gracioso e executor travado

**What**: batimento a cada 15 s (`worker_instancias`, usado só para o resumo e o diagnóstico); no sinal, para de capturar, espera `WORKER_SHUTDOWN_GRACE`, grava `encerrado_em` e sai; detecção de executor que não devolve o controle: para de renovar, grava `executor_travado` (conta tentativa) e, se as goroutines abandonadas chegarem a `WORKER_CONCURRENCY`, encerra a instância **depois** de gravar as tentativas.
**Where**: `api/internal/platform/jobs/worker/heartbeat.go`, `hung.go` (novos)
**Depends on**: T9
**Reuses**: `logx`.
**Requirement**: REC-01 AC5, AC12; OBS-01 AC2.
**Impacto**: nenhum.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Sinal com job em curso: espera até o prazo, grava `encerrado_em`
- [ ] Executor bloqueado além do prazo: o runner, ainda com a concessão válida, grava `executor_travado` (conta tentativa) antes de qualquer autoencerramento
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full + Race
**Commit**: `feat(platform): batimento, encerramento e deteccao de executor travado`

---

#### T12: Recuperação de concessões perdidas (ajuste obrigatório 4, revisado)

**What**: recuperação na partida e a cada 30 s (P3): toda concessão vencida volta a `PENDENTE` **sem consumir tentativa** e **sem nunca ir a `FALHOU`**, com espera `min(30 s · 2^(n-1), 15 min)`; tentativa `concessao_perdida` com `instancia_viva` só como diagnóstico; `perdas_concessao_consecutivas` zerado em toda tentativa decidida pelo runner com concessão válida; alerta `job_suspeito_de_derrubar_worker` na 5ª perda seguida; banco indisponível não muda estado (reconexão com espera).
**Where**: `api/internal/platform/jobs/worker/recovery.go` (novo)
**Depends on**: T11
**Reuses**: condição cercada de T8; `executor_travado` de T11 (única forma de travamento que consome tentativa).
**Requirement**: REC-01 AC1–4, AC7–9; T-06, T-06b, T-07, T-07b.
**Impacto**: nenhum.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Instância morta no meio do job → outra conclui (T-06)
- [ ] Concessão perdida com a instância **viva** (renovação pausada) → volta sem consumir tentativa; o diagnóstico registra `instancia_viva = true` e o tratamento é o mesmo (T-06b)
- [ ] Banco parado e reiniciado no meio da execução (testcontainers `Stop`/`Start`) → nenhum `FALHOU` (T-07)
- [ ] 5 quedas seguidas no mesmo job → alerta, espera crescente até 15 min, nunca `FALHOU`; reinício gracioso tratado igual (T-07b)
- [ ] Nenhum caminho da recuperação lê o batimento para decidir estado ou tentativas (teste de tabela sobre as combinações de batimento)
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full + Race
**Commit**: `feat(platform): recuperacao de concessoes perdidas sem consumir tentativas`

---

#### T13: `cmd/worker`

**What**: binário com a composição de P9: `LoadWorker`, banco, `Recorder` com resolvedor, `BuildMatrix(composition.Contributions())`, **verificação** da versão do RBAC (sem sincronizar; espera até `WORKER_RBAC_WAIT`), `ServicePrincipals`, registro dos tipos da plataforma, `worker.Run`; sem porta; saída diferente de zero com a causa em falha de configuração, banco ou versão.
**Where**: `api/cmd/worker/main.go` (novo)
**Depends on**: T10, T12
**Pré-requisito externo**: tasks `02` T10, T11.
**Reuses**: composição de `cmd/api/main.go`.
**Requirement**: OBS-01 AC6–7; SVX-01 AC2; TRN-01 AC11 (tasks `02`); ARQ-01 AC7.
**Impacto**: novo binário; `cmd/api` não muda.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Partida com versão gravada menor espera e sobe quando a API sincroniza; com versão maior ou hash diferente, sai com erro
- [ ] Conta de serviço declarada inexistente → não sobe
- [ ] Nenhuma porta aberta; teste arquitetural verde
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full + Architecture
**Commit**: `feat(worker): binario do worker`

---

#### T14: Suíte de falhas e idempotência do worker

**What**: suíte dedicada que repete os cenários críticos com várias instâncias e `-race`, sobre um efeito "financeiro" de teste com chave única durável (`cobranca:<id>:receita`), cobrindo cada linha da tabela "Como duplicação, reinício e falha parcial não duplicam efeito financeiro": captura concorrente, sobreposição depois da perda, falha entre efeito e conclusão, banco reiniciado, quedas repetidas, executor travado e encerramento gracioso. Cada cenário afirma "um efeito, nenhum perdido, nenhuma obrigação abandonada".
**Where**: `api/internal/platform/jobs/worker/failure_suite_test.go` (novo, só teste)
**Depends on**: T13
**Reuses**: fixtures de T8–T12.
**Requirement**: Success Criteria da spec `01` (T-04, T-04b, T-04c, T-06, T-07); REC-01; LEA-01.
**Impacto**: nenhum em produção.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Cada cenário roda 50 vezes sem efeito duplo nem perda
- [ ] Nenhum job com efeito termina em `FALHOU` por perda de concessão
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full + Race
**Commit**: `test(platform): suite de falhas e idempotencia do worker`

---

### Phase 4: Outbox, inbox e identidade da entrega (J-4)

#### T15: Publicação e registro de inscrições

**What**: `Event`, `Publisher` por módulo (exige transação, recusa prefixo alheio e envelope incompleto), `Subscription` e registro de inscrições; validação na partida contra `modulegraph` (seta `E`) e `tolerante` exige `Convergent`; `Publish` cria um job de entrega por inscrição com `conta_servico` **da inscrição**, gravada só pela plataforma.
**Where**: `api/internal/platform/outbox/event.go`, `publish.go`, `subscriptions.go` (novos)
**Depends on**: T1, T5, T13
**Reuses**: `Enqueuer` interno da plataforma.
**Requirement**: OUT-01 AC1–6; SVX-01 AC1, AC7; T-08, T-09.
**Impacto**: nenhum: nenhum produtor ainda.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Publicar e confirmar → evento e um job por consumidor; desfazer → nada (T-08)
- [ ] Inscrição numa seta proibida → partida falha (T-09)
- [ ] A conta gravada em cada entrega é a da inscrição
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(platform): outbox e inscricoes de eventos`

---

#### T16: Entrega com inbox e modos de ordem

**What**: executor transacional `jobs.entregar_evento`: inbox `(consumidor, evento_id)`; modos `independente`, `estrita` (erro transitório fora de sequência) e `tolerante` (`OBSOLETO`); `versoes_aplicadas` com `FOR UPDATE`; tratador; atualização da versão; conclusão cercada — tudo numa transação.
**Where**: `api/internal/platform/outbox/deliver.go` (novo)
**Depends on**: T15
**Reuses**: runner transacional (T9).
**Requirement**: INB-01 AC1–9; T-10, T-11, T-11b.
**Impacto**: nenhum.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Mesmo evento 3 vezes → 1 efeito (T-10)
- [ ] v2 antes de v1 nos dois modos (T-11)
- [ ] v2 e v3 simultâneos com barreira, 200 repetições, com `-race` (T-11b)
- [ ] Falha do tratador não grava inbox nem versão (AC9)
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full + Race
**Commit**: `feat(platform): entrega de eventos com inbox e ordem`

---

#### T17: Identidade da entrega (ajuste obrigatório 2)

**What**: para `jobs.entregar_evento`, o runner ignora a conta do tipo e revalida a da linha antes da transação: igual à inscrição atual, `SERVICE`, ativa, do módulo do consumidor; falha (exceto inatividade) → `FALHOU` permanente `conta_de_entrega_invalida`, sem tratador; inbox, versão e conclusão sem `Principal`; o tratador recebe só o `Principal` da sua conta.
**Where**: `api/internal/platform/outbox/delivery_identity.go` (novo)
**Depends on**: T16
**Reuses**: `ServicePrincipals.For`, adiamento de conta inativa (T10).
**Requirement**: SVX-01 AC1, AC8–9; T-13b, T-13c.
**Impacto**: nenhum.

**Tools**: MCP: NONE · Skill: security-best-practices

**Done when**:
- [ ] Dois consumidores de módulos diferentes no mesmo evento: cada tratador vê só a sua conta; auditorias com contas distintas (T-13b)
- [ ] Linha adulterada com a conta do outro consumidor, conta de outro módulo ou inscrição removida → `conta_de_entrega_invalida`, nenhum tratador chamado (T-13c)
- [ ] Conta do consumidor inativa → adiamento, sem `FALHOU`
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(platform): identidade da conta de servico na entrega de eventos`

---

#### T18: Suíte de falhas e idempotência do outbox

**What**: suíte dedicada com várias instâncias e `-race`: queda entre a inbox e a conclusão, instância "zumbi" no consumidor, evento duplicado concorrente, publicação concorrente da mesma versão de agregado (recusada pela chave única), consumidor lento com perda de concessão. Cada cenário afirma "efeito confirmado no máximo uma vez por (consumidor, evento)" e "nenhuma entrega perdida".
**Where**: `api/internal/platform/outbox/failure_suite_test.go` (novo, só teste)
**Depends on**: T17
**Reuses**: fixtures de T15–T17.
**Requirement**: INV-P4; OUT-01 AC6; INB-01; Success Criteria da spec `01`.
**Impacto**: nenhum em produção.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Cada cenário roda 50 vezes sem efeito duplo nem entrega perdida
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full + Race
**Commit**: `test(platform): suite de falhas e idempotencia do outbox`

---

### Phase 5: Agendamentos, limpeza e alertas (J-5)

#### T19: Agendamentos

**What**: `Schedule`, janela calculada no banco, `chave_agendamento`, encadeamento atômico na transação de conclusão, verificação independente a cada 60 s com `agendamento_recriado`.
**Where**: `api/internal/platform/jobs/schedule.go`, `api/internal/platform/jobs/worker/ensure.go` (novos)
**Depends on**: T13
**Reuses**: runner transacional.
**Requirement**: SCH-01 AC1–6; INV-P10; T-15, T-15b, T-15c.
**Impacto**: nenhum.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] 3 runners → 1 execução por janela (T-15)
- [ ] Falha injetada na conclusão → nem conclusão nem próxima janela; recuperado pela concessão (T-15b)
- [ ] Job da próxima janela removido → recriado em até 60 s (T-15c)
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full + Race
**Commit**: `feat(platform): agendamentos periodicos`

---

#### T20: Limpeza com deduplicação durável

**What**: `jobs.limpeza` diário em lotes de 500: eventos cujas entregas estão todas `CONCLUIDO`/`ENCERRADO_MANUALMENTE` além da retenção, com reconferência sob `FOR UPDATE`; jobs sem evento `CONCLUIDO`; nunca `FALHOU`, `ENCERRADO_MANUALMENTE`, `EM_EXECUCAO` nem `versoes_aplicadas`; `worker_instancias` sem batimento há 7 dias e sem concessão.
**Where**: `api/internal/platform/jobs/worker/cleanup.go` (novo)
**Depends on**: T19
**Reuses**: conta `svc.jobs-manutencao` (T6).
**Requirement**: RET-01; RTY-01 AC8; T-11d, T-11e.
**Impacto**: nenhum.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Entrega `FALHOU` além da retenção → evento e inbox preservados; reenfileirar → um efeito (T-11d)
- [ ] Limpeza concorrente com reenfileiramento → evento preservado (T-11e)
- [ ] Instância com concessão ativa nunca é removida
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full + Race
**Commit**: `feat(platform): limpeza de jobs com deduplicacao duravel`

---

#### T21: Alertas diários de contingência e de obrigação em falha

**What**: agendamentos `jobs.alerta_contingencia` (revisões de contingência abertas, tasks `02` T19) e `jobs.alerta_obrigacoes` (`obrigacao_em_falha` para todo `FALHOU` com `Effect != nenhum`), ambos com `svc.jobs-manutencao` e log `error`/`warn` estruturado.
**Where**: `api/internal/platform/jobs/worker/alerts.go` (novo)
**Depends on**: T19
**Pré-requisito externo**: tasks `02` T19.
**Reuses**: consulta de revisões abertas publicada por `identity/app`.
**Requirement**: REC-01 AC10; SVC-01 AC21 (tasks `02`).
**Impacto**: nenhum.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Job com efeito em `FALHOU` → alerta diário até sair de `FALHOU`; a obrigação continua no módulo
- [ ] Revisão de contingência aberta → alerta diário
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(platform): alertas de contingencia e de obrigacao em falha`

---

### Phase 6: Operação (J-6)

#### T22: Permissões `jobs` e concessões da D-01-1

**What**: definições `jobs:job:read`, `jobs:job:retry` e `jobs:job:close`, **todas `sensivel`** (spec `01`, tabela de permissões; D-01-1), na contribuição `jobs`; concessões **explícitas**: `ADMIN_SISTEMA` read e retry, `PRESIDENTE` read — nenhuma chega por concessão automática de permissões `normal`; `jobs:job:close` **sem concessão**; `MatrixVersion` sobe; snapshot "Depois do G1" conferido.
**Where**: `api/internal/composition/jobs_contribution.go`
**Depends on**: T6
**Pré-requisito externo**: tasks `02` **T29** (matriz final de identity: D-02-3, D-02-5 e a permissão `identity:rbac:rollback_authorize`). Esta tarefa verifica a matriz final de permissões; executá-la antes de `02/T29` faria a verificação falhar.
**Reuses**: regras de classe de `BuildMatrix`.
**Requirement**: D-01-1; DLQ-01 AC1–3.
**Impacto**: `ADMIN_SISTEMA` 11 → 13; `PRESIDENTE` +1 (33 ao fim do G1, com a D-02-5).

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] As três permissões têm classe `sensivel`; retirar a concessão explícita de `jobs:job:read` do `PRESIDENTE` num teste faz ele perder a permissão (prova de que não é automática)
- [ ] Snapshot confere as contagens finais (`PRESIDENTE` 33, `ADMIN_SISTEMA` 13) e `jobs:job:close` sem papel
- [ ] Gate check passes

**Tests**: unit
**Gate**: Full
**Commit**: `feat(platform): permissoes de operacao de jobs`

---

#### T23: Serviço de operação: listar, resumo, reenfileirar e encerrar

**What**: `jobs.Admin`: listagem sem payload; resumo (estados, idade do mais antigo, `FALHOU` com obrigação, suspeitos de derrubar worker, instâncias, `nenhum_worker_ativo`); reenfileirar com `FOR UPDATE`, só de `FALHOU`, política do módulo (`ExtraPermission`), novo ciclo, auditoria `job.retry`; encerrar com `ClosePermission` e obrigação resolvida, auditoria `job.close`, sem marcar obrigação como cumprida.
**Where**: `api/internal/platform/jobs/admin.go` (novo)
**Depends on**: T22
**Reuses**: `audit.Recorder`.
**Requirement**: DLQ-01 AC1–9; OBS-01 AC3–5; SVX-01 AC6; T-14, T-14b, T-16.
**Impacto**: nenhum para usuários atuais além da consulta nova.

**Tools**: MCP: NONE · Skill: security-best-practices

**Done when**:
- [ ] Reenfileirar e encerrar com e sem permissões; fora de `FALHOU` → `conflict` (T-14)
- [ ] Encerrar com obrigação pendente → `obrigacao_pendente`; resolvida → encerrado sem marcar a obrigação (T-14b)
- [ ] Sem batimento recente → `nenhum_worker_ativo` (T-16)
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `feat(platform): operacao de jobs com politica do modulo dono`

---

#### T24: Contrato e rotas de operação

**What**: `GET /api/v1/jobs`, `GET /api/v1/jobs/resumo`, `POST /api/v1/jobs/{id}/reenfileirar`, `POST /api/v1/jobs/{id}/encerrar` em `platform.yaml` e handlers na cadeia autenticada.
**Where**: `api/openapi/platform.yaml`, `api/internal/platform/jobs/http.go` (novo)
**Depends on**: T23
**Reuses**: padrão do `audit.Handler`; `httpx.WriteProblem`.
**Requirement**: DLQ-01; OBS-01 AC3; AD-010/012/013.
**Impacto**: `web`: tipos gerados; nenhuma tela neste grupo.

**Tools**: MCP: NONE · Skill: security-best-practices

**Done when**:
- [ ] Contrato válido; sessão, CSRF e permissão verificados; problem+json
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full + Contract
**Commit**: `feat(platform): rotas de operacao de jobs`

---

### Phase 7: Validação com PostgreSQL real e preparação da implantação (sem implantar)

#### T25: Ensaio das migrations do G1 sobre dados sintéticos

**What**: teste que aplica `000001` → `000012` em PostgreSQL real sobre dados **sintéticos** representativos (usuários de todos os papéis, sessões, tokens de recuperação, auditoria antiga, contas e lançamentos fictícios), confere invariantes (todas as contas `HUMAN`, auditoria intacta, gatilhos ativos, `rbac_sync_state = 1`), sobe a API e o worker simulados na ordem do runbook e testa os `down` de `000012` → `000010` (recusados quando há dados novos). Mede o tempo de cada migration.
**Where**: `api/internal/platform/database/g1_rehearsal_migration_test.go` (novo, só teste)
**Depends on**: T3, T13
**Pré-requisito externo**: tasks `02` T9, T11, T21.
**Reuses**: `platform/testutil/postgres.go`.
**Requirement**: TRN-01 AC6–7, AC10 (tasks `02`); JOB-01; design `02` "Compatibilidade com a base atual".
**Impacto**: nenhum em produção; nenhum dado real usado.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Ensaio verde, com tempos registrados no `STATE.md`
- [ ] Nenhum dado pessoal real nem cópia de produção usada
- [ ] Gate check passes

**Tests**: integration
**Gate**: Full
**Commit**: `test(platform): ensaio das migrations do g1 com postgresql real`

---

#### T26: Empacotamento do worker

**What**: `Dockerfile` ganha `go build -o /out/worker ./cmd/worker` na mesma imagem (`ENTRYPOINT` continua `/api`); `docker-compose.yml` ganha o serviço `worker` no profile `worker`, com o mesmo `DATABASE_URL` (`tj_app`). Só build local; nada é publicado nem implantado.
**Where**: `api/Dockerfile`, `docker-compose.yml`
**Depends on**: T13
**Reuses**: build multi-stage distroless atual.
**Requirement**: design `01` P9 (implantação proposta).
**Impacto**: imagem um pouco maior; o serviço `api` não muda.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] `docker build` local gera `/api` e `/worker`; `docker compose --profile worker config` válido
- [ ] Nenhum segredo novo em arquivo versionado; `.env.example` só com nomes e valores de desenvolvimento
- [ ] Gate check passes

**Tests**: none
**Gate**: Build
**Commit**: `build(worker): segundo binario na imagem e servico no compose`

---

#### T27: Runbook do worker e checklist de prontidão do G1 (sem implantar)

**What**: completar o runbook das tasks `02` (T32) com a parte do worker: ordem (migrations → API nova → worker → retirada das antigas), variáveis `WORKER_*`, verificações somente leitura (resumo de jobs, batimentos, `rbac_sync_state`), leitura dos alertas (`job_suspeito_de_derrubar_worker`, `obrigacao_em_falha`, `nenhum_worker_ativo`), parada segura do worker e limites dos `down`. Checklist de prontidão (go/no-go) do G1 para o mantenedor.
**Where**: `docs/operations/g1-worker-implantacao.md` (novo)
**Depends on**: T25, T26
**Pré-requisito externo**: tasks `02` T32.
**Reuses**: design `01` P3, P8, P9; design `02` C10.
**Requirement**: OBS-01; REC-01; D-01-2.
**Impacto**: nenhum.

**Tools**: MCP: NONE · Skill: NONE

**Done when**:
- [ ] Cada passo tem verificação objetiva e critério de interrupção
- [ ] Nenhum segredo, dado pessoal ou dado financeiro interno
- [ ] Revisado pelo mantenedor; a implantação continua exigindo autorização explícita
- [ ] Gate check passes

**Tests**: none
**Gate**: Docs
**Commit**: `docs(platform): runbook do worker e prontidao do g1`

---

## Avisos do validador e justificativas

`validate_tasks.py` (2026-10-14): 0 erros, 10 avisos. Nenhum foi eliminado artificialmente.

| Categoria | Tarefas | Justificativa concreta |
| --- | --- | --- |
| `Where` cita `up`, `down` e o teste de migration | T3 | Padrão do repositório para migrations |
| Arquivos do mesmo pacote para um único mecanismo | T4 (tipo, registro, erros), T8 (captura e concessão), T9 (runner e espera), T11 (batimento e executor travado), T15 (evento, publicação e inscrições) | São a divisão interna de um mecanismo que só é testável inteiro: a captura sem a concessão não tem cerca; a publicação sem as inscrições não cria entregas. Cada tarefa tem critérios de conclusão próprios e independentes das seguintes |
| Pacote da plataforma e o seu ponto de ligação no worker | T19 (`schedule.go` + `worker/ensure.go`) | A declaração do agendamento e a verificação independente formam a mesma garantia (INV-P10) |
| Contrato e handler juntos (contract-first) | T24 | AD-010/012 |
| `Tests: none` | T26, T27 | T26 é empacotamento, verificado pelo gate `Build` (`go build` dos três binários e `docker build` local); T27 é documentação, verificada pela revisão do mantenedor |

## Compatibilidade entre código e banco ao longo da sequência

- A fase 1 (J-1) é só teste e dados e pode começar antes das tasks `02`.
- A partir da fase 2, todo pré-requisito de banco de outra feature está declarado como "Pré-requisito externo" (migrations `000010`/`000011` antes de `000012`; contas de serviço, `ServicePrincipals` e versão da matriz antes do registro de tipos e do `cmd/worker`).
- `000012` é criada em T3 e só é usada a partir de T5; nenhuma tarefa posterior a altera.
- O worker não sincroniza o RBAC: T13 só lê `rbac_sync_state`, criada nas tasks `02` (T8 e T11).
- Nenhum commit intermediário é implantado; a implantação, quando autorizada, segue o runbook (T27).

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1, T5, T7, T10, T12, T17, T20, T21, T22, T23 | 1 componente | ✅ Granular |
| T2 | 1 arquivo de teste | ✅ Granular |
| T3 | migration + teste de migration | ✅ Aceitável (padrão do repositório) |
| T4 | 3 arquivos do mesmo pacote, um conceito (registro de tipos) | ✅ Aceitável |
| T8, T11 | 2 arquivos do mesmo mecanismo | ✅ Aceitável |
| T9 | runner + espera | ✅ Aceitável |
| T13 | 1 binário | ✅ Granular |
| T14, T18, T25 | 1 suíte de teste | ✅ Granular |
| T15 | publicação + inscrições (dependem uma da outra) | ✅ Aceitável |
| T16, T19 | 1–2 arquivos de um mecanismo | ✅ Aceitável |
| T24 | contrato + handler (contract-first) | ✅ Aceitável |
| T26 | 2 arquivos de empacotamento | ✅ Aceitável |
| T6, T27 | 1 arquivo | ✅ Granular |

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | — | ✅ Match |
| T2 | T1 | T1 | ✅ Match |
| T3, T4 | None | — | ✅ Match |
| T5 | T3, T4 | T3, T4 | ✅ Match |
| T6 | None | — | ✅ Match |
| T7 | None | — | ✅ Match |
| T8 | T7, T3 e T5 (fase anterior) | T7 | ✅ Match |
| T9 | T8 | T8 | ✅ Match |
| T10 | T9 | T9 | ✅ Match |
| T11 | T9 | T9 | ✅ Match |
| T12 | T11 | T11 | ✅ Match |
| T13 | T10, T12 | T10, T12 | ✅ Match |
| T14 | T13 | T13 | ✅ Match |
| T15 | T1, T5, T13 (fases anteriores) | — | ✅ Match |
| T16 | T15 | T15 | ✅ Match |
| T17 | T16 | T16 | ✅ Match |
| T18 | T17 | T17 | ✅ Match |
| T19 | T13 (fase anterior) | — | ✅ Match |
| T20, T21 | T19 | T19 | ✅ Match |
| T22 | T6 (fase anterior) | — | ✅ Match |
| T23 | T22 | T22 | ✅ Match |
| T24 | T23 | T23 | ✅ Match |
| T25 | T3, T13 (fases anteriores) | — | ✅ Match |
| T26 | T13 (fase anterior) | — | ✅ Match |
| T27 | T25, T26 | T25, T26 | ✅ Match |

## Test Co-location Validation

| Task | Code Layer Created/Modified | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1, T2 | `modulegraph`, teste arquitetural | unit | unit | ✅ OK |
| T3 | migration | integration | integration | ✅ OK |
| T4 | `platform/jobs` (tipos) | unit | unit | ✅ OK |
| T5 | `platform/jobs` (enfileirar) | integration | integration | ✅ OK |
| T6, T22 | `composition` | unit | unit | ✅ OK |
| T7 | `config` | unit | unit | ✅ OK |
| T8–T12, T14 | `platform/jobs/worker` | integration + concorrência | integration (Gate Race) | ✅ OK |
| T13 | `cmd/worker` | integration | integration | ✅ OK |
| T15–T18 | `platform/outbox` | integration + concorrência | integration (Gate Race) | ✅ OK |
| T19–T21 | agendamentos, limpeza, alertas | integration | integration | ✅ OK |
| T23, T24 | admin + handlers | integration (+ Contract) | integration | ✅ OK |
| T25 | ensaio de migrations | integration | integration | ✅ OK |
| T26 | empacotamento | build | none (Gate Build) | ✅ OK |
| T27 | documentação | none | none | ✅ OK |

## Requirement Coverage

| Requisito | Tarefas |
| --- | --- |
| JOB-01 | T3, T4, T5, T25 |
| LEA-01 | T8, T9, T14 |
| RTY-01 | T9, T20 |
| REC-01 | T4, T9, T11, T12, T14, T21, T27 |
| OUT-01 | T1, T15, T18 |
| INB-01 | T16, T18 |
| SVX-01 | T4, T5, T6, T10, T13, T15, T17, T23 |
| DLQ-01 | T4, T22, T23, T24 |
| SCH-01 | T19 |
| RET-01 | T20 |
| OBS-01 | T7, T11, T13, T23, T27 |
| ARQ-01 | T1, T2, T13 |
