import { BRAND_NAME, BrandMark, type BrandMarkSize } from "@/components/app/brand";

import { SpecimenBlock, SpecimenCaption, SpecimenSection } from "./section";

// Marca (DSF-05): o escudo oficial nos quatro tamanhos, sobre o conteúdo claro
// e sobre o shell escuro, e o nome como texto.

const SIZES: readonly BrandMarkSize[] = [32, 40, 64, 96];

function Sizes({ surface }: { surface: "claro" | "shell" }) {
  return (
    <ul
      className={
        surface === "shell"
          ? "flex flex-wrap items-end gap-6 rounded-lg bg-sidebar p-4 text-sidebar-foreground md:p-6"
          : "flex flex-wrap items-end gap-6 rounded-lg border bg-background p-4 md:p-6"
      }
    >
      {SIZES.map((size) => (
        <li key={size} className="flex flex-col items-center gap-2">
          <BrandMark size={size} />
          <span className={surface === "shell" ? "text-xs text-sidebar-muted-foreground" : "text-xs text-muted-foreground"}>
            {size} px
          </span>
        </li>
      ))}
    </ul>
  );
}

export function BrandSection() {
  return (
    <SpecimenSection
      id="marca"
      title="Marca"
      description="Escudo oficial versionado em public/brand, exibido como está: proporção 1:1, mínimo de 32 px, respiro de um oitavo do diâmetro, sem recolorir. O nome é texto, nunca imagem."
    >
      <SpecimenBlock title="Escudo sobre o conteúdo claro">
        <Sizes surface="claro" />
      </SpecimenBlock>
      <SpecimenBlock title="Escudo sobre o shell escuro">
        <Sizes surface="shell" />
      </SpecimenBlock>
      <SpecimenBlock title={`Nome: ${BRAND_NAME}`}>
        <div className="flex flex-col gap-4 md:flex-row">
          <div className="flex flex-1 items-center justify-center rounded-lg bg-sidebar p-4 text-sidebar-foreground md:p-6">
            <BrandMark size={40} withName />
          </div>
          <div className="flex flex-1 items-center justify-center rounded-lg border bg-background p-4 md:p-6">
            <BrandMark size={64} withName direction="column" />
          </div>
        </div>
        <SpecimenCaption>BrandMark com withName: ao lado do escudo (shell) e abaixo dele (acesso).</SpecimenCaption>
      </SpecimenBlock>
    </SpecimenSection>
  );
}
