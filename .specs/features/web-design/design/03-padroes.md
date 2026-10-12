# Padrões Compartilhados — Design

**Spec**: `.specs/features/web-design/spec/03-padroes.md`
**Status**: Aprovado pelo mantenedor em 2026-10-11
**Depende de**: `01-fundacao-visual.md`, `02-componentes.md`.

Cobre: padrões compostos, formulários, tabelas e listagens, estado na URL, responsividade por padrão e acessibilidade.

---

## Architecture Overview

```
app/(app)/<rota>/page.tsx        Server Component: metadados, RequirePermission, Suspense
  └─ features/<módulo>/<tela>    Client Component
       ├─ PageContainer › PageHeader (Breadcrumb · PageTitle · ações)
       ├─ FilterBar ── useQueryState ──► URL (query string)
       ├─ DataList (tabela ou lista) › RowActions
       ├─ ShowMore
       └─ estados: LoadingState · EmptyState · ErrorState · NoAccess
```

Componentes novos em `web/components/app/`; utilitários novos em `web/lib/url-state/` e `web/lib/responsive/`.

---

## Decisão de design (aprovada em 2026-10-11, WDS-D-026)

### DD-2 — Como a tabela vira lista abaixo de 768 px

**Aprovada a opção A, com três condições do mantenedor:** semântica e acessibilidade adequadas em cada layout; CSS sempre que viável (a priorização de colunas e todo o restante do responsivo são só CSS; a consulta de mídia em JavaScript fica restrita à troca entre tabela e lista); e compatibilidade com SSR e hidratação do Next. Para a última, o hook de consulta de mídia usa `useSyncExternalStore` com retrato de servidor fixo (desktop), de modo que servidor e primeira renderização do cliente coincidem, e tem teste de renderização no servidor. A validação é feita em 390, 768, 820, 1024 e 1440 px.

| | A. Um componente, uma árvore, escolhida por consulta de mídia (recomendada) | B. Só CSS sobre a mesma `<table>` | C. Duas árvores, uma escondida por CSS |
| --- | --- | --- | --- |
| Como | `DataList` renderiza `<table>` ou `<ul>` conforme `matchMedia("(min-width: 768px)")` | `display: block` em linhas e células, rótulo por `data-label` | Tabela e lista renderizadas sempre; `hidden md:block` |
| Semântica | Cada layout com a sua: tabela é tabela, lista é lista | Mudar `display` de elementos de tabela faz alguns navegadores perderem a semântica de tabela | Correta, mas conteúdo duplicado |
| Testes atuais | Continuam: sem `matchMedia`, o padrão é a tabela | Continuam | Quebram: jsdom vê os dois, com nomes e diálogos duplicados |
| Custo | Um hook de consulta de mídia; uma definição de colunas serve aos dois layouts | Layout da lista limitado ao que o CSS consegue | Diálogos e ações em dobro |

Recomendação: A. Não há risco de divergência entre servidor e cliente, porque as listagens só são renderizadas no cliente, depois de `GET /auth/me` (`web/app/(app)/layout.tsx`). A priorização de colunas entre 768 e 1023 px continua sendo só CSS, na mesma tabela.

---

## Estrutura de página

### PageContainer (`components/app/page-container.tsx`)

```ts
type PageContainerProps = { width?: "narrow" | "default" | "wide"; children: ReactNode };
```

- Renderiza o `<main id="conteudo" tabIndex={-1}>` da página, com largura máxima pelo token e margem lateral `--page-gutter`. O `id` é o alvo do link de pular.
- Substitui as seis larguras avulsas atuais (`max-w-3xl` a `max-w-6xl`).

| Largura | Telas |
| --- | --- |
| `narrow` (640) | Saldo, troca de senha, sem acesso |
| `default` (1120) | Início, plano de contas, produtos, detalhe do produto, detalhe do lançamento |
| `wide` (1440) | Lançamentos, usuários |

### PageHeader (`components/app/page-header.tsx`)

```ts
type PageHeaderProps = {
  title: string;
  description?: ReactNode;
  breadcrumb?: { label: string; href?: string }[];
  actions?: ReactNode;      // no máximo um Button primary
  meta?: ReactNode;         // por exemplo o StatusBadge do registro
};
```

