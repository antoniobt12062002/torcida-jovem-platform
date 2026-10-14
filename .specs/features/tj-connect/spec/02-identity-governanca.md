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
- [ ] A transição para a classificação mantém **idênticas** as permissões efetivas de todos os papéis e usuários existentes, com uma única exceção expressamente aprovada (D-02-3), que só entra em vigor com a regra de nomeação institucional do Conselho Fiscal (INS-01).
- [ ] Processos automatizados agem com contas de serviço identificáveis, não interativas e de privilégio mínimo. Sem juízo discricionário, executam **regras determinísticas aprovadas** (por exemplo, confirmar pagamento) ou **decisões humanas registradas** (por exemplo, reembolso aprovado), e a diferença entre os dois casos é explícita.
- [ ] Toda entrada nova de auditoria identifica um autor (pessoa ou conta de serviço) e, quando automatizada, a sua causa e o fluxo.
- [ ] O papel `FISCAL` existe, e o contrato de escopo por jogo está definido para a spec `07`.
- [ ] O Conselho Fiscal só recebe permissões de uma **lista explícita e versionada de permissões autorizadas** (fail-closed), com a restrição de verbos como segunda barreira.

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
| Aplicar concessões reais (inclusive a D-02-3, já aprovada) | Esta spec **especifica** a mudança; a alteração das concessões no código acontece só na implementação autorizada |

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
| Desativar conta de serviço (chave de emergência) | `identity:service_account:deactivate` | sensivel |
| Reativar conta de serviço | `identity:service_account:reactivate` | sensivel (mais restrita: quem reativa ≠ quem desativou; contingência com espera e revisão posterior) |
| Registrar a nomeação ou a destituição de membro do Conselho Fiscal (ato institucional registrado; não inclui bloqueio técnico nem redefinição de senha) | `identity:conselho_fiscal:manage` (D-02-4) | sensivel |
| Atribuir papéis, promover, revogar (existentes) | `identity:role:assign`, `identity:admin:grant`, `identity:admin:revoke` | sensivel (com concessão explícita de transição) |
| Verificar a separação de funções (função de plataforma usada pelos módulos) | — | — |
| Consultar a elegibilidade de identidade de um fiscal (consulta publicada para `acesso`) | — (chamada entre módulos, autorizada no caso de uso de `acesso`) | — |

### Responsabilidades e restrições dos papéis

| Papel | Responsabilidade | Restrições novas |
|---|---|---|
| `ASSOCIADO` | Papel base de toda conta humana: só os próprios dados | Só recebe permissões `normal` de leitura pública e `propria`; **nunca** `sensivel` ou `servico`. Não indica adesão paga (alerta do §3 dos requisitos) |
| `PRESIDENTE` | Presidência | Recebe automaticamente só `normal` e `propria`; `sensivel` só por concessão explícita no módulo dono |
| `DIRETORIA`, `TESOURARIA`, `ESTOQUE_LOJA`, `EVENTOS` | Como hoje | `sensivel` só por concessão explícita |
| `CONSELHO_FISCAL` | Fiscalização | Só permissões da lista explícita autorizada (hoje 11, ver CFS-01); cada uma também precisa ter ação `read`, `opine` ou `approve` sobre `prestacao_contas`; nunca `servico` |
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

`ATIVA → DESATIVADA` exige `identity:service_account:deactivate`; `DESATIVADA → ATIVA` exige `identity:service_account:reactivate` e uma pessoa **diferente** de quem desativou. As duas transições exigem motivo e auditoria. A sincronização também pode desativar uma conta removida do código, mas **nunca a exclui** e **nunca reativa** uma conta desativada.

### Registro de decisão (contrato que os módulos implementam)

`APROVADA → EXECUTADA` ou `APROVADA → REVOGADA`. A execução e a revogação são únicas, e os dois estados finais são terminais.

### Dois tipos de execução automatizada

| Tipo | Quando | O que autoriza | O que a auditoria registra | Exemplos |
|---|---|---|---|---|
| **Execução de regra aprovada** (determinística) | O efeito decorre de um fato verificável e de uma regra fixada em spec, código e ADR, sem juízo | A permissão `servico` do processo, sem decisão humana caso a caso | `regra` (nome e versão da regra), `causation_id` (fato de origem: notificação, evento, janela de agendamento), `correlation_id` | Confirmar pagamento consultado no provedor; registrar a receita da cobrança paga; ativar a adesão paga com vaga; expirar a reserva depois da reconciliação; conciliar |
| **Execução de decisão administrativa** | O efeito depende de juízo discricionário, que só uma pessoa exerce | A permissão `servico` **marcada como "exige decisão"**, mais um `decision_id` `APROVADA`, do mesmo escopo e não executado | Além do acima, `decision_id` e `on_behalf_of_user_id` (quem decidiu) | Executar no provedor o reembolso aprovado; aplicar a suspensão decidida |

