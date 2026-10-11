"use client";

import Link from "next/link";
import { useState } from "react";

import { RequirePermission } from "@/components/app/require-permission";
import { EmptyState, ErrorState, LoadingState } from "@/components/app/states";
import { Field, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";

import { estoqueCatalog } from "../errors";
import { CreateProdutoDialog } from "./create-produto-dialog";
import { type Produto, useProdutos } from "./hooks";
import { searchProdutos } from "./search";

// Lista de produtos (EWB-01 AC1-2). Sem saldo por linha (WEB-D-015).

export function ProdutosPage() {
  return (
    <RequirePermission permission="estoque:produto:read" mode="page">
      <ProdutosContent />
    </RequirePermission>
  );
}

function ProdutosContent() {
  const query = useProdutos();
  const [term, setTerm] = useState("");

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-xl font-semibold">Produtos</h1>
        <RequirePermission permission="estoque:produto:create" mode="action">
          <CreateProdutoDialog />
        </RequirePermission>
      </div>

      {query.isPending ? (
        <LoadingState />
      ) : query.isError ? (
        <ErrorState error={query.error} catalog={estoqueCatalog} onRetry={() => void query.refetch()} />
      ) : query.data.length === 0 ? (
        <EmptyState title="Nenhum produto cadastrado." />
      ) : (
        <>
          <Field className="max-w-sm">
            <FieldLabel>Buscar por código ou nome</FieldLabel>
            <Input type="search" value={term} onChange={(event) => setTerm(event.target.value)} />
          </Field>
          <ProdutosTable produtos={searchProdutos(query.data, term)} />
        </>
      )}
    </div>
  );
}

function ProdutosTable({ produtos }: { produtos: Produto[] }) {
  if (produtos.length === 0) return <EmptyState title="Nenhum produto encontrado para a busca." />;
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Código</TableHead>
          <TableHead>Nome</TableHead>
          <TableHead>Unidade de medida</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {produtos.map((p) => (
          <TableRow key={p.id}>
            <TableCell className="font-mono">{p.codigo}</TableCell>
            <TableCell>
              <Link href={`/estoque/produtos/${p.id}`} className="underline-offset-4 hover:underline">
                {p.nome}
              </Link>
            </TableCell>
            <TableCell>{p.unidade_medida}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