- `PageTitle` (`h1`) com o filete vermelho de 32 por 2 px abaixo. É o único `h1` da página.
- A partir de 768 px: título à esquerda, ações à direita, na mesma linha.
- Abaixo de 768 px: trilha reduzida a "voltar", título, depois as ações em largura total, com a primária por último.
- Substitui os 16 `h1` montados à mão.

### Seções

`SectionTitle` (`h2`) com régua de 2 px `border-strong` abaixo; conteúdo a 16 px; seções separadas por 32 px. Blocos dentro de seção usam `h3`.

---

## Estados de tela (`components/app/states.tsx`)

| Estado | Componente | Conteúdo | Acessibilidade |
| --- | --- | --- | --- |
| Carregando | `LoadingState` com `shape`: `table`, `list`, `detail`, `value` | Esqueleto no formato do conteúdo | `role="status"`, texto oculto "Carregando…" |
| Vazio | `EmptyState` `kind="empty"` | Ícone `Inbox`, título, frase de orientação, ação de criar para quem pode | Título como `h2` |
| Sem resultado | `EmptyState` `kind="no-results"` | Ícone `SearchX`, "Nenhum resultado para os filtros", ação "Limpar filtros" | idem |
| Erro | `ErrorState` | `Alert` de erro com a mensagem de FND-06 e "Tentar de novo" | `role="alert"` |
| Sem acesso | `NoAccess` | Ícone `Lock`, título, frase, "Voltar ao início" | Um `h1` |

- Os estados ocupam o lugar da lista; cabeçalho de página e barra de filtros permanecem (DSP-02 AC6).
- `EmptyState` já aceita descrição e ação; ganha `kind` e ícone. Toda listagem passa a informar a orientação e a ação, que hoje faltam (D-06).
- Ícones de 24 px em `muted-foreground`, sem círculo de fundo e sem ilustração.

---

## Feedback e confirmação

| Situação | Padrão |
| --- | --- |
| Escrita com sucesso | `toast.success` com frase curta no passado ("Lançamento criado."); a lista ou o detalhe refletem a mudança pelo cache de consultas |
| Ação destrutiva ou sensível | `ConfirmDialog` com `subject`, `consequence` e `confirmLabel` com o verbo da ação |
| Motivo obrigatório | Campo no diálogo; confirmar desabilitado enquanto vazio (FND-05 AC6, mantido) |
| Falha na ação confirmada | `Alert` de erro dentro do diálogo, que continua aberto |
| Erro de validação | No campo; ver formulários |
| Erro sem campo | `Alert` no topo do formulário |
| Erro de permissão vindo da API | Mensagem de FND-04 AC5; estado da tela inalterado |
| Ação indisponível pelo estado do registro | Controle desabilitado com o motivo em texto ao lado (`aria-describedby`), nunca só em tooltip |

Ações destrutivas da V1 que usam o padrão: desativar conta, cancelar lançamento, desativar usuário, retirar acesso administrativo, ajuste de estoque.

---

## Formulários

### FormLayout (`components/app/form-layout.tsx`)

```ts
type FormLayoutProps = ComponentProps<"form"> & { error?: ReactNode; footer: ReactNode };
```

- Coluna única, campos a 16 px, grupos a 24 px, largura máxima `narrow`.
- `error` aparece como `Alert` no topo e recebe foco quando surge.
- `footer`: ações à direita a partir de 768 px (secundária, depois primária); abaixo disso, em largura total, com a primária por último. Dentro de diálogo em tela cheia, o rodapé fica fixo na base.

### Regras

| Tema | Regra | Como |
| --- | --- | --- |
| Rótulo | Acima do campo, sempre visível | `Field` |
| Opcional | "(opcional)" no rótulo; obrigatório sem marca | `FieldLabel optional` |
| Validação | Ao sair do campo e ao enviar | React Hook Form em `mode: "onTouched"`, já dependência (WEB-D-003) |
| Foco no erro | Primeiro campo inválido ao enviar | Comportamento padrão do React Hook Form para campos registrados; confirmado por teste |
| Erros da API | `applyProblemToForm` de `lib/forms/`, sem mudança | FND-06 |
| Envio | Botão `loading` com rótulo "Salvando…" | `Button` |
| Tipo e teclado | `type` e `inputMode` corretos; `autocomplete` declarado | Já praticado; revisado tela a tela |
| E-mail e código | `spellCheck={false}`, `autoCapitalize="none"` | Achado da `web-design-guidelines` |
| Colar | Nunca bloqueado | — |
| Texto em toque | 16 px por `--control-font-size` | Evita zoom automático no iOS |

