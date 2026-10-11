"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";

import { apiErrorMessage } from "@/components/app/api-error";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { toast } from "@/components/ui/toast";
import { isApiError, networkError } from "@/lib/api/problem";
import { applyProblemToForm } from "@/lib/forms/apply-problem";

import { estoqueCatalog, MESSAGE_CODIGO_DUPLICADO } from "../errors";
import { type CreateProdutoRequest, useCreateProduto } from "./hooks";

// Criação de produto (EWB-01 AC2-3).

const EMPTY: CreateProdutoRequest = { codigo: "", nome: "", unidade_medida: "" };

export function CreateProdutoDialog() {
  const [open, setOpen] = useState(false);
  const mutation = useCreateProduto();
  const {
    register,
    handleSubmit,
    setError,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<CreateProdutoRequest>({ defaultValues: EMPTY });

  function changeOpen(next: boolean) {
    // Enquanto envia, o diálogo não fecha por fora.
    if (!next && isSubmitting) return;
    setOpen(next);
    if (!next) reset(EMPTY);
  }

  function close() {
    setOpen(false);
    reset(EMPTY);
  }

  async function onValid(values: CreateProdutoRequest) {
    try {
      await mutation.mutateAsync({
        codigo: values.codigo.trim(),
        nome: values.nome.trim(),
        unidade_medida: values.unidade_medida.trim(),
      });
    } catch (cause) {
      const error = isApiError(cause) ? cause : networkError(cause);
      if (error.code === "codigo_duplicado") {
        setError("codigo", { type: error.code, message: MESSAGE_CODIGO_DUPLICADO });
      } else if (error.errors.length > 0) {
        applyProblemToForm(
          error,
          setError,
          { codigo: "codigo", nome: "nome", unidade_medida: "unidade_medida" },
          { catalog: estoqueCatalog },
        );
      } else {
        setError("root", { type: error.code, message: apiErrorMessage(error, estoqueCatalog) });
      }
      return;
    }
    toast.success("Produto criado.");
    close();
  }

  return (
    <Dialog open={open} onOpenChange={changeOpen}>
      <DialogTrigger render={<Button />}>Novo produto</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Novo produto</DialogTitle>
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

          <Field invalid={!!errors.codigo}>
            <FieldLabel>Código</FieldLabel>
            <Input autoComplete="off" {...register("codigo", { validate: required("Informe o código.") })} />
            <FieldError>{errors.codigo?.message}</FieldError>
          </Field>

          <Field invalid={!!errors.nome}>
            <FieldLabel>Nome</FieldLabel>
            <Input autoComplete="off" {...register("nome", { validate: required("Informe o nome.") })} />
            <FieldError>{errors.nome?.message}</FieldError>
          </Field>

          <Field invalid={!!errors.unidade_medida}>
            <FieldLabel>Unidade de medida</FieldLabel>
            <Input
              autoComplete="off"
              placeholder="UN, KG, M…"
              {...register("unidade_medida", { validate: required("Informe a unidade de medida.") })}
            />
            <FieldError>{errors.unidade_medida?.message}</FieldError>
          </Field>

          <DialogFooter>
            <Button type="button" variant="outline" disabled={isSubmitting} onClick={() => changeOpen(false)}>
              Cancelar
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Criando…" : "Criar produto"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

function required(message: string) {
  return (value: string) => value.trim() !== "" || message;
}
