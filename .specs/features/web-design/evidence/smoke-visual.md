# Evidência — smoke completo no navegador (T13, DS-INT, rodada 7)

- **Data**: 2026-10-13
- **Commit exercitado**: `2cc62da` (`feature/web-design`), depois da remoção dos alias de compatibilidade (`663febe`). Build de produção servido por `next start` em `http://localhost:3000`.
- **Ambiente**: API Go local (`:8080`, `healthz` = `ok`), PostgreSQL 16 e S3 local (Garage) em contêineres. Banco com os dados do smoke da rodada 6 (3 usuários, 4 contas, lançamentos, 2 produtos), acrescidos dos registros criados neste smoke.
- **Contas**: fictícias (`presidente@local.test`, `conselho@local.test`; `operador@local.test` só como alvo de desativação). Senhas lidas de arquivo fora do repositório e usadas só em memória. Este arquivo e as capturas não contêm senha, token, cookie nem dado pessoal.
- **Requisitos conferidos**: DSM-06 AC3, AC4, AC5 (AC7 alimentado pela lista de defeitos), DSS-02 AC7, INT-03 (refeito para este ciclo).

## Ferramenta

O Playwright MCP não conectou nesta sessão (tempo de conexão esgotado). O smoke foi feito com o **Microsoft Edge 155 (Chromium) instalado, em modo headless (`--headless=new`)**, dirigido pelo protocolo CDP com scripts Node 22 próprios, fora do repositório (scratchpad da sessão), sem instalar nenhum pacote. Recursos usados: `Emulation.setDeviceMetricsOverride` (larguras e `deviceScaleFactor`), `Emulation.setTouchEmulationEnabled` + `Emulation.setEmulatedMedia` (`pointer`/`any-pointer: coarse`, `hover`/`any-hover: none`, `prefers-reduced-motion`), `Input.dispatchKeyEvent` (teclado), `Input.dispatchMouseEvent` (cliques reais de ponteiro nos fluxos funcionais), `DOM.setFileInputFiles` (anexo), `Page.captureScreenshot` (viewport), `Runtime.consoleAPICalled`, `Runtime.exceptionThrown` e `Log.entryAdded` (console). O perfil temporário do Edge foi apagado ao final.

Diferença em relação à rodada 6: lá o toque era emulado por CDP sob um contexto Playwright sem `hasTouch`; aqui a emulação de toque é a do próprio CDP (`navigator.maxTouchPoints = 5`, `matchMedia('(pointer: coarse)')` e `(hover: none)` verdadeiros).

## 1. Fluxos funcionais (API real)

### Acesso e teclado

| Passo | Resultado | Observação |
| --- | --- | --- |
| Login só por teclado (presidente) | Passou | Ver seção 3 |
| `/entrar`, `/recuperar-acesso`, `/redefinir-senha` sem sessão, 320 a 1440 px | Passou | Títulos "Entrar", "Recuperar acesso", "Redefinir senha"; sem rolagem horizontal |
| Página não encontrada (`/nao-existe-xyz`) | Passou | Documento 404, "Página não encontrada" e "Ir para o início", nas seis larguras |
| Primeiro Tab em 1440, 1024, 820 e 390 | Passou | "Pular para o conteúdo" visível (176 × 44, canto superior esquerdo), `:focus-visible`; Enter leva o foco a `main#conteudo`. Fecha a limitação da rodada 6 (lá 1024 e 1440 não foram feitos com teclado) |

### Identity (`/admin/usuarios`)

