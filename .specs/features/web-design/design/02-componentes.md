# Componentes — Design

**Spec**: `.specs/features/web-design/spec/02-componentes.md`
**Status**: Aprovado pelo mantenedor em 2026-10-11
**Depende de**: `01-fundacao-visual.md` (tokens).

Cobre: componentes fundamentais, de domínio e seus estados. Os componentes compostos de página e de listagem estão em `03-padroes.md`; o shell, em `04-shell-e-acesso.md`.

---

## Architecture Overview

```
components/ui/        primitivos sobre @base-ui/react · sem regra de negócio · sem sessão
      ▲
components/app/       compostos e de domínio · podem ler sessão e formatadores
      ▲
features/<módulo>/    telas e mapeamentos do módulo (status → tom, colunas, filtros)
```

- `components/ui` não importa de `components/app`, de `features` nem de `lib/api`.
- `components/app` não importa de `features`. Hoje `status-badge.tsx` importa um tipo do financeiro; isso sai (DSC-05 AC4).
- Variantes são declaradas com `class-variance-authority`, como o kit já faz.
- Nenhuma dependência nova: Radio, Switch, Tabs e Tooltip existem em `@base-ui/react` (conferido em `web/node_modules/@base-ui/react/`), mas só são criados no primeiro uso (WDS-D-022).

---

## Regras comuns a todo componente interativo

| Regra | Como | Requisito |
| --- | --- | --- |
| Foco | `outline-none` sempre acompanhado de `focus-visible:ring-2 ring-ring ring-offset-2`; no shell, `ring-sidebar-ring` | DSC-01 AC2-3 |
| Alvo | Altura pelos tokens `--control-height-*`, que sobem a 44 px em toque. Controles menores que 24 px (caixa de seleção, fechar toast) ganham área de toque por pseudo-elemento | DSC-01 AC4-5 |
| Estados | Repouso, hover, pressionado, foco, desabilitado; campos também inválido e somente leitura | DSC-01 AC1 |
| Transição | Só `background-color`, `border-color`, `color`, `opacity`, com `--duration-fast` | DSF-04 AC5 |
| Ícones | 16 px, `aria-hidden` quando há texto | DSC-01 AC6 |
| Toque | `touch-action: manipulation` em botões e links de ação | `web-design-guidelines` |

---

## Fundamentais

### Button — ajustar (`components/ui/button.tsx`)

| Variante | Aparência | Uso |
| --- | --- | --- |
| `primary` (padrão) | Fundo `primary`, texto branco; hover `primary-hover` | Uma por região: a ação que conclui a tarefa |
| `secondary` | Contorno `input`, fundo `background`; hover `accent` | Ações alternativas, "Cancelar" |
| `ghost` | Sem fundo; hover `accent` | Ações de baixa ênfase, barras de ferramenta |
| `destructive` | Fundo `destructive`, texto branco; hover `destructive-emphasis` | Só o botão que confirma a ação destrutiva, dentro da confirmação |
| `link` | Texto sublinhado | Ação em linha de texto |

- Tamanhos: `sm` (32), `md` (36, padrão), `lg` (44). Os tamanhos `xs` e `icon-xs` do kit (24 px) saem (aprovado, WDS-D-027). Ícones podem ter 24 px; a área interativa segue DSC-01 AC4 e AC5.
- `loading`: propriedade que desabilita, põe `aria-busy`, mostra `LoaderCircle` e troca o rótulo por `loadingLabel` (obrigatório quando `loading` é usado, terminado em "…").
- O botão que abre uma ação destrutiva (por exemplo "Cancelar lançamento" no detalhe) é `secondary`; só o botão final da confirmação é `destructive`. Assim o vermelho não se espalha pela tela.
- Renomeação: a variante `outline` passa a `secondary`; a `secondary` atual (preenchimento cinza) sai. A variante `default` passa a `primary`.
- **Compatibilidade temporária**: `default` e `outline` continuam aceitas como alias de `primary` e `secondary`, e `icon-sm` como alias do IconButton compacto, até DS-INT removê-las. Em 2026-10-11 havia, fora do kit, 20 usos de `outline`, 17 de `destructive`, 6 de `ghost` e 5 de `size="sm"`.
- Link com aparência de botão: `buttonVariants` aplicado a `Link`, nunca `onClick` com `router.push` (DSC-02 AC5).

