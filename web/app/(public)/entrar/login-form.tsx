"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { useEffect, useRef } from "react";
import { useForm } from "react-hook-form";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { isApiError, networkError } from "@/lib/api/problem";
import { applyProblemToForm } from "@/lib/forms/apply-problem";
import { formatRetryAfter } from "@/lib/session/retry-after";
import {
  routes,
  safeNext,
  SESSION_ENDED_PARAM,
  SESSION_ENDED_VALUE,
} from "@/lib/session/routes";
import { LoginError, useSession } from "@/lib/session/session-provider";

// Login (FND-03 AC1-4, AC10).

type LoginValues = { email: string; password: string };

const MESSAGE_INVALID_CREDENTIALS = "E-mail ou senha inválidos.";

function blockedMessage(retryAfter: string | null): string {
  const wait = formatRetryAfter(retryAfter);
  const base = "O acesso está temporariamente bloqueado por excesso de tentativas.";
  return wait ? `${base} Tente de novo em ${wait}.` : `${base} Tente de novo mais tarde.`;
}

export function LoginForm() {
  const { status, context, login } = useSession();
  const router = useRouter();
  const searchParams = useSearchParams();
  const next = searchParams.get("next");
  const sessionEnded = searchParams.get(SESSION_ENDED_PARAM) === SESSION_ENDED_VALUE;

  // Depois de um login feito aqui, o destino é decidido pelo envio do
  // formulário, não pelo redirecionamento de quem já chegou autenticado.
  const submitted = useRef(false);

  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<LoginValues>({ defaultValues: { email: "", password: "" } });

  useEffect(() => {
    if (status !== "authenticated" || submitted.current) return;
    router.replace(context?.must_change_password ? routes.contaSenha : routes.inicio);
  }, [status, context, router]);

  async function onValid({ email, password }: LoginValues) {
    submitted.current = true;
    try {
      const session = await login(email.trim(), password);
      router.replace(session.must_change_password ? routes.contaSenha : safeNext(next));
    } catch (cause) {
      submitted.current = false;
      const error = isApiError(cause) ? cause : networkError(cause);
      const retryAfter = cause instanceof LoginError ? cause.retryAfter : null;
      applyProblemToForm(
        error,
        setError,
        { email: "email", password: "password" },
        {
          catalog: {
            invalid_credentials: MESSAGE_INVALID_CREDENTIALS,
            login_blocked: blockedMessage(retryAfter),
          },
        },
      );
    }
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-1">
        <h1 className="text-xl font-semibold">Entrar</h1>
        <p className="text-sm text-muted-foreground">Use o e-mail e a senha da sua conta.</p>
      </div>

      {sessionEnded ? (
        <Alert>
          <AlertDescription>Sua sessão foi encerrada. Entre de novo para continuar.</AlertDescription>
        </Alert>
      ) : null}

      <form className="flex flex-col gap-4" noValidate onSubmit={(event) => void handleSubmit(onValid)(event)}>
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

        <Field invalid={!!errors.password}>
          <FieldLabel>Senha</FieldLabel>
          <Input
            type="password"
            autoComplete="current-password"
            {...register("password", { required: "Informe a senha." })}
          />
          <FieldError>{errors.password?.message}</FieldError>
        </Field>

        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Entrando…" : "Entrar"}
        </Button>
      </form>

      <Link
        href={routes.recuperarAcesso}
        className="text-sm text-muted-foreground underline underline-offset-4 hover:text-foreground"
      >
        Esqueci minha senha
      </Link>
    </div>
  );
}
