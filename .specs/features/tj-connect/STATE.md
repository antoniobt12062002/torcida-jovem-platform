# TJ Connect 2027 — STATE

Documento de coordenação do programa TJ Connect, que atravessa `identity`, `associados`, `cobrancas`, `acesso`, `financeiro`, `estoque`, `platform` e `web`. Segue o padrão de `financeiro/STATE.md` e `estoque/STATE.md`. As mudanças em `financeiro` e `estoque` ganham sub-specs **dentro das pastas desses módulos**, referenciadas aqui.

Numeração local: **TJC-D-NNN** para decisões que valem só para o TJ Connect. As decisões de arquitetura do projeto inteiro são as ADRs `011`–`018` (`Proposed`), que entram como `AD-NNN` no `.specs/STATE.md` só quando forem aceitas.

**Estado atual:** Specify do **G1 aprovado** (2026-10-14). **Design do G1 escrito** (`design/02-identity-governanca.md`, `design/01-platform-processamento.md`), aguardando revisão; Tasks e Execute **não autorizados**. Specs: `spec/02-identity-governanca.md` e `spec/01-platform-processamento.md`, ambos com `validate_spec.py` 0 erros e 0 avisos. G2 e seguintes **não autorizados**. Nenhuma migração nem código. A documentação de planejamento (Discuss, Specify e Design do G1) foi publicada na branch `feature/tj-connect-discuss` para revisão externa, sem PR e sem merge.

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
| 1 | `02-identity-governanca` | `tj-connect/spec/` | `identity`, `platform/authz`, `platform/audit` | Classes de permissão e fail-closed; transição sem perda; contas de serviço; separação entre solicitação, decisão e execução; autoria e causalidade; `FISCAL` e contrato de escopo por jogo; Conselho Fiscal por lista permitida; nomeação institucional (INS-01). **Spec aprovada; Design escrito** | 013, 016 | **G1** |
| 2 | `01-platform-processamento` | `tj-connect/spec/` | `platform`, `cmd/worker` | Jobs, captura, cerca, tentativas, falha definitiva, recuperação, outbox e inbox, ordem de eventos, agendamento, conta de serviço no worker, observabilidade, **lista de setas do `architecture_test`** (movida da `02`, porque trata de fronteiras de processamento e contratos de evento). **Spec aprovada; Design escrito** | 011, 012, 014 | **G1** |
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

### Decisões do G1 (consolidadas em 2026-10-14, após o parecer sobre `5aef06e`)

**Nenhuma concessão real foi alterada no código**; as concessões valem como especificação e design.

