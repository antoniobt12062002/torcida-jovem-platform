# Web V1 — Relatório de verificação independente (ciclo 1)

## Validation: Web V1 - FAIL

**Result**: FAIL ❌

**Data**: 2026-10-08
**Ciclo**: 1 (de no máximo 3 ciclos de correção e nova verificação; o ciclo 0 também terminou reprovado)
**Specs**: `.specs/features/web/spec/01-fundacao.md` a `06-integracao.md`
**Intervalo de commits verificado**: `ee1b497..8bf79ec` em `feature/web-v1` (`ee1b497` é a especificação; `develop` = `6382fe1`; 42 commits)
**Verificador**: agente independente (autor ≠ verificador), task T4 de `tasks/06-integracao.md`. Não participou do ciclo 0.
**Fontes de verdade**: as 6 specs, designs, tasks, `STATE.md` (WEB-D-*, correção de WEB-D-009), ADR-010 com as Notas. Toda citação `arquivo:linha` deste relatório vem da leitura do código e dos testes em `8bf79ec`. O relatório do ciclo 0, `.specs/LESSONS.md`, `lessons.json` e relatórios de autores não foram usados como prova. `evidence/f1-rewrite.md` e `smoke.md` serviram só de pista e foram refeitos (seção "Smoke repetido").

Caminhos de código e de teste são relativos a `web/`, salvo indicação.

### Por que a Web V1 não foi aprovada neste ciclo

As duas lacunas do ciclo 0 estão fechadas (seção "Fechamento de L1 e L2"), os gates estão todos em `EXIT=0`, os 10 mutantes mínimos de INT-04 AC3 morreram e o smoke repetido não teve falha. O que reprova é novo: **9 dos 34 mutantes próprios sobreviveram**, em duas classes.

1. **Permissão trocada por outra que os perfis de teste sempre trazem junto** (8 mutantes, O02 a O09). Trocar a permissão que libera uma ação ou uma rota por outra permissão do mesmo módulo não derruba nenhum teste, porque os testes dessas telas só usam "todas as permissões" ou "só leitura". A spec nomeia a permissão de cada ação ("WHEN uma pessoa com `X` ...") e FND-04 AC3 manda decidir a exibição pelas permissões efetivas.
2. **Ordem do histórico de estoque** (O01). Reordenar o histórico por data, em vez de manter a ordem da API (EWB-02 AC1), não derruba nenhum teste, porque todas as movimentações das fixtures têm a mesma data.

Nenhum dos nove indica defeito no código de produção: o código em `8bf79ec` usa a permissão certa em cada ponto e mantém a ordem da API. São lacunas de teste. Com a matriz de papéis atual da API, cada papel tem todas as escritas de um módulo ou só as leituras, então a troca de permissão não muda o que nenhum papel de hoje vê. A regra desta verificação, porém, é que mutante relevante sobrevivente reprova, e a spec existe justamente para a interface não depender do agrupamento por papel. As correções são pequenas e só de teste (seção "Lacunas").

---

## Task Completion

| Sub-spec | Tasks | Status | Observação |
| --- | --- | --- | --- |
| 01 Fundação | T1-T8 | ✅ Concluídas | — |
| 02 Identity acesso | T1-T2 | ✅ Concluídas | — |
| 03 Identity usuários | T1-T3 | ✅ Concluídas | — |
| 04 Financeiro | T1-T5 | ✅ Concluídas | — |
| 05 Estoque | T1-T3 | ✅ Concluídas | — |
| 06 Integração | T1-T3 | ✅ Concluídas | — |
| 06 Integração | T4 | ❌ Não concluída | Veredito reprovado; caixas de T4 não marcadas (`.specs/features/web/tasks/06-integracao.md:117`) |

---

## Fechamento de L1 e L2 (lacunas do ciclo 0)

### L1 — gate Audit: fechada

| Verificação | Observado |
| --- | --- |
| `pnpm audit --audit-level high` | `EXIT=0`; saída: `1 vulnerabilities found / Severity: 1 high (1 ignored)` |
| GHSA-68fv-2mgg-jv7q | Não aparece: 0 ocorrências de `68fv` e de `source-map` no JSON do audit |
| Única exceção ignorada | `GHSA-vfj7-8cjw-p6xm`, em `pnpm-workspace.yaml:10-12`. Conferido por fora: o audit de uma cópia de `package.json` e `pnpm-lock.yaml` **sem** a lista de exceções devolve um único aviso, `GHSA-vfj7-8cjw-p6xm` (`braces`, alta) |
| `pnpm why source-map-js` | `Found 1 version of source-map-js`: `1.2.2`. No lockfile, só `source-map-js@1.2.2` (`pnpm-lock.yaml:3307`, `:7148`) |
| `package.json` | Sem mudança relacionada: no intervalo, só `57cfbad` (F1) altera o arquivo, acrescentando `@tanstack/react-query`, `openapi-fetch`, `react-hook-form` e `msw`. O PR #62 (`6382fe1`) e o merge `8bf79ec` alteram só `pnpm-lock.yaml` (2 inserções, 8 remoções). `pnpm-workspace.yaml` é igual ao de `develop` |

### L2 — FWB-05 AC5, URL assinada fora do cache: fechada

O teste acrescentado em `e2bf7c7` é `features/financeiro/comprovantes/comprovantes-panel.test.tsx:256-271`: depois do clique em "Baixar", serializa os dados de todas as consultas e os dados e variáveis de todas as mutações do `QueryClient` e afirma `expect(cached).not.toContain("assinada-secreta")`.

Recriei três mutantes em `features/financeiro/comprovantes/comprovantes-panel.tsx:42-50`, um por vez, no worktree temporário. Os três morreram pelo teste da linha 256, sem timeout:

| Mutante | Como | Resultado |
| --- | --- | --- |
| L2a | `queryClient.setQueryData(["financeiro","comprovante-url",id], url)` depois de obter a URL (o mutante que sobreviveu no ciclo 0) | ✅ Morto (1 falha, 688 passam) |
| L2b | `useMutation({ mutationFn: downloadComprovante })` e `mutateAsync` no clique | ✅ Morto (1 falha, 688 passam) |
| L2c | `queryClient.fetchQuery({ queryKey, queryFn })` no clique (consulta, como `useQuery`) | ✅ Morto (1 falha, 688 passam) |

---

## Spec-Anchored Acceptance Criteria

Legenda: ✅ o teste afirma o resultado que a spec define; ❌ **LACUNA** falta asserção discriminante (mutante sobrevivente citado); ⚠️ a spec não define o resultado com precisão; 🔎 aceite manual ou por inspeção, conferido por evidência e smoke repetido.

### 01 — Fundação