Regras:

- Uma confirmação de pagamento **não** exige nova decisão humana: é regra aprovada.
- Aprovar reembolso, decidir suspensão, liberar vaga, corrigir presença e conceder liberação comercial são **sempre** atos humanos `sensivel`. Nenhum papel de serviço os recebe, e a execução automática deles só existe como "execução de decisão administrativa".
- Toda permissão `servico` declara, no catálogo, se **exige decisão**. A classe e esse atributo são verificados na partida.

### Invariantes

- **INV-1**: toda permissão do catálogo agregado tem exatamente uma classe.
- **INV-2**: nenhum papel humano tem permissão `servico`; nenhum papel de serviço tem permissão que não seja `servico`.
- **INV-3**: `PRESIDENTE` tem uma permissão `sensivel` se, e somente se, um módulo a concedeu explicitamente.
- **INV-4**: `CONSELHO_FISCAL` só tem permissões que constam da lista explícita e versionada de permissões autorizadas, e cada uma tem ação `read`, `opine`, ou `approve` em recurso `prestacao_contas`.
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
2. As que passam a `sensivel` e que um papel tem hoje recebem uma **concessão explícita e justificada** a esse papel, no módulo dono. O `PRESIDENTE` recebe 16; as concessões de TESOURARIA, ESTOQUE_LOJA, ADMIN_SISTEMA e CONSELHO_FISCAL já são explícitas.
3. A matriz efetiva resultante é comparada com o retrato atual: tem de ser idêntica, papel a papel, exceto pela **exceção aprovada D-02-3** (retirada de duas permissões do `PRESIDENTE`), aplicada só com a regra de nomeação institucional (INS-01).
4. As permissões novas de classe `sensivel` do TJ Connect, quando existirem, entram sem nenhuma concessão.

### F3a — Execução automatizada de regra aprovada (contrato)

1. Um fato verificável chega (por exemplo, a consulta ao provedor confirma o pagamento).
2. A conta de serviço executa com a sua permissão `servico` que **não** exige decisão; o caso de uso aplica a regra fixada na spec do módulo.
3. A auditoria registra a conta de serviço, a `regra` (nome e versão), o `causation_id` e o `correlation_id`. Não há decisão humana caso a caso.

### F3 — Execução automatizada de uma decisão humana (contrato)

1. Uma pessoa com permissão `sensivel` registra a decisão: motivo, valor ou escopo e evidência. O módulo dono cria um registro de decisão `APROVADA` com `decision_id`.
2. O módulo enfileira o trabalho de execução (spec `01`) com `decision_id`, `correlation_id` e `on_behalf_of_user_id`.
3. A conta de serviço executa com a sua permissão `servico`; o caso de uso exige o `decision_id` `APROVADA`, do mesmo escopo e ainda não executado.
4. A decisão passa a `EXECUTADA` na mesma transação do efeito; a auditoria registra o serviço e o `decision_id`.

**Alternativo F3a**: decisão revogada, já executada, inexistente ou de outro escopo → a execução é recusada e auditada como `denied`, sem efeito.

### F4 — Desativação de emergência de uma conta de serviço

