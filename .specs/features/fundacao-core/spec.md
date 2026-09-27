# Fundação Core Specification

Status: Aprovada com ajustes em 2026-09-26. Feature 1 de 2 da fundação (a outra é `fundacao-documentos`). A modelagem de Identity/RBAC foi revisada em 2026-09-26 com os ajustes do mantenedor (IDN-01, IDN-04, IDN-05, IDN-06, RBAC-01 e RBAC-03) e foi aprovada em 2026-09-26. Narrativa em português; os critérios de aceite (ACs) seguem o padrão EARS em inglês, exigido pelo validador `validate_spec.py`.

## Problem Statement

A plataforma só tem um esqueleto (`/healthz` e uma página de status). Todo módulo de negócio (financeiro, estoque, loja, associados) depende das mesmas capacidades: identificar quem age, verificar permissões, registrar auditoria imutável, tratar dinheiro sem erro de arredondamento, guardar documentos, testar contra um banco real e expor um contrato de API estável. Se cada módulo construir isso por conta própria, as regras dos ADR-003, 004, 005 e 006 divergem e ficam caras de corrigir.

A fundação foi dividida em duas features. Esta, `fundacao-core`, entrega identidade, autorização, auditoria, dinheiro, o contrato OpenAPI e a infraestrutura de testes, em `platform` e `identity`. Documentos e object storage ficam em `fundacao-documentos`, executada depois desta.

## Goals

- [ ] Um usuário autenticado por sessão com cookie seguro, com autorização por permissão (nunca só por nome de papel) aplicada em todo caso de uso.
- [ ] Toda alteração de dados auditados grava, na mesma transação, um registro que nem a aplicação consegue alterar ou apagar.
- [ ] Valores monetários em centavos inteiros, com formatação, serialização, rateio e percentuais exatos no Go e no TypeScript.
- [ ] Testes de integração contra PostgreSQL real no CI, verificação automática de fronteiras entre módulos e testes de front.
- [ ] Contrato OpenAPI como fonte de verdade, um por módulo, com código do servidor e tipos TypeScript gerados e conferidos no CI.

## Out of Scope

| Feature | Reason |
|---|---|
| Telas (login, painel, gestão de usuários) | A fundação entrega API e infraestrutura; a UI de identidade é a feature seguinte |
| Provedor concreto de e-mail (Resend, SES, SendGrid, SMTP) | A capacidade `platform/email` e a interface de envio são entregues; o adaptador de um provedor real fica para quando ele for escolhido |
| MFA, OAuth e provedor externo de identidade | ADR-005 adia; pede novo ADR |
| Entidades e regras do financeiro: lançamentos, receitas, despesas, plano de contas, produtos e eventos | Módulos de domínio posteriores; esta feature entrega só o núcleo monetário compartilhado (`platform/money`), que eles importam |
| Documentos e object storage | Feature `fundacao-documentos`, executada depois desta |
| Hospedagem e deploy reais | Decisão adiada; ver ADR-002 e o plano de hospedagem |
| Purga e retenção de auditoria e sessões | Política de retenção depende de definição jurídica e de LGPD |
| Fechamento de ano e ajuste extraordinário | Regras do financeiro (FIN-001, seção 16) |
| Dupla aprovação para promoção administrativa | Evolução futura, fora da V1: hoje uma pessoa com `identity:admin:grant` promove, sempre com motivo e auditoria |
| Vínculo entre usuário e associado | Pertence ao módulo dono do domínio: `associados.associados.user_id`, criado na spec de `associados`; `identity` não conhece associados |
| Convite por e-mail e primeiro acesso por link temporário | Evolução futura da criação de usuários |
| Troca de e-mail (`ChangeEmail`) | Evolução futura: o e-mail é identificador de login e canal de recuperação, então a troca exigirá senha atual, auditoria e, possivelmente, confirmação por e-mail |
| Rate limit unificado de credenciais, por IP e dispositivo, e MFA na recuperação | Evolução futura |
| Rate limiting global da API | Só o bloqueio de tentativas de login está no escopo |

---

## Assumptions & Open Questions

Todas as ambiguidades estão resolvidas ou registradas aqui. "Confirmed? n" significa que o padrão escolhido aguarda a sua confirmação na aprovação da spec.

