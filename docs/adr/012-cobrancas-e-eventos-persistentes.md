# ADR-012: Módulo `cobrancas`, ciclo de vida da cobrança, eventos persistentes e conciliação

- **Date**: 2026-10-13 (revisão 2, com as decisões M-1 a M-9 do mantenedor)
- **Status**: Proposed. Aprovado conceitualmente; ainda não aceito; não autoriza implementação.
- **Deciders**: @antoniobt12062002
- **Tags**: architecture, payments, outbox, idempotency, tj-connect

## Context and Problem Statement

O TJ Connect cobra planos pelo Mercado Pago, **Checkout Pro via Orders API** (decisão M-2), com Pix e cartão em até 3x (TJ-PAG-001). Os requisitos exigem:

- o retorno do navegador não prova pagamento (TJ-PAG-004, AC-PAG-01);
- um webhook repetido não ativa duas vezes (AC-PAG-02);
- pagamento tardio sem vaga não ativa ingresso (TJ-PAG-003);
- conciliação recuperável após webhook perdido ou API indisponível.

A [ADR-011](011-mapa-de-modulos-do-tj-connect.md) cria `cobrancas` sem import de `associados`.

### Fatos do provedor

**Verificados** na documentação oficial do Mercado Pago (Context7 e páginas `checkout-pro-orders`, 2026-10-13):

- Checkout Pro via Orders: `POST /v1/orders` com `type: online` e `processing_mode: manual`; header `X-Idempotency-Key` obrigatório, um por tentativa lógica; `external_reference`; `expiration_time` (duração ISO 8601, ex. `P1D`); `config` com restrições de meios de pagamento; resposta com `checkout_url` e status inicial `created`.
- Orders API: consulta `GET /v1/orders/{id}`; cancelamento `POST /v1/orders/{id}/cancel` só com status `created`; reembolso `POST /v1/orders/{id}/refund` com chave de idempotência, sobre order `processed`, e o reembolso pode ficar `processing`.
- Webhook: header `x-signature` (`ts`, `v1`) e `x-request-id`; HMAC-SHA256 do manifesto `id:[data.id];request-id:[x-request-id];ts:[ts];`, comparado em tempo constante.
- No Checkout Pro com preferências (alternativa), `payment_methods.installments` limita parcelas e `excluded_payment_types` exclui categorias.

**Não verificados** (critério de aceite da sub-spec `04` e verificação obrigatória no Design, em sandbox):

1. Pix e cartão disponíveis no Checkout Pro via Orders, e exclusão de qualquer outro meio (boleto, saldo em conta, débito etc.).
2. Limite de 3 parcelas configurável na order. **Suporte documental confirmado pelo mantenedor** (`config.payment_method.max_installments`, 2026-10-13); falta comprovar na integração.
3. Parcelamento com juros pagos pelo comprador, com total e encargos exibidos antes da confirmação.
4. Distinção, nos dados da order ou do pagamento, entre juros do comprador e taxa de processamento do recebedor.
5. Vigência: efeito de `expiration_time` sobre um Pix já gerado e sobre um cartão em análise.
6. Tópico e formato do webhook de orders; lista completa de status e `status_detail`.
7. Reembolso parcial (a Orders API documenta reembolso total para alguns produtos).
8. Contestação (chargeback) refletida na order.

**Regra de escolha (M-2)**: se a verificação confirmar 1 a 6, adota-se a Orders API. Se algum desses itens falhar, a alternativa é o **Checkout Pro com preferências** (`/checkout/preferences`, com `installments` e `excluded_payment_types` verificados), atrás da mesma porta e com o mesmo ciclo de vida interno. A troca é decisão registrada, não improviso. Nenhum meio de pagamento além de Pix e cartão de crédito é habilitado.

## Decision Drivers

- Nenhum dinheiro e nenhuma ativação sem confirmação consultada no provedor pelo servidor.
- Exatamente um efeito por fato gerador, com repetição, reinício e entrega fora de ordem.
- Nenhuma transação de banco aberta durante uma chamada HTTP ao provedor.
- Sem funcionalidade comercial da loja (carrinho, cupom, recorrência).

## Considered Options

1. **Outbox transacional, consumidores idempotentes e conciliação ativa**, com o provedor atrás de uma porta.
2. **Callback síncrono** de `cobrancas` para uma interface de `associados` dentro da transação do webhook.
3. **Polling de `associados`** sobre o estado das cobranças.

## Decision Outcome

Chosen option: **1**.

### Ciclo de vida da cobrança (interno, independente do provedor)

