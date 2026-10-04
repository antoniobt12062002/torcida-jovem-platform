# ADR-010: Acesso do navegador à API por rewrite same-origin do Next.js

- **Date**: 2026-10-04
- **Status**: Accepted
- **Deciders**: @antoniobt12062002
- **Tags**: web, security, api

## Context and Problem Statement

A [ADR-005](005-autenticacao-e-rbac.md) definiu a sessão por cookie `httpOnly` com `SameSite=Lax` e proteção CSRF, mas deixou em aberto "o domínio do cookie e o uso de proxy pelo Next", por depender de uma hospedagem ainda não definida. O front Web V1 (`.specs/features/web/`) precisa dessa resposta para começar.

A API em Go não emite nenhum cabeçalho CORS. O middleware `Origin` só aceita escritas vindas das origens de `ALLOWED_ORIGINS`, e o cookie `tj_session` é `SameSite=Lax`. Na prática, o navegador só consegue usar a API se as chamadas forem de mesma origem (same-origin) em relação à página.

## Decision Drivers

- Não alterar o modelo de segurança já implementado e testado na API: sessão, `Origin`, CSRF e `SameSite=Lax`.
- Não exigir mudança no backend só para atender o front.
- Funcionar já no desenvolvimento local e continuar válido quando a hospedagem for escolhida.
- Não expor o token de sessão ao JavaScript.

## Considered Options

- (a) Rewrite do Next.js: `/api/v1/*` na origem do front é encaminhado para a API Go.
- (b) Proxy reverso externo (por exemplo, Caddy ou nginx) servindo front e API sob o mesmo domínio.
- (c) CORS na API, com cookie de sessão entre sites.

## Decision Outcome

Chosen option: **(a) rewrite same-origin do Next.js**. O navegador só fala com a origem do front, e o Next encaminha `/api/v1/*` para a API, cujo endereço vem da variável `API_URL`, já usada pelo `web/`.

Consequências diretas:

- O cookie `tj_session` é emitido na origem do front, como cookie host-only enquanto `COOKIE_DOMAIN` ficar vazio.
- `ALLOWED_ORIGINS` contém a origem do front, o que o `.env.example` já faz (`http://localhost:3000`).
- Escritas saem do navegador com o `Origin` do front e com o cabeçalho `X-CSRF-Token`, que vem de `GET /api/v1/auth/me`. É exatamente o modelo que a API valida.
- Nenhuma mudança na API, nos contratos OpenAPI nem no cookie.
- A proteção de rotas no front **não** usa `proxy.ts` (o antigo `middleware`). No Next 16, `proxy` guarda em buffer o corpo das requisições com limite padrão de 10 MB e, acima disso, **trunca em silêncio**, enquanto o upload de comprovantes aceita até 11 MiB. A proteção de rotas fica no layout autenticado (Client Component). A API continua sendo a autoridade de acesso.
- (b) continua possível depois: um proxy reverso na frente dos dois, com o mesmo caminho `/api/v1/*`, mantém o navegador same-origin sem mudar o front.

### Positive Consequences

- A segurança da API fica intocada; o front só consome o modelo existente.
- Um único endereço para o navegador; sem CORS e sem cookie entre sites.
- A troca futura para proxy externo não exige mudança de código no front.

### Negative Consequences

- O servidor Next passa a estar no caminho de todas as chamadas à API (latência extra de um salto e dependência do processo do Next).
- O encaminhamento de `Set-Cookie`, `Origin` e de corpos grandes (upload) pelo rewrite **não é detalhado na documentação** do Next. Por isso é verificado contra a API real como critério de aceite da fundação Web (`.specs/features/web/`, F1), e não presumido.

## Pros and Cons of the Options

### (a) Rewrite same-origin do Next.js ✅ Chosen

- ✅ Sem mudança no backend; funciona já no desenvolvimento
- ✅ Compatível com (b) no futuro
- ❌ Um salto a mais e dependência do processo do Next

### (b) Proxy reverso externo

- ✅ Sem salto extra pelo Next
- ❌ Depende da decisão de hospedagem, que ainda não existe; não resolve o desenvolvimento local

### (c) CORS com cookie entre sites

- ✅ Front e API totalmente independentes
- ❌ Exige mudar a API (CORS e preflight) e o cookie (`SameSite=None`), contrariando o `SameSite=Lax` confirmado na ADR-005
- ❌ Amplia a superfície de CSRF

## Links

- [ADR-005](005-autenticacao-e-rbac.md): resolve a pendência "uso de proxy pelo Next" registrada nas notas
- [ADR-002](002-stack-backend-e-frontend.md)
- `.specs/features/web/STATE.md` (decisão `WEB-D-001`)
- Documentação do Next.js 16 consultada no pacote instalado (`node_modules/next/dist/docs`): `rewrites` (destino externo) e `proxyClientMaxBodySize` (limite de 10 MB e truncamento silencioso), em 2026-10-04
