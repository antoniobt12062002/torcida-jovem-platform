# Estoque (Web) — Specification

## Problem Statement

O `estoque` V1 está em `develop` com 6 operações HTTP: produtos (SKU), movimentações (entrada, saída, devolução), ajuste com motivo e saldo por produto. O saldo é sempre derivado das movimentações (ADR-007). Nada disso é alcançável sem chamar a API diretamente. Esta sub-spec entrega as telas, sem regra nova.

## Goals

- [ ] O responsável de estoque executa pela Web todo o ciclo do `estoque` V1.
- [ ] Diretoria e Conselho Fiscal consultam produtos, movimentações e saldo sem ver ações de escrita.

## Out of Scope

| Feature | Reason |
| --- | --- |
| Saldo na listagem de produtos | WEB-D-015: sem saldo em lote na API. |
| Editar produto | Não existe endpoint (EST-D-007). |
| Origens `VENDA`, `COMPRA`, `EVENTO` como opção manual | WEB-D-007: reservadas para integrações. |
| Nome do responsável | WEB-D-014. |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Detalhe do produto | Rota `/estoque/produtos/[id]` que lê o produto do cache da lista e busca saldo e movimentações | Não há `GET /produtos/{id}` | y |
| Origem nas movimentações manuais | Entrada, saída e devolução enviam `origem = INVENTARIO`, exibida e não editável | WEB-D-007 | y |
| Ajuste | Só pelo fluxo de ajuste, com `AJUSTE_MANUAL` implícito na API, quantidade com sinal e diferente de zero, motivo obrigatório | WEB-D-007 e contrato de `CreateAjusteRequest` | y |
| Devolução | Iniciada a partir de uma movimentação `ENTRADA` ou `SAIDA` do histórico, que vira `movimentacao_de_id` | A API recusa referência a `AJUSTE` ou `DEVOLUCAO` (`devolucao_invalida`) | y |
| Busca de produtos | Por código ou nome, no cliente | WEB-D-016 | y |

**Open questions:** none.

---

## User Stories

### EWB-01: Produtos e saldo

**User Story**: Como responsável de estoque, quero cadastrar produtos e ver o saldo de cada um.

**Acceptance Criteria**:

1. WHEN uma pessoa com `estoque:produto:read` abre `/estoque/produtos` THEN o sistema SHALL chamar `GET /api/v1/estoque/produtos` e listar código, nome e unidade de medida, com busca por código ou nome no cliente.
2. WHEN uma pessoa com `estoque:produto:create` envia código, nome e unidade de medida THEN o sistema SHALL chamar `POST /api/v1/estoque/produtos` e mostrar o produto na lista.
3. IF a API responde `409 codigo_duplicado` THEN o sistema SHALL mostrar o erro no campo de código; IF responde `422 validation_failed` THEN SHALL mostrar cada erro no campo correspondente.
4. WHEN uma pessoa com `estoque:saldo:read` abre o detalhe de um produto THEN o sistema SHALL chamar `GET /api/v1/estoque/produtos/{id}/saldo` e mostrar o saldo com a unidade de medida, destacando saldo negativo.

**Independent Test**: com MSW, criar produto, receber `codigo_duplicado`, abrir detalhe com saldo negativo.

---

### EWB-02: Movimentações

**User Story**: Como responsável de estoque, quero registrar entradas, saídas e devoluções e ver o histórico de cada produto.

**Acceptance Criteria**:

