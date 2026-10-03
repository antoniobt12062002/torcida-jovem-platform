# Estoque — STATE

Documento de coordenação da feature `estoque`, decomposta em 5 sub-specs (`01`-`03` o domínio V1, `04` RBAC, `05` sua camada HTTP) — mesmo padrão estrutural de `financeiro/STATE.md`. Mantém enxuto o `.specs/STATE.md` do projeto: aqui ficam as decisões e o mapa específicos desta feature.

Convenção de numeração local: **EST-D-NNN** para decisões que valem só para `estoque`. Nunca reaproveita o numerador `AD-NNN` do `.specs/STATE.md` global.

---

## Mapa das sub-specs

| # | Spec | Requisitos | Status |
|---|---|---|---|
| 01 | `spec/01-produtos.md` | PRD-01, PRD-02 | Em andamento (T1/5) |
| 02 | `spec/02-movimentacoes.md` | MOV-01 a MOV-04 | Especificada; não implementada |
| 03 | `spec/03-ajustes-e-saldo.md` | AJS-01, AJS-02 | Especificada; não implementada |
| 04 | `spec/04-permissoes.md` | PERM-01 | Especificada; não implementada |
| 05 | `spec/05-api-http.md` | API-01 a API-04 | Especificada; não implementada |

## Dependências

```
01-produtos
       │  (produto_id)
       ▼
02-movimentacoes ──► 03-ajustes-e-saldo   (mesma tabela movimentacoes_estoque; permissão e regra de bloqueio diferentes)

04-permissoes: transversal — depende de 01-03 publicarem suas constantes Perm*, não depende de nenhuma delas em código real.
05-api-http: depende de 01-04 completas (precisa dos 6 casos de uso e das 6 permissões).
```

Nenhuma dependência circular. `02⇢03` é unidirecional de verdade (03 lê a mesma tabela e o mesmo repositório de 02, nunca o contrário) — mais simples que a relação `02⇢01` de `financeiro` (que tinha uma leitura pontual em sentido contrário à ordem de implementação); aqui a ordem de dependência e a ordem de implementação coincidem.

## Fronteiras de responsabilidade

| Dado/operação | Dona | Nota |
|---|---|---|
| `Produto` (SKU mínimo de estoque) | `01` | única spec que insere; nunca editado nem desativado (`EST-D-007`) |
| `Movimentacao` (todos os tipos, incluindo `AJUSTE`) | `02` | única spec que insere a linha (inclusive quando `03` cria um `AJUSTE` — usa o mesmo repositório de `02`, com uma porta mais estreita) |
| Regra de bloqueio de saldo negativo (`SaldoComLock`) | `02` | a regra de quando bloquear (saída e devolução-de-entrada) é de `02`; `03` explicitamente não a usa |
| Catálogo de permissões e matriz de papéis | `04` | única spec que define a `Contribution` |

`01`-`03` são sub-specs do mesmo módulo Go (`api/internal/estoque/`), não módulos de negócio separados — mesma disciplina de `financeiro`.

### Estrutura de pacotes

Mesmo esqueleto de `architecture-overview.md`, idêntico ao de `identity`/`financeiro`:

```
api/internal/estoque/
├── module.go     # estoque.Contribution() (04) e estoque.New(...) (05)
├── domain/       # produto.go, movimentacao.go
├── app/          # um arquivo por caso de uso
├── infra/        # produto_repository.go, movimentacao_repository.go
└── http/         # vazio até 05-api-http
```

---

## Decisões fechadas (EST-D-NNN)

### EST-D-001 — SKU mínimo é dono por `estoque`; catálogo comercial fica para `loja` — **Aprovada**
**Decisão**: `estoque` possui uma representação mínima de SKU (código, nome, unidade de medida) — nenhum dado comercial (preço, descrição, imagem, categoria). O catálogo comercial, quando `loja` for especificada, poderá referenciar o SKU criado aqui por id. Nenhum módulo `produtos` independente é criado agora.
**Escopo**: `01-produtos`.

### EST-D-002 — Sem reserva no V1 — **Aprovada**
**Decisão**: nenhum mecanismo de reserva, expiração de reserva ou carrinho reservado. A responsabilidade por decidir quando uma venda deve reservar estoque fica para a futura especificação de `loja`. `estoque` V1 só tem saldo e movimentações.
**Escopo**: `02-movimentacoes` (não implementa nada disso).

### EST-D-003 — Controle por SKU atômico, sem variação estruturada — **Aprovada**
**Decisão**: cada SKU representa uma unidade distinta de estoque. Tamanho, cor ou outra característica que exija saldo independente resulta em um SKU diferente (ex. `CAMISA-TJ-P-PRETA`, `CAMISA-TJ-M-PRETA` são SKUs independentes). Sem `produto_variacao`, sem relação pai/filho, sem tabela de atributos de variação.
**Escopo**: `01-produtos`.

