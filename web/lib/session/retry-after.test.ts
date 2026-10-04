import { describe, expect, it } from "vitest";

import { formatRetryAfter } from "./retry-after";

// FND-03 AC3: o tempo de espera vem do cabeçalho Retry-After (segundos, ou uma
// data HTTP) e aparece em português.
describe("formatRetryAfter", () => {
  it.each([
    ["1", "1 segundo"],
    ["45", "45 segundos"],
    ["60", "1 minuto"],
    ["61", "2 minutos"],
    ["900", "15 minutos"],
    ["3600", "1 hora"],
    ["5400", "1 hora e 30 minutos"],
    ["7200", "2 horas"],
    ["7260", "2 horas e 1 minuto"],
  ])("Retry-After %s vira %s", (header, text) => {
    expect(formatRetryAfter(header)).toBe(text);
  });

  it("aceita a forma de data HTTP", () => {
    const now = Date.parse("2026-10-04T12:00:00Z");
    expect(formatRetryAfter("Sun, 04 Oct 2026 12:05:00 GMT", now)).toBe("5 minutos");
  });

  it.each([null, undefined, "", "abc", "-5", "0", "1.5"])(
    "sem tempo válido (%s) devolve null",
    (header) => {
      expect(formatRetryAfter(header)).toBeNull();
    },
  );

  it("data no passado devolve null", () => {
    const now = Date.parse("2026-10-04T12:00:00Z");
    expect(formatRetryAfter("Sun, 04 Oct 2026 11:00:00 GMT", now)).toBeNull();
  });
});
