import type { Identifier } from "@/components/app/format";

// Identificador de um lançamento sem número legível (DSP-07 AC7): o tipo
// seguido do código curto em fonte monoespaçada.
export function IdentifierText({ identifier }: { identifier: Identifier }) {
  return (
    <>
      {identifier.type} <span className="font-mono">{identifier.code}</span>
    </>
  );
}
