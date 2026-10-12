"use client";

import Link from "next/link";
import { type ReactNode, useState } from "react";

import { ApiErrorAlert } from "@/components/app/api-error";
import { ConfirmDialog } from "@/components/app/confirm-dialog";
import { authorLabel, formatDateTime, formatIdentifier, type Identifier, shortId } from "@/components/app/format";
import { MoneyValue } from "@/components/app/money-value";
import { PageHeader } from "@/components/app/page-header";
import { RequirePermission } from "@/components/app/require-permission";
import { EmptyState, ErrorState, LoadingState } from "@/components/app/states";
import { StatusBadge } from "@/components/app/status-badge";
import { Button, buttonVariants } from "@/components/ui/button";
import { toast } from "@/components/ui/toast";
import { type ApiError, isApiError, networkError } from "@/lib/api/problem";
import { useSession } from "@/lib/session";

import { ComprovantesPanel } from "@/features/financeiro/comprovantes/comprovantes-panel";

import { ACTION_PERMISSION, type LancamentoAction, lancamentoActions } from "./actions";
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
import { IdentifierText } from "./identifier";
import { LancamentoFormDialog } from "./lancamento-form";
import { FORMA_PAGAMENTO_LABEL, type Lancamento, STATUS_DEF, TIPO_LABEL } from "./labels";

// Detalhe de um lançamento, lido do cache da lista (FWB-03 AC6), com a trilha
// de volta à listagem, o status ao lado do título e as ações de workflow no
// cabeçalho (FWB-04).

const LIST_HREF = "/financeiro/lancamentos";
const LIST_CRUMB = { label: "Lançamentos", href: LIST_HREF };

/** Título enquanto o lançamento não está disponível. */
const FALLBACK_TITLE = "Lançamento";

/** Ordem de leitura das ações no cabeçalho: secundárias, a destrutiva e a principal por último. */
const DISPLAY_ORDER: readonly LancamentoAction[] = ["editar", "devolver", "cancelar", "receber", "pagar"];

export function LancamentoDetail({ id }: { id: string }) {
  const query = useLancamento(id);

  if (query.isSuccess && query.data !== null) return <LancamentoView lancamento={query.data} />;

  return (
    <>
      <PageHeader title={FALLBACK_TITLE} breadcrumb={[LIST_CRUMB, { label: FALLBACK_TITLE }]} />
      {query.isPending ? (
        <LoadingState shape="detail" />
      ) : query.isError ? (
        <ErrorState error={query.error} catalog={lancamentosCatalog} onRetry={() => void query.refetch()} />
      ) : (
        <EmptyState
          title={MESSAGE_NAO_ENCONTRADO}
          action={
            <Link href={LIST_HREF} className={buttonVariants({ variant: "secondary" })}>
              Voltar para lançamentos
            </Link>
          }
        />
      )}
    </>
  );
}

/** Valor líquido com o sentido do lançamento; cancelado não move o saldo e fica riscado. */
function netVariant(l: Lancamento) {
  if (l.status === "CANCELADA") return "neutral";
  return l.tipo === "RECEITA" ? "income" : "expense";
}

