import type { ReactNode, Ref } from "react";

import { Breadcrumb, type BreadcrumbItem } from "@/components/ui/breadcrumb";
import { PageTitle } from "@/components/ui/typography";
import { cn } from "@/lib/utils";

// Cabeçalho de página (DSP-01 AC2 a AC6): o único `h1`, com o filete de marca,
// descrição, trilha e ações. A partir de 768 px, título à esquerda e ações à
// direita; abaixo disso, as ações vão para baixo do título, em largura total.
// Quem usa passa as ações na ordem de leitura, com a primária por último.

export type PageHeaderProps = {
  title: string;
  description?: ReactNode;
  /** Caminho até a página atual; o último item é a página atual, sem link. */
  breadcrumb?: readonly BreadcrumbItem[];
  /** No máximo um `Button` primário; as demais ações são secundárias ou vão em menu. */
  actions?: ReactNode;
  /** Informação do registro ao lado do título, por exemplo o `StatusBadge`. */
  meta?: ReactNode;
  /**
   * Referência ao título, focável só por script (`tabIndex={-1}`): destino do
   * foco quando uma ação some da tela depois de concluída.
   */
  titleRef?: Ref<HTMLHeadingElement>;
  className?: string;
};

export function PageHeader({ title, description, breadcrumb, actions, meta, titleRef, className }: PageHeaderProps) {
  return (
    <header data-slot="page-header" className={cn("flex flex-col gap-4", className)}>
      {breadcrumb?.length ? <Breadcrumb items={breadcrumb} /> : null}
      <div className="flex flex-col gap-4 md:flex-row md:items-start md:justify-between">
        <div className="flex min-w-0 flex-col gap-2">
          <div className="flex flex-wrap items-start gap-x-3 gap-y-2">
            <PageTitle
              ref={titleRef}
              tabIndex={-1}
              className="min-w-0 rounded-sm wrap-break-word outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-ring-offset"
            >
              {title}
            </PageTitle>
            {meta}
          </div>
          {description ? <div className="text-sm text-muted-foreground">{description}</div> : null}
        </div>
        {actions ? (
          <div
            data-slot="page-header-actions"
            className="flex flex-col gap-2 *:w-full md:shrink-0 md:flex-row md:items-center md:*:w-auto"
          >
            {actions}
          </div>
        ) : null}
      </div>
    </header>
  );
}
