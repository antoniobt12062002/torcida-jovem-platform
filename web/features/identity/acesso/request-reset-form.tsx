"use client";

import { useMutation } from "@tanstack/react-query";
import Link from "next/link";
import { useForm } from "react-hook-form";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { api, unwrap } from "@/lib/api/client";
import { isApiError, networkError } from "@/lib/api/problem";
import { applyProblemToForm } from "@/lib/forms/apply-problem";
import { routes } from "@/lib/session/routes";

// Pedido de recuperação de acesso (ACS-02). A resposta é a mesma para qualquer
// e-mail: a tela mostra a mensagem neutra da API, sem indicar se há conta.

type RequestResetValues = { email: string };

export function RequestResetForm() {
  const mutation = useMutation({
    mutationFn: (email: string) =>
      unwrap(api.identity.POST("/api/v1/auth/password-reset/request", { body: { email } })),
  });

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<RequestResetValues>({ defaultValues: { email: "" } });

  async function onValid({ email }: RequestResetValues) {
    try {
      await mutation.mutateAsync(email.trim());
    } catch (cause) {
      applyProblemToForm(isApiError(cause) ? cause : networkError(cause), setError, {
        email: "email",
      });
    }
  }

  const backToLogin = (
    <Link
      href={routes.entrar}
      className="text-sm text-muted-foreground underline underline-offset-4 hover:text-foreground"
    >
      Voltar para entrar
    </Link>
  );

  if (mutation.isSuccess) {
    return (
      <div className="flex flex-col gap-6">
        <h1 className="text-xl font-semibold">Recuperar acesso</h1>
        <Alert>
          <AlertDescription>{mutation.data.message}</AlertDescription>
        </Alert>
        {backToLogin}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h1 className="text-xl font-semibold">Recuperar acesso</h1>
        <p className="text-sm text-muted-foreground">
          Informe o e-mail da sua conta para receber um link de redefinição de senha.
        </p>
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

        <Field invalid={!!errors.email}>
          <FieldLabel>E-mail</FieldLabel>
          <Input
            type="email"
            autoComplete="username"
            {...register("email", {
              validate: (value) => value.trim() !== "" || "Informe o e-mail.",
            })}
          />
          <FieldError>{errors.email?.message}</FieldError>
        </Field>

        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Enviando…" : "Enviar link"}
        </Button>
      </form>

      {backToLogin}
    </div>
  );
}
