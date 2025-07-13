package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TxDB interface {
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	BeginTx(ctx context.Context, opts pgx.TxOptions) (TxDB, error)
	Rollback(ctx context.Context) error
	Commit(ctx context.Context) error
}

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close()
}

type Row interface {
	Scan(dest ...any) error
}

type PgxPoolTxDB struct {
	Pool *pgxpool.Pool
}

func (p *PgxPoolTxDB) Query(ctx context.Context, sql string, args ...any) (Rows, error) {
	return p.Pool.Query(ctx, sql, args...)
}

func (p *PgxPoolTxDB) QueryRow(ctx context.Context, sql string, args ...any) Row {
	return p.Pool.QueryRow(ctx, sql, args...)
}

func (p *PgxPoolTxDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return p.Pool.Exec(ctx, sql, args...)
}

func (p *PgxPoolTxDB) BeginTx(ctx context.Context, opts pgx.TxOptions) (TxDB, error) {
	tx, err := p.Pool.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &PgxTx{Tx: tx}, nil
}

func (p *PgxPoolTxDB) Rollback(ctx context.Context) error {
	return nil
}

func (p *PgxPoolTxDB) Commit(ctx context.Context) error {
	return nil
}

type PgxTx struct {
	Tx pgx.Tx
}

func (t *PgxTx) Query(ctx context.Context, sql string, args ...any) (Rows, error) {
	return t.Tx.Query(ctx, sql, args...)
}

func (t *PgxTx) QueryRow(ctx context.Context, sql string, args ...any) Row {
	return t.Tx.QueryRow(ctx, sql, args...)
}

func (t *PgxTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return t.Tx.Exec(ctx, sql, args...)
}

func (t *PgxTx) BeginTx(ctx context.Context, opts pgx.TxOptions) (TxDB, error) {
	return nil, errors.New("nested transactions are not supported")
}

func (t *PgxTx) Rollback(ctx context.Context) error {
	return t.Tx.Rollback(ctx)
}

func (t *PgxTx) Commit(ctx context.Context) error {
	return t.Tx.Commit(ctx)
}
