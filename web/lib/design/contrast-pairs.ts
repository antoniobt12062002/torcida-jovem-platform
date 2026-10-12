// Pares de contraste do tema. Cada par cita tokens semânticos pelo nome (sem
// `--`). O teste `test/design/contrast.test.ts` resolve os valores em
// `app/globals.css` e exige o mínimo de cada tipo; o espécime mostra a razão.

export type ContrastKind = "text" | "graphic";

export type ContrastPair = {
  foreground: string;
  background: string;
  kind: ContrastKind;
  /** Onde o par aparece na interface. */
  use: string;
};

/** WCAG 2.2 AA: 4,5:1 em texto; 3:1 em gráfico, foco e borda de campo. */
export const CONTRAST_MINIMUM: Record<ContrastKind, number> = { text: 4.5, graphic: 3 };

/** Tokens de texto sem par exigido, com o motivo. */
export const CONTRAST_EXEMPT: Record<string, string> = {
  "disabled-foreground": "Controle desabilitado é isento pela WCAG 2.2 (1.4.3); mantido legível.",
};

export const CONTRAST_PAIRS: readonly ContrastPair[] = [
  // content sobre surface
  { foreground: "foreground", background: "background", kind: "text", use: "Texto principal" },
  { foreground: "foreground", background: "muted", kind: "text", use: "Cabeçalho de tabela e barra de filtros" },
  { foreground: "card-foreground", background: "card", kind: "text", use: "Texto de cartão" },
  { foreground: "popover-foreground", background: "popover", kind: "text", use: "Texto de menu e diálogo" },
  { foreground: "accent-foreground", background: "accent", kind: "text", use: "Linha e item de menu em hover" },
  { foreground: "secondary-foreground", background: "secondary", kind: "text", use: "Preenchimento discreto" },
  { foreground: "muted-foreground", background: "background", kind: "text", use: "Texto secundário" },
  { foreground: "muted-foreground", background: "muted", kind: "text", use: "Texto secundário sobre fundo discreto" },

  // shell
  { foreground: "sidebar-foreground", background: "sidebar", kind: "text", use: "Texto do shell" },
  { foreground: "sidebar-accent-foreground", background: "sidebar-accent", kind: "text", use: "Item do shell em hover" },
  { foreground: "sidebar-muted-foreground", background: "sidebar", kind: "text", use: "Rótulo de grupo do shell" },
  { foreground: "sidebar-muted-foreground", background: "sidebar-accent", kind: "text", use: "Texto secundário do shell em hover" },
  { foreground: "sidebar-primary-foreground", background: "sidebar-primary", kind: "text", use: "Texto sobre o acento do shell (raro)" },
  { foreground: "sidebar-primary", background: "sidebar", kind: "graphic", use: "Indicador do item ativo do shell" },

  // action
  { foreground: "primary-foreground", background: "primary", kind: "text", use: "Botão primário" },
  { foreground: "primary-foreground", background: "primary-hover", kind: "text", use: "Botão primário em hover" },

  // brand
  { foreground: "brand-foreground", background: "brand", kind: "text", use: "Texto sobre o vermelho de marca (raro)" },
  { foreground: "brand", background: "background", kind: "graphic", use: "Filete do título de página e indicador de aba" },
  { foreground: "brand", background: "sidebar", kind: "graphic", use: "Acento de marca sobre o shell" },

  // state
  { foreground: "destructive", background: "background", kind: "text", use: "Mensagem de erro" },
  { foreground: "destructive-foreground", background: "destructive", kind: "text", use: "Botão destrutivo" },
  { foreground: "destructive-foreground", background: "destructive-emphasis", kind: "text", use: "Botão destrutivo em hover" },
  { foreground: "destructive-emphasis", background: "destructive-subtle", kind: "text", use: "Alerta e badge de erro" },
  { foreground: "success", background: "background", kind: "text", use: "Texto e ícone de sucesso" },
  { foreground: "success-emphasis", background: "success-subtle", kind: "text", use: "Alerta e badge de sucesso" },
  { foreground: "warning", background: "background", kind: "text", use: "Texto e ícone de aviso" },
  { foreground: "warning-emphasis", background: "warning-subtle", kind: "text", use: "Alerta e badge de aviso" },
  { foreground: "info", background: "background", kind: "text", use: "Informação neutra" },
  { foreground: "info", background: "info-subtle", kind: "text", use: "Alerta e badge neutros" },

  // bordas e foco
  { foreground: "input", background: "background", kind: "graphic", use: "Borda de campo" },
  { foreground: "destructive", background: "background", kind: "graphic", use: "Borda de campo inválido" },
  { foreground: "border-strong", background: "background", kind: "graphic", use: "Régua de tabela e de título de seção" },
  { foreground: "ring", background: "ring-offset", kind: "graphic", use: "Foco no conteúdo" },
  { foreground: "sidebar-ring", background: "sidebar", kind: "graphic", use: "Foco no shell" },
  { foreground: "sidebar-ring", background: "sidebar-accent", kind: "graphic", use: "Foco no shell sobre item em hover" },
];
