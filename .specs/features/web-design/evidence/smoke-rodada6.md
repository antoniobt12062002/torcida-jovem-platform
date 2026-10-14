# Evidência — smoke real no navegador, gate da rodada 6 (web-design)

- **Data**: 2026-10-12
- **Commit exercitado**: `fc73eac` (`feature/web-design`), build de produção servido por `next start` em `http://localhost:3000`
- **Ambiente**: API Go local (`:8080`, `healthz` = `ok`), PostgreSQL 16 e S3 local (Garage) em contêineres; banco recém-criado, só com dados de teste
- **Navegador**: Chromium via Playwright MCP, sem `hasTouch` no contexto (toque emulado por CDP, ver abaixo)
- **Contas**: fictícias (`presidente@local.test`, `conselho@local.test`, `operador@local.test`). Senhas geradas na hora e guardadas fora do repositório. Este arquivo não contém senha, token, cookie nem dado pessoal.
- **Requisitos conferidos**: DSM-02, DSM-03, DSM-04, DSM-06 AC3 e AC5 (parcial, ver "Limitações"), DSS-01, DSS-02 (AC1 a AC7), FND-05 AC2 revisado por WDS-D-019, FND-05 AC8

## 1. Fluxos

### Acesso

| Passo | Resultado | Observação |
| --- | --- | --- |
| Login com senha errada | Passou | "E-mail ou senha inválidos." num `role=alert`, que recebe o foco |
| Login do presidente (troca obrigatória) | Passou | Vai para `/conta/senha`, título "Trocar senha", layout de acesso, sem `nav`; único botão além do formulário é "Sair". Campos `type=password` com `autocomplete` correto |
| Navegar direto para `/financeiro/contas` durante a troca obrigatória | Passou | Volta para `/conta/senha` (conferido com o conselheiro) |
| Troca concluída | Passou | Vai para `/inicio` (mesmo comportamento da Web V1, que ignora `next` depois da troca) |
| `/recuperar-acesso` | Passou | Mensagem neutra: "Se existir uma conta vinculada ao e-mail informado, enviaremos instruções." |
| `/redefinir-senha` sem token | Passou | "O link de recuperação é inválido." e "Pedir um novo link" |
| Endereço inexistente | Passou | Status 404, título "Página não encontrada \| TJ Platform", texto em português e "Ir para o início" |
| Logout pelo menu da pessoa | Passou | Menu: nome, e-mail, "Trocar senha", "Sair" → `/entrar` |
| `/inicio` sem sessão | Passou | `/entrar?next=%2Finicio`; com filtro: `/entrar?next=%2Ffinanceiro%2Flancamentos%3Ftipo%3DRECEITA` |
| Cookie apagado e clique na navegação | Passou | `/entrar?next=%2Ffinanceiro%2Flancamentos&sessao=encerrada` com "Sua sessão foi encerrada. Entre de novo para continuar." |

### Identity (`/admin/usuarios`)

| Passo | Resultado | Observação |
| --- | --- | --- |
| Criar `conselho@local.test` e `operador@local.test` | Passou | Aviso "Usuário criado. A pessoa terá de trocar a senha no primeiro acesso."; linha entra com "Associado" e troca "Pendente" |
| Menu "Ações de Conselheiro Teste" | Passou | `menu` com nome acessível do usuário; itens "Promover a administrativo", "Gerar senha temporária", separador, "Desativar" (nomes acessíveis incluem o usuário) |
| Promover a Conselho Fiscal com motivo | Passou | "Promover" desabilitado sem motivo; aviso "Acesso administrativo concedido…"; linha passa a "Associado, Conselho Fiscal", acesso administrativo "Sim" |
| Filtro por papel | Passou | URL `?papel=CONSELHO_FISCAL`; recarregar mantém o filtro e o resultado |
| Filtro por situação | Passou | `?situacao=inativos` (estado vazio com "Limpar filtros") e `?situacao=ativos`; recarregar mantém |
| Desativar e reativar o operador | Passou | `alertdialog` "Desativar Operador Teste?" com nome e e-mail e botão "Desativar"; reativar com "Reativar Operador Teste?" e botão "Reativar"; foco volta ao gatilho "Ações de Operador Teste" |
| Senha temporária | Não necessária | O conselheiro entrou com a senha inicial |

