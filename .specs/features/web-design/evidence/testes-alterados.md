# Evidência — testes existentes alterados no ciclo (web-design)

- **Data**: 2026-10-12
- **Unidade**: DS-INT (T11 de `tasks/05-migracao-web-v1.md`), requisito DSM-01 AC4
- **Base do ciclo**: `7c30c6a` (base documental, 69 arquivos de teste / 851 testes)
- **Ponto conferido**: `445bc68` (rodadas 1 a 6 integradas) mais as alterações de T11 neste commit
- **Fontes**: `git diff 7c30c6a..HEAD` restrito aos 69 arquivos de teste que existiam em `7c30c6a`; mensagens de commit; registro de execução em `../STATE.md`

## Nenhum teste removido

Lista de arquivos `web/**/*.test.ts(x)` em `7c30c6a` comparada com a de HEAD (`git ls-tree -r --name-only <rev> -- web`):

- `7c30c6a`: 69 arquivos; HEAD: 108 arquivos (107 em `445bc68` mais `web/test/design/global.guard.test.ts`).
- `comm -23 base head`: vazio. Os 69 arquivos da base continuam presentes; nenhum foi renomeado nem apagado.

Os 39 arquivos novos (guardas, kit, compostos, espécime, shell e telas) não entram na tabela abaixo: não existiam na base.

## Critério

Uma linha removida ou trocada num teste da base conta como alteração. Acréscimos puros (testes novos num arquivo existente) e troca de `import` não alteram asserção existente e ficam agrupados no fim. Para cada alteração: o antes, o depois e a decisão ou o requisito que a motiva. Nenhuma alteração enfraquece o comportamento coberto: quando a asserção antiga saiu, ou a forma nova a substitui no mesmo arquivo, ou o comportamento é coberto em outro teste citado.

## Alterações de asserção, por arquivo

| Arquivo (`web/`) | Commit(s) | Antes | Depois | Motivo |
| --- | --- | --- | --- | --- |
| `components/app/app-shell.test.tsx` | `e9aeeb5` | Clicar em "Trocar senha" no menu da pessoa chamava `router.push("/conta/senha")` (conferido pelo mock de navegação) | "Trocar senha" é um link (`A`, `href="/conta/senha"`) e "Sair" é um botão | DSS-01 AC7: item que navega é link, não `onClick` |
| `components/app/nav-items.test.ts` | `0f4e4d5`, `fc73eac` | "Início" igual a `{ label, href }`; demais itens comparados por igualdade estrita | "Início" com `icon` e `description`; demais itens projetados em rótulo, caminho e permissão (mesmos valores) | DSS-01 AC3 e DSS-06 AC2: itens ganham ícone e descrição; caminhos e permissões inalterados |
| `features/financeiro/nav.test.ts` | `fc73eac` | Itens com `label`, `href`, `permission` | Mesmos valores com `icon` e `description` | DSS-01 AC3 (ícones e descrições da navegação) |
| `features/estoque/nav.test.ts` | `fc73eac` | Idem | Idem | DSS-01 AC3 |
| `features/identity/nav.test.ts` | `fc73eac` | Idem | Idem | DSS-01 AC3 |
| `features/estoque/movimentacoes/movimentacoes-panel.test.tsx` | `647fbab` | Botão de linha com nome exato "Devolver" (8 asserções) | Nome que começa por "Devolver " e identifica a movimentação (`/^Devolver /`) | DSP-05 AC10: ação de linha com nome acessível que identifica o registro |
| `features/financeiro/comprovantes/comprovantes-panel.test.tsx` | `bad9c25`, `32184e8` | Linhas por `data-testid="comprovante-row"`; mensagem "limite permitido de 9,9 MB" | Linhas da tabela nomeada "Comprovantes" (DataList); mensagem "limite permitido de 9 MB" (o limite em bytes não mudou) | DSP-05 (listagem pelo componente do kit); DSP-07 AC6 (megabytes inteiros, para baixo) |
| `features/financeiro/lancamentos/lancamentos-page.test.tsx` | `46a8163`, `7ebae02` | Linhas por `data-testid="lancamento-row"`; link "Detalhes" | Linhas da tabela nomeada "Lançamentos"; link "Detalhes lançamento Receita {id curto}" (mesmo `href`) | DSP-05 (DataList) e DSP-05 AC10 (nome que identifica a linha) |
| `features/financeiro/lancamentos/lancamento-detail.test.tsx` | `8c6a245` | Valor líquido "R$ 97,50" | "+R$ 97,50" | DSC-06 e WDS-D-024: receita com sinal "+" |
| `features/financeiro/lancamentos/workflow.test.tsx` | `8c6a245` | Botão final da confirmação "Confirmar recebimento" / "Confirmar pagamento" | "Receber receita" / "Pagar despesa"; requisições e erros conferidos iguais | DSP-03 AC2 e AC3: confirmação com o verbo da ação e o objeto |
| `features/financeiro/saldo/saldo-page.test.tsx` | `8ff8835` | Saldo negativo "-R$ 987,65" (hífen) | "−R$ 987,65" (sinal de menos U+2212) | DSC-06 AC2 (MoneyValue) |
| `features/identity/usuarios/users-page.test.tsx` | `8770ba4` | Botão "Carregar mais" | "Mostrar mais" (mesmas requisições com `cursor`) | DSP-05 AC12 (componente `ShowMore` do kit) |
| `features/identity/usuarios/users-page.permissions.test.tsx` | `8770ba4` | Itens de menu por nome exato ("Editar papéis", "Retirar acesso administrativo") | Nome do item com o usuário (`menuItemName(acao, nome)`); mesma presença e ausência por permissão | DSP-05 AC10 e DSM-02 (menu "Ações de {nome}") |
| `components/ui/display.test.tsx` | T11 (DS-INT) | `Alert variant="destructive"` com papel `alert` | `Alert variant="error"`, mesma asserção de papel e texto | DSM-05: alias `destructive` removido |
| `components/app/status-badge.test.tsx` | T11 (DS-INT) | `StatusBadge` com status em texto (`CRIADA`, `RECEBIDA`, `PAGA`, `CANCELADA`, `ativo`, `inativo`) e `activeStatus()` | Ativo e inativo por `activeStatusDef()`; recusa de tipo da forma em texto (`@ts-expect-error`) | DSM-05 e DSC-05 AC4: a forma antiga saiu. Rótulos, ícones e tons de lançamento (com DD-3, `CANCELADA` neutra) seguem cobertos em `features/financeiro/lancamentos/lancamentos-page.test.tsx` ("status") |

