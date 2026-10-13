import { Button as ButtonPrimitive } from "@base-ui/react/button"
import { cn } from "cn"
import type { ReactNode } from "react"

import { buttonVariants } from "@/components/ui/button"

const SQUARE = {
  sm: "size-control-sm",
  md: "size-control-md",
  lg: "size-control-lg",
} as const

type IconButtonProps = Omit<ButtonPrimitive.Props, "children" | "aria-label"> & {
  /** Nome acessível, obrigatório: numa linha de lista, identifica o item. */
  label: string
  /** O ícone. */
  children: ReactNode
  variant?: "ghost" | "secondary"
  size?: keyof typeof SQUARE
}

function IconButton({
  label,
  className,
  variant = "ghost",
  size = "md",
  ...props
}: IconButtonProps) {
  return (
    <ButtonPrimitive
      data-slot="icon-button"
      aria-label={label}
      className={cn(buttonVariants({ variant, size: null }), "whitespace-nowrap", SQUARE[size], className)}
      {...props}
    />
  )
}

export { IconButton }
export type { IconButtonProps }
