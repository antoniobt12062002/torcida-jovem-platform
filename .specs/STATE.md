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
- **Decision**: Autenticação por sessão no servidor com cookie `httpOnly` e proteção CSRF, sem JWT; autorização RBAC baseada em permissões; papéis iniciais definidos em AD-011 (substituem a lista original `ADMIN`, `PRESIDENTE`, `DIRETOR`, `FINANCEIRO`, `CONSELHO_FISCAL`, `ASSOCIADO`).
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

### AD-008
- **Decision**: Documentos financeiros ficam em object storage compatível com S3, com bucket privado, URL assinada e metadados no banco; nenhum provedor específico é definido.
- **Reason**: Durabilidade e custo adequados sem misturar binários ao banco transacional nem travar o provedor.
- **Trade-off**: Um serviço a mais para proteger; consistência entre banco e bucket fica na aplicação.
- **Scope**: Módulos que anexam documentos (financeiro em especial). Ver `docs/adr/006-armazenamento-de-documentos.md`.
- **Date**: 2026-09-26
- **Status**: active

### AD-009
- **Decision**: O módulo `estoque` é a fonte de verdade das quantidades, baseada em razão de movimentações; o saldo é calculado pelos movimentos; estoque negativo é bloqueado por padrão e só admite exceção via ajuste autorizado e auditado.
- **Reason**: Uma única fonte de verdade rastreável e consistente com venda e compra.
- **Trade-off**: Modelagem mais elaborada e controle de concorrência no banco.
- **Scope**: `estoque`, `loja`, `eventos` e o fluxo de compra do `financeiro`. Ver `docs/adr/007-fonte-de-verdade-do-estoque.md`.
- **Date**: 2026-09-26
- **Status**: active

### AD-010
- **Decision**: O contrato OpenAPI 3.0.3 é a fonte de verdade da comunicação, em fluxo contract-first (contrato, lint, geração, implementação, validação). Há um contrato por módulo em `api/openapi/` mais `common.yaml`, com código gerado por módulo. Ferramentas: `oapi-codegen`, `kin-openapi`, `openapi-typescript` e Redocly; `openapi-fetch` só na primeira tela real. O contrato descreve comunicação e não contém regras de negócio, que ficam nas specs e nos módulos.
- **Reason**: Uma única fonte de verdade entre servidor, cliente e testes, sem arquivo gigante e respeitando as fronteiras de módulo.
- **Trade-off**: Código gerado versionado e um job de CI extra; migração para a 3.1 exige revisar os contratos.
- **Scope**: Toda a API e o front. Ver `docs/adr/008-versao-do-contrato-openapi.md`.
- **Date**: 2026-09-26
- **Status**: active

### AD-011
- **Decision**: Usuário e associado são entidades separadas e o vínculo fica em `associados.associados.user_id`. Papéis iniciais: `ASSOCIADO`, `PRESIDENTE`, `DIRETORIA`, `TESOURARIA`, `ESTOQUE_LOJA`, `EVENTOS`, `CONSELHO_FISCAL`, `ADMIN_SISTEMA`. O acesso administrativo é um vínculo (`AdminMembership`, "por que") separado do papel ("o que pode fazer"); ninguém concede permissão que não possui nem altera o próprio acesso; toda alteração de permissão é auditada. Dupla aprovação fica como evolução futura.
- **Reason**: Rastreabilidade do poder administrativo e menor dano de uma conta comprometida, sem quebrar a fronteira entre módulos.
- **Trade-off**: A autorização decide só pelas permissões efetivas, sem regra por nome de papel; com a matriz provisória a concessão institucional fica no PRESIDENTE; mais uma entidade e um caso de uso a manter.
- **Scope**: `identity`, `associados` e todo módulo que declara permissões. Ver `docs/adr/009-modelo-de-identidade-e-papeis-iniciais.md`.
- **Date**: 2026-09-26
- **Status**: active