1. Um administrador com `identity:service_account:deactivate` desativa a conta, com motivo.
2. A partir daí, todo caso de uso chamado com o `Principal` dessa conta é recusado (o comportamento do worker está na spec `01`).
3. A reativação exige `identity:service_account:reactivate`, motivo e uma pessoa diferente de quem desativou. É auditada.

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
| Classificação das 30 permissões existentes | Tabela "Classificação das permissões existentes" abaixo: `sensivel` para tudo o que altera acesso de terceiros, movimenta recursos, altera saldos ou estrutura contábil, ou é ato institucional; `normal` para leituras e atos sem efeito patrimonial | **D-02-2 aprovada** em 2026-10-14 (18 sensíveis, 12 normais) | y |
| Preservação na transição | Concessão explícita, justificada e revisada em PR, a cada papel que já tem hoje uma permissão que passa a `sensivel`; a única exceção aprovada é a D-02-3 | Diretriz do mantenedor: preservar acessos legítimos por concessões explícitas quando justificadas | n |
| Onde os papéis de serviço são declarados | No mesmo lugar único que já declara os papéis (`identity/domain/role.go`, protegido por teste); o módulo dono concede as permissões `servico` | Mantém a regra "só um arquivo soletra papéis" | n |
| E-mail das contas de serviço | Endereço no domínio reservado `.invalid` (RFC 2606), por exemplo `svc.cobrancas-sincronizacao@service.invalid` | A coluna `email` é obrigatória e única; o domínio `.invalid` garante que nenhum e-mail seja entregue | n |
| `CommonRead` | Continua permitido só para ação `read` e passa a exigir classe `normal` | Compatível com a regra atual de `authz.Definition` | n |
| Quem recebe `identity:service_account:deactivate` e `:reactivate` | `ADMIN_SISTEMA` e `PRESIDENTE`, por concessão explícita, com reativação por pessoa diferente e procedimento de contingência (SVC-01 AC18–21) | **D-02-1 aprovada** em 2026-10-14; concessão ainda não aplicada no código | y |
| Prazo de espera e revisão da reativação de contingência | 30 minutos desde a desativação; revisão posterior por outra pessoa com `identity:service_account:reactivate` ou pelo Conselho Fiscal (leitura da auditoria) | Evita bloqueio operacional sem anular a separação de funções; prazo configurável | n |
| Lista explícita do Conselho Fiscal | As 11 permissões que o Conselho Fiscal tem hoje (tabela em CFS-01) | Preserva as autorizações legítimas atuais; todo acréscimo exige mudança revisada da lista | n |
| Nomeação institucional do Conselho Fiscal (INS-01) | Regra da **D-02-4**, sem a qual a D-02-3 não pode ser aplicada; a lista de papéis incompatíveis além de `PRESIDENTE` fica sujeita ao estatuto | **D-02-4 aprovada como direção arquitetural** e **D-02-5 aprovada** (`identity:conselho_fiscal:manage` ao `PRESIDENTE`) em 2026-10-14 | y |
| Separação de funções "se o regulamento exigir" | Regra ligada (mais estrita) por padrão | ADR-016 | y |
| Ordem de chegada da classificação | A `02` define as classes e a transição; as permissões novas do TJ Connect são declaradas por cada spec de módulo | Cada módulo é dono das suas permissões | y |

**Open questions:** none — as decisões ainda abertas estão listadas acima como premissas não confirmadas e na seção "Pendências que impedem o Design".

### Classificação das permissões existentes (D-02-2, aprovada em 2026-10-14)

Critério do mantenedor: é `sensivel` toda operação que movimenta recursos, altera saldos, altera a estrutura contábil ou patrimonial, altera direitos ou acesso de terceiros, ou é ato institucional. A revisão anterior (10 sensíveis) foi ampliada para **18**.

| Permissão | Classe | Motivo | Quem tem hoje | Concessão explícita após a classificação |
|---|---|---|---|---|
| `identity:user:read` | normal | Leitura | ADMIN_SISTEMA, DIRETORIA, PRESIDENTE | PRESIDENTE automático; os outros já explícitos |
| `identity:user:create`, `:update`, `:reset_password`; `identity:role:assign`; `identity:admin:grant`, `:revoke` | sensivel | Alteram acesso de terceiros | ADMIN_SISTEMA (explícito), PRESIDENTE (automático) | ADMIN_SISTEMA: mantém. **PRESIDENTE: explícita, justificada** |
| `audit:log:read` | normal | Leitura | ADMIN_SISTEMA, CONSELHO_FISCAL, PRESIDENTE | Sem mudança |
| `financeiro:conta:create`, `:update`, `:deactivate` | **sensivel** (antes normal) | Alteram a estrutura contábil em que os saldos são classificados | TESOURARIA (explícito), PRESIDENTE (automático) | TESOURARIA: mantém. **PRESIDENTE: explícita, justificada** |
| `financeiro:lancamento:create`, `:update`, `:receive`, `:pay` | **sensivel** (antes normal) | Movimentam recursos e alteram saldos | TESOURARIA (explícito), PRESIDENTE (automático) | TESOURARIA: mantém. **PRESIDENTE: explícita, justificada** |
| `financeiro:lancamento:cancel` | sensivel | Altera saldo já liquidado | TESOURARIA, PRESIDENTE | TESOURARIA: mantém. **PRESIDENTE: explícita, justificada** |
| `financeiro:conta:read`, `:lancamento:read`, `:saldo:read`, `:comprovante:read`, `:prestacao_contas:read` | normal | Leitura | TESOURARIA, DIRETORIA, CONSELHO_FISCAL (exceto `prestacao_contas:read` em DIRETORIA e TESOURARIA), PRESIDENTE | Sem mudança |
| `financeiro:comprovante:create` | normal | Anexa evidência; não altera saldo | TESOURARIA, PRESIDENTE | Sem mudança |
| `financeiro:prestacao_contas:approve`, `financeiro:parecer:opine` | sensivel | Atos institucionais do Conselho Fiscal | CONSELHO_FISCAL (explícito), PRESIDENTE (automático) | CONSELHO_FISCAL: mantém. **PRESIDENTE: retirada (D-02-3, exceção aprovada)** |
| `estoque:movimentacao:create`, `:adjust` | **sensivel** (`create` antes normal) | Alteram saldo físico e patrimônio | ESTOQUE_LOJA (explícito), PRESIDENTE (automático) | ESTOQUE_LOJA: mantém. **PRESIDENTE: explícita, justificada** |
| `estoque:produto:create` | normal | Cria SKU; não altera saldo | ESTOQUE_LOJA, PRESIDENTE | Sem mudança |
| `estoque:produto:read`, `:movimentacao:read`, `:saldo:read` | normal | Leitura | ESTOQUE_LOJA, DIRETORIA, CONSELHO_FISCAL, PRESIDENTE | Sem mudança |

