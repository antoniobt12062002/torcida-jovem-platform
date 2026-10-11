# Financeiro (Web) — Specification

## Problem Statement

O `financeiro` V1 está em `develop` com 15 operações HTTP: plano de contas, lançamentos com workflow (`CRIADA` → `RECEBIDA`/`PAGA`, cancelamento com motivo), devolução, saldo e comprovantes. Nada disso é alcançável sem chamar a API diretamente. Esta sub-spec entrega as telas do ciclo completo, sem nenhuma regra nova: a API decide, o front conduz e explica.

## Goals

- [ ] A tesouraria executa pela Web todo o ciclo do `financeiro` V1.
- [ ] Diretoria e Conselho Fiscal consultam contas, lançamentos, saldo e comprovantes sem ver ações de escrita.

## Out of Scope

| Feature | Reason |
| --- | --- |
| Prestação de contas e parecer | Sem endpoint (STATE, "Fora do V1"). |
| Relatórios, exportações, gráficos | Sem endpoint nem regra (STATE, "Fora do V1"). |
| Paginação e filtro no servidor | FIN-D-021 / WEB-D-016: filtro e ordenação no cliente. |
| Nome de quem criou ou cancelou | WEB-D-014: "você" ou identificador curto. |
| Remover comprovante | Não existe endpoint; documentos são versionados (ADR-006). |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Detalhe do lançamento | Rota `/financeiro/lancamentos/[id]` que lê o item do cache da lista (`GET /lancamentos`) | Não há `GET /lancamentos/{id}`; a lista é completa (WEB-D-016) | y |
| Filtros da lista de lançamentos | Tipo, status, conta e período de criação, no cliente; ordem do mais novo para o mais antigo | WEB-D-016 | y |
| Plano de contas | Árvore por `parent_id`, com contas inativas visíveis e marcadas | Contas nunca são excluídas (AD-007) | y |
| Contas oferecidas num lançamento | Só contas ativas cujo `tipo` é o do lançamento; a API decide no fim (`conta_invalida`, `lancamento_tipo_incompativel`) | Evita erro previsível sem criar regra | y |
| Valor líquido | Exibido como prévia (bruto − taxa) durante a digitação; o valor gravado é o devolvido pela API | A API calcula `valor_liquido_cents` | y |
| Tipos de arquivo de comprovante | Seletor sugere `.pdf`, `.jpg`, `.jpeg`, `.png`, `.webp` e recusa antes do envio arquivos acima de 10 MiB − 64 KiB (10.420.224 bytes); a API decide tipo e limite (10 MiB) | Lista e limite de `platform/documents`; margem para o envelope `multipart/form-data`, porque o rewrite do Next trunca corpos acima de cerca de 10 MiB (WEB-D-009, correção de 2026-10-04) | y |
| Atualização do saldo | Toda escrita que muda status (receber, pagar, cancelar) invalida `queryKeys.financeiro.saldo` | WEB-D-010 | y |

**Open questions:** none.

---

## User Stories

### FWB-01: Plano de contas

**User Story**: Como tesouraria, quero manter o plano de contas; como diretoria ou conselho fiscal, quero consultá-lo.

**Acceptance Criteria**:

1. WHEN uma pessoa com `financeiro:conta:read` abre `/financeiro/contas` THEN o sistema SHALL chamar `GET /api/v1/financeiro/contas` e mostrar as contas em árvore por `parent_id`, com tipo e situação.
2. WHEN uma pessoa com `financeiro:conta:create` cria uma conta raiz ou subconta THEN o sistema SHALL chamar `POST /api/v1/financeiro/contas` com `tipo`, `nome` e, para subconta, `parent_id`, e mostrar a conta na árvore.
3. WHEN a criação é de subconta THEN o formulário SHALL fixar o `tipo` igual ao da conta pai.
4. WHEN uma pessoa com `financeiro:conta:update` renomeia uma conta THEN o sistema SHALL chamar `PATCH /api/v1/financeiro/contas/{id}`; IF a API responde `409 conta_ja_utilizada` THEN o sistema SHALL explicar que contas já usadas em lançamento não podem ser renomeadas.
5. WHEN uma pessoa com `financeiro:conta:deactivate` confirma a desativação THEN o sistema SHALL chamar `POST /api/v1/financeiro/contas/{id}/deactivate` e mostrar a conta como inativa.
6. IF a API responde `conta_nao_encontrada` ou `conta_tipo_incompativel` THEN o sistema SHALL mostrar a mensagem correspondente.

