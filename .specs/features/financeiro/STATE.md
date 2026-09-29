# Financeiro — STATE

Documento de coordenação da feature `financeiro`, decomposta em 5 sub-specs. Mantém enxuto o `.specs/STATE.md` do projeto: aqui ficam as decisões e o mapa específicos desta feature; lá fica só uma linha apontando para cá, mais qualquer decisão que vire padrão para módulos futuros (nenhuma ainda).

Convenção de numeração local: **FIN-D-NNN** para decisões que valem só para `financeiro`. Nunca reaproveita o numerador `AD-NNN` do `.specs/STATE.md` global, que é reservado a decisões de arquitetura de todo o projeto.

---

## Mapa das sub-specs

| # | Spec | Requisitos | Status |
|---|---|---|---|
| 01 | `spec/01-plano-de-contas.md` | PC-01 a PC-04 | In Design |
| 02 | `spec/02-lancamentos.md` | LAN-01 a LAN-04 | In Design |
| 03 | `spec/03-workflow-e-saldo.md` | WKF-01 a WKF-03 | In Design |
| 04 | `spec/04-comprovantes.md` | CMP-01 a CMP-03 | In Design |
| 05 | `spec/05-permissoes.md` | PERM-01 a PERM-03 | In Design |

## Dependências

```
01-plano-de-contas
       │  (conta_id)
       ▼
02-lancamentos ──┬──► 03-workflow-e-saldo   (regras de status sobre a mesma entidade)
                 └──► 04-comprovantes        (owner_id)

02-lancamentos ⇢ 01-plano-de-contas   (leitura pontual, ver FIN-D-008; nunca escrita)

05-permissoes: transversal — citada por 01-04, não depende de nenhuma.
```

Nenhuma dependência circular real: a seta tracejada `02⇢01` é leitura, não escrita, e não inverte a ordem de implementação (`01` continua antes de `02`).

## Fronteiras de responsabilidade

| Dado/operação | Dona | Nota |
|---|---|---|
| `ContaContabil` | `01` | única spec que escreve |
| `LancamentoFinanceiro` (todos os campos, incluindo `devolucao_de_id`) | `02` | única spec que insere/edita a linha |
| Regras de transição de `status` e cálculo de saldo | `03` | a coluna `status` é fisicamente parte da entidade de `02`; a regra de quando ela muda é de `03` |
| Existência de documento anexado a um lançamento | `platform/documents` | não é dado do módulo `financeiro` |
| Convenção de `owner_type`/permissão para consultar isso | `04` | documenta a convenção, não guarda o dado |
| Catálogo de permissões e matriz de papéis | `05` | única spec que define a `Contribution` |

`01` e `02` são sub-specs do mesmo módulo Go (`api/internal/financeiro/`), não módulos de negócio separados — a disciplina acima é sobre ownership de dado (quem escreve em quê), não uma fronteira equivalente à de `architecture_test.go` entre `financeiro` e `estoque`.

---

## Decisões fechadas (FIN-D-NNN)

### FIN-D-001 — Regime de caixa
**Decisão**: o V1 opera em regime de caixa. O saldo considera só dinheiro efetivamente recebido ou pago. Estados-base: Receita `CRIADA→RECEBIDA`; Despesa `CRIADA→PAGA`. Sem estados de competência/confirmação no V1.
**Escopo**: `03-workflow-e-saldo`.

### FIN-D-002 — Sem pagamento parcial
**Decisão**: um lançamento é liquidado integralmente. Pagamento/recebimento parcial e parcelamento ficam para evolução futura (V2).
**Escopo**: `02-lancamentos`, `03-workflow-e-saldo`.

### FIN-D-003 — Devolução
**Decisão**: não há mecanismo formal de "estorno". Devolução de dinheiro já recebido corretamente é registrada como uma **nova despesa** com o campo `devolucao_de_id` apontando para a receita original. Relacionamento **unidirecional** — só a despesa carrega o campo; a receita original nunca é escrita por causa disso, permanecendo `RECEBIDA`. O campo só pode ser preenchido quando: o lançamento atual é `DESPESA`; o lançamento referenciado é `RECEITA`; a receita referenciada está `RECEBIDA`.
**Escopo**: `02-lancamentos`.

