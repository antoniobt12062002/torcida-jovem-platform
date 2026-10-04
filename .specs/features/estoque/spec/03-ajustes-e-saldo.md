# Ajustes e Saldo Specification

## Problem Statement

`AD-009`/`ADR-007` preveem uma válvula de escape para quando o saldo precisa ser corrigido além do que entrada/saída/devolução permitem — inclusive para negativo — sempre com motivo e auditoria. Esta sub-spec implementa essa correção (`ajuste`) e a consulta de saldo, ambas com permissão própria, separadas de `02-movimentacoes` porque têm regra de bloqueio e autorização diferentes.

## Goals

- [ ] Um ajuste pode levar o saldo a negativo, mas só com permissão própria e motivo obrigatório.
- [ ] O saldo consultado é sempre a soma de todas as movimentações (entrada, saída, devolução, ajuste) do SKU, nunca uma coluna própria.

## Out of Scope

| Feature | Reason |
| --- | --- |
| Vínculo entre um ajuste e a movimentação que ele corrige | `ADR-007`: "correção de erro é uma nova movimentação de ajuste com motivo, nunca edição" — não exige nem prevê um campo de referência; um `devolucao_de_id`-like para ajuste criaria uma regra nova não solicitada. |
| Estoque mínimo / alerta de saldo baixo | Fora do V1 (decisão do mantenedor) — mesma categoria de feature que `financeiro` já adiou (alertas). |
| Múltiplos locais/depósitos no saldo | Fora do V1 — um único saldo nacional por SKU. |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Quantidade do ajuste é um delta assinado, não um valor absoluto | O operador informa `+3` ou `-3` (o quanto corrigir), nunca "o saldo deve passar a ser X" | Mantém a filosofia de razão de movimentações (`ADR-007`): todo registro é um delta, nunca um estado absoluto. Calcular o delta necessário para atingir um valor-alvo seria uma função adicional não solicitada por nenhum documento. | y (EST-D-010) |
| Ajuste de quantidade zero | Recusado com `quantidade_invalida` | Um ajuste de zero não corrige nada e não tem razão de existir — mesma regra de validação de `02-movimentacoes` para entrada/saída. | y |

**Open questions:** none.

---

## User Stories

### P1: Registrar ajuste ⭐ MVP

**User Story**: Como responsável de estoque, quero registrar um ajuste com motivo, para corrigir o saldo quando entrada/saída/devolução não bastam — mesmo que o resultado seja negativo.

**Why P1**: É a única forma de corrigir um erro de estoque sem editar uma movimentação já registrada (proibido por `AD-009`).

**Acceptance Criteria**:

1. WHEN um ator com `estoque:movimentacao:adjust` envia `POST /estoque/ajustes` com `produto_id`, `quantidade` (inteiro diferente de zero, podendo ser negativo) e `motivo` não vazio THEN o sistema SHALL registrar a movimentação com `tipo=AJUSTE` e a quantidade exatamente como informada, mesmo que o saldo resultante seja negativo, e responder `201`.
2. IF `motivo` for vazio ou ausente THEN o sistema SHALL recusar com `motivo_obrigatorio` (`422`), antes de qualquer leitura ou escrita — mesma ordem de `CancelarLancamento.Execute` em `financeiro`.
3. IF `quantidade` for zero THEN o sistema SHALL recusar com `quantidade_invalida`.
4. IF `produto_id` não existe THEN o sistema SHALL recusar com `produto_nao_encontrado`.
5. IF o ator não tem `estoque:movimentacao:adjust` THEN o sistema SHALL recusar com `forbidden` — inclusive um ator que tem `estoque:movimentacao:create` mas não `adjust` (as duas são permissões distintas, decisão do mantenedor).

**Independent Test**: criar um SKU com saldo zero, ajustar em `-3` (sucesso, saldo passa a `-3`, motivo auditado), ajustar sem motivo (recusado), ajustar com um ator que só tem `movimentacao:create` (recusado).

---

### P1: Consultar saldo ⭐ MVP

**User Story**: Como responsável de estoque, diretoria ou conselho fiscal, quero consultar o saldo atual de um SKU, para saber quanto existe sem somar movimentações manualmente.

**Why P1**: É a pergunta mais básica que `AD-009`/`ADR-007` promete responder de forma confiável.

**Acceptance Criteria**:

1. WHEN um ator com `estoque:saldo:read` consulta o saldo de um `produto_id` THEN o sistema SHALL responder com a soma de todas as movimentações desse SKU (entrada positiva, saída negativa, devolução com o sinal já definido em `02-movimentacoes`, ajuste com o sinal informado), incluindo zero e valores negativos.
2. IF `produto_id` não existe THEN o sistema SHALL recusar com `produto_nao_encontrado`.
3. IF o ator não tem `estoque:saldo:read` THEN o sistema SHALL recusar com `forbidden` — inclusive um ator que só tem `estoque:movimentacao:read` (permissão própria, mesma lógica de `FIN-D-013` em `financeiro`).

**Independent Test**: criar um SKU, entrar com 10, sair com 3, ajustar em `-2`, consultar saldo e ver `5`; ajustar em `-10` adicional e ver o saldo `-5`.

---

## Edge Cases

- IF um SKU nunca recebeu nenhuma movimentação THEN o sistema SHALL responder saldo `0`, nunca um erro (mesma garantia de `ConsultarSaldo` em `financeiro`, que nunca falha por "livro-razão vazio").
- IF um ajuste muito grande, somado ao histórico, ultrapassar o limite seguro de inteiro THEN o sistema SHALL recusar de forma clara — mesma classe de proteção que `platform/money.ErrOutOfRange` já aplica a valores monetários; a contrapartida para quantidade inteira é tratada no design.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| AJS-01 | P1: Registrar ajuste | Implementing | Verified |
| AJS-02 | P1: Consultar saldo | Implementing | Verified |

**Coverage:** 2 total, 2 mapped to tasks, 0 unmapped.

---

## Success Criteria

- [ ] Um ajuste nunca é bloqueado por saldo negativo (é a válvula de escape autorizada).
- [ ] O saldo consultado é sempre consistente com a soma real das movimentações, inclusive imediatamente após um ajuste.
