# ADR-003: Representar valores monetários em centavos inteiros

- **Date**: 2026-09-26
- **Status**: Accepted
- **Deciders**: @antoniobt12062002
- **Tags**: finance, data

## Context and Problem Statement

O módulo financeiro trata valor bruto, taxas, valor líquido, parcelamentos e rateios (ver [FIN-001](../finance/FIN-001-VISAO-GERAL.md)). Números de ponto flutuante acumulam erros de arredondamento e produzem diferenças de centavos que quebram a conferência de saldo.

## Decision Drivers

- Exatidão em somas, parcelas e rateios.
- Comparações e agregações simples e rápidas no banco.
- Contrato de API sem ambiguidade entre Go, PostgreSQL e TypeScript.

## Considered Options

- Inteiro em centavos (`BIGINT`)
- `NUMERIC(14,2)` no banco e tipo decimal nas aplicações
- `float`

## Decision Outcome

Chosen option: **inteiro em centavos**. `10000` representa R$ 100,00. A moeda é apenas o real (BRL). `float` é proibido para qualquer valor financeiro.

- No banco: `BIGINT`, com sufixo `_cents` no nome da coluna (ex.: `gross_amount_cents`).
- Na API: campos inteiros com o mesmo sufixo. A formatação em reais acontece só na borda (interface e relatórios).
- Percentuais e taxas usam inteiros (pontos-base ou milésimos), não `float`.

### Positive Consequences

- Somas e comparações exatas; o tipo é o mesmo em todas as camadas.
- Sem dependência de biblioteca decimal.

### Negative Consequences

- Rateios e parcelas exigem uma regra explícita de arredondamento e de distribuição do resto de centavos (a definir na spec do módulo financeiro).
- Valores em banco exigem conversão para exibição; leitura direta no SQL é menos amigável.
- Suporte a outra moeda no futuro exigirá uma coluna de moeda.

## Pros and Cons of the Options

### Inteiro em centavos ✅ Chosen

- ✅ Exato e simples em todas as camadas
- ❌ Regras de arredondamento em rateio ficam por conta da aplicação

### NUMERIC + decimal

- ✅ Legível no banco
- ❌ Exige biblioteca decimal em Go e cuidado com serialização para JavaScript

### float

- ✅ Nenhuma
- ❌ Erros de arredondamento inaceitáveis para dinheiro

## Links

- [FIN-001 — Visão Geral do Módulo Financeiro](../finance/FIN-001-VISAO-GERAL.md)
