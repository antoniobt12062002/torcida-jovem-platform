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
- [ ] A execução é **pelo menos uma vez**: um job pode ser executado fisicamente mais de uma vez e, depois de perder a concessão, uma instância antiga pode continuar executando em paralelo com a nova. A plataforma **não promete execução física única**; garante que só a instância com o token atual muda o estado do job e exige que todo efeito de negócio seja idempotente e condicional.
- [ ] Todo evento publicado é entregue a cada consumidor inscrito pelo menos uma vez, e o efeito do consumidor é **confirmado no máximo uma vez** por (consumidor, evento), pela inbox na mesma transação e pelas chaves duráveis do próprio módulo.
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
| Consultar jobs, tentativas, eventos e instâncias do worker (estado operacional) | Operador | `jobs:job:read` | sensivel (concessão explícita; não chega ao `PRESIDENTE` automaticamente) |
| Reenfileirar job com falha definitiva | Operador | `jobs:job:retry` | sensivel (reenfileirar só reexecuta trabalho idempotente) |
| Encerrar job manualmente | Operador | `jobs:job:close` **e** a permissão de encerramento declarada pelo módulo dono do tipo (ex. `cobrancas:processamento:close`) | sensivel; não pode silenciar obrigação pendente |

`jobs` é um pseudo-módulo de plataforma com `Contribution` própria, como o `audit` atual. **D-01-1 aprovada** (2026-10-14), ainda não aplicada no código: `ADMIN_SISTEMA` recebe `jobs:job:read` e `jobs:job:retry`; `PRESIDENTE` recebe `jobs:job:read` por concessão explícita. `jobs:job:close` não tem concessão aprovada, e o encerramento exige, além dela, a permissão de encerramento do módulo dono do tipo.

**Obrigações pendentes**: um tipo de job que represente uma obrigação (por exemplo, sincronizar pagamento, executar reembolso aprovado, liberar vaga, entregar evento financeiro ou associativo) declara-se **portador de obrigação** e fornece uma verificação do módulo dono. Encerrar manualmente um job desses **nunca** dá a obrigação por cumprida. O encerramento só é aceito se a verificação do módulo confirmar que a obrigação foi resolvida por outro caminho; caso contrário, é recusado.

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
- **INV-P2**: em qualquer instante, no máximo um token de concessão de um job é o **atual**. Isso não impede que uma instância com token antigo ainda esteja executando (ver "Limites da cerca").
- **INV-P3**: só a instância dona do token atual consegue mudar o estado de um job `EM_EXECUCAO`.
- **INV-P4**: para cada par (consumidor, evento), o efeito do consumidor é **confirmado** no máximo uma vez.
- **INV-P5**: nenhum job sai de `FALHOU` sem um ato humano auditado.
- **INV-P6**: nenhum código da API executa jobs nem entrega eventos.
- **INV-P7**: um módulo só registra jobs e publica eventos dos tipos com o seu prefixo.
- **INV-P8**: a cerca protege só as escritas da plataforma (estado do job, tentativas, concessão). Os efeitos de negócio são protegidos pelas escritas condicionais e pelas chaves únicas duráveis do módulo dono; as chamadas externas, pela chave de idempotência do sistema externo.
- **INV-P9**: o registro de inbox de um evento existe enquanto existir qualquer caminho capaz de entregar esse evento a esse consumidor (ver "Retenção e deduplicação durável").
- **INV-P10**: para todo agendamento declarado, existe a qualquer momento um job não terminal da janela atual ou da próxima, ou essa ausência é detectada e corrigida em até um intervalo de verificação.

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
3. O job de entrega abre a transação do consumidor, grava `(consumidor, evento_id)` na inbox, aplica a regra do modo de ordem (com a linha de versão bloqueada, nos modos `estrita` e `tolerante`) e chama o tratador do consumidor. Inbox, versão e efeito são confirmados juntos.
4. Se a inbox já contém o par → conclui sem chamar o tratador (duplicata).

### F6 — Ação humana sobre falha definitiva

