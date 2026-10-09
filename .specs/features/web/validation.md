# Web V1 — Validation (ciclo 2)

**Date**: 2026-10-08
**Veredito**: FAIL ❌
**Ciclo**: 2 (terceira verificação independente; os ciclos 0 e 1 terminaram em FAIL)
**Spec**: `.specs/features/web/spec/01-fundacao.md` a `06-integracao.md`
**Diff range**: `ee1b497..b57ee9e` em `feature/web-v1` (`ee1b497` é a especificação; `develop` = `6382fe1`)
**Verifier**: sub-agente independente (autor ≠ verificador), em `feature/web-verify3`

Este relatório substitui por inteiro o do ciclo 1. Toda citação `arquivo:linha` vem da leitura do código e dos testes no commit `b57ee9e`, feita neste ciclo. O relatório anterior, `LESSONS.md`, `lessons.json` e os relatórios de autores não foram usados como evidência; o relatório do ciclo 1 foi lido só no fim, para conferir o fechamento das lacunas. `evidence/f1-rewrite.md` e `smoke.md` foram tratados como pista e re-verificados contra a stack real.

Caminhos sem prefixo são relativos a `web/`.

---

## Resumo

| Item | Resultado |
| --- | --- |
| ACs verificados (sub-specs 01 a 05 e INT-01 a INT-03) | 129 verificados, todos com teste e código citados em `arquivo:linha`; 128 com evidência discriminante completa; 1 com evidência parcial (FND-03 AC6, lacuna L3); 0 sem nenhuma evidência |
| Casos de borda das specs | 13 de 13 com evidência |
| Sensor de discriminação | 118 mutantes: 111 mortos por teste pertinente; 7 sobreviventes, dos quais 2 relevantes (S06 e S06c, a mesma lacuna L3) e 5 equivalentes ou fora da spec, justificados |
| Gates | 10 de 10 com `EXIT=0`; `git diff --name-only develop -- api/` vazio |
| Smoke repetido contra a stack real | 101 verificações, 0 falha |
| Fechamento das lacunas dos ciclos 0 e 1 | L1, L2 e O01 a O09 fechadas |
| Critérios de PASS | Não atendidos: há um AC sem evidência discriminante completa e um mutante relevante sobrevivente (a mesma lacuna, L3). Gates e smoke atendidos |
| Lacunas novas | 1 — L3, FND-03 AC6: a variante `401 unauthenticated` não tem teste discriminante (severidade média; correção só de teste) |
| Pendência humana | interação visual no navegador (não é lacuna de teste) |

---

## Task Completion

| Tasks | Status | Notas |
| --- | --- | --- |
| 01-fundacao T1 a T8 | ✅ Done | caixas marcadas |
| 02-identity-acesso T1 a T2 | ✅ Done | caixas marcadas |
| 03-identity-usuarios T1 a T3 | ✅ Done | caixas marcadas |
| 04-financeiro-web T1 a T5 | ✅ Done | caixas marcadas |
| 05-estoque-web T1 a T3 | ✅ Done | caixas marcadas |
| 06-integracao T1 a T3 | ✅ Done | caixas marcadas |
| 06-integracao T4 | ❌ Não concluída | esta verificação |

Conferido por busca: nas seis listas de tasks as únicas caixas abertas antes deste ciclo eram as três de T4 (`.specs/features/web/tasks/06-integracao.md:117-119`).

Correções desde `8bf79ec` (`git diff --name-only 8bf79ec b57ee9e`): só arquivos `*.test.ts(x)`, `features/estoque/test-helpers.ts`, `features/financeiro/lancamentos/test-helpers.ts` e `.specs/` (`LESSONS.md`, `lessons.json`, `validation.md`). Nenhum arquivo de produção mudou.

---

## Spec-Anchored Acceptance Criteria

Cada linha traz o resultado definido pela spec, o teste que o afirma e o código que o implementa. ✅ = o valor afirmado é o da spec. ⚠️ = gap de precisão da spec (seção própria; não derruba o veredito).

### 01 — Fundação (FND)

| AC | Resultado definido na spec | Teste (`arquivo:linha` — afirmação) | Código | Resultado |
| --- | --- | --- | --- | --- |
| FND-01 AC1 | `/api/v1/*` vai ao mesmo caminho em `API_URL`; método, corpo e cabeçalhos preservados nos dois sentidos | `next.config.test.ts:13-20` — `toEqual([{ source: "/api/v1/:path*", destination: "http://localhost:8080/api/v1/:path*" }])`; `:38-47` lê `API_URL` do ambiente. Preservação observada no smoke: `Set-Cookie`, `Content-Type` (`application/json`, `application/problem+json`, multipart), `Origin`, `X-CSRF-Token`, `Retry-After=899`, `Cache-Control: no-store` | `next.config.ts:15-20` | ✅ ⚠️ G1 |
| FND-01 AC2 | Cliente `openapi-fetch` sobre os tipos gerados, `credentials: "same-origin"`; nenhuma tela chama `fetch` direto | `lib/api/client.test.ts:49-67` — `expect(seen[0].url).toBe(ORIGIN + path)` e `expect(seen[0].credentials).toBe("same-origin")` para os três contratos. Busca por `fetch(` em `app/`, `components/`, `features/`, `lib/`: única ocorrência de produção em `lib/api/client.ts:116` | `lib/api/client.ts:111-127` | ✅ |
| FND-01 AC3 | Escritas levam `X-CSRF-Token` com o token da sessão | `lib/api/client.test.ts:72-108` — POST, PUT, PATCH e DELETE: `expect(seen[0].csrf).toBe(TOKEN)`; `:110-114` GET: `toBeNull()`; `lib/session/session-provider.test.tsx:139` e `:190` — token do login e de `/auth/me` | `lib/api/client.ts:66-71` | ✅ |
| FND-01 AC4 | `problem+json` vira erro tipado com `status`, `code`, `errors[]` | `lib/api/client.test.ts:208-228` — `toBeInstanceOf(ApiError)` e `toMatchObject({ status: 422, code: "validation_failed", errors: [{ field: "nome", code: "required" }] })`; `lib/api/problem.test.ts:24-45` | `lib/api/client.ts:135-147`, `lib/api/problem.ts:59-68` | ✅ |
| FND-01 AC5 | Aceite contra a API real: (a) cookie na origem do front, (b) escrita passa pelo `Origin`, (c) comprovante ≥ 9,5 MiB íntegro | Aceite manual. Smoke repetido neste ciclo (seção "Smoke"): (a) `Set-Cookie: tj_session=<redigido>; Path=/; Max-Age=28800; HttpOnly; SameSite=Lax`, sem `Domain`; (b) `POST /financeiro/contas` → `201`, e com `Origin` externa → `403 origin_not_allowed`; (c) arquivo de 10.420.224 bytes (9,94 MiB) → `201`, SHA-256 do download igual. Pista do autor: `.specs/features/web/evidence/f1-rewrite.md:12-16` | `next.config.ts:15-20` | ✅ |
| FND-01 AC6 | Se o AC5 falhar, F1 para e reporta | Condição de parada de processo, não acionada: os três itens do AC5 passaram (`evidence/f1-rewrite.md:18`, confirmado pelo smoke deste ciclo). `git diff --name-only develop -- api/` vazio | — | ✅ |
| FND-02 AC1 | Kit em `components/ui/` com campo, rótulo, grupo com erro, seleção, área de texto, caixa de seleção, diálogo, confirmação, tabela, menu, esqueleto, alerta, toast | `components/ui/field.test.tsx:29-37` (campo e rótulo), `:81-95` (área de texto), `:97-112` (caixa de seleção), `:114-135` (seleção); `components/ui/overlays.test.tsx:41-61` (diálogo), `:63-92` (confirmação), `:114-130` (menu); `components/ui/display.test.tsx:19-42` (tabela), `:44-56` (alerta), `:58-65` (esqueleto), `:67-78` (toast) | `components/ui/*.tsx` (17 componentes) | ✅ |
| FND-02 AC2 | Erro associado por `aria-describedby` e campo com `aria-invalid` | `components/ui/field.test.tsx:39-53` — `expect(input.getAttribute("aria-invalid")).toBe("true")` e `expect(describedByText(input)).toContain("Informe um e-mail válido.")`; idem `:82-94`, `:98-111`, `:115-134` | `components/ui/field.tsx` | ✅ |
| FND-02 AC3 | Dependência nova para o kit: F4 para e reporta, sem tocar `package.json` | Condição de parada, não acionada: `git log ee1b497..b57ee9e -- web/package.json web/pnpm-lock.yaml` lista só `57cfbad` (F1) e a atualização de segurança `6382fe1`/`8bf79ec` | — | ✅ |
| FND-03 AC1 | Login chama `POST /auth/login`, guarda o contexto e leva a `next` interno ou ao início | `app/(public)/entrar/login-form.test.tsx:48-55` — `navigations()` igual a `["/financeiro/contas?tipo=RECEITA"]` e corpo `{ email, password }`; `:57-63` sem `next` → `["/inicio"]`; `lib/session/routes.test.ts:20-27` | `app/(public)/entrar/login-form.tsx:58-62`, `lib/session/session-provider.tsx:161-179`, `lib/session/routes.ts:43-46` | ✅ |
| FND-03 AC2 | `401 invalid_credentials` → "E-mail ou senha inválidos.", sem apontar o campo | `login-form.test.tsx:90-99` — `findByText("E-mail ou senha inválidos.")` e `aria-invalid` diferente de `"true"` nos dois campos | `login-form.tsx:27`, `:73` | ✅ |
| FND-03 AC3 | `429 login_blocked` informa o bloqueio e o tempo de `Retry-After` | `login-form.test.tsx:104-112` — `Retry-After: 900` → texto contém "15 minutos"; `lib/session/retry-after.test.ts:8-20` | `login-form.tsx:29-33`, `lib/session/session-provider.tsx:169-173`, `lib/session/retry-after.ts:19-34` | ✅ ⚠️ G4 |
| FND-03 AC4 | `must_change_password` leva a `/conta/senha` e nenhuma outra tela aparece | `login-form.test.tsx:79-85` — `["/conta/senha"]` mesmo com `next`; `app/(app)/layout.test.tsx:64-70` — redireciona e `queryByText("conteúdo protegido")` nulo | `login-form.tsx:62`, `app/(app)/layout.tsx:23-34` | ✅ |
| FND-03 AC5 | Sair chama `POST /auth/logout`, limpa cache e contexto e leva a `/entrar` | `lib/session/session-provider.test.tsx:196-225` — `navigations()` `["/entrar"]`, CSRF enviado, `status` `anonymous`, `getQueryData` indefinido e nenhuma consulta com dado; `components/app/app-shell.test.tsx:110-123` | `lib/session/session-provider.tsx:125-129`, `:181-199` | ✅ |
| FND-03 AC6 | `401` limpa tudo e leva a `/entrar` com aviso e `next` | `session-provider.test.tsx:259-279` — `pathname` `/entrar`, `next` = `/financeiro/contas?tipo=RECEITA`, `sessao` = `encerrada`, cache limpo; `login-form.test.tsx:165-168` — aviso "sessão foi encerrada"; `lib/api/client.test.ts:131-137`. Todas essas afirmações usam `code = session_expired`; a variante `unauthenticated` não é discriminada (mutantes S06 e S06c sobrevivem) | `lib/api/client.ts:80-81`, `session-provider.tsx:141-148` | ❌ GAP parcial (L3) |
| FND-03 AC7 | `403 password_change_required` leva a `/conta/senha` | `session-provider.test.tsx:355-367` — `navigations()` `["/conta/senha"]`; `client.test.ts:148-154` | `client.ts:82-83`, `session-provider.tsx:149-151` | ✅ |
| FND-03 AC8 | `403 csrf_invalid` rebusca `/auth/me`, pede para repetir, não repete sozinho | `session-provider.test.tsx:372-396` — `me.count` vai a 2, texto "A sessão foi atualizada. Repita a ação, por favor.", `writes` = 1 | `session-provider.tsx:152-156` | ✅ |
| FND-03 AC9 | Token CSRF e contexto só em memória | `session-provider.test.tsx:159-176` — `localStorage.length` 0, `sessionStorage.length` 0, `document.cookie` inalterado | `session-provider.tsx:126`, `:175` (só `QueryClient`) | ✅ |
| FND-03 AC10 | Link "Esqueci minha senha" para `/recuperar-acesso` | `login-form.test.tsx:158-162` — `href` = `/recuperar-acesso` | `login-form.tsx:128-133` | ✅ |
| FND-04 AC1 | Sem sessão: `/entrar` com `next`, sem conteúdo | `app/(app)/layout.test.tsx:42-51` — `pathname` `/entrar`, `next` = caminho pedido, conteúdo ausente | `app/(app)/layout.tsx:26-28`, `session-provider.tsx:201-203` | ✅ |
| FND-04 AC2 | Enquanto consulta a sessão: carregando, sem conteúdo | `layout.test.tsx:27-38` — `getByRole("status")` com "Carregando" e conteúdo ausente | `app/(app)/layout.tsx:49-58` | ✅ |
| FND-04 AC3 | Exibição pelas permissões efetivas, nunca pelo papel | `session-provider.test.tsx:78-99` — papéis `ADMIN_SISTEMA`/`PRESIDENTE` sem permissão → `false`; `components/app/nav.test.ts:41-43`; `components/app/require-permission.test.tsx:102-121`. Isolamento por ação: ver USR, FWB e EWB abaixo | `session-provider.tsx:227-237`, `components/app/require-permission.tsx:21-31` | ✅ |
| FND-04 AC4 | Sem a leitura da área: "Sem acesso", sem chamar a API da área | `require-permission.test.tsx:60-73` — título "Sem acesso" e `calls.count` 0. Por rota: `app/(app)/financeiro/contas/page.test.tsx:40-48`, `app/(app)/financeiro/saldo/page.test.tsx:39-47`, `features/financeiro/lancamentos/lancamentos-page.test.tsx:317-324`, `features/financeiro/lancamentos/lancamento-detail.test.tsx:207-216`, `app/(app)/admin/usuarios/page.test.tsx:60-69`, `app/(app)/estoque/produtos/page.test.tsx:45-53` e `:64-76` | `require-permission.tsx:26-30` e a `page.tsx` de cada rota | ✅ |
| FND-04 AC5 | `403 forbidden` → "Você não tem permissão para esta ação.", estado inalterado | `components/app/api-error.test.tsx:14-16` — `toBe("Você não tem permissão para esta ação.")`; `features/financeiro/contas/contas-page.test.tsx:228-244` e `:341-355` — mensagem exibida e árvore/conta inalterada | `components/app/api-error.tsx:9`, `:15-20` | ✅ |
| FND-05 AC1 | Shell com nome, "Trocar senha", "Sair" e navegação filtrada por permissão | `components/app/app-shell.test.tsx:55-74` — `linkNames(nav)` exatos para dois conjuntos de permissão; `:102-108` — itens `["Trocar senha", "Sair"]` e navegação a `/conta/senha`; `components/app/nav.test.ts:26-39` | `components/app/app-shell.tsx:37-38`, `:66-79`, `components/app/nav.ts:31-42` | ✅ |
| FND-05 AC2 | Abaixo do ponto de quebra a navegação vira menu recolhível, nada inacessível | `app-shell.test.tsx:87-97` — "Abrir menu" abre o diálogo "Menu" com os mesmos itens e fecha ao escolher | `app-shell.tsx:44-55` (`md:hidden`), `:84` (`hidden … md:block`) | ✅ ⚠️ G2 |
| FND-05 AC3 | Estados padronizados de carregamento, vazio e erro com tentar de novo | `components/app/states.test.tsx:12-15`, `:19-28`, `:32-38` — alerta com a mensagem e `onRetry` chamado 1 vez | `components/app/states.tsx:17`, `:37`, `:60` | ✅ |
| FND-05 AC4 | Dinheiro em pt-BR convertido por `parseBRL` em centavos inteiros, nunca float | `components/app/money-field.test.tsx:35-45` — vetores de `api/internal/platform/money/testdata/vectors.json`: `expect(last).toBe(cents)` e `Number.isSafeInteger(last)`; `:112-119` — `{ valor: 123456 }`; `lib/money.test.ts:25-27` | `components/app/money-field.tsx:14-23`, `lib/money.ts:36-46` | ✅ |
| FND-05 AC5 | Quantidade só inteira, sinal configurável | `components/app/quantity-field.test.tsx:40-64` (positivo: recusa decimal, `0`, `-3`), `:67-90` (diferente de zero com sinal) | `components/app/quantity-field.tsx:16-28` | ✅ |
| FND-05 AC6 | Confirmação explícita antes da API; com motivo obrigatório, confirmar desabilitado se vazio ou só espaços | `components/app/confirm-dialog.test.tsx:16-42` — `onConfirm` não chamado antes de confirmar nem ao cancelar; `:74-87` — `disabled` `true` com `""` e com `"  \n\t "`; `:89-96` — motivo aparado | `components/app/confirm-dialog.tsx:80`, `:87-98` | ✅ |
| FND-05 AC7 | `formatBRL` e datas pt-BR no fuso de São Paulo | `lib/money.test.ts:7-9` (vetores do Go); `components/app/format.test.ts:8-19` — `"2026-10-04T02:00:00Z"` → `"03/10/2026"`; `:23-26` | `lib/money.ts:23-29`, `components/app/format.ts:4-42` | ✅ |
| FND-05 AC8 | Com troca obrigatória o shell esconde a navegação e mostra só "Sair" | `app-shell.test.tsx:128-137` — `queryByRole("navigation")` nulo, nenhum link, itens do menu `["Sair"]` | `app-shell.tsx:37`, `:73-76` | ✅ |
| FND-06 AC1 | `422 validation_failed`: cada erro no campo correspondente | `lib/forms/apply-problem.test.ts:21-42`; `login-form.test.tsx:134-145`; `features/estoque/produtos/produtos-page.test.tsx:157-181` — `aria-invalid` e `aria-describedby` do campo certo | `lib/forms/apply-problem.ts:39-51` | ✅ |
| FND-06 AC2 | `code` conhecido → mensagem do catálogo | `lib/api/problem.test.ts:63-68`; `components/app/api-error.test.tsx:19-23` | `lib/api/problem.ts:84-86` | ✅ |
| FND-06 AC3 | `code` desconhecido → genérica com o status, sem detalhe interno | `problem.test.ts:70-85` — contém `"409"`, não contém `code`, `title` nem `detail`; `api-error.test.tsx:26-31` | `problem.ts:89` | ✅ |
| FND-06 AC4 | `413` → conteúdo grande demais | `problem.test.ts:87-91`; `api-error.test.tsx:34-36` — `toBe("O conteúdo enviado é grande demais.")` | `problem.ts:87` | ✅ |
| FND-06 AC5 | `503` ou rede → indisponível, com tentar de novo | `problem.test.ts:93-105`; `api-error.test.tsx:46-52` — botão "Tentar de novo" chama `onRetry`; `states.test.tsx:32-38`; `app/(app)/layout.test.tsx:83-99` | `problem.ts:76-78`, `:88`, `api-error.tsx:35-39` | ✅ ⚠️ G3 |

