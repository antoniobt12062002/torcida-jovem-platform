# TJ Connect 2027 — STATE

Documento de coordenação do programa TJ Connect, que atravessa `identity`, `associados`, `cobrancas`, `acesso`, `financeiro`, `estoque`, `platform` e `web`. Segue o padrão de `financeiro/STATE.md` e `estoque/STATE.md`. As mudanças em `financeiro` e `estoque` ganham sub-specs **dentro das pastas desses módulos**, referenciadas aqui.

Numeração local: **TJC-D-NNN** para decisões que valem só para o TJ Connect. As decisões de arquitetura do projeto inteiro são as ADRs `011`–`018` (`Proposed`), que entram como `AD-NNN` no `.specs/STATE.md` só quando forem aceitas.

**Estado atual:** Specify do **G1** escrito e aguardando revisão do mantenedor: `spec/02-identity-governanca.md` e `spec/01-platform-processamento.md`, ambos com `validate_spec.py` 0 erros e 0 avisos. G2 e seguintes **não autorizados**. Nenhum design, migração ou código. A documentação de planejamento (Discuss e G1) foi publicada na branch `feature/tj-connect-discuss` para revisão externa, sem PR e sem merge.

---

## Princípio: especificar inteiro, implementar fatiado

As entregas R0–R4 dividem **só a implementação**. A Onda A especifica **integralmente os contratos e os estados** de pagamento, quota, adesão, jogos e frequência, mesmo quando a tela ou a operação administrativa correspondente só for implementada numa entrega posterior:

- **Pagamento**: os 11 estados da cobrança, todas as divergências, reembolso, chargeback e conciliação ficam na spec `04`, embora a R1 venda só o plano básico.
- **Quota**: a linha de quota, a reserva, o pagamento tardio, a redução de capacidade, a liberação de vaga e a liberação comercial ficam na spec `05`, embora o Arquibancada só seja vendido na R2.
- **Adesão**: uma máquina de estados única na `06` (`AGUARDANDO_PAGAMENTO`, `ATIVA`, `NAO_CONCLUIDA`, `EM_REGULARIZACAO`, `CANCELAMENTO_SOLICITADO`, `ENCERRADA`, `EXPIRADA`), com os pontos de contato da foto e da camiseta.
- **Jogos e frequência**: os **fatos** (elegibilidade congelada, check-ins, correções, suspensões, início do benefício) e a apuração ficam na `07` e na `08`. São **coletados desde o 1º jogo da R2**; a regra dos 6 jogos limita só a instauração. O rito do procedimento (`14`) pode vir depois porque o seu efeito (suspensão bloqueia a credencial e sai do denominador) e a sua guarda já estão fixados na `08`.

**Regra de corte**: uma sub-spec só vai para a Onda B se o seu ponto de contato com a Onda A estiver fixado numa spec da Onda A.

---

## Sequência definitiva da Onda A (Specify)

Ordem pela dependência de contratos. Cada spec passa por `validate_spec.py`. Há um **ponto de revisão do mantenedor** ao fim de cada grupo, e o Design é feito por grupo.