| Estado | Significado | Terminal? |
|---|---|---|
| `CRIADA` | Registro local gravado; a order ainda não existe no provedor | Não |
| `AGUARDANDO_PAGAMENTO` | Order criada (`created`), `checkout_url` disponível, dentro da validade | Não |
| `EM_PROCESSAMENTO` | Há um pagamento em análise ou pendente no provedor | Não |
| `PAGA` | O provedor confirmou o pagamento integral; valor, moeda e `external_reference` conferidos | Não (pode ir a reembolso ou contestação) |
| `NAO_CONCLUIDA` | Expirou, foi recusada ou cancelada **com confirmação do provedor** de que não há pagamento possível | Sim |
| `FALHA_CRIACAO` | A order não pôde ser criada depois das tentativas com a mesma chave de idempotência | Sim |
| `REEMBOLSO_EM_ANDAMENTO` | Reembolso solicitado ao provedor, aguardando conclusão | Não |
| `REEMBOLSADA` | O provedor confirmou o reembolso do valor devolvido | Sim |
| `CONTESTADA` | Chargeback aberto | Não |
| `CONTESTACAO_PERDIDA` | O dinheiro voltou ao comprador por contestação | Sim |
| `DIVERGENTE` | O estado do provedor contradiz o local ou é inconsistente (ver tabela de divergências); exige resolução humana | Não (sai por decisão auditada) |

Transições permitidas:

`CRIADA → AGUARDANDO_PAGAMENTO | FALHA_CRIACAO`; `AGUARDANDO_PAGAMENTO → EM_PROCESSAMENTO | PAGA | NAO_CONCLUIDA`; `EM_PROCESSAMENTO → PAGA | NAO_CONCLUIDA`; `NAO_CONCLUIDA → PAGA` (pagamento tardio confirmado pelo provedor, marcado como `tardio`); `PAGA → REEMBOLSO_EM_ANDAMENTO | CONTESTADA`; `REEMBOLSO_EM_ANDAMENTO → REEMBOLSADA | PAGA` (reembolso recusado pelo provedor); `CONTESTADA → PAGA | CONTESTACAO_PERDIDA`. Qualquer estado pode ir a `DIVERGENTE`.

Regras:

- Toda transição é um `UPDATE` com comparação do estado atual (`WHERE estado IN (origens permitidas)`) e um contador de versão. Aplicar duas vezes a mesma verdade não muda nada.
- Um retrocesso informado pelo provedor (por exemplo, `processed` e depois pendente) nunca é aplicado: vira `DIVERGENTE`.
- **Reembolso parcial** (decisão A-1): o domínio representa reembolsos integrais e parciais, cada um com valor, motivo, aprovação e rastreabilidade. A soma dos reembolsos aprovados e concluídos nunca excede o valor elegível, o que é garantido no banco. **Nenhuma regra automática** calcula ou processa reembolso parcial sem validação jurídica e sem a confirmação do suporte do Mercado Pago (verificação 7). Até lá, a execução no provedor de um valor parcial fica bloqueada.
- **Contestação** (decisão A-2): `CONTESTADA` (pendente) e `CONTESTACAO_PERDIDA` (pagamento definitivamente revertido) são estados distintos. A contestação não suspende o benefício nem altera a adesão automaticamente. Andamento e desfecho financeiro ficam registrados; qualquer consequência associativa segue procedimento administrativo próprio.

### Relação com a adesão (dono: `associados`)

A cobrança não conhece adesão; `associados` reage aos eventos:

| Evento de `cobrancas` | Efeito em `associados` (detalhado na [ADR-018](018-invariantes-transacionais-de-quota-e-check-in.md)) |
|---|---|
| `cobranca.paga.v1` (não tardia) | Reserva válida → ocupação; adesão `ATIVA`. Plano básico: adesão `ATIVA`, sem quota |
| `cobranca.paga.v1` (tardia) | Revalida a quota sob bloqueio: com vaga, `ATIVA`; sem vaga, `EM_REGULARIZACAO` e pendência (nunca excede, nunca converte em outro plano) |
| `cobranca.nao_concluida.v1` | Adesão `NAO_CONCLUIDA`; a reserva é liberada (a liberação só acontece depois desta confirmação) |
| `cobranca.reembolsada.v1` | Registra o reembolso no cancelamento ou na regularização; encerra a adesão se a decisão previu isso |
| `cobranca.contestada.v1` / `cobranca.contestacao_perdida.v1` | Registra a contestação, o andamento e o desfecho, e abre pendência administrativa. **Nenhum efeito automático** sobre a adesão ou o ingresso (decisão A-2); qualquer consequência segue procedimento administrativo |
| `cobranca.divergente.v1` | Nenhuma ativação; pendência visível |

- Uma adesão aponta para **uma** cobrança de origem. A adesão nunca é editada para outro plano.
- O histórico é uma sequência de adesões por pessoa e temporada. Uma mudança de plano futura (M-6, fora do MVP) poderá ser modelada como nova adesão ligada à anterior, sem regra comercial antecipada agora.

### Garantias de idempotência (por camada)

| Camada | Mecanismo |
|---|---|
| Chamada ao provedor (criar, cancelar, reembolsar) | `X-Idempotency-Key` determinística (`cobranca:<id>:<operação>:<n>`), **gravada antes da chamada**; uma nova tentativa reutiliza a mesma chave |
| Notificação recebida | Unicidade por identificador da notificação do provedor; duplicata é aceita (2xx) e descartada |
| Sincronização | Um job por order com chave de deduplicação `cobranca:<id>:sync`; notificações em rajada se aglutinam |
| Verdade do provedor | O job **sempre consulta** `GET /v1/orders/{id}`; o payload do webhook nunca é aplicado; a ordem de chegada das notificações é irrelevante |
| Transição | `UPDATE` condicional por estado e versão |
| Efeito financeiro | `chave_idempotencia` única no lançamento ([ADR-017](017-evolucao-aditiva-de-financeiro-e-estoque.md)) |
| Evento | Gravado na mesma transação da transição, que acontece uma única vez; consumidores deduplicam pelo id (inbox) |