### Financeiro

| Passo | Resultado | Observação |
| --- | --- | --- |
| Criar contas (Mensalidades/receita, Viagens e Material de bateria/despesa) | Passou | Estado vazio com "Nova conta"; ações por linha com nome da conta |
| Criar subconta "Mensalidades de sócios" | Passou | Diálogo "Nova subconta em Mensalidades"; recuo de 33 px em 1440 |
| Criar lançamentos (receita 150,00 com taxa 2,50; despesa 80,00; receita 30,00; despesa 45,90) | Passou | Prévia "Valor líquido (prévia): R$ 97,00" em `aria-live` |
| Receber | Passou | `alertdialog` "Receber esta receita?", botão "Receber receita"; aviso e status "Recebida" |
| Pagar | Passou | "Pagar esta despesa?", botão "Pagar despesa"; aviso "Pagamento registrado." |
| Anexar PDF (193 bytes) | Passou | Aviso "Comprovante anexado."; lista com nome, tipo, tamanho, versão; limite "até 9 MB" em texto |
| Baixar | Passou | Pede `GET /api/v1/financeiro/comprovantes/{id}/url` e abre a URL assinada do S3 local (conferido só que a URL assinada é pedida) |
| Cancelar com motivo | Passou | "Cancelar lançamento" desabilitado com motivo vazio; status "Cancelada", motivo e autor exibidos |
| Filtros na URL | Passou | `?status=RECEBIDA&tipo=RECEITA&conta=…&de=2026-10-01`; recarregar mantém todos |
| Status com rótulo e ícone distintos | Passou | Criada (circle-dashed), Recebida (arrow-down-to-line), Paga (arrow-up-from-line), Cancelada (ban) |
| Saldo | Passou | R$ 67,50 (147,50 − 80,00); valor em Barlow Condensed, 36 px, `tabular-nums` |
| Desativar conta | Passou | `alertdialog` "Desativar a conta Viagens?"; selo "Inativo" com ícone `circle-slash` |

### Estoque

| Passo | Resultado | Observação |
| --- | --- | --- |
| Criar produtos (CAM-001, BAN-001) | Passou | Estado vazio com "Novo produto" |
| Buscar | Passou | `?busca=camis`; recarregar mantém termo e resultado |
| Entrada de 10 | Passou | Saldo 10 UN |
| Saída de 15 | Passou | "Saldo insuficiente para esta saída." em `role=alert` dentro do diálogo (409 da API esperado no console) |
| Saída de 4 | Passou | Saldo 6 UN |
| Devolução a partir do histórico | Passou | Botões "Devolver entrada de 10 UN em 12/10/2026 20:46" e "Devolver saída de 4 UN em 12/10/2026 20:47"; diálogo "Devolução da saída de 4 UN em …"; saldo 8 UN |
| Ajuste −11 com motivo | Passou | Prévia "Saldo resultante −3 UN" com aviso "O saldo ficará negativo. O ajuste é permitido…"; "Confirmar ajuste" desabilitado sem motivo; selo "Saldo negativo" com ícone `triangle-alert` |

### Permissões (Conselho Fiscal)

| Passo | Resultado | Observação |
| --- | --- | --- |
| Login com troca obrigatória | Passou | Layout de acesso sem navegação |
| Navegação | Passou | Início, Lançamentos, Contas, Saldo, Produtos; sem "Usuários" |
| Contas, lançamentos, saldo, produtos, detalhe do produto | Passou | Nenhum botão de escrita, nenhum campo de arquivo |
| Detalhe do lançamento | Passou | Único botão: "Baixar comprovante-teste.pdf" |
| `/admin/usuarios` | Passou | "Sem acesso" com "Voltar ao início" |

## 2. Rolagem horizontal do documento

