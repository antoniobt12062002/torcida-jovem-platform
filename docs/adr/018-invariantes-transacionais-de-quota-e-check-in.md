# ADR-018: Invariantes transacionais de quota, check-in e histórico de frequência

- **Date**: 2026-10-13 (revisão 2, decisões M-1 a M-9); 2026-10-14 (revisão 3, parecer externo: projeção de presença efetiva; elegibilidade sob bloqueio; cenário 9)
- **Status**: Proposed. Aprovado conceitualmente; ainda não aceito; não autoriza implementação.
- **Deciders**: @antoniobt12062002
- **Tags**: concurrency, postgresql, integrity, tj-connect

## Context and Problem Statement

Três invariantes do TJ Connect precisam valer sob concorrência e falhas, garantidas pelo banco:

- **I-QUOTA**: em toda transição, ocupações + reservas válidas ≤ capacidade da temporada (TJ-CAP-001, TJ-CAP-003, §16).
- **I-CHECKIN**: no máximo um check-in efetivo por associado e jogo (TJ-JOG-006, AC-JOG-02).
- **I-FREQ**: os fatos de frequência são registrados desde o 1º jogo; só a **instauração** do procedimento depende de 6 jogos elegíveis concluídos (TJ-JOG-004; cuidado 2 e ajuste de planejamento do mantenedor).

## Decision Drivers

- O invariante é declarado no esquema (restrição ou `UPDATE` condicional sob bloqueio de linha), nunca só validado antes pela aplicação.
- Nenhuma chamada HTTP ao provedor dentro da transação que mexe na quota.
- A reserva não é liberada só pelo relógio enquanto o pagamento ainda puder ser aprovado.
- Os resultados são reproduzíveis e auditáveis.

## Considered Options (quota)

1. **Linha de quota por temporada** com contadores, `CHECK` e `UPDATE` condicional, na mesma transação da reserva ou adesão.
2. **Unidades de quota materializadas** (N linhas, captura com `SKIP LOCKED`).
3. **Contagem sob `SERIALIZABLE`** com repetição.

## Decision Outcome

### Mecanismo da quota: opção 1

- **Linha de quota** (dona: `associados`): `capacidade`, `ocupadas`, `reservadas`, com `CHECK (ocupadas >= 0 AND reservadas >= 0 AND ocupadas + reservadas <= capacidade)`.
- Toda operação é um `UPDATE` condicional nessa linha **na mesma transação** que cria ou altera a reserva ou a adesão:
  - reservar: `reservadas + 1 WHERE ocupadas + reservadas < capacidade`;
  - converter: `reservadas - 1, ocupadas + 1`;
  - liberar reserva: `reservadas - 1`;
  - ocupar tardio: `ocupadas + 1 WHERE ocupadas + reservadas < capacidade`;
  - liberar vaga: `ocupadas - 1`.

  Zero linhas afetadas significa recusa. O `CHECK` é a última barreira contra um bug de aplicação.
- **Restrições de unicidade**: uma reserva ativa por pessoa e temporada; uma adesão com ingresso **não encerrada** por pessoa e temporada; uma liberação de vaga por adesão.
- **Redução de capacidade** (`associados:quota:update`): `UPDATE capacidade = :nova`; o `CHECK` recusa valores abaixo dos compromissos.
- **Consistência**: um job compara os contadores com a contagem das linhas e abre uma pendência se houver diferença. Ele **nunca corrige sozinho**.
- O plano básico não toca a linha (AC-CAP-02). Intenção, ausência e suspensão não alteram contadores (TJ-CAP-002).
- A venda de Arquibancada exige liberação comercial vigente (M-5), verificada na mesma transação da reserva.

### Check-in, frequência e correção

Correção da revisão 3 (parecer externo): a revisão 2 previa um "índice único parcial sobre os check-ins aceitos e **não anulados**". Isso é incompatível com o histórico append-only, porque a anulação é **outra linha** e um índice parcial não enxerga outra tabela. Para "desligar" o check-in anulado do índice seria preciso alterar a linha original, o que violaria o append-only. A solução separa o histórico imutável de uma projeção transacional.

- **Histórico imutável** (append-only, sem `UPDATE` nem `DELETE`):
  - tentativas de leitura (fiscal, jogo, credencial, resultado, motivo, horário);
  - check-ins aceitos;
  - correções (anulação de um check-in aceito; presença reconhecida administrativamente).
