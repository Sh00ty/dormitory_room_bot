package pgxbalancer

import (
	"context"
	"fmt"
	"time"

	"github.com/Sh00ty/dormitory_room_bot/internal/logger"
	metric "github.com/Sh00ty/dormitory_room_bot/internal/metrics"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

type TransactionBalancer struct {
	connPool *pgxpool.Pool
}

const (
	poolMaxConns          = 20
	poolMinConns          = 4
	poolMaxConnLifetime   = time.Minute
	poolMaxConnIdleTime   = 5 * time.Second
	poolHealthCheckPeriod = 3 * time.Second
)

type txKey int

const (
	txContextKey txKey = iota
	nestedContextKey
)

type Runner interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

func New(ctx context.Context, host, password, dbname, user string, port uint16) (TransactionBalancer, error) {

	dsn := fmt.Sprintf("user=%s dbname=%s  password=%s host=%s port=%d pool_max_conns=%d pool_min_conns=%d pool_max_conn_lifetime=%s pool_max_conn_idle_time=%s pool_health_check_period=%s",
		user, dbname, password, host, port, poolMaxConns, poolMinConns, poolMaxConnLifetime, poolMaxConnIdleTime, poolHealthCheckPeriod)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return TransactionBalancer{}, err
	}
	p, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return TransactionBalancer{}, err
	}
	return TransactionBalancer{p}, nil
}

func (t *TransactionBalancer) BeginTx(ctx context.Context) (context.Context, error) {
	if _, ok := ctx.Value(txContextKey).(pgx.Tx); ok {
		return context.WithValue(ctx, nestedContextKey, true), nil
	}
	tx, err := t.connPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ctx, fmt.Errorf("failed to begin transaction: %w", err)
	}
	ctx = context.WithValue(ctx, nestedContextKey, false)
	return context.WithValue(ctx, txContextKey, tx), nil
}

func (t *TransactionBalancer) RollBackTx(ctx context.Context) (err error) {
	if tx, ok := ctx.Value(txContextKey).(pgx.Tx); ok {
		if nested, ok := ctx.Value(nestedContextKey).(bool); ok && nested {
			return nil
		}
		if err = tx.Rollback(ctx); err == pgx.ErrTxClosed {
			return fmt.Errorf("failed to rollback transaction: %w", err)
		}
		return err
	}
	return fmt.Errorf("failed to rollback transaction")
}

func (t *TransactionBalancer) CommitTx(ctx context.Context) (err error) {
	if tx, ok := ctx.Value(txContextKey).(pgx.Tx); ok {
		if nested, ok := ctx.Value(nestedContextKey).(bool); ok && nested {
			return nil
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		return nil
	}
	return fmt.Errorf("failed to commit transaction")
}

func (t *TransactionBalancer) GetRunnner(ctx context.Context) Runner {
	metric.TotalPostgresRequests.WithLabelValues("GetRunner").Inc()
	if tx, ok := ctx.Value(txContextKey).(pgx.Tx); ok {
		return tx
	}
	return t.connPool
}

func (t *TransactionBalancer) Atomic(ctx context.Context, action func(context.Context) error) error {
	ctx, err := t.BeginTx(ctx)
	if err != nil {
		logger.Errorf("Atomic : %s", err)
		return err
	}
	metric.TotalPostgresRequests.WithLabelValues("Atomic").Inc()

	err = action(ctx)

	if err != nil {
		err2 := t.RollBackTx(ctx)
		if err2 != nil {
			logger.Debugf("Atomic : %v", err2)
		}
		return err
	}

	err = t.CommitTx(ctx)
	if err != nil {
		logger.Errorf("Atomic : %v", err)
	}
	metric.TotalPostgresRequests.WithLabelValues("GetRunner").Inc()
	return err
}
