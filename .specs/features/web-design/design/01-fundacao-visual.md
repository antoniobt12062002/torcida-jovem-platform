# Fundação Visual — Design

**Spec**: `.specs/features/web-design/spec/01-fundacao-visual.md`
**Status**: Aprovado pelo mantenedor em 2026-10-11
**Decisões**: WDS-D-001 a WDS-D-025 (`../STATE.md`), AD-018, AD-002, WEB-D-003.

Cobre: arquitetura de tokens, paleta e tokens semânticos, tipografia e escala, espaçamento, raio, bordas, sombras, estados, iconografia, movimento, marca e espécime.

---

## Architecture Overview

```
app/layout.tsx ── next/font ──► --font-geist-sans · --font-geist-mono · --font-barlow-condensed
        │
app/globals.css
  ├─ 1. Primitivos        --tj-ink · --tj-red · --tj-steel · --tj-neutral-* · --tj-green-* …
  ├─ 2. Semânticos        brand · surface · content · action · state · focus   (um bloco = um tema)
  ├─ 3. De componente     --control-height-* · --table-row-height · --shell-nav-width …
  └─ @theme inline        expõe 2 e 3 como utilitários do Tailwind (bg-*, text-*, border-*, h-*, …)
        │
components/ui  ──►  components/app  ──►  features/*  ──►  app/*
   (só utilitários de tokens semânticos ou de componente; nunca primitivo, nunca literal)
```

Regra de dependência: cada camada só lê a anterior. Primitivos nunca aparecem em utilitário do Tailwind, porque não são expostos no `@theme`. Assim um componente não tem como usar `--tj-red` diretamente.

---

## Decisões de design (aprovadas em 2026-10-11, WDS-D-026)

### DD-1 — Vocabulário dos tokens semânticos

**Aprovada a opção A.** Mantém-se a compatibilidade com os nomes semânticos do kit, sem renomeações desnecessárias, e nenhum componente usa cor primitiva.

Você pediu tokens separados por conceito (`brand`, `surface`, `content`, `state`, `focus`). Há duas formas de entregar isso.

| | A. Grupos conceituais com os nomes do kit (recomendada) | B. Renomear tudo por prefixo de grupo |
| --- | --- | --- |
| Como fica | A camada semântica é organizada e documentada nos cinco grupos; os nomes de uso continuam os do shadcn (`background`, `foreground`, `muted-foreground`, `border`, `destructive`, `sidebar-*`), mais os que faltam (`brand`, `success`, `warning`, `info`, `disabled`) | Nomes novos: `surface-page`, `content-primary`, `state-error`, `focus-on-shell` |
| Custo | Nenhuma reescrita de classe nas ~40 telas e 20 arquivos do kit | Reescrita de toda classe de cor do projeto |
| Kit | `shadcn add` continua gerando código compatível | Todo componente novo do kit precisa ser traduzido à mão |
| Risco | O nome `accent` do shadcn significa "superfície de hover", não "acento de marca"; por isso o vermelho se chama `brand` | Dois vocabulários convivem durante a migração |

Nos dois casos os componentes ficam desacoplados dos hexadecimais, que é o objetivo. A recomendação é A; o restante deste documento usa A.

### DD-4 — Valor de destaque em Barlow Condensed

**Aprovada condicionalmente à avaliação visual do espécime.** Se a leitura de números financeiros ficar prejudicada, o valor de destaque usa Geist.

O arquivo da Barlow Condensed Bold (repositório `google/fonts`) declara os recursos OpenType `tnum` e `pnum`, então algarismos tabulares existem. Proposta: só o valor de destaque isolado (saldo) usa Barlow Condensed 700 com `tabular-nums`. Toda tabela, lista, formulário e valor em linha usa Geist. Se, no espécime, o valor em Barlow Condensed não tiver boa leitura, ele passa a Geist 600, sem mudar componente (é um token de fonte do componente de valor).

Não verificado: se o subconjunto servido pelo `next/font` preserva o `tnum`. Confere-se no espécime, medindo a largura de "1111" contra "0000".

---

## Tokens

### 1. Primitivos (não expostos ao Tailwind)

