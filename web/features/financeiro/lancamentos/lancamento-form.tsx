"use client";

import { Controller, useForm, useWatch } from "react-hook-form";

import { MESSAGE_FORBIDDEN } from "@/components/app/api-error";
import { MoneyField } from "@/components/app/money-field";
import { toast } from "@/components/ui/toast";
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
import { type ErrorCatalog, isApiError, messageFor, networkError } from "@/lib/api/problem";
import { applyProblemToForm } from "@/lib/forms/apply-problem";
import { formatBRL } from "@/lib/money";

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
import { SelectField } from "./select-field";

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
  description?: string;
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
      <DialogContent className="sm:max-w-md">
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
  const contaOptions = contasParaLancamento(contas.data?.items, tipo).map((c) => ({ value: c.id, label: c.nome }));

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
    <form className="flex flex-col gap-4" noValidate onSubmit={(event) => void handleSubmit(onValid)(event)}>
      {errors.root ? (
        <Alert variant="destructive">
          <AlertDescription>{errors.root.message}</AlertDescription>
        </Alert>
      ) : null}

      {tipoEditable ? (
        <Controller
          name="tipo"
          control={control}
          render={({ field }) => (
            <SelectField
              label="Tipo"
              {...field}
              onChange={(value) => {
                field.onChange(value);
                setValue("conta_id", "");
              }}
              options={TIPOS.map((t) => ({ value: t, label: TIPO_LABEL[t] }))}
              error={errors.tipo?.message}
            />
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
          <SelectField
            label="Conta"
            {...field}
            placeholder="Selecione a conta"
            options={contaOptions}
            error={fieldState.error?.message}
          />
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
          {bruto !== null && taxa !== null ? formatBRL(bruto - taxa) : "—"}
        </span>
      </p>

      <Controller
        name="forma_pagamento"
        control={control}
        rules={{ required: "Escolha a forma de pagamento." }}
        render={({ field, fieldState }) => (
          <SelectField
            label="Forma de pagamento"
            {...field}
            placeholder="Selecione a forma de pagamento"
            options={FORMAS_PAGAMENTO.map((f) => ({ value: f, label: FORMA_PAGAMENTO_LABEL[f] }))}
            error={fieldState.error?.message}
          />
        )}
      />

      <DialogFooter>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Salvando…" : submitLabel}
        </Button>
      </DialogFooter>
    </form>
  );
}
