"use client";

import { PlusIcon } from "lucide-react";
import { Suspense } from "react";

import { PageContainer } from "@/components/app/page-container";
import { PageHeader } from "@/components/app/page-header";
import { Button } from "@/components/ui/button";
import { toast } from "@/components/ui/toast";

import { ExampleListing } from "../example-listing";
import { ExampleShell } from "../example-shell";

export function ExamplePage() {
  return (
    <ExampleShell>
      <PageContainer width="wide">
        <PageHeader
          title="Lançamentos"
          description="Receitas e despesas da torcida (dados de exemplo)."
          breadcrumb={[
            { label: "Espécime", href: "/design-system#pagina-de-exemplo" },
            { label: "Financeiro", href: "/design-system#pagina-de-exemplo" },
            { label: "Lançamentos" },
          ]}
          actions={
            <Button onClick={() => toast.show("Exemplo: o formulário fica na seção Sobreposições do espécime.")}>
              <PlusIcon aria-hidden="true" />
              Novo lançamento
            </Button>
          }
        />
        <Suspense>
          <ExampleListing headingLevel={2} />
        </Suspense>
      </PageContainer>
    </ExampleShell>
  );
}
