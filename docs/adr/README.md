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
| [008](008-versao-do-contrato-openapi.md) | Versão do contrato OpenAPI | Accepted |
| [009](009-modelo-de-identidade-e-papeis-iniciais.md) | Modelo de identidade, vínculo administrativo e papéis iniciais | Accepted |
| [010](010-acesso-do-navegador-a-api-por-rewrite-same-origin.md) | Acesso do navegador à API por rewrite same-origin do Next.js | Accepted |
| [011](011-mapa-de-modulos-do-tj-connect.md) | Mapa de módulos e dependências do TJ Connect | Proposed |
| [012](012-cobrancas-e-eventos-persistentes.md) | Módulo `cobrancas`, confirmação por evento persistente e conciliação | Proposed |
| [013](013-contas-de-servico-e-autoria-de-processos.md) | Contas de serviço e autoria de processos automatizados | Proposed |
| [014](014-worker-e-processamento-persistente.md) | `cmd/worker` e processamento persistente em PostgreSQL | Proposed |
| [015](015-superficies-web-e-rotas-publicas-do-tj-connect.md) | Superfícies web, rotas públicas e entrada de webhook do TJ Connect | Proposed |
| [016](016-permissoes-sensiveis-e-escopo-por-jogo.md) | Permissões sensíveis, papel `FISCAL` e autorização por jogo | Proposed |
| [017](017-evolucao-aditiva-de-financeiro-e-estoque.md) | Evolução aditiva de `financeiro` e `estoque` | Proposed |
| [018](018-invariantes-transacionais-de-quota-e-check-in.md) | Invariantes transacionais de quota, check-in e frequência | Proposed |

Novos ADRs seguem o formato MADR e a numeração sequencial (`NNN-titulo-em-kebab-case.md`). Decisões de projeto que valem para todas as features também entram como `AD-NNN` em [`.specs/STATE.md`](../../.specs/STATE.md).
