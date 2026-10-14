# ADR-016: Permissões sensíveis, papel `FISCAL`, autorização por jogo e Conselho Fiscal fail-closed

- **Date**: 2026-10-13 (revisão 2, decisões M-1 a M-9); 2026-10-14 (revisão 3, parecer externo: Conselho Fiscal por permissões explícitas; "exige decisão")
- **Status**: Proposed. Aprovado conceitualmente; ainda não aceito; não autoriza implementação.
- **Deciders**: @antoniobt12062002
- **Tags**: security, rbac, governance, tj-connect

## Context and Problem Statement

`identity/app/roles_matrix.go` (`BuildMatrix`) concede ao `PRESIDENTE` **toda** permissão do catálogo agregado, automaticamente. O TJ Connect traz operações de governança: reembolso, check-in, procedimento de suspensão, quota, liberação comercial. Para elas, o regulamento pode exigir outra autoridade ou separação de funções (TJ requisitos §11 e P1-14). O mantenedor exigiu:

- permissões explícitas para cada operação sensível;
- nenhuma concessão sensível automática ao `PRESIDENTE`;
- impedimentos do Conselho Fiscal *fail-closed*;
- proibição de auto-check-in do fiscal verificada no backend (M-7).

Hoje:

- `authz.Definition` não tem classificação;
- a proteção do Conselho Fiscal é uma **lista de verbos proibidos** (`forbiddenToConselhoFiscal`), e um verbo novo passaria sem alerta;
- o RBAC não restringe por instância (por exemplo, por jogo).

## Decision Drivers

- Menor privilégio e separação de funções em operações sensíveis.
- A autorização decide só por permissão efetiva (ADR-009), mais regras de instância no caso de uso.
- Fail-closed: o que não foi classificado ou permitido explicitamente é recusado, e a partida falha.
- A diretoria do clube não tem conta no MVP (D5).

## Considered Options

1. **Classificação obrigatória no catálogo**, concessão explícita das sensíveis, escopo por instância no caso de uso e Conselho Fiscal por lista permitida.
2. **Manter a concessão total ao `PRESIDENTE`** e controlar por processo.
3. **Motor de políticas (ABAC).**

## Decision Outcome

Chosen option: **1**.

### Classificação (obrigatória)

`authz.Definition` passa a exigir uma de quatro classes. **Sem classe, a partida falha.**

| Classe | Concessão automática ao `PRESIDENTE` | Pode ir a papel humano | Pode ir a papel de serviço |
|---|---|---|---|
| `normal` | Sim (regra atual) | Sim | Não |
| `sensivel` | **Não**; só por concessão explícita revisada em PR | Sim | **Não** (processo não decide, [ADR-013](013-contas-de-servico-e-autoria-de-processos.md)) |
| `propria` (dados do próprio titular) | Sim | Sim, normalmente ao papel base `ASSOCIADO` | Não |
| `servico` | Não | **Não** | Sim |

Todas as permissões já existentes são classificadas na sub-spec `02`, sem mudar quem as tem hoje, salvo decisão explícita. Decisões do mantenedor (2026-10-14), ainda sem concessão aplicada no código:

- **D-02-2 (parcial)**: a classificação vale como base de compatibilidade. A proposta revisada torna `sensivel` toda operação que movimenta recursos ou altera saldos (18 de 30), preservando acessos legítimos por concessões explícitas justificadas.
- **D-02-3 (aprovada)**: o `PRESIDENTE` deixa de ter `financeiro:prestacao_contas:approve` e `financeiro:parecer:opine`. É a única exceção aprovada à preservação da matriz.
- **D-02-4 (proposta, pendente)**: a D-02-3 só pode entrar em vigor com uma regra de nomeação institucional do Conselho Fiscal (`identity:conselho_fiscal:manage`, decisão registrada, `Covers` sem as permissões institucionais e incompatibilidade `PRESIDENTE` × `CONSELHO_FISCAL` por identidade). Sem ela, `authz.Covers` impediria qualquer pessoa de nomear ou retirar conselheiros (spec `02`, INS-01).

### Catálogo das operações exigidas (nomes finais no Specify)

