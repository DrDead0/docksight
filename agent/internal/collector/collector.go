package collector

// Collector will gather container metrics and status from the local Docker host.
// Intentionally empty in the foundation phase.
type Collector struct{}

// NewCollector creates a metrics/status collector placeholder.
func NewCollector() *Collector {
	return &Collector{}
}