| Token | Valor | Origem |
| --- | --- | --- |
| `--tj-ink` | `#111111` | Paleta oficial |
| `--tj-red` | `#D81E28` | Paleta oficial |
| `--tj-white` | `#FFFFFF` | Paleta oficial |
| `--tj-steel` | `#A7A9AC` | Paleta oficial |
| `--tj-neutral-50` … `-950` | `#F6F6F7`, `#ECEDEE`, `#DCDDDF`, `#C4C6C8`, `#A7A9AC`, `#8A8D91`, `#6B6E72`, `#4A4D50`, `#2A2C2E`, `#1C1C1E`, `#111111` | Derivados do aço (400 é o próprio aço; 950 é o preto) |
| `--tj-red-50`, `-700`, `-800` | `#FDECEC`, `#B3141D`, `#8F0F17` | Derivados do vermelho, só para erro |
| `--tj-green-50`, `-700`, `-800` | `#E8F5EC`, `#15803D`, `#166534` | Estado de sucesso (WDS-D-018) |
| `--tj-amber-50`, `-700`, `-800` | `#FDF3E3`, `#B45309`, `#92400E` | Estado de aviso (WDS-D-018) |

### 2. Semânticos (o tema)

Contrastes calculados pela fórmula WCAG; o teste de contraste (abaixo) é a autoridade final.

**brand**

| Token | Valor | Uso | Contraste |
| --- | --- | --- | --- |
| `brand` | `--tj-red` | Indicador de item ativo, filete do título de página, indicador de aba | 5,08 sobre branco; 3,72 sobre preto |
| `brand-foreground` | `--tj-white` | Texto sobre `brand` (raro) | 5,08 |

**surface**

| Token | Valor | Uso |
| --- | --- | --- |
| `background` | `--tj-white` | Fundo do conteúdo |
| `card`, `popover` | `--tj-white` | Cartões, menus, diálogos |
| `muted` | `--tj-neutral-50` | Cabeçalho de tabela, barra de filtros, esqueleto |
| `accent` | `--tj-neutral-100` | Hover de linha, de item de menu e de botão discreto |
| `secondary` | `--tj-neutral-100` | Preenchimento discreto |
| `disabled` | `--tj-neutral-100` | Fundo de controle desabilitado |
| `sidebar` | `--tj-ink` | Shell: cabeçalho e navegação |
| `sidebar-accent` | `--tj-neutral-800` | Hover de item do shell |
| `border` | `--tj-neutral-200` | Divisórias e contornos decorativos |
| `input` | `--tj-neutral-500` | Borda de campo (3,33 sobre branco) |
| `border-strong` | `--tj-ink` | Régua sob o cabeçalho de tabela e sob o título de seção |
| `sidebar-border` | `--tj-neutral-800` | Divisórias do shell |
| `overlay` | `--tj-ink` a 50% | Véu atrás de diálogo e gaveta |

**content**

| Token | Valor | Uso | Contraste |
| --- | --- | --- | --- |
| `foreground` (e `card-`, `popover-`, `accent-`, `secondary-foreground`) | `--tj-ink` | Texto principal | 18,88 sobre branco; 17,48 sobre `muted` |
| `muted-foreground` | `--tj-neutral-600` | Texto secundário, ajuda, metadado | 5,12 sobre branco; 4,74 sobre `muted` |
| `disabled-foreground` | `--tj-neutral-600` | Rótulo de controle desabilitado | 4,37 sobre `disabled` (isento pela WCAG, mantido legível) |
| `sidebar-foreground` | `--tj-white` | Texto do shell | 18,88 |
| `sidebar-muted-foreground` | `--tj-steel` | Rótulo de grupo e texto secundário do shell | 8,02 sobre preto; 5,95 sobre `sidebar-accent` |

**action**

| Token | Valor | Uso | Contraste |
| --- | --- | --- | --- |
| `primary` | `--tj-ink` | Botão primário | 18,88 com `primary-foreground` |
| `primary-foreground` | `--tj-white` | Texto do botão primário | |
| `primary-hover` | `--tj-neutral-800` | Hover do botão primário | 14,02 |
| `sidebar-primary` | `--tj-red` | Indicador do item ativo do shell | 3,72 sobre `sidebar` (gráfico) |

**state**

