# Migração da Web V1 — Tasks

## Execution Protocol (MANDATORY — do not skip)

Implementar com a skill `tlc-spec-driven`: ativar pelo nome e seguir o fluxo de Execute e as Critical Rules. Testes primeiro, derivados dos critérios de aceite; um commit atômico por task; a task é marcada como concluída neste arquivo antes do commit e no mesmo commit. **Se a skill não puder ser ativada, PARE e avise.**

Antes de escrever código Next: ler `web/AGENTS.md` e a documentação do pacote instalado em `web/node_modules/next/dist/docs/`.

Autorização: a aprovação destas tasks autoriza implementação e commits locais na branch da unidade. `git push`, PR, merge e deploy exigem autorização explícita para cada ação.

---

**Spec**: `.specs/features/web-design/spec/05-migracao-web-v1.md`
**Design**: `.specs/features/web-design/design/05-migracao-web-v1.md`
**Status**: Draft (2026-10-11), aguardando aprovação do mantenedor.

**Pré-requisito**: **portão do espécime aprovado pelo mantenedor** e registrado em `STATE.md`. Sem isso, nenhuma task deste arquivo começa.

**Unidades de execução** (propriedade e DAG em `design/05-migracao-web-v1.md`):

| Unidade | Tasks | Rodada | Entrada | Saída |
| --- | --- | --- | --- | --- |
| DS-ID Identity | T1, T2 | 6 | Portão aprovado; base congelada | Telas de usuários migradas; USR-01 a USR-05 preservados |
| DS-FA Financeiro A | T3, T4 | 6 | idem | Plano de contas e saldo migrados; FWB-01 e FWB-02 preservados |
| DS-FB Financeiro B | T5 a T7 | 6 | idem | Lançamentos e comprovantes migrados; FWB-03 a FWB-05 preservados |
| DS-ES Estoque | T8 a T10 | 6 | idem | Produtos, movimentações e ajuste migrados; EWB-01 a EWB-04 preservados |
| DS-INT Integração | T11 a T13 | 7 | As cinco frentes da rodada 6 integradas (inclui DS-S, `tasks/04-shell-e-acesso.md`) | Alias removidos, guarda global, guia, smoke completo |
| VERIFY | T14 | 7 | DS-INT concluída | `validation.md` com veredito |

As quatro frentes de módulo e DS-S rodam em paralelo, cada uma num agente e num worktree próprios. São independentes: não há `import` real entre elas (conferido em 2026-10-11 com o Graphify local 0.9.86, arestas `EXTRACTED`, e com os `import` do código; as duas arestas `INFERRED` entre frentes eram coincidência de nome entre testes).

## Restrições comuns a todas as tasks deste arquivo

- Cada frente altera só os próprios diretórios (gate Propriedade). `package.json` e `pnpm-lock.yaml` não mudam.
- **Arquivos congelados** para as frentes da rodada 6: `components/ui/`, `components/app/`, `lib/url-state/`, `lib/responsive/`, `test/` e os mocks de `lib/session`. Necessidade de mudança compartilhada: parar a task, reportar ao coordenador e aguardar a unidade de integração. Fixtures específicas do módulo vão para o `test-helpers.ts` do próprio módulo.
- **Comportamento intocado**: hooks de dados, clientes da API, catálogos de erro, filtros puros, regras de permissão, conversão de dinheiro e de quantidade não mudam. Muda a casca: contêiner, cabeçalho, origem do estado dos filtros, componente de listagem e aparência.
- Nenhuma mudança em `api/`, em contratos, em `lib/api/`, `lib/forms/` ou `lib/session/`.
- Teste de comportamento não é removido nem enfraquecido. Toda alteração de teste cita a decisão ou o requisito que a motiva e entra no relatório da unidade (DSM-01 AC4).
- As observações de produto do smoke visual (conta inativa com "Nova subconta", lançamento cancelado com "Anexar") não são corrigidas aqui.

## Test Coverage Matrix

> Gerada do código, das diretrizes do projeto e da spec. Diretrizes consultadas: `CLAUDE.md`, `web/AGENTS.md`, `docs/development/web.md`, `web/vitest.config.ts`, `.specs/features/web/tasks/*.md`.

