"use client";

import { hashKey, useMutation, useQueryClient } from "@tanstack/react-query";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { useEffect, useRef, useState } from "react";
import { useForm } from "react-hook-form";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { toast } from "@/components/ui/toast";
import { api, unwrap } from "@/lib/api/client";
import { isApiError, networkError } from "@/lib/api/problem";
import { queryKeys } from "@/lib/api/query-keys";
import { applyProblemToForm } from "@/lib/forms/apply-problem";
import { routes } from "@/lib/session/routes";

import {
  MESSAGE_CONFIRM_MISMATCH,
  MESSAGE_INVALID_RESET_TOKEN,
  MIN_LENGTH_HINT_UNKNOWN,
  passwordFieldError,
} from "./errors";
import { readResetToken } from "./read-reset-token";

// Redefinição de senha pelo link do e-mail (ACS-03). O token vem do fragmento
// (`#token=`), fica só em memória (ref deste componente) e sai da barra de
// endereço logo depois da leitura. Nunca vai para query string, armazenamento
// ou log.

type ResetPasswordValues = { new_password: string; confirm_password: string };

type LinkState = "reading" | "ready" | "missing" | "expired";

const SESSION_HASH = hashKey(queryKeys.session.me);

export function ResetPasswordForm() {
  const router = useRouter();
  const queryClient = useQueryClient();
  // undefined = fragmento ainda não lido; null = link sem token.
  const token = useRef<string | null | undefined>(undefined);
  const [linkState, setLinkState] = useState<LinkState>("reading");

  useEffect(() => {
    // A leitura acontece uma vez: um segundo efeito (StrictMode) já encontra o
    // fragmento removido e reaproveita o token em memória.
    if (token.current === undefined) {
      token.current = readResetToken(window.location.hash);
      if (window.location.hash) {
        window.history.replaceState(null, "", `${window.location.pathname}${window.location.search}`);
      }
    }
    setLinkState(token.current ? "ready" : "missing");
  }, []);

  const mutation = useMutation({
    mutationFn: (body: { token: string; new_password: string }) =>
      unwrap(api.identity.POST("/api/v1/auth/password-reset/confirm", { body })),
  });

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<ResetPasswordValues>({ defaultValues: { new_password: "", confirm_password: "" } });

  async function onValid({ new_password }: ResetPasswordValues) {
    if (!token.current) return;
    try {
      await mutation.mutateAsync({ token: token.current, new_password });
    } catch (cause) {
      const error = isApiError(cause) ? cause : networkError(cause);
      if (error.code === "invalid_reset_token") {
        setLinkState("expired");
        return;
      }
      const placed = passwordFieldError(error.code, MIN_LENGTH_HINT_UNKNOWN);
      if (placed?.field === "new_password") {
        setError("new_password", { type: error.code, message: placed.message });
        return;
      }
      applyProblemToForm(error, setError, { new_password: "new_password" });
      return;
    }
    token.current = null;
    // A API encerrou todas as sessões da pessoa: o front esquece a que tinha.
    queryClient.setQueryData(queryKeys.session.me, null);
    queryClient.removeQueries({ predicate: (q) => q.queryHash !== SESSION_HASH });
    toast.success("Senha redefinida. Entre com a nova senha.");
    router.replace(routes.entrar);
  }

  if (linkState === "reading") return null;

  if (linkState === "missing" || linkState === "expired") {
    return (
      <div className="flex flex-col gap-6">
        <h1 className="text-xl font-semibold">Redefinir senha</h1>
        <Alert variant="destructive">
          <AlertDescription>
            {linkState === "expired" ? MESSAGE_INVALID_RESET_TOKEN : "O link de recuperação é inválido."}
          </AlertDescription>
        </Alert>
        <Link
          href={routes.recuperarAcesso}
          className="text-sm underline underline-offset-4 hover:text-foreground"
        >
          Pedir um novo link
        </Link>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h1 className="text-xl font-semibold">Redefinir senha</h1>
        <p className="text-sm text-muted-foreground">Escolha uma nova senha para a sua conta.</p>
      </div>

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

        <Field invalid={!!errors.new_password}>
          <FieldLabel>Nova senha</FieldLabel>
          <Input
            type="password"
            autoComplete="new-password"
            {...register("new_password", { required: "Informe a nova senha." })}
          />
          <FieldDescription>{MIN_LENGTH_HINT_UNKNOWN}</FieldDescription>
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
          {isSubmitting ? "Redefinindo…" : "Redefinir senha"}
        </Button>
      </form>
    </div>
  );
}
