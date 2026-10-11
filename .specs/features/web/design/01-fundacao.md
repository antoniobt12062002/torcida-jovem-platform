# Fundação Web — Design

**Spec**: `.specs/features/web/spec/01-fundacao.md`
**Decisões**: WEB-D-001 a WEB-D-017 (`.specs/features/web/STATE.md`), AD-017, ADR-010.

## Architecture Overview

```
Navegador (origem do front)
  │  Client Components (telas)
  │    └─ hooks de feature (TanStack Query) ─→ lib/api/client.ts (openapi-fetch)
  │                                             ├─ middleware: X-CSRF-Token em escritas
  │                                             └─ middleware: 401 / 403 → lib/session
  │  fetch same-origin /api/v1/*
  ▼
Servidor Next  ── rewrite (next.config.ts) ──→ API Go (API_URL)/api/v1/*
```

- O navegador nunca fala com `API_URL`. O cookie `tj_session` é host-only na origem do front.
- `lib/session` é o único dono do contexto de sessão (`AuthContext` de `/auth/me`) e do token CSRF, em memória.
- `app/(app)/layout.tsx` é o portão das rotas autenticadas. `app/(public)/` tem as rotas sem sessão.
- A API é sempre a autoridade. O front só esconde ações e reage aos erros.

## Code Reuse Analysis

### Existing Components to Leverage

| Existente | Uso |
| --- | --- |
| `web/lib/api/{identity,financeiro,estoque,common}.d.ts` | Tipos de `paths` para `openapi-fetch` (gerados; inalterados). |
| `web/lib/money.ts` (`formatBRL`, `parseBRL`) | Campo de dinheiro e exibição. |
| `web/lib/utils.ts` (`cn`) | Classes de componente. |
| `web/components/ui/{button,badge}.tsx` | Mantidos; o kit cresce ao lado. |
| `web/components.json` (estilo `base-nova`, `@base-ui/react`) | Instalação dos componentes do kit. |
| `web/vitest.config.ts` (jsdom, Testing Library) | Base do harness de testes. |

### Integration Points

| Sistema | Como |
| --- | --- |
| API Go | Só por `/api/v1/*` via rewrite. Sem mudança. |
| `.env` / `API_URL` | Lido pelo `next.config.ts`. O rewrite é calculado quando o Next carrega a configuração; F1 confirma se `next build` precisa da variável e registra o resultado no guia de desenvolvimento que a INT escreve. |

## Components

### `next.config.ts` (F1)

- `rewrites()` devolve uma única regra: `source` = `/api/v1/:path*` e `destination` = `${API_URL}/api/v1/:path*`.
- Se `API_URL` não estiver definida, a configuração falha ao carregar, com mensagem clara (nunca um destino vazio).
- Sem `proxy.ts`, sem `middleware.ts` (WEB-D-009).

### `lib/api/client.ts` (F1)

- Um cliente `openapi-fetch` por contrato (`identity`, `financeiro`, `estoque`), todos com `baseUrl` igual à origem atual e `credentials: "same-origin"`.
- Middleware de requisição: em métodos de escrita, adiciona `X-CSRF-Token` a partir de `getCsrfToken()`, uma função registrada por `lib/session`. Sem token, a requisição segue sem o cabeçalho e a API devolve `403 csrf_invalid`, que é tratado.
- Middleware de resposta: em `401` (exceto `POST /auth/login`), chama `onUnauthenticated()`; em `403 password_change_required`, chama `onPasswordChangeRequired()`; em `403 csrf_invalid`, chama `onCsrfInvalid()`. Os três são registrados por `lib/session`, para que `lib/api` não importe `lib/session` (sem ciclo).

### `lib/api/problem.ts` (F1)

- `ApiError { status, code, title, detail?, errors: {field, code}[] }` e `toApiError(response, body)`.
- Mensagens genéricas para `413`, `503`, falha de rede e `code` desconhecido (FND-06).
- `messageFor(error, catalog)`: usa o catálogo da feature e cai na mensagem genérica.

### `lib/api/query-keys.ts` (F1, WEB-D-010)

Chaves de todos os recursos do contrato, por exemplo:

