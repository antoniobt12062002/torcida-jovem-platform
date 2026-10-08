# Web V1 — Relatório de verificação independente

## Validation: Web V1 - FAIL

**Data**: 2026-10-08
**Specs**: `.specs/features/web/spec/01-fundacao.md` a `06-integracao.md`
**Intervalo de commits verificado**: `ee1b497..2c50de6` (código da Web V1; `ee1b497` é a especificação; `develop` = `1a508d1`)
**Verificador**: agente independente (autor ≠ verificador), task T4 de `tasks/06-integracao.md`
**Fontes de verdade usadas**: as 6 specs, designs, tasks, `STATE.md` (WEB-D-*), ADR-010. Os relatórios dos autores não foram usados como prova; `evidence/f1-rewrite.md` e `smoke.md` serviram só de pista e foram refeitos (seção "Smoke repetido").

### Por que FAIL

1. **O gate Audit está vermelho** (`pnpm audit --audit-level high` → `EXIT=1`): GHSA-68fv-2mgg-jv7q (`source-map-js` 1.2.1, alta). O aviso não vem do diff da Web V1 (o lockfile de `develop` tem a mesma versão e falha igual hoje), mas o gate é obrigatório em T4 e faz parte da definição de "pronta para uso" (`STATE.md`, item 7); o job `web` do CI falharia no mesmo passo.
2. **Um mutante sobreviveu** (O06): a URL assinada do comprovante pode ser gravada no cache de consultas sem nenhum teste falhar. A cláusula "sem guardá-la no cache de consultas" de FWB-05 AC5 não tem asserção discriminante.

Todo o resto passou: os 10 mutantes mínimos de INT-04 AC3 foram mortos (16 variantes), o smoke repetido contra a stack real não teve falha e os demais gates estão verdes. As correções são pequenas (ver "Lacunas").

---

## Task Completion

| Sub-spec | Tasks | Status | Observação |
| --- | --- | --- | --- |
| 01 Fundação | T1-T8 | ✅ Concluídas | Todas as caixas marcadas em `tasks/01-fundacao.md` |
| 02 Identity acesso | T1-T2 | ✅ Concluídas | — |
| 03 Identity usuários | T1-T3 | ✅ Concluídas | — |
| 04 Financeiro | T1-T5 | ✅ Concluídas | — |
| 05 Estoque | T1-T3 | ✅ Concluídas | — |
| 06 Integração | T1-T3 | ✅ Concluídas | — |
| 06 Integração | T4 | ❌ Não concluída | Veredito FAIL; caixas de T4 não marcadas (`.specs/features/web/tasks/06-integracao.md:117`) |

---

## Spec-Anchored Acceptance Criteria

Caminhos relativos a `web/`, salvo indicação. "T:" é o teste; "C:" é o código de produção. Resultado: ✅ PASS, ❌ GAP, ⚠️ gap de precisão da spec.

### 01 — Fundação

