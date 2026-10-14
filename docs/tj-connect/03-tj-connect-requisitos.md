# TJ Connect 2027 — Requisitos Funcionais Consolidados

> **Projeto:** Plataforma Torcida Jovem Campo Mourão Futsal  
> **Versão:** 1.0 — baseline funcional para Specify/Discuss  
> **Estado:** Documento de negócio consolidado; **não é autorização de implementação**  
> **Data de consolidação:** outubro de 2026  
> **Público:** Presidência e diretoria TJ, responsáveis de produto, arquitetura, UX, desenvolvimento, testes e assessoria jurídica.

---

## 0. Como utilizar este documento

Este arquivo registra as **decisões de produto expressamente aprovadas** durante a discussão do TJ Connect 2027. Ele deve acompanhar **separadamente** os documentos históricos:

1. `01-plano-estrategico-tj-2027-2030.pdf` — fundamento institucional, proposta de associação, benefícios e objetivos de 2027–2030 (documento interno, não versionado).
2. `02-tj-connect-apresentacao.pdf` — visão de jornadas, experiências de associado, fiscal e diretoria e referências conceituais de UX (documento interno, não versionado).
3. **`03-tj-connect-requisitos.md` (este arquivo)** — regras consolidadas, alterações aprovadas, restrições e questões ainda abertas.
4. **Código do repositório existente** — fonte para arquitetura atual, contratos já implementados, infraestrutura, autenticação, módulos e padrões de desenvolvimento; deve ser inspecionado na branch de trabalho atual.

### 0.1. Hierarquia em caso de conflito

- Para **comportamento funcional desejado**: decisões aprovadas neste documento prevalecem sobre propostas antigas contidas nos PDFs.
- Para **estado técnico atual do sistema**: o repositório inspecionado é a fonte de verdade. Os detalhes técnicos mencionados neste documento são contexto e **precisam ser confirmados** no código.
- Para **normas legais, termos do clube e do intermediador de pagamento**: legislação aplicável, instrumentos formalizados e documentação vigente prevalecem. Este documento **não substitui** validação jurídica/contratual.
- Requisitos aqui marcados **PENDENTE** ou **PROPOSTA** não devem ser tratados como decisão aprovada. Em caso de conflito entre documentos, abrir registro de divergência, sem presumir a solução.

**Regra de execução:** o primeiro contato do Claude Code com estes materiais é de **leitura, descoberta e especificação**, sem gerar migrations, endpoints, telas, cobranças nem commits de implementação.

### 0.2. Convenções

- **APROVADO:** decidido pelo responsável do produto em conversa.
- **DIRETRIZ:** consequência lógica/critério necessário para cumprir decisão aprovada; sujeita a refinamento técnico.
- **PENDENTE:** decisão de negócio ou validação externa ainda necessária.
- **PROPOSTA:** solução técnica ou UX útil para discussão; não obrigatória.
- Identificadores `TJ-*` servem para rastreabilidade. IDs que já haviam sido estabelecidos nas conversas foram preservados; novos foram adicionados para organização.

---

## 1. Contexto, objetivo e princípios

### 1.1. Contexto

A Torcida Jovem do Campo Mourão Futsal já mantém uma plataforma web com área administrativa e módulos de identidade/acesso, Financeiro e Estoque. O TJ Connect amplia **a mesma plataforma** para atender torcedores, associados, fiscais credenciados e administradores. Não é uma aplicação independente, não é inicialmente um aplicativo mobile nativo e não deve duplicar cadastros ou operações financeiras.

A plataforma centraliza o relacionamento associativo e ajuda a gerir benefícios decorrentes de uma parceria com o **clube Campo Mourão Futsal**. A **Torcida Jovem** e o **clube** possuem direções distintas, competências distintas e não devem ser tratados como uma única entidade administrativa.

### 1.2. Objetivos

1. Tornar a Torcida Jovem acessível publicamente por um site institucional.
2. Permitir criação de conta gratuita e posterior associação em planos anuais.
3. Arrecadar contribuições com conciliação confiável e baixo trabalho manual.
4. Garantir que a contratação de planos com ingresso respeite a quota pactuada com o clube.
5. Disponibilizar carteirinha digital e uma jornada de presença simples, segura e auditável.
6. Produzir estimativas de público para o clube, sem condicionar a entrada à resposta antecipada.
7. Dar à diretoria governança sobre frequência, procedimentos individualizados e histórico.
8. Gerir camisetas personalizadas sob encomenda e por lotes, aproveitando o Estoque existente.
9. Preservar privacidade dos associados, permissões de acesso e identidade visual existente.

### 1.3. Princípios não negociáveis

- **Uma plataforma:** site, Meu TJ, operação fiscal e administração usam a mesma identidade e backend existente.
- **Uma conta não é uma associação:** registro de torcedor, adesão paga, ingresso da temporada e poderes administrativos são dimensões separadas.
- **Quota é um limite real:** nenhuma concorrência, expiração ou compensação de pagamento permite sobrecontratar o direito a ingresso.
- **Ingresso não é transferido por ausência:** intenção, não resposta, ausência ou suspensão temporária jamais geram vaga para outro associado.
- **Direitos não são suspensos automaticamente por percentual:** exige procedimento humano, motivação e garantias previstas no regulamento válido.
- **Checkout não prova pagamento:** ativação somente mediante confirmação segura pelo servidor.
- **QR TJ não substitui ingresso do clube.**
- **MVP online:** não haverá validação offline nem sincronização posterior de check-ins.
- **Sem prejuízo ao sistema atual:** evolução incremental, com proteção dos módulos e dados existentes.

---

## 2. Escopo do produto

### 2.1. Experiências contempladas

| Experiência | Necessidade |
|---|---|
| Visitante público | Conhecer a torcida, trajetória, planos, informações e conteúdos públicos de transparência conforme publicação autorizada. |
| Torcedor cadastrado gratuito | Criar conta e entrar na plataforma sem pagar, acompanhar oportunidades e contratar uma associação posteriormente. |
| Associado pago | Acompanhar plano, situação, carteira, jogos, presença, pedidos de camiseta, pagamentos e solicitações. |
| Associado Arquibancada | Validar fotografia, verificar autorização para jogos, apresentar credencial ao fiscal e acompanhar frequência. |
| Fiscal TJ | Verificar credenciais online e registrar check-ins confiáveis em jogos para os quais esteja autorizado. |
| Presidente/diretoria TJ | Administrar planos, temporadas, quota, adesões, pagamentos e incidentes, jogos, presença, procedimentos e lotes. |
| Diretoria do clube | Receber estimativas/relatórios autorizados para planejamento; não ganha automaticamente função administrativa na TJ. |

### 2.2. Escopo pretendido para a primeira entrega funcional

- Home institucional pública, apresentação dos três planos e chamadas à associação.
- Cadastro e autenticação centralizados, com cadastro gratuito.
- Temporadas, catálogo de planos e adesão anual.
- Reserva de quota para checkout e integração Mercado Pago.
- Histórico de cobranças, solicitações de cancelamento e tratamento de exceções.
- Portal do associado, identificação e carteirinha digital.
- Cadastro administrativo de jogos e intenção opcional de presença.
- Controle de ingresso/benefício, QR e check-in exclusivamente online.
- Frequência, avisos e suporte ao procedimento de suspensão individual.
- Encomendas de camisetas, lotes, recebimento e entregas.
- Integração controlada com Financeiro e Estoque existentes.
- Trilhas de auditoria, permissões, privacidade e operação administrativa.

