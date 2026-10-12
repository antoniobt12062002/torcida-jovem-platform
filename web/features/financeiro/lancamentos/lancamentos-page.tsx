"use client";

import { EyeIcon, PlusIcon } from "lucide-react";
import { useState } from "react";

import { type Column, DataList } from "@/components/app/data-list";
import { FilterBar } from "@/components/app/filter-bar";
import { formatDate, formatIdentifier } from "@/components/app/format";
import { MoneyValue } from "@/components/app/money-value";
import { PageHeader } from "@/components/app/page-header";
import { RequirePermission } from "@/components/app/require-permission";
import { RowActionButton } from "@/components/app/row-actions";
import { EmptyState, ErrorState, LoadingState } from "@/components/app/states";
import { StatusBadge } from "@/components/app/status-badge";
import { Button } from "@/components/ui/button";
import { Field, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { NativeSelect } from "@/components/ui/native-select";
import { ShowMore } from "@/components/ui/show-more";
import { codecs, useQueryState } from "@/lib/url-state";

import { LANCAMENTO_PERMISSIONS } from "./actions";
import { lancamentosCatalog } from "./errors";
import { filterLancamentos, type LancamentoFilters } from "./filters";
import { useContaNome, useContasOptions, useCreateLancamento, useLancamentos } from "./hooks";
import { LancamentoFormDialog, NEW_LANCAMENTO } from "./lancamento-form";
import {
  type Conta,
  FORMA_PAGAMENTO_LABEL,
  type Lancamento,
  STATUS_DEF,
  STATUSES,
  TIPO_LABEL,
  TIPOS,
} from "./labels";

// Lista de lançamentos com filtros no cliente (FWB-03 AC1) e criação (AC2).
// Os filtros vivem na URL (DSP-05 AC2 a AC4): um link copiado reproduz a
// consulta. A lista mostra 50 por vez (DSP-05 AC12), e o filtro considera
// todos os registros carregados (AC13).

/** Lote do "Mostrar mais" (DSP-05 AC12). */
export const PAGE_SIZE = 50;

const tipoCodec = codecs.oneOf(TIPOS, "");
const statusCodec = codecs.oneOf(STATUSES, "");
const contaCodec = codecs.string();
const dateCodec = codecs.isoDate();

const NOUN = ["lançamento", "lançamentos"] as const;

/** Identifica o lançamento nos nomes das ações de linha (DSP-05 AC10). */
export function lancamentoRowLabel(l: Pick<Lancamento, "id" | "tipo">): string {
  return `lançamento ${formatIdentifier(TIPO_LABEL[l.tipo], l.id).label}`;
}

/**
 * Conta do filtro: uma conta que não existe mais na lista de contas é ignorada
 * (caso de borda de DSP-05). Sem a lista carregada, vale o que veio da URL.
 */
function existingConta(contaId: string, contas: readonly Conta[] | undefined): string {
  if (contaId === "" || contas === undefined) return contaId;
  return contas.some((c) => c.id === contaId) ? contaId : "";
}

function useFilters(contas: readonly Conta[] | undefined) {
  const [tipo, setTipo] = useQueryState("tipo", tipoCodec);
  const [status, setStatus] = useQueryState("status", statusCodec);
  const [conta, setConta] = useQueryState("conta", contaCodec);
  const [de, setDe] = useQueryState("de", dateCodec);
  const [ate, setAte] = useQueryState("ate", dateCodec);
  const filters: LancamentoFilters = { tipo, status, contaId: existingConta(conta, contas), de, ate };
  const activeCount = Object.values(filters).filter((value) => value !== "").length;
  const clear = () => {
    setTipo("");
    setStatus("");
    setConta("");
    setDe("");
    setAte("");
  };
  return { filters, activeCount, clear, setTipo, setStatus, setConta, setDe, setAte };
}

export function LancamentosPage() {
  const lancamentos = useLancamentos();
  const contas = useContasOptions();
  const contaNome = useContaNome();
  const [creating, setCreating] = useState(false);
  const create = useCreateLancamento();
  const { filters, activeCount, clear, setTipo, setStatus, setConta, setDe, setAte } = useFilters(contas.data?.items);

  const items = lancamentos.data?.items ?? [];
  const visible = filterLancamentos(items, filters);

  // Cada mudança de filtro volta ao primeiro lote.
  const [limit, setLimit] = useState(PAGE_SIZE);
  const filtersKey = JSON.stringify(filters);
  const [limitKey, setLimitKey] = useState(filtersKey);
  if (limitKey !== filtersKey) {
    setLimitKey(filtersKey);
    setLimit(PAGE_SIZE);
  }
  const shown = visible.slice(0, limit);

  const columns: Column<Lancamento>[] = [
    {
      id: "data",
      header: "Criado em",
      priority: 1,
      inList: "field",
      wrap: "nowrap",
      cell: (l) => formatDate(l.criado_em),
    },
    { id: "tipo", header: "Tipo", priority: 2, inList: "field", cell: (l) => TIPO_LABEL[l.tipo] },
    { id: "conta", header: "Conta", priority: 1, inList: "title", cell: (l) => contaNome(l.conta_id) },
    {
      id: "bruto",
      header: "Valor bruto",
      numeric: true,
      priority: 3,
      inList: false,
      cell: (l) => <MoneyValue cents={l.valor_bruto_cents} />,
    },
    {
      id: "taxa",
      header: "Taxa",
      numeric: true,
      priority: 3,
      inList: false,
      cell: (l) => <MoneyValue cents={l.taxa_cents} />,
    },
    {
      id: "liquido",
      header: "Valor líquido",
      numeric: true,
      priority: 1,
      // Cancelado: o valor fica riscado (mapa de status do design).
      cell: (l) => (
        <MoneyValue cents={l.valor_liquido_cents} className={l.status === "CANCELADA" ? "line-through" : undefined} />
      ),
    },
    {
      id: "forma",
      header: "Forma de pagamento",
      priority: 2,
      inList: false,
      cell: (l) => FORMA_PAGAMENTO_LABEL[l.forma_pagamento],
    },
    {
      id: "status",
      header: "Status",
      priority: 1,
      inList: "meta",
      cell: (l) => <StatusBadge status={STATUS_DEF[l.status]} />,
    },
  ];

  return (
    <>
      <PageHeader
        title="Lançamentos"
        description="Receitas e despesas da torcida."
        actions={
          <RequirePermission permission={LANCAMENTO_PERMISSIONS.create} mode="action">
            <Button onClick={() => setCreating(true)}>
              <PlusIcon aria-hidden="true" />
              Novo lançamento
            </Button>
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
        }
      />

      {lancamentos.isPending ? (
        <LoadingState shape="table" />
      ) : lancamentos.isError ? (
        <ErrorState error={lancamentos.error} catalog={lancamentosCatalog} onRetry={() => void lancamentos.refetch()} />
      ) : items.length === 0 ? (
        <EmptyState title="Nenhum lançamento registrado." description="Os lançamentos criados aparecem aqui." />
      ) : (
        <div className="flex flex-col gap-4">
          <FilterBar activeCount={activeCount} resultCount={visible.length} resultNoun={NOUN} onClear={clear}>
            <Field>
              <FieldLabel>Tipo</FieldLabel>
              <NativeSelect value={filters.tipo} onChange={(event) => setTipo(tipoCodec.parse(event.target.value))}>
                <option value="">Todos</option>
                {TIPOS.map((t) => (
                  <option key={t} value={t}>
                    {TIPO_LABEL[t]}
                  </option>
                ))}
              </NativeSelect>
            </Field>
            <Field>
              <FieldLabel>Status</FieldLabel>
              <NativeSelect
                value={filters.status}
                onChange={(event) => setStatus(statusCodec.parse(event.target.value))}
              >
                <option value="">Todos</option>
                {STATUSES.map((s) => (
                  <option key={s} value={s}>
                    {STATUS_DEF[s].label}
                  </option>
                ))}
              </NativeSelect>
            </Field>
            <Field>
              <FieldLabel>Conta</FieldLabel>
              <NativeSelect value={filters.contaId} onChange={(event) => setConta(event.target.value)}>
                <option value="">Todas</option>
                {(contas.data?.items ?? []).map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.ativo ? c.nome : `${c.nome} (inativa)`}
                  </option>
                ))}
              </NativeSelect>
            </Field>
            <Field>
              <FieldLabel>Criado de</FieldLabel>
              <Input type="date" value={filters.de} onChange={(event) => setDe(dateCodec.parse(event.target.value))} />
            </Field>
            <Field>
              <FieldLabel>Criado até</FieldLabel>
              <Input type="date" value={filters.ate} onChange={(event) => setAte(dateCodec.parse(event.target.value))} />
            </Field>
          </FilterBar>

          {visible.length === 0 ? (
            <EmptyState
              kind="no-results"
              title="Nenhum lançamento corresponde aos filtros."
              description="Mude ou limpe os filtros para ver outros lançamentos."
              onClear={clear}
            />
          ) : (
            <>
              <DataList
                label="Lançamentos"
                columns={columns}
                rows={shown}
                rowKey={(l) => l.id}
                rowLabel={lancamentoRowLabel}
                primaryAction={(l) => (
                  <RowActionButton
                    label="Detalhes"
                    rowLabel={lancamentoRowLabel(l)}
                    icon={EyeIcon}
                    href={`/financeiro/lancamentos/${l.id}`}
                  />
                )}
              />
              {visible.length > PAGE_SIZE ? (
                <ShowMore
                  shown={shown.length}
                  total={visible.length}
                  noun={NOUN}
                  onShowMore={() => setLimit((current) => current + PAGE_SIZE)}
                />
              ) : null}
            </>
          )}
        </div>
      )}
    </>
  );
}
