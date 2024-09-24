package repository

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

var ErrAlreadyExists = errors.New("already exists")
var ErrNotFound = errors.New("not found")

func convertErr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.UniqueViolation == pgErr.Code {
		return ErrAlreadyExists
	}
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	return err
}