| Passo | Resultado | Observação |
| --- | --- | --- |
| Filtro por papel pela URL | Passou | `?papel=CONSELHO_FISCAL` abre com o seletor "Papel" em `CONSELHO_FISCAL` e uma linha (Conselheiro Teste) |
| Menu "Ações de Operador Teste" | Passou | `menu` com "Promover a administrativo", "Gerar senha temporária", separador, "Desativar"; nomes acessíveis incluem o usuário |
| Desativar com confirmação | Passou | `alertdialog` "Desativar Operador Teste?" com nome, e-mail e consequência; botão "Desativar"; aviso "Usuário desativado."; linha "Inativo"; foco volta a "Ações de Operador Teste" |
| Reativar com confirmação | Passou | Menu troca "Desativar" por "Reativar"; `alertdialog` "Reativar Operador Teste?"; linha volta a "Ativo"; foco volta ao gatilho |

### Financeiro

| Passo | Resultado | Observação |
| --- | --- | --- |
| Criar lançamento (receita R$ 25,00, Mensalidades, Dinheiro) | Passou | Aviso "Lançamento criado."; linha nova no topo com status "Criada" |
| Receber | Passou | `alertdialog` "Receber esta receita?", botão "Receber receita"; aviso "Recebimento registrado."; status "Recebida" |
| Anexar PDF mínimo (193 bytes) | Passou | Aviso "Comprovante anexado."; linha `t13-comprovante.pdf · PDF · 193 bytes · 1 · você` |
| Baixar | Passou | "Baixar t13-comprovante.pdf" pede `GET /api/v1/financeiro/comprovantes/{id}/url` e abre a URL assinada do S3 local (`localhost:3900/tj-local/documents/{id}?…`; assinatura omitida aqui). Conferido que a URL assinada é pedida e aberta; o conteúdo do arquivo baixado não foi conferido |
| Saldo atualizado | Passou | R$ 67,50 antes, R$ 92,50 depois (+ R$ 25,00) |
| Criar e cancelar com motivo (por teclado) | Passou | Ver seção 3 |

### Estoque (Bandeira grande, saldo inicial 0 UN)

| Passo | Resultado | Observação |
| --- | --- | --- |
| Entrada de 5 | Passou | Aviso "Movimentação registrada."; saldo 5 UN |
| Saída de 999 (acima do saldo) | Passou | Diálogo continua aberto com "Saldo insuficiente para esta saída." em `role=alert`; saldo continua 5 UN (409 da API esperado). Foco se perde, ver defeito N2 |
| Ajuste −2 com motivo | Passou | "Confirmar ajuste" desabilitado sem motivo e habilitado com motivo; aviso "Ajuste registrado."; saldo 3 UN; histórico "Ajuste −2 · Ajuste manual · Contagem fisica T13" |
| Prévia de ajuste negativo (−5) | Passou | "Saldo atual 3 UN", "Saldo resultante −2 UN" e o aviso "O saldo ficará negativo. O ajuste é permitido…" |
| Entrada de 20 por teclado (Camiseta oficial 2026) | Passou | Saldo −3 → 17 UN (conferido também na API, `saldo: 17`) |

### Permissões (Conselho Fiscal)

| Passo | Resultado | Observação |
| --- | --- | --- |
| Login (senha atual do conselheiro) | Passou | Vai para `/inicio` |
| Navegação | Passou | Início, Lançamentos, Contas, Saldo, Produtos; sem "Usuários" |
| Início, contas, lançamentos, saldo, produtos, detalhe do produto | Passou | Nenhum botão em `main`, nenhum campo de arquivo, nenhum formulário |
| Detalhe do lançamento | Passou | Único botão: "Baixar comprovante-teste.pdf" |
| `/admin/usuarios` em 1440, 1024, 820, 768, 390 e 320 | Passou | "Sem acesso" com "Voltar ao início", sem rolagem horizontal, console limpo |

## 2. Rolagem horizontal do documento e tabelas

`document.documentElement.scrollWidth === clientWidth`, sessão do presidente (exceto "Entrar" e "Sem acesso"), build de produção, altura 900. Quando a página rola na vertical, `clientWidth` desconta a barra (375 em 390, 305 em 320, 753 em 768). "Tabela" = contêiner `[data-slot=data-list-table]` sem rolagem interna (`scrollWidth − clientWidth = 0`) e nenhum botão ou link da última coluna além da borda do contêiner.