| Critério | Resultado definido na spec | Evidência (`arquivo:linha` + asserção) | Resultado |
| --- | --- | --- | --- |
| FND-01 AC1 | `/api/v1/*` → mesmo caminho em `API_URL`; preserva método, corpo, `Cookie`, `Origin`, `Content-Type`, `X-CSRF-Token`; devolve status, corpo, `Set-Cookie`, `Content-Type`, `Retry-After`, `Cache-Control` | T: `next.config.test.ts:13` `toEqual([{source:"/api/v1/:path*", destination:"http://localhost:8080/api/v1/:path*"}])`; C: `next.config.ts:15`. Cabeçalhos: smoke repetido (itens S1-S4, S20-S23) | ✅ PASS (cabeçalhos só pelo smoke: ⚠️ SP-5) |
| FND-01 AC2 | `openapi-fetch`, `credentials: "same-origin"`; nenhuma tela chama `fetch` direto | T: `lib/api/client.test.ts:49` `expect(seen[0].credentials).toBe("same-origin")` e URL na origem; C: `lib/api/client.ts:112`. "Nenhum `fetch` direto": busca no código de produção só achou `lib/api/client.ts:116` | ✅ PASS (a segunda cláusula é conferida por inspeção, sem guarda automática) |
| FND-01 AC3 | `X-CSRF-Token` em todo método que não seja GET/HEAD/OPTIONS | T: `lib/api/client.test.ts:72` (POST), `:80` (PUT), `:89` (PATCH), `:98` (DELETE) `expect(seen[0].csrf).toBe(TOKEN)`; `:110` GET `toBeNull()`; C: `lib/api/client.ts:66` | ✅ PASS (mutante M01 morto) |
| FND-01 AC4 | Erro tipado com `status`, `code`, `errors[]` | T: `lib/api/client.test.ts:208` `toMatchObject({status:422, code:"validation_failed", errors:[{field:"nome",code:"required"}]})`; `lib/api/problem.test.ts:24`; C: `lib/api/problem.ts:59` | ✅ PASS |
| FND-01 AC5 (aceite manual) | (a) cookie `tj_session` na origem do front; (b) escrita passa pelo `Origin`; (c) comprovante ≥ 9,5 MiB íntegro | Smoke repetido: S2 (`Set-Cookie ... HttpOnly; SameSite=Lax`, sem `Domain`), S5 (`201`; origem externa `403 origin_not_allowed`), S9 (10.420.224 bytes → `201`, SHA-256 do download igual) | ✅ PASS |
| FND-01 AC6 | Parar e reportar se AC5 falhar | Não acionado (AC5 passou). O achado do limite do rewrite foi reportado e resolvido pelo mantenedor: `.specs/features/web/STATE.md:152` | ✅ PASS (condicional, sem teste possível) |
| FND-02 AC1 | Componentes listados em `components/ui/` | C: `components/ui/input.tsx`, `label.tsx`, `field.tsx`, `select.tsx`, `textarea.tsx`, `checkbox.tsx`, `dialog.tsx`, `alert-dialog.tsx`, `table.tsx`, `dropdown-menu.tsx`, `skeleton.tsx`, `alert.tsx`, `toast.tsx`; T: `components/ui/display.test.tsx:19`, `:44`, `:58`, `:67`; `components/ui/overlays.test.tsx:41`, `:63`, `:114` | ✅ PASS |
| FND-02 AC2 | Erro associado por `aria-describedby` e campo com `aria-invalid` | T: `components/ui/field.test.tsx:39` `expect(input.getAttribute("aria-invalid")).toBe("true")` e `describedByText(input)` contém a mensagem; `:82`, `:98`, `:115`; C: `components/ui/field.tsx:213` | ✅ PASS (mutante O03 morto) |
| FND-02 AC3 | F4 não altera `package.json` | `git log ee1b497..2c50de6 -- web/package.json web/pnpm-lock.yaml` → só `57cfbad` (F1) | ✅ PASS (processo) |
| FND-03 AC1 | `POST /auth/login`, guarda contexto, vai a `next` interno ou à inicial | T: `app/(public)/entrar/login-form.test.tsx:48` `navigations()` = `["/financeiro/contas?tipo=RECEITA"]` e corpo `{email, password}`; `:57` `["/inicio"]`; `lib/session/routes.test.ts:19`; C: `app/(public)/entrar/login-form.tsx:62`, `lib/session/routes.ts:43` | ✅ PASS (mutante M04 morto) |
| FND-03 AC2 | "E-mail ou senha inválidos." sem indicar o campo | T: `login-form.test.tsx:90` `findByText("E-mail ou senha inválidos.")` e `aria-invalid` ausente nos dois campos; C: `login-form.tsx:27` | ✅ PASS |
| FND-03 AC3 | Bloqueio temporário e tempo de espera a partir de `Retry-After` | T: `login-form.test.tsx:104` `Retry-After: 900` → texto contém "15 minutos"; C: `login-form.tsx:29` | ✅ PASS (formato do tempo não definido: ⚠️ SP-2) |
| FND-03 AC4 | `must_change_password` → `/conta/senha`, nenhuma outra tela | T: `login-form.test.tsx:79` `["/conta/senha"]`; `app/(app)/layout.test.tsx:64` conteúdo protegido ausente; C: `app/(app)/layout.tsx:23` | ✅ PASS (mutante O08 morto) |
| FND-03 AC5 | Logout: `POST /auth/logout`, limpa cache e contexto, vai a `/entrar` | T: `lib/session/session-provider.test.tsx:196` `navigations()` = `["/entrar"]`, `getQueryData(...)` `toBeUndefined()`, nenhuma consulta com dado; C: `lib/session/session-provider.tsx:181` | ✅ PASS (mutante O15 morto) |
| FND-03 AC6 | `401` → limpa cache e contexto, `/entrar` com aviso e `next` | T: `session-provider.test.tsx:259` `next` = `/financeiro/contas?tipo=RECEITA`, `sessao` = `encerrada`; `login-form.test.tsx:165` aviso exibido; C: `session-provider.tsx:141` | ✅ PASS (mutante M02 morto) |
| FND-03 AC7 | `403 password_change_required` → `/conta/senha` | T: `session-provider.test.tsx:355` `["/conta/senha"]`; C: `session-provider.tsx:149` | ✅ PASS |
| FND-03 AC8 | `403 csrf_invalid` → rebusca `/auth/me`, pede para repetir, sem repetir a escrita | T: `session-provider.test.tsx:372` `me.count` = 2, mensagem "A sessão foi atualizada. Repita a ação, por favor.", `writes` = 1; C: `session-provider.tsx:152` | ✅ PASS (mutante O16 morto) |
| FND-03 AC9 | Token e contexto só em memória | T: `session-provider.test.tsx:159` `localStorage.length` = 0, `sessionStorage.length` = 0, `document.cookie` inalterado; C: `session-provider.tsx:175` | ✅ PASS (mutante O14 morto) |
| FND-03 AC10 | Link "Esqueci minha senha" → `/recuperar-acesso` | T: `login-form.test.tsx:158` `href` = `/recuperar-acesso`; C: `login-form.tsx:128` | ✅ PASS |
| FND-04 AC1 | Sem sessão → `/entrar` com `next`, sem conteúdo | T: `app/(app)/layout.test.tsx:42`; C: `app/(app)/layout.tsx:27` | ✅ PASS |
| FND-04 AC2 | Carregando, sem conteúdo protegido | T: `layout.test.tsx:27` `getByRole("status")` contém "Carregando" e conteúdo ausente; C: `layout.tsx:49` | ✅ PASS |
| FND-04 AC3 | Exibição por `permissions`, nunca pelo papel | T: `session-provider.test.tsx:78` papel `PRESIDENTE` sem permissão → `false`; `components/app/nav.test.ts:41`; `components/app/require-permission.test.tsx:102`; C: `session-provider.tsx:227`, `components/app/require-permission.tsx:26` | ✅ PASS (mutantes M03a e M03b mortos) |
| FND-04 AC4 | Sem a permissão de leitura: "Sem acesso", sem chamar a API | T: `require-permission.test.tsx:60` `calls.count` = 0; por área: `app/(app)/financeiro/contas/page.test.tsx:40`, `app/(app)/financeiro/saldo/page.test.tsx:39`, `features/financeiro/lancamentos/lancamentos-page.test.tsx:275`, `app/(app)/admin/usuarios/page.test.tsx:30`, `features/estoque/produtos/produtos-page.test.tsx:191` | ✅ PASS |
| FND-04 AC5 | `403 forbidden` → "Você não tem permissão para esta ação.", estado inalterado | T: `components/app/api-error.test.tsx:14`; `features/financeiro/contas/contas-page.test.tsx:228` (linha não criada) e `:341` (conta continua "Ativo"); C: `components/app/api-error.tsx:9` | ✅ PASS |
| FND-05 AC1 | Shell: nome, "Trocar senha", "Sair", navegação filtrada por permissão | T: `components/app/app-shell.test.tsx:55`, `:66`, `:102`, `:110`; C: `components/app/app-shell.tsx:38` | ✅ PASS |
| FND-05 AC2 | Abaixo do ponto de quebra de tablet a navegação vira menu recolhível | T: `app-shell.test.tsx:87` menu "Abrir menu" com os mesmos itens e fecha ao navegar; C: `app-shell.tsx:45` (`md:hidden`) e `:84` (`hidden md:block`). A troca pelo ponto de quebra não é testável em jsdom nem foi vista em navegador | ⚠️ SP-1 (evidência parcial) |
| FND-05 AC3 | Estados de carregamento, vazio e erro com tentar de novo | T: `components/app/states.test.tsx:12`, `:19`, `:32` | ✅ PASS |
| FND-05 AC4 | pt-BR → centavos inteiros por `parseBRL`, nunca float | T: `components/app/money-field.test.tsx:35` `expect(last).toBe(cents)` e `Number.isSafeInteger(last)`; `:112` `{valor: 123456}`; C: `components/app/money-field.tsx:17` | ✅ PASS (mutante M05 morto) |
| FND-05 AC5 | Só inteiros; sinal configurável | T: `components/app/quantity-field.test.tsx:40`, `:51`, `:58`, `:67`, `:78`; C: `components/app/quantity-field.tsx:19` | ✅ PASS |
| FND-05 AC6 | Confirmação explícita; com motivo obrigatório, confirmar desabilitado se vazio ou só espaços | T: `components/app/confirm-dialog.test.tsx:16` `onConfirm` não chamado antes de confirmar; `:74` e `:81` `disabled` = `true`; C: `components/app/confirm-dialog.tsx:80` | ✅ PASS (mutantes M06a e M07a mortos) |
| FND-05 AC7 | `formatBRL`; datas pt-BR no fuso de São Paulo | T: `components/app/format.test.ts:8`, `:12` (`"03/10/2026"`), `:23` (`"04/10/2026 09:00"`); C: `components/app/format.ts:4` | ✅ PASS |
| FND-05 AC8 | Com troca obrigatória o shell esconde a navegação e mostra só sair | T: `app-shell.test.tsx:128` sem `navigation`, sem links, menu = `["Sair"]`; C: `app-shell.tsx:37` | ✅ PASS (mutante O10 morto) |
| FND-06 AC1 | `422 validation_failed`: cada erro no campo | T: `lib/forms/apply-problem.test.ts:21`; `login-form.test.tsx:134` `aria-invalid` = `true` e "Formato inválido."; `features/estoque/produtos/produtos-page.test.tsx:156`; C: `lib/forms/apply-problem.ts:39` | ✅ PASS |
| FND-06 AC2 | `code` conhecido → mensagem do catálogo | T: `lib/api/problem.test.ts:63`; `api-error.test.tsx:19`; C: `lib/api/problem.ts:85` | ✅ PASS |
| FND-06 AC3 | `code` desconhecido → mensagem genérica com o status, sem detalhes internos | T: `lib/api/problem.test.ts:70` contém "409", não contém o `code`, `title` nem `detail`; C: `problem.ts:89` | ✅ PASS (mutante O17 morto) |
| FND-06 AC4 | `413` → conteúdo grande demais | T: `lib/api/problem.test.ts:87`; `api-error.test.tsx:34` `"O conteúdo enviado é grande demais."`; C: `problem.ts:87` | ✅ PASS |
| FND-06 AC5 | `503` ou rede → indisponível, com tentar de novo | T: `lib/api/problem.test.ts:93`, `:100`; `api-error.test.tsx:46`; `layout.test.tsx:83`; C: `problem.ts:76` | ✅ PASS |
| Borda: dois `401` | Um único redirecionamento | T: `session-provider.test.tsx:281` `navigations()` com 1 item; C: `session-provider.tsx:110` | ✅ PASS |
| Borda: `next` externo | Ignorado, vai à inicial | T: `lib/session/routes.test.ts:29` (14 casos → `"/inicio"`); `login-form.test.tsx:65` | ✅ PASS |
| Borda: autenticado em `/entrar` | Vai à inicial | T: `login-form.test.tsx:177` `["/inicio"]`; C: `login-form.tsx:53` | ✅ PASS |

### 02 — Identity: acesso e senha