`document.documentElement.scrollWidth === clientWidth`, sessão do presidente, build de produção. Quando a página tem rolagem vertical, `clientWidth` desconta a barra de rolagem (por exemplo 375 em 390).

| Tela | 320 | 390 | 768 | 820 | 1024 | 1440 |
| --- | --- | --- | --- | --- | --- | --- |
| Início | OK | OK | OK | OK | OK | OK |
| Usuários | OK | OK | OK | OK | OK | OK |
| Plano de contas | OK | OK | OK | OK | OK | OK |
| Lançamentos | OK | OK | OK* | OK* | OK* | OK |
| Detalhe do lançamento | OK | OK | OK | OK | OK | OK |
| Saldo | OK | OK | OK | OK | OK | OK |
| Produtos | OK | OK | OK | OK | OK | OK |
| Detalhe do produto | OK | OK | OK | OK | OK | OK |
| Entrar (sem sessão) | OK | OK | — | — | — | — |

\* O documento não rola, mas a tabela de lançamentos rola por dentro: 61 px em 768, 9 px em 820, 61 px em 1024 (ver defeito D1). Nenhuma outra tabela rola por dentro em 768, 820, 1024, 1280 ou 1440.

## 3. Navegação responsiva (FND-05 AC2 revisado, DSS-02)

**A verificação anterior de FND-05 AC2, feita na Web V1 com ponto de quebra em 768 px, NÃO foi reaproveitada.** Tudo abaixo foi executado de novo nesta rodada, com o ponto de quebra de 1024 px (WDS-D-019).

| Largura | Verificação | Resultado |
| --- | --- | --- |
| 390, 768, 820 | Botão "Abrir menu" visível, 44 × 44 px, `aria-expanded="false"`; navegação fixa ausente | Passou nas 8 telas |
| 390, 768, 820 | Gaveta abre como diálogo "Menu"; foco vai para "Fechar"; item ativo com `aria-current="page"` | Passou |
| 390, 768, 820 | Esc fecha e o foco volta para "Abrir menu" | Passou |
| 390, 768, 820 | Escolher "Produtos" navega, fecha a gaveta e o foco volta para "Abrir menu" | Passou |
| 390, 768, 820 | Tab fica preso na gaveta (passa pelo guarda de foco do Base UI e volta) | Passou |
| 1024, 1440 | Navegação fixa visível, sem botão de menu | Passou nas 8 telas |
| todas | Primeiro Tab = "Pular para o conteúdo", visível; Enter leva o foco a `main#conteudo` | Passou (390, 768, 820 com teclado; 1024/1440 pela mesma estrutura) |
| todas | Item ativo com `aria-current="page"` (no detalhe, a seção e o item da trilha) | Passou nas 40 combinações |
| giro | Gaveta aberta em 820, janela passa a 1024: gaveta some, navegação fixa aparece, nada fica `inert` nem com `overflow` travado | Passou |

## 4. Toque e movimento reduzido

**Toque emulado** (CDP `setTouchEmulationEnabled` + `pointer: coarse`, `hover: none`), 390 px. Elementos interativos visíveis com caixa menor que 44 × 44:

| Tela | Elementos abaixo de 44 × 44 |
| --- | --- |
| Usuários (lista e menu "Ações de …" aberto) | nenhum |
| Lançamentos (lista e painel "Filtros" aberto) | nenhum |
| Detalhe do lançamento | nenhum |
| Produtos | links "Bandeira grande" (123 × 24) e "Camiseta oficial 2026" (164 × 24) — defeito D2 |
| Detalhe do produto e diálogo "Registrar entrada" | nenhum |

O diálogo de movimentação em 390 ocupa a tela inteira (375 × 844, descontada a barra de rolagem) com "Registrar" e "Cancelar" fixos na base (DSM-04 AC5).

**Movimento reduzido** (`prefers-reduced-motion: reduce`): gaveta, fundo da gaveta, diálogo "Novo lançamento" e `alertdialog` de cancelamento com `transition-duration` e `animation-duration` de 0,00001 s; 30 ms depois de abrir, a gaveta já está em `x=0` e o diálogo com opacidade 1.

