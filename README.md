# TJ Platform

Plataforma de gestão institucional da Torcida Jovem do Campo Mourão Futsal.

## Sobre

O projeto nasce dentro de um processo de profissionalização da torcida, substituindo controles descentralizados (planilhas, conversas em aplicativos de mensagem, registros individuais) por um sistema integrado de gestão.

## Documentação

- [Como contribuir](./docs/CONTRIBUTING.md) — fluxo de branches, commits e releases
- [Visão geral da arquitetura](./docs/architecture/architecture-overview.md) e [fronteiras de domínio](./docs/architecture/domain-boundaries.md)
- [Decisões de arquitetura (ADRs)](./docs/adr/README.md)
- [Especificações (`.specs/`)](./.specs/README.md) — fluxo Spec Driven
- [Ambiente de IA](./docs/development/ai-environment.md) — skills usadas
- [Ferramentas e versões](./docs/development/tooling.md) — versões adotadas e fontes consultadas
- [Segurança](./SECURITY.md) e [licença](./LICENSE) (todos os direitos reservados)
- [00-CONTEXTO-PROJETO.md](./docs/00-CONTEXTO-PROJETO.md) — contexto institucional e objetivos
- [01-VISAO-PRODUTO.md](./docs/01-VISAO-PRODUTO.md) — visão do produto e perfis de usuário
- [02-MODULOS-SISTEMA.md](./docs/02-MODULOS-SISTEMA.md) — módulos do sistema
- [financeiro/FIN-001-VISAO-GERAL.md](./docs/finance/FIN-001-VISAO-GERAL.md) — especificação do módulo financeiro

## Desenvolvimento local

1. `cp .env.example .env`
2. `docker compose up -d db` e `docker compose --profile tools run --rm migrate` (a API usa o papel `tj_app`; as migrações, o `tj_owner`)
3. API: `cd api && go run ./cmd/api` (lê `DATABASE_URL` e `PORT` do ambiente)
4. Front: `cd web && pnpm install && pnpm dev`
