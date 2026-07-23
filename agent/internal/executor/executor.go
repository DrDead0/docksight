package executor

// Executor will run container lifecycle commands received from the backend.
// Intentionally empty in the foundation phase.
type Executor struct{}

// NewExecutor creates a command executor placeholder.
func NewExecutor() *Executor {
	return &Executor{}
}