| Critério | Resultado definido na spec | Evidência | Resultado |
| --- | --- | --- | --- |
| ACS-01 AC1 | `POST /auth/password` com `current_password` e `new_password`; em `204`, confirmação | T: `features/identity/acesso/change-password-form.test.tsx:69` corpo e CSRF exatos, "Senha alterada com sucesso."; C: `features/identity/acesso/change-password-form.tsx:59` | ✅ PASS |
| ACS-01 AC2 | Confirmação diferente: erro no campo de confirmação, sem chamar a API | T: `change-password-form.test.tsx:89` `requests` = `[]`, `aria-invalid` na confirmação; C: `change-password-form.tsx:151` | ✅ PASS |
| ACS-01 AC3 | Troca obrigatória: rebusca `/auth/me` e vai ao destino ou à inicial | T: `change-password-form.test.tsx:107` `["/inicio"]`, `state.calls` = 2; `:120` `next`; C: `features/identity/acesso/use-change-password.ts:38` | ✅ PASS (mutante O13 morto) |
| ACS-01 AC4 | Explica a obrigatoriedade, sem navegação, mantém sair | T: `change-password-form.test.tsx:139`; `app-shell.test.tsx:128`; C: `change-password-form.tsx:94` | ✅ PASS |
| ACS-01 AC5 | Cinco códigos de senha, mensagem no campo certo | T: `change-password-form.test.tsx:184` (5 casos, campo marcado e os outros não); C: `features/identity/acesso/errors.ts:31` | ✅ PASS |
| ACS-01 AC6 | `429 password_change_blocked` com tempo de `Retry-After` | T: `change-password-form.test.tsx:224` "Muitas tentativas. Tente de novo em 15 minutos."; C: `errors.ts:63` | ✅ PASS (⚠️ SP-2) |
| ACS-02 AC1 | `POST /auth/password-reset/request`; em `202`, mensagem da API | T: `features/identity/acesso/request-reset-form.test.tsx:47` corpo `{email}`; `:64` mostra o texto devolvido; C: `features/identity/acesso/request-reset-form.tsx:58` | ✅ PASS |
| ACS-02 AC2 | Mesma mensagem e comportamento para qualquer e-mail | T: `request-reset-form.test.tsx:54` `expect(second).toBe(first)`; smoke S21 (duas respostas `202` com a mesma mensagem) | ✅ PASS |
| ACS-02 AC3 | Link em `/entrar` | T: `login-form.test.tsx:158` | ✅ PASS |
| ACS-03 AC1 | Lê o token do fragmento, guarda em memória, remove o fragmento sem recarregar | T: `features/identity/acesso/reset-password-form.test.tsx:78` `window.location.hash` = `""`, `navigations()` = `[]`; `features/identity/acesso/read-reset-token.test.ts:7`; C: `features/identity/acesso/reset-password-form.tsx:50` | ✅ PASS (mutante M10 morto) |
| ACS-03 AC2 | Sem token: link inválido e oferta de `/recuperar-acesso` | T: `reset-password-form.test.tsx:91` (3 casos); C: `reset-password-form.tsx:98` | ✅ PASS |
| ACS-03 AC3 | `POST .../confirm` com `token` e `new_password`; em `204`, `/entrar` com aviso | T: `reset-password-form.test.tsx:103` corpo exato, `["/entrar"]`, "Senha redefinida. Entre com a nova senha."; C: `reset-password-form.tsx:73` | ✅ PASS |
| ACS-03 AC4 | `400 invalid_reset_token`: link expirou ou já usado, oferta de novo pedido | T: `reset-password-form.test.tsx:151` "O link expirou ou já foi usado."; smoke S22 (`400 invalid_reset_token`); C: `reset-password-form.tsx:76` | ✅ PASS |
| ACS-03 AC5 | Erros de senha no campo, token mantido | T: `reset-password-form.test.tsx:168` segunda tentativa envia o mesmo `token`; C: `reset-password-form.tsx:80` | ✅ PASS |
| ACS-03 AC6 | Token nunca em query string, armazenamento ou log | T: `reset-password-form.test.tsx:65` (`expectTokenNotLeaked`), usado em `:85`, `:118`, `:162`, `:197`; `:116` `search` = `""` | ✅ PASS |
| Borda: autenticado em `/redefinir-senha` | Redefine e vai a `/entrar` | T: `reset-password-form.test.tsx:121` sessão vira `null` | ✅ PASS |
| Borda: troca voluntária | Funciona dentro do shell | T: `change-password-form.test.tsx:156`; `app/(app)/conta/senha/page.test.tsx:17` | ✅ PASS |

### 03 — Identity: usuários

| Critério | Resultado definido na spec | Evidência | Resultado |
| --- | --- | --- | --- |
| USR-01 AC1 | `GET /users`; nome, e-mail, situação, papéis, vínculo, troca pendente | T: `features/identity/usuarios/users-page.test.tsx:52` células exatas e `requests` = `[{limit:"50"}]`; C: `features/identity/usuarios/users-table.tsx:57` | ✅ PASS |
| USR-01 AC2 | Filtros refazem a consulta com `active` e `role`, do início | T: `users-page.test.tsx:98`, `:115`, `:138` (sem `cursor` depois do filtro); C: `features/identity/usuarios/hooks.ts:23` | ✅ PASS (mutante O20 morto) |
| USR-01 AC3 | "Carregar mais" com `next_cursor` | T: `users-page.test.tsx:163` `[{limit:"50"},{limit:"50",cursor:"pagina-2"}]`; `:180`; C: `hooks.ts:41` | ✅ PASS (mutante O04 morto) |
| USR-01 AC4 | Sem `identity:user:read`: "Sem acesso" e item fora da navegação | T: `app/(app)/admin/usuarios/page.test.tsx:30` `requests` = `[]`; `features/identity/nav.test.ts:18` | ✅ PASS |
| USR-02 AC1 | `POST /users`; em `201`, pessoa na lista e aviso da troca | T: `features/identity/usuarios/dialogs/create-user-dialog.test.tsx:68` corpo e CSRF exatos; C: `features/identity/usuarios/dialogs/create-user-dialog.tsx:68` | ✅ PASS |
| USR-02 AC2 | `409 email_taken` no campo de e-mail | T: `create-user-dialog.test.tsx:98` (1º caso) | ✅ PASS |
| USR-02 AC3 | Cinco códigos no campo correspondente | T: `create-user-dialog.test.tsx:98` (demais casos); C: `create-user-dialog.tsx:32` | ✅ PASS |
| USR-02 AC4 | Sem `identity:user:create`, sem "Novo usuário" | T: `create-user-dialog.test.tsx:58`; C: `features/identity/usuarios/users-page.tsx:87` | ✅ PASS |
| USR-03 AC1 | Desativar: `POST .../deactivate`; em `204`, inativo | T: `features/identity/usuarios/dialogs/status-dialogs.test.tsx:73` `[{body:null, csrf}]`, situação "Inativo" | ✅ PASS |
| USR-03 AC2 | Reativar; aviso de que papéis e vínculo não voltam | T: `status-dialogs.test.tsx:98`; C: `features/identity/usuarios/dialogs/status-dialogs.tsx:15` | ✅ PASS |
| USR-03 AC3 | Confirmação explícita informando o fim das sessões | T: `status-dialogs.test.tsx:83` texto do diálogo e `requests` = `[]` antes de confirmar; C: `status-dialogs.tsx:38` | ✅ PASS |
| USR-03 AC4 | `self_change_forbidden`, `last_admin`, `privilege_escalation`: mensagem e estado mantido | T: `status-dialogs.test.tsx:122` (3 casos), `:139` | ✅ PASS |
| USR-04 AC1 | Promover: `POST .../admin-membership`; aviso de sessões e troca de senha | T: `features/identity/usuarios/dialogs/admin-dialogs.test.tsx:129` corpo `{roles:["DIRETORIA","TESOURARIA"], reason}`; C: `features/identity/usuarios/dialogs/admin-dialogs.tsx:106` | ✅ PASS |
| USR-04 AC2 | Envio desabilitado sem papel ou com motivo < 10 caracteres úteis | T: `admin-dialogs.test.tsx:108` (9 caracteres → desabilitado; 10 → habilitado); C: `features/identity/usuarios/dialogs/reason.ts:9` | ✅ PASS (mutante O12 morto) |
| USR-04 AC3 | `PUT .../roles`; lista vazia avisa antes | T: `admin-dialogs.test.tsx:210`, `:230` (`requests` = `[]` até "Salvar sem papéis"); C: `admin-dialogs.tsx:161` | ✅ PASS (mutante O11 morto) |
| USR-04 AC4 | Retirar acesso com motivo; só `ASSOCIADO` | T: `admin-dialogs.test.tsx:248` corpo `{reason:"Deixou a tesouraria"}`; C: `admin-dialogs.tsx:227` | ✅ PASS (mínimo do motivo não definido: ⚠️ SP-3) |
| USR-04 AC5 | Dez códigos com mensagem em português | T: `admin-dialogs.test.tsx:159` (10 casos); C: `features/identity/usuarios/errors.ts:15` | ✅ PASS |
| USR-04 AC6 | Só a ação compatível com o estado | T: `admin-dialogs.test.tsx:74`, `:85`; C: `features/identity/usuarios/actions.ts:33` | ✅ PASS |
| USR-05 AC1 | Motivo ≥ 10; `POST .../password-reset`; senha uma vez, com copiar | T: `features/identity/usuarios/dialogs/temporary-password-dialog.test.tsx:74`, `:92`; C: `features/identity/usuarios/dialogs/temporary-password-dialog.tsx:63` | ✅ PASS |
| USR-05 AC2 | Ao fechar, some da tela; fora do cache, do armazenamento e do log | T: `temporary-password-dialog.test.tsx:126` DOM; `:135` consultas; `:136` mutações; `:137` armazenamento; `:138` console | ✅ PASS (mutantes M09a e M09b mortos) |
| USR-05 AC3 | Quatro códigos com mensagem | T: `temporary-password-dialog.test.tsx:151` | ✅ PASS |
| Borda: própria linha | Esconde as ações | T: `status-dialogs.test.tsx:54`; C: `actions.ts:29` | ✅ PASS (mutante O07 morto) |
| Borda: invalidação | Ações invalidam as consultas de usuários | T: lista muda depois da escrita em `status-dialogs.test.tsx:90`, `admin-dialogs.test.tsx:147`, `:224`, `:267`; C: `hooks.ts:51` | ✅ PASS |

