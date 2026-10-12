import * as React from "react"
import { cva } from "class-variance-authority"
import {
  CircleAlertIcon,
  CircleCheckIcon,
  InfoIcon,
  TriangleAlertIcon,
  type LucideIcon,
} from "lucide-react"
import { cn } from "@/lib/utils"

type AlertTone = "neutral" | "success" | "warning" | "error"

const alertVariants = cva(
  "group/alert relative flex w-full items-start gap-2 rounded-lg border px-3 py-2 text-left text-sm has-data-[slot=alert-action]:pr-18 [&>svg]:mt-0.5 [&>svg]:size-4 [&>svg]:shrink-0",
  {
    variants: {
      tone: {
        neutral: "border-info bg-info-subtle text-info",
        success: "border-success bg-success-subtle text-success-emphasis",
        warning: "border-warning bg-warning-subtle text-warning-emphasis",
        error: "border-destructive bg-destructive-subtle text-destructive-emphasis",
      },
    },
  }
)

// Cada tom tem ícone fixo: o estado nunca depende só da cor. Erro e aviso
// interrompem a leitura (`alert`); os demais são anunciados sem interromper.
const TONES: Record<AlertTone, { icon: LucideIcon; role: "alert" | "status" }> = {
  neutral: { icon: InfoIcon, role: "status" },
  success: { icon: CircleCheckIcon, role: "status" },
  warning: { icon: TriangleAlertIcon, role: "alert" },
  error: { icon: CircleAlertIcon, role: "alert" },
}

type AlertVariant =
  | AlertTone
  // Alias de compatibilidade: removido em DS-INT.
  | "default"
  // Alias de compatibilidade: removido em DS-INT.
  | "destructive"

function Alert({
  className,
  variant = "default",
  children,
  ...props
}: React.ComponentProps<"div"> & { variant?: AlertVariant | null }) {
  // Alias de compatibilidade: removido em DS-INT. `default` (ou nenhuma
  // variante) é o tom neutro com o papel `alert` que as telas ainda esperam;
  // `destructive` é o tom de erro.
  const legacy = variant === "default" || variant === null
  const tone: AlertTone = legacy
    ? "neutral"
    : variant === "destructive"
      ? "error"
      : variant
  const Icon = TONES[tone].icon
  return (
    <div
      data-slot="alert"
      role={legacy ? "alert" : TONES[tone].role}
      className={cn(alertVariants({ tone }), className)}
      {...props}
    >
      <Icon aria-hidden="true" />
      <div className="grid min-w-0 flex-1 gap-0.5">{children}</div>
    </div>
  )
}

function AlertTitle({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="alert-title"
      className={cn(
        "font-medium [&_a]:underline [&_a]:underline-offset-3",
        className
      )}
      {...props}
    />
  )
}

function AlertDescription({
  className,
  ...props
}: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="alert-description"
      className={cn(
        "text-sm text-pretty [&_a]:underline [&_a]:underline-offset-3 [&_p:not(:last-child)]:mb-4",
        className
      )}
      {...props}
    />
  )
}

function AlertAction({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="alert-action"
      className={cn("absolute top-2 right-2", className)}
      {...props}
    />
  )
}

export { Alert, AlertTitle, AlertDescription, AlertAction }
