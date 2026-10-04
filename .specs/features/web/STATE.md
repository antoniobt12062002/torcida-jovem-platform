# Web V1 — STATE

Documento de coordenação da feature `web`: a camada Web (Next.js, em `web/`) que torna utilizáveis os módulos já existentes na API (`identity`, `financeiro` e `estoque`). Decisões locais usam o prefixo **WEB-D-NNN**, que nunca reaproveita o numerador `AD-NNN` de `.specs/STATE.md`.

Esta feature **não altera a API**: nenhum contrato OpenAPI, handler, migration ou regra de negócio muda. O front consome a API como ela é. Qualquer lacuna que só se resolva no backend é registrada aqui e vira decisão própria, fora deste ciclo.

---

## O que significa "Web V1 pronta para uso"

A Web V1 está pronta quando, com a API, o PostgreSQL e o armazenamento S3 rodando e o front servido pelo Next:

1. Uma pessoa com conta consegue entrar, trocar a senha (inclusive a troca obrigatória do primeiro acesso), recuperar o acesso pelo link do e-mail (`/redefinir-senha#token=`) e sair, sem chamar a API diretamente.
2. Um administrador (`ADMIN_SISTEMA` ou `PRESIDENTE`) consegue cadastrar a pessoa, promovê-la a administrativa com os papéis certos (por exemplo, `TESOURARIA` ou `ESTOQUE_LOJA`), ajustar papéis, retirar o acesso, desativar, reativar e gerar uma senha temporária. Sem isso, ninguém consegue usar `financeiro` nem `estoque` sem chamar a API diretamente.
3. A tesouraria executa pela Web todo o ciclo do `financeiro` V1: plano de contas, lançamentos, devolução, recebimento, pagamento, cancelamento com motivo, saldo e comprovantes (anexar, listar e baixar).
4. O responsável de estoque executa pela Web todo o ciclo do `estoque` V1: produtos, entrada, saída, devolução, movimentações, ajuste com motivo e saldo.
5. Diretoria e Conselho Fiscal veem o que suas permissões de leitura permitem e **não** veem ações que não podem executar. A API continua sendo a autoridade: uma ação negada pela API vira uma mensagem clara, nunca um erro genérico.
6. Todos os erros documentados no contrato de cada operação (status e `code` de `problem+json`) aparecem para a pessoa como mensagem compreensível em português.
7. Os gates de `web` passam (lint, typecheck, testes, build, audit, contrato) e o checklist manual de smoke contra a stack real foi executado e registrado.

---

## Mapa das sub-specs

| # | Spec | Requisitos | Execução | Status |
|---|---|---|---|---|
| 01 | `spec/01-fundacao.md` | FND-01 a FND-06 | F1, F4, F2, F5 | Especificada |
| 02 | `spec/02-identity-acesso.md` | ACS-01 a ACS-03 | F3 | Especificada |
| 03 | `spec/03-identity-usuarios.md` | USR-01 a USR-05 | ID-ADM | Especificada |
| 04 | `spec/04-financeiro-web.md` | FWB-01 a FWB-05 | FIN-a, FIN-b | Especificada |
| 05 | `spec/05-estoque-web.md` | EWB-01 a EWB-04 | EST | Especificada |
| 06 | `spec/06-integracao.md` | INT-01 a INT-04 | INT, VERIFY | Especificada |

Cada sub-spec tem `spec/`, `design/` e `tasks/` com o mesmo nome. A numeração `T1…Tn` é local a cada arquivo de tasks. As dependências entre sub-specs aparecem como pré-requisitos e no DAG global abaixo.

---

## DAG global (unidades de execução)

```
F1 (infra/API) ─┐
                ├─→ F2 (sessão e acesso) ─┬─→ F3 (senha) ──────────────────────────────┐
F4 (UI kit) ────┘                         └─→ F5 (componentes de app) ─┬─→ ID-ADM ──┤
                                                                       ├─→ FIN-a ───┤
                                                                       ├─→ FIN-b ───┼─→ INT ─→ VERIFY
                                                                       └─→ EST ─────┘
```