### 04 — Financeiro

| Critério | Resultado definido na spec | Evidência | Resultado |
| --- | --- | --- | --- |
| FWB-01 AC1 | `GET /contas`, árvore por `parent_id`, tipo e situação | T: `features/financeiro/contas/contas-page.test.tsx:97`; `features/financeiro/contas/tree.test.ts`; C: `features/financeiro/contas/contas-page.tsx:87` | ✅ PASS |
| FWB-01 AC2 | `POST /contas` com `tipo`, `nome`, `parent_id` | T: `contas-page.test.tsx:171` `{tipo:"DESPESA", nome:"Patrocínios"}`; `:188` com `parent_id`; C: `features/financeiro/contas/conta-dialogs.tsx:72` | ✅ PASS (mutante O19 morto) |
| FWB-01 AC3 | Subconta fixa o tipo do pai | T: `contas-page.test.tsx:188` rádios `disabled`, o do pai marcado | ✅ PASS |
| FWB-01 AC4 | `PATCH`; `409 conta_ja_utilizada` explicado | T: `contas-page.test.tsx:264`, `:283` | ✅ PASS |
| FWB-01 AC5 | Desativar com confirmação; conta inativa | T: `contas-page.test.tsx:318` `requests` = `[]` ao cancelar | ✅ PASS (mutante M07b morto) |
| FWB-01 AC6 | `conta_nao_encontrada`, `conta_tipo_incompativel` | T: `contas-page.test.tsx:228`, `:283`, `:341`; C: `features/financeiro/contas/errors.ts:7` | ✅ PASS |
| FWB-02 AC1 | `GET /saldo`, `formatBRL`, zero e negativo | T: `features/financeiro/saldo/saldo-page.test.tsx:40` (`"R$ 1.234,56"`, `"R$ 0,00"`, `"-R$ 987,65"`); C: `features/financeiro/saldo/saldo-page.tsx:34` | ✅ PASS |
| FWB-02 AC2 | Explicação do cálculo | T: `saldo-page.test.tsx:54`; C: `saldo-page.tsx:39` | ✅ PASS |
| FWB-03 AC1 | `GET /lancamentos`, campos e filtros no cliente | T: `features/financeiro/lancamentos/lancamentos-page.test.tsx:84`, `:117` (`list.count` = 1), `:127`; `features/financeiro/lancamentos/filters.test.ts:22`; C: `features/financeiro/lancamentos/filters.ts:31` | ✅ PASS |
| FWB-03 AC2 | `POST /lancamentos` em centavos inteiros | T: `lancamentos-page.test.tsx:164` `valor_bruto_cents: 15000`, `taxa_cents: 250`; smoke S6 (a API recusa decimal com `422`) | ✅ PASS |
| FWB-03 AC3 | `PUT /lancamentos/{id}` | T: `features/financeiro/lancamentos/lancamento-detail.test.tsx:117` corpo exato | ✅ PASS |
| FWB-03 AC4 | Editar só em `CRIADA`; `lancamento_imutavel` explicado e lista atualizada | T: `lancamento-detail.test.tsx:145`, `:151` (`list.count` = 2); `features/financeiro/lancamentos/actions.test.ts:34` | ✅ PASS |
| FWB-03 AC5 | Três códigos junto do campo | T: `lancamentos-page.test.tsx:244`; C: `features/financeiro/lancamentos/errors.ts:32` | ✅ PASS |
| FWB-03 AC6 | Detalhe com todos os campos, cancelamento e devolução | T: `lancamento-detail.test.tsx:61`, `:77`, `:93` | ✅ PASS |
| FWB-04 AC1 | Receber: `POST .../receive`, `RECEBIDA`, saldo atualizado | T: `features/financeiro/lancamentos/workflow.test.tsx:71` `saldoInvalidated` = `true`; C: `features/financeiro/lancamentos/hooks.ts:79` | ✅ PASS (mutantes M07c e O02 mortos) |
| FWB-04 AC2 | Pagar: `POST .../pay`, `PAGA`, saldo atualizado | T: `workflow.test.tsx:102` | ✅ PASS |
| FWB-04 AC3 | Cancelar com `reason`; `CANCELADA` com o motivo | T: `workflow.test.tsx:151` corpo `{reason:"Lançado em duplicidade"}` | ✅ PASS |
| FWB-04 AC4 | Motivo vazio ou só espaços: desabilitado, sem API | T: `workflow.test.tsx:137` `disabled` = `true`, `requests` = `[]` | ✅ PASS (mutante M06a morto) |
| FWB-04 AC5 | Devolução com `devolucao_de_id`, conta de despesa, valores | T: `workflow.test.tsx:208` corpo exato | ✅ PASS (mutante O05a morto) |
| FWB-04 AC6 | Devolução só em receita `RECEBIDA`; `devolucao_invalida` | T: `workflow.test.tsx:238`, `:258`; `actions.test.ts:29` | ✅ PASS |
| FWB-04 AC7 | Receber, pagar e cancelar só nos estados certos; três códigos atualizam a lista | T: `workflow.test.tsx:86`, `:116`, `:178`, `:258`; `actions.test.ts:34` | ✅ PASS |
| FWB-05 AC1 | Lista de comprovantes, do mais novo para o mais antigo | T: `features/financeiro/comprovantes/comprovantes-panel.test.tsx:104` | ✅ PASS |
| FWB-05 AC2 | `multipart/form-data`, campo `file`, "enviando", lista atualizada | T: `comprovantes-panel.test.tsx:122` `contentType` casa `^multipart/form-data; boundary=`; C: `features/financeiro/comprovantes/hooks.ts:39` | ✅ PASS |
| FWB-05 AC3 | > 10.420.224 bytes recusa sem requisição; até 10.420.224 envia | T: `comprovantes-panel.test.tsx:159` constante; `:173` envia; `:182` `requests` = `[]`; C: `features/financeiro/comprovantes/limits.ts:6`, `:21`; smoke S9 | ✅ PASS (mutantes O01a e O01b mortos) |
| FWB-05 AC4 | `413` e três `422` com mensagem | T: `comprovantes-panel.test.tsx:203`; C: `features/financeiro/comprovantes/errors.ts:15` | ✅ PASS |
| FWB-05 AC5 | URL assinada buscada no clique e aberta, **sem guardar no cache de consultas** | T: `comprovantes-panel.test.tsx:235` (`calls.count` = 2, `window.open` com a URL). Nenhuma asserção inspeciona o `QueryClient`; C: `features/financeiro/comprovantes/hooks.ts:52` | ❌ GAP parcial (mutante O06 sobreviveu) |
| FWB-05 AC6 | `document_not_found`, `lancamento_nao_encontrado` | T: `comprovantes-panel.test.tsx:209`, `:256` | ✅ PASS |
| Borda: só leitura | Nenhuma ação de escrita | T: `contas-page.test.tsx:373`; `lancamentos-page.test.tsx:269`; `workflow.test.tsx:269`; `comprovantes-panel.test.tsx:284`; smoke S15-S17 | ✅ PASS |
| Borda: id fora da lista | "Lançamento não encontrado" | T: `lancamento-detail.test.tsx:103` | ✅ PASS |
| Borda: mudança de status | Invalida a lista e o saldo | T: `workflow.test.tsx:82`, `:112`, `:174` | ✅ PASS |