| Code Layer | Required Test Type | Coverage Expectation | Location Pattern | Run Command |
| --- | --- | --- | --- | --- |
| Telas de módulo (`features/`, `app/(app)/<módulo>`) | component com MSW | Testes de comportamento existentes passando; cada AC de DSM e DSP aplicável coberto (URL, lista empilhada, nome de ação, estados) | `web/features/**/*.test.tsx`, `web/app/**/*.test.tsx` | `pnpm test` |
| Guarda da frente | unit | Nenhum valor solto nos diretórios da frente | `web/features/**/*.guard.test.ts` | `pnpm test` |
| Guarda global e remoção de alias | unit | Nenhum valor solto em `app`, `components` e `features`; nenhum alias em uso | `web/test/design/*.test.ts` | `pnpm test` |
| Documentação e revisão estrutural | revisão registrada | Guia cobre o fluxo de tela nova; relações conferidas com os `import` reais | `docs/development/web.md`, `.specs/features/web-design/evidence/` | revisão manual |
| Aparência e fluxo | aceite manual no navegador | Todas as telas em 390, 768, 820, 1024 e 1440 px; teclado; zoom; movimento reduzido | `.specs/features/web-design/evidence/` | roteiro manual |
| Verificação final | verificador independente | Cada AC com evidência; sensor de discriminação; revisão `web-design-guidelines` | `.specs/features/web-design/validation.md` | `validate_state.py` |

## Gate Check Commands

Todos rodam a partir de `web/`, com o código de saída capturado de verdade. Um gate só passa com saída `0`.

| Level | When | Command |
| --- | --- | --- |
| Web | toda task | `pnpm lint && pnpm exec next typegen && pnpm exec tsc --noEmit && pnpm test && pnpm build` |
| Contract | toda task | `pnpm lint:api && pnpm gen:api:check` |
| Backend intacto | toda task | `git diff --name-only develop -- api/` vazio |
| Dependências intactas | toda task | `git diff --name-only develop -- web/package.json web/pnpm-lock.yaml` vazio |
| Ownership | toda task | `git diff --name-only <base da unidade>` só contém caminhos da unidade |
| Congelados intactos | T1 a T10 | `git diff --name-only <base da unidade> -- web/components web/lib web/test` vazio |
| Audit | última task de cada unidade | `pnpm audit --audit-level high` |
| Revisão de interface | T13 e T14 | skill `web-design-guidelines` sobre `web/app`, `web/components` e `web/features` |
| Fechamento | T14 | `validate_state.py` com saída `0` |

Nenhum gate pode ser enfraquecido, pulado ou mascarado. Teste não é apagado nem afrouxado para passar.

---

## Execution Plan

```
T1 → T2
T3 → T4
T5 → T6
T6 → T7
T8 → T9
T9 → T10
T2 → T11
T4 → T11
T7 → T11
T10 → T11
T11 → T12
T12 → T13
T13 → T14
```

Quatro cadeias independentes na rodada 6 (T1 e T2; T3 e T4; T5 a T7; T8 a T10), cada uma de um agente. T11 também depende de DS-S concluída (`tasks/04-shell-e-acesso.md`). T11 a T13 são do coordenador ou de um agente de integração. T14 é de um agente independente dos autores.

---

## Task Breakdown

### T1: Usuários — listagem

**Unidade**: DS-ID
**What**: migrar a tela de usuários para cabeçalho de página, barra de filtros com estado na URL, listagem responsiva e ações de linha em menu.
**Where**: `web/features/identity/usuarios/users-page.tsx`
**Arquivos**: `web/features/identity/usuarios/users-page.tsx`, `users-table.tsx`, mapas de rótulo e status do diretório, `web/features/identity/nav.ts`, `web/app/(app)/admin/` e os testes desses arquivos
**Depends on**: None
**Reuses**: hooks e filtros atuais (`active`, `role`, `next_cursor`); `PageContainer`, `PageHeader`, `FilterBar`, `DataList`, `RowActions`, `ShowMore`, `NativeSelect`, `StatusBadge`, `RoleBadge`, `useQueryState`.
**Requirement**: DSM-02 (AC1 a AC3, AC5), DSP-05 (AC1 a AC7, AC10, AC11, AC14), DSP-02 (AC1 a AC5), USR-01 (preservado)

**Mudanças previstas**: página em largura larga com `Suspense`; filtros `situacao` e `papel` na URL, ainda enviados à API como `active` e `role`; colunas por prioridade (nome e situação; e-mail e papéis; acesso administrativo); menu "Ações de {nome}"; "Mostrar mais" pelo cursor; estados vazio, sem resultado e erro; ícone e descrição em `nav.ts`; os dois `<select>` avulsos saem.

**Done when**:
- [ ] Os testes existentes da tela passam; mudam só as leituras de filtro, agora pela URL (DSP-05 AC2)
- [ ] Mudar um filtro atualiza a URL; abrir a URL com filtro restaura a consulta; valor inválido é ignorado
- [ ] Com `mockViewport` estreito, a tela mostra lista empilhada com todas as ações acessíveis
- [ ] O menu de cada usuário tem nome acessível que identifica o usuário; a própria linha continua sem ações
- [ ] Nenhum `<select>` fora do kit e nenhum contêiner ou título montado à mão
- [ ] Gates passam