### IconButton — criar (`components/ui/icon-button.tsx`)

```ts
type IconButtonProps = Omit<ButtonProps, "children"> & {
  label: string;            // nome acessível, obrigatório
  children: ReactNode;      // o ícone
};
```

- Quadrado, nos mesmos três tamanhos. Variantes `ghost` (padrão) e `secondary`.
- `label` vira `aria-label`. Sem `label`, não compila (DSC-02 AC3).
- Tooltip não é usado: a informação está no nome acessível, e em toque não há hover.

### Input e Textarea — ajustar

- Altura `--control-height-md`; borda `input`; raio `--radius-md`; texto `--control-font-size`.
- Inválido: borda `destructive`. Somente leitura: fundo `muted`. Desabilitado: fundo `disabled`.
- Textarea: altura mínima de três linhas, redimensionável na vertical.
- Placeholder só como exemplo de formato, terminado em "…"; nunca no lugar do rótulo.

### NativeSelect — criar no kit (`components/ui/native-select.tsx`)

```ts
type NativeSelectProps = ComponentProps<"select"> & { invalid?: boolean };
```

- `<select>` com `appearance: none`, seta própria (`ChevronDown`) e as mesmas medidas e estados do Input.
- Cor de fundo e de texto explícitas (`background`, `foreground`), também nas opções.
- Substitui `features/financeiro/lancamentos/select-field.tsx` e o `<select>` de `features/identity/usuarios/users-page.tsx` (WDS-D-007).
- Usado sempre dentro de `Field`, que cuida de rótulo e erro.

### Select (rico) — manter (`components/ui/select.tsx`)

Sem uso nas telas da V1. Fica no kit para o primeiro caso com busca ou conteúdo rico. Ajuste só de tokens.

### Field — ajustar (`components/ui/field.tsx`)

- Composição: `Field` › `FieldLabel` › controle › `FieldDescription` › `FieldError`.
- O `Field` do Base UI liga o rótulo e o erro ao controle. Como o achado de `produtos-page.tsx:49` não foi conclusivo na leitura do código, a tarefa do kit inclui um teste que confirma a associação com `Input`, `Textarea` e `NativeSelect` (`getByLabelText`), e a correção se falhar.
- `FieldError` ganha ícone (`CircleAlert`) à frente do texto.
- `FieldLabel` ganha a propriedade `optional`, que acrescenta "(opcional)" em `muted-foreground`.

### Checkbox — ajustar; Radio e Switch — especificados, não criados

- Checkbox: caixa de 16 px, raio `--radius-sm`, marcado em `primary`. Controle e rótulo no mesmo `label`, formando um alvo único de 44 px de altura em toque.
- Radio: escolha única entre até cinco opções visíveis. Switch: configuração de efeito imediato. Ambos seguem o alvo único. Criados no primeiro uso real.

### Badge — ajustar (`components/ui/badge.tsx`)

| Tom | Fundo | Texto |
| --- | --- | --- |
| `neutral` | `info-subtle` | `info` |
| `success` | `success-subtle` | `success-emphasis` |
| `warning` | `warning-subtle` | `warning-emphasis` |
| `error` | `destructive-subtle` | `destructive-emphasis` |
| `outline` | transparente, borda `border` | `foreground` |

- Pílula, 22 px de altura, texto de 12 px peso 500, ícone de 12 px opcional.
- Sem `transition-all`. As variantes `default` (preto), `ghost` e `link` saem: badge não é ação.

### Alert — ajustar

Tons `neutral`, `success`, `warning` e `error`, cada um com ícone fixo, fundo suave e borda de 1 px no tom. `role="alert"` nos de erro e aviso; `role="status"` nos demais.

### Dialog e AlertDialog — ajustar

| Aspecto | Regra |
| --- | --- |
| Papel | Dialog: formulário e detalhe. AlertDialog: confirmar ação destrutiva ou sensível |
| Largura | 480 px (padrão) ou 640 px; abaixo de 640 px de tela, tela cheia |
| Estrutura | Cabeçalho (título `h2` em Geist 600), corpo rolável, rodapé de ações |
| Tela cheia | Rodapé fixo na base, com `padding-bottom: env(safe-area-inset-bottom)`; botão fechar de 44 px no cabeçalho |
| Rolagem | `overscroll-behavior: contain` no corpo |
| Foco | Entra no primeiro campo (formulário) ou no botão cancelar (confirmação); volta ao gatilho ao fechar. É o comportamento do Base UI; confirma-se por teste |
| Fechar | Esc e botão; clique fora fecha o Dialog, não o AlertDialog. Com ação em andamento, nada fecha |
| Aninhamento | Proibido. Um formulário em diálogo que precise confirmar usa a própria área do diálogo |
| Sombra | Nível 2, sobre `overlay` |