| Ordem | Sub-spec | Pasta | Módulo dono | Conteúdo principal | ADRs | Grupo de Design |
|---|---|---|---|---|---|---|
| 1 | `02-identity-governanca` | `tj-connect/spec/` | `identity`, `platform/authz`, `platform/audit` | Classes de permissão e fail-closed; transição sem perda; contas de serviço; separação entre solicitação, decisão e execução; autoria e causalidade; `FISCAL` e contrato de escopo por jogo; Conselho Fiscal por lista permitida. **Escrita (rascunho)** | 013, 016 | **G1** |
| 2 | `01-platform-processamento` | `tj-connect/spec/` | `platform`, `cmd/worker` | Jobs, captura, cerca, tentativas, falha definitiva, recuperação, outbox e inbox, ordem de eventos, agendamento, conta de serviço no worker, observabilidade, **lista de setas do `architecture_test`** (movida da `02`, porque trata de fronteiras de processamento e contratos de evento). **Escrita (rascunho)** | 011, 012, 014 | **G1** |
| 3 | `03-identity-autocadastro` | `tj-connect/spec/` | `identity`, `platform/httpx` | Cadastro gratuito, confirmação de e-mail, reenvio, limite por IP, cadeia pública e de webhook | 015 | G2 |
| 4 | `04-cobrancas` | `tj-connect/spec/` | `cobrancas` | Ciclo completo, idempotência, webhook, conciliação, divergências, reembolso, contestação; **os 8 itens de verificação do Mercado Pago como critérios de aceite** | 012, 018 | **G3** |
| 5 | `07-receitas-de-cobranca` (nome provisório) | `financeiro/spec/` | `financeiro` | Origem e chave; receita, devolução e tarifa de cobrança; indicador de contribuições; conta transitória | 017 | **G3** |
| 6 | `05-associados-temporadas-quota` | `tj-connect/spec/` | `associados` | Temporada, plano por temporada, liberação comercial, quota, reserva, tardio, redução, liberação de vaga | 016, 018 | **G3** |
| 7 | `06-associados-adesao` | `tj-connect/spec/` | `associados` | Perfil, número permanente (M-8), ciclo da adesão, carteirinha, cancelamento e reembolso, condição de foto, contato com camiseta, histórico (M-6) | 012, 016, 018 | **G3** |
| 8 | `07-acesso-jogos-checkin` | `tj-connect/spec/` | `acesso` | Jogos, elegibilidade, conclusão, intenção, relatório ao clube, designação, credencial/QR, tentativas, check-in, auto-check-in proibido, correção | 016, 018 | **G4** |
| 9 | `08-acesso-frequencia` | `tj-connect/spec/` | `acesso` | Fatos, apuração pura, efeito da suspensão vigente, guarda de instauração | 018 | **G4** |
| 10 | `10-web-tj-connect` | `tj-connect/spec/` | `web` | Superfícies, navegação pós-login, home pública, retorno do checkout, estados de UI | 015 | G5 |

Os números no nome das sub-specs (`01`–`14`) são identificadores estáveis; a coluna "Ordem" é a sequência de trabalho.

**Pontos de revisão**: depois de G1 (fundação de autorização e processamento); depois de G3 (núcleo de pagamento, quota e adesão); depois de G4 (jogos e frequência).

### Onda B (os pontos de contato já estão fixados na Onda A)

| Sub-spec | Pasta | Contato fixado em | Bloqueio de negócio |
|---|---|---|---|
| `11-associados-foto` | `tj-connect/spec/` | `06` | Política de foto (P1-8) |
| `12-associados-camisetas-lotes` | `tj-connect/spec/` | `06` | Processo de camiseta (P1-13) |
| `06-referencias-externas` (nome provisório) | `estoque/spec/` | `12` | — |
| `14-acesso-procedimento-suspensao` | `tj-connect/spec/` | `08` | **TJ-JOG-007 / P0-3 (jurídico)** |

---

## Entregas (fatias de implementação)

| Entrega | Conteúdo | Sub-specs | Condição técnica | Liberação comercial exigida (M-5) |
|---|---|---|---|---|
| **R0 Fundação** | Worker, outbox, contas de serviço, classes de permissão, cadeias pública e de webhook, receita de cobrança no Financeiro | 02, 01, 07 (Financeiro), parte da 03 | Nada visível; pode ir a produção desligada | — |
| **R1 Cadastro e plano básico** | Site, cadastro, Meu TJ, temporada, plano Associado TJ, cobrança completa, carteirinha sem ingresso, **cancelamento e reembolso** (M-3) | 03, 04, 05 (sem publicar Arquibancada), 06, 10 | E-mail real; credenciais Mercado Pago de produção; verificação dos 8 itens concluída | Regularidade institucional e conta recebedora, estatuto e regulamento, termos e LGPD, validação jurídica do arrependimento |
| **R2 Arquibancada e jogos** | Quota e reserva ativas, tardio, foto, jogos, intenção, relatório ao clube, fiscal e check-in, **coleta de frequência desde o 1º jogo** | 05, 06 (completas), 07, 08, 11 | S3 de produção; HTTPS; teste de pico (P2-19); matriz sensível aprovada | Quota formalizada com o clube, prazo da reserva |
| **R3 Camisetas** | Encomendas, lotes, recebimento, entrega; **publicação do Arquibancada + Camisa** (M-4) | 12, 06 (Estoque) | — | Fornecedor, grade, prazo, termos da camiseta |
| **R4 Procedimento** | Rito de frequência e suspensão | 14 | Fatos coletados desde a R2 | Regulamento disciplinar válido (TJ-JOG-007) |

---

## Pendências classificadas pelo que realmente bloqueiam