| Tela | 320 | 390 | 768 | 820 | 1024 | 1440 | Tabela (768 a 1440) |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Entrar (sem sessão) | OK | OK | OK | OK | OK | OK | — |
| Recuperar acesso / Redefinir senha | OK | OK | OK | OK | OK | OK | — |
| Início | OK | OK | OK | OK | OK | OK | — |
| Usuários | OK | OK | OK | OK | OK | OK | inteira (720, 772, 720, 1136 px) |
| Plano de contas | OK | OK | OK | OK | OK | OK | — (árvore) |
| Lançamentos | OK | OK | OK | OK | OK | OK | inteira (720, 772, 720, 1136 px) |
| Detalhe do lançamento | OK | OK | OK | OK | OK | OK | comprovantes inteira |
| Saldo | OK | OK | OK | OK | OK | OK | — |
| Produtos | OK | OK | OK | OK | OK | OK | inteira |
| Detalhe do produto | OK | OK | OK | OK | OK | OK | movimentações inteira |
| Sem acesso (conselheiro em `/admin/usuarios`) | OK | OK | OK | OK | OK | OK | — |
| Página não encontrada | OK | OK | OK | OK | OK | OK | — |

Em 390 e 320 as listagens são empilhadas (nenhum `data-list-table` visível). D1 da rodada 6 (tabela de lançamentos rolando 61 px em 768 e 1024) **não se repete**: a tabela cabe no contêiner em todas as larguras.

## 3. Percurso só por teclado (DSM-06 AC4)

1440 px, sem nenhum evento de mouse. Teclas enviadas por `Input.dispatchKeyEvent`. "Foco visível" = elemento com `:focus-visible` e anel (`box-shadow`) aplicado; foi verdadeiro em **todas** as paradas abaixo.

| Percurso | Sequência resumida | Resultado |
| --- | --- | --- |
| Login | `/entrar`: Tab → "E-mail" (digita) → Tab → "Senha" (digita) → Tab → "Entrar" (44 px de altura) → Enter | `/inicio`, título "Início \| TJ Platform" |
| Criar lançamento | `/financeiro/lancamentos`: Tab ×10 → "Novo lançamento" → Enter (diálogo abre com foco em "Tipo") → Tab → "Conta", ↓ (Mensalidades) → Tab → "Valor bruto", digita 12,34 → Tab ×2 → "Forma de pagamento", ↓ (Pix) → prévia "Valor líquido (prévia): R$ 12,34" → Tab ×2 → "Criar lançamento" → Enter | Aviso "Lançamento criado."; diálogo fecha; foco volta para "Novo lançamento". Esc também fecha e devolve o foco ao gatilho |
| Cancelar com motivo | Na lista: Tab ×18 → link "Detalhes lançamento Receita 1ea87074" → Enter (detalhe; foco no início do documento) → Tab ×3 → "Cancelar lançamento" → Enter (`alertdialog` "Cancelar lançamento?", foco em "Motivo do cancelamento"; confirmar desabilitado) → digita o motivo (confirmar habilita) → Tab ×2 → "Cancelar lançamento" → Enter | Status "Cancelada", motivo exibido. Depois da confirmação o foco vai para `body`, ver defeito N3 |
| Movimentação de estoque | `/estoque/produtos/{Camiseta}`: Tab ×12 → "Registrar entrada" → Enter (foco em "Quantidade") → digita 20 → Tab ×2 → "Registrar" → Enter | Aviso "Movimentação registrada."; diálogo fecha; foco volta a "Registrar entrada"; saldo 17 UN |
| Saída acima do saldo por teclado | Tab ×11 → "Registrar saída" → Enter → digita 999 → Enter | Mensagem "Saldo insuficiente para esta saída." aparece, mas o foco vai para `body` (fora do diálogo), ver defeito N2. Esc fecha e devolve o foco a "Registrar saída" |

