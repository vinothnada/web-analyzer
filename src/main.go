package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/vinothnada/web-analyzer/internal/config"
	"github.com/vinothnada/web-analyzer/internal/http/handlers/analyzer"
)

func main() {

	// Initialize logrus logger
	logger := logrus.New()

	// Configure logrus (optional: set log format, log level, etc.)
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg := config.MustLoad()
	logger.WithFields(logrus.Fields{
		"address": cfg.Addr,
	}).Info("Configuration loaded")

	// Initialize the router
	router := http.NewServeMux()
	logger.Debug("Router initialized")

	// Register the handler
	router.HandleFunc("/api/analyze", analyzer.GetResults)
	logger.Debug("Handler for /api/analyze registered")

	// Create the server
	server := &http.Server{
		Addr:           cfg.Addr,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// Start the server
	logger.WithFields(logrus.Fields{
		"address": cfg.Addr,
	}).Info("Starting server")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	logger.Debug("Signal notification registered")

	go func() {
		logger.Info("Server is listening for incoming requests...")
		err := server.ListenAndServe()
		if err != nil {
			logger.WithFields(logrus.Fields{
				"error": err.Error(),
			}).Error("Failed to start the server")
		}
	}()

	<-done
	logger.Info("Received shutdown signal, stopping the server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("Failed to shutdown the server")
	} else {
		logger.Info("Server shutdown successfully")
	}
}
