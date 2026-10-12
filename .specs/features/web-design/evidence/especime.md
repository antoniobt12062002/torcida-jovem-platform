# Evidência do espécime visual (DS-E, T7)

- **Data**: 2026-10-12
- **Código**: branch `feature/web-design`, commits `7f72ba6` (T5) e `d529b4f` (T6), sobre as rodadas 1 a 4
- **Navegador**: Chromium via Playwright MCP, `next dev` local (porta 3201), sem API; dados fictícios
- **Produção**: `next build` + `next start` → `GET /design-system` respondeu **404** (`notFound()` na página e em `generateMetadata`); `/entrar` respondeu 200. As subpáginas também respondem 404 (relato da unidade; a página principal foi reconferida pelo coordenador)
- **Erros de console**: só `GET /api/v1/auth/me` 500, do `SessionProvider` do layout raiz, porque a API não estava no ar; a página do espécime não faz chamada própria

## Capturas (`evidence/especime/`)

| Largura | Arquivo | O que mostra |
| --- | --- | --- |
| 1440 | `1440-01-topo.png` | título, índice, escudo sobre claro e sobre o shell |
| 1440 | `1440-02-shell.png` | shell com navegação, ícones, item ativo e menu da pessoa |
| 1440 | `1440-03-status-e-valores.png` | status, badges, MoneyValue e comparação DD-4 |
| 1440 | `1440-04-tabela-e-lista.png` | DataList automático, tabela forçada, lista forçada, tabela pequena |
| 1440 | `1440-05-campos.png` a `1440-09-cores.png` | campos, ações, estados de tela, tipografia, cores com contraste calculado |
| 1440 | `1440-10-pagina-de-exemplo.png` | página de exemplo completa no shell |
| 1440 | `1440-11-acesso.png` | login em duas colunas |
| 1024 | `1024-01-pagina-de-exemplo.png` | navegação fixa no limite de 1024 px |
| 820 | `820-01-pagina-de-exemplo.png` | gaveta, filtros em uma linha, tabela sem a coluna de prioridade 3 |
| 768 | `768-01-pagina-de-exemplo.png` | gaveta e tabela no limite de 768 px |
| 390 | `390-01-pagina-de-exemplo.png` | lista empilhada, busca visível, botão "Filtros", ação primária em largura total |
| 390 | `390-02-acesso.png` | login em coluna única |

## Critérios de aprovação

| # | Critério | Resultado |
| --- | --- | --- |
| 1 | Identidade | Moldura preta, título condensado em caixa alta com filete vermelho, escudo no shell e no login. Reconhecível sem depender de vermelho |
| 2 | Vermelho só como acento | Item ativo, filete do título, erro e saldo negativo. Nenhuma área de fundo vermelha |
| 3 | Tipografia | Barlow Condensed em título de página, rótulos de grupo, nome da marca e valor de destaque; resto em Geist |
| 4 | DD-4 | Barlow Condensed 700 e Geist 600 lado a lado; algarismos tabulares preservados ("R$ 1.111,11" e "R$ 8.888,88" com a mesma largura). **Aprovada: Barlow Condensed 700 (WDS-D-031)** |
| 5 | Contraste e foco | Todos os pares exibidos acima do limite (teste de contraste). Foco de 2 px corrigido na rodada 4 |
| 6 | Status e valores | Cada status com rótulo e ícone; `CANCELADA` neutra e com valor riscado; receita "+", despesa "−", saldo negativo com ícone |
| 7 | Responsividade | Sem rolagem horizontal de página em 390 (medido: `scrollWidth` 390). Lista abaixo de 768, tabela a partir de 768, coluna de prioridade 3 só a partir de 1024, navegação fixa a partir de 1024 |
| 8 | Toque | Conferido na segunda conferência, com emulação |
| 9 | Movimento reduzido | Conferido na segunda conferência, com emulação |
| 10 | Marca | Escudo íntegro (teste de hash), 1:1, sem efeito; nenhum elemento de marca inventado |