Observação: os seletores nativos (`select`) mudam de valor com ↓ sem abrir a lista, como esperado no Windows.

## 4. Ampliação a 200% e 320 px (DSM-06 AC5)

**Forma escolhida**: janela de 1280 × 900 a 200% reproduzida como viewport CSS de **640 × 450 com `deviceScaleFactor` 2**. É o que o zoom do navegador faz: o viewport em pixels CSS cai à metade, as media queries e as container queries reagem (refluxo) e a renderização fica com o dobro de pixels físicos. `document.documentElement.style.zoom = '2'` em 1280 foi descartado porque não altera o viewport CSS: as media queries continuariam no layout de 1280 e o teste não mostraria refluxo.

| Tela | 200% (640 CSS) | 320 px |
| --- | --- | --- |
| Início | OK (625/625) | OK |
| Usuários | OK | OK |
| Plano de contas | OK | OK |
| Lançamentos | OK | OK |
| Detalhe do lançamento | OK | OK |
| Saldo | OK (640/640) | OK |
| Produtos | OK | OK |
| Detalhe do produto | OK | OK |
| Diálogo "Novo lançamento" | 480 px, dentro do viewport, sem rolagem horizontal interna; rola na vertical (ações abaixo da dobra) | 305 px (tela inteira), sem rolagem horizontal |

A 200% (640 CSS), a navegação vira gaveta, as listagens ficam empilhadas e "Novo lançamento"/"Filtros" ocupam a largura toda (capturas `640-*`).

## 5. Toque (390 px, ponteiro grosso)

Elementos interativos visíveis (`a[href]`, `button`, campos, `menuitem` etc.) com caixa menor que 44 × 44:

| Tela | Abaixo de 44 × 44 |
| --- | --- |
| Usuários (lista e com o menu "Ações de …" aberto) | nenhum |
| Lançamentos (lista e com o painel "Filtros" aberto) | nenhum |
| Produtos | nenhum (D2 da rodada 6 corrigido) |
| Detalhe do produto (e diálogo "Registrar entrada") | nenhum |
| Detalhe do lançamento, Plano de contas, gaveta aberta | nenhum |
| Início | os links dos cartões medem 24 px de altura, mas usam `::after` absoluto com `inset: 0` sobre o cartão inteiro (309 px de largura); um toque abaixo do título atinge o link. Não é defeito |

## 6. Movimento reduzido

Medido cerca de 30 ms depois de abrir, com e sem `prefers-reduced-motion: reduce`:

| Sobreposição | Sem preferência | Com `reduce` |
| --- | --- | --- |
| Diálogo "Novo lançamento" (conteúdo e fundo) | `transition`/`animation` 0,3 s, opacidade 0,22, escala 0,96 | 0,00001 s, opacidade 1, sem `transform` |
| `alertdialog` "Cancelar lançamento?" | 0,3 s, opacidade 0,08, escala 0,95 | 0,00001 s, opacidade 1, sem `transform` |
| Gaveta (768 px, conteúdo e fundo) | `transition` 0,3 s, opacidade 0,07, `x = −37` | 0,00001 s, opacidade 1, `x = 0` |

## 7. Console e hidratação

Capturados `console.*`, exceções e entradas do `Log` em cada carregamento: todas as telas do presidente nas seis larguras, as telas públicas nas seis larguras, as telas do conselheiro, e uma volta por navegação do lado do cliente (barra lateral: Lançamentos → Contas → Saldo → Produtos → Usuários → Início), filtros pela URL e o erro de saída acima do saldo.

**Nenhum aviso de hidratação** (busca por `hydrat`, `#418`, `#423`, `#425`, `did not match`), nenhum `warning` e nenhuma exceção. Erros, todos esperados:

