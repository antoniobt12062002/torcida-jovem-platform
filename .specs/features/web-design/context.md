# Web Design — Context

**Gathered:** 2026-10-11
**Specs:** `.specs/features/web-design/spec/`
**Status:** Aprovado em 2026-10-11; detalhado em `design/`

Saída do Discuss. Registra a direção visual aprovada e a proposta de fundação que o Design vai detalhar e o espécime vai validar. Os valores marcados como "proposta" podem ser ajustados no Design sem reabrir decisão; as regras (o que cada coisa pode ou não fazer) estão nas specs.

---

## Feature Boundary

Fundação visual, tokens, componentes, padrões compartilhados, shell, telas de acesso, páginas de erro e migração das telas existentes da Web V1 (WDS-D-010). Nada de API, regra de negócio, módulo novo, dashboard ou área pública.

---

## Briefing

| Tema | Resposta |
| --- | --- |
| O que | Linguagem visual e Design System da Web, aplicados a Identity, Financeiro e Estoque e prontos para módulos futuros. |
| Quem usa | Diretoria, tesouraria, estoque e loja, Conselho Fiscal. Uso administrativo, com leitura prolongada, tabelas e dinheiro. |
| Onde | Desktop como referência; celular como uso real para operações rápidas, consulta, estoque e movimentações (WDS-D-014). |
| Restrições técnicas | Next 16, React 19, Tailwind v4, shadcn `base-nova` sobre `@base-ui/react`, `lucide-react`, sem dependência nova, pt-BR. |
| Assets | Só o escudo. Sem manual de marca, sem A FERA. |
| Referências visuais | O escudo e a direção confirmada pelos pares de contraste (WDS-D-015). Nenhuma referência externa foi adotada. |

---

## Direção visual consolidada

**Humor**: firme, direto, sóbrio no conteúdo e com presença no contorno.

| Dimensão | Direção |
| --- | --- |
| Estrutura | Moldura escura (cabeçalho e navegação em preto, com o escudo) e conteúdo claro. |
| Cor | Preto sustenta e age; vermelho marca e sinaliza; branco e neutros derivados do aço organizam. |
| Tipografia | Barlow Condensed forte em título de página e de seção; Geist em todo o texto funcional. |
| Densidade | Compacta e agrupada: 14 px em tabelas, controles de 36 px, 44 px em toque. |
| Formas | Cantos firmes (4, 6, 8 px), bordas definem blocos, sombra só em sobreposições. |
| Iconografia | Lucide, traço único, 16 e 20 px, acompanhada de texto. |
| Marca | Escudo no shell, no login e nas páginas de erro, sempre o arquivo oficial. |
| A FERA | Ausente até existir asset oficial; nunca na área operacional. |
| Estados | Cada estado tem cor, ícone e rótulo. |

**Como a identidade aparece sem virar "dashboard genérico com vermelho"**

- Proporção: título de página condensado, em caixa alta e peso 700, contra um corpo discreto. O contraste de escala é o principal sinal de marca no conteúdo.
- Preto como estrutura: moldura, botão principal, texto e régua de cabeçalho de tabela usam o mesmo preto. A interface é reconhecível mesmo sem vermelho na tela.
- Vermelho raro e preciso: uma barra no item ativo da navegação, um filete sob o título de página, o indicador de aba ativa. Nunca fundo de área, nunca botão comum.
- Detalhes do shell: rótulos de grupo em condensada com espaçamento largo, escudo com área de respiro, separadores finos em aço.
- Planura: sem gradiente, sem brilho, sem sombra em repouso. Blocos se separam por borda e espaço, e listas não são embrulhadas em cartão sem necessidade.

**Do escudo, o que se preserva, moderniza ou não serve**

| Elemento do escudo | Tratamento no produto |
| --- | --- |
| Tríade preto, vermelho e branco, com cinza de apoio | Preservada como base de toda a paleta. |
| Forma circular com anel | Preservada só no próprio escudo; não vira motivo decorativo de componentes. |
| Letreiro em caixa alta | Modernizado: caixa alta em condensada, restrita a títulos e rótulos. |
| Ilustração do cavalo, com contorno pesado | Não serve como linguagem de interface; fica restrita ao escudo. |
| Cinza quente da ilustração (`#AD9F9C`) | Não vira token; o neutro do sistema deriva do aço oficial `#A7A9AC`. |

---

## Princípios