**Casos de borda (01)**

- [x] Dois `401` quase juntos → um único redirecionamento: `session-provider.test.tsx:281-300` — `navigations()` com 1 item (`session-provider.tsx:110-118`).
- [x] `next` de outra origem ignorado: `lib/session/routes.test.ts:29-46` (14 formas) e `login-form.test.tsx:65-74` — `["/inicio"]` (`routes.ts:27-37`).
- [x] Pessoa autenticada em `/entrar` vai ao início: `login-form.test.tsx:177-182` — `["/inicio"]` mesmo com `next` na URL (`login-form.tsx:53-56`).

### 02 — Identity: acesso e senha (ACS)

| AC | Resultado definido na spec | Teste | Código | Resultado |
| --- | --- | --- | --- | --- |
| ACS-01 AC1 | `POST /auth/password` com `current_password` e `new_password`; em `204`, confirma | `features/identity/acesso/change-password-form.test.tsx:69-84` — corpo `{ current_password, new_password }`, CSRF e "Senha alterada com sucesso." | `features/identity/acesso/use-change-password.ts:14-27`, `change-password-form.tsx:59-62`, `:106-110` | ✅ |
| ACS-01 AC2 | Confirmação diferente: erro no campo de confirmação, sem chamar a API | `change-password-form.test.tsx:89-102` — `aria-invalid` no campo de confirmação e `requests` vazio | `change-password-form.tsx:149-152` | ✅ |
| ACS-01 AC3 | Troca obrigatória com sucesso: rebusca `/auth/me` e vai ao destino ou ao início | `change-password-form.test.tsx:107-118` — `["/inicio"]` e `state.calls` 2; `:120-126` — `["/financeiro/contas"]` | `use-change-password.ts:38`, `change-password-form.tsx:79-82` | ✅ |
| ACS-01 AC4 | Modo obrigatório: explica, sem navegação, mantém sair | `change-password-form.test.tsx:139-154` — texto de obrigatoriedade, nenhum link, "Sair" leva a `/entrar`; `components/app/app-shell.test.tsx:128-137` | `change-password-form.tsx:94-98`, `:162-166` | ✅ |
| ACS-01 AC5 | Cinco códigos de senha, cada um no campo certo | `change-password-form.test.tsx:184-204` — `invalid_current_password` em "Senha atual"; `password_unchanged`, `password_too_short`, `password_too_long`, `password_compromised` em "Nova senha"; os outros campos sem `aria-invalid` | `features/identity/acesso/errors.ts:31-57` | ✅ |
| ACS-01 AC6 | `429 password_change_blocked` informa a espera de `Retry-After` | `change-password-form.test.tsx:224-233` — "Muitas tentativas. Tente de novo em 15 minutos." | `errors.ts:63-71`, `use-change-password.ts:22-25` | ✅ |
| ACS-02 AC1 | `POST /auth/password-reset/request`; em `202`, mensagem neutra da API | `features/identity/acesso/request-reset-form.test.tsx:47-52` — corpo `{ email }`; `:64-76` — mostra o texto devolvido pela API | `request-reset-form.tsx:22-25`, `:53-59` | ✅ |
| ACS-02 AC2 | Mesma mensagem e comportamento para qualquer e-mail | `request-reset-form.test.tsx:54-62` — `expect(second).toBe(first)` | `request-reset-form.tsx:53-63` | ✅ |
| ACS-02 AC3 | Link "Esqueci minha senha" em `/entrar` | `app/(public)/entrar/login-form.test.tsx:158-162` | `login-form.tsx:128-133` | ✅ |
| ACS-03 AC1 | Lê o token do fragmento, guarda em memória, remove o fragmento sem recarregar | `features/identity/acesso/reset-password-form.test.tsx:78-86` — `location.hash` `""`, `navigations()` vazio; `read-reset-token.test.ts:7-21` | `reset-password-form.tsx:43`, `:46-56` | ✅ |
| ACS-03 AC2 | Sem token: link inválido e oferta de `/recuperar-acesso` | `reset-password-form.test.tsx:91-98` — "O link de recuperação é inválido." e `href` `/recuperar-acesso` | `reset-password-form.tsx:98-114` | ✅ |
| ACS-03 AC3 | `POST /auth/password-reset/confirm` com `token` e `new_password`; em `204`, `/entrar` com aviso | `reset-password-form.test.tsx:103-119` — corpo `{ token: TOKEN, new_password }`, `["/entrar"]`, "Senha redefinida. Entre com a nova senha." | `reset-password-form.tsx:58-61`, `:88-93` | ✅ |
| ACS-03 AC4 | `400 invalid_reset_token`: expirou ou já foi usado, oferta de novo pedido | `reset-password-form.test.tsx:151-163` — "O link expirou ou já foi usado." | `reset-password-form.tsx:76-79` | ✅ |
| ACS-03 AC5 | Erros de senha no campo de nova senha, token mantido | `reset-password-form.test.tsx:168-198` — segunda tentativa envia o mesmo `TOKEN` | `reset-password-form.tsx:80-84` | ✅ |
| ACS-03 AC6 | Token nunca em query string, armazenamento ou log | `reset-password-form.test.tsx:65-74` (`expectTokenNotLeaked`: URL, `localStorage`, `sessionStorage`, `console.*`) e `:116` — `search` `""` | `reset-password-form.tsx:43` (só `useRef`) | ✅ |

**Casos de borda (02)**

- [x] Pessoa autenticada em `/redefinir-senha`: `reset-password-form.test.tsx:121-130` — redefine, `getQueryData(session.me)` nulo e `["/entrar"]` (`reset-password-form.tsx:89-93`).
- [x] `/conta/senha` sem troca obrigatória funciona como voluntária: `change-password-form.test.tsx:69-84` e `:156-160`; `app/(app)/conta/senha/page.test.tsx:17-26`.

### 03 — Identity: administração de usuários (USR)