**Tests**: component com MSW
**Verificação visual**: tela em 390, 768, 820, 1024 e 1440 px, comparada com o espécime
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Congelados intactos
**Commit**: `feat(web): usuarios no design system com filtros na url e lista responsiva`

---

### T2: Usuários — diálogos

**Unidade**: DS-ID
**What**: migrar os diálogos de criar usuário, acesso administrativo, situação e senha temporária para o layout de formulário e a confirmação padrão.
**Where**: `web/features/identity/usuarios/dialogs/`
**Arquivos**: os quatro arquivos de `web/features/identity/usuarios/dialogs/`, seus testes, `web/features/identity/usuarios/tokens.guard.test.ts`
**Depends on**: T1
**Reuses**: lógica e mensagens atuais; `FormLayout`, `Field`, `ConfirmDialog` (com `subject` e `consequence`), `Button` nas variantes novas.
**Requirement**: DSM-02 (AC5), DSP-03 (AC1 a AC5), DSP-04 (AC1 a AC9), USR-02 a USR-05 (preservados)

**Mudanças previstas**: formulários em coluna única com `FormLayout`; confirmações nomeiam o usuário e a consequência; botões que abrem ação destrutiva passam a `secondary`, e só o botão final da confirmação é `destructive`; `spellCheck={false}` no e-mail; diálogos em tela cheia abaixo de 640 px. A guarda cobre os diretórios de DS-ID.

**Done when**:
- [ ] Os testes existentes dos diálogos passam sem alteração de asserção de comportamento (validações, e-mail duplicado, não escalada, senha temporária)
- [ ] Cada confirmação mostra o nome do usuário, a consequência e um botão com o verbo da ação
- [ ] Nenhum uso de alias de variante de botão resta nos arquivos de DS-ID
- [ ] A guarda passa nos diretórios de DS-ID
- [ ] Gates passam, incluindo Audit (fim da unidade DS-ID)

**Tests**: component com MSW
**Verificação visual**: os quatro diálogos em 390 e 1440 px
**Gate**: Web + Audit + Contract + Backend intacto + Dependências intactas + Ownership + Congelados intactos
**Commit**: `feat(web): dialogos de usuarios nos padroes de formulario e confirmacao`

---

### T3: Plano de contas

**Unidade**: DS-FA
**What**: migrar a árvore do plano de contas e seus diálogos para a estrutura de página, os estados e os componentes do Design System.
**Where**: `web/features/financeiro/contas/contas-page.tsx`
**Arquivos**: `web/features/financeiro/contas/` (página, diálogos, mapas), `web/features/financeiro/nav.ts`, `web/app/(app)/financeiro/contas/` e os testes desses arquivos
**Depends on**: None
**Reuses**: `tree.ts`, hooks e catálogo de erros atuais; `PageContainer`, `PageHeader`, `EmptyState`, `StatusBadge`, `ConfirmDialog`, `FormLayout`, `NativeSelect`, `RowActions`.
**Requirement**: DSM-03 (AC1, AC7, AC8), DSP-01, DSP-02, DSP-03, FWB-01 (preservado)

**Mudanças previstas**: árvore com recuo por nível em qualquer largura (não usa `DataList`); status ativo e inativo com ícone e texto; estado vazio com orientação e ação "Nova conta" para quem pode; confirmação de desativar com o nome da conta e a consequência; ícones e descrições em `nav.ts`.

**Done when**:
- [ ] Os testes existentes de contas passam sem alteração de asserção de comportamento (tipo fixado pelo pai, desativação, árvore)
- [ ] A hierarquia continua legível em 320 px, por recuo, sem rolagem horizontal
- [ ] Conta inativa é distinguível sem depender de cor
- [ ] Gates passam

**Tests**: component com MSW
**Verificação visual**: tela vazia e com árvore em 390, 768, 820, 1024 e 1440 px
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Congelados intactos
**Commit**: `feat(web): plano de contas no design system`

---

### T4: Saldo

**Unidade**: DS-FA
**What**: migrar a tela de saldo para o valor de destaque padronizado.
**Where**: `web/features/financeiro/saldo/saldo-page.tsx`
**Arquivos**: `web/features/financeiro/saldo/`, `web/app/(app)/financeiro/saldo/`, seus testes, `web/features/financeiro/contas/tokens.guard.test.ts`
**Depends on**: T3
**Reuses**: hook de saldo atual; `PageContainer` estreito, `PageHeader`, `Card`, `MoneyValue` (`balance`, `display`), estados.
**Requirement**: DSM-03 (AC1, AC2, AC8), DSC-06 (AC1 a AC3), FWB-02 (preservado)

