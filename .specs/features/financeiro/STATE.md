# Financeiro — STATE

Documento de coordenação da feature `financeiro`, decomposta em 6 sub-specs (`01`-`05` o domínio V1, `06` sua camada HTTP). Mantém enxuto o `.specs/STATE.md` do projeto: aqui ficam as decisões e o mapa específicos desta feature; lá fica só uma linha apontando para cá, mais qualquer decisão que vire padrão para módulos futuros (nenhuma ainda).

Convenção de numeração local: **FIN-D-NNN** para decisões que valem só para `financeiro`. Nunca reaproveita o numerador `AD-NNN` do `.specs/STATE.md` global, que é reservado a decisões de arquitetura de todo o projeto.

---

## Mapa das sub-specs

| # | Spec | Requisitos | Status |
|---|---|---|---|
| 01 | `spec/01-plano-de-contas.md` | PC-01 a PC-04 | Concluída (T1-T4) |
| 02 | `spec/02-lancamentos.md` | LAN-01 a LAN-04 | Concluída (T1-T6) |
| 03 | `spec/03-workflow-e-saldo.md` | WKF-01 a WKF-03 | Concluída (T1-T4) |
| 04 | `spec/04-comprovantes.md` | CMP-01 a CMP-03 | Concluída (T1-T3) |
| 05 | `spec/05-permissoes.md` | PERM-01 a PERM-03 | Concluída (T1-T3) |
| 06 | `spec/06-api-http.md` | API-01 a API-06 | Especificada e desenhada (Specify+Design+Tasks); implementação não iniciada |

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

### FIN-D-017 — Onde `financeiro.Contribution()` mora
**Decisão**: `financeiro.Contribution()` (a contribuição do módulo para o RBAC agregado, `05-permissoes`) fica no pacote raiz do módulo, em `api/internal/financeiro/module.go` — o mesmo arquivo destinado também a `financeiro.New(...)` quando essa composição existir. Não fica em `financeiro/app` (ex. um `roles_matrix.go` local): `app/` contém casos de uso e portas, não a contribuição do módulo para o mecanismo de RBAC. `identity/app/roles_matrix.go` continua sendo o **consumidor/agregador** (`BuildMatrix`), nunca o dono da contribuição de `financeiro` — o placeholder que ele carrega hoje (`FoundationContributions`) é removido quando `05/T3` publicar `financeiro.Contribution()` de verdade. Achado durante a análise de `05-permissoes`: `design/05-permissoes.md` comparava `financeiro.Contribution()` a "`identity.New`/`FoundationContributions`" como se fossem o mesmo padrão — não são (`identity.New` fica no pacote raiz `identity`, `FoundationContributions` fica em `identity/app`); `design/05-permissoes.md` precisa ser corrigido antes de `05/T2`/`T3` para registrar isso explicitamente.
**Escopo**: `05-permissoes` (T2, T3).
**Nota**: a correção de `design/05-permissoes.md` apontada aqui foi antecipada para `T1` (commit estrutural), porque a mesma análise revelou o achado de `FIN-D-018` abaixo — os dois corrigiam a mesma seção do documento.