| Unidade | Sub-spec / tasks | Depende de | Classificação |
|---|---|---|---|
| F1 | 01 / T1-T3 | — | PARALELA (com F4) |
| F4 | 01 / T4 | — | PARALELA (com F1) |
| F2 | 01 / T5-T6 | F1, F4 | SEQUENCIAL |
| F5 | 01 / T7-T8 | F2 | PARALELA (com F3) |
| F3 | 02 / T1-T2 | F2 | PARALELA (com F5) |
| ID-ADM | 03 / T1-T3 | F5 | PARALELA (com FIN-a, FIN-b, EST) |
| FIN-a | 04 / T1-T2 | F5 | PARALELA (com ID-ADM, FIN-b, EST) |
| FIN-b | 04 / T3-T5 | F5 | PARALELA (com ID-ADM, FIN-a, EST) |
| EST | 05 / T1-T3 | F5 | PARALELA (com ID-ADM, FIN-a, FIN-b) |
| INT | 06 / T1-T3 | F3, ID-ADM, FIN-a, FIN-b, EST | INTEGRAÇÃO |
| VERIFY | 06 / T4 | INT | INTEGRAÇÃO (agente independente) |

Dentro de cada unidade, as tasks são sequenciais e feitas pelo mesmo agente. A exceção é a sub-spec 01: F1 (T1-T3) e F4 (T4) correm em paralelo.

F3 não depende de F5: usa só `components/ui` e `lib/session`. ID-ADM, FIN-a, FIN-b e EST dependem de F5, cujos componentes de app usam. O ponto de partida da rodada 4 é a integração de F3 e F5 juntas, para que nenhuma frente comece sobre uma base que ainda vai mudar.

Nenhuma unidade está classificada como DECISÃO: as decisões que existiam foram fechadas em 2026-10-04 (WEB-D-001 a WEB-D-007).

### Agentes paralelos (planejamento; nada criado ainda)

A regra aprovada: duas unidades rodam ao mesmo tempo só se não compartilham arquivos de forma conflitante, não dependem de estado produzido pela outra e não têm decisão comum pendente. Cada agente trabalha em worktree e branch isolados a partir do ponto de integração anterior. O coordenador integra, confere os diffs e a propriedade dos arquivos, e roda os gates antes de liberar a próxima rodada.

| Rodada | Agentes | Integração pelo coordenador |
|---|---|---|
| 1 | A: F1 · B: F4 | conferir que A não tocou em `components/ui/` e que B não tocou em `package.json`; gates completos |
| 2 | um agente: F2 | gates completos |
| 3 | A: F3 · B: F5 | conferir a propriedade (ver tabela abaixo); gates completos |
| 4 | A: ID-ADM · B: FIN-a · C: FIN-b · D: EST | conferir a propriedade por diretório; gates completos com as quatro frentes juntas |
| 5 | um agente: INT | gates completos e smoke manual |
| 6 | agente independente: VERIFY | relatório `validation.md` |

Cada agente roda, antes de devolver, os gates da sua task (ver cada `tasks/*.md`) e informa task, branch, arquivos alterados, testes, gates, SHA e bloqueios. Uma stop condition encontrada por um agente para só aquela unidade. As outras seguem.

### Propriedade de arquivos (ownership)

Caminhos relativos a `web/`. "Exclusivo" quer dizer que nenhuma outra unidade altera o caminho. Um caminho passa de uma unidade para a seguinte só quando a primeira já foi integrada.