## 5. Console

Nenhum aviso de hidratação e nenhum `warning` em todo o percurso. Erros, todos explicados:

- `401` em `GET /api/v1/auth/me` ao abrir telas sem sessão (é a forma de o front saber que não há sessão).
- `401` em `POST /api/v1/auth/login` no teste de senha errada.
- `404` do documento em `/nao-existe-xyz` (esperado).
- `409` em `POST /api/v1/estoque/movimentacoes` no teste de saída acima do saldo (esperado, gera a mensagem).

## 6. Capturas

Em `.specs/features/web-design/evidence/smoke/`, viewport (não página inteira), altura 900 salvo indicação. Sem campo de senha preenchido visível (em `1440-01-login.png` a senha errada aparece mascarada).

| Arquivo | Viewport | Conteúdo |
| --- | --- | --- |
| `1440-01-login.png` | 1440 × 900 | Entrar com o erro "E-mail ou senha inválidos." |
| `1440-02-troca-obrigatoria.png` | 1440 × 900 | Troca obrigatória, layout de acesso, só "Sair" |
| `1440-03-inicio.png` … `1440-10-produto-detalhe.png` | 1440 × 900 | Início, usuários, contas, lançamentos, detalhe do lançamento, saldo, produtos, detalhe do produto |
| `1440-14-dialogo-cancelamento-movimento-reduzido.png` | 1440 × 900 | `alertdialog` de cancelamento com movimento reduzido |
| `1440-15-conselho-sem-acesso.png` | 1440 × 900 | Conselho Fiscal em `/admin/usuarios` |
| `1024-04-usuarios.png`, `1024-06-lancamentos.png` | 1024 × 900 | Navegação fixa; tabela de lançamentos cortada (D1) |
| `820-04-usuarios.png`, `820-06-lancamentos.png` | 820 × 900 | Gaveta fechada; tabela de lançamentos com rolagem interna |
| `768-11-gaveta-aberta.png` | 768 × 900 | Gaveta aberta sobre lançamentos |
| `390-01-login.png` | 390 × 844 | Entrar |
| `390-03-inicio.png` … `390-10-produto-detalhe.png` | 390 × 900 | As 8 telas no celular (listas empilhadas) |
| `390-12-lancamentos-filtros-toque.png` | 390 × 844 | Painel de filtros aberto com toque emulado |
| `390-13-dialogo-entrada-toque.png` | 390 × 844 | Diálogo de entrada em tela inteira com ações na base |

Total: 28 capturas.

## 7. Antes e depois

Não há capturas da Web V1 anteriores à migração versionadas no repositório. `.specs/features/web/smoke-visual.md` registra que 58 capturas foram inspecionadas, mas elas não foram guardadas. A única referência visual versionada é o espécime aprovado (`evidence/especime/`). As capturas desta rodada seguem o espécime: cabeçalho preto com escudo, título em Barlow Condensed com filete vermelho, navegação lateral preta com barra vermelha no item ativo, selos com ícone, listas empilhadas abaixo de 768 px.

## 8. Defeitos encontrados

Nenhum defeito bloqueia fluxo. Ordenados por gravidade.