**Mudanças previstas**: valor com `MoneyValue` na fonte decidida no portão (DD-4); texto explicativo mantido; saldo negativo com sinal, ícone e texto. A guarda cobre os diretórios de DS-FA.

**Done when**:
- [ ] Os testes existentes de saldo passam sem alteração de asserção de comportamento
- [ ] Saldo negativo de exemplo mostra sinal e ícone, e é anunciado como negativo
- [ ] A guarda passa nos diretórios de DS-FA
- [ ] Gates passam, incluindo Audit (fim da unidade DS-FA)

**Tests**: component com MSW
**Verificação visual**: saldo positivo e negativo em 390 e 1440 px
**Gate**: Web + Audit + Contract + Backend intacto + Dependências intactas + Ownership + Congelados intactos
**Commit**: `feat(web): saldo com valor de destaque do design system`

---

### T5: Lançamentos — listagem

**Unidade**: DS-FB
**What**: migrar a listagem de lançamentos para barra de filtros na URL, listagem responsiva e status distintos.
**Where**: `web/features/financeiro/lancamentos/lancamentos-page.tsx`
**Arquivos**: `web/features/financeiro/lancamentos/lancamentos-page.tsx`, `labels.ts`, `web/app/(app)/financeiro/lancamentos/page.tsx` e os testes desses arquivos
**Depends on**: None
**Reuses**: `filterLancamentos`, hooks e rótulos atuais; `FilterBar`, `DataList`, `ShowMore`, `NativeSelect`, `StatusBadge`, `MoneyValue`, `useQueryState`.
**Requirement**: DSM-03 (AC1 a AC6, AC8), DSP-05 (AC1 a AC13), FWB-03 (AC1, preservado)

**Mudanças previstas**: filtros `tipo`, `status`, `conta`, `de` e `ate` na URL; colunas por prioridade (data, conta, valor líquido, status; tipo e forma de pagamento; valor bruto e taxa); mapa de status em `labels.ts` com tom e ícone (`CANCELADA` neutra); link de detalhe com nome que identifica o lançamento; "Mostrar mais" em lotes de 50; página em largura larga com `Suspense`.

**Done when**:
- [x] Os testes existentes da listagem passam; mudam só o nome acessível do link de detalhe (DSP-05 AC10) e a leitura dos filtros pela URL (DSP-05 AC2)
- [x] `CRIADA`, `RECEBIDA`, `PAGA` e `CANCELADA` têm rótulo e ícone distintos entre si
- [x] Um filtro restaurado da URL que aponta para conta inexistente é ignorado
- [x] Com 51 lançamentos aparecem 50 e "Mostrar mais"; com 50, o controle não aparece; o filtro considera todos os registros
- [x] Com `mockViewport` estreito, cada item mostra data, tipo, conta, valor líquido e status
- [x] Gates passam

**Tests**: component com MSW
**Verificação visual**: listagem vazia, filtrada e cheia em 390, 768, 820, 1024 e 1440 px
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Congelados intactos
**Commit**: `feat(web): lancamentos com filtros na url e lista responsiva`

---

### T6: Lançamentos — formulário, detalhe e workflow

**Unidade**: DS-FB
**What**: migrar o formulário, a tela de detalhe e as ações de workflow do lançamento.
**Where**: `web/features/financeiro/lancamentos/lancamento-detail.tsx`
**Arquivos**: `web/features/financeiro/lancamentos/lancamento-form.tsx`, `devolucao-form.tsx`, `lancamento-detail.tsx`, `select-field.tsx` (remoção), `web/app/(app)/financeiro/lancamentos/[id]/` e os testes desses arquivos
**Depends on**: T5
**Reuses**: `actions.ts`, hooks e catálogo de erros atuais; `FormLayout`, `Field`, `NativeSelect`, `MoneyField`, `Breadcrumb` (via `PageHeader`), `SectionTitle`, `MoneyValue`, `ConfirmDialog`.
**Requirement**: DSM-03 (AC1 a AC3, AC8), DSP-03 (AC1 a AC6), DSP-04, DSP-07 (AC7), DSM-05 (AC2, AC3), FWB-03 e FWB-04 (preservados)

**Mudanças previstas**: formulário em `FormLayout` com `NativeSelect` do kit; `select-field.tsx` e o filtro de data com rótulo manual são removidos; detalhe com trilha, status no cabeçalho, valores com `MoneyValue` (receita com "+", despesa com "−"), identificador como tipo e código curto; ações de workflow com confirmação que nomeia o lançamento e a consequência; cancelamento continua exigindo motivo.

