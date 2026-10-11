"use client";

import { useRouter, useSearchParams } from "next/navigation";
import { useState } from "react";
import { useForm } from "react-hook-form";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { isApiError, networkError } from "@/lib/api/problem";
import { applyProblemToForm } from "@/lib/forms/apply-problem";
import { safeNext } from "@/lib/session/routes";
import { useSession } from "@/lib/session/session-provider";

import {
  changePasswordCatalog,
  MESSAGE_CONFIRM_MISMATCH,
  minLengthHint,
  minPasswordLength,
  PasswordApiError,
  passwordFieldError,
} from "./errors";
import { useChangePassword } from "./use-change-password";

// Troca da própria senha (ACS-01), voluntária ou obrigatória.

type ChangePasswordValues = {
  current_password: string;
  new_password: string;
  confirm_password: string;
};

const EMPTY: ChangePasswordValues = { current_password: "", new_password: "", confirm_password: "" };

export function ChangePasswordForm() {
  const { status, context, logout } = useSession();
  const router = useRouter();
  const searchParams = useSearchParams();
  const next = searchParams.get("next");
  const mutation = useChangePassword();
  const [changed, setChanged] = useState(false);

  const mandatory = !!context?.must_change_password;
  const hint = minLengthHint(minPasswordLength(context?.admin_membership));

  const {
    register,
    handleSubmit,
    setError,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<ChangePasswordValues>({ defaultValues: EMPTY });

  async function onValid(values: ChangePasswordValues) {
    setChanged(false);
    const wasMandatory = mandatory;
    try {
      await mutation.mutateAsync({
        current_password: values.current_password,
        new_password: values.new_password,
      });
    } catch (cause) {
      const error = isApiError(cause) ? cause : networkError(cause);
      const placed = passwordFieldError(error.code, hint);
      if (placed) {
        setError(placed.field, { type: error.code, message: placed.message });
        return;
      }
      const retryAfter = cause instanceof PasswordApiError ? cause.retryAfter : null;
      applyProblemToForm(
        error,
        setError,
        { current_password: "current_password", new_password: "new_password" },
        { catalog: changePasswordCatalog(retryAfter) },
      );
      return;
    }
    if (wasMandatory) {
      router.replace(safeNext(next));
      return;
    }
    reset(EMPTY);
    setChanged(true);
  }

  // O portão de (app) só mostra a página com sessão; o modo e a dica dependem do contexto.
  if (status !== "authenticated") return null;

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h1 className="text-xl font-semibold">Trocar senha</h1>
        {mandatory ? (
          <p className="text-sm text-muted-foreground">
            Você precisa trocar a senha antes de continuar. É o primeiro acesso ou a senha foi
            redefinida pela administração.
          </p>
        ) : (
          <p className="text-sm text-muted-foreground">
            Informe a senha atual e escolha uma nova. As suas outras sessões serão encerradas.
          </p>
        )}
      </div>

      {changed ? (
        <Alert>
          <AlertDescription>Senha alterada com sucesso.</AlertDescription>
        </Alert>
      ) : null}

      <form
        className="flex flex-col gap-4"
        noValidate
        onSubmit={(event) => void handleSubmit(onValid)(event)}
      >
        {errors.root ? (
          <Alert variant="destructive">
            <AlertDescription>{errors.root.message}</AlertDescription>
          </Alert>
        ) : null}

        <Field invalid={!!errors.current_password}>
          <FieldLabel>Senha atual</FieldLabel>
          <Input
            type="password"
            autoComplete="current-password"
            {...register("current_password", { required: "Informe a senha atual." })}
          />
          <FieldError>{errors.current_password?.message}</FieldError>
        </Field>

        <Field invalid={!!errors.new_password}>
          <FieldLabel>Nova senha</FieldLabel>
          <Input
            type="password"
            autoComplete="new-password"
            {...register("new_password", { required: "Informe a nova senha." })}
          />
          <FieldDescription>{hint}</FieldDescription>
          <FieldError>{errors.new_password?.message}</FieldError>
        </Field>

        <Field invalid={!!errors.confirm_password}>
          <FieldLabel>Confirmar nova senha</FieldLabel>
          <Input
            type="password"
            autoComplete="new-password"
            {...register("confirm_password", {
              required: "Confirme a nova senha.",
              validate: (value, values) => value === values.new_password || MESSAGE_CONFIRM_MISMATCH,
            })}
          />
          <FieldError>{errors.confirm_password?.message}</FieldError>
        </Field>

        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Trocando…" : "Trocar senha"}
        </Button>
      </form>

      {mandatory ? (
        <Button variant="outline" onClick={() => void logout()}>
          Sair
        </Button>
      ) : null}
    </div>
  );
}
