# API HTTP Specification

## Problem Statement

`01`-`04` implementam os 6 casos de uso de `estoque` e o RBAC que os protege, mas nenhum deles é alcançável fora do processo Go. Esta sub-spec expõe os 6 via HTTP, contract-first (OpenAPI), reaproveitando integralmente os padrões já consolidados em `financeiro/06-api-http` — composition root, mapeamento de erro por tabela explícita, paridade rota×contrato (`AD-012`).

## Goals

- [ ] Os 6 casos de uso (`CriarProduto`, `ListarProdutos`, `RegistrarMovimentacao`, `ListarMovimentacoes`, `AjustarEstoque`, `ConsultarSaldo`) são alcançáveis via HTTP, com a mesma regra de negócio, sem nenhuma regra nova.
- [ ] `httpapi.NewRouter` continua falhando se rota e contrato divergirem (`AD-012`), agora incluindo `estoque`.

## Out of Scope

| Feature | Reason |
| --- | --- |
| Paginação/filtro/ordenação em qualquer listagem | Decisão do mantenedor, mesma lógica de `FIN-D-021`: nenhum caso de uso de `01`-`03` suporta isso; `estoque` expõe fielmente o que existe. |
| Qualquer regra de negócio nova na camada HTTP | Handlers são finos — leem a sessão, chamam o caso de uso, devolvem a resposta (`AD-014`). |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Forma do corpo de `POST /estoque/movimentacoes` | Um único schema com `tipo` (`ENTRADA`\|`SAIDA`\|`DEVOLUCAO`) discriminando o payload; `movimentacao_de_id` só é exigido quando `tipo=DEVOLUCAO` | Decisão do mantenedor (EST-D-005): endpoint único com `tipo`, em vez de `/entradas`/`/saidas` separados. | y |
| Verbo HTTP de todas as operações de escrita | `POST` — nenhuma tem `PUT`/`PATCH`, porque nenhuma edita um registro existente (toda escrita de `estoque` é a criação de um novo registro append-only) | Reflexo direto de `AD-009`: não existe "editar movimentação" nem "editar SKU" (`EST-D-007`) — não há verbo de atualização para nenhum recurso desta sub-spec. | y |

**Open questions:** none.

---

## User Stories

### API-01: Produtos (SKU) via HTTP

**User Story**: Como responsável de estoque, diretoria ou conselho fiscal, quero criar e listar SKUs pela API.

**Acceptance Criteria**:

1. WHEN um ator com `estoque:produto:create` envia `POST /api/v1/estoque/produtos` com um corpo válido THEN o sistema SHALL chamar `CriarProduto.Execute` e responder `201` com o SKU criado.
2. WHEN um ator com `estoque:produto:read` envia `GET /api/v1/estoque/produtos` THEN o sistema SHALL chamar `ListarProdutos.Execute` e responder `200` com todos os SKUs, sem paginação.
3. IF o ator não tem a permissão de qualquer uma das duas operações THEN o sistema SHALL responder `403 forbidden`, chamando o caso de uso mesmo assim (delegação, mesmo padrão de `financeiro`).

**Independent Test**: criar um SKU via HTTP, listar e ver o criado.

---

### API-02: Movimentações via HTTP

**User Story**: Como responsável de estoque, quero registrar entrada, saída e devolução, e listar o histórico de um SKU, pela API.

**Acceptance Criteria**:

1. WHEN um ator com `estoque:movimentacao:create` envia `POST /api/v1/estoque/movimentacoes` com `tipo`, `produto_id`, `quantidade`, `origem` (e `movimentacao_de_id` quando `tipo=DEVOLUCAO`) THEN o sistema SHALL chamar `RegistrarMovimentacao.Execute` e responder `201` com a movimentação criada.
2. WHEN um ator com `estoque:movimentacao:read` envia `GET /api/v1/estoque/produtos/{id}/movimentacoes` THEN o sistema SHALL chamar `ListarMovimentacoes.Execute` e responder `200` com todas as movimentações daquele SKU, na ordem em que ocorreram.
3. IF o saldo for insuficiente para uma `SAIDA` ou para uma `DEVOLUCAO` que reduz saldo THEN o sistema SHALL responder `409 saldo_insuficiente`.
4. IF o `produto_id` não existir THEN o sistema SHALL responder `404 produto_nao_encontrado`.
5. IF `movimentacao_de_id` for inválido para uma `DEVOLUCAO` (não existe, outro SKU, ou referencia `AJUSTE`/`DEVOLUCAO`) THEN o sistema SHALL responder `422 devolucao_invalida`.

