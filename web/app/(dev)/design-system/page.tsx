import type { Metadata } from "next";

import { PageContainer } from "@/components/app/page-container";
import { PageHeader } from "@/components/app/page-header";
import { buttonVariants } from "@/components/ui/button";

import { BrandSection } from "./brand-section";
import { ColorsSection } from "./colors-section";
import { notFoundInProduction } from "./dev-only";
import { SECTIONS } from "./sections";
import { SpaceSection } from "./space-section";
import { paletteOf, readTokenSheet } from "./tokens";
import { TypographySection } from "./typography-section";

// Espécime do design system (DSF-06): a fundação e os componentes reais numa
// só página, fora do grupo autenticado, sem sessão e sem chamada à API. Em
// produção responde como não encontrada antes de renderizar qualquer coisa.

// Também nos metadados: em produção nem o título da página chega à resposta.
export function generateMetadata(): Metadata {
  notFoundInProduction();
  return { title: "Espécime do design system", robots: { index: false, follow: false } };
}

export default function DesignSystemPage() {
  notFoundInProduction();

  const sheet = readTokenSheet();
  const palette = paletteOf(sheet);

  return (
    <PageContainer width="wide">
      <PageHeader
        title="Espécime do design system"
        description="Referência viva da fundação visual e dos componentes da TJ Platform. Só em desenvolvimento: em produção esta rota responde como página não encontrada."
      />

      <nav id="indice" aria-label="Seções do espécime">
        <ul className="flex flex-wrap gap-x-2 gap-y-1">
          {SECTIONS.map((section) => (
            <li key={section.id}>
              <a href={`#${section.id}`} className={buttonVariants({ variant: "link", size: "sm" })}>
                {section.title}
              </a>
            </li>
          ))}
        </ul>
      </nav>

      <BrandSection />
      <ColorsSection palette={palette} />
      <TypographySection />
      <SpaceSection declarations={sheet.declarations} />
    </PageContainer>
  );
}