| AC | Resultado definido pela spec | Teste: `arquivo:linha` e asserção | Código | |
| --- | --- | --- | --- | --- |
| FND-01 AC1 | `/api/v1/*` vai para o mesmo caminho em `API_URL`, preservando cabeçalhos | `next.config.test.ts:13-20` `expect(apiRewrites("http://localhost:8080")).toEqual([{ source: "/api/v1/:path*", destination: "http://localhost:8080/api/v1/:path*" }])`; `:38-47` lê `API_URL` do ambiente. Cabeçalhos: smoke repetido (`Set-Cookie`, `Content-Type: application/problem+json`, `Retry-After=899`, `Cache-Control: no-store`, `Origin` e `X-CSRF-Token` verificados pela API) | `next.config.ts:7-21` | ✅ |
| FND-01 AC2 | `openapi-fetch` sobre os tipos gerados, `credentials: "same-origin"`; nenhuma tela chama `fetch` direto | `lib/api/client.test.ts:49-67` `expect(seen[0].url).toBe(...)`, `expect(seen[0].credentials).toBe("same-origin")`. Cláusula negativa: inspeção (`fetch(` só em `lib/api/client.ts:116`, fora de testes) | `lib/api/client.ts:111-127` | ✅ (ver gap de precisão P3) |
| FND-01 AC3 | Escritas levam `X-CSRF-Token` da sessão atual | `lib/api/client.test.ts:72-108` POST, PUT, PATCH e DELETE `expect(seen[0].csrf).toBe(TOKEN)`; `:110-114` GET `toBeNull()`; `lib/session/session-provider.test.tsx:139` `expect(csrf).toEqual(["token-do-login"])`, `:190` `["token-do-me"]` | `lib/api/client.ts:66-71`; `lib/session/session-provider.tsx:139-140` | ✅ |
| FND-01 AC4 | Erro tipado com `status`, `code`, `errors[]` | `lib/api/client.test.ts:222-227` `toBeInstanceOf(ApiError)` e `toMatchObject({ status: 422, code: "validation_failed", errors: [{ field: "nome", code: "required" }] })`; `lib/api/problem.test.ts:36-44` | `lib/api/client.ts:135-147`; `lib/api/problem.ts:59-68` | ✅ |
| FND-01 AC5 | (a) cookie da origem do front, (b) escrita sem `origin_not_allowed`, (c) comprovante ≥ 9,5 MiB íntegro | Smoke repetido: `Set-Cookie` com `Path=/; Max-Age=28800; HttpOnly; SameSite=Lax`, sem `Domain`; escritas `201`/`204` com `Origin` do front e `403 origin_not_allowed` com `Origin` externa; arquivo de 10.420.224 bytes `201` e download com SHA-256 igual | — | 🔎 ✅ |
| FND-01 AC6 | F1 para e reporta se o AC5 falhar | AC5 passou; o achado do limite de 10 MiB foi reportado e resolvido pelo mantenedor, sem contorno nem mudança na API (`.specs/features/web/STATE.md:152`, `docs/adr/010-acesso-do-navegador-a-api-por-rewrite-same-origin.md:75`); `git diff --name-only develop -- api/` vazio | — | 🔎 ✅ |
| FND-02 AC1 | Kit em `components/ui/` | `components/ui/field.test.tsx:29-37` (campo e rótulo), `:81-95` (área de texto), `:97-112` (caixa de seleção), `:114-135` (seleção); `components/ui/overlays.test.tsx:41-61` (diálogo), `:63-92` (confirmação), `:114-130` (menu); `components/ui/display.test.tsx:19-42` (tabela), `:44-56` (alerta), `:58-65` (esqueleto), `:67-78` (toast) | `components/ui/*.tsx` (15 componentes, com `button.tsx` e `badge.tsx`) | ✅ |
| FND-02 AC2 | Erro ligado por `aria-describedby`; campo com `aria-invalid` | `components/ui/field.test.tsx:48-49` `expect(input.getAttribute("aria-invalid")).toBe("true")`, `expect(describedByText(input)).toContain("Informe um e-mail válido.")` | `components/ui/field.tsx:210-211` | ✅ |
| FND-02 AC3 | F4 não altera `package.json` | `git log ee1b497..8bf79ec -- web/package.json` lista só `57cfbad` (F1); o commit de F4 (`3020783`) não toca `package.json` nem o lockfile | — | 🔎 ✅ |
| FND-03 AC1 | Login, contexto guardado, destino `next` só interno | `app/(public)/entrar/login-form.test.tsx:53-54` `expect(navigations()).toEqual(["/financeiro/contas?tipo=RECEITA"])`, `expect(bodies).toEqual([{ email, password }])`; `:62` `["/inicio"]`; `:65-74` `next` externo vira `/inicio`; `lib/session/routes.test.ts:29-46` | `app/(public)/entrar/login-form.tsx:58-62`; `lib/session/routes.ts:27-46`; `lib/session/session-provider.tsx:161-179` | ✅ (ver P2) |
| FND-03 AC2 | "E-mail ou senha inválidos." sem indicar o campo | `login-form.test.tsx:95-97` `findByText("E-mail ou senha inválidos.")`; `aria-invalid` ausente nos dois campos | `login-form.tsx:27`, `:73` | ✅ |
| FND-03 AC3 | Bloqueio com tempo de `Retry-After` | `login-form.test.tsx:105-110` `Retry-After: 900` → `expect(alert.textContent).toContain("15 minutos")`; `lib/session/retry-after.test.ts:8-20` | `login-form.tsx:29-33`, `:74`; `lib/session/retry-after.ts:19-34` | ✅ |
| FND-03 AC4 | `must_change_password` leva a `/conta/senha`, sem outra tela | `login-form.test.tsx:84` `["/conta/senha"]`; `app/(app)/layout.test.tsx:68-69` `["/conta/senha"]` e conteúdo ausente | `app/(app)/layout.tsx:23-34` | ✅ |
| FND-03 AC5 | Logout: `POST /auth/logout`, cache e contexto limpos, `/entrar` | `lib/session/session-provider.test.tsx:214-224` `["/entrar"]`, `logoutCsrf` igual a `["token-logout"]`, `getQueryData(["financeiro","contas"])` indefinido, `withData` vazio | `lib/session/session-provider.tsx:125-129`, `:181-199` | ✅ |
| FND-03 AC6 | `401`: limpa, vai a `/entrar` com aviso e `next` | `session-provider.test.tsx:272-278` `pathname` `/entrar`, `next` = `/financeiro/contas?tipo=RECEITA`, `sessao` = `encerrada`, estado `anonymous`, cache limpo; `login-form.test.tsx:165-168` aviso | `lib/api/client.ts:80-81`; `session-provider.tsx:141-148`; `login-form.tsx:88-92` | ✅ |
| FND-03 AC7 | `403 password_change_required` leva a `/conta/senha` | `session-provider.test.tsx:366` `expect(navigations()).toEqual(["/conta/senha"])` | `client.ts:82-83`; `session-provider.tsx:149-151` | ✅ |
| FND-03 AC8 | `403 csrf_invalid`: rebusca `/auth/me`, pede repetição, não repete | `session-provider.test.tsx:389-395` `me.count` 2, texto "A sessão foi atualizada. Repita a ação, por favor.", `writes` 1 | `session-provider.tsx:152-156` | ✅ |
| FND-03 AC9 | Token e contexto só em memória | `session-provider.test.tsx:172-175` `localStorage.length` 0, `sessionStorage.length` 0, `document.cookie` inalterado | `session-provider.tsx:94-100`, `:175` | ✅ |
| FND-03 AC10 | Link "Esqueci minha senha" → `/recuperar-acesso` | `login-form.test.tsx:160-161` `expect(link.getAttribute("href")).toBe("/recuperar-acesso")` | `login-form.tsx:128-133` | ✅ |
| FND-04 AC1 | Sem sessão: `/entrar` com `next`, sem conteúdo | `app/(app)/layout.test.tsx:46-50` | `layout.tsx:26-28`; `session-provider.tsx:201-203` | ✅ |
| FND-04 AC2 | Carregando, sem conteúdo protegido | `layout.test.tsx:36-37` `toContain("Carregando")`, conteúdo ausente | `layout.tsx:49-58` | ✅ |
| FND-04 AC3 | Exibição pelas permissões, nunca pelo papel | `session-provider.test.tsx:78-84` papéis `ADMIN_SISTEMA` e `PRESIDENTE` com `permissions: []` → `"false"`; `components/app/nav.test.ts:41-43`; `components/app/require-permission.test.tsx:102-121` | `session-provider.tsx:227-237`; `components/app/nav.ts:23-25`; `components/app/require-permission.tsx:21-31` | ✅ |
| FND-04 AC4 | Sem a leitura da área: "Sem acesso", sem chamar a API | `require-permission.test.tsx:68-72` (`calls.count` 0); `app/(app)/admin/usuarios/page.test.tsx:34-36`; `app/(app)/financeiro/contas/page.test.tsx:44-47`; `app/(app)/financeiro/saldo/page.test.tsx:43-46`; `features/estoque/produtos/produtos-page.test.tsx:191-195`; `features/estoque/produto-detail.test.tsx:82-88`. Rotas de lançamentos: `features/financeiro/lancamentos/lancamentos-page.test.tsx:275-279` e `lancamento-detail.test.tsx:181-185` só testam a ausência com `financeiro:saldo:read` | `require-permission.tsx:26-30`; `app/(app)/financeiro/lancamentos/page.tsx:13`, `[id]/page.tsx:14` | ❌ **LACUNA G6** nas duas rotas de lançamentos (O08, O09); demais áreas ✅ |
| FND-04 AC5 | `403 forbidden`: mensagem fixa, estado mantido | `components/app/api-error.test.tsx:15` `toBe("Você não tem permissão para esta ação.")`; `features/financeiro/contas/contas-page.test.tsx:228-244`, `:341-355` | `components/app/api-error.tsx:15-20` | ✅ |
| FND-05 AC1 | Nome, "Trocar senha", "Sair", navegação por permissão | `components/app/app-shell.test.tsx:59-62`, `:69-73` listas exatas de links; `:105` itens `["Trocar senha","Sair"]` | `components/app/app-shell.tsx:37-38`, `:66-79` | ✅ |
| FND-05 AC2 | Abaixo do ponto de quebra, menu recolhível; nada inacessível | `app-shell.test.tsx:87-97` o menu "Abrir menu" mostra os mesmos links e fecha ao escolher. A troca pela largura é CSS (`md:hidden`, `hidden md:block`) e não é exercitada no jsdom | `app-shell.tsx:43-56`, `:84` | ⚠️ P1 (pendência visual) |
| FND-05 AC3 | Estados de carregamento, vazio e erro com nova tentativa | `components/app/states.test.tsx:14`, `:21-27`, `:35-37` | `components/app/states.tsx:17-72` | ✅ |
| FND-05 AC4 | pt-BR → centavos inteiros, nunca ponto flutuante | `components/app/money-field.test.tsx:40-41` `expect(last).toBe(cents)`, `Number.isSafeInteger(last)`; `:118` `toEqual({ valor: 123456 })` | `components/app/money-field.tsx:14-23`; `lib/money.ts:36-46` | ✅ |
| FND-05 AC5 | Só inteiros, sinal configurável | `components/app/quantity-field.test.tsx:40-63`, `:67-90` | `components/app/quantity-field.tsx:16-33` | ✅ |
| FND-05 AC6 | Confirmação explícita; motivo vazio ou só espaços desabilita | `components/app/confirm-dialog.test.tsx:27-40`; `:76`, `:84` `expect(confirmButton().disabled).toBe(true)` | `components/app/confirm-dialog.tsx:78-98` | ✅ |
| FND-05 AC7 | `formatBRL`; datas pt-BR em São Paulo | `lib/money.test.ts:7-9`; `components/app/format.test.ts:9`, `:14`, `:24-25` | `lib/money.ts:23-29`; `components/app/format.ts:4-42` | ✅ |
| FND-05 AC8 | Troca obrigatória: sem navegação, só "Sair" | `app-shell.test.tsx:132-136` | `app-shell.tsx:37`, `:73-76` | ✅ |
| FND-06 AC1 | `errors[]` junto do campo | `lib/forms/apply-problem.test.ts:33-41`, `:66-71`; `features/estoque/produtos/produtos-page.test.tsx:172-179` | `lib/forms/apply-problem.ts:30-52` | ✅ |
| FND-06 AC2 | Mensagem do catálogo para o `code` | `lib/api/problem.test.ts:63-68`; `api-error.test.tsx:19-23` | `problem.ts:84-86` | ✅ |
| FND-06 AC3 | Genérica com o status, sem detalhes internos | `problem.test.ts:81-84` `toContain("409")` e sem `code`, `title`, `detail` | `problem.ts:89` | ✅ |
| FND-06 AC4 | `413`: conteúdo grande demais | `api-error.test.tsx:35` `toBe("O conteúdo enviado é grande demais.")` | `problem.ts:11`, `:87` | ✅ |
| FND-06 AC5 | Indisponível, com nova tentativa | `problem.test.ts:93-105`; `api-error.test.tsx:46-52`; `app/(app)/layout.test.tsx:93-98` | `problem.ts:76-78`, `:88`; `api-error.tsx:35-39` | ✅ |
| Borda | Dois `401` → um redirecionamento | `session-provider.test.tsx:298` `toHaveLength(1)` | `session-provider.tsx:104-118` | ✅ |
| Borda | `next` de outra origem ignorado | `routes.test.ts:29-46`; `login-form.test.tsx:65-74` | `routes.ts:27-37` | ✅ |
| Borda | Autenticado em `/entrar` → início | `login-form.test.tsx:181` `["/inicio"]` | `login-form.tsx:53-56` | ✅ |

