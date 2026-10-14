# Platform e processamento — Especificação

Status: **Rascunho para revisão do mantenedor** (2026-10-13). Sub-spec `01` do TJ Connect, grupo G1 da Onda A. Coordenação em `.specs/features/tj-connect/STATE.md`; contexto em `context.md`. Narrativa em português; critérios de aceite em inglês (EARS).

## Problem Statement

A plataforma não tem nenhuma execução em segundo plano: `api/cmd` contém só `api` e `bootstrap-admin`. O TJ Connect exige trabalhos que não podem depender de uma requisição nem da memória de uma instância da API:

- consultar o provedor de pagamento depois de um webhook;
- entregar a confirmação de pagamento a outro módulo;
- expirar reservas só depois de reconciliar;
- conciliar periodicamente;
- executar reembolsos aprovados.

Esta spec define a **capacidade de plataforma** que torna isso confiável:

- um worker separado e persistente;
- jobs registrados na mesma transação do caso de uso;
- eventos de domínio com entrega garantida e consumidores idempotentes;
- recuperação após falhas;
- concorrência entre instâncias;
- estados de falha definitiva tratáveis;
- observabilidade.

Ela não contém nenhuma regra de cobrança, adesão, quota ou check-in.

## Goals

- [ ] Nenhum trabalho aceito se perde por reinício, deploy ou queda de qualquer processo.
- [ ] Nenhum job é executado ao mesmo tempo por duas instâncias, e uma instância que perdeu a concessão não consegue registrar resultado.
- [ ] Todo evento publicado é entregue a cada consumidor inscrito pelo menos uma vez, com efeito aplicado exatamente uma vez por consumidor.
- [ ] Toda falha termina num estado visível, com ação de recuperação auditada; nada é descartado em silêncio.
- [ ] A API só registra trabalho; quem executa é sempre o `cmd/worker`.

## Out of Scope

| Item | Razão |
|---|---|
| Tipos de job e de evento de negócio (sincronizar cobrança, expirar reserva, `cobranca.paga` etc.) | Declarados pelas specs `04`–`08`; aqui só o contrato que eles usam |
| Contas de serviço, classes de permissão e autoria | Spec `02`; esta spec só as usa |
| Tela administrativa de jobs e eventos | Spec `10` ou posterior; aqui só as capacidades e permissões |
| Biblioteca de fila de terceiros | ADR-014: fila própria por padrão; uma biblioteca só entra se o Design provar que enfileira na transação GORM |
| Broker externo, filas em memória, cron do sistema operacional | ADR-014 |
| Ordem global de eventos | Não garantida; ver "Ordem de eventos" |
| Infraestrutura de métricas externa (Prometheus etc.) | Nenhuma existe hoje; a observabilidade usa logs estruturados e consultas no banco (premissa) |

---

## Relação com requisitos e ADRs

| Origem | O que esta spec cobre |
|---|---|
| TJ requisitos §16 (Integridade: idempotência e consistência em falhas parciais; Operação: observabilidade de webhooks, falhas, reservas expiradas) | Jobs, outbox, inbox, recuperação, observabilidade |
| TJ-PAG-002, TJ-PAG-004, TJ-CAP-003 | Capacidades que as specs `04`/`05` usam para expirar e conciliar sem depender do relógio local nem da API |
| Decisão D3 do mantenedor (worker separado, persistente) | `cmd/worker` |
| [ADR-014](../../../../docs/adr/014-worker-e-processamento-persistente.md) | Fila em PostgreSQL, concessão, token de cerca, tentativas, falha definitiva |
| [ADR-012](../../../../docs/adr/012-cobrancas-e-eventos-persistentes.md) | Outbox, inbox, envelope de eventos |
| [ADR-011](../../../../docs/adr/011-mapa-de-modulos-do-tj-connect.md) | Pacote `events` só de tipos; inscrições só nas setas permitidas; regras do `architecture_test` |
| [ADR-013](../../../../docs/adr/013-contas-de-servico-e-autoria-de-processos.md), spec `02` | Execução com conta de serviço; autoria e causalidade |

---

## Atores, casos de uso e permissões

### Atores

| Ator | Papel no processamento |
|---|---|
| Caso de uso de módulo (na API ou no worker) | Registra jobs e publica eventos **na sua transação** |
| `cmd/worker` (N instâncias) | Captura e executa jobs; entrega eventos; roda agendamentos |
| Conta de serviço (spec `02`) | Identidade com que cada executor de job age |
| Operador com permissão | Consulta filas, reenfileira ou encerra jobs com falha definitiva |
| Composição (`cmd/api`, `cmd/worker`) | Registra tipos de job, executores, tipos de evento e inscrições |

