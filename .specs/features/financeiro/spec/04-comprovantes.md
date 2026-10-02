# Comprovantes — Especificação

Status: Aprovada em 2026-09-29. Sub-spec 4 de 5 da feature `financeiro`. Decisões compartilhadas em `.specs/features/financeiro/STATE.md` (`FIN-D-NNN`). Narrativa em português; critérios de aceite em inglês (EARS).

## Problem Statement

`financeiro` precisa anexar comprovantes (nota fiscal, recibo, comprovante PIX, contrato, orçamento, fotos — `FIN-001` §15) a um lançamento, reaproveitando inteiramente `platform/documents` (`fundacao-documentos`, já concluída e validada). Esta spec não guarda nenhum dado novo — só documenta a convenção de uso.

## Goals

- Anexar um comprovante a um lançamento.
- Consultar/baixar um comprovante já anexado.
- Listar os comprovantes de um lançamento.

## Out of Scope

| Item | Razão |
|---|---|
| Qualquer regra de armazenamento, versionamento, validação de extensão/conteúdo/tamanho | Já resolvido por `platform/documents` (`fundacao-documentos`) — esta spec não reimplementa nada disso |
| Alterar `status` do lançamento por causa de um comprovante existir | `FIN-D-005` — "com comprovante" é condição derivada, não estado |
| Excluir comprovante | `platform/documents` não expõe essa operação por desenho (`AD-006`/`AD-015`) — não há nada a especificar aqui |
| Obrigatoriedade de comprovante em qualquer fluxo | Nunca obrigatório no V1 (sem alertas, `financeiro/STATE.md`, seção "Decisões futuras") |
| Versionamento/substituição de comprovante (encadeamento via `Supersedes`) | `FIN-D-016` — decidido: comprovantes são independentes e não encadeados no V1; substituição/versionamento fica para uma evolução futura específica, com nova decisão e novos requisitos |

## Assumptions & Open Questions

| Question | Chosen default | Rationale | Confirmed |
|---|---|---|---|
| `owner_type` | `financeiro.lancamento` | Já usado como exemplo ilustrativo em toda a spec/testes de `fundacao-documentos`; bate com o formato exigido (`[a-z0-9_]+\.[a-z0-9_]+`) | y |
| `owner_id` | `id` do `LancamentoFinanceiro` (dono: `02-lancamentos`) | Direto | y |
| Obrigatoriedade | Nunca obrigatório — pode ser anexado a qualquer momento, mesmo depois do lançamento cancelado | `FIN-001` §25 (alertas de "despesa sem comprovante") pressupõe que existir sem um é possível; alertas ficam fora do V1 | y |
| Efeito no cancelamento | Nenhum — comprovantes já anexados permanecem, `platform/documents` não tem operação de exclusão | Consequência direta do desenho de `platform/documents`, não uma nova decisão | y |
| `Supersedes` (correção de um comprovante já anexado) | Nunca usado — cada anexo é um documento independente, sem ligação com anexos anteriores do mesmo lançamento | `FIN-D-016`, decisão do mantenedor: comprovantes são evidências independentes no V1; se um comprovante errado for anexado, o correto é anexado como um novo documento independente, não uma nova versão do anterior | y |

Open questions: none.

## User Stories

### CMP-01: Anexar comprovante

**User Story**: Como Tesouraria, quero anexar um comprovante a um lançamento, para documentar a movimentação financeira.

**Acceptance Criteria**:
1. WHEN a Tesouraria attaches a document to a lançamento THEN the system SHALL call `documents.Service.Store` with `OwnerType="financeiro.lancamento"`, `OwnerID` equal to the lançamento's id, and `RequiredPermission="financeiro:comprovante:create"`.
2. IF the referenced lançamento does not exist THEN the system SHALL reject the attachment with the error `lancamento_nao_encontrado` before calling `documents.Service.Store`.
3. IF the actor lacks `financeiro:comprovante:create` THEN `documents.Service.Store` SHALL return a forbidden error and SHALL NOT upload anything (behavior already guaranteed by `platform/documents`, DOC-01 AC9).
4. The system SHALL NOT require a comprovante to exist for any lançamento operation defined in `01`, `02` or `03`.
5. The system SHALL NOT pass `Supersedes` when calling `documents.Service.Store` — every comprovante attached to a lançamento is an independent document (`FIN-D-016`), never a new version of a previously attached one.

**Independent Test**: anexar um comprovante a um lançamento `CRIADA` e a um `CANCELADA` — os dois aceitam, sem nenhuma outra regra de negócio interferindo.

---

### CMP-02: Consultar e baixar

**User Story**: Como Tesouraria, Presidente, Diretoria ou Conselho Fiscal, quero baixar um comprovante já anexado, para conferir a movimentação.

**Acceptance Criteria**:
1. WHEN an actor holding `financeiro:comprovante:read` requests access to a comprovante THEN the system SHALL call `documents.Service.AccessURL` with that permission, returning a presigned URL.
2. IF the actor lacks `financeiro:comprovante:read` THEN the system SHALL return a forbidden error and SHALL NOT generate a URL (guaranteed by `platform/documents`, DOC-02 AC2).

**Independent Test**: anexar um comprovante, baixá-lo com sucesso via a URL retornada.

---

### CMP-03: Listar comprovantes de um lançamento

**User Story**: Como Tesouraria, Presidente, Diretoria ou Conselho Fiscal, quero listar os comprovantes de um lançamento, para saber quais documentos já foram anexados.

**Acceptance Criteria**:
1. WHEN an actor holding `financeiro:comprovante:read` lists comprovantes of a lançamento THEN the system SHALL call `documents.Service.ListByOwner` with `OwnerType="financeiro.lancamento"` and the lançamento's id, returning every document attached to it, newest first — each independently `Version 1`, since `CMP-01` never chains attachments via `Supersedes`.
2. The system SHALL NOT generate any signed URL as part of this listing (same behavior as `platform/documents`' `ListByOwner`, DOC-04 AC1).
3. The system SHALL NOT record any audit entry for this listing (same behavior as `platform/documents`' `ListByOwner`, DOC-04 AC4) — nor SHALL it write anything to the lançamento's `status`.

**Independent Test**: anexar dois comprovantes distintos ao mesmo lançamento; listar e ver os dois, mais novo primeiro, sem nenhuma URL nem entrada de auditoria de listagem.

## Edge Cases

- Anexar comprovante a um lançamento cancelado: aceito, sem restrição (CMP-01 AC4).
- Cancelar um lançamento que já tem comprovante: o comprovante permanece acessível — nenhuma tarefa desta spec ou de `03` o remove ou esconde.
- Nenhum caso de uso desta spec jamais escreve em `LancamentoFinanceiro` — só lê o `id` para repassar como `owner_id`.

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
|---|---|---|---|
| CMP-01 | Anexar comprovante | In Tasks | Done (T1) |
| CMP-02 | Consultar e baixar | In Tasks | Done (T2) |
| CMP-03 | Listar comprovantes | In Tasks | Done (T3) |

**Coverage:** 3 total, 3 mapped to tasks, 0 unmapped.

## Success Criteria

- Nenhuma regra de armazenamento é reimplementada — tudo delega a `platform/documents`.
- Nenhuma operação desta spec altera o `status` de um lançamento.