Total: 30 permissões. **18 `sensivel`**: 6 de `identity`, 3 de conta, 5 de lançamento, 2 institucionais e 2 de estoque. **12 `normal`**. O `PRESIDENTE` precisa de **16 concessões explícitas** para manter o que tem hoje: as 18 sensíveis menos as 2 retiradas pela D-02-3. Cada concessão deve ter justificativa registrada no PR que a implementar. TESOURARIA, ESTOQUE_LOJA, ADMIN_SISTEMA e CONSELHO_FISCAL **já** recebem as suas por concessão explícita no código atual, então nada muda para eles.

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
1. WHEN the classified matrix is built for the first time THEN the effective permissions of each of the eight existing roles SHALL be identical to the effective permissions before the classification, except for the changes recorded as approved exceptions.
2. WHEN an existing permission becomes `sensivel` and a role held it before, automatically or explicitly, THEN its owning module SHALL grant it to that role explicitly with a recorded justification, unless an approved exception removes it.
3. The system SHALL keep a versioned snapshot of the pre-classification matrix and SHALL fail its test suite if any existing role gains or loses a permission without an approved change to that snapshot.
4. WHEN a new permission of class `sensivel` is declared by a TJ Connect module THEN it SHALL start granted to no role.
5. WHEN the synchronization applies the classified matrix THEN existing sessions SHALL keep working without a new login, because effective permissions are resolved per request.
6. WHEN the `users` table receives the account type THEN every existing account SHALL be of type `HUMAN`.
7. The system SHALL keep every existing audit entry unchanged, including entries whose `actor_type` is `system`.
8. The system SHALL record the removal of `financeiro:prestacao_contas:approve` and `financeiro:parecer:opine` from `PRESIDENTE` (decision D-02-3, approved on 2026-10-14) as the only approved exception to the preservation of the previous matrix, and SHALL apply it only together with the institutional appointment rule of INS-01.
9. The system SHALL hold a matrix version in code and SHALL fail its test suite when the aggregated matrix changes without a version increase.
10. WHEN a process synchronizes the RBAC THEN the system SHALL apply the matrix only if the process's matrix version is greater than or equal to the stored version, SHALL refuse to start on a lower version or on an equal version with a different matrix, and SHALL reject at the database level any write to permissions, roles or role permissions made by a process that does not declare a version at least equal to the stored one.
11. The worker SHALL NOT synchronize the RBAC; it SHALL verify the stored matrix version and SHALL refuse to start when it differs from its own after the configured wait.
12. WHEN an operator needs to run an older binary THEN the system SHALL require a rollback authorized beforehand by an identified person through an authenticated use case with a specific permission, bound to the target version and to the complete sets of grants the rollback restores and removes, and SHALL record the authorizing person as the author of the decision (`actor_type` `user`) and the technical execution separately, referencing that person, before lowering the stored version; the database SHALL reject any decrease of the stored version outside an authorized execution.
13. IF a rollback plan restores any institutional permission THEN the system SHALL refuse its authorization unless the maintainer's express decision and its document are recorded with it, and SHALL refuse authorization by a person who would receive any of the restored grants.

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
11. WHEN a person with `identity:service_account:deactivate` deactivates a service account with a reason THEN the system SHALL apply the change and SHALL write an audit entry with the reason.
12. IF a deactivation or reactivation has a reason shorter than 10 characters THEN the system SHALL refuse it with `validation_failed`.
13. WHILE a service account is inactive the system SHALL refuse every use case invoked with its principal, with `forbidden`.
14. WHEN users are listed by the existing user administration THEN accounts of type `SERVICE` SHALL NOT appear.
15. WHEN a person with `identity:service_account:read` lists service accounts THEN the system SHALL return name, description, owning module, role, permissions and activity, and never any credential field.
16. WHEN a person with `identity:service_account:reactivate` reactivates a service account with a reason THEN the system SHALL apply the change only if that person is not the one who performed the last deactivation, and SHALL write an audit entry with the reason.
17. IF the person who performed the last deactivation tries an ordinary reactivation of the same service account THEN the system SHALL refuse with code `separacao_de_funcoes`.
18. WHEN no other active person holds `identity:service_account:reactivate` THEN the system SHALL allow a contingency reactivation by the same person only after the configurable waiting time (initially 30 minutes) since the deactivation, with a reason of at least 30 characters and the identification of the emergency.
19. WHEN a contingency reactivation is performed THEN the system SHALL write an audit entry marked `contingencia`, SHALL emit a structured log alert at level `warn` and SHALL open a pending independent-review item visible to every holder of `identity:service_account:reactivate` or `identity:service_account:deactivate` and to every holder of `audit:log:read`.
20. The system SHALL allow the independent-review item to be closed only by a person different from the one who reactivated who holds `identity:service_account:reactivate` or `identity:service_account:deactivate`.
21. WHILE an independent-review item opened by a person is still open the system SHALL refuse a new contingency reactivation by that same person, and SHALL repeat the alert once per day.
22. The system SHALL NOT allow a contingency reactivation of a service account whose deactivation was performed by the synchronization because the account was removed from code.
23. The system SHALL keep the type of an account (`HUMAN` or `SERVICE`) and the type of a role immutable after creation, and SHALL reject at the database level any update that changes them.
24. The system SHALL reject at the database level, on insert and on update, any session, password recovery token, administrative membership or role assignment that is incompatible with the account type, and any service account record that points to a `HUMAN` account, regardless of the code path that writes it.