### Casos de uso

| Caso de uso | Quem | Permissão | Classe |
|---|---|---|---|
| Registrar job | Caso de uso de módulo, só para tipos do próprio módulo | — (interno, na transação) | — |
| Publicar evento | Caso de uso de módulo, só para tipos do próprio módulo | — (interno, na transação) | — |
| Executar job / entregar evento | Worker, com a conta de serviço do tipo | Permissões `servico` do caso de uso chamado | servico |
| Consultar jobs, tentativas, eventos e instâncias do worker | Operador | `jobs:job:read` | normal |
| Reenfileirar job com falha definitiva | Operador | `jobs:job:retry` **e** a permissão de operação declarada pelo tipo de job | sensivel |
| Encerrar job manualmente | Operador | `jobs:job:close` **e** a permissão de operação declarada pelo tipo de job | sensivel |

`jobs` é um pseudo-módulo de plataforma com `Contribution` própria, como o `audit` atual. As três permissões começam **sem concessão** (a matriz sensível não está aprovada). `jobs:job:read` é `normal` e também começa sem concessão até decisão (D-01-1).

---

## Estados, transições e invariantes

### Job

| Estado | Significado | Terminal |
|---|---|---|
| `PENDENTE` | Pronto para executar a partir de `executar_em` | Não |
| `EM_EXECUCAO` | Capturado por uma instância, com concessão válida até `concessao_ate` | Não |
| `CONCLUIDO` | Executado com sucesso | Sim |
| `FALHOU` | Falha permanente ou tentativas esgotadas; aguarda ação humana | Não (só sai por ação humana) |
| `ENCERRADO_MANUALMENTE` | Encerrado por uma pessoa, com motivo | Sim |

Transições:

| De → Para | Gatilho |
|---|---|
| `PENDENTE → EM_EXECUCAO` | Captura |
| `EM_EXECUCAO → CONCLUIDO` | Sucesso, com token válido |
| `EM_EXECUCAO → PENDENTE` | Erro transitório com tentativas restantes, com novo `executar_em` |
| `EM_EXECUCAO → FALHOU` | Erro permanente, ou tentativas esgotadas |
| `EM_EXECUCAO → PENDENTE` | Concessão vencida, por recuperação automática (a tentativa conta como falha por "concessão vencida") |
| `FALHOU → PENDENTE` | Reenfileiramento humano: novo ciclo de tentativas |
| `FALHOU → ENCERRADO_MANUALMENTE` | Encerramento humano |

### Tentativa de job

Cada execução gera um registro append-only: instância, início, fim, resultado (`sucesso`, `erro_transitorio`, `erro_permanente`, `tempo_esgotado`, `concessao_perdida`) e mensagem de erro **sanitizada**.

### Evento

`PUBLICADO` (gravado na outbox). Na publicação, gera **um job de entrega por consumidor inscrito**. Cada entrega segue o ciclo de vida do job.

### Invariantes

- **INV-P1**: um job ou evento só existe se a transação que o registrou foi confirmada; um rollback não deixa nenhum.
- **INV-P2**: em qualquer instante, no máximo uma instância tem concessão válida sobre um job.
- **INV-P3**: só a instância dona do token atual consegue mudar o estado de um job `EM_EXECUCAO`.
- **INV-P4**: para cada par (consumidor, evento), o efeito do consumidor é aplicado no máximo uma vez.
- **INV-P5**: nenhum job sai de `FALHOU` sem um ato humano auditado.
- **INV-P6**: nenhum código da API executa jobs nem entrega eventos.
- **INV-P7**: um módulo só registra jobs e publica eventos dos tipos com o seu prefixo.

---

## Fluxos

### F1 — Registrar trabalho na transação (principal)

1. O caso de uso abre a transação (`WithTx`), altera o seu estado e chama "registrar job" ou "publicar evento".
2. O job ou evento é gravado na mesma transação.
3. Com commit, o trabalho fica disponível; com rollback, deixa de existir (INV-P1).

**F1a**: chamado fora de uma transação → erro imediato, sem gravar nada (mesmo comportamento do recorder de auditoria).
**F1b**: chave de deduplicação igual à de um job não terminal do mesmo tipo → nenhum job novo; o caso de uso recebe o id do existente.

### F2 — Capturar e executar (principal)

1. O worker captura até N jobs prontos (`PENDENTE` e `executar_em <= agora do banco`), ignorando os bloqueados por outra instância. Marca cada um `EM_EXECUCAO` com novo token e `concessao_ate`, e confirma a captura.
2. Verifica se a conta de serviço do tipo está ativa:
   - inativa → devolve o job a `PENDENTE` com `executar_em` 60 s depois, **sem consumir tentativa**, e registra um aviso (no máximo um por minuto por tipo).