1. O operador consulta os jobs `FALHOU` com tipo, tentativas e erros sanitizados.
2. **Reenfileirar**: exige as permissões, um motivo (≥ 10 caracteres) e gera auditoria. O job volta a `PENDENTE` com tentativas zeradas no novo ciclo; o histórico de tentativas fica preservado.
3. **Encerrar**: exige as permissões e um motivo; `ENCERRADO_MANUALMENTE`, com auditoria.

### F7 — Agendamentos periódicos

1. Cada agendamento declarado (tipo, intervalo, conta de serviço) mantém um job por janela, identificado pela chave única `<tipo>:<início da janela>`.
2. **Encadeamento atômico**: a transição do job da janela para `CONCLUIDO` ou `FALHOU` e o registro do job da próxima janela acontecem **na mesma transação** (condicionada ao token). Não existe instante em que uma janela terminou e a próxima ainda não foi registrada.
3. **Verificação independente**: na partida e a cada 60 s, toda instância executa "garantir agendamentos". Para cada agendamento declarado, se não houver job não terminal da janela atual ou da próxima, registra o da janela atual (inserção idempotente pela chave única) e emite aviso `agendamento_recriado`.
4. Depois de um período parado, a próxima execução cobre o atraso, porque o trabalho é definido por estado ("o que está pendente"), não por "o que mudou desde a última vez".

Os dois mecanismos são independentes: uma falha do encadeamento é coberta pela verificação, e vice-versa. Não há ponto único em que um agendamento desapareça em silêncio (INV-P10).

---

## Falhas e concorrência

| Situação | Comportamento exigido |
|---|---|
| Duas instâncias capturam ao mesmo tempo | Cada job é capturado por uma só (`SKIP LOCKED`) |
| Instância pausada além da concessão | Outra recaptura com novo token; a primeira não consegue mudar o estado do job (INV-P3) |
| Execução sobreposta depois da perda de concessão | As duas instâncias podem executar o mesmo executor ao mesmo tempo. A instância antiga, ao falhar a renovação, cancela o seu contexto (redução, não eliminação, da sobreposição). O efeito de negócio é confirmado no máximo uma vez por causa das escritas condicionais e chaves únicas do módulo (INV-P8); a chamada externa repetida usa a mesma chave de idempotência |
| Queda dentro da transação de efeito | Rollback; nova tentativa refaz o efeito inteiro |
| Queda depois do efeito e antes de concluir | Nova tentativa; o executor reconhece o efeito já aplicado (idempotência obrigatória do executor, declarada por tipo) |
| Queda depois de uma chamada externa e antes de gravar | Nova tentativa com a mesma chave de idempotência externa, gravada antes da chamada (contrato para os executores) |
| Banco indisponível | O worker para de capturar, registra o erro e tenta reconectar com espera; nada é marcado como falho por indisponibilidade do banco |
| Evento entregue duas vezes ao mesmo consumidor | A inbox impede um segundo efeito (INV-P4) |
| Eventos fora de ordem | Ver "Ordem de eventos" |
| Conta de serviço desativada | Jobs dos tipos dela ficam adiados, sem gastar tentativas, e geram alerta |
| Encerramento da instância (sinal de parada) | Para de capturar; espera os jobs em curso até o prazo de tolerância; o resto é recapturado quando a concessão vence |

### Limites da cerca (fencing)

- **O que a cerca garante**: toda escrita da plataforma sobre um job `EM_EXECUCAO` (concluir, falhar, renovar, registrar a próxima janela) é condicionada ao token atual. Uma instância com token antigo não altera o estado do job.
- **O que a cerca não garante**:
  - não impede que a instância antiga **continue executando** o código do executor;
  - não protege tabelas de módulos de negócio, porque o token não é verificado nelas;
  - não alcança **sistemas externos**: o provedor de pagamento não conhece o token e aceita a chamada de qualquer instância.
- **Contrato para todo executor**:
  1. As escritas de negócio são condicionais (estado e versão) ou protegidas por chave única durável. A repetição ou a concorrência resulta em "já aplicado", nunca em efeito duplo.
  2. Toda chamada externa com efeito usa uma chave de idempotência **gravada antes da chamada** e reutilizada em toda repetição. Em caso de dúvida, o executor consulta o estado atual no sistema externo.
  3. O executor observa o cancelamento do seu contexto, que a plataforma aciona quando a renovação da concessão falha.

### Retenção e deduplicação durável