### 05 — Estoque

| Critério | Resultado definido na spec | Evidência | Resultado |
| --- | --- | --- | --- |
| EWB-01 AC1 | `GET /produtos`; código, nome, unidade; busca no cliente | T: `features/estoque/produtos/produtos-page.test.tsx:83`, `:101` (`calls.count` = 1) | ✅ PASS |
| EWB-01 AC2 | `POST /produtos` e produto na lista | T: `produtos-page.test.tsx:121` corpo exato | ✅ PASS |
| EWB-01 AC3 | `409 codigo_duplicado` no campo; `422` em cada campo | T: `produtos-page.test.tsx:138`, `:156`; C: `features/estoque/produtos/create-produto-dialog.tsx:62` | ✅ PASS |
| EWB-01 AC4 | Saldo com unidade; negativo destacado | T: `features/estoque/produto-detail.test.tsx:47`, `:55` (`"-3 UN"`, "Saldo negativo"); C: `features/estoque/produto-detail.tsx:106` | ✅ PASS |
| EWB-02 AC1 | Histórico com as colunas, na ordem da API | T: `features/estoque/movimentacoes/movimentacoes-panel.test.tsx:77` | ✅ PASS |
| EWB-02 AC2 | Entrada e saída com `origem = INVENTARIO` | T: `movimentacoes-panel.test.tsx:161` corpos exatos; `features/estoque/movimentacoes/payload.test.ts:14`; C: `features/estoque/movimentacoes/payload.ts:21` | ✅ PASS (mutante M08a morto) |
| EWB-02 AC3 | Devolução com `movimentacao_de_id` | T: `movimentacoes-panel.test.tsx:170`; `payload.test.ts:32` | ✅ PASS |
| EWB-02 AC4 | Nunca oferecer `VENDA`, `COMPRA`, `EVENTO`, `AJUSTE_MANUAL` | T: `movimentacoes-panel.test.tsx:131` sem `combobox`, `radio`, `listbox` nem os rótulos; C: `features/estoque/movimentacoes/movimentacao-dialog.tsx:128` | ✅ PASS (mutante M08b morto) |
| EWB-02 AC5 | Quatro códigos com mensagem | T: `movimentacoes-panel.test.tsx:187`, `:200`; C: `features/estoque/errors.ts:14` | ✅ PASS |
| EWB-02 AC6 | Invalida histórico e saldo | T: `movimentacoes-panel.test.tsx:182` `movCalls` = 4, `saldoCalls` = 4; C: `features/estoque/movimentacoes/hooks.ts:30` | ✅ PASS (mutante O09 morto) |
| EWB-03 AC1 | `POST /ajustes` com `produto_id`, `quantidade`, `motivo` | T: `features/estoque/ajustes/ajuste-dialog.test.tsx:52` corpo `{quantidade:-3, motivo:"Contagem física do inventário"}` | ✅ PASS |
| EWB-03 AC2 | Motivo vazio ou quantidade zero: desabilitado, sem API | T: `ajuste-dialog.test.tsx:94` (5 casos); C: `features/estoque/ajustes/ajuste-dialog.tsx:43` | ✅ PASS (mutante M06b morto) |
| EWB-03 AC3 | Saldo atual e resultante; aviso de negativo, sem impedir | T: `ajuste-dialog.test.tsx:58` (`"0 UN"`, `"-3 UN"`, aviso, botão habilitado) | ✅ PASS (forma da confirmação: ⚠️ SP-4) |
| EWB-03 AC4 | `motivo_obrigatorio`, `quantidade_invalida` | T: `ajuste-dialog.test.tsx:114` | ✅ PASS |
| EWB-04 AC1 | Só leitura: vê tudo, sem escrita | T: `produtos-page.test.tsx:185`; `movimentacoes-panel.test.tsx:231`; `ajuste-dialog.test.tsx:131`; smoke S15-S17 | ✅ PASS |
| EWB-04 AC2 | Sem a permissão da seção: omite e não chama a API | T: `produto-detail.test.tsx:74` (`saldoState.calls` = 0); `movimentacoes-panel.test.tsx:102` (`movCalls` = 0) | ✅ PASS (mutante O18 morto) |
| Borda: id fora da lista | "Produto não encontrado" | T: `produto-detail.test.tsx:37` | ✅ PASS |
| Borda: `AJUSTE` e `DEVOLUCAO` | Sem "Devolver" | T: `movimentacoes-panel.test.tsx:112` `[true,true,false,false]`; `payload.test.ts:54` | ✅ PASS |

### 06 — Integração (INT-01 a INT-03)

| Critério | Resultado definido na spec | Evidência | Resultado |
| --- | --- | --- | --- |
| INT-01 AC1 | `nav-items.ts` junta os itens dos módulos mais "Início", na ordem, sem redefinir | T: `components/app/nav-items.test.ts:14`, `:18`, `:22` (`toBe` da própria seção), `:28`; C: `components/app/nav-items.ts:13` | ✅ PASS |
| INT-01 AC2 | Shell usa `nav-items.ts` e filtra por permissão | T: `app/(app)/inicio/page.test.tsx:122` (5 perfis); C: `app/(app)/layout.tsx:34` | ✅ PASS |
| INT-01 AC3 | `/` → `/inicio` | T: `app/page.test.tsx:19` `redirect` chamado com `"/inicio"`; smoke S1 (`307`, `Location: /inicio`) | ✅ PASS |
| INT-01 AC4 | Atalho por área visível; sem área, mensagem | T: `app/(app)/inicio/page.test.tsx:127`, `:147`; C: `app/(app)/inicio/atalhos.tsx:17` | ✅ PASS |
| INT-01 AC5 | Raiz não consulta `/healthz` | T: `app/page.test.tsx:26` `fetchSpy` não chamado; C: `app/page.tsx:7` | ✅ PASS |
| INT-02 AC1 (manual) | Guia explica variáveis, ordem de subida, rewrite e gates | `docs/development/web.md:19` (variáveis da API), `:39` (`API_URL`), `:43` (ordem), `:155` (rewrite), `:168` (gates). A stack do smoke repetido foi montada por este guia | ✅ PASS |
| INT-02 AC2 (manual) | Comportamento de `API_URL` no build | `docs/development/web.md:39` | ✅ PASS |
| INT-02 AC3 (manual) | Sem segredo nem dado pessoal | Leitura integral: só marcadores `<...>`, credenciais de desenvolvimento do compose e `presidente@local.test` | ✅ PASS |
| INT-03 AC1 (manual) | Cada item com resultado, data e observação | `.specs/features/web/smoke.md:34` (tabela, 12 itens); data única em `smoke.md:3` | ✅ PASS (⚠️ SP-6: data por execução, não por item) |
| INT-03 AC2 (manual) | Cobertura dos itens listados | `smoke.md:36` a `:48`. Repetição independente: seção "Smoke repetido" | ✅ PASS (a interação na tela não foi vista por ninguém; ver "Pendências") |
| INT-03 AC3 (manual) | Declara que o token válido não é executável localmente | `smoke.md:12`, `:40`; confirmado no smoke (log só com `recipient_domain` e `subject`) | ✅ PASS |
| INT-03 AC4 (manual) | Falha vira correção e repetição | `smoke.md:50` (nenhum item falhou) | ✅ PASS (condicional) |
| INT-03 AC5 (manual) | Sem senha, token, cookie ou dado pessoal | Leitura integral de `smoke.md` | ✅ PASS |
| Borda: módulo sem item visível | Seção não aparece | T: `components/app/nav.test.ts:26`; `app/(app)/inicio/page.test.tsx:111` (perfis sem a seção) | ✅ PASS |

