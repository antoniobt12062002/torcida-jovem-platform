# Integração e Verificação da Web V1 — Design

**Spec**: `.specs/features/web/spec/06-integracao.md`
**Pré-requisito**: unidades F3, ID-ADM, FIN-a, FIN-b e EST integradas.

## Architecture Overview

```
features/identity/nav.ts ─┐
features/financeiro/nav.ts ┼─→ components/app/nav-items.ts ─→ AppShell (F5) · /inicio
features/estoque/nav.ts ───┘
app/page.tsx ─→ redirect("/inicio")
```

## Components

| Arquivo | Dono | Função |
| --- | --- | --- |
| `components/app/nav-items.ts` | INT | Junta os itens dos módulos, na ordem da spec |
| `app/(app)/layout.tsx` | INT (troca só a lista passada ao shell) | Passa `navItems` ao `AppShell` |
| `app/(app)/inicio/page.tsx` | INT | Atalhos por área visível |
| `app/page.tsx` | INT | Redirecionamento para `/inicio` |
| `docs/development/web.md` | INT | Guia |
| `.specs/features/web/smoke.md` | INT | Checklist e resultados |
| `.specs/features/web/validation.md` | VERIFY | Relatório do verificador |

`app/(app)/layout.tsx` muda de dono ao longo do DAG (F2 → F5 → INT), sempre de forma sequencial e só depois da integração da unidade anterior.

## Critérios de integração (coordenador)

Aplicados ao fim de cada rodada, antes de liberar a próxima:

1. Cada agente entrega: unidade, branch, SHA, arquivos alterados, testes adicionados, gates com código de saída, bloqueios.
2. O coordenador confere o diff de cada agente contra a tabela de propriedade de `STATE.md`. Arquivo fora da propriedade bloqueia a integração daquela unidade.
3. As branches da rodada são integradas, uma a uma, sobre o ponto de integração anterior. Conflito de merge significa erro de propriedade: a integração para e o DAG é revisto, nunca resolvido por escolha manual de linhas.
4. Com todas as branches da rodada juntas, o coordenador roda todos os gates (Web, Audit, Contract, Backend intacto). Só então a rodada seguinte começa.
5. Uma unidade bloqueada não bloqueia as outras da mesma rodada, mas a rodada seguinte só começa quando as suas dependências estiverem integradas.

## Estratégia do verificador independente

- **Quem**: um agente novo (VERIFY), sem acesso ao histórico dos autores, com modelo de raciocínio médio-alto.
- **Entrada**: as 6 specs, os designs, o intervalo de commits (`develop..<ponto de integração final>`) e `STATE.md`. Nunca os relatórios dos autores como prova.
- **Checagem ancorada na spec**: para cada AC, localizar o teste que o cobre e confirmar que o valor afirmado é o que a spec define (por exemplo, a mensagem, o status, o corpo enviado). AC sem teste ou com teste que não discrimina vira lacuna.
- **Sensor de discriminação**: as dez falhas mínimas de INT-04 AC3, injetadas uma a uma num worktree temporário; cada falha deve derrubar ao menos um teste. Sobrevivente vira task de correção. Ao fim, o worktree é descartado e o coordenador confere `git status --porcelain` igual ao de antes.
- **Smoke**: o verificador relê `smoke.md` e repete, contra a stack local, ao menos login, um ciclo do financeiro com comprovante e a visão de `CONSELHO_FISCAL`.
- **Saída**: `validation.md` (veredito, evidência por AC com `arquivo:linha`, resultado do sensor, intervalo de commits) e lista ordenada de lacunas. Lições aprendidas registradas pelo `lessons.py` só para falhas reais.
- **Fechamento**: `validate_state.py web` passa.

## Risks & Concerns

| Risco | Mitigação |
| --- | --- |
| Smoke depender de envio de e-mail | O provedor `log` não expõe o token. O smoke cobre o pedido e o token inválido; a redefinição com token válido fica nos testes com MSW e é declarada no `smoke.md` (INT-03 AC3). |
| Stack local sem S3 | A API não sobe sem armazenamento (`errStorageRequired`) e o `docker-compose.yml` não tem serviço S3. F1 e o smoke usam um contêiner Garage avulso, com a mesma imagem de `api/internal/platform/testutil/s3.go`, configurado fora do repositório e nunca commitado. O guia (INT-02) documenta o procedimento. Incluir S3 no `docker-compose.yml` é mudança fora da Web e fica registrada como lacuna. |
| Uma frente ter interpretado diferente a lista de navegação | `nav.ts` de cada módulo tem teste próprio; a INT só concatena. |
