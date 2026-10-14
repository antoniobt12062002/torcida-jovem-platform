"use client";

import Link from "next/link";
import { Suspense } from "react";

import { buttonVariants } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";

import { ExampleListing } from "./example-listing";
import { ExampleShell } from "./example-shell";
import { SpecimenBlock, SpecimenCaption, SpecimenSection } from "./section";

// Shell, acesso e página de exemplo. O `ShellFrame` não cria `main` nem `h1`:
// aparece aqui dentro de uma moldura que contém o cabeçalho fixo e o link de
// pular (transform cria o bloco de contenção dos elementos fixos). O
// `AccessLayout`, o `PageHeader` e o "Sem acesso" trazem o próprio `main` ou
// `h1`; para esta página manter um só de cada, ficam em páginas próprias do
// espécime, abertas pelos links abaixo.

export function ShellSection() {
  return (
    <SpecimenSection
      id="shell"
      title="Shell"
      description="ShellFrame com navegação de exemplo, item ativo (Lançamentos) e menu da pessoa, sem sessão. A navegação fica fixa a partir de 1024 px; abaixo, abre na gaveta pelo botão de menu."
    >
      <div data-slot="specimen-frame" className="relative transform-gpu overflow-hidden rounded-lg border">
        <ExampleShell>
          <div className="flex flex-col gap-4 p-gutter">
            <Card>
              <CardHeader>
                <p className="text-base leading-6 font-semibold">Área de conteúdo</p>
              </CardHeader>
              <CardContent>
                As páginas trazem o próprio main pelo PageContainer. Veja a página de exemplo completa, com o shell, o cabeçalho e a listagem.
              </CardContent>
            </Card>
          </div>
        </ExampleShell>
      </div>
      <SpecimenCaption>Moldura com altura mínima de uma tela, como o shell real.</SpecimenCaption>
    </SpecimenSection>
  );
}

export function AccessSection() {
  return (
    <SpecimenSection
      id="acesso"
      title="Acesso"
      description="AccessLayout com um formulário de login de exemplo, sem envio real. A partir de 1024 px, painel de marca à esquerda; abaixo, faixa de marca no topo."
    >
      <SpecimenBlock title="Tela de acesso">
        <p className="text-sm text-muted-foreground">
          O AccessLayout é a página inteira: traz o próprio main e o título h1 da tela. Por isso abre numa página própria do espécime.
        </p>
        <div>
          <Link href="/design-system/acesso" className={buttonVariants({ variant: "secondary" })}>
            Abrir a tela de acesso de exemplo
          </Link>
        </div>
      </SpecimenBlock>
    </SpecimenSection>
  );
}

export function ExamplePageSection() {
  return (
    <SpecimenSection
      id="pagina-de-exemplo"
      title="Página de exemplo"
      description="Barra de filtros e listagem juntas; a busca e os filtros ficam na URL e filtram os dados de exemplo no navegador."
    >
      <SpecimenBlock title="Página completa">
        <p className="text-sm text-muted-foreground">
          Com o shell, a trilha e o cabeçalho de página (que traz o h1), a página de exemplo abre à parte.
        </p>
        <div>
          <Link href="/design-system/pagina-de-exemplo" className={buttonVariants({ variant: "secondary" })}>
            Abrir a página de exemplo completa
          </Link>
        </div>
      </SpecimenBlock>
      <SpecimenBlock title="Filtros e listagem">
        <Suspense>
          <ExampleListing />
        </Suspense>
      </SpecimenBlock>
    </SpecimenSection>
  );
}