| Token | Valor | Uso | Contraste |
| --- | --- | --- | --- |
| `destructive` | `--tj-red-700` | Erro: texto, borda, botão destrutivo | 6,92 sobre branco |
| `destructive-foreground` | `--tj-white` | Texto do botão destrutivo | 6,92 |
| `destructive-subtle` | `--tj-red-50` | Fundo de alerta e de badge de erro | |
| `destructive-emphasis` | `--tj-red-800` | Texto sobre `destructive-subtle`; hover do botão destrutivo | 8,18 |
| `success` | `--tj-green-700` | Ícone e texto de sucesso sobre branco | 5,02 |
| `success-subtle` / `success-emphasis` | `--tj-green-50` / `--tj-green-800` | Fundo e texto de badge e alerta de sucesso | 6,35 |
| `warning` | `--tj-amber-700` | Ícone e texto de aviso sobre branco | 5,02 |
| `warning-subtle` / `warning-emphasis` | `--tj-amber-50` / `--tj-amber-800` | Fundo e texto de badge e alerta de aviso | 6,45 |
| `info` | `--tj-neutral-700` | Informação neutra | 8,51 sobre branco |
| `info-subtle` | `--tj-neutral-100` | Fundo de badge e alerta neutros | 7,26 |

**focus**

| Token | Valor | Uso | Contraste |
| --- | --- | --- | --- |
| `ring` | `--tj-ink` | Foco no conteúdo claro | 18,88 |
| `ring-offset` | `background` | Afastamento do anel | |
| `sidebar-ring` | `--tj-white` | Foco no shell | 18,88 |

Tokens do kit que saem: `--chart-1` a `--chart-5` (não há gráfico, WDS-D-010).

### 3. De componente

| Token | Valor | Em ponteiro de toque (`pointer: coarse`) |
| --- | --- | --- |
| `--control-height-sm` | 32 px | 44 px |
| `--control-height-md` | 36 px | 44 px |
| `--control-height-lg` | 44 px | 44 px |
| `--control-font-size` | 14 px | 16 px |
| `--touch-target` | 44 px | 44 px |
| `--table-head-height` / `--table-row-height` | 40 px / 44 px | 40 px / 48 px |
| `--shell-header-height` / `--shell-nav-width` | 56 px / 240 px | iguais |
| `--content-narrow` / `-default` / `-wide` | 640 / 1120 / 1440 px | iguais |
| `--page-gutter` | 16 px abaixo de 768; 24 px até 1023; 32 px a partir de 1024 | iguais |

O alvo de toque é resolvido no token, não no componente: sob `pointer: coarse` os controles passam a 44 px e os campos a 16 px de texto, sem variante nem classe condicional nos componentes (DSC-01 AC4, DSF-03 AC7).

### Tema e tema futuro

- O bloco `:root` contém os primitivos e a atribuição dos semânticos. Esse segundo conjunto é "o tema".
- Um tema escuro futuro é outro bloco (por exemplo `[data-theme="dark"]`) que reatribui só os semânticos. Nenhum componente muda (WDS-D-003).
- O bloco `.dark` atual é removido: tem valores neutros nunca verificados.
- A linha `@custom-variant dark (&:is(.dark *))` **permanece**. Sem ela, o Tailwind v4 ativaria as classes `dark:` do kit por `prefers-color-scheme`, o que violaria DSF-01 AC6. Com ela, e sem nenhum elemento com a classe `.dark`, essas classes ficam inertes. Na revisão do kit as classes `dark:` são removidas (DSM-05 AC5); a linha fica como proteção para componentes instalados depois.
- `html` declara `color-scheme: light`, para campos nativos e barras de rolagem não seguirem o tema do sistema.

---

## Tipografia

### Carregamento

| Fonte | Como | Variável CSS | Token Tailwind |
| --- | --- | --- | --- |
| Geist (variável) | `next/font/google`, já em `app/layout.tsx` | `--font-geist-sans` | `--font-sans` |
| Geist Mono | idem, já carregada | `--font-geist-mono` | `--font-mono` |
| Barlow Condensed 600 e 700 | `next/font/google` (`Barlow_Condensed`, `weight: ["600", "700"]`, `subsets: ["latin"]`, `display: "swap"`) | `--font-barlow-condensed` | `--font-display` |

