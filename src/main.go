package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vinothnada/web-analyzer/internal/config"
	"github.com/vinothnada/web-analyzer/internal/http/handlers/analyzer"
)

func main() {

	cfg := config.MustLoad()

	router := http.NewServeMux()

	router.HandleFunc("GET /api/analyze", analyzer.GetResults())

	server := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	slog.Info("Server started", slog.String("address", cfg.Addr))

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatal("Failed to start the server")
		}
	}()

	<-done

	slog.Info("Shutting down the server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown the server", slog.String("error", err.Error()))
	}

	slog.Info("Server shutdown successfully")

}
