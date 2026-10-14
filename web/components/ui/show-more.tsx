import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

type ShowMoreProps = {
  /** Quantos itens estão visíveis. */
  shown: number
  /** Total de itens; ausente quando a API pagina por cursor. */
  total?: number
  onShowMore: () => void
  loading?: boolean
  noun: readonly [singular: string, plural: string]
  className?: string
}

const NUMBER = new Intl.NumberFormat("pt-BR")

// "Mostrando 50 de 137 lançamentos" e o botão "Mostrar mais". A contagem fica
// numa região viva, para a mudança ser anunciada.
function ShowMore({
  shown,
  total,
  onShowMore,
  loading = false,
  noun,
  className,
}: ShowMoreProps) {
  const [singular, plural] = noun
  const count = total ?? shown
  const text =
    total === undefined
      ? `Mostrando ${NUMBER.format(shown)} ${count === 1 ? singular : plural}`
      : `Mostrando ${NUMBER.format(shown)} de ${NUMBER.format(total)} ${count === 1 ? singular : plural}`
  const hasMore = total === undefined || shown < total
  return (
    <div
      data-slot="show-more"
      className={cn("flex flex-col items-center gap-3", className)}
    >
      <p aria-live="polite" className="text-sm text-muted-foreground tabular-nums">
        {text}
      </p>
      {hasMore ? (
        <Button
          variant="secondary"
          loading={loading}
          loadingLabel="Carregando…"
          onClick={onShowMore}
        >
          Mostrar mais
        </Button>
      ) : null}
    </div>
  )
}

export { ShowMore }
export type { ShowMoreProps }