| Assumption / decision | Chosen default | Rationale | Confirmed? |
|---|---|---|---|
| Como o navegador envia o cookie à API | Cookie `SameSite=Lax` com `Domain` configurável; `Secure` obrigatório fora de `development`. Web e API no mesmo domínio registrável (ex.: `app.` e `api.`), ou o Next faz proxy | Aprovado pelo mantenedor. A configuração final (domínio, proxy) **depende da estratégia de hospedagem**, ainda não definida | y |
| Formato do token de sessão | 256 bits aleatórios em base64url; no banco só o hash SHA-256 | Vazamento do banco não permite sequestrar sessões | n |
| Tempo de sessão | Sessão administrativa expira em 8 horas (teto absoluto) ou após 60 minutos sem uso; a regra para associados será fixada na spec de associados e, até lá, vale a mesma | 8 horas definido pelo mantenedor; 60 minutos de ociosidade é o padrão mantido, ajustável por configuração | n |
| Política de senha | Quem tem qualquer papel diferente de ASSOCIADO: mínimo 10 caracteres; só ASSOCIADO: mínimo 8; máximo 128 em pontos de código Unicode; sem regras de composição. Hash argon2id com memória 19 MiB, 2 iterações, paralelismo 1, configuráveis | Mínimos definidos pelo mantenedor; parâmetros alinhados às recomendações públicas do OWASP (validar na tarefa contra a versão vigente) | y |
| Bloqueio de login | 5 falhas seguidas para o mesmo e-mail em 15 minutos bloqueiam novas tentativas por 15 minutos (HTTP 429); nunca há bloqueio permanente; contadores no PostgreSQL | Aprovado pelo mantenedor; barra força bruta sem trancar a conta de forma definitiva | y |
| Proteção contra senhas comprometidas | Lista local embutida e versionada de senhas comuns e vazadas, comparada sem diferenciar maiúsculas; fonte e licença verificadas na tarefa. Solução inicial: integração externa (ex.: consulta k-anonymity) permanece como evolução em aberto | Aprovado pelo mantenedor; não cria dependência externa nem envia dados para fora | y |
| Troca obrigatória de senha | Ao conceder o primeiro vínculo administrativo a um usuário que não tinha vínculo ativo (só ASSOCIADO), o sistema marca `must_change_password=true` e exige a troca no próximo login; até lá o usuário só pode trocar a senha, ver `me` ou sair | Aprovado pelo mantenedor; evita que uma senha de 8 caracteres passe a proteger conta administrativa | y |
| Capacidade de e-mail | `platform/email`, transversal: interface `Sender` desacoplada do provedor, mensagem validada (destinatário e ausência de quebras de linha nos cabeçalhos), remetente `log` para desenvolvimento e `disabled` para quando não há provedor; nenhum módulo depende de um provedor. O provedor real é definido depois e entra como novo adaptador. O corpo da mensagem nunca é registrado em log | Aprovado pelo mantenedor: e-mail é capacidade da plataforma, não só da autenticação | y |
| Recuperação de acesso | Por e-mail, com token de recuperação (256 bits, base64url), guardado só como hash SHA-256 em `password_reset_tokens`, válido por 30 minutos (configurável) e de uso único. Não existe senha temporária. A resposta da solicitação é sempre a mesma; o envio é assíncrono, para o tempo de resposta não revelar se a conta existe; no máximo 3 solicitações por e-mail por hora (as demais recebem a mesma resposta e não enviam) | Aprovado pelo mantenedor; o limite de 3 por hora contra spam de e-mail é proposta a confirmar | n |
| Redefinição administrativa | Alternativa à recuperação por e-mail: quem tem `identity:user:reset_password` informa um motivo (mínimo 10 caracteres), e o sistema gera uma senha temporária aleatória, mostrada uma única vez ao administrador na resposta (nunca em log, auditoria ou cache), marca `must_change_password`, revoga as sessões e os tokens de recuperação pendentes do alvo e grava `user.password_reset`. Segue a regra sem escalada e não vale para a própria conta. Nunca são enviados senha antiga, token ou hash | Aprovado pelo mantenedor; a recuperação por e-mail continua sem senha temporária, só com token | y |
| Criação de usuário | O criador informa a senha inicial e o sistema marca `must_change_password=true` automaticamente, para que ela não fique como credencial definitiva conhecida por terceiro. Criação por convite e primeiro acesso por link temporário ficam como evolução futura | Aprovado pelo mantenedor | y |
| Regra sem escalada nas ações sobre outro usuário | Vale para promoção, retirada de acesso, alteração de papéis, desativação, reativação e redefinição administrativa: o ator precisa cobrir (`authz.Covers`) as permissões efetivas que a operação concede, retira ou altera no alvo. Ninguém desativa a própria conta, remove o próprio acesso administrativo, altera os próprios papéis ou redefine a própria senha por esse caminho (403 `self_change_forbidden`, com evento de segurança) | Aprovado pelo mantenedor | y |
| Troca de senha | Exige a senha atual, aplica a política, recusa a senha igual à atual (422 `password_unchanged`) e revoga as outras sessões. Erros de senha atual têm contador próprio, `password_change_attempts`, separado do login: 5 erros em 15 minutos bloqueiam a troca por 15 minutos (429 `password_change_blocked`). Um rate limit unificado de credenciais fica como evolução | Aprovado pelo mantenedor | y |
| Papel-base ASSOCIADO | Toda conta tem ASSOCIADO, que nunca é removido; papéis administrativos são adicionais. A lista de papéis recebida pela API é a dos administrativos (pode ser vazia, deixando o vínculo dormente); `PromoteToAdmin` exige ao menos um papel administrativo. Reativar um usuário não restaura papéis administrativos nem o vínculo encerrado | Aprovado pelo mantenedor | y |
| Listagem de usuários | Cursor por (`created_at`, `id`), 50 por página e no máximo 100; filtros `active` e `role`; nunca devolve hash, token ou dado sensível | Aprovado pelo mantenedor | y |
| Hash de e-mail | HMAC-SHA256 com a chave `AUTH_HASH_KEY` (mínimo 32 bytes, obrigatória fora de `development`) nas tentativas de login e nos eventos de auditoria; o e-mail nunca é gravado em claro nesses lugares | Aprovado pelo mantenedor; um vazamento do banco ou da auditoria não permite descobrir e-mails por dicionário | y |
| E-mail: normalização | Espaços nas pontas removidos, minúsculas e formato básico (contém `@`, até 254 caracteres); sem normalização de domínios internacionalizados nem de aliases | Aprovado pelo mantenedor | y |
| Sessões no login | Cada login emite token novo e revoga a sessão apresentada no cookie, se válida; várias sessões simultâneas por usuário são permitidas; todas caem ao trocar ou redefinir a senha, desativar, promover ou retirar o acesso | Aprovado pelo mantenedor | y |
| Bloqueio de login por e-mail | Limitação conhecida: quem conhece um e-mail pode bloqueá-lo por 15 minutos, mesmo inexistente; tentativas durante o bloqueio não contam nem o estendem; limite por IP fica fora do escopo | Aprovado pelo mantenedor | y |
| Argon2id e lista de senhas | Parâmetros configuráveis (padrão 19 MiB, 2 iterações, paralelismo 1, conferidos na tarefa com a recomendação vigente do OWASP e o custo medido); lista embutida de senhas comuns com fonte e licença registradas no commit | Aprovado pelo mantenedor | y |
| Valor padrão de `APP_ENV` | Variável ausente equivale a `production` | Aprovado pelo mantenedor; padrão seguro, e `development` passa a ser explícito | y |
| `ALLOWED_ORIGINS` vazio | Mantém o comportamento seguro: nenhuma origem é autorizada e requisições de navegador com `Origin` são bloqueadas; em `development` o padrão é `http://localhost:3000` | Aprovado pelo mantenedor; falha fechada em vez de aberta | y |
| Cadastro de usuários | Sem autocadastro. O primeiro administrador nasce por CLI; os demais usuários são criados por quem tem `identity:user:create`, sempre só com o papel ASSOCIADO, e o acesso administrativo só existe por promoção (IDN-06), um único caminho para conceder poder | Sistema institucional fechado; reduz superfície de ataque e concentra a regra de promoção em um caso de uso | n |
| Origem das permissões | Modular: cada módulo declara as suas permissões e concessões por papel em uma contribuição própria; `identity` agrega e valida a matriz final. Não existe matriz central única; até a spec do financeiro existir, a fundação declara em nome dele as três permissões institucionais do Conselho Fiscal | Aprovado pelo mantenedor; evita acoplar todos os módulos a uma matriz gigante | y |
| Auditoria de negações | `authz.denied` é gravado (como evento de segurança) para toda negação, exceto de permissões declaradas como leitura comum (`CommonRead`, permitido só em ação `read`); administração, segurança, RBAC, financeiro e dados institucionais sempre são auditados. Leituras comuns ficam só no log de acesso | Aprovado pelo mantenedor; evita volume excessivo de auditoria sem perder as negações relevantes | y |
| Papéis iniciais | ASSOCIADO, PRESIDENTE, DIRETORIA, TESOURARIA, ESTOQUE_LOJA, EVENTOS, CONSELHO_FISCAL e ADMIN_SISTEMA | Lista proposta pelo mantenedor; substitui ADMIN, DIRETOR e FINANCEIRO | y |
| Matriz inicial de permissões | Provisória e mínima. PRESIDENTE: todas as permissões do catálogo. ADMIN_SISTEMA: `identity:user:read`, `identity:user:create`, `identity:user:update`, `identity:role:assign`, `identity:admin:grant`, `identity:admin:revoke` e `audit:log:read`, sem permissões institucionais (nesta matriz provisória). DIRETORIA: só `identity:user:read`. CONSELHO_FISCAL: `audit:log:read` e as institucionais. TESOURARIA, ESTOQUE_LOJA, EVENTOS e ASSOCIADO: nenhuma da fundação (as de documentos entram em `fundacao-documentos`). Cada spec de módulo amplia a matriz do seu módulo. A matriz é dado: a autorização decide sempre pelas permissões efetivas, nunca pelo nome do papel | Provisória. `audit:log:read` do Conselho Fiscal aprovado pelo mantenedor, mantendo a possibilidade futura de separar auditoria institucional e técnica | y |
| Conselho Fiscal | Nunca recebe `create`, `update`, `delete` nem `cancel`. Não é só leitura: recebe as permissões institucionais `financeiro:prestacao_contas:read` (prestação de contas), `financeiro:prestacao_contas:approve` (aprovação) e `financeiro:parecer:opine` (parecer). O catálogo da fundação as declara; os endpoints nascem na spec do financeiro | Decisão do mantenedor (ADR-005, FIN-001 seção 24), ampliada em 2026-09-26 | y |
| Usuário × Associado | Entidades separadas. O vínculo é `associados.associados.user_id` (o módulo dono do domínio guarda a referência); `identity.users` não tem coluna de associado. O associado pode existir sem usuário, e o usuário administrativo pode não ser associado | Aprovado pelo mantenedor: mantém o vínculo no módulo dono e evita `identity` depender de `associados` | y |
| Vínculo administrativo × papel | `AdminMembership` responde "por que a pessoa tem acesso administrativo" (motivo, quem concedeu, quando, encerramento); papel responde "o que ela pode fazer". São separados: encerrar um não some com o histórico do outro. Regra: papel diferente de ASSOCIADO exige vínculo ativo; um vínculo ativo por usuário; o histórico nunca é apagado | Aprovado pelo mantenedor | y |
| Múltiplos papéis | Um usuário pode ter vários papéis ativos; a permissão efetiva é a união; nega por padrão; não há permissão negativa | Aprovado pelo mantenedor | y |
| Concessão sem escalada | Ninguém concede uma permissão superior à própria: o ator só concede, retira ou altera papéis cujas permissões estejam todas entre as suas, e nunca altera os próprios papéis nem o próprio vínculo | Aprovado pelo mantenedor; a regra vale para qualquer ator e decide só pelas permissões efetivas, sem exceção por papel; o que cada papel consegue conceder decorre da matriz | y |
| Último administrador | O sistema impede desativar, retirar o vínculo ou remover as permissões do último usuário ativo que possui `identity:admin:grant`; a CLI `bootstrap-admin` cria o primeiro (ADMIN_SISTEMA por padrão, ou PRESIDENTE) e só roda sem nenhum vínculo administrativo ativo | Evita ficar sem quem administre acesso; o PRESIDENTE, com todas as permissões, pode conceder qualquer papel | n |
| Conteúdo do registro de auditoria | Campos: `id`, `occurred_at` (UTC), `actor_type` (`user`, `system` ou `anonymous`), `actor_user_id`, `action`, `entity_type` e `entity_id` (o alvo), `before`, `after`, `outcome` (`success`, `denied` ou `failure`), `reason`, `context` (JSON sem segredos) e `request_id`. Sem endereço IP; retenção indefinida nesta fase | IP é dado pessoal (LGPD); o `request_id` correlaciona o evento com os logs de acesso, mantidos separados. Aprovado pelo mantenedor | y |
| Falha ao gravar auditoria | Dois graus. **Alteração crítica de negócio** (promoção, retirada de acesso, mudança de papéis, sincronização da matriz, criação e desativação de usuários e, no futuro, dados financeiros): a auditoria é gravada na mesma transação e, se falhar, a operação é revertida e responde 500 `audit_failed`; só há operação concluída se houver registro. **Evento de segurança ou telemetria** (login, falha de login, bloqueio, logout, negação): não bloqueia a operação principal; se a gravação falhar, o sistema registra um incidente operacional (log de erro com a ação e o `request_id`) e segue, pois nenhum estado crítico foi alterado | Aprovado pelo mantenedor: falha de auditoria de telemetria não pode derrubar login nem ser usada para negar serviço; alteração crítica não existe sem rastro | y |
| Evoluções futuras da auditoria | Avaliar armazenamento controlado de IP em eventos de segurança (requisitos operacionais e LGPD); separação entre auditoria institucional e técnica; versionamento do formato dos eventos; encadeamento criptográfico dos registros. Nenhum entra nesta fase: a proteção atual é append-only, triggers contra alteração e permissões restritas no banco | Aprovado pelo mantenedor | y |
| Catálogo de ações de auditoria | As ações são constantes em código (`domínio.verbo`); o registrador rejeita ação fora do catálogo | Evita nomes divergentes e permite testar que toda mudança de permissão tem evento | n |
| Fuso horário | Tudo em UTC (`timestamptz`); conversão para America/Sao_Paulo na borda | Evita ambiguidade de horário de verão e de servidor | n |
| Arredondamento de percentuais | `ROUND_HALF_UP` (empate arredonda para longe de zero, também nos negativos); taxas em pontos-base (1 bp = 0,01%) | Aprovado pelo mantenedor; segue a semântica de `BigDecimal.ROUND_HALF_UP` | y |
| Rateio de centavos | O resto é distribuído 1 centavo por parcela, da primeira em diante | Determinístico e reproduzível (ex.: 10000 em 3 partes = 3334, 3333, 3333) | n |
| Mapeamento HTTP de `ErrOutOfRange` | O erro fica no domínio do Money; o mapeamento para 422 `amount_out_of_range` acontece na primeira feature com endpoint monetário real | Aprovado pelo mantenedor; nenhum endpoint da fundação recebe valores e o Money não deve conhecer HTTP | y |
| Parse monetário | Só dígitos, pontos de milhar em grupos de três e uma vírgula; sem sinal e sem prefixo `R$`. Se necessário, separar no futuro o parsing de entrada monetária do parsing técnico de centavos | Aprovado pelo mantenedor | y |
| Rateio com pesos desiguais | A soma preserva o total e o resto vai, um centavo por vez, para as primeiras partes; comportamento determinístico | Aprovado pelo mantenedor | y |
| Basis points negativos | Aceitos e invertem o sinal do resultado | Aprovado pelo mantenedor | y |
| Limite de valores na API | Inteiros assinados até 9007199254740991 em módulo | Faixa segura para números do JavaScript no front | n |
| Testes de integração | Ficam atrás da tag de build `integration`; `go test ./...` não exige Docker | Mantém o ciclo local rápido | n |
| Cobertura | Sem meta global; o pacote de dinheiro exige 95% de cobertura de instruções | Lógica pura e crítica, fácil de cobrir | n |
| Formato de erro | `application/problem+json` (RFC 9457) com `code` e `request_id`; JSON malformado dá 400, validação dá 422 | Padrão aberto e previsível para o cliente gerado | n |
| Versão da API | Prefixo `/api/v1`; `/healthz` permanece na raiz | Permite evoluir sem quebrar clientes | n |
| Contrato OpenAPI | OpenAPI 3.0.3 (ADR-008), contract-first, um contrato por módulo em `api/openapi/` (`platform.yaml`, `identity.yaml`, e os dos módulos futuros) mais `common.yaml` compartilhado; código gerado por módulo e versionado; CI falha com diferença | Aprovado pelo mantenedor; evita arquivo único gigante e respeita as fronteiras de módulo | y |
| Ferramentas de geração e lint | `oapi-codegen` (servidor Gin, modo strict), `kin-openapi` (validação de respostas nos testes), `openapi-typescript` (tipos do front) e Redocly (lint); `openapi-fetch` só na primeira tela real. Versões fixadas na tarefa T24 depois de um teste rápido de `application/problem+json`, cookie de sessão e `X-CSRF-Token`; se houver incompatibilidade, a solução volta ao mantenedor antes de ser fixada | Aprovado; nenhuma versão é fixada sem checar a documentação vigente | y |
| Conteúdo do `common.yaml` | Somente componentes realmente compartilhados: erros, paginação, IDs, Money e estruturas comuns; nenhuma entidade de domínio. O componente `Id` (UUID) entra quando o primeiro contrato precisar | Aprovado pelo mantenedor; evita acoplar módulos por meio de um arquivo comum | y |
| Validação de requisições pelo contrato | O contrato valida estrutura, tipos, formatos, enums e campos obrigatórios (corpo, query e path), de forma automática, por middleware próprio em `platform/httpx` sobre o `kin-openapi`; regras de negócio, autenticação e autorização ficam nos módulos de domínio. Rota registrada sem operação no contrato é recusada (o roteador falha na partida e um teste de paridade a detecta), pois o contrato é a fonte de verdade | O `gin-middleware` v1.1.0 entrega ao tratador só a mensagem em texto, sem os erros por campo que API-02.4 exige; o contrato não muda para se adequar à ferramenta. Aprovado pelo mantenedor (opção A) | y |
| Limite de responsabilidade do contrato | O OpenAPI define contratos de comunicação (caminhos, esquemas, segurança, exemplos) e não contém regras de negócio, que permanecem nas specs e nos módulos de domínio | Aprovado pelo mantenedor; evita duplicar regra em dois lugares | y |
| Limpeza de sessões e tentativas | Registros expirados são ignorados e podem ser removidos de forma oportunista; sem job de purga | Evita infraestrutura de tarefas agendadas nesta fase | n |
| Papéis de banco | Migrações rodam com o papel dono; a API roda com um papel de aplicação sem DDL e sem UPDATE/DELETE em auditoria | Defesa em profundidade para o ADR-004 | n |