### 02 — Identity: acesso e senha

| AC | Resultado definido pela spec | Teste | Código | |
| --- | --- | --- | --- | --- |
| ACS-01 AC1 | `POST /auth/password` com `current_password` e `new_password`; sucesso confirmado | `features/identity/acesso/change-password-form.test.tsx:75-81` texto "Senha alterada com sucesso." e `expect(requests).toEqual([{ body: { current_password, new_password }, csrf: "csrf-token-de-teste" }])` | `features/identity/acesso/change-password-form.tsx:59-62`, `:106-110` | ✅ |
| ACS-01 AC2 | Confirmação diferente: erro no campo, sem API | `change-password-form.test.tsx:96-101` `expect(requests).toEqual([])` | `change-password-form.tsx:149-152` | ✅ |
| ACS-01 AC3 | Rebusca `/auth/me`, vai ao destino | `:115-117` `["/inicio"]`, `state.calls` 2; `:125` `next` | `features/identity/acesso/use-change-password.ts:38`; `change-password-form.tsx:79-82` | ✅ |
| ACS-01 AC4 | Explica a obrigatoriedade, sem navegação, com "Sair" | `:149-153` | `change-password-form.tsx:94-98`, `:162-166` | ✅ |
| ACS-01 AC5 | Cinco códigos no campo certo | `:184-204` (`it.each`), `isInvalid(field)` e `describedByText(field)` | `features/identity/acesso/errors.ts:31-57` | ✅ |
| ACS-01 AC6 | `429` com tempo de `Retry-After` | `:230-232` "Muitas tentativas. Tente de novo em 15 minutos." | `errors.ts:63-71`; `use-change-password.ts:21-25` | ✅ |
| ACS-02 AC1 | `POST /password-reset/request`; mensagem da API | `features/identity/acesso/request-reset-form.test.tsx:50`, `:75` | `features/identity/acesso/request-reset-form.tsx:22-25`, `:53-62` | ✅ |
| ACS-02 AC2 | Mesma tela para qualquer e-mail | `request-reset-form.test.tsx:59` `expect(second).toBe(first)` | `request-reset-form.tsx:53-62` | ✅ |
| ACS-02 AC3 | Link em `/entrar` | `app/(public)/entrar/login-form.test.tsx:160-161`; smoke: `/recuperar-acesso` `200` | `login-form.tsx:128-133` | ✅ |
| ACS-03 AC1 | Token do fragmento, só em memória, fragmento removido sem recarregar | `features/identity/acesso/reset-password-form.test.tsx:81-85` `hash` vazio, `navigations()` vazio; `read-reset-token.test.ts:7-21` | `features/identity/acesso/reset-password-form.tsx:46-56` | ✅ |
| ACS-03 AC2 | Sem token: link inválido e `/recuperar-acesso` | `reset-password-form.test.tsx:93-97` | `reset-password-form.tsx:98-115` | ✅ |
| ACS-03 AC3 | `POST /confirm` com `token` e `new_password`; `/entrar` com aviso | `:109-117` | `reset-password-form.tsx:58-61`, `:88-93` | ✅ |
| ACS-03 AC4 | `invalid_reset_token`: link expirado e `/recuperar-acesso` | `:157-161` | `reset-password-form.tsx:76-79` | ✅ |
| ACS-03 AC5 | Erro de senha no campo; token mantido | `:184-196` segunda requisição com o mesmo `TOKEN` | `reset-password-form.tsx:80-84` | ✅ |
| ACS-03 AC6 | Token nunca em query, armazenamento ou log | `:65-74`, `:116` `expect(new URL(requests[0].url).search).toBe("")` | `reset-password-form.tsx:43` | ✅ |
| Borda | Autenticado em `/redefinir-senha` | `:121-130` sessão vira `null`, `/entrar` | `reset-password-form.tsx:88-93` | ✅ |
| Borda | `/conta/senha` voluntária | `change-password-form.test.tsx:83` `navigations()` vazio; `app/(app)/conta/senha/page.test.tsx:17-26` | `change-password-form.tsx:83-84` | ✅ |

### 03 — Identity: usuários

