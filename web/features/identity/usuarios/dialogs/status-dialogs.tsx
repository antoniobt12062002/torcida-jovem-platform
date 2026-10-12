"use client";

import { useState } from "react";

import { apiErrorMessage } from "@/components/app/api-error";
import { ConfirmDialog } from "@/components/app/confirm-dialog";
import { toast } from "@/components/ui/toast";
import { isApiError, networkError } from "@/lib/api/problem";

import { usersCatalog } from "../errors";
import { type User, useDeactivateUser, useReactivateUser } from "../hooks";

// Desativação e reativação (USR-03). A confirmação nomeia a pessoa e a
// consequência (DSP-03 AC2); em erro, a mensagem aparece dentro dela, que
// continua aberta, e a lista não muda (DSP-03 AC5).

export const MESSAGE_DEACTIVATED = "Usuário desativado.";
export const MESSAGE_REACTIVATED =
  "Usuário reativado. Papéis administrativos e vínculo administrativo não são restaurados.";

type StatusDialogProps = {
  user: User;
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

/** A pessoa afetada, como a confirmação a mostra. */
export function subjectOf(user: User): string {
  return `${user.name} · ${user.email}`;
}

/**
 * Executa a ação confirmada. Em erro, guarda a mensagem da API para a
 * confirmação mostrar e rejeita, para ela continuar aberta.
 */
export function useConfirmFailure() {
  const [error, setError] = useState<string | null>(null);
  async function attempt(action: () => Promise<unknown>) {
    setError(null);
    try {
      await action();
    } catch (cause) {
      const failure = isApiError(cause) ? cause : networkError(cause);
      setError(apiErrorMessage(failure, usersCatalog));
      throw failure;
    }
  }
  return { error, attempt };
}

export function DeactivateUserDialog({ user, open, onOpenChange }: StatusDialogProps) {
  const mutation = useDeactivateUser();
  const { error, attempt } = useConfirmFailure();
  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title={`Desativar ${user.name}?`}
      subject={subjectOf(user)}
      consequence="A pessoa não poderá mais entrar, e as sessões dela serão encerradas."
      confirmLabel="Desativar"
      pendingLabel="Desativando…"
      destructive
      error={error}
      onConfirm={async () => {
        await attempt(() => mutation.mutateAsync(user.id));
        toast.success(MESSAGE_DEACTIVATED);
      }}
    />
  );
}

export function ReactivateUserDialog({ user, open, onOpenChange }: StatusDialogProps) {
  const mutation = useReactivateUser();
  const { error, attempt } = useConfirmFailure();
  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title={`Reativar ${user.name}?`}
      subject={subjectOf(user)}
      consequence="A pessoa volta a poder entrar só como Associado: papéis administrativos e vínculo administrativo não são restaurados."
      confirmLabel="Reativar"
      pendingLabel="Reativando…"
      error={error}
      onConfirm={async () => {
        await attempt(() => mutation.mutateAsync(user.id));
        toast.success(MESSAGE_REACTIVATED);
      }}
    />
  );
}