**Open questions:** none - all resolved or logged above.

---

## User Stories

### P1: Primeiro administrador ⭐ MVP

**User Story**: Como operador do sistema, quero criar o primeiro administrador por linha de comando, para poder entrar no sistema sem autocadastro.

**Why P1**: Sem um usuário inicial, nada mais pode ser usado.

**Acceptance Criteria**:
1. WHEN the operator runs `bootstrap-admin` with `--email`, `--name`, an optional `--role` (ADMIN_SISTEMA by default, or PRESIDENTE) and the password supplied through the environment variable `BOOTSTRAP_ADMIN_PASSWORD` THEN the system SHALL create an active user with that role and an active administrative membership with the fixed reason `bootstrap-admin` and no grantor, and record an audit entry with action `user.bootstrap`, with no actor, containing the email, the role and the membership reason and never the password.
2. IF an active administrative membership already exists THEN the system SHALL exit with a non-zero status without creating any user.
3. IF the password is absent or violates the administrator password policy (at least 10 characters) THEN the system SHALL exit with a non-zero status without creating any user.
4. The system SHALL NOT accept the password as a command-line argument.

**Independent Test**: Rodar o comando contra um banco vazio e ver o usuário, o vínculo administrativo e o registro de auditoria; rodar de novo e ver a recusa.