| AC | Resultado definido na spec | Teste | Código | Resultado |
| --- | --- | --- | --- | --- |
| USR-01 AC1 | `GET /users`; nome, e-mail, situação, papéis, vínculo, troca pendente | `features/identity/usuarios/users-page.test.tsx:52-75` — células exatas de duas linhas e consulta `[{ limit: "50" }]` | `features/identity/usuarios/hooks.ts:23-43`, `users-table.tsx:57-65` | ✅ |
| USR-01 AC2 | Filtro refaz a consulta com `active` e `role`, do início | `users-page.test.tsx:98-113` (`role`), `:115-136` (`active=false`/`true`), `:138-158` (sem cursor depois de "Carregar mais") | `hooks.ts:25`, `:32-35`, `users-page.tsx:62-68` | ✅ |
| USR-01 AC3 | "Carregar mais" enquanto houver `next_cursor` | `users-page.test.tsx:163-178` — consulta com `cursor: "pagina-2"`, página acrescentada, botão some; `:180-185` | `hooks.ts:41`, `users-page.tsx:149-158` | ✅ |
| USR-01 AC4 | Sem `identity:user:read`: "Sem acesso" e item fora da navegação | `app/(app)/admin/usuarios/page.test.tsx:40-47`, `:60-69`, `:71-81`; `features/identity/nav.test.ts:20-44` | `app/(app)/admin/usuarios/page.tsx:13`, `features/identity/nav.ts:8` | ✅ |
| USR-02 AC1 | `POST /users`; em `201`, pessoa na lista e aviso da troca no primeiro acesso | `features/identity/usuarios/dialogs/create-user-dialog.test.tsx:68-93` — corpo `{ name, email, password }`, CSRF, aviso e linha nova | `hooks.ts:60-67`, `dialogs/create-user-dialog.tsx:41-42`, `:84` | ✅ |
| USR-02 AC2 | `409 email_taken` no campo de e-mail | `create-user-dialog.test.tsx:98-128` (linha `email_taken`) | `create-user-dialog.tsx:32-39`, `:71-75` | ✅ |
| USR-02 AC3 | `invalid_name`, `invalid_email`, `password_*` no campo correspondente | `create-user-dialog.test.tsx:98-128` (cinco linhas), com os outros campos sem `aria-invalid` | `create-user-dialog.tsx:32-39` | ✅ |
| USR-02 AC4 | Sem `identity:user:create`, "Novo usuário" não aparece | `create-user-dialog.test.tsx:58-63`; `features/identity/usuarios/users-page.permissions.test.tsx:137-151` (positivo isolado e negativo cruzado) | `users-page.tsx:87` | ✅ |
| USR-03 AC1 | `identity:user:update` confirma → `POST /users/{id}/deactivate`; inativo | `dialogs/status-dialogs.test.tsx:73-93` — `requests` vazio antes de confirmar; depois `[{ body: null, csrf }]` e "Inativo" | `hooks.ts:70-78`, `dialogs/status-dialogs.tsx:31-51` | ✅ |
| USR-03 AC2 | Reativar → `POST /users/{id}/reactivate`; ativo; aviso de que papéis e vínculo não voltam | `status-dialogs.test.tsx:98-117` — texto "…não são restaurados." | `hooks.ts:140-148`, `status-dialogs.tsx:15-16`, `:53-72` | ✅ |
| USR-03 AC3 | Desativação exige confirmação e informa o fim das sessões | `status-dialogs.test.tsx:82-86` — `alertdialog` com "as sessões dela serão encerradas" e nenhuma chamada | `status-dialogs.tsx:37-38` | ✅ |
| USR-03 AC4 | `self_change_forbidden`, `last_admin`, `privilege_escalation`: mensagem e estado mantido | `status-dialogs.test.tsx:122-137`, `:139-149` | `status-dialogs.tsx:25-29`, `features/identity/usuarios/errors.ts:24-26` | ✅ |
| USR-04 AC1 | `identity:admin:grant` + papel + motivo → `POST /users/{id}/admin-membership`; aviso de sessões encerradas e troca de senha | `dialogs/admin-dialogs.test.tsx:129-154` — corpo `{ roles: ["DIRETORIA", "TESOURARIA"], reason: "Assumiu a tesouraria" }`; `users-page.permissions.test.tsx:137-151` | `dialogs/admin-dialogs.tsx:102-113`, `hooks.ts:85-95` | ✅ |
| USR-04 AC2 | Sem papel ou motivo < 10 caracteres úteis: envio desabilitado | `admin-dialogs.test.tsx:108-127` — desabilitado sem papel, com `"   123456789   "`; habilitado com 10 | `admin-dialogs.tsx:100`, `dialogs/reason.ts:9-11` | ✅ |
| USR-04 AC3 | `identity:role:assign` → `PUT /users/{id}/roles`; lista vazia avisa do vínculo dormente antes | `admin-dialogs.test.tsx:210-228` — corpo `{ roles: ["TESOURARIA", "EVENTOS"] }`; `:230-243` — aviso e `requests` vazio antes de "Salvar sem papéis"; `users-page.permissions.test.tsx:168-172` | `admin-dialogs.tsx:158-174`, `hooks.ts:98-106` | ✅ |
| USR-04 AC4 | `identity:admin:revoke` com motivo → `POST …/admin-membership/revoke`; só `ASSOCIADO` | `admin-dialogs.test.tsx:248-270` — desabilitado sem motivo; corpo `{ reason: "Deixou a tesouraria" }`; papéis "Associado"; `users-page.permissions.test.tsx:174-178` | `admin-dialogs.tsx:214-235`, `hooks.ts:109-122` | ✅ |
| USR-04 AC5 | Dez códigos com mensagem em português | `admin-dialogs.test.tsx:159-187` — os dez códigos da spec, um por linha | `features/identity/usuarios/errors.ts:24-34` | ✅ |
| USR-04 AC6 | Só a ação compatível com o estado | `admin-dialogs.test.tsx:74-83`; `features/identity/usuarios/actions.test.ts:104-111` e `:113-128` (128 combinações) | `features/identity/usuarios/actions.ts:31-35` | ✅ |
| USR-05 AC1 | `identity:user:reset_password` + motivo ≥ 10 → `POST /users/{id}/password-reset`; senha uma vez, com copiar | `dialogs/temporary-password-dialog.test.tsx:74-87` (motivo), `:92-116` — corpo `{ reason }`, senha exibida, `writeText` chamado com ela | `dialogs/temporary-password-dialog.tsx:48`, `:58-82`, `hooks.ts:129-137` | ✅ |
| USR-05 AC2 | Ao fechar, senha descartada; fora do cache, do armazenamento e do log | `temporary-password-dialog.test.tsx:121-145` — `document.body.innerHTML`, cache de consultas, cache de mutações, `localStorage`/`sessionStorage` e `console.*` sem a senha; reabrir não a mostra | `temporary-password-dialog.tsx:50-56`, `hooks.ts:124-128` | ✅ |
| USR-05 AC3 | `reason_required`, `self_change_forbidden`, `privilege_escalation`, `user_inactive` | `temporary-password-dialog.test.tsx:151-164` | `features/identity/usuarios/errors.ts:24-34` | ✅ |

**Casos de borda (03)**

- [x] Própria linha sem ações: `dialogs/status-dialogs.test.tsx:54-68`; `actions.test.ts:155-170` (`actions.ts:29`).
- [x] Ação invalida as consultas de usuários: `admin-dialogs.test.tsx:147` e `:224` (a linha muda depois da escrita), `temporary-password-dialog.test.tsx:119` (`hooks.ts:51-54`).

### 04 — Financeiro (FWB)

| AC | Resultado definido na spec | Teste | Código | Resultado |
| --- | --- | --- | --- | --- |
| FWB-01 AC1 | `financeiro:conta:read` → `GET /financeiro/contas`; árvore por `parent_id`, com tipo e situação | `features/financeiro/contas/contas-page.test.tsx:97-114` — subcontas dentro do item pai, "Receita"/"Despesa", "Ativo"/"Inativo", `state.gets` 1; `features/financeiro/contas/tree.test.ts:26-38`; `app/(app)/financeiro/contas/page.test.tsx:31-38` | `features/financeiro/contas/contas-page.tsx:87`, `tree.ts:17-29` | ✅ |
| FWB-01 AC2 | `financeiro:conta:create` → `POST /financeiro/contas` com `tipo`, `nome` e `parent_id` na subconta | `contas-page.test.tsx:171-186` — `{ tipo: "DESPESA", nome: "Patrocínios" }`; `:188-210` — `{ tipo, nome, parent_id: ID_DESPESAS }`; `:380-388` (permissão isolada) | `features/financeiro/contas/hooks.ts:37-43`, `conta-dialogs.tsx:72` | ✅ |
| FWB-01 AC3 | Subconta fixa o `tipo` do pai | `contas-page.test.tsx:195-200` — rádio do pai marcado e os dois desabilitados | `conta-dialogs.tsx:69`, `:107` | ✅ |
| FWB-01 AC4 | `financeiro:conta:update` → `PATCH /contas/{id}`; `409 conta_ja_utilizada` explicado | `contas-page.test.tsx:264-280` — `{ id, body: { nome } }`; `:283-299` — "Esta conta já foi usada em um lançamento e não pode ser renomeada." | `hooks.ts:45-52`, `features/financeiro/contas/errors.ts:8` | ✅ |
| FWB-01 AC5 | `financeiro:conta:deactivate` confirma → `POST /contas/{id}/deactivate`; inativa | `contas-page.test.tsx:318-338` — `requests` vazio ao cancelar; depois 1 chamada e "Inativo" | `hooks.ts:54-62`, `contas-page.tsx:96-106` | ✅ |
| FWB-01 AC6 | `conta_nao_encontrada`, `conta_tipo_incompativel` | `contas-page.test.tsx:228-244`, `:283-299`, `:341-355` | `errors.ts:9-10` | ✅ |
| FWB-02 AC1 | `financeiro:saldo:read` → `GET /financeiro/saldo`; `saldo_cents` com `formatBRL`, inclusive zero e negativo | `features/financeiro/saldo/saldo-page.test.tsx:40-50` — "R$ 1.234,56", "R$ 0,00", "-R$ 987,65"; `app/(app)/financeiro/saldo/page.test.tsx:31-37` | `features/financeiro/saldo/saldo-page.tsx:34`, `hooks.ts:14-19` | ✅ |
| FWB-02 AC2 | Explica recebidos menos pagos, pelo status atual | `saldo-page.test.tsx:54-63` — texto exato | `saldo-page.tsx:39-42` | ✅ |
| FWB-03 AC1 | `financeiro:lancamento:read` → `GET /lancamentos`; oito colunas; filtros de tipo, status, conta e período no cliente | `features/financeiro/lancamentos/lancamentos-page.test.tsx:87-118` — linha exata e `list.count` 1; `:120-128` (status, sem nova chamada); `:130-145` (tipo, conta, período); `filters.test.ts:22-24` (ordem) e `:34-70` | `lancamentos-page.tsx:42`, `:123-144`, `filters.ts:31-48` | ✅ |
| FWB-03 AC2 | `financeiro:lancamento:create` → `POST /lancamentos` em centavos inteiros | `lancamentos-page.test.tsx:167-196` — `{ tipo: "RECEITA", conta_id, valor_bruto_cents: 15000, taxa_cents: 250, forma_pagamento: "PIX" }`; `:209-231` — `123456`; `:304-314` (permissão isolada) | `hooks.ts:69-76`, `lancamento-form.tsx:132-138`, `lancamentos-page.tsx:50` | ✅ |
| FWB-03 AC3 | `financeiro:lancamento:update` em `CRIADA` → `PUT /lancamentos/{id}` com conta, bruto, taxa e forma | `lancamento-detail.test.tsx:118-144` — corpo exato, sem `tipo` | `hooks.ts:139-150`, `lancamento-detail.tsx:244-246` | ✅ |
| FWB-03 AC4 | Editar só em `CRIADA`; `409 lancamento_imutavel` explica e atualiza a lista | `lancamento-detail.test.tsx:146-150`; `:152-160` — "Só lançamentos em aberto podem ser editados." e `list.count` 2; `actions.test.ts:40-48` | `actions.ts:30-31`, `errors.ts:8`, `:39-41`, `hooks.ts:60-67` | ✅ |
| FWB-03 AC5 | `conta_invalida`, `lancamento_tipo_incompativel`, `amount_out_of_range` junto do campo | `lancamentos-page.test.tsx:246-267` — `aria-invalid` e `aria-describedby` do campo; `lancamento-detail.test.tsx:162-171` | `errors.ts:32-36`, `lancamento-form.tsx:141-147` | ✅ |
| FWB-03 AC6 | Detalhe com todos os campos, cancelamento e referência de devolução | `lancamento-detail.test.tsx:62-76`, `:78-92` (motivo, autor, data), `:94-102` (link da receita) | `lancamento-detail.tsx:99-126` | ✅ |
| FWB-04 AC1 | `financeiro:lancamento:receive` confirma → `POST …/receive`; `RECEBIDA`; saldo atualizado | `workflow.test.tsx:73-86` — sem chamada antes de confirmar; `[{ body: null, csrf }]`; "Recebida"; `saldoInvalidated` `true` | `hooks.ts:79-97`, `lancamento-detail.tsx:168-191` | ✅ |
| FWB-04 AC2 | `financeiro:lancamento:pay` → `POST …/pay`; `PAGA`; saldo | `workflow.test.tsx:104-116` | `hooks.ts:99-108` | ✅ |
| FWB-04 AC3 | `financeiro:lancamento:cancel` com motivo → `POST …/cancel` com `reason`; `CANCELADA` com o motivo | `workflow.test.tsx:153-178` — `{ reason: "Lançado em duplicidade" }`, "Cancelada", motivo no detalhe | `hooks.ts:110-124`, `lancamento-detail.tsx:194-219` | ✅ |
| FWB-04 AC4 | Motivo vazio ou só espaços: confirmação desabilitada, API não chamada | `workflow.test.tsx:139-151` — `disabled` `true` e `requests` vazio | `components/app/confirm-dialog.tsx:80`, `lancamento-detail.tsx:204`, `:208` | ✅ |
| FWB-04 AC5 | `financeiro:lancamento:create` em receita `RECEBIDA` → `POST /lancamentos/devolucoes` com `devolucao_de_id`, conta de despesa, valores e forma | `workflow.test.tsx:210-238` — corpo exato e opções só de despesa | `devolucao-form.tsx:37-45`, `hooks.ts:128-137` | ✅ |
| FWB-04 AC6 | Devolução só em receita `RECEBIDA`; `devolucao_invalida` | `workflow.test.tsx:260-269`; `:240-249`; `actions.test.ts:29` | `actions.ts:38-39`, `errors.ts:13` | ✅ |
| FWB-04 AC7 | Receber só em receita `CRIADA`, pagar só em despesa `CRIADA`, cancelar só em não cancelados; três códigos mostram mensagem e atualizam a lista | `workflow.test.tsx:260-269`; `:88-99`, `:118-125`, `:180-188` (`list.count` 2); `actions.test.ts:24-56` | `actions.ts:28-41`, `errors.ts:39-46` | ✅ |
| FWB-05 AC1 | `financeiro:comprovante:read` → `GET …/comprovantes`; seis colunas, do mais novo para o mais antigo | `features/financeiro/comprovantes/comprovantes-panel.test.tsx:105-113` — fixture em ordem inversa, linhas exatas | `comprovantes-panel.tsx:31-33`, `:86-93`, `hooks.ts:14-24` | ✅ |
| FWB-05 AC2 | `financeiro:comprovante:create` → `multipart/form-data`, campo `file`; "enviando"; em `201`, atualiza | `comprovantes-panel.test.tsx:123-150` — `Content-Type` `multipart/form-data; boundary=…`, campo `file`, conteúdo, "Enviando…", `list.count` 2; `:317-334` (permissão isolada) | `hooks.ts:30-49`, `comprovantes-panel.tsx:58`, `:163-167` | ✅ |
| FWB-05 AC3 | Acima de 10.420.224 bytes: informa e não chama a API; até 10.420.224: envia | `comprovantes-panel.test.tsx:160-163` — `MAX_UPLOAD_BYTES` `toBe(10_420_224)`; `:174-181` — 10.420.224 enviado; `:183-199` — 10.420.225: mensagem e `requests` vazio (nenhuma requisição de nenhum tipo). Smoke: 10.420.224 bytes → `201`, download íntegro | `limits.ts:6`, `:21-23`, `comprovantes-panel.tsx:132-136` | ✅ |
| FWB-05 AC4 | `413` (`payload_too_large`, `document_too_large`) e `422` (três códigos) | `comprovantes-panel.test.tsx:204-220` — cinco códigos, lista inalterada | `features/financeiro/comprovantes/errors.ts:15-21` | ✅ |
| FWB-05 AC5 | Baixar: `GET /comprovantes/{documentId}/url` no clique, abre a URL, sem cache | `comprovantes-panel.test.tsx:236-255` — 0 chamadas antes do clique, 2 cliques → 2 chamadas e 2 URLs; `:257-272` — cache de consultas e de mutações sem a URL | `hooks.ts:52-59`, `comprovantes-panel.tsx:42-50` | ✅ |
| FWB-05 AC6 | `404 document_not_found` ou `lancamento_nao_encontrado` | `comprovantes-panel.test.tsx:274-281`; `:210` | `errors.ts:21-22` | ✅ |