### FIN-D-004 — Conciliação bancária fora do V1
**Decisão**: nenhuma importação de extrato, matching ou mesmo comparação simplificada de saldo bancário no V1.
**Escopo**: `03-workflow-e-saldo` (não implementa nada disso).

### FIN-D-005 — "Com comprovante" é condição derivada
**Decisão**: a existência de comprovante não é um estado do lançamento — é determinada, sob demanda, consultando `platform/documents` (owner_type="financeiro.lancamento"). `04-comprovantes` nunca escreve no `status` do lançamento.
**Escopo**: `03-workflow-e-saldo` (não modela o estado), `04-comprovantes` (não escreve em `02`).

### FIN-D-006 — Cancelamento, inclusive de lançamento já liquidado
**Decisão**: cancelamento é estado terminal, alcançável de `CRIADA`, `RECEBIDA` ou `PAGA`, sempre com motivo obrigatório e auditado. Cancelar um lançamento `RECEBIDA`/`PAGA` **pode produzir uma diferença temporária entre o saldo calculado pelo sistema e o caixa físico**, até que o lançamento correto seja criado e, quando aplicável, liquidado — essa consequência é aceita no V1. Nenhum mecanismo de vínculo obrigatório entre o cancelamento e o lançamento substituto (`substituido_por_id` e equivalentes ficam fora). Cancelamento nunca é tratado como devolução — devolução é sempre uma nova despesa (FIN-D-003), nunca o cancelamento da receita original.
**Escopo**: `03-workflow-e-saldo`.

### FIN-D-007 — Edição
**Decisão**: um lançamento pode ser editado só enquanto `CRIADA`. Depois de `RECEBIDA`/`PAGA`, é imutável. Erro pós-liquidação: cancelar o original (motivo obrigatório, FIN-D-006) + criar um novo lançamento correto.
**Escopo**: `02-lancamentos` (permite a edição), `03-workflow-e-saldo` (define o limite pelo `status`).

### FIN-D-008 — Renomeação de conta contábil
**Decisão**: uma conta pode ser renomeada enquanto nunca tiver sido usada em nenhum lançamento; depois do primeiro uso, o nome é imutável. A regra e o caso de uso pertencem a `01`. A implementação consulta, por uma **porta/interface de leitura** definida em `01` (ex. `LancamentoExistenceChecker`), se existe algum lançamento para aquela conta — a implementação concreta dessa porta pode consultar a tabela de `02`, mas **`01` nunca escreve em `02`, `02` nunca escreve em `01`**. Sem flag desnormalizada em `ContaContabil`, sem gatilho de banco para esta regra. Desativação (distinta de renomeação) nunca afeta a exibição de lançamentos históricos — só impede escolher a conta em um lançamento novo.
**Escopo**: `01-plano-de-contas` (dona da regra e da porta), `02-lancamentos` (implementa a porta, sem escrever em `01`).

### FIN-D-009 — Sem código contábil formal
**Decisão**: nenhum código hierárquico (tipo "1.1.01") no V1. A hierarquia por `id`/`parent_id` é suficiente.
**Escopo**: `01-plano-de-contas`.

### FIN-D-010 — Permissões de lançamento: create e update separadas
**Decisão**: `financeiro:lancamento:create` e `financeiro:lancamento:update` são permissões distintas desde o início, mesmo que no V1 o mesmo papel (`TESOURARIA`) receba as duas.
**Escopo**: `02-lancamentos`, `05-permissoes`.

### FIN-D-011 — Permissões de liquidação: receive e pay separadas
**Decisão**: `financeiro:lancamento:receive` e `financeiro:lancamento:pay` são permissões distintas, sem uma genérica `settle`. Refletido também no catálogo de auditoria: ações `financeiro.lancamento.receive` e `financeiro.lancamento.pay` distintas.
**Escopo**: `03-workflow-e-saldo`, `05-permissoes`.

