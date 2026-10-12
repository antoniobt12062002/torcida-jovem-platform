"use client";

import type { ReactNode } from "react";
import { Controller, useForm, useWatch } from "react-hook-form";

import { MESSAGE_FORBIDDEN } from "@/components/app/api-error";
import { FormLayout } from "@/components/app/form-layout";
import { MoneyField } from "@/components/app/money-field";
import { MoneyValue } from "@/components/app/money-value";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { NativeSelect } from "@/components/ui/native-select";
import { toast } from "@/components/ui/toast";
import { type ErrorCatalog, isApiError, messageFor, networkError } from "@/lib/api/problem";
import { applyProblemToForm } from "@/lib/forms/apply-problem";

import { FIELD_BY_CODE, lancamentosCatalog } from "./errors";
import { contasParaLancamento } from "./filters";
import { useContasOptions } from "./hooks";
import {
  FORMA_PAGAMENTO_LABEL,
  FORMAS_PAGAMENTO,
  type FormaPagamento,
  TIPO_LABEL,
  type TipoConta,
  TIPOS,
} from "./labels";

// Formulário de valores de um lançamento (criar, editar e devolver), num
// diálogo. Valores em centavos inteiros (MoneyField); o líquido é só prévia,
// o gravado é o que a API devolve (premissa "Valor líquido").

export type LancamentoFormValues = {
  tipo: TipoConta;
  conta_id: string;
  valor_bruto_cents: number | null;
  taxa_cents: number | null;
  forma_pagamento: FormaPagamento | "";
};

export type LancamentoValues = {
  tipo: TipoConta;
  conta_id: string;
  valor_bruto_cents: number;
  taxa_cents: number;
  forma_pagamento: FormaPagamento;
};

const FIELD_MAP = {
  tipo: "tipo",
  conta_id: "conta_id",
  valor_bruto_cents: "valor_bruto_cents",
  taxa_cents: "taxa_cents",
  forma_pagamento: "forma_pagamento",
} as const;

export const NEW_LANCAMENTO: LancamentoFormValues = {
  tipo: "RECEITA",
  conta_id: "",
  valor_bruto_cents: null,
  taxa_cents: 0,
  forma_pagamento: "",
};

export type LancamentoFormDialogProps = {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  /** O registro e o efeito do envio, por exemplo a receita que a devolução referencia. */
  description?: ReactNode;
  defaultValues: LancamentoFormValues;
  /** Com `false`, o tipo fica fixo (edição e devolução). */
  tipoEditable: boolean;
  submitLabel: string;
  successMessage: string;
  /** Mensagens extras por `code`, somadas às de lançamentos. */
  catalog?: ErrorCatalog;
  /** Envia à API; lança ApiError em caso de erro. */
  onSubmit: (values: LancamentoValues) => Promise<unknown>;
};

export function LancamentoFormDialog({ open, onOpenChange, title, description, ...props }: LancamentoFormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {description ? <DialogDescription>{description}</DialogDescription> : null}
        </DialogHeader>
        <LancamentoForm {...props} onDone={() => onOpenChange(false)} />
      </DialogContent>
    </Dialog>
  );
}

type LancamentoFormProps = Omit<LancamentoFormDialogProps, "open" | "onOpenChange" | "title" | "description"> & {
  onDone: () => void;
};

