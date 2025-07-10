package circuit

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	if m.breakers == nil {
		t.Error("breakers map should be initialized")
	}
}

func TestManager_GetBreaker(t *testing.T) {
	m := NewManager()

	// Test creating a new breaker
	breaker1 := m.GetBreaker("backend1", 3, time.Second, 2)
	if breaker1 == nil {
		t.Fatal("GetBreaker() returned nil")
	}
	if breaker1.backend != "backend1" {
		t.Errorf("expected backend %s, got %s", "backend1", breaker1.backend)
	}

	// Test getting the same breaker again
	breaker2 := m.GetBreaker("backend1", 5, time.Second*2, 3)
	if breaker2 != breaker1 {
		t.Error("GetBreaker() should return the same breaker instance")
	}

	// Test creating a different breaker
	breaker3 := m.GetBreaker("backend2", 4, time.Second, 1)
	if breaker3 == breaker1 {
		t.Error("GetBreaker() should return different breaker instances for different backends")
	}
	if breaker3.backend != "backend2" {
		t.Errorf("expected backend %s, got %s", "backend2", breaker3.backend)
	}
}

func TestManager_GetAllStates(t *testing.T) {
	m := NewManager()

	// Create multiple breakers
	breaker1 := m.GetBreaker("backend1", 3, time.Second, 2)
	breaker2 := m.GetBreaker("backend2", 4, time.Second, 1)

	// Test initial states
	states := m.GetAllStates()
	if len(states) != 2 {
		t.Errorf("expected 2 states, got %d", len(states))
	}
	if states["backend1"] != StateClosed {
		t.Errorf("expected backend1 state %d, got %d", StateClosed, states["backend1"])
	}
	if states["backend2"] != StateClosed {
		t.Errorf("expected backend2 state %d, got %d", StateClosed, states["backend2"])
	}

	// Change state of one breaker
	breaker1.Call(func() error { return nil })
	breaker2.Call(func() error { return nil })
	states = m.GetAllStates()
	if states["backend1"] != StateClosed {
		t.Errorf("expected backend1 state %d, got %d", StateClosed, states["backend1"])
	}
	if states["backend2"] != StateClosed {
		t.Errorf("expected backend2 state %d, got %d", StateClosed, states["backend2"])
	}
}

func TestManager_ResetBreaker(t *testing.T) {
	m := NewManager()

	// Test resetting non-existent breaker
	err := m.ResetBreaker("non-existent")
	if err == nil {
		t.Error("expected error when resetting non-existent breaker")
	}

	// Create and test existing breaker
	breaker := m.GetBreaker("backend1", 2, time.Second, 1)

	// Fail to open circuit
	breaker.Call(func() error { return errors.New("test error") })
	breaker.Call(func() error { return errors.New("test error") })

	if breaker.GetState() != StateOpen {
		t.Errorf("expected state %d, got %d", StateOpen, breaker.GetState())
	}

	// Reset breaker
	err = m.ResetBreaker("backend1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if breaker.GetState() != StateClosed {
		t.Errorf("expected state %d after reset, got %d", StateClosed, breaker.GetState())
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	m := NewManager()

	// Test concurrent breaker creation
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			backend := fmt.Sprintf("backend%d", id)
			breaker := m.GetBreaker(backend, 3, time.Second, 2)
			if breaker.backend != backend {
				t.Errorf("expected backend %s, got %s", backend, breaker.backend)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all breakers were created
	states := m.GetAllStates()
	if len(states) != 10 {
		t.Errorf("expected 10 states, got %d", len(states))
	}
}

func TestManager_GetBreakerWithDifferentParams(t *testing.T) {
	m := NewManager()

	// Create breaker with initial params
	breaker1 := m.GetBreaker("backend1", 3, time.Second, 2)

	// Try to get same breaker with different params
	breaker2 := m.GetBreaker("backend1", 5, time.Second*2, 3)

	// Should return the same instance with original params
	if breaker2 != breaker1 {
		t.Error("GetBreaker() should return the same instance regardless of params")
	}

	// Verify original params are preserved
	if breaker2.failureThreshold != 3 {
		t.Errorf("expected failure threshold 3, got %d", breaker2.failureThreshold)
	}
	if breaker2.recoveryTimeout != time.Second {
		t.Errorf("expected recovery timeout %v, got %v", time.Second, breaker2.recoveryTimeout)
	}
	if breaker2.maxRequests != 2 {
		t.Errorf("expected max requests 2, got %d", breaker2.maxRequests)
	}
}