**PENDENTE de fatiamento:** antes das Tasks, definir se tudo compõe um único release de MVP ou entregas incrementais com lançamento condicionado ao funcionamento seguro dos módulos críticos. A funcionalidade anunciada ao público não pode ser vendida sem os controles de quota, pagamento e benefício correspondentes.

### 2.3. Explicitamente fora do MVP

- Aplicativo nativo iOS/Android.
- Validação offline ou PWA de check-in offline.
- Apple Wallet / Google Wallet nativos.
- Renovação automática/assinaturas recorrentes.
- Prioridade especial para associados do ano anterior.
- Redistribuição de vaga por ausência, confirmação negativa ou suspensão temporária.
- Bilheteria do clube, substituição de ingresso oficial ou controle das catracas do clube sem acordo e integração próprios.
- Marketplace/loja virtual completa nesta especificação; deve ser compatível com implementação futura.
- Campanhas avançadas de cupons, fidelidade, transferência de credenciais e mercado secundário de ingressos.

---

## 3. Linguagem de domínio e modelo conceitual

| Termo | Definição de negócio |
|---|---|
| **Conta / Usuário** | Identidade para autenticação, com ou sem vínculo associativo pago. |
| **Papel / Permissão** | Autorização para administrar recursos ou operar como fiscal; independente do pagamento de plano. |
| **Associado** | Pessoa que contratou uma associação; possui status e histórico por temporada. |
| **Temporada** | Exercício anual do plano, com preços, vigência e quota próprios (p.ex. temporada 2027). |
| **Plano** | Modalidade comercial: Associado TJ, Arquibancada TJ ou Arquibancada TJ + Camisa. |
| **Adesão / Associação da temporada** | Compra e vínculo histórico de usuário, plano, temporada, pagamento, benefícios e vigência. |
| **Benefício de ingresso** | Direito contratual administrado pela TJ, condicionado ao acordo com o clube, válido durante o período aplicável e sujeito a suspensão regular. |
| **Quota garantida** | Total máximo de titulares ativos ou temporariamente reservados dos planos com ingresso na temporada, conforme garantia do clube. |
| **Reserva de checkout** | Bloqueio temporário e exclusivo de uma unidade da quota, antes da confirmação definitiva. |
| **Intenção de presença** | Resposta opcional para previsão de público; não é reserva de ingresso nem check-in. |
| **Check-in** | Registro efetivo de presença em jogo elegível, validado pelo fiscal TJ online. |
| **Frequência** | Razão entre presenças e jogos elegíveis para determinado associado, de acordo com o regulamento. |
| **Procedimento de frequência** | Análise individual formal conduzida pela TJ, distinta de simples alerta. |
| **Suspensão temporária** | Restrição datada apenas ao benefício de ingresso; não encerra conta, associação ou quota. |
| **Encomenda de camiseta** | Obrigação de fornecer uma camiseta ao comprador elegível, com tamanho e nome opcional. |
| **Lote de produção** | Agrupamento de encomendas destinado a um fornecedor, com fechamento controlado. |

**Alerta arquitetural:** o papel técnico preexistente `ASSOCIADO`, se mantido no controle de acesso atual, **não deve** ser usado como indicador de assinatura paga, quitação, capacidade ou autorização de ingresso. A associação é uma entidade de negócio distinta.

---

## 4. Regras por domínio — cadastro e identidade

### TJ-ID-001 — Cadastro gratuito simplificado — **APROVADO**

- Qualquer torcedor pode criar conta sem pagar ou escolher plano.
- Campos iniciais: **nome completo, e-mail e senha**.
- Confirmar o e-mail antes de concluir contratação de associação.
- Uma identidade e uma sessão atendem site público autenticado, Meu TJ e administração conforme permissões.
- O cadastro não gera carteirinha ativa, plano vigente nem acesso a jogos.
- Preservar funcionalidades existentes de login, recuperação de acesso e administração de usuários, com adaptações seguras para autoatendimento público.

### TJ-ID-002 — Fotografia proporcional ao benefício — **APROVADO**

- Usuário gratuito: **foto não obrigatória**.
- Associado TJ (R$ 60): **foto não obrigatória**.
- Arquibancada TJ (R$ 160) e Arquibancada TJ + Camisa (R$ 240): **foto obrigatória antes do primeiro uso do benefício de ingresso**.
- É permitido contratar e quitar o plano antes de enviar a foto; a associação pode estar ativa, mas a **credencial de ingresso permanecerá não habilitada** até cumprir este requisito.
- Mudanças de foto precisam de controles de integridade, autorização e histórico, sem exposição pública indiscriminada.

### TJ-ID-003 — CPF e dados complementares — **APROVADO como etapa; PENDENTE quanto ao conjunto completo**

- CPF e outros dados efetivamente necessários são coletados **na contratação**, não no cadastro gratuito inicial.
- Validar formato e unicidade conforme a modelagem de titularidade e tratamento de conflitos.
- Permitir correção de CPF/nome por fluxo seguro e auditável; não tornar um erro cadastral irreparável.
- Evitar coleta excessiva de documentos e dados sensíveis.

### TJ-ID-004 — Segurança e privacidade — **DIRETRIZ**

- Verificação de e-mail, recuperação de acesso segura, proteção antiabuso em cadastro/login e proteção de sessão.
- Termos de associação, política de privacidade e bases legais adequadas aos usos dos dados.
- Fotos de identificação e CPF restritos aos contextos necessários, inclusive em telas de fiscais.
- Definir políticas de exclusão/retenção, atualização e acesso administrativo conforme obrigações legais e operacionais.

**PENDÊNCIAS:** tratamento de menores de idade e eventual responsável; detalhes de comprovação de identidade; mecanismo de atualização de foto; dados obrigatórios adicionais conforme fornecedor de pagamento e regras de associação.

---

## 5. Regras por domínio — temporadas, planos e adesões

### TJ-ASSOC-001 — Temporada e vigência — **APROVADO**

- Os três planos são **anuais por temporada**, com validade até **31 de dezembro** do ano contratado.
- **Preço integral** independentemente do mês da entrada; não há proporcionalidade automática.
- Nova temporada exige nova contratação e pagamento, **sem renovação automática**.
- Preservar histórico de planos, pagamentos, frequência e benefícios; manter identificação contínua do associado, sem reutilizar número para outra pessoa.
- Preços e catálogo configuráveis por temporada; valores abaixo referem-se à proposta 2027, não a constantes permanentes em código.
- Não há prioridade de renovação de ex-associados nesta versão.

### TJ-ASSOC-002 — Três planos de 2027 — **APROVADO**

| Plano | Valor anual 2027 | Direito a ingresso? | Camiseta? |
|---|---:|---|---|
| Associado TJ | **R$ 60,00** | Não | Não |
| Arquibancada TJ | **R$ 160,00** | Sim, conforme quota/condições do clube | Não |
| Arquibancada TJ + Camisa | **R$ 240,00** | Sim, conforme quota/condições do clube | Sim, sob encomenda |

Os benefícios institucionais adicionais descritos no Plano Estratégico (p.ex. carteira digital, condições especiais e participação) deverão ser convertidos em texto comercial verificável e regulamento antes da publicação.

**DIRETRIZ:** o plano básico não consome quota de ingresso e pode ter quantidade de associados independente dessa quota.

### TJ-ASSOC-003 — Cancelamento definitivo e vaga — **APROVADO**