| AC | Resultado definido pela spec | Teste | Código | |
| --- | --- | --- | --- | --- |
| USR-01 AC1 | `GET /users`; seis colunas | `features/identity/usuarios/users-page.test.tsx:57-74` `cellsOf(...)` exato e `requests` = `[{ limit: "50" }]` | `features/identity/usuarios/users-table.tsx:58-65`; `hooks.ts:23-43` | ✅ |
| USR-01 AC2 | Filtros enviados; recomeça do início | `users-page.test.tsx:112`, `:131-135`, `:153-157` | `users-page.tsx:62-68`; `hooks.ts:25` | ✅ |
| USR-01 AC3 | "Carregar mais" com `next_cursor` | `:172-177`, `:184` | `users-page.tsx:149-158`; `hooks.ts:41` | ✅ |
| USR-01 AC4 | Sem leitura: "Sem acesso"; item fora da navegação | `app/(app)/admin/usuarios/page.test.tsx:34-36`; `features/identity/nav.test.ts:19` | `app/(app)/admin/usuarios/page.tsx:13`; `features/identity/nav.ts:8` | ✅ |
| USR-02 AC1 | `POST /users`; pessoa na lista; aviso da troca | `features/identity/usuarios/dialogs/create-user-dialog.test.tsx:82-92` | `dialogs/create-user-dialog.tsx:66-68`, `:84` | ✅ |
| USR-02 AC2-AC3 | Erros no campo certo | `create-user-dialog.test.tsx:98-128` | `create-user-dialog.tsx:32-39`, `:71-75` | ✅ |
| USR-02 AC4 | Sem `identity:user:create`: sem "Novo usuário" | `:58-63` | `users-page.tsx:87-90` | ✅ |
| USR-03 AC1 | Desativar; mostra inativo | `dialogs/status-dialogs.test.tsx:88-92` | `dialogs/status-dialogs.tsx:31-51` | ✅ |
| USR-03 AC2 | Reativar; aviso de não restauração | `status-dialogs.test.tsx:110-116` | `status-dialogs.tsx:53-72` | ✅ |
| USR-03 AC3 | Confirmação com aviso das sessões | `:83-86` texto e `requests` vazio antes de confirmar | `status-dialogs.tsx:37-38` | ✅ |
| USR-03 AC4 | Três códigos; estado mantido | `:122-137`, `:139-149` | `status-dialogs.tsx:25-29`; `errors.ts:24-26` | ✅ |
| USR-04 AC1 | Promoção e aviso | `dialogs/admin-dialogs.test.tsx:142-153` corpo `{ roles: ["DIRETORIA","TESOURARIA"], reason }` | `dialogs/admin-dialogs.tsx:102-113` | ✅ |
| USR-04 AC2 | Sem papel ou motivo < 10: desabilitado | `admin-dialogs.test.tsx:114-126` | `admin-dialogs.tsx:100`; `dialogs/reason.ts:9-11` | ✅ |
| USR-04 AC3 | Com `identity:role:assign`: `PUT /roles`; lista vazia avisa antes | `:218-227`, `:237-242` afirmam corpo e aviso. A visibilidade de "Editar papéis" só é testada com as quatro permissões juntas ou sem nenhuma (`:74-90`) | `features/identity/usuarios/actions.ts:34`; `admin-dialogs.tsx:158-174` | ❌ **LACUNA G2** (O02) na permissão que libera a ação; corpo e aviso ✅ |
| USR-04 AC4 | Com `identity:admin:revoke`: retirar acesso | `:259-269` afirmam corpo e resultado. Mesma limitação de `:74-90` | `actions.ts:35`; `admin-dialogs.tsx:214-235` | ❌ **LACUNA G2** (O02); corpo ✅ |
| USR-04 AC5 | Dez códigos em português | `:159-187`, `:189-200`, `:272-285` | `errors.ts:24-34` | ✅ |
| USR-04 AC6 | Ação compatível com o estado | `:76-82` | `actions.ts:31-35` | ✅ |
| USR-05 AC1 | Motivo ≥ 10; senha uma vez, com cópia | `dialogs/temporary-password-dialog.test.tsx:80-86`, `:109-116` | `dialogs/temporary-password-dialog.tsx:48`, `:58-72` | ✅ |
| USR-05 AC2 | Descartada; fora de cache, armazenamento e log | `:126-140` | `temporary-password-dialog.tsx:50-56`; `hooks.ts:129-137` | ✅ |
| USR-05 AC3 | Quatro códigos | `:151-164` | `temporary-password-dialog.tsx:64-65`, `:96` | ✅ |
| Borda | Própria linha sem ações | `status-dialogs.test.tsx:66-67` | `actions.ts:29` | ✅ |
| Borda | Invalidar consultas de usuários | `status-dialogs.test.tsx:90`; `admin-dialogs.test.tsx:147` | `hooks.ts:51-54` | ✅ |

### 04 — Financeiro

| AC | Resultado definido pela spec | Teste | Código | |
| --- | --- | --- | --- | --- |
| FWB-01 AC1 | Árvore por `parent_id`, tipo e situação | `features/financeiro/contas/contas-page.test.tsx:101-113`; `tree.test.ts:26-38` | `contas/contas-page.tsx:87`, `:112-123`; `contas/tree.ts:17-29` | ✅ |
| FWB-01 AC2 | `POST /contas` com `tipo`, `nome`, `parent_id` | `contas-page.test.tsx:183`, `:206-208` | `contas/conta-dialogs.tsx:67-72` | ✅ |
| FWB-01 AC3 | Subconta com tipo fixo | `:197-200` | `conta-dialogs.tsx:69`, `:107` | ✅ |
| FWB-01 AC4 | `PATCH`; `409 conta_ja_utilizada` explicado | `:276-278`, `:284`, `:296` | `contas/hooks.ts:45-52`; `contas/errors.ts:8` | ✅ |
| FWB-01 AC5 | Desativação confirmada | `:327`, `:334-335` | `contas-page.tsx:53-63`, `:96-106` | ✅ |
| FWB-01 AC6 | Mensagens dos dois códigos | `:229-230`, `:285`, `:342` | `contas/errors.ts:9-10` | ✅ |
| FWB-02 AC1 | `GET /saldo`; `formatBRL` com zero e negativo | `features/financeiro/saldo/saldo-page.test.tsx:41-49` "R$ 1.234,56", "R$ 0,00", "-R$ 987,65" | `saldo/saldo-page.tsx:34` | ✅ |
| FWB-02 AC2 | Explicação do saldo | `saldo-page.test.tsx:58-62` | `saldo-page.tsx:39-42` | ✅ |
| FWB-03 AC1 | Lista e filtros no cliente | `features/financeiro/lancamentos/lancamentos-page.test.tsx:90-100`, `:124` `list.count` 1; `filters.test.ts:22-87` | `lancamentos/lancamentos-page.tsx:41-42`, `:123-134`; `lancamentos/filters.ts:31-48` | ✅ |
| FWB-03 AC2 | Com `financeiro:lancamento:create`: `POST` em centavos | `lancamentos-page.test.tsx:180-189` corpo exato. A visibilidade de "Novo lançamento" só é testada com todas as escritas ou só leitura (`:269-273`) | `lancamentos-page.tsx:50`; `lancamentos/lancamento-form.tsx:130-138` | ❌ **LACUNA G3** (O03) na permissão; corpo ✅ |
| FWB-03 AC3 | `PUT` com conta, valores e forma | `lancamentos/lancamento-detail.test.tsx:132-140` | `lancamento-detail.tsx:244-246` | ✅ |
| FWB-03 AC4 | Editar só em `CRIADA`; `lancamento_imutavel` | `lancamento-detail.test.tsx:145-149`, `:157-158`; `actions.test.ts:34-56` | `lancamentos/actions.ts:30-31`; `lancamentos/errors.ts:39-46` | ✅ |
| FWB-03 AC5 | Três códigos junto do campo | `lancamentos-page.test.tsx:244-264` | `errors.ts:32-36` | ✅ |
| FWB-03 AC6 | Detalhe completo | `lancamento-detail.test.tsx:63-74`, `:87-90`, `:98-100` | `lancamento-detail.tsx:99-126` | ✅ |
| FWB-04 AC1 | Receber; `RECEBIDA`; saldo atualizado | `lancamentos/workflow.test.tsx:78-83` | `lancamento-detail.tsx:168-191`; `lancamentos/hooks.ts:79-97` | ✅ |
| FWB-04 AC2 | Pagar; `PAGA`; saldo atualizado | `workflow.test.tsx:110-113` | `hooks.ts:99-108` | ✅ |
| FWB-04 AC3 | Cancelar com `reason`; `CANCELADA` com motivo | `:169-175` | `lancamento-detail.tsx:194-219` | ✅ |
| FWB-04 AC4 | Motivo vazio: desabilitado, sem API | `:145-148` | `components/app/confirm-dialog.tsx:80` | ✅ |
| FWB-04 AC5 | Devolução com `devolucao_de_id` | `:224-233` | `lancamentos/devolucao-form.tsx:37-45` | ✅ |
| FWB-04 AC6 | Devolução só em receita `RECEBIDA` | `:261`; `actions.test.ts:29`; `:246` | `actions.ts:38-39` | ✅ |
| FWB-04 AC7 | Ações por status; três códigos atualizam a lista | `:258-267`; `:94-96`, `:121-122`, `:184-185` | `actions.ts:28-41`; `errors.ts:39-46` | ✅ |
| FWB-05 AC1 | Lista, do mais novo ao mais antigo | `features/financeiro/comprovantes/comprovantes-panel.test.tsx:108-111` | `comprovantes/comprovantes-panel.tsx:31-33`, `:86-93` | ✅ |
| FWB-05 AC2 | Com `financeiro:comprovante:create`: multipart no campo `file`, "enviando", lista | `comprovantes-panel.test.tsx:137-147`. A visibilidade de "Anexar comprovante" só é testada com tesouraria completa ou só leitura (`:295-306`) | `comprovantes-panel.tsx:58`; `comprovantes/hooks.ts:30-49` | ❌ **LACUNA G4** (O04) na permissão; envio ✅ |
| FWB-05 AC3 | Acima de 10.420.224 recusa sem API; até o limite envia | `:160-161`; `:176-178` (exatamente no limite); `:189-194` 10.420.225, mensagem e `requests` vazio | `comprovantes/limits.ts:6`, `:21-23`; `comprovantes-panel.tsx:132-136` | ✅ |
| FWB-05 AC4 | Mensagens de `413` e `422` | `:204-208` | `comprovantes/errors.ts:15-20` | ✅ |
| FWB-05 AC5 | URL no clique, aberta, fora do cache | `:244-253`; `:266-270` | `comprovantes/hooks.ts:52-59`; `comprovantes-panel.tsx:42-50` | ✅ |
| FWB-05 AC6 | `document_not_found`, `lancamento_nao_encontrado` | `:209`, `:278` | `comprovantes/errors.ts:21-22` | ✅ |
| Borda | Só leitura: nenhuma escrita | `contas-page.test.tsx:376-377`; `lancamentos-page.test.tsx:272`; `workflow.test.tsx:273`; `comprovantes-panel.test.tsx:305` | — | ✅ |
| Borda | Id fora da lista | `lancamento-detail.test.tsx:105` | `lancamento-detail.tsx:51-52` | ✅ |
| Borda | Mudança de status invalida lista e saldo | `workflow.test.tsx:82-83`, `:112-113`, `:174-175` | `hooks.ts:79-86` | ✅ |

