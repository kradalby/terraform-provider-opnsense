package opnsense

import (
	"errors"

	"github.com/hashicorp/terraform/helper/schema"
	uuid "github.com/satori/go.uuid"
)

var ErrExpectedString = errors.New("expected string")
var ErrInvalidUUID = errors.New("invalid UUID")
var ErrMoreThanOneUUIDReturned = errors.New("more than one uuid returned")

const apiInternalErrorMsg = "Internal Error status code received"

func ValidateUUID() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, ErrExpectedString)
			return
		}

		_, err := uuid.FromString(v)
		if err != nil {
			es = append(es, ErrInvalidUUID)
		}

		return
	}
}
