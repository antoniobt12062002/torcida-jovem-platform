"use client";

import Link from "next/link";
import { type ReactNode, useState } from "react";

import { authorLabel, formatDateTime, shortId } from "@/components/app/format";
import { RequirePermission } from "@/components/app/require-permission";
import { EmptyState, ErrorState, LoadingState } from "@/components/app/states";
import { StatusBadge } from "@/components/app/status-badge";
import { Button } from "@/components/ui/button";
import { formatBRL } from "@/lib/money";
import { useSession } from "@/lib/session";

import { LANCAMENTO_PERMISSIONS, lancamentoActions } from "./actions";
import { lancamentosCatalog, MESSAGE_NAO_ENCONTRADO } from "./errors";
import { useContaNome, useLancamento, useUpdateLancamento } from "./hooks";
import { LancamentoFormDialog } from "./lancamento-form";
import { FORMA_PAGAMENTO_LABEL, type Lancamento, TIPO_LABEL } from "./labels";

// Detalhe de um lançamento, lido do cache da lista (FWB-03 AC6).

export function LancamentoDetail({ id }: { id: string }) {
  const query = useLancamento(id);

  return (
    <div className="flex flex-col gap-6">
      <Link
        href="/financeiro/lancamentos"
        className="text-sm underline underline-offset-4 hover:text-foreground"
      >
        Voltar para lançamentos
      </Link>
      {query.isPending ? (
        <LoadingState />
      ) : query.isError ? (
        <ErrorState error={query.error} catalog={lancamentosCatalog} onRetry={() => void query.refetch()} />
      ) : query.data === null ? (
        <EmptyState title={MESSAGE_NAO_ENCONTRADO} />
      ) : (
        <LancamentoView lancamento={query.data} />
      )}
    </div>
  );
}

function LancamentoView({ lancamento: l }: { lancamento: Lancamento }) {
  const { context } = useSession();
  const userId = context?.user.id;
  const contaNome = useContaNome();
  const actions = lancamentoActions(l, context?.permissions ?? []);

  return (
    <>
      <div className="flex flex-wrap items-center gap-3">
        <h1 className="text-xl font-semibold">
          {TIPO_LABEL[l.tipo]} {shortId(l.id)}
        </h1>
        <StatusBadge status={l.status} />
      </div>

      {actions.length > 0 ? (
        <div className="flex flex-wrap gap-2">
          {actions.includes("editar") ? (
            <RequirePermission permission={LANCAMENTO_PERMISSIONS.update} mode="action">
              <EditarAction lancamento={l} />
            </RequirePermission>
          ) : null}
        </div>
      ) : null}

      <dl className="grid gap-x-6 gap-y-3 sm:grid-cols-[max-content_1fr]">
        <Item label="Tipo">{TIPO_LABEL[l.tipo]}</Item>
        <Item label="Conta">{contaNome(l.conta_id)}</Item>
        <Item label="Valor bruto">{formatBRL(l.valor_bruto_cents)}</Item>
        <Item label="Taxa">{formatBRL(l.taxa_cents)}</Item>
        <Item label="Valor líquido">{formatBRL(l.valor_liquido_cents)}</Item>
        <Item label="Forma de pagamento">{FORMA_PAGAMENTO_LABEL[l.forma_pagamento]}</Item>
        <Item label="Criado por">{authorLabel(l.criado_por, userId)}</Item>
        <Item label="Criado em">{formatDateTime(l.criado_em)}</Item>
        <Item label="Atualizado em">{formatDateTime(l.atualizado_em)}</Item>
        {l.devolucao_de_id ? (
          <Item label="Devolução da receita">
            <Link
              href={`/financeiro/lancamentos/${l.devolucao_de_id}`}
              className="underline underline-offset-4 hover:text-foreground"
            >
              {shortId(l.devolucao_de_id)}
            </Link>
          </Item>
        ) : null}
        {l.status === "CANCELADA" ? (
          <>
            <Item label="Motivo do cancelamento">{l.motivo_cancelamento ?? "—"}</Item>
            <Item label="Cancelado por">{l.cancelado_por ? authorLabel(l.cancelado_por, userId) : "—"}</Item>
            <Item label="Cancelado em">{l.cancelado_em ? formatDateTime(l.cancelado_em) : "—"}</Item>
          </>
        ) : null}
      </dl>
    </>
  );
}

function Item({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="contents">
      <dt className="text-sm text-muted-foreground">{label}</dt>
      <dd className="text-sm">{children}</dd>
    </div>
  );
}

function EditarAction({ lancamento: l }: { lancamento: Lancamento }) {
  const [open, setOpen] = useState(false);
  const update = useUpdateLancamento(l.id);
  return (
    <>
      <Button variant="outline" onClick={() => setOpen(true)}>
        Editar
      </Button>
      <LancamentoFormDialog
        open={open}
        onOpenChange={setOpen}
        title="Editar lançamento"
        description="Altere a conta, os valores ou a forma de pagamento."
        defaultValues={{
          tipo: l.tipo,
          conta_id: l.conta_id,
          valor_bruto_cents: l.valor_bruto_cents,
          taxa_cents: l.taxa_cents,
          forma_pagamento: l.forma_pagamento,
        }}
        tipoEditable={false}
        submitLabel="Salvar alterações"
        successMessage="Lançamento atualizado."
        onSubmit={({ conta_id, valor_bruto_cents, taxa_cents, forma_pagamento }) =>
          update.mutateAsync({ conta_id, valor_bruto_cents, taxa_cents, forma_pagamento })
        }
      />
    </>
  );
}
