# Identity e governança — Especificação

Status: **Rascunho para revisão do mantenedor** (2026-10-13). Sub-spec `02` do TJ Connect, grupo G1 da Onda A. Coordenação em `.specs/features/tj-connect/STATE.md`; contexto em `context.md`. Narrativa em português; critérios de aceite em inglês (EARS), como nas specs de `financeiro` e `estoque`.

## Problem Statement

A autorização atual tem quatro lacunas:

- concede ao `PRESIDENTE` toda permissão do catálogo, automaticamente;
- protege o Conselho Fiscal com uma lista de verbos **proibidos**, que deixa passar um verbo novo;
- não tem como representar processos automatizados, porque toda escrita exige uma pessoa com sessão e a auditoria grava `system` sem identificador;
- não restringe por instância (por exemplo, por jogo).

O TJ Connect traz operações de governança (reembolso, check-in, procedimento de suspensão, quota, liberação comercial) e processos sem pessoa na hora (conciliação, ativação, expiração). Esta spec define, **só no nível de plataforma e identidade**, as regras que permitem modelar essas operações sem bypass, sem concessão implícita e sem perda do que já funciona hoje.

## Goals

- [ ] Toda permissão do catálogo tem uma classe obrigatória, e a ausência de classe impede a partida da API e do worker.
- [ ] Nenhuma permissão sensível chega a um papel por regra genérica; o `PRESIDENTE` só recebe sensíveis por concessão explícita.
- [ ] A transição para a classificação mantém **idênticas** as permissões efetivas de todos os papéis e usuários existentes.
- [ ] Processos automatizados agem com contas de serviço identificáveis, não interativas e de privilégio mínimo, sem poder decidir.
- [ ] Toda entrada nova de auditoria identifica um autor (pessoa ou conta de serviço) e, quando automatizada, a sua causa e o fluxo.
- [ ] O papel `FISCAL` existe, e o contrato de escopo por jogo está definido para a spec `07`.
- [ ] O Conselho Fiscal é protegido por uma lista de ações **permitidas** (fail-closed).

## Out of Scope

| Item | Razão |
|---|---|
| Matriz de concessão das permissões sensíveis do TJ Connect (quem recebe reembolso, suspensão, quota etc.) | Depende do regulamento (P0-2, P0-3, P1-14); será aprovada pelo mantenedor. Esta spec garante que comecem **sem concessão** |
| Regras de negócio de reembolso, cancelamento, quota, check-in e procedimento | Specs `04`, `05`, `06`, `07`, `08` e `14`. Aqui só o contrato de decisão e de separação de funções |
| Tabela e casos de uso de designação de fiscais | Pertencem a `acesso` (spec `07`). Aqui só os requisitos de identidade que a designação usa |
| Worker, fila, outbox e o momento em que o worker usa a conta de serviço | Spec `01` |
| Cadastro público, confirmação de e-mail, limite por IP | Spec `03` |
| Renomear o papel `ASSOCIADO` | Sem efeito na autorização (que não lê nomes); fica como evolução |
| Dupla aprovação de promoção administrativa | Já é decisão futura do ADR-009 |
| Interface web de administração de contas de serviço e catálogo | Spec `10` ou posterior |
| Rever concessões herdadas (por exemplo, `PRESIDENTE` com `financeiro:prestacao_contas:approve`) | Esta spec preserva o estado atual; reduzir concessão é uma decisão de governança separada (D-02-3) |

---

## Relação com requisitos e ADRs

| Origem | O que esta spec cobre |
|---|---|
| TJ requisitos §11 (matriz preliminar), §16 (menor privilégio, trilha de auditoria) | Classes, concessão explícita, separação de funções, autoria |
| Decisões M-1, M-3, M-5, M-7, A-2 do mantenedor; cuidado 1 | Permissões sensíveis sem concessão automática; separação entre solicitar, decidir e executar; proibição de agir sobre si mesmo |
| [ADR-013](../../../../docs/adr/013-contas-de-servico-e-autoria-de-processos.md) | Contas de serviço, três tipos de autoria, `actor_type = service` |
| [ADR-016](../../../../docs/adr/016-permissoes-sensiveis-e-escopo-por-jogo.md) | Classes, `FISCAL`, escopo por instância, Conselho Fiscal por lista permitida |
| [ADR-009](../../../../docs/adr/009-modelo-de-identidade-e-papeis-iniciais.md), [ADR-005](../../../../docs/adr/005-autenticacao-e-rbac.md) | Regras preservadas: sem escalonamento (`authz.Covers`), ninguém altera o próprio acesso, `AdminMembership` para papel administrativo |
| [ADR-004](../../../../docs/adr/004-auditoria-e-imutabilidade-financeira.md) | Auditoria append-only; registros antigos intocados |