- Correção de D-01: `--font-sans` passa a apontar para `--font-geist-sans` (hoje aponta para si mesma em `web/app/globals.css:10`).
- `--font-heading`, usado pelo kit em títulos de diálogo, continua igual a `--font-sans`. Título de diálogo é texto funcional.
- O `next/font` serve as fontes da própria aplicação, sem requisição a terceiros em tempo de execução (documentação em `web/node_modules/next/dist/docs/01-app/01-getting-started/13-fonts.md`).
- Barlow Condensed tem a função `Barlow_Condensed` no `next/font` instalado, com os pesos 600 e 700 disponíveis (conferido em `font-data.json`).

### Escala e papéis

A escala é a padrão do Tailwind, o que evita tokens novos de tamanho: 12, 14, 16, 18, 20, 24, 30 e 36 px. Valores arbitrários (`text-[13px]`) são recusados pela guarda.

| Papel | Componente | Fonte e peso | Tamanho / altura | Caixa e espaçamento |
| --- | --- | --- | --- | --- |
| Destaque | `DisplayTitle` | Barlow Condensed 700 | 36 / 40 | alta, 0,01em |
| Título de página | `PageTitle` (`h1`) | Barlow Condensed 700 | 30 / 36; 24 / 28 abaixo de 768 | alta, 0,01em |
| Título de seção | `SectionTitle` (`h2`) | Barlow Condensed 600 | 20 / 24 | frase comum |
| Rótulo de grupo e sobretítulo | `Eyebrow` | Barlow Condensed 600 | 14 / 16 | alta, 0,06em |
| Título de bloco | `h3`, título de diálogo | Geist 600 | 16 / 24 | frase comum |
| Corpo | — | Geist 400 | 14 / 20 | frase comum |
| Corpo maior | — | Geist 400 | 16 / 24 | frase comum |
| Rótulo | rótulo de campo, botão | Geist 500 | 14 / 20 | frase comum |
| Legenda | ajuda, metadado | Geist 400 | 12 / 16 | frase comum |
| Dado | número, dinheiro | Geist 400 ou 500, `tabular-nums` | 14 / 20 | — |
| Valor de destaque | `MoneyValue` tamanho `display` | Barlow Condensed 700, `tabular-nums` (DD-4) | 36 / 40 | — |
| Identificador | código curto | Geist Mono 400 | 14 / 20 | — |

- Os papéis em Barlow Condensed existem só como componentes em `components/ui/typography.tsx`. A guarda recusa `font-display` e `uppercase` fora desse arquivo e do shell, o que torna verificáveis DSF-03 AC3 e AC4.
- Todos os títulos usam `text-wrap: balance`.
- Em relação a `context.md`: o rótulo de grupo passa de 13 para 14 px e o identificador de 13 para 14 px, para ficar na escala padrão.

---

## Espaçamento

Base de 4 px (a do Tailwind, `--spacing: 0.25rem`). Passos permitidos: 1, 2, 3, 4, 5, 6, 8, 10, 12 e 16 (4 a 64 px). Valores arbitrários (`p-[13px]`) são recusados pela guarda.

| Relação | Valor |
| --- | --- |
| Ícone a texto | 8 |
| Rótulo a campo; campo a mensagem | 8 |
| Entre campos | 16 |
| Entre grupos de campos; entre blocos de uma seção | 24 |
| Entre seções da página | 32 |
| Cabeçalho de página ao conteúdo | 24 |
| Padding de célula de tabela | 12 horizontal, 8 vertical |
| Padding de cartão e de diálogo | 16 abaixo de 768; 24 acima |
| Margem lateral de página | `--page-gutter` (16, 24, 32) |

## Raio

| Token | Valor | Uso |
| --- | --- | --- |
| `--radius-sm` | 4 px | Badges retangulares, caixa de seleção, foco de link |
| `--radius-md` | 6 px | Botões, campos, itens de menu |
| `--radius-lg` | 8 px | Cartões, diálogos, toasts, menus |
| `rounded-full` | pílula | Só indicador de status |

Os passos maiores gerados pelo kit (`xl` a `4xl`) são fixados em 8 px, para que nenhum componente instalado depois saia mais arredondado.

## Bordas

