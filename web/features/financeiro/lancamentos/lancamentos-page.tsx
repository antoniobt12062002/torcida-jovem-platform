"use client";

import Link from "next/link";
import { useState } from "react";

import { RequirePermission } from "@/components/app/require-permission";
import { EmptyState, ErrorState, LoadingState } from "@/components/app/states";
import { StatusBadge } from "@/components/app/status-badge";
import { formatDate } from "@/components/app/format";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { formatBRL } from "@/lib/money";

import { LANCAMENTO_PERMISSIONS } from "./actions";
import { lancamentosCatalog } from "./errors";
import { EMPTY_FILTERS, filterLancamentos, type LancamentoFilters } from "./filters";
import { useContaNome, useContasOptions, useCreateLancamento, useLancamentos } from "./hooks";
import { LancamentoFormDialog, NEW_LANCAMENTO } from "./lancamento-form";
import {
  FORMA_PAGAMENTO_LABEL,
  STATUS_LABEL,
  type StatusLancamento,
  STATUSES,
  TIPO_LABEL,
  type TipoConta,
  TIPOS,
} from "./labels";
import { SelectField } from "./select-field";

// Lista de lançamentos com filtros no cliente (FWB-03 AC1) e criação (AC2).

export function LancamentosPage() {
  const lancamentos = useLancamentos();
  const contas = useContasOptions();
  const contaNome = useContaNome();
  const [filters, setFilters] = useState<LancamentoFilters>(EMPTY_FILTERS);
  const [creating, setCreating] = useState(false);
  const create = useCreateLancamento();

  const items = lancamentos.data?.items ?? [];
  const visible = filterLancamentos(items, filters);
  const set = <K extends keyof LancamentoFilters>(key: K, value: LancamentoFilters[K]) =>
    setFilters((current) => ({ ...current, [key]: value }));

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-xl font-semibold">Lançamentos</h1>
        <RequirePermission permission={LANCAMENTO_PERMISSIONS.create} mode="action">
          <Button onClick={() => setCreating(true)}>Novo lançamento</Button>
          <LancamentoFormDialog
            open={creating}
            onOpenChange={setCreating}
            title="Novo lançamento"
            defaultValues={NEW_LANCAMENTO}
            tipoEditable
            submitLabel="Criar lançamento"
            successMessage="Lançamento criado."
            onSubmit={(values) => create.mutateAsync(values)}
          />
        </RequirePermission>
      </div>

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-6 lg:items-end">
        <SelectField
          label="Tipo"
          value={filters.tipo}
          onChange={(v) => set("tipo", v as TipoConta | "")}
          placeholder="Todos"
          options={TIPOS.map((t) => ({ value: t, label: TIPO_LABEL[t] }))}
        />
        <SelectField
          label="Status"
          value={filters.status}
          onChange={(v) => set("status", v as StatusLancamento | "")}
          placeholder="Todos"
          options={STATUSES.map((s) => ({ value: s, label: STATUS_LABEL[s] }))}
        />
        <SelectField
          label="Conta"
          value={filters.contaId}
          onChange={(v) => set("contaId", v)}
          placeholder="Todas"
          options={(contas.data?.items ?? []).map((c) => ({
            value: c.id,
            label: c.ativo ? c.nome : `${c.nome} (inativa)`,
          }))}
        />
        <DateFilter label="Criado de" value={filters.de} onChange={(v) => set("de", v)} />
        <DateFilter label="Criado até" value={filters.ate} onChange={(v) => set("ate", v)} />
        <Button variant="outline" onClick={() => setFilters(EMPTY_FILTERS)}>
          Limpar filtros
        </Button>
      </div>

      {lancamentos.isPending ? (
        <LoadingState />
      ) : lancamentos.isError ? (
        <ErrorState error={lancamentos.error} catalog={lancamentosCatalog} onRetry={() => void lancamentos.refetch()} />
      ) : items.length === 0 ? (
        <EmptyState title="Nenhum lançamento registrado." />
      ) : visible.length === 0 ? (
        <EmptyState title="Nenhum lançamento corresponde aos filtros." />
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Criado em</TableHead>
              <TableHead>Tipo</TableHead>
              <TableHead>Conta</TableHead>
              <TableHead className="text-right">Valor bruto</TableHead>
              <TableHead className="text-right">Taxa</TableHead>
              <TableHead className="text-right">Valor líquido</TableHead>
              <TableHead>Forma de pagamento</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>
                <span className="sr-only">Detalhes</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {visible.map((l) => (
              <TableRow key={l.id} data-testid="lancamento-row">
                <TableCell>{formatDate(l.criado_em)}</TableCell>
                <TableCell>{TIPO_LABEL[l.tipo]}</TableCell>
                <TableCell>{contaNome(l.conta_id)}</TableCell>
                <TableCell className="text-right">{formatBRL(l.valor_bruto_cents)}</TableCell>
                <TableCell className="text-right">{formatBRL(l.taxa_cents)}</TableCell>
                <TableCell className="text-right">{formatBRL(l.valor_liquido_cents)}</TableCell>
                <TableCell>{FORMA_PAGAMENTO_LABEL[l.forma_pagamento]}</TableCell>
                <TableCell>
                  <StatusBadge status={l.status} />
                </TableCell>
                <TableCell>
                  <Link
                    href={`/financeiro/lancamentos/${l.id}`}
                    className="text-sm underline underline-offset-4 hover:text-foreground"
                  >
                    Detalhes
                  </Link>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </div>
  );
}

function DateFilter({ label, value, onChange }: { label: string; value: string; onChange: (v: string) => void }) {
  const id = `filtro-${label.toLowerCase().replace(/\s+/g, "-")}`;
  return (
    <div className="flex flex-col gap-2">
      <label htmlFor={id} className="text-sm font-medium">
        {label}
      </label>
      <Input id={id} type="date" value={value} onChange={(event) => onChange(event.target.value)} />
    </div>
  );
}
