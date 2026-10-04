# Identity — Administração de Usuários (Web) — Specification

## Problem Statement

A API cria usuários, promove a administrativo com papéis, ajusta papéis, retira o acesso, desativa, reativa e gera senha temporária. Sem tela para isso, nenhuma pessoa recebe `TESOURARIA` ou `ESTOQUE_LOJA` sem chamar a API diretamente, e os módulos `financeiro` e `estoque` ficam inutilizáveis pela Web. WEB-D-004 inclui essa administração na Web V1 como feature própria de `identity`.

## Goals

- [ ] Quem tem as permissões de `identity` administra pessoas e papéis pela Web, com motivo onde a API exige e confirmação nas ações sensíveis.

## Out of Scope

| Feature | Reason |
| --- | --- |
| Editar nome ou e-mail de usuário | Não existe endpoint. |
| Histórico de vínculos e auditoria | `GET /audit-logs` fora da V1 (STATE, "Fora do V1"). |
| Exclusão de usuário | Não existe; desativação é o caminho (AD-007). |
| Regras de escalada de privilégio no front | A API decide (`privilege_escalation`, `self_change_forbidden`, `last_admin`); o front mostra o resultado. |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Onde ficam as ações | Na própria listagem, por menu de ações da linha, abrindo diálogos; sem página de detalhe | O contrato não tem `GET /users/{id}`; a linha já traz todos os campos de `User` | y |
| Papéis oferecidos na promoção e na edição | Os sete papéis administrativos do enum `Role` (todos menos `ASSOCIADO`); a API recusa o que o ator não pode conceder | `ASSOCIADO` é sempre mantido pela API (descrição de `SetRolesRequest`); a regra de escalada é da API | y |
| Edição de papéis com lista vazia | Permitida, com aviso de que o vínculo fica dormente | Descrição de `SetRolesRequest` | y |
| Paginação | Botão "Carregar mais" com `next_cursor`, filtros `active` e `role` enviados à API | Contrato de `listUsers` (WEB-D-016) | y |
| Senha temporária | Mostrada uma única vez, com botão de copiar; descartada da memória ao fechar o diálogo | Descrição de `TemporaryPassword` e `Cache-Control: no-store` | y |

**Open questions:** none.

---

## User Stories

### USR-01: Listar usuários

**User Story**: Como administrador, quero ver as pessoas cadastradas, filtrar por situação e papel e carregar mais páginas.

**Acceptance Criteria**:

1. WHEN uma pessoa com `identity:user:read` abre `/admin/usuarios` THEN o sistema SHALL chamar `GET /api/v1/users` e mostrar nome, e-mail, situação (ativo ou inativo), papéis, se tem vínculo administrativo e se a troca de senha está pendente.
2. WHEN a pessoa escolhe um filtro de situação ou de papel THEN o sistema SHALL refazer a consulta com `active` e `role` e recomeçar do início.
3. WHILE a resposta traz `next_cursor` não nulo THE sistema SHALL oferecer "Carregar mais", que acrescenta a próxima página à lista.
4. IF a pessoa não tem `identity:user:read` THEN a rota SHALL mostrar "Sem acesso" e o item não SHALL aparecer na navegação.

**Independent Test**: com MSW, duas páginas de usuários, filtro por papel e "Carregar mais".

---

### USR-02: Criar usuário

**User Story**: Como administrador, quero cadastrar uma pessoa com uma senha inicial.

**Acceptance Criteria**:

1. WHEN uma pessoa com `identity:user:create` envia nome, e-mail e senha inicial THEN o sistema SHALL chamar `POST /api/v1/users` e, em `201`, mostrar a pessoa na lista, informando que ela terá de trocar a senha no primeiro acesso.
2. IF a API responde `409 email_taken` THEN o sistema SHALL mostrar o erro no campo de e-mail.
3. IF a API responde `invalid_name`, `invalid_email`, `password_too_short`, `password_too_long` ou `password_compromised` THEN o sistema SHALL mostrar a mensagem junto do campo correspondente.
4. IF a pessoa não tem `identity:user:create` THEN a ação "Novo usuário" não SHALL aparecer.

**Independent Test**: com MSW, criar com sucesso e com `email_taken`.

---

### USR-03: Desativar e reativar

**User Story**: Como administrador, quero desativar quem não deve mais entrar e reativar quando necessário.

**Acceptance Criteria**:

