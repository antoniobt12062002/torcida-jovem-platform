import { Button as ButtonPrimitive } from "@base-ui/react/button"
import { cva, type VariantProps } from "class-variance-authority"
import { cn } from "cn"
import { LoaderCircleIcon } from "lucide-react"

const PRIMARY =
  "bg-primary text-primary-foreground hover:bg-primary-hover active:bg-primary-hover/90 disabled:bg-disabled disabled:text-disabled-foreground"
const SECONDARY =
  "border-input bg-background text-foreground hover:bg-accent active:bg-border aria-expanded:bg-accent disabled:bg-disabled disabled:text-disabled-foreground"

const buttonVariants = cva(
  "group/button inline-flex shrink-0 touch-manipulation items-center justify-center rounded-md border border-transparent text-sm font-medium whitespace-nowrap transition-colors duration-fast outline-none select-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-ring-offset disabled:pointer-events-none disabled:cursor-default [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
  {
    variants: {
      variant: {
        primary: PRIMARY,
        secondary: SECONDARY,
        ghost:
          "text-foreground hover:bg-accent active:bg-border aria-expanded:bg-accent disabled:text-disabled-foreground",
        destructive:
          "bg-destructive text-destructive-foreground hover:bg-destructive-emphasis active:bg-destructive-emphasis/90 disabled:bg-disabled disabled:text-disabled-foreground",
        link: "text-foreground underline underline-offset-4 hover:text-muted-foreground disabled:text-disabled-foreground",
        // Alias de compatibilidade: removido em DS-INT.
        default: PRIMARY,
        // Alias de compatibilidade: removido em DS-INT.
        outline: SECONDARY,
      },
      size: {
        sm: "h-control-sm gap-1.5 px-3",
        md: "h-control-md gap-2 px-4",
        lg: "h-control-lg gap-2 px-5",
        // Alias de compatibilidade: removido em DS-INT.
        default: "h-control-md gap-2 px-4",
        // Alias de compatibilidade: removido em DS-INT.
        icon: "size-control-md",
        // Alias de compatibilidade: removido em DS-INT.
        "icon-sm": "size-control-sm",
      },
    },
    defaultVariants: {
      variant: "primary",
      size: "md",
    },
  }
)

// O rótulo de progresso é obrigatório quando `loading` é usado e termina em "…".
type LoadingProps =
  | { loading?: undefined; loadingLabel?: undefined }
  | { loading: boolean; loadingLabel: `${string}…` }

type ButtonProps = ButtonPrimitive.Props &
  VariantProps<typeof buttonVariants> &
  LoadingProps

function Button({
  className,
  variant = "primary",
  size = "md",
  loading,
  loadingLabel,
  disabled,
  children,
  ...props
}: ButtonProps) {
  return (
    <ButtonPrimitive
      data-slot="button"
      className={cn(buttonVariants({ variant, size }), className)}
      disabled={disabled || loading}
      aria-busy={loading || undefined}
      {...props}
    >
      {loading ? (
        <>
          <LoaderCircleIcon aria-hidden="true" className="animate-spin" />
          {loadingLabel}
        </>
      ) : (
        children
      )}
    </ButtonPrimitive>
  )
}

export { Button, buttonVariants }
export type { ButtonProps }