---

## Atores, casos de uso e permissões

### Atores

| Ator | Descrição |
|---|---|
| Mantenedor (via código e PR) | Declara permissões, classes, concessões e contas de serviço nas `Contribution` dos módulos. A sincronização aplica isso na partida |
| Sincronização do RBAC | Processo da partida que grava o catálogo, as concessões e as contas de serviço. É uma das duas exceções que gravam `actor_type = system` |
| Administrador com permissão | Atribui papéis, promove, revoga, desativa ou reativa contas de serviço, sempre sob `authz.Covers` |
| Pessoa autenticada | Age com o próprio `Principal` |
| Conta de serviço | Age por um processo do worker com o `Principal` da sua conta |

### Casos de uso (identidade e plataforma)

| Caso de uso | Permissão | Classe proposta |
|---|---|---|
| Sincronizar catálogo, classes, concessões e contas de serviço | — (partida) | — |
| Consultar contas de serviço | `identity:service_account:read` | normal |
| Desativar ou reativar conta de serviço | `identity:service_account:update` | sensivel |
| Atribuir papéis, promover, revogar (existentes) | `identity:role:assign`, `identity:admin:grant`, `identity:admin:revoke` | sensivel (com concessão explícita de transição) |
| Verificar a separação de funções (função de plataforma usada pelos módulos) | — | — |
| Consultar a elegibilidade de identidade de um fiscal (consulta publicada para `acesso`) | — (chamada entre módulos, autorizada no caso de uso de `acesso`) | — |

### Responsabilidades e restrições dos papéis

| Papel | Responsabilidade | Restrições novas |
|---|---|---|
| `ASSOCIADO` | Papel base de toda conta humana: só os próprios dados | Só recebe permissões `normal` de leitura pública e `propria`; **nunca** `sensivel` ou `servico`. Não indica adesão paga (alerta do §3 dos requisitos) |
| `PRESIDENTE` | Presidência | Recebe automaticamente só `normal` e `propria`; `sensivel` só por concessão explícita no módulo dono |
| `DIRETORIA`, `TESOURARIA`, `ESTOQUE_LOJA`, `EVENTOS` | Como hoje | `sensivel` só por concessão explícita |
| `CONSELHO_FISCAL` | Fiscalização | Só ações `read`, `opine` e `approve` sobre `prestacao_contas`; nunca `servico` |
| `ADMIN_SISTEMA` | Administração técnica | Como hoje; não recebe `sensivel` de módulos de negócio sem decisão registrada (D-02-2) |
| `FISCAL` (novo) | Verificação de credenciais em jogo designado | Exige `AdminMembership`; começa sem permissões (concedidas na spec `07`); nunca `servico` |
| Papéis de serviço (novos) | Um por conta de serviço | Só `servico`; nunca atribuíveis a pessoa; não exigem `AdminMembership` |

---

## Estados, transições e invariantes

### Conta de usuário (tipo)

| Tipo | Login | Sessão | Recuperação de senha | `AdminMembership` | Papéis possíveis |
|---|---|---|---|---|---|
| `HUMAN` (todas as contas atuais) | Sim | Sim | Sim | Sim | `ASSOCIADO` mais administrativos |
| `SERVICE` | **Não** | **Não** | **Não** | **Não** | Exatamente um papel de serviço |

O tipo é definido na criação e **nunca muda**.

### Conta de serviço (atividade)

`ATIVA ⇄ DESATIVADA`. As duas transições exigem `identity:service_account:update`, motivo e auditoria. A sincronização também pode desativar uma conta removida do código, mas **nunca a exclui** e **nunca reativa** uma conta desativada manualmente.

