# ADR-017: Evolução aditiva de `financeiro` e `estoque`: origem, deduplicação e tratamento financeiro das cobranças

- **Date**: 2026-10-13 (revisão 2, com as decisões M-1 a M-9 do mantenedor)
- **Status**: Proposed. Aprovado conceitualmente. Autorizado apenas o **planejamento**: nenhuma migração ou código.
- **Deciders**: @antoniobt12062002
- **Tags**: finance, inventory, idempotency, tj-connect

## Context and Problem Statement

O TJ Connect reutiliza o Financeiro e o Estoque (TJ-FIN-001, TJ-EST-001). Verificado no código (`api/migrations/000007`, `000009`):

- `lancamentos` tem `valor_bruto_cents`, `taxa_cents` e `valor_liquido_cents` (o líquido é calculado na criação e nunca recalculado), status `CRIADA/RECEBIDA/PAGA/CANCELADA`, `devolucao_de_id` (FIN-D-003) e `criado_por NOT NULL REFERENCES users`. Não tem origem nem chave de idempotência.
- `movimentacoes_estoque` é append-only e tem `origem` em `VENDA/COMPRA/EVENTO/INVENTARIO/AJUSTE_MANUAL` e `responsavel_id NOT NULL`. Não tem referência nem chave.
- O regime é de caixa (FIN-D-001), sem recebimento parcial (FIN-D-002). A devolução é uma despesa nova apontando para a receita (FIN-D-003). O cancelamento corrige erro de registro (FIN-D-006).

O mantenedor pediu: referências de origem, deduplicação e rastreabilidade; tratamento de bruto, taxa, líquido, estornos e reembolsos sem duplicar nem distorcer indicadores; estoque só como movimentação física, sem nomes personalizados.

## Decision Drivers

- Só o dono altera as suas tabelas; mudança só aditiva; nenhuma decisão FIN-D/EST-D revogada.
- Idempotência por **restrição de banco**.
- Um fato gerador → um lançamento lógico; um evento físico → uma movimentação.
- Indicadores de "contribuições associativas" não podem contar dinheiro devolvido nem juros que não pertencem à TJ.

## Considered Options

1. **Colunas aditivas e casos de uso dedicados nos módulos donos.**
2. **Tabela de vínculos no módulo chamador.**
3. **Reutilizar os casos de uso atuais** com outro ator.

## Decision Outcome

Chosen option: **1**. A unicidade vive junto do dado que não pode duplicar.

### Origem e rastreabilidade

- **`lancamentos`**: colunas opcionais `origem_tipo`, `origem_id` e `chave_idempotencia`, com unicidade parcial (`WHERE chave_idempotencia IS NOT NULL`). Os lançamentos manuais continuam sem elas.
- **`movimentacoes_estoque`**: colunas opcionais `referencia_tipo`, `referencia_id` e `chave_idempotencia`, com unicidade parcial.
- **Cadeia completa** (cada elo é gravado pelo seu dono, numa direção só):

  adesão → cobrança (`originador_*`) → order e pagamentos no provedor (ids externos, em `cobrancas`) → lançamento (`origem_*` = cobrança) → auditoria (`correlation_id`).

  `cobrancas` guarda o id do lançamento que gerou, numa coluna própria; `financeiro` nunca escreve em `cobrancas`.

### Tratamento financeiro

| Fato | Lançamento | Valores | Chave de idempotência |
|---|---|---|---|
| Cobrança paga | Receita `RECEBIDA`, criada e liquidada num caso de uso só, na conta de contribuição do plano | **Bruto** = preço do plano devido à TJ. **Taxa** = taxa de processamento retida pelo provedor nesta transação. **Líquido** = bruto − taxa, conferido com o valor líquido informado pelo provedor; uma diferença vira divergência de conciliação, sem lançamento | `cobranca:<id>:receita` |
| Juros de parcelamento pagos pelo comprador | **Nenhum lançamento** se o juro pertence ao provedor (financiamento do Mercado Pago). Guardado em `cobrancas` para transparência e conciliação | — | — |
| Reembolso concluído | Despesa `PAGA` com `devolucao_de_id` = receita (FIN-D-003), valor = valor devolvido ao comprador | Taxa não restituída pelo provedor, se houver: despesa separada de taxa, também vinculada | `reembolso:<id_externo>:devolucao` e `reembolso:<id_externo>:taxa` |
| Contestação perdida (chargeback) | Despesa com `devolucao_de_id` = receita, mais despesa de tarifa de contestação, se houver | Conforme o provedor | `contestacao:<id_externo>:...` |
| Erro de registro | Cancelamento (FIN-D-006), nunca usado para representar reembolso | — | — |

Regras que evitam distorção:

