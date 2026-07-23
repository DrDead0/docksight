package security

// Validator will enforce agent authentication and message integrity checks.
// Intentionally empty in the foundation phase.
type Validator struct{}

// NewValidator creates a security validator placeholder.
func NewValidator() *Validator {
	return &Validator{}
}
