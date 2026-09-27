# fundacao-core Validation

## Validation: fundacao-core - PASS ✅

**Data**: 2026-09-27
**Spec**: `.specs/features/fundacao-core/spec.md`
**Faixa de commits desta iteração**: `6d9f352..8d53101` (5 commits sobre a fase 12 já mesclada: `a9eb6ba`, `95b7ce7`, `e55e764`, `cb876de`, `8d53101`), branch `feature/fundacao-core-validacao`; 28 arquivos, +959/-157 linhas. Faixa completa da feature: `f769065^..8d53101`.
**Verifier**: sub-agente independente (autor != verificador), iteração 2 de no máximo 3, sem herdar o contexto da iteração 1; a árvore real não foi alterada além deste arquivo (`git status --porcelain` idêntico antes e depois do sensor — ver seção do sensor).

**Veredito: PASS.** As quatro correções que a iteração 1 pediu (F1 a F8) foram verificadas com evidência própria, não com a palavra do relatório anterior: o mapeamento de 503 `service_unavailable` existe e está testado nos três pontos (identity/http, platform/audit/http, platform/httpx/authn), a matriz do ADMIN_SISTEMA e do PRESIDENTE agora é afirmada por igualdade exata, o payload de `admin.promote` e de `user.deactivate`/`user.reactivate` é comparado por decodificação estrutural, `occurred_at`/`created_at`/`granted_at` saem em UTC com teste que prova isso a partir de um fuso não-UTC, o regex do `X-Request-Id` rejeita `_`, `.`, `/`, `:` e `+`, o caso de borda de papéis vazios foi corrigido só na spec (sem tocar em código), existe teste de 8 logins concorrentes com sessões independentes, e a documentação (`STATE.md`, `design.md`, `tasks.md`, a tabela de rastreabilidade e a matriz inicial da spec) está consistente com o código, incluindo a referência real à issue #19 (`SEC-001`). O sensor de mutação desta iteração, focado nos arquivos tocados (18 mutantes, incluindo os 4 antigos sobreviventes M37/M39/M58/M69 reproduzidos ponto a ponto), morreu 18 de 18. Todos os gates (Go, web, contrato, `validate_spec.py`, `validate_tasks.py`) saem com 0.

**Placar**

- Spec-anchored: 25 requisitos (IDN-01 a 08, RBAC-01 a 03, AUD-01 a 04, EML-01, MNY-01 a 03, TST-01 a 03, API-01, API-02, PLT-01), conferidos abaixo com evidência `file:line`. Todos os ACs têm evidência; nenhuma lacuna aberta. 3 spec-precision gaps menores seguem sinalizados (não bloqueiam: não têm resultado preciso definido na spec, e o código faz uma escolha razoável e testada).
- Gate: tudo verde (ver "Gate Check").
- Sensor: 18 mutantes injetados em scratch (worktree temporário), 18 mortos, 0 sobreviventes.
- Lacunas ranqueadas: nenhuma que bloqueie o veredito (ver "Achados adicionais" para os itens cosméticos remanescentes).

---

## Histórico

- **Iteração 1** (2026-09-26, `f769065^..6d9f352`): veredito **FAIL**. 75/79 mutantes mortos, 4 sobreviventes (M37 payload de `admin.promote`, M39 payload de `user.deactivate`/`reactivate`, M58 ADMIN_SISTEMA sem asserção exata, M69 regex do request id); 1 lacuna sem implementação e sem teste (banco indisponível → 503 `service_unavailable`, um teste existente até afirmava 500); 1 conflito interno da spec (caso de borda de papéis vazios contradizia o Assumption aprovado). Gate 100% verde já naquela rodada.
- **O que mudou até esta iteração**: 5 commits (`a9eb6ba`, `95b7ce7`, `e55e764`, `cb876de`, `8d53101`) implementando F1 a F8 do plano de correção aprovado pelo mantenedor: `database.Unavailable` novo, com os três pontos de mapeamento HTTP e o contrato atualizado (F1); asserção exata da matriz do ADMIN_SISTEMA/PRESIDENTE (F2); payload de `admin.promote` por igualdade estrutural (F3); UTC em `occurred_at`/`created_at`/`granted_at` com teste dedicado (F4); regex do request id reforçado e payload de `user.deactivate`/`reactivate` formalizado na spec e testado (F5); caso de borda de papéis vazios corrigido só na spec (F6); teste de login concorrente (F7); documentação (`STATE.md`, `design.md`, `tasks.md`, rastreabilidade, e-mail redigido no log interno) (F8).
- **Esta iteração (2)**: re-verificou as 8 correções com evidência própria (não confiou no relato da iteração 1 nem no changelog dos commits), refez a cobertura ancorada na spec do zero, rodou um sensor de mutação novo (18 mutantes, focado nos arquivos tocados, incluindo os 4 antigos reproduzidos) e todos os gates. Resultado: PASS. Nenhuma lacuna nova foi encontrada; os spec-precision gaps menores já sinalizados na iteração 1 (não normativos) continuam sinalizados, sem bloquear.

---

## Task Completion

Todas as tarefas de `tasks.md` (T1 a T86, fundação) seguem `[x]`; `tasks.md:12` diz `Status: Implemented` e cita corretamente a fase 12 mesclada em `develop` e a validação final em andamento nesta branch. As correções F1 a F8 desta iteração não têm tasks formais próprias (foram tratadas como fix tasks da validação, conforme o fluxo do skill); cada uma foi conferida individualmente abaixo.

| Item | Commit | Status | Evidência própria desta iteração |
|---|---|---|---|
| F1 (503 `service_unavailable`) | `a9eb6ba` | ✅ Done | `api/internal/platform/database/unavailable.go`, 3 pontos de mapeamento, contrato, testes unitário + integração |
| F2 (matriz exata) | `95b7ce7` | ✅ Done | `roles_matrix_test.go:139`, `:160` |
| F3 (payload `admin.promote`) | `95b7ce7` | ✅ Done | `promote_admin_test.go:73` (`assertPromoteBeforeAfter`) |
| F4 (UTC) | `95b7ce7` | ✅ Done | `handler.go:128`, `users_handler.go:24`,`:27`, `audit/query.go:191`; 3 testes dedicados |
| F5 (request id + payload deactivate/reactivate) | `e55e764` | ✅ Done | `requestid.go:18`, `requestid_test.go:56`-`60`; `deactivate_user_test.go:73` (`assertActivationBeforeAfter`) |
| F6 (spec: papéis vazios) | `cb876de` | ✅ Done | `spec.md:584`; só doc, confirmado sem diff de código (`git show cb876de --stat`) |
| F7 (login concorrente) | `cb876de` | ✅ Done | `authenticate_test.go:414` (`TestConcurrentLoginsForTheSameUserCreateIndependentSessions`) |
| F8 (documentação + redação de e-mail) | `8d53101` | ✅ Done | `STATE.md`, `design.md:272`,`:483`-`488`, `tasks.md:12`, `spec.md:595`-`619`; `handler.go:218`-`227` (`emailInErrorMessage`) + `handler_internal_test.go:139` |

