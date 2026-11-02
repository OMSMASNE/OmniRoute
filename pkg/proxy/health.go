package proxy

import (
	"log"
	"net/http"
	"time"
)

// StartHealthChecks spawns a background ticker to monitor backend status
func StartHealthChecks(pool *BackendPool, interval time.Duration) {
	ticker := time.NewTicker(interval)

	// Execute health checks inside a dedicated background goroutine.
	go func() {
		for range ticker.C {
			client := http.Client{
				Timeout: 2 * time.Second, // Don;t let tdead backends hang the checker.
			}

			for _, backend := range pool.GetBackends() {
				healthURL := backend.URL.String() + "/health"

				resp, err := client.Get(healthURL)

				// Evaluate status.
				alive := err == nil && resp.StatusCode == http.StatusOK
				if resp != nil && resp.Body != nil {
					resp.Body.Close()
				}

				// Update backend state.
				backend.SetAlive(alive)

				if !alive {
					log.Printf("[Health Check] WARNING: Backend %s is DOWN", backend.URL.String())
				}
			}
		}
	}()
}
