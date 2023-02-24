package template_repo

import (
	"context"
	"errors"
	"reflect"

	"github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	localerrors "gitlab.com/Sh00ty/dormitory_room_bot/internal/local_errors"
	"gitlab.com/Sh00ty/dormitory_room_bot/pkg/pgxbalancer"
)

func getDtoInfo(dto interface{}) ([]string, []interface{}) {
	pointerVal := reflect.ValueOf(dto)
	val := reflect.Indirect(pointerVal)
	typ := val.Type()
	numFeilds := val.NumField()
	columns := make([]string, 0, numFeilds)
	values := make([]interface{}, 0, numFeilds)
	for i := 0; i < numFeilds; i++ {
		values = append(values, val.Field(i).Interface())
		columns = append(columns, typ.Field(i).Tag.Get("db"))
	}
	return columns, values
}

func getDtoColumns(dto interface{}) []string {
	pointerVal := reflect.ValueOf(dto)
	val := reflect.Indirect(pointerVal)
	typ := val.Type()
	numFeilds := val.NumField()
	columns := make([]string, 0, numFeilds)

	for i := 0; i < numFeilds; i++ {
		columns = append(columns, typ.Field(i).Tag.Get("db"))
	}
	return columns
}

type transactionBalancer interface {
	GetRunnner(ctx context.Context) pgxbalancer.Runner
}

type repo interface {
	transactionBalancer
	TableName() string
}

func Create[T any](ctx context.Context, r repo, value T) error {
	columns, values := getDtoInfo(value)

	sql, values, err := squirrel.Insert(r.TableName()).Columns(columns...).
		Values(values...).
		PlaceholderFormat(squirrel.Dollar).
		ToSql()

	if err != nil {
		return err
	}

	ct, err := r.GetRunnner(ctx).Exec(ctx, sql, values...)
	if ct.RowsAffected() == 0 {
		return localerrors.ErrAlreadyExists
	}
	return err
}

func GetBy[T any](ctx context.Context, r repo, conj squirrel.Sqlizer) (T, error) {
	var t T
	sql, values, err := squirrel.Select(getDtoColumns(t)...).
		From(r.TableName()).
		Where(conj).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return t, err
	}

	rows, err := r.GetRunnner(ctx).Query(ctx, sql, values...)
	if err != nil {
		return t, err
	}

	var dto T
	err = pgxscan.ScanOne(&dto, rows)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return t, localerrors.ErrDoesntExist
		}
		return t, err
	}

	return dto, nil
}

func GetAllBy[T any](ctx context.Context, r repo, conj squirrel.Sqlizer) (result []T, err error) {
	var t T
	sql, values, err := squirrel.Select(getDtoColumns(t)...).
		From(r.TableName()).
		Where(conj).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.GetRunnner(ctx).Query(ctx, sql, values...)
	if err != nil {
		return nil, err
	}

	var dtoList []T
	err = pgxscan.ScanAll(&dtoList, rows)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, localerrors.ErrDoesntExist
		}
		return nil, err
	}
	return dtoList, nil
}

func Delete[T any](ctx context.Context, r repo, conj squirrel.Sqlizer) error {
	sql, values, err := squirrel.Delete(r.TableName()).
		Where(conj).
		PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return localerrors.ErrDoesntExist
		}
		return err
	}

	commandTag, err := r.GetRunnner(ctx).Exec(ctx, sql, values...)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return localerrors.ErrDoesntExist
	}
	return nil
}