- Ausência, resposta de não comparecimento, falta de resposta, pedido de cancelamento ainda pendente ou suspensão temporária: **não liberam vaga**.
- Apenas cancelamento **definitivo e efetivado** do benefício, com questões financeiras regularizadas e substituição permitida no acordo com o clube, pode liberar quota para novo titular.
- Invalidar prontamente autorizações anteriores quando o desligamento for efetivado.
- Manter conta gratuita e histórico do ex-associado.
- Operação de liberação deve ser atômica, autorizada e auditável.

### TJ-ASSOC-004 — Solicitação, cancelamento e reembolso — **APROVADO**

- Associado pode solicitar cancelamento no **Meu TJ**.
- Diretoria autorizada realiza análise administrativa da solicitação, benefícios já utilizados e consequências financeiras aplicáveis.
- Registrar solicitação, data, justificativa (quando pertinente), ator, decisão, comunicação e efetivação de reembolso/estorno, se devido.
- Integrar com Mercado Pago e Financeiro com reconciliação e prevenção de duplicidades.
- **A análise manual não pode afastar ou retardar indevidamente direitos legais obrigatórios**, inclusive quando houver direito de arrependimento ou restituição. O desenho deverá contemplar prazos legais e exceções.
- Cancelar associação e reembolsar pagamento não são eventos automaticamente equivalentes: modelar consistência entre seus estados.

### TJ-ASSOC-005 a 008 — Camiseta e personalização — **APROVADO**

Ver seção 10 para detalhamento de encomenda, lotes, nome opcional e integração com Estoque.

---

## 6. Regras por domínio — quota de ingresso e checkout

### TJ-CAP-001 — Quota garantida por temporada — **APROVADO**

- Quota `N` definida conforme garantia do **clube Campo Mourão Futsal** para a temporada; `60` é **somente exemplo histórico**, jamais número fixo universal.
- Os **dois planos Arquibancada compartilham** a mesma quota.
- O sistema nunca pode ter simultaneamente mais titulares elegíveis/comprometidos e reservas válidas que a capacidade contratada.
- O direito decorre da parceria; ele não substitui exigências de ingresso oficial ou controle de acesso do clube.
- Alterações administrativas de capacidade exigem permissão, justificativa e auditoria; não reduzir abaixo dos compromissos vigentes sem resolução formal e compatível com direitos adquiridos.
- Toda venda de plano Arquibancada depende de existência de quota garantida efetivamente aprovada para a temporada.

### TJ-CAP-002 — Sem redistribuição por jogo — **APROVADO**

- Quota é **ocupação por titular/associação da temporada**, não inventário de assentos recomercializável por jogo.
- Não comparecimento, intenção negativa, ausência de resposta e suspensão temporária **não** devolvem vaga ao estoque de planos.
- Intenção de presença serve somente para previsão de público.
- Cancelamento definitivo pode liberar uma vaga apenas nas condições de `TJ-ASSOC-003`.

### TJ-PAG-002 — Reserva temporária no checkout — **APROVADO**

- Ao iniciar checkout de plano com ingresso, criar reserva **exclusiva e temporária** de uma unidade de quota.
- Calcular disponibilidade considerando compromissos efetivos e todas as reservas ainda válidas.
- Transações concorrentes não podem reservar a mesma unidade de quota ou exceder `N`.
- Reserva converte-se em ocupação definitiva da temporada quando pagamento e validações forem concluídos.
- Reserva é liberada após expiração/cancelamento **e reconciliação suficiente** do estado real da cobrança; não confiar apenas em temporizador local se pagamento ainda puder ser aprovado no provedor.
- Alinhar janela da reserva à validade/modalidade da cobrança efetivamente oferecida pelo Mercado Pago. **Prazo exato ainda PENDENTE**.
- Registrar estados, timestamps e identificadores para investigação.

### TJ-PAG-003 — Pagamento tardio após expiração — **APROVADO**

- Ao receber confirmação tardia, fazer novamente checagem **atômica** de capacidade e demais condições.
- Se houver vaga disponível, associação pode ser concluída sem ultrapassar quota.
- Se não houver vaga, **não ativar ingresso nem exceder quota**, mesmo que o dinheiro tenha sido recebido.
- Criar pendência de conciliação/regularização, alertar administração e comprador; oferecer solução compatível com direitos do comprador, como restituição integral quando aplicável, ou alternativa apenas com **concordância expressa**.
- Nunca converter unilateralmente para o plano básico nem criar receita associativa definitiva fictícia.
- Manter eventos de cobrança, investigação, correções e eventual estorno rastreáveis.

### TJ-CAP-003 — Invariantes de concorrência — **DIRETRIZ técnica obrigatória**

- Ativação e reserva devem operar sob garantias transacionais no banco/API, não sob simples checagem apenas no frontend.
- Webhooks duplicados, retries e múltiplos dispositivos não podem ativar duas vezes nem movimentar mais de uma vaga.
- Bloqueio de venda deve considerar indisponibilidade real, sem depender de número exibido com atraso na UI.
- Padrões de idempotência e mecanismos de expiração/compensação devem ser definidos no Design.

---

## 7. Regras por domínio — Mercado Pago e Financeiro

### TJ-PAG-001 — Formas de pagamento — **APROVADO**

- Provedor pretendido: **Mercado Pago**.
- **Pix** como opção principal.
- **Cartão de crédito** como alternativa, à vista ou parcelado em **até 3 vezes**.
- O comprador suporta acréscimos aplicáveis à modalidade cartão **quando efetivamente permitidos/implementáveis nas condições do provedor e pela lei**, com divulgação clara do total, parcelas, encargos e diferença de preço **antes da confirmação**.
- Não presumir que taxas de processamento do vendedor sejam automaticamente transferidas ao comprador ou que a configuração de juros do parcelamento reproduza a taxa de adquirência. **Validar** API/produto, modelo de checkout, parcelamento e regime de taxas com Mercado Pago e assessoria financeira/jurídica.
- Não há cobrança recorrente automática.

### TJ-PAG-004 — Confirmação e conciliação — **DIRETRIZ**

- O sucesso mostrado no browser ou retorno do checkout **não** confirma quitação.
- Confirmar transação e valor/identidade/ordem por servidor e provedor, com autenticidade da notificação, consulta de status e proteção contra repetição.
- Separar estados do pedido, da transação no Mercado Pago, da reserva de quota e da adesão efetiva.
- Cobrir pagamento pendente, aprovado, recusado, cancelado, expirado, estornado, contestado/chargeback e casos inconclusivos.
- Conciliação recuperável após indisponibilidade da API, webhook não entregue, latência e intervenção administrativa.
- Não criar duas associações ou dois lançamentos financeiros para o mesmo fato gerador.
- Credenciais e ambiente do Mercado Pago sob gestão segura, sem exposição no cliente/repositório.

### TJ-PAG-005 — Identidade do recebedor — **PENDENTE / validação externa**

Antes da comercialização, confirmar a conta recebedora definitiva no Mercado Pago: titularidade institucional, elegibilidade, credenciais, termos e obrigações fiscais. Uma eventual troca de recebedor deve considerar o impacto sobre cobranças já iniciadas, **sem pressupor** que seja possível apenas trocar a chave da API.

### TJ-FIN-001 — Financeiro existente — **APROVADO como integração; detalhamento PENDENTE**

- Reutilizar o módulo **Financeiro** já existente para registrar entradas e estornos, conforme o seu modelo de contas, categorias, centros de custo e lançamentos.
- A contribuição paga precisa de um vínculo rastreável com pedido, associado, temporada e transação externa.
- Prever lançamento/reversão idempotente e reconciliação de taxas, líquidos/brutos e despesas do provedor quando pertinente.
- Revisar tratamento de competência, caixa, recebíveis, parcialidades e conciliação antes de implementar.
- Não copiar ou recriar o Financeiro dentro do módulo de associados.

