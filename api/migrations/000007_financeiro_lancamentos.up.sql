-- Lançamentos do financeiro (fundacao-financeiro, LAN-01 a LAN-04). Este
-- script roda com o papel dono (tj_owner). O papel da aplicação (tj_app)
-- recebe INSERT, SELECT e UPDATE (edição restrita a CRIADA, liquidação e
-- cancelamento mudam `status`), mas nunca DELETE: nenhum lançamento é
-- excluído.

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'tj_app') THEN
    RAISE EXCEPTION 'papel tj_app não existe: crie os papéis do banco (docker/postgres/init/01-roles.sql) antes de migrar';
  END IF;
END
$$;

CREATE TABLE lancamentos (
  id                   uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  tipo                 text NOT NULL CHECK (tipo IN ('RECEITA', 'DESPESA')),
  conta_id             uuid NOT NULL REFERENCES contas_contabeis,
  valor_bruto_cents    bigint NOT NULL CHECK (valor_bruto_cents > 0),
  taxa_cents           bigint NOT NULL DEFAULT 0 CHECK (taxa_cents >= 0),
  -- Calculado pela aplicação na criação (valor_bruto_cents - taxa_cents) e
  -- nunca recomputado depois — preserva o histórico mesmo que a lógica de
  -- cálculo evolua (design/02-lancamentos.md, Tech Decisions).
  valor_liquido_cents  bigint NOT NULL,
  forma_pagamento      text NOT NULL CHECK (forma_pagamento IN ('PIX', 'CARTAO', 'DINHEIRO', 'TRANSFERENCIA', 'OUTROS')),
  -- Transições de status (CRIADA→RECEBIDA/PAGA/CANCELADA) são regra de
  -- 03-workflow-e-saldo; esta migração só garante os 4 valores válidos e o
  -- estado inicial.
  status               text NOT NULL DEFAULT 'CRIADA' CHECK (status IN ('CRIADA', 'RECEBIDA', 'PAGA', 'CANCELADA')),
  -- Auto-referência unidirecional (FIN-D-003): só a devolução (despesa)
  -- aponta para a receita original; a receita nunca é escrita por causa
  -- disso.
  devolucao_de_id      uuid REFERENCES lancamentos,
  motivo_cancelamento  text,
  cancelado_por        uuid REFERENCES users,
  cancelado_em         timestamptz,
  criado_por           uuid NOT NULL REFERENCES users,
  criado_em            timestamptz NOT NULL DEFAULT now(),
  atualizado_em        timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX lancamentos_conta_idx ON lancamentos (conta_id);
CREATE INDEX lancamentos_devolucao_idx ON lancamentos (devolucao_de_id) WHERE devolucao_de_id IS NOT NULL;

GRANT INSERT, SELECT, UPDATE ON lancamentos TO tj_app;