3. Executa com tempo máximo do tipo (menor que a concessão), usando o `Principal` da conta de serviço, o `correlation_id` do job e o `causation_id` = id do job.
4. Sucesso → `CONCLUIDO` (condicionado ao token).

**F2a**: job longo → renova `concessao_ate` periodicamente, condicionado ao token.
**F2b**: o token não confere ao concluir → resultado descartado, tentativa registrada como `concessao_perdida`; o efeito de negócio já é idempotente.

### F3 — Falha e nova tentativa

1. Erro classificado como transitório (rede, 5xx, 429, conflito de serialização, tempo esgotado) → `PENDENTE` com espera exponencial e variação aleatória, enquanto houver tentativas.
2. Erro permanente (validação, estado impossível, 4xx definitivo) → `FALHOU` imediatamente.
3. Tentativas esgotadas → `FALHOU`.
4. `FALHOU` gera registro de alerta e aparece na consulta de operação.

### F4 — Recuperação após queda ou reinício

1. Toda instância, na partida e periodicamente, devolve a `PENDENTE` os jobs `EM_EXECUCAO` com `concessao_ate` vencida, registrando a tentativa como `concessao_perdida`.
2. Nenhum passo manual é necessário.

### F5 — Publicar e entregar evento

1. O produtor publica, na sua transação, um evento com envelope completo (ver "Envelope").
2. Na mesma transação, a plataforma cria um job de entrega para cada consumidor inscrito naquele tipo.
3. O job de entrega abre a transação do consumidor, grava `(consumidor, evento_id)` na inbox e chama o tratador do consumidor. Os dois são confirmados juntos.
4. Se a inbox já contém o par → conclui sem chamar o tratador (duplicata).

### F6 — Ação humana sobre falha definitiva

1. O operador consulta os jobs `FALHOU` com tipo, tentativas e erros sanitizados.
2. **Reenfileirar**: exige as permissões, um motivo (≥ 10 caracteres) e gera auditoria. O job volta a `PENDENTE` com tentativas zeradas no novo ciclo; o histórico de tentativas fica preservado.
3. **Encerrar**: exige as permissões e um motivo; `ENCERRADO_MANUALMENTE`, com auditoria.

### F7 — Agendamentos periódicos

1. Cada agendamento declarado (tipo, intervalo, conta de serviço) mantém um job por janela, identificado pela chave `<tipo>:<início da janela>`.
2. Ao concluir uma janela, o próprio job registra a próxima.
3. Depois de um período parado, a próxima execução cobre o atraso, porque o trabalho é definido por estado ("o que está pendente"), não por "o que mudou desde a última vez".

---

## Falhas e concorrência

| Situação | Comportamento exigido |
|---|---|
| Duas instâncias capturam ao mesmo tempo | Cada job é capturado por uma só (`SKIP LOCKED`) |
| Instância pausada além da concessão | Outra recaptura com novo token; a primeira não consegue concluir (INV-P3) |
| Queda dentro da transação de efeito | Rollback; nova tentativa refaz o efeito inteiro |
| Queda depois do efeito e antes de concluir | Nova tentativa; o executor reconhece o efeito já aplicado (idempotência obrigatória do executor, declarada por tipo) |
| Queda depois de uma chamada externa e antes de gravar | Nova tentativa com a mesma chave de idempotência externa, gravada antes da chamada (contrato para os executores) |
| Banco indisponível | O worker para de capturar, registra o erro e tenta reconectar com espera; nada é marcado como falho por indisponibilidade do banco |
| Evento entregue duas vezes ao mesmo consumidor | A inbox impede um segundo efeito (INV-P4) |
| Eventos fora de ordem | Ver "Ordem de eventos" |
| Conta de serviço desativada | Jobs dos tipos dela ficam adiados, sem gastar tentativas, e geram alerta |
| Encerramento da instância (sinal de parada) | Para de capturar; espera os jobs em curso até o prazo de tolerância; o resto é recapturado quando a concessão vence |

### Envelope de evento

`evento_id` (único), `tipo` com versão (`<modulo>.<fato>.v<n>`), `agregado_tipo`, `agregado_id`, `agregado_versao` (inteiro crescente por agregado, atribuído pelo produtor na mesma transação), `ocorrido_em`, `correlation_id`, `causation_id`, e `payload` com o **estado resultante** da transição (não um delta), sem dado pessoal desnecessário.

### Ordem de eventos

A plataforma **não garante ordem** de entrega, nem global nem por agregado. Cada inscrição declara um modo:

