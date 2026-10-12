# Shell e Acesso — Design

**Spec**: `.specs/features/web-design/spec/04-shell-e-acesso.md`
**Status**: Aprovado pelo mantenedor em 2026-10-11
**Depende de**: `01-fundacao-visual.md`, `02-componentes.md`, `03-padroes.md`.

Cobre: shell e navegação, telas de acesso, páginas de erro, metadados e ícones, página inicial.

---

## Architecture Overview

```
app/layout.tsx                 fontes · metadata · viewport (theme-color) · AppProviders
├─ app/(public)/layout.tsx     AccessLayout: escudo + nome + formulário
│    └─ entrar · recuperar-acesso · redefinir-senha
├─ app/(app)/layout.tsx        portão de sessão (inalterado) → AppShell
│    ├─ AppShell               lê sessão e permissões → ShellFrame
│    │    └─ ShellFrame        apresentação pura: cabeçalho, navegação, gaveta, menu da pessoa
│    ├─ app/(app)/error.tsx    erro dentro do shell
│    └─ páginas → PageContainer (main#conteudo)
├─ app/not-found.tsx           endereço inexistente
├─ app/global-error.tsx        falha no layout raiz
└─ app/(dev)/design-system     espécime (usa ShellFrame com dados de exemplo)
```

O `AppShell` é dividido em duas partes para que o espécime mostre o shell real sem sessão (WDS-D-017): `ShellFrame` não conhece sessão nem API; `AppShell` continua com a mesma interface de hoje (`items`, `children`).

---

## Shell

### Anatomia

```
LG (a partir de 1024)
┌───────────────────────────────────────────────────────────────────┐
│ [escudo] TORCIDA JOVEM                              Nome da pessoa ▾│  cabeçalho 56 · sidebar
├──────────────┬────────────────────────────────────────────────────┤
│ ⌂ Início     │                                                    │
│              │   Trilha                                           │
│ FINANCEIRO   │   TÍTULO DE PÁGINA                    [ Ação ]     │
│▌▤ Lançamentos│   ──                                               │
│ ☰ Contas     │                                                    │
│ ▣ Saldo      │   conteúdo claro (background)                      │
│              │                                                    │
│ ESTOQUE      │                                                    │
│ ▢ Produtos   │                                                    │
│              │                                                    │
│ ADMINISTRAÇÃO│                                                    │
│ ☺ Usuários   │                                                    │
└──────────────┴────────────────────────────────────────────────────┘
  navegação 240 · sidebar

Abaixo de 1024
┌───────────────────────────────────────┐
│ [≡] [escudo] TORCIDA JOVEM    Nome ▾  │   navegação na gaveta
├───────────────────────────────────────┤
│  conteúdo                             │
```

### ShellFrame (`components/app/shell-frame.tsx`)

```ts
type ShellFrameProps = {
  entries: readonly NavEntry[];          // já filtradas por permissão
  currentPath: string;
  userName: string;
  userMenu: ReactNode;                   // itens do menu da pessoa
  showNav: boolean;                      // falso durante a troca obrigatória de senha
  children: ReactNode;
};
```

| Parte | Especificação |
| --- | --- |
| Link de pular | Primeiro elemento focável: "Pular para o conteúdo". Oculto até receber foco; então aparece no canto superior esquerdo, sobre o shell. Leva o foco a `main#conteudo` |
| Cabeçalho | `header` fixo no topo, 56 px, fundo `sidebar`, divisória inferior `sidebar-border`. Respeita `env(safe-area-inset-top)` e as laterais |
| Marca | `BrandMark` de 32 px e "TORCIDA JOVEM" em Barlow Condensed 700. Link para `/inicio`. Abaixo de 400 px de largura o nome some e fica só o escudo, com `alt="Torcida Jovem"` |
| Navegação | `nav` com `aria-label="Navegação principal"`, 240 px, fundo `sidebar`, fixa a partir de 1024, com rolagem própria se passar da altura |
| Grupo | `Eyebrow` em `sidebar-muted-foreground`; grupos separados por 24 px |
| Item | 36 px de altura (44 em toque), ícone de 20 px e rótulo em Geist 14; hover `sidebar-accent`; foco `sidebar-ring` |
| Item ativo | `aria-current="page"`, peso 600, texto `sidebar-foreground`, barra `sidebar-primary` de 3 px à esquerda, sobre fundo `sidebar`. Três sinais além da cor |
| Itens inativos | Texto `sidebar-foreground` a peso 400. O contraste é o mesmo; a diferença é peso e barra |
| Área de conteúdo | Fundo `background`. As páginas trazem o próprio `main` pelo `PageContainer` |