**Independent Test**: declarar uma conta de serviço de teste, partir, ver a conta criada; tentar login, recuperação e sessão e ver as recusas; desativar e ver um caso de uso recusado com o `Principal` dela.

---

### P1: Solicitação, decisão e execução separadas ⭐ MVP

**User Story**: Como presidente ou diretor, quero que operações sensíveis separem quem solicita, quem decide e quem executa, e que processos automáticos executem só regras aprovadas ou decisões humanas registradas, nunca juízo próprio, para que nenhum ato de governança dependa de uma pessoa só nem de um processo.

**Why P1**: Contrato usado por reembolso (`04`/`06`), liberação de vaga (`05`), correção de check-in (`07`) e procedimento (`14`).

**Acceptance Criteria**:
1. The system SHALL provide a platform rule that, given the actor and the authors of previous acts declared incompatible by the use case, refuses the act when the actor is one of those authors.
2. WHEN the separation rule refuses an act THEN the use case SHALL respond `forbidden` with code `separacao_de_funcoes` and SHALL write a `denied` audit entry naming the rule.
3. IF an actor attempts a sensitive act whose subject is the actor's own person THEN the use case SHALL refuse it with code `ato_sobre_si_mesmo` and SHALL write a `denied` audit entry.
4. The system SHALL require every decision record to carry an id, the deciding person, a reason of at least 10 characters, the scope it authorizes, the time and its state.
5. WHEN a service account uses a `servico` permission marked as requiring a decision THEN the use case SHALL require a `decision_id` in state `APROVADA`, of the same scope and not yet executed.
6. IF the `decision_id` is missing, unknown, `REVOGADA`, `EXECUTADA` or of a different scope THEN the use case SHALL refuse the execution, SHALL produce no effect and SHALL write a `denied` audit entry.
7. WHEN two executions of the same decision run concurrently THEN exactly one SHALL succeed and the other SHALL be refused.
8. The system SHALL NOT allow a service account to create, approve or revoke a decision record.
9. The system SHALL require every permission of class `servico` to declare whether it requires a decision, and IF the declaration is missing THEN the API and the worker SHALL refuse to start.
10. WHEN a service account uses a `servico` permission that does not require a decision THEN the use case SHALL execute the approved rule without any human decision record and SHALL record the rule name and version.
11. The system SHALL NOT grant any permission whose act involves discretionary judgment (approve a refund, decide a suspension, release a quota slot, correct attendance, grant a commercial release) to a service role.

