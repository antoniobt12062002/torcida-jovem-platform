-- Identidade: usuários, papéis, permissões, vínculo administrativo, sessões e
-- tentativas de login. Roda com o papel dono (tj_owner).
--
-- Papéis e permissões NÃO são semeados aqui: a sincronização na partida da API
-- os cria a partir do código (e audita o que mudou). O usuário não referencia o
-- associado: o vínculo pertence a associados.associados.user_id.

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'tj_app') THEN
    RAISE EXCEPTION 'papel tj_app não existe: crie os papéis do banco (docker/postgres/init/01-roles.sql) antes de migrar';
  END IF;
END
$$;

CREATE TABLE users (
  id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  email                text NOT NULL,
  name                 text NOT NULL,
  password_hash        text NOT NULL,
  is_active            boolean NOT NULL DEFAULT true,
  must_change_password boolean NOT NULL DEFAULT false,
  created_at           timestamptz NOT NULL DEFAULT now(),
  updated_at           timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX users_email_uq ON users (lower(email));

CREATE TABLE roles (
  id          smallserial PRIMARY KEY,
  name        text NOT NULL UNIQUE,
  description text NOT NULL
);

CREATE TABLE permissions (
  id          serial PRIMARY KEY,
  name        text NOT NULL UNIQUE CHECK (name ~ '^[a-z_]+:[a-z_]+:[a-z_]+$'),
  description text NOT NULL DEFAULT '',
  is_active   boolean NOT NULL DEFAULT true
);

CREATE TABLE role_permissions (
  role_id       smallint NOT NULL REFERENCES roles (id),
  permission_id integer  NOT NULL REFERENCES permissions (id),
  PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE user_roles (
  user_id uuid     NOT NULL REFERENCES users (id),
  role_id smallint NOT NULL REFERENCES roles (id),
  PRIMARY KEY (user_id, role_id)
);

-- Por que o usuário tem acesso administrativo (o papel diz o que ele pode
-- fazer). Um vínculo ativo por usuário; o histórico nunca é apagado.
CREATE TABLE admin_memberships (
  id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       uuid NOT NULL REFERENCES users (id),
  reason        text NOT NULL CHECK (length(btrim(reason)) >= 10),
  granted_by    uuid REFERENCES users (id), -- nulo apenas no bootstrap
  granted_at    timestamptz NOT NULL DEFAULT now(),
  revoked_by    uuid REFERENCES users (id),
  revoke_reason text,
  revoked_at    timestamptz
);
CREATE UNIQUE INDEX admin_memberships_active_uq ON admin_memberships (user_id) WHERE revoked_at IS NULL;

CREATE TABLE sessions (
  id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      uuid NOT NULL REFERENCES users (id),
  token_hash   bytea NOT NULL UNIQUE,
  csrf_token   text NOT NULL,
  created_at   timestamptz NOT NULL DEFAULT now(),
  last_seen_at timestamptz NOT NULL DEFAULT now(),
  expires_at   timestamptz NOT NULL,
  revoked_at   timestamptz
);
CREATE INDEX sessions_user_idx ON sessions (user_id) WHERE revoked_at IS NULL;

CREATE TABLE login_attempts (
  id           bigserial PRIMARY KEY,
  email_hash   bytea NOT NULL,
  attempted_at timestamptz NOT NULL DEFAULT now(),
  success      boolean NOT NULL
);
CREATE INDEX login_attempts_idx ON login_attempts (email_hash, attempted_at DESC);

GRANT SELECT, INSERT, UPDATE, DELETE ON users, roles, permissions, role_permissions, user_roles, sessions, login_attempts TO tj_app;
-- Sem DELETE: o histórico do vínculo administrativo é preservado.
GRANT SELECT, INSERT, UPDATE ON admin_memberships TO tj_app;
GRANT USAGE, SELECT ON SEQUENCE roles_id_seq, permissions_id_seq, login_attempts_id_seq TO tj_app;