---

## 8. Regras por domínio — jogos, previsão de público e entrada

### TJ-JOG-001 — Administração de jogos — **APROVADO**

- **Presidente da Torcida Jovem** cadastra e administra jogos; outras pessoas da diretoria somente conforme permissões delegadas.
- Distinguir jogos elegíveis a benefício (principalmente partidas em casa) de partidas informativas/fora de casa, amistosos etc.
- Registrar informações de partida, status, data/horário, adversário e local; prever ajustes por adiamento/cancelamento.
- Definir elegibilidade de jogos e cobertura do acordo com o clube antes de disponibilizar autorizações de ingresso.

### TJ-JOG-002 — Intenção opcional de comparecimento — **APROVADO**

- A cada partida elegível, associado pode indicar **vou / não vou / não respondeu**.
- Responder é **facultativo**; não há prazo de resposta que bloqueie ingresso.
- Mesmo após responder “não vou”, o associado elegível pode comparecer normalmente, sujeito às regras válidas no dia e ao ingresso oficial do clube.
- Não responder não causa punição ou perda de direito.
- Não redistribuir vaga por resposta negativa, ausência de resposta ou intenção alterada.
- Tratar dados como estimativa, nunca como comparecimento efetivo.

### TJ-JOG-003 — Previsão para a diretoria do clube — **APROVADO**

- A TJ poderá gerar/resumir intenção de comparecimento para auxiliar a **diretoria do Campo Mourão Futsal** no planejamento do jogo.
- A direção do clube não se confunde com a diretoria TJ e não recebe automaticamente permissões para gerenciar associação, suspensões ou finanças.
- Preferir informação agregada e acesso restrito/compartilhamento controlado, conforme finalidade e parceria.
- Identificar claramente quantidade de confirmados, negativas, não respondentes e público efetivamente validado; não confundir estimativa com total real.

### TJ-JOG-006 — Carteirinha digital e QR — **APROVADO quanto à finalidade; mecânica exata PENDENTE**

- Associado ativo pode consultar sua carteirinha digital, número único, identificação, plano, vigência e condição de benefício conforme permissões.
- Para ingresso, autorizar **por jogo**, com consulta atual ao servidor: temporada, plano elegível, pagamento/vínculo, fotografia e eventuais suspensões/impedimentos.
- A foto é condição para uso de benefício em planos Arquibancada, mas não para comprar ou manter o plano.
- Fiscal TJ credenciado usa câmera no navegador web, **exclusivamente online**.
- Na leitura, mostrar o mínimo necessário para comparação visual e status de autorização; registrar decisão, fiscal, horário, jogo, resultado e motivo de recusa.
- Um associado não pode ter duas entradas registradas no mesmo jogo; bloqueio deve ser transacional, inclusive com dois fiscais escaneando simultaneamente.
- QR e tokens não devem expor CPF/dados pessoais em claro nem ser aceitos apenas por validação estática do cliente.
- Política de duração/rotação (ex.: 30 segundos no protótipo) é uma **PROPOSTA NÃO APROVADA** e deve ser desenhada tecnicamente, considerando usabilidade em entrada de ginásio e prevenção de reutilização.
- Check-in TJ **não é** ingresso oficial ou mecanismo de abertura de catraca do clube; registrar essa distinção no produto e nas instruções do fiscal.
- Definir exceções e correções administrativas de presença com auditoria antes da implementação.

---

## 9. Regras por domínio — frequência, alertas e suspensão

### TJ-JOG-004 — Mínimo de seis jogos elegíveis — **APROVADO**

- Referência de presença mínima: **50% dos jogos elegíveis** do associado, com base em check-ins reais.
- Abertura de procedimento de possível suspensão por baixa frequência **somente após pelo menos 6 jogos elegíveis concluídos** para o respectivo associado.
- Jogos ocorridos antes do início válido do benefício não entram no denominador.
- Jogos cancelados/não elegíveis não entram no denominador.
- Jogos durante período de suspensão em que a pessoa não podia utilizar o benefício não são contabilizados como faltas.
- Intenção antecipada, inclusive “vou”, não equivale a check-in e não compõe o numerador.
- O conceito completo de jogo elegível, período de carência e regras excepcionais deve ser fixado em regulamento antes do cálculo produtivo.

### TJ-JOG-005 — Suspensão individual e temporária — **APROVADO**

- Frequência abaixo de 50% **pode** originar procedimento, nunca suspensão automática baseada apenas em cálculo.
- TJ deverá notificar formalmente, permitir apresentação de justificativas e defesa, apreciar individualmente, registrar decisão motivada e assegurar via de contestação/revisão conforme regulamento.
- Se aprovada, suspensão terá **período determinado**, datas de início e término e possibilidade de restabelecimento antecipado motivado.
- Somente suspensão **validamente decidida e vigente** bloqueia a credencial de ingresso e check-ins correspondentes; alerta e investigação, isoladamente, não bloqueiam.
- Associação permanece paga/ativa durante suspensão, salvo outro fundamento legítimo e processo próprio.
- Vaga de temporada continua ocupada e não pode ser revendida/redistribuída.
- Jogos no período de impedimento não geram novas ausências para fins do percentual.
- Registrar responsável, comunicação, base fática, defesa, documentos, decisão, prazo, revisão e status com proteção de acesso.

### TJ-JOG-007 — Condição jurídica e de transparência — **PENDENTE crítica antes da comercialização**

A restrição de um benefício pago no próprio ano exige disciplina expressa, prévia e juridicamente válida no **estatuto/regulamento aceito na contratação**, com proporcionalidade, contraditório, critérios de elegibilidade, competência decisória, formas de ciência, prazos, revisão e tratamento de consequências patrimoniais (inclusive eventuais restituições). **Não estão decididos:** prazo de defesa, duração típica da suspensão, instância recursal e efeitos econômicos. Não preencher tais lacunas com suposições do código ou da IA.

---

## 10. Regras por domínio — camisetas sob encomenda

### TJ-ASSOC-005 — Produção sob encomenda — **APROVADO**

- O plano **Arquibancada TJ + Camisa** dá direito a uma camiseta produzida **sob encomenda**, não necessariamente disponível em estoque no momento do pagamento.
- Pedido individual é registrado após aprovação do pagamento/ativação conforme regras de negócio.
- Selecionar tamanho obrigatoriamente na contratação.
- A demora na produção **não bloqueia** o benefício de ingresso, desde que os demais requisitos estejam cumpridos.
- Prazo estimado de confecção/entrega deve ser comunicado antes da compra, sujeito a parâmetros do fornecedor.

### TJ-ASSOC-006 — Produção por lotes — **APROVADO**

- Diretoria cria, organiza e **fecha lotes manualmente**, conforme demanda e negociação com fornecedor.
- Não há fechamento automático em data periódica obrigatória.
- Datas de fechamento podem ser previstas/configuradas por lote e alteradas por responsável autorizado.
- Lote agrupa encomendas quitadas elegíveis, com totais por tamanho e relação de personalizações individuais.
- Impedir que a mesma camiseta entre em dois lotes ativos ou gere entregas duplicadas.
- Controlar pedido ao fornecedor, produção, recebimento e entrega individual.

### TJ-ASSOC-007 — Nome, sem número — **APROVADO**