**Independent Test**: um módulo de teste com a regra "solicitante ≠ aprovador", uma execução por regra aprovada (sem decisão, com `regra` na auditoria) e uma execução de decisão: ver a recusa do mesmo autor, a recusa da execução de decisão sem `decision_id`, a execução por regra aceita sem decisão e uma única execução sob concorrência.

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
10. WHEN a service account executes an approved deterministic rule THEN the entry context SHALL contain `regra` with the rule name and version, and SHALL NOT require `decision_id`.

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

**Why P1**: Exigência do mantenedor e do parecer externo: a lista atual de verbos proibidos não cobre verbos novos, e restringir só por verbo ainda aceitaria `read` sobre qualquer recurso sensível novo sem revisão.

**Lista explícita de permissões autorizadas ao Conselho Fiscal** (versionada no código, com teste; corresponde exatamente ao que ele tem hoje):

| # | Permissão |
|---|---|
| 1 | `audit:log:read` |
| 2 | `financeiro:conta:read` |
| 3 | `financeiro:lancamento:read` |
| 4 | `financeiro:saldo:read` |
| 5 | `financeiro:comprovante:read` |
| 6 | `financeiro:prestacao_contas:read` |
| 7 | `financeiro:prestacao_contas:approve` |
| 8 | `financeiro:parecer:opine` |
| 9 | `estoque:produto:read` |
| 10 | `estoque:movimentacao:read` |
| 11 | `estoque:saldo:read` |

**Acceptance Criteria**:
1. The system SHALL hold a versioned explicit list of the permissions that may be granted to `CONSELHO_FISCAL`, initially the eleven permissions in the table above.
2. IF a contribution grants `CONSELHO_FISCAL` a permission that is not in the explicit list THEN the API and the worker SHALL refuse to start, even when its action is `read`.
3. IF a permission in the explicit list has an action other than `read` or `opine`, or `approve` on resource `prestacao_contas` THEN the API and the worker SHALL refuse to start.
4. IF a contribution grants `CONSELHO_FISCAL` a permission of class `servico` THEN the API and the worker SHALL refuse to start.
5. WHEN the explicit list replaces the current deny-list THEN `CONSELHO_FISCAL` SHALL keep exactly its current eleven effective permissions.
6. The system SHALL fail its test suite when the explicit list changes without the matching update of its versioned snapshot, so that every new access of the Conselho Fiscal requires a reviewed change.

**Independent Test**: tentar conceder ao Conselho Fiscal `acesso:checkin:create`, `cobrancas:reembolso:approve`, um verbo inventado e uma leitura nova fora da lista (por exemplo `acesso:procedimento:read`) e ver a partida falhar; ver as 11 permissões atuais intactas.

---

### P1: Nomeação institucional do Conselho Fiscal e papéis incompatíveis ⭐ MVP

**User Story**: Como presidente, quero registrar a nomeação e a saída de membros do Conselho Fiscal decididas pela instância competente, sem eu mesmo ter os poderes de aprovar contas, e sem que uma pessoa acumule papéis que comprometam a independência do Conselho.

**Why P1**: Consequência verificada da D-02-3. `authz.Covers` exige que quem atribui um papel, promove, revoga, desativa ou redefine a senha tenha **todas** as permissões do papel ou do alvo (`identity/app/assign_roles.go:87`, `promote_admin.go:78`, `revoke_admin.go:74`, `deactivate_user.go:88`, `admin_reset_password.go:75`). Hoje só o `PRESIDENTE` cobre as permissões do Conselho Fiscal. Depois da D-02-3, **ninguém** as cobriria, porque o Conselho não tem permissões de `identity`. Ninguém mais conseguiria nomear, retirar, desativar ou redefinir a senha de um conselheiro. A D-02-3 só pode ser aplicada junto com esta regra.

