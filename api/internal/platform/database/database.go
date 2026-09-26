package database

import (
	"context"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func Open(dsn string, log *slog.Logger) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), gormConfig(log))
}

func gormConfig(log *slog.Logger) *gorm.Config {
	return &gorm.Config{
		Logger: gormlogger.New(slog.NewLogLogger(log.Handler(), slog.LevelWarn), gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormlogger.Warn,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
		}),
	}
}

func Ping(db *gorm.DB) func(context.Context) error {
	return func(ctx context.Context) error {
		sqlDB, err := db.DB()
		if err != nil {
			return err
		}
		return sqlDB.PingContext(ctx)
	}
}
