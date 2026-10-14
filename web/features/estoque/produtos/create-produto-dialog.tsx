"use client";

import { useState } from "react";
import { useForm } from "react-hook-form";

import { PlusIcon } from "lucide-react";

import { apiErrorMessage } from "@/components/app/api-error";
import { FormLayout } from "@/components/app/form-layout";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { toast } from "@/components/ui/toast";
import { isApiError, networkError } from "@/lib/api/problem";
import { applyProblemToForm } from "@/lib/forms/apply-problem";
import { useFocusInvalidAfterSubmit } from "@/lib/forms/use-focus-invalid-after-submit";

import { estoqueCatalog, MESSAGE_CODIGO_DUPLICADO } from "../errors";
import { type CreateProdutoRequest, useCreateProduto } from "./hooks";

// Criação de produto (EWB-01 AC2-3), no layout de formulário (DSP-04).

const EMPTY: CreateProdutoRequest = { codigo: "", nome: "", unidade_medida: "" };

export function CreateProdutoDialog() {
  const [open, setOpen] = useState(false);
  const mutation = useCreateProduto();
  const form = useForm<CreateProdutoRequest>({ defaultValues: EMPTY });
  const {
    register,
    handleSubmit,
    setError,
    reset,
    formState: { errors, isSubmitting },
  } = form;
  // No erro de campo da API, o foco vai ao primeiro campo inválido (DSP-04 AC3).
  useFocusInvalidAfterSubmit(form, ["codigo", "nome", "unidade_medida"]);

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
      <DialogTrigger render={<Button />}>
        <PlusIcon aria-hidden />
        Novo produto
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Novo produto</DialogTitle>
          <DialogDescription>O código identifica o produto e não se repete.</DialogDescription>
        </DialogHeader>
        <FormLayout
          noValidate
          onSubmit={(event) => void handleSubmit(onValid)(event)}
          error={errors.root?.message}
          footer={
            <>
              <Button type="button" variant="secondary" disabled={isSubmitting} onClick={() => changeOpen(false)}>
                Cancelar
              </Button>
              <Button type="submit" loading={isSubmitting} loadingLabel="Criando…">
                Criar produto
              </Button>
            </>
          }
        >
          <Field invalid={!!errors.codigo}>
            <FieldLabel>Código</FieldLabel>
            <Input
              autoComplete="off"
              spellCheck={false}
              autoCapitalize="none"
              {...register("codigo", { validate: required("Informe o código.") })}
            />
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
        </FormLayout>
      </DialogContent>
    </Dialog>
  );
}

function required(message: string) {
  return (value: string) => value.trim() !== "" || message;
}