- **Outbox**: um evento só pode ser removido quando **todas** as suas entregas estão `CONCLUIDO` ou `ENCERRADO_MANUALMENTE` e a mais recente terminou há mais que o prazo de retenção. Uma entrega em `PENDENTE`, `EM_EXECUCAO` ou `FALHOU` **impede** a remoção, por tempo indeterminado.
- **Inbox**: o registro (consumidor, evento) só pode ser removido **depois** que o próprio evento foi removido. Como não existe caminho de entrega sem o evento, nenhuma entrega pode ocorrer depois da remoção do registro de inbox (INV-P9).
- **Reprocessamento tardio**: reenfileirar um `FALHOU` anos depois continua seguro, porque o evento e a sua inbox não foram removidos.
- **Sem replay**: a plataforma não oferece reentrega de entregas já concluídas. Uma futura função de replay exigiria nova decisão e se apoiaria só nas chaves duráveis dos módulos.
- **Garantia de negócio independente da inbox**: todo consumidor que produz efeito financeiro ou associativo mantém a sua **própria chave única durável**, sem expiração (por exemplo, `chave_idempotencia` do lançamento, ativação única por cobrança). A inbox é a primeira barreira, não a única, e a sua limpeza nunca enfraquece essas garantias.

### Envelope de evento

`evento_id` (único), `tipo` com versão (`<modulo>.<fato>.v<n>`), `agregado_tipo`, `agregado_id`, `agregado_versao` (inteiro crescente por agregado, atribuído pelo produtor na mesma transação), `ocorrido_em`, `correlation_id`, `causation_id`, e `payload` com o **estado resultante** da transição (não um delta), sem dado pessoal desnecessário.

### Ordem de eventos

A plataforma **não garante ordem** de entrega, nem global nem por agregado. Cada inscrição declara um modo, e a spec do consumidor justifica a escolha:

- **`independente`**: cada evento é aplicado uma vez (inbox), sem controle de versão. Só é permitido quando os efeitos do consumidor são comutativos, isto é, o resultado final não depende da ordem.
- **`estrita`**: um evento de versão `v` só é aplicado se a última versão aplicada por aquele consumidor para aquele agregado é exatamente `v - 1`. Caso contrário, a entrega volta a `PENDENTE` como erro transitório até as tentativas acabarem; aí vira `FALHOU`, visível. É o modo obrigatório quando **cada transição** gera um efeito próprio (por exemplo, registrar um reembolso ou ativar uma adesão).
- **`tolerante`**: um evento de versão menor ou igual à última aplicada é concluído **sem chamar o tratador** e registrado como `obsoleto`. Só é permitido quando aplicar apenas o **estado resultante mais recente** preserva todos os efeitos necessários, isto é, o consumidor é convergente e nenhuma transição intermediária tem efeito próprio. A spec do consumidor precisa declarar isso, e a revisão da spec o confirma.