**Contagem**: 129 ACs e 13 casos de borda (142 critérios). Com evidência discriminante: 140. Lacuna: 1 (FWB-05 AC5, cláusula do cache). Evidência parcial por imprecisão da spec: 1 (FND-05 AC2).

**Status**: ❌ Há lacuna.

---

## Discrimination Sensor

Worktree temporário `C:\Users\Michels\Desktop\tjw\vsensor` (`git worktree add ... 2c50de6 --detach`), uma falha por vez no código de produção, `pnpm test` completo (`API_URL=http://localhost:8080`), reversão com `git checkout -- .` antes da seguinte. Linha de base no worktree temporário: 66 arquivos, 688 testes, `EXIT=0`.

### Mínimo obrigatório (INT-04 AC3)

| # | Falha injetada | Arquivo | Resultado | Teste que matou (primeiro) |
| --- | --- | --- | --- | --- |
| M01 | (1) Sem `X-CSRF-Token` nas escritas | `lib/api/client.ts:69` | ✅ Morto (30 falhas) | `lib/api/client.test.ts:72` "X-CSRF-Token vai em POST com o token da sessão atual" |
| M02 | (2) Não redireciona em `401` | `lib/session/session-provider.tsx:145` | ✅ Morto (3) | `lib/session/session-provider.test.tsx:259` "limpa cache e contexto e leva a /entrar com aviso e o caminho atual em next" |
| M03a | (3) `RequirePermission` sempre renderiza | `components/app/require-permission.tsx:26` | ✅ Morto (24) | `components/app/require-permission.test.tsx:60` "sem a permissão, mostra Sem acesso e não chama a API da área" |
| M03b | (3) `lancamentoActions` ignora a permissão | `features/financeiro/lancamentos/actions.ts:50` | ✅ Morto (18) | `features/financeiro/lancamentos/actions.test.ts:45` "editar sem a permissão → false" |
| M04 | (4) `safeNext` aceita `next` externo | `lib/session/routes.ts:45` | ✅ Morto (19) | `lib/session/routes.test.ts:29` "recusa outra origem sem esquema e usa a página inicial" |
| M05 | (5) `MoneyField` entrega reais com decimais | `components/app/money-field.tsx:17` | ✅ Morto (13) | `components/app/money-field.test.tsx:35` "1.234,56 vira 123456 centavos inteiros" |
| M06a | (6) Cancelamento com motivo vazio habilitado e enviado | `components/app/confirm-dialog.tsx:80`, `features/financeiro/lancamentos/lancamento-detail.tsx:208` | ✅ Morto (5) | `features/financeiro/lancamentos/workflow.test.tsx:137` "motivo vazio mantém a confirmação desabilitada e não chama a API" |
| M06b | (6) Ajuste com motivo vazio habilitado | `features/estoque/ajustes/ajuste-dialog.tsx:43` | ✅ Morto (2) | `features/estoque/ajustes/ajuste-dialog.test.tsx:94` "motivo vazio: desabilitada e a API não é chamada" |
| M07a | (7) `ConfirmDialog` confirma ao abrir | `components/app/confirm-dialog.tsx:103` | ✅ Morto (10) | `components/app/confirm-dialog.test.tsx:16` "só chama onConfirm ao confirmar; cancelar não chama" |
| M07b | (7) Desativar conta sem diálogo | `features/financeiro/contas/contas-page.tsx:49` | ✅ Morto (3) | `features/financeiro/contas/contas-page.test.tsx:318` "só chama a API depois de confirmar e mostra a conta como inativa" |
| M07c | (7) Receber e pagar chamam a API no clique | `features/financeiro/lancamentos/lancamento-detail.tsx:175` | ✅ Morto (4) | `features/financeiro/lancamentos/workflow.test.tsx:71` "confirma, chama /receive com CSRF, mostra Recebida e invalida o saldo" |
| M08a | (8) Corpo com origem reservada `VENDA` | `features/estoque/movimentacoes/payload.ts:21` | ✅ Morto (5) | `features/estoque/movimentacoes/payload.test.ts:14` "entrada: tipo, produto, quantidade e origem INVENTARIO, nada mais" |
| M08b | (8) Diálogo oferece seletor de origens | `features/estoque/movimentacoes/movimentacao-dialog.tsx:128` | ✅ Morto (1) | `features/estoque/movimentacoes/movimentacoes-panel.test.tsx:124` "entrada de 10, saída de 4, devolução da saída; corpo sempre com INVENTARIO" |
| M09a | (9) Senha temporária em `setQueryData` | `features/identity/usuarios/dialogs/temporary-password-dialog.tsx:63` | ✅ Morto (1) | `features/identity/usuarios/dialogs/temporary-password-dialog.test.tsx:92` "ao fechar some do DOM, do cache, do armazenamento e do log" |
| M09b | (9) Senha temporária por `useMutation` (cache de mutações) | `features/identity/usuarios/dialogs/temporary-password-dialog.tsx:63` | ✅ Morto (1) | o mesmo teste (`:136`) |
| M10 | (10) Token de recuperação fica na URL | `features/identity/acesso/reset-password-form.tsx:52` | ✅ Morto (6) | `features/identity/acesso/reset-password-form.test.tsx:78` "remove o fragmento da barra de endereço sem recarregar e mostra o formulário" |

### Mutantes próprios

| # | Falha injetada | Arquivo | Resultado | Teste que matou (primeiro) |
| --- | --- | --- | --- | --- |
| O01a | Limite de upload aceita 10.420.225 bytes | `features/financeiro/comprovantes/limits.ts:22` | ✅ Morto (1) | `features/financeiro/comprovantes/comprovantes-panel.test.tsx:182` |
| O01b | Limite recusa exatamente 10.420.224 bytes | `features/financeiro/comprovantes/limits.ts:22` | ✅ Morto (1) | `comprovantes-panel.test.tsx:173` |
| O02 | Mudança de status não invalida o saldo | `features/financeiro/lancamentos/hooks.ts:84` | ✅ Morto (3) | `features/financeiro/lancamentos/workflow.test.tsx:71` |
| O03 | Campo numérico nunca marca `aria-invalid` | `components/app/numeric-text-field.tsx:74` | ✅ Morto (20) | `components/app/money-field.test.tsx:49` |
| O04 | "Carregar mais" não envia o cursor | `features/identity/usuarios/hooks.ts:35` | ✅ Morto (2) | `features/identity/usuarios/users-page.test.tsx:163` |
| O05a | Contas do lançamento sem filtro por tipo | `features/financeiro/lancamentos/filters.ts:53` | ✅ Morto (4) | `features/financeiro/lancamentos/filters.test.ts:99` |
| O05b | Contas do lançamento incluem inativas | `features/financeiro/lancamentos/filters.ts:53` | ✅ Morto (3) | `filters.test.ts:99` |
| O06 | URL assinada gravada no cache de consultas | `features/financeiro/comprovantes/comprovantes-panel.tsx:45` | ❌ **Sobreviveu** (688 passaram) | nenhum → lacuna L2 |
| O07 | Própria linha oferece ações | `features/identity/usuarios/actions.ts:29` | ✅ Morto (1) | `features/identity/usuarios/dialogs/status-dialogs.test.tsx:54` |
| O08 | Portão mostra conteúdo com troca obrigatória | `app/(app)/layout.tsx:34` | ✅ Morto (1) | `app/(app)/layout.test.tsx:64` |
| O09 | Movimentação não invalida o saldo do produto | `features/estoque/movimentacoes/hooks.ts:33` | ✅ Morto (2) | `features/estoque/ajustes/ajuste-dialog.test.tsx:52` |
| O10 | Shell mostra navegação na troca obrigatória | `components/app/app-shell.tsx:37` | ✅ Morto (1) | `components/app/app-shell.test.tsx:128` |
| O11 | Salvar papéis vazios sem o aviso de vínculo dormente | `features/identity/usuarios/dialogs/admin-dialogs.tsx:161` | ✅ Morto (1) | `features/identity/usuarios/dialogs/admin-dialogs.test.tsx:230` |
| O12 | Motivo mínimo aceita 9 caracteres | `features/identity/usuarios/dialogs/reason.ts:10` | ✅ Morto (2) | `admin-dialogs.test.tsx:108` |
| O13 | Troca de senha não rebusca `/auth/me` | `features/identity/acesso/use-change-password.ts:38` | ✅ Morto (1) | `features/identity/acesso/change-password-form.test.tsx:107` |
| O14 | Login grava o token CSRF em `localStorage` | `lib/session/session-provider.tsx:175` | ✅ Morto (1) | `lib/session/session-provider.test.tsx:159` |
| O15 | Logout não limpa cache nem contexto | `lib/session/session-provider.tsx:198` | ✅ Morto (2) | `session-provider.test.tsx:196` |
| O16 | `403 csrf_invalid` não rebusca `/auth/me` | `lib/session/session-provider.tsx:154` | ✅ Morto (1) | `session-provider.test.tsx:372` |
| O17 | Mensagem genérica expõe o `code` interno | `lib/api/problem.ts:89` | ✅ Morto (2) | `lib/api/problem.test.ts:70` |
| O18 | Histórico consultado sem `estoque:movimentacao:read` | `features/estoque/produto-detail.tsx:78` | ✅ Morto (1) | `features/estoque/movimentacoes/movimentacoes-panel.test.tsx:102` |
| O19 | Subconta não envia `parent_id` | `features/financeiro/contas/conta-dialogs.tsx:72` | ✅ Morto (1) | `features/financeiro/contas/contas-page.test.tsx:188` |
| O20 | Filtro de papel não vai para a API | `features/identity/usuarios/hooks.ts:34` | ✅ Morto (2) | `features/identity/usuarios/users-page.test.tsx:98` |
| O21 | `401` redireciona mas não limpa cache nem contexto | `lib/session/session-provider.tsx:146` | ✅ Morto (1) | `session-provider.test.tsx:259` |

