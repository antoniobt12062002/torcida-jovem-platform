"use client";

import type { ReactNode } from "react";

import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { useIsNarrow } from "@/lib/responsive";
import { cn } from "@/lib/utils";

import { type RowAction, RowActions } from "./row-actions";

// Listagem responsiva (DSP-05 AC5 a AC11; DD-2, WDS-D-026). Uma definição de
// colunas serve a dois layouts, e só um deles é renderizado por vez: tabela a
// partir de 768 px, lista empilhada abaixo disso. A escolha entre os dois é a
// única consulta de mídia em JavaScript; a prioridade de colunas é só CSS.
// No servidor e na primeira renderização da hidratação vale a tabela.

export type Column<T> = {
  id: string;
  header: string;
  cell: (row: T) => ReactNode;
  /** À direita, com algarismos tabulares. No item empilhado, é o valor em destaque. */
  numeric?: boolean;
  /** 1 e 2 aparecem sempre na tabela; 3 só a partir de 1024 px (o dado continua no detalhe). */
  priority: 1 | 2 | 3;
  /** Tratamento de texto longo na tabela. Padrão: quebra de linha. */
  wrap?: "nowrap" | "truncate" | "wrap";
  /**
   * Papel no item empilhado: título, campo-chave (entram os três primeiros) ou
   * `meta` (status). `false` deixa a coluna fora do item. Sem valor: a coluna
   * numérica é o valor do item e a primeira coluna de texto é o título, se
   * nenhuma outra for marcada como tal; as demais ficam fora.
   */
  inList?: "title" | "field" | "meta" | false;
};

export type DataListProps<T> = {
  /** Nome acessível da tabela ou da lista. */
  label: string;
  columns: readonly Column<T>[];
  rows: readonly T[];
  rowKey: (row: T) => string;
  /** Identifica a linha; usado nos nomes das ações. */
  rowLabel: (row: T) => string;
  /** A ação visível da linha; use `RowActionButton`, que nomeia a linha. */
  primaryAction?: (row: T) => ReactNode;
  /** As demais ações, reunidas no menu "Ações de {rowLabel}". */
  moreActions?: (row: T) => readonly RowAction[];
  /** `table` mantém a tabela em qualquer largura (até quatro colunas curtas, DSP-05 AC8). */
  layout?: "auto" | "table" | "list";
  className?: string;
};

const MAX_FIELDS = 3;

// Prioridade 3: fora da tabela entre 768 e 1023 px.
const LOW_PRIORITY = "hidden lg:table-cell";

export function DataList<T>({ layout = "auto", ...props }: DataListProps<T>) {
  const narrow = useIsNarrow();
  const asList = layout === "list" || (layout === "auto" && narrow);
  return asList ? <StackedList {...props} /> : <DataTable {...props} />;
}

type LayoutProps<T> = Omit<DataListProps<T>, "layout">;

function actionsOf<T>(row: T, { rowLabel, primaryAction, moreActions }: LayoutProps<T>) {
  return <RowActions rowLabel={rowLabel(row)} primary={primaryAction?.(row)} actions={moreActions?.(row)} />;
}

function DataTable<T>(props: LayoutProps<T>) {
  const { label, columns, rows, rowKey, primaryAction, moreActions, className } = props;
  const hasActions = !!primaryAction || !!moreActions;
  return (
    <Table scroll="none" aria-label={label} className={className}>
      <TableHeader>
        <TableRow>
          {columns.map((column) => (
            <TableHead key={column.id} numeric={column.numeric} className={cn(column.priority === 3 && LOW_PRIORITY)}>
              {column.header}
            </TableHead>
          ))}
          {hasActions ? (
            <TableHead>
              <span className="sr-only">Ações</span>
            </TableHead>
          ) : null}
        </TableRow>
      </TableHeader>
      <TableBody>
        {rows.map((row) => (
          <TableRow key={rowKey(row)}>
            {columns.map((column) => (
              <TableCell
                key={column.id}
                numeric={column.numeric}
                wrap={column.wrap === "wrap" ? undefined : column.wrap}
                className={cn(column.priority === 3 && LOW_PRIORITY)}
              >
                {column.cell(row)}
              </TableCell>
            ))}
            {hasActions ? <TableCell wrap="nowrap">{actionsOf(row, props)}</TableCell> : null}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}

/** Distribui as colunas pelos papéis do item empilhado. */
function listRoles<T>(columns: readonly Column<T>[]) {
  const title =
    columns.find((column) => column.inList === "title") ??
    columns.find((column) => column.inList === undefined && !column.numeric);
  return {
    title,
    meta: columns.filter((column) => column.inList === "meta"),
    fields: columns.filter((column) => column.inList === "field").slice(0, MAX_FIELDS),
    value: columns.find((column) => column.numeric && column.inList === undefined),
  };
}

function StackedList<T>(props: LayoutProps<T>) {
  const { label, columns, rows, rowKey, primaryAction, moreActions, className } = props;
  const { title, meta, fields, value } = listRoles(columns);
  const hasActions = !!primaryAction || !!moreActions;
  return (
    <ul aria-label={label} data-slot="data-list" className={cn("flex flex-col divide-y border-y", className)}>
      {rows.map((row) => (
        <li key={rowKey(row)} className="flex flex-col gap-1 py-3">
          <div className="flex items-start justify-between gap-3">
            {/* Quando o título é um link, ele cobre a linha do título. */}
            <div data-slot="data-list-title" className="line-clamp-3 min-w-0 font-semibold wrap-break-word [&>a]:block">
              {title?.cell(row)}
            </div>
            {meta.length > 0 ? (
              <div data-slot="data-list-meta" className="flex shrink-0 flex-wrap justify-end gap-2">
                {meta.map((column) => (
                  <span key={column.id}>{column.cell(row)}</span>
                ))}
              </div>
            ) : null}
          </div>
          {fields.length > 0 ? (
            <div data-slot="data-list-fields" className="line-clamp-3 text-sm wrap-break-word text-muted-foreground">
              {fields.map((column, index) => (
                <span key={column.id}>
                  {index > 0 ? <span aria-hidden> · </span> : null}
                  <span data-slot="data-list-field">
                    <span className="sr-only">{column.header}: </span>
                    {column.cell(row)}
                  </span>
                </span>
              ))}
            </div>
          ) : null}
          {value || hasActions ? (
            <div className="flex items-center justify-between gap-3">
              <div data-slot="data-list-value" className="min-w-0 tabular-nums">
                {value ? (
                  <>
                    <span className="sr-only">{value.header}: </span>
                    {value.cell(row)}
                  </>
                ) : null}
              </div>
              {hasActions ? actionsOf(row, props) : null}
            </div>
          ) : null}
        </li>
      ))}
    </ul>
  );
}