### Registro de decisão (contrato que os módulos implementam)

`APROVADA → EXECUTADA` ou `APROVADA → REVOGADA`. A execução e a revogação são únicas, e os dois estados finais são terminais.

### Invariantes

- **INV-1**: toda permissão do catálogo agregado tem exatamente uma classe.
- **INV-2**: nenhum papel humano tem permissão `servico`; nenhum papel de serviço tem permissão que não seja `servico`.
- **INV-3**: `PRESIDENTE` tem uma permissão `sensivel` se, e somente se, um módulo a concedeu explicitamente.
- **INV-4**: `CONSELHO_FISCAL` só tem permissões com ação `read`, `opine`, ou `approve` em recurso `prestacao_contas`.
- **INV-5**: nenhuma sessão pertence a uma conta `SERVICE`.
- **INV-6**: toda entrada de auditoria nova com `actor_type` `user` ou `service` tem `actor_user_id`; `system` só aparece nas ações de bootstrap e de sincronização do RBAC.
- **INV-7**: a mesma pessoa não pratica dois atos que a regra de separação de funções do caso de uso declara incompatíveis.

---

## Fluxos

### F1 — Partida com o catálogo classificado (principal)

1. A composição agrega as `Contribution` de todos os módulos.
2. `BuildMatrix` valida as classes (INV-1), as concessões (INV-2, INV-3, INV-4) e as contas de serviço.
3. A sincronização grava o catálogo, as classes, as concessões e as contas de serviço, e audita cada diferença.
4. A API (e o worker) só começam a atender depois de F1 concluído sem erro.

**Alternativo F1a**: alguma validação falha → a partida aborta com uma mensagem que nomeia o módulo e a permissão, e nada é gravado.

### F2 — Primeira partida após a transição (compatibilidade)

1. As 30 permissões existentes recebem a classe da tabela "Classificação das permissões existentes".
2. As que passam a `sensivel` e que o `PRESIDENTE` tem hoje recebem uma **concessão explícita de transição** ao `PRESIDENTE`, no módulo dono.
3. A matriz efetiva resultante é comparada com o retrato atual: tem de ser idêntica, papel a papel.
4. As permissões novas de classe `sensivel` do TJ Connect, quando existirem, entram sem nenhuma concessão.

### F3 — Execução automatizada de uma decisão humana (contrato)

1. Uma pessoa com permissão `sensivel` registra a decisão: motivo, valor ou escopo e evidência. O módulo dono cria um registro de decisão `APROVADA` com `decision_id`.
2. O módulo enfileira o trabalho de execução (spec `01`) com `decision_id`, `correlation_id` e `on_behalf_of_user_id`.
3. A conta de serviço executa com a sua permissão `servico`; o caso de uso exige o `decision_id` `APROVADA`, do mesmo escopo e ainda não executado.
4. A decisão passa a `EXECUTADA` na mesma transação do efeito; a auditoria registra o serviço e o `decision_id`.

**Alternativo F3a**: decisão revogada, já executada, inexistente ou de outro escopo → a execução é recusada e auditada como `denied`, sem efeito.

### F4 — Desativação de emergência de uma conta de serviço

1. Um administrador com `identity:service_account:update` desativa a conta, com motivo.
2. A partir daí, todo caso de uso chamado com o `Principal` dessa conta é recusado (o comportamento do worker está na spec `01`).
3. A reativação é o mesmo fluxo, também auditado.

---

## Falhas e concorrência

