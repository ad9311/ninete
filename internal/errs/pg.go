package errs

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Postgres Connection errors
var (
	ErrPgConnError                = errors.New("database error")
	ErrUniqueConstraintViolation  = errors.New("unique constraint violation")
	ErrForeignConstraintViolation = errors.New("foreign constraint violation")
)

// HandlePgError analyzes a PostgreSQL-related error and returns a more specific error
// based on the error code. If the error is not a PostgreSQL error, it returns the original error unchanged.
func HandlePgError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	pgErr, ok := IsPgError(err)
	if !ok {
		return err
	}

	switch pgErr.Code {
	case "23505":
		return fmt.Errorf("%w for %s", ErrUniqueConstraintViolation, pgErr.ColumnName)
	case "23503":
		return fmt.Errorf("%w for %s", ErrForeignConstraintViolation, pgErr.ColumnName)
	default:
		return WrapMessageWithError(ErrPgConnError, err.Error())
	}
}

// IsPgError checks whether an error is of type *pgconn.PgError.
func IsPgError(err error) (*pgconn.PgError, bool) {
	pgErr, ok := err.(*pgconn.PgError)

	return pgErr, ok
}
