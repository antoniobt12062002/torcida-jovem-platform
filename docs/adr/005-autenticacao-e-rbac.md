# ADR-005: Autenticação por sessão com cookie seguro e RBAC por permissões

- **Date**: 2026-09-26
- **Status**: Accepted
- **Deciders**: @antoniobt12062002
- **Tags**: security, identity

## Context and Problem Statement

O sistema tem usuários com perfis muito diferentes: administração, diretoria, financeiro, Conselho Fiscal e associados com login próprio. O associado autenticado é diferente do visitante do portal público (área aberta e separada). Há dados financeiros e pessoais sensíveis, num repositório público, o que exige uma autenticação simples de acertar e difícil de usar errado.

## Decision Drivers

- Segurança por padrão, com poucas partes móveis.
- Um único cliente próprio (o front Next.js); não há clientes de terceiros nem mobile nativo hoje.
- Revogação imediata de acesso (desligamento de diretor, conta comprometida).
- Autorização que evolua sem reescrever regras.

## Considered Options

- Sessão no servidor com cookie `httpOnly`
- JWT (access e refresh token)
- Serviço de identidade externo (provedor OIDC)

## Decision Outcome

Chosen option: **sessão no servidor com cookie `httpOnly`**, sem JWT neste momento.

- Sessões armazenadas no PostgreSQL; o cookie carrega apenas um identificador aleatório e opaco.
- Cookie com `HttpOnly`, `Secure` e `SameSite`; sessão rotacionada no login e revogável.
- Proteção contra CSRF em requisições que alteram estado (validação de `Origin` e token anti-CSRF).
- Senhas com `argon2id`.
- **RBAC baseado em permissões, não no nome do papel:** papéis agrupam permissões, e o código verifica a permissão (ex.: `financeiro:lancamento:criar`).
- Papéis iniciais: `ADMIN`, `PRESIDENTE`, `DIRETOR`, `FINANCEIRO`, `CONSELHO_FISCAL`, `ASSOCIADO`. A matriz de permissões será definida na spec de identidade.

### Positive Consequences

- Revogação imediata e segredo nunca exposto ao JavaScript do navegador.
- Superfície menor que a de JWT (sem gestão de refresh, rotação de chaves ou lista de revogação).
- Novos papéis e permissões sem alterar o código dos módulos.

### Negative Consequences

- O servidor guarda estado de sessão (uma consulta por requisição, mitigável com cache curto).
- Clientes que não sejam o front próprio (app nativo, integrações) exigirão nova decisão.
- Exige tratamento correto de CSRF e de configuração de cookies entre domínios de staging e produção.

## Pros and Cons of the Options

### Sessão com cookie httpOnly ✅ Chosen

- ✅ Simples, revogável e resistente a roubo por XSS
- ❌ Estado no servidor; CSRF precisa ser tratado

### JWT

- ✅ Sem estado, útil para muitos clientes ou serviços
- ❌ Revogação difícil e mais armadilhas de segurança para o nosso caso

### Provedor externo (OIDC)

- ✅ Reduz o código de autenticação e oferece MFA pronto
- ❌ Dependência e custo externos; decisão adiada, reavaliável por um novo ADR

## Links

- [ADR-004](004-auditoria-e-imutabilidade-financeira.md)
- [01-VISAO-PRODUTO](../01-VISAO-PRODUTO.md)