1. **Preto sustenta, vermelho sinaliza.** Cor tem função.
2. **Dado antes de decoração.** Nada compete com tabela, valor e formulário.
3. **Identidade nos detalhes.** Tipografia, proporção, shell e estados; nunca textura ou efeito.
4. **Um padrão por situação.** Uma forma de confirmar, de errar, de esvaziar, de filtrar.
5. **Acessível por construção.** WCAG 2.2 AA como piso; cor nunca é o único sinal.
6. **Celular é uso real.** Todo padrão define o comportamento estreito.
7. **Dinheiro sem ambiguidade.** Algarismos tabulares, sinal explícito, consequência dita antes de ação irreversível.
8. **Mostra o que a pessoa pode fazer e explica quando nega.**
9. **Português direto.** Frase comum, verbo na ação.
10. **Token primeiro.** Nenhum valor solto em tela.

---

## Arquitetura de tokens

Três camadas, todas na folha de estilos global (`web/app/globals.css`), no formato que o Tailwind v4 já usa (`@theme inline` mais variáveis em `:root`).

| Camada | Exemplo | Quem usa |
| --- | --- | --- |
| Primitivos | `--tj-ink`, `--tj-red`, `--tj-steel`, `--tj-neutral-600`, `--tj-red-700` | Só a camada semântica |
| Semânticos | `--background`, `--foreground`, `--primary`, `--muted-foreground`, `--border`, `--ring`, `--destructive`, `--success`, `--warning`, `--sidebar-*` | Componentes e telas, por utilitários do Tailwind |
| De componente | altura de controle, raio de controle, altura de linha de tabela | Só o componente dono, quando o semântico não basta |

- A camada semântica **mantém os nomes do shadcn** que o kit atual já consome. Trocar a paleta não exige reescrever componentes.
- A família `--sidebar-*`, que já existe, passa a ser a camada semântica do shell escuro.
- Entram os semânticos que faltam: `success`, `warning`, `info`, cada um com par de texto e fundo suave, e `brand` para o acento vermelho.
- **Tema futuro**: um tema é um bloco que redefine só os semânticos. Como componentes não conhecem primitivos nem valores literais, um tema escuro futuro não toca em componente (WDS-D-003).
- O bloco `.dark` atual tem valores neutros nunca verificados e nenhum código o ativa. Proposta: removê-lo, para não entregar tema não testado.
- Uma guarda automatizada, no estilo de `web/app/server-client-boundary.test.ts`, falha quando tela ou componente usa cor literal ou utilitário de paleta crua.

---

## Cores e estados (proposta)

**Primitivos**

| Grupo | Valores |
| --- | --- |
| Marca | `ink #111111`, `red #D81E28`, `white #FFFFFF`, `steel #A7A9AC` |
| Neutros (derivados do aço) | 50 `#F6F6F7`, 100 `#ECEDEE`, 200 `#DCDDDF`, 300 `#C4C6C8`, 400 `#A7A9AC`, 500 `#808387`, 600 `#64676B`, 700 `#4A4D50`, 800 `#2A2C2E`, 900 `#1C1C1E`, 950 `#111111` |
| Vermelhos | 50 `#FDECEC`, 500 `#D81E28`, 700 `#B3141D` |
| Estado | sucesso `#15803D`, aviso `#B45309` |

**Semânticos do conteúdo claro**

| Token | Valor | Contraste verificado |
| --- | --- | --- |
| `background` | `#FFFFFF` | — |
| `muted` (cabeçalho de tabela, barra de filtros) | `#F6F6F7` | — |
| `foreground` | `#111111` | 18,88 sobre branco |
| `muted-foreground` | `#64676B` | 5,68 sobre branco; 5,26 sobre `muted`; 4,85 sobre `accent` |
| `primary` / `primary-foreground` | `#111111` / `#FFFFFF` | 18,88 |
| `brand` (acento) | `#D81E28` | 5,08 sobre branco |
| `border` (decorativa) | `#DCDDDF` | não precisa de 3:1 |
| `input` (borda de campo) | `#808387` | 3,81 sobre branco |
| `ring` (foco) | `#111111` | 18,88 |
| `destructive` | `#B3141D` | 6,92 sobre branco; 6,06 sobre `#FDECEC` |
| `success` | `#15803D` | 5,02 sobre branco |
| `warning` | `#B45309` | 5,02 sobre branco; 4,57 sobre `#FDF3E3` |
| `info` | `#4A4D50` | 7,26 sobre `#ECEDEE` |

