-- Papéis do ambiente local e dos testes de integração.
-- As senhas abaixo são apenas de desenvolvimento: o repositório é público.
-- O entrypoint do Postgres executa este script como superusuário, no banco POSTGRES_DB.
--
--   tj_owner  dono do banco; executa as migrações (DDL).
--   tj_app    usado pela API; sem DDL e sem privilégios além dos concedidos
--             explicitamente por cada migração.

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'tj_owner') THEN
    CREATE ROLE tj_owner LOGIN PASSWORD 'tj_owner_dev' NOSUPERUSER NOCREATEDB NOCREATEROLE;
  END IF;

  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'tj_app') THEN
    CREATE ROLE tj_app LOGIN PASSWORD 'tj_app_dev' NOSUPERUSER NOCREATEDB NOCREATEROLE NOINHERIT;
  END IF;

  EXECUTE format('ALTER DATABASE %I OWNER TO tj_owner', current_database());
END
$$;

GRANT USAGE ON SCHEMA public TO tj_app;
