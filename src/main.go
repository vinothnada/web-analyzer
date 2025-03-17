package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/vinothnada/web-analyzer/internal/config"
	"github.com/vinothnada/web-analyzer/internal/http/handlers/analyzer"
)

var requestCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests received",
	},
	[]string{"path", "method"},
)

func init() {
	prometheus.MustRegister(requestCounter)
}

func requestMetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCounter.WithLabelValues(r.URL.Path, r.Method).Inc()
		next.ServeHTTP(w, r)
	})
}

func main() {
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	cfg := config.MustLoad()
	logger.WithFields(logrus.Fields{
		"address": cfg.Addr,
	}).Info("Configuration loaded")

	// Initialize the router
	router := http.NewServeMux()
	logger.Debug("Router initialized")

	// Register handlers
	router.HandleFunc("/api/analyze", analyzer.GetResults)
	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/debug/pprof/", http.DefaultServeMux.ServeHTTP) // Enable pprof

	logger.Debug("Handlers registered")

	// Wrap router with metrics middleware
	server := &http.Server{
		Addr:           cfg.Addr,
		Handler:        requestMetricsMiddleware(router),
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
		if err := server.ListenAndServe(); err != nil {
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