- **`tolerante`**: a inbox guarda a maior `agregado_versao` aplicada por (consumidor, agregado). Um evento com versão menor ou igual é concluído **sem chamar o tratador** e registrado como `obsoleto`. Como o payload traz o estado resultante, o consumidor converge para o mais recente.
- **`estrita`**: um evento só é aplicado se a versão anterior do mesmo agregado já foi aplicada por aquele consumidor. Se não foi, a entrega volta a `PENDENTE` como erro transitório, até as tentativas acabarem; aí vira `FALHOU`, visível.

---

## Limites entre plataforma e módulos

- A plataforma conhece **tipos**, não regras: não lê o payload além do envelope e não decide nada de negócio.
- Os tipos de job e de evento têm o prefixo do módulo dono. Cada módulo recebe da composição um registrador restrito ao seu prefixo (INV-P7).
- O executor de um tipo de job é declarado pelo módulo dono e registrado **só** na composição do worker (INV-P6).
- As inscrições em eventos só existem nas setas `E` permitidas pela [ADR-011](../../../../docs/adr/011-mapa-de-modulos-do-tj-connect.md); a composição falha na partida se uma inscrição não corresponder a uma seta permitida.
- O payload tem limite de 64 KiB e não contém segredo, credencial, senha, token, CPF ou foto. A regra é verificada por revisão e por uma lista de campos proibidos testada nos tipos declarados.
- As tabelas de jobs, tentativas, outbox, inbox e instâncias do worker são de `platform`. Os módulos nunca as escrevem diretamente.

---

## Assumptions & Open Questions

| Assumption / decision | Chosen default | Rationale | Confirmed? |
|---|---|---|---|
| Valores padrão (configuráveis por ambiente) | Polling de 1 s; concessão de 60 s; tempo máximo padrão de 45 s; até 10 tentativas; espera exponencial a partir de 5 s, até 15 min, com variação aleatória de ±20%; prazo de tolerância no encerramento de 30 s | Volume baixo (centenas de associados); a concessão é maior que o tempo máximo | n |
| Retenção | Jobs `CONCLUIDO` e eventos entregues a todos os consumidores: 90 dias. Registros da inbox: 180 dias. `FALHOU` e `ENCERRADO_MANUALMENTE`: sem expiração | São dados técnicos; o rastro de negócio vive nas tabelas dos módulos e na auditoria. A inbox dura mais que a possível reentrega | n |
| Observabilidade | Logs estruturados com campos fixos, mais consultas de operação na API (sem stack de métricas externa) | Não há infraestrutura de métricas no projeto hoje | n |
| Batimento das instâncias | Cada instância grava um batimento a cada 15 s; sem batimento há mais de 60 s, a instância conta como inativa | Permite ver "nenhum worker ativo" | n |
| Ordem de eventos | Sem garantia; modos `tolerante` e `estrita` por inscrição | Evita bloquear agregados inteiros por padrão; o modo estrito fica disponível onde for necessário | y |
| Fila própria ou biblioteca | Própria (ADR-014), salvo prova no Design | Enfileirar na transação GORM é requisito | y |
| Quem recebe `jobs:job:read`, `jobs:job:retry`, `jobs:job:close` | Ninguém até decisão (D-01-1) | Não presumir concessão | n |
| Mensagens de erro guardadas | Truncadas em 2 KiB, sem cabeçalhos HTTP, tokens nem corpo de resposta do provedor | Evita vazar segredo ou dado pessoal | y |

**Open questions:** none — as decisões ainda abertas estão listadas acima como premissas não confirmadas e em "Pendências que impedem o Design".

---

## User Stories

### P1: Registro transacional de jobs ⭐ MVP

**User Story**: Como módulo de negócio, quero registrar trabalho na mesma transação da minha mudança de estado, para que nenhum trabalho exista sem a mudança nem a mudança sem o trabalho.

**Why P1**: Base de toda execução assíncrona.

**Acceptance Criteria**:
1. WHEN a use case registers a job inside its transaction and the transaction commits THEN the job SHALL exist in state `PENDENTE`.
2. WHEN a use case registers a job inside its transaction and the transaction rolls back THEN no job SHALL exist.
3. IF a job is registered outside a transaction THEN the platform SHALL return an error and SHALL write nothing.
4. IF a module registers a job whose type prefix is not its own module THEN the platform SHALL refuse it with an error.
5. WHEN a job is registered with a deduplication key equal to that of a non-terminal job of the same type THEN the platform SHALL NOT create a new job and SHALL return the id of the existing one.
6. IF a job payload exceeds 64 KiB THEN the platform SHALL refuse the job with an error.
7. The API process SHALL register no job executor and SHALL run no job.

