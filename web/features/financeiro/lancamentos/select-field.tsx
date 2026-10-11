"use client";

import { type ReactNode, type Ref, useId } from "react";

import { cn } from "@/lib/utils";

// Lista de opções nativa (<select>) com rótulo, erro e aria ligados. Usada nos
// formulários e filtros de lançamentos.

export type SelectOption = { value: string; label: string };

export type SelectFieldProps = {
  label: ReactNode;
  value: string;
  onChange: (value: string) => void;
  onBlur?: () => void;
  name?: string;
  options: readonly SelectOption[];
  /** Opção vazia no topo (por exemplo "Selecione" ou "Todos"). */
  placeholder?: string;
  error?: string;
  disabled?: boolean;
  ref?: Ref<HTMLSelectElement>;
  className?: string;
};

export function SelectField({
  label,
  value,
  onChange,
  onBlur,
  name,
  options,
  placeholder,
  error,
  disabled,
  ref,
  className,
}: SelectFieldProps) {
  const id = useId();
  const errorId = `${id}-erro`;
  return (
    <div className={cn("flex flex-col gap-2", className)}>
      <label htmlFor={id} className={cn("text-sm font-medium", error && "text-destructive")}>
        {label}
      </label>
      <select
        ref={ref}
        id={id}
        name={name}
        value={value}
        disabled={disabled}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? errorId : undefined}
        onChange={(event) => onChange(event.target.value)}
        onBlur={onBlur}
        className="h-8 w-full min-w-0 rounded-lg border border-input bg-transparent px-2 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50 aria-invalid:border-destructive aria-invalid:ring-3 aria-invalid:ring-destructive/20 dark:bg-input/30"
      >
        {placeholder !== undefined ? <option value="">{placeholder}</option> : null}
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
      {error ? (
        <p id={errorId} role="alert" className="text-sm text-destructive">
          {error}
        </p>
      ) : null}
    </div>
  );
}
