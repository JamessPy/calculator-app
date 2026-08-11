package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/JamessPy/calculator-app/internal/calculator"
	"github.com/JamessPy/calculator-app/internal/config"
	"github.com/JamessPy/calculator-app/internal/httpapi"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "err", err)
		os.Exit(1)
	}

	// Composition root: dependencies are wired here, by hand.
	calc := calculator.NewService()
	router := httpapi.NewRouter(calc, logger, cfg.AllowedOrigins)
	srv := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	logger.Info("server starting", "addr", srv.Addr)

	if err := srv.ListenAndServe(); err != nil {
		logger.Error("server failed", "err", err)
		os.Exit(1)
	}
}
