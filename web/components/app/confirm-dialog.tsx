"use client";

import { type ReactElement, type ReactNode, useId, useState } from "react";

import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogTrigger,
} from "@/components/ui/alert-dialog";
import { Field, FieldLabel } from "@/components/ui/field";
import { Textarea } from "@/components/ui/textarea";

import { FormErrorAlert } from "./form-error-alert";

// Confirmação de ação irreversível ou sensível (FND-05 AC6, DSP-03). Nada é
// chamado antes do clique em confirmar. Com motivo obrigatório, confirmar fica
// desabilitado enquanto o motivo estiver vazio ou só com espaços. A confirmação
// nomeia o registro (`subject`) e diz o efeito (`consequence`); se a ação
// falhar, continua aberta e mostra o `error`.

export type ConfirmDialogProps = {
  title: string;
  /** Registro afetado, mostrado em destaque: "Receita fb6e3d8f", o nome da conta. */
  subject?: ReactNode;
  /** Frase sobre o efeito da ação. */
  consequence?: ReactNode;
  /** Texto livre da descrição, depois do registro e da consequência. */
  description?: ReactNode;
  /** Texto do botão de confirmar: repete o verbo da ação ("Cancelar lançamento"). */
  confirmLabel: string;
  /** Texto do botão de cancelar. Padrão: "Cancelar". */
  cancelLabel?: string;
  /** Botão de confirmar com estilo destrutivo. */
  destructive?: boolean;
  /** Pede um motivo: `required` exige texto além de espaços. */
  reason?: "required" | "optional";
  /** Rótulo do campo de motivo. Padrão: "Motivo". */
  reasonLabel?: string;
  /**
   * Chamado ao confirmar, com o motivo aparado (undefined sem motivo). Se
   * devolver uma Promise, os botões ficam desabilitados até ela terminar; ao
   * resolver o diálogo fecha, ao rejeitar fica aberto (quem chama mostra o erro).
   */
  onConfirm: (reason: string | undefined) => void | Promise<unknown>;
  /** Abertura controlada; sem `open`, o diálogo controla a si mesmo. */
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  /** Elemento que abre o diálogo, por exemplo `<Button>Cancelar lançamento</Button>`. */
  trigger?: ReactElement;
  /** Ação em andamento controlada por fora (por exemplo `mutation.isPending`). */
  pending?: boolean;
  /** Mensagem da falha da ação confirmada, mostrada dentro do diálogo. */
  error?: ReactNode;
  /**
   * Rótulo do botão de confirmar enquanto a ação está em andamento, terminado
   * em "…" (DSC-02 AC4). Padrão: "Confirmando…". O padrão não repete
   * `confirmLabel` com reticências, porque "Cancelar lançamento…" se leria como
   * um passo que ainda abre outra tela.
   */
  pendingLabel?: `${string}…`;
};

const DEFAULT_PENDING_LABEL = "Confirmando…";

export function ConfirmDialog({
  title,
  subject,
  consequence,
  description,
  confirmLabel,
  cancelLabel = "Cancelar",
  destructive = false,
  reason,
  reasonLabel = "Motivo",
  onConfirm,
  open: openProp,
  onOpenChange,
  trigger,
  pending = false,
  error,
  pendingLabel = DEFAULT_PENDING_LABEL,
}: ConfirmDialogProps) {
  const reasonId = useId();
  const [internalOpen, setInternalOpen] = useState(false);
  const open = openProp ?? internalOpen;
  const [reasonText, setReasonText] = useState("");
  const [running, setRunning] = useState(false);

  // O motivo é esquecido sempre que o diálogo fecha.
  const [wasOpen, setWasOpen] = useState(open);
  if (open !== wasOpen) {
    setWasOpen(open);
    if (!open) setReasonText("");
  }

  const busy = running || pending;
  const trimmed = reasonText.trim();
  const canConfirm = !busy && (reason !== "required" || trimmed !== "");

  function setOpen(next: boolean) {
    if (openProp === undefined) setInternalOpen(next);
    onOpenChange?.(next);
  }

  async function confirm() {
    if (!canConfirm) return;
    setRunning(true);
    try {
      await onConfirm(reason ? trimmed || undefined : undefined);
      setOpen(false);
    } catch {
      // A falha é tratada por quem chama; o diálogo continua aberto.
    } finally {
      setRunning(false);
    }
  }

  return (
    <AlertDialog
      open={open}
      onOpenChange={(next) => {
        if (!next && busy) return;
        setOpen(next);
      }}
    >
      {trigger ? <AlertDialogTrigger render={trigger} /> : null}
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          {subject || consequence || description ? (
            <AlertDialogDescription className="flex flex-col gap-1">
              {subject ? <strong className="font-semibold wrap-break-word text-foreground">{subject}</strong> : null}
              {consequence ? <span>{consequence}</span> : null}
              {description ? <span>{description}</span> : null}
            </AlertDialogDescription>
          ) : null}
        </AlertDialogHeader>
        {reason ? (
          <Field>
            <FieldLabel htmlFor={reasonId} optional={reason === "optional"}>
              {reasonLabel}
            </FieldLabel>
            <Textarea
              id={reasonId}
              name="motivo"
              autoComplete="off"
              value={reasonText}
              required={reason === "required"}
              disabled={busy}
              onValueChange={(value) => setReasonText(String(value))}
            />
          </Field>
        ) : null}
        {/* Recebe o foco: os botões ficam desabilitados durante a ação. */}
        {error ? <FormErrorAlert>{error}</FormErrorAlert> : null}
        <AlertDialogFooter>
          <AlertDialogCancel disabled={busy}>{cancelLabel}</AlertDialogCancel>
          <AlertDialogAction
            variant={destructive ? "destructive" : "primary"}
            disabled={!canConfirm}
            loading={busy}
            loadingLabel={pendingLabel}
            onClick={() => void confirm()}
          >
            {confirmLabel}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