| Unidade | Arquivos exclusivos | Observação |
|---|---|---|
| F1 | `package.json`, `pnpm-lock.yaml`, `next.config.ts`, `vitest.config.ts`, `lib/api/client.ts`, `lib/api/problem.ts`, `lib/api/query-keys.ts`, `lib/forms/`, `test/`, `.specs/features/web/evidence/f1-rewrite.md` | **Única** unidade que altera dependências. Os `lib/api/*.d.ts` gerados não mudam (contrato inalterado). |
| F4 | `components/ui/`, `components.json`, `app/globals.css` | **Única** dona de `components/ui/` e da instalação de componentes compartilhados. Não altera `package.json`; se um componente exigir dependência nova, para e reporta. |
| F2 | `lib/session/`, `app/layout.tsx`, `app/(public)/layout.tsx`, `app/(public)/entrar/`, `app/(app)/layout.tsx`, `app/(app)/sem-acesso/` | Providers raiz (query, sessão, toaster), login, logout e proteção de rotas. |
| F5 | `components/app/` (exceto `nav-items.ts`) | Shell, gate de permissão, estados, confirmação, campos de dinheiro e quantidade. Altera `app/(app)/layout.tsx` (passa a envolver o shell), depois da integração de F2. |
| F3 | `features/identity/acesso/`, `app/(public)/recuperar-acesso/`, `app/(public)/redefinir-senha/`, `app/(app)/conta/senha/` | Feature `identity` (acesso). |
| ID-ADM | `features/identity/nav.ts`, `features/identity/usuarios/`, `app/(app)/admin/usuarios/` | Feature `identity` (usuários). Não é tocada pelos agentes de Financeiro ou Estoque. |
| FIN-a | `features/financeiro/nav.ts`, `features/financeiro/contas/`, `features/financeiro/saldo/`, `app/(app)/financeiro/contas/`, `app/(app)/financeiro/saldo/` | Não altera arquivos de FIN-b. O `nav.ts` lista os três itens do módulo, inclusive o caminho de lançamentos. |
| FIN-b | `features/financeiro/lancamentos/`, `features/financeiro/comprovantes/`, `app/(app)/financeiro/lancamentos/` | Não altera arquivos de FIN-a. Lê contas pela própria consulta, com a chave compartilhada de `lib/api/query-keys.ts`. |
| EST | `features/estoque/`, `app/(app)/estoque/` | Não altera arquivos de Financeiro nem de Identity. |
| INT | `components/app/nav-items.ts`, `app/page.tsx`, `app/(app)/inicio/`, `app/(app)/layout.tsx` (só a lista passada ao shell), `docs/development/web.md`, `.specs/features/web/smoke.md` | Navegação global, página inicial autenticada e redirecionamento da raiz. Lê `features/*/nav.ts` sem alterar. Única unidade que liga os módulos entre si. |
| VERIFY | `.specs/features/web/validation.md` | Não altera código. Lacunas viram tasks na frente dona do arquivo. |

Cada módulo expõe seus itens de navegação em `features/<modulo>/nav.ts`, arquivo próprio do módulo. A INT é a única que junta esses itens em `components/app/nav-items.ts`. Assim o registro global não vira arquivo disputado por agentes paralelos.

`app/(app)/layout.tsx` é o único arquivo que passa por três donos (F2 → F5 → INT), sempre em rodadas diferentes e só depois da integração do dono anterior.

### Gates por task

Definidos em cada `tasks/*.md`. Em todas: **Web** (`pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build`), **Contract** (`pnpm lint:api && pnpm gen:api:check`), **Backend intacto** (`git diff --name-only develop -- api/` vazio) e **Ownership** (diff só nos caminhos da unidade). **Audit** (`pnpm audit --audit-level high`) na task de dependências (01/T1) e na última task de cada unidade. Códigos de saída capturados de verdade (no Windows, `cmd /v:on` com `!ERRORLEVEL!`).

### Configuração de CI necessária para o build

Não é decisão de arquitetura. O rewrite é resolvido ao carregar `next.config.ts`, e `next typegen` e `next build` falham sem `API_URL`, por design: não há valor padrão. O job `web` do CI fornece `API_URL=http://localhost:8080` só nesses dois passos, como valor de validação. Os gates locais rodam com `API_URL` no ambiente. Cada ambiente implantado fornece o seu próprio `API_URL` no build (nota da ADR-010 de 2026-10-04).