| Borda | Largura e cor | Uso |
| --- | --- | --- |
| Decorativa | 1 px `border` | Cartão, divisória, linha de tabela |
| De controle | 1 px `input` | Campos (contraste 3,33) |
| Forte | 2 px `border-strong` | Régua sob o cabeçalho de tabela e sob o título de seção |
| De marca | 2 px `brand` | Filete de 32 px sob o título de página; 3 px no item ativo do shell |
| De erro | 1 px `destructive` | Campo inválido |

A régua preta e o filete vermelho são os dois detalhes gráficos que carregam a identidade no conteúdo, no lugar de cartões e sombras.

## Sombras e profundidade

| Nível | Valor | Uso |
| --- | --- | --- |
| 0 | nenhuma | Tudo em repouso: cartões, tabelas, campos, shell |
| 1 | `0 4px 12px` de `--tj-ink` a 12% | Menu, popover, toast |
| 2 | `0 16px 40px` de `--tj-ink` a 20% | Diálogo, gaveta |

Profundidade no conteúdo é dada por borda e por fundo `muted`. Sem gradiente e sem brilho (DSF-04 AC4).

---

## Estados de interação

| Estado | Conteúdo claro | Shell escuro |
| --- | --- | --- |
| Repouso | tokens base | tokens base |
| Hover | Primário: `primary-hover`. Demais superfícies: `accent` | `sidebar-accent` |
| Pressionado | Um passo além do hover; sem deslocamento | idem |
| Foco visível | Anel de 2 px `ring` com 2 px de afastamento | Anel de 2 px `sidebar-ring` |
| Selecionado ou ativo | Peso 600 e marca gráfica | Peso 600 e barra `sidebar-primary` de 3 px sobre `sidebar` |
| Desabilitado | Fundo `disabled`, texto `disabled-foreground`, cursor padrão | Texto `sidebar-muted-foreground` |
| Inválido | Borda `destructive`, mensagem com ícone | — |
| Somente leitura | Fundo `muted`, sem borda de controle | — |
| Em andamento | Desabilitado, `aria-busy`, rótulo com "…" | — |

O indicador ativo do shell fica sempre sobre `sidebar`, nunca sobre `sidebar-accent`: vermelho sobre a superfície de hover dá 2,76, abaixo de 3:1.

---

## Iconografia

- Biblioteca: `lucide-react`, já instalada. Traço padrão da biblioteca.
- Tamanhos: 16 px em linha, botões e badges; 20 px na navegação e em cabeçalhos; 24 px em estado vazio e página de erro.
- Cor: herdam a cor do texto. Só ícones de estado usam tom semântico.
- Decorativos levam `aria-hidden`. Ícone sem texto ao lado só em `IconButton`, com nome acessível.
- Sem círculo colorido atrás de ícone e sem ilustração.

| Uso | Ícone (export do Lucide, conferido na versão instalada) |
| --- | --- |
| Início | `House` |
| Lançamentos, Contas, Saldo | `Receipt`, `ListTree`, `Wallet` |
| Produtos | `Package` |
| Usuários | `Users` |
| Menu, mais ações | `Menu`, `EllipsisVertical` |
| Criar, buscar, filtros | `Plus`, `Search`, `SlidersHorizontal` |
| Sucesso, aviso, erro, informação | `CircleCheck`, `TriangleAlert`, `CircleAlert`, `Info` |
| Em andamento | `LoaderCircle` |
| Vazio, sem resultado, sem acesso | `Inbox`, `SearchX`, `Lock` |
| Trocar senha, sair | `KeyRound`, `LogOut` |
| Anexo, baixar, devolver | `Paperclip`, `Download`, `Undo2` |

Os ícones de status de domínio estão em `02-componentes.md`.

---

## Movimento

| Token | Valor | Uso |
| --- | --- | --- |
| `--duration-fast` | 120 ms | Hover, foco, troca de cor |
| `--duration-base` | 200 ms | Menu, toast, expansão |
| `--duration-slow` | 300 ms | Diálogo e gaveta |
| `--ease-out` | `cubic-bezier(0.2, 0, 0, 1)` | Entradas |
| `--ease-in` | `cubic-bezier(0.4, 0, 1, 1)` | Saídas |

- Anima-se só `opacity` e `transform` em entradas e saídas, e `background-color`, `border-color` e `color` em estados. `transition-all` sai de `button.tsx` e `badge.tsx`.
- Nenhuma animação decorativa: sem entrada de página, sem animação de lista. O esqueleto pulsa e é o único movimento contínuo.
- `prefers-reduced-motion: reduce`: uma regra na camada base reduz `animation-duration` e `transition-duration` a 0,01 ms em todos os elementos. Cobre também as animações do `tw-animate-css` usadas pelo kit.

