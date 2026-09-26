# ADR-004: Auditoria append-only e imutabilidade dos dados financeiros

- **Date**: 2026-09-26
- **Status**: Accepted
- **Deciders**: @antoniobt12062002
- **Tags**: finance, security, audit

## Context and Problem Statement

A TJ Platform precisa de transparência, rastreabilidade e prestação de contas anual ([FIN-001](../finance/FIN-001-VISAO-GERAL.md), seções 16 a 18). Depois que o ano é aprovado, o passado não pode ser alterado; erros são corrigidos por ajuste extraordinário. Isso exige registro confiável de quem fez o quê, e que dados financeiros nunca sumam.

## Decision Drivers

- Rastreabilidade completa: criação, alteração, responsável, data/hora, valor anterior e valor novo.
- Impossibilidade de apagar ou reescrever o histórico, inclusive por falha de código.
- Auditoria consistente com o dado: sem "alteração sem registro".

## Considered Options

- Tabela de auditoria append-only, gravada na mesma transação da alteração
- Triggers no banco gerando o histórico
- Event sourcing completo
- Apenas colunas `updated_at` e `updated_by`

## Decision Outcome

Chosen option: **tabela de auditoria append-only, gravada pela aplicação na mesma transação da alteração**, mais as regras abaixo.

1. Cada criação ou alteração de dado financeiro grava um registro de auditoria com entidade, identificador, ação, usuário, data/hora, valor anterior e valor novo.
2. O papel de banco usado pela aplicação **não tem permissão de `UPDATE` nem `DELETE`** na tabela de auditoria.
3. **Não existe exclusão definitiva** de informação financeira. Usa-se cancelamento (com motivo e responsável), desativação de categorias e **ajuste extraordinário** após o fechamento do ano.
4. O fechamento de ano bloqueia alterações retroativas.

### Positive Consequences

- Histórico confiável, verificável e consistente com o dado.
- Suporta a prestação de contas e o parecer do Conselho Fiscal.

### Negative Consequences

- Toda escrita financeira passa por um caminho comum de auditoria; quem esquecer perde a garantia (mitigar com testes e revisão).
- Volume de dados de auditoria cresce continuamente; será preciso definir retenção e particionamento.
- Consultas de "estado em uma data" ficam mais trabalhosas que em event sourcing.

## Pros and Cons of the Options

### Append-only na aplicação ✅ Chosen

- ✅ Simples, na mesma transação, com permissões de banco como barreira
- ❌ Depende de disciplina no código

### Triggers no banco

- ✅ Captura qualquer alteração, mesmo fora da aplicação
- ❌ Perde o contexto de negócio (motivo, usuário da aplicação) e é mais difícil de testar

### Event sourcing completo

- ✅ Histórico perfeito por construção
- ❌ Complexidade desproporcional ao estágio atual

### Colunas updated_at / updated_by

- ✅ Trivial
- ❌ Não guarda valor anterior nem histórico; inaceitável para prestação de contas

## Links

- [ADR-003](003-valores-monetarios-em-centavos.md)
- [FIN-001](../finance/FIN-001-VISAO-GERAL.md)