---

## Spec-Anchored Acceptance Criteria

Regra aplicada: para cada AC, o resultado definido pela spec é comparado com a asserção do teste (valor, código, estado), não com a existência de uma asserção. Evidência-ou-zero: AC sem `file:line` conta como não coberto. Convenção de caminhos: relativos a `api/internal/`, salvo `cmd/` (`api/cmd/`), `web/` e `api/openapi/`.

### IDN-01 Primeiro administrador

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| 1 | usuário ativo, papel (ADMIN_SISTEMA padrão ou PRESIDENTE), vínculo `bootstrap-admin` sem concedente, `user.bootstrap` sem ator com e-mail/papel/motivo, nunca a senha | `identity/app/bootstrap_admin_test.go:29` (`TestBootstrapOnAnEmptyDatabaseCreatesTheAdministrator`) e `:119` (`TestBootstrapAuditsWithoutActorAndNeverStoresThePassword`); `cmd/bootstrap-admin/main_test.go:47` | PASS |
| 2 | vínculo ativo existente: recusa sem criar | `bootstrap_admin_test.go:74`, `:90`, `:208` (concorrência) | PASS |
| 3 | senha ausente/fraca: recusa | `bootstrap_admin_test.go:172`; `main_test.go:104` | PASS |
| 4 | senha não é argumento de linha de comando | `main_test.go:88` (`TestRunHasNoPasswordFlagAndNeverEchoesTheValue`) | PASS |

### IDN-02 Login, sessão e logout

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| 1 | 200 igual ao `me`; cookie `tj_session` HttpOnly/Secure/SameSite=Lax/Path=/ | `identity/http/auth_handler_test.go` (citado e reconfirmado íntegro nesta iteração); `httpapi/e2e_test.go:151`,`:210` | PASS |
| 2 | só hash SHA-256 no banco | `identity/infra/session_repository_test.go` | PASS |
| 3, 4 | 401 `invalid_credentials`, corpo idêntico p/ e-mail inexistente e inativo | `identity/app/authenticate_test.go:119`-`180` (reconfirmado nesta iteração) | PASS |
| 5 | ociosidade 60 min / teto 8 h | `identity/domain/session_test.go`; `httpapi/e2e_test.go:553` (`TestE2EAnIdleSessionExpires`) | PASS |
| 6 | logout revoga e limpa cookie | `httpapi/e2e_test.go:644` | PASS |
| 7 | login emite token novo mesmo com cookie enviado | `authenticate_test.go:339` (área) | PASS |
| 8 | 5ª falha em 15 min: 429 + `Retry-After` | `identity/domain/lockout_test.go`; `auth_handler_test.go:152` | PASS |
| 9 | `me` com campos exatos, sem hash/token | `auth_handler_test.go:57`-`90` | PASS |
| 10 | nunca bloqueio permanente | `lockout_test.go:46` | PASS |

### IDN-03 CSRF e origem

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| 1, 2 | POST/PUT/PATCH/DELETE exige `X-CSRF-Token` igual; senão 403 `csrf_invalid` | `platform/httpx/csrf_test.go`; `httpapi/e2e_test.go:318` (`TestE2EStateChangesWithoutCSRFAreRefused`) | PASS |
| 3, 4 | `Origin` fora da lista: 403 `origin_not_allowed`, também no login | `httpapi/e2e_test.go:332` (`TestE2EOriginIsCheckedOnStateChanges`) | PASS |
| 5 | GET/HEAD/OPTIONS não alteram estado | `httpapi/e2e_test.go:625` (`TestE2EReadsNeverChangeState`) | PASS |
| 6 | autenticada, sem `Origin`, `Sec-Fetch-Site: cross-site`: 403 | `e2e_test.go:339`; `csrf_test.go` | PASS |

### IDN-04 Gestão de usuários

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| 1 | cria ativo, só ASSOCIADO, `must_change_password=true`, 201, `user.create` | `identity/app/create_user_test.go`; `httpapi/e2e_test.go:379` | PASS |
| 2 | e-mail repetido (case-insensitive): 409 `email_taken` | `identity/infra/user_repository_test.go` | PASS |
| 3 | desativa: sessões + tokens revogados na mesma transação; `user.deactivate` com before/after **exatamente** `{"active":true}`/`{"active":false}` | `identity/app/deactivate_user_test.go:47` e `:73` (`assertActivationBeforeAfter`, decodificação JSON exata) — **reconfirmado nesta iteração como F5**; `httpapi/e2e_test.go:379` | PASS |
| 4 | último com `identity:admin:grant`: 409 `last_admin` | `deactivate_user_test.go:128`; `assign_roles_test.go:170` | PASS |
| 5, 6 | troca de senha: hash atualizado, outras sessões revogadas, `user.password_change`; senha atual errada: 403 | `identity/app/change_password_test.go` | PASS |
| 7 | `identity:role:assign` substitui o conjunto, audita `user.roles_set`, vale na próxima requisição | `identity/app/assign_roles_test.go:36`; `httpapi/e2e_test.go:396` (`TestE2ERolesChangeApplyOnTheNextRequest`) | PASS |
| 8 | cursor (`created_at`,`id`), 50/pg, filtros, `limit>100`⇒422 | `identity/app/list_users_test.go`; `httpapi/e2e_test.go:577` | PASS |
| 9 | nunca hash/token em resposta, log ou auditoria | `identity/http/users_handler_test.go`; `httpapi/e2e_test.go:597` (`TestE2ENoUserResponseCarriesAHashOrAToken`) | PASS |
| 10 | reativa sem restaurar papel/vínculo; `user.reactivate` before/after exatamente `{"active":false}`/`{"active":true}` | `deactivate_user_test.go:256`-`266` (`assertActivationBeforeAfter`) — **reconfirmado como F5** | PASS |
| 11 | nova senha == atual: 422 `password_unchanged` | `change_password_test.go` | PASS |
| 12 | 5 erros de senha atual em 15 min: 429 `password_change_blocked`, contador próprio | `change_password_test.go`; `auth_handler_test.go:393` | PASS |