- **Sincronização concorrente** (duas instâncias de API e worker partindo juntas): a sincronização é idempotente e serializada no banco; duas execuções simultâneas produzem o mesmo estado final e no máximo um registro de auditoria por diferença real.
- **Conta desativada durante uma execução**: a verificação de atividade acontece na resolução do `Principal`, no início de cada caso de uso. Um caso de uso já iniciado termina; o próximo é recusado.
- **Decisão executada em paralelo**: a passagem `APROVADA → EXECUTADA` é condicional. De duas execuções concorrentes, só uma efetiva; a outra é recusada.
- **Papel revogado entre a designação e o uso** (fiscal): a consulta de elegibilidade é feita no momento do uso, não guardada.

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
|---|---|---|---|
| Classificação das 30 permissões existentes | Tabela "Classificação das permissões existentes" abaixo | Tudo o que altera acesso de outra pessoa, cancela lançamento, ajusta estoque ou é ato institucional é `sensivel`; o resto, `normal` | n |
| Preservação na transição | Concessão explícita de transição ao `PRESIDENTE` para cada permissão existente que vira `sensivel` | Exigência do mantenedor: nenhuma perda e nenhum ganho | n |
| Onde os papéis de serviço são declarados | No mesmo lugar único que já declara os papéis (`identity/domain/role.go`, protegido por teste); o módulo dono concede as permissões `servico` | Mantém a regra "só um arquivo soletra papéis" | n |
| E-mail das contas de serviço | Endereço no domínio reservado `.invalid` (RFC 2606), por exemplo `svc.cobrancas-sincronizacao@service.invalid` | A coluna `email` é obrigatória e única; o domínio `.invalid` garante que nenhum e-mail seja entregue | n |
| `CommonRead` | Continua permitido só para ação `read` e passa a exigir classe `normal` | Compatível com a regra atual de `authz.Definition` | n |
| Quem recebe `identity:service_account:update` | **Ninguém** até decisão (D-02-1) | Não presumir concessão sensível | n |
| Separação de funções "se o regulamento exigir" | Regra ligada (mais estrita) por padrão | ADR-016 | y |
| Ordem de chegada da classificação | A `02` define as classes e a transição; as permissões novas do TJ Connect são declaradas por cada spec de módulo | Cada módulo é dono das suas permissões | y |

**Open questions:** none — as decisões ainda abertas estão listadas acima como premissas não confirmadas e na seção "Pendências que impedem o Design".

### Classificação das permissões existentes (proposta)

| Permissão | Classe | Hoje com o `PRESIDENTE` (automático) | Concessão explícita de transição ao `PRESIDENTE` |
|---|---|---|---|
| `identity:user:read` | normal | sim | — |
| `identity:user:create` | sensivel | sim | sim |
| `identity:user:update` | sensivel | sim | sim |
| `identity:role:assign` | sensivel | sim | sim |
| `identity:admin:grant` | sensivel | sim | sim |
| `identity:admin:revoke` | sensivel | sim | sim |
| `identity:user:reset_password` | sensivel | sim | sim |
| `audit:log:read` | normal | sim | — |
| `financeiro:conta:create`, `:update`, `:deactivate`, `:read` | normal | sim | — |
| `financeiro:lancamento:create`, `:update`, `:read`, `:receive`, `:pay` | normal | sim | — |
| `financeiro:lancamento:cancel` | sensivel | sim | sim |
| `financeiro:saldo:read`, `financeiro:comprovante:create`, `financeiro:comprovante:read` | normal | sim | — |
| `financeiro:prestacao_contas:read` | normal | sim | — |
| `financeiro:prestacao_contas:approve` | sensivel | sim | sim |
| `financeiro:parecer:opine` | sensivel | sim | sim |
| `estoque:produto:create`, `:read`; `estoque:movimentacao:create`, `:read`; `estoque:saldo:read` | normal | sim | — |
| `estoque:movimentacao:adjust` | sensivel | sim | sim |

Total: 30 permissões, 10 `sensivel` com concessão explícita de transição ao `PRESIDENTE`. As concessões atuais aos demais papéis não mudam.

---

## User Stories

### P1: Classe obrigatória e fail-closed ⭐ MVP

**User Story**: Como mantenedor, quero que toda permissão tenha uma classe obrigatória e que a plataforma recuse partir com algo não classificado ou concedido fora da regra, para que nenhuma permissão sensível seja concedida por omissão.

**Why P1**: Base de todas as permissões do TJ Connect; sem isso, a matriz sensível não tem garantia.