### Sheet (gaveta) — ajustar

Usada pela navegação abaixo de 1024 px e pelos filtros abaixo de 768 px. Lados `left` (navegação) e `bottom` (filtros). Mesmas regras de foco, Esc e `overscroll-behavior` do Dialog.

### DropdownMenu — ajustar

- Item com 36 px de altura (44 px em toque), ícone opcional de 16 px.
- Item destrutivo em `destructive`, sempre o último, separado por divisória.
- Item que navega é renderizado como link.

### Toast — ajustar (`components/ui/toast.tsx`)

| Aspecto | Regra |
| --- | --- |
| Tons | `neutral`, `success`, `warning`, `error`, cada um com ícone e filete lateral de 3 px no tom |
| API | `toast.show`, `toast.success`, `toast.error` mantidos; entra `toast.warning` |
| Duração | Sucesso e neutro 5 s; aviso e erro 8 s. Todos dispensáveis; a contagem pausa em hover e foco (comportamento do Base UI, confirmado por teste) |
| Posição | Canto inferior direito; abaixo de 768 px, base da tela em largura total, acima da área segura |
| Anúncio | Região viva do Base UI; erro como `assertive` |
| Uso | Resultado de ação. Erro que exige correção fica na tela (campo, formulário ou confirmação), não só em toast |

### Table — ajustar (`components/ui/table.tsx`)

- Cabeçalho: fundo `muted`, texto de 12 px peso 600 em `muted-foreground`, régua inferior de 2 px `border-strong`, altura `--table-head-height`.
- Linha: altura mínima `--table-row-height`, divisória de 1 px `border`, hover `accent`.
- Célula: padding de 12 por 8. A quebra de linha deixa de ser proibida por padrão; cada coluna declara `nowrap`, `truncate` ou quebra (caso de borda da spec).
- `TableHead` recebe `scope="col"`; a tabela exige `aria-label` ou `aria-labelledby`.
- Coluna numérica: propriedade `numeric` no cabeçalho e na célula, que aplica alinhamento à direita e `tabular-nums`.
- O contêiner com `overflow-x-auto` sai do componente base. A rolagem horizontal vira uma variante explícita (`scrollable`), usada só com justificativa (DSP-05 AC9), com primeira coluna fixa, `tabindex="0"` e nome acessível.
- **Compatibilidade temporária**: até DS-INT, a `Table` usada diretamente pelas telas ainda não migradas mantém a rolagem embutida como padrão (`scroll="legacy"`); a `DataList` já usa a tabela sem rolagem. DS-INT remove o modo legado.

### Skeleton — ajustar

Fundo `muted`, pulso suave, sem animação sob redução de movimento. Formas prontas: linha de texto, linha de tabela, bloco de valor.

### Card — criar (`components/ui/card.tsx`)

Borda de 1 px `border`, raio `--radius-lg`, sem sombra, padding de 16 ou 24. Partes: `Card`, `CardHeader`, `CardTitle` (`h3`), `CardContent`. Usos previstos na V1: atalhos do início, item de lista empilhada, bloco de valor do saldo. Não envolve tabela, lista nem formulário de página (DSC-05 AC5).

### Breadcrumb — criar (`components/ui/breadcrumb.tsx`)

`nav` com `aria-label="Trilha de navegação"` e lista ordenada. Itens são links; o último é texto com `aria-current="page"`. Separador `ChevronRight` decorativo. Abaixo de 768 px mostra só o item anterior, como "voltar" (`ArrowLeft` e o nome da listagem).

### ShowMore — criar (`components/ui/show-more.tsx`)

```ts
type ShowMoreProps = {
  shown: number; total?: number;   // total ausente quando a API pagina por cursor
  onShowMore: () => void; loading?: boolean;
  noun: [singular: string, plural: string];
};
```

