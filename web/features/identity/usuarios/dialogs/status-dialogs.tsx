"use client";

import { apiErrorMessage } from "@/components/app/api-error";
import { ConfirmDialog } from "@/components/app/confirm-dialog";
import { toast } from "@/components/ui/toast";
import { isApiError, networkError } from "@/lib/api/problem";

import { usersCatalog } from "../errors";
import { type User, useDeactivateUser, useReactivateUser } from "../hooks";

// Desativação e reativação (USR-03). Em erro, a mensagem aparece em aviso e a
// lista não muda (o diálogo continua aberto).

export const MESSAGE_DEACTIVATED = "Usuário desativado.";
export const MESSAGE_REACTIVATED =
  "Usuário reativado. Papéis administrativos e vínculo administrativo não são restaurados.";

type StatusDialogProps = {
  user: User;
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

/** Mostra o erro da API em aviso e rejeita, para o diálogo continuar aberto. */
export function reportFailure(cause: unknown): never {
  const error = isApiError(cause) ? cause : networkError(cause);
  toast.error(apiErrorMessage(error, usersCatalog));
  throw error;
}

export function DeactivateUserDialog({ user, open, onOpenChange }: StatusDialogProps) {
  const mutation = useDeactivateUser();
  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title={`Desativar ${user.name}?`}
      description="A pessoa não poderá mais entrar, e as sessões dela serão encerradas."
      confirmLabel="Desativar"
      destructive
      onConfirm={async () => {
        try {
          await mutation.mutateAsync(user.id);
        } catch (cause) {
          reportFailure(cause);
        }
        toast.success(MESSAGE_DEACTIVATED);
      }}
    />
  );
}

export function ReactivateUserDialog({ user, open, onOpenChange }: StatusDialogProps) {
  const mutation = useReactivateUser();
  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title={`Reativar ${user.name}?`}
      description="A pessoa volta a poder entrar só como Associado: papéis administrativos e vínculo administrativo não são restaurados."
      confirmLabel="Reativar"
      onConfirm={async () => {
        try {
          await mutation.mutateAsync(user.id);
        } catch (cause) {
          reportFailure(cause);
        }
        toast.success(MESSAGE_REACTIVATED);
      }}
    />
  );
}
