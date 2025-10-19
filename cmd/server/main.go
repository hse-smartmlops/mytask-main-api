package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"emplacc-api/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx); err != nil {
		log.Fatalf("run application: %v", err)
	}
}
