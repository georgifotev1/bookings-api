package repository

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrDuplicate       = errors.New("already exists")
	ErrForeignKey      = errors.New("referenced record does not exist")
	ErrNullViolation   = errors.New("required field is missing")
	ErrCheckViolation  = errors.New("value violates check constraint")
	ErrSerialization   = errors.New("transaction conflict, please retry")
	ErrDeadlock        = errors.New("deadlock detected, please retry")
	ErrInvalidInput    = errors.New("invalid input value")
	ErrValueTooLong    = errors.New("value exceeds maximum length")
	ErrUndefinedColumn = errors.New("undefined column — possible schema mismatch")
	ErrUndefinedTable  = errors.New("undefined table — possible schema mismatch")
)

func MapError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrDuplicate
		case "23503":
			return ErrForeignKey
		case "23502":
			return ErrNullViolation
		case "23514":
			return ErrCheckViolation
		// Transactions
		case "40001":
			return ErrSerialization
		case "40P01":
			return ErrDeadlock
		// Data errors
		case "22001":
			return ErrValueTooLong
		case "22003", "22P02":
			return ErrInvalidInput
		case "42703":
			return ErrUndefinedColumn
		case "42P01":
			return ErrUndefinedTable
		}
	}
	return err
}
