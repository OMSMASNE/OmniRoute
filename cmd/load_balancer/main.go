package main

import (
	"log"
	"net/http"
	"omniRoute/pkg/proxy"
	"omniRoute/pkg/ratelimit"
	"time"
)

func main() {
	proxyPort := "8080"

	// Define all internal Docker network endpoints of our target cluster
	backendList := []string{
		"http://backend-1:8081",
		"http://backend-2:8081",
		"http://backend-3:8081",
	}

	// Highly optimized connection pooling configurations
	customTransport := &http.Transport{
		// MaxIdleConns controls the maximum number of idle (keep-alive) connections across all hosts.
		MaxIdleConns: 10000,
		// MaxIdleConnsPerHost is the critical setting. It controls how many idle connections
		// we keep open specifically to each backend target.
		MaxIdleConnsPerHost: 2000,
		// IdleConnTimeout defines how long an idle connection stays in the pool before closing.
		IdleConnTimeout: 90 * time.Second,
		// MaxConnsPerHost limits the total number of connections (open + idle) per host.
		MaxConnsPerHost: 0, // 0 means no limit, letting it scale to current traffic demands
	}

	// Initialize our concurrent-safe backend selector pool
	pool, err := proxy.NewBackendPool(backendList, customTransport)
	if err != nil {
		log.Fatalf("Initializtion error for backend pool: %v", err)
	}

	// Launch active health checking every 5 seconds.
	proxy.StartHealthChecks(pool, 5*time.Second)
	log.Println("Active health checking daemon started successfully.")

	// Open the floodgates for the benchmark: 60,000 burst capacity, 50,000 refill rate.
	limiter := ratelimit.NewRateLimiter(60000.0, 50000.0)

	coreHandler := func(w http.ResponseWriter, r *http.Request) {
		peer := pool.GetNextPeer()
		if peer == nil {
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}
		peer.ReverseProxy.ServeHTTP(w, r)
	}

	http.HandleFunc("/", limiter.Middleware(coreHandler))

	log.Printf("Rate-Limited Load Balancer listening on %s...", proxyPort)
	if err := http.ListenAndServe(":"+proxyPort, nil); err != nil {
		log.Fatalf("Load Balancer failed: %v", err)
	}

	// 	coreHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	if !limiter.Bucket().Allow() {
	// 		w.Header().Set("Content-Type", "application/json")
	// 		w.WriteHeader(http.StatusTooManyRequests)
	// 		w.Write([]byte(`{"error": "Too Many Requests"}`))
	// 		return
	// 	}

	// 	peer := pool.GetNextPeer()
	// 	if peer == nil {
	// 		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	// 		return
	// 	}
	// 	peer.ReverseProxy.ServeHTTP(w, r)
	// })

	// log.Printf("Optimized Load Balancer listening on %s...", proxyPort)
	// // Passing coreHandler directly here bypasses Go's default routing multiplexer overhead completely
	// if err := http.ListenAndServe(":"+proxyPort, coreHandler); err != nil {
	// 	log.Fatalf("Load Balancer down: %v", err)
	// }
}
