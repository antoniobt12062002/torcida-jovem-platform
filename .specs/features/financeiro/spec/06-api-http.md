# API HTTP — Especificação

Status: Aprovada em 2026-10-03. Sub-spec 6 de `financeiro`. Decisões compartilhadas em `.specs/features/financeiro/STATE.md` (`FIN-D-NNN`). Narrativa em português; critérios de aceite em inglês (EARS).

## Problem Statement

`financeiro` V1 (`01`-`05`) está completo no domínio: 15 casos de uso, RBAC publicado, auditoria, persistência — mas sem nenhuma via de acesso externo. Não existe `financeiro/http/`, nenhum contrato `api/openapi/financeiro.yaml`, nenhuma rota registrada em `httpapi.NewRouter`, nenhum tipo gerado para o frontend. Esta sub-spec expõe os 15 casos de uso já existentes via HTTP, seguindo exatamente os padrões já estabelecidos por `identity` (`AD-012`, `AD-013`, `AD-014`) — não inventa nenhuma regra de negócio nova, nenhuma permissão nova, nenhum caso de uso novo.

## Goals

- Publicar `api/openapi/financeiro.yaml`, cobrindo os 15 casos de uso.
- Implementar `financeiro/http/` (handlers finos sobre os casos de uso, mesmo padrão de `identity/http`).
- Criar `financeiro.New(...)`/`financeiro.Module` (composition root que faltava, mesmo padrão de `identity.New`).
- Integrar com `httpapi.NewRouter` e `cmd/api/main.go`, incluindo a primeira instanciação real de `platform/documents.Service` (já pronta desde `fundacao-documentos`, nunca antes conectada).
- Gerar os tipos TypeScript (`web/lib/api/financeiro.d.ts`), sem implementar nenhuma tela.

## Out of Scope

| Item | Razão |
|---|---|
| Qualquer caso de uso, regra de domínio ou permissão nova | Fora do escopo desta rodada — só expõe o que `01`-`05` já decidiram |
| Paginação/filtro/ordenação de `GET /financeiro/contas` e `GET /financeiro/lancamentos` | `ListarContas`/`ListarLancamentos` não suportam nenhum dos três hoje; adicioná-los exigiria alterar casos de uso existentes — `FIN-D-021`, decisão aberta registrada, não decidida silenciosamente |
| Telas de frontend | Fora do escopo — só os tipos gerados |
| `financeiro` V2, `estoque` e demais módulos | Fora do escopo desta sub-spec |
| Alterar `BuildMatrix`, `FoundationContributions()` ou qualquer permissão já publicada | `05-permissoes` já está completa; esta sub-spec só consome `financeiro.Contribution()` |
| Verificação automática de expiração de URL assinada | Já é risco residual aceito desde `AD-015`/`fundacao-documentos` — fora do escopo desta sub-spec |

## Assumptions & Open Questions