---

### P1: Login, sessão e logout

**User Story**: Como usuário, quero entrar com e-mail e senha e manter uma sessão segura, para acessar as áreas conforme minhas permissões.

**Why P1**: Toda autorização parte de uma sessão válida.

**Acceptance Criteria**:
1. WHEN a user submits a valid email and password to `POST /api/v1/auth/login` THEN the system SHALL respond 200 with the same body as `GET /api/v1/auth/me`, create a session and set the cookie `tj_session` with the attributes HttpOnly, Secure, SameSite=Lax and Path=/.
2. WHEN a session is created THEN the system SHALL store only the SHA-256 hash of the session token in the database.
3. IF the email does not exist or the password is wrong THEN the system SHALL respond 401 with code `invalid_credentials` and an identical body for both cases.
4. IF the user is inactive THEN the system SHALL respond 401 with code `invalid_credentials` and the same body as for a wrong password.
5. WHILE a session has been idle for more than 60 minutes or is older than 8 hours the system SHALL treat it as expired and respond 401 with code `session_expired`.
6. WHEN a user calls `POST /api/v1/auth/logout` THEN the system SHALL revoke the session and clear the cookie.
7. WHEN a login succeeds THEN the system SHALL issue a new session token even if a session cookie was sent with the request.
8. IF a fifth consecutive login failure for the same email occurs within 15 minutes THEN the system SHALL respond 429 with a `Retry-After` header to every login attempt for that email during the following 15 minutes.
9. WHEN a session is valid THEN `GET /api/v1/auth/me` SHALL respond 200 with the user id, email, name, role names, effective permissions, the administrative membership summary (`null`, or an object with only `reason` and `granted_at`, never the grantor or the history), the CSRF token and the flag `must_change_password`. The system SHALL NOT include hashes, internal tokens or other sensitive data in these responses.
10. The system SHALL NOT lock an account permanently as a result of failed logins.

**Independent Test**: Logar com o administrador criado pela CLI, chamar `/auth/me`, deslogar e ver que a sessão deixou de valer.

---
### P1: Proteção contra CSRF e origem

**User Story**: Como usuário, quero que outro site não consiga agir em meu nome, para que minha sessão não seja abusada.

**Why P1**: Cookies de sessão são enviados automaticamente pelo navegador (ADR-005).

**Acceptance Criteria**:
1. WHEN an authenticated request uses POST, PUT, PATCH or DELETE THEN the system SHALL require the header `X-CSRF-Token` to equal the CSRF token of the session.
2. IF the CSRF header is missing or different THEN the system SHALL respond 403 with code `csrf_invalid`.
3. IF a state-changing request carries an `Origin` header that is not in `ALLOWED_ORIGINS` THEN the system SHALL respond 403 with code `origin_not_allowed`.
4. IF `POST /api/v1/auth/login` carries an `Origin` header that is not in `ALLOWED_ORIGINS` THEN the system SHALL respond 403 with code `origin_not_allowed`.
5. The system SHALL NOT change state in response to GET, HEAD or OPTIONS requests.
6. IF an authenticated state-changing request carries no `Origin` header and has `Sec-Fetch-Site: cross-site` THEN the system SHALL respond 403 with code `origin_not_allowed`; the absence of `Origin` alone SHALL NOT block a request, since clients other than browsers do not send it.

**Independent Test**: Enviar um POST autenticado sem o cabeçalho e com origem estranha e ver os dois 403.

---

### P1: Gestão de usuários

**User Story**: Como administrador, quero criar, desativar e atribuir papéis a usuários, para dar e retirar acesso com rastreabilidade.

**Why P1**: Sem usuários e papéis não há quem use os módulos.

**Acceptance Criteria**:
1. WHEN an actor with `identity:user:create` calls `POST /api/v1/users` with email, name and initial password THEN the system SHALL create an active user holding only the role ASSOCIADO with `must_change_password` set to true, respond 201 and record an audit entry with action `user.create`.
2. IF the email already exists, ignoring letter case, THEN the system SHALL respond 409 with code `email_taken`.
3. WHEN an actor with `identity:user:update` deactivates a user THEN the system SHALL mark the user inactive, revoke all of that user's sessions and pending recovery tokens in the same transaction and record an audit entry with action `user.deactivate`.
4. IF an actor tries to deactivate the last active user that holds `identity:admin:grant` THEN the system SHALL respond 409 with code `last_admin`.
5. WHEN a user calls `POST /api/v1/auth/password` with the correct current password and a valid new password THEN the system SHALL update the hash, revoke all other sessions of that user and record an audit entry with action `user.password_change` containing no password value.
6. IF the current password sent to `POST /api/v1/auth/password` is wrong THEN the system SHALL respond 403 with code `invalid_current_password`.
7. WHEN an actor with `identity:role:assign` sets the roles of a user that has an active administrative membership THEN the system SHALL replace the role set, record the previous and new role names in an audit entry with action `user.roles_set` and apply the new permissions from the next request.
8. WHEN an actor with `identity:user:read` calls `GET /api/v1/users` THEN the system SHALL return users with cursor pagination over (`created_at`, `id`), 50 per page by default, filterable by `active` and `role`, and IF `limit` exceeds 100 THEN the system SHALL respond 422 with code `invalid_limit`.
9. The system SHALL NOT return password hashes or session tokens in any API response, log line or audit entry.
10. WHEN an actor with `identity:user:update` reactivates a user THEN the system SHALL mark the user active and record an audit entry with action `user.reactivate`, and SHALL NOT restore any administrative role or closed administrative membership.
11. IF the new password sent to `POST /api/v1/auth/password` equals the current one THEN the system SHALL respond 422 with code `password_unchanged`.
12. IF five wrong current passwords are submitted to `POST /api/v1/auth/password` by the same user within 15 minutes THEN the system SHALL respond 429 with code `password_change_blocked` and a `Retry-After` header to every change attempt of that user during the following 15 minutes, counting them in `password_change_attempts`, separate from the login attempts.

**Independent Test**: Como PRESIDENTE, criar um usuário, promovê-lo a TESOURARIA (IDN-06), trocar papéis, desativá-lo e ver as sessões dele caírem; conferir tudo na auditoria.

---

### P1: Política de senha

**User Story**: Como mantenedor, quero regras de senha diferentes para quem administra e para o associado, e proteção contra senhas fracas, para reduzir o risco de contas invadidas.

**Why P1**: Contas administrativas acessam dados financeiros; a política precisa existir antes do primeiro usuário.

**Acceptance Criteria**:
1. IF the target user holds any role other than ASSOCIADO and the new password has fewer than 10 characters THEN the system SHALL reject it with 422 and code `password_too_short`.
2. IF the target user holds only the role ASSOCIADO and the new password has fewer than 8 characters THEN the system SHALL reject it with 422 and code `password_too_short`.
3. IF a password has more than 128 characters THEN the system SHALL reject it with 422 and code `password_too_long`.
4. IF a password matches, ignoring letter case, an entry of the embedded list of commonly used or compromised passwords THEN the system SHALL reject it with 422 and code `password_compromised`.
5. WHEN an administrative membership is granted to a user that had no active membership THEN the system SHALL set `must_change_password` to true for that user.
6. WHEN a user changes the password successfully THEN the system SHALL set `must_change_password` to false.
7. The system SHALL measure password length in Unicode code points and SHALL NOT impose composition rules.
8. WHEN an actor creates a user, or an administrator resets a user's password, THEN the system SHALL set `must_change_password` to true for that user.

**Independent Test**: Trocar a senha de uma conta administrativa por uma de 9 caracteres e ver a recusa; promover um ASSOCIADO a DIRETORIA e ver a troca de senha obrigatória.

---

### P1: Capacidade de e-mail

**User Story**: Como desenvolvedor de qualquer módulo, quero enviar e-mails por uma interface única, para não depender de um provedor específico.

**Why P1**: A recuperação de acesso precisa enviar e-mail, e outros módulos (comunicação, eventos) também vão precisar.

