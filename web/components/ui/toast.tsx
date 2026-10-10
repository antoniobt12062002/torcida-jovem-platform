"use client"

import { Toast as ToastPrimitive } from "@base-ui/react/toast"
import { XIcon } from "lucide-react"

import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"

type ToastVariant = "default" | "success" | "destructive"

type ToastOptions = {
  description?: string
  timeout?: number
}

const toastManager = ToastPrimitive.createToastManager()

function show(variant: ToastVariant, title: string, options?: ToastOptions) {
  return toastManager.add({
    title,
    description: options?.description,
    timeout: options?.timeout,
    type: variant,
  })
}

// API imperativa: pode ser chamada de qualquer lugar (handlers, callbacks do cliente da API).
const toast = {
  show: (title: string, options?: ToastOptions) =>
    show("default", title, options),
  success: (title: string, options?: ToastOptions) =>
    show("success", title, options),
  error: (title: string, options?: ToastOptions) =>
    show("destructive", title, options),
  dismiss: (id?: string) => toastManager.close(id),
}

function ToastList() {
  const { toasts } = ToastPrimitive.useToastManager()
  return toasts.map((item) => (
    <ToastPrimitive.Root
      key={item.id}
      toast={item}
      data-slot="toast"
      className={cn(
        "absolute right-0 bottom-0 left-auto z-[calc(1000-var(--toast-index))] w-full origin-bottom rounded-lg border bg-popover text-popover-foreground shadow-lg ring-1 ring-foreground/10 select-none",
        "[--gap:0.75rem] [--peek:0.75rem] [--scale:calc(max(0,1-(var(--toast-index)*0.1)))] [--shrink:calc(1-var(--scale))] [--height:var(--toast-frontmost-height,var(--toast-height))]",
        "h-(--height) data-expanded:h-(--toast-height)",
        "transform-[translateX(var(--toast-swipe-movement-x))_translateY(calc(var(--toast-swipe-movement-y)-(var(--toast-index)*var(--peek))-(var(--shrink)*var(--height))))_scale(var(--scale))]",
        "data-expanded:transform-[translateX(var(--toast-swipe-movement-x))_translateY(calc(var(--toast-offset-y)*-1+var(--toast-index)*var(--gap)*-1+var(--toast-swipe-movement-y)))]",
        "transition-[transform,opacity,height] duration-300 data-ending-style:opacity-0 data-limited:opacity-0 data-starting-style:transform-[translateY(150%)]",
        item.type === "destructive" && "border-destructive/40",
        item.type === "success" && "border-primary/40"
      )}
    >
      <ToastPrimitive.Content className="flex items-start gap-3 overflow-hidden p-3 text-sm transition-opacity data-behind:opacity-0 data-expanded:opacity-100">
        <div className="flex min-w-0 flex-1 flex-col gap-1">
          <ToastPrimitive.Title
            className={cn(
              "font-medium",
              item.type === "destructive" && "text-destructive"
            )}
          />
          <ToastPrimitive.Description className="text-muted-foreground" />
        </div>
        <ToastPrimitive.Close
          aria-label="Fechar aviso"
          render={<Button variant="ghost" size="icon-xs" />}
        >
          <XIcon />
        </ToastPrimitive.Close>
      </ToastPrimitive.Content>
    </ToastPrimitive.Root>
  ))
}

function Toaster() {
  return (
    <ToastPrimitive.Provider toastManager={toastManager}>
      <ToastPrimitive.Portal>
        <ToastPrimitive.Viewport
          data-slot="toaster"
          className="fixed right-4 bottom-4 z-100 w-[calc(100vw-2rem)] sm:w-90"
        >
          <ToastList />
        </ToastPrimitive.Viewport>
      </ToastPrimitive.Portal>
    </ToastPrimitive.Provider>
  )
}

export { Toaster, toast }
