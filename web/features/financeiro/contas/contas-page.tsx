"use client";

import { useState } from "react";

import { apiErrorMessage } from "@/components/app/api-error";
import { ConfirmDialog } from "@/components/app/confirm-dialog";
import { RequirePermission } from "@/components/app/require-permission";
import { EmptyState, ErrorState, LoadingState } from "@/components/app/states";
import { activeStatus, StatusBadge } from "@/components/app/status-badge";
import { Button } from "@/components/ui/button";
import { toast } from "@/components/ui/toast";
import { isApiError, networkError } from "@/lib/api/problem";

import { CreateContaDialog, RenameContaDialog } from "./conta-dialogs";
import { contasCatalog, tipoContaLabel } from "./errors";
import { type Conta, useContas, useDeactivateConta } from "./hooks";
import { buildContaTree, type ContaNode } from "./tree";

// Plano de contas (FWB-01): árvore por parent_id, com contas inativas
// visíveis e marcadas, e as ações de escrita só com a permissão de cada uma.

export const PERMISSION = {
  read: "financeiro:conta:read",
  create: "financeiro:conta:create",
  update: "financeiro:conta:update",
  deactivate: "financeiro:conta:deactivate",
} as const;

type DialogState =
  | { kind: "create"; parent?: Conta }
  | { kind: "rename"; conta: Conta }
  | { kind: "deactivate"; conta: Conta }
  | null;

type Actions = {
  onSubconta: (conta: Conta) => void;
  onRename: (conta: Conta) => void;
  onDeactivate: (conta: Conta) => void;
};

export function ContasPage() {
  const query = useContas();
  const deactivate = useDeactivateConta();
  const [dialog, setDialog] = useState<DialogState>(null);

  const actions: Actions = {
    onSubconta: (parent) => setDialog({ kind: "create", parent }),
    onRename: (conta) => setDialog({ kind: "rename", conta }),
    onDeactivate: (conta) => setDialog({ kind: "deactivate", conta }),
  };
  const close = () => setDialog(null);

  async function confirmDeactivate(conta: Conta) {
    try {
      await deactivate.mutateAsync(conta.id);
    } catch (cause) {
      const error = isApiError(cause) ? cause : networkError(cause);
      toast.error(apiErrorMessage(error, contasCatalog));
      // Mantém o diálogo aberto; a conta continua como estava.
      throw cause;
    }
    toast.success("Conta desativada.");
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex flex-col gap-1">
          <h1 className="text-xl font-semibold">Plano de contas</h1>
          <p className="text-sm text-muted-foreground">
            Contas de receita e de despesa usadas nos lançamentos. Contas nunca são excluídas: quando
            deixam de ser usadas, ficam inativas.
          </p>
        </div>
        <RequirePermission permission={PERMISSION.create} mode="action">
          <Button onClick={() => setDialog({ kind: "create" })}>Nova conta</Button>
        </RequirePermission>
      </div>

      {query.isPending ? (
        <LoadingState label="Carregando contas…" />
      ) : query.isError ? (
        <ErrorState error={query.error} catalog={contasCatalog} onRetry={() => void query.refetch()} />
      ) : query.data.items.length === 0 ? (
        <EmptyState title="Nenhuma conta cadastrada." />
      ) : (
        <ContaList nodes={buildContaTree(query.data.items)} actions={actions} label="Plano de contas" />
      )}

      {dialog?.kind === "create" ? (
        <CreateContaDialog key={dialog.parent?.id ?? "raiz"} parent={dialog.parent} onClose={close} />
      ) : null}
      {dialog?.kind === "rename" ? (
        <RenameContaDialog key={dialog.conta.id} conta={dialog.conta} onClose={close} />
      ) : null}
      {dialog?.kind === "deactivate" ? (
        <ConfirmDialog
          open
          onOpenChange={(open) => !open && close()}
          title={`Desativar a conta ${dialog.conta.nome}?`}
          description="A conta continua no plano de contas, marcada como inativa, e não é excluída."
          confirmLabel="Desativar"
          destructive
          pending={deactivate.isPending}
          onConfirm={() => confirmDeactivate(dialog.conta)}
        />
      ) : null}
    </div>
  );
}

function ContaList({ nodes, actions, label }: { nodes: ContaNode[]; actions: Actions; label?: string }) {
  return (
    <ul aria-label={label} className={label ? "flex flex-col gap-1" : "ml-5 flex flex-col gap-1 border-l pl-3"}>
      {nodes.map((node) => (
        <li key={node.conta.id} className="flex flex-col gap-1">
          <ContaRow conta={node.conta} actions={actions} />
          {node.children.length > 0 ? <ContaList nodes={node.children} actions={actions} /> : null}
        </li>
      ))}
    </ul>
  );
}

function ContaRow({ conta, actions }: { conta: Conta; actions: Actions }) {
  return (
    <div
      data-slot="conta-row"
      className="flex flex-wrap items-center justify-between gap-2 rounded-lg border px-3 py-2"
    >
      <div className="flex flex-wrap items-center gap-2">
        <span data-slot="conta-nome" className={conta.ativo ? "font-medium" : "font-medium text-muted-foreground"}>
          {conta.nome}
        </span>
        <span className="text-xs text-muted-foreground">{tipoContaLabel[conta.tipo]}</span>
        <StatusBadge status={activeStatus(conta.ativo)} />
      </div>
      <div className="flex flex-wrap gap-1">
        <RequirePermission permission={PERMISSION.create} mode="action">
          <Button
            variant="ghost"
            size="sm"
            aria-label={`Nova subconta em ${conta.nome}`}
            onClick={() => actions.onSubconta(conta)}
          >
            Nova subconta
          </Button>
        </RequirePermission>
        <RequirePermission permission={PERMISSION.update} mode="action">
          <Button
            variant="ghost"
            size="sm"
            aria-label={`Renomear ${conta.nome}`}
            onClick={() => actions.onRename(conta)}
          >
            Renomear
          </Button>
        </RequirePermission>
        {conta.ativo ? (
          <RequirePermission permission={PERMISSION.deactivate} mode="action">
            <Button
              variant="ghost"
              size="sm"
              aria-label={`Desativar ${conta.nome}`}
              onClick={() => actions.onDeactivate(conta)}
            >
              Desativar
            </Button>
          </RequirePermission>
        ) : null}
      </div>
    </div>
  );
}