**Acceptance Criteria**:
1. The system SHALL expose an e-mail sending interface in `platform/email` that does not depend on any provider, and modules SHALL depend only on that interface.
2. IF a message has an invalid recipient address, an empty subject or a CR or LF character in the recipient or the subject THEN the sender SHALL reject it with an error and send nothing.
3. WHEN `EMAIL_PROVIDER` is `log` THEN the system SHALL log only the recipient domain and the subject of each message and SHALL NOT log the body; WHEN it is `disabled` THEN the system SHALL send nothing and log an operational warning at startup.
4. IF a delivery fails THEN the system SHALL log an operational incident and SHALL NOT expose the failure to the client.
5. IF `EMAIL_PROVIDER` has an unknown value THEN the API SHALL refuse to start, naming the variable.

**Independent Test**: Enviar com o remetente de teste e ver a mensagem; enviar com destinatário inválido e com quebra de linha no assunto e ver a recusa.

---

### P1: Recuperação de acesso por e-mail

**User Story**: Como usuário que esqueceu a senha, quero recuperar o acesso por e-mail, para voltar a entrar sem depender de um administrador.

**Why P1**: Sem recuperação, todo esquecimento vira chamado ao administrador.

**Acceptance Criteria**:
1. WHEN a client calls `POST /api/v1/auth/password-reset/request` with an email THEN the system SHALL respond 202 with the same body, "Se existir uma conta vinculada ao e-mail informado, enviaremos instruções.", whether or not an account exists, and the response SHALL NOT depend on the delivery of the e-mail.
2. WHEN the account exists and is active THEN the system SHALL create a 256-bit random token, store only its SHA-256 hash in `password_reset_tokens` with an expiry of 30 minutes, invalidate the user's earlier pending tokens and send the instructions with the link through `platform/email`, after the response.
3. The system SHALL NOT store, log or audit the token, the link or any password.
4. IF more than 3 requests for the same email hash arrive within one hour THEN the system SHALL respond as in AC 1 without creating a token or sending an e-mail.
5. WHEN a client calls `POST /api/v1/auth/password-reset/confirm` with a valid token and a new password that satisfies the password policy THEN the system SHALL, in one transaction, update the password hash, mark the token used, invalidate the user's other pending tokens, set `must_change_password` to false, revoke every session of the user and record an audit entry with action `auth.password_reset_completed`.
6. IF the token is unknown, expired or already used THEN the system SHALL respond 400 with code `invalid_reset_token`, identical for the three cases, and record an audit entry with action `auth.password_reset_failed`.
7. IF the new password violates the password policy THEN the system SHALL respond 422 with the policy code and SHALL keep the token valid.
8. WHEN a request or a confirmation is recorded THEN the system SHALL use the actions `auth.password_reset_requested` and `auth.password_reset_failed` as security events, with a context that has no email in clear text, token or link.
9. IF the user is inactive THEN the system SHALL NOT create a token or send an e-mail and SHALL respond as in AC 1.
10. The two routes SHALL be public but SHALL apply the origin check of `ALLOWED_ORIGINS`.

**Independent Test**: Solicitar a recuperação de um e-mail existente e de outro inexistente e ver respostas idênticas; usar o link uma vez, entrar com a senha nova e ver que o token não vale mais e que as sessões antigas caíram.

---

### P1: Redefinição administrativa de senha

**User Story**: Como administrador, quero disparar a redefinição de senha de outro usuário, para socorrer quem não consegue recuperar o acesso sozinho.

**Why P1**: Alternativa administrativa à recuperação por e-mail, sob controle de permissão.

**Acceptance Criteria**:
1. WHEN an actor with `identity:user:reset_password` calls `POST /api/v1/users/{id}/password-reset` with a reason THEN the system SHALL, in one transaction, generate a random temporary password, replace the password hash, set `must_change_password` to true, revoke every session and pending recovery token of the target user and record an audit entry with action `user.password_reset` containing the reason, and SHALL respond 200 with the temporary password exactly once, with `Cache-Control: no-store`.
6. IF the reason is missing or has fewer than 10 characters after trimming THEN the system SHALL respond 422 with code `reason_required` and change nothing.
7. The system SHALL NOT store, log or audit the temporary password, the old password or any hash, and SHALL NOT send them by e-mail.
2. IF the actor lacks a permission held by the target user THEN the system SHALL respond 403 with code `privilege_escalation` and change nothing.
3. IF the actor targets their own account THEN the system SHALL respond 403 with code `self_change_forbidden`.
4. IF the target user is inactive THEN the system SHALL respond 409 with code `user_inactive`.
5. WHEN the request is denied by AC 2 or 3 THEN the system SHALL record an audit entry with action `user.password_reset`, outcome `denied`, as a security event.

**Independent Test**: Como administrador, redefinir a senha de um usuário com motivo, receber a senha temporária uma vez, ver as sessões dele caírem e a troca obrigatória no próximo login; tentar com um alvo mais poderoso e ver 403.

---

### P1: Modelo de permissões e papéis

**User Story**: Como mantenedor, quero permissões granulares agrupadas em papéis, para que a autorização evolua sem reescrever regras.

**Why P1**: ADR-005 exige autorização por permissão, não por nome de papel.

**Acceptance Criteria**:
1. The system SHALL identify each permission by a string of the form `<module>:<resource>:<action>` matching `^[a-z_]+:[a-z_]+:[a-z_]+$` and SHALL reject the registration of any other form.
2. WHEN the API starts THEN the system SHALL synchronize the permissions declared in code and the role-to-permission matrix into the database, idempotently, under a database advisory lock so that instances starting together do not duplicate the work.
3. WHEN the synchronization changes the catalog or the matrix THEN the system SHALL record, in the same transaction, an audit entry with action `rbac.sync`, no actor, `before` and `after` holding the full state (permissions with their active flag and the permissions of each role) and a context listing the permissions added, deactivated and reactivated and the role-permission links added and removed, enough to rebuild the history; WHEN nothing changes THEN it SHALL record nothing.
4. The system SHALL provision the roles ASSOCIADO, PRESIDENTE, DIRETORIA, TESOURARIA, ESTOQUE_LOJA, EVENTOS, CONSELHO_FISCAL and ADMIN_SISTEMA.
5. The system SHALL NOT grant the role CONSELHO_FISCAL any permission whose action is `create`, `update`, `delete` or `cancel`.
6. The system SHALL grant the role CONSELHO_FISCAL the institutional permissions `financeiro:prestacao_contas:read`, `financeiro:prestacao_contas:approve` and `financeiro:parecer:opine`.
7. WHEN a user has more than one active role THEN the effective permissions SHALL be the union of the permissions of those roles.
8. IF a permission stored in the database is no longer declared in code THEN the synchronization SHALL mark it inactive and SHALL NOT delete it, and an inactive permission SHALL NOT count in the effective permissions.
9. The system SHALL build the matrix from the contributions (permissions and role grants) declared by each module, and SHALL refuse to start when the aggregated matrix has an invalid permission, a grant of an undeclared permission, or a forbidden action for CONSELHO_FISCAL.
10. The system SHALL grant PRESIDENTE every permission of the aggregated catalog as explicit role-permission links created by the synchronization, without wildcards; a new permission reaches PRESIDENTE only through the synchronization.

**Independent Test**: Subir a API duas vezes e ver a matriz idêntica e uma única auditoria `rbac.sync`; rodar o teste de invariante do Conselho Fiscal.

---

### P1: Vínculo administrativo e promoção

**User Story**: Como presidente ou administrador de sistema, quero conceder e retirar acesso administrativo com o motivo registrado, separado do que cada papel permite, para saber por que alguém administra e poder retirar isso sem perder o histórico.

**Why P1**: Acesso administrativo é o maior risco de segurança da plataforma; precisa de motivo, rastro e um único caminho de concessão.

