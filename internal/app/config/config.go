package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
	"github.com/shekshuev/shortener/internal/app/logger"
	"go.uber.org/zap"
)

// Config содержит настройки приложения, включая параметры сервера, базы данных и файлового хранилища.
type Config struct {
	ServerAddress          string // Адрес и порт, на котором запускается сервер.
	BaseURL                string // Базовый URL для сокращённых ссылок.
	FileStoragePath        string // Путь к файлу для хранения сокращённых URL.
	DatabaseDSN            string // Строка подключения к базе данных.
	DefaultServerAddress   string // Значение по умолчанию для ServerAddress.
	DefaultBaseURL         string // Значение по умолчанию для BaseURL.
	DefaultFileStoragePath string // Значение по умолчанию для FileStoragePath.
	DefaultDatabaseDSN     string // Значение по умолчанию для DatabaseDSN.
}

type envConfig struct {
	ServerAddress   string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func GetConfig() Config {
	var cfg Config
	cfg.DefaultServerAddress = "localhost:8080"
	cfg.DefaultBaseURL = "http://localhost:8080"
	cfg.DefaultFileStoragePath = "./storage.txt"
	cfg.DefaultDatabaseDSN = ""
	parseFlags(&cfg)
	parsEnv(&cfg)
	return cfg
}

func parseFlags(cfg *Config) {
	if f := flag.Lookup("a"); f == nil {
		flag.StringVar(&cfg.ServerAddress, "a", cfg.DefaultServerAddress, "address and port to run server")
	} else {
		cfg.ServerAddress = cfg.DefaultServerAddress
	}
	if f := flag.Lookup("b"); f == nil {
		flag.StringVar(&cfg.BaseURL, "b", cfg.DefaultBaseURL, "base URL of shorter address")
	} else {
		cfg.BaseURL = cfg.DefaultBaseURL
	}
	if f := flag.Lookup("f"); f == nil {
		flag.StringVar(&cfg.FileStoragePath, "f", cfg.DefaultFileStoragePath, "file storage path")
	} else {
		cfg.FileStoragePath = cfg.DefaultFileStoragePath
	}
	if f := flag.Lookup("d"); f == nil {
		flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DefaultDatabaseDSN, "database connection string")
	} else {
		cfg.DatabaseDSN = cfg.DefaultDatabaseDSN
	}
	flag.Parse()
	parsEnv(cfg)
}

func parsEnv(cfg *Config) {
	l := logger.NewLogger()
	var envCfg envConfig
	err := env.Parse(&envCfg)
	if err != nil {
		l.Log.Error("Error starting server", zap.Error(err))
	}
	if len(envCfg.BaseURL) > 0 {
		cfg.BaseURL = envCfg.BaseURL
	}
	if len(envCfg.ServerAddress) > 0 {
		cfg.ServerAddress = envCfg.ServerAddress
	}
	if len(envCfg.FileStoragePath) > 0 {
		cfg.FileStoragePath = envCfg.FileStoragePath
	}
	if len(envCfg.DatabaseDSN) > 0 {
		cfg.DatabaseDSN = envCfg.DatabaseDSN
	}
}