```
queryKeys.session.me
queryKeys.identity.users(filters)
queryKeys.financeiro.contas
queryKeys.financeiro.lancamentos
queryKeys.financeiro.saldo
queryKeys.financeiro.comprovantes(lancamentoId)
queryKeys.estoque.produtos
queryKeys.estoque.movimentacoes(produtoId)
queryKeys.estoque.saldo(produtoId)
```

### `lib/forms/` (F1)

- `applyProblemToForm(error, setError, fieldMap)`: põe cada `errors[]` no campo do React Hook Form; o que não tem campo vira erro do formulário (`root`).
- Sem biblioteca de schema (WEB-D-003).

### `test/` (F1)

- `test/msw/server.ts`: `setupServer` do MSW, ligado no `vitest.setup` (iniciar, resetar handlers, encerrar).
- `test/msw/fixtures.ts`: fábricas de `AuthContext`, `problem+json` e itens de cada recurso, tipadas pelos `.d.ts`.
- `test/render.tsx`: `renderWithProviders(ui, { session, permissions })` com `QueryClient` novo por teste (sem repetição de consultas) e provedor de sessão.

### Verificação de aceite do rewrite (F1, FND-01 AC5)

Roteiro executado pelo agente com a stack local (`docker compose up -d db`, migrate, armazenamento S3 local, API, `pnpm dev`), registrado em `.specs/features/web/evidence/f1-rewrite.md`. A API não sobe sem armazenamento (`errStorageRequired`) e o `docker-compose.yml` não tem S3: o agente sobe um contêiner Garage avulso com a imagem de `api/internal/platform/testutil/s3.go` (`dxflrs/garage`), com `STORAGE_ENABLED=true` e `S3_*` apenas no `.env` local, nunca commitados. Os passos usados vão para a evidência, sem credenciais.

1. Login pelo navegador (ou `curl` com `Origin: http://localhost:3000` contra `:3000`): resposta `200`, `Set-Cookie: tj_session` sem `Domain`.
2. Escrita autenticada (`POST /api/v1/financeiro/contas`) com `X-CSRF-Token`: `201`, nunca `403 origin_not_allowed`.
3. Upload de PDF de 9,5 MiB em `/api/v1/financeiro/lancamentos/{id}/comprovantes`: `201`, `size_bytes` igual ao do arquivo.
4. Os dados de teste usam conta e senha locais de desenvolvimento, nunca dados pessoais; o arquivo de evidência não contém cookie, token nem senha.

### Kit de UI (F4)

- Instalados pelo CLI do shadcn no estilo do projeto (`base-nova`): `input`, `label`, `field`, `select`, `textarea`, `checkbox`, `dialog`, `alert-dialog`, `table`, `dropdown-menu`, `skeleton`, `alert`, `sheet` (navegação recolhível) e `toast` (sobre `Toast` do `@base-ui/react`).
- Antes de instalar, F4 confere se o componente pede dependência nova. Se pedir, para (FND-02 AC3).
- `app/globals.css` só ganha os tokens que os componentes exigem.

### `lib/session/` (F2)

- `SessionProvider`: consulta `GET /auth/me` com `queryKeys.session.me`; expõe `{ status: "loading" | "authenticated" | "anonymous", context }`.
- `useSession()`, `usePermission(p)`, `useCan(...ps)`.
- `login(email, password)`: `POST /auth/login`, grava o contexto devolvido no cache de `session.me`.
- `logout()`: `POST /auth/logout`, `queryClient.clear()`, navegação para `/entrar`.
- Registra os callbacks de `lib/api/client.ts`. `onUnauthenticated` usa uma trava para fazer um único redirecionamento (caso de borda da spec).
- `routes.ts`: constantes de caminho (`/entrar`, `/conta/senha`, `/recuperar-acesso`, `/redefinir-senha`, `/inicio`, `/sem-acesso`) e `safeNext(next)`.
- `RETRY_AFTER`: formata `Retry-After` (segundos) em texto pt-BR.

### Rotas (F2)

| Rota | Arquivo | Conteúdo |
| --- | --- | --- |
| raiz | `app/layout.tsx` | `lang="pt-BR"`, metadados "TJ Platform", `QueryClientProvider`, `SessionProvider`, toaster |
| públicas | `app/(public)/layout.tsx` | layout simples, centralizado |
| `/entrar` | `app/(public)/entrar/page.tsx` | formulário de login |
| autenticadas | `app/(app)/layout.tsx` | portão: carregando → conteúdo; anônimo → `/entrar?next=`; `must_change_password` fora de `/conta/senha` → `/conta/senha` |
| `/sem-acesso` | `app/(app)/sem-acesso/page.tsx` | página "Sem acesso" |