### Critérios de integração e verificador independente

Definidos em `design/06-integracao.md`: o que cada agente entrega, conferência de propriedade, integração sem resolução manual de conflito, gates completos por rodada, e a estratégia do agente VERIFY (checagem ancorada na spec, sensor com dez falhas mínimas em worktree temporário, smoke repetido, `validation.md`).

---

## Decisões (WEB-D-NNN)

### WEB-D-001 — Rewrite same-origin de `/api/v1/*` pelo Next — **Aprovada (D1)**
O navegador só fala com a origem do front; o Next encaminha `/api/v1/*` para `API_URL`. Sem mudança na API. Registrada em [ADR-010](../../../docs/adr/010-acesso-do-navegador-a-api-por-rewrite-same-origin.md) e `AD-017`.

### WEB-D-002 — Client Components nas telas operacionais — **Aprovada (D2)**
As telas que leem e escrevem na API são Client Components e chamam a API pelo navegador. É o modelo que `Origin` e CSRF da API protegem. Server Components ficam só para o que não depende da sessão (estrutura de página, metadados).

### WEB-D-003 — Bibliotecas: `openapi-fetch`, TanStack Query, React Hook Form e MSW — **Aprovada (D3)**
Nenhuma biblioteca equivalente ou concorrente entra sem nova justificativa. Consequências derivadas (sem decisão nova):
- **Sem biblioteca de schema** (por exemplo `zod`). A validação de formulário usa as regras do React Hook Form, e a fonte de verdade é a resposta `problem+json` da API, mapeada para os campos (`lib/forms/`).
- **Toast com o `Toast` do `@base-ui/react`**, que já é dependência. Nenhuma biblioteca de toast nova.
- `@hookform/resolvers` não entra (não há schema para resolver).

### WEB-D-004 — Administração de usuários na Web V1, como feature própria de `identity` — **Aprovada (D4)**
Inclui criação, promoção, papéis, retirada de acesso, desativação, reativação e redefinição administrativa de senha. Fica na sub-spec 03, separada de `financeiro` e `estoque`.

### WEB-D-005 — Desktop-first, utilizável em tablet e celular — **Aprovada (D5)**
O layout é pensado para desktop. Em telas estreitas, a navegação vira menu recolhível e tabelas viram rolagem horizontal ou listas empilhadas. Nada fica inacessível, mas não há recurso específico de celular.

### WEB-D-006 — Testes da V1: unidade, componente e integração com MSW, mais checklist manual de smoke — **Aprovada (D6)**
Playwright e E2E contra a stack real ficam fora da V1 (evolução registrada em "Fora do V1").

### WEB-D-007 — Origens de movimentação de estoque expostas na UI — **Aprovada (D7)**
`AJUSTE_MANUAL` só existe no fluxo de ajuste. As origens reservadas para integrações (`VENDA`, `COMPRA`, `EVENTO`) não aparecem como opção manual. Na entrada, na saída e na devolução manuais, a UI oferece só `INVENTARIO`.

### WEB-D-008 — Correção do DAG proposto na auditoria (achado técnico, não decisão de negócio)
O DAG da auditoria punha F2 (login) e F3 (senha) antes ou em paralelo a F4, mas F4 é a única dona de `components/ui/`, de onde vêm os campos desses formulários. Isso é dependência, não conflito de arquivo: as telas teriam de ser refeitas. O DAG definitivo é `F1 ∥ F4 → F2 → (F3 ∥ F5) → (ID-ADM ∥ FIN-a ∥ FIN-b ∥ EST) → INT → VERIFY`. Todas as regras de propriedade aprovadas ficam preservadas, e o paralelismo de F3 com F5 e das quatro frentes de módulo continua.

