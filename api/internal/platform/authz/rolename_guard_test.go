package authz

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// The role names of the project (ADR-009). The definition file below is the
// only place allowed to spell them; TestGuardRoleListMatchesTheRoleDefinitions
// keeps this list in sync with it.
var roleNames = []string{
	"ASSOCIADO", "PRESIDENTE", "DIRETORIA", "TESOURARIA", "ESTOQUE_LOJA", "EVENTOS", "CONSELHO_FISCAL", "ADMIN_SISTEMA",
}

const roleDefinitionFile = "identity/domain/role.go"

func isRoleName(s string) bool { return slices.Contains(roleNames, s) }

// roleNameViolations walks the Go sources under root and reports:
//   - a string literal equal to a role name, outside the role definition file;
//   - a read of a `.Roles` field outside the identity module.
//
// Test files and generated code are ignored. RBAC-02.8: authorization decides by
// permission, never by role name.
func roleNameViolations(root string) ([]string, error) {
	var out []string
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") || strings.HasSuffix(name, ".gen.go") {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		file, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		inIdentity := strings.HasPrefix(rel, "identity/") || strings.Contains(rel, "/identity/")
		ast.Inspect(file, func(n ast.Node) bool {
			switch v := n.(type) {
			case *ast.BasicLit:
				if v.Kind == token.STRING && !strings.HasSuffix(rel, roleDefinitionFile) {
					if s, err := strconv.Unquote(v.Value); err == nil && isRoleName(s) {
						out = append(out, rel+": nome de papel literal "+v.Value)
					}
				}
			case *ast.SelectorExpr:
				if v.Sel.Name == "Roles" && !inIdentity {
					out = append(out, rel+": leitura de .Roles fora de identity")
				}
			}
			return true
		})
		return nil
	})
	sort.Strings(out)
	return out, err
}

func TestNoAuthorizationDecisionUsesRoleNamesInTheSourceTree(t *testing.T) {
	violations, err := roleNameViolations(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("percorrer o código: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("decisão por nome de papel (RBAC-02.8): %v", violations)
	}
}

func writeFile(t *testing.T, root, rel, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestGuardDetectsRoleNameLiteralsAndRolesReadsOutsideIdentity(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "financeiro/app/x.go", "package app\nfunc f(r string) bool { return r == \"PRESIDENTE\" }\n")
	writeFile(t, root, "platform/httpx/y.go", "package httpx\ntype P struct{ Roles []string }\nfunc g(p P) int { return len(p.Roles) }\n")

	got, err := roleNameViolations(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || !strings.Contains(got[0], "financeiro/app/x.go") || !strings.Contains(got[1], "platform/httpx/y.go") {
		t.Errorf("violações = %v", got)
	}
}

func TestGuardAllowsTheRoleDefinitionIdentityCodeTestsAndGeneratedFiles(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "identity/domain/role.go", "package domain\nconst Presidente = \"PRESIDENTE\"\n")
	writeFile(t, root, "identity/http/me.go", "package http\ntype P struct{ Roles []string }\nfunc g(p P) int { return len(p.Roles) }\n")
	writeFile(t, root, "financeiro/app/x_test.go", "package app\nvar _ = \"PRESIDENTE\"\n")
	writeFile(t, root, "identity/http/api.gen.go", "package http\nvar _ = \"CONSELHO_FISCAL\"\n")

	if got, err := roleNameViolations(root); err != nil || len(got) != 0 {
		t.Errorf("violações = %v, err = %v", got, err)
	}
}

func TestGuardOnlyFlagsExactRoleNames(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "a/x.go", "package a\nvar _ = \"presidente\"\nvar _ = \"PRESIDENTE_X\"\nvar _ = \"identity:user:read\"\n")

	if got, err := roleNameViolations(root); err != nil || len(got) != 0 {
		t.Errorf("violações = %v, err = %v", got, err)
	}
}

// Keeps the guard's list equal to the constants of the role definition file,
// once that file exists.
func TestGuardRoleListMatchesTheRoleDefinitions(t *testing.T) {
	path := filepath.Join("..", "..", filepath.FromSlash(roleDefinitionFile))
	if _, err := os.Stat(path); err != nil {
		return // the definition file arrives with the role catalog
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var defined []string
	ast.Inspect(file, func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			if s, err := strconv.Unquote(lit.Value); err == nil && s == strings.ToUpper(s) && s != "" {
				defined = append(defined, s)
			}
		}
		return true
	})
	sort.Strings(defined)
	want := append([]string(nil), roleNames...)
	sort.Strings(want)
	if strings.Join(defined, ",") != strings.Join(want, ",") {
		t.Errorf("papéis em role.go = %v, lista do guard = %v", defined, want)
	}
}
