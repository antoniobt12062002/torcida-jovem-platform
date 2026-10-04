# Estoque (Web) — Design

**Spec**: `.specs/features/web/spec/05-estoque-web.md`
**Pré-requisito**: sub-spec 01 completa (F5 integrada).

## Architecture Overview

```
app/(app)/estoque/produtos/page.tsx       ─→ features/estoque/produtos/*
app/(app)/estoque/produtos/[id]/page.tsx  ─→ features/estoque/produto-detail.tsx
                                               ├─ saldo (EWB-01 AC4)
                                               ├─ movimentacoes/* (EWB-02)
                                               └─ ajustes/* (EWB-03)
features/estoque/nav.ts
todas ─→ lib/api/client.ts (estoque) · lib/api/query-keys.ts (estoque.*)
```

## Code Reuse Analysis

### Existing Components to Leverage

| Existente | Uso |
| --- | --- |
| `components/app` (`QuantityField` em `positive` e `nonZero`, `ConfirmDialog` com motivo, estados, `RequirePermission`, `format.ts`) | formulários e exibição |
| `components/ui` (`table`, `dialog`, `input`) | listas e formulários |

### Integration Points

| Ponto | Como |
| --- | --- |
| Navegação | `features/estoque/nav.ts` (item Produtos, `estoque:produto:read`); a INT liga. |

## Components

| Arquivo | Função |
| --- | --- |
| `features/estoque/nav.ts` | `estoqueNav: NavItem[]` |
| `features/estoque/errors.ts` | Catálogo de erros |
| `features/estoque/produtos/hooks.ts` | `useProdutos()`, `useProduto(id)` (do cache), `useSaldo(id)`, criação |
| `features/estoque/produtos/search.ts` | Função pura de busca por código ou nome |
| `features/estoque/produtos/produtos-page.tsx`, `create-produto-dialog.tsx` | Lista e criação |
| `features/estoque/produto-detail.tsx` | Detalhe com seções por permissão (EWB-04 AC2) |
| `features/estoque/movimentacoes/hooks.ts` | `useMovimentacoes(id)`, `useRegistrarMovimentacao()` com invalidação de histórico e saldo |
| `features/estoque/movimentacoes/payload.ts` | Função pura que monta o corpo com `origem = INVENTARIO` (WEB-D-007) |
| `features/estoque/movimentacoes/movimentacoes-panel.tsx`, `movimentacao-dialog.tsx` | Histórico, entrada, saída, devolução |
| `features/estoque/ajustes/ajuste-dialog.tsx`, `hooks.ts` | Ajuste com prévia do saldo |
| `app/(app)/estoque/produtos/page.tsx`, `app/(app)/estoque/produtos/[id]/page.tsx` | Páginas finas |

## Data Models

`Produto`, `ProdutoList`, `Movimentacao`, `MovimentacaoList`, `CreateProdutoRequest`, `CreateMovimentacaoRequest`, `CreateAjusteRequest`, `Saldo` dos `.d.ts`. Rótulos pt-BR para `TipoMovimentacao` e `OrigemMovimentacao`.

## Error Handling Strategy

| `code` | Mensagem |
| --- | --- |
| `codigo_duplicado` | "Já existe um produto com este código." (campo código) |
| `saldo_insuficiente` | "Saldo insuficiente para esta saída." |
| `produto_nao_encontrado` | "Produto não encontrado." |
| `quantidade_invalida` | "Quantidade inválida." |
| `devolucao_invalida` | "Só é possível devolver uma entrada ou saída deste produto." |
| `motivo_obrigatorio` | "Informe o motivo do ajuste." |

## Risks & Concerns

| Risco | Mitigação |
| --- | --- |
| Uma origem reservada aparecer por engano | `payload.ts` fixa `INVENTARIO`; teste confere o corpo enviado; alvo do sensor do verificador. |

## Tech Decisions

| Decisão | Motivo |
| --- | --- |
| Movimentações e ajuste no detalhe do produto | Saldo e histórico são por produto na API. |