1. WHEN uma pessoa com `identity:user:update` confirma a desativação de um usuário ativo THEN o sistema SHALL chamar `POST /api/v1/users/{id}/deactivate` e, em `204`, mostrar o usuário como inativo.
2. WHEN uma pessoa com `identity:user:update` confirma a reativação de um usuário inativo THEN o sistema SHALL chamar `POST /api/v1/users/{id}/reactivate`, mostrar o usuário como ativo e avisar que papéis administrativos e vínculo não são restaurados.
3. The desativação SHALL exigir confirmação explícita, informando que as sessões da pessoa serão encerradas.
4. IF a API responde `self_change_forbidden`, `last_admin` ou `privilege_escalation` THEN o sistema SHALL mostrar a mensagem correspondente e manter o estado exibido.

**Independent Test**: com MSW, desativar, reativar e receber `last_admin`.

---

### USR-04: Acesso administrativo e papéis

**User Story**: Como administrador, quero promover uma pessoa a administrativa com papéis, ajustar os papéis e retirar o acesso, sempre com motivo onde a API exige.

**Acceptance Criteria**:

1. WHEN uma pessoa com `identity:admin:grant` escolhe ao menos um papel e informa um motivo para um usuário sem vínculo THEN o sistema SHALL chamar `POST /api/v1/users/{id}/admin-membership` e, em `204`, avisar que as sessões do usuário foram encerradas e que ele terá de trocar a senha.
2. IF nenhum papel foi escolhido ou o motivo tem menos de 10 caracteres sem contar espaços nas pontas THEN o sistema SHALL manter o envio desabilitado na promoção.
3. WHEN uma pessoa com `identity:role:assign` salva a lista de papéis de um usuário com vínculo THEN o sistema SHALL chamar `PUT /api/v1/users/{id}/roles` e, se a lista estiver vazia, SHALL antes avisar que o vínculo ficará dormente.
4. WHEN uma pessoa com `identity:admin:revoke` confirma a retirada de acesso com motivo THEN o sistema SHALL chamar `POST /api/v1/users/{id}/admin-membership/revoke` e mostrar o usuário só com `ASSOCIADO`.
5. IF a API responde `already_admin`, `not_admin`, `admin_membership_required`, `privilege_escalation`, `self_change_forbidden`, `last_admin`, `user_inactive`, `unknown_role`, `admin_role_required` ou `reason_required` THEN o sistema SHALL mostrar a mensagem em português correspondente.
6. The sistema SHALL oferecer só a ação compatível com o estado do usuário: promover quando não há vínculo; editar papéis e retirar acesso quando há.

**Independent Test**: com MSW, promover, editar papéis (inclusive vazio), retirar acesso e receber `privilege_escalation`.

---

### USR-05: Senha temporária

**User Story**: Como administrador, quero gerar uma senha temporária para quem perdeu o acesso e não consegue usar o e-mail.

**Acceptance Criteria**:

1. WHEN uma pessoa com `identity:user:reset_password` confirma a redefinição com motivo de ao menos 10 caracteres THEN o sistema SHALL chamar `POST /api/v1/users/{id}/password-reset` e mostrar a senha temporária uma única vez, com opção de copiar.
2. WHEN o diálogo da senha temporária é fechado THEN o sistema SHALL descartar a senha da memória da tela e não SHALL guardá-la no cache de consultas, em armazenamento do navegador ou em log.
3. IF a API responde `reason_required`, `self_change_forbidden`, `privilege_escalation` ou `user_inactive` THEN o sistema SHALL mostrar a mensagem correspondente.

**Independent Test**: com MSW, gerar, conferir a exibição, fechar e conferir que a senha não está mais no DOM nem no cache.

---

## Edge Cases

- WHEN a linha é a da própria pessoa autenticada THEN o sistema SHALL esconder desativar, retirar acesso, editar papéis e senha temporária (a API recusaria com `self_change_forbidden`).
- WHEN uma ação muda um usuário THEN o sistema SHALL invalidar as consultas de usuários, para que os filtros reflitam a mudança.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| USR-01 | USR-01: Listar usuários | Tasks | In Tasks |
| USR-02 | USR-02: Criar usuário | Tasks | In Tasks |
| USR-03 | USR-03: Desativar e reativar | Tasks | In Tasks |
| USR-04 | USR-04: Acesso administrativo e papéis | Tasks | In Tasks |
| USR-05 | USR-05: Senha temporária | Tasks | In Tasks |

**Coverage:** 5 total, 5 mapped to tasks, 0 unmapped.

---

## Success Criteria

- [ ] Um `ADMIN_SISTEMA` cria uma pessoa, promove a `TESOURARIA`, e essa pessoa entra, troca a senha e vê o Financeiro.
- [ ] Uma pessoa com só `identity:user:read` (por exemplo `DIRETORIA`) vê a lista sem nenhuma ação de escrita.
