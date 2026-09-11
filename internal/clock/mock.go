package clock

import (
	"sync"
	"time"
)

// Mock is a scripted time source for tests. Successive calls to Now return the
// scripted instants in order; once the script is exhausted the last instant is
// returned for every further call, which is how the original's mocked
// TimeProvider behaved.
type Mock struct {
	mu       sync.Mutex
	instants []time.Time
	calls    int
}

// NewMock returns a time source scripted with the given instants.
func NewMock(instants ...time.Time) *Mock { return &Mock{instants: instants} }

// Returns replaces the script and resets the call counter.
func (m *Mock) Returns(instants ...time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.instants = instants
	m.calls = 0
}

// Now yields the next scripted instant.
func (m *Mock) Now() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.instants) == 0 {
		return time.Time{}
	}
	i := m.calls
	if i >= len(m.instants) {
		i = len(m.instants) - 1
	}
	m.calls++
	return m.instants[i]
}
