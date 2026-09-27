# Fundação Documentos Design

**Spec**: `.specs/features/fundacao-documentos/spec.md`
**Status**: Implementado e mesclado em `develop` (PR #38, 2026-09-27); validação final da feature em andamento. Feature 2 de 2; depende de `fundacao-core` concluída e validada.

Decisões ativas do projeto respeitadas: AD-001 a AD-009, em especial AD-008 (object storage S3, sem provedor definido). Nenhuma é substituída.

---

## Architecture Overview

O serviço de documentos combina permissão, storage e banco. `platform/documents` não declara nenhuma permissão própria: cada chamada recebe o `authz.Principal` do ator e a `authz.Permission` exigida, e só repassa para `authz.Require` — quem decide o nome e o dono da permissão é sempre o módulo chamador (ver Tech Decisions). O arquivo vai ao bucket privado por streaming; os metadados e a auditoria entram numa única transação.

```mermaid
graph TD
    UC[módulo chamador<br/>decide a permissão e o owner_type/owner_id] --> SVC[documents.Service]
    SVC --> AZ["authz.Require(actor, permissão recebida)"]
    SVC --> VAL[validação: extensão, tamanho, conteúdo]
    SVC --> ST[storage.Storage - protocolo S3]
    SVC --> TX[database.WithTx]
    TX --> META[(documents)]
    TX --> AUD[audit.Recorder]
    ST --> S3[(bucket privado)]
```

Ordem em `Store`: permissão (a que o chamador passou), validação de extensão, leitura em streaming com limite de 10 MiB e detecção do tipo pelo conteúdo, envio ao storage calculando o SHA-256, e então a transação com metadados (`status = ACTIVE`) e auditoria. Se a transação falha, apenas o objeto recém-criado é removido. `AccessURL` e `ListByOwner` seguem a mesma regra: primeiro a permissão recebida, depois a operação; `ListByOwner` não audita (leitura de metadados, não de conteúdo) e não gera URL.

---

## Code Reuse Analysis

### Existing Components to Leverage

| Component | Location | How to Use |
|---|---|---|
| Unidade de trabalho e auditoria | `platform/database` e `platform/audit` (da `fundacao-core`) | `WithTx` e `Recorder.Record` na mesma transação |
| Permissões | `platform/authz` | `Require(ctx, actor, permissão recebida do chamador)`; nenhuma permissão declarada por `platform/documents` |
| Infraestrutura de testes | `platform/testutil` | `NewTestDB` e `SharedS3` (novo, T3), no mesmo padrão de `SharedPostgres` |
| Configuração | `api/internal/config/config.go` | Novos campos de S3 e de validade da URL (T1) |

### Integration Points

| System | Integration Method |
|---|---|
| Object storage S3 | Cliente S3 atrás da interface `storage.Storage`, endpoint configurável e estilo de path opcional; AWS SDK for Go v2 (`service/s3`), confirmado na tarefa T4 — o mesmo cliente que valida o emulador em T3 |
| PostgreSQL | Migração `000005_documents`, com `tj_app` sem UPDATE e DELETE |
| Emulador de testes | Garage v2.4.1 (`platform/testutil/s3.go`), único processo com `--single-node --default-bucket`; valida o adaptador S3 e os fluxos da aplicação, mas não é garantia completa do comportamento de um provedor S3 de produção (ver Risks & Concerns) |

---

## Components

### platform/storage

- **Purpose**: Interface mínima para object storage S3, sem provedor definido.
- **Location**: `api/internal/platform/storage/`
- **Interfaces**: `Put(ctx, key string, r io.Reader, size int64, contentType string) error`; `PresignGet(ctx, key string, ttl time.Duration) (string, error)`; `DeleteCreated(ctx, key string) error` (só para desfazer upload órfão).
- **Dependencies**: SDK S3 (na implementação `s3.go`).

### platform/documents

- **Purpose**: Guardar, versionar, listar e dar acesso a documentos, aplicando a permissão que o chamador exigir, sem conhecer nenhuma permissão de negócio.
- **Location**: `api/internal/platform/documents/`
- **Interfaces**:
  - `Service.Store(ctx, StoreInput) (Document, error)` — `StoreInput` carrega `Actor authz.Principal`, `RequiredPermission authz.Permission`, `OwnerType`, `OwnerID`, `Filename`, `Content io.Reader`, `UploadedBy`.
  - `Service.AccessURL(ctx, AccessInput) (string, error)` — `AccessInput` carrega `Actor`, `RequiredPermission`, `ID uuid.UUID`.
  - `Service.ListByOwner(ctx, ListInput) ([]Document, error)` — `ListInput` carrega `Actor`, `RequiredPermission`, `OwnerType`, `OwnerID`; valida só o formato de `OwnerType` (`[a-z0-9_]+\.[a-z0-9_]+`), sem catálogo fixo.
- **Dependencies**: `storage`, `audit`, `authz`, `database`.

---

## Data Models

Migração `000005_documents` (com `down`, e falha com mensagem clara se o papel `tj_app` não existir).

```sql
-- 000005_documents
CREATE TABLE documents (
  id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  owner_type        text NOT NULL,
  owner_id          text NOT NULL,
  original_filename text NOT NULL,
  content_type      text NOT NULL,
  size_bytes        bigint NOT NULL CHECK (size_bytes > 0),
  sha256            char(64) NOT NULL,
  storage_key       text NOT NULL UNIQUE,
  version           int NOT NULL CHECK (version >= 1),
  supersedes_id     uuid UNIQUE REFERENCES documents,
  status            text NOT NULL DEFAULT 'ACTIVE',
  uploaded_by       uuid NOT NULL REFERENCES users,
  uploaded_at       timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX documents_owner_idx ON documents (owner_type, owner_id, uploaded_at DESC);
-- GRANT INSERT, SELECT ON documents TO tj_app (sem UPDATE e DELETE)
```

**Relationships**: `documents.supersedes_id` forma a cadeia de versões, e o `UNIQUE` impede bifurcação. O documento pertence a uma entidade de outro módulo por `owner_type` e `owner_id`, sem chave estrangeira, porque o dono é definido por cada módulo. `status` nasce sempre `ACTIVE`; sem `UPDATE` concedido a `tj_app`, nenhuma transição é possível nesta feature — o campo existe para um estado futuro (ex.: quarentena) sem exigir migração de schema quando isso for decidido, mas nenhuma tarefa desta tabela implementa transição.

---

## Error Handling Strategy

| Error Scenario | Handling | User Impact |
|---|---|---|
| Extensão fora da lista | `document_extension_not_allowed`, antes de ler o conteúdo | Arquivo recusado |
| Tipo divergente da extensão ou fora da lista | `document_type_mismatch` ou `document_type_not_allowed` | Arquivo recusado |
| Arquivo acima de 10 MiB | `document_too_large`, antes de gravar metadados | Arquivo recusado |
| Sem a permissão que o chamador exigiu | Erro de proibido, sem escrita, sem URL e sem linha listada | Acesso negado |
| Documento inexistente | Erro de não encontrado, sem URL | Não encontrado |
| `owner_type` fora do formato técnico | `document_owner_type_invalid`, antes de consultar o banco | Chamada recusada |
| Falha no upload ao storage | Erro antes de gravar metadados | Nenhum documento registrado |
| Falha ao gravar metadados após upload | Remove só o objeto recém-criado | Nenhum documento registrado, sem órfão |
| Storage inalcançável | Devolve o erro do storage, sem gravar metadados | Tentar mais tarde |

---

## Risks & Concerns

| Concern | Location (file:line) | Impact | Mitigation |
|---|---|---|---|
| Objeto enviado e metadados gravados em sistemas diferentes | não se aplica (ainda não implementado) | Objeto órfão se a transação falhar | Remover só o objeto recém-criado e teste de falha simulada (DOC-01.7) |
| Detecção de tipo pelo conteúdo pode ser enganada por arquivos poliglotas | não se aplica | Arquivo malicioso com aparência de imagem | Sem varredura de malware nesta fase (fora de escopo); acesso só por URL assinada curta e permissão |
| URL assinada vaza por compartilhamento do link | não se aplica | Acesso por terceiros durante 300 segundos | Validade curta e auditoria de cada emissão |
| A migração depende do papel `tj_app`, criado fora dela | `docs/architecture/architecture-overview.md` | Falha em ambiente sem o papel | Migração falha com mensagem clara; o script local e o helper de testes criam o papel |
| Nenhum emulador S3 local gratuito verificado aplica a expiração de uma URL assinada (só a assinatura em si) — testado em MinIO (descontinuado antes de testar), LocalStack (descartado antes de testar por exigir conta), SeaweedFS 4.47 e Garage v2.4.1, ambos aceitando uma URL 7s após um TTL de 3s | `api/internal/platform/testutil/s3.go` | A aplicação da expiração em produção não é coberta por teste automatizado neste repositório | DOC-02.4 verifica só o que esta aplicação controla: o TTL pedido chega inalterado ao adaptador e aparece como `X-Amz-Expires` na URL gerada. A aplicação efetiva da expiração é responsabilidade do provedor S3 de produção (comportamento padrão documentado pela AWS) e só é verificável nesse ambiente — risco residual aceito e registrado, não coberto por teste local |

---

## Tech Decisions (only non-obvious ones)

| Decision | Choice | Rationale |
|---|---|---|
| Upload | Pela API, em streaming, sem gravar em disco | Valida tamanho, extensão, conteúdo e permissão antes de persistir |
| Versionamento | Nova linha com `version + 1` e `supersedes_id` | Imutabilidade (ADR-004); o objeto antigo nunca muda |
| Remoção de objeto | Só do objeto recém-criado quando o registro falha | Não viola a proibição de apagar documentos, pois o objeto nunca foi referenciado |
| Chave do objeto | `documents/<uuid>` | Não expõe o nome original e evita colisão |
| Permissão | Parâmetro de cada chamada (`RequiredPermission authz.Permission` no Input), nunca uma constante do pacote | `platform` não decide nome nem dono de permissão de negócio; o módulo consumidor decide |
| "Versão atual" | Não existe no serviço; cada `id` é uma versão específica | Rastrear "o vigente" é estado de negócio do módulo dono, não de infraestrutura |
| Status | Coluna `status`, só `ACTIVE` nesta feature, sem `UPDATE` concedido | Prepara espaço para estados futuros sem exigir nova migração, sem construir workflow agora |
| Formato de `owner_type` | Regex técnica (`[a-z0-9_]+\.[a-z0-9_]+`), sem lista fixa de módulos | `platform/documents` não pode conhecer os módulos que ainda não existem |