### FIN-D-018 — `financeiro` não importa `identity/domain`; papéis reexportados via `identity/app`
**Decisão**: `financeiro.Contribution()` precisa montar `Grants map[domain.Role][]authz.Permission` (tipo de `identity/app.Contribution`), mas `financeiro` não pode importar `identity/domain` diretamente — `api/internal/architecture_test.go` proíbe um módulo de importar o `domain` (ou `infra`/`http`) de outro módulo; só o `app` publicado e o pacote raiz são permitidos (`TestAllowsAModuleUsingAnotherModulesPublishedApp`). Resolução: `identity/app` reexporta `Role` como alias de `identity/domain.Role` (`type Role = domain.Role`, não uma nova definição) e as 8 constantes de papel (`identity/app/roles.go`), referenciando as constantes já existentes — nunca duplicando o valor como string literal (`platform/authz/rolename_guard_test.go` proíbe qualquer string literal igual a um nome de papel fora de `identity/domain/role.go`, em qualquer arquivo do repositório). `financeiro/module.go` usa `app.Role`/`app.RoleConselhoFiscal` etc. Sem ciclo de import: `identity` (app/domain/infra) não referencia `financeiro` em nenhum ponto hoje. Zero mudança de comportamento de RBAC — alias puro, `BuildMatrix` e `FoundationContributions()` inalterados nesta decisão.
**Achado durante a análise de `05-permissoes`**: `design/05-permissoes.md` (seção Components) e `tasks/05-permissoes.md` (T1, campo `Reuses`) presumiam que `financeiro` importaria `identity/domain.RoleConselhoFiscal` diretamente — já corrigidos nesta mesma rodada (commit estrutural de `T1`).
**Escopo**: `05-permissoes` (T1, T2, T3 — toda tarefa que monte um `Grants` usa o re-export).

### FIN-D-019 — As 13 permissões operacionais reaproveitam as constantes de `financeiro/app` (Opção A)
**Decisão**: `financeiro.Contribution()` declara as 13 permissões operacionais referenciando diretamente as constantes `Perm*` já declaradas pelos casos de uso de `01`-`04` em `financeiro/app` (`finapp.PermContaCreate` etc., com alias de import `finapp` para não colidir com `app` de `identity/app`) — nunca as redigitando como strings literais independentes em `module.go`. Decisão do mantenedor (`05-permissoes/T2`), escolhida sobre a alternativa de literais autocontidos (Opção B) porque elimina estruturalmente qualquer risco de drift/typo entre o catálogo de RBAC e o que `platform/authz.Require` de fato verifica em cada caso de uso — uma renomeação na declaração original passa a falhar a compilação de `module.go`, em vez de divergir silenciosamente.
**Consequência arquitetural**: cria a dependência `financeiro` (pacote raiz) → `financeiro/app`, interna ao próprio módulo. Confirmado sem ciclo (`financeiro/app` não importa o pacote raiz `financeiro` em nenhum ponto) e sem violação de `architecture_test.go` (a regra só restringe um módulo de importar `domain`/`infra`/`http` de **outro** módulo; pacote raiz importando o `app` do próprio módulo não é restringido por nenhum caso do teste — confirmado por `TestProductionCodeRespectsTheModuleBoundaries` passando). `design/05-permissoes.md` atualizado para registrar essa dependência explicitamente.
**Escopo**: `05-permissoes` (T2).