### FIN-D-012 — Diretoria e comprovantes
**Decisão**: `DIRETORIA` recebe `financeiro:comprovante:read`. Não recebe nenhuma permissão de criação, alteração ou exclusão de lançamentos por causa disso.
**Escopo**: `04-comprovantes`, `05-permissoes`.

### FIN-D-013 — Permissão própria de saldo
**Decisão**: `financeiro:saldo:read` é uma permissão própria, distinta de `financeiro:lancamento:read` — o saldo é uma consulta com regra própria (FIN-D-001).
**Escopo**: `03-workflow-e-saldo`, `05-permissoes`.

### FIN-D-014 — Permissões institucionais preservadas, fora do escopo funcional do V1
**Decisão**: `financeiro:prestacao_contas:read`, `financeiro:prestacao_contas:approve` e `financeiro:parecer:opine` já existem em `identity/app/roles_matrix.go` (`FoundationContributions`), concedidas a `CONSELHO_FISCAL` desde a fundação. Quando `financeiro` publicar sua própria `Contribution`, ela precisa **preservá-las exatamente como estão** (mesmo nome, mesmo papel), substituindo o placeholder — sem criar nenhum caso de uso ou endpoint que as utilize nesta rodada. Pertencem a uma futura sub-spec de prestação de contas (V2).
**Escopo**: `05-permissoes`.

### FIN-D-015 — Consistência entre tipo do lançamento e tipo da conta
**Decisão**: um lançamento `RECEITA` só pode usar uma conta `RECEITA`; um lançamento `DESPESA` só pode usar uma conta `DESPESA`. É regra de negócio do lançamento (não um detalhe técnico incidental), validada em aplicação antes de gravar, com o erro `conta_tipo_incompativel` quando violada.
**Escopo**: `02-lancamentos` (LAN-01 AC3).

### FIN-D-016 — Comprovantes independentes, sem versionamento
**Decisão**: no V1, cada comprovante anexado a um lançamento representa uma evidência própria e independente. Não há cadeia de versões entre comprovantes; `Supersedes` (mecanismo de versionamento já existente em `platform/documents`) nunca é usado por `financeiro` nesta feature. A listagem trata cada comprovante como um registro independente (`Version 1` sempre, do ponto de vista de `platform/documents`). Se um comprovante for anexado por engano, o correto é anexado como um novo documento independente — não substitui nem encadeia o anterior. Substituição/versionamento de comprovante, se necessário no futuro, será uma evolução específica, com nova decisão e novos requisitos.
**Escopo**: `04-comprovantes` (CMP-01 AC5, CMP-03).

---

## Matriz de permissões do V1 (consolidada — ver detalhe e rationale em `05-permissoes`)

| Permissão | `TESOURARIA` | `DIRETORIA` | `CONSELHO_FISCAL` | `PRESIDENTE` |
|---|---|---|---|---|
| `financeiro:conta:create` | ✓ | | | automático |
| `financeiro:conta:update` | ✓ | | | automático |
| `financeiro:conta:deactivate` | ✓ | | | automático |
| `financeiro:conta:read` | ✓ | ✓ | ✓ | automático |
| `financeiro:lancamento:create` | ✓ | | | automático |
| `financeiro:lancamento:update` | ✓ | | | automático |
| `financeiro:lancamento:read` | ✓ | ✓ | ✓ | automático |
| `financeiro:lancamento:receive` | ✓ | | | automático |
| `financeiro:lancamento:pay` | ✓ | | | automático |
| `financeiro:lancamento:cancel` | ✓ | | | automático |
| `financeiro:saldo:read` | ✓ | ✓ | ✓ | automático |
| `financeiro:comprovante:create` | ✓ | | | automático |
| `financeiro:comprovante:read` | ✓ | ✓ | ✓ | automático |

