-- Plano de contas do financeiro (fundacao-financeiro, PC-01, PC-03).
-- Este script roda com o papel dono (tj_owner). O papel da aplicação (tj_app)
-- recebe INSERT, SELECT e UPDATE (nome e ativo mudam sob regra restrita,
-- FIN-D-008), mas nunca DELETE: nenhuma conta é excluída, só desativada.

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'tj_app') THEN
    RAISE EXCEPTION 'papel tj_app não existe: crie os papéis do banco (docker/postgres/init/01-roles.sql) antes de migrar';
  END IF;
END
$$;

CREATE TABLE contas_contabeis (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tipo       text NOT NULL CHECK (tipo IN ('RECEITA', 'DESPESA')),
  nome       text NOT NULL,
  -- Mesma tipo do pai é validada em aplicação (T2): uma CHECK entre linhas
  -- da mesma tabela exigiria um gatilho, desnecessário para uma regra que a
  -- aplicação já garante antes do INSERT.
  parent_id  uuid REFERENCES contas_contabeis,
  ativo      boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX contas_contabeis_parent_idx ON contas_contabeis (parent_id);

GRANT INSERT, SELECT, UPDATE ON contas_contabeis TO tj_app;