- `401` em `GET /api/v1/auth/me` nas telas públicas sem sessão (é como o front descobre que não há sessão).
- `404` do documento em `/nao-existe-xyz`.
- `409` em `POST /api/v1/estoque/movimentacoes` na saída acima do saldo (gera a mensagem do diálogo).

## 8. Capturas

Em `.specs/features/web-design/evidence/smoke-final/`, só viewport (não página inteira). Altura 900, salvo indicação. Só e-mails `@local.test`; nenhum campo de senha preenchido aparece.

| Arquivo | Viewport | Conteúdo |
| --- | --- | --- |
| `1440-01-login.png`, `390-01-login.png` | 1440 × 900, 390 × 844 | Entrar, vazio |
| `1440-03-inicio.png` … `1440-10-produto-detalhe.png` (8) | 1440 × 900 | Início, usuários, contas, lançamentos, detalhe do lançamento, saldo, produtos, detalhe do produto |
| `390-03-inicio.png` … `390-10-produto-detalhe.png` (8) | 390 × 900 | As mesmas telas no celular |
| `1440-15-conselho-sem-acesso.png`, `390-15-conselho-sem-acesso.png` | 1440 × 900, 390 × 900 | Conselho Fiscal em `/admin/usuarios` |
| `1440-16-nao-encontrada.png`, `390-16-nao-encontrada.png` | 1440 × 900, 390 × 900 | Página não encontrada |
| `1024-04-usuarios.png`, `1024-06-lancamentos.png` | 1024 × 900 | Navegação fixa; tabelas inteiras |
| `820-04-usuarios.png`, `820-06-lancamentos.png` | 820 × 900 | Gaveta fechada; tabelas inteiras |
| `320-04-usuarios.png`, `320-06-lancamentos.png`, `320-09-produtos.png` | 320 × 900 | Listas empilhadas em 320 px |
| `640-04-usuarios-zoom200.png`, `640-06-lancamentos-zoom200.png`, `640-09-produtos-zoom200.png` | 640 × 450 CSS, dpr 2 (imagem 1280 × 900) | Ampliação a 200% |
| `640-23-dialogo-novo-lancamento-zoom200.png` | 640 × 450 CSS, dpr 2 | Diálogo a 200% |
| `1440-14-dialogo-cancelamento-movimento-reduzido.png` | 1440 × 900 | `alertdialog` de cancelamento com movimento reduzido |
| `768-11-gaveta-aberta-movimento-reduzido.png` | 768 × 900 | Gaveta aberta com movimento reduzido |
| `1440-17-teclado-novo-lancamento.png` | 1440 × 900 | Diálogo preenchido por teclado, foco em "Forma de pagamento" |
| `1440-18-teclado-cancelamento.png` | 1440 × 900 | `alertdialog` com motivo digitado por teclado |
| `1440-19-usuarios-menu-acoes.png` | 1440 × 900 | Menu "Ações de Operador Teste" |
| `1440-20-lancamento-recebido-com-comprovante.png` | 1440 × 900 | Receita recebida com comprovante anexado |
| `1440-21-saida-acima-do-saldo.png` | 1440 × 900 | Mensagem de saldo insuficiente no diálogo (mostra também N1) |
| `1440-22-ajuste-saldo-negativo.png` | 1440 × 900 | Prévia de ajuste negativo (mostra também N1) |
| `1440-26-confirmacao-desativar.png` | 1440 × 900 | `alertdialog` "Desativar Operador Teste?" (rodapé correto, para comparação com N1) |
| `1024-25-dialogo-entrada.png`, `768-25-dialogo-entrada.png` | 1024 × 900, 768 × 900 | Diálogo "Registrar entrada" (N1) |
| `390-24-usuarios-menu-toque.png` | 390 × 844 | Menu de ações com toque emulado |
| `390-25-dialogo-entrada.png` | 390 × 844 | Diálogo "Registrar entrada" em tela inteira (N1, faixa de 16 px) |

Total: 46 capturas.

## 9. Comparação com a rodada 6 (depois da remoção dos alias)

