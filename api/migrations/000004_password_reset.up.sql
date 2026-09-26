-- Recuperação de senha por e-mail e contador de tentativas de troca de senha.
-- Roda com o papel dono (tj_owner).
--
-- O token de recuperação NUNCA é gravado: só o SHA-256 dele. O e-mail também não
-- é gravado nas solicitações: só o HMAC-SHA256.

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'tj_app') THEN
    RAISE EXCEPTION 'papel tj_app não existe: crie os papéis do banco (docker/postgres/init/01-roles.sql) antes de migrar';
  END IF;
END
$$;

CREATE TABLE password_reset_tokens (
  id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id    uuid NOT NULL REFERENCES users (id),
  token_hash bytea NOT NULL UNIQUE,
  created_at timestamptz NOT NULL DEFAULT now(),
  expires_at timestamptz NOT NULL,
  used_at    timestamptz,
  request_id text
);
CREATE INDEX password_reset_tokens_user_idx ON password_reset_tokens (user_id) WHERE used_at IS NULL;

CREATE TABLE password_reset_requests (
  id           bigserial PRIMARY KEY,
  email_hash   bytea NOT NULL,
  requested_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX password_reset_requests_idx ON password_reset_requests (email_hash, requested_at DESC);

-- Contador próprio da troca de senha, separado do login (login_attempts).
CREATE TABLE password_change_attempts (
  id           bigserial PRIMARY KEY,
  user_id      uuid NOT NULL REFERENCES users (id),
  attempted_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX password_change_attempts_idx ON password_change_attempts (user_id, attempted_at DESC);

-- Tokens: sem DELETE, o histórico de recuperações é preservado.
GRANT SELECT, INSERT, UPDATE ON password_reset_tokens TO tj_app;
GRANT SELECT, INSERT, DELETE ON password_reset_requests, password_change_attempts TO tj_app;
GRANT USAGE, SELECT ON SEQUENCE password_reset_requests_id_seq, password_change_attempts_id_seq TO tj_app;