### Menu da pessoa

- Gatilho: nome (truncado a 24 caracteres visíveis) e `ChevronDown`, no cabeçalho; 44 px de alvo em toque.
- Conteúdo: nome completo e e-mail no topo, como texto; "Trocar senha" como **link** (`KeyRound`); divisória; "Sair" como botão (`LogOut`).
- Durante a troca obrigatória de senha: só "Sair" (FND-05 AC8, mantido).
- O menu abre sobre o conteúdo claro e usa os tokens de `popover`.

### Navegação responsiva (WDS-D-019)

| Largura | Comportamento |
| --- | --- |
| A partir de 1024 | Navegação fixa; sem botão de menu |
| Abaixo de 1024 | Botão "Abrir menu" (`IconButton`, 44 px) no cabeçalho; navegação em `Sheet` lateral esquerda de 280 px, com o mesmo `NavList` e fundo `sidebar` |

- A gaveta fecha ao escolher um item, com Esc, com o botão de fechar e com clique no véu. O foco volta ao botão de menu.
- Se a janela passar a 1024 px com a gaveta aberta, ela fecha (caso de borda da spec).
- As classes de ponto de quebra do shell passam de `md:` para `lg:`. É a única mudança de comportamento do shell em relação a FND-05, e é reverificada (DSS-02 AC7).

### Modelo de navegação

`NavItem` (`components/app/nav.ts`) ganha o campo opcional `icon: LucideIcon`. Cada módulo define os ícones no próprio `features/<módulo>/nav.ts`, que já é dele. `visibleItems`, a ordem e as permissões não mudam (INT-01).

| Item | Ícone |
| --- | --- |
| Início | `House` |
| Lançamentos | `Receipt` |
| Contas | `ListTree` |
| Saldo | `Wallet` |
| Produtos | `Package` |
| Usuários | `Users` |

---

## Telas de acesso

### AccessLayout (`app/(public)/layout.tsx` e `components/app/access-layout.tsx`)

```
LG                                              Abaixo de 1024
┌───────────────────┬──────────────────────┐    ┌──────────────────────┐
│                   │                      │    │      [escudo 64]     │
│                   │   ENTRAR             │    │    TORCIDA JOVEM     │
│   [escudo 96]     │   ──                 │    ├──────────────────────┤
│   TORCIDA JOVEM   │   E-mail  [_______]  │    │   ENTRAR             │
│                   │   Senha   [_______]  │    │   ──                 │
│                   │   [ Entrar        ]  │    │   E-mail  [_______]  │
│                   │   Esqueci minha senha│    │   Senha   [_______]  │
│   sidebar (preto) │   background (claro) │    │   [ Entrar        ]  │
└───────────────────┴──────────────────────┘    └──────────────────────┘
```

- A partir de 1024: painel de marca à esquerda (40% da largura, fundo `sidebar`), com escudo de 96 px e o nome "Torcida Jovem" em `DisplayTitle`; formulário à direita, centralizado, com largura máxima de 400 px. Aprovado em WDS-D-027.
- Abaixo de 1024: faixa de marca no topo (fundo `sidebar`, escudo de 64 px e nome), formulário abaixo em largura total com margem de 16.
- Sem fotografia, ilustração, gradiente ou A FERA (DSS-03 AC5, WDS-D-013). "Desde 2013" não é repetido como texto decorativo; a inscrição continua só dentro do escudo, que é preservado (WDS-D-027).
- Título de cada tela em `PageTitle` com o filete; formulário em `FormLayout`; botão primário em largura total.
- "TJ Platform" sai do layout público (WDS-D-020).

