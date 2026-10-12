# Fundação Visual — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules. Testes primeiro, derivados dos critérios de aceite; um commit atômico por task; a task é marcada como concluída neste arquivo antes do commit e no mesmo commit. **Se a skill não puder ser ativada, PARE e avise.**

Antes de escrever código Next: ler `web/AGENTS.md` e a documentação do pacote instalado em `web/node_modules/next/dist/docs/`. Rodar `pnpm install` antes de começar.

Autorização: a aprovação destas tasks autoriza implementação e commits locais na branch da unidade. `git push`, PR, merge e deploy exigem autorização explícita para cada ação.

---

**Spec**: `.specs/features/web-design/spec/01-fundacao-visual.md`
**Design**: `.specs/features/web-design/design/01-fundacao-visual.md`
**Status**: Draft (2026-10-11), aguardando aprovação do mantenedor.

**Unidades de execução** (propriedade e DAG em `design/05-migracao-web-v1.md`):

| Unidade | Tasks | Rodada | Entrada | Saída |
| --- | --- | --- | --- | --- |
| DS-F Fundação visual | T1 a T4 | 1 | Base documental na branch; `pnpm install` feito | Tokens, fontes, escudo e guardas integrados; gates completos |
| DS-E Espécime visual | T5 a T7 | 5 | DS-A integrada (`tasks/03-padroes.md`, T3 a T10) | Espécime conferido e **aprovado pelo mantenedor** |

**Pré-requisito externo de T5**: todas as tasks de DS-K, DS-T, DS-L e DS-A integradas. Neste arquivo a seta `T4 → T5` representa só a ordem; a dependência real é a rodada 4.

## Restrições comuns a todas as tasks deste arquivo

- Só os arquivos da unidade são alterados (gate Propriedade). `package.json` e `pnpm-lock.yaml` não mudam.
- Nenhum arquivo de `api/`, nenhum contrato, nenhuma regra de negócio ou de permissão.
- Nenhuma cor literal fora de `app/globals.css`. Nenhuma versão nova da marca; o desenho do escudo não muda.
- Teste existente não é removido nem enfraquecido.
- Sem Stitch, sem serviço externo, sem dependência nova.

## Test Coverage Matrix

> Gerada do código, das diretrizes do projeto e da spec. Diretrizes consultadas: `CLAUDE.md`, `web/AGENTS.md`, `docs/development/web.md`, `web/vitest.config.ts`, `.specs/features/web/tasks/*.md`.

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| Guardas de design (leitura de arquivos) | unit | Cada regra da guarda tem caso que passa e caso que falha | `web/test/design/*.test.ts` | `pnpm test` |
| `app/globals.css` (tokens) | unit (leitura do CSS) | Todo par de contraste declarado; todo token usado existe; camadas presentes | `web/test/design/*.test.ts` | `pnpm test` |
| `app/layout.tsx` (fontes, metadados) | unit + build | Variáveis de fonte ligadas aos tokens; `theme-color`; build de produção | `web/test/design/*.test.ts` | `pnpm test && pnpm build` |
| Marca (`public/brand`, ícones) | unit | Integridade dos traçados; `viewBox`; dimensões do PNG | `web/test/design/brand.test.ts` | `pnpm test` |
| Rota do espécime | component | Seções presentes; indisponível em produção | `web/app/(dev)/**/*.test.tsx` | `pnpm test` |
| Aparência | aceite manual no navegador | 390, 768, 820, 1024 e 1440 px; roteiro registrado | `.specs/features/web-design/evidence/` | roteiro manual |

## Gate Check Commands

Todos rodam a partir de `web/`, com o código de saída capturado de verdade (no Windows: `cmd /v:on /c "<comando> & echo !ERRORLEVEL!"`). Um gate só passa com saída `0`.

| Level | When | Command |
| --- | --- | --- |
| Web | toda task | `pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build` |
| Contract | toda task | `pnpm lint:api && pnpm gen:api:check` |
| Backend intacto | toda task | `git diff --name-only develop -- api/` vazio |
| Dependências intactas | toda task | `git diff --name-only develop -- web/package.json web/pnpm-lock.yaml` vazio |
| Ownership | toda task | `git diff --name-only <base da unidade>` só contém caminhos da unidade |
| Audit | última task de cada unidade | `pnpm audit --audit-level high` |
| Specs e tasks | início da rodada | `validate_spec.py` nas cinco specs e `validate_tasks.py` nos cinco arquivos de tasks, saída `0` |

