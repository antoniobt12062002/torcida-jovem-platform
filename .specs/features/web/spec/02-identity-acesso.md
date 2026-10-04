# Identity — Acesso e Senha (Web) — Specification

## Problem Statement

A API já troca senha (`POST /auth/password`), pede recuperação por e-mail (`POST /auth/password-reset/request`) e confirma a recuperação com token (`POST /auth/password-reset/confirm`). O e-mail de recuperação aponta para `${APP_BASE_URL}/redefinir-senha#token=…`, uma rota que não existe no front. E toda conta criada ou promovida nasce com troca de senha obrigatória: sem a tela de troca, ninguém passa do primeiro acesso.

## Goals

- [ ] A pessoa troca a própria senha, inclusive no primeiro acesso, e recupera o acesso pelo link do e-mail, sem chamar a API diretamente.

## Out of Scope

| Feature | Reason |
| --- | --- |
| Administração de usuários e senha temporária | Sub-spec 03. |
| Medidor de força de senha próprio | A API decide (tamanho mínimo e senha comprometida); o front só mostra o retorno. |
| Login social, MFA | Fora da ADR-005. |

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
| --- | --- | --- | --- |
| Tamanho mínimo exibido como dica | "Mínimo de 10 caracteres" para quem tem vínculo administrativo (`admin_membership` não nulo no contexto) e "mínimo de 8" para os demais; a API continua decidindo | ADR-005 (nota de 2026-09-26) e `password_too_short` da API | y |
| Onde o token de recuperação é lido | Do fragmento da URL (`#token=`), que o navegador não envia ao servidor; removido da barra de endereço logo após a leitura | Formato do link já emitido pela API; o fragmento não aparece em logs de servidor nem em `Referer` | y |
| Depois da troca de senha | Rebuscar `/auth/me` e seguir para o destino; a sessão atual continua (a API encerra só as outras) | Descrição de `204` em `changePassword` | y |
| Depois da redefinição por token | Levar a `/entrar` com aviso de sucesso; a API encerra todas as sessões | Descrição de `204` em `confirmPasswordReset` | y |

**Open questions:** none.

---

## User Stories

### ACS-01: Trocar a própria senha

**User Story**: Como pessoa autenticada, quero trocar a minha senha, e no primeiro acesso quero ser levada a fazer isso antes de usar o sistema.

**Acceptance Criteria**:

1. WHEN a pessoa envia senha atual, nova senha e confirmação iguais em `/conta/senha` THEN o sistema SHALL chamar `POST /api/v1/auth/password` com `current_password` e `new_password` e, em `204`, mostrar confirmação de sucesso.
2. IF a nova senha e a confirmação diferem THEN o sistema SHALL mostrar o erro no campo de confirmação e não SHALL chamar a API.
3. WHEN a troca tem sucesso e a sessão tinha `must_change_password = true` THEN o sistema SHALL rebuscar `GET /api/v1/auth/me` e levar a pessoa ao destino original ou à página inicial.
4. WHILE `must_change_password = true` THE página `/conta/senha` SHALL explicar que a troca é obrigatória e não SHALL oferecer navegação para outras áreas, mantendo a opção de sair.
5. IF a API responde `invalid_current_password`, `password_unchanged`, `password_too_short`, `password_too_long` ou `password_compromised` THEN o sistema SHALL mostrar a mensagem em português correspondente junto do campo certo.
6. IF a API responde `429 password_change_blocked` THEN o sistema SHALL informar quanto tempo esperar, a partir de `Retry-After`.

**Independent Test**: com MSW, sessão com troca obrigatória → troca com sucesso → contexto rebuscado → página inicial; e cada código de erro no campo certo.

---

### ACS-02: Pedir recuperação de acesso

**User Story**: Como pessoa que esqueceu a senha, quero pedir um link de recuperação por e-mail.

**Acceptance Criteria**:

1. WHEN a pessoa envia um e-mail em `/recuperar-acesso` THEN o sistema SHALL chamar `POST /api/v1/auth/password-reset/request` e, em `202`, mostrar a mensagem neutra devolvida pela API.
2. The página SHALL mostrar a mesma mensagem e o mesmo comportamento para qualquer e-mail, sem indicar se existe conta.
3. The página `/entrar` SHALL oferecer o link "Esqueci minha senha" para `/recuperar-acesso` (link criado por F2, FND-03 AC10; conferido no smoke da INT).

**Independent Test**: com MSW, enviar dois e-mails diferentes e conferir a mesma mensagem.

---

### ACS-03: Redefinir a senha pelo link do e-mail

**User Story**: Como pessoa que recebeu o link, quero definir uma nova senha.

**Acceptance Criteria**:

1. WHEN a página `/redefinir-senha` abre com `#token=…` THEN o sistema SHALL ler o token do fragmento, guardá-lo só em memória e remover o fragmento da barra de endereço sem recarregar a página.
2. IF a página abre sem token THEN o sistema SHALL informar que o link é inválido e oferecer `/recuperar-acesso`.
3. WHEN a pessoa envia nova senha e confirmação iguais THEN o sistema SHALL chamar `POST /api/v1/auth/password-reset/confirm` com `token` e `new_password` e, em `204`, levar a pessoa a `/entrar` com aviso de senha redefinida.
4. IF a API responde `400 invalid_reset_token` THEN o sistema SHALL informar que o link expirou ou já foi usado e oferecer `/recuperar-acesso`.
5. IF a API responde `password_too_short`, `password_too_long` ou `password_compromised` THEN o sistema SHALL mostrar a mensagem junto do campo de nova senha, mantendo o token para nova tentativa.
6. The token SHALL nunca ser enviado em query string, gravado em armazenamento do navegador ou escrito em log.

**Independent Test**: com MSW, abrir com token, confirmar que o fragmento sumiu, redefinir com sucesso; repetir com `invalid_reset_token`.

---

## Edge Cases

- WHEN uma pessoa autenticada abre `/redefinir-senha` THEN o sistema SHALL permitir a redefinição normalmente (a API encerra todas as sessões) e depois levar a `/entrar`.
- WHEN `/conta/senha` é aberta sem troca obrigatória THEN a página SHALL funcionar como troca voluntária, dentro do shell.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
| --- | --- | --- | --- |
| ACS-01 | ACS-01: Trocar a própria senha | Tasks | In Tasks |
| ACS-02 | ACS-02: Pedir recuperação | Tasks | In Tasks |
| ACS-03 | ACS-03: Redefinir pelo link | Tasks | In Tasks |

**Coverage:** 3 total, 3 mapped to tasks, 0 unmapped.

---

## Success Criteria

- [ ] Uma conta recém-criada entra, troca a senha e chega à página inicial.
- [ ] O link do e-mail de recuperação funciona de ponta a ponta na stack local.