1. WHEN uma pessoa com `estoque:movimentacao:read` abre o detalhe de um produto THEN o sistema SHALL chamar `GET /api/v1/estoque/produtos/{id}/movimentacoes` e mostrar tipo, quantidade, origem, motivo, referência de devolução, responsável e data, na ordem devolvida pela API.
2. WHEN uma pessoa com `estoque:movimentacao:create` registra uma entrada ou saída com quantidade inteira positiva THEN o sistema SHALL chamar `POST /api/v1/estoque/movimentacoes` com `tipo`, `produto_id`, `quantidade` e `origem = INVENTARIO`.
3. WHEN uma pessoa com `estoque:movimentacao:create` registra a devolução de uma movimentação `ENTRADA` ou `SAIDA` do histórico THEN o sistema SHALL chamar `POST /api/v1/estoque/movimentacoes` com `tipo = DEVOLUCAO`, `movimentacao_de_id` daquela movimentação e `origem = INVENTARIO`.
4. The formulário de movimentação manual SHALL nunca oferecer as origens `VENDA`, `COMPRA`, `EVENTO` ou `AJUSTE_MANUAL`.
5. IF a API responde `409 saldo_insuficiente`, `404 produto_nao_encontrado`, `422 quantidade_invalida` ou `422 devolucao_invalida` THEN o sistema SHALL mostrar a mensagem correspondente.
6. WHEN uma movimentação é registrada THEN o sistema SHALL invalidar o histórico e o saldo daquele produto.

**Independent Test**: com MSW, entrada de 10, saída de 4, devolução da saída, saída com `saldo_insuficiente`.

---

### EWB-03: Ajuste

**User Story**: Como responsável de estoque, quero corrigir o saldo com um ajuste justificado.

**Acceptance Criteria**:

1. WHEN uma pessoa com `estoque:movimentacao:adjust` confirma um ajuste com quantidade inteira diferente de zero e motivo THEN o sistema SHALL chamar `POST /api/v1/estoque/ajustes` com `produto_id`, `quantidade` e `motivo`.
2. IF o motivo está vazio ou só com espaços, ou a quantidade é zero THEN o sistema SHALL manter a confirmação desabilitada e não SHALL chamar a API.
3. The confirmação SHALL mostrar o saldo atual e o saldo resultante, avisando quando o resultado for negativo, sem impedir o ajuste.
4. IF a API responde `422 motivo_obrigatorio` ou `422 quantidade_invalida` THEN o sistema SHALL mostrar a mensagem correspondente.

**Independent Test**: com MSW, ajuste de −3 num produto com saldo 0 (aviso de negativo, sucesso); ajuste sem motivo (API não chamada).

---

### EWB-04: Consulta somente leitura

**User Story**: Como diretoria ou conselho fiscal, quero consultar o estoque sem risco de alterar nada.

**Acceptance Criteria**:

1. WHILE a pessoa tem só `estoque:produto:read`, `estoque:movimentacao:read` e `estoque:saldo:read` THE sistema SHALL mostrar produtos, histórico e saldo e não SHALL mostrar criar produto, entrada, saída, devolução nem ajuste.
2. IF a pessoa não tem `estoque:movimentacao:read` ou `estoque:saldo:read` THEN o detalhe do produto SHALL omitir a seção correspondente, sem chamar a API dela.

**Independent Test**: com MSW, sessão com as três permissões de leitura e sessão só com `estoque:produto:read`.

---

## Edge Cases

- WHEN o detalhe é aberto por endereço de um id que não está na lista THEN o sistema SHALL mostrar "Produto não encontrado".
- WHEN uma movimentação do histórico é `AJUSTE` ou `DEVOLUCAO` THEN o sistema SHALL não oferecer "Devolver" nela.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| EWB-01 | EWB-01: Produtos e saldo | Tasks | Verified |
| EWB-02 | EWB-02: Movimentações | Tasks | Verified |
| EWB-03 | EWB-03: Ajuste | Tasks | Verified |
| EWB-04 | EWB-04: Consulta somente leitura | Tasks | Verified |

**Coverage:** 4 total, 4 mapped to tasks, 0 unmapped.

---

## Success Criteria

- [ ] Uma pessoa com `ESTOQUE_LOJA` executa pela Web: produto → entrada → saída → devolução → ajuste → saldo.
- [ ] Uma pessoa com `CONSELHO_FISCAL` vê tudo e não vê nenhuma ação de escrita.
