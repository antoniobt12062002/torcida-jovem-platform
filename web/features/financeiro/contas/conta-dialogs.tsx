"use client";

import { useForm } from "react-hook-form";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { toast } from "@/components/ui/toast";
import { isApiError, networkError } from "@/lib/api/problem";
import { applyProblemToForm } from "@/lib/forms/apply-problem";

import { contasFormCatalog, tipoContaLabel } from "./errors";
import { type Conta, useCreateConta, useRenameConta } from "./hooks";

// Diálogos de criar conta (raiz ou subconta) e de renomear (FWB-01 AC2-4, AC6).

const NOME_MAX = 200;
const TIPOS = ["RECEITA", "DESPESA"] as const;

const nomeRules = {
  validate: (value: string) => value.trim() !== "" || "Informe o nome.",
  maxLength: { value: NOME_MAX, message: `Use no máximo ${NOME_MAX} caracteres.` },
};

function toApiError(cause: unknown) {
  return isApiError(cause) ? cause : networkError(cause);
}

function RootError({ message }: { message?: string }) {
  if (!message) return null;
  return (
    <Alert variant="destructive">
      <AlertDescription>{message}</AlertDescription>
    </Alert>
  );
}

type CreateValues = { nome: string; tipo: Conta["tipo"] | "" };

export type CreateContaDialogProps = {
  /** Conta pai; sem ela, cria uma conta raiz. */
  parent?: Conta;
  onClose: () => void;
};

export function CreateContaDialog({ parent, onClose }: CreateContaDialogProps) {
  const mutation = useCreateConta();
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<CreateValues>({ defaultValues: { nome: "", tipo: parent?.tipo ?? "" } });

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
        <form className="flex flex-col gap-4" noValidate onSubmit={(event) => void handleSubmit(onValid)(event)}>
          <RootError message={errors.root?.message} />

          <Field invalid={!!errors.nome}>
            <FieldLabel>Nome</FieldLabel>
            <Input autoComplete="off" {...register("nome", nomeRules)} />
            <FieldError>{errors.nome?.message}</FieldError>
          </Field>

          <fieldset className="flex flex-col gap-2" aria-describedby={errors.tipo ? "conta-tipo-erro" : undefined}>
            <legend className="mb-1 text-sm font-medium">Tipo</legend>
            <div className="flex gap-4">
              {TIPOS.map((tipo) => (
                <label key={tipo} className="flex items-center gap-2 text-sm">
                  <input
                    type="radio"
                    value={tipo}
                    disabled={!!parent}
                    {...register("tipo", { required: parent ? false : "Escolha o tipo." })}
                  />
                  {tipoContaLabel[tipo]}
                </label>
              ))}
            </div>
            {errors.tipo ? (
              <p id="conta-tipo-erro" role="alert" className="text-sm text-destructive">
                {errors.tipo.message}
              </p>
            ) : null}
          </fieldset>

          <DialogFooter>
            <DialogClose render={<Button type="button" variant="outline" disabled={isSubmitting} />}>
              Cancelar
            </DialogClose>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Criando…" : "Criar conta"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}

type RenameValues = { nome: string };

export function RenameContaDialog({ conta, onClose }: { conta: Conta; onClose: () => void }) {
  const mutation = useRenameConta();
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isSubmitting },
  } = useForm<RenameValues>({ defaultValues: { nome: conta.nome } });

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
        <form className="flex flex-col gap-4" noValidate onSubmit={(event) => void handleSubmit(onValid)(event)}>
          <RootError message={errors.root?.message} />

          <Field invalid={!!errors.nome}>
            <FieldLabel>Nome</FieldLabel>
            <Input autoComplete="off" {...register("nome", nomeRules)} />
            <FieldError>{errors.nome?.message}</FieldError>
          </Field>

          <DialogFooter>
            <DialogClose render={<Button type="button" variant="outline" disabled={isSubmitting} />}>
              Cancelar
            </DialogClose>
            <Button type="submit" disabled={isSubmitting}>
              {isSubmitting ? "Salvando…" : "Salvar"}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
