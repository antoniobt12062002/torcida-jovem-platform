import { readdirSync, readFileSync } from "node:fs"
import { join } from "node:path"

import { describe, expect, it } from "vitest"

import { findRawValues, GUARD_RULES, WEB_ROOT, type GuardException } from "@/test/design/guard"

// Guarda de tokens do kit: aplica todas as regras da guarda de design (DS-F) a
// cada componente de `components/ui`. Os arquivos de teste ficam fora do
// escopo: citam os nomes das classes proibidas nas próprias asserções.

const KIT = "components/ui"
const files = readdirSync(join(WEB_ROOT, KIT)).filter((name) => /\.tsx?$/.test(name))
const tests = files.filter((name) => /\.test\.tsx?$/.test(name))
const components = files.filter((name) => !/\.test\.tsx?$/.test(name))

// Exceções: só as que o design prevê, cada uma com o motivo.
const EXCEPTIONS: GuardException[] = [
  {
    file: `${KIT}/typography.tsx`,
    rules: ["font-display", "uppercase"],
    reason:
      "Papéis tipográficos de identidade (DisplayTitle, PageTitle, SectionTitle, Eyebrow): com o shell, o único lugar da fonte condensada e da caixa alta (design/01-fundacao-visual.md, Tipografia; DSF-03 AC3-4).",
  },
]

describe("guarda de tokens do kit", () => {
  it("encontra os componentes do kit", () => {
    expect(components).toEqual(expect.arrayContaining(["button.tsx", "typography.tsx", "table.tsx"]))
    expect(components.length).toBeGreaterThanOrEqual(23)
  })

  it("nenhum componente tem valor visual solto", () => {
    const found = findRawValues(
      components.map((name) => `${KIT}/${name}`),
      { exceptions: EXCEPTIONS }
    )
    expect(found.map(({ file, line, rule, match }) => `${file}:${line} ${rule} ${match}`)).toEqual([])
  })

  it("as exceções são só as do arquivo de tipografia, restritas às duas regras de identidade", () => {
    expect(EXCEPTIONS.map((exception) => exception.file)).toEqual([`${KIT}/typography.tsx`])
    expect(EXCEPTIONS[0].rules).toEqual(["font-display", "uppercase"])
    expect(EXCEPTIONS[0].reason.trim()).not.toBe("")
  })

  it("sem as exceções, a guarda só aponta a tipografia, e só por fonte de identidade e caixa alta", () => {
    const found = findRawValues(components.map((name) => `${KIT}/${name}`))
    expect(found.length).toBeGreaterThan(0)
    expect([...new Set(found.map((hit) => hit.file))]).toEqual([`${KIT}/typography.tsx`])
    expect([...new Set(found.map((hit) => hit.rule))].sort()).toEqual(["font-display", "uppercase"])
  })

  it("a guarda aplicada é a de DS-F, com todas as regras", () => {
    expect([...GUARD_RULES].sort()).toEqual(
      [
        "arbitrary-color",
        "arbitrary-radius",
        "arbitrary-spacing",
        "arbitrary-text-size",
        "dark-variant",
        "font-display",
        "literal-color",
        "raw-palette",
        "transition-all",
        "uppercase",
      ].sort()
    )
  })
})

describe("cobertura do kit", () => {
  const sources = tests.map((name) => readFileSync(join(WEB_ROOT, KIT, name), "utf8"))

  it.each(components)("%s é referenciado por pelo menos um teste", (name) => {
    const specifier = `@/components/ui/${name.replace(/\.tsx?$/, "")}`
    expect(sources.some((source) => source.includes(`from "${specifier}"`))).toBe(true)
  })
})