**Controle transacional da versão aplicada** (modos `estrita` e `tolerante`): a última versão aplicada fica numa linha por (consumidor, agregado), lida com bloqueio e atualizada **na mesma transação** do efeito e do registro de inbox, por escrita condicional (`tolerante`: só se a nova versão for maior; `estrita`: só se a atual for `v - 1`). Duas entregas simultâneas do mesmo agregado se serializam nessa linha. Uma versão menor que chega depois de uma maior confirmada é marcada `obsoleto` (`tolerante`) ou aguarda (`estrita`), e nunca há atualização perdida nem aplicação fora de ordem.

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
| Valores padrão (configuráveis por ambiente) | Polling de 1 s; concessão de 60 s; tempo máximo padrão de 45 s; até 10 tentativas; espera exponencial a partir de 5 s, até 15 min, com variação aleatória de ±20%; prazo de tolerância no encerramento de 30 s | Volume baixo (centenas de associados); a concessão é maior que o tempo máximo. **D-01-2 aprovada**: valores iniciais configuráveis | y |
| Retenção | Prazos configuráveis (padrão de 90 dias) e **condicionados** às regras de "Retenção e deduplicação durável": evento só sai com todas as entregas concluídas ou encerradas; inbox só sai depois do evento; `FALHOU` impede a remoção; `FALHOU` e `ENCERRADO_MANUALMENTE` não expiram | A deduplicação técnica nunca expira antes de todo caminho de entrega; as garantias de negócio ficam nas chaves duráveis dos módulos. **D-01-2 aprovada com condição**: a retenção só pode ser usada se a deduplicação de negócio continuar garantida depois da limpeza (RET-01 AC5) | y |
| Observabilidade | Logs estruturados com campos fixos, mais consultas de operação na API (sem stack de métricas externa) | Não há infraestrutura de métricas no projeto hoje | n |
| Batimento das instâncias | Cada instância grava um batimento a cada 15 s; sem batimento há mais de 60 s, a instância conta como inativa | Permite ver "nenhum worker ativo". **D-01-2 aprovada**: valores iniciais de monitoramento configuráveis | y |
| Ordem de eventos | Sem garantia; modos `independente`, `estrita` e `tolerante` por inscrição, com versão aplicada controlada na transação; `tolerante` só para consumidores convergentes declarados | Parecer externo: o modo tolerante não pode perder efeitos de transições intermediárias | y |
| Execução física | Pelo menos uma vez, com sobreposição possível depois da perda de concessão | Parecer externo: a plataforma não promete execução única; a correção vem da idempotência dos efeitos | y |
| Fila própria ou biblioteca | Própria (ADR-014), salvo prova no Design | Enfileirar na transação GORM é requisito | y |
| Quem recebe `jobs:job:read`, `jobs:job:retry`, `jobs:job:close` | `ADMIN_SISTEMA`: `read` e `retry`; `PRESIDENTE`: `read` (explícita); `close`: ninguém até decisão | **D-01-1 aprovada** em 2026-10-14; concessões ainda não aplicadas no código | y |
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
5. WHEN a job is registered with a deduplication key equal to that of a job of the same type in `PENDENTE` or `EM_EXECUCAO` THEN the platform SHALL NOT create a new job and SHALL return the id of the existing one.
6. IF a job payload exceeds 64 KiB THEN the platform SHALL refuse the job with an error.
7. The API process SHALL register no job executor and SHALL run no job.

**Independent Test**: num teste de integração, registrar um job e confirmar, registrar outro e desfazer, registrar fora de transação, e ver só o primeiro existir.

---

### P1: Captura, concessão e token de cerca ⭐ MVP

**User Story**: Como operador, quero que várias instâncias do worker rodem juntas sem capturar o mesmo job ao mesmo tempo e sem que uma instância antiga sobrescreva o estado registrado por outra, sabendo que a execução física pode se repetir e que a correção vem da idempotência dos efeitos.

**Why P1**: Garante INV-P2, INV-P3 e INV-P8.

**Acceptance Criteria**:
1. WHEN several worker instances poll at the same time THEN a given ready job SHALL be claimed by at most one of those polls.
2. WHEN a job is claimed THEN the platform SHALL assign it a new lease token and a lease expiry computed from the database clock.
3. IF an instance tries to complete, fail or renew a job with a lease token that is no longer current THEN the platform SHALL reject the change and SHALL record the attempt as `concessao_perdida`.
4. WHILE a job is running longer than half of its lease the owning instance SHALL renew the lease using its current token.
5. The platform SHALL compare every scheduling and lease time against the database clock, never against the instance clock.
6. IF a lease renewal is rejected because the token is no longer current THEN the instance SHALL cancel the execution context of that job.
7. The platform SHALL NOT promise single physical execution, and every job type SHALL be accepted only with an executor whose business writes are conditional or protected by a durable unique key.
8. WHEN an instance whose lease expired and a new owner execute the same job concurrently THEN the business effect SHALL be committed at most once and only the current owner SHALL change the job state.
9. WHEN an executor performs an external call with side effects THEN it SHALL use an idempotency key persisted before the first call and reused on every repetition.

