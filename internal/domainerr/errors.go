package domainerr

import "errors"

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type ValidationError struct{ Message string }

func (e ValidationError) Error() string { return e.Message }