**Done when**:
- [x] Os testes existentes de formulário, detalhe e workflow passam sem alteração de asserção de comportamento (prévia do líquido, dinheiro inválido, devolução só com contas de despesa ativas, motivo obrigatório)
- [x] `select-field.tsx` não existe mais e nenhum `<select>` avulso resta em DS-FB
- [x] Confirmar recebimento, pagamento, devolução e cancelamento mostram o registro e a consequência, com o verbo da ação no botão
- [x] Gates passam

**Tests**: component com MSW
**Verificação visual**: formulário e detalhe (criado, recebido, cancelado) em 390, 820 e 1440 px
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Congelados intactos
**Commit**: `feat(web): formulario, detalhe e workflow de lancamentos no design system`

---

### T7: Comprovantes

**Unidade**: DS-FB
**What**: migrar o painel de comprovantes e ajustar o texto do limite de anexo.
**Where**: `web/features/financeiro/comprovantes/comprovantes-panel.tsx`
**Arquivos**: `web/features/financeiro/comprovantes/`, seus testes, `web/features/financeiro/lancamentos/tokens.guard.test.ts`
**Depends on**: T6
**Reuses**: hooks, `limits.ts` e catálogo de erros atuais; `SectionTitle`, `DataList` (`layout="table"` se couber em 320 px), `EmptyState`, `Alert`.
**Requirement**: DSM-03 (AC1, AC8), DSP-07 (AC6), DSP-05 (AC8), FWB-05 (preservado; AC3 com texto revisado)

**Mudanças previstas**: lista de comprovantes no padrão de listagem; estado vazio com orientação e ação para quem pode anexar; o limite exibido passa a megabytes inteiros arredondados para baixo ("até 9 MB"); o limite real em bytes não muda. A guarda cobre os diretórios de DS-FB.

**Done when**:
- [ ] Os testes existentes passam; muda só o texto do limite, citando DSP-07 AC6
- [ ] `MAX` em bytes de `limits.ts` continua igual (teste confirma o valor)
- [ ] Baixar continua pedindo a URL assinada só no clique
- [ ] A guarda passa nos diretórios de DS-FB
- [ ] Gates passam, incluindo Audit (fim da unidade DS-FB)

**Tests**: component com MSW
**Verificação visual**: painel vazio e com anexos em 390 e 1440 px
**Gate**: Web + Audit + Contract + Backend intacto + Dependências intactas + Ownership + Congelados intactos
**Commit**: `feat(web): comprovantes no design system com limite de anexo em texto natural`

---

### T8: Produtos — listagem e criação

**Unidade**: DS-ES
**What**: migrar a listagem de produtos, com busca na URL, e o diálogo de criação.
**Where**: `web/features/estoque/produtos/produtos-page.tsx`
**Arquivos**: `web/features/estoque/produtos/`, `web/features/estoque/nav.ts`, `web/app/(app)/estoque/produtos/page.tsx` e os testes desses arquivos
**Depends on**: None
**Reuses**: `searchProdutos`, hooks e catálogo de erros atuais; `FilterBar`, `DataList`, `ShowMore`, `EmptyState`, `FormLayout`, `useQueryState`.
**Requirement**: DSM-04 (AC1, AC2, AC7), DSP-05 (AC1 a AC8, AC12, AC13), DSP-04, EWB-01 (preservado)

**Mudanças previstas**: busca `busca` na URL, com escrita adiada; colunas por prioridade (código, nome, situação; unidade); estado vazio com ação "Novo produto" para quem pode; diálogo de criação em `FormLayout`; ícone e descrição em `nav.ts`; página com `Suspense`.

**Done when**:
- [ ] Os testes existentes passam; mudam só as leituras do termo de busca, pela URL (DSP-05 AC2)
- [ ] Digitar na busca atualiza a URL sem uma entrada de histórico por tecla
- [ ] Código duplicado continua aparecendo no campo
- [ ] Com `mockViewport` estreito, a tela mostra lista empilhada
- [ ] Gates passam

**Tests**: component com MSW
**Verificação visual**: listagem vazia, com busca e cheia em 390, 768, 820, 1024 e 1440 px
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Congelados intactos
**Commit**: `feat(web): produtos com busca na url e lista responsiva`

---

### T9: Detalhe do produto e movimentações

**Unidade**: DS-ES
**What**: migrar o detalhe do produto, o histórico de movimentações e o diálogo de movimentação.
**Where**: `web/features/estoque/movimentacoes/movimentacoes-panel.tsx`
**Arquivos**: `web/features/estoque/produto-detail.tsx`, `web/features/estoque/movimentacoes/`, `web/app/(app)/estoque/produtos/[id]/` e os testes desses arquivos
**Depends on**: T8
**Reuses**: hooks, `payload.ts` e rótulos atuais; `PageHeader` com trilha, `SectionTitle`, `DataList`, `RowActions`, `ShowMore`, `QuantityField`, `StatusBadge`.
**Requirement**: DSM-04 (AC1, AC3 a AC7), DSP-05 (AC5 a AC11), EWB-02 (preservado)

