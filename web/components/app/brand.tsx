import Image from "next/image";

import { DisplayTitle } from "@/components/ui/typography";
import { cn } from "@/lib/utils";

// Escudo oficial da Torcida Jovem (DSF-05). O arquivo é exibido como está:
// proporção 1:1, sem recolorir, recortar, girar nem aplicar efeito. O respiro
// em volta é de um oitavo do diâmetro. O nome é texto, nunca imagem.

export const BRAND_NAME = "Torcida Jovem";

const BRAND_SRC = "/brand/torcida-jovem-logo.svg";

export type BrandMarkSize = 32 | 40 | 64 | 96;

// Respiro (um oitavo do diâmetro) e corpo do nome para cada tamanho.
const SIZES: Record<BrandMarkSize, { clearance: string; name: string }> = {
  32: { clearance: "p-1", name: "text-xl leading-6" },
  40: { clearance: "p-1.25", name: "text-2xl leading-7" },
  64: { clearance: "p-2", name: "text-3xl leading-9" },
  96: { clearance: "p-3", name: "" },
};

export type BrandMarkProps = {
  /** Diâmetro do escudo em px. Padrão: 32, o mínimo permitido. */
  size?: BrandMarkSize;
  /** Mostra o nome "Torcida Jovem" junto do escudo; a imagem passa a ser decorativa. */
  withName?: boolean;
  /** Posição do nome: ao lado (`row`, padrão) ou abaixo do escudo (`column`). */
  direction?: "row" | "column";
  /**
   * Escudo acima da dobra (tela de acesso, cabeçalho): carrega sem espera e com
   * prioridade alta. No Next 16 a prop `priority` do `Image` está obsoleta; a
   * documentação indica `loading="eager"` com `fetchPriority="high"`.
   */
  priority?: boolean;
  className?: string;
};

export function BrandMark({
  size = 32,
  withName = false,
  direction = "row",
  priority = false,
  className,
}: BrandMarkProps) {
  const { clearance, name } = SIZES[size];
  // `div`, não `span`: com o nome, a raiz contém um parágrafo (`DisplayTitle`).
  return (
    <div
      data-slot="brand-mark"
      className={cn("inline-flex items-center", direction === "column" && "flex-col", className)}
    >
      <span className={cn("inline-flex shrink-0", clearance)}>
        {/* SVG não passa pelo otimizador de imagens: é servido como está. */}
        <Image
          src={BRAND_SRC}
          alt={withName ? "" : BRAND_NAME}
          width={size}
          height={size}
          unoptimized
          loading={priority ? "eager" : undefined}
          fetchPriority={priority ? "high" : undefined}
          className="aspect-square shrink-0"
        />
      </span>
      {withName ? <DisplayTitle className={name}>{BRAND_NAME}</DisplayTitle> : null}
    </div>
  );
}
