package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"

	"github.com/relentlessworks/jsonkit/internal/api"
	"github.com/relentlessworks/jsonkit/internal/config"
)

func main() {
	cfg := config.Load()

	// Auto-generate secret if not provided
	if cfg.Secret == "" {
		b := make([]byte, 32)
		rand.Read(b)
		cfg.Secret = hex.EncodeToString(b)
		log.Println("no secret provided, generated random secret")
	}

	srv := api.New()
	addr := cfg.ListenAddr()
	log.Printf("jsonkit listening on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		fmt.Fprintf(log.Writer(), "server error: %v\n", err)
	}
}
