// Contraste WCAG 2.x entre duas cores sRGB em hexadecimal.
// Fórmula: https://www.w3.org/TR/WCAG22/#dfn-contrast-ratio

function channel(value: number) {
  const srgb = value / 255;
  return srgb <= 0.04045 ? srgb / 12.92 : ((srgb + 0.055) / 1.055) ** 2.4;
}

/** Luminância relativa de `#RGB` ou `#RRGGBB`. */
export function relativeLuminance(hex: string) {
  const digits = hex.trim().replace(/^#/, "");
  const full = digits.length === 3 ? [...digits].map((digit) => digit + digit).join("") : digits;
  if (!/^[0-9a-fA-F]{6}$/.test(full)) throw new Error(`Cor inválida: ${hex}`);
  const [red, green, blue] = [0, 2, 4].map((start) => channel(parseInt(full.slice(start, start + 2), 16)));
  return 0.2126 * red + 0.7152 * green + 0.0722 * blue;
}

/** Razão de contraste entre duas cores, de 1 a 21. */
export function contrastRatio(first: string, second: string) {
  const [lighter, darker] = [relativeLuminance(first), relativeLuminance(second)].sort((a, b) => b - a);
  return (lighter + 0.05) / (darker + 0.05);
}
