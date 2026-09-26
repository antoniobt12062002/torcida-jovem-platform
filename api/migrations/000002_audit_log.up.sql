-- Registro de auditoria imutável (ADR-004, AUD-02).
-- Este script roda com o papel dono (tj_owner). O papel da aplicação (tj_app)
-- recebe apenas INSERT e SELECT.

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'tj_app') THEN
    RAISE EXCEPTION 'papel tj_app não existe: crie os papéis do banco (docker/postgres/init/01-roles.sql) antes de migrar';
  END IF;
END
$$;

CREATE TABLE audit_log (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  occurred_at   timestamptz NOT NULL DEFAULT now(),
  actor_type    text NOT NULL CHECK (actor_type IN ('user', 'system', 'anonymous')),
  -- Referência lógica a users, sem chave estrangeira: uma restrição jamais pode
  -- impedir a gravação de um registro de auditoria.
  actor_user_id uuid,
  action        text NOT NULL,
  entity_type   text NOT NULL,
  entity_id     text NOT NULL,
  before        jsonb,
  after         jsonb,
  outcome       text NOT NULL CHECK (outcome IN ('success', 'denied', 'failure')),
  reason        text,
  context       jsonb NOT NULL DEFAULT '{}'::jsonb,
  request_id    text,
  CONSTRAINT audit_log_actor_consistent CHECK ((actor_type = 'user') = (actor_user_id IS NOT NULL))
);

CREATE INDEX audit_log_entity_idx ON audit_log (entity_type, entity_id, occurred_at DESC);
CREATE INDEX audit_log_time_idx   ON audit_log (occurred_at DESC, id DESC);

CREATE FUNCTION audit_log_reject_change() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
  RAISE EXCEPTION 'audit_log é imutável: % não é permitido', TG_OP
    USING ERRCODE = 'restrict_violation';
END
$$;

CREATE TRIGGER audit_log_no_update_delete
  BEFORE UPDATE OR DELETE ON audit_log
  FOR EACH ROW EXECUTE FUNCTION audit_log_reject_change();

CREATE TRIGGER audit_log_no_truncate
  BEFORE TRUNCATE ON audit_log
  FOR EACH STATEMENT EXECUTE FUNCTION audit_log_reject_change();

GRANT INSERT, SELECT ON audit_log TO tj_app;
