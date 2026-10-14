import { readdirSync, readFileSync, statSync } from "node:fs";
import { join } from "node:path";

import { describe, expect, it } from "vitest";

import { findRawValues, WEB_ROOT, type GuardException } from "./guard";

// Guarda global de design (DS-INT, T11; DSM-05 AC1 a AC5, DSF-01 AC7, DSC-05
// AC4): aplica a guarda de DS-F a todo o front (`app`, `components` e
// `features`) e acrescenta as verificações de padrão de página e de alias
// removidos. Ficam fora os testes (citam classes e nomes proibidos nas
// próprias asserções) e `app/globals.css` (a folha de tokens, onde os valores
// primitivos são definidos). As guardas de cada frente continuam valendo; esta
// cobre também o que ficou entre elas.

const ROOTS = ["app", "components", "features"] as const;

function walk(path: string): string[] {
  const full = join(WEB_ROOT, path);
  if (!statSync(full).isDirectory()) return [path];
  return readdirSync(full).flatMap((name) => walk(`${path}/${name}`));
}

const sources = ROOTS.flatMap(walk).filter(
  (file) => /\.(?:tsx?|css)$/.test(file) && !/\.test\.tsx?$/.test(file) && file !== "app/globals.css",
);
const read = (file: string) => readFileSync(join(WEB_ROOT, file), "utf8");

// Exceções consolidadas da guarda de valores soltos, cada uma com o motivo.
const EXCEPTIONS: GuardException[] = [
  {
    file: "components/ui/typography.tsx",
    rules: ["font-display", "uppercase"],
    reason:
      "Papéis tipográficos de identidade (DisplayTitle, PageTitle, SectionTitle, Eyebrow): o único lugar da fonte condensada e da caixa alta (design/01-fundacao-visual.md, Tipografia; DSF-03 AC3 e AC4).",
  },
  {
    file: "app/layout.tsx",
    rules: ["literal-color"],
    reason:
      "A meta theme-color não lê variável CSS: repete o primitivo --tj-ink (#111111), e test/design/fonts.test.ts confere a igualdade (DSS-06).",
  },
];

// Onde cada padrão de página pode ser montado.
const SELECT_OWNER = "components/ui/native-select.tsx";
const H1_OWNERS = ["components/ui/typography.tsx"]; // PageTitle e DisplayTitle; PageHeader, AccessLayout, Sem acesso e erros os usam.
const MAIN_OWNERS = ["components/app/page-container.tsx", "components/app/access-layout.tsx"];
// Larguras de conteúdo (`max-w-content-*`) são do contêiner de página, do layout
// de formulário, do estado "Sem acesso" e do diálogo largo.
const CONTENT_WIDTH_OWNERS = [
  "components/app/page-container.tsx",
  "components/app/form-layout.tsx",
  "components/app/require-permission.tsx",
  "components/ui/dialog.tsx",
];
// Larguras que não são de página: limite de célula e de elemento interno.
const NON_PAGE_WIDTH = /^max-w-(?:0|full|none|min|max|fit|\d+(?:\.\d+)?)$/;

/** Abre cada elemento JSX `<Name ...>` e devolve o texto dos atributos. */
function openingTags(source: string, name: string): string[] {
  const tags: string[] = [];
  const start = new RegExp(`<${name}(?![\\w.])`, "g");
  for (const hit of source.matchAll(start)) {
    let depth = 0;
    let quote: string | null = null;
    let i = hit.index + hit[0].length;
    for (; i < source.length; i++) {
      const char = source[i];
      if (quote) {
        if (char === quote) quote = null;
      } else if (char === '"' || char === "'" || char === "`") {
        quote = char;
      } else if (char === "{") {
        depth++;
      } else if (char === "}") {
        depth--;
      } else if (char === ">" && depth === 0) {
        break;
      }
    }
    tags.push(source.slice(hit.index + hit[0].length, i));
  }
  return tags;
}

