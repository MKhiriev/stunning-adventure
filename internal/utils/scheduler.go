package utils

import "time"

// RunWithTicker executes the provided function periodically using a time.Ticker.
// The function runs asynchronously in a separate goroutine.
//
// Behavior:
//   - Creates a ticker with the given duration
//   - Launches a goroutine that runs `fn()` on every tick
//   - Goroutine lives until program termination; ticker is never stopped
//
// Parameters:
//
//	fn       - function to invoke periodically
//	duration - interval at which `fn` should run
func RunWithTicker(fn func(), duration time.Duration) {
	ticker := time.NewTicker(duration)

	go func() {
		for range ticker.C {
			fn()
		}
	}()
}