- **Projeção de presença efetiva** (estado derivado, mutável, reconstruível a partir do histórico): uma linha por `(jogo_id, associado_id)`, com **chave primária** nesse par, o estado `EFETIVA` ou `ANULADA` e a referência ao fato que a define.
  - Check-in aceito: `INSERT` na projeção como `EFETIVA`, ou `UPDATE` de `ANULADA` para `EFETIVA`, **na mesma transação** que grava o fato no histórico. Se já existe `EFETIVA`, a operação recusa: "entrada já registrada".
  - Anulação: `UPDATE` da projeção de `EFETIVA` para `ANULADA`, na mesma transação do fato de anulação.
  - Presença reconhecida: `INSERT` ou `UPDATE` para `EFETIVA`, com origem "correção", na mesma transação do fato.
  - A **unicidade de presença efetiva** é garantida pela chave primária e pelo bloqueio da linha da projeção. O histórico nunca muda.
  - **Verificação de consistência**: um job reconstrói a projeção a partir do histórico e abre uma pendência se houver diferença. Ele nunca corrige sozinho.
- **Elegibilidade no momento da efetivação**, na mesma transação do check-in e **sob bloqueio compartilhado** (`FOR SHARE`) das linhas que podem revogá-la:
  - jogo publicado e elegível (linha do jogo);
  - designação ativa do fiscal (linha da designação);
  - fiscal ≠ titular (M-7);
  - adesão ativa com plano de ingresso, foto habilitada e benefício iniciado: linha da adesão em `associados`, lida pela consulta publicada que aplica o bloqueio compartilhado dentro da transação do chamador (`WithTx`);
  - nenhuma suspensão vigente: **linha de controle do benefício do associado** em `acesso` (uma por associado e temporada, com versão), que toda decisão de suspensão, revisão ou restabelecimento bloqueia com `FOR UPDATE` e atualiza.

  Toda operação que revoga elegibilidade (encerrar adesão, revogar foto, decidir suspensão, cancelar jogo, revogar designação) pega bloqueio **exclusivo** na mesma linha. O check-in e a revogação ficam serializados: um acontece inteiramente antes do outro (cenário 9).

  O **QR é dinâmico, de curta duração e específico do jogo** (decisão A-3), emitido e validado online pelo servidor. Um QR de outro jogo ou vencido é recusado (AC-JOG-04). Validade, assinatura e renovação ficam para o Design da spec `07`. Um QR válido **não** dispensa a unicidade do check-in (garantida pela projeção) nem a conferência visual da identidade pelo fiscal. Somente online (AC-JOG-03).
- **Correção** é sempre um fato novo e append-only no histórico, refletido na projeção na mesma transação, sem `UPDATE` nem `DELETE` no fato original.
- **Fatos de frequência desde o 1º jogo**:
  - elegibilidade do jogo congelada na conclusão (`acesso:jogo:conclude`), com correção posterior como fato novo auditado;
  - check-ins aceitos e correções (histórico), com a presença efetiva lida da projeção;
  - períodos de suspensão vigentes;
  - início do benefício de cada adesão (lido de `associados`).
- **Apuração**: função pura e determinística sobre os fatos, consultável a qualquer momento.
- **Instauração**: o caso de uso exige ≥ 6 jogos no denominador e presença < 50%, e grava a apuração usada. Nada suspende automaticamente.

## Demonstração por cenários

A notação `T1`/`T2` indica transações concorrentes em `READ COMMITTED`, o padrão do PostgreSQL. Todos os cenários viram testes de integração com PostgreSQL real (testcontainers), com goroutines concorrentes onde indicado.

### 1. Compra simultânea da última vaga

Estado: capacidade 60, ocupadas 59, reservadas 0.

1. `T1` e `T2` executam o `UPDATE … SET reservadas = reservadas + 1 WHERE … AND ocupadas + reservadas < capacidade`.
2. `T1` bloqueia a linha e atualiza (59 + 1 = 60). `T2` espera o bloqueio.
3. `T1` confirma. O PostgreSQL reavalia o `WHERE` de `T2` sobre a versão nova da linha: 59 + 1 < 60 é falso, e `T2` afeta 0 linhas.
4. `T2` desfaz a reserva que ia criar e responde "sem vaga". Nenhuma cobrança é criada para `T2`.