| # | Decisão | Sub-spec | Conteúdo | Situação |
|---|---|---|---|---|
| D-02-1 | Contas de serviço: desativação e reativação | 02 | `ADMIN_SISTEMA` e `PRESIDENTE` recebem `identity:service_account:deactivate` e `:reactivate` por concessão explícita. Reativação por pessoa diferente. Contingência quando só existe uma pessoa habilitada: espera configurável (inicial 30 min, não suficiente sozinha), motivo, identificação da emergência, auditoria, alerta e revisão independente que essa pessoa **não** pode fechar (SVC-01 AC16–22; design `02` C4) | **Aprovada** |
| D-02-2 | Classificação das 30 permissões existentes | 02 | 18 `sensivel` e 12 `normal`; 16 concessões explícitas de transição ao `PRESIDENTE`, só para preservar acessos atuais, com justificativa individual e testes de compatibilidade; permissões sensíveis novas sem concessão automática | **Aprovada** |
| D-02-3 | Retirar do `PRESIDENTE` os atos do Conselho Fiscal | 02 | Retirada de `financeiro:prestacao_contas:approve` e `financeiro:parecer:opine`, junto com a D-02-4 (TRN-01 AC8) | **Aprovada**; a ser aplicada no incremento I-6 do design `02` |
| D-02-4 | Nomeação e destituição institucional do Conselho Fiscal | 02 | Autorização específica (`identity:conselho_fiscal:manage`); referência verificável ao ato da instância competente (documento armazenado); exceção de `Covers` só nas operações listadas (nomeação/destituição e bloqueio técnico; **nunca** a redefinição administrativa de senha); `PRESIDENTE` × `CONSELHO_FISCAL` incompatíveis; usuário já irregular perde só os atos institucionais até regularizar; outras incompatibilidades só pelo estatuto (INS-01; design `02` C7/C8) | **Aprovada como direção arquitetural** |
| D-02-6 | Quem autoriza a reversão de versão do RBAC | 02 | Permissão nova `identity:rbac:rollback_authorize` (`sensivel`); concessão proposta explícita a `ADMIN_SISTEMA`. A decisão humana (autorização pela API, `actor_type = user`) é separada da execução técnica (`system`, com referência à pessoa). Reversão que restaure permissões institucionais exige decisão expressa do mantenedor com documento, registrada antes (design `02` C10; TRN-01 AC12–13) | **Proposta**, pendente; sem concessão até a decisão (fail-closed). Não bloqueia I-1 nem I-2 |
| D-02-5 | Quem registra os atos do Conselho Fiscal | 02 | Concessão explícita de `identity:conselho_fiscal:manage` ao `PRESIDENTE`, limitada ao **registro** de nomeações e destituições formalmente decididas pela instância competente, com evidência documental e auditoria. Não confere competência para decidir unilateralmente a composição do Conselho Fiscal. Aplicada com a D-02-3 (I-6; tasks `02` T29). `jobs:job:close` continua sem concessão | **Aprovada** (2026-10-14) |
| D-01-1 | Operação de jobs | 01 | `ADMIN_SISTEMA` consulta e reenfileira; `PRESIDENTE` consulta por concessão explícita; jobs com efeito financeiro ou associativo exigem a política de reprocessamento do módulo dono (DLQ-01 AC9); `jobs:job:close` sem concessão no G1; encerramento nunca oculta obrigação (DLQ-01 AC6–8) | **Aprovada** |
| D-01-2 | Parâmetros e retenção | 01 | Valores configuráveis; retenção só com deduplicação durável preservada (RET-01) | **Mantida** |
| D-G1-1 | Proteção **versionada** contra a inclusão dos PDFs e de `docs/tj-connect/interno/` | — | `.gitignore` atualizado | **Resolvida** (commit `40b5165`) |

**Pendência que bloqueia só a implantação de I-6**: provedor de e-mail real em produção (recuperação de acesso dos conselheiros, que deixa de ter redefinição administrativa). A concessão de `manage` foi resolvida pela D-02-5.

### Ajustes obrigatórios da aprovação do Design G1 (2026-10-14)

| # | Ajuste | Onde foi tratado | Tasks |
|---|---|---|---|
| 1 | Sincronização de RBAC segura durante o deploy (o advisory lock só serializa; não decide qual matriz vence) | Design `02` C10 (versão da matriz, hash conferido por teste, barreira por gatilho, worker só verifica, ordem de implantação, `rbac-rollback` auditado); spec `02` TRN-01 AC9–12, T-28 | `02`: T10, T11, T12, T32; `01`: T13, T25, T27 |
| 2 | Identidade da entrega de eventos | Design `01` P4 "Identidade da entrega" (conta gravada só pela plataforma a partir da inscrição, revalidada a cada execução, passos internos sem `Principal`, tratador só com a sua conta); spec `01` SVX-01 AC1, AC7–9, T-13b/T-13c | `01`: T5, T15, T17, T18 |
| 3 | Integridade das contas de serviço (tipo imutável; restrições no banco em todos os caminhos de escrita) | Design `02` C5 "Integridade das contas de serviço no banco" e migration `000010`; spec `02` SVC-01 AC23–24, T-29 | `02`: T9 |
| 4 | Recuperação de concessões: falha do executor × falha de infraestrutura, sem abandonar obrigações | Design `01` P2/P3, **revisado após a revisão das Tasks**: toda perda de concessão é recuperável (não consome tentativa, nunca `FALHOU`, espera até 15 min, alerta `job_suspeito_de_derrubar_worker`); só falha identificada pelo runner com concessão válida consome tentativa (erro, panic, `tempo_esgotado`, `executor_travado`); batimento só como diagnóstico; `obrigacao_em_falha` diário; `WORKER_MAX_LEASE_LOSSES` removido. Spec `01` REC-01 AC7–12, T-06b/T-07b/T-07c | `01`: T11, T12, T14, T21 |

