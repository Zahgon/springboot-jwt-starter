// Package clock provides the injectable time source the rest of the
// application reads instead of calling time.Now directly, so tests can script
// the exact sequence of instants a code path observes.
package clock

import "time"

// Provider yields the current instant.
type Provider interface {
	Now() time.Time
}

// System is the production Provider, backed by the wall clock.
type System struct{}

// Now returns the current wall-clock time.
func (System) Now() time.Time { return time.Now() }