**Garantia**: uma única reserva. Se a aplicação omitisse o `WHERE`, o `CHECK` recusaria `T2`. **Teste**: 50 goroutines disputando 1 vaga → exatamente 1 sucesso; contadores = contagem das linhas.

### 2. Expiração da reserva e pagamento posterior

Reserva R ligada à cobrança C, validade alinhada à order (`expiration_time`) mais margem.

1. Na validade, o job de expiração consulta a order.
   - Pago → converte (cenário normal).
   - `created` → pede `cancel`.
2. Se o cancelamento for **confirmado**: numa transação, C → `NAO_CONCLUIDA`, evento `cobranca.nao_concluida`. O consumidor faz `reservadas - 1` e R `EXPIRADA`.
3. Se o cancelamento for recusado porque o pagamento entrou nesse meio tempo: nova consulta → pago → caminho normal de conversão. R ainda é válida.
4. Se, apesar disso, o provedor depois confirmar um pagamento de C já `NAO_CONCLUIDA` (por exemplo, um cartão em análise), C → `PAGA` com `tardio = true`, e o consumidor bloqueia a linha de quota:
   - com vaga: `ocupadas + 1` condicional → adesão `ATIVA`;
   - sem vaga: zero linhas → adesão `EM_REGULARIZACAO`, pendência de reembolso ou alternativa com concordância expressa. **Nenhum contador muda.**

**Garantia**: a reserva nunca é liberada enquanto o pagamento ainda pode ser aprovado sem consulta; o pagamento tardio nunca excede (AC-CAP-03). **Teste**: provedor falso com aprovação depois do cancelamento, com e sem vaga livre.

### 3. Notificações duplicadas e fora de ordem

1. Chegam N1 ("pagamento criado"), N3 ("aprovado") e N3 de novo, com N2 ("em análise") atrasada.
2. Cada notificação é gravada com unicidade; a segunda cópia de N3 é aceita (2xx) e descartada.
3. Cada notificação nova enfileira `cobranca:<id>:sync`. A chave de deduplicação aglutina tudo num job não terminal.
4. O job consulta a order **atual** (paga) e aplica `AGUARDANDO → PAGA` com `UPDATE` condicional. Emite `cobranca.paga` uma vez.
5. Quando N2 chega por último, o novo job consulta de novo (paga) e não há transição. Se o provedor dissesse "em análise" depois de pago, seria retrocesso → `DIVERGENTE`, nunca regressão.
6. O consumidor recebe `cobranca.paga` duas vezes (entrega pelo menos uma vez); a inbox `(consumidor, evento_id)` descarta a segunda.

**Garantia**: uma ativação e um lançamento (AC-PAG-02). **Teste**: sequências embaralhadas e duplicadas de notificações sobre o provedor falso.

### 4. Falha do worker após operação parcial

| Ponto da queda | Resultado |
|---|---|
| Depois de `POST /refund` no provedor, antes de gravar | Nova tentativa com a mesma `X-Idempotency-Key` → o provedor devolve o mesmo reembolso → grava |
| Dentro da transação "transição + lançamento + evento" | Rollback total; nova tentativa refaz tudo |
| Depois de confirmar a transação, antes de concluir o job | Nova tentativa → estado-alvo já atingido → conclui sem efeito |
| Worker pausado além da concessão | Outra instância recaptura (novo `lease_token`); a primeira, ao voltar, falha ao concluir porque o token não confere |

**Garantia**: nenhum efeito parcial e nenhum efeito duplo ([ADR-014](014-worker-e-processamento-persistente.md)). **Teste**: injeção de falha em cada ponto com um executor instrumentado.

### 5. Cancelamento definitivo e liberação da vaga

1. O associado solicita (`associados:cancelamento:request`): a adesão fica `CANCELAMENTO_SOLICITADO`; a quota não muda (AC-CAN-01).
2. A análise e a decisão (`analyze`, `decide`) são feitas por pessoas diferentes, se exigido. Se deferido: adesão `ENCERRADA` com data. O check-in seguinte já lê `ENCERRADA` e recusa; não há cache a invalidar.
3. Reembolso, se devido: solicitado, aprovado por outra pessoa e executado pela conta de serviço com `decision_id`.
4. **Liberação da vaga** (`associados:vaga:release`), numa transação, exige:
   - adesão `ENCERRADA`;
   - reembolso concluído ou declarado não devido;
   - substituição permitida pela liberação comercial vigente (acordo com o clube, P0-1).

   Executa `ocupadas - 1` e grava a liberação (única por adesão).