**Mudanças previstas**: saldo em destaque com algarismos tabulares e indicação de saldo negativo por ícone e texto; histórico com colunas por prioridade (tipo, quantidade, data; origem, responsável; motivo, devolução de) e lista empilhada no celular; "Devolver" com nome acessível que identifica a movimentação; diálogo em tela cheia abaixo de 640 px, com as ações fixas na base.

**Done when**:
- [ ] Os testes existentes passam; muda só o nome acessível do botão "Devolver" (DSP-05 AC10), que continua contendo o texto visível
- [ ] Quantidade decimal continua recusada; origem continua fixa em "Inventário"; saída acima do saldo continua mostrando a mensagem atual
- [ ] Com `mockViewport` estreito, cada movimentação mostra tipo, quantidade, data e responsável, com o motivo limitado a três linhas
- [ ] Gates passam

**Tests**: component com MSW
**Verificação visual**: detalhe com histórico e diálogo de movimentação em 390 px com emulação de toque, e em 1440 px
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership + Congelados intactos
**Commit**: `feat(web): detalhe do produto e movimentacoes no design system`

---

### T10: Ajuste de estoque

**Unidade**: DS-ES
**What**: migrar o diálogo de ajuste para o layout de formulário, com o aviso de saldo negativo no tom de aviso.
**Where**: `web/features/estoque/ajustes/ajuste-dialog.tsx`
**Arquivos**: `web/features/estoque/ajustes/`, seus testes, `web/features/estoque/tokens.guard.test.ts`
**Depends on**: T9
**Reuses**: lógica atual do ajuste; `FormLayout`, `QuantityField`, `Textarea`, `Alert` no tom `warning`.
**Requirement**: DSM-04 (AC1, AC5 a AC7), DSP-03 (AC6), EWB-03 e EWB-04 (preservados)

**Mudanças previstas**: prévia do saldo com algarismos tabulares; aviso de saldo negativo como alerta de aviso, com ícone e texto, sem bloquear; botão desabilitado sem quantidade ou motivo, como hoje. A guarda cobre os diretórios de DS-ES.

**Done when**:
- [ ] Os testes existentes do ajuste passam sem alteração de asserção de comportamento
- [ ] O aviso de saldo negativo não depende só de cor
- [ ] Conta só com leitura (`CONSELHO_FISCAL`) não vê ação de escrita em nenhuma tela de estoque
- [ ] A guarda passa nos diretórios de DS-ES
- [ ] Gates passam, incluindo Audit (fim da unidade DS-ES)

**Tests**: component com MSW
**Verificação visual**: diálogo de ajuste com e sem aviso, em 390 e 1440 px
**Gate**: Web + Audit + Contract + Backend intacto + Dependências intactas + Ownership + Congelados intactos
**Commit**: `feat(web): ajuste de estoque no design system`

---

### T11: Remoção dos alias e guarda global

**Unidade**: DS-INT
**What**: remover os alias de compatibilidade do kit e dos compostos e aplicar a guarda de design a todo o front.
**Where**: `web/test/design/`
**Arquivos**: `web/test/design/global.guard.test.ts`; em `web/components/ui/`: `button.tsx`, `badge.tsx`, `alert.tsx`, `table.tsx`; em `web/components/app/`: `status-badge.tsx`, `states.tsx`, `confirm-dialog.tsx`; os testes desses arquivos
**Depends on**: T2, T4, T7, T10
**Reuses**: biblioteca de guarda de DS-F; relatórios das frentes com a lista de testes alterados.
**Requirement**: DSM-05 (AC1 a AC5), DSF-01 (AC7), DSC-05 (AC4), DSM-01 (AC4, AC5)

**Pré-requisito externo**: DS-S concluída e as cinco frentes integradas pelo coordenador, sem conflito de merge e com os gates completos passando.

**Mudanças previstas**: saem os alias `default`, `outline`, `icon` e `icon-sm` do botão, as variantes antigas de Badge e Alert, o modo legado de rolagem da tabela e a forma antiga de `StatusBadge`; a guarda global cobre `app`, `components` e `features`, e acrescenta as verificações de `<select>` fora do kit e de contêiner ou título de página montado à mão.

**Done when**:
- [ ] `tsc` passa sem nenhum uso de alias (o compilador aponta qualquer resto)
- [ ] A guarda global passa, com a lista de exceções consolidada e justificada
- [ ] A lista consolidada de testes alterados no ciclo, com a decisão de cada um, é registrada em `evidence/testes-alterados.md`, e nenhum teste foi removido
- [ ] Gates passam