**Acceptance Criteria**:
1. The system SHALL require every permission definition to declare exactly one class among `normal`, `sensivel`, `propria` and `servico`.
2. IF any permission in the aggregated catalog has no class THEN the API and the worker SHALL refuse to start and SHALL report the module and the permission name.
3. WHEN the matrix is built THEN `PRESIDENTE` SHALL receive automatically only permissions of class `normal` and `propria`.
4. WHEN the matrix is built THEN a permission of class `sensivel` SHALL be held by a role only if the owning module's contribution grants it to that role explicitly.
5. IF a contribution grants a permission of class `servico` to a human role THEN the API and the worker SHALL refuse to start.
6. IF a contribution grants a permission of any class other than `servico` to a service role THEN the API and the worker SHALL refuse to start.
7. IF a contribution grants a permission of class `sensivel` or `servico` to `ASSOCIADO` THEN the API and the worker SHALL refuse to start.
8. IF a permission is marked `CommonRead` and its class is not `normal` THEN the API and the worker SHALL refuse to start.
9. WHEN the synchronization runs THEN the system SHALL store the class of every permission and SHALL write one audit entry for every permission whose class changed.

**Independent Test**: montar matrizes de teste com cada violação (sem classe; `servico` a humano; `sensivel` sem concessão; `sensivel` a `ASSOCIADO`) e ver `BuildMatrix` falhar com a mensagem certa; montar uma matriz válida e ver o `PRESIDENTE` sem as sensíveis não concedidas.

---

### P1: Transição sem perda nem escalonamento ⭐ MVP

**User Story**: Como usuário atual da plataforma, quero continuar com exatamente as mesmas permissões depois da classificação, para que nada que funciona hoje deixe de funcionar e ninguém ganhe acesso novo.

**Why P1**: Exigência do mantenedor; o sistema atual está em uso.

**Acceptance Criteria**:
1. WHEN the classified matrix is built for the first time THEN the effective permissions of each of the eight existing roles SHALL be identical to the effective permissions before the classification.
2. WHEN an existing permission becomes `sensivel` and `PRESIDENTE` held it automatically before THEN its owning module SHALL grant it to `PRESIDENTE` explicitly.
3. The system SHALL keep a versioned snapshot of the pre-classification matrix and SHALL fail its test suite if any existing role gains or loses a permission without an approved change to that snapshot.
4. WHEN a new permission of class `sensivel` is declared by a TJ Connect module THEN it SHALL start granted to no role.
5. WHEN the synchronization applies the classified matrix THEN existing sessions SHALL keep working without a new login, because effective permissions are resolved per request.
6. WHEN the `users` table receives the account type THEN every existing account SHALL be of type `HUMAN`.
7. The system SHALL keep every existing audit entry unchanged, including entries whose `actor_type` is `system`.

**Independent Test**: comparar o retrato versionado da matriz atual com a matriz classificada, papel a papel; criar uma permissão `sensivel` nova e ver nenhum papel com ela.

---

### P1: Contas de serviço não interativas ⭐ MVP

**User Story**: Como mantenedor, quero que cada processo automatizado tenha uma conta própria, identificável, sem login e com o mínimo de permissões, para que toda ação automática tenha um autor e possa ser interrompida.

**Why P1**: Pré-requisito do worker (spec `01`) e de todo efeito automático do TJ Connect.

**Acceptance Criteria**:
1. The system SHALL support accounts of type `SERVICE`, each with a unique stable name in the form `svc.<modulo>-<processo>`, a description and an owning module.
2. IF the module prefix of a service account name differs from its owning module THEN the API and the worker SHALL refuse to start.
3. WHEN the synchronization runs THEN it SHALL create every service account declared in code that does not exist yet, and SHALL write one audit entry per created account.
4. WHEN a service account is no longer declared in code THEN the synchronization SHALL deactivate it, SHALL NOT delete it, and SHALL write an audit entry.
5. The synchronization SHALL NOT reactivate a service account that was deactivated by a person.
6. IF a login is attempted with the e-mail of a service account THEN the system SHALL respond exactly as for invalid credentials and SHALL NOT create a session.
7. IF a password recovery is requested for the e-mail of a service account THEN the system SHALL respond exactly as for any address and SHALL NOT create a recovery token.
8. The system SHALL reject at the database level any session row whose user is of type `SERVICE`.
9. IF a person tries to assign any role to a service account, promote it, revoke it or reset its password THEN the system SHALL refuse with `forbidden` and SHALL write a `denied` audit entry.
10. The system SHALL hold each service account with exactly one service role, and each service role SHALL belong to exactly one service account.
11. WHEN a person with `identity:service_account:update` deactivates or reactivates a service account with a reason THEN the system SHALL apply the change and SHALL write an audit entry with the reason.
12. IF a deactivation or reactivation has a reason shorter than 10 characters THEN the system SHALL refuse it with `validation_failed`.
13. WHILE a service account is inactive the system SHALL refuse every use case invoked with its principal, with `forbidden`.
14. WHEN users are listed by the existing user administration THEN accounts of type `SERVICE` SHALL NOT appear.
15. WHEN a person with `identity:service_account:read` lists service accounts THEN the system SHALL return name, description, owning module, role, permissions and activity, and never any credential field.