**Independent Test**: com MSW, árvore com raiz, subconta e conta inativa; criar subconta; renomear com `conta_ja_utilizada`.

---

### FWB-02: Saldo

**User Story**: Como tesouraria, diretoria ou conselho fiscal, quero ver o saldo atual.

**Acceptance Criteria**:

1. WHEN uma pessoa com `financeiro:saldo:read` abre `/financeiro/saldo` THEN o sistema SHALL chamar `GET /api/v1/financeiro/saldo` e mostrar `saldo_cents` com `formatBRL`, inclusive zero e negativo.
2. The página SHALL explicar que o saldo é a soma dos lançamentos recebidos menos a dos pagos, pelo status atual.

**Independent Test**: com MSW, saldos positivo, zero e negativo.

---

### FWB-03: Lançamentos

**User Story**: Como tesouraria, quero registrar receitas e despesas e corrigi-las enquanto ainda estão em aberto.

**Acceptance Criteria**:

1. WHEN uma pessoa com `financeiro:lancamento:read` abre `/financeiro/lancamentos` THEN o sistema SHALL chamar `GET /api/v1/financeiro/lancamentos` e mostrar tipo, conta, valor bruto, taxa, valor líquido, forma de pagamento, status e data de criação, com filtros de tipo, status, conta e período aplicados no cliente.
2. WHEN uma pessoa com `financeiro:lancamento:create` envia tipo, conta, valor bruto, taxa e forma de pagamento THEN o sistema SHALL chamar `POST /api/v1/financeiro/lancamentos` com os valores em centavos inteiros.
3. WHEN uma pessoa com `financeiro:lancamento:update` edita um lançamento `CRIADA` THEN o sistema SHALL chamar `PUT /api/v1/financeiro/lancamentos/{id}` com conta, valor bruto, taxa e forma de pagamento.
4. The ação de editar SHALL aparecer só para lançamentos `CRIADA`; IF a API responde `409 lancamento_imutavel` THEN o sistema SHALL explicar que só lançamentos em aberto podem ser editados e atualizar a lista.
5. IF a API responde `conta_invalida`, `lancamento_tipo_incompativel` ou `amount_out_of_range` THEN o sistema SHALL mostrar a mensagem junto do campo correspondente.
6. WHEN uma pessoa abre o detalhe de um lançamento THEN o sistema SHALL mostrar todos os campos, inclusive motivo, autor e data de cancelamento quando houver, e a referência à receita devolvida quando for devolução.

**Independent Test**: com MSW, criar receita, editar enquanto `CRIADA`, tentar editar `RECEBIDA`, filtrar por status.

---

### FWB-04: Workflow, cancelamento e devolução

**User Story**: Como tesouraria, quero marcar receitas como recebidas e despesas como pagas, cancelar com motivo e registrar devoluções.

**Acceptance Criteria**:

1. WHEN uma pessoa com `financeiro:lancamento:receive` confirma o recebimento de uma receita `CRIADA` THEN o sistema SHALL chamar `POST /api/v1/financeiro/lancamentos/{id}/receive`, mostrar `RECEBIDA` e atualizar o saldo.
2. WHEN uma pessoa com `financeiro:lancamento:pay` confirma o pagamento de uma despesa `CRIADA` THEN o sistema SHALL chamar `POST /api/v1/financeiro/lancamentos/{id}/pay`, mostrar `PAGA` e atualizar o saldo.
3. WHEN uma pessoa com `financeiro:lancamento:cancel` confirma o cancelamento com motivo de um lançamento não cancelado THEN o sistema SHALL chamar `POST /api/v1/financeiro/lancamentos/{id}/cancel` com `reason` e mostrar `CANCELADA` com o motivo.
4. IF o motivo do cancelamento está vazio ou só com espaços THEN o sistema SHALL manter a confirmação desabilitada e não SHALL chamar a API.
5. WHEN uma pessoa com `financeiro:lancamento:create` registra a devolução de uma receita `RECEBIDA` THEN o sistema SHALL chamar `POST /api/v1/financeiro/lancamentos/devolucoes` com `devolucao_de_id` daquela receita, conta de despesa, valor bruto, taxa e forma de pagamento.
6. The ação de devolução SHALL aparecer só em receitas `RECEBIDA`; IF a API responde `devolucao_invalida` THEN o sistema SHALL mostrar a mensagem correspondente.
7. The sistema SHALL oferecer receber só em receitas `CRIADA`, pagar só em despesas `CRIADA` e cancelar só em lançamentos não cancelados; IF a API responde `lancamento_nao_pode_ser_recebido`, `lancamento_nao_pode_ser_pago` ou `lancamento_ja_cancelado` THEN o sistema SHALL mostrar a mensagem e atualizar a lista.