**Independent Test**: criar um SKU, entrar com 10 via HTTP, sair com 4, devolver 2 da saída, listar e ver as 3 movimentações.

---

### API-03: Ajustes e saldo via HTTP

**User Story**: Como responsável de estoque, quero registrar um ajuste e consultar o saldo atual, pela API.

**Acceptance Criteria**:

1. WHEN um ator com `estoque:movimentacao:adjust` envia `POST /api/v1/estoque/ajustes` com `produto_id`, `quantidade` (diferente de zero) e `motivo` THEN o sistema SHALL chamar `AjustarEstoque.Execute` e responder `201`, mesmo que o saldo resultante seja negativo.
2. IF `motivo` for vazio ou ausente THEN o sistema SHALL responder `422 motivo_obrigatorio`, nunca chamando a persistência do caso de uso (mesma ordem já garantida por `AjustarEstoque.Execute`).
3. WHEN um ator com `estoque:saldo:read` envia `GET /api/v1/estoque/produtos/{id}/saldo` THEN o sistema SHALL chamar `ConsultarSaldo.Execute` e responder `200` com `{saldo}`, o valor exato que o caso de uso devolve (zero e negativo incluídos).

**Independent Test**: ajustar um SKU em `-3` sem nenhuma movimentação anterior (sucesso, saldo `-3`), consultar e confirmar.

---

### API-04: Tratamento de erros uniforme

**User Story**: Como consumidor da API, quero que todo erro de domínio de `estoque` vire uma resposta `problem+json` previsível, no mesmo formato já usado pelos demais módulos.

**Acceptance Criteria**:

1. The system SHALL mapear cada sentinel de `estoque/domain` a um status e `code` fixos, nunca reaproveitando o `.Error()` do sentinel como `code` (mesmo motivo de `FIN-D-022`: sentinels de domínio são frases em português).
2. `produto_nao_encontrado` SHALL responder `404`; `codigo_duplicado` SHALL responder `409`; `quantidade_invalida`, `motivo_obrigatorio` e `devolucao_invalida` SHALL responder `422`; `saldo_insuficiente` SHALL responder `409`; `forbidden` SHALL responder `403` — tabela completa fechada pelo mantenedor (EST-D-006).

**Independent Test**: provocar cada um dos 7 erros e confirmar o status e `code` exatos.

---

## Edge Cases

- IF o corpo de `POST /estoque/movimentacoes` tiver `tipo=DEVOLUCAO` sem `movimentacao_de_id` THEN o sistema SHALL recusar na validação de forma do contrato (campo obrigatório condicional), antes de qualquer chamada ao caso de uso — a ser detalhado no design (a estratégia depende de como o contrato OpenAPI expressa campo condicional, ver `design/05-api-http.md`).

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| API-01 | API-01: Produtos via HTTP | Implementing | Verified |
| API-02 | API-02: Movimentações via HTTP | Implementing | Verified |
| API-03 | API-03: Ajustes e saldo via HTTP | Implementing | Verified |
| API-04 | API-04: Tratamento de erros uniforme | Implementing | Verified |

**Coverage:** 4 total, 4 mapped to tasks, 0 unmapped.

---

## Success Criteria

- [ ] Todo caso de uso de `01`-`04` é alcançável via HTTP, com a mesma regra de negócio, sem nenhuma regra nova.
- [ ] `httpapi.NewRouter` continua falhando se rota e contrato de `estoque` divergirem.
