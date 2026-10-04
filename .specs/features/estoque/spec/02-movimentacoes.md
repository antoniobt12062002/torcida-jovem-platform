# Movimentações de Estoque Specification

## Problem Statement

`AD-009`/`ADR-007` já decidiram que o saldo de um SKU é calculado a partir de um razão de movimentações imutáveis, nunca uma coluna própria. Esta sub-spec implementa o registro dessas movimentações — entrada, saída e devolução — e sua listagem. Ajuste (a correção excepcional que pode levar o saldo a negativo) é uma sub-spec própria (`03-ajustes-e-saldo`), porque tem permissão e regra de bloqueio diferentes.

## Goals

- [ ] Toda entrada, saída e devolução é um registro append-only, nunca editado.
- [ ] Uma saída nunca resulta em saldo negativo.
- [ ] Uma devolução sempre referencia a movimentação original, sem jamais alterá-la.

## Out of Scope

| Feature | Reason |
| --- | --- |
| Ajuste de estoque | Sub-spec própria (`03-ajustes-e-saldo`) — permissão (`estoque:movimentacao:adjust`) e regra de bloqueio diferentes (pode ir a negativo). |
| Consulta de saldo | Sub-spec própria (`03-ajustes-e-saldo`) — mesma decisão de `financeiro` (`FIN-D-013`): saldo é uma permissão e consulta com regra própria, não parte da sub-spec de escrita. |
| Reserva de estoque para venda pendente | Decisão do mantenedor (EST-D-002) — fora do V1; responsabilidade de uma futura `loja`. |
| Integração real com `financeiro` (compra→entrada) ou `loja` (venda→saída) | Decisão do mantenedor — nenhum dos dois módulos do outro lado existe ainda; `origem` registra o valor (`VENDA`, `COMPRA`, `EVENTO`, `INVENTARIO`, `AJUSTE_MANUAL`) como campo informativo, mas nenhuma chamada real cruza módulos nesta sub-spec. |
| Edição ou exclusão de uma movimentação já registrada | `AD-009`/`ADR-007`: append-only; correção é sempre uma nova movimentação. |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Sinal da devolução | A devolução inverte o sentido da movimentação referenciada: devolver uma `SAIDA` repõe estoque (soma positiva, nunca bloqueada); devolver uma `ENTRADA` remove estoque (soma negativa, sujeita à mesma regra de bloqueio de saldo negativo que `SAIDA`) | Mantém uma única regra de saldo negativo — bloqueia quando o efeito líquido é retirar, nunca quando é repor — sem introduzir o conceito de "devolução que nunca falha" nem duplicar a lógica de bloqueio por tipo. Nenhum documento define a direção explicitamente (`ADR-007` só lista devolução como um dos 4 tipos, sem detalhar o sinal). | y (registrada como EST-D-008, lógica, não de negócio nova) |
| O que uma devolução pode referenciar | Só uma movimentação `ENTRADA` ou `SAIDA` — nunca outra `DEVOLUCAO` nem um `AJUSTE` | Evita cadeias de devolução-de-devolução e evita devolver um ajuste (que já é a ferramenta de correção). Mesmo espírito de `FIN-D-003` em `financeiro`, que restringe o que uma devolução pode referenciar. | y (EST-D-009) |
| Quantidade de uma devolução é igual à da movimentação original? | Não — a devolução recebe sua própria quantidade (positiva, "quantos estão sendo devolvidos"), independente da quantidade original | Permite devolução parcial (ex. devolver 2 de uma saída de 5) sem exigir um caso de uso de devolução parcial separado. Mesmo padrão de `financeiro`: `CriarDevolucaoInput` tem seu próprio valor, independente do lançamento referenciado. | y |
| Quantidade de `ENTRADA`/`SAIDA`/`DEVOLUCAO` no corpo da requisição | Sempre um inteiro positivo (`quantidade > 0`); o sinal é aplicado internamente pelo caso de uso conforme o `tipo` | Mais natural numa API HTTP ("remova 5 unidades" em vez de "-5 unidades"); evita o cliente ter que saber a convenção de sinal interna. | y |

**Open questions:** none — todas resolvidas ou registradas acima como decisão lógica derivada das regras já fechadas (nenhuma delas é uma regra de negócio nova, nenhuma contradiz `AD-009`/`ADR-007`/`EST-D-001` a `EST-D-007`).

---

## User Stories

### P1: Registrar entrada ⭐ MVP

**User Story**: Como responsável de estoque, quero registrar uma entrada de um SKU, para que o saldo reflita o que chegou.

**Why P1**: É o caminho mais básico de crescimento de saldo — sem ele, todo saldo começa e permanece zero.

**Acceptance Criteria**:

1. WHEN um ator com `estoque:movimentacao:create` envia `POST /estoque/movimentacoes` com `tipo=ENTRADA`, `produto_id`, `quantidade` (> 0) e `origem` THEN o sistema SHALL registrar a movimentação com quantidade positiva e responder `201` com os dados completos, incluindo o `id` gerado.
2. IF `produto_id` não existe THEN o sistema SHALL recusar com `produto_nao_encontrado`, sem registrar nada.
3. IF `quantidade` for zero ou negativa THEN o sistema SHALL recusar com `quantidade_invalida`.
4. IF o ator não tem `estoque:movimentacao:create` THEN o sistema SHALL recusar com `forbidden`, chamando o caso de uso mesmo assim (mesmo padrão de `financeiro`: é o caso de uso que recusa, não a camada HTTP).