### WEB-D-009 — Proteção de rotas no layout autenticado, sem `proxy.ts` (derivada de WEB-D-001 e WEB-D-002)
O layout `app/(app)/layout.tsx` consulta `GET /api/v1/auth/me` e redireciona quando não há sessão. Não se usa `proxy.ts`, porque no Next 16 ele trunca em silêncio corpos acima de 10 MB, e o upload de comprovantes aceita até 11 MiB. A segurança real continua na API: o gate do front é só navegação.

**Correção (2026-10-04, achado técnico de F1, resolvido pelo mantenedor):** o próprio rewrite também trunca corpos acima de cerca de 10 MiB, mesmo sem `proxy.ts`; acima disso a pessoa recebe `500` e não o `413` da API (evidência em `evidence/f1-rewrite.md`). Resolução, como mitigação só da camada Web e não como regra de negócio: o front recusa antes do envio arquivos acima de **10 MiB − 64 KiB (10.420.224 bytes)**, margem para o envelope `multipart/form-data` (FWB-05 AC3). O limite da API continua 10 MiB e é a autoridade. Sem `experimental.proxyClientMaxBodySize` e sem proxy externo na V1. Nota correspondente na ADR-010.

### WEB-D-010 — Chaves de consulta centralizadas em `lib/api/query-keys.ts` (derivada de WEB-D-003 e da regra de propriedade)
F1 define as chaves do TanStack Query para todos os recursos do contrato. Assim FIN-b pode ler contas e invalidar o saldo sem importar código de FIN-a, e frentes paralelas compartilham cache sem compartilhar arquivo.

### WEB-D-011 — `package.json` só em F1 (derivada da regra de propriedade)
Todas as dependências da V1 (as de WEB-D-003) entram em F1. Nenhuma outra unidade altera `package.json` nem `pnpm-lock.yaml`. Se precisar, para e reporta.

### WEB-D-012 — Interface em português do Brasil (derivada de AD-006 e das convenções do projeto)
`lang="pt-BR"`, textos em português e dinheiro e datas no formato pt-BR (`lib/money.ts` já existe).

### WEB-D-013 — Rotas dos módulos pertencem aos módulos; navegação global e raiz pertencem à INT (derivada da regra de propriedade aprovada)
No App Router, criar o arquivo de página é registrar a rota. Cada frente cria as páginas só dentro do seu segmento (`app/(app)/financeiro/contas/`, …). A INT é a única que monta a navegação global, a página inicial autenticada e o redirecionamento de `/`. Isso cumpre "a navegação global e o registro final das rotas ficam para INT" sem criar arquivo disputado.

### WEB-D-014 — Autoria exibida como "você" ou identificador curto (derivada de "não alterar backend por conveniência")
`criado_por`, `cancelado_por`, `responsavel_id` e `uploaded_by` são UUIDs. Listar usuários exige `identity:user:read`, que tesouraria e estoque não têm. A UI mostra "você" quando o id é o da sessão e um identificador curto nos demais casos. Exibir nomes exigiria mudança na API: fica registrado como lacuna, fora deste ciclo.

### WEB-D-015 — Saldo de estoque por produto na página do produto (derivada de "não alterar backend por conveniência")
A API não tem saldo em lote. A listagem de produtos não mostra saldo (evita uma chamada por linha), e o saldo aparece no detalhe do produto. Saldo em lote fica registrado como lacuna de API, fora deste ciclo.

### WEB-D-016 — Listagens sem paginação nem filtro no servidor (herdada de FIN-D-021 e das specs de estoque)
Contas, lançamentos, produtos e movimentações chegam inteiros. Busca, filtro e ordenação são feitos no cliente. Usuários têm cursor (`next_cursor`) e filtros `active` e `role` na API, e a UI os usa.

### WEB-D-017 — E2E com Playwright como evolução posterior (derivada de WEB-D-006)
Fica fora da V1 e registrado aqui para um ciclo próprio, porque exige subir Postgres, S3, API e front no CI.

---

## Fora do V1 (não implementar sem nova decisão)