| Question | Chosen default | Rationale | Confirmed |
|---|---|---|---|
| Formato de erro | `application/problem+json`, schema `Problem` já existente em `common.yaml` | Único formato usado em todo o projeto (`identity`, `platform`) | y |
| Formato monetário | Schema `Cents` já existente em `common.yaml` (`x-go-type: money.Cents`) | Já mapeado para o tipo Go real, reutilizável sem nenhuma definição nova | y |
| Autenticação/CSRF | Cookie de sessão (`tj_session`) + `X-CSRF-Token` em toda operação que altera estado — mesmo mecanismo global de `platform/httpx`, não específico de `identity` | `Authn`/`CSRF`/`Origin` já são middlewares de `platform/httpx`, module-agnósticos; nenhuma rota de `financeiro` é pública | y |
| `EditarLancamento` é `PUT` (substituição completa) | `PUT /financeiro/lancamentos/{id}` | `EditarLancamentoInput` exige os 4 campos editáveis sempre — não é um `PATCH` parcial | y |
| `ReceberLancamento`/`PagarLancamento` sem corpo | `POST` sem `requestBody`, mesmo padrão de `deactivateUser`/`reactivateUser` | Os `Input` desses casos de uso só têm `Actor`+`ID` | y |
| `CancelarLancamento` com motivo no corpo | `POST` com `{reason}` no corpo | Mesmo padrão de `revokeAdminMembership`; `motivo` é obrigatório no caso de uso | y |
| `CriarDevolucao` é uma operação própria, não um campo opcional de `CriarLancamento` | `POST /financeiro/lancamentos/devolucoes`, endpoint distinto | São dois estruturas de input Go diferentes (`CriarLancamentoInput` vs. `CriarDevolucaoInput`), casos de uso distintos — a fronteira HTTP espelha a fronteira real do domínio | y |
| `ConsultarSaldo` sem filtro de período/conta | `GET /financeiro/saldo`, sem parâmetros | O caso de uso não aceita nenhum filtro (`ConsultarSaldoInput` só tem `Actor`) | y |
| `money.ErrOutOfRange` → `422 amount_out_of_range` | Mapeamento direto | Já registrado como decisão futura em `.specs/STATE.md` (raiz), "para o primeiro endpoint monetário" — este é o primeiro | y |
| `authz.ErrForbidden` → `403 forbidden` | Reaproveita o mapeamento já usado por `identity/http` | Mesmo sentinel, mesmo código, em todo o projeto | y |
| ETag/cache | Nenhum | Nenhum recurso de `financeiro` é imutável o suficiente para justificar cache — status de lançamento e saldo mudam a qualquer momento | y |
| Paginação/ordenação de listagens de `financeiro` | **Decisão aberta** — ver `FIN-D-021` | `ListarContas`/`ListarLancamentos` não suportam isso hoje; estender os casos de uso está fora do escopo desta rodada | **n — aguardando confirmação** |
| Mapeamento dos 11 sentinels de domínio de `financeiro` para código HTTP | **Decisão técnica proposta** — ver `FIN-D-022` | Diferente de `identity`, os sentinels de `financeiro` são frases em português, não códigos `snake_case` — não podem ser reaproveitados como `code` diretamente | **y, confirmar a tabela proposta** |
| `ErrContaInvalida` → `404` ou `422` | Proposto `422 conta_invalida` | É uma referência de campo inválida (conta inexistente OU inativa), mesma forma de `ErrUnknownRole`/`ErrInvalidEmail` de `identity` (422), não "o recurso que você pediu não existe" (404) | **y, confirmar** |
| Limite de corpo do upload de comprovante | Proposto 11 MiB (`documents.maxSize` de 10 MiB + margem de overhead multipart) na chain dedicada da rota de anexar | `httpx.BodyLimit(limit)` já é parametrizável; o valor exato da margem é a única escolha aberta | **y, confirmar o valor** |
| `platform/documents.Service` sem `STORAGE_ENABLED=true` | A API falha ao iniciar (`cmd/api/main.go`) se `financeiro` for composto sem armazenamento habilitado | `AD-012` exige que toda rota do contrato tenha operação registrada incondicionalmente — não é possível registrar a rota de comprovante "só quando há storage", então a validação deve ser na inicialização, não por requisição | y |

Open questions: `FIN-D-021` (paginação) precisa de decisão do mantenedor antes da implementação. `FIN-D-022` (tabela de erros), o status de `ErrContaInvalida` e o valor exato do limite de upload são decisões técnicas propostas, pendentes de confirmação explícita antes da implementação — ver relatório de entrega.

## User Stories

### API-01: Plano de contas via HTTP

**User Story**: Como Tesouraria, Diretoria ou Conselho Fiscal, quero criar, listar, renomear e desativar contas contábeis pela API, para gerenciar o plano de contas sem acesso direto ao banco.

**Acceptance Criteria**:
1. WHEN an actor holding `financeiro:conta:create` sends `POST /api/v1/financeiro/contas` with a valid body THEN the system SHALL call `CriarConta.Execute` and answer `201` with the created conta.
2. WHEN an actor holding `financeiro:conta:read` sends `GET /api/v1/financeiro/contas` THEN the system SHALL call `ListarContas.Execute` and answer `200` with every conta, exactly as the use case returns them — no pagination, no filter.
3. WHEN an actor holding `financeiro:conta:update` sends `PATCH /api/v1/financeiro/contas/{id}` with a new `nome` THEN the system SHALL call `RenomearConta.Execute` and answer `200` with the renamed conta.
4. WHEN an actor holding `financeiro:conta:deactivate` sends `POST /api/v1/financeiro/contas/{id}/deactivate` THEN the system SHALL call `DesativarConta.Execute` and answer `204`.
5. IF the actor lacks the permission for any of these four operations THEN the system SHALL answer `403 forbidden`, calling the use case regardless so it is the use case's own `Authz.Require` that refuses (same delegation pattern already used by every `financeiro` use case).

**Independent Test**: criar uma conta raiz, criar uma subconta do mesmo tipo, listar as duas, renomear a subconta, desativar a raiz — tudo via HTTP, sem acesso direto ao banco.

---

### API-02: Lançamentos via HTTP

**User Story**: Como Tesouraria, quero criar, editar, devolver e listar lançamentos pela API.

