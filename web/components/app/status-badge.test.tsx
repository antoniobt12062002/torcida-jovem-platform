import { render, screen } from "@testing-library/react";
import { readFileSync } from "node:fs";
import { join } from "node:path";
import { BanIcon, CircleDashedIcon, TriangleAlertIcon } from "lucide-react";
import { describe, expect, it } from "vitest";

import { WEB_ROOT } from "@/test/design/guard";

import { activeStatusDef, StatusBadge, type StatusDef, type Tone } from "./status-badge";

// DS-INT (T11, DSM-05): a forma em texto saiu; cada feature passa o próprio
// `StatusDef`. O mapa de DD-3 vive em `features/financeiro/lancamentos/labels.ts`
// e é coberto por `lancamentos-page.test.tsx`.
describe("StatusBadge", () => {
  it.each<[boolean, string]>([
    [true, "Ativo"],
    [false, "Inativo"],
  ])("mostra ativo=%s como %s", (active, label) => {
    render(<StatusBadge status={activeStatusDef(active)} />);
    expect(screen.getByText(label)).toBeTruthy();
  });

  it("não aceita mais o status como texto (alias removido em DS-INT)", () => {
    // @ts-expect-error a forma antiga, com o status em texto, foi removida.
    const legacy = <StatusBadge status="CANCELADA" />;
    expect(legacy).toBeTruthy();
  });
});

function badge(container: HTMLElement) {
  return container.querySelector("[data-slot=badge]") as HTMLElement;
}

function drawing(container: HTMLElement) {
  return badge(container).querySelector("svg")?.innerHTML ?? "";
}

// DSC-05 AC2 e AC4: tom, ícone e rótulo vêm da feature dona do status.
describe("StatusBadge com definição", () => {
  it("mostra o rótulo e o ícone recebidos, com o ícone escondido de leitores de tela", () => {
    const status: StatusDef = { label: "Em conferência", tone: "warning", icon: TriangleAlertIcon };
    const { container } = render(<StatusBadge status={status} />);
    expect(badge(container).textContent).toBe("Em conferência");
    expect(badge(container).querySelector("svg")?.getAttribute("aria-hidden")).toBe("true");
  });

  it.each<[Tone, string]>([
    ["neutral", "bg-info-subtle"],
    ["success", "bg-success-subtle"],
    ["warning", "bg-warning-subtle"],
    ["error", "bg-destructive-subtle"],
  ])("tom %s usa o fundo %s", (tone, background) => {
    const { container } = render(<StatusBadge status={{ label: "Status", tone, icon: CircleDashedIcon }} />);
    expect(badge(container).className.split(/\s+/)).toContain(background);
    expect(badge(container).getAttribute("data-tone")).toBe(tone);
  });

  // DSC-05 AC3: mesmo tom, ainda distinguíveis por rótulo e ícone.
  it("dois status no mesmo tom continuam distintos por rótulo e ícone", () => {
    const a = render(<StatusBadge status={{ label: "Criada", tone: "neutral", icon: CircleDashedIcon }} />);
    const b = render(<StatusBadge status={{ label: "Cancelada", tone: "neutral", icon: BanIcon }} />);
    expect(badge(a.container).className).toBe(badge(b.container).className);
    expect(badge(a.container).textContent).not.toBe(badge(b.container).textContent);
    expect(drawing(a.container)).not.toBe(drawing(b.container));
  });

  it("não importa tipos de módulo nem de features (DSC-05 AC4)", () => {
    const source = readFileSync(join(WEB_ROOT, "components/app/status-badge.tsx"), "utf8");
    const imports = [...source.matchAll(/from\s+"([^"]+)"/g)].map((match) => match[1]);
    expect(imports.length).toBeGreaterThan(0);
    expect(imports.filter((path) => path.includes("lib/api") || path.includes("features"))).toEqual([]);
  });
});

describe("activeStatusDef", () => {
  it("devolve a definição pronta de ativo e de inativo", () => {
    expect(activeStatusDef(true)).toMatchObject({ label: "Ativo", tone: "success" });
    expect(activeStatusDef(false)).toMatchObject({ label: "Inativo", tone: "neutral" });
    expect(activeStatusDef(true).icon).not.toBe(activeStatusDef(false).icon);
  });
});