---

## Marca

### Escudo

- Arquivo: `web/public/brand/torcida-jovem-logo.svg`, versionado.
- Ajuste único: o atributo `viewBox` passa de `0 0 595.28 841.89` (prancheta A4) para `0 118.275 591.17 591.17`, um quadrado centrado no círculo do escudo (centro 295,58 por 413,86; raio 293,86). Nenhum `path`, classe ou cor muda.
- Teste de integridade (`brand.test.ts`): a lista de atributos `d` e o bloco de estilos do arquivo versionado são comparados, por hash, com os do arquivo recebido. Qualquer alteração de desenho falha.
- Componente `BrandMark` (`components/app/brand.tsx`): `size` (32, 40, 64, 96), `withName`. Com nome, `alt=""`; sem nome, `alt="Torcida Jovem"`. Sempre com `width` e `height`.
- Regras de uso: proporção 1:1, mínimo de 32 px, respiro de um oitavo do diâmetro, sem recolorir, girar, recortar ou aplicar efeito.
- O escudo tem fundo preto próprio. Funciona sobre o shell escuro e sobre o conteúdo claro sem versão negativa.

### Nome

"TORCIDA JOVEM" em Barlow Condensed 700, caixa alta, ao lado do escudo no shell e abaixo dele no login (WDS-D-020). É texto, não imagem.

### Limpeza

Saem de `web/public/` os arquivos do `create-next-app`: `file.svg`, `globe.svg`, `next.svg`, `vercel.svg`, `window.svg` (nenhum é referenciado pelo código).

---

## Espécime

- **Rota**: `web/app/(dev)/design-system/page.tsx`, fora do grupo autenticado, sem sessão.
- **Produção**: a página chama `notFound()` quando `process.env.NODE_ENV === "production"` (DSF-06 AC5). A revisão do mantenedor é feita com `pnpm dev`.
- **Sem componente exclusivo**: a página só compõe os componentes reais. Os dados de exemplo são listas estáticas no próprio diretório da rota. O shell é mostrado pelo `ShellFrame`, a parte de apresentação do `AppShell` (`04-shell-e-acesso.md`), com nome e itens de exemplo.
- **Componentes sem uso na V1** (Radio, Switch, Tabs, Tooltip) não aparecem (WDS-D-022).

| Seção | Conteúdo |
| --- | --- |
| Marca | Escudo nos quatro tamanhos, sobre claro e sobre shell; nome |
| Cores | Primitivos; semânticos por grupo, cada um com o par de contraste e a razão calculada |
| Tipografia | Cada papel da escala com texto real em português, inclusive acentos e caixa alta |
| Espaço, raio, borda, sombra | Amostras com o nome do token |
| Ações | Button em todas as variantes e tamanhos, em repouso, foco, desabilitado e em andamento; IconButton |
| Campos | Field com Input, Textarea, Select nativo, Checkbox, MoneyField e QuantityField, em repouso, foco, inválido, desabilitado e somente leitura |
| Status e valores | Todos os status de domínio; MoneyValue nas quatro variantes e no tamanho de destaque |
| Tabela e lista | A listagem de exemplo nos três comportamentos; uma tabela pequena que continua tabela |
| Estados de tela | Carregando, vazio, sem resultado, erro, sem acesso |
| Feedback | Alertas nos quatro tons; botão que dispara cada toast; confirmação com motivo |
| Sobreposições | Diálogo de formulário, gaveta, menu de ações |
| Shell | `ShellFrame` com navegação, item ativo e menu da pessoa |
| Página de exemplo | Cabeçalho de página, barra de filtros e listagem juntos, como numa tela real |

Exemplos responsivos: o revisor redimensiona a janela (390, 768, 820, 1024, 1440). A listagem de exemplo também é mostrada com o layout forçado (`layout="list"` e `layout="table"`), que é uma propriedade real do componente.

**Portão**: a aprovação do espécime é registrada em `../STATE.md` antes de qualquer tela ser migrada (DSF-06 AC6).

---

## Guardas automatizadas