- A personalização é **exclusivamente por nome**; não há número, apelidos extras em campos diferentes ou outras customizações.
- Nome informado pode ser distinto do nome civil do cadastro, sujeito à política de conteúdo/quantidade de caracteres do fornecedor.
- Nome e tamanho podem ser corrigidos antes do fechamento do lote.
- Depois do fechamento, alterações comuns ficam bloqueadas; excepcionalmente só com autorização administrativa e viabilidade confirmada com fornecedor, mediante registro.

### TJ-ASSOC-008 — Personalização opcional — **APROVADO**

- Associado pode pedir camiseta **sem nome**.
- Tamanho segue obrigatório mesmo sem personalização.
- A decisão de personalizar ou não não altera, por si só, o preço aprovado de R$ 240; eventual taxa adicional futura exigiria nova decisão comercial.

### TJ-EST-001 — Integração com Estoque — **DIRETRIZ**

- Reutilizar produtos/SKUs, entradas físicas, saídas e saldos do **Estoque existente**, quando a modelagem comportar roupas por tamanho e unidades individualizadas.
- **Não registrar estoque físico apenas porque um plano foi pago.** O pagamento gera obrigação/encomenda, não entrada de produto recebido.
- Recebimento de lote deve refletir quantidades realmente recebidas, inclusive entregas parciais e divergências do fornecedor.
- Entrega individual deve dar baixa idempotente conforme o padrão existente e vincular retirada à pessoa correta.
- Não expor nome personalizado desnecessariamente em relatórios públicos de estoque.
- **PENDENTE:** definição de tamanhos permitidos, limite de caracteres, posicionamento/natureza da personalização, regras de conteúdo, custos, fornecedores, prazo de produção, entrega/retirada e tratamento de devoluções e defeitos.

### 10.1. Estados de acompanhamento sugeridos — **PROPOSTA**

`aguardando_lote → incluída_em_lote → lote_fechado → pedido_enviado → em_produção → recebida/disponível → entregue`.

O desenho técnico deve distinguir estado do **lote**, da **encomenda individual** e da **movimentação física de estoque**, incluindo divergência, cancelamento e retorno, em vez de copiar uma única máquina de estados para os três.

---

## 11. Perfis e matriz preliminar de autorização

> **DIRETRIZ** — esta é a intenção funcional. Antes de codificar, inspecionar o RBAC implementado, os papéis existentes e as permissões granulares; não criar papéis duplicados ou escalonar privilégios por ser associado.

| Operação | Público | Torcedor / Associado | Fiscal TJ | Diretoria autorizada | Presidente TJ | Diretoria clube |
|---|---:|---:|---:|---:|---:|---:|
| Ver home, planos e conteúdo público | Sim | Sim | Sim | Sim | Sim | Sim |
| Criar conta gratuita | Sim | — | — | — | — | — |
| Consultar própria associação/carteirinha | Não | Próprio | Próprio | Próprio | Próprio | Não, salvo vínculo próprio |
| Comprar plano e consultar próprios pagamentos | Não | Próprio | Próprio | Próprio | Próprio | Não, salvo vínculo próprio |
| Informar intenção de comparecer | Não | Próprio | Próprio | Próprio | Próprio | Não, salvo vínculo próprio |
| Fazer check-in de outros associados | Não | Não | Jogos autorizados | Só com permissão | Só com permissão | Não |
| Administrar jogos | Não | Não | Não | Se delegado | Sim | Não |
| Analisar cancelamento/reembolso | Não | Não | Não | Se autorizado | Conforme permissão | Não |
| Administrar lotes/camisetas | Não | Não | Não | Se autorizado | Conforme permissão | Não |
| Decidir procedimento de suspensão | Não | Não | Não | Autoridade prevista no regulamento | Conforme regulamento | Não |
| Consultar estimativa agregada do clube | Segundo publicação | Próprios dados | Conforme função | Sim | Sim | Somente relatório autorizado |

**Regra:** papéis administrativos e associação paga podem coexistir na mesma conta. Um presidente pode também ser associado, sem duplicar CPF ou autenticação.

---

## 12. Jornadas essenciais

### J-01 — Visitante torna-se torcedor cadastrado

1. Abre site público da TJ e conhece planos.
2. Escolhe criar conta; informa nome, e-mail e senha.
3. Verifica o e-mail.
4. Acessa área pessoal gratuita; não possui acesso de gestão nem ingresso.
5. Decide posteriormente contratar um plano.

**Aceite:** criar conta não exige CPF, foto ou pagamento. Navegação para área administrativa exige permissão preexistente válida.

### J-02 — Adesão ao plano básico

1. Torcedor autenticado escolhe Associado TJ para a temporada aberta.
2. Complementa dados cadastrais obrigatórios e aceita termos aplicáveis.
3. Escolhe meio de pagamento, visualiza total e quita.
4. Servidor confirma pagamento de forma segura.
5. Associação é ativada, carteirinha é disponibilizada e lançamento financeiro é registrado sem duplicidade.

**Aceite:** plano básico não ocupa quota de Arquibancada e não autoriza check-in de ingresso.

### J-03 — Adesão a Arquibancada

1. Torcedor visualiza disponibilidade real, escolhe plano elegível.
2. Reserva transacional é criada com expiração e cobrança correspondente.
3. Realiza pagamento.
4. Após confirmação verificável, reserva torna-se ocupação de temporada.
5. Se ainda não possui foto, pode enviar depois, mas não usar benefício antes da aprovação/registro da foto.
6. Na partida coberta, apresenta credencial TJ e, separadamente, o ingresso/autorização oficial exigido pelo clube.

**Aceite:** duas compras simultâneas pela última vaga não geram sobrecontratação; pagamento tardio sem vaga é pendência, não autorização de ingresso.

### J-04 — Compra com camiseta

1. Escolhe Arquibancada TJ + Camisa.
2. Seleciona tamanho e, opcionalmente, nome de personalização.
3. Paga e acompanha associação e encomenda de forma independente.
4. Diretoria inclui encomenda em lote; até fechamento, associado pode modificar dados permitidos.
5. Diretoria fecha, envia e acompanha o lote, registra recebimentos.
6. Camiseta recebida fica disponível para retirada; entrega ao titular é registrada e baixa estoque.

**Aceite:** camiseta não recebida não aparece como saldo físico; encomenda nunca consta simultaneamente de lotes ativos incompatíveis.

### J-05 — Planejamento e jogo

1. Presidente TJ cadastra partida elegível.
2. Associados com direito podem indicar intenção de presença, sem obrigação.
3. Sistema mostra estimativa agregada para a TJ e para o clube conforme compartilhamento autorizado.
4. No dia, fiscal TJ verifica carteirinha/identidade/condições de ingresso online.
5. Registro atômico aceita somente um check-in válido por pessoa/jogo; auditoria guarda todas as tentativas relevantes.

**Aceite:** intenção negativa ou não resposta não impede um associado elegível de entrar; ausência não libera quota de temporada.

### J-06 — Baixa frequência e suspensão

1. Após pelo menos seis jogos concluídos elegíveis, o sistema calcula presença real.
2. Percentual abaixo de 50% sinaliza **possibilidade de análise**, sem bloquear ingresso.
3. TJ instaura procedimento, notifica, recebe justificativas, decide fundamentadamente e comunica.
4. Se houver decisão válida, benefício é restringido por período determinado; associação permanece ativa e quota ocupada.
5. Após término/reversão, benefício volta a ser elegível; partidas do período de impedimento não geram faltas.

