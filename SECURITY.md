# Política de segurança

A TJ Platform trata dados financeiros e dados pessoais de associados. Vulnerabilidades são levadas a sério.

## Como reportar

Não abra issue nem PR públicos com detalhes de uma vulnerabilidade.

Use o relato privado do GitHub: aba **Security** do repositório, opção **Report a vulnerability**. Inclua o que foi encontrado, como reproduzir e o impacto potencial.

Se a opção não estiver disponível, entre em contato com o mantenedor do repositório pelo perfil do GitHub [@antoniobt12062002](https://github.com/antoniobt12062002), sem publicar os detalhes.

## O que esperar

O objetivo é confirmar o recebimento em até 7 dias e informar o andamento da correção. Falhas confirmadas são corrigidas em release, com o crédito ao relator, se desejado.

## Versões suportadas

Apenas a versão mais recente publicada em `main` recebe correções de segurança.

## Regras do projeto

- O repositório é público: **nunca** versionar segredos, chaves, tokens, `.env` ou dados pessoais. Segredos ficam em GitHub Secrets por ambiente.
- Dados pessoais de associados seguem a LGPD (Lei 13.709/2018); o portal público não expõe dados pessoais nem documentos.
- Atualizações de dependências são acompanhadas pelo Dependabot. Verificações de vulnerabilidade no CI estão planejadas.
- Autenticação, sessões e autorização seguem o [ADR-005](docs/adr/005-autenticacao-e-rbac.md); dados financeiros seguem o [ADR-004](docs/adr/004-auditoria-e-imutabilidade-financeira.md).