**Acceptance Criteria**:
1. The system SHALL mark `financeiro:prestacao_contas:approve` and `financeiro:parecer:opine` as institutional permissions of the Conselho Fiscal.
2. WHEN a person registers the appointment or the removal of a member of the Conselho Fiscal (assigning or removing the role `CONSELHO_FISCAL`) THEN the use case SHALL require `identity:conselho_fiscal:manage` and `authz.Covers` over every permission involved except the institutional ones.
3. WHEN such an appointment or removal is registered THEN the use case SHALL require a reference to the act of the competent body (body, act identification and date), a stored document of that act and a reason of at least 10 characters, and SHALL record it as the registration of an external act, never as a decision of the actor.
4. IF the actor of an appointment or removal holds the role `CONSELHO_FISCAL` or is the target of the act THEN the use case SHALL refuse with code `separacao_de_funcoes`.
5. WHEN a person deactivates or reactivates the account of a holder of `CONSELHO_FISCAL` (technical blocking) THEN the use case SHALL keep requiring `identity:user:update` and SHALL apply `authz.Covers` over the target's permissions except the institutional ones, without requiring the appointment authorization.
6. The system SHALL keep the full `authz.Covers` rule, with no exception, for the administrative password reset of a holder of `CONSELHO_FISCAL`, whose access recovery SHALL use only the self-service e-mail recovery.
7. The system SHALL apply the exclusion of institutional permissions in `authz.Covers` only to the operations listed in AC2 and AC5, and a test SHALL fail if any other use case uses it.
8. The system SHALL refuse to assign the role `CONSELHO_FISCAL` to a user who holds `PRESIDENTE`, and SHALL refuse to assign `PRESIDENTE` to a user who holds `CONSELHO_FISCAL`, checking by user identity.
9. WHILE an existing user holds both `PRESIDENTE` and `CONSELHO_FISCAL` the system SHALL keep every historical record, SHALL report the user as an irregularity on every start and in the identity administration, and SHALL deny that user the institutional permissions until the irregularity is resolved by removing one of the roles.
10. The system SHALL evaluate every separation-of-duties rule by user identity over the union of all the user's roles, so that holding several roles never bypasses a rule.
11. The system SHALL hold the list of incompatible role pairs as versioned configuration that initially contains only `PRESIDENTE` × `CONSELHO_FISCAL`; any other pair SHALL be added only from a rule of the statute approved by the maintainer.
12. The system SHALL accept an appointment or removal only when the deciding body belongs to a versioned list of competent bodies (initially only the General Assembly), the act date is not in the future and the act document exists, SHALL record the registering person as registrar and the body as the author of the decision, and SHALL provide no path that changes the composition of the Conselho Fiscal without such an act.

**Independent Test**: com a matriz após a D-02-3, mostrar que atribuir `CONSELHO_FISCAL` pelo caminho comum (`AssignRoles`/`PromoteToAdmin`) é recusado com `ato_institucional_obrigatorio`; que o caso de uso institucional sem `identity:conselho_fiscal:manage` é recusado; que com a permissão, o documento do ato e o motivo, passa; que com o próprio ator como alvo, falha; que o bloqueio técnico pelo `PRESIDENTE` passa e a redefinição administrativa de senha do conselheiro é recusada; e que uma pessoa com `PRESIDENTE` + `CONSELHO_FISCAL` é recusada na atribuição e, se já existir, perde só os atos institucionais e aparece como irregularidade.

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
| T-18 | Conselho Fiscal com `read` fora da lista explícita (ex. `acesso:procedimento:read`) → partida falha | Unidade | CFS-01 |
| T-19 | Lista explícita alterada sem atualizar o snapshot → suíte falha | Unidade (snapshot) | CFS-01 |
| T-20 | Execução por regra aprovada (confirmação de pagamento de teste) sem decisão → aceita, auditoria com `regra` e sem `decision_id` | Integração | SOD-01, AUT-01 |
| T-21 | Permissão `servico` sem declarar "exige decisão" → partida falha | Unidade | SOD-01 |
| T-22 | Mesma pessoa desativa e tenta reativar a conta de serviço → `separacao_de_funcoes` | Integração | SVC-01 |
| T-23 | Reativação de contingência antes de 30 min → recusada; depois de 30 min com motivo → aceita, marcada `contingencia` e com revisão posterior aberta | Integração (relógio do banco controlado) | SVC-01 |
| T-24 | Matriz após a D-02-3: `PRESIDENTE` sem `prestacao_contas:approve` e `parecer:opine`; Conselho Fiscal com as 11; diferença do snapshot igual exatamente à exceção aprovada | Unidade (snapshot) | TRN-01 |
| T-25 | Atribuir `CONSELHO_FISCAL` pelo caminho comum → `ato_institucional_obrigatorio`; caso de uso institucional sem `identity:conselho_fiscal:manage` → `forbidden`; com a permissão, documento e motivo → passa; alvo = ator → `separacao_de_funcoes` | Integração | INS-01 |
| T-26 | Usuário com `PRESIDENTE` recebendo `CONSELHO_FISCAL` → recusado; usuário existente com os dois → pendência na partida, sem revogação | Integração | INS-01 |
| T-27 | Usuário com dois papéis, sendo um o autor de um ato anterior incompatível → separação recusada pela identidade, não pelo papel | Unidade | INS-01, SOD-01 |
| T-28 | Binário antigo depois do novo, mesma versão com hash diferente, versão menor, API e worker concorrentes, `rbac-rollback` sem autorização, com plano alterado e com autorização válida → matriz nova preservada, recusas corretas e autoria humana registrada | Integração (PostgreSQL) | TRN-01 |
| T-28b | Autorização de rollback que restaura permissões institucionais sem decisão do mantenedor e documento, ou feita por quem ganharia as concessões → recusada | Integração | TRN-01, INS-01 |
| T-29 | Alterar tipo de usuário ou de papel; sessão, token de recuperação, vínculo administrativo e papel incompatíveis por INSERT e UPDATE, em SQL direto como `tj_app` → recusados pelo banco | Integração (PostgreSQL) | SVC-01 |

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