**Aceite:** não existe bloqueio automático decorrente exclusivamente de percentual.

### J-07 — Cancelamento e reembolso

1. Associado solicita cancelamento no Meu TJ.
2. Administração acompanha o pedido, observa direitos legais, analisa benefício/entregas/pagamentos e decide.
3. Conforme o caso, efetua regularização/estorno e encerra vínculo ou benefício nos termos válidos.
4. Somente com cancelamento definitivo e substituição permitida a vaga pode ser liberada.
5. Conta gratuita e histórico permanecem disponíveis conforme retenção legítima.

**Aceite:** solicitação aberta não libera automaticamente quota nem apaga trilhas financeiras.

---

## 13. Superfícies e rotas — intenção, não imposição técnica

**Público**: página inicial `/` (atualmente, segundo contexto anterior, a raiz pode redirecionar para a área autenticada; **verificar no código**), história/sobre a TJ, página de planos/adesão, informações institucionais, transparência pública nos limites autorizados, login e cadastro.

**Área pessoal**: `Meu TJ` com dados de conta, assinatura/benefícios, carteirinha, jogos, confirmação opcional, frequência, camiseta, histórico de pedidos/pagamentos, solicitações de cancelamento e notificações pertinentes.

**Operação fiscal**: interface responsiva de leitura por câmera e consulta online, seleção de jogo, status de conferência, foto quando permitida e log de operações.

**Administração TJ**: temporadas, planos/preços, quotas e autorizações, pedidos/reservas/pagamentos, associados, jogos, público estimado, frequência/procedimentos, lotes/encomendas/entregas, integração financeiro/estoque e auditoria.

**PROPOSTA de nomes de rota**, sujeita à convenção do repositório: `/`, `/entrar`, `/cadastro`, `/planos`, `/meu-tj`, `/meu-tj/carteirinha`, `/meu-tj/jogos`, `/fiscal`; preservação funcional de `/inicio`, `/financeiro`, `/estoque` e administração atual.

**DIRETRIZ UX:** preservar o Design System existente, hierarquia e linguagem visual da TJ, layout responsivo para público em celular e fiscal em ginásio. Evitar sobrecarga de fluxo de cadastro e garantir mensagens explicativas para pagamentos pendentes, foto ausente, suspensão e indisponibilidade de quota.

---

## 14. Modelo informacional conceitual

> **Não é DDL, migration nem desenho técnico aprovado**. Usar somente como mapa de conceitos que o Specify/Design deve explorar.

- **Usuário/identidade** — referência ao cadastro e às permissões atuais; não duplicar usuários.
- **Temporada** — vigência, estado de publicação, catálogo e quota garantida.
- **Plano ofertado por temporada** — nome, preço, benefícios e regras publicadas versionadas.
- **Adesão** — usuário, temporada, plano contratado, status, pagamento de origem, data de início/fim e histórico.
- **Reserva de quota** — adesão/pedido pretendido, validade, estado, titular e controle de idempotência.
- **Pedido e cobrança** — valores, meio, mercado pago, status, eventos e ligações com Financeiro.
- **Perfil associativo** — número permanente, CPF e dados complementares com histórico/controle de correções.
- **Identificação/foto** — arquivo protegido, metadados e situação de habilitação de ingresso.
- **Jogo** — temporada, local, data, elegibilidade, status e configurações próprias.
- **Intenção de comparecimento** — uma resposta atual por pessoa/jogo e histórico de alterações conforme Design.
- **Autorização de jogo** — avaliação contextual/temporária do direito, sem confundir com check-in efetivado.
- **Tentativa/registro de check-in** — fiscal, jogo, pessoa, resultado, timestamp, motivo, token e auditoria.
- **Apuração de frequência** — jogos elegíveis, presenças, excluídos e justificativas de exclusão.
- **Procedimento de suspensão** — notificação, defesa, decisão, recursos, vigência e revisões.
- **Encomenda individual de camiseta** — associado/adesão, tamanho, nome opcional, lote, andamento, entrega.
- **Lote de produção** — fornecedor, fechamento, itens, quantidades, previsão, recebimento e conferências.
- **Vínculos Financeiro/Estoque** — identificadores das entidades já existentes para evitar duplicidade.
- **Eventos de auditoria/outbox** — operações sensíveis e integrações assíncronas consistentes.

**DIRETRIZ:** o Design deve estudar transações/isolamento do PostgreSQL para invariantes de quota e check-in, índices de unicidade e constraints adequadas, idempotência em webhooks e efeitos financeiros.

---

## 15. Estados e transições que precisam ser projetados

Estados abaixo são **modelos de discussão**, não enums finais:

| Processo | Estados/condições a distinguir |
|---|---|
| Conta | Não verificada; ativa; restringida conforme mecanismo atual. |
| Pedido/checkout | Criado; reserva válida; pagamento pendente; aprovado; recusado; expirado; exceção de conciliação; estornado. |
| Associação | Aguardando pagamento; ativa; cancelamento solicitado; encerrada; expirada; disputa/regularização. |
| Benefício de ingresso | Não aplicável; ativo porém sem foto; habilitado; bloqueado por suspensão válida; encerrado. |
| Jogo | Rascunho; publicado; adiado; cancelado; concluído, com elegibilidade documentada. |
| Frequência | Abaixo de seis jogos; monitorada; alerta; sob procedimento; suspensão vigente; restabelecida. |
| Encomenda/lote | Aguardando lote; incluída; lote fechado; encomendada; em produção; recebida; entregue; divergência/cancelamento. |

**Proibição de acoplamento indevido:** suspender ingresso não transforma associação paga em cancelada; pagar não significa camiseta entregue; intenção de comparecer não é check-in; usuário com role `ASSOCIADO` não necessariamente possui adesão paga.

---

## 16. Requisitos de qualidade e operação

### Segurança

- Sessão centralizada; authorization checks no backend em todas as rotas privadas; princípio de menor privilégio.
- Verificação de autenticidade e reconciliação de eventos de pagamento; armazenamento seguro de segredos.
- Tokens QR com assinatura/expiração/nonce ou mecanismo equivalente discutido no Design; prevenção de replay e uso cruzado entre jogos.
- Trilha de auditoria contra exclusão/alteração indevida para decisões disciplinares, alterações de quota, estornos e validação de entrada.
- Limites de requisições/antiabuso em cadastro, login, pagamento e scanner; proteção contra exposição de dados e scraping.

### Integridade

- Quota: ocupações comprometidas + reservas efetivamente válidas **≤ quota garantida** em todas as transições relevantes.
- Check-in: um resultado aceito por associado e jogo, mesmo sob requisições simultâneas.
- Idempotência de recebimento de webhooks, lançamentos Financeiro, movimentos Estoque e comandos administrativos repetidos.
- Consistência em falhas parciais entre API, provedor de pagamento, banco e módulos existentes; estratégias de recuperação e reconciliação.

### Usabilidade e acessibilidade

- Mobile-first para associados e fiscais, com navegação por teclado quando aplicável, legibilidade e contrastes adequados.
- Respostas claras de acesso autorizado, negado, já utilizado, rede indisponível, foto pendente e suspensão; sem publicar motivos privados no ambiente do ginásio.
- Gestão de erros e estados de carregamento; impedir que cliques repetidos gerem ações duplicadas.

### Privacidade e conformidade