| Item | Motivo |
|---|---|
| Dashboard com indicadores e gráficos | A API não tem endpoint de indicadores; seria cálculo no front sobre listas inteiras. |
| Relatórios e exportações (CSV, PDF) | Não há endpoint nem regra de negócio especificada. |
| Prestação de contas e parecer do Conselho Fiscal | As permissões existem (`financeiro:prestacao_contas:*`, `financeiro:parecer:opine`), mas não há endpoint. |
| Consulta ao registro de auditoria (`GET /api/v1/audit-logs`) | Proposto como fora do V1 na auditoria e não contestado. A API existe; fica para um ciclo próprio, reavaliável. |
| PWA, notificações push, recursos específicos de celular, modo offline | Fora de WEB-D-005. |
| Portal público, loja pública, área do associado | Módulos ainda inexistentes ou sem especificação. |
| Internacionalização e tema escolhido pelo usuário | WEB-D-012; o tema segue o padrão do sistema operacional, se o kit já suportar, sem seletor próprio. |
| E2E com Playwright | WEB-D-017. |
| Nomes de autores e saldo em lote | WEB-D-014 e WEB-D-015: exigem mudança na API. |

---

## Lacunas da API registradas (não resolvidas neste ciclo)

| Lacuna | Classificação | Tratamento na V1 |
|---|---|---|
| Autoria em UUID, sem nome | Questão de UX; exigiria mudança na API | WEB-D-014 |
| Sem saldo de estoque em lote | Questão de UX; exigiria mudança na API | WEB-D-015 |
| Sessão não informa quando expira | Resolvível no front | Reagir a `401 session_expired` (FND-03) |
| Listas sem paginação nem filtro | Resolvível no front | WEB-D-016 |
| `Set-Cookie`, `Origin` e corpos grandes pelo rewrite não documentados no Next | Risco técnico | Critério de aceite de F1 contra a API real (FND-01) |
| `docker-compose.yml` sem serviço S3, mas a API não sobe sem armazenamento (`errStorageRequired`) | Ambiente de desenvolvimento; corrigir exigiria mudança fora de `web/` | F1 e o smoke usam um contêiner Garage avulso, com a imagem de `api/internal/platform/testutil/s3.go`, configurado só no `.env` local; o guia da INT documenta. Incluir S3 no compose fica para decisão própria. |
| Provedor de e-mail local (`log`) não expõe o link de recuperação | Ambiente de desenvolvimento | Smoke cobre pedido e token inválido; token válido coberto por MSW (INT-03 AC3). |

---

## Inconsistências documentais encontradas (não corrigidas nesta rodada)

- `.specs/STATE.md`, Handoff: ainda diz que `financeiro` só está especificado e sem código, e não menciona `estoque`. Os dois V1 estão em `develop`.
- `.specs/features/financeiro/STATE.md`: marca `06-api-http` como concluída, mas não registra a mesclagem da PR #57.
- `.specs/features/estoque/STATE.md`, Handoff: ainda diz "aguardando merge" da correção de RBAC e "aguardar revisão das duas PRs", mas #58 e #59 estão mescladas.
- `.env.example`: traz `STORAGE_ENABLED=false` e diz que "nada na API consome" o armazenamento ainda, mas desde `financeiro` a API recusa subir sem ele (`errStorageRequired` em `api/cmd/api/main.go`). Com o `.env` copiado do exemplo, a API não sobe.
- Ambiente local: `web/node_modules` está defasado em relação ao lockfile (`next@16.3.6`, `cn@0.2.6`). Um `pnpm install` resolve antes da implementação; não é problema do repositório.

---

## Handoff

- **Fase**: Specify, Discuss, Design e Tasks concluídos para as 6 sub-specs, em 2026-10-04, depois das decisões D1 a D7 do mantenedor. Nenhum código, branch, commit, push ou PR.
- **Próximo passo**: aguardar autorização do mantenedor para iniciar a execução pela rodada 1 (F1 ∥ F4).
- **Bloqueios**: nenhum.
