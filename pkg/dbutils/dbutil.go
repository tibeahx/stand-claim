package dbutils

import (
	"database/sql"

	"reflect"

	"github.com/cockroachdb/errors"
	"github.com/jmoiron/sqlx"
)

func NamedSelect[T any](
	db *sqlx.DB,
	query string,
	dest *[]T,
	args map[string]any,
) error {
	if dest == nil {
		return errors.New("dest is nil")
	}
	rows, err := db.NamedQuery(query, args)
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}
	defer rows.Close()

	for rows.Next() {
		var val T
		if err := scanRow(rows, &val); err != nil {
			return errors.Wrap(err, "failed to scan row")
		}
		*dest = append(*dest, val)
	}

	return nil
}

func NamedGet[T any](
	db *sqlx.DB,
	query string,
	dest *T,
	args map[string]any,
) error {
	if dest == nil {
		return errors.New("dest is nil")
	}
	rows, err := db.NamedQuery(query, args)
	if err != nil {
		return errors.Wrap(err, "failed to execute query")
	}
	defer rows.Close()

	if !rows.Next() {
		return sql.ErrNoRows
	}

	if err := scanRow(rows, dest); err != nil {
		return errors.Wrap(err, "failed to scan row")
	}

	return nil
}

func NamedExec(
	db *sqlx.DB,
	query string,
	args map[string]any,
) error {
	_, err := db.NamedExec(query, args)
	return errors.Wrap(err, "failed to execute query")
}

func scanRow[T any](rows *sqlx.Rows, dest *T) error {
	if reflect.TypeOf(*dest).Kind() == reflect.Struct {
		return rows.StructScan(dest)
	}

	return rows.Scan(dest)
}
