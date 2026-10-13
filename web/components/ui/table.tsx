"use client"

import * as React from "react"
import { cn } from "@/lib/utils"

type TableScroll =
  /** Padrão. Sem contêiner de rolagem: a forma usada pela lista responsiva. */
  | "none"
  /** Rolagem horizontal explícita, só com justificativa na spec da tela. */
  | "scrollable"

// A tabela precisa de nome acessível: `aria-label` ou `aria-labelledby`.
function Table({
  className,
  scroll = "none",
  ...props
}: React.ComponentProps<"table"> & { scroll?: TableScroll }) {
  const table = (
    <table
      data-slot="table"
      className={cn(
        "w-full caption-bottom text-sm",
        scroll === "scrollable" &&
          "[&_tr>*:first-child]:sticky [&_tr>*:first-child]:left-0 [&_tbody_tr>*:first-child]:bg-background [&_thead_tr>*:first-child]:bg-muted",
        className
      )}
      {...props}
    />
  )
  if (scroll === "none") return table
  // Região rolável: alcançável pelo teclado e com o mesmo nome da tabela.
  return (
    <div
      data-slot="table-container"
      role="region"
      tabIndex={0}
      aria-label={props["aria-label"]}
      aria-labelledby={props["aria-labelledby"]}
      className="relative w-full overflow-x-auto overscroll-x-contain rounded-md outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-ring-offset"
    >
      {table}
    </div>
  )
}

function TableHeader({ className, ...props }: React.ComponentProps<"thead">) {
  return (
    <thead
      data-slot="table-header"
      className={cn(
        "bg-muted [&_tr]:border-b-2 [&_tr]:border-border-strong [&_tr]:hover:bg-muted",
        className
      )}
      {...props}
    />
  )
}

function TableBody({ className, ...props }: React.ComponentProps<"tbody">) {
  return (
    <tbody
      data-slot="table-body"
      className={cn("[&_tr:last-child]:border-0", className)}
      {...props}
    />
  )
}

function TableFooter({ className, ...props }: React.ComponentProps<"tfoot">) {
  return (
    <tfoot
      data-slot="table-footer"
      className={cn(
        "border-t bg-muted font-medium [&>tr]:last:border-b-0",
        className
      )}
      {...props}
    />
  )
}

function TableRow({ className, ...props }: React.ComponentProps<"tr">) {
  return (
    <tr
      data-slot="table-row"
      className={cn(
        "border-b transition-colors duration-fast hover:bg-accent has-aria-expanded:bg-accent data-[state=selected]:bg-muted",
        className
      )}
      {...props}
    />
  )
}

// Coluna numérica: à direita, com algarismos tabulares.
const NUMERIC = "text-right tabular-nums"

function TableHead({
  className,
  numeric,
  scope = "col",
  ...props
}: React.ComponentProps<"th"> & { numeric?: boolean }) {
  return (
    <th
      data-slot="table-head"
      scope={scope}
      className={cn(
        "h-table-head px-3 text-left align-middle text-xs font-semibold whitespace-nowrap text-muted-foreground [&:has([role=checkbox])]:pr-0",
        numeric && NUMERIC,
        className
      )}
      {...props}
    />
  )
}

// A coluna declara como trata texto longo: quebra (padrão), `nowrap` ou
// `truncate`, que corta com reticências sem alargar a tabela.
function TableCell({
  className,
  numeric,
  wrap,
  ...props
}: React.ComponentProps<"td"> & {
  numeric?: boolean
  wrap?: "nowrap" | "truncate"
}) {
  return (
    <td
      data-slot="table-cell"
      className={cn(
        "h-table-row px-3 py-2 align-middle [&:has([role=checkbox])]:pr-0",
        numeric && NUMERIC,
        wrap === "nowrap" && "whitespace-nowrap",
        wrap === "truncate" && "max-w-0 truncate",
        className
      )}
      {...props}
    />
  )
}

function TableCaption({
  className,
  ...props
}: React.ComponentProps<"caption">) {
  return (
    <caption
      data-slot="table-caption"
      className={cn("mt-4 text-sm text-muted-foreground", className)}
      {...props}
    />
  )
}

export {
  Table,
  TableHeader,
  TableBody,
  TableFooter,
  TableHead,
  TableRow,
  TableCell,
  TableCaption,
}
