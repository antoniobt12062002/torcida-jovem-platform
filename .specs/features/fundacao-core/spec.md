# Fundação Core Specification

Status: Aprovada com ajustes em 2026-09-26. Feature 1 de 2 da fundação (a outra é `fundacao-documentos`). Narrativa em português; os critérios de aceite (ACs) seguem o padrão EARS em inglês, exigido pelo validador `validate_spec.py`.

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
| Recuperação de senha por e-mail | Exige provedor de e-mail, ainda não definido |
| MFA, OAuth e provedor externo de identidade | ADR-005 adia; pede novo ADR |
| Entidades e regras do financeiro: lançamentos, receitas, despesas, plano de contas, produtos e eventos | Módulos de domínio posteriores; esta feature entrega só o núcleo monetário compartilhado (`platform/money`), que eles importam |
| Documentos e object storage | Feature `fundacao-documentos`, executada depois desta |
| Hospedagem e deploy reais | Decisão adiada; ver ADR-002 e o plano de hospedagem |
| Purga e retenção de auditoria e sessões | Política de retenção depende de definição jurídica e de LGPD |
| Fechamento de ano e ajuste extraordinário | Regras do financeiro (FIN-001, seção 16) |
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
| Troca obrigatória de senha | Ao promover um usuário que só era ASSOCIADO a qualquer papel administrativo, o sistema marca `must_change_password=true` e exige a troca no próximo login; até lá o usuário só pode trocar a senha, ver `me` ou sair | Aprovado pelo mantenedor; evita que uma senha de 8 caracteres passe a proteger conta administrativa | y |
| Valor padrão de `APP_ENV` | Variável ausente equivale a `production` | Aprovado pelo mantenedor; padrão seguro, e `development` passa a ser explícito | y |
| `ALLOWED_ORIGINS` vazio | Mantém o comportamento seguro: nenhuma origem é autorizada e requisições de navegador com `Origin` são bloqueadas; em `development` o padrão é `http://localhost:3000` | Aprovado pelo mantenedor; falha fechada em vez de aberta | y |
| Cadastro de usuários | Sem autocadastro; o primeiro ADMIN nasce por CLI e os demais são criados por quem tem permissão | Sistema institucional fechado; reduz superfície de ataque | n |
| Matriz inicial de permissões | Provisória e mínima: ADMIN e PRESIDENTE com todas as permissões da fundação; DIRETOR só `identity:user:read`; FINANCEIRO, CONSELHO_FISCAL e ASSOCIADO sem permissões da fundação (as de documentos entram em `fundacao-documentos`) | Cada spec de módulo ampliará a matriz do seu módulo | n |
| Ações que o Conselho Fiscal nunca recebe | `create`, `update`, `delete` e `cancel`; recebe `read`, `approve` e `opine` | Decisão já tomada pelo mantenedor (ADR-005, FIN-001 seção 24) | y |
| Conteúdo do registro de auditoria | Sem endereço IP; com `request_id`; retenção indefinida nesta fase | IP é dado pessoal (LGPD) e não é necessário agora | n |
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
| Validação de requisições pelo contrato | O contrato valida estrutura, tipos, formatos, enums e campos obrigatórios (corpo, query e path), de forma automática, por middleware próprio em `platform/httpx` sobre o `kin-openapi`; regras de negócio, autenticação e autorização ficam nos módulos de domínio. Rota sem operação no contrato segue para o handler (recusá-la em produção é decisão futura) | O `gin-middleware` v1.1.0 entrega ao tratador só a mensagem em texto, sem os erros por campo que API-02.4 exige; o contrato não muda para se adequar à ferramenta. Aprovado pelo mantenedor (opção A) | y |
| Limite de responsabilidade do contrato | O OpenAPI define contratos de comunicação (caminhos, esquemas, segurança, exemplos) e não contém regras de negócio, que permanecem nas specs e nos módulos de domínio | Aprovado pelo mantenedor; evita duplicar regra em dois lugares | y |
| Limpeza de sessões e tentativas | Registros expirados são ignorados e podem ser removidos de forma oportunista; sem job de purga | Evita infraestrutura de tarefas agendadas nesta fase | n |
| Papéis de banco | Migrações rodam com o papel dono; a API roda com um papel de aplicação sem DDL e sem UPDATE/DELETE em auditoria | Defesa em profundidade para o ADR-004 | n |

