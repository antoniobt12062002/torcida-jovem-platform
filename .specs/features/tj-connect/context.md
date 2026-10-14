# TJ Connect 2027 — Context (Discuss arquitetural)

**Gathered:** 2026-10-13 (revisão 2: decisões M-1 a M-9 e ajustes obrigatórios do mantenedor incorporados)
**Fonte de negócio:** `docs/tj-connect/03-tj-connect-requisitos.md` (prevalece sobre os PDFs, que não são versionados)
**Status:** Aguardando a revisão final do mantenedor. Nada aqui autoriza Specify detalhado, Design, migração, código ou commit.

---

## Feature Boundary

O TJ Connect amplia a plataforma existente com:

- site público e cadastro gratuito;
- adesão anual paga em três planos por temporada;
- quota de ingresso compartilhada pelos dois planos Arquibancada;
- cobrança pelo Mercado Pago (Checkout Pro, Pix e cartão em até 3x);
- carteirinha;
- jogos, intenção de comparecimento e relatório agregado ao clube;
- check-in online por fiscal designado;
- fatos de frequência desde o 1º jogo e procedimento individual de suspensão;
- camisetas sob encomenda em lotes manuais.

Tudo com Financeiro e Estoque existentes, sem duplicá-los. Fora do escopo: app nativo, modo offline, Wallet, renovação automática, prioridade de renovação, redistribuição de vaga, bilheteria do clube, loja virtual e **mudança de plano dentro da temporada** (M-6).

---

## Implementation Decisions

### D1 — `cobrancas` → [ADR-012](../../../docs/adr/012-cobrancas-e-eventos-persistentes.md)

- O módulo é independente do provedor e sem funcionalidade comercial antecipada.
- **Ciclo de vida interno completo** (revisão 3): `CRIADA`, `AGUARDANDO_PAGAMENTO`, `EM_PROCESSAMENTO`, `PAGA`, `PARCIALMENTE_REEMBOLSADA`, `NAO_CONCLUIDA` (não terminal: admite `→ PAGA` tardio), `FALHA_CRIACAO`, `REEMBOLSO_EM_ANDAMENTO`, `REEMBOLSADA`, `CONTESTADA`, `CONTESTACAO_PERDIDA`, `DIVERGENTE`. Reembolso é entidade própria com estados e teto pelo valor pago. Transições por `UPDATE` condicional; retrocesso informado pelo provedor vira `DIVERGENTE`.
- Relação com a adesão por eventos `cobranca.*.v1` (outbox, inbox). A adesão tem uma cobrança de origem e nunca troca de plano.
- **Idempotência em 7 camadas**: chave do provedor gravada antes da chamada; unicidade da notificação; job deduplicado por cobrança; consulta da verdade em vez do payload; transição condicional; chave no lançamento; inbox.
- Conciliação ativa (periódica e diária, com detecção de órfãs) e tabela de divergências local × provedor; resolução humana sensível.

### D2 — Autoria → [ADR-013](../../../docs/adr/013-contas-de-servico-e-autoria-de-processos.md)

- Três tipos de autoria distintos: **humana direta**, **decisão administrativa** (com `decision_id`) e **processo automatizado** (conta de serviço).
- **Processo nunca decide**: as permissões de juízo são `sensivel` e proibidas a papel de serviço; uma execução de serviço exige `decision_id` válido.
- Contas de serviço identificáveis, não interativas (sem credencial, sem sessão, sem `AdminMembership`), com privilégio mínimo e desativação como chave de emergência.
- `audit_log` ganha `actor_type = service` com id obrigatório, mais `correlation_id`, `causation_id`, `on_behalf_of_user_id` e `decision_id`. O ator `system` sem id deixa de ser usado em caminhos novos.

### D3 — Worker → [ADR-014](../../../docs/adr/014-worker-e-processamento-persistente.md)

- `cmd/worker` independente; a API só enfileira na transação do caso de uso, e um teste impede executores na API.
- `SKIP LOCKED`, concessão com prazo e **token de cerca**; relógio do banco.
- Tabela de recuperação para cada ponto de falha; tempo máximo por tipo, com renovação de concessão.
- Erros classificados em transitório ou permanente, com espera exponencial e variação aleatória.
- Falha definitiva mantida, visível, reenfileirável ou encerrável com motivo e auditoria, nunca excluída.