Formulários da V1 são curtos e vivem em diálogo. Em tela cheia (abaixo de 640 px) o padrão é o mesmo: cabeçalho, campos em coluna, rodapé fixo.

---

## Listagens

### Definição de colunas

```ts
type Column<T> = {
  id: string;
  header: string;
  cell: (row: T) => ReactNode;
  numeric?: boolean;
  priority: 1 | 2 | 3;          // 1 sempre visível; 2 a partir de 768; 3 a partir de 1024
  wrap?: "nowrap" | "truncate" | "wrap";
  inList?: "title" | "field" | "meta" | false;   // papel no item empilhado
};
```

### DataList (`components/app/data-list.tsx`)

```ts
type DataListProps<T> = {
  label: string;                               // nome acessível da tabela ou da lista
  columns: Column<T>[];
  rows: T[];
  rowKey: (row: T) => string;
  rowLabel: (row: T) => string;                // identifica a linha; usado nos nomes das ações
  primaryAction?: (row: T) => ReactNode;
  moreActions?: (row: T) => RowAction[];
  layout?: "auto" | "table" | "list";          // padrão "auto"
};
```

| Largura | Resultado |
| --- | --- |
| A partir de 1024 | Tabela com todas as colunas |
| 768 a 1023 | Mesma tabela; colunas de prioridade 3 ocultas por CSS (`hidden lg:table-cell`); o dado continua no detalhe |
| Abaixo de 768 | Lista (`<ul>`) de itens empilhados |
| `layout="table"` | Tabela em qualquer largura: para tabelas de até quatro colunas curtas (DSP-05 AC8) |

**Item empilhado**

```
┌──────────────────────────────────────────────┐
│ Título (coluna "title")              [status] │
│ Campo-chave 1 · Campo-chave 2                 │
│ Valor (numérico, à direita)          [ ⋮ ]    │
└──────────────────────────────────────────────┘
```

- Título em Geist 600; até três campos (`inList: "field"`) em `muted-foreground`; `meta` para status; texto longo limitado a três linhas.
- Itens separados por divisória de 1 px, sem cartão em volta: a lista ocupa a largura da página.
- Quando o item leva ao detalhe, o título é o link; a área de toque do link cobre a linha do título.

### RowActions (`components/app/row-actions.tsx`)

```ts
type RowAction = { label: string; icon?: LucideIcon; onSelect?: () => void; href?: string; destructive?: boolean };
```

- Uma ação visível (`primaryAction`) e as demais em `DropdownMenu` aberto por `IconButton`.
- Nome acessível do menu: "Ações de {rowLabel}". Ações visíveis repetidas ganham o complemento: o texto visível continua "Devolver", e o nome acessível passa a "Devolver {rowLabel}" (DSP-05 AC10; o texto visível faz parte do nome, como pede a WCAG 2.5.3).
- Exemplos de `rowLabel`: "lançamento Receita fb6e3d8f", "saída de 3 unidades em 09/10/2026", o nome do usuário.

### FilterBar (`components/app/filter-bar.tsx`)

```ts
type FilterBarProps = {
  children: ReactNode;          // Fields com NativeSelect, Input de busca ou de data
  activeCount: number;
  resultCount: number; resultNoun: [string, string];
  onClear: () => void;
};
```

| Largura | Comportamento |
| --- | --- |
| A partir de 1024 | Uma linha sobre fundo `muted`: filtros, depois "Limpar filtros" (visível só com filtro ativo) |
| 768 a 1023 | Mesma barra, em duas linhas (grade de três colunas) |
| Abaixo de 768 | A busca, se houver, fica visível; os demais filtros vão para uma gaveta inferior aberta pelo botão "Filtros (2)", com "Aplicar" e "Limpar" no rodapé |

