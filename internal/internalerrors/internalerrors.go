package internalerrors

import (
	"errors"
)

var (
	// ErrAlreadyExists is used when a resource already exists.
	ErrAlreadyExists = errors.New("already exists")
	// ErrNoAuth is used when a user is unauthorized.
	ErrNoAuth = errors.New("unauthorized")
	// ErrNotFound is an error that will be used when there is a resources that is missing.
	ErrNotFound = errors.New("resource not found")
	// ErrNotImplemented is an error that will be used when there is no implementation.
	ErrNotImplemented = errors.New("not implemented")
	// ErrNotValid is used when a resource is not valid.
	ErrNotValid = errors.New("not valid")
)