### IDN-05 Política de senha

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| 1-4 | mínimos 10/8, máximo 128, lista de comprometidas | `identity/domain/user_test.go`; `identity/app/create_user_test.go`; `platform/password/denylist_test.go` | PASS |
| 5, 8 | primeiro vínculo administrativo e criação/reset administrativo: `must_change_password=true` | `promote_admin_test.go:34` (`stored.MustChangePassword`); `create_user_test.go`; `admin_reset_password_test.go` | PASS |
| 6 | troca bem-sucedida: `must_change_password=false` | `change_password_test.go` | PASS |
| 7 | comprimento em pontos de código, sem regra de composição | `domain/user_test.go` | PASS |

### IDN-06 Vínculo administrativo e promoção

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| 1 | numa transação: vínculo, papéis, `must_change_password`, sessões revogadas; `admin.promote` com **before/after estruturalmente exatos** (papéis, `admin_membership`, `must_change_password`) | `identity/app/promote_admin_test.go:34` e `:73`-`104` (`assertPromoteBeforeAfter`, decodifica JSON e afirma chaves e valores exatos) — **reconfirmado nesta iteração como F3, e a mutação que troca `before` por `after` morre (ver sensor M14/F3)** | PASS |
| 2-4 | motivo curto/ausente 422; já admin 409; inativo 409 | `promote_admin_test.go:204` (`TestPromoteValidatesReasonTargetStateAndExistingMembership`) | PASS |
| 5 | um vínculo ativo por usuário, no banco; nunca apaga linha | `identity/infra/admin_membership_repository_test.go`; `platform/database/identity_migration_test.go` | PASS |
| 6, 7 | retirada fecha o vínculo, remove papéis != ASSOCIADO, revoga sessões, `admin.revoke`; sem vínculo: 409 `not_admin` | `identity/app/revoke_admin_test.go` | PASS |
| 8 | ficaria sem `identity:admin:grant`: 409 `last_admin` | `revoke_admin_test.go`; `assign_roles_test.go:170` | PASS |
| 9 | papel != ASSOCIADO sem vínculo: 409 `admin_membership_required` | `assign_roles_test.go:132` (área) | PASS |
| 10 | promoção sem papel administrativo: 422 `admin_role_required` | `promote_admin_test.go:188` (`TestPromoteRequiresAtLeastOneAdministrativeRole`) | PASS |

### IDN-07 Recuperação de acesso por e-mail

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| 1 | 202 idêntico exista ou não a conta | `httpapi/e2e_test.go:461` (`TestE2EPasswordRecoveryByEmail`) | PASS |
| 2, 3 | token 256 bits só hash, 30 min, invalida anteriores, envio depois da resposta; nunca grava/audita o token | `identity/app/request_password_reset_test.go` | PASS |
| 4 | > 3/hora por hash: mesma resposta, nada criado | `request_password_reset_test.go` | PASS |
| 5, 6, 7 | confirma numa transação; token inválido/expirado/usado: 400 idêntico; política violada: 422, token continua válido | `identity/app/reset_password_test.go`; `e2e_test.go:461` | PASS |
| 8 | eventos sem e-mail/token/link | `request_password_reset_test.go`; `reset_password_test.go` | PASS |
| 9 | inativo: mesma resposta, nada criado | `request_password_reset_test.go` | PASS |
| 10 | rotas públicas com checagem de origem | `auth_handler_test.go` (área) | PASS |

### IDN-08 Redefinição administrativa

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| 1 | numa transação: temporária, hash trocado, `must_change_password`, sessões/tokens revogados, `user.password_reset`; 200 com a senha uma vez, `Cache-Control: no-store` | `identity/app/admin_reset_password_test.go`; `httpapi/e2e_test.go:488` (`TestE2EAdministrativePasswordReset`, confere o header) | PASS |
| 2 | ator sem permissão do alvo: 403 `privilege_escalation` | `admin_reset_password_test.go` | PASS |
| 3 | própria conta: 403 `self_change_forbidden` | `admin_reset_password_test.go`; `e2e_test.go:450` | PASS |
| 4 | alvo inativo: 409 `user_inactive` | `admin_reset_password_test.go` | PASS |
| 5 | negada por AC 2/3: `user.password_reset` `denied` | `admin_reset_password_test.go` | PASS |
| 6 | motivo ausente/curto: 422 `reason_required` | `admin_reset_password_test.go` | PASS |
| 7 | nunca guarda/audita/envia a temporária, a antiga ou hash | `admin_reset_password_test.go`; `e2e_test.go:507` (`noSecretsLeaked`) | PASS |

### EML-01 Capacidade de e-mail

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| 1 | interface `platform/email` independente de provedor | `platform/email/email.go`; `internal/architecture_test.go` (platform nunca importa módulo) | PASS |
| 2-5 | validação de mensagem; `log`/`disabled`; falha vira incidente; provedor desconhecido não inicia | `platform/email/email_test.go` | PASS |