`app/page.tsx` (ping de `/healthz`) fica intocado até a INT.

### `components/app/` (F5)

| Componente | Função |
| --- | --- |
| `app-shell.tsx` | Cabeçalho, navegação lateral, menu recolhível (`sheet`), menu da pessoa (Trocar senha, Sair); com `must_change_password`, só "Sair" (FND-05 AC8) |
| `nav.ts` | Tipo `NavItem { label, href, permission }` e `visibleItems(items, permissions)` |
| `require-permission.tsx` | Renderiza filhos só com a permissão; sem ela, mostra "Sem acesso" ou nada (modo de ação) |
| `states.tsx` | `LoadingState`, `EmptyState`, `ErrorState` (com tentar de novo) |
| `confirm-dialog.tsx` | Confirmação, com motivo opcionalmente obrigatório (trim) |
| `money-field.tsx` | Campo pt-BR → centavos inteiros (`parseBRL`) |
| `quantity-field.tsx` | Inteiro; modo `positive` ou `nonZero` |
| `status-badge.tsx` | Badge por status (`CRIADA`, `RECEBIDA`, `PAGA`, `CANCELADA`, `ativo`/`inativo`) |
| `format.ts` | `formatDate`, `formatDateTime` (pt-BR, `America/Sao_Paulo`), `shortId`, `authorLabel(id, sessionUserId)` (WEB-D-014) |
| `api-error.tsx` | Mostra `ApiError` com `messageFor` |

F5 altera `app/(app)/layout.tsx` para envolver o conteúdo autenticado com `AppShell` e passa uma lista de itens vazia. A INT liga a lista real (`components/app/nav-items.ts`).

## Data Models

- `AuthContext`, `Problem` e os recursos vêm dos `.d.ts` gerados. Nenhum tipo de API é redefinido à mão.
- `ApiError` (acima) é o único tipo novo de erro.

## Error Handling Strategy

| Origem | Tratamento |
| --- | --- |
| `401` | `lib/session` limpa e redireciona (FND-03 AC6) |
| `403 password_change_required` | `/conta/senha` (FND-03 AC7) |
| `403 csrf_invalid` | rebusca `/auth/me` e avisa (FND-03 AC8) |
| `403 forbidden` | mensagem de permissão (FND-04 AC5) |
| `422 validation_failed` | `applyProblemToForm` (FND-06 AC1) |
| outro `code` | catálogo da feature → mensagem genérica (FND-06 AC2-3) |
| `413`, `503`, rede | mensagens genéricas (FND-06 AC4-5) |

## Risks & Concerns

| Risco | Mitigação |
| --- | --- |
| O rewrite não repassar `Set-Cookie`, `Origin` ou corpo grande | FND-01 AC5 verificado contra a API real; se falhar, F1 para (AC6) e o mantenedor decide (por exemplo a opção (b) da ADR-010). |
| `API_URL` ser fixada no build | F1 registra o comportamento observado; a INT documenta. |
| `Request` com URL relativa no Node durante testes | `baseUrl` usa a origem do jsdom (`http://localhost:3000`); o harness cobre isso. |
| Um agente instalar dependência fora de F1 | Gate de propriedade (`package.json` e `pnpm-lock.yaml` só no diff de F1). |

## Tech Decisions

| Decisão | Motivo |
| --- | --- |
| Callbacks registrados em vez de `lib/api` importar `lib/session` | Evita dependência circular e permite testar o cliente isolado. |
| Cliente por contrato | Os `.d.ts` são um arquivo por contrato; um cliente por `paths` mantém os tipos exatos. |
| Detalhe de recurso lido do cache da lista quando o contrato não tem `GET` por id | Contrato inalterado; listas são completas (WEB-D-016). |

## Tips

- Antes de escrever código Next, ler `web/AGENTS.md` e a documentação em `web/node_modules/next/dist/docs/` (Next 16 tem mudanças incompatíveis).
- Rodar `pnpm install` antes de começar: o `node_modules` local está defasado.
