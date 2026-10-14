import { Field as FieldPrimitive } from "@base-ui/react/field"

import { controlClasses } from "@/components/ui/input"
import { cn } from "@/lib/utils"

function Textarea({
  className,
  rows,
  ...props
}: FieldPrimitive.Control.Props & { rows?: number }) {
  return (
    <FieldPrimitive.Control
      render={<textarea rows={rows} />}
      data-slot="textarea"
      className={cn(
        controlClasses,
        "flex field-sizing-content min-h-20 resize-y py-2",
        className
      )}
      {...props}
    />
  )
}

export { Textarea }
