package controlplane

import "fmt"

// ReconciliationError represents an error that occurred during reconciliation
type ReconciliationError struct {
	message string
	cause   error
}

// NewReconciliationError creates a new ReconciliationError
func NewReconciliationError(message string, cause error) error {
	return &ReconciliationError{
		message: message,
		cause:   cause,
	}
}

// Error returns the error message
func (e *ReconciliationError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %v", e.message, e.cause)
	}
	return e.message
}

// Unwrap returns the underlying error
func (e *ReconciliationError) Unwrap() error {
	return e.cause
}

// IsReconciliationError checks if the given error is a ReconciliationError
func IsReconciliationError(err error) bool {
	_, ok := err.(*ReconciliationError)
	return ok
}