**Tests**: unit + component
**Verificação visual**: nenhuma tela mudou de aparência com a remoção (conferência rápida de três telas)
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership
**Commit**: `refactor(web): remove alias de compatibilidade e aplica guarda global de design`

---

### T12: Guia de desenvolvimento e revisão estrutural

**Unidade**: DS-INT
**What**: documentar o Design System no guia da Web e fazer a revisão estrutural com o grafo atualizado.
**Where**: `docs/development/web.md`
**Arquivos**: `docs/development/web.md`, `.specs/features/web-design/evidence/revisao-estrutural.md`
**Depends on**: T11
**Reuses**: `docs/development/ai-environment.md` (uso do Graphify local 0.9.86); diagnóstico de 2026-10-11.
**Requirement**: INT-02 (atualizado), DSM-01 (AC2), WDS-D-029

**Mudanças previstas**: seção no guia com tokens, camadas de componentes, guardas, como abrir o espécime e como adicionar uma tela nova; grafo regenerado pelo coordenador com o executável local, versão conferida; revisão registrada: nenhuma dependência nova entre módulos, `components/app` sem importação de `features` além de `nav-items.ts`, `components/ui` dependendo só de `lib/utils`, relações conferidas com os `import` reais.

**Done when**:
- [ ] O guia descreve o fluxo para uma tela nova usando só kit, compostos e padrões
- [ ] A revisão estrutural lista as relações conferidas e declara os limites do grafo (arestas `INFERRED` descartadas, `.d.ts` fora do grafo)
- [ ] Nenhum hook, MCP remoto ou serviço novo foi configurado
- [ ] Gates passam

**Tests**: revisão registrada (documental e estrutural)
**Verificação visual**: não se aplica
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Ownership
**Commit**: `docs(web): guia do design system e revisao estrutural`

---

### T13: Smoke completo no navegador

**Unidade**: DS-INT
**What**: conferir todas as telas num navegador real, num build de produção, e registrar a evidência.
**Where**: `.specs/features/web-design/evidence/smoke-visual.md`
**Arquivos**: `.specs/features/web-design/evidence/smoke-visual.md`
**Depends on**: T12
**Reuses**: procedimento de `.specs/features/web/smoke-visual.md` e de `docs/development/web.md` (stack local com PostgreSQL e armazenamento S3 avulso, contas fictícias, script temporário fora do repositório).
**Requirement**: DSM-06 (AC1 a AC5, AC7), DSS-02 (AC7), INT-03 (refeito para este ciclo)

**Mudanças previstas**: nenhum código. Roteiro por tela, nas cinco larguras; percursos só por teclado; ampliação a 200% e 320 px; movimento reduzido; comparação de cada tela com o espécime aprovado; revisão `web-design-guidelines` pelo coordenador como prévia do gate do verificador.

**Done when**:
- [ ] Todas as telas da matriz do design conferidas em 390, 768, 820, 1024 e 1440 px, sem rolagem horizontal de página
- [ ] Percurso só por teclado: login, criação de lançamento, cancelamento com motivo e movimentação de estoque
- [ ] Console do navegador sem aviso de hidratação em nenhuma tela
- [ ] Visão de `CONSELHO_FISCAL` sem nenhuma ação de escrita; "Sem acesso" em `/admin/usuarios`
- [ ] Divergências em relação ao espécime viram tasks de correção na unidade dona, antes de T14
- [ ] O arquivo não contém senha, token, cookie nem dado pessoal; o que não for executável localmente (redefinição com token válido) é declarado
- [ ] Gates passam, incluindo Audit (fim da unidade DS-INT)

**Tests**: aceite manual no navegador (roteiro registrado)
**Verificação visual**: é o objeto da task
**Gate**: Web + Audit + Contract + Backend intacto + Dependências intactas + Ownership + Revisão de interface
**Commit**: `docs(web-design): evidencia do smoke visual completo`

---

### T14: Verificação independente

**Unidade**: VERIFY
**What**: verificação final por um agente que não participou da implementação, com relatório e veredito.
**Where**: `.specs/features/web-design/validation.md`
**Arquivos**: `.specs/features/web-design/validation.md`
**Depends on**: T13
**Reuses**: estratégia de `design/05-migracao-web-v1.md` ("Estratégia de Verification") e de `.specs/features/web/design/06-integracao.md`.
**Requirement**: DSM-06 (AC1 a AC7), DSM-01 (AC1 a AC6), INT-04 (refeito para este ciclo)

