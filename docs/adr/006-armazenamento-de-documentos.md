# ADR-006: Armazenamento de documentos em object storage compatível com S3

- **Date**: 2026-09-26
- **Status**: Proposed
- **Deciders**: @antoniobt12062002 (aprovação pendente)
- **Tags**: storage, security, finance

## Context and Problem Statement

O financeiro exige gestão documental: nota fiscal, recibo, comprovante PIX, contrato, orçamento e fotos, visíveis apenas para a administração e o Conselho Fiscal ([FIN-001](../finance/FIN-001-VISAO-GERAL.md), seção 15). São arquivos binários, potencialmente numerosos e sensíveis, que precisam sobreviver a troca de servidor e a deploys.

## Decision Drivers

- Durabilidade e backup independentes do servidor da aplicação.
- Acesso restrito, com trilha de quem consultou.
- Custo baixo e sem depender de um provedor específico.
- Não misturar arquivos binários com o banco transacional.

## Considered Options

- Object storage compatível com S3 (Cloudflare R2, Backblaze B2, AWS S3 ou similar)
- Arquivos em `bytea` no PostgreSQL
- Sistema de arquivos local do servidor

## Decision Outcome

Proposed option: **object storage compatível com S3**, com bucket privado e acesso por URL assinada de curta duração emitida pela API após checar a permissão.

- No banco ficam apenas os metadados: identificador, nome original, tipo, tamanho, hash SHA-256, lançamento associado, usuário e data de envio.
- O documento em si é imutável: substituir gera nova versão, nunca sobrescreve (alinhado ao [ADR-004](004-auditoria-e-imutabilidade-financeira.md)).
- Acesso ao documento gera registro de auditoria.
- A escolha do provedor fica para a decisão de hospedagem, e a API usa um cliente S3 genérico para não travar a escolha.

### Positive Consequences

- Durabilidade e custo adequados; migração de provedor sem mudar o código.
- Banco pequeno e com backup rápido.

### Negative Consequences

- Um serviço a mais para configurar e proteger (chaves, políticas do bucket).
- Consistência entre banco e bucket exige cuidado (upload antes do registro, limpeza de órfãos).
- Validação de tipo e tamanho de arquivo, e possível varredura de malware, ficam por conta da aplicação.

## Pros and Cons of the Options

### Object storage S3 ✅ Proposed

- ✅ Durável, barato, independente do servidor
- ❌ Mais uma peça de infraestrutura

### bytea no PostgreSQL

- ✅ Transacional junto com o dado
- ❌ Engorda o banco e os backups; ruim para arquivos grandes

### Sistema de arquivos local

- ✅ Nenhuma infraestrutura extra
- ❌ Perde-se em troca de servidor ou redeploy; não escala; backup manual

## Open Questions

- Provedor (depende da decisão de hospedagem).
- Limite de tamanho, tipos aceitos e política de retenção.
- Varredura de malware nos uploads.

## Links

- [FIN-001](../finance/FIN-001-VISAO-GERAL.md)
- [ADR-004](004-auditoria-e-imutabilidade-financeira.md)
