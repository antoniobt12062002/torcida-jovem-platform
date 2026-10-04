# Fundação Web — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules. Testes primeiro, derivados dos ACs; um commit atômico por task. **Se a skill não puder ser ativada, PARE e avise.**

Antes de escrever código Next: ler `web/AGENTS.md` e a documentação do pacote instalado em `web/node_modules/next/dist/docs/`. Rodar `pnpm install` antes de começar.

---

**Spec**: `.specs/features/web/spec/01-fundacao.md`
**Design**: `.specs/features/web/design/01-fundacao.md`
**Status**: Draft (2026-10-04), aguardando autorização de execução.

**Pré-requisito**: nenhum (primeira sub-spec da Web).

**Unidades de execução** (ver `.specs/features/web/STATE.md`):

| Unidade | Tasks | Paralelismo |
| --- | --- | --- |
| F1 (infra/API) | T1, T2, T3 | paralela a F4 |
| F4 (kit de UI) | T4 | paralela a F1 |
| F2 (sessão e acesso) | T5, T6 | depois de F1 e F4 integradas |
| F5 (componentes de app) | T7, T8 | depois de F2; paralela a F3 (sub-spec 02) |

## Test Coverage Matrix

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| `next.config.ts` (rewrite) | unit + aceite manual | destino calculado a partir de `API_URL`; falha sem `API_URL`; roteiro FND-01 AC5 | `web/next.config.test.ts`, `.specs/features/web/evidence/f1-rewrite.md` | `pnpm test` |
| `lib/api/` (cliente, problem, query keys) | unit com MSW | CSRF só em escritas; callbacks de 401/403; parsing de `problem+json` | `web/lib/api/*.test.ts` | `pnpm test` |
| `lib/forms/` | unit | `errors[]` → campos e `root` | `web/lib/forms/*.test.ts` | `pnpm test` |
| `components/ui/` | component | papel acessível e associação de erro | `web/components/ui/*.test.tsx` | `pnpm test` |
| `lib/session/` | unit + component com MSW | estados, login, logout, 401 único, `safeNext` | `web/lib/session/*.test.ts(x)` | `pnpm test` |
| Rotas `(public)` e `(app)` | component com MSW | portão, redirecionamentos, mensagens | `web/app/**/*.test.tsx` | `pnpm test` |
| `components/app/` | component | ACs de FND-05 | `web/components/app/*.test.tsx` | `pnpm test` |

## Gate Check Commands

Todos rodam a partir de `web/`, com o código de saída capturado de verdade (no Windows: `cmd /v:on /c "<comando> & echo !ERRORLEVEL!"`). Um gate só passa com saída `0`.

| Level | When | Command |
| --- | --- | --- |
| Web | toda task | `pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build` |
| Audit | T1 (dependências novas) e fim de cada unidade | `pnpm audit --audit-level high` |
| Contract | toda task | `pnpm lint:api && pnpm gen:api:check` (o contrato não muda; confirma que nada foi gerado de novo) |
| Backend intacto | toda task | `git diff --name-only develop -- api/` vazio |
| Ownership | toda task | `git diff --name-only <base da unidade>` só contém caminhos da unidade (tabela de propriedade em `STATE.md`) |

Nenhum gate pode ser enfraquecido, pulado ou mascarado (CI inclusive). Teste não é apagado nem afrouxado para passar.

---

## Execution Plan

```
T1 → T2
T2 → T3
T3 → T5
T4 → T5
T5 → T6
T6 → T7
T7 → T8
```

T1-T3 (F1) e T4 (F4) rodam em paralelo, em agentes e worktrees separados. T5 só começa com as duas unidades integradas.

---

## Task Breakdown

### T1: Dependências, rewrite e harness de testes

**What**: adicionar `openapi-fetch`, `@tanstack/react-query`, `react-hook-form` e `msw` (versões estáveis atuais, sem nenhuma outra biblioteca); `rewrites()` no `next.config.ts`; harness MSW e `renderWithProviders` (só com `QueryClientProvider` nesta task; F2 acrescenta a sessão no próprio arquivo de teste dela).
**Where**: `web/package.json`, `web/pnpm-lock.yaml`, `web/next.config.ts`, `web/vitest.config.ts`, `web/test/`
**Depends on**: None
**Reuses**: `vitest.config.ts` existente.
**Requirement**: FND-01 (AC1, AC2 parcial)

**Done when**:
- [x] Só as quatro dependências de WEB-D-003 entram (`pnpm-lock.yaml` coerente, `--frozen-lockfile` passa)
- [x] `next.config.ts` aponta `/api/v1/:path*` para `${API_URL}/api/v1/:path*` e falha com mensagem clara sem `API_URL`
- [x] Harness MSW ligado no setup do Vitest, com teste de fumaça
- [x] Gates Web, Audit, Contract, Backend intacto e Ownership passam

**Tests**: unit (configuração do rewrite; harness)
**Gate**: Web + Audit + Contract + Backend intacto + Ownership
**Commit**: `feat(web): dependencias da v1, rewrite same-origin e harness de testes`

