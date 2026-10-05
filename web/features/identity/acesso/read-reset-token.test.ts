import { describe, expect, it } from "vitest";

import { readResetToken } from "./read-reset-token";

// ACS-03 AC1-2: o token vem do fragmento `#token=` do link do e-mail.
describe("readResetToken", () => {
  it("lê o token do fragmento", () => {
    expect(readResetToken("#token=abc_DEF-123")).toBe("abc_DEF-123");
  });

  it("aceita o fragmento sem o #", () => {
    expect(readResetToken("token=abc123")).toBe("abc123");
  });

  it("ignora outros parâmetros do fragmento", () => {
    expect(readResetToken("#x=1&token=abc123&y=2")).toBe("abc123");
  });

  it("preserva o token como veio, inclusive +, e decodifica escapes", () => {
    expect(readResetToken("#token=a+b%2Fc")).toBe("a+b/c");
  });

  it.each(["", "#", "#token=", "#token=%20", "#outro=abc", "#tokenx=abc", "#token=%E0%A4%A"])(
    "sem token válido (%s) devolve null",
    (hash) => {
      expect(readResetToken(hash)).toBeNull();
    },
  );
});