### RBAC-01 Modelo de permissões e papéis

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| 1 | formato `^[a-z_]+:[a-z_]+:[a-z_]+$` | `platform/authz/permission_test.go` | PASS |
| 2, 3 | sincronização idempotente sob advisory lock; `rbac.sync` com before/after completos | `identity/infra/role_repository_test.go` | PASS |
| 4 | provisiona os 8 papéis | `identity/app/roles_matrix_test.go:23` (`TestFoundationMatrixHasTheEightRoles`) | PASS |
| 5, 6 | Conselho Fiscal sem create/update/delete/cancel; com as 3 institucionais | `roles_matrix_test.go:43` (`TestConselhoFiscalHasTheInstitutionalPermissionsAndNeverCreateUpdateDeleteCancel`, conjunto exato por `slices.Equal`) | PASS |
| 7 | união das permissões de vários papéis | `identity/infra/role_repository_test.go` | PASS |
| 8 | permissão removida do código fica inativa, não conta | `role_repository_test.go` | PASS |
| 9 | recusa iniciar com matriz inválida | `roles_matrix_test.go:183` (`TestBuildMatrixRefusesInvalidContributions`) | PASS |
| 10 | PRESIDENTE com todo o catálogo, vínculos explícitos | `roles_matrix_test.go:83`,`:160` (`TestPresidenteCoversTheWholeCatalog`, conjunto exato) | PASS |
| — | ADMIN_SISTEMA com o conjunto exato, sem permissão institucional | `roles_matrix_test.go:139` (`TestAdminSistemaHasExactlyTheTechnicalPermissionsAndNoInstitutionalOne`, `slices.Equal` + negação de prefixo `financeiro:`) — **F2, reconfirmado; morre no sensor (M6/M18)** | PASS |

### RBAC-02 Autorização negada por padrão

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| 1 | rota fora da lista pública sem sessão: 401 | `httpapi/router_test.go`; `httpapi/e2e_test.go:344` | PASS |
| 2 | sem permissão: 403, sem revelar recurso | `httpapi/e2e_test.go:356` (compara corpo de recurso existente x inexistente) | PASS |
| 3 | `authz.Require` antes de ler/escrever | um teste "ChecksThePermissionFirst" por caso de uso (`create_user_test.go`, `deactivate_user_test.go`, `promote_admin_test.go`, etc.) | PASS |
| 4 | suíte falha se alguma rota não é pública nem protegida | `httpapi/router_test.go` | PASS |
| 5 | inativo invalida toda sessão | `identity/app/session_service_test.go` | PASS |
| 6 | `must_change_password`: 403 em tudo, exceto logout/`me`/troca | `httpapi/e2e_test.go:234` (`TestE2EFirstLoginForcesThePasswordChange`) | PASS |
| 7 | `authz.denied` gravado, exceto leitura comum | `platform/audit/authz_hook_test.go` | PASS |
| 8 | nunca decide por nome de papel | `platform/authz/rolename_guard_test.go` (varre o código-fonte) | PASS |

### RBAC-03 Concessão sem escalada

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| 1, 2 | conceder/remover papel com permissão que o ator não tem: 403 `privilege_escalation` | `promote_admin_test.go:155` (`TestPromoteRefusesRolesWithPermissionsTheActorLacks`); `httpapi/e2e_test.go:432` | PASS |
| 3 | mexer no próprio acesso: 403 `self_change_forbidden` | `promote_admin_test.go:140` (`TestNobodyPromotesThemselves`); `e2e_test.go:450` | PASS |
| 4 | avaliação com permissões efetivas no momento do pedido | `identity/app/session_service_test.go` | PASS |
| 5, 6 | negada por 1/2/3: `role.change_denied`; desativar/reativar/reset negados por escalada: 403 | `promote_admin_test.go:124`,`:140`; `deactivate_user_test.go` | PASS |

### AUD-01 a AUD-04

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| AUD-01.1, .2 | mesma transação; falha: rollback e 500 `audit_failed` | `platform/audit/atomicity_test.go`; um teste `...RolledBackWhenTheAuditFails` por caso de uso | PASS |
| AUD-01.3 | campos completos; **`occurred_at` em UTC** | `platform/audit/query_test.go:14` (`TestRowRecordNormalizesOccurredAtToUTC`, fuso BRT fixo, afirma `Location()==UTC` e sufixo `Z`) — **F4, reconfirmado; morre no sensor (M9)** | PASS |
| AUD-01.4, .5, .6 | motivo obrigatório em cancelamento; redação de campos sensíveis; erro fora de transação | `platform/audit/entry_test.go`; `recorder_test.go` | PASS |
| AUD-02.1, .2 | UPDATE/DELETE/TRUNCATE rejeitados até para o dono; app só INSERT/SELECT | `platform/database/audit_log_migration_test.go:51`,`:73` (reconfirmado nesta iteração, inclui o papel dono) | PASS |
| AUD-02.3 | nenhuma operação da API altera/apaga auditoria | estrutural: `api/openapi/platform.yaml` só declara `GET /audit-logs`; `httpapi/routes_test.go` (paridade rota×contrato) | PASS (estrutural) |
| AUD-03.1-.4 | mais novo primeiro, cursor 50/pg; filtros com E; `limit>100`⇒422; sem permissão⇒403 | `platform/audit/http_test.go` | PASS |
| AUD-04.1-.4 | `auth.login`, `auth.login_failed` (hash de e-mail, sem senha/e-mail), `auth.login_blocked`, `auth.logout` | `identity/app/authenticate_test.go`; `httpapi/e2e_test.go:534` (`TestE2EFailedLoginIsAuditedWithoutThePasswordOrThePlainEmail`) | PASS |
| AUD-04.5 | evento para toda mudança/negação de papéis, permissões, vínculo e senhas | um evento por caso de uso, listado em `httpapi/e2e_test.go:294`-`301` (11 ações consultadas pela API) | PASS |
| AUD-04.6, .7 | evento de segurança em transação própria, falha vira incidente; ação fora do catálogo: erro | `platform/audit/recorder_test.go`; `entry_test.go` | PASS |

### MNY-01 a MNY-03

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| MNY-01.1-.5 | `Cents` int64; soma/sub exatas com `ErrOverflow`; `Cmp`; sem float | `platform/money/money_test.go:14`,`:35`,`:65`,`:77`,`:85` (reconfirmado) | PASS |
| MNY-02.1-.6 | JSON inteiro; string/decimal/expoente/null⇒erro; `ErrOutOfRange`; formatação `R$ 1.234,56`; parse; limites | `platform/money/json_test.go`; `platform/money/format_test.go:41`-`90` (reconfirmado, vetores compartilhados) | PASS |
| MNY-02.7, .8 | esquema barra `real`/`double precision`/`numeric` e `_cents` não-`bigint` | `platform/database/schema_guard_test.go:43`,`:55`,`:81` (reconfirmado nesta iteração) | PASS |
| MNY-02.9 | helper TS com número não seguro: `RangeError` | `web/lib/money.test.ts` | PASS |
| MNY-03.1-.6 | `Allocate` soma o total, resto nas primeiras; pesos inválidos; total negativo; `Percent` half-up; overflow; propriedade | `platform/money/allocate_test.go:20`-`144` (reconfirmado) | PASS |

