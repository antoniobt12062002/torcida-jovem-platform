package database

import (
	"bytes"
	"strings"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
)

func TestGormLogsParameterizedStatementWithoutBoundValues(t *testing.T) {
	var buf bytes.Buffer
	cfg := gormConfig(logx.New("debug", &buf))
	cfg.DisableAutomaticPing = true

	db, err := gorm.Open(postgres.Open("postgres://u:p@127.0.0.1:1/db?sslmode=disable&connect_timeout=1"), cfg)
	if err != nil {
		t.Fatalf("gorm.Open não deveria conectar: %v", err)
	}

	err = db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", "hash-segredo-123", 7).Error
	if err == nil {
		t.Fatal("esperava erro de conexão para forçar o log da query")
	}

	out := buf.String()
	if !strings.Contains(out, "UPDATE users SET password_hash") {
		t.Errorf("o log deveria conter a instrução parametrizada: %s", out)
	}
	if strings.Contains(out, "hash-segredo-123") {
		t.Errorf("o log vazou o valor do parâmetro: %s", out)
	}
}
