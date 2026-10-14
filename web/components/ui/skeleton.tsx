import { cn } from "@/lib/utils"

// Formas prontas: linha de texto, linha de tabela e bloco de valor. Sem
// `shape`, quem chama dá as medidas por `className`.
const SHAPES = {
  text: "h-4 w-full",
  row: "h-table-row w-full",
  block: "h-10 w-40",
} as const

function Skeleton({
  className,
  shape,
  ...props
}: React.ComponentProps<"div"> & { shape?: keyof typeof SHAPES }) {
  return (
    <div
      data-slot="skeleton"
      aria-hidden="true"
      className={cn(
        "animate-pulse rounded-md bg-muted motion-reduce:animate-none",
        shape && SHAPES[shape],
        className
      )}
      {...props}
    />
  )
}

export { Skeleton }
