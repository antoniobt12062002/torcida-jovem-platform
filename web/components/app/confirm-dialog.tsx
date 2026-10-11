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

// Confirmação de ação irreversível ou sensível (FND-05 AC6). Nada é chamado
// antes do clique em confirmar. Com motivo obrigatório, confirmar fica
// desabilitado enquanto o motivo estiver vazio ou só com espaços.

export type ConfirmDialogProps = {
  title: string;
  description?: ReactNode;
  /** Texto do botão de confirmar. Padrão: "Confirmar". */
  confirmLabel?: string;
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
};

export function ConfirmDialog({
  title,
  description,
  confirmLabel = "Confirmar",
  cancelLabel = "Cancelar",
  destructive = false,
  reason,
  reasonLabel = "Motivo",
  onConfirm,
  open: openProp,
  onOpenChange,
  trigger,
  pending = false,
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
          {description ? <AlertDialogDescription>{description}</AlertDialogDescription> : null}
        </AlertDialogHeader>
        {reason ? (
          <Field>
            <FieldLabel htmlFor={reasonId}>
              {reasonLabel}
              {reason === "optional" ? <span className="text-muted-foreground">(opcional)</span> : null}
            </FieldLabel>
            <Textarea
              id={reasonId}
              value={reasonText}
              required={reason === "required"}
              disabled={busy}
              onValueChange={(value) => setReasonText(String(value))}
            />
          </Field>
        ) : null}
        <AlertDialogFooter>
          <AlertDialogCancel disabled={busy}>{cancelLabel}</AlertDialogCancel>
          <AlertDialogAction
            variant={destructive ? "destructive" : "default"}
            disabled={!canConfirm}
            onClick={() => void confirm()}
          >
            {confirmLabel}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
