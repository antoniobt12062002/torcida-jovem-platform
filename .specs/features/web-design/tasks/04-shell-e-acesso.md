# Shell e Acesso — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules. Testes primeiro, derivados dos critérios de aceite; um commit atômico por task; a task é marcada como concluída neste arquivo antes do commit e no mesmo commit. **Se a skill não puder ser ativada, PARE e avise.**

Antes de escrever código Next: ler `web/AGENTS.md` e a documentação do pacote instalado em `web/node_modules/next/dist/docs/` (em especial `error.md`, `not-found.md` e `app-icons.md` em `03-api-reference/03-file-conventions/`).

Autorização: a aprovação destas tasks autoriza implementação e commits locais na branch da unidade. `git push`, PR, merge e deploy exigem autorização explícita para cada ação.

---

**Spec**: `.specs/features/web-design/spec/04-shell-e-acesso.md`
**Design**: `.specs/features/web-design/design/04-shell-e-acesso.md`
**Status**: Draft (2026-10-11), aguardando aprovação do mantenedor.

**Pré-requisito**: **portão do espécime aprovado pelo mantenedor** e registrado em `STATE.md`. Sem isso, nenhuma task deste arquivo começa.

**Unidade de execução** (propriedade e DAG em `design/05-migracao-web-v1.md`):

| Unidade | Tasks | Rodada | Entrada | Saída |
| --- | --- | --- | --- | --- |
| DS-S Shell e acesso | T1 a T5 | 6, paralela a DS-ID, DS-FA, DS-FB e DS-ES | Portão aprovado; DS-T, DS-K, DS-L e DS-A integradas e congeladas | Layouts, telas de acesso, páginas de erro e início migrados; FND-05 AC2 reverificado no navegador |

## Restrições comuns a todas as tasks deste arquivo

- Só os arquivos de DS-S são alterados (gate Propriedade). `package.json` e `pnpm-lock.yaml` não mudam.
- **Arquivos congelados**: `components/ui/`, `components/app/` (exceto `nav-items.ts`), `lib/url-state/`, `lib/responsive/`, `test/` e os mocks de `lib/session`. Se uma mudança neles for necessária, a task para e reporta ao coordenador; não há contorno local.
- O comportamento de FND-03, FND-04 e ACS-01 a ACS-03 não muda: mensagens, redirecionamentos, tratamento de token e de `Retry-After` ficam como estão. `lib/session` e `lib/api` não são alterados.
- Nenhuma página de erro exibe mensagem técnica, pilha ou identificador interno.
- Sem ilustração, fotografia ou A FERA; sem "Desde 2013" fora do escudo.

## Test Coverage Matrix

> Gerada do código, das diretrizes do projeto e da spec. Diretrizes consultadas: `CLAUDE.md`, `web/AGENTS.md`, `docs/development/web.md`, `web/vitest.config.ts`, `.specs/features/web/tasks/*.md`.

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| Layouts e rotas de `app/` | component com MSW | Cada AC de DSS atendido pela rota; testes de comportamento existentes passando | `web/app/**/*.test.tsx` | `pnpm test` |
| Formulários de acesso (`features/identity/acesso`) | component com MSW | Comportamento existente intacto; atributos de campo novos | `web/features/identity/acesso/*.test.tsx` | `pnpm test` |
| Guarda da unidade | unit | Nenhum valor solto nos diretórios de DS-S | `web/app/**/*.guard.test.ts` | `pnpm test` |
| Navegação responsiva | aceite manual no navegador | 390, 768, 820, 1024 e 1440 px, build de produção | `.specs/features/web-design/evidence/` | roteiro manual |

## Gate Check Commands

Todos rodam a partir de `web/`, com o código de saída capturado de verdade. Um gate só passa com saída `0`.

| Level | When | Command |
| --- | --- | --- |
| Web | toda task | `pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build` |
| Contract | toda task | `pnpm lint:api && pnpm gen:api:check` |
| Backend intacto | toda task | `git diff --name-only develop -- api/` vazio |
| Dependências intactas | toda task | `git diff --name-only develop -- web/package.json web/pnpm-lock.yaml` vazio |
| Ownership | toda task | `git diff --name-only <base da unidade>` só contém caminhos de DS-S |
| Congelados intactos | toda task | `git diff --name-only <base da unidade> -- web/components/ui web/components/app web/lib web/test` só lista `web/components/app/nav-items.ts` |
| Audit | última task da unidade | `pnpm audit --audit-level high` |

Nenhum gate pode ser enfraquecido, pulado ou mascarado. Teste não é apagado nem afrouxado para passar.

---

## Execution Plan

```
T1 → T2
T2 → T3
T3 → T4
T4 → T5
```

Um único agente, em worktree próprio, em paralelo às quatro frentes de módulo.

---

## Task Breakdown

### T1: Layout autenticado e página "Sem acesso"

