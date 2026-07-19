// Command server runs the gomock generic mock entity HTTP API.
package main

import (
	"log"
	"net/http"

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
	router := api.NewRouter(handlers)

	addr := ":" + cfg.Port
	log.Printf("gomock API listening on %s (data dir: %s)", addr, cfg.DataDir)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