### TST-01 a TST-03

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| TST-01.1 | sem a tag, `go test ./...` roda sem Docker | Gate desta iteração: `go test -count=1 ./...` saiu 0 sem Docker (ver Gate Check) | PASS |
| TST-01.2-.6 | Postgres 16 via testcontainers, banco isolado, papel de aplicação, CI roda com a tag, falha nomeando Docker sem a tag | `platform/testutil/postgres_test.go`, `db_test.go`; `.github/workflows/ci.yml` | PASS |
| TST-02.1-.5 | módulo não importa domain/infra/http de outro; platform não importa módulo; dependências para dentro; módulos descobertos por diretório | `internal/architecture_test.go` (lido e reconfirmado integralmente nesta iteração: `TestProductionCodeRespectsTheModuleBoundaries`, `TestDetectsAModuleImportingAnotherModulesInternals`, `TestDetectsPlatformImportingABusinessModule`, `TestDetectsOutwardDependenciesInsideAModule`, `TestAFutureModuleIsCoveredWithoutAList`, `TestTestFilesAreIgnored`) | PASS |
| TST-03.1-.3 | `pnpm test` roda `*.test.ts(x)` com Vitest/jsdom; CI antes do build; vetores de `money.ts` | `web/vitest.config.ts`; `web/lib/money.test.ts`; gate desta iteração: 54 testes passam | PASS |

### API-01, API-02

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| API-01.1 | contratos OpenAPI 3.0.3 por módulo + `common.yaml`, fonte das rotas | `api/openapi/*.yaml`; `httpapi/routes_test.go` | PASS |
| API-01.2, .3, .4 | código Go e tipos TS gerados versionados, CI falha se divergir; lint falha o CI | Gate desta iteração: `go generate ./...` sem diff; `pnpm gen:api:check` saída 0; `pnpm lint:api` saída 0 (Redocly válido) | PASS |
| API-01.5 | teste de integração valida status/cabeçalhos/corpo contra o contrato | `httpapi/e2e_test.go:114` (`e.contract.ValidateResponse` em toda chamada, inclusive nas novas do 503 — `audit/http_test.go`, `httpx/authn_test.go`) | PASS |
| API-01.6 | rota sem operação no contrato falha um teste | `httpapi/routes_test.go` | PASS |
| API-02.1-.7 | `problem+json` com os 6 campos; `X-Request-Id`; JSON inválido 400; validação 422 com `errors[]`; paginação; panic 500 sem detalhe; 413 nos dois limites | `platform/httpx/problem_test.go`, `requestid_test.go` (reconfirmado com os 5 casos novos de F5), `middleware_test.go`; contrato descreve o 503 em toda operação que já descrevia o 500 (`api/openapi/identity.yaml`, `platform.yaml`, reconfirmado nesta iteração) | PASS |

### PLT-01 Configuração, logs e migrações

| AC | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| PLT-01.1, .2 | variável obrigatória ausente: sai != 0 nomeando a variável, sem segredo; `COOKIE_SECURE=false` fora de `development`: recusa | `internal/config/config_env_test.go:112`,`:99`,`:151` (reconfirmado nesta iteração) | PASS |
| PLT-01.3, .4 | log estruturado por requisição; chaves sensíveis redigidas | `platform/httpx/middleware_test.go:77`; `platform/logx/logx_test.go` | PASS |
| PLT-01.5, .6, .7 | sem migração automática; papel de aplicação sem DDL; GORM sem valores no log | `platform/database/entrypoint_test.go`; `audit_log_migration_test.go`; `platform/database/database_test.go` | PASS |

### Edge cases

| Caso de borda | Resultado definido pela spec | Evidência | Resultado |
|---|---|---|---|
| Dois criam o mesmo e-mail | um sucede, outro 409 | `identity/infra/user_repository_test.go` | PASS |
| **Dois logins simultâneos do mesmo usuário** | duas sessões independentes | `identity/app/authenticate_test.go:414` (`TestConcurrentLoginsForTheSameUserCreateIndependentSessions`, 8 goroutines, 8 tokens distintos, 8 linhas em `sessions`) — **F7, reconfirmado nesta iteração como teste de integração real, não probe** | PASS |
| Cookie malformado | 401 `unauthenticated`, sem erro interno | `platform/httpx/authn_test.go` | PASS |
| **Banco indisponível durante uma requisição** | **503 `service_unavailable`** | `platform/database/unavailable.go` + `unavailable_test.go` (14 casos positivos, 6 negativos) + `unavailable_integration_test.go` (`*pgconn.ConnectError` real, porta fechada); `identity/http/handler.go:180`-`182`; `platform/audit/http.go:55`-`56` + `audit/http_test.go` (derruba a conexão real via testcontainers e confere 503 sem vazar DSN/driver); `platform/httpx/authn.go:143`-`145` + `authn_test.go:138`-`147`; contrato (`identity.yaml`, `platform.yaml`) descreve 503 em toda operação — **F1, reconfirmado com evidência própria; nenhuma das três rotas vaza DSN, driver ou query (checado por `strings.Contains` negativo no teste de auditoria)** | PASS |
| **Papéis de um usuário substituídos por conjunto vazio** | aceito, vínculo dormente, ASSOCIADO mantido, nunca `validation_failed` | `spec.md:584` (corrigida, alinhada ao Assumption "Papel-base ASSOCIADO" e ao contrato); `identity/app/assign_roles.go:58`-`66` (código já seguia o Assumption, confirmado sem alteração nesta iteração — `git show cb876de --stat` só toca `spec.md` e `authenticate_test.go`); `identity/app/assign_roles_test.go:64` (`TestAnEmptyListLeavesTheMembershipDormantWithOnlyAssociado`) | PASS |
| Dois administradores promovem o mesmo usuário | um sucede, outro 409 `already_admin` | `promote_admin_test.go:237` (`TestConcurrentPromotionsOfTheSameUserLeaveExactlyOneWinner`) | PASS |
| Papel removido | deixa de valer na requisição seguinte | `identity/app/session_service_test.go`; `httpapi/e2e_test.go:396` | PASS |
| `Allocate` com total zero | partes zero | `platform/money/allocate_test.go` | PASS |