## Testes de alias removidos ou convertidos em T11

Testes criados no próprio ciclo (rodadas 2 e 4, arquivos novos ou acréscimos) para garantir a compatibilidade temporária. Com a remoção dos alias, deixam de ter objeto; cada um virou uma verificação de que a forma antiga não existe mais (no tipo, por `@ts-expect-error` conferido pelo `tsc`, e, quando cabe, nas classes geradas).

| Arquivo (`web/`) | Antes | Depois | Testes (antes → depois) |
| --- | --- | --- | --- |
| `components/ui/actions.test.tsx` | 5 casos: `default`/`outline`/`size="default"` iguais aos nomes novos; `icon` e `icon-sm` quadrados | 1 caso: os cinco nomes antigos recusados pelo tipo e sem as classes antigas | 44 → 40 |
| `components/ui/display.test.tsx` | Badge: 3 casos de alias; Alert: `destructive` igual a `error`, e sem variante (alias `default`) neutro com papel `alert` | Badge: 1 caso de recusa; Alert: 1 caso (tom obrigatório e alias recusados) | 48 → 45 |
| `components/ui/data.test.tsx` | `Table` sem propriedades com rolagem embutida (modo legado) | Sem propriedades, sem contêiner de rolagem; modo `legacy` recusado pelo tipo | 33 → 34 |
| `components/app/status-badge.test.tsx` | 8 casos da forma em texto (mapa de DD-3, ícones e rótulos únicos) mais os 7 da base acima | 2 casos (ativo/inativo e recusa da forma em texto); os 8 casos de definição continuam | 23 → 11 |

Acréscimos de T11: `components/app/filter-bar.test.tsx` (23 → 25: sem filtros além da busca, sem botão "Filtros"; "Limpar filtros" junto da busca preenchida) e `test/design/global.guard.test.ts` (novo, 12 testes). Total do front: 1761 → 1757 testes, 107 → 108 arquivos.

## Somente acréscimos ou troca de import (sem asserção alterada)

`app/(app)/conta/senha/page.test.tsx`, `app/(app)/estoque/produtos/page.test.tsx`, `app/(app)/financeiro/contas/page.test.tsx`, `app/(app)/financeiro/saldo/page.test.tsx`, `app/(app)/layout.test.tsx`, `app/(app)/sem-acesso/page.test.tsx`, `components/app/confirm-dialog.test.tsx`, `components/app/format.test.ts`, `components/app/require-permission.test.tsx`, `components/app/states.test.tsx`, `components/ui/field.test.tsx`, `components/ui/overlays.test.tsx`, `features/estoque/ajustes/ajuste-dialog.test.tsx`, `features/estoque/produto-detail.test.tsx`, `features/estoque/produtos/produtos-page.test.tsx`, `features/financeiro/contas/contas-page.test.tsx`. Em todos, as linhas removidas são só `import` reagrupados ou comentários; os testes da base continuam com as mesmas asserções, e os novos cobrem os ACs de DSP, DSM e DSS de cada tela.

## Alterações das correções da verificação (iterações 2 e 3)

Acrescentado em 2026-10-13 (correção da iteração 3, WDS-D-035), sobre o ponto `023558b`. Mesmo critério das seções acima.

### `fb62b99`: espaço não separável entre número e unidade (DSP-07 AC4, achado A2 da iteração 2)