### 05 — Estoque

| AC | Resultado definido pela spec | Teste | Código | |
| --- | --- | --- | --- | --- |
| EWB-01 AC1 | Lista e busca no cliente | `features/estoque/produtos/produtos-page.test.tsx:88-92`, `:107-115` | `produtos/produtos-page.tsx:52`, `:65-79` | ✅ |
| EWB-01 AC2 | `POST /produtos` | `produtos-page.test.tsx:131-133` | `produtos/create-produto-dialog.tsx:53-59` | ✅ |
| EWB-01 AC3 | `codigo_duplicado` no código; `validation_failed` por campo | `:147-152`, `:172-179` | `create-produto-dialog.tsx:62-70` | ✅ |
| EWB-01 AC4 | Saldo com unidade; negativo destacado | `features/estoque/produto-detail.test.tsx:50-51`, `:58-60` | `features/estoque/produto-detail.tsx:85-106` | ✅ |
| EWB-02 AC1 | Histórico **na ordem devolvida pela API** | `features/estoque/movimentacoes/movimentacoes-panel.test.tsx:92-97` afirma as quatro linhas na ordem da fixture, mas todas têm a mesma `criado_em` (`:28-39`) | `movimentacoes/movimentacoes-panel.tsx:51` | ❌ **LACUNA G1** (O01) |
| EWB-02 AC2 | Com `estoque:movimentacao:create`: entrada e saída com `INVENTARIO` | `movimentacoes-panel.test.tsx:161-169` corpos exatos. A visibilidade só é testada com todas as permissões ou só leitura (`:231-238`) | `produto-detail.tsx:67`; `movimentacoes/payload.ts:16-25` | ❌ **LACUNA G5** (O05) na permissão; corpo ✅ |
| EWB-02 AC3 | Com `estoque:movimentacao:create`: devolução com `movimentacao_de_id` | `:170-179` corpo exato. Mesma limitação | `movimentacoes-panel.tsx:23`, `:65` | ❌ **LACUNA G5** (O07) na permissão; corpo ✅ |
| EWB-02 AC4 | Nunca oferece origem reservada | `:131-136`; `payload.test.ts:44-49` | `movimentacoes/movimentacao-dialog.tsx:126-129` | ✅ |
| EWB-02 AC5 | Quatro códigos | `:193`, `:200-212` | `features/estoque/errors.ts:14-21` | ✅ |
| EWB-02 AC6 | Invalida histórico e saldo | `:182-183` `movCalls` 4, `saldoCalls` 4 | `movimentacoes/hooks.ts:30-35` | ✅ |
| EWB-03 AC1 | Com `estoque:movimentacao:adjust`: `POST /ajustes` | `features/estoque/ajustes/ajuste-dialog.test.tsx:67-72` corpo exato. A visibilidade só é testada com todas, só leitura ou sem `saldo:read` (`:131-142`) | `produto-detail.tsx:108`; `ajustes/ajuste-dialog.tsx:59-63` | ❌ **LACUNA G5** (O06) na permissão; corpo ✅ |
| EWB-03 AC2 | Motivo vazio ou quantidade zero: desabilitado | `ajuste-dialog.test.tsx:94-109` | `ajuste-dialog.tsx:43` | ✅ |
| EWB-03 AC3 | Saldo atual e resultante; aviso sem impedir | `:58-61` | `ajuste-dialog.tsx:44`, `:111-129` | ✅ |
| EWB-03 AC4 | Dois códigos | `:114-126` | `errors.ts:18`, `:20` | ✅ |
| EWB-04 AC1 | Só leitura: nenhuma escrita | `produtos-page.test.tsx:188`; `movimentacoes-panel.test.tsx:235-237`; `ajuste-dialog.test.tsx:134` | `produtos-page.tsx:35-37`; `produto-detail.tsx:67`, `:108` | ✅ |
| EWB-04 AC2 | Seção omitida sem chamar a API | `produto-detail.test.tsx:77-79`; `movimentacoes-panel.test.tsx:105-106` | `produto-detail.tsx:74-80` | ✅ |
| Borda | Id fora da lista | `produto-detail.test.tsx:39-41` | `produto-detail.tsx:44` | ✅ |
| Borda | Sem "Devolver" em `AJUSTE` e `DEVOLUCAO` | `movimentacoes-panel.test.tsx:118` `[true, true, false, false]` | `payload.ts:28-30` | ✅ |

### 06 — Integração (INT-01 a INT-03)

| AC | Resultado definido pela spec | Evidência | Código | |
| --- | --- | --- | --- | --- |
| INT-01 AC1 | Registro junta os módulos, na ordem, sem redefinir | `components/app/nav-items.test.ts:19`, `:23-25` (`toBe`), `:29-46` | `components/app/nav-items.ts:11-13` | ✅ |
| INT-01 AC2 | Shell usa o registro, filtrado | `app/(app)/inicio/page.test.tsx:122-125` cinco perfis | `app/(app)/layout.tsx:34` | ✅ |
| INT-01 AC3 | `/` → `/inicio` | `app/page.test.tsx:23-25`; smoke: `307`, `Location: /inicio` | `app/page.tsx:7-9` | ✅ |
| INT-01 AC4 | Atalhos por área; mensagem sem área | `inicio/page.test.tsx:136-141`, `:149` | `app/(app)/inicio/atalhos.tsx:15-34` | ✅ |
| INT-01 AC5 | Raiz sem `/healthz` | `app/page.test.tsx:26` `expect(fetchSpy).not.toHaveBeenCalled()` | `app/page.tsx:1-9` | ✅ |
| Borda | Seção sem item visível some | `components/app/nav.test.ts:27`; `inicio/page.test.tsx:137-141` | `components/app/nav.ts:35-36` | ✅ |
| INT-02 AC1 | Variáveis, ordem, rewrite e gates | `docs/development/web.md:13-41`, `:43-143`, `:155-166`, `:168-183`. Segui o guia no smoke repetido e cheguei a `/entrar` `200` | — | 🔎 ✅ |
| INT-02 AC2 | `API_URL` no build | `docs/development/web.md:39` | — | 🔎 ✅ |
| INT-02 AC3 | Sem segredo | Leitura integral: só marcadores `<...>` (`docs/development/web.md:32`) e credenciais de desenvolvimento do compose | — | 🔎 ✅ |
| INT-03 AC1 | Resultado, data e observação por item | `.specs/features/web/smoke.md:3`, `:34-48` | — | 🔎 ✅ |
| INT-03 AC2 | Cobertura do checklist | `.specs/features/web/smoke.md:36-47` (itens 1 a 11) | — | 🔎 ✅ |
| INT-03 AC3 | Token válido não executável localmente | `.specs/features/web/smoke.md:12`, `:40` | — | 🔎 ✅ |
| INT-03 AC4 | Falhas registradas e repetidas | `.specs/features/web/smoke.md:50` (nenhuma) | — | 🔎 ✅ |
| INT-03 AC5 | Sem senha, token, cookie ou dado real | Leitura integral: valores como `<redigido>`, contas `*@local.test` | — | 🔎 ✅ |

**Contagem**: 129 ACs e 13 casos de borda (142 critérios). Com evidência discriminante completa: 133. Com lacuna em uma cláusula (permissão ou ordem): 9 (FND-04 AC4, USR-04 AC3, USR-04 AC4, FWB-03 AC2, FWB-05 AC2, EWB-02 AC1, EWB-02 AC2, EWB-02 AC3, EWB-03 AC1). Nenhum AC ficou sem teste algum. Gaps de precisão: 3 (P1 a P3).

**Status**: ❌ Lacunas presentes.

---

## Discrimination Sensor

Worktree temporário `tjw/vsensor2` em `8bf79ec` (`git -c core.autocrlf=false worktree add --detach`), `pnpm install --frozen-lockfile`, `API_URL=http://localhost:8080`. Um mutante por vez, só em código de produção, `pnpm test` completo (689 testes), `git checkout -- .` e conferência de `git status --porcelain` vazio depois de cada um. Nenhuma morte foi por timeout (0 ocorrências de "Test timed out" nos 51 registros).