| Bloqueia | Pendência | Responsável |
|---|---|---|
| **Specify da Onda A** | **Nada.** As questões abertas viram premissas explícitas nas specs (seção "Assumptions & Open Questions"), sujeitas ao gate de fechamento | — |
| **Design do G3** | Verificação dos 8 itens do Mercado Pago em sandbox (ADR-012). Exige credenciais de teste e uma **autorização específica para um spike de integração**, que é código | Mantenedor (credenciais e autorização) |
| **Design da sub-spec `07` do Financeiro** | Conta transitória do recebido sem vaga; data do lançamento (aprovação ou liberação); tratamento da taxa no reembolso | Contador |
| **Operação dos fluxos sensíveis** | Matriz de papéis das permissões sensíveis (proposta na `02`, aprovada pelo mantenedor conforme o regulamento) | Mantenedor e regulamento |
| **Produção** | Provedor de e-mail (R1); provedor S3 (R2); conta Mercado Pago do recebedor definitivo (TJ-PAG-005) | Mantenedor |
| **Comercialização** (via liberação comercial, nunca presumida) | P0-1 quota com o clube; P0-2 regularidade institucional, estatuto e regulamento; P0-3 disciplina de suspensão; P0-4 Mercado Pago; P0-5 prazo da reserva; P0-6 termos e arrependimento | Diretoria, jurídico, clube |
| **Numeração operacional** (não a spec) | Levantamento da base de 2026 (M-8) | Diretoria |

---

## Decisões do mantenedor

### Registradas (2026-10-13)

M-1 a M-9: ver `context.md`, tabela "Decisões M-1 a M-9".

A-1 a A-4 (2026-10-13):

| # | Decisão | Onde está refletida |
|---|---|---|
| A-1 | O domínio representa reembolso integral e parcial (valor, motivo, aprovação, rastreio, teto pelo valor elegível); nenhuma regra automática de reembolso parcial sem validação jurídica e confirmação do Mercado Pago | ADR-012; specs `04` e `06` |
| A-2 | Contestação não suspende o benefício automaticamente; contestação pendente ≠ pagamento definitivamente revertido; consequência associativa só por procedimento | ADR-012; specs `04` e `06` |
| A-3 | QR dinâmico, de curta duração, específico do jogo e validado online; validade, assinatura e renovação no Design; não dispensa a unicidade do check-in nem a conferência visual | ADR-018; spec `07` |
| A-4 | Versionar o `03-requisitos.md` sem informação interna; PDFs fora do Git | Trechos reescritos (ver abaixo); originais em nota privada não versionada |

Atualização do Mercado Pago: o mantenedor confirmou na documentação o `config.payment_method.max_installments` (limite de 3 parcelas). Falta comprovar na integração; os demais itens continuam pendentes (ADR-012). **Nenhum teste externo nesta etapa.**

### Abertas no G1 (não impedem o Design; fechar até as Tasks)

| # | Decisão | Sub-spec | Recomendação |
|---|---|---|---|
| D-02-1 | Quem recebe `identity:service_account:update` (chave de emergência das contas de serviço) | 02 | `ADMIN_SISTEMA` e `PRESIDENTE`, por concessão explícita |
| D-02-2 | Confirmar a classificação proposta das 30 permissões existentes (10 sensíveis, com concessão explícita de transição ao `PRESIDENTE`) | 02 | Aprovar a tabela da spec |
| D-02-3 | Rever concessões herdadas pelo `PRESIDENTE` que são atos do Conselho Fiscal (`financeiro:prestacao_contas:approve`, `financeiro:parecer:opine`) | Fora do G1 | Decisão de governança separada; o G1 preserva o estado atual |
| D-01-1 | Quem recebe `jobs:job:read`, `jobs:job:retry` e `jobs:job:close` | 01 | `ADMIN_SISTEMA` para leitura e reenfileiramento; encerramento exige também a permissão do módulo dono do tipo |
| D-01-2 | Valores padrão (concessão, tentativas, espera) e retenção (90 e 180 dias) | 01 | Aprovar os padrões da spec |
| D-G1-1 | Proteção **versionada** contra a inclusão dos PDFs e de `docs/tj-connect/interno/` | — | **Resolvida** (autorizada pelo mantenedor): `.gitignore` atualizado com `docs/tj-connect/*.pdf` e `docs/tj-connect/interno/` |

---

## Revisão do documento de requisitos para versionamento (M-9)

`docs/tj-connect/03-tj-connect-requisitos.md` foi revisado por completo:

