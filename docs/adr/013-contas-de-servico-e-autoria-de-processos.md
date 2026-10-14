# ADR-013: Contas de serviço e autoria de processos automatizados

- **Date**: 2026-10-13 (revisão 2, decisões M-1 a M-9); 2026-10-14 (revisão 3, parecer externo: regra aprovada × decisão administrativa)
- **Status**: Proposed. Aprovado conceitualmente; ainda não aceito; não autoriza implementação.
- **Deciders**: @antoniobt12062002
- **Tags**: security, identity, audit, rbac, tj-connect

## Context and Problem Statement

Hoje toda escrita nasce de uma pessoa com sessão:

- os casos de uso exigem um `authz.Principal` com a permissão (por exemplo `financeiro/app/criar_lancamento.go`);
- `lancamentos.criado_por` e `movimentacoes_estoque.responsavel_id` são `NOT NULL REFERENCES users`;
- `platform/audit` aceita o ator `system`, mas **sem identificador** (`platform/audit/entry.go`), e a tabela `audit_log` impõe `(actor_type = 'user') = (actor_user_id IS NOT NULL)`.

O TJ Connect introduz ações sem pessoa na hora: sincronizar cobrança, registrar receita, ativar adesão, expirar reserva, conciliar, executar um reembolso aprovado. O mantenedor exige:

- contas não interativas, identificáveis e com permissões mínimas;
- origem de toda ação auditável;
- distinção clara entre **autoria humana**, **processo automatizado** e **decisão administrativa**;
- nenhum bypass genérico de autorização e nenhum registro sem autoria identificável.

## Decision Drivers

- Toda escrita tem um autor nomeado: uma pessoa ou um processo registrado.
- A autorização continua sendo por permissão verificada no caso de uso, sem "modo sistema".
- Processos executam regras determinísticas; **juízo discricionário é sempre humano**.
- Chaves estrangeiras e tabelas existentes mudam o mínimo possível.

## Considered Options

1. **Contas de serviço em `identity`**, com permissões exclusivas de serviço, autoria tipada e cadeia de causalidade na auditoria.
2. **Colunas de autoria polimórficas** (`criado_por` anulável e `criado_por_processo`) em cada tabela.
3. **Porta interna sem permissão**, confiando no chamador.

## Decision Outcome

Chosen option: **1**.

### Três tipos de autoria

| Tipo | Quem age | Exemplo | Registro na auditoria |
|---|---|---|---|
| **Humana direta** | Pessoa com sessão, pela própria permissão | Associado solicita cancelamento; fiscal registra check-in | `actor_type = user`, `actor_user_id` = a pessoa |
| **Decisão administrativa** | Pessoa com permissão **sensível**, com motivação registrada | Diretor aprova reembolso; autoridade decide suspensão; responsável registra liberação comercial | Igual à humana, mais o **registro de decisão** no módulo dono (decisão, motivo, evidência, versão) com id próprio (`decision_id`) |
| **Processo automatizado: execução de regra aprovada** | Conta de serviço, aplicando uma regra determinística fixada em spec, código e ADR, a partir de um fato verificável | Confirmar o pagamento consultado no provedor, registrar a receita da cobrança paga, ativar a adesão paga com vaga, expirar a reserva depois da reconciliação, conciliar | `actor_type = service` (novo), `actor_user_id` = a conta de serviço, `regra` (nome e versão), `causation_id`, `correlation_id` e, quando houver, `on_behalf_of_user_id`. **Sem decisão humana caso a caso** |
| **Processo automatizado: execução de decisão administrativa** | Conta de serviço, executando o efeito de uma decisão humana já registrada | Executar no provedor o reembolso **já aprovado**; aplicar uma suspensão **já decidida** | Igual à linha anterior, mais `decision_id` obrigatório |

Regras:

- **Processo nunca exerce juízo discricionário.** As permissões de juízo (aprovar reembolso, decidir suspensão, liberar vaga, corrigir presença, conceder liberação comercial) são `sensivel` e **não podem** ser concedidas a papel de serviço.
- **Regra aprovada não exige decisão por caso.** Confirmar um pagamento, registrar a receita correspondente, ativar a adesão paga com vaga ou expirar uma reserva reconciliada são aplicações determinísticas de regras já aprovadas. Exigir uma decisão humana para cada pagamento seria artificial e é expressamente rejeitado.
- **Execução de decisão exige `decision_id`.** Toda permissão `servico` declara se "exige decisão". Uma permissão que exige (por exemplo `cobrancas:reembolso:execute`) só é aceita pelo caso de uso junto com o `decision_id` de uma decisão humana `APROVADA`, do mesmo escopo e ainda não executada. Sem isso, a recusa é auditada.
- **Humano não executa como processo**: as permissões exclusivas de serviço não podem ser concedidas a papel humano.
- A matriz (`BuildMatrix`, [ADR-016](016-permissoes-sensiveis-e-escopo-por-jogo.md)) recusa os dois casos e **impede a partida** da API e do worker.