| Tela | Mudança |
| --- | --- |
| `/entrar` | Layout novo; `spellCheck={false}` no e-mail; link "Esqueci minha senha" abaixo do botão |
| `/recuperar-acesso`, `/redefinir-senha` | Layout novo; link "Voltar para entrar" |
| `/conta/senha` | Durante a troca obrigatória, usa o `AccessLayout` dentro do grupo autenticado, com "Sair" no canto; fora dela, é uma página comum do shell, com largura `narrow` |

Nenhum comportamento de FND-03 e de ACS-01 a ACS-03 muda: mensagens, redirecionamentos, tratamento do token e do `Retry-After` ficam como estão.

---

## Páginas de erro

Convenções conferidas em `web/node_modules/next/dist/docs/01-app/03-api-reference/03-file-conventions/` (`error.md`, `not-found.md`).

| Arquivo | Quando aparece | Layout | Conteúdo |
| --- | --- | --- | --- |
| `app/(app)/error.tsx` | Exceção de renderização numa rota autenticada | Dentro do shell (o `error.js` é envolvido pelo layout do mesmo segmento) | `h1` "Algo deu errado", frase "Não foi possível mostrar esta página.", "Tentar de novo" (primário) e "Voltar ao início" (link) |
| `app/not-found.tsx` | Endereço inexistente | `AccessLayout`, sem formulário | `h1` "Página não encontrada", frase, link "Ir para o início" |
| `app/global-error.tsx` | Exceção no layout raiz | Documento próprio (`html` e `body`), como a convenção exige | Escudo, `h1` "Algo deu errado", "Tentar de novo" |

- São Client Components (`"use client"`), como a convenção exige para `error` e `global-error`.
- O foco vai para o `h1` ao montar (`tabIndex={-1}`), e o título do documento muda.
- Nenhuma mostra `error.message`, pilha nem `digest` (DSS-04 AC4).
- "Tentar de novo" usa a função de nova tentativa que a convenção entrega. A documentação instalada recomenda `retry` (a propriedade `unstable_retry` existe desde a 16.2) no lugar de `reset`; o nome exato é confirmado na documentação instalada no momento da implementação, como `web/AGENTS.md` manda.
- `global-error` não recebe os estilos globais do layout raiz. Ele importa `globals.css` e aplica as variáveis de fonte por conta própria. A confirmar no build: se a importação não for aceita ali, a página usa um bloco mínimo de estilos locais com os mesmos valores dos tokens, registrado como exceção na guarda.
- `global-not-found` (experimental) não é usado: há um único layout raiz.

---

## Metadados e ícones

| Item | Onde | Valor |
| --- | --- | --- |
| Idioma | `app/layout.tsx` | `lang="pt-BR"` (já existe) |
| Título | `metadata.title` | `{ default: "TJ Platform", template: "%s | TJ Platform" }` (já existe; mantido por WDS-D-020) |
| Descrição | `metadata.description` | Mantida |
| `theme-color` | `export const viewport` em `app/layout.tsx` | `#111111`, a cor do shell |
| `color-scheme` | `viewport.colorScheme` e CSS | `light` |
| Ícone da aba | `app/icon.svg` | O SVG do escudo, com o enquadramento corrigido, sem outra alteração |
| `favicon.ico` | `app/favicon.ico` | Removido: é o ícone padrão do Next. O `icon.svg` o substitui |
| Ícone do iOS | `app/apple-icon.png` | PNG de 180 por 180 px rasterizado do escudo (DD-5) |

**DD-5 — Ícone do iOS (aprovada, WDS-D-026).** A convenção só aceita `apple-icon` em PNG ou JPG. Um PNG de 180 por 180 px é gerado por rasterização do próprio SVG do escudo, sem redesenho nem modificação da arte, e versionado ao lado do `icon.svg` (WDS-D-021).