**Open questions:** none - all resolved or logged above.

---

## User Stories

### P1: Primeiro administrador ⭐ MVP

**User Story**: Como operador do sistema, quero criar o primeiro ADMIN por linha de comando, para poder entrar no sistema sem autocadastro.

**Why P1**: Sem um usuário inicial, nada mais pode ser usado.

**Acceptance Criteria**:
1. WHEN the operator runs `bootstrap-admin` with `--email` and the password supplied through the environment variable `BOOTSTRAP_ADMIN_PASSWORD` THEN the system SHALL create an active user with the role ADMIN and record an audit entry with action `user.bootstrap`.
2. IF a user with the role ADMIN already exists THEN the system SHALL exit with a non-zero status without creating any user.
3. IF the password is absent or violates the administrator password policy (at least 10 characters) THEN the system SHALL exit with a non-zero status without creating any user.
4. The system SHALL NOT accept the password as a command-line argument.

**Independent Test**: Rodar o comando contra um banco vazio e ver o usuário e o registro de auditoria; rodar de novo e ver a recusa.

---

### P1: Login, sessão e logout

**User Story**: Como usuário, quero entrar com e-mail e senha e manter uma sessão segura, para acessar as áreas conforme minhas permissões.

**Why P1**: Toda autorização parte de uma sessão válida.

**Acceptance Criteria**:
1. WHEN a user submits a valid email and password to `POST /api/v1/auth/login` THEN the system SHALL respond 200, create a session and set the cookie `tj_session` with the attributes HttpOnly, Secure, SameSite=Lax and Path=/.
2. WHEN a session is created THEN the system SHALL store only the SHA-256 hash of the session token in the database.
3. IF the email does not exist or the password is wrong THEN the system SHALL respond 401 with code `invalid_credentials` and an identical body for both cases.
4. IF the user is inactive THEN the system SHALL respond 401 with code `invalid_credentials` and the same body as for a wrong password.
5. WHILE a session has been idle for more than 60 minutes or is older than 8 hours the system SHALL treat it as expired and respond 401 with code `session_expired`.
6. WHEN a user calls `POST /api/v1/auth/logout` THEN the system SHALL revoke the session and clear the cookie.
7. WHEN a login succeeds THEN the system SHALL issue a new session token even if a session cookie was sent with the request.
8. IF a fifth consecutive login failure for the same email occurs within 15 minutes THEN the system SHALL respond 429 with a `Retry-After` header to every login attempt for that email during the following 15 minutes.
9. WHEN a session is valid THEN `GET /api/v1/auth/me` SHALL respond 200 with the user id, email, name, role names, effective permissions, the CSRF token and the flag `must_change_password`.
10. The system SHALL NOT lock an account permanently as a result of failed logins.

**Independent Test**: Logar com o ADMIN criado pela CLI, chamar `/auth/me`, deslogar e ver que a sessão deixou de valer.

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

**Independent Test**: Enviar um POST autenticado sem o cabeçalho e com origem estranha e ver os dois 403.

---

### P1: Gestão de usuários

**User Story**: Como administrador, quero criar, desativar e atribuir papéis a usuários, para dar e retirar acesso com rastreabilidade.

**Why P1**: Sem usuários e papéis não há quem use os módulos.

**Acceptance Criteria**:
1. WHEN an actor with `identity:user:create` calls `POST /api/v1/users` with email, name, initial password and role names THEN the system SHALL create the user, respond 201 and record an audit entry with action `user.create`.
2. IF the email already exists, ignoring letter case, THEN the system SHALL respond 409 with code `email_taken`.
3. WHEN an actor with `identity:user:update` deactivates a user THEN the system SHALL mark the user inactive, revoke all of that user's sessions in the same transaction and record an audit entry with action `user.deactivate`.
4. IF an actor tries to deactivate the last active ADMIN THEN the system SHALL respond 409 with code `last_admin`.
5. WHEN a user calls `POST /api/v1/auth/password` with the correct current password and a valid new password THEN the system SHALL update the hash, revoke all other sessions of that user and record an audit entry with action `user.password_change` containing no password value.
6. IF the current password sent to `POST /api/v1/auth/password` is wrong THEN the system SHALL respond 403 with code `invalid_current_password`.
7. WHEN an actor with `identity:role:assign` sets the roles of a user THEN the system SHALL replace the role set, record the previous and new role names in an audit entry with action `user.roles_set` and apply the new permissions from the next request.
8. WHEN an actor with `identity:user:read` calls `GET /api/v1/users` THEN the system SHALL return users with cursor pagination.
9. The system SHALL NOT return password hashes or session tokens in any API response, log line or audit entry.

