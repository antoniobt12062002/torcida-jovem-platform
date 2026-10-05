"use client";

import Link from "next/link";
import { type ReactNode, useState } from "react";

import { ApiErrorAlert } from "@/components/app/api-error";
import { ConfirmDialog } from "@/components/app/confirm-dialog";
import { authorLabel, formatDateTime, shortId } from "@/components/app/format";
import { RequirePermission } from "@/components/app/require-permission";
import { EmptyState, ErrorState, LoadingState } from "@/components/app/states";
import { StatusBadge } from "@/components/app/status-badge";
import { Button } from "@/components/ui/button";
import { toast } from "@/components/ui/toast";
import { type ApiError, isApiError, networkError } from "@/lib/api/problem";
import { formatBRL } from "@/lib/money";
import { useSession } from "@/lib/session";

import { ComprovantesPanel } from "@/features/financeiro/comprovantes/comprovantes-panel";

import { ACTION_PERMISSION, lancamentoActions } from "./actions";
import { DevolucaoAction } from "./devolucao-form";
import { lancamentosCatalog, MESSAGE_NAO_ENCONTRADO } from "./errors";
import {
  useCancelarLancamento,
  useContaNome,
  useLancamento,
  usePagarLancamento,
  useReceberLancamento,
  useUpdateLancamento,
} from "./hooks";
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
  const [actionError, setActionError] = useState<ApiError | null>(null);
  const report = (error: unknown) => setActionError(isApiError(error) ? error : networkError(error));

  return (
    <>
      <div className="flex flex-wrap items-center gap-3">
        <h1 className="text-xl font-semibold">
          {TIPO_LABEL[l.tipo]} {shortId(l.id)}
        </h1>
        <StatusBadge status={l.status} />
      </div>

      {actionError ? <ApiErrorAlert error={actionError} catalog={lancamentosCatalog} /> : null}

      {actions.length > 0 ? (
        <div className="flex flex-wrap gap-2">
          {actions.map((action) => (
            <RequirePermission key={action} permission={ACTION_PERMISSION[action]} mode="action">
              {action === "editar" ? <EditarAction lancamento={l} /> : null}
              {action === "receber" ? (
                <StatusAction kind="receber" lancamento={l} onStart={() => setActionError(null)} onError={report} />
              ) : null}
              {action === "pagar" ? (
                <StatusAction kind="pagar" lancamento={l} onStart={() => setActionError(null)} onError={report} />
              ) : null}
              {action === "cancelar" ? (
                <CancelarAction lancamento={l} onStart={() => setActionError(null)} onError={report} />
              ) : null}
              {action === "devolver" ? <DevolucaoAction receita={l} /> : null}
            </RequirePermission>
          ))}
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

      <RequirePermission permission="financeiro:comprovante:read" mode="action">
        <ComprovantesPanel lancamentoId={l.id} />
      </RequirePermission>
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

type ActionProps = {
  lancamento: Lancamento;
  onStart: () => void;
  onError: (error: unknown) => void;
};

const STATUS_ACTION = {
  receber: {
    label: "Receber",
    title: "Confirmar recebimento?",
    description: "A receita passa a Recebida e entra no saldo.",
    confirmLabel: "Confirmar recebimento",
    success: "Recebimento registrado.",
  },
  pagar: {
    label: "Pagar",
    title: "Confirmar pagamento?",
    description: "A despesa passa a Paga e sai do saldo.",
    confirmLabel: "Confirmar pagamento",
    success: "Pagamento registrado.",
  },
} as const;

/** Receber (FWB-04 AC1) ou pagar (AC2), com confirmação. */
function StatusAction({ kind, lancamento, onStart, onError }: ActionProps & { kind: "receber" | "pagar" }) {
  const receber = useReceberLancamento(lancamento.id);
  const pagar = usePagarLancamento(lancamento.id);
  const mutation = kind === "receber" ? receber : pagar;
  const text = STATUS_ACTION[kind];
  return (
    <ConfirmDialog
      trigger={<Button>{text.label}</Button>}
      title={text.title}
      description={text.description}
      confirmLabel={text.confirmLabel}
      pending={mutation.isPending}
      onConfirm={async () => {
        onStart();
        try {
          await mutation.mutateAsync();
          toast.success(text.success);
        } catch (error) {
          onError(error);
        }
      }}
    />
  );
}

/** Cancelamento com motivo obrigatório (FWB-04 AC3-4). */
function CancelarAction({ lancamento, onStart, onError }: ActionProps) {
  const cancelar = useCancelarLancamento(lancamento.id);
  return (
    <ConfirmDialog
      trigger={<Button variant="destructive">Cancelar lançamento</Button>}
      title="Cancelar lançamento?"
      description="O lançamento não é excluído: fica como Cancelado, com o motivo registrado."
      confirmLabel="Cancelar lançamento"
      cancelLabel="Voltar"
      destructive
      reason="required"
      reasonLabel="Motivo do cancelamento"
      pending={cancelar.isPending}
      onConfirm={async (reason) => {
        if (!reason) return;
        onStart();
        try {
          await cancelar.mutateAsync(reason);
          toast.success("Lançamento cancelado.");
        } catch (error) {
          onError(error);
        }
      }}
    />
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