### Mínimo obrigatório (INT-04 AC3) e L2

| # | Mutante | Arquivo | Resultado | Teste que matou |
| --- | --- | --- | --- | --- |
| 1 (M01) | Remover `X-CSRF-Token` das escritas | `lib/api/client.ts:69` | ✅ Morto (30 falhas) | `lib/api/client.test.ts:72` e as asserções de `csrf` de cada tela (por exemplo `workflow.test.tsx:170`) |
| 2 (M02) | Não redirecionar em `401` | `lib/session/session-provider.tsx:145` | ✅ Morto (3 falhas) | `session-provider.test.tsx:259`, `:281`, `:302` |
| 3 (M03) | Exibir ação sem a permissão | `components/app/require-permission.tsx:29` | ✅ Morto (15 falhas) | `require-permission.test.tsx:102`; `ajuste-dialog.test.tsx:131` |
| 4 (M04) | Aceitar `next` externo | `lib/session/routes.ts:45` | ✅ Morto (19 falhas) | `login-form.test.tsx:65`; `routes.test.ts:29` |
| 5 (M05) | Dinheiro como número com casas decimais | `features/financeiro/lancamentos/lancamento-form.tsx:135-136` | ✅ Morto (4 falhas) | `lancamentos-page.test.tsx:164`, `:206`; `lancamento-detail.test.tsx:117`; `workflow.test.tsx:208` |
| 6a (M06a) | Cancelamento com motivo vazio | `components/app/confirm-dialog.tsx:80`; `lancamento-detail.tsx:208` | ✅ Morto (5 falhas) | `workflow.test.tsx:137`; `confirm-dialog.test.tsx:74`, `:81` |
| 6b (M06b) | Ajuste com motivo vazio | `features/estoque/ajustes/ajuste-dialog.tsx:43` | ✅ Morto (2 falhas) | `ajuste-dialog.test.tsx:94` (motivo vazio e só espaços) |
| 7a (M07a) | Diálogo de confirmação confirma ao abrir | `components/app/confirm-dialog.tsx:78` | ✅ Morto (9 falhas) | `confirm-dialog.test.tsx:16`; `workflow.test.tsx:71`, `:102` |
| 7b (M07b) | Desativar conta sem confirmação | `features/financeiro/contas/contas-page.tsx:49` | ✅ Morto (3 falhas) | `contas-page.test.tsx:318`, `:341` |
| 8a (M08a) | Saída manual enviada com origem `VENDA` | `features/estoque/movimentacoes/payload.ts:21` | ✅ Morto (2 falhas) | `payload.test.ts:23`; `movimentacoes-panel.test.tsx:124` |
| 8b (M08b) | Origens reservadas oferecidas no formulário | `movimentacoes/movimentacao-dialog.tsx:128` | ✅ Morto (1 falha) | `movimentacoes-panel.test.tsx:124` |
| 9a (M09a) | Senha temporária em `setQueryData` | `dialogs/temporary-password-dialog.tsx:63` | ✅ Morto (1 falha) | `temporary-password-dialog.test.tsx:92` |
| 9b (M09b) | Senha temporária em `useMutation` | `dialogs/temporary-password-dialog.tsx:63` | ✅ Morto (1 falha) | `temporary-password-dialog.test.tsx:92` |
| 10 (M10) | Token de recuperação mantido na URL | `features/identity/acesso/reset-password-form.tsx:51-53` | ✅ Morto (6 falhas) | `reset-password-form.test.tsx:78`, `:103`, `:167` |
| L2a | URL assinada em `setQueryData` | `comprovantes/comprovantes-panel.tsx:45` | ✅ Morto (1 falha) | `comprovantes-panel.test.tsx:256` |
| L2b | URL assinada por `useMutation` | `comprovantes/comprovantes-panel.tsx:45` | ✅ Morto (1 falha) | `comprovantes-panel.test.tsx:256` |
| L2c | URL assinada por `fetchQuery` | `comprovantes/comprovantes-panel.tsx:45` | ✅ Morto (1 falha) | `comprovantes-panel.test.tsx:256` |

Os 10 mutantes mínimos (14 variantes) e os 3 de L2: 17 de 17 mortos.

### Mutantes próprios

| # | Mutante | Arquivo | Resultado | Teste que matou |
| --- | --- | --- | --- | --- |
| O01 | Histórico reordenado por data (mais novo primeiro) em vez da ordem da API | `features/estoque/movimentacoes/movimentacoes-panel.tsx:51` | ❌ **Sobreviveu** (689 passam) | nenhum → G1 |
| O02 | Permissões de "Editar papéis" e "Retirar acesso" trocadas entre si | `features/identity/usuarios/actions.ts:34-35` | ❌ **Sobreviveu** | nenhum → G2 |
| O03 | "Novo lançamento" liberado por `lancamento:update` | `features/financeiro/lancamentos/lancamentos-page.tsx:50` | ❌ **Sobreviveu** | nenhum → G3 |
| O04 | "Anexar comprovante" liberado por `lancamento:create` | `features/financeiro/comprovantes/comprovantes-panel.tsx:58` | ❌ **Sobreviveu** | nenhum → G4 |
| O05 | Entrada e saída liberadas por `movimentacao:adjust` | `features/estoque/produto-detail.tsx:67` | ❌ **Sobreviveu** | nenhum → G5 |
| O06 | Ajuste liberado por `movimentacao:create` | `features/estoque/produto-detail.tsx:108` | ❌ **Sobreviveu** | nenhum → G5 |
| O07 | "Devolver" liberado por `movimentacao:adjust` | `features/estoque/movimentacoes/movimentacoes-panel.tsx:23`, `:65` | ❌ **Sobreviveu** | nenhum → G5 |
| O08 | `/financeiro/lancamentos` protegida por `conta:read` | `app/(app)/financeiro/lancamentos/page.tsx:13` | ❌ **Sobreviveu** | nenhum → G6 |
| O09 | `/financeiro/lancamentos/[id]` protegida por `conta:read` | `app/(app)/financeiro/lancamentos/[id]/page.tsx:14` | ❌ **Sobreviveu** | nenhum → G6 |
| O10 | Tela de saldo chama `fetch` direto, sem o cliente | `features/financeiro/saldo/hooks.ts:17` | ✅ Morto | `saldo-page.test.tsx:66` (o erro `503` deixa de virar mensagem) |
| O11 | Esquecer a sessão sem limpar o cache | `lib/session/session-provider.tsx:127-128` | ✅ Morto | `session-provider.test.tsx:259`, `:196` |
| O12 | Troca de senha sem rebuscar `/auth/me` | `features/identity/acesso/use-change-password.ts:38` | ✅ Morto | `change-password-form.test.tsx:107` |
| O13 | Mudança de status sem invalidar o saldo | `features/financeiro/lancamentos/hooks.ts:84` | ✅ Morto | `workflow.test.tsx:71`, `:102`, `:151` |
| O14 | Limite de upload exclusivo (`<`) | `comprovantes/limits.ts:22` | ✅ Morto | `comprovantes-panel.test.tsx:173` |
| O15 | Motivo mínimo contando espaços | `usuarios/dialogs/reason.ts:10` | ✅ Morto | `admin-dialogs.test.tsx:108`; `temporary-password-dialog.test.tsx:74` |
| O16 | `useCan` decide pelo papel | `lib/session/session-provider.tsx:236` | ✅ Morto | `session-provider.test.tsx:78` |
| O17 | Seção de saldo sem `estoque:saldo:read` | `features/estoque/produto-detail.tsx:74-76` | ✅ Morto | `produto-detail.test.tsx:74`; `ajuste-dialog.test.tsx:137` |
| O18 | Navegação visível com troca obrigatória | `components/app/app-shell.tsx:37` | ✅ Morto | `app-shell.test.tsx:128` |
| O19 | Ações na própria linha | `usuarios/actions.ts:29` | ✅ Morto | `status-dialogs.test.tsx:54` |
| O20 | Token de recuperação também na query string | `acesso/reset-password-form.tsx:60` | ✅ Morto | `reset-password-form.test.tsx:103` |
| O21 | Token CSRF em `sessionStorage` | `lib/session/session-provider.tsx:175` | ✅ Morto | `session-provider.test.tsx:159` |
| O22 | Movimentação sem invalidar o saldo do produto | `estoque/movimentacoes/hooks.ts:33` | ✅ Morto | `ajuste-dialog.test.tsx:52`; `movimentacoes-panel.test.tsx:124` |
| O23 | Bloqueio de login ignora `Retry-After` | `app/(public)/entrar/login-form.tsx:74` | ✅ Morto | `login-form.test.tsx:104` |
| O25 | Conteúdo protegido durante o carregamento | `app/(app)/layout.tsx:34` | ✅ Morto (12 falhas) | `layout.test.tsx:27`; `inicio/page.test.tsx:147` |
| O26 | Lançamentos sem ordenar do mais novo | `lancamentos/filters.ts:47` | ✅ Morto | `filters.test.ts:22`, `:48` |
| O27 | Lista vazia de papéis sem o aviso | `dialogs/admin-dialogs.tsx:161` | ✅ Morto | `admin-dialogs.test.tsx:230` |
| O28 | Tipo da subconta deixa de ser fixo | `contas/conta-dialogs.tsx:107` | ✅ Morto | `contas-page.test.tsx:188` |
| O29 | Token descartado depois de erro de senha | `acesso/reset-password-form.tsx:80` | ✅ Morto | `reset-password-form.test.tsx:167` |
| O30 | Ajuste que negativa o saldo é impedido | `ajustes/ajuste-dialog.tsx:43` | ✅ Morto | `ajuste-dialog.test.tsx:52` |
| O31 | `csrf_invalid` sem rebuscar `/auth/me` | `lib/session/session-provider.tsx:154` | ✅ Morto | `session-provider.test.tsx:372` |
| O36 | Mensagem genérica expõe o `code` | `lib/api/problem.ts:89` | ✅ Morto | `api-error.test.tsx:26`; `problem.test.ts:70` |
| O38 | Devolução em qualquer receita | `lancamentos/actions.ts:39` | ✅ Morto (7 falhas) | `actions.test.ts:34`; `workflow.test.tsx:258` |
| O52 | Atalhos de `/inicio` sem filtrar por permissão | `app/(app)/inicio/atalhos.tsx:15` | ✅ Morto (6 falhas) | `inicio/page.test.tsx:127`, `:147` |

