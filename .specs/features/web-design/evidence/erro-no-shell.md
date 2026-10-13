# Evidência — página de erro dentro do shell numa exceção real (DSS-04)

- **Data**: 2026-10-13
- **Task**: T2 de `tasks/04-shell-e-acesso.md` (unidade DS-S), item pendente "uma exceção forçada mostra a página de erro dentro do shell"
- **Requisitos**: DSS-04 AC1, AC4 e AC5 (e AC2 por contraste: o `global-error` não é acionado)
- **Autorização**: WDS-D-035 (fechamento da T2 em ambiente controlado)
- **Código conferido**: `023558b` (branch `ds/v3-fix`), Next 16.4.0, React 19.3.0
- **Status**: **comprovado** para erro lançado no cliente e no servidor

## Ambiente controlado

Nenhum código de erro entrou na produção. A simulação rodou num worktree descartável (`git worktree add <scratchpad>/err-sim HEAD --detach`), fora do repositório, com duas rotas temporárias:

| Rota temporária | Tipo | Como falha |
| --- | --- | --- |
| `app/(app)/zz-falha/cliente/page.tsx` | Client Component (`"use client"`) | lança `Error` em toda renderização e soma 1 a `window.__zzRenders` antes de lançar |
| `app/(app)/zz-falha/servidor/page.tsx` | Server Component (`force-dynamic`) | registra uma linha no log do servidor e lança `Error` em toda renderização |

As mensagens lançadas continham o texto "SIMULACAO zz-falha ... detalhe tecnico que nao pode aparecer", para conferir que nada técnico aparece na tela.

Build de produção (`pnpm build`) e `pnpm start -p 3140` no worktree, com `API_URL=http://localhost:8080` e a API local. Edge headless por CDP, script Node fora do repositório. Sessão do presidente (conta fictícia de desenvolvimento), aberta pelo formulário de login; a senha foi lida do arquivo local de contas e não foi impressa nem gravada. O cookie de sessão de `localhost` vale para as duas portas.

Ao final, o servidor da porta 3140 e o Edge foram encerrados e o worktree foi destruído (`robocopy /MIR` de pasta vazia, por causa de caminhos longos do `node_modules`, e `git worktree prune`). `git worktree list` não mostra mais `err-sim`; `git status --porcelain` do worktree de correção só mostra as mudanças desta rodada. O front da porta 3000 não foi tocado.

## Resultado, ponto a ponto (1440 × 900)

| Verificação | Cliente (`/zz-falha/cliente`) | Servidor (`/zz-falha/servidor`) |
| --- | --- | --- |
| Boundary acionado é `app/(app)/error.tsx` | sim: texto "Não foi possível mostrar esta página…", ações "Tentar de novo" e "Voltar ao início" | sim, igual |
| Não é o `global-error` | sim: sem "Não foi possível abrir o sistema" e sem `[data-slot=access-layout]` | sim |
| Não é a tela padrão do Next | sim: sem "Application error", "client-side exception", "This page couldn't load" nem "Internal Server Error" | sim |
| Shell presente | `[data-slot=shell]`, cabeçalho com o link da marca para `/inicio`, `nav` "Navegação principal", nome da pessoa no menu | sim, igual |
| Uma região principal | 1 `main` | 1 `main` |
| Texto em português, um `h1` | `h1` único "Algo deu errado"; título da aba "Algo deu errado \| TJ Platform" | sim, igual |
| Foco no `h1` ao aparecer | `document.activeElement` é o `h1` | sim |
| Nenhuma mensagem técnica, pilha ou `digest` na tela | `innerText` sem "SIMULACAO", "detalhe tecnico", "digest", "Error:", "stack trace" nem "zz-falha" | sim (o servidor registra `digest: '2110239581'` só no log dele, como o Next faz em produção) |
| Em repouso, 5 s depois de aparecer | 0 requisições; renderizações da rota paradas em 6; `TaskDuration` do navegador +0 ms | 0 requisições; `TaskDuration` +0 ms |
| "Tentar de novo" chama a nova tentativa | sim: as renderizações da rota passam de 6 para 12 e sai `GET /zz-falha/cliente?_rsc=…` | sim: sai `GET /zz-falha/servidor?_rsc=…` e o servidor registra a segunda renderização (2 no total: abertura e nova tentativa) |
| Falha persistente: a página de erro reaparece | sim, com `h1` focado de novo | sim |
| Sem loop, 5 s depois da nova tentativa | 0 requisições; renderizações paradas em 12; `TaskDuration` +61 ms (o clique e a nova renderização), `ScriptDuration` +47 ms | 0 requisições; `TaskDuration` +51 ms, `ScriptDuration` +34 ms |
| CPU dos processos, 5 s parado na página de erro depois da nova tentativa | servidor Next: 0 ms; Edge headless (9 processos): 0 ms (`TotalProcessorTime` do Windows) | — (medido só no cliente) |
| "Voltar ao início" | leva a `/inicio`, `h1` "Início", shell presente, sem página de erro | sim, igual |
| Avisos de hidratação no console | nenhum | nenhum |

A nova tentativa também dispara o `prefetch` dos links da navegação (`/inicio`, `/financeiro/lancamentos`, `/financeiro/contas`, `/financeiro/saldo`, `/estoque/produtos`, `/admin/usuarios`, uma vez cada), comportamento normal do Next ao renovar a árvore; não se repetem.

As 6 renderizações por tentativa no cliente vêm das novas tentativas internas do React e do Next antes de entregar o erro ao boundary (a conta inclui a hidratação da página pré-renderizada). O que importa para o loop é que o número é o mesmo em cada tentativa e não cresce sem ação.

### Console

- Cliente: 2 entradas `console.error` com a mensagem e a pilha da exceção simulada. É o registro que o React faz de toda exceção capturada por um boundary; fica só no console do navegador e não aparece na página.
- Servidor: 2 entradas `console.error` "Minified React error #441" (erro de Server Component com a mensagem omitida em produção). Nenhuma mensagem técnica chega à página.
- Nenhum erro ou aviso de hidratação (`hydrat`, `#418`, `#423`, `#425`, "did not match") nos dois casos.

## Capturas (`erro-no-shell/`)

| Arquivo | Conteúdo |
| --- | --- |
| `cliente-1440-erro.png` | erro do cliente dentro do shell, 1440 px |
| `cliente-1440-apos-tentar.png` | depois de "Tentar de novo": a página de erro reaparece |
| `cliente-390-erro.png` | o mesmo em 390 px (gaveta fechada, ações em largura total) |
| `cliente-1440-voltar-inicio.png` | "Voltar ao início" abre `/inicio` |
| `servidor-1440-erro.png`, `servidor-1440-apos-tentar.png`, `servidor-390-erro.png`, `servidor-1440-voltar-inicio.png` | os mesmos quatro passos para o erro lançado no servidor |

## Limitações

- A medida de CPU por `TotalProcessorTime` tem a resolução do Windows (cerca de 15 ms); o número relevante é que não houve trabalho contínuo, confirmado também pelo `TaskDuration` do navegador e pela contagem de requisições e de renderizações.
- A exceção no layout raiz (DSS-04 AC2, `global-error.tsx`) não foi simulada nesta rodada; segue coberta pelos testes de `app/error-pages.test.tsx`.
