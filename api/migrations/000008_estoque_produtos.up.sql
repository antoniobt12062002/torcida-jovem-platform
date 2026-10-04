-- SKU mínimo de estoque (estoque/01-produtos, PRD-01, EST-D-001/EST-D-007).
-- Este script roda com o papel dono (tj_owner). O papel da aplicação
-- (tj_app) recebe INSERT e SELECT, nunca UPDATE nem DELETE: o SKU é estável
-- após a criação (sem update, sem deactivate, sem exclusão física).
-- Nome da tabela `produtos_estoque`, não `produtos`: evita colisão futura
-- com um catálogo comercial em `loja` (EST-D-011).

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'tj_app') THEN
    RAISE EXCEPTION 'papel tj_app não existe: crie os papéis do banco (docker/postgres/init/01-roles.sql) antes de migrar';
  END IF;
END
$$;

CREATE TABLE produtos_estoque (
  id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  codigo          text NOT NULL UNIQUE,
  nome            text NOT NULL,
  unidade_medida  text NOT NULL,
  criado_em       timestamptz NOT NULL DEFAULT now()
);

GRANT INSERT, SELECT ON produtos_estoque TO tj_app;
