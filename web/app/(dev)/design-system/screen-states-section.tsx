"use client";

import { PlusIcon } from "lucide-react";
import Link from "next/link";

import { EmptyState, ErrorState, LoadingState, type LoadingShape } from "@/components/app/states";
import { Button, buttonVariants } from "@/components/ui/button";
import { toast } from "@/components/ui/toast";

import { SpecimenBlock, SpecimenCaption, SpecimenSection } from "./section";

// Estados de tela (DSP-02): ocupam o lugar do conteúdo. "Sem acesso" traz o
// `h1` da página, por isso é mostrado numa página própria do espécime.

const SHAPES: readonly LoadingShape[] = ["table", "list", "detail", "value"];

export function ScreenStatesSection() {
  return (
    <SpecimenSection id="estados-de-tela" title="Estados de tela" description="Carregando, vazio, sem resultado, erro e sem acesso.">
      <SpecimenBlock title="Carregando (LoadingState)">
        <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
          {SHAPES.map((shape) => (
            <div key={shape} className="flex flex-col gap-1 rounded-lg border px-4">
              <LoadingState shape={shape} label={`Carregando (forma ${shape})…`} />
              <SpecimenCaption>shape=&quot;{shape}&quot;</SpecimenCaption>
            </div>
          ))}
        </div>
      </SpecimenBlock>

      <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
        <SpecimenBlock title="Vazio">
          <div className="rounded-lg border">
            <EmptyState
              headingLevel={3}
              title="Nenhum lançamento ainda"
              description="Registre a primeira receita ou despesa da torcida."
              action={
                <Button onClick={() => toast.show("Exemplo: o formulário fica na seção Sobreposições.")}>
                  <PlusIcon aria-hidden="true" />
                  Novo lançamento
                </Button>
              }
            />
          </div>
        </SpecimenBlock>
        <SpecimenBlock title="Sem resultado">
          <div className="rounded-lg border">
            <EmptyState
              kind="no-results"
              headingLevel={3}
              description="Mude a busca ou os filtros."
              onClear={() => toast.show("Exemplo: filtros limpos.")}
            />
          </div>
        </SpecimenBlock>
      </div>

      <SpecimenBlock title="Erro">
        <ErrorState error={new Error("exemplo")} onRetry={() => toast.show("Exemplo: nova tentativa.")} />
      </SpecimenBlock>

      <SpecimenBlock title="Sem acesso">
        <p className="text-sm text-muted-foreground">
          O estado &quot;Sem acesso&quot; traz o único h1 da página. Para não haver dois títulos de página aqui, ele aparece dentro do shell, numa página própria.
        </p>
        <div>
          <Link href="/design-system/sem-acesso" className={buttonVariants({ variant: "secondary" })}>
            Abrir o estado sem acesso
          </Link>
        </div>
      </SpecimenBlock>
    </SpecimenSection>
  );
}
