package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shekshuev/shortener/internal/app/config"
	"github.com/shekshuev/shortener/internal/app/handler"
	"github.com/shekshuev/shortener/internal/app/logger"
	"github.com/shekshuev/shortener/internal/app/service"
	"github.com/shekshuev/shortener/internal/app/store"

	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"

	_ "net/http/pprof"
)

func main() {
	l := logger.NewLogger()
	cfg := config.GetConfig()
	var urlStore store.URLStore = nil
	if cfg.DatabaseDSN == cfg.DefaultDatabaseDSN {
		urlStore = store.NewMemoryURLStore(&cfg)
	} else {
		urlStore = store.NewPostgresURLStore(&cfg)
	}
	urlService := service.NewURLService(urlStore, &cfg)
	urlHandler := handler.NewURLHandler(urlService)
	server := &http.Server{
		Addr:    cfg.ServerAddress,
		Handler: urlHandler.Router,
	}
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Log.Error("Error starting server", zap.Error(err))
		}
	}()
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	l.Log.Info("Server started")
	<-done
	l.Log.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := urlStore.Close(); err != nil {
		l.Log.Error("Error closing store", zap.Error(err))
	} else {
		l.Log.Info("Store closed")
	}
	if err := server.Shutdown(ctx); err != nil {
		l.Log.Error("Server forced to shutdown", zap.Error(err))
	} else {
		l.Log.Info("Server shutdown gracefully")
	}
}
