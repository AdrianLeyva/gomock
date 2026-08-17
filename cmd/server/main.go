// Command server runs the gomock generic mock entity HTTP API.
package main

import (
	"log"
	"net/http"
	"time"

	"gomock/internal/api"
	"gomock/internal/config"
	"gomock/internal/store"
)

func main() {
	cfg := config.Load()

	initial, err := store.LoadDir(cfg.DataDir)
	if err != nil {
		log.Fatalf("failed to load entity data from %q: %v", cfg.DataDir, err)
	}

	s := store.New(initial)
	handlers := api.NewHandlers(s)
	limiter := api.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	router := api.NewRouter(handlers, limiter)

	// Explicit timeouts and header cap bound slow-client (Slowloris) and
	// oversized-header attacks that the default ListenAndServe leaves open.
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	log.Printf("gomock API listening on %s (data dir: %s)", srv.Addr, cfg.DataDir)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