**Independent Test**: com MSW, receber, pagar, cancelar sem e com motivo, devolver uma receita recebida.

---

### FWB-05: Comprovantes

**User Story**: Como tesouraria, quero anexar comprovantes a um lançamento; como diretoria ou conselho fiscal, quero consultá-los.

**Acceptance Criteria**:

1. WHEN uma pessoa com `financeiro:comprovante:read` abre o detalhe de um lançamento THEN o sistema SHALL chamar `GET /api/v1/financeiro/lancamentos/{id}/comprovantes` e listar nome, tipo, tamanho, versão, autor e data, do mais novo para o mais antigo.
2. WHEN uma pessoa com `financeiro:comprovante:create` escolhe um arquivo THEN o sistema SHALL enviá-lo como `multipart/form-data` no campo `file` para `POST /api/v1/financeiro/lancamentos/{id}/comprovantes`, mostrar o progresso como "enviando" e, em `201`, atualizar a lista.
3. IF o arquivo escolhido tem mais de 10.420.224 bytes (10 MiB − 64 KiB, limite efetivo do cliente) THEN o sistema SHALL informar que o arquivo excede o limite permitido e não SHALL chamar a API; WHEN o arquivo tem até 10.420.224 bytes THEN o sistema SHALL enviá-lo. Essa margem é mitigação da camada Web (WEB-D-009); o limite da API continua 10 MiB.
4. IF a API responde `413` (`payload_too_large` ou `document_too_large`) ou `422` (`document_extension_not_allowed`, `document_type_not_allowed`, `document_type_mismatch`) THEN o sistema SHALL mostrar a mensagem correspondente.
5. WHEN uma pessoa com `financeiro:comprovante:read` pede para baixar THEN o sistema SHALL chamar `GET /api/v1/financeiro/comprovantes/{documentId}/url` no momento do clique e abrir a URL assinada, sem guardá-la no cache de consultas.
6. IF a API responde `404 document_not_found` ou `lancamento_nao_encontrado` THEN o sistema SHALL mostrar a mensagem correspondente.

**Independent Test**: com MSW, anexar um PDF, aceitar um arquivo de exatamente 10.420.224 bytes, recusar um de 10.420.225 bytes sem nenhuma requisição, mostrar a mensagem do `413 document_too_large` devolvido pela API, baixar pela URL assinada.

---

## Edge Cases

- WHEN a pessoa tem só permissões de leitura THEN nenhuma ação de escrita (criar, editar, receber, pagar, cancelar, devolver, anexar, renomear, desativar) SHALL aparecer em nenhuma tela do financeiro.
- WHEN o detalhe é aberto por endereço de um id que não está na lista THEN o sistema SHALL mostrar "Lançamento não encontrado".
- WHEN um lançamento muda de status THEN o sistema SHALL invalidar a lista de lançamentos e o saldo.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| FWB-01 | FWB-01: Plano de contas | Tasks | Verified |
| FWB-02 | FWB-02: Saldo | Tasks | Verified |
| FWB-03 | FWB-03: Lançamentos | Tasks | Verified |
| FWB-04 | FWB-04: Workflow, cancelamento e devolução | Tasks | Verified |
| FWB-05 | FWB-05: Comprovantes | Tasks | Verified |

**Coverage:** 5 total, 5 mapped to tasks, 0 unmapped.

---

## Success Criteria

- [ ] Uma pessoa com `TESOURARIA` executa pela Web o ciclo: conta → receita → recebimento → comprovante → devolução → cancelamento → saldo.
- [ ] Uma pessoa com `CONSELHO_FISCAL` vê tudo e não vê nenhuma ação de escrita.