**Independent Test**: num teste de integração, registrar um job e confirmar, registrar outro e desfazer, registrar fora de transação, e ver só o primeiro existir.

---

### P1: Captura, concessão e token de cerca ⭐ MVP

**User Story**: Como operador, quero que várias instâncias do worker rodem juntas sem executar o mesmo job ao mesmo tempo nem sobrescrever o resultado umas das outras, para escalar e reiniciar sem risco.

**Why P1**: Garante INV-P2 e INV-P3.

**Acceptance Criteria**:
1. WHEN several worker instances poll at the same time THEN each ready job SHALL be claimed by at most one instance.
2. WHEN a job is claimed THEN the platform SHALL assign it a new lease token and a lease expiry computed from the database clock.
3. IF an instance tries to complete, fail or renew a job with a lease token that is no longer current THEN the platform SHALL reject the change and SHALL record the attempt as `concessao_perdida`.
4. WHILE a job is running longer than half of its lease the owning instance SHALL renew the lease using its current token.
5. The platform SHALL compare every scheduling and lease time against the database clock, never against the instance clock.

**Independent Test**: 5 instâncias e 100 jobs → cada job executado uma vez; pausar uma instância além da concessão e ver a sua conclusão rejeitada.

---

### P1: Tentativas, tempo máximo e falha definitiva ⭐ MVP

**User Story**: Como operador, quero que erros passageiros sejam repetidos automaticamente e que erros definitivos parem num estado visível, para não perder trabalho nem repetir para sempre.

**Why P1**: Exigência D3 (tentativas e operações definitivamente malsucedidas).

**Acceptance Criteria**:
1. WHEN an executor returns a transient error and attempts remain THEN the job SHALL return to `PENDENTE` with a next run time given by exponential backoff with jitter.
2. WHEN an executor returns a permanent error THEN the job SHALL move to `FALHOU` without further attempts.
3. WHEN a job exhausts its maximum attempts THEN it SHALL move to `FALHOU`.
4. IF an execution exceeds the maximum run time of its type THEN the platform SHALL cancel it and SHALL record the attempt as `tempo_esgotado`, counted as a transient error.
5. The platform SHALL record every attempt as an append-only row with instance, start, end, result and an error message truncated to 2 KiB.
6. The platform SHALL NOT store HTTP headers, tokens, secrets or provider response bodies in attempt records.
7. WHEN a job moves to `FALHOU` THEN the platform SHALL emit a structured log at level `error` with job id, type, attempts and correlation id.
8. The platform SHALL NOT delete a job in `FALHOU` or `ENCERRADO_MANUALMENTE`.

**Independent Test**: executores falsos que falham de forma transitória, permanente e por tempo, e verificar estados, esperas e tentativas registradas.

---

### P1: Recuperação após falhas e reinícios ⭐ MVP

**User Story**: Como mantenedor, quero que uma queda em qualquer ponto não perca nem duplique trabalho, para não depender de nenhuma instância estar viva.

**Why P1**: Exigência D3 (recuperação após falha; nada em memória).

**Acceptance Criteria**:
1. WHEN a worker instance starts THEN it SHALL return to `PENDENTE` every job in `EM_EXECUCAO` whose lease has expired.
2. WHILE any instance is running it SHALL periodically return expired-lease jobs to `PENDENTE`.
3. WHEN an instance crashes in the middle of a job THEN another instance SHALL execute that job again after the lease expires, with no manual step.
4. WHEN the database is unavailable THEN the worker SHALL stop claiming, SHALL retry the connection with backoff and SHALL NOT mark any job as failed because of that unavailability.
5. WHEN an instance receives a stop signal THEN it SHALL stop claiming new jobs and SHALL wait up to the grace period for running jobs before exiting.
6. The platform SHALL require every job type to declare that its executor is idempotent, and SHALL refuse to register a type without that declaration.

**Independent Test**: matar uma instância durante um job e ver outra concluí-lo; reiniciar o banco durante a execução e ver nenhum job marcado como falho por isso.

---

### P1: Outbox e entrega de eventos ⭐ MVP

**User Story**: Como módulo produtor, quero publicar um fato na minha transação e ter certeza de que cada consumidor inscrito o receberá, para comunicar módulos sem dependência circular.

**Why P1**: Contrato da confirmação de pagamento (ADR-012).

**Acceptance Criteria**:
1. WHEN a use case publishes an event inside its transaction and the transaction commits THEN the event and one delivery job per subscribed consumer SHALL exist.
2. WHEN the publishing transaction rolls back THEN neither the event nor any delivery job SHALL exist.
3. The platform SHALL require every event to carry `evento_id`, a versioned `tipo`, `agregado_tipo`, `agregado_id`, `agregado_versao`, `ocorrido_em`, `correlation_id` and `causation_id`.
4. IF a module publishes an event whose type prefix is not its own module THEN the platform SHALL refuse it with an error.
5. IF the composition subscribes a consumer to an event type through an edge not allowed by ADR-011 THEN the worker and the API SHALL refuse to start.
6. WHEN a consumer is subscribed to an event type THEN every event of that type published after the subscription SHALL be delivered to that consumer at least once.