Nenhum gate pode ser enfraquecido, pulado ou mascarado. Teste não é apagado nem afrouxado para passar.

---

## Execution Plan

```
T1 → T2
T2 → T3
T3 → T4
T4 → T5
T5 → T6
T6 → T7
```

T1 a T4 (DS-F) são a rodada 1, num único agente. T5 a T7 (DS-E) são a rodada 5. Entre T4 e T5 acontecem as rodadas 2, 3 e 4. Depois de T7 há o **portão**: nada da rodada 6 começa sem a aprovação visual do mantenedor.

---

## Task Breakdown

### T1: Biblioteca de guarda de design

**Unidade**: DS-F
**What**: funções que leem arquivos de `web/` e apontam valores visuais soltos, para uso pelas guardas de cada unidade.
**Where**: `web/test/design/`
**Arquivos**: `web/test/design/guard.ts`, `web/test/design/guard.test.ts`, arquivos de exemplo em `web/test/design/fixtures/`
**Depends on**: None
**Reuses**: modelo de `web/app/server-client-boundary.test.ts` (leitura de arquivos, sem navegador).
**Requirement**: DSF-01 (AC7), DSF-03 (AC3 a AC5), DSF-04 (AC1 a AC5)

**Mudanças previstas**: `findRawValues(paths, options)` devolve arquivo, linha e regra violada. Regras: cor literal (hex, `rgb`, `hsl`, `oklch`); utilitário de paleta crua do Tailwind; valor arbitrário de cor, de tamanho de texto ou de espaço; `font-display` e `uppercase` fora dos arquivos permitidos; `transition-all`; classe `dark:`. A lista de exceções é explícita, com o motivo de cada uma.

**Done when**:
- [x] Cada regra tem um exemplo que passa e um que falha, e a falha informa arquivo e linha
- [x] A função aceita a lista de caminhos como parâmetro, para cada unidade guardar só os seus
- [x] Nenhuma guarda é aplicada ainda ao código existente nesta task (as unidades donas fazem isso)
- [x] Gates Web, Contract, Backend intacto, Dependências intactas e Ownership passam

**Tests**: unit
**Verificação visual**: não se aplica
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership
**Commit**: `test(web): biblioteca de guarda de valores visuais soltos`

---

### T2: Tokens em três camadas e teste de contraste

**Unidade**: DS-F
**What**: reescrever os tokens da folha global conforme o design, com o teste que calcula o contraste de cada par.
**Where**: `web/app/globals.css`
**Arquivos**: `web/app/globals.css`, `web/lib/design/contrast-pairs.ts`, `web/test/design/contrast.test.ts`, `web/test/design/tokens.test.ts`
**Depends on**: T1
**Reuses**: bloco `@theme inline` e nomes semânticos atuais; `tw-animate-css`; `shadcn/tailwind.css`.
**Requirement**: DSF-01 (AC1 a AC6), DSF-02 (AC1 a AC10), DSF-04 (AC1 a AC8)

**Mudanças previstas**: primitivos `--tj-*` (não expostos ao Tailwind); semânticos por grupo (`brand`, `surface`, `content`, `action`, `state`, `focus`) com os nomes do kit mantidos e os novos (`brand`, `success`, `warning`, `info`, `disabled`, `border-strong`, `primary-hover`, variantes `-subtle` e `-emphasis`); tokens de componente (alturas de controle, tabela, shell, larguras de conteúdo, margem de página) com ajuste sob `pointer: coarse`; raios 4, 6 e 8; sombras de sobreposição; durações e curvas; regra global de `prefers-reduced-motion`; `color-scheme: light`; remoção do bloco `.dark` e dos tokens `--chart-*`; `@custom-variant dark` mantido; `--font-sans` apontando para `--font-geist-sans` e `--font-display` declarado.

