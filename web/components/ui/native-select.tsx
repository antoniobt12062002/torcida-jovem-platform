import * as React from "react"
import { Field as FieldPrimitive } from "@base-ui/react/field"
import { ChevronDownIcon } from "lucide-react"

import { controlClasses } from "@/components/ui/input"
import { cn } from "@/lib/utils"

type NativeSelectProps = React.ComponentProps<"select"> & { invalid?: boolean }

// Select nativo do kit. Dentro de `Field`, o rótulo e o erro são ligados pelo
// Base UI, como no Input; fora dele, `invalid` marca o erro.
function NativeSelect({ className, children, invalid, ...props }: NativeSelectProps) {
  return (
    <div data-slot="native-select" className="relative w-full">
      <FieldPrimitive.Control
        render={<select>{children}</select>}
        data-slot="native-select-control"
        aria-invalid={invalid || undefined}
        className={cn(
          controlClasses,
          "h-control-md appearance-none pr-9 [&_option]:bg-background [&_option]:text-foreground",
          className
        )}
        {...(props as FieldPrimitive.Control.Props)}
      />
      <ChevronDownIcon
        aria-hidden="true"
        className="pointer-events-none absolute top-1/2 right-3 size-4 -translate-y-1/2 text-muted-foreground"
      />
    </div>
  )
}

export { NativeSelect }
export type { NativeSelectProps }