**Independent Test**: publicar e confirmar, publicar e desfazer, e ver só no primeiro caso um job de entrega por consumidor inscrito.

---

### P1: Consumidores idempotentes e eventos fora de ordem ⭐ MVP

**User Story**: Como módulo consumidor, quero aplicar cada evento exatamente uma vez e ser protegido de eventos atrasados, para que repetições e atrasos não causem efeitos duplicados ou regressões.

**Why P1**: INV-P4; exigência de tratamento fora de ordem.

**Acceptance Criteria**:
1. WHEN a delivery job runs THEN the platform SHALL write the pair (consumer, `evento_id`) in the inbox in the same transaction as the consumer's effect.
2. IF the pair (consumer, `evento_id`) already exists in the inbox THEN the platform SHALL complete the delivery without calling the consumer handler.
3. WHERE a subscription uses mode `tolerante` the platform SHALL complete without calling the handler, and SHALL record the delivery as `obsoleto`, when the event's `agregado_versao` is lower than or equal to the highest version already applied by that consumer for that aggregate.
4. WHERE a subscription uses mode `estrita` the platform SHALL treat a delivery as a transient error when the previous `agregado_versao` of the same aggregate has not been applied by that consumer.
5. WHEN a consumer handler fails THEN the inbox pair SHALL NOT be written and the delivery SHALL follow the job retry rules.

**Independent Test**: entregar o mesmo evento 3 vezes (1 efeito); entregar v2 antes de v1 nos dois modos e ver `obsoleto` no tolerante e espera no estrito.

---

### P1: Execução com conta de serviço e auditoria ⭐ MVP

**User Story**: Como auditor, quero que todo efeito produzido pelo worker apareça como ato de uma conta de serviço nomeada, ligado ao job e ao fluxo de origem, para reconstruir o porquê de cada ação automática.

**Why P1**: Integra a spec `02` ao processamento.

**Acceptance Criteria**:
1. The platform SHALL require every job type and every event subscription to declare exactly one service account.
2. IF the declared service account does not exist or is not of type `SERVICE` THEN the worker SHALL refuse to start.
3. WHEN the worker runs a job THEN it SHALL invoke the executor with the principal of the declared service account, the job's `correlation_id` and `causation_id` equal to the job id.
4. WHILE the declared service account is inactive the worker SHALL postpone the jobs of that type without consuming attempts, retrying after 60 seconds, and SHALL emit a structured log at level `warn` at most once per minute per job type.
5. The platform SHALL NOT write an audit entry for each job attempt; business effects SHALL be audited by the module use cases with the service account as actor.
6. WHEN a person re-enqueues or closes a failed job THEN the platform SHALL write an audit entry with the person as actor, the job id and the reason.

**Independent Test**: um job de teste que chama um caso de uso auditado; ver a entrada com `actor_type` `service`, `causation_id` = job e `correlation_id` do fluxo; desativar a conta e ver o adiamento.

---

### P1: Recuperação humana de falhas ⭐ MVP

**User Story**: Como operador autorizado, quero reenfileirar ou encerrar um job que falhou de vez, com motivo registrado, para resolver incidentes sem acesso direto ao banco.

**Why P1**: Exigência de ações de recuperação para falhas definitivas.

**Acceptance Criteria**:
1. WHEN a person with `jobs:job:read` lists jobs by state, type or `correlation_id` THEN the platform SHALL return them with attempts and sanitized errors, without the payload.
2. WHEN a person holding `jobs:job:retry` and the operation permission declared by the job type re-enqueues a job in `FALHOU` with a reason of at least 10 characters THEN the job SHALL return to `PENDENTE` with a new attempt cycle and its previous attempts kept.
3. WHEN a person holding `jobs:job:close` and the operation permission declared by the job type closes a job in `FALHOU` with a reason of at least 10 characters THEN the job SHALL move to `ENCERRADO_MANUALMENTE`.
4. IF re-enqueue or close is requested for a job not in `FALHOU` THEN the platform SHALL refuse it with `conflict`.
5. IF the person lacks either required permission THEN the platform SHALL refuse with `forbidden` and SHALL write a `denied` audit entry.

**Independent Test**: provocar um `FALHOU`, reenfileirar e ver o novo ciclo com histórico preservado; tentar sem a permissão do tipo e ver a recusa.

