"use client";

import { type FormEvent, useState } from "react";

import { FormLayout } from "@/components/app/form-layout";
import { MoneyField } from "@/components/app/money-field";
import { RowActions } from "@/components/app/row-actions";
import { Button } from "@/components/ui/button";
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { NativeSelect } from "@/components/ui/native-select";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { toast } from "@/components/ui/toast";

import { SpecimenBlock, SpecimenCaption, SpecimenSection } from "./section";

// Sobreposições: diálogo de formulário com FormLayout, gaveta inferior e menu
// de ações de linha. Foco preso, Esc fecha e o foco volta ao gatilho.

function NewEntryDialog() {
  const [open, setOpen] = useState(false);
  const [descricao, setDescricao] = useState("");
  const [valor, setValor] = useState<number | null>(null);
  const [submitted, setSubmitted] = useState(false);
  const missing = descricao.trim() === "";

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitted(true);
    if (missing || valor === null) return;
    setOpen(false);
    setSubmitted(false);
    setDescricao("");
    setValor(null);
    toast.success("Lançamento criado (exemplo).");
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button />}>Novo lançamento</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Novo lançamento</DialogTitle>
          <DialogDescription>Exemplo sem envio: nada é gravado.</DialogDescription>
        </DialogHeader>
        <FormLayout
          noValidate
          onSubmit={handleSubmit}
          error={submitted && (missing || valor === null) ? "Corrija os campos destacados." : undefined}
          footer={
            <>
              <DialogClose render={<Button variant="secondary" type="button" />}>Cancelar</DialogClose>
              <Button type="submit">Salvar</Button>
            </>
          }
        >
          <Field invalid={submitted && missing}>
            <FieldLabel>Descrição</FieldLabel>
            <Input value={descricao} onChange={(event) => setDescricao(event.target.value)} />
            {submitted && missing ? <FieldError>Informe a descrição.</FieldError> : null}
          </Field>
          <MoneyField
            label="Valor"
            value={valor}
            onChange={setValor}
            error={submitted && valor === null ? "Informe o valor." : undefined}
          />
          <Field>
            <FieldLabel optional>Conta</FieldLabel>
            <NativeSelect defaultValue="">
              <option value="">Selecione a conta</option>
              <option value="associacoes">Associações</option>
              <option value="loja">Loja</option>
            </NativeSelect>
          </Field>
        </FormLayout>
      </DialogContent>
    </Dialog>
  );
}

export function OverlaysSection() {
  return (
    <SpecimenSection
      id="sobreposicoes"
      title="Sobreposições"
      description="Abaixo de 640 px o diálogo ocupa a tela inteira, com o rodapé preso ao diálogo."
    >
      <div className="grid grid-cols-1 gap-6 md:grid-cols-3">
        <SpecimenBlock title="Diálogo com formulário">
          <div>
            <NewEntryDialog />
          </div>
          <SpecimenCaption>Dialog com FormLayout: erro sem campo no topo, ações no rodapé.</SpecimenCaption>
        </SpecimenBlock>

        <SpecimenBlock title="Gaveta inferior">
          <div>
            <Sheet>
              <SheetTrigger render={<Button variant="secondary" />}>Abrir gaveta</SheetTrigger>
              <SheetContent side="bottom">
                <SheetHeader>
                  <SheetTitle>Resumo do mês</SheetTitle>
                  <SheetDescription>Receitas e despesas de março (exemplo).</SheetDescription>
                </SheetHeader>
                <p className="px-4 text-sm">A gaveta inferior guarda os filtros em tela estreita e conteúdos curtos.</p>
                <SheetFooter>
                  <SheetClose render={<Button />}>Fechar</SheetClose>
                </SheetFooter>
              </SheetContent>
            </Sheet>
          </div>
          <SpecimenCaption>Sheet side=&quot;bottom&quot;.</SpecimenCaption>
        </SpecimenBlock>

        <SpecimenBlock title="Menu de ações de linha">
          <RowActions
            rowLabel="lançamento Venda de bonés"
            actions={[
              { label: "Editar", onSelect: () => toast.show("Exemplo: editar.") },
              { label: "Anexar comprovante", onSelect: () => toast.show("Exemplo: anexar.") },
              { label: "Cancelar lançamento", destructive: true, onSelect: () => toast.warning("Exemplo: cancelar.") },
            ]}
          />
          <SpecimenCaption>DropdownMenu pelo RowActions: a ação destrutiva vai por último, após o separador.</SpecimenCaption>
        </SpecimenBlock>
      </div>
    </SpecimenSection>
  );
}