| ID | Gravidade | Tela / largura | Como reproduzir | Arquivo provável |
| --- | --- | --- | --- | --- |
| D1 | Média | Lançamentos, 768, 820 e 1024 px (sem rolagem interna em 1280 e 1440; larguras entre 1024 e 1280 não medidas) | Abrir `/financeiro/lancamentos` com 4 lançamentos em 1024 px: a tabela rola 61 px por dentro e o botão "Detalhes" da última coluna aparece cortado ("Det"); em 768 também 61 px, em 820, 9 px. As colunas de prioridade 3 (valor bruto, taxa) já saem, mas forma de pagamento (prioridade 2) e a ação continuam. A ação principal da linha fica parcialmente fora da área visível justo no ponto de quebra da navegação fixa | `web/features/financeiro/lancamentos/lancamentos-page.tsx` (prioridades das colunas) e `web/components/app/data-list.tsx` (`LOW_PRIORITY = "hidden lg:table-cell @max-4xl:hidden"`) |
| D2 | Baixa | Produtos, 390 px, ponteiro de toque | Abrir `/estoque/produtos` em 390 com toque emulado: o nome do produto é o único alvo do item e mede 24 px de altura (123 × 24, 164 × 24). Passa no mínimo de 24 px da WCAG 2.5.8, mas não nos 44 px adotados no projeto; nas outras listas a ação é um botão de 44 px | `web/features/estoque/produtos/produtos-page.tsx` (coluna `inList: "title"` com link) ou `web/components/app/data-list.tsx` |
| D3 | Baixa | Detalhe do lançamento, logo depois de abrir pela URL | Abrir direto `/financeiro/lancamentos/{id}` de um lançamento de subconta: por um instante a conta aparece como prefixo do id (`3c246d92`) até o plano de contas carregar; depois mostra "Mensalidades de sócios". Não se repetiu nas aberturas seguintes, com cache | `web/features/financeiro/lancamentos/lancamento-detail.tsx` (uso de `shortId` de `web/components/app/format.ts` como fallback) |
| D4 | Baixa (visual) | Detalhe do lançamento, 1440 px | Em "Anexar comprovante", o campo de arquivo nativo ("Escolher arquivo Nenhum arquivo escolhido") fica com o texto colado no topo da caixa de 36 px, sem centralizar na vertical (ver `1440-07-lancamento-detalhe.png`) | `web/features/financeiro/comprovantes/comprovantes-panel.tsx` / `Input` do kit com `type=file` |
| D5 | Baixa (a11y) | Todas, depois de qualquer ação com aviso | O aviso (toast do Base UI) é exposto como `role="dialog"` não modal com `tabindex=0` dentro da região "Avisos"; leitores de tela anunciam "diálogo" e ele entra na ordem de Tab enquanto visível. Comportamento padrão da biblioteca, registrar ou ajustar | `web/components/ui/` (Toast) |

Observações que não são defeito desta feature:

- Em 390, a lista empilhada de usuários não mostra "Acesso administrativo" nem "Troca de senha" (escolha de prioridade; a informação continua em 768 px ou mais).
- Conta inativa ainda oferece "Nova subconta" e "Renomear"; lançamento cancelado ainda oferece "Anexar comprovante" (já registradas como regra de negócio fora do escopo, `05-migracao-web-v1.md`).
- Saldo de estoque negativo usa hífen ("-3 UN"); valores em dinheiro usam o sinal de menos ("−R$ 80,00").
- "Devolução de" no histórico de estoque e o nome dos lançamentos usam o prefixo do id (comportamento da Web V1).

## 9. Limitações

- O percurso só por teclado de DSM-06 AC4 (login, criação de lançamento, cancelamento, movimentação) não foi feito completo; só a navegação por teclado do shell, da gaveta e dos diálogos.
- A ampliação a 200% (DSM-06 AC5) não foi testada; a largura de 320 px foi, em todas as 8 telas e em "Entrar".
- Toque emulado por CDP (o contexto do Playwright MCP não tem `hasTouch`), então os cliques foram de mouse sob `pointer: coarse`.
- Primeiro Tab conferido com teclado em 390, 768 e 820; em 1024 e 1440 conferida a mesma estrutura (link de pular é o primeiro foco do documento), sem repetir o Tab.


## Correções depois deste smoke (coordenador)

- D1 corrigido em `7ebae02` ("Forma de pagamento" para a prioridade 3) e `c0b1925` (coluna de prioridade 3 só com área de conteúdo de 64rem). Reconferido em `next start` com a sessão real: lançamentos, usuários e produtos em 768, 1024, 1200, 1240, 1280 e 1440 px sem rolagem do documento e com a tabela inteira dentro do contêiner.
- D2, D3 e D4 corrigidos em `32184e8`.
- D5 (toast exposto como `role="dialog"` pelo Base UI) fica registrado para a rodada 7.
