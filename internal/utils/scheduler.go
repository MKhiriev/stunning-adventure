package utils

import "time"

func RunWithTicker(fn func(), duration time.Duration) {
	ticker := time.NewTicker(duration)

	go func() {
		for range ticker.C {
			fn()
		}
	}()
}