### Revisão preliminar das Tasks do G1 (2026-10-14)

Estrutura aprovada preliminarmente (32 + 27 tarefas), condicionada à revisão externa dos arquivos completos. Correções feitas antes da publicação:

| # | Ponto | Correção |
|---|---|---|
| 1 | D-02-5 sem decisão unilateral | Instância competente restrita a `CompetentBodies` (inicialmente só a Assembleia Geral; `CHECK` no banco), data não futura, documento obrigatório, registrador gravado como registrador e a instância como autora; nenhum outro caminho altera a composição (INS-01 AC12; tasks `02` T27, T29) |
| 2 | Barreira de versão do RBAC | Monotonicidade no banco (`rbac_sync_state_monotonic`, `tj_app` sem `INSERT`/`DELETE`), histórico `rbac_matrix_versions`, gatilho testado nas três tabelas e três operações, rollback que mostra e exige confirmação das concessões restauradas e grava a lista (`rbac_version_rollbacks`) antes de reduzir a versão (TRN-01 AC12; tasks `02` T8, T11, T12) |
| 3 | Recuperação de jobs | Classificação por batimento **descartada**: toda perda de concessão é recuperável (sem consumir tentativa, nunca `FALHOU`, espera até 15 min, alerta após 5); só falha identificada pelo runner com concessão válida consome tentativa (erro, panic, `tempo_esgotado`, `executor_travado`); batimento só como diagnóstico (REC-01 AC7–8; design `01` P3; tasks `01` T3, T11, T12) |
| 4 | Avisos dos validadores | 35 avisos justificados um a um nas seções "Avisos do validador e justificativas"; seção "Compatibilidade entre código e banco ao longo da sequência" em cada arquivo de tasks |
| 5 | Revisão externa das Tasks (2026-10-14) | Dependência `01/T22` → `02/T29` e ordem consolidada corrigidas; `jobs:job:read` como `sensivel` com concessões explícitas; `rbac-rollback` em três passos (plano, autorização humana auditada, execução técnica), confirmação pelo `plan_hash` completo, decisão do mantenedor para reverter a D-02-3; variável de sessão documentada como barreira de compatibilidade, não autenticação |
### Matriz de permissões proposta (catálogo atual e permissões novas do G1)

Legenda: **A** = decorre de decisão aprovada; "=" = sem mudança. Nenhuma concessão pendente desde a D-02-5. Detalhe por permissão e local no código: design `02`, seção "Matriz de permissões: antes e depois".

| Papel | Hoje | Proposta | Mudanças |
|---|---|---|---|
| `PRESIDENTE` | 30 (todas automáticas) | **33**: das 30 existentes, 28 (12 `normal` automáticas + 16 `sensivel` explícitas com justificativa); mais 5 novas | **A**: −`financeiro:prestacao_contas:approve`, −`financeiro:parecer:opine` (D-02-3, com a D-02-4); 16 concessões explícitas de transição (D-02-2); +`identity:service_account:deactivate`, +`:reactivate` (D-02-1); +`jobs:job:read` (D-01-1); +`identity:service_account:read` (normal, automática); +`identity:conselho_fiscal:manage` (D-02-5, só a partir de I-6). Até I-6, `approve` e `opine` ficam como concessões explícitas transitórias |
| `ADMIN_SISTEMA` | 8 (`identity` ×7, `audit:log:read`) | **13** | **A**: +`identity:service_account:read`, +`:deactivate`, +`:reactivate` (D-02-1); +`jobs:job:read`, +`jobs:job:retry` (D-01-1) |
| `CONSELHO_FISCAL` | 11 | **11** = | Agora na lista explícita (CFS-01); mantém `prestacao_contas:approve` e `parecer:opine` |
| `TESOURARIA` | 13 | **13** = | 9 delas passam a `sensivel` (D-02-2); a concessão já é explícita no código |
| `ESTOQUE_LOJA` | 6 | **6** = | 2 passam a `sensivel` (D-02-2); concessão já explícita |
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

### Requisito → design → tasks (G1)