- **Reembolso nunca cancela a receita.** A receita fica `RECEBIDA` e a devolução a compensa: no caixa, o saldo líquido fica correto e o histórico, intacto.
- **Indicador de contribuições** = receitas de origem `cobrancas` − devoluções vinculadas a elas (`devolucao_de_id`). A especificação do indicador vai na sub-spec `financeiro/spec/07-receitas-de-cobranca` (nome provisório).
- **Pagamento tardio sem vaga** (TJ-PAG-003): o dinheiro é real e entra no caixa, mas **não é contribuição associativa**. É registrado numa conta transitória de "valores a regularizar", fora do indicador. Ao resolver: devolução vinculada ou, com concordância expressa do comprador, reclassificação (cancelamento do registro transitório e nova receita, FIN-D-006/FIN-D-007).
- **Parcelamento**: com juros pagos pelo comprador, o recebedor recebe o valor do plano de uma vez (sujeito ao prazo de liberação). FIN-D-002 continua válido. **A verificar** junto com os itens da [ADR-012](012-cobrancas-e-eventos-persistentes.md). Se o produto exigir parcelamento custeado pelo recebedor ou repasse em parcelas, isso volta como decisão.
- **Data do lançamento**: data de aprovação ou data de liberação do dinheiro. **Pendência contábil.** As duas datas ficam guardadas em `cobrancas`.

Os **casos de uso dedicados** do Financeiro ("registrar receita de cobrança", "registrar devolução de cobrança", "registrar tarifa de cobrança") recebem a permissão de classe `servico` ([ADR-013](013-contas-de-servico-e-autoria-de-processos.md), [ADR-016](016-permissoes-sensiveis-e-escopo-por-jogo.md)). Uma repetição com a mesma chave devolve o registro existente, sem erro. `criado_por` é a conta de serviço; a auditoria carrega `correlation_id` e `decision_id` (para reembolso aprovado).

### Estoque: só movimentação física

- **Novas origens** (nomes finais no Specify): `PRODUCAO_ENCOMENDA` (ENTRADA de unidades recebidas do fornecedor) e `ENTREGA_BENEFICIO` (SAIDA ao associado). Os valores atuais não mudam.
- **Recebimento de lote**: uma ENTRADA por recebimento físico e por SKU de tamanho (EST-D-003), com a quantidade realmente recebida. Recebimento parcial e divergência (esperado × recebido) ficam no lote, em `associados`; o estoque registra só o que entrou. Chave: `lote:<id>:recebimento:<n>:<sku>`.
- **Entrega**: uma SAIDA de 1 unidade do SKU do tamanho. Chave: `encomenda:<id>:entrega`. Saldo negativo bloqueado (ADR-007).
- **Defeito ou devolução**: DEVOLUCAO referenciando a movimentação original (EST-D-008/009).
- **Sem dados pessoais**: `referencia_id` é um UUID opaco de lote ou encomenda. **Nenhum nome personalizado, nome civil ou titular** chega às colunas, ao motivo ou à auditoria do estoque. Um teste de contrato garante que o caso de uso do estoque não aceita esses campos.
- **Ação humana**: recebimento e entrega são feitos por pessoa com permissão do estoque; `responsavel_id` é essa pessoa.
- **Pagamento não gera estoque**: o pagamento cria só a obrigação de encomenda em `associados` (AC-CAM-03).

### Positive Consequences

- AC-PAG-02 e AC-CAM-05 garantidos por restrição; nenhum histórico existente muda.
- Contribuições líquidas de devoluções e livres de juros de terceiros; dinheiro sem vaga separado.

### Negative Consequences

- Duas sub-specs e migrations novas em módulos entregues, com regressão a verificar.
- Dois caminhos de criação no Financeiro (manual e por cobrança); telas e auditoria distinguem os dois.
- Três pendências contábeis bloqueiam o Design da sub-spec `financeiro/spec/07-receitas-de-cobranca`: conta transitória, data do lançamento e taxa no reembolso.

## Pros and Cons of the Options

### 1. Aditivo nos donos ✅ Chosen

- ✅ Unicidade no dado; respeita o dono.
- ❌ Muda módulos entregues.

### 2. Vínculo no chamador

- ❌ Uma falha entre "criar lançamento" e "gravar vínculo" duplica no reprocessamento.

### 3. Casos de uso atuais

- ❌ Sem idempotência; criar e liquidar separados abrem uma janela inconsistente.

## Links

- [ADR-004](004-auditoria-e-imutabilidade-financeira.md), [ADR-007](007-fonte-de-verdade-do-estoque.md), [ADR-012](012-cobrancas-e-eventos-persistentes.md), [ADR-013](013-contas-de-servico-e-autoria-de-processos.md), [ADR-016](016-permissoes-sensiveis-e-escopo-por-jogo.md).
- Decisões preservadas: `FIN-D-001`, `FIN-D-002`, `FIN-D-003`, `FIN-D-006`, `FIN-D-007`, `EST-D-003`, `EST-D-008`, `EST-D-009`.