**Unidade**: DS-S
**What**: levar o portão das rotas autenticadas e a página "Sem acesso" aos estados e à estrutura padrão.
**Where**: `web/app/(app)/layout.tsx`
**Arquivos**: `web/app/(app)/layout.tsx`, `web/app/(app)/sem-acesso/`, e os testes existentes desses dois
**Depends on**: None
**Reuses**: `AppShell`, `LoadingState`, `ErrorState`, `NoAccess`, `PageContainer`.
**Requirement**: DSP-02 (AC1, AC4, AC5), DSS-01 (AC6), FND-04 (AC1, AC2, preservados)

**Mudanças previstas**: os dois blocos `main` montados à mão (carregamento e erro de sessão) passam a usar os estados padrão dentro de `PageContainer`; a lógica do portão (redirecionamentos, troca obrigatória de senha) não muda.

**Done when**:
- [x] Os testes existentes de `layout.test.tsx` e de `sem-acesso` passam sem alteração de asserção de comportamento
- [x] Carregamento e erro de sessão usam os componentes de estado, com uma única região principal
- [x] Gates passam

**Tests**: component com MSW
**Verificação visual**: carregamento, erro de sessão e "Sem acesso" em 390 e 1440 px
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Congelados intactos
**Commit**: `refactor(web): portao autenticado e sem acesso nos padroes do design system`

---

### T2: Páginas de erro em português

**Unidade**: DS-S
**What**: criar as páginas de erro de renderização e de endereço inexistente, com identidade e sem detalhe técnico.
**Where**: `web/app/(app)/error.tsx`
**Arquivos**: `web/app/(app)/error.tsx`, `web/app/not-found.tsx`, `web/app/global-error.tsx`, `web/app/error-pages.test.tsx`
**Depends on**: T1
**Reuses**: `PageContainer`, `AccessLayout`, `BrandMark`, `Button`; convenções `error`, `global-error` e `not-found` do Next instalado.
**Requirement**: DSS-04 (AC1 a AC5)

**Mudanças previstas**: `error.tsx` dentro do shell, com "Tentar de novo" e "Voltar ao início"; `not-found.tsx` no layout de acesso; `global-error.tsx` com documento próprio, escudo e "Tentar de novo". O nome da função de nova tentativa é o da documentação instalada. Se `global-error` não aceitar importar a folha global, usa um bloco mínimo de estilos locais com os valores dos tokens, declarado como exceção na guarda.

**Done when**:
- [ ] Cada página tem um `h1` em português e leva o foco a ele ao montar
- [ ] Nenhuma exibe `error.message`, pilha ou `digest` (teste com um erro de exemplo)
- [ ] "Tentar de novo" chama a função de nova tentativa recebida
- [ ] Num build de produção, um endereço inexistente mostra "Página não encontrada" e uma exceção forçada mostra a página de erro dentro do shell
- [ ] Gates passam

**Tests**: component
**Verificação visual**: as três páginas num build de produção, em 390 e 1440 px
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Congelados intactos
**Commit**: `feat(web): paginas de erro em portugues com a identidade do sistema`

---

### T3: Telas de acesso

**Unidade**: DS-S
**What**: aplicar o layout de acesso e o layout de formulário a login, recuperação, redefinição e troca de senha.
**Where**: `web/app/(public)/layout.tsx`
**Arquivos**: `web/app/(public)/` (layout, `entrar/`, `recuperar-acesso/`, `redefinir-senha/`), `web/app/(app)/conta/`, `web/features/identity/acesso/` e os testes existentes desses diretórios
**Depends on**: T2
**Reuses**: `AccessLayout`, `FormLayout`, `Field`, `PageTitle`; formulários e testes atuais.
**Requirement**: DSS-03 (AC1 a AC6), DSP-04 (AC1 a AC9), ACS-01 a ACS-03 e FND-03 (preservados)

**Mudanças previstas**: layout público passa a `AccessLayout` (duas colunas no desktop, coluna única abaixo de 1024, marca "Torcida Jovem"); formulários em `FormLayout`, com botão primário em largura total; `spellCheck={false}` e `autoCapitalize="none"` nos campos de e-mail; durante a troca obrigatória, `/conta/senha` usa o layout de acesso; fora dela, é página comum de largura estreita.

**Done when**:
- [ ] Os testes existentes de login, recuperação, redefinição e troca de senha passam sem alteração de asserção de comportamento
- [ ] O escudo e "Torcida Jovem" aparecem nas quatro telas; "TJ Platform" não aparece como marca
- [ ] Campos de e-mail e senha mantêm `autocomplete` de autenticação; e-mail sem correção ortográfica
- [ ] Enviar vazio leva o foco ao primeiro campo inválido
- [ ] Gates passam

**Tests**: component com MSW
**Verificação visual**: as quatro telas em 390, 820 e 1440 px, comparadas com o espécime
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Congelados intactos
**Commit**: `feat(web): telas de acesso com a identidade da torcida`

---

### T4: Navegação global e página inicial

