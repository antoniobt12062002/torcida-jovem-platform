import type { ReactNode } from "react";

import { cn } from "@/lib/utils";

// Contêiner de página (DSP-01 AC1): o `main` da página, numa das três larguras
// nomeadas, com a margem lateral do token. O `id` é o alvo do link "Pular para
// o conteúdo" do shell, que leva o foco para cá.

export type PageWidth = "narrow" | "default" | "wide";

const WIDTHS: Record<PageWidth, string> = {
  narrow: "max-w-content-narrow",
  default: "max-w-content-default",
  wide: "max-w-content-wide",
};

export type PageContainerProps = { width?: PageWidth; children: ReactNode };

export function PageContainer({ width = "default", children }: PageContainerProps) {
  return (
    <main
      id="conteudo"
      tabIndex={-1}
      className={cn("mx-auto flex w-full flex-1 flex-col gap-6 px-gutter py-6 outline-none md:gap-8 md:py-8", WIDTHS[width])}
    >
      {children}
    </main>
  );
}