**Independent Test**: 5 instâncias e 100 jobs → nenhuma captura simultânea e cada job concluído uma vez; pausar a instância A no meio do executor além da concessão, deixar B executar e concluir, retomar A e ver um único efeito de negócio, a conclusão de A rejeitada e a tentativa registrada como `concessao_perdida`.

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
3. The platform SHALL require every subscription to declare exactly one mode among `independente`, `estrita` and `tolerante`.
4. WHERE a subscription uses mode `estrita` or `tolerante` the platform SHALL read the last applied version of (consumer, aggregate) with a row lock and SHALL update it with a conditional write in the same transaction as the effect and the inbox pair.
5. WHERE a subscription uses mode `tolerante` the platform SHALL complete without calling the handler, and SHALL record the delivery as `obsoleto`, when the event's `agregado_versao` is lower than or equal to the last applied version.
6. WHERE a subscription uses mode `estrita` the platform SHALL treat a delivery as a transient error when the last applied version is not exactly `agregado_versao - 1`.
7. WHEN deliveries of two different versions of the same aggregate run concurrently for the same consumer THEN the last applied version SHALL end equal to the highest applied version and no lower version SHALL be applied after a higher one has committed.
8. IF a subscription declares mode `tolerante` THEN its consumer spec SHALL state that applying only the most recent resulting state preserves every required effect, and a subscription whose consumer has a distinct effect per transition SHALL use mode `estrita` or `independente`.
9. WHEN a consumer handler fails THEN the inbox pair and the version row SHALL NOT be written and the delivery SHALL follow the job retry rules.

**Independent Test**: entregar o mesmo evento 3 vezes (1 efeito); entregar v2 antes de v1 nos modos estrito e tolerante; entregar v2 e v3 **simultaneamente** (goroutines com barreira) e ver, no tolerante, versão final 3 e v2 nunca aplicada depois de v3; no estrito, v3 aplicada só depois de v2.

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
2. WHEN a person holding `jobs:job:retry` re-enqueues a job in `FALHOU` with a reason of at least 10 characters and the job type's reprocessing policy allows it THEN the job SHALL return to `PENDENTE` with a new attempt cycle and its previous attempts kept.
3. WHEN a person holding `jobs:job:close` and the close permission declared by the module that owns the job type closes a job in `FALHOU` with a reason of at least 10 characters THEN the job SHALL move to `ENCERRADO_MANUALMENTE`.
4. IF re-enqueue or close is requested for a job not in `FALHOU` THEN the platform SHALL refuse it with `conflict`.
5. IF the person lacks either required permission THEN the platform SHALL refuse with `forbidden` and SHALL write a `denied` audit entry.
6. The platform SHALL require every job type to declare whether it carries an obligation and, IF it does, to provide a resolution check owned by its module.
7. IF a person tries to close a job whose type carries an obligation and the module's resolution check does not confirm that the obligation was resolved by another path THEN the platform SHALL refuse the close with code `obrigacao_pendente`.
8. The platform SHALL NOT treat the manual close of a job as the fulfilment of the obligation it carried.
9. The platform SHALL require every job type with a financial or associative effect to declare a reprocessing policy owned by its module that checks idempotency, the obligation's validity and any additional permission, and IF the policy refuses THEN the re-enqueue SHALL be refused with the policy's code.

**Independent Test**: provocar um `FALHOU`, reenfileirar e ver o novo ciclo com histórico preservado; tentar sem a permissão do tipo e ver a recusa.

---

### P1: Agendamentos periódicos ⭐ MVP

**User Story**: Como módulo, quero declarar uma tarefa periódica (como conciliação ou expiração) que rode uma vez por janela mesmo com várias instâncias, e que recupere janelas perdidas, para não depender de cron externo.

**Why P1**: Conciliação e expiração das specs `04`/`05`.

**Acceptance Criteria**:
1. WHEN the worker starts THEN it SHALL ensure one non-terminal job exists for the current window of every declared schedule.
2. The platform SHALL guarantee at most one job per schedule and window, identified by a unique key made of the schedule type and the window start.
3. WHEN a scheduled job moves to `CONCLUIDO` or `FALHOU` THEN the platform SHALL register the job for the next window in the same transaction as that state change.
4. IF the transaction that completes or fails a scheduled job rolls back THEN neither the state change nor the next-window job SHALL exist, and the job SHALL stay recoverable by the lease rules.
5. WHILE any worker instance runs it SHALL check every 60 seconds that each declared schedule has a non-terminal job for the current or next window, SHALL register the current window when none exists, and SHALL emit a structured log `agendamento_recriado` at level `warn`.
6. WHEN the worker was stopped for several windows THEN the platform SHALL register only the current window, and the executor SHALL process all pending state in that run.

