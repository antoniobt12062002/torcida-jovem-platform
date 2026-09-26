# ADR-008: Versão do contrato OpenAPI

- **Date**: 2026-09-26
- **Status**: Accepted
- **Deciders**: @antoniobt12062002
- **Tags**: api, contract

## Context and Problem Statement

O contrato OpenAPI é a fonte de verdade da comunicação entre a API em Go e o front em Next.js: dele saem as interfaces do servidor, os modelos, os tipos TypeScript e a validação das respostas nos testes. Toda a cadeia depende de uma única versão do formato, e trocá-la depois obriga a revisar a configuração de todas as ferramentas. É preciso fixar a versão antes de escrever o primeiro contrato.

## Decision Drivers

- Suporte comprovado nas quatro ferramentas da cadeia: `oapi-codegen`, `kin-openapi`, `openapi-typescript` e o linter Redocly.
- Risco baixo: o contrato é a base de todos os módulos.
- Poder migrar para uma versão mais nova sem reescrever os contratos.

## Considered Options

- OpenAPI 3.0.3
- OpenAPI 3.1.x
- Swagger 2.0 (OpenAPI 2.0)

## Decision Outcome

Chosen option: **OpenAPI 3.0.3**, porque é a versão coberta por todas as ferramentas escolhidas e suficiente para o que a plataforma precisa hoje (esquemas, parâmetros, segurança por cookie, respostas `application/problem+json`, paginação por cursor).

Os contratos ficam em `api/openapi/`, um por módulo, mais `common.yaml` com os componentes compartilhados. O OpenAPI descreve **contratos de comunicação** e não contém regras de negócio, que permanecem nas specs e nos módulos de domínio.

### Positive Consequences

- Cadeia de geração e validação com o menor risco: nenhuma ferramenta depende de recursos novos.
- Contratos simples de migrar: evitar recursos exclusivos da 3.0 fora do necessário (por exemplo, usar `nullable: true` só onde houver valor nulo de fato).

### Negative Consequences

- Sem recursos exclusivos da 3.1 (tipos como lista, alinhamento total com JSON Schema, `webhooks`).
- Uma migração futura exigirá revisar os contratos e regerar o código.

### Migração futura

A 3.1 é a evolução natural e pode ser adotada quando houver necessidade concreta e as ferramentas estiverem maduras nesse uso. Ao migrar: novo ADR que substitui este, ajuste de `nullable` para tipos em lista, regeração de todo o código e reexecução dos testes de contrato. A migração é feita por módulo ou de uma vez, conforme o tamanho dos contratos na época.

## Pros and Cons of the Options

### OpenAPI 3.0.3 ✅ Chosen

- ✅ Suporte estabelecido em todas as ferramentas
- ✅ Suficiente para as necessidades atuais
- ❌ Sem recursos exclusivos da 3.1

### OpenAPI 3.1.x

- ✅ Alinhada ao JSON Schema; mais expressiva
- ✅ As ferramentas já anunciam suporte (`oapi-codegen` v2.8.0 e `openapi-typescript` 7.x citam 3.1; `kin-openapi` v0.149.0 a lista como alvo)
- ❌ A maturidade desse suporte nas nossas ferramentas não foi verificada; risco maior para a base de todos os módulos

### Swagger 2.0

- ✅ Nenhuma
- ❌ Formato antigo; sem suporte no `oapi-codegen` nem no `openapi-typescript` atuais

## Links

- [ADR-002](002-stack-backend-e-frontend.md)
- [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen), [kin-openapi](https://github.com/getkin/kin-openapi) e [openapi-typescript](https://openapi-ts.dev/introduction), consultados em 2026-09-26
