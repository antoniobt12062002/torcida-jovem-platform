"use client";

import { CircleAlertIcon } from "lucide-react";
import { useForm } from "react-hook-form";

import { FormLayout } from "@/components/app/form-layout";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Field, FieldError, FieldLabel, FieldLegend, FieldSet } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { toast } from "@/components/ui/toast";
import { isApiError, networkError } from "@/lib/api/problem";
import { applyProblemToForm } from "@/lib/forms/apply-problem";
import { useFocusInvalidAfterSubmit } from "@/lib/forms/use-focus-invalid-after-submit";

import { contasFormCatalog, tipoContaLabel } from "./errors";
import { type Conta, useCreateConta, useRenameConta } from "./hooks";

// Diálogos de criar conta (raiz ou subconta) e de renomear (FWB-01 AC2-4, AC6),
// no padrão de formulário (DSP-04): erro sem campo no topo, ações no rodapé.

const NOME_MAX = 200;
const TIPOS = ["RECEITA", "DESPESA"] as const;

const nomeRules = {
  validate: (value: string) => value.trim() !== "" || "Informe o nome.",
  maxLength: { value: NOME_MAX, message: `Use no máximo ${NOME_MAX} caracteres.` },
};

function toApiError(cause: unknown) {
  return isApiError(cause) ? cause : networkError(cause);
}

type CreateValues = { nome: string; tipo: Conta["tipo"] | "" };

export type CreateContaDialogProps = {
  /** Conta pai; sem ela, cria uma conta raiz. */
  parent?: Conta;
  onClose: () => void;
};

export function CreateContaDialog({ parent, onClose }: CreateContaDialogProps) {
  const mutation = useCreateConta();
  const form = useForm<CreateValues>({ defaultValues: { nome: "", tipo: parent?.tipo ?? "" } });
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = form;
  // No erro de campo da API, o foco vai ao primeiro campo inválido (DSP-04 AC3).
  useFocusInvalidAfterSubmit(form, ["nome", "tipo"]);

  const title = parent ? `Nova subconta em ${parent.nome}` : "Nova conta";

  async function onValid(values: CreateValues) {
    // A subconta tem sempre o tipo do pai (FWB-01 AC3).
    const tipo = parent ? parent.tipo : (values.tipo as Conta["tipo"]);
    const nome = values.nome.trim();
    try {
      await mutation.mutateAsync(parent ? { tipo, nome, parent_id: parent.id } : { tipo, nome });
    } catch (cause) {
      applyProblemToForm(toApiError(cause), setError, { nome: "nome", tipo: "tipo" }, { catalog: contasFormCatalog });
      return;
    }
    toast.success("Conta criada.");
    onClose();
  }

  return (
    <Dialog open onOpenChange={(open) => !open && !isSubmitting && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{title}</DialogTitle>
          {parent ? (
            <DialogDescription>A subconta tem sempre o mesmo tipo da conta pai.</DialogDescription>
          ) : null}
        </DialogHeader>
        <FormLayout
          noValidate
          onSubmit={(event) => void handleSubmit(onValid)(event)}
          error={errors.root?.message}
          footer={
            <>
              <DialogClose render={<Button type="button" variant="secondary" disabled={isSubmitting} />}>
                Cancelar
              </DialogClose>
              <Button type="submit" loading={isSubmitting} loadingLabel="Criando…">
                Criar conta
              </Button>
            </>
          }
        >
          <Field invalid={!!errors.nome}>
            <FieldLabel>Nome</FieldLabel>
            <Input autoComplete="off" {...register("nome", nomeRules)} />
            <FieldError>{errors.nome?.message}</FieldError>
          </Field>

          <FieldSet className="gap-2" aria-describedby={errors.tipo ? "conta-tipo-erro" : undefined}>
            <FieldLegend variant="label">Tipo</FieldLegend>
            <div className="flex flex-wrap gap-x-6 gap-y-2">
              {TIPOS.map((tipo) => (
                <label key={tipo} className="flex min-h-control-sm items-center gap-2 text-sm">
                  <input
                    type="radio"
                    value={tipo}
                    disabled={!!parent}
                    className="size-4 accent-primary"
                    {...register("tipo", { required: parent ? false : "Escolha o tipo." })}
                  />
                  {tipoContaLabel[tipo]}
                </label>
              ))}
            </div>
            {errors.tipo ? (
              // Mesmo desenho de FieldError, que só funciona dentro de um Field de controle único.
              <p id="conta-tipo-erro" role="alert" className="flex items-start gap-2 text-sm text-destructive">
                <CircleAlertIcon aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
                {errors.tipo.message}
              </p>
            ) : null}
          </FieldSet>
        </FormLayout>
      </DialogContent>
    </Dialog>
  );
}

type RenameValues = { nome: string };

export function RenameContaDialog({ conta, onClose }: { conta: Conta; onClose: () => void }) {
  const mutation = useRenameConta();
  const form = useForm<RenameValues>({ defaultValues: { nome: conta.nome } });
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = form;
  // No erro de campo da API, o foco vai ao primeiro campo inválido (DSP-04 AC3).
  useFocusInvalidAfterSubmit(form, ["nome"]);

  async function onValid(values: RenameValues) {
    try {
      await mutation.mutateAsync({ id: conta.id, nome: values.nome.trim() });
    } catch (cause) {
      applyProblemToForm(toApiError(cause), setError, { nome: "nome" }, { catalog: contasFormCatalog });
      return;
    }
    toast.success("Conta renomeada.");
    onClose();
  }

  return (
    <Dialog open onOpenChange={(open) => !open && !isSubmitting && onClose()}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Renomear conta</DialogTitle>
          <DialogDescription>
            Só contas que nunca foram usadas em um lançamento podem ser renomeadas.
          </DialogDescription>
        </DialogHeader>
        <FormLayout
          noValidate
          onSubmit={(event) => void handleSubmit(onValid)(event)}
          error={errors.root?.message}
          footer={
            <>
              <DialogClose render={<Button type="button" variant="secondary" disabled={isSubmitting} />}>
                Cancelar
              </DialogClose>
              <Button type="submit" loading={isSubmitting} loadingLabel="Salvando…">
                Salvar
              </Button>
            </>
          }
        >
          <Field invalid={!!errors.nome}>
            <FieldLabel>Nome</FieldLabel>
            <Input autoComplete="off" {...register("nome", nomeRules)} />
            <FieldError>{errors.nome?.message}</FieldError>
          </Field>
        </FormLayout>
      </DialogContent>
    </Dialog>
  );
}
