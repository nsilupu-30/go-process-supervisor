package restart

import (
	"context"
	"math"
	"time"

	"github.com/nsilupu-30/go-process-supervisor/internal/config"
)

// CalculateDelay calcula el tiempo de espera para el intento especificado
// aplicando la fórmula de retardo exponencial con un límite máximo (MaxSeconds).
func CalculateDelay(attempt int, cfg config.BackoffConfig) time.Duration {
	if attempt <= 1 {
		initial := time.Duration(cfg.InitialSeconds) * time.Second
		return applyMaxCap(initial, cfg.MaxSeconds)
	}

	factor := cfg.Factor
	if factor < 1.0 {
		factor = 1.0
	}

	// fórmula: espera = initial * (factor ^ (attempt - 1))
	delaySec := float64(cfg.InitialSeconds) * math.Pow(factor, float64(attempt-1))

	maxSec := float64(cfg.MaxSeconds)
	if maxSec > 0 && delaySec > maxSec {
		delaySec = maxSec
	}

	return time.Duration(delaySec * float64(time.Second))
}

func applyMaxCap(delay time.Duration, maxSeconds int) time.Duration {
	if maxSeconds > 0 {
		maxDur := time.Duration(maxSeconds) * time.Second
		if delay > maxDur {
			return maxDur
		}
	}
	return delay
}

// Wait realiza la espera del tiempo delay.
// Si el context es cancelado durante la espera, se interrumpe de inmediato y retorna ctx.Err().
func Wait(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
