// Package internal holds the architecture test: it checks the module boundaries
// of docs/architecture/domain-boundaries.md against the imports of production code.
package internal

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
)

const importPrefix = "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/"

// pkg is a package under api/internal, by its path relative to it ("identity/domain"),
// with the import paths its production files use.
type pkg struct {
	Path    string
	Imports []string
}

var layers = []string{"domain", "app", "infra", "http"}

func segments(rel string) (module, layer string) {
	parts := strings.Split(rel, "/")
	module = parts[0]
	if len(parts) > 1 && slices.Contains(layers, parts[1]) {
		layer = parts[1]
	}
	return module, layer
}

// internalImport returns the path of an import relative to api/internal, or "" for
// the standard library and third-party packages.
func internalImport(imp string) string {
	if rest, ok := strings.CutPrefix(imp, importPrefix); ok {
		return rest
	}
	return ""
}

// businessModules finds the modules from the directories: a directory of internal/
// other than platform that has a domain package. There is no list to keep.
func businessModules(pkgs []pkg) []string {
	seen := map[string]bool{}
	for _, p := range pkgs {
		module, layer := segments(p.Path)
		if module != "platform" && layer == "domain" {
			seen[module] = true
		}
	}
	out := make([]string, 0, len(seen))
	for m := range seen {
		out = append(out, m)
	}
	sort.Strings(out)
	return out
}

// checkBoundaries returns the violations of the boundary rules:
//
//  1. a module does not import the domain, infra or http package of another module;
//  2. platform does not import any business module;
//  3. inside a module the dependencies point inward: domain imports no other layer of
//     the module and no third-party package, app imports neither infra nor http, and
//     infra and http do not import each other.
func checkBoundaries(pkgs []pkg) []string {
	modules := businessModules(pkgs)
	isModule := func(m string) bool { return slices.Contains(modules, m) }
	var out []string
	add := func(p pkg, imp, why string) { out = append(out, p.Path+" importa "+imp+": "+why) }

	for _, p := range pkgs {
		module, layer := segments(p.Path)
		for _, imp := range p.Imports {
			rel := internalImport(imp)
			if rel == "" {
				if module != "platform" && layer == "domain" && strings.Contains(strings.SplitN(imp, "/", 2)[0], ".") {
					add(p, imp, "o domain não usa pacotes de terceiros (adaptadores externos)")
				}
				continue
			}
			target, targetLayer := segments(rel)
			switch {
			case module == "platform" && isModule(target):
				add(p, rel, "platform não importa módulo de negócio")
			case isModule(module) && isModule(target) && target != module && slices.Contains([]string{"domain", "infra", "http"}, targetLayer):
				add(p, rel, "módulo não importa domain, infra nem http de outro módulo")
			case isModule(module) && target == module:
				switch {
				case layer == "domain" && targetLayer != "" && targetLayer != "domain":
					add(p, rel, "o domain não importa as outras camadas do módulo")
				case layer == "app" && (targetLayer == "infra" || targetLayer == "http"):
					add(p, rel, "app não importa infra nem http")
				case layer == "infra" && targetLayer == "http", layer == "http" && targetLayer == "infra":
					add(p, rel, "infra e http não se importam")
				}
			}
		}
	}
	sort.Strings(out)
	return out
}