Texto "Mostrando 50 de 137 lançamentos" e botão `secondary` "Mostrar mais". Com cursor e total desconhecido: "Mostrando 50 usuários". A contagem fica em região `aria-live="polite"`.

### Tabs e Tooltip — especificados, não criados

- Tabs: setas alternam as abas, Tab entra no painel; aba ativa com peso 600 e indicador `brand` de 2 px; aba ativa refletida na URL (`03-padroes.md`).
- Tooltip: só para reforço; nunca carrega informação necessária; não é usado em toque.

### Tipografia — criar (`components/ui/typography.tsx`)

`DisplayTitle`, `PageTitle`, `SectionTitle`, `Eyebrow`: os únicos lugares, com o shell, onde `font-display` e caixa alta aparecem (`01-fundacao-visual.md`).

---

## De domínio (`components/app/`)

### StatusBadge — refatorar (`components/app/status-badge.tsx`)

```ts
type Tone = "neutral" | "success" | "warning" | "error";
type StatusDef = { label: string; tone: Tone; icon: LucideIcon };
function StatusBadge(props: { status: StatusDef; className?: string }): JSX.Element;
```

- O componente deixa de conhecer `StatusLancamento`. Cada feature exporta o próprio mapa (`features/financeiro/lancamentos/labels.ts`, `features/identity/usuarios/…`, `features/estoque/…`).
- `activeStatus(boolean)` continua em `components/app`, porque "ativo e inativo" é comum aos três módulos.

**Mapa proposto (DD-3)**

| Status | Rótulo | Tom | Ícone |
| --- | --- | --- | --- |
| `CRIADA` | Criada | neutral | `CircleDashed` |
| `RECEBIDA` | Recebida | success | `ArrowDownToLine` |
| `PAGA` | Paga | success | `ArrowUpFromLine` |
| `CANCELADA` | Cancelada | neutral, texto riscado no valor | `Ban` |
| ativo | Ativo | success (contorno) | `CircleCheck` |
| inativo | Inativo | neutral | `CircleSlash` |
| Saldo negativo | Saldo negativo | error | `TriangleAlert` |

`RECEBIDA` e `PAGA` são ambas conclusões e ficam no mesmo tom, distintas por rótulo e por ícone de direção (entrada e saída). **DD-3 (aprovada, WDS-D-026)**: `CANCELADA` passa de vermelha a neutra, com ícone e texto, porque cancelamento não é erro; o vermelho de erro fica reservado a falhas e a saldo negativo.

### MoneyValue — criar (`components/app/money-value.tsx`)

```ts
type MoneyValueProps = {
  cents: number;
  variant?: "neutral" | "income" | "expense" | "balance";
  size?: "body" | "display";
  className?: string;
};
```

| Variante | Sinal | Tom | Extra |
| --- | --- | --- | --- |
| `neutral` | o do próprio valor | `foreground` | — |
| `income` | "+" | `success` | — |
| `expense` | "−" (U+2212) | `foreground` | — |
| `balance` | "−" quando negativo | `foreground`; negativo em `destructive` | Negativo acrescenta `TriangleAlert` e o texto oculto "saldo negativo" |

- Formatação por `formatBRL` (`web/lib/money.ts`), `tabular-nums`, sem quebra entre "R$" e o número.
- `size="display"`: 36 px, fonte do token `--font-money-display` (Barlow Condensed por DD-4, trocável por Geist num só lugar).
- Em tabelas, valores vão em coluna `numeric` com `MoneyValue` `neutral`; o tipo (receita ou despesa) já é outra coluna.

### MoneyField e QuantityField — ajustar só a aparência

Comportamento, conversão e testes inalterados (FND-05 AC4-5). Passam a usar `Field` e os tokens de controle.

### RoleBadge — criar (`components/app/role-badge.tsx`)

`Badge` `outline` com o nome do papel em português, a partir do mapa de rótulos que já existe em `features/identity/usuarios/`. Como o mapa é da feature, o componente recebe o rótulo pronto.

### ConfirmDialog — ajustar (`components/app/confirm-dialog.tsx`)

