# ADR-007: O módulo de estoque é a fonte de verdade, baseada em movimentações

- **Date**: 2026-09-26
- **Status**: Accepted
- **Deciders**: @antoniobt12062002
- **Tags**: domain, stock, consistency

## Context and Problem Statement

Vários módulos tocam em quantidades de produtos: a loja vende, o financeiro registra a compra, os eventos consomem itens. Sem um dono único, cada módulo mantém sua própria contagem e os números divergem. É preciso definir quem é a fonte de verdade do estoque e como ela muda.

## Decision Drivers

- Uma única fonte de verdade para "quantos itens existem".
- Rastreabilidade: cada mudança de quantidade tem origem, responsável e motivo.
- Consistência transacional com venda e compra ([ADR-001](001-adotar-monolito-modular.md)).
- Compatível com a imutabilidade e auditoria ([ADR-004](004-auditoria-e-imutabilidade-financeira.md)).

## Considered Options

- Razão de movimentações (ledger) no módulo `estoque`, com saldo derivado
- Coluna de quantidade mutável no produto, atualizada por quem vende ou compra
- Event sourcing completo do estoque

## Decision Outcome

Chosen option: **o módulo `estoque` é o único dono das quantidades, registradas como movimentações imutáveis** (entrada, saída, ajuste, devolução). O saldo é calculado a partir dos movimentos e mantido na mesma transação.

- **Estoque baseado em razão de movimentações; o saldo é calculado pelos movimentos.**
- **Estoque negativo é bloqueado por padrão.** Exceções só ocorrem por ajuste autorizado (permissão específica) e auditado ([ADR-004](004-auditoria-e-imutabilidade-financeira.md)).
- Loja, compras e eventos **não alteram quantidades diretamente**: chamam o serviço de aplicação de `estoque`.
- Cada movimentação guarda produto, quantidade, tipo, origem (venda, compra, evento, inventário), usuário e data/hora.
- Correção de erro é uma nova movimentação de ajuste com motivo, nunca edição.
- Compras seguem o fluxo de [FIN-001](../finance/FIN-001-VISAO-GERAL.md), seção 10: pedido, entrega, entrada no estoque e despesa.

### Positive Consequences

- Histórico completo e explicável de qualquer saldo.
- Fim das divergências entre módulos; inventário vira conferência do saldo contra a contagem real.

### Negative Consequences

- Mais tabelas e mais disciplina que uma simples coluna de quantidade.
- Como o negativo é bloqueado, a concorrência (duas vendas do último item) exige bloqueio ou restrição no banco.
- Reservas para a loja virtual (carrinho, pagamento pendente) ainda precisam de definição.

## Pros and Cons of the Options

### Razão de movimentações ✅ Chosen

- ✅ Auditável, consistente e coerente com o ADR-004
- ❌ Modelagem mais elaborada

### Coluna de quantidade no produto

- ✅ Simples
- ❌ Sem histórico; qualquer módulo pode corromper o número

### Event sourcing completo

- ✅ Reconstrução total do passado
- ❌ Complexidade desproporcional

## Open Questions

- Reservas e expiração para pagamentos pendentes.
- Estoque por variação (tamanho e cor de camiseta) ou por produto.

## Links

- [Fronteiras de domínio](../architecture/domain-boundaries.md)
- [FIN-001](../finance/FIN-001-VISAO-GERAL.md)