**Acceptance Criteria**:
1. WHEN an actor with `identity:admin:grant` calls `POST /api/v1/users/{id}/admin-membership` with role names and a reason THEN the system SHALL, in one transaction, create an active administrative membership (user, reason, grantor, timestamp), assign the roles, set `must_change_password` according to IDN-05.5, revoke every session of the target user and record an audit entry with action `admin.promote` containing the previous and new role names.
2. IF the reason is missing or has fewer than 10 characters after trimming THEN the system SHALL respond 422 with code `reason_required` and change nothing.
3. IF the target user already has an active administrative membership THEN the system SHALL respond 409 with code `already_admin`.
4. IF the target user is inactive THEN the system SHALL respond 409 with code `user_inactive`.
5. The system SHALL keep at most one active administrative membership per user, enforced by the database, and SHALL NOT delete membership rows.
6. WHEN an actor with `identity:admin:revoke` calls `POST /api/v1/users/{id}/admin-membership/revoke` with a reason in the body THEN the system SHALL, in one transaction, close the membership recording who, when and why, remove every role other than ASSOCIADO, revoke every session of the target user and record an audit entry with action `admin.revoke`.
7. IF the target user has no active administrative membership THEN the system SHALL respond 409 with code `not_admin`.
8. IF the operation would leave no active user that holds `identity:admin:grant` THEN the system SHALL respond 409 with code `last_admin` and change nothing.
9. The system SHALL NOT allow a user to hold a role other than ASSOCIADO without an active administrative membership; adding such a role without one SHALL respond 409 with code `admin_membership_required`.
10. IF the roles of a promotion contain no role other than ASSOCIADO THEN the system SHALL respond 422 with code `admin_role_required`; the role ASSOCIADO SHALL always remain with the user, and no operation removes it.

**Independent Test**: Como PRESIDENTE, promover um usuário a TESOURARIA com motivo, ver a troca de senha obrigatória e as sessões caírem; retirar o acesso e ver o histórico do vínculo preservado.

---

### P1: Concessão sem escalada de privilégio

**User Story**: Como mantenedor, quero que ninguém consiga dar a outra pessoa mais poder do que já tem, para que uma conta comprometida ou mal-intencionada não amplie o próprio acesso.

**Why P1**: Sem essa regra, quem administra acesso poderia criar contas com permissões que ele próprio não possui.

**Acceptance Criteria**:
1. IF an actor tries to grant, in a promotion or in a role assignment, a role with any permission the actor does not hold THEN the system SHALL respond 403 with code `privilege_escalation` and change nothing.
2. IF an actor tries to remove a role, or the administrative membership, of a user when that role holds a permission the actor does not hold THEN the system SHALL respond 403 with code `privilege_escalation` and change nothing.
3. IF an actor tries to change their own roles or administrative membership, deactivate their own account or reset their own password as an administrator THEN the system SHALL respond 403 with code `self_change_forbidden` and change nothing.
4. The system SHALL evaluate these rules with the effective permissions of the actor at the moment of the request.
5. WHEN a change is denied by AC 1, 2 or 3 THEN the system SHALL record an audit entry with action `role.change_denied` containing the actor, the target and the requested roles, in its own transaction; WHEN a deactivation, reactivation or administrative password reset is denied by AC 3 or 6 THEN it SHALL record the action of the operation (`user.deactivate`, `user.reactivate` or `user.password_reset`) with outcome `denied`, as a security event.
6. IF an actor tries to deactivate, reactivate or reset the password of a user whose effective permissions are not all held by the actor THEN the system SHALL respond 403 with code `privilege_escalation` and change nothing.

**Independent Test**: Como ator cujas permissões efetivas não cobrem CONSELHO_FISCAL, tentar promover alguém a esse papel e ver 403 `privilege_escalation`; como PRESIDENTE, tentar alterar o próprio papel e ver 403 `self_change_forbidden`.

---

### P1: Autorização negada por padrão

**User Story**: Como mantenedor, quero que nada seja acessível sem autenticação e permissão explícitas, para não expor dados por esquecimento.

**Why P1**: Um endpoint esquecido sem proteção é o erro mais provável e mais caro.

**Acceptance Criteria**:
1. IF a request reaches a route that is not on the public allowlist and carries no valid session THEN the system SHALL respond 401 with code `unauthenticated`.
2. IF the authenticated principal lacks the permission required by a use case THEN the system SHALL respond 403 with code `forbidden` and SHALL NOT reveal whether the requested resource exists.
3. WHEN a use case is invoked THEN it SHALL check the required permission through `authz.Require` before performing any read or write.
4. The test suite SHALL fail if any registered route is neither on the public allowlist nor protected by the authentication middleware.
5. WHILE a user is inactive the system SHALL treat every session of that user as invalid.
6. WHILE `must_change_password` is true for the authenticated user the system SHALL respond 403 with code `password_change_required` to every route except `POST /api/v1/auth/logout`, `GET /api/v1/auth/me` and `POST /api/v1/auth/password`.
7. WHEN a permission check fails THEN the system SHALL record an audit entry with action `authz.denied`, outcome `denied`, the user as actor and the required permission as target, as a security event, unless the permission is declared as a common read; a failure to record it SHALL NOT change the 403 response.
8. The system SHALL NOT make an authorization decision from a role name, and the test suite SHALL fail if a role name literal appears in source code outside the role definitions, or if the roles of a principal are read outside the identity module.

**Independent Test**: Chamar cada rota sem sessão e ver 401; chamar com um papel sem a permissão e ver 403.

---

### P1: Registro de auditoria atômico

**User Story**: Como diretor e conselheiro fiscal, quero que toda alteração relevante fique registrada junto com a própria alteração, para confiar no histórico.

**Why P1**: É a base da prestação de contas (ADR-004).

**Acceptance Criteria**:
1. WHEN a use case that changes audited data commits THEN the system SHALL have written the audit entry in the same database transaction as the change.
2. IF writing the audit entry fails THEN the system SHALL roll back the whole transaction and respond 500 with code `audit_failed`.
3. The system SHALL record for every entry an id (UUID), `occurred_at` (UTC), `actor_type`, `actor_user_id` (present only when `actor_type` is `user`), `action`, the target as `entity_type` and `entity_id`, `before` (JSON or null), `after` (JSON or null), `outcome`, an optional `reason`, a `context` object (JSON, no secrets) and the `request_id`.
4. IF an entry has an action of cancellation or extraordinary adjustment and no reason THEN the recorder SHALL reject it with an error.
5. IF `before` or `after` contains a key named `password`, `password_hash`, `token`, `session_token` or `csrf_token` THEN the recorder SHALL replace its value with `[redacted]`.
6. IF the recorder is called outside a database transaction THEN it SHALL return an error and SHALL NOT write the entry.

**Independent Test**: Forçar falha na gravação da auditoria e ver a alteração desfeita.

---

### P1: Eventos de segurança

**User Story**: Como presidente, conselheiro fiscal ou administrador de sistema, quero que login, falhas de autenticação e mudanças de acesso fiquem registrados de forma imutável, para investigar tentativas de invasão e abusos de poder.

**Why P1**: Sem esse rastro, não há como saber quem tentou entrar, quem ganhou ou perdeu acesso e quem tentou escalar privilégio.

**Acceptance Criteria**:
1. WHEN a login succeeds THEN the system SHALL record an entry with action `auth.login` and the user as actor.
2. WHEN a login fails because the email is unknown, the password is wrong or the user is inactive THEN the system SHALL record an entry with action `auth.login_failed`, outcome `failure`, actor type `anonymous`, the user as target when it exists and a context with the failure category and a hash of the email, and SHALL NOT store the password or the plain email.
3. WHEN a login attempt is refused because of the lockout THEN the system SHALL record an entry with action `auth.login_blocked` and outcome `denied`.
4. WHEN a user logs out THEN the system SHALL record an entry with action `auth.logout`.
5. The system SHALL record an audit entry for every change or denied change of roles, permissions or administrative membership: `admin.promote`, `admin.revoke`, `user.roles_set`, `rbac.sync` and `role.change_denied`; and for passwords `auth.password_reset_requested`, `auth.password_reset_completed`, `auth.password_reset_failed` and `user.password_reset`.
6. WHEN a security event is recorded THEN the system SHALL write it in its own transaction; IF that write fails THEN the system SHALL log an operational incident with the action and the `request_id` and SHALL NOT change the response to the client.
7. IF an entry carries an action that is not in the catalog of actions declared in code THEN the recorder SHALL reject it with an error.

**Independent Test**: Errar a senha até o bloqueio e ver `auth.login_failed` e `auth.login_blocked`; promover, retirar e tentar escalar privilégio e ver os eventos correspondentes na consulta de auditoria.

---

### P1: Auditoria imutável

