"use client";

import { ArrowDownToLineIcon, ArrowUpFromLineIcon, type LucideIcon, Undo2Icon } from "lucide-react";
import { useState } from "react";
import { Controller, useForm } from "react-hook-form";

import { apiErrorMessage } from "@/components/app/api-error";
import { formatDateTime } from "@/components/app/format";
import { FormErrorAlert } from "@/components/app/form-error-alert";
import { QuantityField } from "@/components/app/quantity-field";
import { rowActionName } from "@/components/app/row-actions";
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
import { useFocusInvalidAfterSubmit } from "@/lib/forms/use-focus-invalid-after-submit";

import { estoqueCatalog, origemLabels, tipoLabels } from "../errors";
import type { Produto } from "../produtos/hooks";
import { type Movimentacao, useRegistrarMovimentacao } from "./hooks";
import { ORIGEM_MANUAL } from "./payload";

// Entrada, saída e devolução manuais (EWB-02 AC2-5). A origem é sempre
// Inventário, mostrada como texto e nunca oferecida como escolha (WEB-D-007).
// Abaixo de 640 px o diálogo ocupa a tela, com as ações presas à base.

export type MovimentacaoDialogProps =
  | { produto: Produto; tipo: "ENTRADA" | "SAIDA"; referencia?: undefined }
  | { produto: Produto; tipo: "DEVOLUCAO"; referencia: Movimentacao };

const TEXTS = {
  ENTRADA: { trigger: "Registrar entrada", title: "Registrar entrada" },
  SAIDA: { trigger: "Registrar saída", title: "Registrar saída" },
  DEVOLUCAO: { trigger: "Devolver", title: "Registrar devolução" },
} as const;

const ICONS: Record<MovimentacaoDialogProps["tipo"], LucideIcon> = {
  ENTRADA: ArrowDownToLineIcon,
  SAIDA: ArrowUpFromLineIcon,
  DEVOLUCAO: Undo2Icon,
};

/**
 * Identifica uma movimentação em texto ("saída de 4 UN em 04/10/2026 09:00"):
 * nome das ações da linha no histórico (DSP-05 AC10) e descrição da devolução.
 */
export function movimentacaoLabel(mov: Movimentacao, unidade: string): string {
  return `${tipoLabels[mov.tipo].toLowerCase()} de ${Math.abs(mov.quantidade)} ${unidade} em ${formatDateTime(mov.criado_em)}`;
}

type Values = { quantidade: number | null };

export function MovimentacaoDialog(props: MovimentacaoDialogProps) {
  const { produto, tipo } = props;
  const texts = TEXTS[tipo];
  const [open, setOpen] = useState(false);
  const mutation = useRegistrarMovimentacao();
  const form = useForm<Values>({ defaultValues: { quantidade: null } });
  const {
    control,
    handleSubmit,
    setError,
    reset,
    formState: { errors, isSubmitting },
  } = form;
  // O campo fica desabilitado durante o envio; no erro, o foco volta ao
  // diálogo: no campo inválido ou no alerta (N2 do smoke T13).
  useFocusInvalidAfterSubmit(form, ["quantidade"]);

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

  const Icon = ICONS[tipo];
  const trigger =
    props.tipo === "DEVOLUCAO" ? (
      <Button
        variant="secondary"
        size="sm"
        aria-label={rowActionName(texts.trigger, movimentacaoLabel(props.referencia, produto.unidade_medida))}
      />
    ) : (
      <Button variant={tipo === "ENTRADA" ? "primary" : "secondary"} />
    );

  return (
    <Dialog open={open} onOpenChange={changeOpen}>
      <DialogTrigger render={trigger}>
        <Icon aria-hidden />
        {texts.trigger}
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{texts.title}</DialogTitle>
          <DialogDescription>
            {props.tipo === "DEVOLUCAO"
              ? `Devolução da ${movimentacaoLabel(props.referencia, produto.unidade_medida)}.`
              : produto.nome}
          </DialogDescription>
        </DialogHeader>
        <form
          className="flex flex-1 flex-col gap-4"
          noValidate
          onSubmit={(event) => void handleSubmit(onValid)(event)}
        >
          {errors.root ? <FormErrorAlert>{errors.root.message}</FormErrorAlert> : null}

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

          <dl className="flex flex-col gap-1 text-sm">
            <dt className="font-medium">Origem</dt>
            <dd>{origemLabels[ORIGEM_MANUAL]}</dd>
          </dl>

          <DialogFooter>
            <Button type="button" variant="secondary" disabled={isSubmitting} onClick={() => changeOpen(false)}>
              Cancelar
            </Button>
            <Button type="submit" loading={isSubmitting} loadingLabel="Registrando…">
              Registrar
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
