-- Metadados de documentos (fundacao-documentos, DOC-01).
-- Este script roda com o papel dono (tj_owner). O papel da aplicação (tj_app)
-- recebe apenas INSERT e SELECT: sem UPDATE nem DELETE, nenhuma linha muda ou
-- desaparece depois de gravada (AD-015 / docs/adr, imutabilidade de versão).

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'tj_app') THEN
    RAISE EXCEPTION 'papel tj_app não existe: crie os papéis do banco (docker/postgres/init/01-roles.sql) antes de migrar';
  END IF;
END
$$;

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
  -- UNIQUE impede que duas versões diferentes declarem suceder a mesma
  -- versão anterior (bifurcação da cadeia).
  supersedes_id     uuid UNIQUE REFERENCES documents,
  -- Nasce sempre ACTIVE; sem UPDATE concedido a tj_app, nenhuma transição é
  -- possível nesta feature. A coluna existe para um estado futuro (ex.:
  -- quarentena) sem exigir nova migração quando isso for decidido.
  status            text NOT NULL DEFAULT 'ACTIVE',
  uploaded_by       uuid NOT NULL REFERENCES users,
  uploaded_at       timestamptz NOT NULL DEFAULT now()
);

-- Suporta ListByOwner (DOC-04): mais novo primeiro, por dono.
CREATE INDEX documents_owner_idx ON documents (owner_type, owner_id, uploaded_at DESC);

GRANT INSERT, SELECT ON documents TO tj_app;
