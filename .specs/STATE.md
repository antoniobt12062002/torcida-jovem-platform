# STATE

## Decisions

### AD-001
- **Decision**: A plataforma é um monólito modular em Go, com módulos `identity`, `financeiro`, `estoque`, `loja`, `associados`, `eventos`, `acesso`, `transparencia` e `comunicacao` sobre um núcleo `platform`.
- **Reason**: Consistência transacional entre venda, estoque e financeiro, e baixa complexidade operacional para uma equipe pequena.
- **Trade-off**: Fronteiras dependem de verificação automática; implantação e escala são do sistema inteiro.
- **Scope**: Todo o backend. Ver `docs/adr/001-adotar-monolito-modular.md` e `docs/architecture/domain-boundaries.md`.
- **Date**: 2026-09-26
- **Status**: active

### AD-002
- **Decision**: Stack fixa em Go + Gin + GORM, PostgreSQL 16, `golang-migrate` (sem `AutoMigrate`), Next.js + TypeScript + shadcn/ui + pnpm, REST com OpenAPI.
- **Reason**: Implantação simples, tipagem forte e transações explícitas; alternativas avaliadas no ADR.
- **Trade-off**: Menos recursos prontos que frameworks maiores; GORM pode exigir SQL explícito em relatórios.
- **Scope**: Todo o repositório. Mudança estrutural de schema só por migration revisada em PR. Ver `docs/adr/002-stack-backend-e-frontend.md`.
- **Date**: 2026-09-26
- **Status**: active

### AD-003
- **Decision**: Valores monetários são inteiros em centavos (`BIGINT`, sufixo `_cents`); `float` é proibido para dinheiro.
- **Reason**: Exatidão em somas, parcelas e rateios.
- **Trade-off**: Regra explícita de arredondamento e de resto de centavos em rateios.
- **Scope**: Banco, API e frontend. Ver `docs/adr/003-valores-monetarios-em-centavos.md`.
- **Date**: 2026-09-26
- **Status**: active

### AD-004
- **Decision**: Auditoria append-only na mesma transação da alteração; sem exclusão definitiva de dados financeiros (cancelamento, desativação e ajuste extraordinário).
- **Reason**: Rastreabilidade e prestação de contas anual com ano bloqueado.
- **Trade-off**: Todo caminho de escrita financeira precisa passar pela auditoria; volume de dados cresce.
- **Scope**: Módulos financeiros e qualquer dado sujeito à prestação de contas. Ver `docs/adr/004-auditoria-e-imutabilidade-financeira.md`.
- **Date**: 2026-09-26
- **Status**: active

### AD-005
- **Decision**: Autenticação por sessão no servidor com cookie `httpOnly` e proteção CSRF, sem JWT; autorização RBAC baseada em permissões; papéis iniciais `ADMIN`, `PRESIDENTE`, `DIRETOR`, `FINANCEIRO`, `CONSELHO_FISCAL`, `ASSOCIADO`.
- **Reason**: Revogação imediata, menor superfície de ataque, um único cliente próprio.
- **Trade-off**: Estado de sessão no servidor; outros tipos de cliente exigirão nova decisão.
- **Scope**: Toda a API e o frontend. O associado autenticado é distinto do visitante do portal público. Ver `docs/adr/005-autenticacao-e-rbac.md`.
- **Date**: 2026-09-26
- **Status**: active

### AD-006
- **Decision**: `.specs/` é a fonte oficial das specs (fluxo Specify, Design, Tasks, Implement, Validate); `docs/` guarda documentação de produto, arquitetura (`docs/architecture/`), ADRs (`docs/adr/`) e histórico (`docs/archive/`). Skills de IA não são versionadas; ficam documentadas em `docs/development/ai-environment.md`.
- **Reason**: Separar especificação executável de documentação de referência e manter o repositório independente de ferramentas locais.
- **Trade-off**: Os docs de produto existentes precisam ser convertidos gradualmente para o formato de spec.
- **Scope**: Organização do repositório e do processo de desenvolvimento.
- **Date**: 2026-09-26
- **Status**: active

### AD-007
- **Decision**: O repositório é público, mas o código é proprietário (todos os direitos reservados), sem licença open source.
- **Reason**: Transparência institucional sem abrir o código para uso e redistribuição.
- **Trade-off**: Terceiros não podem reutilizar o código; contribuições externas exigem acordo prévio.
- **Scope**: Todo o repositório. Nunca versionar segredos ou dados pessoais. Ver `LICENSE` e `SECURITY.md`.
- **Date**: 2026-09-26
- **Status**: active

## Handoff

- **Feature**: governança (sem feature em `.specs/features/`)
- **Phase / Task**: Governança concluída, aguardando aprovação do mantenedor
- **Completed**: CLAUDE.md, `.specs/`, `docs/architecture/`, ADR-001 a ADR-007, SECURITY.md, LICENSE, `docs/development/ai-environment.md`, reorganização de `docs/`
- **In-progress** (file:line): none
- **Next step**: Após a aprovação, definir a spec de fundação (auth/RBAC, auditoria, dinheiro, testes com banco real, OpenAPI) e resolver as decisões pendentes listadas na entrega
- **Blockers**: Aprovação dos ADR-006 e ADR-007 (Proposed) e decisões em aberto
- **Uncommitted files**: none
- **Branch**: feature/governanca