Comparação pixel a pixel (diferença por canal > 30) entre as capturas de mesmo nome em `evidence/smoke/` e `evidence/smoke-final/`, feita no próprio navegador com `canvas`:

| Captura | Pixels diferentes | Explicação |
| --- | --- | --- |
| `1440-03`, `1440-05`, `1440-08`, `1440-10`, `1024-04`, `820-04`, `390-03`, `390-04`, `390-05`, `390-06`, `390-08` | 0% | Idênticas |
| `1440-04-usuarios` | 0,03% (linhas 440–451) | Faixa de 12 px numa linha da tabela; sem mudança de layout. Provável renderização de texto diferente entre o Chromium do Playwright (rodada 6) e o Edge (esta rodada) |
| `1440-06-lancamentos` | 0,02% (linhas 245–258) | Ícone nativo dos campos de data ("Criado de", "Criado até"): o Edge desenha um ícone de calendário diferente do Chromium. Controle nativo, não é mudança do código |
| `1440-07-lancamento-detalhe` | 0,23% (linhas 608–625) | Campo de arquivo com o texto centralizado na vertical (correção de D4 em `32184e8`); o rótulo nativo muda de "Escolher arquivo" (Chromium) para "Escolher Arquivo" (Edge) |
| `1440-09-produtos`, `390-09-produtos` | 0,46%, 2,8% | Nome do produto com alvo maior (correção de D2) |
| `390-07`, `390-10` | 0,02% | Antisserrilhado de texto |
| `1024-06-lancamentos`, `820-06-lancamentos` | 4,75%, 5,99% | Correção de D1: "Forma de pagamento" saiu nessas larguras e a tabela cabe no contêiner (antes "Detalhes" aparecia cortado e havia barra de rolagem interna) |

Nenhuma diferença atribuível à remoção dos alias. Como as capturas da rodada 6 já tinham sido comparadas ao espécime aprovado (`evidence/especime/`) e as atuais são iguais a elas, exceto pelas correções, a fidelidade ao espécime se mantém: cabeçalho preto com escudo, título em Barlow Condensed com filete vermelho, navegação lateral preta com barra vermelha no item ativo, selos com ícone, listas empilhadas abaixo de 768 px.

Situação dos defeitos da rodada 6: D1, D2 e D4 conferidos como corrigidos; D3 não reproduzido (o detalhe abriu com "Mensalidades de sócios"); **D5 continua** (o aviso "Lançamento criado." é exposto como `role="dialog"` com botão "Fechar aviso").

## 10. Defeitos

Nenhum bloqueia fluxo. Por gravidade.

