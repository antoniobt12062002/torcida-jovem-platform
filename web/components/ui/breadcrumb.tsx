import Link from "next/link"
import { ArrowLeftIcon, ChevronRightIcon } from "lucide-react"

import { cn } from "@/lib/utils"

type BreadcrumbItem = { label: string; href?: string }

const LINK =
  "rounded-sm underline-offset-4 outline-none hover:text-foreground hover:underline focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-ring-offset"

// Caminho até a página atual. O último item é a página atual, sem link. Abaixo
// de 768 px mostra só o item anterior, como "voltar".
function Breadcrumb({
  items,
  className,
}: {
  items: readonly BreadcrumbItem[]
  className?: string
}) {
  const last = items.length - 1
  const previous = items[last - 1]
  return (
    <nav
      data-slot="breadcrumb"
      aria-label="Trilha de navegação"
      className={cn("text-sm text-muted-foreground", className)}
    >
      {previous?.href ? (
        <Link
          data-slot="breadcrumb-back"
          href={previous.href}
          className={cn(
            LINK,
            "inline-flex min-h-control-sm touch-manipulation items-center gap-2 md:hidden"
          )}
        >
          <ArrowLeftIcon aria-hidden="true" className="size-4 shrink-0" />
          {previous.label}
        </Link>
      ) : null}
      <ol className="hidden flex-wrap items-center gap-2 md:flex">
        {items.map((item, index) => (
          <li key={index} className="inline-flex items-center gap-2">
            {index > 0 ? (
              <ChevronRightIcon aria-hidden="true" className="size-4 shrink-0" />
            ) : null}
            {index === last || !item.href ? (
              <span
                aria-current={index === last ? "page" : undefined}
                className={cn(index === last && "font-medium text-foreground")}
              >
                {item.label}
              </span>
            ) : (
              <Link href={item.href} className={LINK}>
                {item.label}
              </Link>
            )}
          </li>
        ))}
      </ol>
    </nav>
  )
}

export { Breadcrumb }
export type { BreadcrumbItem }