**Independent Test**: declarar uma conta de serviço de teste, partir, ver a conta criada; tentar login, recuperação e sessão e ver as recusas; desativar e ver um caso de uso recusado com o `Principal` dela.

---

### P1: Solicitação, decisão e execução separadas ⭐ MVP

**User Story**: Como presidente ou diretor, quero que operações sensíveis separem quem solicita, quem decide e quem executa, e que processos automáticos só executem decisões humanas registradas, para que nenhum ato de governança dependa de uma pessoa só ou de um processo sem juízo.

**Why P1**: Contrato usado por reembolso (`04`/`06`), liberação de vaga (`05`), correção de check-in (`07`) e procedimento (`14`).

**Acceptance Criteria**:
1. The system SHALL provide a platform rule that, given the actor and the authors of previous acts declared incompatible by the use case, refuses the act when the actor is one of those authors.
2. WHEN the separation rule refuses an act THEN the use case SHALL respond `forbidden` with code `separacao_de_funcoes` and SHALL write a `denied` audit entry naming the rule.
3. IF an actor attempts a sensitive act whose subject is the actor's own person THEN the use case SHALL refuse it with code `ato_sobre_si_mesmo` and SHALL write a `denied` audit entry.
4. The system SHALL require every decision record to carry an id, the deciding person, a reason of at least 10 characters, the scope it authorizes, the time and its state.
5. WHEN a service account executes a decision THEN the use case SHALL require a `decision_id` in state `APROVADA`, of the same scope and not yet executed.
6. IF the `decision_id` is missing, unknown, `REVOGADA`, `EXECUTADA` or of a different scope THEN the use case SHALL refuse the execution, SHALL produce no effect and SHALL write a `denied` audit entry.
7. WHEN two executions of the same decision run concurrently THEN exactly one SHALL succeed and the other SHALL be refused.
8. The system SHALL NOT allow a service account to create, approve or revoke a decision record.

**Independent Test**: um módulo de teste com a regra "solicitante ≠ aprovador" e uma execução por serviço: ver a recusa do mesmo autor, a recusa sem decisão, uma única execução sob concorrência.

---

### P1: Autoria e rastreabilidade ⭐ MVP

**User Story**: Como membro do Conselho Fiscal ou auditor, quero que toda ação registrada diga quem agiu (pessoa ou processo), por quê e em nome de quem, para reconstruir qualquer efeito automático até a origem.

**Why P1**: Exigência de autoria identificável (D2) e base da auditoria de pagamentos e check-ins.

**Acceptance Criteria**:
1. The system SHALL accept `actor_type` values `user`, `service`, `anonymous` and `system`.
2. IF an audit entry has `actor_type` `user` or `service` without an actor id THEN the system SHALL reject the entry and SHALL fail the enclosing transaction.
3. IF an audit entry has `actor_type` `service` and its actor id is not an account of type `SERVICE` THEN the system SHALL reject the entry.
4. IF an audit entry has `actor_type` `user` and its actor id is an account of type `SERVICE` THEN the system SHALL reject the entry.
5. The system SHALL allow `actor_type` `system` only for the bootstrap and RBAC synchronization actions, and a test SHALL fail if any other action uses it.
6. WHEN a service account writes an audit entry THEN the entry context SHALL contain `correlation_id` and `causation_id`.
7. WHEN an action executes a human decision THEN the entry context SHALL contain `decision_id`.
8. WHERE an automated flow was started by a person THEN the entry context SHALL contain `on_behalf_of_user_id` with that person's id.
9. WHEN a person with `audit:log:read` filters the audit log by `correlation_id` THEN the system SHALL return every entry of that flow ordered by time.

