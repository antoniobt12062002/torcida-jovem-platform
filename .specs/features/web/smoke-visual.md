# Conferência visual da Web V1 no navegador

- **Data**: 2026-10-09
- **Código exercitado**: `feature/web-v1`, build de produção (`next build` + `next start`), nos commits `b92f9ea` (antes da correção abaixo) e `5514b38` (depois)
- **Stack**: PostgreSQL 16 (compose), Garage `dxflrs/garage:v2.4.1` avulso, API Go, front com `API_URL=http://localhost:8080`
- **Navegador**: Microsoft Edge 155 em modo headless, dirigido pelo protocolo de depuração (CDP) por um script temporário fora do repositório. Nenhuma dependência foi adicionada ao projeto.

Esta conferência cobre a pendência registrada em `smoke.md`: cliques, digitação, diálogos, mensagens e redirecionamentos foram exercitados numa página real, e 58 capturas de tela foram inspecionadas. As contas eram fictícias (`*@local.test`), com senhas geradas na hora e apagadas ao fim. Este arquivo não contém senha, token, cookie nem chave.

## Defeito funcional encontrado e corrigido

**`/financeiro/contas` não abria no build de produção** (FWB-01 AC1). A página caía na tela de erro do Next com `TypeError: ... is not iterable`.

- **Causa**: `app/(app)/financeiro/contas/page.tsx` é um Server Component e importava a constante `PERMISSION` de `contas-page.tsx`, um módulo `"use client"`. No servidor, um valor que não é componente vindo de um módulo cliente é uma referência opaca: `PERMISSION.read` chegava `undefined` ao `RequirePermission`.
- **Por que passou pela verificação**: os testes rodam em jsdom, onde essa fronteira servidor/cliente não existe, e o smoke anterior foi feito por HTTP, sem navegador.
- **Correção** (commit `5514b38`): as páginas de contas e de saldo declaram a permissão da rota como texto, como as demais. A de saldo funcionava, mas usava o mesmo padrão frágil.
- **Guarda**: `web/app/server-client-boundary.test.ts` falha se um arquivo de servidor em `app/` importar de um módulo `"use client"` algo que não seja componente. Conferido: passa com a correção e falha com o código anterior, apontando as duas constantes.
- **Reconferido no navegador** depois do novo build: a página abre, com estado vazio e todo o fluxo abaixo.

## Resultado por área

| Área | Conferido | Resultado |
| --- | --- | --- |
| Login | campos vazios, senha errada ("E-mail ou senha inválidos."), sucesso, link "Esqueci minha senha", `/` → `/entrar?next=` | Passou |
| Troca de senha obrigatória | redirecionamento para `/conta/senha`, sem navegação lateral, confirmação diferente, sucesso → `/inicio` | Passou |
| Shell e navegação | cabeçalho, menu lateral por seção, item ativo, menu da pessoa (Trocar senha, Sair), atalhos de `/inicio` | Passou |
| Usuários | lista, filtro por papel, criar (validação, e-mail duplicado no campo), promover com papéis e motivo, menu conforme o estado do usuário, própria linha sem ações, toast | Passou |
| Plano de contas | estado vazio, criar raiz, subconta com tipo fixado pelo pai, confirmação de desativar, conta inativa marcada, árvore | Passou depois da correção |
| Lançamentos | estado vazio, validação, dinheiro inválido, prévia do líquido, criar, filtros, detalhe | Passou |
| Workflow | confirmar recebimento, devolução só com contas de despesa ativas, cancelamento com motivo obrigatório (botão desabilitado com vazio ou só espaços), dados de cancelamento no detalhe | Passou |
| Comprovantes | anexar PDF, lista com autor "você", URL assinada pedida no clique | Passou |
| Saldo | valor depois do recebimento (R$ 198,50) e texto explicativo | Passou |
| Produtos | estado vazio, validação, código duplicado no campo, busca, detalhe com saldo | Passou |
| Movimentações | quantidade decimal recusada, origem fixa "Inventário" sem seletor, entrada, saída acima do saldo ("Saldo insuficiente para esta saída."), devolução a partir do histórico | Passou |
| Ajuste | botão desabilitado sem quantidade ou motivo, prévia do saldo, aviso de saldo negativo sem bloquear, selo "Saldo negativo" | Passou |
| Permissões (`CONSELHO_FISCAL`) | navegação sem Administração; nenhuma ação de escrita em contas, lançamentos, detalhe (só "Baixar"), produtos e detalhe; "Sem acesso" em `/admin/usuarios` | Passou |
| Sessão | logout; cookie removido e clique → `/entrar?next=…&sessao=encerrada` com o aviso; recarga sem sessão → `/entrar?next=…` | Passou |
| Recuperação | pedido com mensagem neutra; `/redefinir-senha` sem token; com token inválido o fragmento some da URL e aparece "O link expirou ou já foi usado." | Passou |
| Responsividade | 390 px: menu recolhível abre e fecha ao navegar, sem rolagem horizontal da página; 820 px: navegação lateral visível | Passou |
| Estados | carregamento (esqueleto), vazio, erro de validação, erro da API e sucesso (toast) vistos nas telas acima | Passou |

Nenhum problema de segurança foi observado. Não executável localmente, como já registrado: redefinição com token válido.

## Observações de comportamento (sem defeito; para decisão de produto)

- Conta inativa ainda oferece "Nova subconta". A spec não define; a API decide.
- Lançamento cancelado ainda oferece "Anexar comprovante". A spec não proíbe; a API decide.
- Uma exceção de renderização mostra a tela de erro padrão do Next, em inglês ("This page couldn't load"). Não há `app/error.tsx` em português.

As pendências de aparência estão em `design-pendencias.md`.
