package db

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect opens a GORM connection to Postgres, retrying while the database starts.
//
// It uses the simple query protocol, so no prepared statements are created. Supabase's
// pooler can hand a new connection a backend that still holds statements from an earlier
// process, which fails with `prepared statement "stmtcache_..." already exists`.
func Connect(dsn string) (*gorm.DB, error) {
	var (
		conn *gorm.DB
		err  error
	)
	for i := 0; i < 15; i++ {
		conn, err = gorm.Open(postgres.New(postgres.Config{DSN: dsn, PreferSimpleProtocol: true}), &gorm.Config{PrepareStmt: false})
		if err == nil {
			sqlDB, derr := conn.DB()
			if derr == nil && sqlDB.Ping() == nil {
				sqlDB.SetMaxOpenConns(20)
				sqlDB.SetMaxIdleConns(5)
				sqlDB.SetConnMaxLifetime(30 * time.Minute)
				return conn, nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("postgres not reachable: %w", err)
}
