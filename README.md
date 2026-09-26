# TJ Platform

Plataforma de gestão institucional da Torcida Jovem do Campo Mourão Futsal.

## Sobre

O projeto nasce dentro de um processo de profissionalização da torcida, substituindo controles descentralizados (planilhas, conversas em aplicativos de mensagem, registros individuais) por um sistema integrado de gestão.

## Documentação

- [Como contribuir](./docs/CONTRIBUTING.md) — fluxo de branches, commits e releases
- A documentação de produto (contexto, visão, módulos e módulo financeiro) está em revisão pelo PO e será publicada em `docs/` após a aprovação.

## Desenvolvimento local

1. `cp .env.example .env`
2. `docker compose up -d db` e `docker compose --profile tools run --rm migrate`
3. API: `cd api && go run ./cmd/api` (lê `DATABASE_URL` e `PORT` do ambiente)
4. Front: `cd web && pnpm install && pnpm dev`
