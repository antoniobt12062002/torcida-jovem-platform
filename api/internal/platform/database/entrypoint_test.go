package database

import (
	"os"
	"strings"
	"testing"
)

// PLT-01.5: a API não aplica migrações ao iniciar; elas rodam por ferramenta separada.
func TestAPIEntrypointDoesNotRunMigrations(t *testing.T) {
	src, err := os.ReadFile("../../../cmd/api/main.go")
	if err != nil {
		t.Fatal(err)
	}
	text := strings.ToLower(string(src))
	if strings.Contains(text, "migrate") {
		t.Error("cmd/api/main.go não deve chamar migrações na partida")
	}
}
