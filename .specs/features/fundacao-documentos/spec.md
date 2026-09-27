# Fundação Documentos Specification

Status: Implementada e mesclada em `develop` (PR #38, 2026-09-27); validação final da feature em andamento. Feature 2 de 2 da fundação, executada **depois** de `fundacao-core` (concluída e validada). Narrativa em português; os critérios de aceite (ACs) seguem o padrão EARS em inglês, exigido pelo validador `validate_spec.py`.

## Problem Statement

O financeiro precisa anexar comprovantes (nota fiscal, recibo, comprovante PIX, contrato, orçamento, fotos) de modo que não possam ser trocados nem apagados, e que só a administração e o Conselho Fiscal consultem (FIN-001, seção 15). Esta feature entrega o armazenamento de documentos em object storage compatível com S3, com versionamento imutável, validações de arquivo, acesso por URL assinada e auditoria, sobre a fundação entregue por `fundacao-core` (auditoria, permissões, unidade de trabalho, infraestrutura de testes e migrações).

## Goals

- [ ] Documentos guardados em object storage compatível com S3, imutáveis e versionados.
- [ ] Upload com validação de tamanho (10 MiB), extensão e tipo real do conteúdo (PDF, JPG, PNG e WebP).
- [ ] Acesso por URL assinada de curta duração, com permissão verificada e acesso auditado.
- [ ] Consulta de metadados por dono (`owner_type`, `owner_id`), sem gerar URL.
- [ ] Nenhum provedor específico: só o protocolo S3, com emulador local para desenvolvimento e testes.
- [ ] `platform/documents` não possui nenhuma permissão de negócio própria: quem chama informa qual permissão vale para cada operação.

## Out of Scope

| Feature | Reason |
|---|---|
| Endpoints HTTP de documentos | O serviço é da plataforma; cada módulo expõe o que precisa |
| Escolha do provedor de object storage | ADR-006 deixa para depois; usa só o protocolo S3 |
| Varredura de malware em uploads | Em aberto no ADR-006; entra em feature própria |
| Identidade, permissões, auditoria e migrações-base | Entregues por `fundacao-core` |
| Retenção e purga de documentos | Decisão arquitetural desta camada: não existe operação de exclusão (definitiva ou lógica) em `platform/documents`, coerente com ADR-004. Política de retenção, se e quando existir, pertence a cada módulo consumidor (ex.: o financeiro decide quando um documento deixa de ser relevante, nunca apaga o que já foi gravado) |
| Workflow de estado do documento | O campo `status` nasce só com o valor `ACTIVE`; nenhuma transição de estado é implementada nesta feature (sem UPDATE concedido à `tj_app` na tabela) |
| Telas de upload e download | A UI é feita nas features dos módulos |

---

## Assumptions & Open Questions

Todas as ambiguidades estão resolvidas ou registradas aqui. "Confirmed? n" aguarda confirmação na aprovação.

| Assumption / decision | Chosen default | Rationale | Confirmed? |
|---|---|---|---|
| Documentos aceitos | Até 10 MiB; PDF, JPG, PNG e WebP; extensões fora dessa lista são bloqueadas; o tipo é detectado pelo conteúdo e deve coincidir com a extensão | Definido pelo mantenedor; cobre FIN-001 seção 15 | y |
| Fluxo de upload e download | Upload passa pela API (streaming para o S3); download por URL assinada com validade de 300 segundos | Permite validar tipo, tamanho e permissão antes de gravar | y |
| Emulador de S3 local | Garage v2.4.1 (AGPL-3.0, ativo), só para desenvolvimento e testes — MinIO (arquivado em 2026-04) e LocalStack (exige conta desde 2026-03) foram descartados. O emulador valida o adaptador S3 e os fluxos da aplicação, mas não é garantia completa do comportamento de um provedor S3 de produção (ver "Limite do emulador S3 local") | Não define provedor de produção (ADR-006); AGPL só se aplica ao servidor de teste, nunca ao código deste repositório | y |
| Autorização parametrizada | `platform/documents` não declara nem embute nenhuma permissão de negócio. Cada operação (`Store`, `AccessURL`, `ListByOwner`) recebe do chamador o `authz.Principal` e a `authz.Permission` exigida, e só aplica `authz.Require`; o nome e o dono da permissão (ex.: `financeiro:documento:create`, quando o financeiro tiver spec própria) são decisão do módulo consumidor, nunca de `platform` | Aprovado pelo mantenedor: nenhuma regra de negócio de documentos entra em `platform` | y |
| Formato de `owner_type` | Validação só técnica (não vazio, tamanho razoável, caracteres de identificador com um separador — ex.: `financeiro.lancamento`); nenhum catálogo fixo de valores aceitos, para não acoplar `platform/documents` aos módulos que ainda não existem | Aprovado pelo mantenedor | y |
| Versão atual de um documento | `platform/documents` não rastreia "qual é a versão mais recente" de um documento; cada versão é sua própria linha e seu próprio `id`. O módulo consumidor guarda qual `id` é o vigente e o atualiza ao gravar uma nova versão | Aprovado pelo mantenedor: isso é estado de negócio, não infraestrutura | y |
| Status técnico do documento | Campo `status`, com `ACTIVE` como único valor emitido nesta feature; existe para permitir estados futuros (ex.: quarentena, rejeitado) sem migração de schema, mas nenhum workflow de transição é implementado agora | Aprovado pelo mantenedor | y |
| Limite do emulador S3 local | Nenhum emulador S3 local gratuito verificado (MinIO, LocalStack, SeaweedFS, Garage) aplica a expiração de uma URL assinada — só a assinatura em si. É lacuna documentada do ecossistema (ex.: o S3Mock da Adobe declara isso abertamente), não falha de escolha. A aplicação é responsável só por pedir e gerar a URL com o TTL correto (verificado localmente, decodificando `X-Amz-Expires`); a aplicação efetiva da expiração é responsabilidade do provedor S3 e só é verificável de fato num ambiente com um provedor de produção | Aprovado pelo mantenedor, depois de testar MinIO (descontinuado), LocalStack (exige conta) e SeaweedFS e Garage (nenhum aplica expiração) | y |
| Dependências da `fundacao-core` | Usa `WithTx`, `audit.Recorder`, `authz.Require`, `testutil` e o sistema de migrações da core | Evita duplicar infraestrutura | y |
| SDK de S3 e emulador | Escolhidos e fixados nas tarefas T3 e T4, após checar a documentação vigente | Aprovado: nenhuma versão é fixada sem checar a documentação vigente | y |

**Open questions:** none - all resolved or logged above.

---

## User Stories

### P2: Armazenamento de documentos

**User Story**: Como financeiro, quero anexar comprovantes de forma que não possam ser trocados nem apagados, para sustentar a prestação de contas.

**Why P2**: O financeiro precisa, mas a fundação já se prova sem ele.

**Acceptance Criteria**:
1. WHEN a document is stored THEN the system SHALL write the object to the private bucket under the key `documents/<uuid>` and insert a metadata row with id, owner type, owner id, original filename, content type, size, SHA-256, storage key, version, superseded id, status, uploader and upload time.
2. WHEN a document is stored THEN the system SHALL compute the SHA-256 over the streamed bytes and store it in hexadecimal.
3. IF the size exceeds 10 MiB THEN the system SHALL reject the document with the error `document_too_large` before writing any metadata.
4. IF the content type detected from the bytes is not application/pdf, image/jpeg, image/png or image/webp THEN the system SHALL reject the document with the error `document_type_not_allowed`.
5. WHEN a new version of an existing document is stored THEN the system SHALL set its version to the previous version plus one, reference the previous document as superseded and leave the previous object unchanged.
6. The system SHALL NOT provide any operation that overwrites or deletes a stored document object or its metadata row.
7. IF inserting the metadata fails after the object was uploaded THEN the system SHALL delete only the object that was just created and return the error.
8. WHEN a document is stored THEN the system SHALL record an audit entry with action `document.create` in the same transaction as the metadata.
9. IF the caller-supplied permission is not held by the actor THEN the system SHALL return a forbidden error and SHALL NOT write to storage, without the service knowing or deciding what that permission is named.
10. IF the extension of the original filename is not `.pdf`, `.jpg`, `.jpeg`, `.png` or `.webp`, ignoring letter case, THEN the system SHALL reject the document with the error `document_extension_not_allowed` before reading its content.
11. IF the extension does not match the content type detected from the bytes THEN the system SHALL reject the document with the error `document_type_mismatch`.
12. WHEN a document is stored THEN the system SHALL set its `status` to `ACTIVE`, with no operation in this feature that changes it afterward.

**Independent Test**: Guardar um PDF, guardar uma nova versão e ver a cadeia de versões com o objeto antigo intacto.

---

### P2: Acesso a documentos por URL assinada

**User Story**: Como conselheiro fiscal, quero abrir documentos por link temporário e ter o acesso registrado, para consultar comprovantes sem expor o bucket.

**Why P2**: Completa o ciclo dos documentos com controle de acesso (ADR-006).

**Acceptance Criteria**:
1. WHEN an actor holding the caller-supplied required permission requests access to a document by its exact id THEN the system SHALL return a presigned GET URL valid for 300 seconds and record an audit entry with action `document.access` naming the document and the actor.
2. IF the actor lacks the caller-supplied required permission THEN the system SHALL return a forbidden error and SHALL NOT generate a URL.
3. IF the document id does not exist THEN the system SHALL return a not-found error and SHALL NOT generate a URL.
4. WHEN the system requests a presigned GET URL from the storage adapter THEN it SHALL pass the configured TTL through unchanged, and the generated URL SHALL carry that same value in its `X-Amz-Expires` query parameter, verified in an integration test that decodes the URL. Enforcing that a request made after the URL's expiry is rejected is the storage provider's responsibility, not this system's, and is out of reach of a local test: see the Assumption "Limite do emulador S3 local".
5. IF an object is requested from the bucket without a valid signature THEN the storage SHALL reject the request, verified in an integration test.
6. The system SHALL NOT resolve "the current version" of a document on the caller's behalf: access is always requested by the exact id of one specific version, and tracking which version is current is the calling module's responsibility.

**Independent Test**: Gerar a URL, conferir o `X-Amz-Expires` na própria URL, baixar o arquivo e ver que uma cópia sem assinatura é recusada.

---

### P2: Consulta de documentos por dono

**User Story**: Como módulo consumidor, quero listar os documentos de um dono (ex.: um lançamento financeiro), para montar minha própria tela sem duplicar metadados.

**Why P2**: Sem consulta, nenhum módulo consegue saber quais documentos já existem para o que ele possui.

**Acceptance Criteria**:
1. WHEN an actor holding the caller-supplied required permission calls the listing with an owner type and an owner id THEN the system SHALL return every document version whose owner type and owner id match, ordered from newest to oldest, without generating any signed URL.
2. IF the actor lacks the caller-supplied required permission THEN the system SHALL return a forbidden error and SHALL NOT return any row.
3. IF the owner type is empty, exceeds 100 characters, or does not match the technical format `[a-z0-9_]+\.[a-z0-9_]+` THEN the system SHALL reject the call with the error `document_owner_type_invalid`, with no fixed catalog of accepted values.
4. The system SHALL NOT record an audit entry for a listing call: only `document.access` (DOC-02) audits disclosure of a document's content.

**Independent Test**: Guardar duas versões de um documento e uma de outro dono; listar pelo primeiro dono e ver só as duas versões, mais novas primeiro.

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
- IF two different callers pass two different required permissions for the same kind of operation THEN the system SHALL check exactly the permission it was given for that call, never a permission it remembers from a previous call.

---

## Requirement Traceability

| Requirement ID | Story | Phase | Status |
|---|---|---|---|
| DOC-01 | P2: Armazenamento de documentos | Done | Implemented |
| DOC-02 | P2: Acesso a documentos por URL assinada | Done | Implemented |
| DOC-03 | P2: Configuração do storage | Done | Implemented |
| DOC-04 | P2: Consulta de documentos por dono | Done | Implemented |

**Coverage:** 4 total, 4 mapped to tasks, 0 unmapped (mapeamento detalhado em `tasks.md`, seção Requirement Coverage).

---

## Success Criteria

- [ ] Um documento é guardado, versionado, acessado por URL assinada e o acesso aparece na auditoria.
- [ ] Extensão fora da lista, tipo divergente da extensão e arquivo acima de 10 MiB são recusados antes de gravar metadados.
- [ ] `platform/documents` compila e roda sem depender de nenhuma permissão nomeada dentro do próprio pacote.
- [ ] Listar por dono devolve as versões certas, mais novas primeiro, sem gerar URL nem auditoria.
- [ ] `go test -tags=integration ./...` passa no CI com o emulador S3.