**Profundidade**: completa para caminhos críticos (autenticação, dinheiro, permissões, uploads).
**Resultado do sensor**: 39 mutantes injetados, 38 mortos, 1 sobrevivente (O06). Os 10 obrigatórios de INT-04 AC3 morreram nas 16 variantes; dos 23 próprios, 22 morreram. Em todos os mortos, os testes que falharam são os que cobrem o comportamento mutado (nomes conferidos um a um; nenhuma morte por tempo esgotado).

**Isolamento**: `git status --porcelain` do worktree real (`tjw/verify`) e do repositório principal (`torcida-jovem`) vazio antes e depois do sensor; worktree temporário removido ao fim.

---

## Smoke repetido (stack real, 2026-10-08)

Montado pelo `docs/development/web.md` a partir deste worktree: `docker compose up -d db` e migrações 1 a 9, contêiner Garage avulso (`dxflrs/garage:v2.4.1`, porta 3900, chaves geradas na hora), `bootstrap-admin --role PRESIDENTE`, binário de `./cmd/api` na 8080 e `next build` + `next start` na 3000 com `API_URL=http://localhost:8080`. As chamadas são as que as telas fazem, sempre por `http://localhost:3000` (nunca direto na API), com `Origin`, cookie e `X-CSRF-Token`. Senhas geradas em memória; nenhum segredo foi gravado em arquivo.

**Sem navegador automatizado**: o servidor MCP do Playwright não conectou e a V1 não tem E2E (WEB-D-017). Renderização, cliques, diálogos e redirecionamentos feitos pelo JavaScript não foram vistos; ficam cobertos só pelos testes com MSW.

| # | Verificação | Resultado |
| --- | --- | --- |
| S1 | `GET /` → `307` com `Location: /inicio`; `/entrar`, `/inicio`, `/financeiro/lancamentos`, `/financeiro/saldo`, `/estoque/produtos` → `200` com o título certo | Passou |
| S2 | Login: senha errada → `401 invalid_credentials`; certa → `200`, `Set-Cookie: tj_session=<redigido>; Path=/; Max-Age=28800; HttpOnly; SameSite=Lax`, sem `Domain`; `must_change_password = true` | Passou |
| S3 | Antes da troca: `GET /financeiro/contas` → `403 password_change_required`; troca sem CSRF → `403 csrf_invalid`; com CSRF → `204`; `/auth/me` → `must_change_password = false` | Passou |
| S4 | `PRESIDENTE` cria duas pessoas (`201`) e promove a `TESOURARIA` e a `CONSELHO_FISCAL` (`204`); cada uma entra e faz a troca obrigatória (`204`) | Passou |
| S5 | Tesouraria: contas `RECEITA`, `DESPESA` e subconta → `201`; `Origin` externa → `403 origin_not_allowed`; sem CSRF → `403 csrf_invalid` | Passou |
| S6 | Lançamento `RECEITA` de 15000 centavos → `201 CRIADA`; valor decimal (`150.5`) → `422 validation_failed` | Passou |
| S7 | `PUT` → `200`, líquido 19850; receber → `204`; receber de novo → `409 lancamento_nao_pode_ser_recebido`; editar recebida → `409 lancamento_imutavel` | Passou |
| S8 | Comprovante de 200.000 bytes → `201`, `size_bytes` igual; sem CSRF → `403 csrf_invalid` | Passou |
| S9 | Comprovante de **exatamente 10.420.224 bytes** (corpo multipart de 10.420.401 bytes) → `201`, `size_bytes = 10420224`; URL assinada → `200`; download com 10.420.224 bytes e SHA-256 igual ao enviado | Passou |
| S10 | Lista de comprovantes → `200`, 2 itens | Passou |
| S11 | Devolução → `201` (`DESPESA`, `CRIADA`, `devolucao_de_id` certo); pagar → `204` | Passou |
| S12 | Cancelar com motivo só de espaços → `422 motivo_obrigatorio`; com motivo → `204`; lista mostra `CANCELADA` com o motivo | Passou |
| S13 | Saldo → `200`, `saldo_cents = 14850` (19850 − 5000) | Passou |
| S14 | Tesouraria em `GET /users` → `403 forbidden` | Passou |
| S15 | `CONSELHO_FISCAL`: `/auth/me` com 11 permissões, nenhuma de escrita | Passou |
| S16 | `CONSELHO_FISCAL`, leituras → `200`: contas, lançamentos, saldo, comprovantes, URL do comprovante, produtos, movimentações, saldo de estoque | Passou |
| S17 | `CONSELHO_FISCAL`, escritas → `403 forbidden`: criar, renomear e desativar conta; criar, editar, receber, pagar, cancelar e devolver lançamento; anexar comprovante; criar produto; movimentação; ajuste. `GET /users` → `403`. Saldo inalterado depois | Passou |
| S18 | Logout sem CSRF → `403 csrf_invalid`; com CSRF → `204` e `Set-Cookie ... Max-Age=0`; cookie antigo → `401` | Passou |
| S19 | Sem cookie: `GET /auth/me` → `401` | Passou |
| S20 | Erro pelo rewrite chega com `Content-Type: application/problem+json` (`422 reason_required`) | Passou |
| S21 | Senha temporária → `200` com `Cache-Control: no-store` devolvido ao cliente; pedido de recuperação para conta existente e inexistente → `202` com a mesma mensagem | Passou |
| S22 | Confirmar com token inválido → `400 invalid_reset_token`; o log da API registrou só `recipient_domain` e `subject` | Passou |
| S23 | Seis senhas erradas → `429 login_blocked` com `Retry-After: 899` devolvido ao cliente | Passou |

Total: 97 verificações em duas execuções (83 + 14), nenhuma falha. A primeira tentativa da segunda execução falhou por erro do próprio verificador (API iniciada antes das migrações); foi corrigida e repetida, sem relação com o produto.

**Desmontagem**: API e `next start` encerrados, contêiner Garage removido, `docker compose down -v`, binários, logs e configuração temporária apagados. Nenhum `.env` foi criado.

---

## Gate Check

Rodados em `tjw/verify/web` com `API_URL=http://localhost:8080`, código de saída real por `cmd /v:on /c "<cmd> & echo EXIT=!ERRORLEVEL!"`.

