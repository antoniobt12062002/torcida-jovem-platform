# Financeiro — STATE

Documento de coordenação da feature `financeiro`, decomposta em 5 sub-specs. Mantém enxuto o `.specs/STATE.md` do projeto: aqui ficam as decisões e o mapa específicos desta feature; lá fica só uma linha apontando para cá, mais qualquer decisão que vire padrão para módulos futuros (nenhuma ainda).

Convenção de numeração local: **FIN-D-NNN** para decisões que valem só para `financeiro`. Nunca reaproveita o numerador `AD-NNN` do `.specs/STATE.md` global, que é reservado a decisões de arquitetura de todo o projeto.

---

## Mapa das sub-specs

| # | Spec | Requisitos | Status |
|---|---|---|---|
| 01 | `spec/01-plano-de-contas.md` | PC-01 a PC-04 | In Progress (T1, T2, T3 concluídas) |
| 02 | `spec/02-lancamentos.md` | LAN-01 a LAN-04 | Concluída (T1-T6) |
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

### Estrutura de pacotes (decisão do mantenedor, 2026-09-29)

As 5 sub-specs são unidades **documentais e de execução** (spec, design, tasks, testes, PR) — nunca pacotes Go independentes. `financeiro` segue o esqueleto já documentado em `docs/architecture/architecture-overview.md`, o mesmo de `identity`:

```
api/internal/financeiro/
├── module.go     # financeiro.New(...) e financeiro.Contribution() (05), quando existirem
├── domain/       # entidades e erros — um arquivo por conceito (conta.go, lancamento.go), não por sub-spec
├── app/          # um arquivo por caso de uso (criar_conta.go, criar_lancamento.go, receber_lancamento.go, anexar_comprovante.go, ...), mesma granularidade de identity/app; usecase.go compartilhado (Authorizer/Auditor/TxFunc)
├── infra/        # um arquivo por repositório/porta concreta (conta_repository.go, lancamento_repository.go, lancamento_existence_checker.go, ...)
└── http/         # vazio até existir uma tarefa de rota HTTP — nenhuma das 5 sub-specs do V1 inclui isso ainda
```

Isso corrige um caminho equivocado que estava nos `tasks/0N-*.md` originais (pacotes flat por sub-spec, ex. `financeiro/planocontas/`, `financeiro/lancamentos/`) — passou por 4 rodadas de revisão sem eu sinalizar a tensão com `architecture-overview.md`, encontrado só ao começar a `T2` de `01`. `T1` (migração) não precisou de nenhuma reorganização: só tocou `api/migrations/` e o teste de migração em `platform/database` (mesmo padrão já usado por `documents`/`identity`/`password_reset`), nunca um pacote `financeiro/`. Todos os campos `Where` de `tasks/01-05` já foram corrigidos para os caminhos reais acima.

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
**Decisão**: `financeiro:lancamento:receive` e `financeiro:lancamento:pay` são permissões distintas, sem uma genérica `settle`. Refletido também no catálogo de auditoria: ações `lancamento.receive` e `lancamento.pay` distintas (formato `dominio.verbo` exigido por `platform/audit.Register`; nomes corrigidos de `financeiro.lancamento.receive`/`.pay`, que tinham três segmentos — achado técnico da T2, ver Handoff).
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

`CONSELHO_FISCAL` nunca recebe `create`/`update`/`deactivate`/`receive`/`pay`/`cancel` — todas as 6 hoje impedidas em código por `roles_matrix.go`'s `forbiddenToConselhoFiscal` (`fix(identity): reforca a invariante do Conselho Fiscal no RBAC`, PR #40, mesclado em `develop`). Até esse reforço, só `create`/`update`/`cancel` eram cobertas (achado da revisão de consistência das sub-specs); `deactivate`/`receive`/`pay` foram adicionadas à lista especificamente por causa desse achado, fechando a lacuna antes do V1 do financeiro usá-las de verdade. `ASSOCIADO`, `ESTOQUE_LOJA`, `EVENTOS`, `ADMIN_SISTEMA`: nenhuma permissão de `financeiro` no V1. As 3 permissões institucionais (FIN-D-014) não aparecem nesta tabela — nenhum papel ganha caso de uso novo por causa delas nesta rodada.

---

## Ordem real de execução entre sub-specs

