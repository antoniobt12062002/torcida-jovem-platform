# Architecture Decision Records

Registro das decisões de arquitetura da TJ Platform. Cada ADR é imutável: uma decisão nova substitui a antiga por meio de um novo ADR (status "Superseded by"), sem editar o histórico.

| ADR | Título | Status |
|---|---|---|
| [001](001-adotar-monolito-modular.md) | Arquitetura de monólito modular | Accepted |
| [002](002-stack-backend-e-frontend.md) | Stack de backend, frontend e banco de dados | Accepted |
| [003](003-valores-monetarios-em-centavos.md) | Valores monetários em centavos inteiros | Accepted |
| [004](004-auditoria-e-imutabilidade-financeira.md) | Auditoria append-only e imutabilidade financeira | Accepted |
| [005](005-autenticacao-e-rbac.md) | Autenticação por sessão e RBAC por permissões | Accepted |
| [006](006-armazenamento-de-documentos.md) | Armazenamento de documentos em object storage S3 | Accepted |
| [007](007-fonte-de-verdade-do-estoque.md) | Estoque como fonte de verdade baseada em movimentações | Accepted |

Novos ADRs seguem o formato MADR e a numeração sequencial (`NNN-titulo-em-kebab-case.md`). Decisões de projeto que valem para todas as features também entram como `AD-NNN` em [`.specs/STATE.md`](../../.specs/STATE.md).
