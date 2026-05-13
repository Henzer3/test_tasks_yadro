package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"test.task.log.server/internal/adapters/rest"
	"test.task.log.server/internal/adapters/rest/middleware"
	"test.task.log.server/internal/config"
	db "test.task.log.server/internal/repository"
	"test.task.log.server/internal/usercase"
)

const gracefulShutdownTime = time.Second * 3

func main() {
	cfg := config.MustLoad()

	log := mustMakeLogger(cfg.LogLevel)

	log.Info("starting server")
	log.Debug("debug messages are enabled")

	// create logRepository
	storage, err := db.New(log, cfg.DBAddress)
	if err != nil {
		log.Error("failed to connect to db", "err", err)
		return
	}

	defer func() {
		if err := storage.Close(); err != nil {
			log.Error("close conn in database adapter", "err", err)
		}
	}()

	// migrations :TODO

	if err := storage.Migrate(); err != nil {
		log.Error("failed to migrate db", "err", err)
		return
	}

	// create service
	service := usercase.New(log, storage)

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/parse", rest.NewParseHandler(log, service))
	mux.HandleFunc("GET /api/v1/log/{log_id}", rest.NewLogHandler(log, service))
	mux.HandleFunc("GET /api/v1/topology/{log_id}", rest.NewTopologyHandler(log, service))
	mux.HandleFunc("GET /api/v1/node/{node_id}", rest.NewNodeHandler(log, service))
	mux.HandleFunc("GET /api/v1/port/{node_id}", rest.NewPortsHandler(log, service))

	var handler http.Handler = mux

	handler = middleware.Logging(handler, log)

	server := http.Server{
		Addr:        cfg.Rest.Address,
		ReadTimeout: cfg.Rest.Timeout,
		Handler:     handler,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	chanError := make(chan error, 1)
	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		ctxShutdown, cancel := context.WithTimeout(context.Background(), gracefulShutdownTime)
		defer cancel()

		chanError <- server.Shutdown(ctxShutdown)
	}()

	if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		log.Error("server closed unexpectedly", "error", err)
		return
	}

	// Make sure the program doesn't exit and waits instead for Shutdown to return.
	if err := <-chanError; err != nil {
		log.Debug("hard stop server", "err", err)
	}
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
