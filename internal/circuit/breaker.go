package circuit

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"kalshi/pkg/metrics"
)

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

type CircuitBreaker struct {
	mu               sync.RWMutex
	state            State
	failureCount     int
	successCount     int
	failureThreshold int
	recoveryTimeout  time.Duration
	maxRequests      int
	lastFailureTime  time.Time
	backend          string
}

func NewCircuitBreaker(backend string, failureThreshold int, recoveryTimeout time.Duration, maxRequests int) *CircuitBreaker {
	if failureThreshold <= 0 || recoveryTimeout <= 0 || maxRequests <= 0 {
		panic("circuit breaker parameters must be positive")
	}

	cb := &CircuitBreaker{
		state:            StateClosed,
		failureThreshold: failureThreshold,
		recoveryTimeout:  recoveryTimeout,
		maxRequests:      maxRequests,
		backend:          backend,
	}

	// Update metrics
	metrics.CircuitBreakerState.WithLabelValues(backend).Set(float64(StateClosed))

	return cb
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if request is allowed
	switch state := cb.state; state {
	case StateClosed:
		// Allow request
	case StateOpen:
		if time.Since(cb.lastFailureTime) > cb.recoveryTimeout {
			cb.state = StateHalfOpen
			cb.successCount = 0
			metrics.CircuitBreakerState.WithLabelValues(cb.backend).Set(float64(StateHalfOpen))
		} else {
			return ErrCircuitBreakerOpen
		}
	case StateHalfOpen:
		if cb.successCount >= cb.maxRequests {
			return ErrCircuitBreakerOpen
		}
	default:
		return ErrCircuitBreakerOpen
	}

	// Execute the function
	err := fn()

	// Record the result
	if err == nil {
		cb.successCount++

		switch cb.state {
		case StateHalfOpen:
			if cb.successCount >= cb.maxRequests {
				cb.state = StateClosed
				cb.failureCount = 0
				metrics.CircuitBreakerState.WithLabelValues(cb.backend).Set(float64(StateClosed))
			}
		case StateClosed:
			cb.failureCount = 0
		}
	} else {
		cb.failureCount++
		cb.lastFailureTime = time.Now()

		// Debug logging
		fmt.Printf("Failure count: %d, threshold: %d, state: %d\n", cb.failureCount, cb.failureThreshold, cb.state)

		if cb.state == StateHalfOpen {
			// On failure in half-open, transition to open and reset failureCount
			cb.state = StateOpen
			cb.failureCount = cb.failureThreshold
			metrics.CircuitBreakerState.WithLabelValues(cb.backend).Set(float64(StateOpen))
			fmt.Printf("Transitioning to Open state from HalfOpen\n")
		} else if cb.failureCount == cb.failureThreshold {
			cb.state = StateOpen
			metrics.CircuitBreakerState.WithLabelValues(cb.backend).Set(float64(StateOpen))
			fmt.Printf("Transitioning to Open state\n")
		}
	}

	return err
}

func (cb *CircuitBreaker) GetState() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	metrics.CircuitBreakerState.WithLabelValues(cb.backend).Set(float64(StateClosed))
}

// SetState manually sets the circuit breaker state
func (cb *CircuitBreaker) SetState(state State) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = state
	cb.failureCount = 0
	cb.successCount = 0
	metrics.CircuitBreakerState.WithLabelValues(cb.backend).Set(float64(state))
}

var ErrCircuitBreakerOpen = errors.New("circuit breaker is open")