### EST-D-004 — RBAC: 6 permissões — **Aprovada**
**Decisão**: `estoque:produto:create`, `estoque:produto:read`, `estoque:movimentacao:create`, `estoque:movimentacao:adjust`, `estoque:movimentacao:read`, `estoque:saldo:read`. `ESTOQUE_LOJA` recebe as 6; `DIRETORIA` e `CONSELHO_FISCAL` recebem as 3 de leitura; `PRESIDENTE` automático; demais papéis nenhuma. A separação de `movimentacao:create`/`movimentacao:adjust` é intencional — ajuste é operação excepcional, com motivo obrigatório e autorização própria.
**Escopo**: `04-permissoes`.

### EST-D-005 — API de movimentações: endpoint único com `tipo` + endpoint próprio de ajuste — **Aprovada**
**Decisão**: `POST /api/v1/estoque/movimentacoes` com `tipo` (`ENTRADA`\|`SAIDA`\|`DEVOLUCAO`) no corpo; `POST /api/v1/estoque/ajustes` como endpoint próprio. Sem `/entradas`/`/saidas` separados. Devolução referencia a movimentação original por `movimentacao_de_id`; a original nunca é editada.
**Escopo**: `05-api-http`.

### EST-D-006 — Mapeamento de erros HTTP — **Aprovada**
**Decisão**: `saldo_insuficiente`→`409` (conflito com o estado atual do estoque, não erro estrutural de payload); `motivo_obrigatorio`→`422`; `produto_nao_encontrado`→`404`; `forbidden`→`403`. Fechados no design, consistentes com a mesma régua semântica já usada em `financeiro` (`FIN-D-022`: recurso inexistente→404, conflito com estado atual→409, validação de campo/referência→422): `codigo_duplicado`→`409` (conflito de identidade, análogo a `financeiro:conta_ja_utilizada`), `devolucao_invalida`→`422` (validação de referência de campo), `quantidade_invalida`→`422`.
**Escopo**: `05-api-http` (API-04).

### EST-D-007 — SKU imutável após a criação — **Aprovada**
**Decisão**: sem `produto:update`, sem `produto:deactivate`, sem exclusão física. O campo `ativo` é removido do modelo (não há funcionalidade que o justifique no V1). A identidade do SKU é estável após a criação. Desativação/aposentadoria de SKU, se necessária no futuro, é evolução própria, com nova decisão.
**Escopo**: `01-produtos`.

### EST-D-008 — Sinal da devolução inverte o sentido da movimentação referenciada (achado lógico, registrado com justificativa)
**Achado**: nenhum documento (`ADR-007` incluído) define a direção do efeito de uma devolução. **Decisão derivada, não uma regra de negócio nova**: devolver uma `SAIDA` repõe estoque (soma positiva, nunca bloqueada); devolver uma `ENTRADA` remove estoque (soma negativa, sujeita à mesma regra de bloqueio de saldo negativo que `SAIDA`). Mantém uma única regra de bloqueio (bloqueia quando o efeito líquido é retirar, nunca quando é repor), sem introduzir o conceito de "devolução que nunca falha" nem duplicar a lógica de bloqueio por tipo.
**Escopo**: `02-movimentacoes`.

### EST-D-009 — Devolução só pode referenciar `ENTRADA` ou `SAIDA` (achado lógico, registrado com justificativa)
**Achado**: evita cadeias de devolução-de-devolução e evita devolver um ajuste (que já é a própria ferramenta de correção). Mesmo espírito de `FIN-D-003` em `financeiro` (que restringe o que uma devolução pode referenciar).
**Escopo**: `02-movimentacoes`.

### EST-D-010 — Ajuste é um delta assinado, nunca um valor absoluto (achado lógico, registrado com justificativa)
**Achado**: mantém a filosofia de razão de movimentações (`ADR-007`) — todo registro é um delta, nunca um estado absoluto. Calcular o delta necessário para atingir um valor-alvo seria uma função adicional não solicitada por nenhum documento.
**Escopo**: `03-ajustes-e-saldo`.

### EST-D-011 — Nomes de tabela `produtos_estoque`/`movimentacoes_estoque` (achado mecânico)
**Achado**: evita colisão de nome com uma futura tabela de catálogo comercial em `loja`, que pode usar "produto" num sentido mais rico e incompatível com o SKU mínimo de `estoque` (`EST-D-001`).
**Escopo**: `01-produtos`, `02-movimentacoes`.

---

## Regras herdadas de `AD-009`/`ADR-007` (não redecididas aqui, só referenciadas)

