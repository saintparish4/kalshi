package circuit

import (
	"errors"
	"testing"
	"time"
)

func TestNewCircuitBreaker(t *testing.T) {
	tests := []struct {
		name             string
		backend          string
		failureThreshold int
		recoveryTimeout  time.Duration
		maxRequests      int
		shouldPanic      bool
	}{
		{
			name:             "valid parameters",
			backend:          "test-backend",
			failureThreshold: 3,
			recoveryTimeout:  time.Second,
			maxRequests:      2,
			shouldPanic:      false,
		},
		{
			name:             "zero failure threshold",
			backend:          "test-backend",
			failureThreshold: 0,
			recoveryTimeout:  time.Second,
			maxRequests:      2,
			shouldPanic:      true,
		},
		{
			name:             "zero recovery timeout",
			backend:          "test-backend",
			failureThreshold: 3,
			recoveryTimeout:  0,
			maxRequests:      2,
			shouldPanic:      true,
		},
		{
			name:             "zero max requests",
			backend:          "test-backend",
			failureThreshold: 3,
			recoveryTimeout:  time.Second,
			maxRequests:      0,
			shouldPanic:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil && !tt.shouldPanic {
					t.Errorf("NewCircuitBreaker() panicked unexpectedly: %v", r)
				} else if r == nil && tt.shouldPanic {
					t.Error("NewCircuitBreaker() should have panicked")
				}
			}()

			cb := NewCircuitBreaker(tt.backend, tt.failureThreshold, tt.recoveryTimeout, tt.maxRequests)
			if !tt.shouldPanic {
				if cb.backend != tt.backend {
					t.Errorf("expected backend %s, got %s", tt.backend, cb.backend)
				}
				if cb.failureThreshold != tt.failureThreshold {
					t.Errorf("expected failure threshold %d, got %d", tt.failureThreshold, cb.failureThreshold)
				}
				if cb.recoveryTimeout != tt.recoveryTimeout {
					t.Errorf("expected recovery timeout %v, got %v", tt.recoveryTimeout, cb.recoveryTimeout)
				}
				if cb.maxRequests != tt.maxRequests {
					t.Errorf("expected max requests %d, got %d", tt.maxRequests, cb.maxRequests)
				}
				if cb.state != StateClosed {
					t.Errorf("expected initial state %d, got %d", StateClosed, cb.state)
				}
			}
		})
	}
}

func TestCircuitBreaker_ClosedState(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, time.Second, 1)

	// Test successful calls in closed state
	err := cb.Call(func() error { return nil })
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if cb.GetState() != StateClosed {
		t.Errorf("expected state %d, got %d", StateClosed, cb.GetState())
	}
	if cb.failureCount != 0 {
		t.Errorf("expected failure count 0, got %d", cb.failureCount)
	}

	// Test failure in closed state
	testErr := errors.New("test error")
	err = cb.Call(func() error { return testErr })
	if err != testErr {
		t.Errorf("expected error %v, got %v", testErr, err)
	}
	if cb.GetState() != StateClosed {
		t.Errorf("expected state %d, got %d", StateClosed, cb.GetState())
	}
	if cb.failureCount != 1 {
		t.Errorf("expected failure count 1, got %d", cb.failureCount)
	}
}

func TestCircuitBreaker_OpenState(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, time.Millisecond*100, 1)

	// Fail enough times to open circuit
	var err error
	for i := 0; i < 2; i++ {
		err = cb.Call(func() error { return errors.New("test error") })
		// After the last failure, the state should transition to Open
		if i == 1 {
			if cb.GetState() != StateOpen {
				t.Errorf("expected state %d (StateOpen), got %d", StateOpen, cb.GetState())
			}
		}
	}

	// Test that calls are rejected in open state
	err = cb.Call(func() error { return nil })
	if err != ErrCircuitBreakerOpen {
		t.Errorf("expected error %v, got %v", ErrCircuitBreakerOpen, err)
	}

	// Wait for recovery timeout
	time.Sleep(time.Millisecond * 150)
	// Before the call, state should still be Open
	if cb.GetState() != StateOpen {
		t.Errorf("expected state %d (StateOpen) before call, got %d", StateOpen, cb.GetState())
	}
	// The first call after timeout should transition to half-open and allow the call
	err = cb.Call(func() error { return nil })
	if err != nil {
		t.Errorf("expected no error after recovery timeout, got %v", err)
	}
	// After the call, with maxRequests=1, state should be Closed
	if cb.GetState() != StateClosed {
		t.Errorf("expected state %d (StateClosed), got %d", StateClosed, cb.GetState())
	}
}

func TestCircuitBreaker_HalfOpenState(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, time.Millisecond*50, 2)

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error { return errors.New("test error") })
	}

	// Wait for recovery timeout
	time.Sleep(time.Millisecond * 100)

	// Test successful calls in half-open state
	err := cb.Call(func() error { return nil })
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if cb.GetState() != StateHalfOpen {
		t.Errorf("expected state %d, got %d", StateHalfOpen, cb.GetState())
	}

	// Second successful call should close the circuit
	err = cb.Call(func() error { return nil })
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if cb.GetState() != StateClosed {
		t.Errorf("expected state %d, got %d", StateClosed, cb.GetState())
	}
}

func TestCircuitBreaker_HalfOpenStateFailure(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, time.Millisecond*50, 2)

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error { return errors.New("test error") })
	}

	// Wait for recovery timeout
	time.Sleep(time.Millisecond * 100)

	// The first call after timeout should transition to half-open
	err := cb.Call(func() error { return errors.New("test error") })
	if err == nil {
		t.Error("expected error, got nil")
	}
	// After a failure in half-open, state should be Open
	if cb.GetState() != StateOpen {
		t.Errorf("expected state %d (StateOpen), got %d", StateOpen, cb.GetState())
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, time.Second, 1)

	// Fail once
	cb.Call(func() error { return errors.New("test error") })

	// Reset
	cb.Reset()

	if cb.GetState() != StateClosed {
		t.Errorf("expected state %d, got %d", StateClosed, cb.GetState())
	}
	if cb.failureCount != 0 {
		t.Errorf("expected failure count 0, got %d", cb.failureCount)
	}
	if cb.successCount != 0 {
		t.Errorf("expected success count 0, got %d", cb.successCount)
	}
}

func TestCircuitBreaker_GetState(t *testing.T) {
	cb := NewCircuitBreaker("test", 1, time.Second, 1)

	// Test initial state
	if state := cb.GetState(); state != StateClosed {
		t.Errorf("expected initial state %d, got %d", StateClosed, state)
	}

	// Open circuit
	cb.Call(func() error { return errors.New("test error") })
	if state := cb.GetState(); state != StateOpen {
		t.Errorf("expected state %d, got %d", StateOpen, state)
	}
}

func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	cb := NewCircuitBreaker("test", 10, time.Millisecond*10, 5)

	// Run concurrent calls
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			cb.Call(func() error { return nil })
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify state is still consistent
	state := cb.GetState()
	if state != StateClosed && state != StateOpen && state != StateHalfOpen {
		t.Errorf("invalid state %d", state)
	}
}