### D4 — Módulos → [ADR-011](../../../docs/adr/011-mapa-de-modulos-do-tj-connect.md)

- Grafo **real** separado por tipo de relação: chamada direta (D, importa `B/app`), contrato de evento (E, importa `B/events`) e contribuição de RBAC (R).
- Hoje só existem `financeiro →R identity` e `estoque →R identity`. A seta `financeiro → estoque` é permitida e não usada.
- Ausência de ciclos provada em três níveis: compilação (ordem topológica conferida seta a seta), cadeia síncrona numa transação e laço assíncrono com término.
- O `architecture_test` passa a verificar a lista explícita de setas e o tipo pelo pacote, o pacote `events` só de tipos, a aciclicidade da lista e dos imports, e a obrigatoriedade de entrada para módulo novo.
- Camisetas pertencem a `associados`; o estoque recebe só referência opaca.

### D5 — Fiscal e clube → [ADR-016](../../../docs/adr/016-permissoes-sensiveis-e-escopo-por-jogo.md)

- `FISCAL` via `AdminMembership`, mais designação por jogo em `acesso`.
- **Auto-check-in proibido no backend** (M-7): fiscal ≠ titular, na mesma transação do check-in.
- Clube sem conta: relatório agregado exportado, com auditoria.

### D6 — Web → [ADR-015](../../../docs/adr/015-superficies-web-e-rotas-publicas-do-tj-connect.md)

- Os grupos `(site)`, `(meu-tj)` e `(fiscal)` são novos; `(app)` é preservado. Login unificado; navegação por permissões e vínculo.
- A **liberação comercial não bloqueia** site, catálogo nem cadastro (M-5). A API recusa o checkout sem liberação vigente.
- Checkout Pro por redirecionamento. O retorno mostra só "em verificação".

### D7 — Financeiro e Estoque → [ADR-017](../../../docs/adr/017-evolucao-aditiva-de-financeiro-e-estoque.md)

- Só **planejamento**.
- Origem e chave de idempotência únicas no banco; cadeia de rastreio adesão → cobrança → provedor → lançamento → auditoria.
- **Tratamento financeiro**:
  - receita = preço do plano;
  - taxa retida separada;
  - líquido conferido com o provedor;
  - juros do comprador fora da receita;
  - reembolso e chargeback como devolução vinculada (nunca cancelamento);
  - indicador de contribuições líquido de devoluções;
  - pagamento tardio sem vaga numa conta transitória fora do indicador.
- Estoque: novas origens `PRODUCAO_ENCOMENDA` e `ENTREGA_BENEFICIO`, chave por recebimento e por entrega, **nenhum dado pessoal ou personalização**.

### Decisões M-1 a M-9 do mantenedor

| # | Decisão | Onde está refletida |
|---|---|---|
| M-1 | ADRs aprovadas conceitualmente; ficam `Proposed` até a conferência final | Cabeçalho das ADRs 011–018 |
| M-2 | Checkout Pro, priorizando a Orders API se todos os requisitos forem atendidos; só Pix e cartão de crédito | ADR-012: fatos verificados, 8 itens não verificados como critério de aceite, regra de escolha e alternativa por preferências |
| M-3 | Cancelamento e reembolso na R1, com auditoria; arrependimento sob validação jurídica | ADR-016 (permissões de solicitar, analisar, decidir, aprovar e executar), STATE (R1) |
| M-4 | +Camisa vendido só com a R3 funcional | STATE (R2 vende só Arquibancada; R3 publica o +Camisa) |
| M-5 | Liberação comercial com evidência, responsável, data, versão e validade; site e cadastro não bloqueados; nada por presunção técnica | ADR-011 (dono `associados`), ADR-016 (`liberacao_comercial:grant`/`revoke`, `plano:publish`), ADR-015, ADR-018 |
| M-6 | Sem upgrade nem downgrade; histórico que permita evolução futura | ADR-012: uma adesão por contrato, plano imutável, ligação futura possível |
| M-7 | Fiscal não registra a própria presença; verificação no backend | ADR-016 e ADR-018 (cenário 6) |
| M-8 | Preservar números confiáveis de 2026, se existirem; senão, TJ-0001 em 2027; levantamento antes; sem importação automática de inconsistentes nem duplicatas; número permanente, independente da temporada | Sub-spec `06` (o número é do perfil associativo, nunca da adesão; unicidade no banco; importação só validada) |
| M-9 | Versionar o `03-requisitos.md` após revisão; PDFs fora | Revisão feita (ver STATE, "Revisão do documento de requisitos"); PDFs e `docs/tj-connect/interno/` excluídos pelo `.gitignore` |

