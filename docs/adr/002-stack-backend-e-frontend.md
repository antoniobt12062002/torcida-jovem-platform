# ADR-002: Stack de backend, frontend e banco de dados

- **Date**: 2026-09-26
- **Status**: Accepted
- **Deciders**: @antoniobt12062002
- **Tags**: architecture, stack, database

## Context and Problem Statement

O scaffold inicial já usava Go e Next.js, mas isso não basta como justificativa: uma stack escolhida por inércia é cara de trocar depois. Esta decisão avalia as alternativas de forma independente do scaffold e registra as limitações conhecidas.

O sistema é um ERP institucional: forte no relacional e transacional (financeiro, estoque, auditoria), com painel administrativo para a diretoria, área autenticada do associado e portal público, todos precisando funcionar bem no celular.

## Decision Drivers

- Correção transacional e integridade dos dados financeiros.
- Simplicidade e custo de implantação (artefato pequeno, poucas dependências de runtime).
- Tipagem forte e testabilidade.
- Ecossistema maduro para PostgreSQL, autenticação e geração de PDF/planilhas.
- Manutenibilidade por uma equipe pequena.

## Considered Options

- Backend: Go + Gin + GORM
- Backend: Node/TypeScript full-stack (Next.js ou NestJS)
- Backend: Java/Kotlin com Spring Boot
- Backend: Python com Django ou FastAPI
- Frontend: Next.js + shadcn/ui, contra SPA em Vite ou outros frameworks (Remix, SvelteKit)

## Decision Outcome

Chosen option: **Go + Gin + GORM no backend, PostgreSQL 16 como banco, Next.js (App Router) + TypeScript + shadcn/ui + pnpm no frontend**, com Docker para implantação.

- **Migrações:** `golang-migrate`, com SQL versionado e revisado. `AutoMigrate` do GORM é proibido: mudança estrutural só entra por migration em PR.
- **Contrato:** REST descrito em OpenAPI, com cliente TypeScript gerado.

### Positive Consequences

- Binário estático e imagem pequena (a imagem do scaffold tem cerca de 63 MB), simples de implantar.
- Tipagem forte e transações explícitas, adequadas ao financeiro.
- Next.js atende painel autenticado e portal público com o mesmo projeto; shadcn/ui dá componentes acessíveis e responsivos sem dependência de runtime pesada.

### Negative Consequences

- Menos "baterias incluídas" que Django ou Rails (painel administrativo, autenticação e formulários são construídos aqui).
- O GORM esconde SQL e pode causar consultas ineficientes (N+1) se mal usado; relatórios financeiros complexos tendem a exigir SQL explícito.
- Duas linguagens (Go e TypeScript) no mesmo repositório.

### Limitações e mitigação

- Se o GORM atrapalhar consultas de relatório ou controle de transação, os repositórios de um módulo podem migrar para `sqlc`/`pgx` sem afetar os demais, porque o acesso a dados fica atrás de interfaces por módulo.
- A escolha de Gin e GORM é reversível por módulo; a de Go e PostgreSQL não é, na prática.

## Pros and Cons of the Options

### Go + Gin + GORM ✅ Chosen

- ✅ Implantação simples, tipagem forte, bom suporte a PostgreSQL
- ❌ Mais código para o que frameworks maiores entregam prontos

### Node/TypeScript full-stack

- ✅ Uma linguagem só, compartilhamento de tipos
- ❌ Regras financeiras no mesmo runtime do front misturam responsabilidades; tipagem em runtime exige validação extra

### Java/Kotlin com Spring Boot

- ✅ Ecossistema corporativo maduro, ótimo para regras complexas
- ❌ Mais pesado de operar e hospedar para o porte da instituição

### Python com Django ou FastAPI

- ✅ Painel administrativo pronto (Django), desenvolvimento rápido
- ❌ Tipagem e desempenho menos previsíveis; implantação com mais partes móveis

## Links

- [ADR-001](001-adotar-monolito-modular.md)
- [Visão geral da arquitetura](../architecture/architecture-overview.md)