---

### T2: Cliente da API, erros, chaves de consulta e formulários

**What**: `lib/api/client.ts` (clientes por contrato, middleware de CSRF e de 401/403 com callbacks registráveis), `lib/api/problem.ts`, `lib/api/query-keys.ts`, `lib/forms/`.
**Where**: `web/lib/api/client.ts`, `web/lib/api/problem.ts`, `web/lib/api/query-keys.ts`, `web/lib/forms/`
**Depends on**: T1
**Reuses**: tipos gerados em `web/lib/api/*.d.ts` (sem alterar).
**Requirement**: FND-01 (AC2-4), FND-06

**Done when**:
- [x] `X-CSRF-Token` presente em `POST`/`PUT`/`PATCH`/`DELETE` e ausente em `GET`
- [x] `401` em rota diferente de `POST /auth/login` chama `onUnauthenticated`; `403 password_change_required` e `403 csrf_invalid` chamam seus callbacks; `403 forbidden` não chama nenhum
- [x] `problem+json` vira `ApiError` com `status`, `code` e `errors[]`; `413`, `503`, rede e `code` desconhecido têm mensagens genéricas
- [x] `applyProblemToForm` põe erros nos campos e o resto em `root`
- [x] Chaves de consulta cobrem todos os recursos do design
- [x] Gates passam

**Tests**: unit com MSW
**Gate**: Web + Contract + Backend intacto + Ownership
**Commit**: `feat(web): cliente tipado da api com csrf e tratamento de problem+json`

---

### T3: Verificação de aceite do rewrite contra a API real

**What**: executar o roteiro do design (login, escrita autenticada, upload de 9,5 MiB) com a stack local e registrar a evidência. Registrar também se `next build` exige `API_URL` no momento do build.
**Where**: `.specs/features/web/evidence/f1-rewrite.md`
**Depends on**: T2
**Reuses**: `docker-compose.yml`, `.env.example`, comandos de `CLAUDE.md`.
**Requirement**: FND-01 (AC5, AC6)

**Done when**:
- [x] Os três itens de FND-01 AC5 demonstrados, com status HTTP e cabeçalhos relevantes registrados (sem cookie, token ou senha no arquivo)
- [x] Se algum item falhar: unidade F1 parada, falha registrada e reportada ao coordenador, sem contorno
- [x] Gates passam

**Tests**: aceite manual contra a stack real (roteiro registrado)
**Gate**: Web + Contract + Backend intacto + Ownership
**Commit**: `docs(web): evidencia do rewrite same-origin contra a api real`

---

### T4: Kit de UI

**What**: instalar os componentes listados no design com o CLI do shadcn no estilo `base-nova`, mais o toaster sobre `Toast` do `@base-ui/react`; testes de acessibilidade dos campos.
**Where**: `web/components/ui/`, `web/components.json`, `web/app/globals.css`
**Depends on**: None
**Reuses**: `button.tsx`, `badge.tsx`, `lib/utils.ts`.
**Requirement**: FND-02

**Done when**:
- [x] Todos os componentes de FND-02 AC1 presentes
- [x] Campos com erro têm `aria-invalid` e `aria-describedby` apontando para a mensagem
- [x] `package.json` e `pnpm-lock.yaml` intactos; se algum componente exigir dependência nova, a task para e reporta
- [x] Gates passam

**Tests**: component
**Gate**: Web + Contract + Backend intacto + Ownership
**Commit**: `feat(web): kit de componentes de interface`

---

### T5: Sessão e providers raiz

**What**: `lib/session/` (provider, hooks de permissão, login, logout, callbacks do cliente, `routes.ts`, `safeNext`, formatação de `Retry-After`) e `app/layout.tsx` com `lang="pt-BR"`, metadados e providers.
**Where**: `web/lib/session/`, `web/app/layout.tsx`
**Depends on**: T3, T4
**Reuses**: `lib/api/client.ts`, `lib/api/query-keys.ts`, toaster de `components/ui`.
**Requirement**: FND-03 (AC5-9), FND-04 (AC3)

**Done when**:
- [x] Token CSRF e contexto só em memória (teste confere que `localStorage` e `sessionStorage` ficam vazios)
- [x] Logout limpa todo o cache de consultas
- [x] Dois `401` simultâneos produzem um único redirecionamento
- [x] `safeNext` recusa `//x`, `https://x` e caminhos sem `/` inicial
- [x] Gates passam

**Tests**: unit + component com MSW
**Gate**: Web + Contract + Backend intacto + Ownership
**Commit**: `feat(web): sessao em memoria, csrf e providers raiz`

---

### T6: Login, portão das rotas autenticadas e "Sem acesso"

**What**: `/entrar`, layout público, layout autenticado (portão) e página `/sem-acesso`.
**Where**: `web/app/(public)/layout.tsx`, `web/app/(public)/entrar/`, `web/app/(app)/layout.tsx`, `web/app/(app)/sem-acesso/`
**Depends on**: T5
**Reuses**: `lib/session`, `components/ui`.
**Requirement**: FND-03 (AC1-4, AC10), FND-04 (AC1, AC2, AC4)