**Casos de borda (04)**

- [x] Só leitura não vê nenhuma escrita: `contas-page.test.tsx:373-378`; `lancamentos-page.test.tsx:272-276`; `workflow.test.tsx:271-277`; `comprovantes-panel.test.tsx:302-307`.
- [x] Id fora da lista → "Lançamento não encontrado": `lancamento-detail.test.tsx:104-107` (`lancamento-detail.tsx:51-52`).
- [x] Mudança de status invalida lista e saldo: `workflow.test.tsx:84-85`, `:114-115`, `:176-177` (`hooks.ts:79-86`).

### 05 — Estoque (EWB)

| AC | Resultado definido na spec | Teste | Código | Resultado |
| --- | --- | --- | --- | --- |
| EWB-01 AC1 | `estoque:produto:read` → `GET /estoque/produtos`; código, nome, unidade; busca no cliente | `features/estoque/produtos/produtos-page.test.tsx:84-94`; `:102-117` (busca sem nova chamada); `search.test.ts:20-35` | `produtos-page.tsx:52`, `:71-81`, `search.ts:10-16` | ✅ |
| EWB-01 AC2 | `estoque:produto:create` → `POST /estoque/produtos`; produto na lista | `produtos-page.test.tsx:122-136` — corpo `{ codigo, nome, unidade_medida }` e CSRF; `:219-229` (permissão isolada) | `produtos/hooks.ts:50-56`, `produtos-page.tsx:35` | ✅ |
| EWB-01 AC3 | `409 codigo_duplicado` no campo de código; `422 validation_failed` em cada campo | `produtos-page.test.tsx:139-155`, `:157-181` | `create-produto-dialog.tsx:62-70` | ✅ |
| EWB-01 AC4 | `estoque:saldo:read` → `GET …/saldo`; saldo com unidade; negativo destacado | `features/estoque/produto-detail.test.tsx:47-61` — "12 UN", "-3 UN", "Saldo negativo"; `:63-69` | `produto-detail.tsx:74-76`, `:98-106`, `produtos/hooks.ts:37-47` | ✅ |
| EWB-02 AC1 | `estoque:movimentacao:read` → `GET …/movimentacoes`; sete colunas, na ordem da API | `features/estoque/movimentacoes/movimentacoes-panel.test.tsx:77-99`; `:272-282` (fixture fora de ordem cronológica, linhas na ordem da resposta); `:284-293` (resposta invertida) | `movimentacoes-panel.tsx:51-61`, `movimentacoes/hooks.ts:15-27` | ✅ |
| EWB-02 AC2 | `estoque:movimentacao:create` → `POST /estoque/movimentacoes` com `tipo`, `produto_id`, `quantidade`, `origem = INVENTARIO` | `movimentacoes-panel.test.tsx:161-169` — corpos exatos de entrada e saída; `payload.test.ts:14-30` | `payload.ts:16-25`, `produto-detail.tsx:67-72` | ✅ |
| EWB-02 AC3 | Devolução de `ENTRADA`/`SAIDA` com `tipo = DEVOLUCAO`, `movimentacao_de_id`, `INVENTARIO` | `movimentacoes-panel.test.tsx:151-158`, `:170-179`; `payload.test.ts:32-42` | `payload.ts:23`, `movimentacoes-panel.tsx:64-68` | ✅ |
| EWB-02 AC4 | Nunca oferece `VENDA`, `COMPRA`, `EVENTO`, `AJUSTE_MANUAL` | `movimentacoes-panel.test.tsx:130-136` — sem `combobox`, `radio`, `listbox` nem os oito rótulos; `payload.test.ts:44-49` | `movimentacao-dialog.tsx:126-129` | ✅ |
| EWB-02 AC5 | `saldo_insuficiente`, `produto_nao_encontrado`, `quantidade_invalida`, `devolucao_invalida` | `movimentacoes-panel.test.tsx:187-198`, `:200-212` | `features/estoque/errors.ts:14-21` | ✅ |
| EWB-02 AC6 | Registro invalida histórico e saldo | `movimentacoes-panel.test.tsx:181-183` — `movCalls` 4 e `saldoCalls` 4 depois de três registros | `movimentacoes/hooks.ts:30-35`, `:43` | ✅ |
| EWB-03 AC1 | `estoque:movimentacao:adjust` → `POST /estoque/ajustes` com `produto_id`, `quantidade`, `motivo` | `features/estoque/ajustes/ajuste-dialog.test.tsx:52-78` — `{ produto_id, quantidade: -3, motivo }`; `:148-161` (permissão isolada) | `ajustes/hooks.ts:17-23`, `ajuste-dialog.tsx:59-71`, `produto-detail.tsx:108` | ✅ |
| EWB-03 AC2 | Motivo vazio ou só espaços, ou quantidade zero: desabilitado, API não chamada | `ajuste-dialog.test.tsx:94-109` — cinco casos | `ajuste-dialog.tsx:43` | ✅ |
| EWB-03 AC3 | Mostra saldo atual e resultante; avisa negativo sem impedir | `ajuste-dialog.test.tsx:58-61` — "0 UN", "-3 UN", aviso e botão habilitado; `:80-89` | `ajuste-dialog.tsx:44`, `:111-129` | ✅ |
| EWB-03 AC4 | `422 motivo_obrigatorio`, `422 quantidade_invalida` | `ajuste-dialog.test.tsx:114-126` | `features/estoque/errors.ts:18-20` | ✅ |
| EWB-04 AC1 | Só as três leituras: vê tudo, sem criar produto, entrada, saída, devolução nem ajuste | `movimentacoes-panel.test.tsx:231-238`; `produtos-page.test.tsx:186-190`; `ajuste-dialog.test.tsx:131-135` | `produto-detail.tsx:67`, `:108`, `movimentacoes-panel.tsx:23` | ✅ |
| EWB-04 AC2 | Sem `movimentacao:read` ou `saldo:read`: seção omitida, API não chamada | `produto-detail.test.tsx:74-80`, `:118-153`; `movimentacoes-panel.test.tsx:102-107` | `produto-detail.tsx:74-80` | ✅ |

**Casos de borda (05)**

- [x] Id fora da lista → "Produto não encontrado": `produto-detail.test.tsx:37-42`, sem consultar saldo nem histórico (`produto-detail.tsx:41-45`).
- [x] `AJUSTE` e `DEVOLUCAO` sem "Devolver": `movimentacoes-panel.test.tsx:112-119` — `[true, true, false, false]`; `payload.test.ts:53-61` (`payload.ts:28-30`).

### 06 — Integração (INT-01 a INT-03)

| AC | Resultado definido na spec | Evidência | Código / documento | Resultado |
| --- | --- | --- | --- | --- |
| INT-01 AC1 | `nav-items.ts` junta os três `nav.ts` mais "Início", na ordem, sem redefinir | `components/app/nav-items.test.ts:14-26` — `toBe(financeiroNav)`, `toBe(estoqueNav)`, `toBe(identityNavSection)`; `:28-47` (caminhos e permissões) | `components/app/nav-items.ts:11-13` | ✅ |
| INT-01 AC2 | Shell usa `nav-items.ts`, só itens permitidos | `app/(app)/inicio/page.test.tsx:122-125` — renderiza pelo `AppLayout` real e compara os links da navegação para cinco perfis | `app/(app)/layout.tsx:34` | ✅ |
| INT-01 AC3 | `/` leva a `/inicio` | `app/page.test.tsx:19-27` — `redirect` chamado com `"/inicio"`; smoke: `GET /` → `307 /inicio` | `app/page.tsx:7-9` | ✅ |
| INT-01 AC4 | `/inicio`: um atalho por área visível; sem área, mensagem | `inicio/page.test.tsx:127-142` (atalhos e seções por perfil); `:147-155` — "Sua conta ainda não tem acesso a nenhum módulo." | `app/(app)/inicio/atalhos.tsx:15-34` | ✅ |
| INT-01 AC5 | Raiz não consulta `/healthz` | `app/page.test.tsx:20`, `:26` — `fetch` não chamado | `app/page.tsx:1-9` | ✅ |
| INT-02 AC1 | Guia com variáveis, ordem de subida, rewrite e gates | Leitura de `docs/development/web.md:13-41` (variáveis da API e do front), `:43-143` (banco, S3, administrador, API, front), `:155-166` (rewrite e limite), `:168-183` (gates). Stack deste ciclo montada pelo guia até o login | `docs/development/web.md` | ✅ |
| INT-02 AC2 | Comportamento de `API_URL` no build | `docs/development/web.md:39` (sem padrão; fixado no build). Conferido: `.next/routes-manifest.json` do build do gate contém `localhost:8080` e o `next start` o usou | `docs/development/web.md:39-41` | ✅ |
| INT-02 AC3 | Sem segredo, senha, cookie ou dado pessoal | `docs/development/web.md:5`; busca por valor de cookie, chave `GK…`, token e e-mail pessoal sem ocorrência; só marcadores `<...>` e contas `*@local.test` | `docs/development/web.md` | ✅ |
| INT-03 AC1 | Cada item com resultado, data e observação | `.specs/features/web/smoke.md:3` (data) e `:34-48` (12 itens com resultado e observação) | `.specs/features/web/smoke.md` | ✅ |
| INT-03 AC2 | Checklist cobre os onze assuntos | `smoke.md:36-47`: login (1), troca obrigatória (2), recuperação (3), token inválido (4), criação e promoção (5), senha temporária (6), financeiro com comprovante (7), estoque com ajuste negativo (8), `CONSELHO_FISCAL` (9), sessão expirada (10), logout (11) | `smoke.md` | ✅ |
| INT-03 AC3 | Declara que o token válido não é executável localmente | `smoke.md:12` e `:40` | `smoke.md` | ✅ |
| INT-03 AC4 | Falha registrada, corrigida e repetida | Condição não acionada: `smoke.md:50` ("Nenhum item falhou"); o smoke repetido neste ciclo também não teve falha | `smoke.md` | ✅ |
| INT-03 AC5 | Sem senha, token, cookie ou dado pessoal | `smoke.md:14`; mesma busca de INT-02 AC3, sem ocorrência | `smoke.md` | ✅ |

**Caso de borda (06)**

- [x] Módulo sem item visível some da navegação: `components/app/nav.test.ts:26-28`; `inicio/page.test.tsx:118` (`ASSOCIADO`: só "Início") (`components/app/nav.ts:35-36`).

**Status**: ❌ Gap presente — 128 de 129 ACs e 13 de 13 casos de borda com evidência discriminante; FND-03 AC6 com evidência parcial (L3); ⚠️ 4 gaps de precisão da spec sinalizados (seção própria).

---

## Fechamento das lacunas dos ciclos 0 e 1

Cada item foi recriado no código de produção, no worktree temporário, e observado neste ciclo. Conferência final com o relatório do ciclo 1 (lido só depois de fechada a evidência acima): as lacunas G1 a G6 dele correspondem aos mutantes O01 a O09 desta tabela, todos mortos agora; a lacuna L3 deste ciclo não constava nos ciclos anteriores.

| Lacuna | O que foi feito | O que foi observado | Estado |
| --- | --- | --- | --- |
| L1 — gate Audit | `pnpm audit --audit-level high` no worktree de verificação; no worktree temporário, removida a exceção e rodado `pnpm audit --json`; `pnpm why source-map-js` | Com a exceção: "1 high (1 ignored)", `EXIT=0`. Sem a exceção: `EXIT=1` e uma única advisory, `GHSA-vfj7-8cjw-p6xm` (`braces` 3.0.3). `GHSA-68fv-2mgg-jv7q` ausente. `pnpm why`: "Found 1 version of source-map-js" (1.2.2). Única exceção em `pnpm-workspace.yaml:10-12` | Fechada |
| L2 — FWB-05 AC5 (URL assinada sem cache) | Três mutantes em `features/financeiro/comprovantes/comprovantes-panel.tsx`: `setQueryData` (L2a), `useMutation` (L2b), `fetchQuery` (L2c) | Os três morrem em `comprovantes-panel.test.tsx:257` ("não guarda a URL assinada no cache de consultas nem no de mutações") | Fechada |
| O01 — ordem do histórico | `movimentacoes-panel.tsx:51`: ordenação decrescente, crescente e `reverse()` | Mortos por `movimentacoes-panel.test.tsx:272` e `:284` | Fechada |
| O02 — `identity:role:assign` × `identity:admin:revoke` trocadas | `features/identity/usuarios/actions.ts:34-35` | 16 testes falham: `actions.test.ts:82-100` e `users-page.permissions.test.tsx:137-151`, `:168-178` | Fechada |
| O03 — "Novo lançamento" por `lancamento:update` | `lancamentos-page.tsx:50` | Morto por `lancamentos-page.test.tsx:304` e `:310` | Fechada |
| O04 — "Anexar comprovante" por `lancamento:create` | `comprovantes-panel.tsx:58` | Morto por `comprovantes-panel.test.tsx:317`, `:327`, `:336`, `:341` | Fechada |
| O05 — entrada/saída por `movimentacao:adjust` | `features/estoque/produto-detail.tsx:67` | Morto por `movimentacoes-panel.test.tsx:317`, `:327`, `:339`, `:349`, `:370` | Fechada |
| O06 — ajuste por `movimentacao:create` | `produto-detail.tsx:108` | Morto por `ajuste-dialog.test.tsx:148`, `:163`, `:170` e `movimentacoes-panel.test.tsx:317-356` | Fechada |
| O07 — "Devolver" por `movimentacao:adjust` | `movimentacoes-panel.tsx:23` e `:65`: os dois pontos, só a coluna, só o `RequirePermission` | As três variantes morrem em `movimentacoes-panel.test.tsx:317` e `:349` (e `:327`, `:339` nas duas primeiras) | Fechada |
| O08 — rota `/financeiro/lancamentos` por `conta:read` | `app/(app)/financeiro/lancamentos/page.tsx:13` | Morto por `lancamentos-page.test.tsx:317` (duas sessões) e `:326` | Fechada |
| O09 — rota `/financeiro/lancamentos/[id]` por `conta:read` | `app/(app)/financeiro/lancamentos/[id]/page.tsx:14` | Morto por `lancamento-detail.test.tsx:207` (duas sessões) e `:218` | Fechada |