**Profundidade**: completa para caminho crítico (autenticação, dinheiro, permissões): 51 mutações manuais.
**Resultado do sensor**: 42 de 51 mortos; 9 sobreviventes, todos entre os mutantes próprios. ❌

**Isolamento**: `git worktree remove --force` executado; `tjw/vsensor2` não existe mais. `git status --porcelain` vazio em `tjw/verify2` antes de escrever este relatório e vazio em `C:\Users\Michels\Desktop\torcida-jovem` (`develop` em `6382fe1`).

---

## Smoke repetido (stack real, 2026-10-08)

Montagem pelo `docs/development/web.md`, a partir do worktree `tjw/verify2`: `docker compose up -d db`, migrações (1 a 9) **antes** da API, contêiner Garage avulso (`dxflrs/garage:v2.4.1`, porta 3900), `bootstrap-admin --role PRESIDENTE`, binário de `./cmd/api`, e `next start` sobre o `next build` dos gates (`API_URL=http://localhost:8080`). Todas as chamadas foram feitas em `http://localhost:3000`, com `Origin: http://localhost:3000` e `X-CSRF-Token` nas escritas, com os mesmos caminhos e corpos que os hooks das telas montam. Senhas, cookies, token CSRF e chaves S3 existiram só na memória de um único processo; nenhum foi gravado em arquivo. Contas fictícias `*.v2@local.test`.

**51 de 51 verificações passaram.**

| Item | Observado |
| --- | --- |
| Rotas HTML | `/` → `307` com `Location: /inicio`; `/entrar`, `/inicio`, `/recuperar-acesso`, `/redefinir-senha`, `/conta/senha`, `/admin/usuarios`, `/financeiro/lancamentos`, `/financeiro/contas`, `/financeiro/saldo`, `/estoque/produtos`, `/sem-acesso` → `200` com o título esperado |
| Login | Senha errada → `401 invalid_credentials` em `application/problem+json`. Senha certa → `200`; `Set-Cookie: tj_session=<redigido>; Path=/; Max-Age=28800; HttpOnly; SameSite=Lax`, **sem `Domain`**; `must_change_password=true`, papéis `ASSOCIADO, PRESIDENTE`, 30 permissões, `csrf_token` presente |
| Troca obrigatória | Antes: `403 password_change_required`. Sem `X-CSRF-Token`: `403 csrf_invalid`. Com ele: `204`; `/auth/me` com `must_change_password=false` |
| Origem | Escrita com `Origin: http://evil.example` → `403 origin_not_allowed` |
| Usuários | Criação `201` e promoção `204` a `TESOURARIA` e a `CONSELHO_FISCAL`; primeiro login e troca obrigatória das duas contas → `200` e `204`. Senha temporária → `200` com `Cache-Control: no-store` |
| Financeiro (`TESOURARIA`) | Contas raiz e subconta → `201`. Receita de 15000 centavos → `201 CRIADA`; edição para 20000 com taxa 150 → `200`, líquido 19850; receber → `204`; receber de novo → `409 lancamento_nao_pode_ser_recebido`. Devolução de 5000 → `201` com `devolucao_de_id` certo; pagar → `204`. Despesa cancelada: motivo só de espaços → `422 motivo_obrigatorio`; com motivo → `204`, `CANCELADA`, motivo gravado. Saldo → `200`, `saldo_cents=14850` |
| Comprovantes | Sem CSRF → `403 csrf_invalid`. PDF de 200.052 bytes → `201`, `size_bytes` igual. **Arquivo de exatamente 10.420.224 bytes** (corpo multipart de 10.420.403 bytes) → `201`, `size_bytes=10420224`. Lista → 2 itens. Download pela URL assinada dos dois: `200`, tamanho e SHA-256 iguais aos enviados |
| `CONSELHO_FISCAL` | `/auth/me`: 11 permissões, nenhuma de escrita. Leituras `200`: contas, lançamentos, saldo, comprovantes, URL do comprovante (download íntegro), produtos, movimentações, saldo de estoque. Escritas `403 forbidden`: criar, renomear e desativar conta; criar, editar, receber, pagar e cancelar lançamento; devolução; anexar comprovante; criar produto; movimentação; ajuste. `GET /users` → `403 forbidden`. Saldo inalterado depois das tentativas |
| Recuperação | Pedido para conta existente e inexistente → `202` com corpo idêntico. Token inválido → `400 invalid_reset_token` |
| Bloqueio de login | Depois das falhas seguidas: `429 login_blocked` com `Retry-After: 899` (cabeçalho preservado pelo rewrite) |
| Sessão e logout | Sem cookie → `401 unauthenticated`. Logout sem CSRF → `403 csrf_invalid`; com CSRF → `204` e `Set-Cookie` com `Max-Age=0`; cookie antigo → `401` |

**Não exercitado**: a interação visual no navegador. Não há navegador automatizado nesta verificação (Playwright está fora da V1, WEB-D-017). Cliques, renderização, diálogos, redirecionamentos feitos pelo JavaScript, a remoção do fragmento `#token=` e o menu recolhível em tela estreita não foram vistos. É pendência humana (roteiro em `.specs/features/web/smoke.md:58-71`), não lacuna de teste. A redefinição com token válido continua não executável localmente (INT-03 AC3).

**Desmontagem**: API e `next start` encerrados, Garage removido, `docker compose down -v`, binários e configuração do Garage apagados. Não foi criado `.env`. Os contêineres que já existiam (parados) não foram tocados. Portas 3000, 8080, 5432 e 3900 livres ao fim.

---

## Gate Check

Em `tjw/verify2/web`, com `API_URL=http://localhost:8080`, código de saída real por `cmd /v:on /c "<cmd> & echo EXIT=!ERRORLEVEL!"`.

| Gate | Comando | Saída |
| --- | --- | --- |
| Instalação | `pnpm install --frozen-lockfile` | `EXIT=0` |
| Lint | `pnpm lint` | `EXIT=0` |
| Tipos de rota | `pnpm exec next typegen` | `EXIT=0` |
| Tipos | `pnpm exec tsc --noEmit` | `EXIT=0` |
| Testes | `pnpm test` | `EXIT=0`: 66 arquivos, 689 testes, 0 falhas, 0 pulados |
| Build | `pnpm build` | `EXIT=0`: 15 rotas |
| Audit | `pnpm audit --audit-level high` | `EXIT=0`: 1 alta, ignorada (`GHSA-vfj7-8cjw-p6xm`) |
| Contrato | `pnpm lint:api` | `EXIT=0` |
| Contrato | `pnpm gen:api:check` | `EXIT=0` |
| Backend intacto | `git diff --name-only develop -- api/` | vazio |

- **Arquivos de teste antes da feature** (`ee1b497`): 2. **Depois** (`8bf79ec`): 66.
- **Testes pulados**: nenhum.
- **Estado**: `validate_state.py --root tjw/verify2 .specs/features/web` → `EXIT=1`, como esperado para um relatório reprovado.

---

## Code Quality