function LancamentoView({ lancamento: l }: { lancamento: Lancamento }) {
  const { context } = useSession();
  const userId = context?.user.id;
  const contaNome = useContaNome();
  const identifier = formatIdentifier(TIPO_LABEL[l.tipo], l.id);
  const available = new Set(lancamentoActions(l, context?.permissions ?? []));
  const actions = DISPLAY_ORDER.filter((action) => available.has(action));
  const [actionError, setActionError] = useState<ApiError | null>(null);
  const report = (error: unknown) => setActionError(isApiError(error) ? error : networkError(error));
  const common = { lancamento: l, identifier, onStart: () => setActionError(null), onError: report };

  return (
    <>
      <PageHeader
        title={identifier.label}
        breadcrumb={[LIST_CRUMB, { label: identifier.label }]}
        meta={<StatusBadge status={STATUS_DEF[l.status]} />}
        actions={
          actions.length > 0
            ? actions.map((action) => (
                <RequirePermission key={action} permission={ACTION_PERMISSION[action]} mode="action">
                  {action === "editar" ? <EditarAction lancamento={l} identifier={identifier} /> : null}
                  {action === "receber" ? <StatusAction kind="receber" {...common} /> : null}
                  {action === "pagar" ? <StatusAction kind="pagar" {...common} /> : null}
                  {action === "cancelar" ? <CancelarAction {...common} /> : null}
                  {action === "devolver" ? <DevolucaoAction receita={l} /> : null}
                </RequirePermission>
              ))
            : undefined
        }
      />

      {actionError ? <ApiErrorAlert error={actionError} catalog={lancamentosCatalog} /> : null}

      <dl className="grid gap-x-6 gap-y-3 sm:grid-cols-[max-content_1fr]">
        <Item label="Tipo">{TIPO_LABEL[l.tipo]}</Item>
        <Item label="Conta">{contaNome(l.conta_id)}</Item>
        <Item label="Valor bruto">
          <MoneyValue cents={l.valor_bruto_cents} />
        </Item>
        <Item label="Taxa">
          <MoneyValue cents={l.taxa_cents} />
        </Item>
        <Item label="Valor líquido">
          <MoneyValue
            cents={l.valor_liquido_cents}
            variant={netVariant(l)}
            className={l.status === "CANCELADA" ? "line-through" : undefined}
          />
        </Item>
        <Item label="Forma de pagamento">{FORMA_PAGAMENTO_LABEL[l.forma_pagamento]}</Item>
        <Item label="Criado por">{authorLabel(l.criado_por, userId)}</Item>
        <Item label="Criado em">{formatDateTime(l.criado_em)}</Item>
        <Item label="Atualizado em">{formatDateTime(l.atualizado_em)}</Item>
        {l.devolucao_de_id ? (
          <Item label="Devolução da receita">
            <Link
              href={`${LIST_HREF}/${l.devolucao_de_id}`}
              className="font-mono underline underline-offset-4 hover:decoration-2"
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
      <dd className="text-sm wrap-break-word">{children}</dd>
    </div>
  );
}

type ActionProps = {
  lancamento: Lancamento;
  identifier: Identifier;
  onStart: () => void;
  onError: (error: unknown) => void;
};

// O botão de confirmar repete o verbo da ação (DSP-03 AC3).
const STATUS_ACTION = {
  receber: {
    label: "Receber",
    title: "Receber esta receita?",
    consequence: "A receita passa a Recebida e entra no saldo.",
    confirmLabel: "Receber receita",
    pendingLabel: "Recebendo…",
    success: "Recebimento registrado.",
  },
  pagar: {
    label: "Pagar",
    title: "Pagar esta despesa?",
    consequence: "A despesa passa a Paga e sai do saldo.",
    confirmLabel: "Pagar despesa",
    pendingLabel: "Pagando…",
    success: "Pagamento registrado.",
  },
} as const;

/** Receber (FWB-04 AC1) ou pagar (AC2), com confirmação. */
function StatusAction({ kind, lancamento, identifier, onStart, onError }: ActionProps & { kind: "receber" | "pagar" }) {
  const receber = useReceberLancamento(lancamento.id);
  const pagar = usePagarLancamento(lancamento.id);
  const mutation = kind === "receber" ? receber : pagar;
  const text = STATUS_ACTION[kind];
  return (
    <ConfirmDialog
      trigger={<Button>{text.label}</Button>}
      title={text.title}
      subject={<Subject lancamento={lancamento} identifier={identifier} />}
      consequence={text.consequence}
      confirmLabel={text.confirmLabel}
      pendingLabel={text.pendingLabel}
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
function CancelarAction({ lancamento, identifier, onStart, onError }: ActionProps) {
  const cancelar = useCancelarLancamento(lancamento.id);
  return (
    <ConfirmDialog
      trigger={<Button variant="destructive">Cancelar lançamento</Button>}
      title="Cancelar lançamento?"
      subject={<Subject lancamento={lancamento} identifier={identifier} />}
      consequence="O lançamento não é excluído: fica como Cancelado, com o motivo registrado."
      confirmLabel="Cancelar lançamento"
      cancelLabel="Voltar"
      pendingLabel="Cancelando…"
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

/** O registro afetado nas confirmações: identificador e valor líquido. */
function Subject({ lancamento, identifier }: { lancamento: Lancamento; identifier: Identifier }) {
  return (
    <>
      <IdentifierText identifier={identifier} />, <MoneyValue cents={lancamento.valor_liquido_cents} />
    </>
  );
}

function EditarAction({ lancamento: l, identifier }: { lancamento: Lancamento; identifier: Identifier }) {
  const [open, setOpen] = useState(false);
  const update = useUpdateLancamento(l.id);
  return (
    <>
      <Button variant="secondary" onClick={() => setOpen(true)}>
        Editar
      </Button>
      <LancamentoFormDialog
        open={open}
        onOpenChange={setOpen}
        title="Editar lançamento"
        description={
          <>
            Altere a conta, os valores ou a forma de pagamento da <IdentifierText identifier={identifier} />.
          </>
        }
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
