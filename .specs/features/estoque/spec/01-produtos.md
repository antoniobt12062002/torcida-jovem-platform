# Produtos (SKU de Estoque) Specification

## Problem Statement

`estoque` precisa de uma unidade mínima e estável para identificar o que está sendo movimentado. Sem ela, nenhuma movimentação (entrada, saída, ajuste, devolução) tem a quê se referir. Esta sub-spec entrega só essa identidade mínima — nunca um catálogo comercial.

## Goals

- [ ] Toda movimentação de estoque referencia um SKU que existe e é identificável de forma estável.
- [ ] Criar e listar SKUs pela API, sem nenhum dado comercial (preço, descrição, imagem, categoria).

## Out of Scope

| Feature | Reason |
| --- | --- |
| Atualizar/renomear SKU | Decisão do mantenedor (EST-D-007): identidade do SKU é estável após a criação; sem caso de uso nem permissão de update. |
| Desativar/excluir SKU | Decisão do mantenedor (EST-D-007): sem exclusão física, sem desativação — se necessário no futuro, é evolução própria, com nova decisão. |
| Catálogo comercial (preço, descrição, imagem, categoria) | Decisão do mantenedor (EST-D-001): pertence ao futuro módulo `loja`, que poderá referenciar o SKU criado aqui. |
| Variação estruturada (tamanho/cor como atributos, tabela pai/filho) | Decisão do mantenedor (EST-D-003): cada variação é um SKU atômico independente — sem `produto_variacao`, sem relação pai/filho. |
| Múltiplos locais/depósitos | Fora do V1 (decisão do mantenedor) — um único saldo nacional por SKU. |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Nome da tabela | `produtos_estoque` (não `produtos`) | Evita colisão de nome com uma futura tabela de catálogo comercial em `loja`, que pode usar "produto" num sentido mais rico (EST-D-011). | y |
| Unicidade do código | `codigo` é único globalmente, case-sensitive, `NOT NULL` | É o identificador de negócio (ex. `CAMISA-TJ-P-PRETA`) — duplicidade quebraria a referência de qualquer movimentação. Nenhum documento define normalização de maiúsculas/minúsculas; mantido case-sensitive por ser o comportamento padrão do `UNIQUE` do Postgres, sem lógica adicional. | y |
| Formato do código | string livre, sem máscara nem validação de padrão | Nenhum documento define um formato (ex. prefixos obrigatórios); impor uma máscara agora seria regra de negócio não solicitada. | y |

**Open questions:** none — todas resolvidas ou registradas acima.

---

## User Stories

### P1: Criar SKU de estoque ⭐ MVP

**User Story**: Como responsável de estoque (`ESTOQUE_LOJA`), quero criar um SKU com código, nome e unidade de medida, para poder movimentar e consultar saldo desse item.

**Why P1**: Sem isso, nenhuma movimentação é possível — é a base de toda a sub-spec `02-movimentacoes`.

**Acceptance Criteria**:

1. WHEN um ator com `estoque:produto:create` envia `codigo`, `nome` e `unidade_medida` válidos THEN o sistema SHALL criar o SKU e devolver seus dados, incluindo o `id` gerado.
2. IF o `codigo` informado já existe THEN o sistema SHALL recusar a criação com o erro `codigo_duplicado`, sem alterar nenhum dado.
3. IF `codigo`, `nome` ou `unidade_medida` estiver vazio THEN o sistema SHALL recusar a criação com um erro de validação, sem persistir nada.
4. IF o ator não tem `estoque:produto:create` THEN o sistema SHALL recusar com `forbidden`, nunca revelando se o `codigo` já existe.

**Independent Test**: criar um SKU com um código novo (sucesso), repetir o mesmo código (recusado), tentar com um ator sem a permissão (recusado sem vazar informação).

---

### P1: Listar SKUs de estoque ⭐ MVP

**User Story**: Como responsável de estoque, diretoria ou conselho fiscal, quero listar todos os SKUs cadastrados, para saber o que existe antes de movimentar ou consultar saldo.

**Why P1**: É a única forma de descobrir quais SKUs existem sem acesso direto ao banco.

**Acceptance Criteria**:

1. WHEN um ator com `estoque:produto:read` lista os SKUs THEN o sistema SHALL devolver todos, sem paginação, filtro ou ordenação (decisão deliberada do V1, mesma lógica de `FIN-D-021` em `financeiro`).
2. IF o ator não tem `estoque:produto:read` THEN o sistema SHALL recusar com `forbidden`.

**Independent Test**: criar dois SKUs, listar e ver os dois, exatamente como criados.

---

## Edge Cases

- IF `unidade_medida` tiver um valor nunca usado antes (não é um enum fechado) THEN o sistema SHALL aceitar — é texto livre, não uma lista fixa (nenhum documento define um catálogo de unidades).
- IF dois SKUs forem criados com `codigo` idêntico em requisições concorrentes THEN o sistema SHALL garantir a unicidade via constraint do banco (`UNIQUE`), recusando a segunda com `codigo_duplicado` mesmo sob concorrência.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| PRD-01 | P1: Criar SKU | Implementing | Verified |
| PRD-02 | P1: Listar SKUs | Implementing | Verified |

**Coverage:** 2 total, 2 mapped to tasks, 0 unmapped.

---

## Success Criteria

- [ ] Toda movimentação futura (`02-movimentacoes`) tem um SKU válido para referenciar.
- [ ] Nenhum dado comercial (preço, descrição, imagem) existe na tabela de SKU.
