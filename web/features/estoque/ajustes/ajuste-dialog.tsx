"use client";

import { useState } from "react";

import { apiErrorMessage } from "@/components/app/api-error";
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
import { Field, FieldLabel } from "@/components/ui/field";
import { Textarea } from "@/components/ui/textarea";
import { toast } from "@/components/ui/toast";
import { type ApiError, isApiError, networkError } from "@/lib/api/problem";
import { cn } from "@/lib/utils";

import { estoqueCatalog } from "../errors";
import type { Produto } from "../produtos/hooks";
import { useAjustarEstoque } from "./hooks";

// Ajuste de estoque (EWB-03): quantidade com sinal e diferente de zero, motivo
// obrigatório e prévia do saldo resultante. Saldo negativo é avisado, sem
// impedir: a API aceita (ADR-007). Nada é enviado antes de confirmar.

export type AjusteDialogProps = { produto: Produto; saldoAtual: number };

export function AjusteDialog({ produto, saldoAtual }: AjusteDialogProps) {
  const [open, setOpen] = useState(false);
  const [quantidade, setQuantidade] = useState<number | null>(null);
  const [motivo, setMotivo] = useState("");
  const [error, setError] = useState<ApiError | null>(null);
  const mutation = useAjustarEstoque();

  const pending = mutation.isPending;
  const motivoAparado = motivo.trim();
  const canConfirm = !pending && quantidade !== null && quantidade !== 0 && motivoAparado !== "";
  const resultante = quantidade === null ? null : saldoAtual + quantidade;
  const unidade = produto.unidade_medida;

  function clear() {
    setQuantidade(null);
    setMotivo("");
    setError(null);
  }

  function changeOpen(next: boolean) {
    if (!next && pending) return;
    setOpen(next);
    if (!next) clear();
  }

  async function confirm() {
    if (!canConfirm || quantidade === null) return;
    setError(null);
    try {
      await mutation.mutateAsync({ produto_id: produto.id, quantidade, motivo: motivoAparado });
    } catch (cause) {
      setError(isApiError(cause) ? cause : networkError(cause));
      return;
    }
    toast.success("Ajuste registrado.");
    setOpen(false);
    clear();
  }

  return (
    <Dialog open={open} onOpenChange={changeOpen}>
      <DialogTrigger render={<Button variant="outline" />}>Ajustar estoque</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Ajustar estoque</DialogTitle>
          <DialogDescription>
            Corrige o saldo de {produto.nome}. Use sinal negativo para reduzir, por exemplo -3.
          </DialogDescription>
        </DialogHeader>

        <div className="flex flex-col gap-4">
          {error ? (
            <Alert variant="destructive">
              <AlertDescription>{apiErrorMessage(error, estoqueCatalog)}</AlertDescription>
            </Alert>
          ) : null}

          <QuantityField
            mode="nonZero"
            label="Quantidade"
            description={`Em ${unidade}, com sinal.`}
            value={quantidade}
            onChange={setQuantidade}
            disabled={pending}
            required
          />

          <Field>
            <FieldLabel>Motivo</FieldLabel>
            <Textarea
              value={motivo}
              required
              disabled={pending}
              onValueChange={(value) => setMotivo(String(value))}
            />
          </Field>

          <dl className="grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm">
            <dt className="text-muted-foreground">Saldo atual</dt>
            <dd data-testid="ajuste-saldo-atual" className="tabular-nums">{`${saldoAtual} ${unidade}`}</dd>
            <dt className="text-muted-foreground">Saldo resultante</dt>
            <dd
              data-testid="ajuste-saldo-resultante"
              className={cn("tabular-nums", resultante !== null && resultante < 0 && "font-semibold text-destructive")}
            >
              {resultante === null ? "—" : `${resultante} ${unidade}`}
            </dd>
          </dl>

          {resultante !== null && resultante < 0 ? (
            <Alert>
              <AlertDescription>
                O saldo ficará negativo. O ajuste é permitido; confira a quantidade antes de confirmar.
              </AlertDescription>
            </Alert>
          ) : null}
        </div>

        <DialogFooter>
          <Button type="button" variant="outline" disabled={pending} onClick={() => changeOpen(false)}>
            Cancelar
          </Button>
          <Button type="button" disabled={!canConfirm} onClick={() => void confirm()}>
            Confirmar ajuste
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