**User Story**: Como conselheiro fiscal, quero a garantia técnica de que o histórico não pode ser reescrito, para emitir parecer com confiança.

**Why P1**: A imutabilidade é o requisito central do ADR-004.

**Acceptance Criteria**:
1. IF any UPDATE, DELETE or TRUNCATE is executed against `audit_log` THEN the database SHALL reject it with an exception, for any role.
2. The application database role SHALL hold only the INSERT and SELECT privileges on `audit_log`.
3. The system SHALL NOT expose any API operation that updates or deletes audit entries.

**Independent Test**: Conectar com o papel da aplicação e tentar UPDATE e DELETE; ambos falham.

---

### P2: Consulta de auditoria

**User Story**: Como administrador ou presidente, quero consultar o histórico, para investigar quem alterou o quê.

**Why P2**: Necessária para usar a auditoria, mas a gravação (P1) vem antes.

**Acceptance Criteria**:
1. WHEN an actor with `audit:log:read` calls `GET /api/v1/audit-logs` THEN the system SHALL return entries newest first with cursor pagination, 50 per page by default.
2. WHEN the filters `entity_type`, `entity_id`, `actor_user_id`, `action`, `outcome`, `from` or `to` are supplied THEN the system SHALL return only the entries matching all of them.
3. IF `limit` exceeds 100 THEN the system SHALL respond 422 with code `invalid_limit`.
4. IF the actor lacks `audit:log:read` THEN the system SHALL respond 403 with code `forbidden`.

**Independent Test**: Criar um usuário e achar o registro `user.create` pelo filtro de entidade.

---

### P1: Tipo monetário em centavos

**User Story**: Como desenvolvedor, quero um tipo de dinheiro exato, para nunca introduzir erro de arredondamento.

**Why P1**: ADR-003; toda regra financeira depende dele.

**Acceptance Criteria**:
1. The money package SHALL represent amounts as the type `Cents`, an int64 count of centavos of BRL.
2. WHEN two amounts are added or subtracted THEN the system SHALL return the exact result.
3. IF an addition or subtraction overflows int64 THEN the system SHALL return `ErrOverflow` and SHALL NOT wrap around.
4. WHEN two amounts are compared THEN `Cmp` SHALL return -1, 0 or 1.
5. The money package SHALL NOT expose any function that accepts or returns float32 or float64.

**Independent Test**: Testes unitários do pacote, com cobertura mínima de 95%.

---

### P1: Serialização e formatação de dinheiro

**User Story**: Como desenvolvedor, quero dinheiro serializado e exibido de forma idêntica no Go, na API e no TypeScript, para que o valor nunca mude no caminho.

**Why P1**: Um valor que muda entre camadas quebra a conferência de saldo.

**Acceptance Criteria**:
1. WHEN a `Cents` value is marshaled to JSON THEN the system SHALL emit an integer literal with no quotes, no decimal point and no exponent.
2. IF JSON being unmarshaled into `Cents` is a string, a decimal, an exponent form or null THEN the system SHALL return an error.
3. IF the magnitude of a JSON amount being unmarshaled exceeds 9007199254740991 THEN the money package SHALL return `ErrOutOfRange`, and the API layer of the first feature with a monetary endpoint SHALL map it to 422 with code `amount_out_of_range`.
4. WHEN 123456 is formatted THEN the Go and TypeScript formatters SHALL both return `R$ 1.234,56`, and for -123456 they SHALL return `-R$ 1.234,56`, using an ordinary space.
5. WHEN the text `1.234,56` is parsed THEN the system SHALL return 123456.
6. IF the text to parse has more than two decimal places or characters outside digits, dots and one comma THEN the system SHALL return an error.
7. IF a migration adds a column of type `real`, `double precision` or `numeric` THEN the schema guard test SHALL fail.
8. IF a column name ends with `_cents` and its type is not `bigint` THEN the schema guard test SHALL fail.
9. IF a TypeScript money helper receives a number that is not a safe integer THEN it SHALL throw a RangeError.

**Independent Test**: Os mesmos vetores de teste passam no Go e no TypeScript.

---

### P2: Rateio e percentuais

**User Story**: Como desenvolvedor do financeiro, quero dividir valores e aplicar taxas sem perder centavos, para suportar parcelamento, rateio e taxas.

**Why P2**: Usada pelo financeiro logo após a fundação, mas não é pré-requisito dela.

**Acceptance Criteria**:
1. WHEN `Allocate(total, weights)` is called with positive integer weights THEN the system SHALL return parts that sum exactly to `total`, giving the remainder cents one each to the first parts in order.
2. IF `weights` is empty or contains a value less than or equal to zero THEN the system SHALL return an error.
3. WHEN `total` is negative THEN the system SHALL allocate its magnitude and negate every part, so that the parts sum to `total`.
4. WHEN `Percent(amount, basisPoints)` is called THEN the system SHALL return the amount times basis points divided by 10000, rounded half up, with negative amounts rounded symmetrically away from zero.
5. IF the intermediate product of `Percent` overflows int64 THEN the system SHALL still return the exact result when it fits in int64 and SHALL return `ErrOverflow` otherwise.
6. The test suite SHALL include a property test showing that the parts returned by `Allocate` always sum to `total`.

**Independent Test**: `Allocate(10000, [1,1,1])` retorna 3334, 3333, 3333 e `Percent(10000, 400)` retorna 400.

---

### P1: Infraestrutura de testes de integração (Go)

**User Story**: Como desenvolvedor, quero testar contra um PostgreSQL real, para pegar problemas de SQL, transação e permissão que mocks escondem.

**Why P1**: Sem ela, as garantias de auditoria e papéis de banco não são verificáveis.

**Acceptance Criteria**:
1. WHERE the build tag `integration` is absent THEN `go test ./...` SHALL run without Docker and SHALL NOT execute integration tests.
2. WHEN integration tests run THEN the system SHALL start PostgreSQL 16 through testcontainers once per package, apply all migrations and give each test an isolated database.
3. WHEN a test finishes THEN its data SHALL NOT be visible to any other test, proven by a test that writes in one test and reads in another.
4. WHEN integration tests connect to the database THEN they SHALL use the non-privileged application role so that grants and triggers are exercised.
5. WHEN CI runs the `api` job THEN it SHALL execute `go test -tags=integration ./...`.
6. IF Docker is unavailable while the tag `integration` is set THEN the tests SHALL fail with a message naming Docker and SHALL NOT skip silently.

**Independent Test**: `go test -tags=integration ./...` passa no CI; `go test ./...` passa numa máquina sem Docker.

---

### P2: Verificação de fronteiras entre módulos

**User Story**: Como mantenedor, quero que as regras de fronteira sejam verificadas por teste, para que o monólito continue modular.

**Why P2**: `domain-boundaries.md` (regra 8) deixou essa verificação para esta spec.

**Acceptance Criteria**:
1. The test suite SHALL fail if a package under `api/internal/<module>/` imports the `domain`, `infra` or `http` package of a different business module.
2. The test suite SHALL fail if a package under `api/internal/platform/` imports any business module package.
3. WHEN the checker is given a fixture that violates each rule THEN it SHALL report a violation, proven by negative tests.

**Independent Test**: Adicionar um import proibido num fixture e ver o teste falhar.

---

### P2: Infraestrutura de testes do front

**User Story**: Como desenvolvedor, quero testes de unidade no front, para validar o formatador de dinheiro e o que vier depois.

**Why P2**: Necessária para o TypeScript de dinheiro, mas não bloqueia o backend.

**Acceptance Criteria**:
1. WHEN `pnpm test` runs in `web/` THEN the system SHALL execute all `*.test.ts` and `*.test.tsx` files with Vitest in a jsdom environment.
2. WHEN CI runs the `web` job THEN it SHALL run `pnpm test` before the build.
3. The repository SHALL contain tests for `web/lib/money.ts` that cover the vectors of the money serialization story.

**Independent Test**: `pnpm test` passa localmente e no CI.

---

### P1: Contrato OpenAPI

**User Story**: Como desenvolvedor, quero um contrato único da API, para que servidor, cliente e testes não divirjam.

**Why P1**: O front consome tipos gerados; sem o contrato, as camadas divergem em silêncio.