**Status dos ACs**: ✅ todos os ACs e casos de borda têm evidência e o valor asserido coincide com o resultado que a spec define. 3 spec-precision gaps menores, não normativos, seguem sinalizados (ver "Achados adicionais"): o `code` `login_blocked` do 429 de login não está literalmente escrito na spec (só o status e o header); o bootstrap marcar `must_change_password=true` é decisão de design razoável, não exigida literalmente pela AC de IDN-01.1; a condição de IDN-05.5 ("usuário que não tinha vínculo ativo") é sempre verdadeira na prática, porque `PromoteToAdmin` já recusa com `already_admin` quando há vínculo ativo — nenhum dos três é um comportamento errado, e nenhum foi alterado por esta iteração porque não fazia parte do plano F1-F8 aprovado.

---

## Discrimination Sensor

**Profundidade**: reforçada nos arquivos tocados por esta iteração (prioridade pedida pela tarefa), com alguns mutantes adicionais em caminhos críticos adjacentes (RBAC-03 em `assign_roles.go`). 18 mutações manuais (mínimo pedido: 15).

**Isolamento**: `git status --porcelain` da árvore real vazio antes (`?? .specs/features/fundacao-core/validation.md`, o único arquivo não versionado, esperado) e idêntico depois (`diff` contra o baseline: sem diferença). Um `git worktree` temporário em detached HEAD `8d53101`, fora da árvore real (`...\scratchpad\verifier-wt2`), removido com `git worktree remove --force`; `git worktree list` mostra só a árvore real depois. Nenhum `git stash` foi usado.

**Método**: cada mutante troca um trecho exato de código de produção (validado por `assert` no script de aplicação — falha se o trecho não existir), roda o teste que a spec ou o commit desta iteração associam ao comportamento (`go test -count=1` ou, quando exige banco, `go test -count=1 -tags=integration -run <Teste>`), confirma FAIL, e é revertido com `git checkout -- <arquivo>` antes do próximo. Docker funcionou em todas as tentativas (nenhuma instabilidade a repetir).

| # | Mutação (arquivo) | Descrição | Morto? |
|---|---|---|---|
| M1 | `platform/database/unavailable.go` | remove o reconhecimento de `57P01/57P02/57P03` (admin/crash shutdown) | ✅ `TestUnavailableRecognizesConnectivityFailures` |
| M2 | `platform/database/unavailable.go` | desliga a checagem da classe SQLSTATE `08xxx` | ✅ mesmo teste (casos `08001`, `08006`) |
| M3 | `platform/database/unavailable.go` | desliga o reconhecimento de "pool fechado" (`sql: database is closed`) | ✅ mesmo teste |
| M4 | `platform/database/unavailable.go` | desliga o reconhecimento de `net.Error` | ✅ mesmo teste |
| M5 | `platform/httpx/requestid.go` | regex do request id passa a aceitar `_` e `.` | ✅ `TestRequestIDGeneratesWhenMissingOrInvalid` (F5) |
| M6 | `identity/app/roles_matrix.go` | concede `financeiro:prestacao_contas:read` ao ADMIN_SISTEMA | ✅ `TestAdminSistemaHasExactlyTheTechnicalPermissionsAndNoInstitutionalOne` (F2 — reproduz o antigo M58) |
| M7 | `identity/http/handler.go` | quebra o regex `emailInErrorMessage` (redação de e-mail no log) | ✅ `TestLogInternalRedactsAnEmailInTheErrorMessage` (F8) |
| M8 | `identity/http/users_handler.go` | remove `.UTC()` de `CreatedAt` em `userOut` | ✅ `TestUserOutNormalizesTimestampsToUTC` (F4) |
| M9 | `platform/audit/query.go` | remove `.UTC()` de `OccurredAt` em `row.record()` | ✅ `TestRowRecordNormalizesOccurredAtToUTC` (F4) |
| M10 | `platform/httpx/authn.go` | desliga o `case database.Unavailable(err)` (503) | ✅ `TestAuthnAnswersServiceUnavailableWhenTheDatabaseIsDown` (F1) |
| M11 | `identity/http/handler.go` | desliga o `case database.Unavailable(err)` em `writeError` | ✅ `TestEveryDomainErrorMapsToItsStatusAndCode` (F1) |
| M12 | `identity/http/handler.go` | remove `.UTC()` de `GrantedAt` em `authContext` | ✅ `TestMeShowsTheMembershipSummaryOnlyWithReasonAndDateOrNull` (integração, F4) |
| M13 | `platform/audit/http.go` | desliga o `case database.Unavailable(err)` (503) | ✅ `TestAuditQueryAnswersServiceUnavailableWhenTheDatabaseIsDown` (integração, F1) |
| M14 | `identity/app/promote_admin.go` | `Before.roles` passa a usar `after` em vez de `before` (`admin.promote`) | ✅ `TestPromoteGrantsMembershipAndRolesForcesPasswordChangeRevokesSessionsAndAudits` (integração, F3 — reproduz o antigo M37) |
| M15 | `identity/app/deactivate_user.go` | `Before`/`After` do payload passam a usar o mesmo valor (`active`) | ✅ `TestDeactivateRevokesSessionsAndPendingRecoveryTokensAndAudits` (integração, F5 — reproduz o antigo M39) |
| M16 | `identity/app/assign_roles.go` | desliga a checagem de auto-alteração de papéis | ✅ `TestNobodyChangesTheirOwnRoles` (integração, RBAC-03.3) |
| M17 | `platform/database/unavailable.go` | `err == nil` passa a devolver `true` | ✅ `TestUnavailableRejectsBusinessAndNilErrors` |
| M18 | `identity/app/roles_matrix.go` | concede `identity:user:create` a DIRETORIA | ✅ `TestFoundationProvisionalMatrix` |

**Sobreviventes e equivalência**: nenhum. Os 4 sobreviventes da iteração 1 (M37, M39, M58, M69) foram reproduzidos ponto a ponto nesta rodada (M14, M15, M6, M5) e morrem todos, confirmando que as correções F1 a F5 realmente fecharam as lacunas de discriminação encontradas antes — não apenas por leitura de código, mas por mutação executada.

**Sensor**: 18 injeções, 18 mortas, 0 sobreviventes. **Result**: 18/18 — PASS ✅.

---

