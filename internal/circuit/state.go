package circuit

import (
	"fmt"
	"sync"
	"time"
)

type Manager struct {
	breakers map[string]*CircuitBreaker
	mu       sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

func (m *Manager) GetBreaker(backend string, failureThreshold int, recoveryTimeout time.Duration, maxRequests int) *CircuitBreaker {
	m.mu.RLock()
	breaker, exists := m.breakers[backend]
	m.mu.RUnlock()

	if exists {
		return breaker
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double Check Pattern
	if breaker, exists := m.breakers[backend]; exists {
		return breaker
	}

	breaker = NewCircuitBreaker(backend, failureThreshold, recoveryTimeout, maxRequests)
	m.breakers[backend] = breaker
	return breaker
}

func (m *Manager) GetAllStates() map[string]State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := make(map[string]State)
	for name, breaker := range m.breakers {
		states[name] = breaker.GetState()
	}
	return states
}

func (m *Manager) ResetBreaker(backend string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	breaker, exists := m.breakers[backend]
	if !exists {
		return fmt.Errorf("circuit breaker for backend %s not found", backend)
	}

	breaker.Reset()
	return nil
}