**Unidade**: DS-S
**What**: completar a navegação global com ícone e descrição de "Início" e migrar a página inicial para os atalhos em cartão.
**Where**: `web/app/(app)/inicio/`
**Arquivos**: `web/components/app/nav-items.ts`, `web/app/(app)/inicio/` e seus testes, `web/app/(app)/shell-acesso.guard.test.ts`
**Depends on**: T3
**Reuses**: `navItems`, `visibleItems`, `Card`, `PageContainer`, `PageHeader`, `EmptyState`; biblioteca de guarda de DS-F.
**Requirement**: DSS-06 (AC1 a AC4), INT-01 (AC1, AC4, preservados), DSF-01 (AC7), DSM-05 (AC1)

**Mudanças previstas**: `inicioNavItem` ganha ícone e descrição; a ordem e a composição de `navItems` não mudam; `/inicio` usa `PageHeader` e uma grade de cartões que são links, com ícone, nome e descrição quando o módulo a fornecer; sem indicadores. A guarda da unidade cobre `app/(public)/`, `app/(app)/layout.tsx`, `inicio/`, `sem-acesso/`, `conta/`, as páginas de erro e `features/identity/acesso/`.

**Done when**:
- [ ] Os testes existentes de `nav-items` e de `inicio` passam sem alteração de asserção de comportamento
- [ ] Os atalhos são links, respeitam as permissões e ficam em uma coluna abaixo de 768 px (classe conferida em teste)
- [ ] Conta sem nenhuma área vê o estado vazio com a explicação
- [ ] A guarda passa nos diretórios de DS-S, sem contêiner nem título montados à mão
- [ ] Gates passam

**Tests**: component + unit
**Verificação visual**: `/inicio` com duas contas de permissões diferentes, em 390, 820 e 1440 px
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Congelados intactos
**Commit**: `feat(web): pagina inicial e navegacao global no design system`

---

### T5: Revalidação de FND-05 AC2 e da navegação no navegador

**Unidade**: DS-S
**What**: reexecutar, num navegador real, o critério de navegação responsiva com o novo ponto de quebra e registrar a evidência.
**Where**: `.specs/features/web-design/evidence/shell-responsivo.md`
**Arquivos**: `.specs/features/web-design/evidence/shell-responsivo.md`
**Depends on**: T4
**Reuses**: procedimento de `.specs/features/web/smoke-visual.md` e de `docs/development/web.md` (stack local, contas fictícias, script temporário fora do repositório).
**Requirement**: DSS-02 (AC1 a AC7), FND-05 (AC2, revisado por WDS-D-019), DSS-01 (AC4, AC5)

**Mudanças previstas**: nenhum código. Roteiro executado num build de produção, com a stack real, em 390, 768, 820, 1024 e 1440 px.

**Done when**:
- [ ] Em 390, 768 e 820 px: botão de menu presente, gaveta abre, fecha ao navegar e com Esc, foco volta ao botão, nenhuma ação inacessível
- [ ] Em 1024 e 1440 px: navegação fixa presente, sem botão de menu
- [ ] Link de pular funciona por teclado; item ativo anunciado; nenhuma rolagem horizontal de página a partir de 320 px
- [ ] A evidência registra que a verificação anterior de FND-05 AC2 (ponto de quebra de 768 px) não foi reaproveitada
- [ ] O arquivo não contém senha, token, cookie nem dado pessoal
- [ ] IF algum item falhar THEN a unidade para, registra a falha e reporta ao coordenador
- [ ] Gates passam, incluindo Audit (fim da unidade DS-S)

**Tests**: aceite manual no navegador (roteiro registrado)
**Verificação visual**: é o objeto da task
**Gate**: Web + Audit + Contract + Backend intacto + Dependências intactas + Ownership + Congelados intactos
**Commit**: `docs(web-design): evidencia da navegacao responsiva no novo ponto de quebra`

---

## Phase Execution Map

```
T1 → T2
T2 → T3
T3 → T4
T4 → T5
```

Rodada 6, em paralelo às frentes de módulo de `tasks/05-migracao-web-v1.md`.

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1 | 1 layout e 1 página pequena | ✅ Granular |
| T2 | 3 páginas de erro da mesma convenção | ✅ Aceitável |
| T3 | 4 telas do mesmo fluxo e do mesmo layout | ✅ Aceitável |
| T4 | 1 página e o registro de navegação | ✅ Granular |
| T5 | 1 roteiro e 1 arquivo de evidência | ✅ Granular |

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | None | ✅ Match |
| T2 | T1 | T1 | ✅ Match |
| T3 | T2 | T2 | ✅ Match |
| T4 | T3 | T3 | ✅ Match |
| T5 | T4 | T4 | ✅ Match |

## Test Co-location Validation

| Task | Code Layer | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1 | layouts e rotas | component com MSW | component com MSW | ✅ OK |
| T2 | rotas de erro | component | component | ✅ OK |
| T3 | rotas e formulários de acesso | component com MSW | component com MSW | ✅ OK |
| T4 | rota e guarda | component + unit | component + unit | ✅ OK |
| T5 | navegação responsiva | aceite manual | aceite manual | ✅ OK |
