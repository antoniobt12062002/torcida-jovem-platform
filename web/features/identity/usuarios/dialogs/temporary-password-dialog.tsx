"use client";

import { useId, useState } from "react";

import { ApiErrorAlert } from "@/components/app/api-error";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "@/components/ui/toast";
import { type ApiError, isApiError, networkError } from "@/lib/api/problem";

import { usersCatalog } from "../errors";
import { requestTemporaryPassword, type User, useInvalidateUsers } from "../hooks";
import { isReasonValid, REASON_HINT } from "./reason";

// Senha temporária (USR-05). A senha vive só no estado deste diálogo: não
// passa pelo QueryClient, por armazenamento do navegador nem por log, e é
// descartada ao fechar.

export const MESSAGE_SHOWN_ONCE =
  "Esta senha é mostrada uma única vez. Entregue-a à pessoa por um canal seguro: ela terá de trocá-la no primeiro acesso.";

export function TemporaryPasswordDialog({
  user,
  open,
  onOpenChange,
}: {
  user: User;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const reasonId = useId();
  const invalidate = useInvalidateUsers();
  const [reason, setReason] = useState("");
  const [pending, setPending] = useState(false);
  const [error, setError] = useState<ApiError | null>(null);
  const [password, setPassword] = useState<string | null>(null);

  const canSubmit = !pending && isReasonValid(reason);

  function close() {
    if (pending) return;
    setPassword(null);
    setReason("");
    setError(null);
    onOpenChange(false);
  }

  async function generate() {
    if (!canSubmit) return;
    setPending(true);
    setError(null);
    try {
      setPassword(await requestTemporaryPassword(user.id, reason.trim()));
    } catch (cause) {
      setError(isApiError(cause) ? cause : networkError(cause));
      return;
    } finally {
      setPending(false);
    }
    // A troca de senha passa a ser obrigatória: a lista precisa refletir.
    void invalidate();
  }

  async function copy() {
    if (password === null) return;
    try {
      await navigator.clipboard.writeText(password);
      toast.success("Senha copiada.");
    } catch {
      toast.error("Não foi possível copiar. Selecione a senha e copie manualmente.");
    }
  }

  return (
    <Dialog open={open} onOpenChange={(next) => (next ? onOpenChange(true) : close())}>
      <DialogContent showCloseButton={password === null}>
        <DialogHeader>
          <DialogTitle>Senha temporária para {user.name}</DialogTitle>
          <DialogDescription>
            A senha atual deixa de valer e a pessoa terá de trocar a senha no próximo acesso.
          </DialogDescription>
        </DialogHeader>

        {password === null ? (
          <>
            {error ? <ApiErrorAlert error={error} catalog={usersCatalog} /> : null}
            <Field>
              <FieldLabel htmlFor={reasonId}>Motivo</FieldLabel>
              <Textarea
                id={reasonId}
                value={reason}
                disabled={pending}
                onValueChange={(value) => setReason(String(value))}
              />
              <FieldDescription>{REASON_HINT}</FieldDescription>
            </Field>
            <DialogFooter>
              <Button variant="outline" disabled={pending} onClick={close}>
                Cancelar
              </Button>
              <Button disabled={!canSubmit} onClick={() => void generate()}>
                Gerar senha temporária
              </Button>
            </DialogFooter>
          </>
        ) : (
          <>
            <Alert>
              <AlertDescription>{MESSAGE_SHOWN_ONCE}</AlertDescription>
            </Alert>
            <div className="flex items-center gap-2">
              <code
                className="flex-1 rounded-md border bg-muted px-3 py-2 font-mono text-sm break-all select-all"
              >
                {password}
              </code>
              <Button variant="outline" onClick={() => void copy()}>
                Copiar
              </Button>
            </div>
            <DialogFooter>
              <Button onClick={close}>Fechar</Button>
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