**Done when**:
- [x] Os quatro primitivos de marca têm exatamente `#111111`, `#D81E28`, `#FFFFFF` e `#A7A9AC`
- [x] O teste de contraste lê a folha global, resolve cada par de `contrast-pairs.ts` e exige 4,5:1 em texto e 3:1 em gráfico e borda de campo; todos os pares passam
- [x] O teste de tokens confirma: três camadas presentes; nenhum primitivo no `@theme`; todo semântico com valor; token de erro diferente do vermelho de marca; ausência de `.dark` e de `prefers-color-scheme`
- [x] O teste confirma a regra de `prefers-reduced-motion` e os tokens sob `pointer: coarse`
- [x] A guarda de T1 aplicada a `app/globals.css` só encontra as exceções declaradas (a própria folha de tokens)
- [x] Os testes existentes do projeto continuam passando sem alteração
- [x] Gates passam

**Tests**: unit (leitura do CSS)
**Verificação visual**: abrir `/entrar` e uma tela autenticada com `pnpm dev` e confirmar paleta nova sem quebra de layout (registro simples no relatório da task)
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership
**Commit**: `feat(web): tokens de design em camadas com a paleta oficial`

---

### T3: Fontes e metadados do documento

**Unidade**: DS-F
**What**: carregar a Barlow Condensed e declarar `theme-color` e `color-scheme` no layout raiz.
**Where**: `web/app/layout.tsx`
**Arquivos**: `web/app/layout.tsx`, `web/test/design/fonts.test.ts`
**Depends on**: T2
**Reuses**: carregamento atual de Geist e Geist Mono por `next/font/google`.
**Requirement**: DSF-03 (AC1, AC2), DSS-05 (AC1, AC2)

**Mudanças previstas**: `Barlow_Condensed` com pesos 600 e 700, subconjunto `latin`, `display: "swap"`, variável `--font-barlow-condensed`; `export const viewport` com `themeColor: "#111111"` e `colorScheme: "light"`; `lang`, título e descrição mantidos. `AppProviders` não muda.

**Done when**:
- [x] O teste confirma que o layout declara as três variáveis de fonte e que a folha global as liga a `--font-sans`, `--font-mono` e `--font-display`
- [x] `pnpm build` passa e nenhuma dependência entra em `package.json`
- [x] No navegador, a fonte computada de um parágrafo é Geist (pendência D-01 resolvida), e não a fonte padrão
- [x] O título continua no formato "Página | TJ Platform"
- [x] Gates passam

**Tests**: unit + build
**Verificação visual**: fonte computada de corpo e de um título de teste, conferida no navegador e registrada no relatório da task
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership
**Commit**: `feat(web): tipografia do design system e metadados de tema`

---

### T4: Escudo oficial, ícones e limpeza de `public/`

**Unidade**: DS-F
**What**: versionar o escudo com o enquadramento corrigido, derivar os ícones da aplicação e remover os arquivos de exemplo.
**Where**: `web/public/brand/`
**Arquivos**: `web/public/brand/torcida-jovem-logo.svg`, `web/app/icon.svg`, `web/app/apple-icon.png`, `web/app/favicon.ico` (remoção), `web/public/file.svg`, `globe.svg`, `next.svg`, `vercel.svg` e `window.svg` (remoção), `web/test/design/brand.test.ts`
**Depends on**: T3
**Reuses**: o arquivo recebido do escudo, sem alteração de desenho.
**Requirement**: DSF-05 (AC1, AC6, AC8), DSS-05 (AC3, AC4)

**Mudanças previstas**: só o atributo `viewBox` do SVG muda, para um quadrado centrado no círculo (`0 118.275 591.17 591.17`, a confirmar renderizando o arquivo). `app/icon.svg` é cópia do mesmo SVG. `app/apple-icon.png` é a rasterização do mesmo SVG em 180 por 180 px, feita com o navegador já instalado na máquina, sem ferramenta nova e sem edição.

**Done when**:
- [x] O teste compara, por hash, os atributos `d` e o bloco de estilos do SVG versionado com os do arquivo recebido, e falha com qualquer diferença de desenho
- [x] O teste confirma `viewBox` quadrado e PNG de 180 por 180 px
- [x] Renderizado, o escudo aparece inteiro, sem corte, em 32, 64 e 96 px
- [x] Nenhuma referência aos cinco SVGs de exemplo no código; `favicon.ico` padrão removido
- [x] IF o escudo ficar ilegível em 16 px THEN a necessidade de asset oficial reduzido é registrada em `STATE.md`, sem criar versão simplificada
- [x] Gates passam, incluindo Audit (fim da unidade DS-F)