**Nenhuma** pendência muda a estrutura do Design. Situação das decisões do mantenedor em 2026-10-14 (nenhuma concessão real foi aplicada no código):

- **D-02-1 — aprovada**: `ADMIN_SISTEMA` e `PRESIDENTE` recebem `identity:service_account:deactivate` e `:reactivate` por concessão explícita. A reativação exige pessoa diferente de quem desativou; a contingência (espera configurável, inicial 30 min, motivo, emergência, auditoria, alerta e revisão independente que quem reativou não pode fechar) cobre o caso de uma única pessoa habilitada (SVC-01 AC16–22).
- **D-02-2 — aprovada**: 18 sensíveis e 12 normais; 16 concessões explícitas de transição ao `PRESIDENTE`, com justificativa individual e testes de compatibilidade.
- **D-02-3 — aprovada**: retirada de `financeiro:prestacao_contas:approve` e `financeiro:parecer:opine` do `PRESIDENTE`, registrada como única exceção à preservação da matriz (TRN-01 AC8). **Só pode ser aplicada junto com a D-02-4**, senão ninguém mais consegue nomear ou retirar membros do Conselho Fiscal.
- **D-02-4 — aprovada como direção arquitetural**: autorização institucional específica, referência verificável ao ato, exceção de `Covers` só nas operações listadas, nomeação separada do bloqueio técnico, incompatibilidade `PRESIDENTE` × `CONSELHO_FISCAL` por identidade, irregularidades preservadas e bloqueadas até regularização, outras incompatibilidades só pelo estatuto (INS-01).
- **D-02-5 — aprovada** (2026-10-14): concessão explícita de `identity:conselho_fiscal:manage` ao `PRESIDENTE`, limitada ao **registro** de nomeações e destituições formalmente decididas pela instância competente, com evidência documental e auditoria. Não confere competência para decidir unilateralmente a composição do Conselho Fiscal. Aplicada no mesmo incremento que a D-02-3. `jobs:job:close` continua sem concessão no G1.
- **Ajustes obrigatórios do Design (2026-10-14)**: compatibilidade do RBAC durante o deploy (TRN-01 AC9–12) e integridade das contas de serviço no banco (SVC-01 AC23–24).

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
|---|---|---|---|
| CLS-01 | P1: Classe obrigatória e fail-closed | Tasks | Mapped |
| TRN-01 | P1: Transição sem perda nem escalonamento | Tasks | Mapped |
| SVC-01 | P1: Contas de serviço não interativas | Tasks | Mapped |
| SOD-01 | P1: Solicitação, decisão e execução separadas | Tasks | Mapped |
| AUT-01 | P1: Autoria e rastreabilidade | Tasks | Mapped |
| FIS-01 | P1: Papel FISCAL e contrato de escopo por jogo | Tasks | Mapped |
| CFS-01 | P1: Conselho Fiscal fail-closed | Tasks | Mapped |
| INS-01 | P1: Nomeação institucional do Conselho Fiscal e papéis incompatíveis | Tasks | Mapped |

**Coverage:** 8 total, 8 mapped to tasks, 0 unmapped (tasks em revisão; nenhuma implementada).

## Success Criteria

- [ ] A matriz efetiva dos 8 papéis atuais é idêntica antes e depois, salvo a exceção aprovada D-02-3 (teste de retrato verde com a diferença registrada).
- [ ] Depois da D-02-3, continua possível nomear e retirar membros do Conselho Fiscal (INS-01), e ninguém acumula `PRESIDENTE` e `CONSELHO_FISCAL`.
- [ ] Nenhuma permissão sensível nova está concedida a nenhum papel ao fim desta spec, salvo as concessões aprovadas D-02-1, D-01-1 e D-02-5.
- [ ] Toda entrada de auditoria nova tem autor identificável; `system` só no bootstrap e na sincronização.
- [ ] Uma conta de serviço não consegue sessão por nenhum caminho (aplicação e banco).
