"use client";

import Link from "next/link";
import { useState } from "react";

import { type Column, DataList } from "@/components/app/data-list";
import { FilterBar } from "@/components/app/filter-bar";
import { PageHeader } from "@/components/app/page-header";
import { RequirePermission } from "@/components/app/require-permission";
import { EmptyState, ErrorState, LoadingState } from "@/components/app/states";
import { Field, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { ShowMore } from "@/components/ui/show-more";
import { codecs, SEARCH_DEBOUNCE_MS, useQueryState } from "@/lib/url-state";

import { estoqueCatalog } from "../errors";
import { CreateProdutoDialog } from "./create-produto-dialog";
import { type Produto, useProdutos } from "./hooks";
import { searchProdutos } from "./search";

// Lista de produtos (EWB-01 AC1-2). Sem saldo por linha (WEB-D-015). A busca
// fica na URL, em `busca`, com a escrita adiada (DSP-05 AC2 a AC4); a lista
// continua inteira no cliente (WEB-D-016) e aparece de 50 em 50.
// O produto não tem situação na API: a coluna "Situação" do design não existe.

const searchCodec = codecs.string();

/** Itens por vez na lista; "Mostrar mais" só aparece com mais do que isso. */
export const PRODUTOS_PAGE_SIZE = 50;

const CREATE = "estoque:produto:create";

export function ProdutosPage() {
  return (
    <RequirePermission permission="estoque:produto:read" mode="page">
      <ProdutosContent />
    </RequirePermission>
  );
}

function ProdutosContent() {
  const query = useProdutos();
  const [busca, setBusca] = useQueryState("busca", searchCodec, { debounceMs: SEARCH_DEBOUNCE_MS });
  // Sem nenhum produto, criar fica no estado vazio, sem repetir no cabeçalho.
  const empty = query.isSuccess && query.data.length === 0;
  const createAction = (
    <RequirePermission permission={CREATE} mode="action">
      <CreateProdutoDialog />
    </RequirePermission>
  );

  return (
    <>
      <PageHeader
        title="Produtos"
        description="Itens controlados pelo estoque da torcida."
        actions={empty ? undefined : createAction}
      />

      {query.isPending ? (
        <LoadingState shape="table" />
      ) : query.isError ? (
        <ErrorState error={query.error} catalog={estoqueCatalog} onRetry={() => void query.refetch()} />
      ) : empty ? (
        <EmptyState
          title="Nenhum produto cadastrado."
          description="Cadastre o primeiro produto para registrar entradas e saídas."
          action={createAction}
        />
      ) : (
        <ProdutosListing produtos={query.data} busca={busca} onBusca={setBusca} />
      )}
    </>
  );
}

type ListingProps = { produtos: Produto[]; busca: string; onBusca: (next: string) => void };

function ProdutosListing({ produtos, busca, onBusca }: ListingProps) {
  const rows = searchProdutos(produtos, busca);
  const clear = () => onBusca("");

  return (
    <div className="flex flex-col gap-4">
      <FilterBar
        activeCount={busca.trim() === "" ? 0 : 1}
        resultCount={rows.length}
        resultNoun={["produto", "produtos"]}
        onClear={clear}
        search={
          <Field>
            <FieldLabel>Buscar por código ou nome</FieldLabel>
            <Input type="search" value={busca} onChange={(event) => onBusca(event.target.value)} />
          </Field>
        }
      >
        {null}
      </FilterBar>
      {rows.length === 0 ? (
        <EmptyState
          kind="no-results"
          title="Nenhum produto encontrado para a busca."
          description="Confira o código ou o nome, ou limpe a busca."
          onClear={clear}
        />
      ) : (
        // Mudou a busca, a lista volta aos primeiros itens.
        <ProdutosList key={busca} rows={rows} />
      )}
    </div>
  );
}

const COLUMNS: readonly Column<Produto>[] = [
  {
    id: "codigo",
    header: "Código",
    priority: 1,
    inList: "field",
    wrap: "nowrap",
    cell: (p) => <span className="font-mono">{p.codigo}</span>,
  },
  {
    id: "nome",
    header: "Nome",
    priority: 1,
    inList: "title",
    cell: (p) => (
      <Link href={`/estoque/produtos/${p.id}`} className="font-medium underline-offset-4 hover:underline">
        {p.nome}
      </Link>
    ),
  },
  { id: "unidade", header: "Unidade de medida", priority: 2, inList: "field", cell: (p) => p.unidade_medida },
];

function ProdutosList({ rows }: { rows: Produto[] }) {
  const [shown, setShown] = useState(PRODUTOS_PAGE_SIZE);
  return (
    <>
      <DataList
        label="Produtos"
        columns={COLUMNS}
        rows={rows.slice(0, shown)}
        rowKey={(p) => p.id}
        rowLabel={(p) => `produto ${p.nome}`}
      />
      {rows.length > PRODUTOS_PAGE_SIZE ? (
        <ShowMore
          shown={Math.min(shown, rows.length)}
          total={rows.length}
          noun={["produto", "produtos"]}
          onShowMore={() => setShown((current) => current + PRODUTOS_PAGE_SIZE)}
        />
      ) : null}
    </>
  );
}