A ordem de **especificação** (`01→02→03→04→05`, já revisada e aprovada) não é idêntica à ordem real de **implementação de código nem à de PRs**, porque algumas tarefas têm pré-requisitos de schema entre sub-specs que a numeração local de cada `tasks/0N-*.md` não expressa (o validador de tasks só entende dependências dentro do mesmo arquivo). Consolidado aqui, agora em dois níveis: a ordem de tarefas e o agrupamento em PRs.

### Ordem de tarefas (dependência real)

1. `01/T1` (migração `contas_contabeis`)
2. `02/T1` (migração `lancamentos` — depende de `01/T1` pela FK)
3. `01/T2`, `01/T3` (criar, listar, desativar conta)
4. `02/T6` (implementação concreta de `LancamentoExistenceChecker`, depende de `02/T1`)
5. `01/T4` (renomear conta — só testável de ponta a ponta com `02/T1` e `02/T6` já existindo)
6. `02/T2`, `02/T3`, `02/T4`, `02/T5` (criar, editar, devolução, listar lançamento — `T4`/devolução usa fixture SQL direta para a pré-condição `RECEBIDA` em teste, **sem dependência real de código de `03`**, apesar de fazer sentido temático depois de `03/T1` existir)
7. `03/T1`-`T4` (receber, pagar, cancelar, saldo — dependem só de `02/T1`+`T2`)
8. `04/T1`-`T3` (comprovantes — dependem só de `02/T1`+`T2` e de `platform/documents`, já mesclada; **sem dependência de `03`**)
9. `05/T1`-`T3` (institucionais, catálogo operacional, substituição do placeholder — **sem dependência de código real de `01`-`04`**: cada sub-spec referencia os nomes de permissão como literais diretos, mesmo padrão já usado em `identity` — `create_user.go` usa `"identity:user:create"` direto, não importa `roles_matrix.go`; os 13 nomes já estão fixados em `financeiro/STATE.md` desde a especificação. `05` é sequenciada por último por conveniência de verificação — nesse ponto os 13 nomes já foram exercitados de verdade pelos testes de `01`-`04` — não por bloqueio técnico)

### Agrupamento em PRs (uma sub-spec por PR, salvo exceção documentada)

| # | PR | Tarefas | Depende de |
|---|---|---|---|
| 1 | `01` (parte 1) | T1-T3 | nenhuma |
| 2 | `02` (completa) | T1-T6 | PR 1 mesclada |
| 3 | `01` (parte 2) | T4 | PR 2 mesclada |
| 4 | `03` (completa) | T1-T4 | PR 2 mesclada |
| 5 | `04` (completa) | T1-T3 | PR 2 mesclada |
| 6 | `05` (completa) | T1-T3 | nenhuma tecnicamente — sequenciada por último por conveniência |

**Exceção documentada** (a única mistura entre sub-specs em nível de PR, e é uma divisão, não uma mistura): `01` fica em duas PRs porque `T4` só pode ser codificada e testada de ponta a ponta depois que `02/T1` (schema) e `02/T6` (implementação concreta da porta) existirem — não há como evitar isso sem violar a fronteira de ownership (`01` nunca escreve em `02`, `02` nunca escreve em `01`, `FIN-D-008`) ou sem misturar `T4` dentro da PR de `02` (o que misturaria duas sub-specs numa PR, a regra que realmente importa preservar). PR 3, PR 4 e PR 5 não dependem umas das outras — só de PR 2 — e podem ser feitas em qualquer ordem relativa entre si; a ordem 3→4→5 é só a mais conveniente (a menor primeiro), não uma dependência real.
10. `04/T1`, `04/T2`, `04/T3` (comprovantes — dependem de `02/T2`)
11. `05/T1` (permissões institucionais — pode ser feita a qualquer momento, inclusive em paralelo com o restante)
12. `05/T2`, `05/T3` (catálogo operacional completo e substituição do placeholder — por último, depois que `01`-`04` já fixaram seus nomes de permissão)

## Decisões futuras (V2, fora do V1 — não implementar sem nova aprovação)

