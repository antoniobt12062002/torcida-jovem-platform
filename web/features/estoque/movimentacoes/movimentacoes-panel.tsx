"use client";

import { authorLabel, formatDateTime, shortId } from "@/components/app/format";
import { RequirePermission } from "@/components/app/require-permission";
import { EmptyState, ErrorState, LoadingState } from "@/components/app/states";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { useCan, useSession } from "@/lib/session/session-provider";

import { estoqueCatalog, origemLabels, tipoLabels } from "../errors";
import type { Produto } from "../produtos/hooks";
import { useMovimentacoes } from "./hooks";
import { MovimentacaoDialog } from "./movimentacao-dialog";
import { podeDevolver } from "./payload";

// Histórico de movimentações do produto (EWB-02 AC1), com "Devolver" nas
// entradas e saídas para quem pode registrar (EWB-02 AC3).

const DASH = "—";

export function MovimentacoesPanel({ produto }: { produto: Produto }) {
  const query = useMovimentacoes(produto.id);
  const { context } = useSession();
  const canCreate = useCan("estoque:movimentacao:create");

  return (
    <section className="flex flex-col gap-2" aria-labelledby="movimentacoes-titulo">
      <h2 id="movimentacoes-titulo" className="text-base font-semibold">
        Movimentações
      </h2>
      {query.isPending ? (
        <LoadingState />
      ) : query.isError ? (
        <ErrorState error={query.error} catalog={estoqueCatalog} onRetry={() => void query.refetch()} />
      ) : query.data.length === 0 ? (
        <EmptyState title="Nenhuma movimentação registrada." />
      ) : (
        <Table aria-labelledby="movimentacoes-titulo">
          <TableHeader>
            <TableRow>
              <TableHead>Tipo</TableHead>
              <TableHead>Quantidade</TableHead>
              <TableHead>Origem</TableHead>
              <TableHead>Motivo</TableHead>
              <TableHead>Devolução de</TableHead>
              <TableHead>Responsável</TableHead>
              <TableHead>Data</TableHead>
              {canCreate ? <TableHead className="sr-only">Ações</TableHead> : null}
            </TableRow>
          </TableHeader>
          <TableBody>
            {query.data.map((m) => (
              <TableRow key={m.id}>
                <TableCell>{tipoLabels[m.tipo]}</TableCell>
                <TableCell className="tabular-nums">{m.quantidade}</TableCell>
                <TableCell>{origemLabels[m.origem]}</TableCell>
                <TableCell className="whitespace-normal">{m.motivo ?? DASH}</TableCell>
                <TableCell className="font-mono">
                  {m.movimentacao_de_id ? shortId(m.movimentacao_de_id) : DASH}
                </TableCell>
                <TableCell>{authorLabel(m.responsavel_id, context?.user.id)}</TableCell>
                <TableCell>{formatDateTime(m.criado_em)}</TableCell>
                {canCreate ? (
                  <TableCell>
                    {podeDevolver(m) ? (
                      <RequirePermission permission="estoque:movimentacao:create" mode="action">
                        <MovimentacaoDialog produto={produto} tipo="DEVOLUCAO" referencia={m} />
                      </RequirePermission>
                    ) : null}
                  </TableCell>
                ) : null}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </section>
  );
}