**Semânticos do shell escuro**

| Token | Valor | Contraste verificado |
| --- | --- | --- |
| `sidebar` | `#111111` | — |
| `sidebar-foreground` | `#FFFFFF` | 18,88 |
| texto secundário do shell | `#A7A9AC` | 8,02 |
| `sidebar-accent` (hover) | `#2A2C2E` | branco 14,02; aço 5,95 |
| indicador de item ativo | `#D81E28` | 3,72 sobre `#111111` (gráfico, mínimo 3:1) |
| `sidebar-border` | `#2A2C2E` | — |
| foco no shell | `#FFFFFF` | 18,88 |

**Limites que os números impõem**

- Vermelho de marca sobre preto dá 3,72: serve para indicador e texto grande, não para texto comum no shell.
- Vermelho de marca sobre a superfície de hover do shell (`#2A2C2E`) dá 2,76: o indicador ativo fica sobre o preto base, não sobre a superfície de hover.
- Aço sobre branco dá 2,36: no conteúdo claro o aço é só decorativo.
- Sucesso `#15803D` sobre o fundo suave verde dá 4,47: o texto sobre fundo suave precisa de um verde mais escuro, a fechar no Design.
- Erro e marca são vermelhos próximos. A distinção vem de tom (`#B3141D` contra `#D81E28`), de ícone e de texto, nunca só da cor.

**Estados de interação**

| Estado | Regra |
| --- | --- |
| Hover | Um passo de neutro (botão preto vai a `#2A2C2E`; superfície clara vai a `muted`). |
| Pressionado | Um passo além do hover, sem deslocamento exagerado. |
| Foco visível | Anel de 2 px com afastamento: preto no conteúdo, branco no shell. Nunca vermelho, para não parecer erro. |
| Desabilitado | Fundo neutro 100, texto neutro 600, sem redução de opacidade no rótulo. |
| Inválido | Borda e mensagem em `destructive`, com ícone. |
| Selecionado / ativo | Peso do texto, marca gráfica e, quando navegação, indicador vermelho. |
| Em andamento | Controle desabilitado, rótulo de progresso com "…". |

**Estados de domínio**: neutro, sucesso, aviso, erro e informação, cada um com ícone próprio. Convenção de valores em P8 de `STATE.md`.

---

## Escala tipográfica (proposta)

| Papel | Fonte e peso | Tamanho / altura | Caixa | Onde |
| --- | --- | --- | --- | --- |
| Destaque | Barlow Condensed 700 | 36 / 40 | alta | Login, páginas de erro |
| Título de página | Barlow Condensed 700 | 30 / 36 (24 / 28 no celular) | alta | `h1` |
| Título de seção | Barlow Condensed 600 | 20 / 24 | frase comum | `h2` |
| Rótulo de grupo | Barlow Condensed 600 | 13 / 16, espaçamento 0,06em | alta | Grupos da navegação, sobretítulos |
| Título de bloco | Geist 600 | 16 / 24 | frase comum | `h3`, diálogos, cartões |
| Corpo | Geist 400 | 14 / 20 | frase comum | Texto, tabelas |
| Corpo maior | Geist 400 | 16 / 24 | frase comum | Campos em toque, texto corrido |
| Rótulo | Geist 500 | 14 / 20 | frase comum | Rótulos de campo, botões |
| Legenda | Geist 400 | 12 / 16 | frase comum | Ajuda, metadados |
| Dado | Geist 400/500, algarismos tabulares | 14 / 20 | — | Números e dinheiro |
| Identificador | Geist Mono 400 | 13 / 20 | — | Códigos e ids curtos |

- Pesos carregados: Barlow Condensed 600 e 700; Geist variável; Geist Mono 400.
- Caixa alta só em título de página, rótulo de grupo e sobretítulo. Nunca em corpo, rótulo de campo, botão, célula ou mensagem.
- Tamanho mínimo: 12 px.
- A verificar no Design: se a Barlow Condensed oferece algarismos tabulares. Se não, valores de destaque (saldo) ficam em Geist.

---

## Espaçamento, forma e movimento (proposta)