**Tests**: unit
**Verificação visual**: aba do navegador com o ícone novo; escudo renderizado nos três tamanhos
**Gate**: Web + Audit + Contract + Backend intacto + Dependências intactas + Ownership
**Commit**: `feat(web): escudo oficial versionado e icones da aplicacao`

---

### T5: Espécime — rota e seções de fundação

**Unidade**: DS-E
**What**: criar a rota interna do espécime, indisponível em produção, com as seções de marca, cores, tipografia, espaço, raio, borda e sombra.
**Where**: `web/app/(dev)/design-system/`
**Arquivos**: `page.tsx`, componentes de seção e dados de exemplo dentro de `web/app/(dev)/design-system/`, `page.test.tsx`
**Depends on**: T4
**Reuses**: `BrandMark`, componentes de tipografia e todos os tokens; nenhum componente ou estilo exclusivo.
**Requirement**: DSF-06 (AC1, AC2, AC5)

**Mudanças previstas**: página que chama `notFound()` quando `NODE_ENV` é `production`; seções com âncoras; cada cor semântica mostrada com o par de contraste e a razão calculada a partir de `contrast-pairs.ts`; cada papel tipográfico com texto real em português, com acentos e caixa alta.

**Done when**:
- [ ] O teste confirma que a página responde como não encontrada com `NODE_ENV=production` e renderiza as seções fora de produção
- [ ] A guarda de design aplicada a `app/(dev)/` passa: nenhum valor solto, nenhum estilo próprio
- [ ] Nenhum componente novo foi criado fora de `components/ui` e `components/app`
- [ ] Gates passam

**Tests**: component
**Verificação visual**: seções de fundação conferidas em 390 e 1440 px
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership
**Commit**: `feat(web): especime do design system com as secoes de fundacao`

---

### T6: Espécime — componentes, padrões, shell e página de exemplo

**Unidade**: DS-E
**What**: completar o espécime com todos os componentes implementados, seus estados, os padrões compostos e o shell.
**Where**: `web/app/(dev)/design-system/`
**Arquivos**: seções adicionais e dados de exemplo dentro de `web/app/(dev)/design-system/`, `page.test.tsx` (ampliado)
**Depends on**: T5
**Reuses**: kit, compostos, `ShellFrame`, `DataList` com `layout` forçado.
**Requirement**: DSF-06 (AC1 a AC4), DSC-03 (AC9)

**Mudanças previstas**: seções Ações, Campos, Status e valores, Tabela e lista, Estados de tela, Feedback, Sobreposições, Shell e Página de exemplo, como no design. Cada componente interativo em repouso, foco, desabilitado, inválido e em andamento, quando o estado existir. Saldo em Barlow Condensed e em Geist lado a lado, para a decisão de DD-4. Radio, Switch, Tabs e Tooltip não aparecem.

**Done when**:
- [ ] Todos os componentes do inventário marcados como "usado na V1" aparecem, com seus estados
- [ ] A listagem de exemplo aparece como tabela completa, com colunas por prioridade e como lista empilhada; uma tabela pequena continua tabela
- [ ] O shell aparece com navegação, item ativo e menu da pessoa, com dados de exemplo e sem sessão
- [ ] O teste confirma a presença de cada seção e a ausência dos quatro componentes adiados
- [ ] Gates passam

**Tests**: component
**Verificação visual**: todas as seções conferidas em 390, 768, 820, 1024 e 1440 px
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership
**Commit**: `feat(web): especime com componentes, padroes e shell`

---

### T7: Evidência do espécime e portão de aprovação visual

**Unidade**: DS-E
**What**: registrar a conferência do espécime nas cinco larguras e submeter ao mantenedor. A execução para aqui.
**Where**: `.specs/features/web-design/evidence/especime.md`
**Arquivos**: `.specs/features/web-design/evidence/especime.md`
**Depends on**: T6
**Reuses**: roteiro de conferência visual da Web V1 (`.specs/features/web/smoke-visual.md`), com capturas feitas por script temporário fora do repositório.
**Requirement**: DSF-06 (AC6), DSF-03 (AC6), DSF-02 (AC5, AC6)

