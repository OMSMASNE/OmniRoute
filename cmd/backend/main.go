package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Dynamically look for a PORT env variable, defaulting to 8081 if not found
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Fetch backend identifier name to distinguish instances in logs.
	serverName := os.Getenv("SERVER_NAME")
	if serverName == "" {
		serverName = "Unknown_backend"
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader((http.StatusOK))
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Log the incoming request locally to see the load balancer in action later.
		log.Printf("[%s] Recieved request: %s %s", serverName, r.Method, r.URL.Path)
		w.Header().Set("Content-type", "application/json")
		fmt.Fprintf(w, `{"message": "Hello from %s}`, serverName)
	})

	log.Printf("Starting dummy backend server [%s] on the port %s.", serverName, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start %v", err)
	}
}