| ID | Gravidade | Tela / largura | Como reproduzir | Arquivo provável |
| --- | --- | --- | --- | --- |
| N1 | Média (visual) | Diálogos com `DialogFooter`: "Registrar entrada", "Registrar saída", devolução, "Ajustar estoque" e, pelo mesmo componente, "Senha temporária". 768, 820, 1024 e 1440 px; em 390 também | Abrir `/estoque/produtos/{id}` em 1440 e "Registrar entrada": o rodapé cinza fica 24 px acima da borda do diálogo (rodapé termina em 591, diálogo em 615), cobre os últimos 24 px do corpo ("Inventário" aparece cortado; no ajuste, a segunda linha do aviso de saldo negativo) e deixa uma faixa branca embaixo. Em 390 (tela inteira) a faixa é de 16 px. Causa provável: `sticky bottom-0` com `-mb-4`/`md:-mb-6` dentro do contêiner que rola com `p-4`/`md:p-6`; o elemento grudado não entra no padding e sobe o valor da margem negativa. O `alertdialog` ("Desativar", "Cancelar lançamento") usa outro rodapé e está correto. Já existia na rodada 6 (faixa de 16 px visível em `smoke/390-13-dialogo-entrada-toque.png`), mas não foi apontado | `web/components/ui/dialog.tsx` (`DialogFooter`, linha 116) |
| N2 | Média (a11y, teclado) | Diálogos de movimentação de estoque, todas as larguras | Em "Registrar saída", digitar 999 e Enter: a mensagem "Saldo insuficiente para esta saída." aparece, mas o foco vai para `body`, fora do diálogo modal. O campo "Quantidade" fica `disabled` durante o envio e perde o foco; o alerta de erro não recebe o foco (o `FormLayout` faz isso nas telas de acesso). Comportamento herdado da Web V1 (`disabled={isSubmitting}` já existia em `develop`) | `web/features/estoque/movimentacoes/movimentacao-dialog.tsx`; padrão provável também em `create-produto-dialog.tsx`, `conta-dialogs.tsx`, `create-user-dialog.tsx` |
| N3 | Baixa (a11y, teclado) | Detalhe do lançamento, todas as larguras | Cancelar um lançamento pelo teclado: depois da confirmação, o botão "Cancelar lançamento" some (o lançamento fica "Cancelada") e o foco cai em `body`; o próximo Tab recomeça do início do documento. Esperado: foco no título ou no status atualizado | `web/features/financeiro/lancamentos/lancamento-detail.tsx` |
| D5 | Baixa (a11y) | Todas, depois de qualquer ação com aviso | Continua da rodada 6: o aviso é `role="dialog"` não modal na região "Avisos" | `web/components/ui/toast.tsx` |

Observações que não são defeito desta feature:

- Usuário inativo ainda oferece "Promover a administrativo" e "Gerar senha temporária" no menu (regra de negócio, fora do escopo, como as já registradas: conta inativa com "Nova subconta", lançamento cancelado com "Anexar").
- A prévia de saldo do ajuste ("Saldo resultante") não está numa região `aria-live`; a do valor líquido do lançamento está. Vale avaliar na revisão `web-design-guidelines`.
- Depois de navegar da lista para o detalhe com Enter, o foco volta ao início do documento (comportamento padrão de rota do Next.js).
- A 200%, o diálogo "Novo lançamento" rola por dentro e os botões ficam abaixo da dobra (o rodapé não é fixo nesse diálogo). Não há perda de conteúdo.

## 11. Limitações

- **Redefinição de senha com token válido não executável localmente**: o provedor de e-mail local não expõe o link. Conferidos só `/recuperar-acesso` e `/redefinir-senha` sem token (rolagem e console).
- Download: conferido que a URL assinada é pedida (`GET …/comprovantes/{id}/url`) e aberta; o arquivo baixado não foi aberto.
- Navegador headless: Edge 155 sem janela; o zoom de 200% foi reproduzido por viewport CSS de 640 px com `deviceScaleFactor` 2 (seção 4), não pelo controle de zoom da interface.
- Larguras 768 a 1440 conferidas na sessão do presidente; o conselheiro foi medido só em "Sem acesso" (as demais telas dele foram conferidas em 1440 quanto a ações de escrita).
- A revisão `web-design-guidelines` (gate de T13 e T14) não faz parte deste roteiro e fica com o coordenador/verificador.
- Dados criados neste smoke ficam no banco local: receitas de R$ 12,34 (cancelada, motivo "Teste de teclado T13") e R$ 25,00 (recebida, com `t13-comprovante.pdf`), movimentações em "Bandeira grande" (entrada 5, ajuste −2) e "Camiseta oficial 2026" (entrada 20); operador desativado e reativado.


## Correções depois deste smoke (coordenador)

N1 corrigido em `ad67ac5` (rodapé do diálogo colado à borda, medido 0 px em 390, 820 e 1440 px; capturas `fix-N1-*`). N2 e N3 corrigidos em `2903f32` (foco no alerta ou no primeiro campo inválido dentro do diálogo; foco no `h1` depois de receber, pagar e cancelar), conferidos no navegador. D5 permanece como limitação registrada do Base UI.