`CONSELHO_FISCAL` nunca recebe `create`/`update`/`deactivate`/`receive`/`pay`/`cancel`. **Atenção, achado da revisão de consistência**: só `create`/`update`/`cancel` são impedidos em código, por `roles_matrix.go`'s `forbiddenToConselhoFiscal = []string{"create", "update", "delete", "cancel"}` — `deactivate`, `receive` e `pay` **não estão** nessa lista hoje. Para essas três, a garantia de que `CONSELHO_FISCAL` nunca as recebe depende inteiramente de `financeiro`'s própria `Contribution` simplesmente não as conceder — nada em `BuildMatrix` bloqueiria um erro futuro que concedesse `financeiro:lancamento:receive` ao Conselho Fiscal, por exemplo. Ver detalhe em `05-permissoes`. `ASSOCIADO`, `ESTOQUE_LOJA`, `EVENTOS`, `ADMIN_SISTEMA`: nenhuma permissão de `financeiro` no V1. As 3 permissões institucionais (FIN-D-014) não aparecem nesta tabela — nenhum papel ganha caso de uso novo por causa delas nesta rodada.

---

## Ordem real de execução entre sub-specs

A ordem de **especificação** (`01→02→03→04→05`, já revisada e aprovada) não é idêntica à ordem real de **implementação de código**, porque algumas tarefas têm pré-requisitos de schema entre sub-specs que a numeração local de cada `tasks/0N-*.md` não expressa (o validador de tasks só entende dependências dentro do mesmo arquivo). Ordem real recomendada, consolidada aqui para não se perder:

1. `01/T1` (migração `contas_contabeis`)
2. `02/T1` (migração `lancamentos` — depende de `01/T1` pela FK)
3. `01/T2`, `01/T3` (criar, listar, desativar conta)
4. `02/T6` (implementação concreta de `LancamentoExistenceChecker`, depende de `02/T1`)
5. `01/T4` (renomear conta — só testável de ponta a ponta com `02/T1` e `02/T6` já existindo)
6. `02/T2`, `02/T3`, `02/T5` (criar, editar, listar lançamento)
7. `03/T1`, `03/T2` (receber, pagar — dependem de `02/T2`)
8. `02/T4` (devolução — teste usa fixture SQL direta para a pré-condição `RECEBIDA`, não depende de código de `03`, mas faz mais sentido depois de `03/T1` existir de verdade)
9. `03/T3`, `03/T4` (cancelar, saldo)
10. `04/T1`, `04/T2`, `04/T3` (comprovantes — dependem de `02/T2`)
11. `05/T1` (permissões institucionais — pode ser feita a qualquer momento, inclusive em paralelo com o restante)
12. `05/T2`, `05/T3` (catálogo operacional completo e substituição do placeholder — por último, depois que `01`-`04` já fixaram seus nomes de permissão)

## Decisões futuras (V2, fora do V1 — não implementar sem nova aprovação)

Parcelamento e pagamento parcial (FIN-D-002); centro de resultado e rateio; fornecedores; orçamento; prestação de contas anual (ativaria FIN-D-014); alertas (ex. despesas sem comprovante); exportações (PDF/Excel/CSV); integração com `estoque` (fluxo compra→despesa+entrada em estoque); gateway de pagamento; portal da transparência; conciliação bancária (FIN-D-004); regime de competência, caso o caixa se mostre insuficiente; `substituido_por_id` ou operação atômica de cancelamento+substituição, caso a janela de saldo do FIN-D-006 se mostrar um problema real na prática; código contábil formal (FIN-D-009), caso necessário depois; versionamento/substituição de comprovante via `Supersedes` (FIN-D-016), caso "corrigir um comprovante já anexado" se mostre uma necessidade real, não só "anexar mais um".

## Handoff

- **Fase**: documentação das 5 sub-specs (spec/design/tasks) em criação nesta sessão, branch `feature/financeiro-especificacao`.
- **Próximo passo**: revisão humana da estrutura documental completa; código não iniciado.
- **Bloqueios**: nenhum — todas as decisões de negócio necessárias para especificar o V1 estão fechadas acima.