| Gate | Comando | Resultado |
| --- | --- | --- |
| Web | `pnpm lint` | `EXIT=0` |
| Web | `pnpm exec next typegen` | `EXIT=0` |
| Web | `pnpm exec tsc --noEmit` | `EXIT=0` |
| Web | `pnpm test` | `EXIT=0` (66 arquivos, 688 testes, 0 falhas, 0 pulados) |
| Web | `pnpm build` | `EXIT=0` (15 rotas) |
| **Audit** | `pnpm audit --audit-level high` | **`EXIT=1`**: GHSA-68fv-2mgg-jv7q, alta, `source-map-js` `>=1.0.0 <1.2.2` (lockfile com 1.2.1, `web/pnpm-lock.yaml:3307`, via `@tailwindcss/node@4.3.3` e `css-tree@3.2.1`). "2 high (1 ignored)": o ignorado é GHSA-vfj7-8cjw-p6xm (`web/pnpm-workspace.yaml:12`) |
| Contract | `pnpm lint:api` | `EXIT=0` |
| Contract | `pnpm gen:api:check` | `EXIT=0` (com um aviso de validação de exemplo em `identity.yaml`, sem falha) |
| Backend intacto | `git diff --name-only develop -- api/` | vazio |
| State | `validate_state.py --root <worktree> .specs/features/web` | `EXIT=1`, esperado com veredito FAIL ("validation.md verdict is FAIL"). Atenção: com o argumento literal `web` o script resolve o diretório de código `web/` da raiz (que existe) e acusa "no validation.md"; use o caminho `.specs/features/web` |

- **Testes antes da feature** (`ee1b497`): 2 arquivos (`lib/money.test.ts` e `lib/smoke.test.ts`), ambos mantidos sem alteração. **Depois**: 66 arquivos, 688 testes. Nenhum teste removido, enfraquecido ou pulado.
- **Sobre o Audit**: o lockfile de `develop` (`1a508d1`) tem a mesma versão e `pnpm audit --audit-level high` em `develop` também sai com 1 hoje. O aviso não foi introduzido pela Web V1; ele passou a reprovar depois da rodada dos autores.

---

## Code Quality

| Princípio | Status |
| --- | --- |
| Código mínimo, sem recurso além do pedido | ✅ |
| Mudanças cirúrgicas (só `web/`, `docs/`, `.specs/features/web/` e `API_URL` no CI, previsto em `STATE.md:110`) | ✅ |
| Sem dependência além de WEB-D-003; `package.json` só em F1 | ✅ |
| Asserções com o valor definido na spec | ✅ com 1 exceção (FWB-05 AC5) |
| Cobertura por camada (felizes, borda e erro por rota) | ✅ |
| Diretrizes seguidas: `CLAUDE.md` (dinheiro em centavos inteiros, sem JWT, CSRF, RBAC por permissão) | ✅ |

Observações sem peso no veredito: `features/identity/usuarios/actions.ts` não tem teste unitário próprio (a regra é coberta pelos testes de tela, e o mutante O07 morreu); a cláusula "nenhuma tela chama `fetch` diretamente" (FND-01 AC2) não tem guarda automática (por exemplo, regra de lint).

---

## Lacunas (ordenadas por severidade)

### L1 — Gate Audit vermelho (bloqueante)

- **Critério**: gate Audit de T4; `STATE.md`, "pronta para uso", item 7.
- **Onde**: `web/pnpm-lock.yaml:3307` (`source-map-js@1.2.1`), trazido por `@tailwindcss/node@4.3.3` e `css-tree@3.2.1`.
- **Causa**: GHSA-68fv-2mgg-jv7q (negação de serviço em `source-map-js` < 1.2.2). Já existia no lockfile de `develop`.
- **Dona provável**: F1 (única unidade que altera `package.json` e `pnpm-lock.yaml`, WEB-D-011).
- **Correção sugerida**: levar `source-map-js` a `>=1.2.2` (atualizar `@tailwindcss/postcss`, ou `overrides` do pnpm para `source-map-js`). Exceção de auditoria só se não houver versão corrigida, o que não é o caso (1.2.2 já está no lockfile por outro caminho). **Feito quando**: `pnpm audit --audit-level high` sai com 0 e os demais gates continuam verdes.

### L2 — FWB-05 AC5 sem asserção sobre o cache (menor)

- **Critério**: FWB-05 AC5, "sem guardá-la no cache de consultas".
- **Onde**: `web/features/financeiro/comprovantes/comprovantes-panel.test.tsx:235` (só conta as chamadas e confere `window.open`).
- **Evidência**: mutante O06 (gravar a URL com `setQueryData` em `comprovantes-panel.tsx:45`) passou nos 688 testes. O código atual está correto (`web/features/financeiro/comprovantes/hooks.ts:52` não usa o `QueryClient`); falta o teste que impediria a regressão.
- **Dona provável**: FIN-b.
- **Correção sugerida**: no teste de download, depois dos cliques, afirmar que nenhuma consulta nem mutação do `QueryClient` contém a URL assinada (como `web/features/identity/usuarios/dialogs/temporary-password-dialog.test.tsx:127` faz com a senha). **Feito quando**: o mutante O06 morre.

### Pendências (não são lacunas de teste; exigem pessoa)

- **Interação no navegador**: ninguém, nem os autores nem este verificador, viu as telas num navegador. O roteiro está em `.specs/features/web/smoke.md:58`. Recomenda-se que o mantenedor o execute antes de declarar a Web V1 pronta.
- **FND-05 AC2**: a troca da navegação pelo ponto de quebra só é observável no navegador (SP-1).

---

## Gaps de precisão da spec

| # | Critério | Imprecisão | Efeito |
| --- | --- | --- | --- |
| SP-1 | FND-05 AC2 | "Ponto de quebra de tablet" sem valor nem forma de verificação; jsdom não avalia media queries | O teste confere o menu recolhível, não a troca pelo ponto de quebra |
| SP-2 | FND-03 AC3, ACS-01 AC6 | "Quanto tempo esperar" sem formato | Os testes fixam "15 minutos", escolha do autor |
| SP-3 | USR-04 AC4 | "Com motivo", sem tamanho mínimo (a promoção exige 10 caracteres; a API aceitou 5 na retirada, segundo `smoke.md:55`) | O front exige só motivo não vazio |
| SP-4 | EWB-03 AC3 | "A confirmação" não diz se é um passo separado | Implementado como diálogo único com "Confirmar ajuste" |
| SP-5 | FND-01 AC1 | Preservação de cabeçalhos sem tipo de teste definido | Só o smoke exercita; o teste unitário cobre a configuração do rewrite |
| SP-6 | INT-03 AC1 | "Data" por item ou por execução | `smoke.md` traz uma data para a execução inteira |

---

## Lições registradas (`lessons.py`, candidatas)

| Lição | Sinal | Origem |
| --- | --- | --- |
| L-002 | `gate_fail` | L1 (gate Audit) |
| L-003 | `surviving_mutant` | L2 (mutante O06) |
| L-004 a L-009 | `spec_precision_gap` | SP-1 a SP-6 |

---

## Requirement Traceability Update

Veredito FAIL: nenhum Status foi alterado nas specs e as caixas de T4 não foram marcadas. Situação proposta para depois das correções:

| Requisito | Status atual | Situação nesta verificação |
| --- | --- | --- |
| FND-01 a FND-06 | In Tasks | Verificados (FND-05 AC2 com SP-1) |
| ACS-01 a ACS-03 | In Tasks | Verificados |
| USR-01 a USR-05 | In Tasks | Verificados |
| FWB-01 a FWB-04 | In Tasks | Verificados |
| FWB-05 | In Tasks | ❌ Precisa de correção (L2) |
| EWB-01 a EWB-04 | In Tasks | Verificados |
| INT-01 a INT-03 | In Tasks | Verificados |
| INT-04 | In Tasks | ❌ Aberto até L1 e L2 fecharem |

---

## Summary

**Overall**: ❌ Not Ready

**Checagem ancorada na spec**: 140 de 142 critérios com evidência discriminante; 1 lacuna (FWB-05 AC5); 6 gaps de precisão sinalizados.
**Sensor**: 38 de 39 mutantes mortos (os 10 obrigatórios, todos); 1 sobrevivente.
**Gates**: Web, Contract e Backend intacto verdes; **Audit vermelho**.
**Smoke repetido**: 97 verificações, nenhuma falha; sem navegador.

**O que funciona**: acesso same-origin com cookie, `Origin` e CSRF; login, troca obrigatória, sessão expirada e logout; administração de usuários; ciclo do financeiro com comprovante no limite de 10.420.224 bytes; estoque; visão somente leitura.

**Próximos passos**: F1 corrige L1; FIN-b corrige L2; nova verificação (ciclo 1 de 3). A re-verificação pode se limitar ao gate Audit, ao mutante O06 e aos gates completos.