---

## Discrimination Sensor

Worktree temporário `C:\Users\Michels\Desktop\tjw\vsensor3` (`git worktree add --detach … b57ee9e`), `pnpm install --frozen-lockfile`, `API_URL=http://localhost:8080`. Uma falha de comportamento por vez no código de produção, `pnpm test` completo (845 testes), leitura dos testes que falharam, `git checkout -- .` e conferência de `git status --porcelain` vazio depois de cada mutante. Total: 118 mutantes. Nenhuma morte foi por timeout (0 ocorrências de "timed out" nos logs); a única falha sem relação com o mutante está registrada em "Lacunas por severidade", como observação.

### Dez falhas obrigatórias (INT-04 AC3), L2 e O01 a O09

| # | Falha injetada | Arquivo:linha | Resultado | Teste que matou (primeiro; +N = outros que também falharam) |
| --- | --- | --- | --- | --- |
| M01 | remove X-CSRF-Token das escritas | `lib/api/client.ts:69` | ✅ Morto | `lib/api/client.test.ts` — X-CSRF-Token > vai em POST com o token da sessão atual (+30) |
| M02a | 401 nao redireciona (so limpa a sessao) | `lib/session/session-provider.tsx:145` | ✅ Morto | `lib/session/session-provider.test.tsx` — 401 numa chamada autenticada > limpa cache e contexto e leva a /entrar com aviso e o caminho atual em next (+2) |
| M02b | cliente nao avisa 401 (onUnauthenticated nunca chamado) | `lib/api/client.ts:81` | ✅ Morto | `lib/api/client.test.ts` — callbacks de 401 e 403 > 401 numa chamada autenticada chama onUnauthenticated (+3) |
| M03 | RequirePermission modo action exibe sem a permissao | `components/app/require-permission.tsx:26` | ✅ Morto | `components/app/require-permission.test.tsx` — modo ação > sem a permissão, não renderiza nada (+46) |
| M04 | safeNext aceita next externo | `lib/session/routes.ts:45` | ✅ Morto | `lib/session/routes.test.ts` — safeNext > recusa outra origem sem esquema e usa a página inicial (+17) |
| M05a | lancamento envia valor bruto em reais com casas decimais | `features/financeiro/lancamentos/lancamento-form.tsx:135` | ✅ Morto | `features/financeiro/lancamentos/lancamento-detail.test.tsx` — edição > edita um lançamento CRIADA com PUT, valores em centavos e CSRF (+3) |
| M05b | parseBRL devolve reais com casas decimais | `lib/money.ts:45` | ✅ Morto | `lib/money.test.ts` — parseBRL (mesmos vetores do Go) > lê 1.234,56 como 123456 (+22) |
| M06a | ConfirmDialog habilita confirmar com motivo obrigatorio vazio (cancelamento) | `components/app/confirm-dialog.tsx:80`, `features/financeiro/lancamentos/lancamento-detail.tsx:208`, `features/financeiro/lancamentos/lancamento-detail.tsx:210` | ✅ Morto | `components/app/confirm-dialog.test.tsx` — motivo obrigatório > desabilitado com motivo vazio (+4) |
| M06b | ajuste de estoque confirma com motivo vazio | `features/estoque/ajustes/ajuste-dialog.tsx:43` | ✅ Morto | `features/estoque/ajustes/ajuste-dialog.test.tsx` — confirmação desabilitada > motivo vazio: desabilitada e a API não é chamada (+1) |
| M06c | so o ConfirmDialog deixa de exigir o motivo (a guarda do cancelamento fica) | `components/app/confirm-dialog.tsx:80` | ✅ Morto | `components/app/confirm-dialog.test.tsx` — motivo obrigatório > desabilitado com motivo vazio (+4) |
| M07a | desativar conta sem confirmacao | `features/financeiro/contas/contas-page.tsx:49` | ✅ Morto | `features/financeiro/contas/contas-page.test.tsx` — desativar conta > só chama a API depois de confirmar e mostra a conta como inativa (+2) |
| M07b | receber/pagar sem confirmacao (botao chama a API direto) | `features/financeiro/lancamentos/lancamento-detail.tsx:174` | ✅ Morto | `features/financeiro/lancamentos/workflow.test.tsx` — receber > confirma, chama /receive com CSRF, mostra Recebida e invalida o saldo (+3) |
| M07c | desativar usuario sem confirmacao (dialogo confirma sozinho ao abrir) | `features/identity/usuarios/dialogs/status-dialogs.tsx:3`, `features/identity/usuarios/dialogs/status-dialogs.tsx:34` | ✅ Morto | `features/identity/usuarios/dialogs/status-dialogs.test.tsx` — desativar > pede confirmação avisando do fim das sessões e, em 204, mostra inativo |
| M08a | movimentacao manual envia origem reservada VENDA | `features/estoque/movimentacoes/payload.ts:21` | ✅ Morto | `features/estoque/movimentacoes/movimentacoes-panel.test.tsx` — registro de movimentações > entrada de 10, saída de 4, devolução da saída; corpo sempre com INVENTARIO (+4) |
| M08b | dialogo de movimentacao oferece origens reservadas num seletor | `features/estoque/movimentacoes/movimentacao-dialog.tsx:128` | ✅ Morto | `features/estoque/movimentacoes/movimentacoes-panel.test.tsx` — registro de movimentações > entrada de 10, saída de 4, devolução da saída; corpo sempre com INVENTARIO |
| M09a | senha temporaria gravada no QueryClient por setQueryData | `features/identity/usuarios/dialogs/temporary-password-dialog.tsx:3`, `features/identity/usuarios/dialogs/temporary-password-dialog.tsx:43`, `features/identity/usuarios/dialogs/temporary-password-dialog.tsx:65` | ✅ Morto | `features/identity/usuarios/dialogs/temporary-password-dialog.test.tsx` — geração e descarte > mostra a senha uma vez, copia, e ao fechar some do DOM, do cache, do armazenamento e do log |
| M09b | senha temporaria obtida por useMutation (fica no cache de mutacoes) | `features/identity/usuarios/dialogs/temporary-password-dialog.tsx:3`, `features/identity/usuarios/dialogs/temporary-password-dialog.tsx:43`, `features/identity/usuarios/dialogs/temporary-password-dialog.tsx:65` | ✅ Morto | `features/identity/usuarios/dialogs/temporary-password-dialog.test.tsx` — geração e descarte > mostra a senha uma vez, copia, e ao fechar some do DOM, do cache, do armazenamento e do log |
| M10 | token de recuperacao mantido na URL (sem replaceState) | `features/identity/acesso/reset-password-form.tsx:51` | ✅ Morto | `features/identity/acesso/reset-password-form.test.tsx` — abrir o link com token > remove o fragmento da barra de endereço sem recarregar e mostra o formulário (+5) |
| L2a | URL assinada gravada no QueryClient por setQueryData | `features/financeiro/comprovantes/comprovantes-panel.tsx:3`, `features/financeiro/comprovantes/comprovantes-panel.tsx:38`, `features/financeiro/comprovantes/comprovantes-panel.tsx:47` | ✅ Morto | `features/financeiro/comprovantes/comprovantes-panel.test.tsx` — baixar > não guarda a URL assinada no cache de consultas nem no de mutações |
| L2b | URL assinada obtida por useMutation (cache de mutacoes) | `features/financeiro/comprovantes/comprovantes-panel.tsx:3`, `features/financeiro/comprovantes/comprovantes-panel.tsx:38`, `features/financeiro/comprovantes/comprovantes-panel.tsx:47` | ✅ Morto | `features/financeiro/comprovantes/comprovantes-panel.test.tsx` — baixar > não guarda a URL assinada no cache de consultas nem no de mutações |
| L2c | URL assinada obtida por useQuery/fetchQuery (cache de consultas) | `features/financeiro/comprovantes/comprovantes-panel.tsx:3`, `features/financeiro/comprovantes/comprovantes-panel.tsx:38`, `features/financeiro/comprovantes/comprovantes-panel.tsx:47` | ✅ Morto | `features/financeiro/comprovantes/comprovantes-panel.test.tsx` — baixar > não guarda a URL assinada no cache de consultas nem no de mutações |
| O01a | historico reordenado por data (mais novo primeiro) | `features/estoque/movimentacoes/movimentacoes-panel.tsx:51` | ✅ Morto | `features/estoque/movimentacoes/movimentacoes-panel.test.tsx` — ordem do histórico > mostra as linhas exatamente na ordem da resposta, sem reordenar por data (+1) |
| O01b | historico reordenado por data (mais antigo primeiro) | `features/estoque/movimentacoes/movimentacoes-panel.tsx:51` | ✅ Morto | `features/estoque/movimentacoes/movimentacoes-panel.test.tsx` — ordem do histórico > mostra as linhas exatamente na ordem da resposta, sem reordenar por data (+1) |
| O01c | historico invertido (reverse) | `features/estoque/movimentacoes/movimentacoes-panel.tsx:51` | ✅ Morto | `features/estoque/movimentacoes/movimentacoes-panel.test.tsx` — histórico > chama o histórico e mostra todas as colunas, na ordem da API (+9) |
| O02 | permissoes de Editar papeis e Retirar acesso trocadas | `features/identity/usuarios/actions.ts:34` | ✅ Morto | `features/identity/usuarios/actions.test.ts` — availableActions > identity:role:assign > sozinha (com vínculo, ativo): libera só a sua ação (+15) |
| O03 | Novo lancamento liberado por financeiro:lancamento:update | `features/financeiro/lancamentos/lancamentos-page.tsx:50` | ✅ Morto | `features/financeiro/lancamentos/lancamentos-page.test.tsx` — permissões > leituras + só lancamento:create vê Novo lançamento (+1) |
| O04 | Anexar comprovante liberado por financeiro:lancamento:create | `features/financeiro/comprovantes/comprovantes-panel.tsx:58` | ✅ Morto | `features/financeiro/comprovantes/comprovantes-panel.test.tsx` — permissões > comprovante:read + só comprovante:create pode anexar (+3) |
| O05 | entrada/saida liberadas por estoque:movimentacao:adjust | `features/estoque/produto-detail.tsx:67` | ✅ Morto | `features/estoque/ajustes/ajuste-dialog.test.tsx` — isolamento de estoque:movimentacao:adjust > todas as permissões menos estoque:movimentacao:adjust: sem Ajustar estoque (+5) |
| O06 | ajuste liberado por estoque:movimentacao:create | `features/estoque/produto-detail.tsx:108` | ✅ Morto | `features/estoque/ajustes/ajuste-dialog.test.tsx` — isolamento de estoque:movimentacao:adjust > leitura do saldo + só estoque:movimentacao:adjust: ajusta e envia o POST /ajustes (+6) |
| O07a | Devolver liberado por estoque:movimentacao:adjust (useCan e RequirePermission) | `features/estoque/movimentacoes/movimentacoes-panel.tsx:23`, `features/estoque/movimentacoes/movimentacoes-panel.tsx:65` | ✅ Morto | `features/estoque/movimentacoes/movimentacoes-panel.test.tsx` — isolamento das permissões de escrita no detalhe > leituras + só estoque:movimentacao:create: entrada, saída e Devolver, sem Ajustar (+3) |
| O07b | Devolver: so a coluna (useCan) passa a adjust | `features/estoque/movimentacoes/movimentacoes-panel.tsx:23` | ✅ Morto | `features/estoque/movimentacoes/movimentacoes-panel.test.tsx` — isolamento das permissões de escrita no detalhe > leituras + só estoque:movimentacao:create: entrada, saída e Devolver, sem Ajustar (+3) |
| O07c | Devolver: so o RequirePermission passa a adjust | `features/estoque/movimentacoes/movimentacoes-panel.tsx:65` | ✅ Morto | `features/estoque/movimentacoes/movimentacoes-panel.test.tsx` — isolamento das permissões de escrita no detalhe > leituras + só estoque:movimentacao:create: entrada, saída e Devolver, sem Ajustar (+1) |
| O08 | rota /financeiro/lancamentos protegida por financeiro:conta:read | `app/(app)/financeiro/lancamentos/page.tsx:13` | ✅ Morto | `features/financeiro/lancamentos/lancamentos-page.test.tsx` — permissões > sem lancamento:read, só com as leituras vizinhas: Sem acesso e nenhuma consulta (+2) |
| O09 | rota /financeiro/lancamentos/[id] protegida por financeiro:conta:read | `app/(app)/financeiro/lancamentos/[id]/page.tsx:14` | ✅ Morto | `features/financeiro/lancamentos/lancamento-detail.test.tsx` — permissões > sem lancamento:read, só com as leituras vizinhas: Sem acesso e nenhuma consulta (+2) |

### Mutantes próprios do verificador

Escolhidos onde havia suspeita de teste fraco, com ênfase em frentes não corrigidas neste ciclo (plano de contas e saldo, senha, shell e navegação, `/inicio`) e em ordenações e filtros.

