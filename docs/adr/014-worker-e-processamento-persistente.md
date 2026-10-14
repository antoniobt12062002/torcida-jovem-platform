# ADR-014: `cmd/worker` e processamento persistente em PostgreSQL

- **Date**: 2026-10-13 (revisão 2, com as decisões M-1 a M-9 do mantenedor)
- **Status**: Proposed. Aprovado conceitualmente; ainda não aceito; não autoriza implementação.
- **Deciders**: @antoniobt12062002
- **Tags**: architecture, jobs, reliability, tj-connect

## Context and Problem Statement

Hoje `api/cmd` tem só `api` e `bootstrap-admin`, sem nenhuma execução em segundo plano. O TJ Connect precisa de:

- sincronização de cobranças;
- entrega de eventos ([ADR-012](012-cobrancas-e-eventos-persistentes.md));
- expiração de reservas com reconciliação;
- conciliação periódica;
- execução de reembolsos aprovados;
- e-mails transacionais.

O mantenedor decidiu por um `cmd/worker` independente da API. Exigiu recuperação após falhas, concorrência entre instâncias, novas tentativas, tratamento de jobs que excedem o tempo e de falhas definitivas, e proibiu depender da memória de uma instância da API.

## Decision Drivers

- Um trabalho aceito sobrevive a reinício, deploy e queda.
- O trabalho é enfileirado na **mesma transação** do caso de uso que o origina.
- Várias instâncias rodam juntas sem executar o mesmo job ao mesmo tempo, e sem que uma instância "zumbi" sobrescreva o resultado de outra.
- Toda falha termina num estado visível e tratável; nada some.
- Nenhuma infraestrutura nova além do PostgreSQL 16.

## Considered Options

1. **Fila própria em PostgreSQL** (`platform/jobs`), com captura `FOR UPDATE SKIP LOCKED`, concessão com prazo e token de cerca.
2. **Biblioteca de fila sobre PostgreSQL** (por exemplo River).
3. **Broker externo** (Redis, RabbitMQ, SQS).
4. **Goroutines e timers na API.**

## Decision Outcome

Chosen option: **1**. A opção 2 só a substitui se o Design comprovar, com teste, que a biblioteca enfileira na mesma transação GORM/pgx usada pelos casos de uso, sem segunda conexão. Isso **não foi verificado**.

### Modelo

- **Tabela de jobs** (dona: `platform`), com:
  - id, tipo, payload (sem dado pessoal desnecessário);
  - `run_at`, estado (`pendente`, `em_execucao`, `concluido`, `falhou`, `encerrado_manualmente`);
  - `tentativas` e `max_tentativas`;
  - `lease_token` (UUID por captura), `locked_by`, `locked_until`;
  - `ultimo_erro` (classificado), `correlation_id`;
  - `chave_dedup`, opcional e única entre os jobs não terminais.
- **A API só enfileira** (`INSERT` na transação do caso de uso). Nenhum código de negócio da API roda em goroutine de fundo. Um teste de arquitetura impede `cmd/api` de registrar executores.
- **Relógio**: todas as comparações de tempo (`run_at`, `locked_until`) usam `now()` do banco, nunca o relógio da instância.

### Captura e concorrência entre instâncias

- Uma transação curta faz `SELECT … WHERE estado = 'pendente' AND run_at <= now() … FOR UPDATE SKIP LOCKED LIMIT n` e marca `em_execucao`, com um novo `lease_token` e `locked_until = now() + lease`. A transação é confirmada antes da execução.
- **Token de cerca (fencing)**: concluir, falhar ou renovar é `UPDATE … WHERE id = ? AND lease_token = ?`. Uma instância que perdeu a concessão (pausa longa, rede) não consegue registrar nada, porque o token já mudou.
- **Serialização por agregado** quando a regra exige (por exemplo, uma sincronização por cobrança por vez): a `chave_dedup` impede dois jobs não terminais iguais, e os efeitos de negócio usam `UPDATE` condicional por estado e versão.

### Recuperação após falha