**Independent Test**: gravar entradas válidas e inválidas para cada combinação de tipo de ator e conta; percorrer um fluxo de teste pelo `correlation_id`.

---

### P1: Papel FISCAL e contrato de escopo por jogo ⭐ MVP

**User Story**: Como presidente, quero designar fiscais para jogos específicos e ter certeza de que só eles, e nunca sobre si mesmos, registram entradas nesses jogos, para que o check-in seja confiável.

**Why P1**: A spec `07` depende deste contrato.

**Acceptance Criteria**:
1. The system SHALL provide the role `FISCAL`, which requires an active `AdminMembership` like every human role other than `ASSOCIADO`.
2. WHEN `FISCAL` is created THEN it SHALL hold no permission until a module contribution grants one.
3. The system SHALL publish an identity query that reports whether a user is an active `HUMAN` account with an active `AdminMembership` and the role `FISCAL`.
4. WHEN the role `FISCAL` or the membership of a user is revoked THEN the identity query SHALL report that user as not eligible from that moment on.
5. The system SHALL NOT grant any permission to a user because of a game designation; a designation SHALL only narrow where an existing permission can be used.
6. The system SHALL require that instance-scoped permissions (such as check-in) are checked in the use case together with the scope record, never by the route alone.

**Independent Test**: promover uma pessoa a `FISCAL`, consultar a elegibilidade, revogar o papel e ver a consulta mudar.

---

### P1: Conselho Fiscal fail-closed ⭐ MVP

**User Story**: Como Conselho Fiscal, quero que a plataforma garanta que eu só fiscalizo e nunca opero, mesmo quando módulos novos surgirem, para preservar a minha independência.

**Why P1**: Exigência do mantenedor; a lista atual de verbos proibidos não cobre verbos novos.

**Acceptance Criteria**:
1. WHEN the matrix is built THEN every permission granted to `CONSELHO_FISCAL` SHALL have action `read` or `opine`, or action `approve` on resource `prestacao_contas`.
2. IF a contribution grants `CONSELHO_FISCAL` a permission outside that list THEN the API and the worker SHALL refuse to start.
3. IF a contribution grants `CONSELHO_FISCAL` a permission of class `servico` THEN the API and the worker SHALL refuse to start.
4. WHEN the allow-list replaces the current deny-list THEN `CONSELHO_FISCAL` SHALL keep exactly its current effective permissions.

**Independent Test**: tentar conceder `acesso:checkin:create`, `cobrancas:reembolso:approve` e um verbo inventado ao Conselho Fiscal e ver a partida falhar; ver as permissões atuais intactas.

---

## Edge Cases

- IF two instances synchronize at the same time THEN the system SHALL reach the same final state and SHALL write one audit entry per real difference.
- IF a code change moves an existing permission from `normal` to `sensivel` without an explicit grant to the roles that held it THEN the snapshot test SHALL fail.
- IF a service account is deactivated while one of its use cases is in progress THEN that use case SHALL finish and the next one SHALL be refused.
- WHEN a person is both `FISCAL` and a member being checked in THEN the self-action rule SHALL refuse the act (the rule itself is applied by spec `07`).
- IF the reason of a decision record has fewer than 10 characters THEN the system SHALL refuse the record with `validation_failed`.

---

## Cenários de teste