### AD-012
- **Decision**: O contrato OpenAPI (`api/openapi/*.yaml`) é a fonte única das rotas. A partida da API falha se uma rota registrada não tem operação no contrato, ou se uma operação não tem rota; o CI falha se o código gerado ou os tipos do front divergem do contrato.
- **Reason**: Evita rotas esquecidas fora do contrato (sem validação, sem documentação e sem tipos) e contrato que promete o que a API não faz.
- **Trade-off**: Toda rota nova começa pelo contrato, e a geração de código entra no fluxo de trabalho.
- **Scope**: `api/openapi/`, `api/internal/httpapi/router.go` e o job de contrato do CI. Ver `docs/architecture/architecture-overview.md`.
- **Date**: 2026-09-27
- **Status**: active

### AD-013
- **Decision**: Rotas são negadas por padrão: só a lista pública explícita (`/healthz`, login e as duas rotas de recuperação de acesso) dispensa sessão. A cadeia pública aplica origem, limite de corpo de 64 KiB e validação do contrato; a autenticada aplica limite de 1 MiB, origem, sessão, CSRF e validação do contrato, nessa ordem, para que quem não entrou não aprenda a estrutura da API.
- **Reason**: Um endpoint esquecido sem proteção é o erro mais provável e mais caro; a ordem da cadeia decide o que cada tipo de chamador consegue observar.
- **Trade-off**: Toda rota pública nova exige decisão explícita e teste; a autenticação vem antes da validação de estrutura.
- **Scope**: `api/internal/httpapi/` e `api/internal/platform/httpx/`. Ver AD-005 e `docs/adr/005-autenticacao-e-rbac.md`.
- **Date**: 2026-09-27
- **Status**: active

### AD-014
- **Decision**: Cada módulo é composto em um único ponto de entrada (`identity.New`), que liga repositórios e casos de uso; os handlers HTTP são finos (leem o principal, chamam o caso de uso, devolvem a resposta) e os erros de domínio viram HTTP numa tabela única. As fronteiras entre módulos e entre camadas (`domain` → `app` → `infra` e `http`) são verificadas por `api/internal/architecture_test.go`, que descobre os módulos pelos diretórios de `internal/`.
- **Reason**: Regra de negócio fica num só lugar, o mapeamento de erros não diverge entre rotas e um módulo novo herda as fronteiras sem configuração.
- **Trade-off**: Handlers dependem do módulo composto; quem adiciona um módulo segue o mesmo esqueleto de pacotes (`domain`, `app`, `infra`, `http`).
- **Scope**: Todos os módulos de negócio. Ver `docs/architecture/domain-boundaries.md`.
- **Date**: 2026-09-27
- **Status**: active

## Handoff

- **Feature**: `.specs/features/fundacao-core/`
- **Phase / Task**: Fases 1 a 11 mescladas (T1 a T82); fase 12 (guardrails e documentação, T83 a T86) implementada na branch `feature/fundacao-core-fase-12`, aguardando PR e merge
- **Completed**: T1 a T86 (fundacao-core)
- **In-progress** (file:line): none
- **Next step**: Mesclar o PR da fase 12, rodar a validação final da `fundacao-core` (Verifier independente e `validation.md`), analisar o SEC-001 e apresentar a `fundacao-documentos`
- **Blockers**: none
- **Future decisions**: ErrOutOfRange → 422 `amount_out_of_range` no primeiro endpoint monetário; adicionar o componente `Id` ao `common.yaml` quando o primeiro contrato precisar; dupla aprovação para promoção administrativa; separar auditoria institucional e técnica (`audit:log:read` do Conselho Fiscal aprovado); convite por e-mail, primeiro acesso por link temporário e fluxo de onboarding; `GET /roles` (papéis, descrições e permissões) quando houver interface administrativa; troca de e-mail (`ChangeEmail`, com senha atual, auditoria e possível confirmação por e-mail); rate limit unificado de credenciais, por IP e dispositivo, e MFA na recuperação; impedir `EMAIL_PROVIDER=disabled` em produção; provedor concreto de e-mail (adaptador atrás de `platform/email`); armazenamento controlado de IP em eventos de segurança (LGPD); versionamento dos eventos de auditoria; encadeamento criptográfico dos registros
- **Uncommitted files**: none
- **Branch**: feature/fundacao-core-fase-12