### FIN-D-020 — T3 (remoção do placeholder) exigiu ajustar 10 arquivos de teste fora de `financeiro/`, sem nenhuma regra nova
**Decisão/achado**: ao remover o bloco `financeiro` de `FoundationContributions()` (T3), `TESOURARIA`/`DIRETORIA`/`CONSELHO_FISCAL` passaram a só ter conteúdo real de RBAC quando a matriz é construída como em produção — `FoundationContributions()` (identity, audit) **mais** `financeiro.Contribution()`. Qualquer ambiente de teste que construísse a matriz só com `FoundationContributions()` (prática comum antes da T3, quando isso já bastava) passou a divergir da produção. Dois tipos de ajuste, nenhum deles uma regra nova — só restauração de fidelidade ao comportamento já aprovado:
1. **Ambiente de teste desatualizado** (`identity/app/{session_service_test.go,usecase_env_test.go}`, `identity/http/env_test.go`, `httpapi/{e2e_test.go,router_integration_test.go}`, e o helper `matrix(t)` de `identity/infra/role_repository_test.go`): passam a agregar `financeiro.Contribution()` junto de `FoundationContributions()`, igual aos dois composition roots reais. Sem isso, `PRESIDENTE` e `ADMIN_SISTEMA` ficavam com permissões idênticas nesses ambientes (a vantagem de `PRESIDENTE` vinha inteiramente das permissões de `financeiro`), quebrando testes de "ator não pode agir sobre alguém mais poderoso" (`identity/app`, `identity/http`, `httpapi`) sem nenhuma mudança real de comportamento em produção.
2. **Fixture que usava o placeholder como "permissão descartável"** (`identity/app/roles_matrix_test.go`: 2 testes cujo `want` incluía as 3 institucionais via `FoundationContributions()` isolada — reduzido para refletir que elas não vêm mais de lá; `identity/infra/role_repository_test.go`: 3 testes que filtravam `c.Module != "financeiro"` para simular "uma permissão saiu do código" — substituído por `removableFixture()`, uma `Contribution` sintética de módulo `estoque` com 3 permissões, sem nenhuma relação com `financeiro` real).
3. **Achado genuíno de negócio, com decisão do mantenedor** (não uma correção mecânica): `identity/app/{promote_admin_test.go,admin_reset_password_test.go}` usavam `domain.RoleTesouraria` como "papel sem conteúdo de RBAC" para provar que `ADMIN_SISTEMA` pode agir livremente sobre ele — verdade só porque `TESOURARIA` nunca teve permissão alguma antes de `05-permissoes`. Com as 13 permissões reais, `ADMIN_SISTEMA` (sem nenhuma permissão de `financeiro`) deixou de cobrir `TESOURARIA` — `authz.Covers`, mecanismo preexistente e intocado, passou a recusar corretamente pela primeira vez. Apresentadas duas alternativas (trocar o papel-alvo do fixture vs. documentar a nova recusa como comportamento real); o mantenedor escolheu **trocar o fixture para `domain.RoleEventos`** (ainda genuinamente sem nenhuma `Contribution`, convenção já usada nos mesmos arquivos) — preserva a intenção original do teste, não introduz nem documenta nenhuma mudança de comportamento de `ADMIN_SISTEMA`.
4. **Lacuna de cobertura real, achada por mutação**: nenhum teste provava que `cmd/bootstrap-admin/main.go` de fato sincroniza as permissões de `financeiro` no banco — um mutante que removia inteiramente a importação e o wiring passava limpo. Adicionada uma asserção em `TestBootstrapAdminWorksOnAFreshDatabaseBeforeTheAPIEverStarted` (`cmd/bootstrap-admin/bootstrap_integration_test.go`), contando `permissions WHERE name LIKE 'financeiro:%'` (= 16). `cmd/api` não tem nenhum arquivo de teste (condição pré-existente, não criada por esta tarefa); a omissão de wiring lá já é pega em tempo de compilação (import não utilizado), verificado manualmente por mutação.

Nenhuma alteração de comportamento de produção nesta tarefa além da já prevista (placeholder removido, `financeiro.Contribution()` wired nos dois composition roots) — confirmado por `git diff`: só `identity/app/roles_matrix.go`, `cmd/api/main.go` e `cmd/bootstrap-admin/main.go` mudam fora de arquivos `_test.go`.
**Escopo**: `05-permissoes` (T3).

### FIN-D-021 — Paginação/filtro/ordenação ficam fora do escopo de `06-api-http`
**Decisão**: `ListarContas` e `ListarLancamentos` (`01`/`02`) não suportam paginação, filtro ou ordenação — devolvem a lista inteira, sempre, confirmado lendo os dois casos de uso por completo. `06-api-http` expõe isso fielmente (`GET` sem parâmetros, resposta sem `next_cursor`), sem estender nenhum caso de uso existente para adicionar esses recursos — fora do escopo desta sub-spec. Lançamentos crescem sem limite ao longo do tempo; isso é uma lacuna de escala conhecida, não resolvida aqui.
**Escopo**: `06-api-http` (`API-01` AC2, `API-02` AC4). Estender os casos de uso com paginação, se necessário, é uma decisão futura separada.