**Independent Test**: criar um SKU, registrar uma entrada de 10 unidades, consultar saldo (fora desta sub-spec, mas verificável pela listagem de movimentações) e ver a movimentação com quantidade `+10`.

---

### P1: Registrar saída ⭐ MVP

**User Story**: Como responsável de estoque, quero registrar uma saída de um SKU, para que o saldo reflita o que saiu — nunca ficando negativo.

**Why P1**: É o mecanismo central da regra de `AD-009` (saldo negativo bloqueado por padrão).

**Acceptance Criteria**:

1. WHEN um ator com `estoque:movimentacao:create` envia `POST /estoque/movimentacoes` com `tipo=SAIDA`, `produto_id`, `quantidade` (> 0) e `origem`, E o saldo atual é suficiente THEN o sistema SHALL registrar a movimentação com quantidade negativa e responder `201`.
2. IF o saldo atual do SKU for menor que a `quantidade` solicitada THEN o sistema SHALL recusar com `saldo_insuficiente` (`409`), sem registrar nada.
3. IF `produto_id` não existe THEN o sistema SHALL recusar com `produto_nao_encontrado`.
4. IF `quantidade` for zero ou negativa THEN o sistema SHALL recusar com `quantidade_invalida`.
5. WHILE duas saídas concorrentes disputam o mesmo saldo (ex. as duas pedem a última unidade) THEN o sistema SHALL garantir, via lock na mesma transação, que no máximo uma delas seja aceita — a outra recebe `saldo_insuficiente`.

**Independent Test**: criar um SKU, entrar com 5 unidades, sair com 5 (sucesso, saldo zero), tentar sair com 1 a mais (recusado); depois, com um teste de integração real, disparar duas saídas concorrentes pedindo a última unidade e confirmar que só uma é aceita.

---

### P1: Registrar devolução ⭐ MVP

**User Story**: Como responsável de estoque, quero registrar uma devolução vinculada a uma entrada ou saída já existente, para corrigir o efeito dela sem jamais editá-la.

**Why P1**: É a terceira forma de movimentação que `ADR-007` já define — sem ela, qualquer devolução teria que ser simulada como ajuste, perdendo o vínculo com a movimentação original.

**Acceptance Criteria**:

1. WHEN um ator com `estoque:movimentacao:create` envia `POST /estoque/movimentacoes` com `tipo=DEVOLUCAO`, `produto_id`, `quantidade` (> 0), `origem` e `movimentacao_de_id` referenciando uma `ENTRADA` ou `SAIDA` existente do mesmo `produto_id` THEN o sistema SHALL registrar a devolução com o sinal invertido da movimentação referenciada e responder `201`.
2. IF `movimentacao_de_id` não existe, referencia outro `produto_id`, ou referencia um `AJUSTE`/outra `DEVOLUCAO` THEN o sistema SHALL recusar com `devolucao_invalida`.
3. IF a devolução invertida resultaria em saldo negativo (devolução de uma `ENTRADA`, com saldo insuficiente) THEN o sistema SHALL recusar com `saldo_insuficiente`, pela mesma regra de `SAIDA`.
4. A movimentação original referenciada SHALL permanecer inalterada após o registro da devolução.

**Independent Test**: registrar uma entrada de 10, uma saída de 4, devolver 2 da saída (sucesso, saldo volta a 8), devolver 10 da entrada original (recusado por saldo insuficiente, pois já saíram/devolveram-se unidades).

---

### P1: Listar movimentações de um SKU ⭐ MVP

**User Story**: Como responsável de estoque, diretoria ou conselho fiscal, quero listar o histórico de movimentações de um SKU, para auditar o que aconteceu.

**Why P1**: É o histórico que `ADR-007` promete como consequência da imutabilidade — sem listagem, a movimentação existe só no banco.

**Acceptance Criteria**:

1. WHEN um ator com `estoque:movimentacao:read` lista as movimentações de um `produto_id` THEN o sistema SHALL devolver todas, na ordem em que ocorreram, sem paginação (mesma decisão de `FIN-D-021`), incluindo ajustes (ainda que `03-ajustes-e-saldo` seja quem os cria).
2. IF o ator não tem `estoque:movimentacao:read` THEN o sistema SHALL recusar com `forbidden`.

**Independent Test**: registrar entrada, saída e devolução de um SKU, listar e ver as três, na ordem de criação.

---

## Edge Cases

- IF a mesma movimentação for enviada duas vezes (requisição duplicada do cliente) THEN o sistema SHALL registrar duas movimentações distintas — não há deduplicação nesta sub-spec (nenhum documento exige idempotência por chave de requisição; introduzi-la seria uma regra nova não solicitada).
- IF `quantidade` for um número não inteiro THEN o sistema SHALL recusar na validação de forma do contrato (tipo inteiro), antes de qualquer lógica de negócio.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| MOV-01 | P1: Registrar entrada | Implementing | Verified |
| MOV-02 | P1: Registrar saída | Implementing | Verified |
| MOV-03 | P1: Registrar devolução | Implementing | Verified |
| MOV-04 | P1: Listar movimentações | Implementing | Verified |

**Coverage:** 4 total, 4 mapped to tasks, 0 unmapped.

---

## Success Criteria

- [ ] Nenhuma saída ou devolução-de-entrada resulta em saldo negativo, inclusive sob concorrência real (testado com Postgres).
- [ ] Nenhuma movimentação é editada ou excluída em nenhum caminho de código.