// Atributos de formulário (DSP-04 AC6 e AC7). Campos de motivo, filtros e
// busca declaram `name` (nos filtros, igual à chave da URL) e
// `autoComplete="off"`; os campos de e-mail e de código desativam a correção
// ortográfica.
const FORM_CONTROLS = ["Input", "Textarea", "NativeSelect"] as const;
const NAMED_CONTROLS: Record<string, readonly string[]> = {
  "components/app/confirm-dialog.tsx": ["motivo"],
  "features/estoque/ajustes/ajuste-dialog.tsx": ["motivo"],
  "features/identity/usuarios/dialogs/admin-dialogs.tsx": ["motivo"],
  "features/identity/usuarios/dialogs/temporary-password-dialog.tsx": ["motivo"],
  "features/estoque/produtos/produtos-page.tsx": ["busca"],
  "features/financeiro/lancamentos/lancamentos-page.tsx": ["tipo", "status", "conta", "de", "ate"],
  "features/identity/usuarios/users-page.tsx": ["situacao", "papel"],
};
const controlTags = (source: string) => FORM_CONTROLS.flatMap((name) => openingTags(source, name));
const isEmailOrCode = (tag: string) => /\btype="email"|register\("(?:email|codigo)"/.test(tag);

type AliasCheck = { name: string; find: (source: string) => boolean };

// Alias de compatibilidade removidos em DS-INT. O compilador já recusa todos;
// a guarda mantém a recusa explícita e legível no relato.
const ALIASES: readonly AliasCheck[] = [
  {
    name: 'Button com variant="default" ou "outline"',
    find: (source) => openingTags(source, "Button").some((tag) => /\bvariant="(?:default|outline)"/.test(tag)),
  },
  {
    name: 'Button com size="default", "icon" ou "icon-sm"',
    find: (source) =>
      openingTags(source, "Button").some((tag) => /\bsize="(?:default|icon|icon-sm)"/.test(tag)),
  },
  {
    name: 'Badge com variant="default", "secondary" ou "destructive"',
    find: (source) =>
      openingTags(source, "Badge").some((tag) => /\bvariant="(?:default|secondary|destructive)"/.test(tag)),
  },
  {
    name: "Alert sem tom",
    find: (source) => openingTags(source, "Alert").some((tag) => !/\bvariant=/.test(tag)),
  },
  {
    name: 'Alert com variant="default" ou "destructive"',
    find: (source) => openingTags(source, "Alert").some((tag) => /\bvariant="(?:default|destructive)"/.test(tag)),
  },
  {
    name: 'Table com scroll="legacy"',
    find: (source) => openingTags(source, "Table").some((tag) => /\bscroll="legacy"/.test(tag)),
  },
  {
    name: "StatusBadge com status em texto",
    find: (source) => openingTags(source, "StatusBadge").some((tag) => /\bstatus=["']/.test(tag)),
  },
  {
    name: "activeStatus ou tipo Status legado",
    find: (source) => /\bactiveStatus\(|\btype Status\b|\bStatus\b[^;]*from "@\/components\/app\/status-badge"/.test(source),
  },
];

describe("guarda global de design", () => {
  it("lê os três diretórios do front, sem testes e sem a folha de tokens", () => {
    expect(sources).toEqual(
      expect.arrayContaining([
        "app/layout.tsx",
        "app/(app)/layout.tsx",
        "app/(dev)/design-system/page.tsx",
        "components/ui/button.tsx",
        "components/app/page-header.tsx",
        "features/financeiro/lancamentos/lancamentos-page.tsx",
        "features/estoque/produtos/produtos-page.tsx",
        "features/identity/usuarios/users-page.tsx",
      ]),
    );
    expect(sources.some((file) => /\.test\.tsx?$/.test(file))).toBe(false);
    expect(sources).not.toContain("app/globals.css");
  });

  it("nenhum valor visual solto fora das exceções consolidadas", () => {
    const found = findRawValues(sources, { exceptions: EXCEPTIONS });
    expect(found.map(({ file, line, rule, match }) => `${file}:${line} ${rule} ${match}`)).toEqual([]);
  });

  it("cada exceção tem motivo e é necessária: sem ela, a guarda aponta exatamente aquele arquivo e aquelas regras", () => {
    for (const exception of EXCEPTIONS) {
      expect(exception.reason.trim(), exception.file).not.toBe("");
      const rules = [...new Set(findRawValues([exception.file]).map((hit) => hit.rule))].sort();
      expect(rules, exception.file).toEqual([...exception.rules].sort());
    }
    expect(EXCEPTIONS.map((exception) => exception.file)).toEqual(["components/ui/typography.tsx", "app/layout.tsx"]);
  });

  it(`nenhum <select> fora de ${SELECT_OWNER}`, () => {
    const hits = sources.filter((file) => file !== SELECT_OWNER && /<select[\s>]/.test(read(file)));
    expect(hits).toEqual([]);
    expect(read(SELECT_OWNER)).toMatch(/<select[\s>]/);
  });

  it("nenhum h1 montado à mão: o título de página vem de PageTitle, por PageHeader, AccessLayout, Sem acesso e páginas de erro", () => {
    const hits = sources.filter((file) => !H1_OWNERS.includes(file) && /<h1[\s>]/.test(read(file)));
    expect(hits).toEqual([]);
  });

  it("nenhum main montado à mão: o contêiner de página vem de PageContainer ou AccessLayout", () => {
    const hits = sources.filter((file) => !MAIN_OWNERS.includes(file) && /<main[\s>]/.test(read(file)));
    expect(hits).toEqual([]);
  });

  it("nenhuma largura de página montada à mão", () => {
    const hits = sources.flatMap((file) =>
      [...read(file).matchAll(/(?<![\w-])(?:[a-z0-9]+:)*(max-w-[\w./-]+)/g)]
        .map((match) => match[1])
        .filter((width) => {
          if (width.startsWith("max-w-content-")) return !CONTENT_WIDTH_OWNERS.includes(file);
          // Larguras fora da escala de conteúdo só no kit e nos compostos (gaveta, menu, célula).
          if (file.startsWith("components/")) return false;
          return !NON_PAGE_WIDTH.test(width);
        })
        .map((width) => `${file}: ${width}`),
    );
    expect(hits).toEqual([]);
  });

  it('nenhum título montado com "text-xl font-semibold"', () => {
    const hits = sources.filter((file) => /text-xl font-semibold|font-semibold text-xl/.test(read(file)));
    expect(hits).toEqual([]);
  });

  it("campos de motivo, filtros e busca declaram name (a chave da URL nos filtros) e autoComplete=\"off\"", () => {
    for (const [file, names] of Object.entries(NAMED_CONTROLS)) {
      const tags = controlTags(read(file));
      expect(tags.map((tag) => /\bname="([^"]+)"/.exec(tag)?.[1] ?? null).sort(), file).toEqual([...names].sort());
      expect(tags.filter((tag) => !/\bautoComplete="off"/.test(tag)), file).toEqual([]);
    }
  });

  it("todo NativeSelect e Textarea do front declara name e autoComplete, e todo QuantityField declara name (DSP-04 AC6)", () => {
    const missing = sources
      .filter((file) => !file.startsWith("components/ui/") && file.endsWith(".tsx"))
      .flatMap((file) => {
        const source = read(file);
        const textual = ["NativeSelect", "Textarea"].flatMap((name) =>
          openingTags(source, name)
            .filter((tag) => !/\bname=/.test(tag) || !/\bautoComplete=/.test(tag))
            .map((tag) => `${file}: <${name}${tag.slice(0, 60)}`),
        );
        // `{...field}` do Controller do React Hook Form já traz o `name`.
        const quantity = openingTags(source, "QuantityField")
          .filter((tag) => !/\bname=|\{\.\.\.field\}/.test(tag))
          .map((tag) => `${file}: <QuantityField${tag.slice(0, 60)}`);
        return [...textual, ...quantity];
      });
    expect(missing).toEqual([]);
  });

  // Verificação, iteração 3, D2 e D3: a guarda acima não cobria `Input` nem
  // `MoneyField`. Todo `Input` declara `name` e `autoComplete`; o de arquivo
  // (`type="file"`) só `name`, porque `autocomplete` não se aplica a ele. Todo
  // `MoneyField` declara `name` (o componente não repassa `autoComplete`).
  // `{...field}` do Controller e `{...register(...)}` do React Hook Form já
  // trazem o `name`; o `autoComplete` continua exigido no `Input`.
  it("todo Input do front declara name e autoComplete (o de arquivo só name), e todo MoneyField declara name", () => {
    const spreadName = /\{\.\.\.(?:field\}|register\()/;
    const missing = sources
      .filter((file) => !file.startsWith("components/ui/") && file.endsWith(".tsx"))
      .flatMap((file) => {
        const source = read(file);
        const inputs = openingTags(source, "Input")
          .filter((tag) => {
            const named = /\bname=/.test(tag) || spreadName.test(tag);
            const isFile = /\btype="file"/.test(tag);
            return !named || (!isFile && !/\bautoComplete=/.test(tag));
          })
          .map((tag) => `${file}: <Input${tag.slice(0, 60)}`);
        const money = openingTags(source, "MoneyField")
          .filter((tag) => !/\bname=/.test(tag) && !spreadName.test(tag))
          .map((tag) => `${file}: <MoneyField${tag.slice(0, 60)}`);
        return [...inputs, ...money];
      });
    expect(missing).toEqual([]);
  });

  it("número e unidade de quantidade separados por espaço não separável (DSP-07 AC4)", () => {
    const offenders = sources
      .filter((file) => file.endsWith(".tsx"))
      .flatMap((file) =>
        read(file)
          .split("\n")
          .map((line, index) => ({ line, index }))
          .filter(({ line }) => /\} \$\{(?:produto\.)?(?:unidade|unidade_medida)\}/.test(line))
          .map(({ index }) => `${file}:${index + 1}`),
      );
    expect(offenders).toEqual([]);
  });

  it("campos de e-mail e de código desativam a correção ortográfica e a maiúscula automática", () => {
    const tags = sources.flatMap((file) => controlTags(read(file)).filter(isEmailOrCode).map((tag) => ({ file, tag })));
    expect(tags.map(({ file }) => file)).toEqual(
      expect.arrayContaining(["features/estoque/produtos/create-produto-dialog.tsx", "app/(public)/entrar/login-form.tsx"]),
    );
    expect(tags.filter(({ tag }) => !/\bspellCheck=\{false\}/.test(tag)).map(({ file }) => file)).toEqual([]);
    const codigo = tags.filter(({ tag }) => /register\("codigo"/.test(tag));
    expect(codigo).toHaveLength(1);
    expect(codigo[0].tag).toMatch(/\bautoCapitalize="none"/);
  });

  it("todo placeholder literal termina em reticências (…)", () => {
    const hits = sources.flatMap((file) =>
      [...read(file).matchAll(/\bplaceholder="([^"]*)"/g)]
        .filter((match) => !match[1].endsWith("…"))
        .map((match) => `${file}: ${match[1]}`),
    );
    expect(hits).toEqual([]);
  });

  it("nenhum uso de alias removido", () => {
    const hits = sources.flatMap((file) => {
      const source = read(file);
      return ALIASES.filter(({ find }) => find(source)).map(({ name }) => `${file}: ${name}`);
    });
    expect(hits).toEqual([]);
  });
});

describe("controles da guarda global", () => {
  it("a verificação de alias aponta usos reais, inclusive com atributos em várias linhas e funções", () => {
    const legacy = [
      '<Button onClick={() => go(">")} variant="outline">x</Button>',
      '<Button\n  size="icon-sm"\n  aria-label="Abrir"\n/>',
      '<Badge variant="secondary">x</Badge>',
      "<Alert>\n  <AlertDescription>x</AlertDescription>\n</Alert>",
      '<Alert variant="destructive">',
      '<Table scroll="legacy" aria-label="x">',
      '<StatusBadge status="CANCELADA" />',
      "const s = activeStatus(true);",
    ].join("\n");
    expect(ALIASES.filter(({ find }) => find(legacy)).map(({ name }) => name)).toEqual([
      'Button com variant="default" ou "outline"',
      'Button com size="default", "icon" ou "icon-sm"',
      'Badge com variant="default", "secondary" ou "destructive"',
      "Alert sem tom",
      'Alert com variant="default" ou "destructive"',
      'Table com scroll="legacy"',
      "StatusBadge com status em texto",
      "activeStatus ou tipo Status legado",
    ]);
  });

  it("a forma nova não é apontada", () => {
    const current = [
      '<Button variant="secondary" size="sm" onClick={() => go()}>x</Button>',
      '<Badge variant="outline">Papel</Badge>',
      '<Alert\n  variant="warning">\n<AlertDescription>x</AlertDescription>\n</Alert>',
      '<Table scroll="scrollable" aria-label="x">',
      "<StatusBadge status={activeStatusDef(true)} />",
      '<AlertDialogAction variant="destructive">',
      'import { activeStatusDef, type StatusDef } from "@/components/app/status-badge";',
    ].join("\n");
    expect(ALIASES.filter(({ find }) => find(current)).map(({ name }) => name)).toEqual([]);
  });

  it("a guarda de valores soltos lê de fato os arquivos: sem exceções, aponta a tipografia", () => {
    const found = findRawValues(sources);
    expect([...new Set(found.map((hit) => hit.file))].sort()).toEqual(
      ["app/layout.tsx", "components/ui/typography.tsx"].sort(),
    );
  });
});
