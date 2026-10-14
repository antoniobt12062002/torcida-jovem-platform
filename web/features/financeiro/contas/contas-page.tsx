"use client";

import { BanIcon, PencilIcon, PlusIcon } from "lucide-react";
import { useState } from "react";

import { apiErrorMessage } from "@/components/app/api-error";
import { ConfirmDialog } from "@/components/app/confirm-dialog";
import { PageHeader } from "@/components/app/page-header";
import { RequirePermission } from "@/components/app/require-permission";
import { RowActionButton } from "@/components/app/row-actions";
import { EmptyState, ErrorState, LoadingState } from "@/components/app/states";
import { activeStatusDef, StatusBadge } from "@/components/app/status-badge";
import { Button } from "@/components/ui/button";
import { toast } from "@/components/ui/toast";
import { isApiError, networkError } from "@/lib/api/problem";
import { useCan } from "@/lib/session/session-provider";
import { cn } from "@/lib/utils";

import { CreateContaDialog, RenameContaDialog } from "./conta-dialogs";
import { contasCatalog, tipoContaLabel } from "./errors";
import { type Conta, useContas, useDeactivateConta } from "./hooks";
import { buildContaTree, type ContaNode } from "./tree";

// Plano de contas (FWB-01, DSM-03 AC7): árvore por parent_id, com recuo por
// nível em qualquer largura (não vira lista empilhada nem tabela), contas
// inativas visíveis e marcadas com ícone e texto, e as ações de escrita só
// com a permissão de cada uma. As ações ficam visíveis na linha, sem menu:
// são no máximo três, curtas, e quebram para baixo do nome em tela estreita.

export const PERMISSION = {
  read: "financeiro:conta:read",
  create: "financeiro:conta:create",
  update: "financeiro:conta:update",
  deactivate: "financeiro:conta:deactivate",
} as const;

export const DESCRIPTION =
  "Contas de receita e de despesa usadas nos lançamentos. Contas nunca são excluídas: quando deixam de ser usadas, ficam inativas.";

export const DEACTIVATE_CONSEQUENCE = "A conta continua no plano de contas, marcada como inativa, e não é excluída.";

export const EMPTY_CAN_CREATE =
  "Comece pelas contas principais de receita e de despesa; as subcontas entram depois, dentro delas.";

export const EMPTY_READ_ONLY = "Quando a tesouraria cadastrar as contas, elas aparecem aqui.";

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
  const canCreate = useCan(PERMISSION.create);
  const [dialog, setDialog] = useState<DialogState>(null);
  const [deactivateError, setDeactivateError] = useState<string>();

  const actions: Actions = {
    onSubconta: (parent) => setDialog({ kind: "create", parent }),
    onRename: (conta) => setDialog({ kind: "rename", conta }),
    onDeactivate: (conta) => {
      setDeactivateError(undefined);
      setDialog({ kind: "deactivate", conta });
    },
  };
  const close = () => setDialog(null);

  async function confirmDeactivate(conta: Conta) {
    setDeactivateError(undefined);
    try {
      await deactivate.mutateAsync(conta.id);
    } catch (cause) {
      const error = isApiError(cause) ? cause : networkError(cause);
      // O diálogo continua aberto, com a mensagem; a conta continua como estava.
      setDeactivateError(apiErrorMessage(error, contasCatalog));
      throw cause;
    }
    toast.success("Conta desativada.");
  }

  const empty = query.isSuccess && query.data.items.length === 0;
  const newConta = (
    <RequirePermission permission={PERMISSION.create} mode="action">
      <Button onClick={() => setDialog({ kind: "create" })}>
        <PlusIcon aria-hidden="true" />
        Nova conta
      </Button>
    </RequirePermission>
  );

  return (
    <>
      {/* Com o plano vazio, "Nova conta" vai para o estado vazio: um só primário na página. */}
      <PageHeader title="Plano de contas" description={DESCRIPTION} actions={empty ? undefined : newConta} />

      {query.isPending ? (
        <LoadingState label="Carregando contas…" shape="list" />
      ) : query.isError ? (
        <ErrorState error={query.error} catalog={contasCatalog} onRetry={() => void query.refetch()} />
      ) : empty ? (
        <EmptyState
          title="Nenhuma conta cadastrada."
          description={canCreate ? EMPTY_CAN_CREATE : EMPTY_READ_ONLY}
          action={canCreate ? newConta : undefined}
        />
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
          subject={dialog.conta.nome}
          consequence={DEACTIVATE_CONSEQUENCE}
          confirmLabel="Desativar"
          pendingLabel="Desativando…"
          destructive
          pending={deactivate.isPending}
          error={deactivateError}
          onConfirm={() => confirmDeactivate(dialog.conta)}
        />
      ) : null}
    </>
  );
}

/**
 * Um nível da árvore. A raiz é a lista nomeada da página; cada nível abaixo
 * fica dentro do item da conta pai, recuado e com um filete à esquerda.
 */
function ContaList({ nodes, actions, label }: { nodes: ContaNode[]; actions: Actions; label?: string }) {
  return (
    <ul
      aria-label={label}
      data-slot={label ? "conta-tree" : "conta-subtree"}
      className={label ? "flex flex-col gap-2" : "ml-2 flex flex-col gap-2 border-l pl-3 md:ml-4 md:pl-4"}
    >
      {nodes.map((node) => (
        <li key={node.conta.id} className="flex flex-col gap-2">
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
      className="flex flex-wrap items-center justify-between gap-x-3 gap-y-2 rounded-lg border bg-card px-3 py-2"
    >
      <div className="flex min-w-0 flex-wrap items-center gap-x-2 gap-y-1">
        <span data-slot="conta-nome" className={cn("font-medium wrap-break-word", !conta.ativo && "text-muted-foreground")}>
          {conta.nome}
        </span>
        <span className="text-xs text-muted-foreground">{tipoContaLabel[conta.tipo]}</span>
        <StatusBadge status={activeStatusDef(conta.ativo)} />
      </div>
      <div data-slot="conta-actions" className="flex flex-wrap gap-1">
        <RequirePermission permission={PERMISSION.create} mode="action">
          <RowActionButton
            label="Nova subconta"
            rowLabel={`em ${conta.nome}`}
            icon={PlusIcon}
            onClick={() => actions.onSubconta(conta)}
          />
        </RequirePermission>
        <RequirePermission permission={PERMISSION.update} mode="action">
          <RowActionButton label="Renomear" rowLabel={conta.nome} icon={PencilIcon} onClick={() => actions.onRename(conta)} />
        </RequirePermission>
        {conta.ativo ? (
          <RequirePermission permission={PERMISSION.deactivate} mode="action">
            <RowActionButton
              label="Desativar"
              rowLabel={conta.nome}
              icon={BanIcon}
              onClick={() => actions.onDeactivate(conta)}
            />
          </RequirePermission>
        ) : null}
      </div>
    </div>
  );
}
