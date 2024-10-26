package main

import (
	"net/http"

	"github.com/shekshuev/shortener/internal/app/config"
	"github.com/shekshuev/shortener/internal/app/handler"
	"github.com/shekshuev/shortener/internal/app/logger"
	"github.com/shekshuev/shortener/internal/app/service"
	"github.com/shekshuev/shortener/internal/app/store"

	"go.uber.org/zap"
)

func main() {
	l := logger.GetInstance()
	cfg := config.GetConfig()
	urlStore := store.NewURLStore(&cfg)
	urlService := service.NewURLService(urlStore, &cfg)
	urlHandler := handler.NewURLHandler(urlService)
	if err := http.ListenAndServe(cfg.ServerAddress, urlHandler.Router); err != nil {
		l.Log.Error("Error starting server", zap.Error(err))
	}
}
