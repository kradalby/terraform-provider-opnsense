package opnsense

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	uuid "github.com/satori/go.uuid"
)

func ValidateUUID() schema.SchemaValidateFunc {
	return func(i interface{}, k string) (s []string, es []error) {
		v, ok := i.(string)
		if !ok {
			es = append(es, fmt.Errorf("expected type of %s to be string", k))
			return
		}

		_, err := uuid.FromString(v)
		if err != nil {
			es = append(es, fmt.Errorf(
				"expected %s to contain a valid UUID, got: %s", k, v))
		}

		return
	}
}