## Interactive UAT Results

Não aplicável: a fundação entrega API e infraestrutura, sem tela. A checagem de ponta a ponta é `httpapi/e2e_test.go` (fluxo completo: bootstrap, login, criar, promover, trocar senha obrigatória, recuperar por e-mail, redefinir por administrador, atribuir papel, retirar acesso, desativar, consultar auditoria — `TestE2EFullIdentityFlow`, linha 250).

---

## Code Quality

| Princípio | Status |
|---|---|
| Sem funcionalidade além do pedido | ✅ as 5 correções desta iteração tocam só o que F1-F8 pediram; nenhum endpoint ou regra nova fora do plano aprovado |
| Sem abstração de uso único | ✅ `database.Unavailable` é uma função pura reaproveitada nos 3 pontos, sem abstração extra; `assertPromoteBeforeAfter`/`assertActivationBeforeAfter` são helpers de teste enxutos e reaproveitados |
| Só arquivos exigidos | ✅ os 28 arquivos do diff (`a9eb6ba^..HEAD`) batem com o que F1-F8 descrevem: código de produção só em `unavailable.go`, `handler.go`, `audit/http.go`, `query.go`, `authn.go`, `roles_matrix` (não tocado), `identity.yaml`/`platform.yaml`, `api.gen.go` (gerado); o resto são testes e documentação |
| Segue os padrões existentes | ✅ `database.Unavailable` segue o padrão de erro sentinela + `errors.Is`/`errors.AsType` já usado no resto do código; nenhuma camada nova |
| Testes mapeiam para ACs, não são rasos | ✅ conferido por leitura direta de `unavailable_test.go`, `promote_admin_test.go`, `deactivate_user_test.go`, `roles_matrix_test.go`, `query_test.go`, `handler_internal_test.go`, `authn_test.go`, `requestid_test.go`, `http_test.go` (audit) nesta sessão — todas as asserções são por valor exato ou decodificação estrutural, não `strings.Contains` frouxo |
| Cobertura por camada | ✅ F1 cobre unitário (classificador) e integração (falha real de conexão via `pgconn.Connect` e via `testcontainers` + `sqlDB.Close()`) |
| Todo teste mapeia para AC/borda | ✅ por amostragem, todos os testes novos citam o AC ou a lacuna que corrigem em comentário |
| Diretrizes documentadas seguidas | ✅ `CLAUDE.md`, `docs/CONTRIBUTING.md`, `domain-boundaries.md`: sem `AutoMigrate`, sem JWT, auditoria na mesma transação, contrato primeiro (contrato atualizado no mesmo commit do código) |

---

## Edge Cases

- [x] Dois criam o mesmo e-mail: um sucede, outro 409 (`identity/infra/user_repository_test.go`)
- [x] Dois logins simultâneos: sessões independentes, com teste real de concorrência (`authenticate_test.go:414`) — **antes só probe, agora teste no repositório (F7)**
- [x] Cookie malformado: 401 sem erro interno (`platform/httpx/authn_test.go`)
- [x] Banco indisponível: 503 `service_unavailable` implementado e testado nos 3 pontos (`unavailable.go` + 3 arquivos de teste) — **antes GAP, agora fechado (F1)**
- [x] Papéis vazios: aceito, vínculo dormente, sem `validation_failed` — **antes conflito interno da spec, agora resolvido só na spec (F6)**
- [x] Dois administradores promovem o mesmo usuário: um 409 `already_admin` (`promote_admin_test.go:237`)
- [x] Papel removido deixa de valer na requisição seguinte (`session_service_test.go`)
- [x] `Allocate` com total zero (`allocate_test.go`)

---

## Gate Check

Executado na árvore real; nada destrutivo, sem push, Docker só pelos testes de integração.

| Comando | Saída | Resultado |
|---|---|---|
| `cd api && go vet ./...` | 0 | sem achados |
| `cd api && go vet -tags=integration ./...` | 0 | sem achados |
| `cd api && go build ./...` | 0 | ok |
| `cd api && go test -p 2 -count=1 -tags=integration ./...` | 0 | 18 pacotes `ok`, 4 sem arquivo de teste, 0 falha, nenhuma repetição por instabilidade de Docker |
| `cd api && go test -count=1 ./...` (sem a tag, TST-01.1) | 0 | todos `ok` |
| `cd api && go test -cover ./internal/platform/money/...` | 0 | 100,0% das instruções (exigido: 95%) |
| `cd api && go generate ./...` | 0 | sem diff (`git status --porcelain` inalterado) — API-01.2 |
| `cd web && pnpm lint` | 0 | ok |
| `cd web && pnpm lint:api` | 0 | Redocly: `platform` e `identity` válidos |
| `cd web && pnpm gen:api:check` | 0 | tipos em dia (aviso cosmético do `$ref` em `example`, já conhecido da iteração 1, inalterado) |
| `cd web && pnpm exec next typegen` | 0 | ok |
| `cd web && pnpm exec tsc --noEmit` | 0 | ok |
| `cd web && pnpm test` | 0 | 2 arquivos, 54 testes, 0 falha |
| `cd web && pnpm build` | 0 | Next 16.3.6, compilação e tipos ok |
| `python .claude/skills/tlc-spec-driven/scripts/validate_spec.py fundacao-core` | 0 | 0 erros, 0 avisos |
| `python .claude/skills/tlc-spec-driven/scripts/validate_tasks.py fundacao-core` | 0 | 0 erros, 18 avisos ("Tests: none", esperados para tarefas de contrato/CI/docs — mesmo resultado da iteração 1) |

- **Testes**: todos os pacotes Go com testes saem `ok`; Vitest 54/54. Nenhum `t.Skip`/`it.skip` encontrado nos arquivos tocados por esta iteração.
- **Delta desta iteração**: +959/-157 linhas em 28 arquivos; testes novos ou reforçados em `unavailable_test.go` (novo), `unavailable_integration_test.go` (novo), `roles_matrix_test.go`, `promote_admin_test.go`, `deactivate_user_test.go`, `authenticate_test.go`, `handler_internal_test.go`, `auth_handler_test.go`, `audit/http_test.go`, `audit/query_test.go`, `httpx/authn_test.go`, `httpx/requestid_test.go`; nenhum teste removido ou enfraquecido.

---

## Requirement Traceability Update