| Grupo | Valores |
| --- | --- |
| Escala de espaço (base 4) | 4, 8, 12, 16, 20, 24, 32, 40, 48, 64 |
| Ritmo | rótulo a campo 8; entre campos 16; entre grupos 24; entre seções 32 |
| Margem de página | 16 (celular), 24 (tablet), 32 (desktop) |
| Raio | 4 (pequeno), 6 (controles), 8 (diálogos e cartões); pílula só em status |
| Borda | 1 px; 2 px só em foco e indicador ativo |
| Sombra | nenhuma em repouso; dois níveis para menu e diálogo |
| Controles | 32 (compacto), 36 (padrão), 44 (toque) |
| Ícones | 16 e 20 |
| Tabela | linha de 44, cabeçalho de 40 |
| Navegação | 240 de largura |
| Largura de conteúdo | `narrow` 640, `default` 1120, `wide` 1440 |
| Pontos de quebra | 768 e 1024 (os do Tailwind) |
| Movimento | 120, 200 e 300 ms, ease-out; só `transform` e `opacity`; nada decorativo |

---

## Padrões responsivos

| Elemento | Abaixo de 768 | 768 a 1023 | A partir de 1024 |
| --- | --- | --- | --- |
| Navegação | Gaveta aberta por botão | Gaveta aberta por botão (P3) | Lateral fixa de 240 |
| Cabeçalho de página | Título, depois ações em largura total | Título e ações na mesma linha | Igual |
| Tabelas | Lista empilhada; tabela pequena continua tabela | Colunas por prioridade | Tabela completa |
| Filtros | Recolhidos em painel, com contagem de ativos | Barra em duas linhas | Barra em uma linha |
| Formulários | Coluna única, ações em largura total | Coluna única | Coluna única, largura `narrow` |
| Diálogos | Tela cheia abaixo de 640, ações fixas embaixo | Centralizado | Centralizado |
| Ações de linha | Uma visível, demais em menu | Igual | Igual |
| Cartões e atalhos | Uma coluna | Duas colunas | Três colunas |

Rolagem horizontal só em tabela declarada como comparativa, com justificativa registrada, primeira coluna fixa e região com nome acessível.

---

## Acessibilidade

Requisito do Design System, verificado em cada spec.

| Tema | Regra |
| --- | --- |
| Contraste | 4,5:1 em texto; 3:1 em texto grande, bordas de campo e elementos gráficos essenciais |
| Foco | Sempre visível, 2 px, 3:1 contra o fundo, no conteúdo e no shell |
| Teclado | Toda ação alcançável; Esc fecha sobreposição e devolve o foco; link "Pular para o conteúdo" |
| Alvos | 44 × 44 em ponteiro de toque; nunca menos de 24 × 24 |
| Rótulos e nomes | Todo campo com rótulo associado; todo botão de ícone com nome; ação de linha nomeia a linha |
| Erros | Junto do campo, ligados por `aria-describedby`, com ícone e texto que diz o próximo passo |
| Hierarquia | Um `h1` por página, níveis sem salto, regiões nomeadas |
| Estados | Cor nunca sozinha: sempre ícone ou texto |
| Movimento | `prefers-reduced-motion` respeitado |
| Leitores de tela | Avisos e toasts em região viva; ícones decorativos ocultos |
| Zoom e refluxo | Utilizável a 200% e a 320 px de largura, sem rolagem horizontal de página |

---

## Agent's Discretion

Ficam a critério do Design, dentro das regras das specs: valores finais dos neutros e dos tons de sucesso e aviso, alturas exatas de linha, ícone de cada status e de cada item de navegação, tempo de exibição dos toasts, nomes dos tokens de componente e organização interna dos arquivos do kit.

## Specific References

- O escudo oficial (`web/public/brand/torcida-jovem-logo.svg`) é a única referência de marca.
- Pares de contraste aprovados em WDS-D-015.
- Nenhuma referência externa foi adotada. Se o espécime deixar dúvida em alguma dimensão, a comparação com referências externas é proposta nesse momento, dizendo qual decisão ajudaria a resolver.

## Deferred Ideas

- Tema escuro (arquitetura pronta, valores e verificação em ciclo próprio).
- Aviso de alterações não salvas ao sair de um formulário.
- Navegação em trilho de ícones no tablet, se a gaveta se mostrar incômoda.
- Componentes de indicador e de gráfico, quando existir endpoint de indicadores.
- Uso de A FERA em login, estados vazios e páginas de erro, quando houver asset oficial.
- Paginação no servidor e virtualização, se os volumes crescerem.
