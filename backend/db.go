package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func connectDB(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("нет связи с базой: %w", err)
	}
	return pool, nil
}

// миграции лежат в migrations/NNN_название.sql и применяются по порядку один раз
func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations (version INT PRIMARY KEY)`)
	if err != nil {
		return err
	}

	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return err
	}
	names := []string{}
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)

	for _, name := range names {
		version, err := strconv.Atoi(strings.SplitN(name, "_", 2)[0])
		if err != nil {
			return fmt.Errorf("неверное имя миграции %s", name)
		}

		var applied bool
		err = pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)`, version).Scan(&applied)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		sql, err := migrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			return err
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, string(sql)); err != nil {
			tx.Rollback(ctx)
			return fmt.Errorf("миграция %s: %w", name, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (version) VALUES ($1)`, version); err != nil {
			tx.Rollback(ctx)
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return err
		}
		log.Printf("применена миграция %s", name)
	}
	return nil
}
