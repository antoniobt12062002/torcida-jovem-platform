"use client"

import * as React from "react"
import { Dialog as DialogPrimitive } from "@base-ui/react/dialog"
import { cn } from "@/lib/utils"

import { Button } from "@/components/ui/button"
import { IconButton } from "@/components/ui/icon-button"
import { XIcon } from "lucide-react"

// Soma a área segura do aparelho ao espaçamento da base (o passo 4 da escala).
const SAFE_AREA_BOTTOM = {
  paddingBottom: "calc(var(--spacing) * 4 + env(safe-area-inset-bottom))",
} as const

function Dialog({ ...props }: DialogPrimitive.Root.Props) {
  return <DialogPrimitive.Root data-slot="dialog" {...props} />
}

function DialogTrigger({ ...props }: DialogPrimitive.Trigger.Props) {
  return <DialogPrimitive.Trigger data-slot="dialog-trigger" {...props} />
}

function DialogPortal({ ...props }: DialogPrimitive.Portal.Props) {
  return <DialogPrimitive.Portal data-slot="dialog-portal" {...props} />
}

function DialogClose({ ...props }: DialogPrimitive.Close.Props) {
  return <DialogPrimitive.Close data-slot="dialog-close" {...props} />
}

function DialogOverlay({
  className,
  ...props
}: DialogPrimitive.Backdrop.Props) {
  return (
    <DialogPrimitive.Backdrop
      data-slot="dialog-overlay"
      className={cn(
        "fixed inset-0 isolate z-50 bg-overlay duration-slow data-open:animate-in data-open:fade-in-0 data-closed:animate-out data-closed:fade-out-0",
        className
      )}
      {...props}
    />
  )
}

// Abaixo de 640 px ocupa a tela inteira; a partir daí é uma caixa centrada de
// 480 px (`default`) ou 640 px (`wide`). A rolagem fica contida no diálogo.
// Com rodapé, o diálogo não tem padding inferior: o elemento grudado com
// `bottom-0` para na borda do padding de quem rola, e o rodapé subiria.
function DialogContent({
  className,
  children,
  showCloseButton = true,
  size = "default",
  ...props
}: DialogPrimitive.Popup.Props & {
  showCloseButton?: boolean
  size?: "default" | "wide"
}) {
  return (
    <DialogPortal>
      <DialogOverlay />
      <DialogPrimitive.Popup
        data-slot="dialog-content"
        data-size={size}
        className={cn(
          "fixed inset-0 z-50 flex w-full flex-col gap-4 overflow-y-auto overscroll-contain bg-popover p-4 text-sm text-popover-foreground shadow-elevation-2 duration-slow outline-none focus-visible:ring-2 focus-visible:ring-ring data-open:animate-in data-open:fade-in-0 data-closed:animate-out data-closed:fade-out-0 sm:inset-auto sm:top-1/2 sm:left-1/2 sm:max-h-11/12 sm:-translate-x-1/2 sm:-translate-y-1/2 sm:rounded-lg sm:data-open:zoom-in-95 sm:data-closed:zoom-out-95 md:p-6 has-data-[slot=dialog-footer]:pb-0 md:has-data-[slot=dialog-footer]:pb-0",
          size === "wide" ? "sm:max-w-content-narrow" : "sm:max-w-120",
          className
        )}
        {...props}
      >
        {children}
        {showCloseButton && (
          <DialogPrimitive.Close
            data-slot="dialog-close"
            render={
              <IconButton
                label="Fechar"
                size="sm"
                className="absolute top-2 right-2 max-sm:size-control-lg"
              >
                <XIcon />
              </IconButton>
            }
          />
        )}
      </DialogPrimitive.Popup>
    </DialogPortal>
  )
}

function DialogHeader({ className, ...props }: React.ComponentProps<"div">) {
  return (
    <div
      data-slot="dialog-header"
      className={cn("flex flex-col gap-2 pr-10", className)}
      {...props}
    />
  )
}

// Preso à base do diálogo (não à janela), colado à borda inferior e acima da
// área segura do aparelho; o corpo rola por trás dele.
function DialogFooter({
  className,
  showCloseButton = false,
  children,
  ...props
}: React.ComponentProps<"div"> & {
  showCloseButton?: boolean
}) {
  return (
    <div
      data-slot="dialog-footer"
      className={cn(
        "sticky bottom-0 -mx-4 mt-auto flex flex-col-reverse gap-2 border-t bg-muted p-4 sm:flex-row sm:justify-end md:-mx-6 md:px-6",
        className
      )}
      {...props}
      style={{ ...SAFE_AREA_BOTTOM, ...props.style }}
    >
      {children}
      {showCloseButton && (
        <DialogPrimitive.Close render={<Button variant="secondary" />}>
          Fechar
        </DialogPrimitive.Close>
      )}
    </div>
  )
}

function DialogTitle({ className, ...props }: DialogPrimitive.Title.Props) {
  return (
    <DialogPrimitive.Title
      data-slot="dialog-title"
      className={cn(
        "font-heading text-base leading-6 font-semibold",
        className
      )}
      {...props}
    />
  )
}

function DialogDescription({
  className,
  ...props
}: DialogPrimitive.Description.Props) {
  return (
    <DialogPrimitive.Description
      data-slot="dialog-description"
      className={cn(
        "text-sm text-muted-foreground *:[a]:underline *:[a]:underline-offset-3",
        className
      )}
      {...props}
    />
  )
}

export {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogOverlay,
  DialogPortal,
  DialogTitle,
  DialogTrigger,
}