function LancamentoForm({
  defaultValues,
  tipoEditable,
  submitLabel,
  successMessage,
  catalog: extraCatalog,
  onSubmit,
  onDone,
}: LancamentoFormProps) {
  const contas = useContasOptions();
  const catalog: ErrorCatalog = { forbidden: MESSAGE_FORBIDDEN, ...lancamentosCatalog, ...extraCatalog };

  const {
    control,
    handleSubmit,
    setError,
    setValue,
    formState: { errors, isSubmitting },
  } = useForm<LancamentoFormValues>({ defaultValues });

  const tipo = useWatch({ control, name: "tipo" });
  const bruto = useWatch({ control, name: "valor_bruto_cents" });
  const taxa = useWatch({ control, name: "taxa_cents" });
  const contaOptions = contasParaLancamento(contas.data?.items, tipo);

  async function onValid(values: LancamentoFormValues) {
    try {
      await onSubmit({
        tipo: values.tipo,
        conta_id: values.conta_id,
        valor_bruto_cents: values.valor_bruto_cents as number,
        taxa_cents: values.taxa_cents as number,
        forma_pagamento: values.forma_pagamento as FormaPagamento,
      });
    } catch (cause) {
      const error = isApiError(cause) ? cause : networkError(cause);
      const field = Object.hasOwn(FIELD_BY_CODE, error.code)
        ? FIELD_BY_CODE[error.code as keyof typeof FIELD_BY_CODE]
        : undefined;
      if (field) {
        setError(field, { type: error.code, message: messageFor(error, catalog) });
        return;
      }
      applyProblemToForm(error, setError, FIELD_MAP, { catalog });
      return;
    }
    toast.success(successMessage);
    onDone();
  }

  return (
    <FormLayout
      noValidate
      onSubmit={(event) => void handleSubmit(onValid)(event)}
      error={errors.root?.message}
      footer={
        <>
          <DialogClose render={<Button variant="secondary" type="button" />}>Cancelar</DialogClose>
          <Button type="submit" loading={isSubmitting} loadingLabel="Salvando…">
            {submitLabel}
          </Button>
        </>
      }
    >
      {tipoEditable ? (
        <Controller
          name="tipo"
          control={control}
          render={({ field, fieldState }) => (
            <Field invalid={!!fieldState.error}>
              <FieldLabel>Tipo</FieldLabel>
              <NativeSelect
                ref={field.ref}
                name={field.name}
                value={field.value}
                onBlur={field.onBlur}
                onChange={(event) => {
                  field.onChange(event.target.value);
                  setValue("conta_id", "");
                }}
              >
                {TIPOS.map((t) => (
                  <option key={t} value={t}>
                    {TIPO_LABEL[t]}
                  </option>
                ))}
              </NativeSelect>
              {fieldState.error ? <FieldError>{fieldState.error.message}</FieldError> : null}
            </Field>
          )}
        />
      ) : (
        <p className="text-sm">
          <span className="text-muted-foreground">Tipo: </span>
          {TIPO_LABEL[tipo]}
        </p>
      )}

      <Controller
        name="conta_id"
        control={control}
        rules={{ required: "Escolha a conta." }}
        render={({ field, fieldState }) => (
          <Field invalid={!!fieldState.error}>
            <FieldLabel>Conta</FieldLabel>
            <NativeSelect
              ref={field.ref}
              name={field.name}
              value={field.value}
              onBlur={field.onBlur}
              onChange={(event) => field.onChange(event.target.value)}
            >
              <option value="">Selecione a conta</option>
              {contaOptions.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.nome}
                </option>
              ))}
            </NativeSelect>
            {fieldState.error ? <FieldError>{fieldState.error.message}</FieldError> : null}
          </Field>
        )}
      />

      <Controller
        name="valor_bruto_cents"
        control={control}
        rules={{ validate: (v) => v !== null || "Informe o valor bruto." }}
        render={({ field, fieldState }) => (
          <MoneyField label="Valor bruto" {...field} error={fieldState.error?.message} />
        )}
      />

      <Controller
        name="taxa_cents"
        control={control}
        rules={{ validate: (v) => v !== null || "Informe a taxa (use 0,00 se não houver)." }}
        render={({ field, fieldState }) => (
          <MoneyField label="Taxa" {...field} error={fieldState.error?.message} />
        )}
      />

      <p className="text-sm" aria-live="polite">
        <span className="text-muted-foreground">Valor líquido (prévia): </span>
        <span data-testid="previa-liquido">
          {bruto !== null && taxa !== null ? <MoneyValue cents={bruto - taxa} /> : "—"}
        </span>
      </p>

      <Controller
        name="forma_pagamento"
        control={control}
        rules={{ required: "Escolha a forma de pagamento." }}
        render={({ field, fieldState }) => (
          <Field invalid={!!fieldState.error}>
            <FieldLabel>Forma de pagamento</FieldLabel>
            <NativeSelect
              ref={field.ref}
              name={field.name}
              value={field.value}
              onBlur={field.onBlur}
              onChange={(event) => field.onChange(event.target.value)}
            >
              <option value="">Selecione a forma de pagamento</option>
              {FORMAS_PAGAMENTO.map((f) => (
                <option key={f} value={f}>
                  {FORMA_PAGAMENTO_LABEL[f]}
                </option>
              ))}
            </NativeSelect>
            {fieldState.error ? <FieldError>{fieldState.error.message}</FieldError> : null}
          </Field>
        )}
      />
    </FormLayout>
  );
}