---

### P1: Agendamentos periódicos ⭐ MVP

**User Story**: Como módulo, quero declarar uma tarefa periódica (como conciliação ou expiração) que rode uma vez por janela mesmo com várias instâncias, e que recupere janelas perdidas, para não depender de cron externo.

**Why P1**: Conciliação e expiração das specs `04`/`05`.

**Acceptance Criteria**:
1. WHEN the worker starts THEN it SHALL ensure one pending job exists for the current window of every declared schedule.
2. The platform SHALL guarantee at most one job per schedule and window, identified by the schedule type and the window start.
3. WHEN a scheduled job completes or fails permanently THEN the platform SHALL register the job for the next window.
4. WHEN the worker was stopped for several windows THEN the platform SHALL register only the current window, and the executor SHALL process all pending state in that run.

**Independent Test**: 3 instâncias com um agendamento de 1 minuto por 5 minutos → 5 execuções; parar tudo por 3 janelas e ver uma execução que cobre o atraso.

---

### P1: Observabilidade e condições de operação ⭐ MVP

**User Story**: Como operador, quero saber se há worker ativo, quanto trabalho está atrasado e o que falhou, para agir antes que o atraso afete pagamentos ou reservas.

**Why P1**: TJ requisitos §16 Operação.

**Acceptance Criteria**:
1. The worker SHALL emit one structured log line per attempt with job id, type, attempt number, result, duration in milliseconds and `correlation_id`.
2. WHILE a worker instance runs it SHALL write a heartbeat every 15 seconds with instance id, start time and version.
3. WHEN a person with `jobs:job:read` requests the operational summary THEN the platform SHALL return counts per state and type, the age of the oldest `PENDENTE` job, the number of `FALHOU` jobs and the list of instances with their last heartbeat.
4. IF no instance has a heartbeat newer than 60 seconds THEN the operational summary SHALL report `nenhum_worker_ativo`.
5. IF the oldest `PENDENTE` job is older than 5 minutes THEN the worker SHALL emit a structured log at level `warn` once per minute while the condition holds.
6. IF the worker configuration is invalid or the database is unreachable at start THEN the worker SHALL exit with a non-zero code and a message naming the problem.
7. The worker SHALL expose no inbound network port.

**Independent Test**: parar todas as instâncias e ver `nenhum_worker_ativo`; atrasar jobs e ver o aviso.

---

### P1: Limites arquiteturais verificados ⭐ MVP

**User Story**: Como mantenedor, quero que o teste arquitetural impeça setas não permitidas, ciclos e lógica fora do lugar, para que o grafo da ADR-011 seja garantido por código.

**Why P1**: Exigência do mantenedor na ADR-011.

**Acceptance Criteria**:
1. The architecture test SHALL hold an explicit allow-list of module edges with their types (`D`, `E`, `R`) and SHALL fail on any import between modules not in the list.
2. The architecture test SHALL fail if a type `D` edge imports anything other than `<modulo>/app`, a type `E` edge anything other than `<modulo>/events`, or a type `R` edge anything other than `identity/app`.
3. The architecture test SHALL fail if a `<modulo>/events` package imports `app`, `infra`, `http` or another business module.
4. The architecture test SHALL fail if the allow-list or the set of actual imports contains a cycle.
5. The architecture test SHALL fail if a module directory has no entry in the allow-list.
6. The architecture test SHALL report, as a warning without failing, every allowed edge with no actual import.
7. The architecture test SHALL fail if `cmd/api` registers a job executor or an event subscription handler.

**Independent Test**: em pacotes de teste, criar um import não permitido, um ciclo e um `events` com lógica, e ver o teste falhar em cada caso.

---

## Edge Cases

- IF a job type is registered twice in the composition THEN the worker SHALL refuse to start.
- IF an event type has no subscriber THEN publishing it SHALL store the event and create no delivery job.
- WHEN the payload is exactly 64 KiB THEN the platform SHALL accept it.
- IF the clock of an instance drifts THEN leases and schedules SHALL still be correct because only the database clock is used.
- IF a consumer is subscribed after events were published THEN those earlier events SHALL NOT be delivered to it (no backfill in this spec).
- IF a re-enqueued job fails permanently again THEN it SHALL return to `FALHOU` with both attempt cycles in its history.

---

## Cenários de teste