### FIN-D-022 — Tabela de mapeamento de erros de `06-api-http` (proposta)
**Decisão proposta**: os 11 sentinels de `financeiro/domain` são frases em português (`"conta não encontrada"`), não códigos `snake_case` como os de `identity/domain` (`"email_taken"`) — não podem ser reaproveitados como `code` de `problem+json` diretamente (`identity/http`'s `sentinel.Error()` só funciona porque os sentinels de `identity` já nascem no formato certo). A tabela completa (status + code por erro) está em `design/06-api-http.md`. Pendente de confirmação do mantenedor antes de `06/T3`.
**Escopo**: `06-api-http` (`API-06`).

### FIN-D-023 — `cmd/api` falha ao iniciar sem `STORAGE_ENABLED=true` quando `financeiro` é composto
**Decisão**: diferente do precedente de `email.Sender` "disabled" (sempre funciona via `disabledSender`), `platform/documents.Service` não pode ter uma rota condicionalmente registrada — `AD-012` exige paridade incondicional entre rota e operação do contrato. Resolução: `cmd/api/main.go` exige `STORAGE_ENABLED=true` para compor `financeiro` (falha no startup, mesmo padrão de outras validações obrigatórias de `config.Load()`), nunca um `503` por requisição.
**Escopo**: `06-api-http` (`T1`, `T8`).

### FIN-D-024 — Chain de upload dedicada para o único endpoint de multipart de `financeiro`
**Decisão**: `platform/httpx.BodyLimit(limit)` já é parametrizável; a chain `authenticated` global usa `1 MiB` (`AD-013`/API-02.7), insuficiente para os `10 MiB` já decididos em `platform/documents` (`DOC-01 AC3`). O mecanismo (uma segunda chain, só para a rota de anexar comprovante) está fechado; o valor exato da margem acima de `10 MiB` (proposto `11 MiB`) está pendente de confirmação.
**Escopo**: `06-api-http` (`T7`).

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
9. `05/T1` (permissões institucionais — pode ser feita a qualquer momento, inclusive em paralelo com o restante; **sem dependência de código real de `01`-`04`**)
10. `05/T2`, `05/T3` (catálogo operacional completo e substituição do placeholder — por último, depois que `01`-`04` já fixaram seus nomes de permissão: cada sub-spec referencia os nomes de permissão como literais diretos, mesmo padrão já usado em `identity` — `create_user.go` usa `"identity:user:create"` direto, não importa `roles_matrix.go`; os 13 nomes já estão fixados em `financeiro/STATE.md` desde a especificação. `05/T2`-`T3` são sequenciadas por último por conveniência de verificação — nesse ponto os 13 nomes já foram exercitados de verdade pelos testes de `01`-`04` — não por bloqueio técnico)

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

## Decisões futuras (V2, fora do V1 — não implementar sem nova aprovação)

Parcelamento e pagamento parcial (FIN-D-002); centro de resultado e rateio; fornecedores; orçamento; prestação de contas anual (ativaria FIN-D-014); alertas (ex. despesas sem comprovante); exportações (PDF/Excel/CSV); integração com `estoque` (fluxo compra→despesa+entrada em estoque); gateway de pagamento; portal da transparência; conciliação bancária (FIN-D-004); regime de competência, caso o caixa se mostre insuficiente; `substituido_por_id` ou operação atômica de cancelamento+substituição, caso a janela de saldo do FIN-D-006 se mostrar um problema real na prática; código contábil formal (FIN-D-009), caso necessário depois; versionamento/substituição de comprovante via `Supersedes` (FIN-D-016), caso "corrigir um comprovante já anexado" se mostre uma necessidade real, não só "anexar mais um".

## Handoff

- **Fase**: PRs #41-#51 e #53-#56 mescladas em `develop` (squash + bypass administrativo) — `01-plano-de-contas`, `02-lancamentos`, `03-workflow-e-saldo`, `04-comprovantes` e `05-permissoes` (T1-T3) completas; auditoria final de `financeiro` feita (duas pendências documentais corrigidas pela PR #56). `develop` em `87f4f62`. PR #52 (bump npm de rotina, Dependabot) permanece aberta e intocada, bloqueada por `pnpm audit` numa vulnerabilidade pré-existente em `braces@3.0.3` sem patch upstream — tratada separadamente do fluxo de `financeiro`. **`06-api-http` especificada e desenhada** (branch `feature/financeiro-06-api-http`, a partir de `develop`) — `spec`/`design`/`tasks` completos, validados, nenhuma linha de código de produção escrita. Quatro decisões técnicas (`FIN-D-021` a `FIN-D-024`) pendentes de confirmação do mantenedor antes do início de `T1`/`T3`/`T7` (ver relatório de entrega).
- **Progresso `01`**: Concluída (T1-T4) — tests-first, mutação e gates completos por tarefa, um commit atômico cada. `T4` usa a porta `LancamentoExistenceChecker` declarada em `app/renomear_conta.go` (consumidor) e satisfeita estruturalmente por `infra.LancamentoExistenceChecker` (já implementado em `02/T6`). Corrigida também uma inconsistência documental em `design/01-plano-de-contas.md`: a assinatura da porta estava documentada com `contaID uuid.UUID`, mas todo id de `financeiro` é `string` desde a T2 — `02/T6` já tinha sido implementado corretamente com `string`; só a documentação estava desatualizada.
- **Progresso `03`**: Concluída (T1-T4) — tests-first, mutação e gates completos por tarefa. `T1`/`T2`: `CRIADA→RECEBIDA` só para `RECEITA` e `CRIADA→PAGA` só para `DESPESA` (erro dedicado por direção — `ErrLancamentoNaoPodeSerRecebido`/`ErrLancamentoNaoPodeSerPago`); audita `lancamento.receive`/`lancamento.pay`. `T3`: cancela de `CRIADA`, `RECEBIDA` ou `PAGA` (`ErrLancamentoJaCancelado` se já `CANCELADA`), motivo obrigatório validado antes de qualquer leitura/escrita (`ErrMotivoObrigatorio`), persiste `motivo_cancelamento`/`cancelado_por`/`cancelado_em` (colunas já existentes desde `02/T1`), audita `lancamento.cancel` com o motivo no campo `Reason` — `platform/audit`'s `requiresReason()` reforça isso automaticamente. `FIN-D-006` validada: cancelar um lançamento `RECEBIDA`/`PAGA` é permitido sem nenhuma mitigação. `T4` (`ConsultarSaldo`): consulta pura, sem `Tx`/`Audit` — soma `RECEBIDA` menos soma `PAGA` por status **atual**, numa única query agregada (`COALESCE`+`FILTER`, conjunto vazio retorna zero); `financeiro:saldo:read` é permissão própria (FIN-D-013), nunca reaproveita `lancamento:read`. Validado com `lancamentos` reais que um lançamento cancelado **depois** de `RECEBIDA`/`PAGA` deixa de compor o saldo — consequência direta de `FIN-D-006` que a consulta por status atual garante automaticamente, sem lógica extra. Nenhuma alteração de schema em nenhuma das quatro tarefas.
- **Progresso `02`**: `T1`-`T6` concluídas — tests-first, mutação e gates completos por tarefa. `T1` ajustou também os testes de down-migration de `01` (`financeiro_contas_migration_test.go`) e de `identity` (`identity_migration_test.go`) para desfazer `000007_financeiro_lancamentos` antes de `contas_contabeis`/`users`, já que `lancamentos` referencia as duas por FK. `T2` acrescentou `env_test.go`: `newUser`/`userActor`, porque `lancamentos.criado_por` é uma FK real para `users` (diferente de `contas_contabeis`, sem essa FK) — `actor()` (UUID sintético) continua válido só para os casos de uso de `01`. `T3` (decisão do mantenedor, registrada como `spec/02-lancamentos.md` LAN-02 AC5): a edição reaplica a mesma validação de conta de LAN-01 AC2/AC3 (existe, ativa, mesmo tipo) sempre que `conta_id` é fornecido — a invariante de FIN-D-015 vale para todo o ciclo de vida do lançamento, não só na criação. `T4` reusa `PermLancamentoCreate`/`ActionLancamentoCreate` (FIN-D-003: devolução não é operação distinta), com `devolucao_de_id` no payload `After`; `Tipo` é sempre `DESPESA`, fixado pelo caso de uso, não recebido como input. `T5` ordena por `criado_em`. `T6` (decisão do mantenedor): implementa só o tipo concreto `infra.LancamentoExistenceChecker` (`TemLancamento`), sem criar `financeiro/module.go` — nenhuma tarefa aprovada de `01`/`02` previu esse composition root; o wiring real fica pendente para quando `01-plano-de-contas/T4` (ou a camada HTTP) precisar dele. Testado diretamente em `infra_test` (primeiro teste de `financeiro/infra`, mesmo padrão de `identity/infra`), já que não há caso de uso em `app` que o exercite ainda.
- **Progresso `04`**: Concluída (T1-T3) — tests-first, mutação e gates completos por tarefa. Composição pura sobre `platform/documents.Service`: nenhum caso de uso tem `Authz` próprio — todos repassam `RequiredPermission` e deixam o próprio `documents.Service` recusar internamente (decisão do mantenedor, confirma CMP-01 AC3/CMP-02 AC2/CMP-03 AC1, DOC-01 AC9/DOC-02 AC2 literalmente, sem duplicar a checagem). `T1` confirma que o lançamento existe (reaproveita `LancamentoReader.Buscar`, de `02`) **antes** de chamar `Store`; nunca passa `Supersedes` (`FIN-D-016`). `T2`/`T3` não dependem de `02-lancamentos` em código — um comprovante é identificado pelo próprio id, e `ListarComprovantes` nem checa a existência do lançamento (`LancamentoID` sem documentos só devolve a lista vazia que `ListByOwner` já devolveria). `T1` foi testada só com double de `documents.Service` (`app.DocumentStore`); `T2` combinou doubles (`app.DocumentAccessor`) com um único teste de integração real (Postgres+Garage/S3) provando o roundtrip `ConsultarComprovante→AccessURL→URL assinada→download HTTP real→mesmo conteúdo` — exigência explícita do "Done when" de `tasks/04-comprovantes.md`. `T3` (`app.DocumentLister`) usou só doubles — seu "Done when" não pede roundtrip real, e `ListByOwner` nunca toca em armazenamento — mais dois testes mínimos com Postgres real (sem S3) provam, por comportamento, as garantias negativas de CMP-03 AC3 (status do lançamento e `audit_log` inalterados); a própria ausência dos campos `Lancamentos`/`Auditor` na struct já é a garantia arquitetural de que `ListarComprovantes` não poderia escrever em nenhum dos dois mesmo por engano. Nenhum teste de `financeiro` reverifica comportamento interno de `platform/documents` (assinatura, expiração, `PresignGet`, armazenamento, versionamento, ordenação) — isso continua sendo evidência exclusiva da suíte de `fundacao-documentos`.
- **Progresso `05`**: Concluída (T1-T3). `T1`/`T2` mescladas em `develop` (#53, #54). `T3` (`FIN-D-020`): placeholder removido de `FoundationContributions()`; `financeiro.Contribution()` wired em `cmd/api/main.go` e `cmd/bootstrap-admin/main.go`. A matriz de produção final (confirmada pelo teste `TestContributionCombinedWithFoundationContributionsBuildsTheFullProductionMatrix`, em `financeiro/module_test.go`) dá a `TESOURARIA` as 13 operacionais, a `DIRETORIA` as 4 leituras + `identity:user:read` (este último de `FoundationContributions()`, não de `financeiro`), a `CONSELHO_FISCAL` as 4 leituras + as 3 institucionais + `audit:log:read`, a `PRESIDENTE` tudo automaticamente, e nenhuma permissão de `financeiro` a `ASSOCIADO`/`ESTOQUE_LOJA`/`EVENTOS`/`ADMIN_SISTEMA`. 10 arquivos de teste fora de `financeiro/` precisaram de ajuste como consequência direta (`FIN-D-020` detalha os 4 tipos de ajuste, incluindo a troca de fixture `RoleTesouraria`→`RoleEventos` em 2 testes de `identity/app` e uma lacuna de cobertura real fechada em `cmd/bootstrap-admin`). Nenhuma mudança de comportamento de produção além do previsto pela T3 — confirmado por diff restrito a `roles_matrix.go` + os dois `main.go` fora de arquivos `_test.go`.
- **Progresso `06`**: `Specify`+`Discuss`+`Design`+`Tasks` concluídos, nenhuma implementação iniciada. Auditoria dos padrões de `identity/http`, `httpapi/router.go`, `platform/httpx` e `api/openapi/*.yaml` feita diretamente no código antes de qualquer proposta (não presumida). Achados relevantes: `platform/documents.Service`/`storage.NewS3` existem desde `fundacao-documentos` mas nunca foram instanciados em `cmd/api/main.go` (`FIN-D-023`); `ListarContas`/`ListarLancamentos` não suportam paginação/filtro/ordenação (`FIN-D-021`); sentinels de `financeiro/domain` são frases em português, não códigos — não podem reaproveitar `.Error()` como `identity` faz (`FIN-D-022`); limite de corpo padrão (`1 MiB`) é insuficiente para o upload de comprovante (`10 MiB`, `FIN-D-024`). 8 tarefas (`T1`-`T8`), divididas por unidade arquiteturalmente verificável (composition root, contrato, infraestrutura de erro, 4 grupos de recurso, wiring final) — nunca por arquivo isolado. `validate_spec.py`/`validate_tasks.py`: 0 erros nos três documentos.
- **Bloqueios**: `06-api-http` tem 4 decisões técnicas pendentes de confirmação explícita do mantenedor antes do início da implementação (`FIN-D-021` a `FIN-D-024`, ver relatório de entrega) — nenhuma é uma decisão de negócio, todas têm uma proposta já fechada por precedente, mas nenhuma foi assumida silenciosamente. Fora isso, nenhum bloqueio — todas as decisões de negócio necessárias para `01`-`05` estão fechadas acima.
- **Nota técnica (achado da T2 de `01`, resolvido)**: `platform/audit.Register` exige o formato de duas partes `dominio.verbo` (ex. `user.create`, `document.create`). Os nomes de três partes que apareciam em `design/02-lancamentos.md`, `spec/03-workflow-e-saldo.md`, `design/03-workflow-e-saldo.md`, `tasks/03-workflow-e-saldo.md` e nesta FIN-D-011 foram corrigidos para o formato aceito. Catálogo final de ações de `lancamento`: `lancamento.create` (LAN-01 AC5; reaproveitada por `CriarDevolucao`, LAN-03 — devolução não é operação de negócio distinta, FIN-D-003; `devolucao_de_id` vai no payload `After`), `lancamento.update` (LAN-02 AC4), `lancamento.receive`/`lancamento.pay`/`lancamento.cancel` (`03-workflow-e-saldo`, fora do escopo desta PR). `owner_type="financeiro.lancamento"` usado por `platform/documents` (FIN-D-005, `04-comprovantes`) é um namespace diferente (formato próprio de `documents`, não de `audit.Action`) e não precisou de ajuste.
