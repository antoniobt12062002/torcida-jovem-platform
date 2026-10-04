# Financeiro (Web) — Design

**Spec**: `.specs/features/web/spec/04-financeiro-web.md`
**Pré-requisito**: sub-spec 01 completa (F5 integrada).

## Architecture Overview

Duas frentes paralelas, sem arquivo em comum:

```
FIN-a  features/financeiro/nav.ts
       features/financeiro/contas/*   ─→ app/(app)/financeiro/contas/page.tsx
       features/financeiro/saldo/*    ─→ app/(app)/financeiro/saldo/page.tsx

FIN-b  features/financeiro/lancamentos/*  ─→ app/(app)/financeiro/lancamentos/page.tsx
       features/financeiro/comprovantes/*     app/(app)/financeiro/lancamentos/[id]/page.tsx

ambas  ─→ lib/api/client.ts (financeiro) · lib/api/query-keys.ts (financeiro.*)
```

FIN-b precisa de contas (seleção de conta, nome na lista) e invalida o saldo. Nos dois casos usa só as chaves de `lib/api/query-keys.ts` (WEB-D-010): FIN-b tem a sua própria consulta `useContasOptions()` sobre `queryKeys.financeiro.contas`, sem importar nada de `features/financeiro/contas/`. As duas consultas dividem o cache porque usam a mesma chave e a mesma função de busca (`GET /contas` devolvendo `ContaList`).

## Code Reuse Analysis

### Existing Components to Leverage

| Existente | Uso |
| --- | --- |
| `components/app` (`MoneyField`, `ConfirmDialog` com motivo, `StatusBadge`, estados, `RequirePermission`, `format.ts`) | formulários, workflow, exibição |
| `components/ui` (`table`, `select`, `dialog`, `dropdown-menu`) | listas e formulários |
| `lib/money.ts` | exibição e prévia do líquido |

### Integration Points

| Ponto | Como |
| --- | --- |
| Navegação | `features/financeiro/nav.ts` (dono: FIN-a) lista os três itens do módulo, inclusive o caminho `/financeiro/lancamentos` de FIN-b, com as permissões de leitura. FIN-b não altera `nav.ts`. A INT liga. |
| Saldo | FIN-b invalida `queryKeys.financeiro.saldo` sem importar `features/financeiro/saldo/`. |

## Components

### FIN-a

| Arquivo | Função |
| --- | --- |
| `features/financeiro/nav.ts` | `financeiroNav: NavItem[]` (Contas, Lançamentos, Saldo) |
| `features/financeiro/contas/hooks.ts` | `useContas()`, mutações de criar, renomear, desativar |
| `features/financeiro/contas/tree.ts` | Função pura `ContaList → árvore` (contas órfãs na raiz) |
| `features/financeiro/contas/errors.ts` | Catálogo de erros de contas |
| `features/financeiro/contas/contas-page.tsx` | Árvore, ações por permissão, diálogos |
| `features/financeiro/saldo/saldo-page.tsx` | Saldo e explicação |
| `app/(app)/financeiro/contas/page.tsx`, `app/(app)/financeiro/saldo/page.tsx` | Páginas finas com `RequirePermission` |

### FIN-b

| Arquivo | Função |
| --- | --- |
| `features/financeiro/lancamentos/hooks.ts` | `useLancamentos()`, `useLancamento(id)` (seleciona do cache da lista), `useContasOptions()`, mutações (criar, editar, receber, pagar, cancelar, devolver) com invalidação de `lancamentos` e `saldo` |
| `features/financeiro/lancamentos/filters.ts` | Função pura de filtro e ordenação |
| `features/financeiro/lancamentos/actions.ts` | Função pura `(lançamento, permissões) → ações disponíveis` (FWB-03 AC4, FWB-04 AC6-7, casos de borda) |
| `features/financeiro/lancamentos/errors.ts` | Catálogo de erros de lançamentos |
| `features/financeiro/lancamentos/lancamento-form.tsx` | Criar e editar (prévia do líquido) |
| `features/financeiro/lancamentos/devolucao-form.tsx` | Devolução |
| `features/financeiro/lancamentos/lancamentos-page.tsx`, `lancamento-detail.tsx` | Lista e detalhe |
| `features/financeiro/comprovantes/hooks.ts` | `useComprovantes(id)`, `useUploadComprovante(id)`, `downloadComprovante(documentId)` (chamada direta, sem cache) |
| `features/financeiro/comprovantes/errors.ts`, `comprovantes-panel.tsx` | Catálogo e painel no detalhe |
| `app/(app)/financeiro/lancamentos/page.tsx`, `app/(app)/financeiro/lancamentos/[id]/page.tsx` | Páginas finas |

Upload: `openapi-fetch` com `bodySerializer` que monta `FormData` com o campo `file`; o navegador define o `Content-Type` com o boundary. O middleware de CSRF de F1 vale para o upload como para qualquer escrita.

## Data Models

`Conta`, `ContaList`, `Lancamento`, `LancamentoList`, `Saldo`, `Comprovante`, `ComprovanteList`, `ComprovanteUrl` e os `*Request` dos `.d.ts`. `FormaPagamento`, `TipoConta`, `StatusLancamento` ganham rótulos pt-BR no front.

## Error Handling Strategy

| `code` | Mensagem |
| --- | --- |
| `conta_ja_utilizada` | "Esta conta já foi usada em um lançamento e não pode ser renomeada." |
| `conta_tipo_incompativel` | "A subconta precisa ter o mesmo tipo da conta pai." |
| `conta_invalida` | "Escolha uma conta ativa do mesmo tipo do lançamento." |
| `lancamento_tipo_incompativel` | "O tipo do lançamento não combina com o tipo da conta." |
| `amount_out_of_range` | "Valor fora da faixa permitida." |
| `lancamento_imutavel` | "Só lançamentos em aberto podem ser editados." |
| `lancamento_nao_pode_ser_recebido` / `_pago` | "Este lançamento não está mais em aberto." |
| `lancamento_ja_cancelado` | "Este lançamento já foi cancelado." |
| `motivo_obrigatorio` | "Informe o motivo." |
| `devolucao_invalida` | "Só é possível devolver uma receita recebida." |
| `document_*`, `payload_too_large` | tipo ou tamanho de arquivo não aceito |

## Risks & Concerns

| Risco | Mitigação |
| --- | --- |
| Upload grande pelo rewrite | Verificado em F1 (FND-01 AC5): o rewrite trunca corpos acima de cerca de 10 MiB. O painel recusa antes do envio arquivos acima de 10.420.224 bytes (10 MiB − 64 KiB), constante única em `features/financeiro/comprovantes/`, com mensagem de arquivo acima do limite (WEB-D-009, correção de 2026-10-04). |
| URL assinada expirar se guardada | Buscada no clique, nunca em cache. |
| FIN-a e FIN-b divergirem na consulta de contas | Mesma chave e mesma forma de retorno; teste de FIN-b usa a fixture de contas de `test/msw/fixtures.ts`. |

## Tech Decisions

| Decisão | Motivo |
| --- | --- |
| `actions.ts` como função pura | Concentra a regra de visibilidade num ponto testável e alvo claro para o sensor do verificador. |
| Detalhe pelo cache da lista | Contrato sem `GET` por id. |