| # | Falha injetada | Arquivo:linha | Resultado | Teste que matou (primeiro; +N = outros que também falharam) |
| --- | --- | --- | --- | --- |
| P01 | rota /financeiro/contas protegida por financeiro:conta:create | `app/(app)/financeiro/contas/page.tsx:13` | ✅ Morto | `app/(app)/financeiro/contas/page.test.tsx` — /financeiro/contas > com financeiro:conta:read, chama GET /contas e mostra o plano |
| P02 | rota /financeiro/saldo protegida por financeiro:conta:read | `app/(app)/financeiro/saldo/page.tsx:13` | ✅ Morto | `app/(app)/financeiro/saldo/page.test.tsx` — /financeiro/saldo > com financeiro:saldo:read, chama GET /saldo e mostra o valor (+1) |
| P03 | Renomear conta liberado por financeiro:conta:create | `features/financeiro/contas/contas-page.tsx:149` | ✅ Morto | `features/financeiro/contas/contas-page.test.tsx` — ações por permissão > financeiro:conta:create mostra só a ação correspondente (+1) |
| P04 | Desativar conta liberado por financeiro:conta:update | `features/financeiro/contas/contas-page.tsx:160` | ✅ Morto | `features/financeiro/contas/contas-page.test.tsx` — ações por permissão > financeiro:conta:update mostra só a ação correspondente (+1) |
| P05 | Nova subconta liberada por financeiro:conta:update (Nova conta fica em create) | `features/financeiro/contas/contas-page.tsx:139` | ✅ Morto | `features/financeiro/contas/contas-page.test.tsx` — ações por permissão > financeiro:conta:create mostra só a ação correspondente (+1) |
| P06 | Desativar oferecido tambem em conta inativa | `features/financeiro/contas/contas-page.tsx:159` | ✅ Morto | `features/financeiro/contas/contas-page.test.tsx` — ações por permissão > conta já inativa não oferece desativar |
| P07 | subconta nao fixa o tipo do pai (radio habilitado e tipo do formulario) | `features/financeiro/contas/conta-dialogs.tsx:107`, `features/financeiro/contas/conta-dialogs.tsx:69` | ✅ Morto | `features/financeiro/contas/contas-page.test.tsx` — criar conta > subconta fixa o tipo do pai e envia parent_id |
| P08 | item Contas da navegacao exige financeiro:lancamento:read | `features/financeiro/nav.ts:10` | ✅ Morto | `components/app/nav-items.test.ts` — navItems > mantém os caminhos e as permissões definidos pelos módulos (+2) |
| P09 | arvore de contas ignora parent_id (tudo na raiz) | `features/financeiro/contas/tree.ts:25` | ✅ Morto | `features/financeiro/contas/contas-page.test.tsx` — árvore do plano de contas > mostra as contas em árvore por parent_id, com tipo e situação (+5) |
| P10 | dica de tamanho minimo invertida (10 sem vinculo, 8 com vinculo) | `features/identity/acesso/errors.ts:19` | ✅ Morto | `features/identity/acesso/change-password-form.test.tsx` — dica de tamanho mínimo > 10 caracteres para quem tem vínculo administrativo (+3) |
| P11 | troca de senha nao rebusca /auth/me | `features/identity/acesso/use-change-password.ts:38` | ✅ Morto | `features/identity/acesso/change-password-form.test.tsx` — troca obrigatória com sucesso > rebusca /auth/me e leva à página inicial |
| P12 | troca voluntaria tambem redireciona | `features/identity/acesso/change-password-form.tsx:79` | ✅ Morto | `features/identity/acesso/change-password-form.test.tsx` — troca voluntária com sucesso > envia senha atual e nova senha e confirma o sucesso |
| P13 | invalid_current_password vai para o campo de nova senha | `features/identity/acesso/errors.ts:32` | ✅ Morto | `features/identity/acesso/change-password-form.test.tsx` — erros de senha da API no campo certo > 403 invalid_current_password |
| P14 | token de recuperacao gravado em sessionStorage | `features/identity/acesso/reset-password-form.tsx:50` | ✅ Morto | `features/identity/acesso/reset-password-form.test.tsx` — abrir o link com token > remove o fragmento da barra de endereço sem recarregar e mostra o formulário (+5) |
| P15 | token de recuperacao enviado tambem em query string | `features/identity/acesso/reset-password-form.tsx:60` | ✅ Morto | `features/identity/acesso/reset-password-form.test.tsx` — redefinição com sucesso > envia token e nova senha no corpo e leva a /entrar com aviso |
| P16 | pedido de recuperacao mostra texto fixo do front, nao o da API | `features/identity/acesso/request-reset-form.tsx:58` | ✅ Morto | `features/identity/acesso/request-reset-form.test.tsx` — pedido de recuperação > envia o e-mail e mostra a mensagem neutra devolvida pela API (+2) |
| P17 | modo obrigatorio sem o botao Sair | `features/identity/acesso/change-password-form.tsx:162` | ✅ Morto | `features/identity/acesso/change-password-form.test.tsx` — modo obrigatório > explica a obrigatoriedade, não oferece navegação e mantém Sair |
| P18 | shell mostra navegacao com troca de senha obrigatoria | `components/app/app-shell.tsx:37` | ✅ Morto | `components/app/app-shell.test.tsx` — com troca de senha obrigatória > esconde a navegação e mantém só Sair |
| P19 | visibleItems ignora a permissao do item | `components/app/nav.ts:24` | ✅ Morto | `components/app/app-shell.test.tsx` — navegação por permissão > mostra o nome da pessoa, o conteúdo e só os itens com permissão (conjunto 1) (+27) |
| P20 | visibleItems mantem secoes vazias | `components/app/nav.ts:36` | ✅ Morto | `components/app/nav.test.ts` — visibleItems > sem permissões, mantém só os itens sem permissão exigida e remove seções vazias (+19) |
| P21 | layout autenticado passa lista vazia ao shell | `app/(app)/layout.tsx:34` | ✅ Morto | `app/(app)/inicio/page.test.tsx` — perfil TESOURARIA > a navegação do shell mostra exatamente os itens permitidos, na ordem (+4) |
| P22 | useCan aceita qualquer uma das permissoes (some em vez de every) | `lib/session/session-provider.tsx:236` | ✅ Morto | `components/app/require-permission.test.tsx` — modo página > com várias permissões exigidas, falta de uma basta para negar (+1) |
| P23 | portao mostra o conteudo enquanto a sessao carrega | `app/(app)/layout.tsx:34` | ✅ Morto | `app/(app)/layout.test.tsx` — enquanto consulta a sessão, mostra carregando e nada do conteúdo (+11) |
| P24 | portao nao barra troca obrigatoria | `app/(app)/layout.tsx:34` | ✅ Morto | `app/(app)/layout.test.tsx` — com sessão > com troca obrigatória fora de /conta/senha, leva a /conta/senha sem mostrar o conteúdo |
| P25 | ordem da navegacao global trocada (Estoque antes de Financeiro) | `components/app/nav-items.ts:13` | ✅ Morto | `components/app/nav-items.test.ts` — navItems > segue com Financeiro, Estoque e Administração, nessa ordem (+4) |
| P26 | logout nao limpa o cache | `lib/session/session-provider.tsx:197` | ✅ Morto | `lib/session/session-provider.test.tsx` — logout > chama POST /auth/logout, limpa todo o cache e leva a /entrar |
| P27 | csrf_invalid nao rebusca /auth/me | `lib/session/session-provider.tsx:154` | ✅ Morto | `lib/session/session-provider.test.tsx` — 403 csrf_invalid > rebusca /auth/me, pede para repetir e não repete a escrita |
| P28 | token CSRF copiado para sessionStorage no login | `lib/session/session-provider.tsx:175` | ✅ Morto | `lib/session/session-provider.test.tsx` — token CSRF e contexto só em memória > não grava nada em localStorage, sessionStorage nem cookie |
| P29 | pessoa ja autenticada em /entrar vai para next | `app/(public)/entrar/login-form.tsx:55` | ✅ Morto | `app/(public)/entrar/login-form.test.tsx` — pessoa já autenticada em /entrar > vai para a página inicial |
| P30 | login ignora must_change_password | `app/(public)/entrar/login-form.tsx:62` | ✅ Morto | `app/(public)/entrar/login-form.test.tsx` — troca de senha obrigatória > leva a /conta/senha mesmo com next |
| P31 | 403 forbidden vira mensagem generica | `components/app/api-error.tsx:16` | ✅ Morto | `components/app/api-error.test.tsx` — apiErrorMessage > 403 forbidden vira a mensagem de permissão (+2) |
| P32 | /inicio mostra atalhos sem filtrar por permissao | `app/(app)/inicio/atalhos.tsx:15` | ✅ Morto | `app/(app)/inicio/page.test.tsx` — perfil TESOURARIA > /inicio mostra um atalho para cada área visível, agrupado pelo módulo (+5) |
| P33 | /inicio nunca mostra a mensagem de conta sem modulos | `app/(app)/inicio/atalhos.tsx:17` | ✅ Morto | `app/(app)/inicio/page.test.tsx` — perfil ASSOCIADO > /inicio mostra um atalho para cada área visível, agrupado pelo módulo (+1) |
| P34 | lancamentos do mais antigo para o mais novo | `features/financeiro/lancamentos/filters.ts:47` | ✅ Morto | `features/financeiro/lancamentos/filters.test.ts` — ordenação > sem filtros, devolve todos do mais novo para o mais antigo (+6) |
| P35 | lancamentos sem ordenacao (ordem da API) | `features/financeiro/lancamentos/filters.ts:47` | ✅ Morto | `features/financeiro/lancamentos/filters.test.ts` — ordenação > sem filtros, devolve todos do mais novo para o mais antigo (+3) |
| P36 | filtro de periodo exclui a data inicial | `features/financeiro/lancamentos/filters.ts:42` | ✅ Morto | `features/financeiro/lancamentos/filters.test.ts` — filtros > por período de criação, com as duas pontas inclusivas (+1) |
| P37 | filtro de periodo usa UTC | `features/financeiro/lancamentos/filters.ts:19` | ✅ Morto | `features/financeiro/lancamentos/filters.test.ts` — filtros > o período usa a data no fuso de São Paulo |
| P38 | filtro de conta ignorado | `features/financeiro/lancamentos/filters.ts:39` | ✅ Morto | `features/financeiro/lancamentos/filters.test.ts` — filtros > por conta (+1) |
| P39 | comprovantes do mais antigo para o mais novo | `features/financeiro/comprovantes/comprovantes-panel.tsx:32` | ✅ Morto | `features/financeiro/comprovantes/comprovantes-panel.test.tsx` — lista > lista nome, tipo, tamanho, versão, autor e data, do mais novo para o mais antigo |
| P40 | contas oferecidas incluem inativas | `features/financeiro/lancamentos/filters.ts:53` | ✅ Morto | `features/financeiro/lancamentos/filters.test.ts` — contas oferecidas num lançamento > só contas ativas do tipo de receita, por nome (+2) |
| P41 | busca de produtos so por nome | `features/estoque/produtos/search.ts:14` | ✅ Morto | `features/estoque/produtos/produtos-page.test.tsx` — listagem de produtos > busca por código ou por nome, no cliente (+1) |
| P42 | filtro de usuarios nao envia role | `features/identity/usuarios/hooks.ts:34` | ✅ Morto | `features/identity/usuarios/users-page.test.tsx` — filtros > filtro de papel refaz a consulta com role e mostra só o resultado da API (+1) |
| P43 | filtro de usuarios fora da chave (nao recomeca do inicio) | `features/identity/usuarios/hooks.ts:25` | ✅ Morto | `features/identity/usuarios/users-page.test.tsx` — filtros > filtro de papel refaz a consulta com role e mostra só o resultado da API (+2) |
| P44 | limite de upload exclusivo (< em vez de <=) | `features/financeiro/comprovantes/limits.ts:22` | ✅ Morto | `features/financeiro/comprovantes/comprovantes-panel.test.tsx` — limite de tamanho > exatamente no limite (10.420.224 bytes), envia |
| P45 | limite de upload de 10 MiB cheios | `features/financeiro/comprovantes/limits.ts:6` | ✅ Morto | `features/financeiro/comprovantes/comprovantes-panel.test.tsx` — limite de tamanho > o limite é 10 MiB − 64 KiB (10.420.224 bytes) (+1) |
| P46 | motivo minimo sem aparar espacos | `features/identity/usuarios/dialogs/reason.ts:10` | ✅ Morto | `features/identity/usuarios/dialogs/admin-dialogs.test.tsx` — promover > mantém o envio desabilitado sem papel ou com motivo de menos de 10 caracteres úteis (+1) |
| P47 | editar papeis vazio sem aviso de vinculo dormente | `features/identity/usuarios/dialogs/admin-dialogs.tsx:161` | ✅ Morto | `features/identity/usuarios/dialogs/admin-dialogs.test.tsx` — editar papéis > lista vazia: avisa que o vínculo fica dormente antes de enviar |
| P48 | acoes aparecem na propria linha | `features/identity/usuarios/actions.ts:29` | ✅ Morto | `features/identity/usuarios/actions.test.ts` — availableActions > própria linha > nenhuma ação, mesmo com todas as permissões (sem vínculo, ativo) (+20) |
| P49 | senha temporaria liberada por identity:user:update | `features/identity/usuarios/actions.ts:36` | ✅ Morto | `features/identity/usuarios/actions.test.ts` — availableActions > identity:user:reset_password > sozinha (sem vínculo, ativo): libera só a sua ação (+28) |
| P50 | Novo usuario liberado por identity:user:update | `features/identity/usuarios/users-page.tsx:87` | ✅ Morto | `features/identity/usuarios/users-page.permissions.test.tsx` — permissão de cada ação, isolada > leitura + só identity:user:create: aparece só o que ela libera (+10) |
| P51 | receber nao invalida o saldo | `features/financeiro/lancamentos/hooks.ts:84` | ✅ Morto | `features/financeiro/lancamentos/workflow.test.tsx` — receber > confirma, chama /receive com CSRF, mostra Recebida e invalida o saldo (+2) |
| P52 | movimentacao nao invalida o saldo do produto | `features/estoque/movimentacoes/hooks.ts:33` | ✅ Morto | `features/estoque/ajustes/ajuste-dialog.test.tsx` — ajuste com sucesso > −3 num produto com saldo 0: avisa o negativo, envia e atualiza saldo e histórico (+2) |
| P53 | painel de comprovantes liberado por financeiro:lancamento:read | `features/financeiro/lancamentos/lancamento-detail.tsx:128` | ✅ Morto | `features/financeiro/comprovantes/comprovantes-panel.test.tsx` — permissões > sem financeiro:comprovante:read não mostra o painel nem consulta comprovantes (+2) |
| P54 | devolucao liberada por financeiro:lancamento:update | `features/financeiro/lancamentos/actions.ts:25` | ✅ Morto | `features/financeiro/lancamentos/actions.test.ts` — todas as combinações de tipo × status × permissão > RECEITA RECEBIDA: devolver com a permissão → true (+3) |
| P55 | saldo de estoque zero destacado como negativo | `features/estoque/produto-detail.tsx:106` | ✅ Morto | `features/estoque/produto-detail.test.tsx` — saldo > saldo zero não é destacado |
| P56 | ajuste com resultado negativo bloqueado | `features/estoque/ajustes/ajuste-dialog.tsx:43` | ✅ Morto | `features/estoque/ajustes/ajuste-dialog.test.tsx` — ajuste com sucesso > −3 num produto com saldo 0: avisa o negativo, envia e atualiza saldo e histórico (+2) |
| P57 | credentials include no cliente | `lib/api/client.ts:114` | ✅ Morto | `lib/api/client.test.ts` — clientes por contrato > chamam a API pela origem atual com credentials same-origin |
| S01 | contas consultadas mesmo sem financeiro:conta:read | `features/financeiro/lancamentos/hooks.ts:47` | ⚠️ Sobreviveu — fora da spec (ver abaixo) | — |
| S02 | X-CSRF-Token tambem enviado em HEAD e OPTIONS | `lib/api/client.ts:15` | ⚠️ Sobreviveu — equivalente (ver abaixo) | — |
| S03 | botao do menu recolhivel visivel em qualquer largura (sem md:hidden) | `components/app/app-shell.tsx:45` | ⚠️ Sobreviveu — visual, G2 (ver abaixo) | — |
| S04 | guarda redundante do cancelamento removida (if (!reason) return) | `features/financeiro/lancamentos/lancamento-detail.tsx:208`, `features/financeiro/lancamentos/lancamento-detail.tsx:210` | ⚠️ Sobreviveu — equivalente (ver abaixo) | — |
| S05 | RequirePermission interno do Devolver removido (redundante com useCan) | `features/estoque/movimentacoes/movimentacoes-panel.tsx:65` | ⚠️ Sobreviveu — equivalente (ver abaixo) | — |
| S06 | 401 de session_expired tratado, unauthenticated ignorado | `lib/api/client.ts:81` | ❌ Sobreviveu — relevante, lacuna L3 | — |
| S07 | devolução não rebusca a lista de lançamentos | `features/financeiro/lancamentos/hooks.ts:133` | ✅ Morto | `features/financeiro/lancamentos/workflow.test.tsx` — devolução > envia devolucao_de_id da receita, conta de despesa, valores e forma, com CSRF |
| S08 | `payload_too_large` sem mensagem própria no envio de comprovante | `features/financeiro/comprovantes/errors.ts:16` | ✅ Morto | `features/financeiro/comprovantes/comprovantes-panel.test.tsx` — erros do envio > 413 payload_too_large |
| S09 | detalhe do lancamento nao mostra autor do cancelamento | `features/financeiro/lancamentos/lancamento-detail.tsx:122` | ✅ Morto | `features/financeiro/lancamentos/lancamento-detail.test.tsx` — detalhe > cancelado: mostra motivo, autor e data do cancelamento |
| S10 | troca obrigatoria ignora next e vai sempre ao inicio | `features/identity/acesso/change-password-form.tsx:80` | ✅ Morto | `features/identity/acesso/change-password-form.test.tsx` — troca obrigatória com sucesso > leva ao destino original pedido em next |
| S11 | reativacao sem aviso de que papeis nao sao restaurados | `features/identity/usuarios/dialogs/status-dialogs.tsx:68` | ✅ Morto | `features/identity/usuarios/dialogs/status-dialogs.test.tsx` — reativar > em 204, mostra ativo e avisa que papéis e vínculo não são restaurados |
| S12 | criacao de usuario sem aviso da troca no primeiro acesso | `features/identity/usuarios/dialogs/create-user-dialog.tsx:84` | ✅ Morto | `features/identity/usuarios/dialogs/create-user-dialog.test.tsx` — criação com sucesso > envia nome, e-mail e senha, avisa da troca no primeiro acesso e mostra a pessoa na lista |
| S06b | 401 de unauthenticated tratado, session_expired ignorado | `lib/api/client.ts:81` | ✅ Morto | `lib/api/client.test.ts` — callbacks de 401 e 403 > 401 numa chamada autenticada chama onUnauthenticated (+2) |
| S06c | sessao so reage a 401 com code session_expired (filtro no provedor) | `lib/session/session-provider.tsx:141` | ❌ Sobreviveu — relevante, lacuna L3 | — |
| S13 | 401 leva a /entrar sem o aviso de sessao encerrada | `lib/session/session-provider.tsx:145` | ✅ Morto | `lib/session/session-provider.test.tsx` — 401 numa chamada autenticada > limpa cache e contexto e leva a /entrar com aviso e o caminho atual em next |
| S14 | 401 leva a /entrar sem next | `lib/session/session-provider.tsx:145` | ✅ Morto | `lib/session/session-provider.test.tsx` — 401 numa chamada autenticada > limpa cache e contexto e leva a /entrar com aviso e o caminho atual em next |
| S15 | 401 redireciona mas nao limpa o cache das outras consultas | `lib/session/session-provider.tsx:145` | ✅ Morto | `lib/session/session-provider.test.tsx` — 401 numa chamada autenticada > limpa cache e contexto e leva a /entrar com aviso e o caminho atual em next |
| S16 | login descarta o Retry-After do 429 | `lib/session/session-provider.tsx:172` | ✅ Morto | `app/(public)/entrar/login-form.test.tsx` — acesso bloqueado > informa o bloqueio e o tempo de espera a partir de Retry-After |
| S17 | troca de senha descarta o Retry-After do 429 | `features/identity/acesso/use-change-password.ts:24` | ✅ Morto | `features/identity/acesso/change-password-form.test.tsx` — troca bloqueada por excesso de tentativas > informa quanto tempo esperar a partir de Retry-After |
| S18 | Receber chama a mutacao de pagar | `features/financeiro/lancamentos/lancamento-detail.tsx:171` | ✅ Morto | `features/financeiro/lancamentos/workflow.test.tsx` — receber > confirma, chama /receive com CSRF, mostra Recebida e invalida o saldo (+3) |
| S19 | saldo resultante do ajuste calculado com sinal trocado | `features/estoque/ajustes/ajuste-dialog.tsx:44` | ✅ Morto | `features/estoque/ajustes/ajuste-dialog.test.tsx` — ajuste com sucesso > −3 num produto com saldo 0: avisa o negativo, envia e atualiza saldo e histórico (+1) |
| S20 | promocao aceita zero papeis | `features/identity/usuarios/dialogs/admin-dialogs.tsx:100` | ✅ Morto | `features/identity/usuarios/dialogs/admin-dialogs.test.tsx` — promover > mantém o envio desabilitado sem papel ou com motivo de menos de 10 caracteres úteis |
| S21 | senha temporaria sem motivo minimo | `features/identity/usuarios/dialogs/temporary-password-dialog.tsx:48` | ✅ Morto | `features/identity/usuarios/dialogs/temporary-password-dialog.test.tsx` — motivo mínimo > mantém o envio desabilitado com menos de 10 caracteres úteis |
| S22 | portao sem sessao redireciona sem next | `lib/session/session-provider.tsx:202` | ✅ Morto | `app/(app)/layout.test.tsx` — sem sessão > leva a /entrar com o caminho pedido em next, sem mostrar o conteúdo (+1) |
| S23 | 403 password_change_required ignorado pela sessao | `lib/session/session-provider.tsx:150` | ✅ Morto | `lib/session/session-provider.test.tsx` — 403 password_change_required > leva a /conta/senha |
| S24 | filtro de status dos lancamentos ignorado | `features/financeiro/lancamentos/filters.ts:38` | ✅ Morto | `features/financeiro/lancamentos/filters.test.ts` — filtros > por status CRIADA (+6) |
| S25 | desativar conta nao rebusca a lista | `features/financeiro/contas/hooks.ts:58` | ✅ Morto | `features/financeiro/contas/contas-page.test.tsx` — desativar conta > só chama a API depois de confirmar e mostra a conta como inativa |