- Princípios da LGPD: finalidade, necessidade, segurança, retenção, compartilhamento e exercício de direitos.
- Explicar tratamento de CPF, imagem e dados de presença e divulgação de relatórios públicos; evitar tornar listas nominais publicamente acessíveis.
- Estatuto, regulamento de benefícios, parceria com clube e condições comerciais devem estar formalizados e refletidos no produto.
- Não publicar horários, percentuais, nomes, CPF, foto ou andamento de procedimento disciplinar sem base legítima e autorização apropriada.

### Operação

- Observabilidade de webhooks, falhas de conciliação, reservas expiradas, cobranças recebidas sem associação ativa, tentativas duplicadas de QR e divergências de estoque.
- Backups/recuperação e logs de auditoria conforme práticas já implementadas.
- Indicadores administrativos suficientes para acompanhamento da temporada, sem exposição indevida.

---

## 17. Critérios mínimos de aceite por risco

| Código | Cenário de teste | Resultado obrigatório |
|---|---|---|
| AC-ID-01 | Novo visitante se cadastra | Conta criada com nome/e-mail/senha; sem CPF/foto/pagamento exigidos. |
| AC-ID-02 | Usuário gratuito tenta acessar administração | Acesso negado sem permissão administrativa. |
| AC-ID-03 | Compra Arquibancada sem foto | Adesão pode existir após pagamento, mas credencial de ingresso fica impedida até cadastro de foto. |
| AC-PL-01 | Compra em mês avançado da temporada | Cobra valor anual integral; vigência encerra em 31/12 daquela temporada. |
| AC-PL-02 | Início de nova temporada | Sem renovação ou débito automáticos; número/histórico preservados. |
| AC-CAP-01 | Dois compradores disputam última vaga | No máximo uma vaga comprometida/reservada por unidade disponível. |
| AC-CAP-02 | Plano básico é comprado | Não reduz quota de Arquibancada. |
| AC-CAP-03 | Pagamento tardio de reserva liberada sem vaga | Dinheiro é conciliado; ingresso não é ativado; ocorrência e reembolso/regularização rastreados. |
| AC-CAP-04 | Associado não confirma presença ou informa ausência | Vaga permanece comprometida e ingresso continua permitido se elegível. |
| AC-CAP-05 | Benefício temporariamente suspenso | Vaga não fica disponível para venda. |
| AC-CAP-06 | Cancelamento definitivo regularmente efetivado | Autorização do antigo titular revogada; quota só liberada se condições comerciais/contratuais cumpridas. |
| AC-PAG-01 | Retorno visual de sucesso sem confirmação do servidor | Associação não é considerada quitada automaticamente. |
| AC-PAG-02 | Webhook aprovado é reenviado várias vezes | Apenas uma ativação e um lançamento financeiro lógico. |
| AC-PAG-03 | Cartão parcelado | Até três parcelas conforme capacidade confirmada do provedor; preço total e eventuais encargos expostos antes de pagar. |
| AC-JOG-01 | Associado informou “não vou”, mas comparece | Fiscal valida normalmente, se demais condições de entrada estiverem válidas. |
| AC-JOG-02 | Dois fiscais escaneiam simultaneamente a mesma pessoa | Um único check-in confirmado; outro retorna situação não duplicável. |
| AC-JOG-03 | Fiscal está offline | Não realiza nem armazena check-in local como válido no MVP. |
| AC-JOG-04 | QR associado é apresentado para outro jogo | Não concede automaticamente entrada no jogo incorreto. |
| AC-FRQ-01 | Associado tem cinco jogos elegíveis e 0 presença | Não pode iniciar procedimento de suspensão por critério de frequência ainda. |
| AC-FRQ-02 | Associado tem seis jogos elegíveis e menos de 50% | Gera elegibilidade para análise/aviso, **não** suspensão automática. |
| AC-FRQ-03 | Jogo ocorre durante impedimento vigente | Não aumenta faltas na apuração. |
| AC-FRQ-04 | Suspensão é decidida sem notificação/defesa prevista | Fluxo deve impedir registro como decisão regularmente efetivada. |
| AC-CAM-01 | Compra + Camisa sem nome | Compra permitida, com tamanho obrigatório. |
| AC-CAM-02 | Nome/tamanho muda antes do fechamento | Pedido do lote reflete alteração autorizada. |
| AC-CAM-03 | Associação paga, lote não recebido | Não há entrada fictícia de estoque. |
| AC-CAM-04 | Mesma camiseta tentada em dois lotes | Segunda alocação incompatível é rejeitada. |
| AC-CAM-05 | Camiseta entregue ao associado | Registro individual e baixa de estoque coerentes e idempotentes. |
| AC-CAN-01 | Cancelamento solicitado mas pendente | Quota não liberada e histórico preservado. |
| AC-CAN-02 | Pedido enquadra-se em direito obrigatório de reembolso | Fluxo respeita obrigação legal, independentemente de preferência por análise manual. |
| AC-SEC-01 | Club board acessa relatório autorizado | Visualiza somente escopo autorizado, sem dados privados administrativos da TJ. |

A suíte final deve incluir testes de domínio, integração, contrato de API, segurança/RBAC, concorrência, UI e fluxos end-to-end conforme arquitetura existente.

---

## 18. Inventário das decisões aprovadas

| Assunto | Decisão consolidada |
|---|---|
| Produto | Uma aplicação/site; não projeto nativo separado. |
| Cadastro | Público, gratuito; nome, e-mail, senha. |
| E-mail | Verificação antes da contratação. |
| CPF | Somente na associação; correção com proteção/auditoria. |
| Foto | Obrigatória apenas para Arquibancada, antes do primeiro uso. |
| Planos | R$ 60 / R$ 160 / R$ 240 em 2027. |
| Vigência | Até 31/12 da temporada; cobrança integral sem proporcionalidade. |
| Renovação | Manual/nova contratação; sem renovação automática ou prioridade do ano anterior. |
| Quota | Garantia configurável por temporada; dois planos Arquibancada somados. |
| Checkout | Reserva temporária atômica, alinhada ao pagamento. |
| Pagamento tardio | Revalidar quota; sem overbooking; regularizar/reembolsar se sem vaga. |
| Pagamento | Mercado Pago; Pix e cartão até 3x; transparência de acréscimos aplicáveis. |
| Cancelamento | Solicitação no Meu TJ, análise administrativa com salvaguardas legais. |
| Liberação de vaga | Somente cancelamento definitivo efetivado e condições contratuais cumpridas. |
| Jogos | Presidente TJ cria/administra; delegações por permissão. |
| Confirmação | Opcional e informativa; não limita direito de comparecer. |
| Relatório clube | Projeção agregada enviada à direção do clube, distinta da TJ. |
| Scanner | Fiscal TJ, online, QR + identificação + check-in único auditável. |
| Ingresso | Credencial TJ não substitui ingresso oficial do clube. |
| Frequência | 50% de check-ins efetivos, ao menos seis jogos elegíveis antes de análise. |
| Suspensão | Possível na temporada, nunca automática; defesa, decisão motivada, prazo certo, revisão; vaga permanece ocupada. |
| Camiseta | Sob encomenda; tamanho obrigatório; nome opcional, sem número. |
| Lotes | Agrupamento e fechamento manuais pela diretoria. |
| Estoque | Recebimento físico e entrega vinculados à encomenda; sem estoque fictício. |
| Offline/Wallet | Fora do MVP. |

---

## 19. Pendências de produto, contrato e operação

> **Esta seção não reabre decisões aprovadas.** Consolida apenas parâmetros, documentos externos ou exceções que ainda precisam ser definidos. A IA não deve decidir silenciosamente nenhum destes pontos.

