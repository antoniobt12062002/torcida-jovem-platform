# ADR-009: Modelo de identidade, vínculo administrativo e papéis iniciais

- **Date**: 2026-09-26
- **Status**: Proposed
- **Deciders**: @antoniobt12062002
- **Tags**: security, identity, rbac

## Context and Problem Statement

A ADR-005 fixou sessão no servidor e RBAC por permissões, e listou seis papéis iniciais (`ADMIN`, `PRESIDENTE`, `DIRETOR`, `FINANCEIRO`, `CONSELHO_FISCAL`, `ASSOCIADO`). Antes de implementar identidade é preciso decidir como o usuário se relaciona com o associado, como se separa "por que alguém administra" de "o que pode fazer", como se promove um associado a administrativo sem abrir caminho para escalada de privilégio, e quais papéis existem de fato. A lista da ADR-005 não cobre o estoque, a loja e os eventos, e o `ADMIN` genérico mistura poder técnico com poder institucional.

## Decision Drivers

- O módulo dono do domínio guarda os seus vínculos (`docs/architecture/domain-boundaries.md`).
- Acesso administrativo é o maior risco de segurança e precisa de motivo, rastro e revogação limpa.
- Ninguém deve conceder mais poder do que possui.
- O Conselho Fiscal tem função institucional própria, que vai além de leitura.

## Considered Options

- Manter os seis papéis da ADR-005 e tratar acesso administrativo só como papel.
- Separar vínculo administrativo de papel, adotar oito papéis e a regra sem escalada.
- Motor de políticas (por exemplo Casbin) para expressar tudo em regras.

## Decision Outcome

Chosen option: **"Separar vínculo administrativo de papel, com oito papéis e a regra sem escalada"**, porque deixa o poder rastreável e limitado sem introduzir um motor de políticas.

- **Usuário × associado**: entidades separadas. O vínculo é `associados.associados.user_id`, no módulo dono do domínio; `identity.users` não tem coluna de associado e `identity` não importa `associados`.
- **Papéis iniciais**: `ASSOCIADO`, `PRESIDENTE`, `DIRETORIA`, `TESOURARIA`, `ESTOQUE_LOJA`, `EVENTOS`, `CONSELHO_FISCAL`, `ADMIN_SISTEMA`. Substituem a lista da ADR-005.
- **Permissões e papéis**: permissões são um catálogo versionado no código; papéis são agrupadores; um usuário pode ter vários papéis ativos; a permissão efetiva é a união; o padrão é negar.
- **Conselho Fiscal**: nunca recebe `create`, `update`, `delete` nem `cancel`; recebe permissões institucionais de prestação de contas, aprovação e parecer, além de leitura.
- **`AdminMembership`** ("por que possui acesso administrativo") é separado de papel ("o que pode fazer"): motivo obrigatório, quem concedeu, quando, encerramento, um vínculo ativo por usuário, histórico nunca apagado. Papel diferente de `ASSOCIADO` exige vínculo ativo.
- **Promoção**: caso de uso único, numa transação: cria o vínculo, atribui papéis, marca `must_change_password`, revoga as sessões do alvo e audita.
- **Sem escalada**: ninguém concede, retira ou altera papel com permissão que não possui, nem altera o próprio acesso.
- **Auditoria**: toda alteração de permissão (promoção, retirada, papéis, sincronização da matriz, tentativa negada) é auditada.
- **Dupla aprovação** para promoção fica como evolução futura, fora da V1.

### Positive Consequences

- Acesso administrativo com motivo e histórico; retirar o acesso não perde o rastro.
- Uma conta comprometida com poder técnico não cria contas com poder institucional.
- Fronteira de módulos preservada na ligação entre usuário e associado.

### Negative Consequences

- O `ADMIN_SISTEMA` não concede `PRESIDENTE` nem `CONSELHO_FISCAL`; a concessão fica concentrada no `PRESIDENTE`.
- Mais uma entidade (`AdminMembership`) e mais um caso de uso a manter.
- Sem dupla aprovação, uma única pessoa com `identity:admin:grant` promove.

## Pros and Cons of the Options

### Manter os seis papéis e tratar acesso administrativo só como papel

- ✅ Nenhuma entidade nova.
- ❌ Não registra por que alguém tem acesso; retirar o papel apaga o motivo.
- ❌ Não cobre estoque, loja e eventos.

### Separar vínculo administrativo de papel, com oito papéis e a regra sem escalada ✅ Chosen

- ✅ Rastro, motivo e revogação limpa.
- ✅ Limita o dano de uma conta comprometida.
- ❌ Concentra a concessão no `PRESIDENTE`.

### Motor de políticas

- ✅ Expressa regras arbitrárias.
- ❌ DSL e operação a mais para um mapa papel-permissão simples.

## Links

- Detalha e substitui a lista de papéis iniciais de [ADR-005](005-autenticacao-e-rbac.md).
- Spec: `.specs/features/fundacao-core/spec.md` (IDN-06, RBAC-01, RBAC-03) e `design.md` (seção Modelo de identidade).