Saldo é sempre derivado das movimentações, nunca uma coluna mutável; movimentações são append-only; nenhuma é editada para corrigir histórico; correção ocorre por nova movimentação de ajuste; saída não pode resultar em saldo negativo; ajuste autorizado pode corrigir o saldo inclusive para negativo; ajuste exige motivo; devolução é uma nova movimentação que referencia a original sem jamais alterá-la. Concorrência tratada por `pg_advisory_xact_lock`, dentro da mesma transação, com teste de integração real obrigatório para duas operações concorrentes disputando o último saldo (exigido explicitamente em `tasks/02-movimentacoes.md`/T3).

## Fora do V1 (não implementar sem nova decisão)

Reserva de estoque (`EST-D-002`); integração real com `financeiro` (compra→entrada) ou `loja` (venda→saída) — nenhum dos dois módulos do outro lado existe; custo de aquisição/métodos de custeio; estoque mínimo/alertas; múltiplos locais/depósitos; quantidade fracionária; variação estruturada (`EST-D-003`); qualquer catálogo comercial (`EST-D-001`).

---

## Matriz de permissões do V1 (consolidada)

| Permissão | `ESTOQUE_LOJA` | `DIRETORIA` | `CONSELHO_FISCAL` | `PRESIDENTE` |
|---|---|---|---|---|
| `estoque:produto:create` | ✓ | | | automático |
| `estoque:produto:read` | ✓ | ✓ | ✓ | automático |
| `estoque:movimentacao:create` | ✓ | | | automático |
| `estoque:movimentacao:adjust` | ✓ | | | automático |
| `estoque:movimentacao:read` | ✓ | ✓ | ✓ | automático |
| `estoque:saldo:read` | ✓ | ✓ | ✓ | automático |

`ASSOCIADO`, `TESOURARIA`, `EVENTOS`, `ADMIN_SISTEMA`: nenhuma permissão de `estoque` no V1.

---

## Ordem real de execução entre sub-specs

A ordem de especificação (`01→02→03→04→05`) coincide com a ordem real de implementação — mais simples que `financeiro`, que teve uma exceção documentada (`01`/T4 dependendo de `02`). Aqui:

1. `01` completa (migration+domínio+infra+2 casos de uso)
2. `02` completa (depende de `01` só pela FK de `produto_id`)
3. `03` completa (depende de `02` pela mesma tabela/repositório)
4. `04` completa (depende de `01`-`03` só para referenciar as constantes `Perm*` já publicadas)
5. `05` completa (depende de `01`-`04`)

Nenhuma PR precisa misturar sub-specs — cada uma é uma unidade de execução própria, mesma regra de `financeiro`.

## Decisões futuras (V2, fora do V1 — não implementar sem nova aprovação)

Ver seção "Fora do V1" acima — repetida aqui por paralelismo com `financeiro/STATE.md`: reserva, integração real com `financeiro`/`loja`, custo de aquisição, estoque mínimo/alertas, múltiplos locais, quantidade fracionária, variação estruturada, catálogo comercial.

## Progresso de implementação

- **`01-produtos`/`T1`**: migration `000008_estoque_produtos` (tabela `produtos_estoque`, `codigo UNIQUE NOT NULL`, `GRANT INSERT, SELECT` — nunca `UPDATE`/`DELETE`, `EST-D-007`). Teste de migração dedicado (`estoque_produtos_migration_test.go`, 5 testes: coluna armazenada, unicidade de código, grants corretos, down/up roundtrip, falha clara sem `tj_app`). Nenhuma alteração em outras migrations — `produtos_estoque` não é referenciada por nenhuma FK ainda. Gate completo (full, integration, lint, arquitetura) 100% verde, sem exceção.

## Handoff

- **Fase**: `Specify`+`Discuss`+`Design`+`Tasks` concluídos para as 5 sub-specs de `estoque` (`01-produtos`, `02-movimentacoes`, `03-ajustes-e-saldo`, `04-permissoes`, `05-api-http`), validados por `validate_spec.py` (0 erros/0 avisos nas 5) e `validate_tasks.py` (0 erros nas 5; avisos de granularidade aceitos, mesma categoria dos já aceitos em `financeiro`). `EST-D-001` a `EST-D-011` fechadas pelo mantenedor em 2026-10-03 (`EST-D-008`-`EST-D-010` são achados lógicos derivados das regras já fechadas, não decisões de negócio novas — registrados com justificativa explícita, nunca escolhidos silenciosamente). Nenhum código, migration, branch, commit ou PR criado nesta rodada.
- **Nada implementado ainda**: `api/internal/estoque/` não existe no código — só em `.specs/features/estoque/`.
- **Bloqueios**: nenhum — todas as decisões de negócio necessárias para `01`-`05` estão fechadas acima.
- **Próximo passo**: aguardar autorização explícita do mantenedor para iniciar a implementação (`T1`-`T5` de `01-produtos`, na ordem real de execução descrita acima), seguindo o mesmo protocolo autônomo já usado em `financeiro` (tests-first, mutação quando aplicável, gate check por tarefa, um commit atômico por tarefa, PR ao final de cada sub-spec, merge sempre por autorização separada).