### P0 — Impedimentos para comercialização/operação de benefícios

1. **Parceria com clube:** quota efetivamente garantida para 2027, cobertura exata dos jogos, forma de ingresso oficial, autoridade para alterar quota, substituição de titular após cancelamento e formalização da relação.
2. **Base associativa/legal:** regularidade institucional da associação, estatuto e regulamento aceitos, legitimidade dos órgãos decisores, direitos do associado, suspensão de benefícios pagos e consequências patrimoniais.
3. **Disciplina de suspensão:** notificação válida, prazo de defesa, instância decisória, duração máxima/típica, recurso/revisão, exceções de jogos elegíveis, restituição ou ajustes financeiros quando aplicáveis.
4. **Mercado Pago:** ambiente e titularidade real do recebedor, produto de checkout, Pix/cartão/parcelamento até 3x, encargos repassáveis, validade da cobrança, webhook/consulta/estorno e obrigações fiscais.
5. **Reserva de quota:** tempo de retenção, expiração segura por modalidade, conciliação tardia, limite de pedidos simultâneos por pessoa e política de ressarcimento/regularização sem overbooking.
6. **Termos comerciais:** direitos obrigatórios de arrependimento/cancelamento, comunicação de prazo de camiseta, regras da oferta, eventuais restrições e clareza entre carteirinha TJ e ingresso do clube.

### P1 — Definições de negócio necessárias para especificação completa

7. Dados complementares da associação e procedimento de correção de CPF/identidade; menores/responsáveis.
8. Política exata de foto: dimensões/revisão/atualização e prevenção de abuso, sem coleta em cadastro gratuito.
9. Regras detalhadas do QR/validade, acesso fiscal, eventual correção administrativa de check-in e contingência durante falta de rede (sem registrar check-in offline como válido).
10. Definição de jogo elegível: partidas em casa cobertas, canceladas/adiadas, jogos após adesão, situações excepcionais de acesso impedido por terceiro.
11. Publicação e compartilhamento de estimativas com o clube; forma de acesso e privacidade.
12. Benefícios não financeiros do plano básico, condições de loja/antecipação e conteúdo exato que aparecerá na oferta pública.
13. Processos de camiseta: grade de tamanhos, modelo 2027, caracteres/nome permitido, fornecedor, prazo estimado, custo e logística de retirada; exceções de defeitos e divergências.
14. Quem pode autorizar reembolso, corrigir presença, fechar lote, mudar quota ou decidir suspensões em termos de permissões individuais.

### P2 — Planejamento técnico/lançamento

15. Auditoria da implementação atual de identidade, sessão e RBAC; interfaces e integrações já existentes.
16. Mapeamento Financeiro (bruto, líquido, taxas, estornos) e Estoque (tamanhos/SKU/baixa física).
17. Estratégia de implantação por fases sem publicar benefícios antes da prontidão operacional.
18. Transparência pública: escopo de relatórios, governança de publicação, periodicidade e anonimização.
19. Métricas não funcionais e operação real: volume esperado, teste de pico de entrada, conectividade no ginásio e recuperação de incidentes.

---

## 20. Pontos de divergência entre PDFs antigos e decisões consolidadas

| Tema | Risco de interpretação ao ler a referência antiga | Regra que prevalece |
|---|---|---|
| Quota | Valores de 50 ou 60 vagas ilustrativas | Configurar `N` garantido por temporada, confirmado com o clube. |
| Intenção de comparecer | Tratar resposta ou confirmação como reserva e liberar vagas | Resposta é só estimativa; nunca redistribuir vaga por falta ou resposta. |
| Frequência | Percentual usado apenas como alerta ou bloqueio automático | Após ao menos seis jogos elegíveis, pode haver procedimento individual, sem punição automática; suspensão temporária é possível conforme regulamento válido. |
| Offline | Check-in com listas sincronizadas offline | MVP **somente online**. |
| Foto | Exigir em todo cadastro | Não no cadastro gratuito; para Arquibancada, antes do primeiro uso do ingresso. |
| Carteirinha QR | Confundir QR de associado com ingresso do clube | Verificação TJ por jogo é adicional ao ingresso oficial exigido pelo clube. |
| Camiseta | Simplificar como item sempre disponível em estoque | Sob encomenda, lotes manuais, nome opcional, tamanho obrigatório. |
| Renovação | Prioridade automática ou reserva para associados anteriores | **Não** oferecer prioridade do ano anterior nesta versão. |

---

## 21. Diretrizes para a fase Specify/Discuss/Design/Tasks

### Primeiro passo: descoberta, sem alterar sistema

1. Ler este documento e os dois PDFs separadamente; registrar divergências e decisões substituídas.
2. Investigar branch e arquitetura reais do repositório (Go, PostgreSQL, Next.js, mecanismos de autenticação, permissões, Financeiro e Estoque).
3. Construir mapa de reutilização de módulos, dependências, riscos e contratos existentes, **sem pressupor** que o código tem exatamente as características anteriormente descritas.
4. Produzir requisitos rastreáveis por ID e testes de aceite, classificação P0/P1/P2 e proposta de divisão incremental.
5. Identificar decisões abertas indispensáveis e apresentar opções com trade-offs para discussão com responsável do produto.
6. Não criar código/migrations nem efetuar integração ao Mercado Pago neste estágio; aguardar encerramento dos gates Specify → Discuss → Design → Tasks.

### Design posterior — entregáveis esperados

- Contextos e limites dos módulos (identidade, associação, temporada, capacidade, pagamento, jogos, check-in, frequência, camisetas, Financeiro, Estoque).
- Entidades/relações e migrations **propostas**, com migração segura do banco existente.
- Fluxos de estado, conciliação, expiração e compensação em falhas.
- Matriz detalhada de RBAC/ownership e proteção de dados.
- Contratos HTTP/eventos, idempotência, locking transacional e integração com provedor.
- Wireflows/UI aderentes ao Design System existente e acessibilidade.
- Plano de testes que comprova invariantes e casos de concorrência.
- Plano de rollout, observabilidade e contingência operacional.

### Critério para seguir da especificação ao desenvolvimento

O responsável do produto valida as pendências P0, aprova regras comerciais e procedimento disciplinar dentro das exigências legais, confirma quota/ingresso com o clube, e o time técnico aprova arquitetura e estratégia de testes. Apenas então avançar para Tasks/implementação.

---

## 22. Resumo executivo de fechamento

O TJ Connect 2027 será uma expansão integrada da plataforma existente. O torcedor pode se cadastrar gratuitamente e adquirir um dos três planos anuais (R$ 60, R$ 160 ou R$ 240). Os dois planos Arquibancada compartilham uma quota sazonal realmente garantida pelo clube; checkout e pagamentos Mercado Pago não podem sobrecontratar essa quota. A confirmação de comparecimento é opcional e jamais substitui o direito adquirido. O fiscal TJ registra entradas pelo QR online, sem substituir o ingresso oficial do clube. A frequência de 50% pode levar a uma suspensão temporária **somente** após seis jogos elegíveis e processo individual regular; sem liberação da vaga. Camisetas são personalizáveis apenas com nome, opcionalmente, produzidas sob encomenda por lotes fechados manualmente, com integração física ao Estoque. O Financeiro existente registra os efeitos efetivos dos pagamentos, sem duplicidade.

**Próximo marco:** usar este documento como baseline de produto, confrontá-lo com os dois PDFs e a base de código no **Specify**, e resolver os pontos P0 antes de comercialização/ativação real de benefícios.