5. Uma liberação repetida viola a unicidade e não tem efeito. Uma reserva concorrente de outro comprador espera o bloqueio da linha e vê a vaga só depois da confirmação.

**Garantia**: a vaga só volta com cancelamento definitivo e condições cumpridas (TJ-ASSOC-003, AC-CAP-06). **Teste**: liberação concorrente duplicada; liberação concorrente com reserva.

### 6. Check-in simultâneo do mesmo associado por fiscais diferentes

1. F1 e F2, ambos designados para o jogo J e nenhum deles titular, leem a credencial de A ao mesmo tempo.
2. `T1` e `T2` passam pela elegibilidade e tentam gravar a presença efetiva `(J, A)` na projeção, junto com o fato "check-in aceito" no histórico.
3. A chave primária `(J, A)` da projeção faz `T2` esperar `T1`. Com `T1` confirmada, `T2` encontra a linha `EFETIVA`, recusa e **desfaz também o fato** que ia gravar no histórico.
4. `T2` grava uma **tentativa** "entrada já registrada" (numa transação nova) e responde com o horário e o fiscal de `T1`.

**Garantia**: uma presença efetiva e um único fato "check-in aceito" (AC-JOG-02). **Teste**: 10 goroutines com fiscais distintos → 1 aceito, 9 tentativas "já registrada", 1 linha `EFETIVA` e 1 fato aceito no histórico.

### 7. Correção auditada de check-in

1. Um check-in foi aceito por engano (por exemplo, credencial de outra pessoa). Quem tem `acesso:checkin:correct` registra uma **anulação**, com motivo e evidência. Não pode ser o titular nem o fiscal original.
2. Numa transação: o fato de anulação é acrescentado ao histórico (append-only, único por check-in aceito), e a projeção `(J, A)` passa de `EFETIVA` para `ANULADA` por `UPDATE` condicional (`WHERE estado = 'EFETIVA' AND fato = :checkin_anulado`). O fato original **não muda**. Um novo check-in legítimo do titular no mesmo jogo, se ainda ocorrer, leva a projeção de `ANULADA` a `EFETIVA` com o novo fato. Uma anulação e um novo check-in concorrentes se serializam na linha da projeção.
3. Uma presença reconhecida administrativamente (por exemplo, falha de rede comprovada na entrada) é um fato de correção do tipo "presença reconhecida", com motivo e evidência, nunca um check-in retroativo disfarçado. O MVP continua sem check-in offline (AC-JOG-03).
4. A apuração seguinte inclui as correções. Se um procedimento já instaurado usou a apuração anterior, ele recebe um aviso de fato superveniente para revisão; nada muda automaticamente.

**Garantia**: o histórico nunca é editado; a projeção sempre corresponde ao histórico (verificação de consistência); toda correção tem autor, motivo e evidência. **Teste**: anulação, nova apuração, unicidade da anulação, anulação concorrente com novo check-in, e reconstrução da projeção a partir do histórico igual à projeção corrente.

### 8. Frequência com jogos inelegíveis e suspensão

Associado A, benefício iniciado antes do J1. Jogos J1–J12:

| Jogo | Situação | Conta no denominador? | Check-in |
|---|---|---|---|
| J1, J2 | Em casa, elegível, concluído | Sim | J1 |
| J3 | Cancelado | Não | — |
| J4 | Elegível | Sim | J4 |
| J5 | Fora de casa (não elegível) | Não | — |
| J6, J7 | Elegíveis | Sim | — |
| J8, J9 | **Suspensão vigente** de A | Não (não são faltas) | — (check-in recusado) |
| J10, J11, J12 | Elegíveis | Sim | J10 |

Denominador = 8 (J1, J2, J4, J6, J7, J10, J11, J12); numerador = 3 → 37,5%. Com ≥ 6 e < 50%, **a instauração é possível**. Isso não é suspensão (AC-FRQ-02), e as intenções "vou" não contam.