| Falha | O que acontece |
|---|---|
| A instância cai durante o job | A concessão vence; outra instância recaptura o job (novo token) e o executa de novo |
| Queda dentro da transação de efeito | Rollback; nada parcial no banco |
| Queda depois da chamada externa e antes de gravar o efeito | Na nova tentativa, a mesma `X-Idempotency-Key`, gravada antes da chamada, faz o provedor devolver o mesmo resultado; o efeito é gravado |
| Queda depois de gravar o efeito e antes de concluir o job | Na nova tentativa, o executor vê o estado-alvo já atingido e conclui sem novo efeito |
| Reinício do worker | Na partida, os jobs com concessão vencida voltam automaticamente ao fluxo; nenhum passo manual |
| API fora do ar | Os jobs já enfileirados continuam; o worker não depende da API |

Todo executor é **obrigatoriamente idempotente**, e esse é um critério de aceite em cada spec que declara um tipo de job.

### Tempo de execução

- Cada tipo de job declara um tempo máximo; o executor roda com um `context` que vence antes da concessão.
- Um job longo legítimo renova a concessão periodicamente (com o token).
- Um job que excede o tempo máximo é interrompido pelo `context` e conta como tentativa falhada (`timeout`).

### Novas tentativas e falha definitiva

- **Classificação do erro**:
  - transitório (rede, 5xx, 429, falha de serialização) → nova tentativa;
  - permanente (validação, 4xx definitivo, estado impossível) → `falhou` imediatamente, sem gastar tentativas.
- **Espera** exponencial com variação aleatória, limitada por tipo; `max_tentativas` por tipo.
- **Falha definitiva**: estado `falhou`, mantido para sempre.
  - Gera métrica e alerta.
  - Aparece no painel operacional para quem tem `jobs:job:read`; reenfileirar ou encerrar exige `jobs:job:retry` ou `jobs:job:close` **e** a permissão de operação declarada pelo módulo dono do tipo (ex. `cobrancas:processamento:retry`), conforme a spec `01`.
  - Uma pessoa pode **reenfileirar** (novo ciclo de tentativas) ou **encerrar manualmente** com motivo (`encerrado_manualmente`), sempre com auditoria. Nunca há exclusão.
  - Um job com falha não bloqueia os outros: o isolamento é por linha.

### Agendamentos periódicos

- Conciliação, expiração e verificação de consistência são jobs que se reprogramam.
- `chave_dedup` por período (ex. `conciliacao:2027-03-01T10`) garante uma execução por janela, mesmo com várias réplicas.
- Depois de um período parado, a próxima execução cobre o atraso (consulta por estado, não por "o que mudou desde a última vez").

### Encerramento e observabilidade

- No sinal de parada, a instância deixa de capturar e espera os jobs em curso até um prazo de tolerância. O que não terminar é recapturado por outra instância quando a concessão vencer.
- Métricas: tamanho da fila, idade do job mais antigo, falhas por tipo, concessões vencidas, tentativas. Logs estruturados com `correlation_id`.

### Positive Consequences

- Trabalho e mudança de estado atômicos; reinícios e instâncias concorrentes seguros por construção.
- Sem broker; o backup do PostgreSQL cobre a fila.

### Negative Consequences

- Fila, cerca e agendador são código do projeto, com testes de concorrência reais.
- Polling no banco (pequeno no volume esperado), com índice e intervalo configuráveis.
- Um processo a mais para implantar e monitorar.

## Pros and Cons of the Options

### 1. Fila própria em PostgreSQL ✅ Chosen

- ✅ Transação única com o caso de uso; sem dependência nova; cerca explícita.
- ❌ Código de infraestrutura próprio.

### 2. Biblioteca sobre PostgreSQL

- ✅ Captura, tentativas e agendamento prontos.
- ❌ Enfileiramento na transação GORM não verificado; dependência e migrations novas.

### 3. Broker externo

- ❌ Sem atomicidade com o banco sem outbox de qualquer forma; operação extra; desproporcional.

### 4. Tarefas em memória da API

- ❌ Perdidas em reinício; vetado.

## Links

- [ADR-012](012-cobrancas-e-eventos-persistentes.md), [ADR-013](013-contas-de-servico-e-autoria-de-processos.md), [ADR-018](018-invariantes-transacionais-de-quota-e-check-in.md).
- Atualiza [architecture-overview](../architecture/architecture-overview.md) (segundo binário), quando aceito.