| Princípio | Status |
| --- | --- |
| Código mínimo, sem escopo extra | ✅ |
| Mudanças só nos caminhos de cada unidade; nada em `api/` | ✅ |
| Padrões consistentes (cliente único, `RequirePermission`, catálogos de erro) | ✅ |
| Valor afirmado igual ao da spec | ✅ nos 133 critérios completos; ❌ na cláusula de permissão ou de ordem dos 9 com lacuna |
| Cobertura por camada (regra pura 1:1; telas com sucesso, borda e erro) | ✅ |
| Todo teste mapeia para AC, borda ou "Done when" | ✅ |
| Diretrizes documentadas: `docs/development/web.md` (gates, MSW com `onUnhandledRequest: "error"`) | ✅ |

---

## Lacunas (ordenadas por severidade)

Todas são lacunas de teste; nenhuma pede mudança em código de produção. Severidade **menor** para as seis: a API continua sendo a autoridade, e com a matriz de papéis atual nenhum papel tem só parte das escritas de um módulo. Ainda assim reprovam o ciclo, porque são mutantes relevantes sobreviventes.

### G1 — EWB-02 AC1: a ordem da API não é discriminada (O01)

- **Onde**: `web/features/estoque/movimentacoes/movimentacoes-panel.test.tsx:28-39` e `:92-97`.
- **Causa**: as quatro movimentações da fixture têm a mesma `criado_em`; qualquer reordenação estável por data preserva a ordem. A API devolve `ORDER BY criado_em` crescente, então um "mais novo primeiro" inverteria a tela real sem derrubar teste.
- **Unidade dona provável**: EST.
- **Correção sugerida**: dar `criado_em` distintos e fora de ordem cronológica às linhas da fixture e manter a asserção da ordem igual à da resposta.

### G2 — USR-04 AC3 e AC4: `identity:role:assign` e `identity:admin:revoke` não são isoladas (O02)

- **Onde**: `web/features/identity/usuarios/dialogs/admin-dialogs.test.tsx:74-90`.
- **Unidade dona provável**: ID-ADM.
- **Correção sugerida**: dois casos na linha de quem tem vínculo: com `identity:user:read` e `identity:role:assign`, o menu é só `["Editar papéis"]`; com `identity:user:read` e `identity:admin:revoke`, é só `["Retirar acesso administrativo"]`.

### G3 — FWB-03 AC2: `financeiro:lancamento:create` não é isolada em "Novo lançamento" (O03)

- **Onde**: `web/features/financeiro/lancamentos/lancamentos-page.test.tsx:269-273`.
- **Unidade dona provável**: FIN-b.
- **Correção sugerida**: leitura mais `lancamento:update`, `receive`, `pay` e `cancel`, sem `create` → sem "Novo lançamento"; leitura mais só `create` → com o botão.

### G4 — FWB-05 AC2: `financeiro:comprovante:create` não é isolada em "Anexar comprovante" (O04)

- **Onde**: `web/features/financeiro/comprovantes/comprovantes-panel.test.tsx:295-306`.
- **Unidade dona provável**: FIN-b.
- **Correção sugerida**: `comprovante:read` mais `lancamento:create`, sem `comprovante:create` → sem o campo de anexar; `comprovante:read` mais só `comprovante:create` → com o campo.

### G5 — EWB-02 AC2, EWB-02 AC3 e EWB-03 AC1: `movimentacao:create` e `movimentacao:adjust` não são isoladas (O05, O06, O07)

- **Onde**: `web/features/estoque/movimentacoes/movimentacoes-panel.test.tsx:231-238` e `web/features/estoque/ajustes/ajuste-dialog.test.tsx:131-142`.
- **Unidade dona provável**: EST.
- **Correção sugerida**: leituras mais só `estoque:movimentacao:create` → "Registrar entrada", "Registrar saída" e "Devolver" visíveis, "Ajustar estoque" ausente; leituras mais só `estoque:movimentacao:adjust` → o inverso.

### G6 — FND-04 AC4 nas rotas de lançamentos: `financeiro:lancamento:read` não é isolada (O08, O09)

- **Onde**: `web/features/financeiro/lancamentos/lancamentos-page.test.tsx:275-288` e `web/features/financeiro/lancamentos/lancamento-detail.test.tsx:181-194`.
- **Causa**: o caso negativo usa só `financeiro:saldo:read` e o positivo usa as três leituras juntas.
- **Unidade dona provável**: FIN-b.
- **Correção sugerida**: caso negativo com `financeiro:conta:read` e `financeiro:comprovante:read`, sem `lancamento:read` → "Sem acesso" e nenhum `GET /lancamentos`; caso positivo só com `financeiro:lancamento:read`.

### Pendências (não são lacunas de teste; exigem pessoa)

- Smoke visual no navegador pelo roteiro de `.specs/features/web/smoke.md:58-71`, incluindo o menu recolhível em tela estreita (FND-05 AC2).
- Redefinição de senha com token válido, não executável na stack local (INT-03 AC3).

---

## Gaps de precisão da spec

Não reprovam o veredito por si sós.

| # | AC | Gap |
| --- | --- | --- |
| P1 | FND-05 AC2 | A spec fala em "ponto de quebra de tablet" sem valor. O código usa `md` do Tailwind (`components/app/app-shell.tsx:45`, `:84`). O teste cobre o menu recolhível, não a troca pela largura (CSS, fora do jsdom). Fica no smoke visual |
| P2 | FND-03 AC1 | A spec diz "caminho interno começando com `/` e sem `//`". O código recusa `//` só no início (`lib/session/routes.ts:30`); um `next` como `/a//b` é aceito. Continua na mesma origem, então a intenção de segurança está atendida, mas a leitura literal da frase não. Convém a spec dizer "sem `//` no início" ou o código recusar em qualquer posição |
| P3 | FND-01 AC2 | "Nenhuma tela SHALL chamar `fetch` diretamente" não tem guarda automatizada (regra de lint ou teste de código-fonte). Hoje vale por inspeção: `fetch(` só em `lib/api/client.ts:116`. O mutante ingênuo O10 morreu por efeito no tratamento de erro, mas uma chamada direta que trate erros passaria. Sugestão: `no-restricted-globals` ou `no-restricted-syntax` para `fetch` fora de `lib/api/` |

---

## Lições registradas (`lessons.py`)

Só para as falhas reais novas deste ciclo, com sinal `surviving_mutant`:

| Fonte | Lição |
| --- | --- |
| O02 a O09 | Teste de visibilidade por permissão isola a permissão da ação: um caso só com ela e um caso com as outras escritas do módulo sem ela |
| O01 | Fixture de lista cuja ordem a spec define usa valores distintos e fora de ordem na chave de ordenação provável |

---

## Requirement Traceability Update

Nenhum status foi alterado nas specs, porque o veredito é de reprovação.

| Requisito | Status atual | Situação neste ciclo |
| --- | --- | --- |
| FND-01, FND-02, FND-03, FND-05, FND-06 | In Tasks | Pronto para `Verified` |
| FND-04 | In Tasks | ❌ Precisa de correção (G6) |
| ACS-01 a ACS-03 | In Tasks | Pronto para `Verified` |
| USR-01, USR-02, USR-03, USR-05 | In Tasks | Pronto para `Verified` |
| USR-04 | In Tasks | ❌ Precisa de correção (G2) |
| FWB-01, FWB-02, FWB-04 | In Tasks | Pronto para `Verified` |
| FWB-03 | In Tasks | ❌ Precisa de correção (G3) |
| FWB-05 | In Tasks | ❌ Precisa de correção (G4); L2 fechada |
| EWB-01, EWB-04 | In Tasks | Pronto para `Verified` |
| EWB-02, EWB-03 | In Tasks | ❌ Precisa de correção (G1, G5) |
| INT-01 a INT-03 | In Tasks | Pronto para `Verified` |
| INT-04 | In Tasks | ❌ Aberto até G1 a G6 fecharem |

---

## Summary

**Overall**: ❌ Não pronta neste ciclo.

**Checagem ancorada na spec**: 133 de 142 critérios com evidência discriminante completa; 9 com lacuna em uma cláusula; 3 gaps de precisão.
**Sensor**: 42 de 51 mutantes mortos (17 de 17 entre os obrigatórios e L2; 25 de 34 entre os próprios).
**Gates**: todos em `EXIT=0`; 689 testes passam.
**Smoke repetido**: 51 de 51.

**O que funciona**: L1 e L2 fechadas; os dez mutantes mínimos de INT-04 AC3 mortos; login, sessão, CSRF, dinheiro em centavos, confirmações, motivos obrigatórios, origem de estoque, segredos fora de cache e de URL; ciclo do financeiro com comprovante no limite exato de 10.420.224 bytes contra a stack real; `CONSELHO_FISCAL` só com leituras.

**O que falta**: seis correções só de teste (G1 a G6), nas frentes EST, ID-ADM e FIN-b.

**Próximos passos**: as frentes donas acrescentam os casos; nova verificação (ciclo 2 de 3), que pode se limitar aos mutantes O01 a O09, aos gates completos e ao `validate_state.py`.