Associado B, benefício iniciado no J7, mesmos jogos e nenhuma suspensão: denominador = J7, J10, J11, J12 = 4. **A instauração é recusada** (AC-FRQ-01), mas as 4 presenças ou ausências **já estão registradas** e entram nas apurações seguintes.

**Garantia**: a coleta começa no 1º jogo; a regra dos 6 só limita a instauração (AC-FRQ-03). **Teste**: tabela de casos da função de apuração, mais o teste do caso de uso de instauração.

### 9. Check-in concorrente com cancelamento ou suspensão do benefício

O fiscal F lê a credencial de A no jogo J (`T1`) no mesmo instante em que uma pessoa autorizada efetiva o encerramento da adesão de A, ou decide a suspensão dele (`T2`).

**Ordem a: `T2` bloqueia primeiro.**

1. `T2` pega o bloqueio exclusivo da linha da adesão (encerramento) ou da linha de controle do benefício de A (suspensão), muda o estado e confirma.
2. `T1` pede `FOR SHARE` na mesma linha e espera `T2`.
3. Liberado o bloqueio, `T1` lê a versão confirmada (`ENCERRADA` ou suspensão vigente).
4. `T1` recusa com "benefício não habilitado" (sem expor o motivo privado no ginásio, §16) e grava a tentativa.

**Ordem b: `T1` bloqueia primeiro.**

1. `T1` segura `FOR SHARE` nas linhas de elegibilidade, verifica tudo válido, grava o check-in e a projeção e confirma.
2. `T2` esperou o bloqueio e só então efetiva o encerramento ou a suspensão.
3. A entrada foi registrada **antes** de a revogação existir, o que é correto: a elegibilidade valia no momento da efetivação.
4. A auditoria mostra a ordem pelos horários de confirmação.

**Nunca**: `T1` confirmar o check-in com base numa leitura feita antes da confirmação de `T2`. Isso exigiria que `T1` lesse sem bloqueio e confirmasse depois de `T2`, e o `FOR SHARE` torna essa sequência impossível. O mesmo vale para revogação de foto, cancelamento do jogo e revogação da designação do fiscal, que pegam bloqueio exclusivo na linha correspondente.

**Garantia**: a elegibilidade é verdadeira no momento da efetivação do check-in (TJ-ASSOC-003 "invalidar prontamente"; TJ-JOG-005). **Teste**: goroutines com barreira e pausas controladas executando as duas ordens, repetido 200 vezes; para cada execução, se o check-in foi aceito, a confirmação de `T1` é anterior à de `T2`; se `T2` foi anterior, o check-in é recusado. Mais a mesma bateria para suspensão, revogação de foto e revogação de designação.

### Positive Consequences

- AC-CAP-01, AC-CAP-03..06, AC-PAG-02, AC-JOG-02..04, AC-FRQ-01..03 e AC-CAN-01 garantidos por restrição de banco ou por função pura testada.
- Overbooking impossível mesmo com bug de aplicação.

### Negative Consequences

- Contadores redundantes às linhas exigem o job de consistência.
- A projeção de presença efetiva duplica informação do histórico e exige reconstrução e verificação periódicas.
- O check-in segura bloqueios compartilhados em linhas de `associados` e `acesso`. Revogações concorrentes esperam alguns milissegundos; o volume de entrada no ginásio é baixo, mas o teste de pico (P2-19) deve medir isso.
- A linha de quota serializa as reservas de uma temporada; é aceitável em dezenas de vagas.
- A suíte de concorrência com PostgreSQL real aumenta o tempo do CI de integração.

## Pros and Cons of the Options (quota)

### 1. Contadores com `CHECK` e `UPDATE` condicional ✅ Chosen

- ✅ Invariante no esquema; disponibilidade O(1).
- ❌ Contador redundante.

### 2. Unidades materializadas

- ❌ Alterar a capacidade cria ou remove linhas; o histórico de reocupação complica.

### 3. `SERIALIZABLE`

- ❌ O invariante depende de todo caminho usar o isolamento correto.

## Links

- [ADR-011](011-mapa-de-modulos-do-tj-connect.md), [ADR-012](012-cobrancas-e-eventos-persistentes.md), [ADR-014](014-worker-e-processamento-persistente.md), [ADR-016](016-permissoes-sensiveis-e-escopo-por-jogo.md), [ADR-017](017-evolucao-aditiva-de-financeiro-e-estoque.md).