| # | Cenário | Tipo | Requisitos |
|---|---|---|---|
| T-01 | Registrar job e confirmar / desfazer / fora de transação | Integração (PostgreSQL) | JOB-01 |
| T-02 | Prefixo de outro módulo → recusado | Unidade | JOB-01, OUT-01 |
| T-03 | 5 instâncias × 100 jobs → cada job uma vez | Integração (concorrência) | LEA-01 |
| T-04 | Instância pausada além da concessão → conclusão rejeitada (cerca) | Integração | LEA-01 |
| T-05 | Erros transitório, permanente e por tempo → estados e esperas corretos | Unidade com relógio do banco controlado | RTY-01 |
| T-06 | Matar a instância no meio do job → outra conclui | Integração | REC-01 |
| T-07 | Banco indisponível durante a execução → nenhum `FALHOU` por isso | Integração | REC-01 |
| T-08 | Publicar e confirmar / desfazer → jobs de entrega só no primeiro | Integração | OUT-01 |
| T-09 | Inscrição numa seta proibida → partida falha | Unidade (composição) | OUT-01 |
| T-10 | Mesmo evento 3 vezes → 1 efeito | Integração | INB-01 |
| T-11 | v2 antes de v1: `tolerante` → `obsoleto`; `estrita` → espera e depois aplica na ordem | Integração | INB-01 |
| T-12 | Conta de serviço inativa → adiamento sem gastar tentativas | Integração | SVX-01 |
| T-13 | Entrada de auditoria do efeito com ator `service`, `causation_id` e `correlation_id` | Integração | SVX-01 |
| T-14 | Reenfileirar e encerrar com e sem permissões; job fora de `FALHOU` → `conflict` | Integração | DLQ-01 |
| T-15 | 3 instâncias com agendamento → 1 execução por janela; janelas perdidas cobertas | Integração | SCH-01 |
| T-16 | Sem batimento → `nenhum_worker_ativo` | Integração | OBS-01 |
| T-17 | Import não permitido, ciclo, `events` com lógica, executor registrado na API → teste arquitetural falha | Teste de código-fonte | ARQ-01 |

---

## Dependências e riscos

| Tipo | Item | Mitigação |
|---|---|---|
| Dependência | Contas de serviço, classes e autoria da spec `02` | `02` especificada antes |
| Dependência | Tipos de job e de evento das specs `04`–`08` | Esta spec fixa o contrato que elas usam |
| Risco | Executores não idempotentes causam efeitos duplicados | Declaração obrigatória por tipo (REC-01); revisão; testes de reexecução em cada spec de módulo |
| Risco | Carga de polling no banco | Índice por estado e `executar_em`; intervalo configurável; volume baixo |
| Risco | Crescimento das tabelas técnicas | Retenção de 90 e 180 dias com limpeza por job agendado |
| Risco | O modo `estrita` trava um agregado se uma versão falhar de vez | O travamento fica visível como `FALHOU` e é resolvido pela recuperação humana |
| Risco | Um segundo binário a implantar | Mesma imagem e configuração da API, outro comando; o deploy é decisão posterior com autorização |

---

## Pendências que impedem o Design

**Nenhuma.** As decisões abaixo podem ser fechadas até as Tasks:

- **D-01-1**: quem recebe `jobs:job:read`, `jobs:job:retry` e `jobs:job:close`. Recomendação: `ADMIN_SISTEMA` para leitura e reenfileiramento técnico. O encerramento manual exige também a permissão de operação do tipo, que é do módulo dono.
- **D-01-2**: confirmar os valores padrão e a retenção.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
|---|---|---|---|
| JOB-01 | P1: Registro transacional de jobs | Specify | Pending |
| LEA-01 | P1: Captura, concessão e token de cerca | Specify | Pending |
| RTY-01 | P1: Tentativas, tempo máximo e falha definitiva | Specify | Pending |
| REC-01 | P1: Recuperação após falhas e reinícios | Specify | Pending |
| OUT-01 | P1: Outbox e entrega de eventos | Specify | Pending |
| INB-01 | P1: Consumidores idempotentes e eventos fora de ordem | Specify | Pending |
| SVX-01 | P1: Execução com conta de serviço e auditoria | Specify | Pending |
| DLQ-01 | P1: Recuperação humana de falhas | Specify | Pending |
| SCH-01 | P1: Agendamentos periódicos | Specify | Pending |
| OBS-01 | P1: Observabilidade e condições de operação | Specify | Pending |
| ARQ-01 | P1: Limites arquiteturais verificados | Specify | Pending |

**Coverage:** 11 total, 0 mapped to tasks, 11 unmapped (Design não iniciado).

## Success Criteria

- [ ] Nenhum cenário de queda (T-04, T-06, T-07) perde ou duplica efeito.
- [ ] Com 5 instâncias, nenhum job executa ao mesmo tempo em duas.
- [ ] Todo `FALHOU` é visível e só sai por ato humano auditado.
- [ ] O teste arquitetural falha em cada violação de T-17.