A contagem ("12 lançamentos") fica abaixo da barra, em região `aria-live="polite"`, para anunciar o efeito do filtro.

### Estado na URL (`lib/url-state/`)

```ts
type Codec<T> = { parse: (raw: string | null) => T; serialize: (value: T) => string | null };

function useQueryState<T>(
  key: string,
  codec: Codec<T>,
  options?: { history?: "replace" | "push" },
): [value: T, setValue: (next: T) => void];

const codecs: {
  string: (fallback?: string) => Codec<string>;
  oneOf: <T extends string>(values: readonly T[], fallback: T | "") => Codec<T | "">;
  isoDate: () => Codec<string>;
};
```

- Leitura por `useSearchParams`. Escrita por `window.history.replaceState` ou `pushState`, que o Next integra ao roteador e sincroniza com `useSearchParams` (documentação em `web/node_modules/next/dist/docs/01-app/01-getting-started/04-linking-and-navigating.md`).
- `serialize` devolve `null` para o valor padrão, o que remove o parâmetro: a URL sem filtro fica limpa.
- `parse` devolve o valor padrão para entrada inválida (DSP-05 AC4). Filtro que aponta para registro inexistente é descartado pela tela.
- Filtros e busca usam `replace`; aba usa `push` (caso de borda da spec).
- Busca: o campo responde a cada tecla; a escrita na URL espera 300 ms. O Safari recusa mais de 100 chamadas de histórico em 30 s.
- "Mostrar mais" não vai para a URL.
- Como `useSearchParams` exige um limite de `Suspense` em página pré-renderizada, cada `page.tsx` de listagem envolve a feature em `<Suspense>` com `LoadingState` (o login já faz isso).

| Tela | Parâmetros |
| --- | --- |
| Lançamentos | `tipo`, `status`, `conta`, `de`, `ate` |
| Usuários | `situacao`, `papel` |
| Produtos | `busca` |

### Listas longas

- Listas carregadas inteiras (lançamentos, contas, produtos, movimentações): estado local `visíveis = 50`, mais 50 por clique em `ShowMore`. Filtro e busca operam sobre todos os registros; mudar o filtro volta a 50.
- Usuários: `ShowMore` chama a página seguinte pelo `next_cursor`, como a tela já faz.
- Sem virtualização (WDS-D-012). Se os volumes crescerem, a paginação no servidor é reavaliada.

---

## Responsividade por padrão

Contextos: **XS** abaixo de 640, **SM** 640 a 767, **MD** 768 a 1023, **LG** a partir de 1024. "Toque" é `pointer: coarse`, em qualquer largura.

| Padrão | XS | SM | MD | LG | Toque |
| --- | --- | --- | --- | --- | --- |
| Navegação | Gaveta | Gaveta | Gaveta | Lateral fixa de 240 | Itens de 44 |
| Margem de página | 16 | 16 | 24 | 32 | — |
| Cabeçalho de página | Empilhado; ações em largura total | Empilhado | Linha única | Linha única | Botões de 44 |
| Título de página | 24 / 28 | 24 / 28 | 30 / 36 | 30 / 36 | — |
| Trilha de navegação | Só "voltar" | Só "voltar" | Completa | Completa | Alvo de 44 |
| Barra de filtros | Busca visível; resto em gaveta | idem | Duas linhas | Uma linha | Campos de 44, texto de 16 |
| Listagem | Lista empilhada | Lista empilhada | Tabela, prioridades 1 e 2 | Tabela completa | Linha de 48 |
| Tabela pequena | Tabela | Tabela | Tabela | Tabela | — |
| Ações de linha | Uma visível e menu | idem | idem | idem | Alvo de 44 |
| Formulário | Coluna única, ações em largura total | Coluna única | Coluna única, ações à direita | idem | Campos de 44, texto de 16 |
| Diálogo | Tela cheia, rodapé fixo | Centralizado, 480 | Centralizado | Centralizado | Fechar de 44 |
| Confirmação | Centralizada, margem de 16, ações empilhadas | Centralizada | Centralizada | Centralizada | Botões de 44 |
| Toast | Base, largura total | Base, largura total | Canto inferior direito | idem | Fechar de 44 |
| Cartões e atalhos | 1 coluna | 2 colunas | 2 colunas | 3 colunas | — |
| Seções | 24 entre seções | 24 | 32 | 32 | — |
| Estados de tela | Padding de 32 | 32 | 48 | 48 | — |