**Independent Test**: 3 instâncias com um agendamento de 1 minuto por 5 minutos → 5 execuções; derrubar a instância no momento da conclusão e ver a próxima janela existir ou ser recriada em até 60 s; remover à força o job da próxima janela e ver `agendamento_recriado`; parar tudo por 3 janelas e ver uma execução que cobre o atraso.

---

### P1: Retenção sem perda de deduplicação ⭐ MVP

**User Story**: Como mantenedor, quero limpar dados técnicos antigos sem nunca permitir que um evento seja aplicado duas vezes, nem mesmo quando um job com falha for reenfileirado muito tempo depois, para manter o banco enxuto sem enfraquecer a idempotência.

**Why P1**: Parecer externo: a retenção de 90/180 dias conflitava com o reprocessamento indefinido de falhas.

**Acceptance Criteria**:
1. IF any delivery of an event is in `PENDENTE`, `EM_EXECUCAO` or `FALHOU` THEN the cleanup SHALL NOT remove that event nor its inbox records.
2. WHEN every delivery of an event is `CONCLUIDO` or `ENCERRADO_MANUALMENTE` and the latest ended more than the configured retention ago THEN the cleanup MAY remove the event and SHALL remove its inbox records only in the same transaction or after it.
3. The platform SHALL NOT offer re-delivery of a delivery that is already `CONCLUIDO`.
4. WHEN a job in `FALHOU` is re-enqueued at any age THEN its event and inbox records SHALL still exist and the delivery SHALL be deduplicated.
5. The platform SHALL require every consumer with a financial or associative effect to declare the durable business unique key that makes its effect idempotent independently of the inbox.

**Independent Test**: deixar uma entrega em `FALHOU`, avançar o relógio do banco além da retenção, rodar a limpeza e ver evento e inbox preservados; reenfileirar e ver um único efeito; concluir todas as entregas, avançar o relógio e ver evento e inbox removidos juntos.

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
- IF an old instance keeps running after losing its lease and calls the payment provider THEN the call SHALL reuse the persisted idempotency key and the provider SHALL return the same result as for the current owner.
- WHEN the cleanup and a re-enqueue of the same event run concurrently THEN the re-enqueue SHALL win (the cleanup re-checks delivery states under lock) and the event SHALL NOT be removed.

---

## Cenários de teste

