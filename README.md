# TJ Platform

Plataforma de gestão institucional da Torcida Jovem do Campo Mourão Futsal.

## Sobre

O projeto nasce dentro de um processo de profissionalização da torcida, substituindo controles descentralizados (planilhas, conversas em aplicativos de mensagem, registros individuais) por um sistema integrado de gestão.

## Documentação

- [Como contribuir](./docs/CONTRIBUTING.md) — fluxo de branches, commits e releases
- [00-CONTEXTO-PROJETO.md](./docs/00-CONTEXTO-PROJETO.md) — contexto institucional e objetivos
- [01-VISAO-PRODUTO.md](./docs/01-VISAO-PRODUTO.md) — visão do produto e perfis de usuário
- [02-MODULOS-SISTEMA.md](./docs/02-MODULOS-SISTEMA.md) — módulos do sistema
- [financeiro/FIN-001-VISAO-GERAL.md](./docs/financeiro/FIN-001-VISAO-GERAL.md) — especificação do módulo financeiro

## Desenvolvimento local

1. `cp .env.example .env`
2. `docker compose up -d db` e `docker compose --profile tools run --rm migrate`
3. API: `cd api && go run ./cmd/api` (lê `DATABASE_URL` e `PORT` do ambiente)
4. Front: `cd web && pnpm install && pnpm dev`