### Cuidados obrigatórios → garantia

| Cuidado | Garantia |
|---|---|
| 1. Permissões sensíveis explícitas | ADR-016: classe obrigatória, nenhuma concessão sensível automática, matriz sujeita a aprovação |
| 2. Frequência desde o 1º jogo | ADR-018: fatos imutáveis e apuração pura (cenário 8) |
| 3. Reserva transacional; tardio nunca excede | ADR-018: `CHECK` e `UPDATE` condicional (cenários 1 e 2) |
| 4. Foto só para o benefício Arquibancada | Sub-spec `06` (condição) e `11` (upload) |
| 5. Camisetas sob encomenda, lotes manuais, só nome | Sub-specs `12` e `13`; ADR-017 |
| 6. Bloqueios legais não presumidos | Liberação comercial (M-5) e matriz sensível vazia até aprovação |
| 7. PDFs e finanças internas fora do Git | `.gitignore` (PDFs e `docs/tj-connect/interno/`); nenhum número interno nas specs |

### Agent's Discretion

Nenhuma área foi delegada. Ficam para o Design, dentro dos limites das ADRs: nomes de tabelas, colunas e eventos; tempos de concessão e intervalos; forma da restrição que impede sessão de conta de serviço.

### Declined / Undiscussed Gray Areas → Assumptions

Cada item vira uma pergunta em aberto na sub-spec indicada:

| Tema | Premissa padrão até a decisão | Sub-spec |
|---|---|---|
| Reembolso parcial | **Decidido (A-1)**: o domínio representa integral e parcial com teto; nenhuma regra automática de parcial sem validação jurídica e confirmação do Mercado Pago | 04, 06 |
| Efeito de chargeback sobre o benefício | **Decidido (A-2)**: nenhum efeito automático; contestação pendente ≠ revertida; consequência só por procedimento | 04, 06 |
| Mecânica do QR | **Decidido (A-3)**: dinâmico, curto, específico do jogo, validado online; parâmetros no Design | 07 |
| Leitor de QR no navegador | Suporte em iOS e Android a verificar | 10 |
| Prazo da reserva | Alinhado ao `expiration_time` da order, mais margem; valor numérico pendente (P0-5) | 05 |

---

## Specific References

- Mercado Pago (consultado em 2026-10-13): Checkout Pro via Orders (`/v1/orders`, `type: online`, `processing_mode: manual`, `expiration_time`, `X-Idempotency-Key`, `checkout_url`); Orders API (consultar, `cancel` só com `created`, `refund` com idempotência); webhook `x-signature` com manifesto `id:…;request-id:…;ts:…;`; preferências com `installments` e `excluded_payment_types` (alternativa).
- Divergências dos PDFs que não prevalecem: nome e CPF definitivos; carteira "na hora"; "presença já registrada hoje"; aviso obrigatório de ausência; "peso na renovação"; upgrade pelo app; relatórios PDF/CSV.

---

## Deferred Ideas

- Mudança de plano dentro da temporada (M-6).
- Wallet e modo offline como "fase 2" (PDF 02).
- **Correção (2026-10-14)**: o número permanente do associado **não** é ideia adiada. Ele faz parte do MVP (TJ-ASSOC-001, AC-PL-02; M-8). A renovação não é automática, mas uma nova adesão em temporada posterior **preserva o mesmo número e o histórico** (sub-spec `06`).
- Exportações além do relatório agregado ao clube.
- Portal de transparência e balancete público (PDF 01).
- Loja, pré-venda e condições especiais para associados.
