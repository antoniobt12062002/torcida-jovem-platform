# TJ Connect 2027 — STATE

Documento de coordenação do programa TJ Connect, que atravessa `identity`, `associados`, `cobrancas`, `acesso`, `financeiro`, `estoque`, `platform` e `web`. Segue o padrão de `financeiro/STATE.md` e `estoque/STATE.md`. As mudanças em `financeiro` e `estoque` ganham sub-specs **dentro das pastas desses módulos**, referenciadas aqui.

Numeração local: **TJC-D-NNN** para decisões que valem só para o TJ Connect. As decisões de arquitetura do projeto inteiro são as ADRs `011`–`018` (`Proposed`), que entram como `AD-NNN` no `.specs/STATE.md` só quando forem aceitas.

**Estado atual:** Specify do **G1** revisado conforme o parecer externo e as decisões do mantenedor de 2026-10-14, publicado para conferência externa. **Design não autorizado.** Specs: `spec/02-identity-governanca.md` e `spec/01-platform-processamento.md`, ambos com `validate_spec.py` 0 erros e 0 avisos. G2 e seguintes **não autorizados**. Nenhum design, migração ou código. A documentação de planejamento (Discuss e G1) foi publicada na branch `feature/tj-connect-discuss` para revisão externa, sem PR e sem merge.

---

## Princípio: especificar inteiro, implementar fatiado

As entregas R0–R4 dividem **só a implementação**. A Onda A especifica **integralmente os contratos e os estados** de pagamento, quota, adesão, jogos e frequência, mesmo quando a tela ou a operação administrativa correspondente só for implementada numa entrega posterior:

- **Pagamento**: os 12 estados da cobrança (ADR-012, revisão 3), o ciclo do reembolso integral e parcial, todas as divergências, reembolso, chargeback e conciliação ficam na spec `04`, embora a R1 venda só o plano básico.
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

Decisões do mantenedor de 2026-10-14. **Nenhuma concessão real foi alterada no código**; todas as concessões abaixo valem só como especificação.

| # | Decisão | Sub-spec | Conteúdo | Situação |
|---|---|---|---|---|
| D-02-1 | Contas de serviço: desativação e reativação | 02 | `ADMIN_SISTEMA` e `PRESIDENTE` recebem `identity:service_account:deactivate` e `:reactivate` por concessão explícita; reativação por pessoa diferente, com contingência (espera de 30 min, motivo e revisão posterior) quando não houver outra pessoa (SVC-01 AC16–21) | **Aprovada**; parâmetros da contingência aguardam confirmação |
| D-02-2 | Classificação das 30 permissões existentes | 02 | Base de compatibilidade aprovada. Proposta revisada: `sensivel` para operações que movimentam recursos, alteram saldos, estrutura contábil ou acesso de terceiros, e atos institucionais (18 de 30), com 16 concessões explícitas justificadas ao `PRESIDENTE` | **Parcial**: base aprovada; classificação de segurança revisada **aguarda aprovação** |
| D-02-3 | Retirar do `PRESIDENTE` os atos do Conselho Fiscal | 02 | Retirada de `financeiro:prestacao_contas:approve` e `financeiro:parecer:opine`; única exceção aprovada à preservação da matriz (TRN-01 AC8) | **Aprovada**; aplicação **condicionada à D-02-4** |
| D-02-4 | Nomeação institucional do Conselho Fiscal (consequência verificada da D-02-3) | 02 | `identity:conselho_fiscal:manage`; decisão registrada com referência ao ato institucional; `Covers` sem as permissões institucionais; `PRESIDENTE` × `CONSELHO_FISCAL` incompatíveis por identidade; separação de funções por identidade sobre a união dos papéis (INS-01) | **Nova; aguarda aprovação** (e a lista de papéis incompatíveis, sujeita ao estatuto) |
| D-01-1 | Operação de jobs | 01 | `ADMIN_SISTEMA`: consultar e reenfileirar; `PRESIDENTE`: consultar por concessão explícita (`jobs:job:read` passa a `sensivel`); encerramento com autorização do módulo dono e sem silenciar obrigações (DLQ-01 AC6–8) | **Aprovada**; quem recebe `jobs:job:close` continua sem decisão |
| D-01-2 | Parâmetros e retenção | 01 | Valores iniciais configuráveis; retenção só com deduplicação de negócio garantida (RET-01) | **Aprovada com condição** |
| D-G1-1 | Proteção **versionada** contra a inclusão dos PDFs e de `docs/tj-connect/interno/` | — | `.gitignore` atualizado | **Resolvida** (commit `40b5165`) |