| Requisito | Design | Tasks |
|---|---|---|
| CLS-01 | `02` C1, C2 | `02`: T2–T6, T15 |
| TRN-01 | `02` C2, C10, snapshot | `02`: T1, T3–T6, T8, T10–T13, T21, T29, T31, T32; `01`: T25 |
| SVC-01 | `02` C3, C4, C5, C10 | `02`: T8, T9, T13–T16, T18–T20 |
| SOD-01 | `02` C1, C9 | `02`: T2, T17, T24 |
| AUT-01 | `02` C6 | `02`: T12, T21–T23 |
| FIS-01 | `02` C3 | `02`: T13 |
| CFS-01 | `02` C2 | `02`: T7 |
| INS-01 | `02` C1, C7, C8 | `02`: T17, T25–T30 |
| JOB-01 | `01` P1 | `01`: T3–T5, T25 |
| LEA-01 | `01` P2 | `01`: T8, T9, T14 |
| RTY-01 | `01` P2, P6 | `01`: T9, T20 |
| REC-01 | `01` P2, P3 | `01`: T4, T9, T11, T12, T14, T21, T27 |
| OUT-01 | `01` P4, P7 | `01`: T1, T15, T18 |
| INB-01 | `01` P4 | `01`: T16, T18 |
| SVX-01 | `01` P2, P4, P9; `02` C4, C6 | `01`: T4–T6, T10, T13, T15, T17, T23 |
| DLQ-01 | `01` P8 | `01`: T4, T22–T24 |
| SCH-01 | `01` P5 | `01`: T19 |
| RET-01 | `01` P6 | `01`: T20 |
| OBS-01 | `01` P8 | `01`: T7, T11, T13, T23, T27 |
| ARQ-01 | `01` P7 | `01`: T1, T2, T13 |

---

## Decisões fechadas (TJC-D-NNN)

Nenhuma ainda. As decisões de arquitetura estão nas ADRs 011–018 (`Proposed`); as de produto, em `context.md` (M-1 a M-9).

## Handoff

- **Fase**: Design do G1 aprovado com quatro ajustes obrigatórios (2026-10-14), já incorporados; **Tasks do G1 escritas, aprovadas preliminarmente e publicadas para revisão externa**. Execute, G2, deploy e merge não autorizados. ADRs continuam `Proposed`.
- **Branch**: `feature/tj-connect-discuss`, publicada para revisão externa (sem PR, sem merge na `develop`).
- **Arquivos**:
  - `.specs/features/tj-connect/{context.md,STATE.md}`;
  - `docs/adr/011`–`018` e o índice em `docs/adr/README.md`;
  - nota de reconciliação no Handoff de `.specs/STATE.md`.
- **Specs do G1**: `.specs/features/tj-connect/spec/02-identity-governanca.md` e `spec/01-platform-processamento.md` (atualizadas com os limites da D-02-4, a contingência da D-02-1 e a política de reprocessamento da D-01-1).
- **Designs do G1**: `.specs/features/tj-connect/design/02-identity-governanca.md` e `design/01-platform-processamento.md` (com os ajustes 1–4 e a D-02-5).
- **Tasks do G1**: `.specs/features/tj-connect/tasks/02-identity-governanca.md` (32 tarefas, I-1 a I-6 + preparação) e `tasks/01-platform-processamento.md` (27 tarefas, J-1 a J-6 + preparação).
- **Ordem consolidada (corrigida em 2026-10-14)**:
  1. `01` T1–T2 (J-1; sem dependências, a qualquer momento);
  2. `02` T1–T6 (I-1) e T7 (I-2);
  3. `02` T8–T20 (I-3);
  4. `02` T21–T23 (I-4);
  5. `02` T24 (I-5) e T25–T30 (I-6: D-02-3, D-02-4, D-02-5);
  6. `01` T3–T24 (J-2 a J-6). **`01/T22` depende de `02/T29`**, porque verifica a matriz final de permissões;
  7. `02` T31–T32 e `01` T25–T27 (verificação final e preparação, sem implantar).
- **Execute autorizado (2026-10-14)** só para J-1 (`01` T1–T2) e I-1 (`02` T1–T6), na branch `feature/tj-connect-g1-execute`, com commits locais sem push. I-1 preserva integralmente a matriz efetiva atual, inclusive as concessões transitórias de `approve`/`opine` ao `PRESIDENTE`.
- **Próximo passo**: revisão de J-1 e I-1 pelo mantenedor antes da fase seguinte.