Piso: tudo funciona a 320 px de largura e com zoom de 200%, sem rolagem horizontal de página.

---

## Acessibilidade

### Achados da `web-design-guidelines` e onde são resolvidos

| Achado (diagnóstico) | Solução no Design System | Verificado por |
| --- | --- | --- |
| Fonte de corpo caindo em Times (`globals.css:10`) | `--font-sans` corrigida | Conferência no navegador |
| `transition-all` em Button e Badge | Propriedades listadas; guarda recusa `transition-all` | Guarda de tokens |
| Sem `prefers-reduced-motion` | Regra global na camada base | Teste de CSS e smoke com a preferência ativa |
| Sem link de pular para o conteúdo | Primeiro elemento do shell; alvo `main#conteudo` | Teste do shell e percurso por teclado |
| Navegação por `onClick` com `router.push` | Itens de menu que navegam são links | Teste do menu da pessoa |
| Nomes de ação repetidos entre linhas ("Detalhes", "Devolver") | `rowLabel` em `DataList` e `RowActions` | Teste por nome acessível único |
| Colunas de dinheiro sem `tabular-nums` | Coluna `numeric` e `MoneyValue` | Teste de classe; espécime |
| Texto longo sem tratamento em tabela | `wrap` por coluna; três linhas no item empilhado | Smoke em 320 px |
| Filtros só em memória | `useQueryState` | Teste de URL por tela |
| E-mail sem `spellCheck={false}` | Regra de formulários | Revisão da skill |
| Diálogo e gaveta sem `overscroll-behavior` | Regra do Dialog e do Sheet | Revisão da skill |
| Sem `touch-action: manipulation` | Regra comum dos componentes | Revisão da skill |
| Sem `safe-area-inset` | Shell, rodapé fixo de diálogo e toasts | Smoke em 390 px |
| Sem `theme-color` e `color-scheme` | `viewport` em `app/layout.tsx`; `color-scheme: light` | Conferência no navegador |
| `<select>` nativo sem cores explícitas | `NativeSelect` | Teste do kit |
| Listas inteiras sem limite | `ShowMore` de 50 | Teste da listagem |
| Títulos sem equilíbrio de linha | `text-wrap: balance` nos componentes de título | Guarda de tokens |
| Estados vazios sem orientação | `EmptyState` com `kind`, frase e ação | Teste por tela |
| Associação de rótulo não confirmada na busca de produtos | `Field` com teste de `getByLabelText` | Teste do kit |

### Requisitos permanentes

| Tema | Regra | Onde |
| --- | --- | --- |
| Contraste | 4,5:1 em texto; 3:1 em gráfico e borda de campo | Teste de contraste |
| Foco visível | Anel de 2 px em todo controle, no conteúdo e no shell | Regras comuns dos componentes |
| Teclado | Tudo alcançável; Esc fecha; foco volta ao gatilho; ordem de foco igual à visual | Kit e smoke por teclado |
| Alvos | 44 em toque; 24 no mínimo | Tokens de componente |
| Rótulos | Todo campo com rótulo visível associado | `Field` |
| Erros | No campo, com ícone, texto e `aria-describedby`; foco no primeiro erro | `Field`, `FormLayout` |
| Estados vazios | Título, orientação e ação | `EmptyState` |
| Diálogo e gaveta | Foco preso, Esc, retorno do foco, rolagem contida, sem aninhamento | Kit |
| Cor | Nunca o único sinal: status, valores, erro e item ativo têm ícone, texto ou forma | `StatusBadge`, `MoneyValue`, shell |
| Hierarquia | Um `h1`; níveis sem salto; regiões nomeadas | `PageHeader`, shell |
| Região viva | Toasts, contagem de resultados, carregamento | Kit |
| Movimento | Redução respeitada | Camada base |
| Zoom e refluxo | 200% e 320 px sem rolagem horizontal de página | Smoke |
| Idioma | `lang="pt-BR"` | `app/layout.tsx` |

