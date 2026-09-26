//go:build integration

package testutil

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const defaultPostgresImage = "postgres:16-alpine"

// Postgres is a PostgreSQL 16 container with the tj_owner and tj_app roles
// created by docker/postgres/init/01-roles.sql, the same script used locally.
type Postgres struct {
	container *postgres.PostgresContainer
	host      string
}

var (
	sharedOnce sync.Once
	sharedPG   *Postgres
	sharedErr  error
)

// SharedPostgres starts PostgreSQL once per test binary (one package). The
// testcontainers reaper removes the container when the process exits.
func SharedPostgres(t testing.TB) *Postgres {
	t.Helper()
	sharedOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		sharedPG, sharedErr = startPostgres(ctx)
	})
	if sharedErr != nil {
		t.Fatalf("%v", sharedErr)
	}
	return sharedPG
}

func startPostgres(ctx context.Context) (*Postgres, error) {
	image := os.Getenv("TJ_TEST_POSTGRES_IMAGE")
	if image == "" {
		image = defaultPostgresImage
	}
	c, err := postgres.Run(ctx, image,
		postgres.WithDatabase("tj"),
		postgres.WithUsername("tj"),
		postgres.WithPassword("tj"),
		postgres.WithInitScripts(rolesScript()),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		logs := ""
		if c != nil {
			logs = containerLogs(ctx, c)
			_ = testcontainers.TerminateContainer(c)
		}
		return nil, fmt.Errorf("postgres de teste: falha ao iniciar via Docker (verifique se o Docker está em execução; é necessário com a tag integration): %w\n%s", err, logs)
	}
	dsn, err := c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("postgres de teste: endereço do contêiner: %w", err)
	}
	u, err := url.Parse(dsn)
	if err != nil {
		return nil, fmt.Errorf("postgres de teste: DSN inválida: %w", err)
	}
	return &Postgres{container: c, host: u.Host}, nil
}

// AdminDSN connects as the bootstrap superuser, used only to create and drop
// databases.
func (p *Postgres) AdminDSN(database string) string { return p.dsn("tj", "tj", database) }

// OwnerDSN connects as tj_owner, the role that runs migrations.
func (p *Postgres) OwnerDSN(database string) string {
	return p.dsn("tj_owner", "tj_owner_dev", database)
}

// AppDSN connects as tj_app, the role the API uses.
func (p *Postgres) AppDSN(database string) string { return p.dsn("tj_app", "tj_app_dev", database) }

func (p *Postgres) dsn(user, password, database string) string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, password),
		Host:     p.host,
		Path:     "/" + database,
		RawQuery: "sslmode=disable",
	}
	return u.String()
}

func rolesScript() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "docker", "postgres", "init", "01-roles.sql")
}

func containerLogs(ctx context.Context, c testcontainers.Container) string {
	r, err := c.Logs(ctx)
	if err != nil {
		return "(logs indisponíveis: " + err.Error() + ")"
	}
	defer r.Close()
	b, _ := io.ReadAll(io.LimitReader(r, 8<<10))
	return string(b)
}