**Mudanças previstas**: arquivo de evidência com a lista de critérios de aprovação abaixo, o resultado de cada um e as observações. Nenhuma alteração de código.

**Critérios de aprovação do espécime** (cada um com resultado e observação):

1. Identidade: a interface é reconhecível como da Torcida Jovem e não como um painel genérico com vermelho.
2. Vermelho: aparece só como acento; nenhuma área de fundo vermelha; contagem de usos na página de exemplo.
3. Tipografia: Barlow Condensed só em títulos e rótulos de identidade; corpo em Geist; caixa alta sem excesso.
4. DD-4: leitura do saldo em Barlow Condensed comparada com Geist; decisão do mantenedor registrada.
5. Contraste: todos os pares exibidos acima do limite; foco visível no conteúdo e no shell.
6. Status e valores: cada status distinguível sem cor; `CANCELADA` neutra; receita, despesa e saldo negativo distintos por sinal e ícone.
7. Responsividade: 390, 768, 820, 1024 e 1440 px sem rolagem horizontal de página; tabela e lista nos três comportamentos; navegação fixa só a partir de 1024.
8. Toque: controles com 44 px em emulação de toque.
9. Movimento: com redução de movimento ativa, diálogo, gaveta e toast sem animação.
10. Marca: escudo íntegro, com respiro, sem distorção; nenhum elemento de marca inventado.

**Done when**:
- [ ] Os dez critérios conferidos num navegador real e registrados, sem senha, token ou dado pessoal
- [ ] Divergências entre o espécime e o design listadas, cada uma com proposta de correção
- [ ] Gates passam, incluindo Audit (fim da unidade DS-E)
- [ ] **A execução para.** A aprovação do mantenedor, com a decisão de DD-4, é registrada em `STATE.md` antes de qualquer task da rodada 6
- [ ] IF o mantenedor pedir ajustes THEN eles viram tasks de correção na unidade dona do arquivo, e o portão é refeito

**Tests**: aceite manual no navegador (roteiro registrado)
**Verificação visual**: é o objeto da task
**Gate**: Web + Audit + Contract + Backend intacto + Dependências intactas + Ownership
**Commit**: `docs(web-design): evidencia do especime para aprovacao visual`

---

## Phase Execution Map

```
T1 → T2
T2 → T3
T3 → T4
T4 → T5
T5 → T6
T6 → T7
```

Rodada 1: T1 a T4. Rodada 5: T5 a T7, seguidas do portão.

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1 | 1 biblioteca de teste e seus casos | ✅ Granular |
| T2 | 1 folha de tokens e os testes que a verificam | ✅ Aceitável (a folha é um único contrato; dividir deixaria tokens sem teste) |
| T3 | 1 arquivo de layout | ✅ Granular |
| T4 | 1 asset e os ícones derivados dele | ✅ Aceitável |
| T5 | 1 rota, seções de fundação | ✅ Granular |
| T6 | mesma rota, seções de componentes | ✅ Aceitável |
| T7 | 1 arquivo de evidência | ✅ Granular |

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | None | ✅ Match |
| T2 | T1 | T1 | ✅ Match |
| T3 | T2 | T2 | ✅ Match |
| T4 | T3 | T3 | ✅ Match |
| T5 | T4 | T4 | ✅ Match |
| T6 | T5 | T5 | ✅ Match |
| T7 | T6 | T6 | ✅ Match |

## Test Co-location Validation

| Task | Code Layer | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1 | guardas de design | unit | unit | ✅ OK |
| T2 | `app/globals.css` | unit (leitura do CSS) | unit (leitura do CSS) | ✅ OK |
| T3 | `app/layout.tsx` | unit + build | unit + build | ✅ OK |
| T4 | marca | unit | unit | ✅ OK |
| T5 | rota do espécime | component | component | ✅ OK |
| T6 | rota do espécime | component | component | ✅ OK |
| T7 | aparência | aceite manual | aceite manual | ✅ OK |
