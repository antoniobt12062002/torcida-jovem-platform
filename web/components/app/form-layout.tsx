"use client";

import type { ComponentProps, ReactNode } from "react";

import { cn } from "@/lib/utils";

import { FormErrorAlert } from "./form-error-alert";

// Layout de formulário (DSP-04): coluna única na largura estreita, com o erro
// que não pertence a nenhum campo num alerta no topo e as ações no rodapé.
// Quem usa passa as ações na ordem de leitura: secundária, depois a principal.

export type FormLayoutProps = ComponentProps<"form"> & {
  /** Erro sem campo (por exemplo o da API). Aparece no topo e recebe o foco ao surgir. */
  error?: ReactNode;
  /** Ações: em largura total abaixo de 768 px; à direita a partir daí. */
  footer: ReactNode;
};

export function FormLayout({ error, footer, className, children, ...props }: FormLayoutProps) {
  return (
    <form
      data-slot="form-layout"
      className={cn("flex w-full max-w-content-narrow flex-col gap-4", className)}
      {...props}
    >
      {/* O alerta recebe o foco quando surge, não a cada renderização. */}
      {error ? <FormErrorAlert>{error}</FormErrorAlert> : null}
      {children}
      <div
        data-slot="form-footer"
        className="mt-2 flex flex-col gap-2 *:w-full md:flex-row md:justify-end md:*:w-auto"
      >
        {footer}
      </div>
    </form>
  );
}
