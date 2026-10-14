"use client";

import { BanIcon, EyeIcon, PaperclipIcon } from "lucide-react";

import { type Column, DataList, type DataListProps } from "@/components/app/data-list";
import { MoneyValue } from "@/components/app/money-value";
import { RowActionButton } from "@/components/app/row-actions";
import { StatusBadge } from "@/components/app/status-badge";
import { toast } from "@/components/ui/toast";

import { type Lancamento, STATUS_LANCAMENTO, TIPO_LABEL } from "./data";

// Listagem de lançamentos de exemplo: uma definição de colunas para os três
// comportamentos do `DataList`. As ações só mostram um aviso: nada sai da página.

export const LANCAMENTO_COLUMNS: readonly Column<Lancamento>[] = [
  { id: "descricao", header: "Descrição", priority: 1, inList: "title", cell: (row) => row.descricao },
  { id: "tipo", header: "Tipo", priority: 2, inList: "field", cell: (row) => TIPO_LABEL[row.tipo] },
  { id: "conta", header: "Conta", priority: 3, inList: "field", cell: (row) => row.conta },
  { id: "data", header: "Data", priority: 2, inList: "field", wrap: "nowrap", cell: (row) => row.data },
  { id: "status", header: "Status", priority: 1, inList: "meta", cell: (row) => <StatusBadge status={STATUS_LANCAMENTO[row.status]} /> },
  {
    id: "valor",
    header: "Valor",
    numeric: true,
    priority: 1,
    // Cancelada: o valor fica riscado (design/02-componentes.md, mapa de status).
    cell: (row) => <MoneyValue cents={row.valorCents} className={row.status === "CANCELADA" ? "line-through" : undefined} />,
  },
];

export function rowLabelOf(row: Lancamento) {
  return `lançamento ${row.descricao}`;
}

export function LancamentosList({
  rows,
  layout,
  label,
}: {
  rows: readonly Lancamento[];
  layout?: DataListProps<Lancamento>["layout"];
  label: string;
}) {
  return (
    <DataList
      label={label}
      layout={layout}
      columns={LANCAMENTO_COLUMNS}
      rows={rows}
      rowKey={(row) => row.id}
      rowLabel={rowLabelOf}
      primaryAction={(row) => (
        <RowActionButton
          label="Detalhes"
          rowLabel={rowLabelOf(row)}
          icon={EyeIcon}
          onClick={() => toast.show(`Exemplo: detalhes de ${row.descricao}.`)}
        />
      )}
      moreActions={(row) => [
        {
          label: "Anexar comprovante",
          icon: PaperclipIcon,
          onSelect: () => toast.show("Exemplo: nenhum arquivo é enviado."),
        },
        {
          label: "Cancelar lançamento",
          icon: BanIcon,
          destructive: true,
          disabled: row.status === "CANCELADA",
          onSelect: () => toast.warning("Exemplo: a confirmação real fica na seção Feedback."),
        },
      ]}
    />
  );
}
