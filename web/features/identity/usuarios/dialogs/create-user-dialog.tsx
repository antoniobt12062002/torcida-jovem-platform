"use client";

import { useForm } from "react-hook-form";

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
import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { toast } from "@/components/ui/toast";
import { isApiError, networkError } from "@/lib/api/problem";
import { applyProblemToForm } from "@/lib/forms/apply-problem";

import { usersCatalog } from "../errors";
import { useCreateUser } from "../hooks";

// Criação de usuário (USR-02). A API cria a pessoa só com ASSOCIADO e com
// troca de senha obrigatória no primeiro acesso.

type CreateUserValues = { name: string; email: string; password: string };

const EMPTY: CreateUserValues = { name: "", email: "", password: "" };

/** Erros da API que pertencem a um campo (o corpo traz só o `code`). */
const FIELD_OF_CODE: Readonly<Record<string, keyof CreateUserValues>> = {
  email_taken: "email",
  invalid_email: "email",
  invalid_name: "name",
  password_too_short: "password",
  password_too_long: "password",
  password_compromised: "password",
};

export const MESSAGE_USER_CREATED =
  "Usuário criado. A pessoa terá de trocar a senha no primeiro acesso.";

export function CreateUserDialog({
  open,
  onOpenChange,
}: {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}) {
  const mutation = useCreateUser();
  const {
    register,
    handleSubmit,
    setError,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<CreateUserValues>({ defaultValues: EMPTY });

  function close() {
    reset(EMPTY);
    mutation.reset();
    onOpenChange(false);
  }

  async function onValid(values: CreateUserValues) {
    try {
      await mutation.mutateAsync({ name: values.name, email: values.email, password: values.password });
    } catch (cause) {
      const error = isApiError(cause) ? cause : networkError(cause);
      const field = Object.hasOwn(FIELD_OF_CODE, error.code) ? FIELD_OF_CODE[error.code] : undefined;
      if (field) {
        setError(field, { type: error.code, message: usersCatalog[error.code] });
        return;
      }
      applyProblemToForm(
        error,
        setError,
        { name: "name", email: "email", password: "password" },
        { catalog: usersCatalog },
      );
      return;
    }
    toast.success(MESSAGE_USER_CREATED);
    close();
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(next) => {
        if (next) onOpenChange(true);
        else if (!isSubmitting) close();
      }}
    >
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Novo usuário</DialogTitle>
          <DialogDescription>
            A pessoa entra só com o papel Associado e terá de trocar a senha no primeiro acesso.
          </DialogDescription>
        </DialogHeader>
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

          <Field invalid={!!errors.name}>
            <FieldLabel>Nome</FieldLabel>
            <Input autoComplete="off" {...register("name", { required: "Informe o nome." })} />
            <FieldError>{errors.name?.message}</FieldError>
          </Field>

          <Field invalid={!!errors.email}>
            <FieldLabel>E-mail</FieldLabel>
            <Input type="email" autoComplete="off" {...register("email", { required: "Informe o e-mail." })} />
            <FieldError>{errors.email?.message}</FieldError>
          </Field>

          <Field invalid={!!errors.password}>
            <FieldLabel>Senha inicial</FieldLabel>
            <Input
              type="password"
              autoComplete="new-password"
              {...register("password", { required: "Informe a senha inicial." })}
            />
            <FieldDescription>Mínimo de 8 caracteres.</FieldDescription>
            <FieldError>{errors.password?.message}</FieldError>
          </Field>

          <DialogFooter>
            <Button type="button" variant="outline" disabled={isSubmitting} onClick={close}>
              Cancelar
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Criando…" : "Criar usuário"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