### Sobreviventes

**Relevantes (lacuna L3)**

- **S06** (`lib/api/client.ts:81`): o cliente deixa de avisar a sessão quando o `401` traz `code = unauthenticated`; só `session_expired` continua levando a `/entrar`. Os 845 testes passam.
- **S06c** (`lib/session/session-provider.tsx:141`): o mesmo filtro, posto no provedor de sessão. Os 845 testes passam.
- Controle: **S06b**, o filtro inverso (ignorar `session_expired`), morre em três testes. Ou seja, os testes discriminam um dos dois códigos que FND-03 AC6 nomeia, e não o outro.

**Equivalentes ou fora da spec** — justificados na seção "Mutantes equivalentes".

**Sensor depth**: P0-full (autenticação, sessão, dinheiro, permissões)
**Result**: 111 de 118 mortos; 5 sobreviventes equivalentes ou fora da spec; 2 sobreviventes relevantes (uma lacuna, L3) — FAIL ❌

**Isolamento**: `git status --porcelain` vazio em `tjw\verify3` antes do sensor e depois dele (fora os arquivos deste relatório), e vazio em `C:\Users\Michels\Desktop\torcida-jovem`. O worktree temporário foi removido ao fim.

---

## Smoke repetido contra a stack real

Stack montada pelo guia `docs/development/web.md`, a partir do worktree de verificação: `docker compose -p tjv3 up -d db`, migrações 1 a 9 antes da API, Garage avulso `dxflrs/garage:v2.4.1` na porta 3900 com chaves geradas na hora, `bootstrap-admin --role PRESIDENTE`, binário de `./cmd/api`, `next start -p 3000` sobre o build do gate (`API_URL=http://localhost:8080`). Docker Engine 28.3.2.

Todas as chamadas foram feitas contra `http://localhost:3000` (pelo rewrite, nunca direto no `:8080`), com `Origin: http://localhost:3000`, o cookie da sessão e `X-CSRF-Token` nas escritas, nos mesmos caminhos e corpos que os hooks de `features/**` montam. Senhas, chaves, cookies e tokens ficaram só em memória e num arquivo temporário fora do repositório, apagado ao fim. Nenhum valor desses aparece aqui.

| Bloco | Verificações | Resultado |
| --- | --- | --- |
| Login | Senha errada → `401 invalid_credentials`. Login `PRESIDENTE` → `200`, `must_change_password=true`, 30 permissões, `csrf_token` presente. `Set-Cookie: tj_session=<redigido>; Path=/; Max-Age=28800; HttpOnly; SameSite=Lax`, sem `Domain`. Antes da troca, leitura → `403 password_change_required`. Troca sem `X-CSRF-Token` → `403 csrf_invalid`; com ele → `204`; `GET /auth/me` → `must_change_password=false`. Escrita com `Origin` externa → `403 origin_not_allowed` | Passou (8 de 8) |
| Contas de teste | `POST /users` → `201`; promoção a `TESOURARIA`, `CONSELHO_FISCAL` e `ESTOQUE_LOJA` → `204`; login com troca obrigatória → `200` e `204`; `GET /auth/me` com o papel (13, 11 e 6 permissões) | Passou (15 de 15) |
| Ciclo do financeiro | Contas `RECEITA`, `DESPESA` e subconta → `201`; renomear → `200`. Receita de 15000 centavos → `201 CRIADA`; edição (20000, taxa 150) → `200`, líquido 19850; receber → `204`; receber de novo → `409 lancamento_nao_pode_ser_recebido`; editar recebida → `409 lancamento_imutavel`. Devolução de 5000 → `201` com `devolucao_de_id`; pagar → `204`. Despesa de 3000 → `201`; cancelar só com espaços → `422 motivo_obrigatorio`; com motivo → `204`, lista com `CANCELADA` e o motivo. Saldo → `14850`. Desativar conta → `204` | Passou |
| Comprovantes | Sem `X-CSRF-Token` → `403 csrf_invalid`. PDF de 200.052 bytes → `201`, `size_bytes` igual, `version=1`; URL assinada → `200`; download com SHA-256 igual. **Arquivo de exatamente 10.420.224 bytes → `201`, `size_bytes=10420224`, download de 10.420.224 bytes com SHA-256 igual ao enviado.** Lista → 2 itens | Passou |
| Estoque (apoio) | Produto → `201`; entrada 10 com `INVENTARIO` → `201`; ajuste −2 com motivo → `201`, origem `AJUSTE_MANUAL`; saldo → `8`. `TESOURARIA` no estoque → `403 forbidden` | Passou |
| `CONSELHO_FISCAL` | `GET /auth/me` sem nenhuma permissão de escrita. Leituras → `200`: contas, lançamentos, saldo, comprovantes, URL do comprovante (e download íntegro), produtos, movimentações, saldo de estoque. Escritas → `403 forbidden`: criar, renomear e desativar conta; criar, editar, receber, pagar, cancelar e devolver lançamento; anexar comprovante; criar produto; movimentação; ajuste. `GET /users` → `403`. Saldo inalterado (14850) | Passou (25 de 25) |
| Sessão e logout | Sem cookie → `401 unauthenticated`. Logout sem `X-CSRF-Token` → `403 csrf_invalid`; com ele → `204` e `Set-Cookie` com `Max-Age=0; HttpOnly; SameSite=Lax`; cookie antigo → `401` | Passou |
| Rotas HTML | `/entrar`, `/inicio`, `/financeiro/lancamentos`, `/financeiro/lancamentos/{id}`, `/financeiro/contas`, `/financeiro/saldo`, `/estoque/produtos`, `/admin/usuarios`, `/conta/senha`, `/recuperar-acesso`, `/redefinir-senha`, `/sem-acesso` → `200` com o `<title>` esperado; `/` → `307` para `/inicio` | Passou (13 de 13) |
| Cabeçalhos pelo rewrite (FND-01 AC1) | Senha temporária → `200` com `Cache-Control: no-store`; e-mail repetido → `409 email_taken` em `application/problem+json`; sexto login errado → `429 login_blocked` com `Retry-After: 899`; recuperação → `202` com a mesma mensagem para e-mail existente e inexistente; token inválido → `400 invalid_reset_token` | Passou (8 de 8) |

