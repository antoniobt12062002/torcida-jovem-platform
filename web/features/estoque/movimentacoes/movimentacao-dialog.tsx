"use client";

import { useState } from "react";
import { Controller, useForm } from "react-hook-form";

import { apiErrorMessage } from "@/components/app/api-error";
import { formatDateTime } from "@/components/app/format";
import { QuantityField } from "@/components/app/quantity-field";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { toast } from "@/components/ui/toast";
import { isApiError, networkError } from "@/lib/api/problem";
import { applyProblemToForm } from "@/lib/forms/apply-problem";

import { estoqueCatalog, origemLabels, tipoLabels } from "../errors";
import type { Produto } from "../produtos/hooks";
import { type Movimentacao, useRegistrarMovimentacao } from "./hooks";
import { ORIGEM_MANUAL } from "./payload";

// Entrada, saída e devolução manuais (EWB-02 AC2-5). A origem é sempre
// Inventário, mostrada como texto e nunca oferecida como escolha (WEB-D-007).

export type MovimentacaoDialogProps =
  | { produto: Produto; tipo: "ENTRADA" | "SAIDA"; referencia?: undefined }
  | { produto: Produto; tipo: "DEVOLUCAO"; referencia: Movimentacao };

const TEXTS = {
  ENTRADA: { trigger: "Registrar entrada", title: "Registrar entrada" },
  SAIDA: { trigger: "Registrar saída", title: "Registrar saída" },
  DEVOLUCAO: { trigger: "Devolver", title: "Registrar devolução" },
} as const;

type Values = { quantidade: number | null };

export function MovimentacaoDialog(props: MovimentacaoDialogProps) {
  const { produto, tipo } = props;
  const texts = TEXTS[tipo];
  const [open, setOpen] = useState(false);
  const mutation = useRegistrarMovimentacao();
  const {
    control,
    handleSubmit,
    setError,
    reset,
    formState: { errors, isSubmitting },
  } = useForm<Values>({ defaultValues: { quantidade: null } });

  function changeOpen(next: boolean) {
    if (!next && isSubmitting) return;
    setOpen(next);
    if (!next) reset({ quantidade: null });
  }

  async function onValid({ quantidade }: Values) {
    if (quantidade === null) return;
    try {
      await mutation.mutateAsync(
        props.tipo === "DEVOLUCAO"
          ? { tipo: "DEVOLUCAO", produtoId: produto.id, quantidade, movimentacaoDeId: props.referencia.id }
          : { tipo: props.tipo, produtoId: produto.id, quantidade },
      );
    } catch (cause) {
      const error = isApiError(cause) ? cause : networkError(cause);
      if (error.errors.length > 0) {
        applyProblemToForm(error, setError, { quantidade: "quantidade" }, { catalog: estoqueCatalog });
      } else {
        setError("root", { type: error.code, message: apiErrorMessage(error, estoqueCatalog) });
      }
      return;
    }
    toast.success("Movimentação registrada.");
    setOpen(false);
    reset({ quantidade: null });
  }

  return (
    <Dialog open={open} onOpenChange={changeOpen}>
      <DialogTrigger render={<Button variant={tipo === "DEVOLUCAO" ? "outline" : "default"} size={tipo === "DEVOLUCAO" ? "sm" : "default"} />}>
        {texts.trigger}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{texts.title}</DialogTitle>
          <DialogDescription>
            {props.tipo === "DEVOLUCAO"
              ? `Devolução da ${tipoLabels[props.referencia.tipo].toLowerCase()} de ${Math.abs(props.referencia.quantidade)} ${produto.unidade_medida} em ${formatDateTime(props.referencia.criado_em)}.`
              : produto.nome}
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

          <Controller
            control={control}
            name="quantidade"
            rules={{ validate: (v) => v !== null || "Informe a quantidade." }}
            render={({ field }) => (
              <QuantityField
                {...field}
                mode="positive"
                label="Quantidade"
                description={`Em ${produto.unidade_medida}.`}
                error={errors.quantidade?.message}
                disabled={isSubmitting}
              />
            )}
          />

          <div className="flex flex-col gap-1 text-sm">
            <span className="font-medium">Origem</span>
            <span>{origemLabels[ORIGEM_MANUAL]}</span>
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" disabled={isSubmitting} onClick={() => changeOpen(false)}>
              Cancelar
            </Button>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Registrando…" : "Registrar"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
