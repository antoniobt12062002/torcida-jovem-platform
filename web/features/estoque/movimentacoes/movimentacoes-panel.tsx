"use client";

import { type Column, DataList } from "@/components/app/data-list";
import { authorLabel, formatDateTime, shortId } from "@/components/app/format";
import { RequirePermission } from "@/components/app/require-permission";
import { EmptyState, ErrorState, LoadingState } from "@/components/app/states";
import { SectionTitle } from "@/components/ui/typography";
import { useIsNarrow } from "@/lib/responsive";
import { useCan, useSession } from "@/lib/session/session-provider";

import { estoqueCatalog, origemLabels, tipoLabels } from "../errors";
import type { Produto } from "../produtos/hooks";
import { type Movimentacao, useMovimentacoes } from "./hooks";
import { MovimentacaoDialog, movimentacaoLabel } from "./movimentacao-dialog";
import { podeDevolver } from "./payload";

// Histórico de movimentações do produto (EWB-02 AC1), com "Devolver" nas
// entradas e saídas para quem pode registrar (EWB-02 AC3). Tabela a partir de
// 768 px; abaixo, lista empilhada com tipo, quantidade, data e responsável, e
// o motivo por último, cortado em três linhas (DSM-04 AC3).

const DASH = "—";

type Columns = Record<
  "tipo" | "quantidade" | "origem" | "motivo" | "devolucao" | "responsavel" | "data",
  Column<Movimentacao>
>;

function columnsFor(sessionUserId: string | undefined): Columns {
  return {
    tipo: { id: "tipo", header: "Tipo", priority: 1, inList: "title", cell: (m) => tipoLabels[m.tipo] },
    quantidade: { id: "quantidade", header: "Quantidade", priority: 1, numeric: true, cell: (m) => m.quantidade },
    origem: { id: "origem", header: "Origem", priority: 2, inList: false, cell: (m) => origemLabels[m.origem] },
    motivo: { id: "motivo", header: "Motivo", priority: 3, inList: "field", cell: (m) => m.motivo ?? DASH },
    devolucao: {
      id: "devolucao",
      header: "Devolução de",
      priority: 3,
      inList: false,
      cell: (m) => <span className="font-mono">{m.movimentacao_de_id ? shortId(m.movimentacao_de_id) : DASH}</span>,
    },
    responsavel: {
      id: "responsavel",
      header: "Responsável",
      priority: 2,
      inList: "field",
      cell: (m) => authorLabel(m.responsavel_id, sessionUserId),
    },
    data: { id: "data", header: "Data", priority: 1, inList: "field", wrap: "nowrap", cell: (m) => formatDateTime(m.criado_em) },
  };
}

export function MovimentacoesPanel({ produto }: { produto: Produto }) {
  const query = useMovimentacoes(produto.id);
  const { context } = useSession();
  const canCreate = useCan("estoque:movimentacao:create");
  const narrow = useIsNarrow();

  const c = columnsFor(context?.user.id);
  // Na tabela, a ordem de sempre; no item empilhado os campos entram nesta
  // ordem, e o motivo, por último, é o que o corte de três linhas encurta.
  const columns = narrow
    ? [c.tipo, c.quantidade, c.data, c.responsavel, c.motivo, c.origem, c.devolucao]
    : [c.tipo, c.quantidade, c.origem, c.motivo, c.devolucao, c.responsavel, c.data];
  const unidade = produto.unidade_medida;

  return (
    <section className="flex flex-col gap-4" aria-labelledby="movimentacoes-titulo">
      <SectionTitle id="movimentacoes-titulo">Movimentações</SectionTitle>
      {query.isPending ? (
        <LoadingState shape={narrow ? "list" : "table"} />
      ) : query.isError ? (
        <ErrorState error={query.error} catalog={estoqueCatalog} onRetry={() => void query.refetch()} />
      ) : query.data.length === 0 ? (
        <EmptyState headingLevel={3} title="Nenhuma movimentação registrada." />
      ) : (
        <DataList
          label="Movimentações"
          columns={columns}
          rows={query.data}
          rowKey={(m) => m.id}
          rowLabel={(m) => movimentacaoLabel(m, unidade)}
          primaryAction={
            canCreate
              ? (m) =>
                  podeDevolver(m) ? (
                    <RequirePermission permission="estoque:movimentacao:create" mode="action">
                      <MovimentacaoDialog produto={produto} tipo="DEVOLUCAO" referencia={m} />
                    </RequirePermission>
                  ) : null
              : undefined
          }
        />
      )}
    </section>
  );
}
