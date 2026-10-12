import * as React from "react"
import { Input as InputPrimitive } from "@base-ui/react/input"
import { cn } from "@/lib/utils"

// Borda, texto e estados comuns a Input, Textarea e NativeSelect. A cor do
// texto vai com a variante `[&]:` porque `cn` trata `text-control` (tamanho)
// e uma cor de texto como a mesma propriedade e descartaria o tamanho.
const controlClasses =
  "w-full min-w-0 rounded-md border border-input bg-background px-3 text-control [&]:text-foreground transition-colors duration-fast outline-none placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-ring-offset aria-invalid:border-destructive aria-readonly:border-transparent aria-readonly:bg-muted [&[readonly]]:border-transparent [&[readonly]]:bg-muted disabled:cursor-default disabled:border-border disabled:bg-disabled disabled:text-disabled-foreground"

function Input({ className, type, ...props }: React.ComponentProps<"input">) {
  return (
    <InputPrimitive
      type={type}
      data-slot="input"
      className={cn(
        controlClasses,
        "h-control-md file:inline-flex file:h-6 file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground",
        className
      )}
      {...props}
    />
  )
}

export { Input, controlClasses }
