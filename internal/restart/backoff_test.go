package restart

import (
	"context"
	"testing"
	"time"

	"github.com/nsilupu-30/go-process-supervisor/internal/config"
)

func TestShouldRestart(t *testing.T) {
	tests := []struct {
		name           string
		policy         config.RestartPolicy
		exitCode       int
		currentRetries int
		maxRetries     int
		expected       bool
	}{
		{
			name:           "Policy Never -> Always false",
			policy:         config.RestartNever,
			exitCode:       1,
			currentRetries: 0,
			maxRetries:     5,
			expected:       false,
		},
		{
			name:           "Policy Always with exit code 0 -> true",
			policy:         config.RestartAlways,
			exitCode:       0,
			currentRetries: 0,
			maxRetries:     5,
			expected:       true,
		},
		{
			name:           "Policy Always with exit code 1 -> true",
			policy:         config.RestartAlways,
			exitCode:       1,
			currentRetries: 0,
			maxRetries:     5,
			expected:       true,
		},
		{
			name:           "Policy OnFailure with exit code 0 -> false",
			policy:         config.RestartOnFailure,
			exitCode:       0,
			currentRetries: 0,
			maxRetries:     5,
			expected:       false,
		},
		{
			name:           "Policy OnFailure with exit code 1 -> true",
			policy:         config.RestartOnFailure,
			exitCode:       1,
			currentRetries: 0,
			maxRetries:     5,
			expected:       true,
		},
		{
			name:           "Exceeded max retries -> false",
			policy:         config.RestartAlways,
			exitCode:       1,
			currentRetries: 5,
			maxRetries:     5,
			expected:       false,
		},
		{
			name:           "Unlimited retries (maxRetries=0) -> true",
			policy:         config.RestartAlways,
			exitCode:       1,
			currentRetries: 100,
			maxRetries:     0,
			expected:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldRestart(tt.policy, tt.exitCode, tt.currentRetries, tt.maxRetries)
			if got != tt.expected {
				t.Errorf("ShouldRestart() = %v, se esperaba %v", got, tt.expected)
			}
		})
	}
}

func TestCalculateDelay(t *testing.T) {
	cfg := config.BackoffConfig{
		InitialSeconds: 1,
		Factor:         2.0,
		MaxSeconds:     10,
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{attempt: 1, expected: 1 * time.Second},  // 1 * (2.0^0) = 1s
		{attempt: 2, expected: 2 * time.Second},  // 1 * (2.0^1) = 2s
		{attempt: 3, expected: 4 * time.Second},  // 1 * (2.0^2) = 4s
		{attempt: 4, expected: 8 * time.Second},  // 1 * (2.0^3) = 8s
		{attempt: 5, expected: 10 * time.Second}, // 1 * (2.0^4) = 16s -> tope maxSeconds = 10s
		{attempt: 6, expected: 10 * time.Second}, // tope maxSeconds = 10s
	}

	for _, tt := range tests {
		got := CalculateDelay(tt.attempt, cfg)
		if got != tt.expected {
			t.Errorf("CalculateDelay(attempt=%d) = %v, se esperaba %v", tt.attempt, got, tt.expected)
		}
	}
}

func TestWait_Completion(t *testing.T) {
	ctx := context.Background()
	start := time.Now()
	err := Wait(ctx, 50*time.Millisecond)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Wait() retorno error inesperado: %v", err)
	}

	if duration < 40*time.Millisecond {
		t.Errorf("Wait() retorno antes de tiempo, duracion: %v", duration)
	}
}

func TestWait_Cancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	start := time.Now()
	err := Wait(ctx, 10*time.Second) // Intento de esperar 10s, pero cancelado a los 50ms
	duration := time.Since(start)

	if err == nil {
		t.Fatalf("Wait() debio retornar error por contexto cancelado")
	}

	if duration > 1*time.Second {
		t.Errorf("Wait() no respondio a la cancelacion del contexto a tiempo, tardo %v", duration)
	}
}
