import { describe, expect, it } from "vitest";

describe("infraestrutura de testes do front", () => {
  it("executa em ambiente jsdom", () => {
    expect(typeof window).toBe("object");
    expect(document.createElement("div").tagName).toBe("DIV");
  });
});