- **Não contém**: dados pessoais, CPF, nomes de pessoas, e-mails, telefones, valores de caixa, receitas ou despesas internas, credenciais.
- **Contém só preços públicos dos planos** (R$ 60, R$ 160, R$ 240), que serão divulgados de qualquer forma.
- **Trechos reescritos por decisão A-4**:
  - §7 TJ-PAG-005: virou o bloqueio genérico "confirmar a conta recebedora definitiva (titularidade institucional, elegibilidade, credenciais, termos, obrigações fiscais) antes da comercialização";
  - §19 P0-2: a menção ao andamento institucional específico virou "regularidade institucional da associação";
  - §0: os dois PDFs são citados como "documento interno, não versionado".
- **Originais preservados** fora do Git, em `docs/tj-connect/interno/`, diretório excluído pelo `.gitignore`.

O arquivo de requisitos é versionado apenas nesta versão sanitizada.

---

## Rastreabilidade (requisito de negócio → sub-spec)

| Requisitos | Sub-spec |
|---|---|
| TJ-ID-001, TJ-ID-004 (parcial), AC-ID-01, AC-ID-02 | 03, 10 |
| TJ-ID-002, AC-ID-03 | 06 (condição), 11 (upload) |
| TJ-ID-003 | 06 |
| TJ-ASSOC-001, TJ-ASSOC-002, AC-PL-01, AC-PL-02 | 05, 06 |
| TJ-ASSOC-003, TJ-ASSOC-004, AC-CAN-01, AC-CAN-02, AC-CAP-06 | 05 (liberação de vaga), 06 (cancelamento), 04 (reembolso) |
| TJ-ASSOC-005..008, TJ-EST-001, AC-CAM-01..05 | 06 (contato), 12, 06-Estoque |
| TJ-CAP-001..003, TJ-PAG-002, TJ-PAG-003, AC-CAP-01..05 | 05, 06 |
| TJ-PAG-001, TJ-PAG-004, TJ-PAG-005, AC-PAG-01..03 | 04 |
| TJ-FIN-001 | 07-Financeiro, 04 |
| TJ-JOG-001..003, TJ-JOG-006, AC-JOG-01..04, AC-SEC-01 | 07 |
| TJ-JOG-004, AC-FRQ-01..03 | 08 |
| TJ-JOG-005, TJ-JOG-007, AC-FRQ-04 | 08 (efeito e guarda), 14 (rito) |
| §11 (matriz de autorização) | 02 e cada sub-spec de módulo |
| §13 (superfícies) | 10 |
| §16 (segurança, integridade, operação) | 01, 02, 03, 04 e transversal |

### Requisitos das sub-specs do G1

| Sub-spec | IDs | Origem principal |
|---|---|---|
| `02-identity-governanca` | CLS-01 (classes, fail-closed), TRN-01 (transição sem perda), SVC-01 (contas de serviço), SOD-01 (solicitação, decisão e execução), AUT-01 (autoria), FIS-01 (`FISCAL` e escopo por jogo), CFS-01 (Conselho Fiscal) | §11, §16; M-3, M-5, M-7, A-2; ADR-013, ADR-016 |
| `01-platform-processamento` | JOB-01 (registro transacional), LEA-01 (concessão e cerca), RTY-01 (tentativas e falha definitiva), REC-01 (recuperação), OUT-01 (outbox), INB-01 (idempotência e ordem), SVX-01 (conta de serviço e auditoria), DLQ-01 (recuperação humana), SCH-01 (agendamento), OBS-01 (observabilidade), ARQ-01 (limites arquiteturais) | §16; D3; ADR-011, ADR-012, ADR-014 |

---

## Decisões fechadas (TJC-D-NNN)

Nenhuma ainda. As decisões de arquitetura estão nas ADRs 011–018 (`Proposed`); as de produto, em `context.md` (M-1 a M-9).

## Handoff

- **Fase**: Specify do G1 escrito (02 e 01, validados); **parado para revisão do mantenedor**. G2 não autorizado.
- **Branch**: `feature/tj-connect-discuss`, publicada para revisão externa (sem PR, sem merge na `develop`).
- **Arquivos**:
  - `.specs/features/tj-connect/{context.md,STATE.md}`;
  - `docs/adr/011`–`018` e o índice em `docs/adr/README.md`;
  - nota de reconciliação no Handoff de `.specs/STATE.md`.
- **Specs do G1**: `.specs/features/tj-connect/spec/02-identity-governanca.md` e `spec/01-platform-processamento.md`.
- **Próximo passo**: revisão externa e do mantenedor sobre o G1 (decisões D-02-1, D-02-2, D-02-3, D-01-1, D-01-2); depois, com autorização, Design do G1 ou Specify do G2.
