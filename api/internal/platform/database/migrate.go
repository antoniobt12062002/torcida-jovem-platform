package database

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Migrate applies every pending migration found in dir, using the owner
// connection. It is idempotent: with nothing pending it returns nil. The API
// never calls it on startup; migrations run from a separate tool.
func Migrate(ownerDSN, dir string) error {
	u, err := url.Parse(ownerDSN)
	if err != nil {
		return errors.New("migrate: DSN do dono inválida")
	}
	u.Scheme = "pgx5"

	src, err := iofs.New(os.DirFS(dir), ".")
	if err != nil {
		return fmt.Errorf("migrate: ler migrações em %s: %w", dir, err)
	}
	m, err := migrate.NewWithSourceInstance("iofs", src, u.String())
	if err != nil {
		return fmt.Errorf("migrate: preparar migração: %w", err)
	}
	defer func() { _, _ = m.Close() }()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate: aplicar migrações: %w", err)
	}
	return nil
}
