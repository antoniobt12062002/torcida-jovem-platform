"use client";

import { useId, useState } from "react";

import { ApiErrorAlert } from "@/components/app/api-error";
import { ConfirmDialog } from "@/components/app/confirm-dialog";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
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
import { type User, useGrantAdmin, useRevokeAdmin, useSetRoles } from "../hooks";
import { ADMIN_ROLES, type Role, roleLabel } from "../roles";
import { isReasonValid, REASON_HINT } from "./reason";
import { reportFailure } from "./status-dialogs";

// Acesso administrativo e papéis (USR-04): promover, editar papéis e retirar
// o acesso. A API decide a escalada de privilégio; o front mostra o resultado.

export const MESSAGE_GRANTED =
  "Acesso administrativo concedido. As sessões da pessoa foram encerradas e ela terá de trocar a senha.";
export const MESSAGE_ROLES_SAVED = "Papéis atualizados.";
export const MESSAGE_REVOKED = "Acesso administrativo retirado. A pessoa fica só com o papel Associado.";
export const MESSAGE_DORMANT =
  "Sem nenhum papel, o vínculo administrativo fica dormente: a pessoa mantém o vínculo, mas sem nenhuma permissão administrativa.";

type AdminDialogProps = {
  user: User;
  open: boolean;
  onOpenChange: (open: boolean) => void;
};

function toApiError(cause: unknown): ApiError {
  return isApiError(cause) ? cause : networkError(cause);
}

/** Papéis na ordem do contrato, para um corpo estável. */
function ordered(roles: ReadonlySet<Role>): Role[] {
  return ADMIN_ROLES.filter((role) => roles.has(role));
}

function RoleCheckboxes({
  selected,
  onChange,
  disabled,
}: {
  selected: ReadonlySet<Role>;
  onChange: (next: Set<Role>) => void;
  disabled?: boolean;
}) {
  const baseId = useId();
  return (
    <fieldset className="flex flex-col gap-2">
      <legend className="mb-1 text-sm font-medium">Papéis</legend>
      {ADMIN_ROLES.map((role) => {
        const labelId = `${baseId}-${role}`;
        return (
          <div key={role} className="flex items-center gap-2">
            <Checkbox
              aria-labelledby={labelId}
              checked={selected.has(role)}
              disabled={disabled}
              onCheckedChange={(checked) => {
                const next = new Set(selected);
                if (checked) next.add(role);
                else next.delete(role);
                onChange(next);
              }}
            />
            <span id={labelId} className="text-sm">
              {roleLabel(role)}
            </span>
          </div>
        );
      })}
    </fieldset>
  );
}

export function GrantAdminDialog({ user, open, onOpenChange }: AdminDialogProps) {
  const reasonId = useId();
  const mutation = useGrantAdmin();
  const [roles, setRoles] = useState<Set<Role>>(new Set());
  const [reason, setReason] = useState("");
  const [error, setError] = useState<ApiError | null>(null);

  const pending = mutation.isPending;
  const canSubmit = !pending && roles.size > 0 && isReasonValid(reason);

  async function submit() {
    if (!canSubmit) return;
    setError(null);
    try {
      await mutation.mutateAsync({ id: user.id, body: { roles: ordered(roles), reason: reason.trim() } });
    } catch (cause) {
      setError(toApiError(cause));
      return;
    }
    toast.success(MESSAGE_GRANTED);
    onOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={(next) => (pending ? undefined : onOpenChange(next))}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Promover {user.name} a administrativo</DialogTitle>
          <DialogDescription>
            As sessões da pessoa serão encerradas e ela terá de trocar a senha no próximo acesso.
          </DialogDescription>
        </DialogHeader>
        {error ? <ApiErrorAlert error={error} catalog={usersCatalog} /> : null}
        <RoleCheckboxes selected={roles} onChange={setRoles} disabled={pending} />
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
          <Button variant="outline" disabled={pending} onClick={() => onOpenChange(false)}>
            Cancelar
          </Button>
          <Button disabled={!canSubmit} onClick={() => void submit()}>
            Promover
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export function SetRolesDialog({ user, open, onOpenChange }: AdminDialogProps) {
  const mutation = useSetRoles();
  const [roles, setRoles] = useState<Set<Role>>(
    () => new Set(user.roles.filter((role) => role !== "ASSOCIADO")),
  );
  const [confirmingEmpty, setConfirmingEmpty] = useState(false);
  const [error, setError] = useState<ApiError | null>(null);
  const pending = mutation.isPending;

  async function save() {
    if (pending) return;
    // Lista vazia: avisa antes que o vínculo fica dormente (USR-04 AC3).
    if (roles.size === 0 && !confirmingEmpty) {
      setConfirmingEmpty(true);
      return;
    }
    setError(null);
    try {
      await mutation.mutateAsync({ id: user.id, body: { roles: ordered(roles) } });
    } catch (cause) {
      setError(toApiError(cause));
      return;
    }
    toast.success(MESSAGE_ROLES_SAVED);
    onOpenChange(false);
  }

  return (
    <Dialog open={open} onOpenChange={(next) => (pending ? undefined : onOpenChange(next))}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Editar papéis de {user.name}</DialogTitle>
          <DialogDescription>O papel Associado é sempre mantido.</DialogDescription>
        </DialogHeader>
        {error ? <ApiErrorAlert error={error} catalog={usersCatalog} /> : null}
        <RoleCheckboxes
          selected={roles}
          disabled={pending}
          onChange={(next) => {
            setRoles(next);
            setConfirmingEmpty(false);
          }}
        />
        {confirmingEmpty ? (
          <Alert>
            <AlertDescription>{MESSAGE_DORMANT}</AlertDescription>
          </Alert>
        ) : null}
        <DialogFooter>
          <Button
            variant="outline"
            disabled={pending}
            onClick={() => (confirmingEmpty ? setConfirmingEmpty(false) : onOpenChange(false))}
          >
            {confirmingEmpty ? "Voltar" : "Cancelar"}
          </Button>
          <Button disabled={pending} onClick={() => void save()}>
            {confirmingEmpty ? "Salvar sem papéis" : "Salvar papéis"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}

export function RevokeAdminDialog({ user, open, onOpenChange }: AdminDialogProps) {
  const mutation = useRevokeAdmin();
  return (
    <ConfirmDialog
      open={open}
      onOpenChange={onOpenChange}
      title={`Retirar o acesso administrativo de ${user.name}?`}
      description="A pessoa fica só com o papel Associado e as sessões dela serão encerradas. O histórico do vínculo é preservado."
      confirmLabel="Retirar acesso"
      destructive
      reason="required"
      onConfirm={async (reason) => {
        try {
          await mutation.mutateAsync({ id: user.id, body: { reason: reason ?? "" } });
        } catch (cause) {
          reportFailure(cause);
        }
        toast.success(MESSAGE_REVOKED);
      }}
    />
  );
}
