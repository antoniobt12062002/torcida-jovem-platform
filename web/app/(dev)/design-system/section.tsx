import type { ReactNode } from "react";

import { SectionTitle } from "@/components/ui/typography";

// Composição das seções do espécime: o título de seção real (`h2`) e uma
// âncora. Nenhum estilo além de utilitários de layout e de tokens.

export type SpecimenSectionProps = {
  id: string;
  title: string;
  description?: ReactNode;
  children: ReactNode;
};

export function SpecimenSection({ id, title, description, children }: SpecimenSectionProps) {
  return (
    <section id={id} aria-labelledby={`${id}-titulo`} className="flex scroll-mt-6 flex-col gap-6">
      <div className="flex flex-col gap-2">
        <SectionTitle id={`${id}-titulo`}>{title}</SectionTitle>
        {description ? <p className="text-sm text-pretty text-muted-foreground">{description}</p> : null}
      </div>
      {children}
    </section>
  );
}

/** Título de bloco (`h3`, Geist 600 16/24), dentro de uma seção. */
export function SpecimenBlock({ title, children }: { title: string; children: ReactNode }) {
  return (
    <div className="flex flex-col gap-3">
      <h3 className="text-base leading-6 font-semibold text-balance">{title}</h3>
      {children}
    </div>
  );
}

/** Legenda de uma amostra: o nome do token ou do componente. */
export function SpecimenCaption({ children }: { children: ReactNode }) {
  return <p className="text-xs text-muted-foreground">{children}</p>;
}