Parcelamento e pagamento parcial (FIN-D-002); centro de resultado e rateio; fornecedores; orçamento; prestação de contas anual (ativaria FIN-D-014); alertas (ex. despesas sem comprovante); exportações (PDF/Excel/CSV); integração com `estoque` (fluxo compra→despesa+entrada em estoque); gateway de pagamento; portal da transparência; conciliação bancária (FIN-D-004); regime de competência, caso o caixa se mostre insuficiente; `substituido_por_id` ou operação atômica de cancelamento+substituição, caso a janela de saldo do FIN-D-006 se mostrar um problema real na prática; código contábil formal (FIN-D-009), caso necessário depois; versionamento/substituição de comprovante via `Supersedes` (FIN-D-016), caso "corrigir um comprovante já anexado" se mostre uma necessidade real, não só "anexar mais um".

## Handoff

- **Fase**: `01-plano-de-contas` PR 1 (`T1`-`T3`) e `02-lancamentos` PR 2 (`T1`-`T6`) concluídas, todos os commits locais na branch `feature/financeiro-especificacao`; nenhum push, PR ou merge feito. Próxima etapa prevista (aguardando autorização): PR 3 — `01-plano-de-contas/T4` (`RenomearConta`), agora desbloqueada.
- **Progresso `01`**: `T1` (migração `contas_contabeis`), `T2` (`CriarConta`+`ListarContas`) e `T3` (`DesativarConta`) concluídas — tests-first, mutação e gates completos por tarefa, um commit atômico cada. `T4` (`RenomearConta`) fica para depois que `02/T1` (tabela `lancamentos`) existir — agora existe (abaixo).
- **Progresso `02`**: `T1`-`T6` concluídas — tests-first, mutação e gates completos por tarefa. `T1` ajustou também os testes de down-migration de `01` (`financeiro_contas_migration_test.go`) e de `identity` (`identity_migration_test.go`) para desfazer `000007_financeiro_lancamentos` antes de `contas_contabeis`/`users`, já que `lancamentos` referencia as duas por FK. `T2` acrescentou `env_test.go`: `newUser`/`userActor`, porque `lancamentos.criado_por` é uma FK real para `users` (diferente de `contas_contabeis`, sem essa FK) — `actor()` (UUID sintético) continua válido só para os casos de uso de `01`. `T3` (decisão do mantenedor, registrada como `spec/02-lancamentos.md` LAN-02 AC5): a edição reaplica a mesma validação de conta de LAN-01 AC2/AC3 (existe, ativa, mesmo tipo) sempre que `conta_id` é fornecido — a invariante de FIN-D-015 vale para todo o ciclo de vida do lançamento, não só na criação. `T4` reusa `PermLancamentoCreate`/`ActionLancamentoCreate` (FIN-D-003: devolução não é operação distinta), com `devolucao_de_id` no payload `After`; `Tipo` é sempre `DESPESA`, fixado pelo caso de uso, não recebido como input. `T5` ordena por `criado_em`. `T6` (decisão do mantenedor): implementa só o tipo concreto `infra.LancamentoExistenceChecker` (`TemLancamento`), sem criar `financeiro/module.go` — nenhuma tarefa aprovada de `01`/`02` previu esse composition root; o wiring real fica pendente para quando `01-plano-de-contas/T4` (ou a camada HTTP) precisar dele. Testado diretamente em `infra_test` (primeiro teste de `financeiro/infra`, mesmo padrão de `identity/infra`), já que não há caso de uso em `app` que o exercite ainda.
- **Bloqueios**: nenhum — todas as decisões de negócio necessárias para especificar o V1 estão fechadas acima.
- **Nota técnica (achado da T2 de `01`, resolvido)**: `platform/audit.Register` exige o formato de duas partes `dominio.verbo` (ex. `user.create`, `document.create`). Os nomes de três partes que apareciam em `design/02-lancamentos.md`, `spec/03-workflow-e-saldo.md`, `design/03-workflow-e-saldo.md`, `tasks/03-workflow-e-saldo.md` e nesta FIN-D-011 foram corrigidos para o formato aceito. Catálogo final de ações de `lancamento`: `lancamento.create` (LAN-01 AC5; reaproveitada por `CriarDevolucao`, LAN-03 — devolução não é operação de negócio distinta, FIN-D-003; `devolucao_de_id` vai no payload `After`), `lancamento.update` (LAN-02 AC4), `lancamento.receive`/`lancamento.pay`/`lancamento.cancel` (`03-workflow-e-saldo`, fora do escopo desta PR). `owner_type="financeiro.lancamento"` usado por `platform/documents` (FIN-D-005, `04-comprovantes`) é um namespace diferente (formato próprio de `documents`, não de `audit.Action`) e não precisou de ajuste.