Se o escudo ficar ilegível em 16 px, isso é registrado em `../STATE.md` como necessidade de asset oficial reduzido. Nenhuma versão simplificada é criada.

---

## Página inicial

- `PageContainer` `default`, `PageHeader` com o título "Início" e a descrição "Olá, {primeiro nome}."
- Atalhos: grade de `Card` clicáveis (o cartão inteiro é um link), cada um com ícone de 20 px, nome da área em Geist 600 e uma frase de descrição. Uma, duas ou três colunas conforme a largura.
- A lista vem de `navItems` filtrada por permissão, como hoje (INT-01 AC4). A frase de descrição entra como campo opcional `description` em `NavItem`, definida por cada módulo.
- Sem indicador, gráfico ou dado de exemplo (WDS-D-010).
- Sem nenhuma área disponível: `EmptyState` explicando que não há áreas liberadas para a conta.

---

## Code Reuse Analysis

| Existente | Local | Uso |
| --- | --- | --- |
| `AppShell`, `NavList`, `NavLink` | `components/app/app-shell.tsx` | Viram `ShellFrame`; `AppShell` fica como adaptador de sessão |
| `visibleItems`, `isNavSection`, tipos | `components/app/nav.ts` | Mantidos; entram `icon` e `description` opcionais |
| `nav-items.ts` e `features/*/nav.ts` | — | Mantidos; ganham ícones |
| Portão de sessão | `app/(app)/layout.tsx` | Lógica inalterada; os dois `main` avulsos de carregamento e erro passam a usar `LoadingState` e `ErrorState` |
| `Sheet`, `DropdownMenu` | `components/ui` | Gaveta e menu da pessoa |
| Formulários de acesso | `app/(public)/entrar/login-form.tsx`, `features/identity/acesso/*` | Só layout e atributos de campo |

## Risks & Concerns

| Concern | Location | Impact | Mitigation |
| --- | --- | --- | --- |
| Mudança do ponto de quebra altera requisito verificado | `components/app/app-shell.tsx:45,84` (`md:hidden`, `md:block`) | FND-05 AC2 deixa de estar coberto pela verificação anterior | Reexecução do critério e smoke em 390, 768, 820, 1024 e 1440 (DSS-02 AC7) |
| Testes do shell dependem da marca "TJ Platform" e da estrutura atual | `components/app/app-shell.test.tsx` | Falham com a marca nova | Alterados com citação de WDS-D-020; lista em `05-migracao-web-v1.md` |
| `global-error` sem estilos globais | `app/global-error.tsx` | Página de erro sem identidade | Importação explícita, com reserva em estilos locais; conferida no build de produção |
| Escudo ilegível em 16 px | `app/icon.svg` | Aba pouco reconhecível | Registrar a dependência; não simplificar o desenho |
| Vermelho sobre hover do shell abaixo de 3:1 | item ativo | Indicador pouco visível | Indicador sempre sobre `sidebar`; item ativo não muda de fundo no hover |
| SVG servido por componente de imagem | `BrandMark` | O otimizador de imagens do Next trata SVG de forma restrita | `BrandMark` usa `img` simples com `width` e `height`, ou `next/image` com `unoptimized`; a escolha é confirmada na documentação instalada na implementação |
| Gaveta e navegação fixa com o mesmo `nav` | `ShellFrame` | Duas regiões de navegação com o mesmo nome | A gaveta só existe no DOM quando aberta; nomes distintos ("Navegação principal" e "Menu") |

## Tech Decisions

| Decisão | Escolha | Motivo |
| --- | --- | --- |
| Shell em duas partes | `ShellFrame` e `AppShell` | Espécime com o shell real, sem sessão |
| Ícones e descrições da navegação | Campos opcionais em `NavItem`, definidos por módulo | Mantém a propriedade de arquivos por módulo |
| Erro autenticado | `app/(app)/error.tsx` | Fica dentro do shell; a pessoa não perde a navegação |
| Tela de acesso em duas colunas no desktop | Painel de marca e formulário | Identidade forte só onde não há operação |
| `theme-color` | Preto do shell | A barra do navegador no celular continua a moldura |
