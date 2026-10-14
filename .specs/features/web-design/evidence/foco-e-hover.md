# Evidência — anel de foco dos links (D1) e hover da marca (D4)

- **Data**: 2026-10-13
- **Origem**: `validation.md`, iteração 3, itens D1 e D4; correção autorizada em WDS-D-035
- **Código conferido**: `023558b` (branch `ds/v3-fix`)
- **Ambiente**: `pnpm build` e `pnpm start -p 3141` no worktree de correção, `API_URL=http://localhost:8080`; Edge headless por CDP, 1440 × 900; sessão do presidente (conta fictícia de desenvolvimento), sem credencial gravada. Servidor e navegador encerrados ao final.

## D1 — anel do kit nos links de produto e de devolução (DSC-01 AC2)

Foco levado só por teclado (Tab a partir do início da página), lido com `getComputedStyle` no elemento focado.

| Link | Tabs | `:focus-visible` | `box-shadow` computado | `outline` | Captura |
| --- | --- | --- | --- | --- | --- |
| Nome do produto na tabela de `/estoque/produtos` ("Bandeira grande") | 12 | sim | `rgb(255, 255, 255) 0 0 0 2px, rgb(17, 17, 17) 0 0 0 4px` | `none` | `foco-e-hover/d1-link-produto-foco.png` |
| Receita devolvida no detalhe de uma despesa de devolução (código curto `f655839f`) | 14 | sim | `rgb(255, 255, 255) 0 0 0 2px, rgb(17, 17, 17) 0 0 0 4px` | `none` | `foco-e-hover/d1-link-devolucao-foco.png` |

Leitura: afastamento branco de 2 px (`ring-offset-2`, `--ring-offset` branco) e anel de 2 px em `#111111` (`ring-2`, `--ring`), igual ao do `Button`. Raio de 4 px (`rounded-sm`).

Testes que exigem o anel (`13f5ad7`): `features/estoque/produtos/produtos-page.test.tsx` ("o link do nome tem o anel de foco do kit, nunca vermelho") e `features/financeiro/lancamentos/lancamento-detail.test.tsx` ("tem o anel de foco do kit, nunca vermelho"), no padrão de `components/ui/actions.test.tsx:117`.

Mutantes, aplicados numa cópia do arquivo e revertidos (sem `git stash`): remover `rounded-sm outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-ring-offset` do link.

| Mutante | Arquivo | Resultado |
| --- | --- | --- |
| M28 | `features/estoque/produtos/produtos-page.tsx:140` | **morto**: 1 teste falha (26 passam) |
| M29 | `features/financeiro/lancamentos/lancamento-detail.tsx:162` | **morto**: 1 teste falha (27 passam) |

## D4 — hover no link da marca do shell

`components/app/shell-frame.tsx:157` recebe `transition-colors duration-fast hover:bg-sidebar-accent` (`023558b`), os mesmos tokens e a mesma transição dos itens da navegação (`shell-frame.tsx:254,258`). Escudo, raio, alvo e foco inalterados.

| Estado | Medida (`getComputedStyle`) | Captura |
| --- | --- | --- |
| Repouso (ponteiro fora) | fundo `rgba(0, 0, 0, 0)` sobre o cabeçalho `rgb(17, 17, 17)`; raio 6 px; caixa 156 × 44 px; transição de cor em 0,12 s | `foco-e-hover/d4-marca-repouso.png` |
| Hover (ponteiro no centro do link, `:hover` verdadeiro) | fundo `rgb(42, 44, 46)` = `--sidebar-accent` `#2a2c2e`; texto `rgb(255, 255, 255)` | `foco-e-hover/d4-marca-hover.png` |
| Foco por teclado (2º Tab, depois de "Pular para o conteúdo") | `:focus-visible`; `box-shadow` `rgb(255, 255, 255) 0 0 0 2px` (anel branco do shell, `--sidebar-ring` `#fff`); `outline: none` | `foco-e-hover/d4-marca-foco.png` |

Contraste: branco sobre `sidebar-accent` e o anel branco sobre `sidebar-accent` já estão nos pares calculados de `lib/design/contrast-pairs.ts:37,72` ("Item do shell em hover", "Foco no shell sobre item em hover"), verificados por `test/design/contrast.test.ts`.

Teste: `components/app/shell-frame.test.tsx` ("o link da marca tem hover com o fundo dos itens do shell e mantém o foco do shell") falhou antes da correção e passa depois; exige `hover:bg-sidebar-accent`, `rounded-md`, `transition-colors`, `duration-fast` e o anel do shell, e recusa outro `hover:` e qualquer animação além da transição de cor.

Console: só os dois `401` de `/api/v1/auth/me` da tela de login antes da sessão; nenhum erro nas telas medidas.
