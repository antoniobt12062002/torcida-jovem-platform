"use client";

import { useMutation } from "@tanstack/react-query";
import Link from "next/link";
import { useForm } from "react-hook-form";

import { FormLayout } from "@/components/app/form-layout";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button, buttonVariants } from "@/components/ui/button";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { PageTitle } from "@/components/ui/typography";
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
    <div>
      <Link href={routes.entrar} className={buttonVariants({ variant: "link", size: "sm" })}>
        Voltar para entrar
      </Link>
    </div>
  );

  if (mutation.isSuccess) {
    return (
      <div className="flex flex-col gap-6">
        <PageTitle>Recuperar acesso</PageTitle>
        <Alert variant="success">
          <AlertDescription>{mutation.data.message}</AlertDescription>
        </Alert>
        {backToLogin}
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <PageTitle>Recuperar acesso</PageTitle>
        <p className="text-sm text-muted-foreground">
          Informe o e-mail da sua conta para receber um link de redefinição de senha.
        </p>
      </div>

      <FormLayout
        noValidate
        onSubmit={(event) => void handleSubmit(onValid)(event)}
        error={errors.root?.message}
        footer={
          <Button type="submit" size="lg" className="w-full md:w-full" disabled={isSubmitting}>
            {isSubmitting ? "Enviando…" : "Enviar link"}
          </Button>
        }
      >

        <Field invalid={!!errors.email}>
          <FieldLabel>E-mail</FieldLabel>
          <Input
            type="email"
            autoComplete="username"
            spellCheck={false}
            autoCapitalize="none"
            {...register("email", {
              validate: (value) => value.trim() !== "" || "Informe o e-mail.",
            })}
          />
          <FieldError>{errors.email?.message}</FieldError>
        </Field>

      </FormLayout>

      {backToLogin}
    </div>
  );
}
