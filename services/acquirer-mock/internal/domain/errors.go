package domain

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrInvalidState        = errors.New("invalid state")
	ErrIdempotencyConflict = errors.New("idempotency conflict")
	ErrAmountExceeded      = errors.New("amount exceeded")
	ErrAmountMismatch      = errors.New("amount mismatch")
)
