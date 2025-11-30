package observer

import (
	"fmt"
	"strings"
)

// MultiObserverError aggregates all errors from observers
type MultiObserverError struct {
	Errors []error
}

func (m MultiObserverError) Error() string {
	errs := make([]string, 0, len(m.Errors))

	for _, err := range m.Errors {
		errs = append(errs, err.Error())
	}

	return fmt.Sprintf("errors from observers: %v", strings.Join(errs, "; "))
}
