import * as React from "react"
import { cn } from "@/lib/utils"

// Papéis tipográficos de identidade. Com o shell, este é o único lugar onde a
// fonte condensada (`font-display`) e a caixa alta aparecem; a guarda de design
// recusa as duas fora daqui. Todos usam quebra de linha equilibrada.
//
// O design pede espaçamento de 0,01em nos títulos e de 0,06em no sobretítulo.
// Sem token para esses valores, valem os passos mais próximos da escala padrão.

/** Destaque (36/40), em caixa alta. Não cria título na estrutura da página. */
function DisplayTitle({ className, ...props }: React.ComponentProps<"p">) {
  return (
    <p
      data-slot="display-title"
      className={cn(
        "font-display text-4xl font-bold tracking-normal text-balance uppercase",
        className
      )}
      {...props}
    />
  )
}

/** Título de página: o único `h1`, com o filete de marca de 32 por 2 px. */
function PageTitle({ className, ...props }: React.ComponentProps<"h1">) {
  return (
    <h1
      data-slot="page-title"
      className={cn(
        "font-display text-2xl leading-7 font-bold tracking-normal text-balance uppercase after:mt-2 after:block after:h-0.5 after:w-8 after:bg-brand md:text-3xl md:leading-9",
        className
      )}
      {...props}
    />
  )
}

/** Título de seção (`h2`), em frase comum, com a régua forte de 2 px. */
function SectionTitle({ className, ...props }: React.ComponentProps<"h2">) {
  return (
    <h2
      data-slot="section-title"
      className={cn(
        "border-b-2 border-border-strong pb-2 font-display text-xl leading-6 font-semibold text-balance",
        className
      )}
      {...props}
    />
  )
}

/** Rótulo de grupo e sobretítulo (14/16), em caixa alta. */
function Eyebrow({ className, ...props }: React.ComponentProps<"p">) {
  return (
    <p
      data-slot="eyebrow"
      className={cn(
        "font-display text-sm leading-4 font-semibold tracking-wider text-balance uppercase",
        className
      )}
      {...props}
    />
  )
}

export { DisplayTitle, PageTitle, SectionTitle, Eyebrow }
