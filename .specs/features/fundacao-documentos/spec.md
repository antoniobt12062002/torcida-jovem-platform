# Fundação Documentos Specification

Status: Aprovada com ajustes em 2026-09-26. Feature 2 de 2 da fundação, executada **depois** de `fundacao-core`. Narrativa em português; os critérios de aceite (ACs) seguem o padrão EARS em inglês, exigido pelo validador `validate_spec.py`.

## Problem Statement

O financeiro precisa anexar comprovantes (nota fiscal, recibo, comprovante PIX, contrato, orçamento, fotos) de modo que não possam ser trocados nem apagados, e que só a administração e o Conselho Fiscal consultem (FIN-001, seção 15). Esta feature entrega o armazenamento de documentos em object storage compatível com S3, com versionamento imutável, validações de arquivo, acesso por URL assinada e auditoria, sobre a fundação entregue por `fundacao-core` (auditoria, permissões, unidade de trabalho, infraestrutura de testes e migrações).

## Goals

- [ ] Documentos guardados em object storage compatível com S3, imutáveis e versionados.
- [ ] Upload com validação de tamanho (10 MiB), extensão e tipo real do conteúdo (PDF, JPG, PNG e WebP).
- [ ] Acesso por URL assinada de curta duração, com permissão verificada e acesso auditado.
- [ ] Nenhum provedor específico: só o protocolo S3, com emulador local para desenvolvimento e testes.

## Out of Scope

| Feature | Reason |
|---|---|
| Endpoints HTTP de documentos | O serviço é da plataforma; cada módulo expõe o que precisa |
| Escolha do provedor de object storage | ADR-006 deixa para depois; usa só o protocolo S3 |
| Varredura de malware em uploads | Em aberto no ADR-006; entra em feature própria |
| Identidade, permissões, auditoria e migrações-base | Entregues por `fundacao-core` |
| Retenção e purga de documentos | Depende de definição jurídica; ADR-004 proíbe exclusão definitiva |
| Telas de upload e download | A UI é feita nas features dos módulos |

---

## Assumptions & Open Questions

Todas as ambiguidades estão resolvidas ou registradas aqui. "Confirmed? n" aguarda confirmação na aprovação.

| Assumption / decision | Chosen default | Rationale | Confirmed? |
|---|---|---|---|
| Documentos aceitos | Até 10 MiB; PDF, JPG, PNG e WebP; extensões fora dessa lista são bloqueadas; o tipo é detectado pelo conteúdo e deve coincidir com a extensão | Definido pelo mantenedor; cobre FIN-001 seção 15 | y |
| Fluxo de upload e download | Upload passa pela API (streaming para o S3); download por URL assinada com validade de 300 segundos | Permite validar tipo, tamanho e permissão antes de gravar | n |
| Emulador de S3 local | Um emulador compatível com S3 será escolhido na tarefa correspondente, apenas para desenvolvimento e testes | Não define provedor de produção (ADR-006); a escolha exige checar licença e manutenção atuais | n |
| Permissões de documentos | ADMIN e PRESIDENTE: `document:file:create` e `document:file:read`; FINANCEIRO: create e read; CONSELHO_FISCAL: só read; DIRETOR e ASSOCIADO: nenhuma | Alinhado a FIN-001 seção 15 e às permissões do Conselho Fiscal aprovadas; revisada na spec do financeiro | n |
| Dependências da `fundacao-core` | Usa `WithTx`, `audit.Recorder`, `authz.Require`, `testutil` e o sistema de migrações da core | Evita duplicar infraestrutura | y |
| SDK de S3 e emulador | Escolhidos e fixados nas tarefas T3 e T4, após checar a documentação vigente | Aprovado: nenhuma versão é fixada sem checar a documentação vigente | y |

**Open questions:** none - all resolved or logged above.

---

## User Stories

### P2: Armazenamento de documentos

**User Story**: Como financeiro, quero anexar comprovantes de forma que não possam ser trocados nem apagados, para sustentar a prestação de contas.

**Why P2**: O financeiro precisa, mas a fundação já se prova sem ele.