Três arquivos da base tiveram asserções trocadas. Em todas, só o espaço entre o número e a unidade mudou, de espaço comum (U+0020) para espaço não separável (U+00A0), porque a tela passou a usar U+00A0 (`produto-detail.tsx:107`, `ajuste-dialog.tsx:120,127`, `movimentacao-dialog.tsx:57`). O número, a unidade e o restante do texto são os mesmos; a asserção ficou mais estrita (exige o caractere certo), não mais fraca.

| Arquivo (`web/`) | Antes | Depois | Asserções |
| --- | --- | --- | --- |
| `features/estoque/produto-detail.test.tsx` | saldo `"12 UN"`, `"-3 UN"`, `"0 UN"`, `"7 UN"` (linhas 50, 58, 66, 120, 148) com espaço comum | os mesmos textos com U+00A0 | 5 |
| `features/estoque/ajustes/ajuste-dialog.test.tsx` | saldo e prévia do ajuste (saldo atual e resultante: `"0 UN"`, `"-3 UN"`, `"5 UN"`, `"7 UN"`, `"3 UN"`, `"1 UN"`; linhas 55, 59-60, 65, 83, 86-87, 126, 170, 202) com espaço comum | os mesmos textos com U+00A0 | 10 |
| `features/estoque/movimentacoes/movimentacoes-panel.test.tsx` | saldo depois de cada movimentação (`saldoText("0 UN")` e afins, linhas 131, 144, 153, 162, 194, 201, 209, 222, 253 e 471) e nome acessível de "Devolver" (`"Devolver entrada de 10 UN em …"`, `"Devolver saída de 4 UN em …"`, linhas 423-424 e 451) com espaço comum | os mesmos textos com U+00A0 | 13 |

Antes desta rodada, `produto-detail.test.tsx` e `ajuste-dialog.test.tsx` estavam na lista "Somente acréscimos" abaixo; a partir de `fb62b99` têm estas trocas. Nenhuma outra asserção desses arquivos mudou.

**Por que `movimentacoes-panel.test.tsx:461` ficou com espaço comum**: a linha é `within(dialog).getByText("Devolução da saída de 4 UN em 04/10/2026 09:00.")`. O `getByText` da Testing Library aplica o normalizador padrão ao texto do DOM (`trim()` e troca de `\s+` por um espaço comum). Em JavaScript, `\s` inclui U+00A0, então o U+00A0 da tela vira espaço comum antes da comparação, e a cadeia esperada precisa ter espaço comum. Com U+00A0 na cadeia esperada a busca falharia mesmo com a tela correta. O caractere da tela nessa frase é conferido pela guarda `test/design/global.guard.test.ts:267` (nenhum `} ${unidade}` com espaço comum no código) e o do nome de "Devolver", pelas linhas 423-424 e 451, que comparam `aria-label` e nome acessível sem normalizador.

### `13f5ad7`, `9e95fb9`, `131a5c4` e `023558b`: correção da iteração 3 (D1 a D4)

Nenhuma asserção existente foi trocada ou removida; só acréscimos.

| Arquivo (`web/`) | Commit | Acréscimo | Testes |
| --- | --- | --- | --- |
| `features/estoque/produtos/produtos-page.test.tsx` | `13f5ad7` | o link do nome tem o anel de foco do kit, nunca vermelho (D1, mutante M28) | +1 |
| `features/financeiro/lancamentos/lancamento-detail.test.tsx` | `13f5ad7` | o link da devolução tem o anel de foco do kit, nunca vermelho (D1, mutante M29) | +1 |
| `features/financeiro/comprovantes/comprovantes-panel.test.tsx` | `9e95fb9` | o campo de arquivo declara `name="comprovante"`, ligado ao rótulo, sem `autocomplete` (D3) | +1 |
| `test/design/global.guard.test.ts` | `131a5c4` | todo `Input` do front declara `name` e `autoComplete` (o de arquivo só `name`) e todo `MoneyField` declara `name`; `{...field}` e `{...register(...)}` contam como `name` (D2, D3) | +1 |
| `app/(dev)/design-system/page.test.tsx` | `131a5c4` | todo campo do espécime tem `name` sem repetição (fora dos grupos de rádio e dos espelhos ocultos do Base UI) e os de texto têm `autocomplete` (D2) | +1 |
| `app/(dev)/design-system/subpages.test.tsx` | `131a5c4` | o login de exemplo usa `name` `email`/`senha` e `autocomplete` `username`/`current-password` (D2) | +1 |
| `components/app/shell-frame.test.tsx` | `023558b` | o link da marca tem `hover:bg-sidebar-accent`, a transição de cor e o anel do shell (D4) | +1 |

Total do front: 1787 → 1794 testes, 110 arquivos (nenhum arquivo novo).

## Desvios de teste já registrados

- Os 9 testes de `features/` e `lib/session` que dependem do toast de erro anunciado como `polite` não mudaram (desvio aceito na rodada 2, `../STATE.md`); T11 não os altera.
- Os 33 consumidores do mock de navegação passaram sem alteração na rodada 2 (DS-T).
