package circuit

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestDebugCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker("test", 2, time.Millisecond*100, 1)

	fmt.Printf("Initial state: %d\n", cb.GetState())
	fmt.Printf("Initial failure count: %d\n", cb.failureCount)

	// First failure
	err := cb.Call(func() error { return errors.New("test error") })
	fmt.Printf("After first failure - state: %d, failure count: %d, err: %v\n", cb.GetState(), cb.failureCount, err)

	// Second failure
	err = cb.Call(func() error { return errors.New("test error") })
	fmt.Printf("After second failure - state: %d, failure count: %d, err: %v\n", cb.GetState(), cb.failureCount, err)

	// Check state
	state := cb.GetState()
	fmt.Printf("Final state: %d (expected: %d)\n", state, StateOpen)
}