| Teste | Local | O que verifica | Requisito |
| --- | --- | --- | --- |
| Guarda de tokens | `web/test/design/tokens-guard.test.ts` | Em `app`, `components` e `features` (fora `globals.css`): nenhuma cor literal, nenhum utilitário de paleta crua (`red-500`), nenhum valor arbitrário de cor, tamanho de texto ou espaço, nenhum `font-display` ou `uppercase` fora dos arquivos permitidos, nenhum `transition-all`, nenhuma classe `dark:` | DSF-01 AC7, DSF-03 AC3-5, DSF-04 AC1-5, DSM-05 |
| Contraste | `web/test/design/contrast.test.ts` | Lê `globals.css`, resolve cada par declarado em `web/lib/design/contrast-pairs.ts` e exige 4,5:1 (texto) ou 3:1 (gráfico) | DSF-02 AC1-3 |
| Tokens órfãos | mesmo arquivo da guarda | Todo token usado em utilitário existe no `@theme`; todo semântico tem valor | Caso de borda da spec |
| Integridade do escudo | `web/test/design/brand.test.ts` | Hash dos traçados e estilos igual ao do arquivo recebido; `viewBox` quadrado | DSF-05 AC1 |

As guardas seguem o modelo de `web/app/server-client-boundary.test.ts`: leem arquivos e não dependem de navegador.

---

## Code Reuse Analysis

| Existente | Local | Uso |
| --- | --- | --- |
| Bloco `@theme inline` e variáveis semânticas | `web/app/globals.css` | Mantida a estrutura; trocam-se os valores e entram os tokens que faltam |
| Carregamento de Geist e Geist Mono | `web/app/layout.tsx` | Mantido; entra a Barlow Condensed |
| `tw-animate-css`, `shadcn/tailwind.css` | `web/app/globals.css` | Mantidos |
| Modelo de guarda por leitura de arquivos | `web/app/server-client-boundary.test.ts` | Base das guardas novas |
| `lucide-react` | `package.json` | Única biblioteca de ícones |

## Risks & Concerns

| Concern | Location | Impact | Mitigation |
| --- | --- | --- | --- |
| `--font-sans` referencia a si mesma | `web/app/globals.css:10` | Toda a interface em Times | Corrigido na primeira tarefa da fundação; conferido no navegador pela fonte computada |
| Classes `dark:` do kit ativadas pelo sistema | `web/components/ui/*.tsx` | Tema escuro não verificado apareceria por `prefers-color-scheme` | Manter `@custom-variant dark`; remover as classes na revisão do kit; a guarda recusa `dark:` |
| Componentes com valores arbitrários do kit (`rounded-[min(...)]`, `text-[0.8rem]`) | `web/components/ui/button.tsx:24-31` | A guarda falharia no próprio kit | A revisão do kit reescreve esses pontos com tokens de componente |
| Subconjunto da Barlow Condensed sem `tnum` | `next/font` | Valor de destaque desalinhado | Conferir no espécime; reserva em Geist (DD-4) |
| Vermelho de erro confundido com o de marca | tokens `brand` e `destructive` | Leitura errada de estado | Tons distintos, erro sempre com ícone e texto, foco nunca vermelho |
| Espécime exposto em produção | `app/(dev)/design-system` | Página interna pública | `notFound()` em produção, com teste que renderiza a página com `NODE_ENV=production` |
| Guarda rígida demais atrasa o trabalho | `tokens-guard.test.ts` | Falsos positivos | Lista explícita de arquivos e padrões permitidos, no próprio teste, com o motivo de cada exceção |

## Tech Decisions

| Decisão | Escolha | Motivo |
| --- | --- | --- |
| Nomes de uso dos tokens | Os do kit, agrupados por conceito (DD-1) | Sem reescrita em massa; compatível com `shadcn add` |
| Alvo de toque | Token de altura sob `pointer: coarse` | Componentes não ganham variante nem lógica |
| Papéis tipográficos de marca | Componentes, não classes soltas | Torna verificável onde a condensada e a caixa alta aparecem |
| Escala de texto e de espaço | As padrão do Tailwind, com valores arbitrários proibidos | Menos tokens para manter |
| Espécime em rota de desenvolvimento | `notFound()` em produção | WDS-D-017; referência viva sem exposição |
