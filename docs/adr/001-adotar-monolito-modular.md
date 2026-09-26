# ADR-001: Adotar arquitetura de monólito modular

- **Date**: 2026-09-26
- **Status**: Accepted
- **Deciders**: @antoniobt12062002
- **Tags**: architecture

## Context and Problem Statement

A TJ Platform reúne módulos fortemente relacionados: uma venda baixa o estoque, gera receita e alimenta o histórico do associado (ver [01-VISAO-PRODUTO](../01-VISAO-PRODUTO.md)). A equipe é pequena e voluntária, e a instituição precisa de uma operação simples de manter. É preciso escolher a forma de organizar e implantar o sistema.

## Decision Drivers

- Consistência transacional entre módulos (venda, estoque e financeiro).
- Baixa complexidade operacional: poucos serviços para implantar, monitorar e pagar.
- Manutenção viável por uma equipe pequena.
- Possibilidade de evoluir fronteiras sem reescrever o sistema.

## Considered Options

- Monólito modular (um binário, módulos com fronteiras explícitas)
- Microsserviços
- Monólito em camadas, sem fronteiras entre módulos

## Decision Outcome

Chosen option: **"Monólito modular"**, porque atende à consistência transacional e à simplicidade operacional, mantendo fronteiras que permitem extrair serviços no futuro, se algum dia for necessário.

Os módulos são: `identity`, `financeiro`, `estoque`, `loja`, `associados`, `eventos`, `acesso`, `transparencia` e `comunicacao`, mais um núcleo compartilhado `platform` (configuração, HTTP, auditoria, autorização, dinheiro). Regras e dependências em [domain-boundaries](../architecture/domain-boundaries.md).

### Positive Consequences

- Operações que atravessam módulos rodam em uma única transação de banco.
- Um único artefato para implantar; menos custo e menos partes para falhar.
- Fronteiras explícitas reduzem o acoplamento acidental.

### Negative Consequences

- As fronteiras dependem de disciplina e de verificação automática (linter de dependências), pois o compilador sozinho não as impede.
- Escala e implantação são do sistema inteiro, não por módulo.
- Uma falha grave em um módulo pode afetar todo o processo.

## Pros and Cons of the Options

### Monólito modular ✅ Chosen

- ✅ Transações locais e simples entre módulos
- ✅ Operação e custo mínimos
- ❌ Exige disciplina para manter as fronteiras

### Microsserviços

- ✅ Implantação e escala independentes
- ❌ Consistência distribuída (sagas, eventos) para operações que hoje são uma transação
- ❌ Custo operacional desproporcional ao estágio da instituição

### Monólito em camadas sem fronteiras

- ✅ Início mais rápido
- ❌ Tende a acoplar tudo a tudo, dificultando a evolução e o teste isolado

## Links

- [Visão geral da arquitetura](../architecture/architecture-overview.md)
- [Fronteiras de domínio](../architecture/domain-boundaries.md)