---

## Texto da interface

| Tema | Regra | Exemplo |
| --- | --- | --- |
| Caixa | Frase comum | "Novo lançamento" |
| Ações | Verbo e objeto | "Desativar conta", "Anexar comprovante" |
| Confirmação | Repete o verbo | "Cancelar lançamento" |
| Progresso | Termina em "…" | "Salvando…" |
| Erro do front | O que houve e o próximo passo | "O arquivo passa de 9 MB. Escolha um arquivo menor." |
| Unidades | Espaço não separável | "9 MB" |
| Limite de anexo | Megabytes inteiros, para baixo | "até 9 MB" (o limite real do front, 10.420.224 bytes, não muda) |
| Identificador | Tipo e código curto em fonte monoespaçada | "Receita `fb6e3d8f`" |
| Vazio | Orientação, não constatação | "Nenhum lançamento ainda. Crie o primeiro para começar o controle." |

---

## Code Reuse Analysis

| Existente | Local | Uso |
| --- | --- | --- |
| `LoadingState`, `EmptyState`, `ErrorState` | `components/app/states.tsx` | Estendidos, sem troca de nome |
| `RequirePermission`, `NoAccess` | `components/app/require-permission.tsx` | Lógica mantida; `NoAccess` ganha o layout padrão |
| `ConfirmDialog` | `components/app/confirm-dialog.tsx` | Estendido |
| `applyProblemToForm`, catálogos de erro | `lib/forms/`, `features/*/errors.ts` | Sem mudança |
| Filtros puros (`filterLancamentos`, `searchProdutos`) | `features/*/filters.ts`, `search.ts` | Sem mudança; só a origem do estado muda |
| `navigation-mock` | `lib/session/navigation-mock.ts` | Estendido para reagir a `replaceState` e `pushState` |
| `useSearchParams` com `Suspense` | `app/(public)/entrar/page.tsx` | Mesmo arranjo nas listagens |

## Risks & Concerns

| Concern | Location | Impact | Mitigation |
| --- | --- | --- | --- |
| Mock de navegação não reage a mudanças de histórico | `web/lib/session/navigation-mock.ts:23` | Testes de filtro na URL não veriam a atualização | A unidade de `lib` torna o mock reativo, com teste próprio |
| `matchMedia` ausente no jsdom | `web/test/setup.ts` | O hook de mídia falharia | O hook trata a ausência como desktop; testes de lista usam um mock explícito |
| Layout escolhido por JavaScript | `DataList` | Um quadro com o layout errado ao abrir | As listagens só existem depois da sessão carregar no cliente; o hook lê `matchMedia` de forma síncrona na primeira renderização |
| `Suspense` esquecido numa página | `app/(app)/**/page.tsx` | `next build` falha | O build é gate de toda tarefa |
| Limite de chamadas de histórico no Safari | `useQueryState` | Erro ao digitar rápido na busca | Escrita na URL com espera de 300 ms |
| Lista filtrada e "Mostrar mais" | listagens | Pessoa não vê que há mais resultados | `ShowMore` sempre mostra "X de Y" |
| Testes que buscam ações por nome exato | `movimentacoes-panel.test.tsx`, `lancamentos-page.test.tsx` | Falham com o nome que identifica a linha | Alterados com citação de DSP-05 AC10 (lista em `05-migracao-web-v1.md`) |

## Tech Decisions

| Decisão | Escolha | Motivo |
| --- | --- | --- |
| Tabela para lista | Uma árvore por consulta de mídia (DD-2) | Semântica correta e testes atuais preservados |
| Prioridade de colunas | CSS na mesma tabela | Sem JavaScript e sem duplicação |
| Estado na URL | `useSearchParams` e API de histórico nativa | WDS-D-011; sem biblioteca |
| Contagem de resultados | Região viva abaixo dos filtros | Leitor de tela percebe o efeito do filtro |
| Validação | `onTouched` | Erro aparece ao sair do campo, sem acusar enquanto a pessoa digita |