**Acceptance Criteria**:
1. WHEN an actor holding `financeiro:lancamento:create` sends `POST /api/v1/financeiro/lancamentos` with `tipo`, `conta_id`, `valor_bruto_cents`, `taxa_cents` and `forma_pagamento` THEN the system SHALL call `CriarLancamento.Execute` and answer `201` with the created lançamento, including the computed `valor_liquido_cents`.
2. WHEN an actor holding `financeiro:lancamento:update` sends `PUT /api/v1/financeiro/lancamentos/{id}` with `conta_id`, `valor_bruto_cents`, `taxa_cents` and `forma_pagamento` THEN the system SHALL call `EditarLancamento.Execute`, replacing every editable field, and answer `200` with the updated lançamento.
3. WHEN an actor holding `financeiro:lancamento:create` sends `POST /api/v1/financeiro/lancamentos/devolucoes` with `conta_id`, `valor_bruto_cents`, `taxa_cents`, `forma_pagamento` and `devolucao_de_id` THEN the system SHALL call `CriarDevolucao.Execute` and answer `201` with the created despesa.
4. WHEN an actor holding `financeiro:lancamento:read` sends `GET /api/v1/financeiro/lancamentos` THEN the system SHALL call `ListarLancamentos.Execute` and answer `200` with every lançamento, exactly as the use case returns them (including `CANCELADA`) — no pagination, no filter, no guaranteed order beyond what the use case already returns.
5. IF the lançamento's `status` is not `CRIADA` WHEN `PUT /api/v1/financeiro/lancamentos/{id}` is called THEN the system SHALL answer `409 lancamento_imutavel` (`domain.ErrLancamentoImutavel`).

**Independent Test**: criar um lançamento, editá-lo enquanto `CRIADA`, tentar editá-lo depois de outra transição (`409`), criar uma devolução referenciando uma receita recebida, listar e ver todos.

---

### API-03: Workflow de liquidação e cancelamento via HTTP

**User Story**: Como Tesouraria, quero receber, pagar e cancelar lançamentos pela API.

**Acceptance Criteria**:
1. WHEN an actor holding `financeiro:lancamento:receive` sends `POST /api/v1/financeiro/lancamentos/{id}/receive` THEN the system SHALL call `ReceberLancamento.Execute` and answer `204`.
2. WHEN an actor holding `financeiro:lancamento:pay` sends `POST /api/v1/financeiro/lancamentos/{id}/pay` THEN the system SHALL call `PagarLancamento.Execute` and answer `204`.
3. WHEN an actor holding `financeiro:lancamento:cancel` sends `POST /api/v1/financeiro/lancamentos/{id}/cancel` with `{reason}` THEN the system SHALL call `CancelarLancamento.Execute` and answer `204`.
4. IF `reason` is empty or missing WHEN cancelling THEN the system SHALL answer `422 motivo_obrigatorio` (`domain.ErrMotivoObrigatorio`), never calling the use case's persistence step (same pre-check already in `CancelarLancamento.Execute`).
5. IF the lançamento cannot receive/pay/is already cancelled THEN the system SHALL answer `409` with the matching code (`lancamento_nao_pode_ser_recebido`, `lancamento_nao_pode_ser_pago`, `lancamento_ja_cancelado`).

**Independent Test**: criar uma receita, recebê-la via HTTP, cancelá-la com motivo depois de recebida (permitido, `FIN-D-006`) — e ver o saldo refletir a mudança.

---

### API-04: Saldo via HTTP

**User Story**: Como Tesouraria, Diretoria ou Conselho Fiscal, quero consultar o saldo atual pela API.

**Acceptance Criteria**:
1. WHEN an actor holding `financeiro:saldo:read` sends `GET /api/v1/financeiro/saldo` THEN the system SHALL call `ConsultarSaldo.Execute` and answer `200` with `{saldo_cents}`, the exact value the use case returns (zero included, never an error for an empty ledger).

**Independent Test**: consultar o saldo com zero lançamentos (`0`), criar e receber uma receita, consultar de novo e ver o valor líquido refletido.

---

### API-05: Comprovantes via HTTP

**User Story**: Como Tesouraria, quero anexar comprovantes pela API; como Tesouraria, Diretoria ou Conselho Fiscal, quero consultar e listar comprovantes.