### Matriz de permissões proposta (catálogo atual e permissões novas do G1)

Legenda: **A** = decorre de decisão aprovada; **C** = condicionada (aguarda aprovação); "=" = sem mudança.

| Papel | Hoje | Proposta | Mudanças |
|---|---|---|---|
| `PRESIDENTE` | 30 (todas automáticas) | **33**: das 30 existentes, 28 (12 `normal` automáticas + 16 `sensivel` explícitas); mais 5 novas | **A**: −`financeiro:prestacao_contas:approve`, −`financeiro:parecer:opine` (D-02-3, aplicada só com a D-02-4); +`identity:service_account:deactivate`, +`:reactivate` (D-02-1); +`jobs:job:read` (D-01-1); +`identity:service_account:read` (normal, automática). **C**: as 16 explícitas dependem da classificação revisada (D-02-2); +`identity:conselho_fiscal:manage` (D-02-4) |
| `ADMIN_SISTEMA` | 8 (`identity` ×7, `audit:log:read`) | **13** | **A**: +`identity:service_account:deactivate`, +`:reactivate` (D-02-1); +`jobs:job:read`, +`jobs:job:retry` (D-01-1). **C**: +`identity:service_account:read` (necessária para operar a D-02-1; concessão explícita a confirmar) |
| `CONSELHO_FISCAL` | 11 | **11** = | Agora na lista explícita (CFS-01); mantém `prestacao_contas:approve` e `parecer:opine` |
| `TESOURARIA` | 13 | **13** = | 9 delas passam a `sensivel` (C), mas a concessão já é explícita no código |
| `ESTOQUE_LOJA` | 6 | **6** = | 2 passam a `sensivel` (C), concessão já explícita |
| `DIRETORIA` | 8 (leituras) | **8** = | — |
| `ASSOCIADO`, `EVENTOS` | 0 | **0** = | — |
| `FISCAL` (novo) | — | **0** | Concessões só na spec `07` |
| Papéis de serviço (novos) | — | **0** | Concessões `servico` só nas specs dos módulos |

**Exceção à preservação**: a única diferença do snapshot nas 30 permissões existentes é a D-02-3 (−2 no `PRESIDENTE`). O teste de snapshot de TRN-01 registra essa diferença como alteração aprovada. Não há outra perda nem ganho em permissões existentes.

### Impacto verificado da D-02-3

- **`authz.Covers`** exige que quem atribui um papel, promove, revoga, desativa ou redefine a senha tenha **todas** as permissões envolvidas (`identity/app/assign_roles.go:87`, `promote_admin.go:78`, `revoke_admin.go:74`, `deactivate_user.go:88`, `admin_reset_password.go:75`). Depois da D-02-3, só o Conselho Fiscal teria as duas permissões institucionais, e ele não tem permissões de `identity`. Resultado sem correção: **ninguém** conseguiria nomear, retirar, desativar ou redefinir a senha de um conselheiro. Por isso a D-02-3 só pode ser aplicada com a D-02-4.
- **União de permissões**: uma pessoa com `PRESIDENTE` e `CONSELHO_FISCAL` recuperaria pela união as permissões retiradas. A D-02-4 torna os dois papéis incompatíveis, verificando pela identidade do usuário. Quem já acumular os dois aparece como pendência na partida, sem revogação automática.
- **Separação de funções**: toda regra compara a identidade do usuário sobre a união dos seus papéis, nunca o nome do papel (INS-01 AC7).

## Parecer externo de 2026-10-14 sobre o commit `40b5165`

Parecer favorável à arquitetura, com aprovação condicionada a 9 ajustes. Situação de cada um:

| # | Ajuste pedido | Onde foi tratado |
|---|---|---|
| 1 | Sem promessa de execução física única; limites da cerca; concorrência depois da perda de concessão | Spec `01`: Goals, INV-P2/P4/P8, "Limites da cerca", LEA-01 AC6–9, T-04b/T-04c; ADR-014 |
| 2 | Inbox × retenção; deduplicação segura depois da limpeza; garantias próprias dos módulos | Spec `01`: "Retenção e deduplicação durável", INV-P9, nova história **RET-01**, T-11d/T-11e; ADR-014 |
| 3 | Versão aplicada transacional; `tolerante` só para consumidor convergente; testes simultâneos | Spec `01`: "Ordem de eventos" (modos `independente`, `estrita`, `tolerante`), INB-01 AC3–9, T-11b/T-11c |
| 4 | Próxima janela de agendamento recuperável | Spec `01`: F7 (encadeamento atômico e verificação independente), INV-P10, SCH-01 AC3–5, T-15b/T-15c; ADR-014 |
| 5 | Conselho Fiscal por permissões explícitas | Spec `02`: CFS-01 (lista de 11), T-18/T-19; ADR-016 |
| 6 | Regra aprovada × decisão administrativa | Spec `02`: "Dois tipos de execução automatizada", F3a, SOD-01 AC9–11, AUT-01 AC10, T-20/T-21; ADR-013, ADR-016 |
| 7 | Histórico append-only × anulação × índice; projeção de presença; check-in × cancelamento ou suspensão | ADR-018: projeção de presença efetiva, elegibilidade sob bloqueio, cenários 6, 7 e **9**. Fica para a spec `07` antes do Design de jogos |
| 8 | `NAO_CONCLUIDA` não terminal; reembolso parcial | ADR-012: tabela de estados, `PARCIALMENTE_REEMBOLSADA`, entidade `Reembolso`, evento `cobranca.reembolso_concluido.v1`; ADR-011 |
| 9 | Recomendações de governança | Tabela "Abertas no G1" e impacto na matriz, acima; specs `01` e `02` (Pendências) |

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
| `02-identity-governanca` | CLS-01 (classes, fail-closed), TRN-01 (transição sem perda), SVC-01 (contas de serviço), SOD-01 (solicitação, decisão e execução), AUT-01 (autoria), FIS-01 (`FISCAL` e escopo por jogo), CFS-01 (Conselho Fiscal), INS-01 (nomeação institucional e papéis incompatíveis) | §11, §16; M-3, M-5, M-7, A-2; ADR-013, ADR-016 |
| `01-platform-processamento` | JOB-01 (registro transacional), LEA-01 (concessão e cerca), RTY-01 (tentativas e falha definitiva), REC-01 (recuperação), OUT-01 (outbox), INB-01 (idempotência e ordem), SVX-01 (conta de serviço e auditoria), DLQ-01 (recuperação humana), SCH-01 (agendamento), RET-01 (retenção sem perda de deduplicação), OBS-01 (observabilidade), ARQ-01 (limites arquiteturais) | §16; D3; ADR-011, ADR-012, ADR-014 |

---

## Decisões fechadas (TJC-D-NNN)

Nenhuma ainda. As decisões de arquitetura estão nas ADRs 011–018 (`Proposed`); as de produto, em `context.md` (M-1 a M-9).

## Handoff

- **Fase**: Specify do G1 revisado conforme o parecer externo e as decisões do mantenedor de 2026-10-14 (specs `01` e `02` validadas; ADRs 011–014, 016 e 018 em revisão 3, ainda `Proposed`). Publicado na branch para **conferência externa final**. Design e G2 não autorizados.
- **Branch**: `feature/tj-connect-discuss`, publicada para revisão externa (sem PR, sem merge na `develop`).
- **Arquivos**:
  - `.specs/features/tj-connect/{context.md,STATE.md}`;
  - `docs/adr/011`–`018` e o índice em `docs/adr/README.md`;
  - nota de reconciliação no Handoff de `.specs/STATE.md`.
- **Specs do G1**: `.specs/features/tj-connect/spec/02-identity-governanca.md` e `spec/01-platform-processamento.md`.
- **Próximo passo**: conferência externa da versão publicada; aprovação da classificação revisada (D-02-2), da D-02-4 e dos parâmetros da contingência de reativação; depois, com autorização, Design do G1.
