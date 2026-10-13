"use client";

import { TriangleAlertIcon } from "lucide-react";

import { PageHeader } from "@/components/app/page-header";
import { RequirePermission } from "@/components/app/require-permission";
import { EmptyState, ErrorState, LoadingState } from "@/components/app/states";
import { type StatusDef, StatusBadge } from "@/components/app/status-badge";
import { SectionTitle } from "@/components/ui/typography";
import { cn } from "@/lib/utils";

import { AjusteDialog } from "./ajustes/ajuste-dialog";
import { estoqueCatalog, MESSAGE_PRODUTO_NAO_ENCONTRADO } from "./errors";
import { MovimentacaoDialog } from "./movimentacoes/movimentacao-dialog";
import { MovimentacoesPanel } from "./movimentacoes/movimentacoes-panel";
import { type Produto, useProduto, useSaldo } from "./produtos/hooks";

// Detalhe do produto (EWB-01 AC4, EWB-04 AC2). O produto vem do cache da
// lista; cada seção só é montada com a sua permissão de leitura, então sem
// ela a API correspondente não é chamada.

const PRODUTOS = { label: "Produtos", href: "/estoque/produtos" };

/** Saldo abaixo de zero: ícone e texto, além da cor (DSM-04 AC6). */
const SALDO_NEGATIVO: StatusDef = { label: "Saldo negativo", tone: "error", icon: TriangleAlertIcon };

export function ProdutoDetail({ id }: { id: string }) {
  return (
    <RequirePermission permission="estoque:produto:read" mode="page">
      <ProdutoDetailContent id={id} />
    </RequirePermission>
  );
}

function ProdutoDetailContent({ id }: { id: string }) {
  const query = useProduto(id);

  if (query.isPending) return <LoadingState shape="detail" />;
  if (query.isError) {
    return <ErrorState error={query.error} catalog={estoqueCatalog} onRetry={() => void query.refetch()} />;
  }
  if (!query.produto) {
    return (
      <>
        <PageHeader title="Produto" breadcrumb={[PRODUTOS, { label: "Produto" }]} />
        <EmptyState title={MESSAGE_PRODUTO_NAO_ENCONTRADO} description="Volte à lista de produtos e escolha outro." />
      </>
    );
  }
  return <ProdutoSections produto={query.produto} />;
}

function ProdutoSections({ produto }: { produto: Produto }) {
  return (
    <>
      <PageHeader
        title={produto.nome}
        breadcrumb={[PRODUTOS, { label: produto.nome }]}
        description={
          <dl className="flex flex-wrap gap-x-6 gap-y-1">
            <div className="flex gap-1">
              <dt>Código:</dt>
              <dd className="font-mono">{produto.codigo}</dd>
            </div>
            <div className="flex gap-1">
              <dt>Unidade de medida:</dt>
              <dd>{produto.unidade_medida}</dd>
            </div>
          </dl>
        }
        actions={
          <RequirePermission permission="estoque:movimentacao:create" mode="action">
            <MovimentacaoDialog produto={produto} tipo="SAIDA" />
            <MovimentacaoDialog produto={produto} tipo="ENTRADA" />
          </RequirePermission>
        }
      />

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
    <section className="flex flex-col gap-4" aria-labelledby="saldo-titulo">
      <SectionTitle id="saldo-titulo">Saldo</SectionTitle>
      {query.isPending ? (
        <LoadingState shape="value" />
      ) : query.isError ? (
        <ErrorState error={query.error} catalog={estoqueCatalog} onRetry={() => void query.refetch()} />
      ) : (
        <div className="flex flex-wrap items-center gap-3">
          <span
            data-testid="saldo-valor"
            data-negative={query.data < 0}
            className={cn("text-2xl font-semibold tabular-nums", query.data < 0 && "text-destructive")}
          >
            {`${query.data} ${produto.unidade_medida}`}
          </span>
          {query.data < 0 ? <StatusBadge status={SALDO_NEGATIVO} /> : null}
          {/* O ajuste fica aqui porque a confirmação mostra o saldo atual (EWB-03 AC3). */}
          <RequirePermission permission="estoque:movimentacao:adjust" mode="action">
            <AjusteDialog produto={produto} saldoAtual={query.data} />
          </RequirePermission>
        </div>
      )}
    </section>
  );
}
