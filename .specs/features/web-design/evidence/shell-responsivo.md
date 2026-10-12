# Evidência — navegação responsiva do shell (DSS-02, FND-05 AC2 revisado)

**Task**: T5 de `tasks/04-shell-e-acesso.md` (unidade DS-S, rodada 6).
**Requisitos**: DSS-02 AC1 a AC7, FND-05 AC2 (revisado por WDS-D-019: ponto de quebra de 768 para 1024 px), DSS-01 AC4 e AC5.
**Status**: **PENDENTE** — o roteiro com o shell real exige sessão autenticada na stack local, que a frente DS-S não usa. A conferência é do coordenador na integração.

## Sobre a verificação anterior

A verificação de FND-05 AC2 feita na Web V1, com a navegação fixa a partir de 768 px, **não foi reaproveitada** (DSS-02 AC7). O ponto de quebra mudou para 1024 px (WDS-D-019), e o critério precisa ser reexecutado num navegador real, em build de produção, com a stack real.

## O que falta e por quê

- O shell (`ShellFrame` via `AppShell`) só é renderizado pelo portão de `app/(app)/layout.tsx` com sessão autenticada. Sem sessão, toda rota autenticada leva a `/entrar?next=…` (conferido em `next start`: `/inicio` → `/entrar?next=%2Finicio`).
- O espécime (`/design-system`) mostra o `ShellFrame` sem sessão, mas responde 404 no build de produção (conferido), e a task exige build de produção.
- A frente DS-S recebeu a instrução de não usar credenciais nem dados da API local. Por isso nenhum dos itens abaixo foi executado.

## Roteiro a executar (coordenador)

Build de produção (`pnpm build && pnpm exec next start`), API local, conta fictícia de desenvolvimento conforme `docs/development/web.md`, script temporário fora do repositório. Nenhuma senha, token, cookie ou dado pessoal neste arquivo.

| Largura | Verificação | Resultado |
| --- | --- | --- |
| 390, 768, 820 | Botão "Abrir menu" presente, alvo de pelo menos 44 por 44 px | pendente |
| 390, 768, 820 | Gaveta abre; fecha ao escolher um item e com Esc; o foco volta ao botão de menu | pendente |
| 390, 768, 820 | Nenhuma ação inacessível; cabeçalho visível durante a rolagem | pendente |
| 1024, 1440 | Navegação fixa ao lado do conteúdo, sem botão de menu | pendente |
| todas | Primeiro Tab foca "Pular para o conteúdo", visível; Enter leva o foco a `main#conteudo` | pendente |
| todas | Item ativo com `aria-current="page"`, peso maior e barra vermelha | pendente |
| 320 a 1440 | Nenhuma rolagem horizontal de página (`scrollWidth` igual a `clientWidth`) em `/inicio` e numa listagem | pendente |
| giro | Gaveta aberta abaixo de 1024 px; janela passa a 1024 px: gaveta fecha e navegação fixa aparece | pendente |

Se algum item falhar, a unidade para, a falha é registrada aqui e reportada (critério da T5).

## O que a frente DS-S conferiu sem sessão (complementar, não substitui o roteiro)

Build de produção (`next start`, porta 3110), Playwright:

- `/entrar`, `/recuperar-acesso` e `/nao-existe` em 390, 820 e 1440 px: layout de acesso com escudo e "Torcida Jovem"; uma coluna abaixo de 1024 px e duas a partir daí; igual ao espécime aprovado (`especime/1440-11-acesso.png`, `especime/390-02-acesso.png`).
- `/entrar` e `/nao-existe` em 320 px: sem rolagem horizontal (`scrollWidth` 320 = `clientWidth` 320).
- `/nao-existe`: status 404, título "Página não encontrada | TJ Platform", foco no `h1` ao abrir.
- `/recuperar-acesso`: enviar vazio leva o foco ao campo de e-mail, com a mensagem "Informe o e-mail.".
