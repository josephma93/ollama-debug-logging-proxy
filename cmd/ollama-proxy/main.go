package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joseph/ollama-logging-proxy/internal/config"
	"github.com/joseph/ollama-logging-proxy/internal/proxy"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	handler, err := proxy.New(proxy.Options{
		Target:       cfg.Target,
		MaxBodyBytes: cfg.MaxBodyBytes,
	})
	if err != nil {
		log.Fatalf("failed to configure proxy: %v", err)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	logger.Printf("ollama-logging-proxy listening on %s", cfg.Listen)
	logger.Printf("forwarding upstream to %s", cfg.Target.String())

	server := &http.Server{
		Addr:    cfg.Listen,
		Handler: handler,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("proxy server failed: %v", err)
	}
}