**Independent Test**: Como ADMIN, criar um usuário FINANCEIRO, atribuir e trocar papéis, desativá-lo e ver as sessões dele caírem; conferir tudo na auditoria.

---

### P1: Política de senha

**User Story**: Como mantenedor, quero regras de senha diferentes para quem administra e para o associado, e proteção contra senhas fracas, para reduzir o risco de contas invadidas.

**Why P1**: Contas administrativas acessam dados financeiros; a política precisa existir antes do primeiro usuário.

**Acceptance Criteria**:
1. IF the target user holds any role other than ASSOCIADO and the new password has fewer than 10 characters THEN the system SHALL reject it with 422 and code `password_too_short`.
2. IF the target user holds only the role ASSOCIADO and the new password has fewer than 8 characters THEN the system SHALL reject it with 422 and code `password_too_short`.
3. IF a password has more than 128 characters THEN the system SHALL reject it with 422 and code `password_too_long`.
4. IF a password matches, ignoring letter case, an entry of the embedded list of commonly used or compromised passwords THEN the system SHALL reject it with 422 and code `password_compromised`.
5. WHEN an actor adds a role other than ASSOCIADO to a user who held only the role ASSOCIADO THEN the system SHALL set `must_change_password` to true for that user.
6. WHEN a user changes the password successfully THEN the system SHALL set `must_change_password` to false.
7. The system SHALL measure password length in Unicode code points and SHALL NOT impose composition rules.

**Independent Test**: Criar um usuário FINANCEIRO com senha de 9 caracteres e ver a recusa; promover um ASSOCIADO a DIRETOR e ver a troca de senha obrigatória.

---

### P1: Modelo de permissões e papéis

**User Story**: Como mantenedor, quero permissões granulares agrupadas em papéis, para que a autorização evolua sem reescrever regras.

**Why P1**: ADR-005 exige autorização por permissão, não por nome de papel.

**Acceptance Criteria**:
1. The system SHALL identify each permission by a string of the form `<module>:<resource>:<action>` matching `^[a-z_]+:[a-z_]+:[a-z_]+$` and SHALL reject the registration of any other form.
2. WHEN the API starts THEN the system SHALL synchronize the permissions declared in code and the role-to-permission matrix into the database, idempotently.
3. The system SHALL provision the roles ADMIN, PRESIDENTE, DIRETOR, FINANCEIRO, CONSELHO_FISCAL and ASSOCIADO.
4. The system SHALL NOT grant the role CONSELHO_FISCAL any permission whose action is `create`, `update`, `delete` or `cancel`.
5. WHEN a user has more than one role THEN the effective permissions SHALL be the union of the permissions of those roles.
6. IF a permission stored in the database is no longer declared in code THEN the synchronization SHALL mark it inactive and SHALL NOT delete it.

**Independent Test**: Subir a API duas vezes e ver a matriz idêntica; rodar o teste de invariante do Conselho Fiscal.

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

**Independent Test**: Chamar cada rota sem sessão e ver 401; chamar com um papel sem a permissão e ver 403.

---

### P1: Registro de auditoria atômico

**User Story**: Como diretor e conselheiro fiscal, quero que toda alteração relevante fique registrada junto com a própria alteração, para confiar no histórico.

**Why P1**: É a base da prestação de contas (ADR-004).