**Acceptance Criteria**:
1. The system SHALL keep the HTTP contract as OpenAPI 3.0.3 documents under `api/openapi/`, one per module plus the shared `common.yaml`, as the source of truth for every HTTP route under `/api/v1`.
2. WHEN a module contract changes THEN the generated server interface and models of that module SHALL be regenerated and committed, and CI SHALL fail if generation produces a difference.
3. WHEN a module contract changes THEN the TypeScript types under `web/lib/api/` SHALL be regenerated and committed, and CI SHALL fail if generation produces a difference.
4. IF any contract fails linting THEN CI SHALL fail.
5. WHEN an integration test calls an implemented endpoint THEN the response status, headers and body SHALL be validated against the contract of the module that owns the endpoint, and a mismatch SHALL fail the test.
6. IF a route is registered in the Gin router without a matching operation in any module contract THEN a test SHALL fail.

**Independent Test**: Alterar o YAML de um módulo sem regenerar e ver o CI falhar.

---

### P1: Convenções da API

**User Story**: Como consumidor da API, quero erros, paginação e identificadores previsíveis, para tratar respostas de forma uniforme.

**Why P1**: Toda rota nova herda estas convenções.

**Acceptance Criteria**:
1. WHEN an error response is returned THEN the system SHALL use `application/problem+json` with the fields `type`, `title`, `status`, `detail`, `code` and `request_id`.
2. WHEN a request arrives THEN the system SHALL use the `X-Request-Id` header when it matches `^[A-Za-z0-9-]{8,64}$`, otherwise generate a UUID v4, and SHALL echo the id in the response header `X-Request-Id`.
3. IF the request body is not valid JSON THEN the system SHALL respond 400 with code `invalid_json`.
4. IF the request body fails validation THEN the system SHALL respond 422 with code `validation_failed` and an `errors` list whose items carry `field` and `code`.
5. WHEN a list endpoint is called THEN the system SHALL accept `limit` (default 50, maximum 100) and `cursor`, and SHALL return `items` and `next_cursor`.
6. IF an unhandled panic occurs THEN the system SHALL respond 500 with code `internal_error` and the `request_id`, and SHALL NOT include internal messages or stack traces in the response.
7. IF a request body exceeds 1 MiB on an authenticated route, or 64 KiB on a public route, THEN the system SHALL respond 413 with code `payload_too_large`, and the contract SHALL describe this response.

**Independent Test**: Provocar cada erro e validar o corpo contra o contrato.

---

### P2: Configuração, logs e migrações

**User Story**: Como operador, quero uma API que falhe cedo em configuração errada e não vaze segredos em logs, para operar com segurança.

**Why P2**: Endurece a operação; as demais histórias funcionam sem ela, mas não devem ir a produção sem ela.

**Acceptance Criteria**:
1. IF a required environment variable is missing or invalid at startup THEN the process SHALL exit with a non-zero status naming the variable and SHALL NOT print any secret value.
2. IF `APP_ENV` is not `development` and `COOKIE_SECURE` is false THEN the process SHALL refuse to start.
3. WHEN a request completes THEN the system SHALL write one structured JSON log line with method, path without query string, status, duration in milliseconds, `request_id` and the user id when present.
4. IF a log field key is `password`, `password_hash`, `token`, `session_token`, `csrf_token`, `cookie` or `authorization` THEN the logger SHALL replace its value with `[redacted]`.
5. The system SHALL NOT apply migrations automatically when the API starts.
6. The API SHALL connect to PostgreSQL with the application role, which holds no DDL privileges.
7. IF GORM logs a query THEN it SHALL log the parameterized statement and SHALL NOT log bound parameter values.

**Independent Test**: Iniciar com variável faltando e ver a mensagem; provocar um log com senha e ver `[redacted]`.

---

## Edge Cases

- IF two requests create the same email concurrently THEN the system SHALL let exactly one succeed and respond 409 to the other.
- IF two logins for the same user happen concurrently THEN the system SHALL create two independent sessions.
- IF the session cookie is present but malformed THEN the system SHALL respond 401 with code `unauthenticated` and SHALL NOT raise an internal error.
- IF the database is unavailable during a request THEN the system SHALL respond 503 with code `service_unavailable`.
- IF the roles of a user are replaced with an empty role set THEN the system SHALL respond 422 with code `validation_failed`.
- IF two administrators promote the same user concurrently THEN the system SHALL let exactly one succeed and respond 409 `already_admin` to the other.
- WHEN a role is removed from a user THEN the system SHALL stop granting the permissions of that role from the next request.
- WHEN `Allocate` receives a total of zero THEN the system SHALL return parts that are all zero.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
|---|---|---|---|
| IDN-01 | P1: Primeiro administrador | In Tasks | Implementing |
| IDN-02 | P1: Login, sessão e logout | In Tasks | Implementing |
| IDN-03 | P1: Proteção contra CSRF e origem | In Tasks | Implementing |
| IDN-04 | P1: Gestão de usuários | In Tasks | Implementing |
| IDN-05 | P1: Política de senha | In Tasks | Implementing |
| IDN-06 | P1: Vínculo administrativo e promoção | In Tasks | Implementing |
| IDN-07 | P1: Recuperação de acesso por e-mail | In Tasks | Implementing |
| IDN-08 | P1: Redefinição administrativa de senha | In Tasks | Implementing |
| EML-01 | P1: Capacidade de e-mail | In Tasks | Implementing |
| RBAC-01 | P1: Modelo de permissões e papéis | In Tasks | Implementing |
| RBAC-02 | P1: Autorização negada por padrão | In Tasks | Implementing |
| RBAC-03 | P1: Concessão sem escalada de privilégio | In Tasks | Implementing |
| AUD-01 | P1: Registro de auditoria atômico | In Tasks | Implementing |
| AUD-02 | P1: Auditoria imutável | In Tasks | Implementing |
| AUD-03 | P2: Consulta de auditoria | In Tasks | Implementing |
| AUD-04 | P1: Eventos de segurança | In Tasks | Implementing |
| MNY-01 | P1: Tipo monetário em centavos | In Tasks | Implementing |
| MNY-02 | P1: Serialização e formatação de dinheiro | In Tasks | Implementing |
| MNY-03 | P2: Rateio e percentuais | In Tasks | Implementing |
| TST-01 | P1: Infraestrutura de testes de integração (Go) | In Tasks | Pending |
| TST-02 | P2: Verificação de fronteiras entre módulos | In Tasks | Pending |
| TST-03 | P2: Infraestrutura de testes do front | In Tasks | Pending |
| API-01 | P1: Contrato OpenAPI | In Tasks | Implementing |
| API-02 | P1: Convenções da API | In Tasks | Implementing |
| PLT-01 | P2: Configuração, logs e migrações | In Tasks | Implementing |

**Coverage:** 25 total, 25 mapped to tasks, 0 unmapped (mapeamento detalhado em `tasks.md`, seção Requirement Coverage).

---

## Success Criteria

- [ ] `go test -tags=integration ./...` e `pnpm test` passam no CI, junto com o lint, o build e as verificações de contrato OpenAPI.
- [ ] Nas rotas da fundação, um PRESIDENTE cria um usuário, promove com motivo, atribui papel, retira o acesso e desativa, tudo com registro de auditoria consultável pela API.
- [ ] Ninguém concede permissão que não possui nem altera o próprio acesso, e toda alteração de permissão (promoção, retirada, papéis, sincronização da matriz) fica na auditoria.
- [ ] Um usuário recupera o acesso por e-mail com token de uso único, e a resposta da solicitação é idêntica para contas existentes e inexistentes; a redefinição administrativa segue a regra sem escalada.
- [ ] Senhas fora da política (10 para qualquer papel diferente de ASSOCIADO, 8 para associados, lista de comprometidas) são recusadas e a troca obrigatória bloqueia o resto da API.
- [ ] O teste de invariante confirma que CONSELHO_FISCAL não tem nenhuma permissão `create`, `update`, `delete` ou `cancel` e tem as institucionais de prestação de contas, aprovação e parecer.
- [ ] O papel de banco da aplicação não consegue fazer UPDATE, DELETE ou TRUNCATE em `audit_log`, provado por teste.
- [ ] Os mesmos vetores de dinheiro passam no Go e no TypeScript, e o teste de esquema barra colunas `float`, `double` e `numeric`.