**Acceptance Criteria**:
1. WHEN a document is stored THEN the system SHALL write the object to the private bucket under the key `documents/<uuid>` and insert a metadata row with id, owner type, owner id, original filename, content type, size, SHA-256, storage key, version, superseded id, uploader and upload time.
2. WHEN a document is stored THEN the system SHALL compute the SHA-256 over the streamed bytes and store it in hexadecimal.
3. IF the size exceeds 10 MiB THEN the system SHALL reject the document with the error `document_too_large` before writing any metadata.
4. IF the content type detected from the bytes is not application/pdf, image/jpeg, image/png or image/webp THEN the system SHALL reject the document with the error `document_type_not_allowed`.
5. WHEN a new version of an existing document is stored THEN the system SHALL set its version to the previous version plus one, reference the previous document as superseded and leave the previous object unchanged.
6. The system SHALL NOT provide any operation that overwrites or deletes a stored document object or its metadata row.
7. IF inserting the metadata fails after the object was uploaded THEN the system SHALL delete only the object that was just created and return the error.
8. WHEN a document is stored THEN the system SHALL record an audit entry with action `document.create` in the same transaction as the metadata.
9. IF the actor lacks `document:file:create` THEN the system SHALL return a forbidden error and SHALL NOT write to storage.
10. IF the extension of the original filename is not `.pdf`, `.jpg`, `.jpeg`, `.png` or `.webp`, ignoring letter case, THEN the system SHALL reject the document with the error `document_extension_not_allowed` before reading its content.
11. IF the extension does not match the content type detected from the bytes THEN the system SHALL reject the document with the error `document_type_mismatch`.

**Independent Test**: Guardar um PDF, guardar uma nova versão e ver a cadeia de versões com o objeto antigo intacto.

---

### P2: Acesso a documentos por URL assinada

**User Story**: Como conselheiro fiscal, quero abrir documentos por link temporário e ter o acesso registrado, para consultar comprovantes sem expor o bucket.

**Why P2**: Completa o ciclo dos documentos com controle de acesso (ADR-006).

**Acceptance Criteria**:
1. WHEN an actor with `document:file:read` requests access to a document THEN the system SHALL return a presigned GET URL valid for 300 seconds and record an audit entry with action `document.access` naming the document and the actor.
2. IF the actor lacks `document:file:read` THEN the system SHALL return a forbidden error and SHALL NOT generate a URL.
3. IF the document id does not exist THEN the system SHALL return a not-found error and SHALL NOT generate a URL.
4. IF a presigned URL is used after its expiry THEN the storage SHALL reject the request, verified in an integration test with a one-second validity.
5. IF an object is requested from the bucket without a valid signature THEN the storage SHALL reject the request, verified in an integration test.

**Independent Test**: Gerar a URL, baixar o arquivo, esperar expirar e ver a recusa.

---

---

### P2: Configuração do storage

**User Story**: Como operador, quero que a API falhe cedo se o storage estiver mal configurado, para não descobrir o erro no primeiro upload.

**Why P2**: Evita falhas silenciosas e mantém o provedor intercambiável.

**Acceptance Criteria**:
1. IF storage is enabled and any of `S3_ENDPOINT`, `S3_REGION`, `S3_BUCKET`, `S3_ACCESS_KEY` or `S3_SECRET_KEY` is missing THEN the process SHALL exit with a non-zero status naming the variable and SHALL NOT print any secret value.
2. IF `DOCUMENT_URL_TTL_SECONDS` is less than or equal to zero THEN the process SHALL refuse to start.
3. WHERE `S3_USE_PATH_STYLE` is true THEN the storage client SHALL address the bucket by path, as required by local emulators.

**Independent Test**: Iniciar com `S3_BUCKET` ausente e ver a mensagem; subir com o emulador em modo path-style.

---

## Edge Cases

- IF the same file is stored twice as separate documents THEN the system SHALL create two distinct documents with distinct keys.
- WHEN the presigned URL validity is configured to zero or negative THEN the system SHALL return a configuration error at startup.
- IF the storage is unreachable during a store operation THEN the system SHALL return the storage error and SHALL NOT write metadata.
- WHEN a document has several versions THEN requesting access without a version SHALL return the URL of the latest version.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
|---|---|---|---|
| DOC-01 | P2: Armazenamento de documentos | In Tasks | Pending |
| DOC-02 | P2: Acesso a documentos por URL assinada | In Tasks | Pending |
| DOC-03 | P2: Configuração do storage | In Tasks | Pending |

**Coverage:** 3 total, 3 mapped to tasks, 0 unmapped (mapeamento detalhado em `tasks.md`, seção Requirement Coverage).

---

## Success Criteria

- [ ] Um documento é guardado, versionado, acessado por URL assinada e o acesso aparece na auditoria.
- [ ] Extensão fora da lista, tipo divergente da extensão e arquivo acima de 10 MiB são recusados antes de gravar metadados.
- [ ] `go test -tags=integration ./...` passa no CI com o emulador S3.