| # | Cenário | Tipo | Requisitos |
|---|---|---|---|
| T-01 | Catálogo com uma permissão sem classe → partida falha nomeando módulo e permissão | Unidade (`BuildMatrix`) | CLS-01 |
| T-02 | `servico` concedida a `TESOURARIA` → partida falha | Unidade | CLS-01 |
| T-03 | `sensivel` nova sem concessão → `PRESIDENTE` não a tem | Unidade | CLS-01, TRN-01 |
| T-04 | Retrato da matriz atual = matriz classificada, para os 8 papéis | Unidade (snapshot) | TRN-01 |
| T-05 | Sessão existente continua válida depois da sincronização | Integração | TRN-01 |
| T-06 | Login e recuperação com e-mail de conta de serviço → mesma resposta genérica, sem sessão nem token | Integração | SVC-01 |
| T-07 | `INSERT` direto de sessão para conta `SERVICE` → recusado pelo banco | Integração | SVC-01 |
| T-08 | Conta removida do código → desativada, não excluída; reativação manual não é desfeita | Integração | SVC-01 |
| T-09 | Caso de uso com o `Principal` de uma conta desativada → `forbidden` | Integração | SVC-01 |
| T-10 | Solicitante tenta aprovar o próprio pedido → `separacao_de_funcoes` | Unidade | SOD-01 |
| T-11 | Execução sem decisão, com decisão revogada ou de outro escopo → recusada, sem efeito | Integração | SOD-01 |
| T-12 | Duas execuções concorrentes da mesma decisão → uma efetiva | Integração (concorrência) | SOD-01 |
| T-13 | Entrada de auditoria `service` com id humano e vice-versa → rejeitada | Unidade | AUT-01 |
| T-14 | Uso de `system` fora do bootstrap e da sincronização → teste falha | Teste de código-fonte | AUT-01 |
| T-15 | Filtro por `correlation_id` devolve o fluxo ordenado | Integração | AUT-01 |
| T-16 | Revogação de `FISCAL` → consulta de elegibilidade falsa | Integração | FIS-01 |
| T-17 | Conselho Fiscal com verbo novo → partida falha; permissões atuais intactas | Unidade | CFS-01 |

---

## Dependências e riscos

| Tipo | Item | Mitigação |
|---|---|---|
| Dependência | A spec `01` usa as contas de serviço e a autoria definidas aqui | `02` antes de `01` na sequência do G1 |
| Dependência | As specs `04`–`08` declaram permissões com classe e aplicam a separação de funções | Contratos fixados aqui |
| Risco | A classificação errada de uma permissão existente retira acesso legítimo | Retrato versionado da matriz (TRN-01); transição só com concessão explícita |
| Risco | Uma conta de serviço com permissão demais | Papel único por conta, só `servico`, partida falha fora disso |
| Risco | Migrations em tabelas de `identity` e `audit_log` em produção | Colunas com padrão compatível (`HUMAN`); restrições novas só sobre linhas novas; planejamento no Design, sem execução agora |
| Risco | Sem matriz sensível aprovada, os fluxos do TJ Connect não operam | Comportamento desejado; aprovação da matriz é pré-requisito de operação, não de Design |

---

## Pendências que impedem o Design

**Nenhuma.** As decisões abertas abaixo não mudam a estrutura do Design e podem ser fechadas até as Tasks:

- **D-02-1**: quem recebe `identity:service_account:update` (chave de emergência). Recomendação: `ADMIN_SISTEMA` e `PRESIDENTE`, por concessão explícita.
- **D-02-2**: confirmar a classificação proposta das 30 permissões existentes.
- **D-02-3** (governança, fora desta spec): rever concessões herdadas pelo `PRESIDENTE`, por exemplo `financeiro:prestacao_contas:approve` e `financeiro:parecer:opine`, que são atos do Conselho Fiscal. Esta spec as **preserva**.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
|---|---|---|---|
| CLS-01 | P1: Classe obrigatória e fail-closed | Specify | Pending |
| TRN-01 | P1: Transição sem perda nem escalonamento | Specify | Pending |
| SVC-01 | P1: Contas de serviço não interativas | Specify | Pending |
| SOD-01 | P1: Solicitação, decisão e execução separadas | Specify | Pending |
| AUT-01 | P1: Autoria e rastreabilidade | Specify | Pending |
| FIS-01 | P1: Papel FISCAL e contrato de escopo por jogo | Specify | Pending |
| CFS-01 | P1: Conselho Fiscal fail-closed | Specify | Pending |

**Coverage:** 7 total, 0 mapped to tasks, 7 unmapped (Design não iniciado).

## Success Criteria

- [ ] A matriz efetiva dos 8 papéis atuais é idêntica antes e depois (teste de retrato verde).
- [ ] Nenhuma permissão sensível nova está concedida a nenhum papel ao fim desta spec.
- [ ] Toda entrada de auditoria nova tem autor identificável; `system` só no bootstrap e na sincronização.
- [ ] Uma conta de serviço não consegue sessão por nenhum caminho (aplicação e banco).
