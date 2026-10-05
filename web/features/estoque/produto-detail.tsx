"use client";

import Link from "next/link";

import { RequirePermission } from "@/components/app/require-permission";
import { ErrorState, LoadingState } from "@/components/app/states";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";

import { estoqueCatalog, MESSAGE_PRODUTO_NAO_ENCONTRADO } from "./errors";
import { MovimentacaoDialog } from "./movimentacoes/movimentacao-dialog";
import { MovimentacoesPanel } from "./movimentacoes/movimentacoes-panel";
import { type Produto, useProduto, useSaldo } from "./produtos/hooks";

// Detalhe do produto (EWB-01 AC4, EWB-04 AC2). O produto vem do cache da
// lista; cada seção só é montada com a sua permissão de leitura, então sem
// ela a API correspondente não é chamada.

export function ProdutoDetail({ id }: { id: string }) {
  return (
    <RequirePermission permission="estoque:produto:read" mode="page">
      <ProdutoDetailContent id={id} />
    </RequirePermission>
  );
}

function ProdutoDetailContent({ id }: { id: string }) {
  const query = useProduto(id);

  if (query.isPending) return <LoadingState />;
  if (query.isError) {
    return <ErrorState error={query.error} catalog={estoqueCatalog} onRetry={() => void query.refetch()} />;
  }

  return (
    <div className="flex flex-col gap-6">
      <Link href="/estoque/produtos" className="text-sm text-muted-foreground underline-offset-4 hover:underline">
        Voltar para produtos
      </Link>
      {query.produto ? (
        <ProdutoSections produto={query.produto} />
      ) : (
        <p className="text-sm">{MESSAGE_PRODUTO_NAO_ENCONTRADO}</p>
      )}
    </div>
  );
}

function ProdutoSections({ produto }: { produto: Produto }) {
  return (
    <>
      <header className="flex flex-col gap-1">
        <h1 className="text-xl font-semibold">{produto.nome}</h1>
        <dl className="flex flex-wrap gap-x-6 gap-y-1 text-sm text-muted-foreground">
          <div className="flex gap-1">
            <dt>Código:</dt>
            <dd className="font-mono">{produto.codigo}</dd>
          </div>
          <div className="flex gap-1">
            <dt>Unidade de medida:</dt>
            <dd>{produto.unidade_medida}</dd>
          </div>
        </dl>
      </header>

      <RequirePermission permission="estoque:movimentacao:create" mode="action">
        <div className="flex flex-wrap gap-2">
          <MovimentacaoDialog produto={produto} tipo="ENTRADA" />
          <MovimentacaoDialog produto={produto} tipo="SAIDA" />
        </div>
      </RequirePermission>

      <RequirePermission permission="estoque:saldo:read" mode="action">
        <SaldoSection produto={produto} />
      </RequirePermission>

      <RequirePermission permission="estoque:movimentacao:read" mode="action">
        <MovimentacoesPanel produto={produto} />
      </RequirePermission>
    </>
  );
}

function SaldoSection({ produto }: { produto: Produto }) {
  const query = useSaldo(produto.id);

  return (
    <section className="flex flex-col gap-2" aria-labelledby="saldo-titulo">
      <h2 id="saldo-titulo" className="text-base font-semibold">
        Saldo
      </h2>
      {query.isPending ? (
        <LoadingState />
      ) : query.isError ? (
        <ErrorState error={query.error} catalog={estoqueCatalog} onRetry={() => void query.refetch()} />
      ) : (
        <div className="flex items-center gap-3">
          <span
            data-testid="saldo-valor"
            data-negative={query.data < 0}
            className={cn("text-2xl font-semibold tabular-nums", query.data < 0 && "text-destructive")}
          >
            {`${query.data} ${produto.unidade_medida}`}
          </span>
          {query.data < 0 ? <Badge variant="destructive">Saldo negativo</Badge> : null}
        </div>
      )}
    </section>
  );
}