## Divergências encontradas

1. **Barra de filtros a partir de 1024 px**: empilhada. Corrigida em `74dd8fe` (ver segunda conferência).
1a. **Overflow em 1024 px** (apontado pelo mantenedor): corrigido em `6414cd9`.
2. **Campos numéricos**: `MoneyField` e `QuantityField` não aceitam somente leitura nem a marca "(opcional)"; o espécime mostra só repouso, inválido e desabilitado.
3. **Escudo em 16 px** (aba): ilegível, já registrado como dependência de asset oficial reduzido.

## Segunda conferência (2026-10-12, depois do portão)

Correções pedidas pelo mantenedor: overflow em 1024 px, barra de filtros e conferência real de toque, teclado e movimento reduzido. Navegador Chromium via Playwright MCP, `next dev`, sem API. O indicador de desenvolvimento do Next foi ocultado nas capturas por estilo injetado só no navegador.

**Viewports das capturas.** As capturas de viewport têm a largura pedida; a área útil do documento é 15 px menor por causa da barra de rolagem vertical (o arquivo de 390 px da primeira conferência era uma captura de página inteira, por isso media 375 px). Novas capturas, todas de viewport:

| Arquivo | Viewport | Área útil do documento |
| --- | --- | --- |
| `390-01-pagina-de-exemplo.png` | 390 × 844 | 375 px |
| `820-01-pagina-de-exemplo.png` | 820 × 1180 | 820 px (sem barra) |
| `1024-01-pagina-de-exemplo.png` | 1024 × 900 | 1009 px |
| `1440-10-pagina-de-exemplo.png` | 1440 × 900 | 1425 px |
| `1440-12-filtros-no-especime.png` | 1440 × 900, recorte da seção | — |

**Rolagem horizontal do documento** (`scrollWidth` igual a `clientWidth`) em `/design-system/pagina-de-exemplo` e `/design-system`: sem rolagem em 390, 768, 820, 1000, 1024, 1040, 1100, 1280 e 1440 px. A unidade de correção também varreu de 768 a 1440 px de 8 em 8 px, sem rolagem. Em 1024 px todas as ações das linhas ficam dentro do contêiner da tabela. Antes da correção, 1024 px dava 1065/1009.

**Barra de filtros**: busca numa linha e dois filtros abaixo quando o contêiner é médio (768 a ~1100 px com navegação); os três numa linha quando há espaço; o mesmo comportamento dentro do espécime e na página de exemplo.

**Teclado**: o primeiro Tab foca "Pular para o conteúdo" (anel branco de 2 px sobre o cabeçalho escuro); Enter leva o foco a `main#conteudo`. Os botões de linha têm nome que identifica a linha ("Detalhes lançamento Mensalidades de março") e anel de foco. O menu de ações abre por Enter e fecha por Esc, devolvendo o foco ao botão "Ações de lançamento …". Em 820 px a gaveta abre por Enter em "Abrir menu" e fecha por Esc, com o foco de volta ao botão.

**Hidratação após redimensionar**: nenhum erro de console além do `GET /api/v1/auth/me` 500 (API fora do ar), durante a varredura de larguras nas duas páginas.

**Movimento reduzido** (`prefers-reduced-motion: reduce` emulado): a consulta casa; duração de transição e de animação de botões e da gaveta caem a 0,00001 s.

**Ponteiro de toque**: emulado por CDP (emulação de toque mais `pointer: coarse` e `any-pointer: coarse`); `matchMedia('(pointer: coarse)')` casou, controles passaram a 44 px (`--control-height-md` = 2,75rem) e campos a 16 px. Em 390 e 820 px, nenhum botão, link, campo ou select ficou abaixo de 44 × 44 px depois da correção do link da marca (que media 40 × 44). **Limitação**: a emulação não reproduz o tamanho físico do dedo nem a área de toque estendida do sistema; medem-se as caixas renderizadas.
