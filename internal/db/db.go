package db

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DB struct {
	db *sql.DB
}

func NewDB(ctx context.Context, dbConnect string) (*DB, error) {
	db, err := sql.Open("pgx", dbConnect)
	if err != nil {
		return nil, err
	}

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}

	_, err = db.ExecContext(
		ctx,
		`create table if not exists public.user
		(
			id       serial primary key,
			login    varchar not null,
			password varchar not null
		);
	`)
	if err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}