### Conta de serviço

- **Identificável**: linha em `users` com tipo `SERVICE` (coluna nova em `identity`), nome estável e legível (`svc.cobrancas-sincronizacao`), descrição, módulo dono e data de criação. A lista é declarada em código e provisionada pela sincronização da partida, com auditoria, como os papéis.
- **Não interativa**:
  - sem credencial verificável (o hash de senha é um valor que nenhuma senha satisfaz);
  - `Authenticate`, recuperação de senha e criação de sessão recusam `SERVICE`;
  - a tabela de sessões só aceita usuário humano, garantido no banco por gatilho ou restrição equivalente, a decidir no Design;
  - não pode receber `AdminMembership` nem papel humano.
- **Privilégio mínimo**: um papel de serviço por processo, só com as permissões exclusivas de serviço que esse processo usa. Cada papel declara a lista fechada, e a partida falha se um papel de serviço tiver permissão não classificada como exclusiva de serviço.
- **Como obtém o Principal**: não há token nem login. O worker constrói o `Principal` da conta de serviço a partir das permissões efetivas na matriz. A fronteira de confiança é o próprio processo do worker (credenciais do banco), já protegida pelo ambiente; o abuso é detectável pela auditoria.
- **Desativação como chave de emergência**: o worker verifica `is_active` da conta antes de cada job. Uma conta desativada suspende o processo, auditado, sem deploy.
- **Listagens**: telas e contagens de usuários filtram `SERVICE` por padrão.

### Auditoria

- `audit_log.actor_type` ganha `service`, com id **obrigatório**. A restrição da tabela passa a ser "`user` e `service` exigem `actor_user_id`; `anonymous` não tem". Muda por migration em `platform`; aqui fica só o planejamento.
- O ator `system` sem id deixa de ser aceito em caminhos novos. Um teste proíbe o seu uso fora da partida e do bootstrap, que ficam documentados como as únicas exceções.
- `context` padroniza `correlation_id` (fluxo inteiro, ex. do checkout à ativação), `causation_id` (evento, notificação ou job imediato), `on_behalf_of_user_id` (pessoa cuja ação originou o fluxo, ex. quem iniciou o checkout), `regra` (nome e versão da regra aprovada aplicada) e `decision_id` (decisão humana executada, só no segundo tipo de execução).
- **Desativação e reativação** de conta de serviço usam permissões distintas (`identity:service_account:deactivate` e `:reactivate`); a reativação exige pessoa diferente de quem desativou. Ver spec `02` (SVC-01).
- Com isso, para qualquer efeito automático é possível responder: qual processo, disparado por quê, em nome de quem e sob qual decisão.

### Positive Consequences

- `criado_por` e `responsavel_id` continuam `NOT NULL REFERENCES users`, sem migração nessas tabelas.
- O RBAC é uniforme e sem bypass; juízo e execução ficam separados por construção.
- A autoria é nominal em toda ação automatizada.

### Negative Consequences

- `users` passa a conter contas não humanas; consultas e telas precisam filtrar.
- Migrations em `identity` (tipo de conta, proteção de sessão) e em `platform` (restrição do `audit_log`).
- O ADR-009 precisa da exceção "papel de serviço não exige `AdminMembership`, e humano não recebe papel de serviço".

## Pros and Cons of the Options

### 1. Contas de serviço ✅ Chosen

- ✅ Sem mudança de FK; RBAC uniforme; autoria nominal; juízo humano obrigatório.
- ❌ Contas não humanas em `users`.

### 2. Autoria polimórfica

- ❌ Migração em toda tabela com autor; FK anulável; a autorização dos processos continua sem modelo.

### 3. Porta interna sem permissão

- ❌ É o bypass vetado.

## Links

- Altera [ADR-009](009-modelo-de-identidade-e-papeis-iniciais.md) e o comportamento de `platform/audit` ([ADR-004](004-auditoria-e-imutabilidade-financeira.md)).
- [ADR-012](012-cobrancas-e-eventos-persistentes.md), [ADR-014](014-worker-e-processamento-persistente.md), [ADR-016](016-permissoes-sensiveis-e-escopo-por-jogo.md), [ADR-017](017-evolucao-aditiva-de-financeiro-e-estoque.md).
