package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"memorialcandle/internal/config"
	"memorialcandle/internal/memorial"
	"memorialcandle/internal/protocol"
	"memorialcandle/internal/store"
)

func main() {
	cfg := config.FromEnv()
	if !cfg.Valid() {
		log.Fatal("invalid configuration")
	}
	db, err := store.Open(cfg.Database)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	service := memorial.NewService(db, cfg.Seed)
	server := &http.Server{Addr: cfg.Addr, Handler: protocol.NewServer(service, log.Default()).Routes()}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()
	log.Printf("memorial candle service listening on %s", cfg.Addr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
