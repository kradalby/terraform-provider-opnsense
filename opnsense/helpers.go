package opnsense

import (
	"errors"
)

var (
	ErrExpectedString          = errors.New("expected string")
	ErrInvalidUUID             = errors.New("invalid UUID")
	ErrMoreThanOneUUIDReturned = errors.New("more than one uuid returned")
	ErrStatusNotOk             = errors.New("api status message not ok")
)

const apiInternalErrorMsg = "Internal Error status code received"
