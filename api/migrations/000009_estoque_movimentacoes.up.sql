-- Razão de movimentações de estoque (estoque/02-movimentacoes, MOV-01 a
-- MOV-04; AD-009/ADR-007). Este script roda com o papel dono (tj_owner). O
-- papel da aplicação (tj_app) recebe INSERT e SELECT, nunca UPDATE nem
-- DELETE: toda movimentação é append-only, correção é sempre uma nova
-- movimentação de ajuste (EST-D-010), nunca edição.

DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'tj_app') THEN
    RAISE EXCEPTION 'papel tj_app não existe: crie os papéis do banco (docker/postgres/init/01-roles.sql) antes de migrar';
  END IF;
END
$$;

CREATE TABLE movimentacoes_estoque (
  id                  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  produto_id          uuid NOT NULL REFERENCES produtos_estoque,
  tipo                text NOT NULL CHECK (tipo IN ('ENTRADA', 'SAIDA', 'AJUSTE', 'DEVOLUCAO')),
  -- Já assinada pela aplicação (positiva ou negativa, nunca zero): o saldo é
  -- SUM(quantidade) GROUP BY produto_id, sem CASE por tipo (design/02).
  quantidade          bigint NOT NULL CHECK (quantidade <> 0),
  origem              text NOT NULL CHECK (origem IN ('VENDA', 'COMPRA', 'EVENTO', 'INVENTARIO', 'AJUSTE_MANUAL')),
  -- Só preenchido quando tipo = AJUSTE (EST-D-005/motivo obrigatório).
  motivo              text,
  -- Auto-referência unidirecional: só a devolução aponta para a
  -- movimentação original (ENTRADA ou SAIDA); a original nunca é escrita
  -- por causa disso (EST-D-008/EST-D-009).
  movimentacao_de_id  uuid REFERENCES movimentacoes_estoque,
  responsavel_id      uuid NOT NULL REFERENCES users,
  criado_em           timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX movimentacoes_estoque_produto_idx ON movimentacoes_estoque (produto_id);
CREATE INDEX movimentacoes_estoque_devolucao_idx ON movimentacoes_estoque (movimentacao_de_id) WHERE movimentacao_de_id IS NOT NULL;

GRANT INSERT, SELECT ON movimentacoes_estoque TO tj_app;