**Done when**:
- [ ] Login com sucesso, `invalid_credentials`, `login_blocked` (com tempo de espera) e `must_change_password` cobertos
- [ ] `/entrar` tem o link "Esqueci minha senha" para `routes.recuperarAcesso`
- [ ] Rota autenticada sem sessão leva a `/entrar?next=…`, sem renderizar o conteúdo
- [ ] Pessoa autenticada em `/entrar` vai para a página inicial
- [ ] Gates passam

**Tests**: component com MSW
**Gate**: Web + Contract + Backend intacto + Ownership
**Commit**: `feat(web): login e protecao das rotas autenticadas`

---

### T7: Shell, gate de permissão e estados de tela

**What**: `app-shell.tsx`, `nav.ts`, `require-permission.tsx`, `states.tsx`, `api-error.tsx`; `app/(app)/layout.tsx` passa a envolver o conteúdo no shell, com lista de itens vazia.
**Where**: `web/components/app/`, `web/app/(app)/layout.tsx`
**Depends on**: T6
**Reuses**: `components/ui` (`sheet`, `dropdown-menu`, `skeleton`, `alert`), `lib/session`, `lib/api/problem.ts`.
**Requirement**: FND-05 (AC1-3, AC8), FND-04 (AC5), FND-06 (apresentação)

**Done when**:
- [ ] Shell mostra só itens com permissão; menu recolhível abaixo do ponto de quebra
- [ ] Com `must_change_password = true`, o shell esconde a navegação e mantém só "Sair"
- [ ] `RequirePermission` cobre os modos página e ação
- [ ] Estados de carregamento, vazio e erro com tentar de novo
- [ ] Gates passam

**Tests**: component
**Gate**: Web + Contract + Backend intacto + Ownership
**Commit**: `feat(web): shell da aplicacao, gate de permissao e estados de tela`

---

### T8: Campos de dinheiro e quantidade, confirmação, status e formatação

**What**: `money-field.tsx`, `quantity-field.tsx`, `confirm-dialog.tsx`, `status-badge.tsx`, `format.ts`.
**Where**: `web/components/app/`
**Depends on**: T7
**Reuses**: `lib/money.ts`, `components/ui` (`alert-dialog`, `field`, `input`, `textarea`, `badge`).
**Requirement**: FND-05 (AC4-7)

**Done when**:
- [ ] Campo de dinheiro entrega inteiro em centavos (nunca `float`), com os vetores de `lib/money.test.ts`
- [ ] Campo de quantidade recusa decimal; modos `positive` e `nonZero`
- [ ] Confirmação com motivo obrigatório não habilita o botão com motivo vazio ou só espaços
- [ ] Datas em pt-BR no fuso de São Paulo; `authorLabel` devolve "você" para o id da sessão
- [ ] Gates passam (incluindo Audit, fim da unidade F5)

**Tests**: component + unit
**Gate**: Web + Audit + Contract + Backend intacto + Ownership
**Commit**: `feat(web): campos de dinheiro e quantidade e dialogo de confirmacao`

---

## Phase Execution Map

```
T1 → T2
T2 → T3
T3 → T5
T4 → T5
T5 → T6
T6 → T7
T7 → T8
```

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1 | dependências + rewrite + harness: o mínimo para testar qualquer outra coisa | ✅ Aceitável |
| T2 | camada `lib/api` + `lib/forms`, um único contrato interno | ✅ Aceitável |
| T3 | 1 roteiro de aceite + 1 arquivo de evidência | ✅ Granular |
| T4 | 1 diretório, componentes instalados por ferramenta | ✅ Aceitável |
| T5 | 1 diretório + layout raiz | ✅ Granular |
| T6 | 4 rotas pequenas do mesmo fluxo | ✅ Aceitável |
| T7 | estrutura do shell | ✅ Granular |
| T8 | componentes de formulário e formatação | ✅ Granular |

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | None | ✅ Match |
| T2 | T1 | T1 | ✅ Match |
| T3 | T2 | T2 | ✅ Match |
| T4 | None | None | ✅ Match |
| T5 | T3, T4 | T3, T4 | ✅ Match |
| T6 | T5 | T5 | ✅ Match |
| T7 | T6 | T6 | ✅ Match |
| T8 | T7 | T7 | ✅ Match |

## Test Co-location Validation

| Task | Code Layer | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1 | config + harness | unit | unit | ✅ OK |
| T2 | `lib/api`, `lib/forms` | unit com MSW | unit com MSW | ✅ OK |
| T3 | evidência | aceite manual | aceite manual | ✅ OK |
| T4 | `components/ui` | component | component | ✅ OK |
| T5 | `lib/session` | unit + component | unit + component | ✅ OK |
| T6 | rotas | component com MSW | component com MSW | ✅ OK |
| T7 | `components/app` | component | component | ✅ OK |
| T8 | `components/app` | component + unit | component + unit | ✅ OK |
