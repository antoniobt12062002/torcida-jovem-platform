"use client";

import { useEffect, useRef } from "react";

import { PageTitle } from "@/components/ui/typography";

// Título das páginas de erro (DSS-04 AC5): o único `h1`, que recebe o foco ao
// aparecer para o leitor de tela anunciar a falha. Com `documentTitle`, também
// troca o título da aba, já que as páginas de erro do Next não exportam
// metadados.

export type ErrorHeadingProps = {
  children: string;
  documentTitle?: string;
};

export function ErrorHeading({ children, documentTitle }: ErrorHeadingProps) {
  const ref = useRef<HTMLHeadingElement>(null);

  useEffect(() => {
    if (documentTitle) document.title = documentTitle;
    ref.current?.focus();
  }, [documentTitle]);

  return (
    <PageTitle ref={ref} tabIndex={-1} className="outline-none">
      {children}
    </PageTitle>
  );
}
