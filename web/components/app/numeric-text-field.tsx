"use client";

import { type ReactNode, type Ref, useState } from "react";

import { Field, FieldDescription, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";

// Base comum dos campos de dinheiro e quantidade: guarda o texto digitado,
// entrega ao formulário só o inteiro lido (ou null) e mostra o erro de
// formato depois que a pessoa sai do campo. Uso interno de components/app.

export type ParseResult = { value: number; error: null } | { value: null; error: string | null };

/** Props comuns, compatíveis com o `field` do Controller do React Hook Form. */
export type NumericFieldProps = {
  label: ReactNode;
  /** Valor atual (inteiro) ou null quando vazio ou inválido. */
  value: number | null;
  onChange: (value: number | null) => void;
  onBlur?: () => void;
  name?: string;
  id?: string;
  /** Erro vindo do formulário (regras do React Hook Form ou `errors[]` da API). */
  error?: string;
  description?: ReactNode;
  disabled?: boolean;
  required?: boolean;
  placeholder?: string;
  ref?: Ref<HTMLInputElement>;
  className?: string;
};

type Internal = NumericFieldProps & {
  parse: (text: string) => ParseResult;
  format: (value: number) => string;
  inputMode: "decimal" | "numeric" | "text";
  prefix?: ReactNode;
};

export function NumericTextField({
  label,
  value,
  onChange,
  onBlur,
  name,
  id,
  error,
  description,
  disabled,
  required,
  placeholder,
  ref,
  className,
  parse,
  format,
  inputMode,
  prefix,
}: Internal) {
  const [text, setText] = useState(() => (value === null ? "" : format(value)));
  const [shownValue, setShownValue] = useState(value);
  const [touched, setTouched] = useState(false);

  // Valor trocado por fora (reset do formulário, valor inicial carregado): o
  // texto passa a mostrar o novo valor, a menos que já o represente.
  if (value !== shownValue) {
    setShownValue(value);
    if (parse(text).value !== value) setText(value === null ? "" : format(value));
  }

  const parsed = parse(text);
  const message = (touched && parsed.error) || error;

  return (
    <Field invalid={!!message} className={className}>
      <FieldLabel>{label}</FieldLabel>
      <div className="flex items-center gap-2">
        {prefix ? <span className="text-sm text-muted-foreground">{prefix}</span> : null}
        <Input
          ref={ref}
          id={id}
          name={name}
          type="text"
          inputMode={inputMode}
          autoComplete="off"
          disabled={disabled}
          required={required}
          placeholder={placeholder}
          value={text}
          onChange={(event) => {
            const next = event.target.value;
            setText(next);
            const result = parse(next);
            setShownValue(result.value);
            onChange(result.value);
          }}
          onBlur={() => {
            setTouched(true);
            onBlur?.();
          }}
        />
      </div>
      {description ? <FieldDescription>{description}</FieldDescription> : null}
      {message ? <FieldError>{message}</FieldError> : null}
    </Field>
  );
}
