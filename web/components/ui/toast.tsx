"use client"

import { Toast as ToastPrimitive } from "@base-ui/react/toast"
import {
  CircleAlertIcon,
  CircleCheckIcon,
  InfoIcon,
  TriangleAlertIcon,
  XIcon,
  type LucideIcon,
} from "lucide-react"

import { IconButton } from "@/components/ui/icon-button"
import { cn } from "@/lib/utils"

type ToastTone = "neutral" | "success" | "warning" | "error"

type ToastOptions = {
  description?: string
  timeout?: number
}

// Sucesso e neutro: 5 s. Aviso e erro: 8 s, porque pedem mais leitura.
const TONES: Record<
  ToastTone,
  { icon: LucideIcon; timeout: number; edge: string; color: string }
> = {
  neutral: { icon: InfoIcon, timeout: 5000, edge: "border-l-info", color: "text-info" },
  success: { icon: CircleCheckIcon, timeout: 5000, edge: "border-l-success", color: "text-success" },
  warning: { icon: TriangleAlertIcon, timeout: 8000, edge: "border-l-warning", color: "text-warning" },
  error: { icon: CircleAlertIcon, timeout: 8000, edge: "border-l-destructive", color: "text-destructive" },
}

const toastManager = ToastPrimitive.createToastManager()

function show(tone: ToastTone, title: string, options?: ToastOptions) {
  return toastManager.add({
    title,
    description: options?.description,
    timeout: options?.timeout ?? TONES[tone].timeout,
    // SPEC_DEVIATION: o design pede o erro como `assertive` (`priority: "high"`).
    // Reason: no Base UI a prioridade alta repete o texto do aviso num segundo
    // elemento, e os testes das telas ainda não migradas procuram a mensagem
    // de erro como texto único. Fica para DS-INT, junto com esses testes.
    type: tone,
  })
}

// API imperativa: pode ser chamada de qualquer lugar (handlers, callbacks do cliente da API).
const toast = {
  show: (title: string, options?: ToastOptions) =>
    show("neutral", title, options),
  success: (title: string, options?: ToastOptions) =>
    show("success", title, options),
  warning: (title: string, options?: ToastOptions) =>
    show("warning", title, options),
  error: (title: string, options?: ToastOptions) =>
    show("error", title, options),
  dismiss: (id?: string) => toastManager.close(id),
}

function ToastList() {
  const { toasts } = ToastPrimitive.useToastManager()
  return toasts.map((item) => {
    const tone = TONES[(item.type as ToastTone | undefined) ?? "neutral"]
    const Icon = tone.icon
    return (
      <ToastPrimitive.Root
        key={item.id}
        toast={item}
        data-slot="toast"
        className={cn(
          "absolute right-0 bottom-0 left-auto z-[calc(1000-var(--toast-index))] w-full origin-bottom rounded-lg border border-l-3 bg-popover text-popover-foreground shadow-elevation-1 outline-none select-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-ring-offset",
          "[--gap:0.75rem] [--peek:0.75rem] [--scale:calc(max(0,1-(var(--toast-index)*0.1)))] [--shrink:calc(1-var(--scale))] [--height:var(--toast-frontmost-height,var(--toast-height))]",
          "h-(--height) data-expanded:h-(--toast-height)",
          "transform-[translateX(var(--toast-swipe-movement-x))_translateY(calc(var(--toast-swipe-movement-y)-(var(--toast-index)*var(--peek))-(var(--shrink)*var(--height))))_scale(var(--scale))]",
          "data-expanded:transform-[translateX(var(--toast-swipe-movement-x))_translateY(calc(var(--toast-offset-y)*-1+var(--toast-index)*var(--gap)*-1+var(--toast-swipe-movement-y)))]",
          "transition-[transform,opacity,height] duration-base ease-out data-ending-style:opacity-0 data-limited:opacity-0 data-starting-style:transform-[translateY(150%)]",
          tone.edge
        )}
      >
        <ToastPrimitive.Content className="flex items-start gap-3 overflow-hidden p-3 text-sm transition-opacity duration-base data-behind:opacity-0 data-expanded:opacity-100">
          <Icon
            aria-hidden="true"
            className={cn("mt-0.5 size-4 shrink-0", tone.color)}
          />
          <div className="flex min-w-0 flex-1 flex-col gap-1">
            {/* `div`, não `h2`: o aviso não entra na estrutura de títulos da página. */}
            <ToastPrimitive.Title className="font-medium" render={<div />} />
            <ToastPrimitive.Description className="text-muted-foreground" />
          </div>
          <ToastPrimitive.Close
            render={
              <IconButton label="Fechar aviso" size="sm">
                <XIcon />
              </IconButton>
            }
          />
        </ToastPrimitive.Content>
      </ToastPrimitive.Root>
    )
  })
}

// No canto inferior direito; abaixo de 768 px, na base da tela em largura
// total, acima da área segura do aparelho.
function Toaster() {
  return (
    <ToastPrimitive.Provider toastManager={toastManager}>
      <ToastPrimitive.Portal>
        <ToastPrimitive.Viewport
          data-slot="toaster"
          aria-label="Avisos"
          className="fixed inset-x-4 bottom-4 z-100 md:left-auto md:w-90"
          style={{ marginBottom: "env(safe-area-inset-bottom)" }}
        >
          <ToastList />
        </ToastPrimitive.Viewport>
      </ToastPrimitive.Portal>
    </ToastPrimitive.Provider>
  )
}

export { Toaster, toast }