**Acceptance Criteria**:
1. WHEN a use case that changes audited data commits THEN the system SHALL have written the audit entry in the same database transaction as the change.
2. IF writing the audit entry fails THEN the system SHALL roll back the whole transaction and respond 500 with code `audit_failed`.
3. The system SHALL record for every entry an id (UUID), `occurred_at` (UTC), `actor_user_id` (null only for system actions), `action`, `entity_type`, `entity_id`, `before` (JSON or null), `after` (JSON or null), an optional `reason` and the `request_id`.
4. IF an entry has an action of cancellation or extraordinary adjustment and no reason THEN the recorder SHALL reject it with an error.
5. IF `before` or `after` contains a key named `password`, `password_hash`, `token`, `session_token` or `csrf_token` THEN the recorder SHALL replace its value with `[redacted]`.
6. IF the recorder is called outside a database transaction THEN it SHALL return an error and SHALL NOT write the entry.

**Independent Test**: Forçar falha na gravação da auditoria e ver a alteração desfeita.

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
2. WHEN the filters `entity_type`, `entity_id`, `actor_user_id`, `from` or `to` are supplied THEN the system SHALL return only the entries matching all of them.
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
7. IF a request body exceeds 1 MiB THEN the system SHALL respond 413 with code `payload_too_large`.

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
- IF a user is created, or its roles are replaced, with an empty role set THEN the system SHALL respond 422 with code `validation_failed`.
- WHEN a role is removed from a user THEN the system SHALL stop granting the permissions of that role from the next request.
- WHEN `Allocate` receives a total of zero THEN the system SHALL return parts that are all zero.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
|---|---|---|---|
| IDN-01 | P1: Primeiro administrador | In Tasks | Pending |
| IDN-02 | P1: Login, sessão e logout | In Tasks | Pending |
| IDN-03 | P1: Proteção contra CSRF e origem | In Tasks | Pending |
| IDN-04 | P1: Gestão de usuários | In Tasks | Pending |
| IDN-05 | P1: Política de senha | In Tasks | Pending |
| RBAC-01 | P1: Modelo de permissões e papéis | In Tasks | Pending |
| RBAC-02 | P1: Autorização negada por padrão | In Tasks | Pending |
| AUD-01 | P1: Registro de auditoria atômico | In Tasks | Pending |
| AUD-02 | P1: Auditoria imutável | In Tasks | Pending |
| AUD-03 | P2: Consulta de auditoria | In Tasks | Pending |
| MNY-01 | P1: Tipo monetário em centavos | In Tasks | Implementing |
| MNY-02 | P1: Serialização e formatação de dinheiro | In Tasks | Implementing |
| MNY-03 | P2: Rateio e percentuais | In Tasks | Implementing |
| TST-01 | P1: Infraestrutura de testes de integração (Go) | In Tasks | Pending |
| TST-02 | P2: Verificação de fronteiras entre módulos | In Tasks | Pending |
| TST-03 | P2: Infraestrutura de testes do front | In Tasks | Pending |
| API-01 | P1: Contrato OpenAPI | In Tasks | Implementing |
| API-02 | P1: Convenções da API | In Tasks | Implementing |
| PLT-01 | P2: Configuração, logs e migrações | In Tasks | Implementing |

**Coverage:** 19 total, 19 mapped to tasks, 0 unmapped (mapeamento detalhado em `tasks.md`, seção Requirement Coverage).

---

## Success Criteria

- [ ] `go test -tags=integration ./...` e `pnpm test` passam no CI, junto com o lint, o build e as verificações de contrato OpenAPI.
- [ ] Nas rotas da fundação, um ADMIN cria um usuário, atribui papel e desativa, tudo com registro de auditoria consultável pela API.
- [ ] Senhas fora da política (10 para administrativos, 8 para associados, lista de comprometidas) são recusadas e a troca obrigatória bloqueia o resto da API.
- [ ] O teste de invariante confirma que CONSELHO_FISCAL não tem nenhuma permissão `create`, `update`, `delete` ou `cancel`.
- [ ] O papel de banco da aplicação não consegue fazer UPDATE, DELETE ou TRUNCATE em `audit_log`, provado por teste.
- [ ] Os mesmos vetores de dinheiro passam no Go e no TypeScript, e o teste de esquema barra colunas `float`, `double` e `numeric`.