| # | Cenário | Tipo | Requisitos |
|---|---|---|---|
| T-01 | Registrar job e confirmar / desfazer / fora de transação | Integração (PostgreSQL) | JOB-01 |
| T-02 | Prefixo de outro módulo → recusado | Unidade | JOB-01, OUT-01 |
| T-03 | 5 instâncias × 100 jobs → nenhuma captura simultânea; cada job concluído uma vez | Integração (concorrência) | LEA-01 |
| T-04 | Instância pausada além da concessão → conclusão rejeitada (cerca) | Integração | LEA-01 |
| T-04b | Execução sobreposta: A pausada no meio do executor, B executa e conclui, A retoma → um efeito de negócio, conclusão de A rejeitada, contexto de A cancelado na renovação | Integração (concorrência) | LEA-01 |
| T-04c | Chamada externa repetida por A e B → mesma chave de idempotência; provedor falso registra um efeito | Integração | LEA-01 |
| T-05 | Erros transitório, permanente e por tempo → estados e esperas corretos | Unidade com relógio do banco controlado | RTY-01 |
| T-06 | Matar a instância no meio do job → outra conclui | Integração | REC-01 |
| T-07 | Banco indisponível durante a execução → nenhum `FALHOU` por isso | Integração | REC-01 |
| T-08 | Publicar e confirmar / desfazer → jobs de entrega só no primeiro | Integração | OUT-01 |
| T-09 | Inscrição numa seta proibida → partida falha | Unidade (composição) | OUT-01 |
| T-10 | Mesmo evento 3 vezes → 1 efeito | Integração | INB-01 |
| T-11 | v2 antes de v1: `tolerante` → `obsoleto`; `estrita` → espera e depois aplica na ordem | Integração | INB-01 |
| T-11b | v2 e v3 entregues **simultaneamente** ao mesmo consumidor (barreira entre goroutines), repetido 200 vezes: `tolerante` termina sempre em versão 3 sem aplicar v2 depois de v3; `estrita` aplica v2 antes de v3 | Integração (concorrência) | INB-01 |
| T-11c | Inscrição `tolerante` sem declaração de consumidor convergente → revisão da spec do consumidor recusa | Revisão de spec (checklist) | INB-01 |
| T-11d | Entrega em `FALHOU` além da retenção → limpeza preserva evento e inbox; reenfileirar → um efeito | Integração (relógio do banco controlado) | RET-01 |
| T-11e | Limpeza concorrente com reenfileiramento do mesmo evento → evento preservado | Integração (concorrência) | RET-01 |
| T-12 | Conta de serviço inativa → adiamento sem gastar tentativas | Integração | SVX-01 |
| T-13 | Entrada de auditoria do efeito com ator `service`, `causation_id` e `correlation_id` | Integração | SVX-01 |
| T-14 | Reenfileirar e encerrar com e sem permissões; job fora de `FALHOU` → `conflict` | Integração | DLQ-01 |
| T-14b | Encerrar job portador de obrigação (ex. reembolso aprovado) sem resolução confirmada pelo módulo → `obrigacao_pendente`; com resolução confirmada → encerrado, e a obrigação não é marcada como cumprida pelo encerramento | Integração | DLQ-01 |
| T-15 | 3 instâncias com agendamento → 1 execução por janela; janelas perdidas cobertas | Integração | SCH-01 |
| T-15b | Falha injetada na transação de conclusão de um job agendado → nem conclusão nem próxima janela; job recuperado pela concessão | Integração | SCH-01 |
| T-15c | Job da próxima janela removido à força → recriado em até 60 s com `agendamento_recriado` | Integração | SCH-01 |
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
| Risco | Crescimento das tabelas técnicas | Limpeza agendada condicionada à deduplicação durável (RET-01); `FALHOU` cresce só com incidentes reais, que ficam visíveis |
| Risco | Executor de módulo sem escrita condicional causa efeito duplo em execução sobreposta | LEA-01 AC7 recusa o registro do tipo sem a declaração; cada spec de módulo inclui o teste de execução sobreposta |
| Risco | O modo `estrita` trava um agregado se uma versão falhar de vez | O travamento fica visível como `FALHOU` e é resolvido pela recuperação humana |
| Risco | Um segundo binário a implantar | Mesma imagem e configuração da API, outro comando; o deploy é decisão posterior com autorização |

---

## Pendências que impedem o Design

**Nenhuma.** As decisões abaixo podem ser fechadas até as Tasks:

- **D-01-1 — aprovada** (2026-10-14): `ADMIN_SISTEMA` consulta e reenfileira; `PRESIDENTE` consulta o estado operacional por concessão explícita. O encerramento exige autorização adicional do módulo dono e não pode silenciar obrigações pendentes (DLQ-01 AC6–8). Quem recebe `jobs:job:close` continua sem decisão, o que não afeta o Design.
- **D-01-2 — aprovada com condição** (2026-10-14): valores iniciais configuráveis; retenção só com a deduplicação de negócio garantida (RET-01).

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
| RET-01 | P1: Retenção sem perda de deduplicação | Specify | Pending |
| OBS-01 | P1: Observabilidade e condições de operação | Specify | Pending |
| ARQ-01 | P1: Limites arquiteturais verificados | Specify | Pending |

**Coverage:** 12 total, 0 mapped to tasks, 12 unmapped (Design não iniciado).

## Success Criteria

- [ ] Nenhum cenário de queda ou sobreposição (T-04, T-04b, T-04c, T-06, T-07) perde ou duplica efeito de negócio.
- [ ] Com 5 instâncias, nenhum job é capturado simultaneamente por duas; a execução sobreposta depois da perda de concessão nunca confirma efeito duplo.
- [ ] Nenhuma limpeza técnica permite reaplicar um evento (T-11d, T-11e), e nenhum agendamento desaparece em silêncio (T-15b, T-15c).
- [ ] Todo `FALHOU` é visível e só sai por ato humano auditado.
- [ ] O teste arquitetural falha em cada violação de T-17.
