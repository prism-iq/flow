package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"flow/internal/config"
	"flow/pkg/logger"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	db  *sql.DB
	log *logger.Logger
}

func NewPostgresDB(cfg *config.Config, log *logger.Logger) (*PostgresDB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	pdb := &PostgresDB{
		db:  db,
		log: log.WithComponent("postgres"),
	}

	if err := pdb.initAGE(ctx); err != nil {
		log.Warn().Err(err).Msg("AGE initialization warning")
	}

	return pdb, nil
}

func (p *PostgresDB) initAGE(ctx context.Context) error {
	queries := []string{
		"LOAD 'age'",
		"SET search_path = ag_catalog, \"$user\", public",
	}

	for _, q := range queries {
		if _, err := p.db.ExecContext(ctx, q); err != nil {
			return err
		}
	}

	return nil
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}

func (p *PostgresDB) DB() *sql.DB {
	return p.db
}

func (p *PostgresDB) Health(ctx context.Context) error {
	return p.db.PingContext(ctx)
}