- Comportamento mantido (FND-05 AC6).
- `confirmLabel` passa a ser obrigatório, para não existir "Confirmar" genérico (DSP-03 AC3). As seis chamadas existentes já informam o rótulo.
- Entram `subject` (nome do registro, mostrado em destaque) e `consequence` (frase sobre o efeito), que compõem a descrição (DSP-03 AC2).
- Entra `error`, mostrado como `Alert` dentro do diálogo quando a ação falha (DSP-03 AC5).

---

## Inventário final

| Componente | Arquivo | Ação | Usado na V1 |
| --- | --- | --- | --- |
| Button | `ui/button.tsx` | Ajustar variantes, tamanhos, `loading` | Sim |
| IconButton | `ui/icon-button.tsx` | Criar | Sim |
| Input, Textarea, Label, Separator | `ui/*` | Ajustar tokens | Sim |
| NativeSelect | `ui/native-select.tsx` | Criar | Sim |
| Select | `ui/select.tsx` | Ajustar tokens | Não |
| Field | `ui/field.tsx` | Ajustar; `optional`; ícone de erro | Sim |
| Checkbox | `ui/checkbox.tsx` | Ajustar | Sim |
| Badge, Alert, Skeleton | `ui/*` | Ajustar tons | Sim |
| Dialog, AlertDialog, Sheet, DropdownMenu | `ui/*` | Ajustar; tela cheia; rolagem contida | Sim |
| Toast | `ui/toast.tsx` | Ajustar tons, ícone, `warning` | Sim |
| Table | `ui/table.tsx` | Ajustar; `numeric`; sem rolagem embutida | Sim |
| Card, Breadcrumb, ShowMore, tipografia | `ui/*` | Criar | Sim |
| Radio, Switch, Tabs, Tooltip | — | Especificados; criar no primeiro uso | Não |
| StatusBadge, ConfirmDialog, MoneyField, QuantityField | `app/*` | Ajustar | Sim |
| MoneyValue, RoleBadge, BrandMark | `app/*` | Criar | Sim |

---

## Testes

- Cada arquivo de `components/ui` e `components/app` criado ou alterado tem teste de papel, nome acessível e teclado (DSC-07 AC1).
- Os testes atuais (`display.test.tsx`, `field.test.tsx`, `overlays.test.tsx` e os de `components/app`) continuam valendo. Mudam só onde a API muda por decisão desta feature:

| Teste | Mudança | Motivo |
| --- | --- | --- |
| `components/app/status-badge.test.tsx` | Passa a receber a definição de status | DSC-05 AC4 |
| `components/app/confirm-dialog.test.tsx` | `confirmLabel` obrigatório | DSP-03 AC3 |
| `components/ui/display.test.tsx` | Nomes de variante de Button e Badge | Este documento |

## Risks & Concerns

| Concern | Location | Impact | Mitigation |
| --- | --- | --- | --- |
| Renomear variantes do Button quebra chamadas | `variant="outline"` e `variant="destructive"` em features | Erro de tipo em várias telas | O TypeScript aponta todas; a troca é feita na unidade do kit junto com as chamadas, ou a variante antiga fica como alias até a migração da tela (decidir em Tasks) |
| `StatusBadge` acoplado ao financeiro | `components/app/status-badge.tsx:2` | Componente comum dependente de módulo | Inversão descrita acima |
| Associação de rótulo não confirmada | `features/estoque/produtos/produtos-page.tsx:49` | Campo sem nome acessível | Teste de `getByLabelText` no kit |
| Botão de 24 px do kit | `components/ui/button.tsx:24` | Alvo abaixo do mínimo em toque | Tamanhos `xs` removidos |
| Tela cheia do diálogo com teclado virtual | Dialog abaixo de 640 px | Rodapé fixo pode cobrir o campo em foco | Rodapé fixo ao contêiner do diálogo, não à janela; conferido no smoke em 390 px |
| Classes `dark:` e valores arbitrários no kit | `components/ui/*` | Guarda de tokens falha | Reescrita na unidade do kit |

## Tech Decisions

| Decisão | Escolha | Motivo |
| --- | --- | --- |
| Nome acessível do IconButton | Propriedade obrigatória no tipo | Erro em tempo de compilação, não em revisão |
| Mapa de status | Na feature dona | Mantém `components/app` sem dependência de módulo |
| Rolagem horizontal da tabela | Variante explícita | O padrão deixa de ser rolar (WDS-D-006) |
| Destrutivo | Sólido só no botão final da confirmação | Mantém o vermelho raro na tela |