| Área | Permissão | Classe | Regra adicional no caso de uso |
|---|---|---|---|
| **Reembolso** | `associados:cancelamento:request` | propria | Só para a própria adesão |
| | `associados:cancelamento:analyze` | sensivel | Registra a análise (benefícios usados, prazos legais), sem decidir |
| | `associados:cancelamento:decide` | sensivel | Decisão motivada; **não** pode ser de quem analisou, se o regulamento exigir |
| | `cobrancas:reembolso:request` | sensivel | Solicita valor e motivo, vinculado à decisão de cancelamento ou de regularização |
| | `cobrancas:reembolso:approve` | sensivel | Aprovador ≠ solicitante (sempre) |
| | `cobrancas:reembolso:execute` | servico (**exige decisão**) | Exige `decision_id` aprovado e não executado |
| | `cobrancas:reembolso:read` | normal | — |
| **Jogos** | `acesso:jogo:create` | normal | — |
| | `acesso:jogo:update` | normal | Publicar, editar, adiar, cancelar, com motivo |
| | `acesso:jogo:conclude` | sensivel | Congela a elegibilidade usada na frequência |
| | `acesso:jogo:read` | normal | — |
| | `acesso:fiscal:designate` | sensivel | Designa ou revoga fiscais por jogo |
| **Check-in** | `acesso:checkin:create` | sensivel | Designação ativa para o jogo; **fiscal ≠ titular da credencial (M-7)** |
| | `acesso:checkin:correct` | sensivel | Motivo e evidência; corretor ≠ titular; corretor ≠ fiscal do registro original |
| | `acesso:checkin:read` | normal | — |
| | `acesso:estimativa:export` | normal | Só dados agregados ao clube |
| **Procedimento** | `acesso:procedimento:open` | sensivel | Exige a guarda dos 6 jogos e < 50% ([ADR-018](018-invariantes-transacionais-de-quota-e-check-in.md)) |
| | `acesso:procedimento:decide` | sensivel | Decisor ≠ quem instaurou; exige notificação e defesa registradas (AC-FRQ-04) |
| | `acesso:procedimento:review` | sensivel | Revisor ≠ decisor; reavaliação, recurso, restabelecimento antecipado |
| | `acesso:procedimento:read` | sensivel | Dado disciplinar: leitura restrita |
| **Quota e venda** | `associados:quota:update` | sensivel | Justificativa; nunca abaixo dos compromissos |
| | `associados:liberacao_comercial:grant` | sensivel | Evidência, versão e validade de cada bloqueio (M-5) |
| | `associados:liberacao_comercial:revoke` | sensivel | Motivo; interrompe novas vendas do escopo |
| | `associados:plano:publish` | sensivel | Exige liberação comercial vigente para o plano |
| | `associados:vaga:release` | sensivel | Exige cancelamento definitivo e condições do acordo |
| **Camisetas** | `associados:lote:close` | sensivel | Fechamento manual; alteração depois do fechamento só com registro |
| **Operação** | `cobrancas:conciliacao:resolve` | sensivel | Resolve `DIVERGENTE` com evidência |
| | `cobrancas:processamento:close` | sensivel | Autorização do módulo para encerrar job de cobrança falho (junto com `jobs:job:close`); recusado com obrigação pendente |
| **Contas de serviço** | `identity:service_account:deactivate` | sensivel | Desativação de emergência, com motivo |
| | `identity:service_account:reactivate` | sensivel | Reativação com motivo; quem reativa ≠ quem desativou |

**Permissões `servico` e o atributo "exige decisão"** (revisão 3, ver [ADR-013](013-contas-de-servico-e-autoria-de-processos.md)): toda permissão `servico` declara se exige decisão humana. As de **regra aprovada** (por exemplo, sincronizar e confirmar cobrança, registrar receita de cobrança, ativar adesão paga, expirar reserva reconciliada) **não** exigem decisão caso a caso. As de **execução de decisão administrativa** (por exemplo, `cobrancas:reembolso:execute`, aplicar suspensão decidida) exigem `decision_id`. Os nomes finais ficam nas specs de cada módulo.

Os papéis que recebem cada permissão sensível **dependem do regulamento** (P0-2, P0-3, P1-14). Até o mantenedor aprovar uma matriz, **ninguém** as recebe e o fluxo correspondente não opera. Esse bloqueio é deliberado e nada é presumido. A sub-spec `02` garante que essas permissões nasçam sem concessão; a matriz proposta será submetida à aprovação do mantenedor antes da operação, nas specs de cada módulo.