`spec.md:595`-`619` já está `Done`/`Implemented` para as 25 linhas de rastreabilidade, reconfirmado nesta iteração (conferido diretamente, sem alteração necessária).

| Requisito | Status |
|---|---|
| IDN-01 a IDN-08, EML-01, RBAC-01 a RBAC-03, AUD-01 a AUD-04, MNY-01 a MNY-03, TST-01 a TST-03, API-01, API-02, PLT-01 | ✅ Verified |

---

## Summary

**Overall**: ✅ Ready

**Spec-anchored check**: 25/25 requisitos com cobertura completa; nenhuma lacuna; 3 spec-precision gaps menores sinalizados (não normativos, não bloqueiam)
**Sensor**: 18/18 mutantes mortos, incluindo os 4 antigos sobreviventes reproduzidos e agora mortos
**Gate**: todos os comandos com saída 0; Go e Vitest sem falhas; cobertura de `money` 100,0%

**What works**: tudo o que a iteração 1 já certificava, mais: 503 `service_unavailable` quando o banco cai (nos 3 pontos que traduzem erro para HTTP, sem vazar DSN/driver), matriz do ADMIN_SISTEMA e do PRESIDENTE com conjunto exato garantido por teste, payload de auditoria de promoção e de (des)ativação verificado por decodificação estrutural, timestamps de resposta sempre em UTC independente do fuso do processo, regex do request id restrito ao alfabeto correto, spec sem contradição interna no caso de papéis vazios, login concorrente com sessões independentes provado por teste real, documentação de acompanhamento (`STATE.md`, `design.md`, `tasks.md`, rastreabilidade) alinhada ao código e à branch real, e log interno que redige e-mail antes de gravar.

**Issues found**: nenhuma que bloqueie o veredito. Itens cosméticos remanescentes, sem prioridade (ver "Achados adicionais").

**Next steps**: nenhuma ação corretiva pendente para `fundacao-core`. Recomenda-se seguir o `STATE.md`: tratar a issue SEC-001 (#19) e iniciar a spec de `fundacao-documentos`.

---

## Achados adicionais

### Segurança (por leitura do código e execução dos testes)

- **503 sem vazamento**: os três pontos que respondem 503 usam uma mensagem estática (`"Serviço temporariamente indisponível..."`), nunca `err.Error()` do driver; `audit/http_test.go` confirma isso ativamente (busca por `tj_app_dev`, `connection`, `sql:`, `gorm`, `pgconn` no corpo da resposta, nenhum presente) e `handler_internal_test.go` confirma o mesmo para o caminho de `identity/http` com uma DSN fabricada contendo `segredo-de-conexao`.
- **Redação de e-mail no log interno (F8)**: `logInternal` (`identity/http/handler.go:222`-`227`) aplica `emailInErrorMessage.ReplaceAllString` antes de logar; `TestLogInternalRedactsAnEmailInTheErrorMessage` reproduz um erro real de violação de unicidade do PostgreSQL contendo um e-mail e confirma que o log final não contém o e-mail, só a marca `[e-mail redigido]`. Reproduzi o cenário lendo o teste linha a linha (não só a existência dele): a asserção nega a presença literal do e-mail e exige a presença da marca de redação — não é um teste vazio.
- **Nenhum segredo novo em log/auditoria/resposta**: as mensagens estáticas de 503, a senha temporária, os tokens de recuperação e os hashes continuam fora de log e auditoria, sem mudança nesta iteração (`noSecretsLeaked` em `e2e_test.go` continua cobrindo o fluxo completo).

### Itens cosméticos remanescentes (não bloqueiam)

- `api/openapi/identity.yaml:146` ainda tem um `$ref` dentro de um `example` que `pnpm gen:api:check` não resolve (saída 0, só aviso); já conhecido da iteração 1, sem prioridade.
- Os 3 spec-precision gaps menores (código `login_blocked` não literal na spec; `must_change_password` no bootstrap não exigido literalmente pela AC; condição de IDN-05.5 sempre verdadeira na prática) permanecem como observações, não como defeitos: não faziam parte do plano F1-F8 aprovado pelo mantenedor e nenhum deles corresponde a um comportamento incorreto.
- `STATE.md:122` registra "F8 ... em andamento neste commit" — uma frase que ficou congelada no próprio commit que concluiu F8 (o commit que a escreveu é o mesmo que fecha o item). É uma redação um pouco confusa lida depois do fato, mas não é uma informação errada nem afeta nenhum gate; não é motivo de FAIL.

### O que ficou fora do escopo (evolução futura; não é falso FAIL)

Confirmado em `spec.md` (Out of Scope) e `STATE.md` (Future decisions), sem mudança nesta iteração:

- Telas de login, painel e gestão de usuários (feature seguinte de UI).
- Provedor concreto de e-mail (só a interface, `log` e `disabled`); impedir `EMAIL_PROVIDER=disabled` em produção.
- MFA, OAuth, provedor externo de identidade; rate limit unificado de credenciais por IP/dispositivo; rate limiting global da API.
- Entidades do financeiro; mapeamento de `ErrOutOfRange` para 422 no primeiro endpoint monetário real.
- Documentos e object storage (`fundacao-documentos`).
- Hospedagem e deploy reais; domínio final do cookie e proxy do Next.
- Purga/retenção de auditoria e sessões (LGPD); armazenamento controlado de IP em eventos de segurança; separação entre auditoria institucional e técnica; versionamento e encadeamento criptográfico dos eventos.
- Fechamento de ano e ajuste extraordinário; dupla aprovação para promoção administrativa.
- Vínculo usuário × associado (spec de `associados`); convite por e-mail e onboarding; troca de e-mail (`ChangeEmail`).
- `GET /roles`; componente `Id` no `common.yaml`; `openapi-fetch` só na primeira tela real.
- Consulta externa de senhas comprometidas (k-anonymity).
- **Item novo registrado no `STATE.md` nesta rodada**: uniformizar o tempo de resposta da recuperação de senha entre conta existente e inexistente (hash falso ou custo artificial) — hoje há uma pequena diferença porque o token só é criado no banco quando a conta existe; registrado como decisão futura, não como lacuna desta feature (a spec já exige resposta idêntica e envio assíncrono, o que está implementado e testado; a diferença residual é de tempo de banco, não de conteúdo da resposta).