**Mudanças previstas**: nenhum código de produção. Entrada: as cinco specs, os designs, o intervalo de commits e `STATE.md`; nunca os relatórios dos autores como prova. O verificador: (1) checa cada AC de DSF, DSC, DSP, DSS e DSM contra teste ou evidência, com `arquivo:linha`; (2) confirma que os testes da Web V1 passam e que só os testes listados mudaram; (3) reverifica FND-05 AC2 no navegador; (4) executa a skill `web-design-guidelines` como **gate obrigatório**; (5) roda o sensor de discriminação com as 14 falhas mínimas do design, num worktree temporário; (6) repete parte do smoke (login, um ciclo do financeiro, visão de `CONSELHO_FISCAL`); (7) confirma que `api/`, contratos, permissões e regras de negócio não mudaram.

**Done when**:
- [ ] `validation.md` tem veredito, evidência por AC, resultado do sensor por falha, saída da `web-design-guidelines` e o intervalo de commits
- [ ] Cada achado da `web-design-guidelines` está corrigido ou casado com um conflito registrado em `STATE.md`; achado sem um dos dois reprova
- [ ] Toda falha do sensor que sobreviveu virou task de correção na unidade dona
- [ ] O worktree do sensor foi descartado e `git status --porcelain` é igual ao de antes
- [ ] `validate_state.py` passa
- [ ] Lições registradas pelo `lessons.py` só para falhas reais
- [ ] IF o veredito for FAIL THEN as lacunas viram tasks de correção e a verificação é refeita, no limite de três iterações antes de escalar ao mantenedor

**Tests**: verificador independente (spec, sensor, smoke, `web-design-guidelines`)
**Verificação visual**: reverificação de FND-05 AC2 e amostra do smoke
**Gate**: Web + Contract + Backend intacto + Dependências intactas + Revisão de interface + Fechamento
**Commit**: `docs(web-design): relatorio de verificacao independente`

---

## Phase Execution Map

```
T1 → T2
T3 → T4
T5 → T6
T6 → T7
T8 → T9
T9 → T10
T2 → T11
T4 → T11
T7 → T11
T10 → T11
T11 → T12
T12 → T13
T13 → T14
```

Rodada 6: quatro cadeias em paralelo (mais DS-S). Rodada 7: T11 a T13, depois T14.

## Task Granularity Check

| Task | Scope | Status |
| --- | --- | --- |
| T1 | 1 tela de listagem | ✅ Granular |
| T2 | 4 diálogos do mesmo recurso | ✅ Aceitável |
| T3 | 1 tela e seus diálogos | ✅ Aceitável |
| T4 | 1 tela pequena | ✅ Granular |
| T5 | 1 tela de listagem | ✅ Granular |
| T6 | formulário e detalhe do mesmo recurso (o detalhe abre o formulário) | ✅ Aceitável |
| T7 | 1 painel | ✅ Granular |
| T8 | 1 listagem e seu diálogo de criação | ✅ Aceitável |
| T9 | 1 tela de detalhe e seu histórico | ✅ Aceitável |
| T10 | 1 diálogo | ✅ Granular |
| T11 | remoção coordenada de alias (precisa ser atômica para compilar) | ✅ Aceitável |
| T12 | 1 guia e 1 registro de revisão | ✅ Granular |
| T13 | 1 roteiro e 1 arquivo de evidência | ✅ Granular |
| T14 | 1 relatório de verificação | ✅ Granular |

## Diagram-Definition Cross-Check

| Task | Depends On (task body) | Diagram Shows | Status |
| --- | --- | --- | --- |
| T1 | None | None | ✅ Match |
| T2 | T1 | T1 | ✅ Match |
| T3 | None | None | ✅ Match |
| T4 | T3 | T3 | ✅ Match |
| T5 | None | None | ✅ Match |
| T6 | T5 | T5 | ✅ Match |
| T7 | T6 | T6 | ✅ Match |
| T8 | None | None | ✅ Match |
| T9 | T8 | T8 | ✅ Match |
| T10 | T9 | T9 | ✅ Match |
| T11 | T2, T4, T7, T10 | T2, T4, T7, T10 | ✅ Match |
| T12 | T11 | T11 | ✅ Match |
| T13 | T12 | T12 | ✅ Match |
| T14 | T13 | T13 | ✅ Match |

## Test Co-location Validation

| Task | Code Layer | Matrix Requires | Task Says | Status |
| --- | --- | --- | --- | --- |
| T1 a T10 | telas de módulo e guarda da frente | component com MSW | component com MSW | ✅ OK |
| T11 | guarda global e remoção de alias | unit | unit + component | ✅ OK |
| T12 | documentação e revisão estrutural | revisão registrada | revisão documental e estrutural | ✅ OK |
| T13 | aparência e fluxo | aceite manual | aceite manual | ✅ OK |
| T14 | verificação final | verificador independente | verificador independente | ✅ OK |
