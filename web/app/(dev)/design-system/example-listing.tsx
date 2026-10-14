"use client";

import { EmptyState } from "@/components/app/states";
import { FilterBar } from "@/components/app/filter-bar";
import { Field, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { NativeSelect } from "@/components/ui/native-select";
import { codecs, SEARCH_DEBOUNCE_MS, useQueryState } from "@/lib/url-state";

import { LANCAMENTOS, STATUS_LANCAMENTO, type StatusLancamento, TIPO_LABEL, type Tipo } from "./data";
import { LancamentosList } from "./lancamentos-list";

// Barra de filtros e listagem juntas, como numa tela real (DSP-05): busca e
// dois filtros guardados na URL, filtrando os dados de exemplo no navegador.
// Quem usa fica dentro de `<Suspense>` (exigência de `useSearchParams`).

const TIPOS = ["RECEITA", "DESPESA"] as const satisfies readonly Tipo[];
const STATUS = ["CRIADA", "RECEBIDA", "PAGA", "CANCELADA"] as const satisfies readonly StatusLancamento[];

const searchCodec = codecs.string();
const tipoCodec = codecs.oneOf(TIPOS, "");
const statusCodec = codecs.oneOf(STATUS, "");

function normalize(text: string) {
  return text.normalize("NFD").replace(/\p{Diacritic}/gu, "").toLowerCase();
}

export function ExampleListing({ headingLevel = 3 }: { headingLevel?: 2 | 3 }) {
  const [busca, setBusca] = useQueryState("busca", searchCodec, { debounceMs: SEARCH_DEBOUNCE_MS });
  const [tipo, setTipo] = useQueryState("tipo", tipoCodec);
  const [status, setStatus] = useQueryState("status", statusCodec);

  const term = normalize(busca.trim());
  const rows = LANCAMENTOS.filter(
    (row) =>
      (term === "" || normalize(row.descricao).includes(term)) &&
      (tipo === "" || row.tipo === tipo) &&
      (status === "" || row.status === status),
  );
  const activeCount = [busca.trim(), tipo, status].filter((value) => value !== "").length;

  function clear() {
    setBusca("");
    setTipo("");
    setStatus("");
  }

  return (
    <div className="flex flex-col gap-4">
      <FilterBar
        activeCount={activeCount}
        resultCount={rows.length}
        resultNoun={["lançamento", "lançamentos"]}
        onClear={clear}
        search={
          <Field>
            <FieldLabel>Buscar</FieldLabel>
            <Input
              type="search"
              name="busca"
              autoComplete="off"
              value={busca} placeholder="Descrição…" onChange={(event) => setBusca(event.target.value)} />
          </Field>
        }
      >
        <Field>
          <FieldLabel>Tipo</FieldLabel>
          <NativeSelect name="tipo" autoComplete="off" value={tipo} onChange={(event) => setTipo(tipoCodec.parse(event.target.value))}>
            <option value="">Todos</option>
            {TIPOS.map((value) => (
              <option key={value} value={value}>
                {TIPO_LABEL[value]}
              </option>
            ))}
          </NativeSelect>
        </Field>
        <Field>
          <FieldLabel>Status</FieldLabel>
          <NativeSelect name="status" autoComplete="off" value={status} onChange={(event) => setStatus(statusCodec.parse(event.target.value))}>
            <option value="">Todos</option>
            {STATUS.map((value) => (
              <option key={value} value={value}>
                {STATUS_LANCAMENTO[value].label}
              </option>
            ))}
          </NativeSelect>
        </Field>
      </FilterBar>
      {rows.length > 0 ? (
        <LancamentosList label="Lançamentos" rows={rows} />
      ) : (
        <EmptyState kind="no-results" headingLevel={headingLevel} description="Mude a busca ou os filtros." onClear={clear} />
      )}
    </div>
  );
}