### Separação de funções

Implementada no caso de uso comparando os autores registrados nos atos anteriores, nunca por nome de papel. As regras marcadas "se o regulamento exigir" ficam **ligadas por padrão** (a opção mais estrita) e só são relaxadas por decisão registrada.

### Papel `FISCAL` e escopo por jogo

- `FISCAL` entra na lista de papéis do ADR-009. É concedido por `AdminMembership` (motivo, quem concedeu, revogação). Recebe `acesso:checkin:create`, `acesso:jogo:read` e a leitura mínima da credencial no momento da leitura.
- `acesso` é dono da **designação de fiscais por jogo** (jogo, fiscal, quem designou, vigência, revogação).
- O check-in exige, **no backend, na mesma transação**:
  1. a permissão;
  2. uma designação ativa do autor para o jogo;
  3. **autor ≠ titular da credencial lida**.

  A falha em 1 ou 2 responde como falta de permissão (403 genérico). A falha em 3 responde "verificação por outro fiscal necessária" e é registrada como tentativa. O `PRESIDENTE` não é fiscal por ser presidente.

### Conselho Fiscal: fail-closed

Revisão 3 (parecer externo): restringir só por verbo ainda aceitaria `read` sobre qualquer recurso sensível novo (por exemplo, dados disciplinares) sem revisão. A proteção passa a ser por **permissões explicitamente autorizadas**:

- `forbiddenToConselhoFiscal` (lista de verbos proibidos) é substituída por uma **lista explícita e versionada de permissões** que o `CONSELHO_FISCAL` pode receber. Ela começa com exatamente as 11 que ele tem hoje: `audit:log:read`; `financeiro:conta:read`, `:lancamento:read`, `:saldo:read`, `:comprovante:read`, `:prestacao_contas:read`, `:prestacao_contas:approve`, `financeiro:parecer:opine`; `estoque:produto:read`, `:movimentacao:read`, `:saldo:read`.
- **Primeira barreira**: qualquer concessão ao `CONSELHO_FISCAL` fora da lista faz `BuildMatrix` falhar, **mesmo que a ação seja `read`**.
- **Segunda barreira**: toda permissão da lista precisa ter ação `read`, `opine` ou `approve` sobre `prestacao_contas`, e nunca classe `servico`.
- Acrescentar uma permissão à lista exige mudança revisada do arquivo versionado e do seu snapshot de teste. Um novo acesso do Conselho Fiscal nunca acontece por omissão.
- As autorizações legítimas atuais ficam preservadas exatamente (spec `02`, CFS-01).

### Relatório ao clube, sem conta

Um usuário da TJ com `acesso:estimativa:export` gera o relatório **agregado** de intenção de comparecimento por jogo (confirmados, negativas, não respondentes e, depois do jogo, total validado), sem nomes. A geração é auditada.

### Positive Consequences

- Operações de governança deixam de ser poder implícito do cargo; o que não foi classificado não funciona.
- Um verbo novo não escapa da proteção do Conselho Fiscal.
- O auto-check-in é impossível pela API, qualquer que seja a interface.

### Negative Consequences

- Muda `authz.Definition` e `BuildMatrix` (`platform` e `identity`), com spec e testes próprios, e exige classificar o catálogo atual.
- Até a matriz aprovada, os fluxos sensíveis ficam inoperantes. É o comportamento desejado, mas condiciona as entregas.

## Pros and Cons of the Options

### 1. Classificação + concessão explícita + escopo no caso de uso ✅ Chosen

- ✅ Fecha por padrão; reaproveita a matriz; sem motor novo.
- ❌ Mudança em `identity` e `platform`.

### 2. Concessão total mantida

- ❌ Contraria a diretriz do mantenedor.

### 3. Motor de políticas

- ❌ Desproporcional (já recusado no ADR-009).

## Links

- Altera [ADR-009](009-modelo-de-identidade-e-papeis-iniciais.md) e [ADR-005](005-autenticacao-e-rbac.md).
- [ADR-013](013-contas-de-servico-e-autoria-de-processos.md), [ADR-015](015-superficies-web-e-rotas-publicas-do-tj-connect.md), [ADR-018](018-invariantes-transacionais-de-quota-e-check-in.md).
