package validators

import "context"

type Validator interface {
	Validate(context.Context, any, ...string) error
}
