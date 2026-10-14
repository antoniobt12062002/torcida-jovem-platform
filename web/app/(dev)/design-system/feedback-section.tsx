"use client";

import { BanIcon } from "lucide-react";

import { ConfirmDialog } from "@/components/app/confirm-dialog";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { toast } from "@/components/ui/toast";

import { SpecimenBlock, SpecimenSection } from "./section";

// Feedback (DSP-03): alertas nos quatro tons, um botão para cada toast e a
// confirmação com motivo obrigatório. Nada chama a API.

const WAIT_MS = 800;

export function FeedbackSection() {
  return (
    <SpecimenSection id="feedback" title="Feedback" description="Cada tom tem ícone fixo: o estado nunca depende só da cor.">
      <SpecimenBlock title="Alert">
        <div className="flex flex-col gap-3">
          <Alert variant="neutral">
            <AlertTitle>Prestação de contas de março</AlertTitle>
            <AlertDescription>O relatório fica disponível após a reunião do conselho.</AlertDescription>
          </Alert>
          <Alert variant="success">
            <AlertTitle>Lançamento criado</AlertTitle>
            <AlertDescription>A receita já aparece no saldo da conta Associações.</AlertDescription>
          </Alert>
          <Alert variant="warning">
            <AlertTitle>Comprovante pendente</AlertTitle>
            <AlertDescription>Anexe o comprovante antes de fechar o mês.</AlertDescription>
          </Alert>
          <Alert variant="error">
            <AlertTitle>Não foi possível salvar</AlertTitle>
            <AlertDescription>Confira a conexão e tente de novo.</AlertDescription>
          </Alert>
        </div>
      </SpecimenBlock>

      <SpecimenBlock title="Toast">
        <div className="flex flex-wrap gap-3">
          <Button variant="secondary" onClick={() => toast.show("Relatório em preparação.")}>
            Toast neutro
          </Button>
          <Button variant="secondary" onClick={() => toast.success("Lançamento criado.")}>
            Toast de sucesso
          </Button>
          <Button
            variant="secondary"
            onClick={() => toast.warning("Comprovante pendente.", { description: "Anexe antes de fechar o mês." })}
          >
            Toast de aviso
          </Button>
          <Button
            variant="secondary"
            onClick={() => toast.error("Não foi possível salvar.", { description: "Tente de novo em instantes." })}
          >
            Toast de erro
          </Button>
        </div>
      </SpecimenBlock>

      <SpecimenBlock title="Confirmação com motivo (ConfirmDialog)">
        <div>
          <ConfirmDialog
            title="Cancelar lançamento?"
            subject="Receita fb6e3d8f · Mensalidades de março"
            consequence="O lançamento sai do saldo e fica no histórico como cancelado. Esta ação não pode ser desfeita."
            confirmLabel="Cancelar lançamento"
            cancelLabel="Voltar"
            pendingLabel="Cancelando…"
            destructive
            reason="required"
            onConfirm={() =>
              new Promise<void>((resolve) => {
                setTimeout(() => {
                  toast.success("Lançamento cancelado (exemplo).");
                  resolve();
                }, WAIT_MS);
              })
            }
            trigger={
              <Button variant="destructive">
                <BanIcon aria-hidden="true" />
                Cancelar lançamento
              </Button>
            }
          />
        </div>
      </SpecimenBlock>
    </SpecimenSection>
  );
}