### Eventos persistentes

- `platform/outbox` é infraestrutura, como `platform/audit`. Gravar um evento fora de uma transação falha.
- O tipo é versionado (`cobranca.paga.v1`). O payload traz ids, valores em centavos e flags; nenhum dado pessoal. Inclui `correlation_id` (fluxo do checkout) e `causation_id` (notificação ou job).
- Os contratos ficam em `cobrancas/events` (pacote só de tipos, [ADR-011](011-mapa-de-modulos-do-tj-connect.md)).
- A entrega é pelo menos uma vez, pelo worker ([ADR-014](014-worker-e-processamento-persistente.md)). Cada consumidor grava `(consumidor, evento_id)` com unicidade na mesma transação do efeito.

### Conciliação e reprocessamento

- **Ativa**:
  - um job periódico consulta as cobranças não terminais sem atualização recente;
  - as reservas próximas de expirar disparam consulta antecipada;
  - um job diário busca as orders do período no provedor e compara com o banco, encontrando órfãs e diferenças.
- **Reprocessamento**: jobs e entregas de evento que esgotarem as tentativas ficam em estado de falha permanente, nunca descartados em silêncio. Uma pessoa com `jobs:job:retry` ou `jobs:job:close` **e** `cobrancas:processamento:retry` (todas sensíveis; spec `01`) pode reenfileirar ou encerrar com motivo, com auditoria.
- **Divergências** (estado local × provedor):

| Situação | Tratamento |
|---|---|
| Local `AGUARDANDO`/`EM_PROCESSAMENTO`, provedor pago | Transição normal para `PAGA` |
| Local `NAO_CONCLUIDA`, provedor pago | `PAGA` com `tardio = true`; `associados` revalida a quota |
| Local `PAGA`, provedor reembolsado fora do sistema (pelo painel do provedor) | `REEMBOLSADA` com origem "externa"; devolução registrada no Financeiro; alerta |
| Local `PAGA`, provedor contestado | `CONTESTADA` |
| Valor ou moeda diferentes do esperado | `DIVERGENTE`; nenhuma ativação |
| Order no provedor com `external_reference` desconhecida | Registro de divergência órfã; nenhum efeito |
| Local `CRIADA` e order inexistente após as tentativas | `FALHA_CRIACAO`; a reserva associada segue o fluxo de liberação |
| Retrocesso de estado | `DIVERGENTE` |
| Reembolso em `processing` por muito tempo | Nova consulta; depois do limite, pendência |

A resolução de `DIVERGENTE` exige `cobrancas:conciliacao:resolve` (sensível), motivo e evidência, e gera transição auditada.

### Webhook

O webhook valida a assinatura, grava a notificação (com unicidade) e enfileira a sincronização numa transação, e responde 2xx. Assinatura inválida é registrada e recusada sem detalhe. A cadeia HTTP está na [ADR-015](015-superficies-web-e-rotas-publicas-do-tj-connect.md).

### Positive Consequences

- Nenhuma confirmação se perde entre banco e provedor; reprocessar é seguro por construção.
- Divergências viram estados explícitos, não exceções silenciosas.
- O provedor fica atrás de uma porta: a alternativa por preferências e a troca de recebedor (TJ-PAG-005) não mudam o ciclo interno.

### Negative Consequences

- Consistência eventual (segundos) entre o pagamento e a carteirinha.
- Worker, outbox, inbox e painel de pendências a construir e operar.
- Itens 1 a 8 ainda não verificados podem forçar a alternativa por preferências.

## Pros and Cons of the Options

### 1. Outbox, idempotência e conciliação ✅ Chosen

- ✅ Atômico com a transição; reprocessável; sem ciclo.
- ❌ Infraestrutura nova.

### 2. Callback síncrono

- ❌ Uma falha em `associados` reverte o registro da confirmação; a regra de quota passaria a rodar dentro da transação de `cobrancas`.

### 3. Polling

- ❌ Latência e carga constantes; `associados` precisaria conhecer estados internos de `cobrancas`.

## Links

- [ADR-011](011-mapa-de-modulos-do-tj-connect.md), [ADR-013](013-contas-de-servico-e-autoria-de-processos.md), [ADR-014](014-worker-e-processamento-persistente.md), [ADR-015](015-superficies-web-e-rotas-publicas-do-tj-connect.md), [ADR-017](017-evolucao-aditiva-de-financeiro-e-estoque.md), [ADR-018](018-invariantes-transacionais-de-quota-e-check-in.md).
- Mercado Pago: Checkout Pro via Orders (criar order), Orders API (consultar, cancelar, reembolsar), notificações (webhooks).