// loadPackages reads the production files (not _test.go) under dir, which is api/internal.
func loadPackages(t *testing.T, dir string) []pkg {
	t.Helper()
	byDir := map[string]map[string]bool{}
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(dir, filepath.Dir(path))
		rel = filepath.ToSlash(rel)
		if byDir[rel] == nil {
			byDir[rel] = map[string]bool{}
		}
		for _, imp := range file.Imports {
			byDir[rel][strings.Trim(imp.Path.Value, `"`)] = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("ler pacotes: %v", err)
	}
	var out []pkg
	for rel, imports := range byDir {
		if rel == "." {
			continue
		}
		p := pkg{Path: rel}
		for imp := range imports {
			p.Imports = append(p.Imports, imp)
		}
		sort.Strings(p.Imports)
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func in(rel string) string { return importPrefix + rel }

// A minimal healthy tree: a module with its layers, the platform and a composition package.
func healthy() []pkg {
	return []pkg{
		{Path: "platform/authz"},
		{Path: "identity/domain", Imports: []string{in("platform/authz"), "crypto/sha256"}},
		{Path: "identity/app", Imports: []string{in("identity/domain"), in("platform/authz")}},
		{Path: "identity/infra", Imports: []string{in("identity/app"), in("identity/domain"), "gorm.io/gorm"}},
		{Path: "identity/http", Imports: []string{in("identity/app"), "github.com/gin-gonic/gin"}},
		{Path: "loja/domain"},
		{Path: "loja/app", Imports: []string{in("identity/app"), in("loja/domain")}},
		{Path: "httpapi", Imports: []string{in("identity/infra"), in("loja/http")}},
	}
}

func with(extra pkg, base []pkg) []pkg {
	out := slices.Clone(base)
	for i, p := range out {
		if p.Path == extra.Path {
			out[i].Imports = append(slices.Clone(p.Imports), extra.Imports...)
			return out
		}
	}
	return append(out, extra)
}

// TST-02.1 e .2 no código real: nenhuma violação, e o teste não passa no vazio.
func TestProductionCodeRespectsTheModuleBoundaries(t *testing.T) {
	pkgs := loadPackages(t, ".")

	if modules := businessModules(pkgs); !slices.Contains(modules, "identity") || slices.Contains(modules, "platform") {
		t.Fatalf("os módulos deveriam ser descobertos pelos diretórios (identity sim, platform não): %v", modules)
	}
	if len(pkgs) < 15 {
		t.Fatalf("poucos pacotes lidos (%d): o teste estaria passando no vazio", len(pkgs))
	}
	if violations := checkBoundaries(pkgs); len(violations) > 0 {
		t.Errorf("violações das fronteiras:\n  %s", strings.Join(violations, "\n  "))
	}
}

func TestAHealthyTreeHasNoViolations(t *testing.T) {
	if v := checkBoundaries(healthy()); len(v) > 0 {
		t.Errorf("violações: %v", v)
	}
}

// TST-02.1 e .3: um módulo importando domain, infra ou http de outro.
func TestDetectsAModuleImportingAnotherModulesInternals(t *testing.T) {
	for _, target := range []string{"identity/domain", "identity/infra", "identity/http"} {
		v := checkBoundaries(with(pkg{Path: "loja/app", Imports: []string{in(target)}}, healthy()))

		if len(v) != 1 || !strings.Contains(v[0], "loja/app importa "+target) {
			t.Errorf("%s: violações = %v", target, v)
		}
	}
}

// A interface publicada (app) e o pacote raiz do outro módulo são permitidos.
func TestAllowsAModuleUsingAnotherModulesPublishedApp(t *testing.T) {
	v := checkBoundaries(with(pkg{Path: "loja/app", Imports: []string{in("identity/app"), in("identity")}}, healthy()))

	if len(v) != 0 {
		t.Errorf("violações: %v", v)
	}
}

// TST-02.2 e .3: platform importando um módulo de negócio, em qualquer camada.
func TestDetectsPlatformImportingABusinessModule(t *testing.T) {
	for _, target := range []string{"identity/app", "identity/domain", "identity", "loja/domain"} {
		v := checkBoundaries(with(pkg{Path: "platform/authz", Imports: []string{in(target)}}, healthy()))

		if len(v) != 1 || !strings.Contains(v[0], "platform não importa módulo de negócio") {
			t.Errorf("%s: violações = %v", target, v)
		}
	}
}

// TST-02.4: as dependências apontam para dentro.
func TestDetectsOutwardDependenciesInsideAModule(t *testing.T) {
	cases := map[string]pkg{
		"domain importa app":    {Path: "identity/domain", Imports: []string{in("identity/app")}},
		"domain importa infra":  {Path: "identity/domain", Imports: []string{in("identity/infra")}},
		"domain importa http":   {Path: "identity/domain", Imports: []string{in("identity/http")}},
		"domain importa o gorm": {Path: "identity/domain", Imports: []string{"gorm.io/gorm"}},
		"domain importa o gin":  {Path: "identity/domain", Imports: []string{"github.com/gin-gonic/gin"}},
		"app importa infra":     {Path: "identity/app", Imports: []string{in("identity/infra")}},
		"app importa http":      {Path: "identity/app", Imports: []string{in("identity/http")}},
		"infra importa http":    {Path: "identity/infra", Imports: []string{in("identity/http")}},
		"http importa infra":    {Path: "identity/http", Imports: []string{in("identity/infra")}},
	}
	for name, bad := range cases {
		if v := checkBoundaries(with(bad, healthy())); len(v) != 1 {
			t.Errorf("%s: violações = %v", name, v)
		}
	}
}

// O domain pode usar a biblioteca padrão e o platform; app, infra e http podem usar terceiros.
func TestAllowsTheDomainToUseTheStandardLibraryAndPlatform(t *testing.T) {
	v := checkBoundaries(with(pkg{Path: "identity/domain", Imports: []string{"regexp", "time", in("platform/password")}}, healthy()))

	if len(v) != 0 {
		t.Errorf("violações: %v", v)
	}
}

// TST-02.5: um módulo novo é coberto desde o primeiro commit, sem lista.
func TestAFutureModuleIsCoveredWithoutAList(t *testing.T) {
	pkgs := append(healthy(), pkg{Path: "estoque/domain"}, pkg{Path: "estoque/app", Imports: []string{in("estoque/domain")}})

	if !slices.Contains(businessModules(pkgs), "estoque") {
		t.Fatalf("o módulo estoque deveria ser descoberto: %v", businessModules(pkgs))
	}
	v := checkBoundaries(with(pkg{Path: "estoque/app", Imports: []string{in("identity/infra")}}, pkgs))
	if len(v) != 1 || !strings.Contains(v[0], "estoque/app importa identity/infra") {
		t.Errorf("violações = %v", v)
	}
	v = checkBoundaries(with(pkg{Path: "platform/authz", Imports: []string{in("estoque/app")}}, pkgs))
	if len(v) != 1 {
		t.Errorf("platform importando o módulo novo: %v", v)
	}
}

// TST-02.5: arquivos _test.go não contam como imports.
func TestTestFilesAreIgnored(t *testing.T) {
	dir := t.TempDir()
	write := func(rel, src string) {
		path := filepath.Join(dir, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("identity/domain/user.go", "package domain\n")
	write("identity/domain/user_test.go", "package domain\n\nimport _ \""+importPrefix+"identity/infra\"\nimport _ \"gorm.io/gorm\"\n")
	write("identity/infra/repo.go", "package infra\n\nimport _ \""+importPrefix+"identity/domain\"\n")

	pkgs := loadPackages(t, dir)

	if v := checkBoundaries(pkgs); len(v) != 0 {
		t.Errorf("o import de teste não deveria contar: %v", v)
	}
	write("identity/domain/leak.go", "package domain\n\nimport _ \""+importPrefix+"identity/infra\"\n")
	if v := checkBoundaries(loadPackages(t, dir)); len(v) != 1 {
		t.Errorf("o import de produção deveria ser detectado: %v", v)
	}
}

// O platform nunca é um módulo de negócio, mesmo que tenha um pacote chamado domain.
func TestPlatformIsNeverABusinessModule(t *testing.T) {
	pkgs := append(healthy(), pkg{Path: "platform/domain"})

	if slices.Contains(businessModules(pkgs), "platform") {
		t.Error("platform não é módulo de negócio")
	}
}