**Acceptance Criteria**:
1. WHEN an actor holding `financeiro:comprovante:create` sends `POST /api/v1/financeiro/lancamentos/{id}/comprovantes` as `multipart/form-data` with a `file` field THEN the system SHALL call `AnexarComprovante.Execute`, delegating entirely to `platform/documents.Service.Store`, and answer `201` with the stored document's metadata.
2. IF the lançamento does not exist THEN the system SHALL answer `404 lancamento_nao_encontrado` before any call to `documents.Service.Store` (same guarantee already in `AnexarComprovante.Execute`).
3. WHEN an actor holding `financeiro:comprovante:read` sends `GET /api/v1/financeiro/comprovantes/{documentId}/url` THEN the system SHALL call `ConsultarComprovante.Execute`, delegating entirely to `documents.Service.AccessURL`, and answer `200` with `{url}`, with `Cache-Control: no-store`.
4. WHEN an actor holding `financeiro:comprovante:read` sends `GET /api/v1/financeiro/lancamentos/{id}/comprovantes` THEN the system SHALL call `ListarComprovantes.Execute`, delegating entirely to `documents.Service.ListByOwner`, and answer `200` with the list, in whatever order the use case already returns (newest first).
5. The system SHALL NOT reimplement any storage, versioning, file-type or file-size rule already enforced by `platform/documents` — every such error (`document_extension_not_allowed`, `document_type_not_allowed`, `document_type_mismatch`, `document_too_large`, `document_not_found`) SHALL be mapped to HTTP exactly as `platform/documents` names it, with no new validation added by `financeiro`.

**Independent Test**: anexar um comprovante a um lançamento, consultar sua URL assinada e baixar o conteúdo, listar e ver o único comprovante — tudo via HTTP, sem nenhum acesso direto a `platform/documents` fora da composição de `financeiro`.

---

### API-06: Tratamento de erros uniforme

**User Story**: Como consumidor da API (frontend ou qualquer cliente), quero que todo erro de `financeiro` siga o mesmo formato `problem+json` já usado pelo resto da API.

**Acceptance Criteria**:
1. WHEN any `financeiro` use case returns one of its 11 domain sentinels THEN the system SHALL answer with the status and `code` of the mapping table in `design/06-api-http.md` (`FIN-D-022`), never a generic `500`.
2. WHEN any `financeiro` use case returns `authz.ErrForbidden` THEN the system SHALL answer `403 forbidden`, the same mapping `identity/http` already uses.
3. WHEN any `financeiro` use case returns `money.ErrOutOfRange` THEN the system SHALL answer `422 amount_out_of_range`.
4. WHEN any `platform/documents` sentinel reaches the `financeiro` HTTP layer THEN the system SHALL map it using the sentinel's own `.Error()` string as the `code` (already `snake_case`, e.g. `document_too_large`), never redefining it.
5. The system SHALL NOT introduce a new error-handling mechanism: the mapping function follows exactly the shape of `identity/http/handler.go`'s `writeError`/`sentinels` table.

**Independent Test**: provocar cada um dos 11 sentinels de domínio de `financeiro` (ex. renomear uma conta já usada, cancelar duas vezes, criar lançamento em conta inativa) e confirmar o `status`/`code` exatos da tabela.

## Edge Cases

- Ator sem permissão para qualquer operação: a chamada chega ao caso de uso, que recusa via `Authz.Require` — a camada HTTP nunca verifica a permissão por conta própria (mesmo padrão de `identity/http`, nenhuma duplicação de checagem).
- `PUT` de um lançamento que não existe: `404 lancamento_nao_encontrado`.
- `GET /financeiro/saldo` com zero lançamentos: `200 {"saldo_cents": 0}`, nunca um erro.
- Upload de comprovante maior que o limite da chain dedicada: `413 payload_too_large` (limite genérico de corpo) ou `413 document_too_large` (limite de `platform/documents`), dependendo de qual dos dois limites for excedido primeiro — ambos já existem, nenhum novo.
- Cancelamento de lançamento já `RECEBIDA`/`PAGA`: aceito (`FIN-D-006`), nunca bloqueado pela camada HTTP.

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
|---|---|---|---|
| API-01 | Plano de contas via HTTP | In Design | Not Started |
| API-02 | Lançamentos via HTTP | In Design | Not Started |
| API-03 | Workflow via HTTP | In Design | Not Started |
| API-04 | Saldo via HTTP | In Design | Not Started |
| API-05 | Comprovantes via HTTP | In Design | Not Started |
| API-06 | Tratamento de erros uniforme | In Design | Not Started |

**Coverage:** 6 total, 6 mapped to tasks, 0 unmapped.

## Success Criteria

- Todo caso de uso de `01`-`05` é alcançável via HTTP, com a mesma regra de negócio, sem nenhuma regra nova.
- `httpapi.NewRouter` continua falhando se rota e contrato divergirem (`AD-012`), agora incluindo `financeiro`.
- Nenhuma regra de `platform/documents` é reimplementada.
- `go test ./... -tags=integration`, `golangci-lint`, teste de fronteiras e `validate_spec.py`/`validate_tasks.py` continuam verdes.