Total: 101 verificações (93 na primeira execução, 8 na complementar), 0 falha.

**Não exercitado**: a interação visual no navegador (cliques, renderização, diálogos, redirecionamentos feitos pelo JavaScript, remoção do fragmento `#token=` da barra de endereço, menu recolhível em tela estreita). Não havia navegador automatizado (WEB-D-017). É pendência humana, coberta pelo roteiro de `smoke.md` ("Roteiro para o mantenedor"); não é lacuna de teste, porque cada comportamento tem teste de componente com MSW citado acima. A redefinição com token válido não é executável localmente (provedor de e-mail `log`, INT-03 AC3).

**Desmontagem**: API e `next start` encerrados (portas 8080 e 3000 livres), contêiner Garage removido, `docker compose -p tjv3 down -v` (volume e rede removidos), diretório temporário com chaves, senhas e `garage.toml` apagado. Nenhum `.env` foi criado. Os contêineres preexistentes (parados, de outros projetos) não foram tocados.

---

## Gate Check

Worktree `C:\Users\Michels\Desktop\tjw\verify3`, `web/`, `API_URL=http://localhost:8080`, código de saída real por `cmd /v:on /c "<cmd> & echo EXIT=!ERRORLEVEL!"`.

| Gate | Comando | Resultado |
| --- | --- | --- |
| Instalação | `pnpm install --frozen-lockfile` | `EXIT=0` |
| Web | `pnpm lint` | `EXIT=0` |
| Web | `pnpm exec next typegen` | `EXIT=0` |
| Web | `pnpm exec tsc --noEmit` | `EXIT=0` |
| Web | `pnpm test` | `EXIT=0` — 68 arquivos, 845 testes passaram, 0 falhou, 0 pulado |
| Web | `pnpm build` | `EXIT=0` — 15 rotas |
| Audit | `pnpm audit --audit-level high` | `EXIT=0` — "1 high (1 ignored)" |
| Audit (apoio) | `pnpm why source-map-js` | `EXIT=0` — só `1.2.2` |
| Contract | `pnpm lint:api` | `EXIT=0` — quatro contratos válidos |
| Contract | `pnpm gen:api:check` | `EXIT=0` |
| Backend intacto | `git diff --name-only develop -- api/` | vazio |
| State | `validate_state.py --root <worktree> .specs/features/web` | `EXIT=1`, com um único erro: "validation.md verdict is FAIL". É o esperado enquanto L3 estiver aberta; o relatório em si é aceito (presente, com veredito e citações) |

- **Test count before feature** (`develop`): 2 arquivos de teste em `web/`
- **Test count after feature** (`b57ee9e`): 68 arquivos, 845 testes
- **Skipped tests**: nenhum (`it.skip`, `it.only`, `it.todo` sem ocorrência)
- **Failures**: nenhuma
- **`// SPEC_DEVIATION`**: nenhuma ocorrência em `web/`
- **Fora de `web/`, `docs/` e `.specs/` no intervalo**: só `.github/workflows/ci.yml` (fornece `API_URL` a `next typegen` e `next build`, como previsto em `STATE.md`, "Configuração de CI necessária para o build")

---

## Code Quality

| Princípio | Status |
| --- | --- |
| Minimum code | ✅ |
| Surgical changes (só `web/`, guia, specs e o `API_URL` do CI; `api/` intacto) | ✅ |
| No scope creep (nada da lista "Fora do V1" foi implementado) | ✅ |
| Matches patterns (um cliente tipado, chaves de consulta centralizadas, catálogo de erros por feature) | ✅ |
| Spec-anchored outcome check (valores afirmados são os da spec) | ✅ |
| Per-layer Coverage Expectation (funções puras 1:1; telas com caminho feliz, borda e erro) | ✅ |
| Todo teste mapeia para AC, caso de borda ou premissa da spec | ✅ |
| Diretrizes documentadas seguidas: `CLAUDE.md` (dinheiro em centavos inteiros, sessão por cookie `httpOnly` com CSRF, sem JWT), ADR-010, WEB-D-001 a WEB-D-017 | ✅ |

---

## Lacunas por severidade

### Média

**L3 — FND-03 AC6: a variante `401 unauthenticated` não tem teste discriminante** (nova neste ciclo)

- **AC**: "WHEN qualquer chamada autenticada recebe `401` (`unauthenticated` ou `session_expired`) THEN o sistema SHALL limpar o cache e o contexto de sessão e levar a pessoa a `/entrar`, com aviso de sessão encerrada e o caminho atual em `next`."
- **O que os testes afirmam hoje**: o redirecionamento só com `session_expired` (`lib/session/session-provider.test.tsx:259-279`, resposta em `:267`; `lib/api/client.test.ts:131-137`, resposta em `:132`). O código `unauthenticated` aparece em `session-provider.test.tsx:289`, mas junto de um `session_expired` simultâneo (`:288`), que sozinho já produz o único redirecionamento esperado; e em `client.test.ts:176-181`, onde se afirma que o callback **não** é chamado (callbacks desregistrados).
- **Evidência**: mutantes S06 e S06c sobrevivem (845 de 845 testes passam com `401 unauthenticated` ignorado).
- **Por que é relevante**: `401 unauthenticated` é o que a API real devolve quando o cookie some no meio do uso (observado no smoke deste ciclo: sem cookie → `401 unauthenticated`; item 10 de `smoke.md`). Uma regressão que deixe a pessoa numa tela autenticada sem sessão passaria em todos os gates.
- **O código de produção está correto**: `lib/api/client.ts:80-81` e `lib/session/session-provider.tsx:141-148` reagem a qualquer `401`, sem olhar o `code`. A lacuna é só de teste.
- **Unidade dona provável**: F2 (`lib/session/session-provider.test.tsx`) e F1 (`lib/api/client.test.ts`).
- **Correção sugerida**: parametrizar os dois testes pelos dois códigos. Em `session-provider.test.tsx`, `it.each(["session_expired", "unauthenticated"])` no teste de `:259`, mantendo as afirmações de `pathname`, `next`, `sessao=encerrada`, sessão anônima e cache limpo. Em `client.test.ts`, `it.each` no teste de `:131`, afirmando `onUnauthenticated` chamado uma vez para cada código.
- **Done when**: S06 e S06c morrem; S06b continua morrendo; gates `EXIT=0`.

### Alta, baixa

Nenhuma.

### Observação (não é lacuna)

- Numa execução do sensor feita enquanto a stack do smoke compilava (Go e Docker em paralelo), o teste `features/financeiro/lancamentos/lancamento-detail.test.tsx:62` falhou uma vez por esgotar o prazo de 1 s do `findByRole`, num mutante que não toca aquele código (P01). Repetido sem carga, P01 morreu só pelo teste pertinente, e o teste passou nas outras 117 execuções do sensor e no gate. É sensibilidade a carga da máquina, não defeito do teste nem do código.

---

## Gaps de precisão da spec

Não derrubam o veredito. São pontos em que a spec não fixa um valor que um teste de componente possa afirmar.

- **G1 — FND-01 AC1 (preservação de cabeçalhos)**: o encaminhamento é feito pelo rewrite do Next, não por código do projeto. O teste de unidade afirma origem e destino; a preservação de método, corpo e cabeçalhos só é observável contra a stack real, e foi observada no smoke deste ciclo para todos os cabeçalhos citados no AC.
- **G2 — FND-05 AC2 (ponto de quebra)**: "menor que o ponto de quebra de tablet" é resolvido por classes CSS (`md:hidden`, `hidden md:block`), que o jsdom não avalia. O teste afirma que o menu recolhível existe, tem os mesmos itens e fecha ao escolher; a troca de layout na largura certa é conferência visual (pendência humana).
- **G3 — FND-06 AC5 ("oferecer tentar de novo")**: nas telas de consulta há botão "Tentar de novo" (`ErrorState`, `ApiErrorAlert`). Nos formulários, a indisponibilidade aparece como mensagem do formulário e a nova tentativa é o próprio botão de envio; a spec não diz se isso basta.
- **G4 — textos sem redação fixa**: FND-03 AC3 e ACS-01 AC6 ("informar quanto tempo esperar"), USR-03 AC2/AC3 e USR-04 AC1 ("avisar que…") não fixam a frase. Os testes afirmam a frase implementada e o dado que a spec exige (o tempo de `Retry-After`, o aviso de sessões encerradas).

## Mutantes equivalentes ou fora da spec (justificados)

Sobreviveram e não são lacuna, porque não mudam comportamento observável definido pela spec.

- **S01 — contas consultadas sem `financeiro:conta:read`** (`features/financeiro/lancamentos/hooks.ts:47`): com o mutante, a tela de lançamentos faz um `GET /financeiro/contas` a mais, que a API recusa; a tela mostra o identificador curto da conta nos dois casos (`hooks.ts:57`). FND-04 AC4 trata da rota da própria área (`/financeiro/contas`, coberta em `app/(app)/financeiro/contas/page.test.tsx:40-48`); nenhuma spec define o que a tela de lançamentos faz com as contas quando falta essa leitura. Fora da spec. Sugestão sem peso no veredito: um teste que afirme `contas.count` 0 com só `financeiro:lancamento:read`.
- **S02 — `X-CSRF-Token` também em `HEAD` e `OPTIONS`** (`lib/api/client.ts:15`): FND-01 AC3 exige o cabeçalho nas escritas e é afirmado para POST, PUT, PATCH e DELETE, e a ausência em GET. O contrato não tem operação `HEAD` nem `OPTIONS`, então nenhuma tela as emite. Equivalente.
- **S03 — botão do menu sem `md:hidden`** (`components/app/app-shell.tsx:45`): só muda em que largura o botão aparece, o que o jsdom não avalia. É o gap de precisão G2 e a pendência visual.
- **S04 — guarda `if (!reason) return` do cancelamento removida** (`features/financeiro/lancamentos/lancamento-detail.tsx:208`): redundante com o `ConfirmDialog`, que mantém o botão desabilitado e retorna cedo sem motivo (`components/app/confirm-dialog.tsx:80`, `:88`). Quando a regra do diálogo cai (M06a, M06c), os testes matam. Equivalente.
- **S05 — `RequirePermission` interno do "Devolver" trocado por uma permissão que a tela sempre tem** (`features/estoque/movimentacoes/movimentacoes-panel.tsx:65`): redundante com o `useCan` da linha `:23`, que já remove a coluna inteira. Quando os dois pontos ou qualquer um deles passa a `adjust` (O07a, O07b, O07c), os testes matam. Equivalente.

---

## Lições registradas (`lessons.py`)

Só a falha real nova deste ciclo:

- **L-012** (candidata; sinal `surviving_mutant`; fonte: S06, `web/lib/api/client.ts:81`, FND-03 AC6): "Quando um AC nomeia mais de um code da API para o mesmo comportamento, parametrize o teste por todos os codes nomeados".

Os gaps de precisão G1 a G4 e os mutantes equivalentes não geraram lição: não são falhas.

---

## Pendência humana

- **Smoke visual no navegador**: os dez passos de `.specs/features/web/smoke.md`, "Roteiro para o mantenedor". Não é lacuna de teste nem bloqueia este veredito, mas ninguém viu as telas renderizadas nesta verificação.

---

## Requirement Traceability Update

Nenhuma alteração nas specs: com veredito FAIL, a coluna Status das seis tabelas "Requirement Traceability" continua `In Tasks` e as caixas de T4 ficam abertas.

| Requirement | Status atual | Situação nesta verificação |
| --- | --- | --- |
| FND-03 | In Tasks | ❌ Needs Fix (teste): lacuna L3 no AC6 |
| FND-01, FND-02, FND-04, FND-05, FND-06 | In Tasks | Sem lacuna; passam a Verified quando L3 fechar |
| ACS-01 a ACS-03 | In Tasks | Sem lacuna |
| USR-01 a USR-05 | In Tasks | Sem lacuna |
| FWB-01 a FWB-05 | In Tasks | Sem lacuna |
| EWB-01 a EWB-04 | In Tasks | Sem lacuna |
| INT-01 a INT-03 | In Tasks | Sem lacuna |
| INT-04 | In Tasks | Em andamento (este relatório) |

---

## Summary

**Overall**: ❌ Not Ready — uma lacuna de teste (L3). O código de produção atende às specs em tudo o que foi verificado.

**Spec-anchored check**: 128 de 129 ACs e 13 de 13 casos de borda com o resultado da spec afirmado em teste discriminante; FND-03 AC6 afirmado só para um dos dois códigos que nomeia (L3); 4 gaps de precisão sinalizados.
**Sensor**: 111 de 118 mortos; 5 sobreviventes equivalentes ou fora da spec; 2 sobreviventes relevantes (uma lacuna, L3) — FAIL ❌
**Gate**: 845 testes passaram, 0 falhou; todos os gates `EXIT=0`.
**Smoke**: 101 verificações contra a stack real, 0 falha.

**What works**: o caminho same-origin com cookie, `Origin` e CSRF; login, troca obrigatória, recuperação e logout; administração de usuários; ciclo completo do financeiro com comprovantes até o limite de 10.420.224 bytes; ciclo do estoque; visão somente leitura; navegação e página inicial por permissão.

**Next steps**: fechar L3 com a parametrização dos dois testes (F2 e F1), repetir S06, S06b e S06c e os gates, e então marcar T4 e a rastreabilidade. Este é o terceiro veredito FAIL: pela regra de INT-04 AC5 (no máximo 3 ciclos de correção), a decisão sobre um novo ciclo cabe ao mantenedor. O smoke visual no navegador continua pendente de execução humana.
