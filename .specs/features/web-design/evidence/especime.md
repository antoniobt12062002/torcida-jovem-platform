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
| 4 | DD-4 | Barlow Condensed 700 e Geist 600 lado a lado; algarismos tabulares preservados ("R$ 1.111,11" e "R$ 8.888,88" com a mesma largura). **Decisão do mantenedor pendente** |
| 5 | Contraste e foco | Todos os pares exibidos acima do limite (teste de contraste). Foco de 2 px corrigido na rodada 4 |
| 6 | Status e valores | Cada status com rótulo e ícone; `CANCELADA` neutra e com valor riscado; receita "+", despesa "−", saldo negativo com ícone |
| 7 | Responsividade | Sem rolagem horizontal de página em 390 (medido: `scrollWidth` 390). Lista abaixo de 768, tabela a partir de 768, coluna de prioridade 3 só a partir de 1024, navegação fixa a partir de 1024 |
| 8 | Toque | Alvos pelos tokens sob `pointer: coarse` (cobertos por teste); não emulado nesta conferência |
| 9 | Movimento reduzido | Regra global coberta por teste; não emulado nesta conferência |
| 10 | Marca | Escudo íntegro (teste de hash), 1:1, sem efeito; nenhum elemento de marca inventado |

## Divergências encontradas

1. **Barra de filtros em 1440 px**: os três filtros ficam empilhados em coluna, um por linha, em vez de uma linha só (em 820 px a barra aparece corretamente em uma linha). Defeito de `components/app/filter-bar.tsx` no ponto de quebra `lg`; correção proposta: manter a grade de três colunas (ou a linha flexível) também a partir de 1024 px. Corrigir antes da rodada 6.
2. **Campos numéricos**: `MoneyField` e `QuantityField` não aceitam somente leitura nem a marca "(opcional)"; o espécime mostra só repouso, inválido e desabilitado.
3. **Escudo em 16 px** (aba): ilegível, já registrado como dependência de asset oficial reduzido.
