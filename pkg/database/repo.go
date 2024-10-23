package database

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/navikt/nada-backend/pkg/database/gensql"
	"sync"
	// Pin version of sqlc and goose for cli
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	_ "github.com/sqlc-dev/sqlc"
)

var (
	//go:embed migrations/*.sql
	embedMigrations embed.FS
	once            sync.Once
)

type Repo struct {
	Querier Querier
	queries *gensql.Queries
	db      *sql.DB
}

type Querier interface {
	gensql.Querier
	WithTx(tx *sql.Tx) *gensql.Queries
}

type Transacter interface {
	Commit() error
	Rollback() error
}

// WithTx is a helper function that returns a function that will return a new transaction and the querier
// to be used within the transaction. It allows us to define a subset of the queries to be used within the
// transaction.
func WithTx[T any](r *Repo) func() (T, Transacter, error) {
	return func() (T, Transacter, error) {
		tx, err := r.db.Begin()
		if err != nil {
			return *new(T), nil, fmt.Errorf("begin tx: %w", err)
		}

		return any(r.queries.WithTx(tx)).(T), tx, nil
	}
}

func New(dbConnDSN string, maxIdleConn, maxOpenConn int) (*Repo, error) {
	db, err := sql.Open("pgx", dbConnDSN)
	if err != nil {
		return nil, fmt.Errorf("open sql connection: %w", err)
	}
	db.SetMaxIdleConns(maxIdleConn)
	db.SetMaxOpenConns(maxOpenConn)

	once.Do(func() {
		goose.SetBaseFS(embedMigrations)
	})

	if err := goose.Up(db, "migrations"); err != nil {
		return nil, fmt.Errorf("goose up: %w", err)
	}

	queries := gensql.New(db)
	return &Repo{
		Querier: queries,
		queries: queries,
		db:      db,
	}, nil
}

func (r *Repo) GetDB() *sql.DB {
	return r.db
}
