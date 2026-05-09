package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"test.task.dns/internal/adapter/dns"
	grpcserver "test.task.dns/internal/adapter/grpc"
	"test.task.dns/internal/config"
	"test.task.dns/internal/usercase"
	dnspb "test.task.pkg/generated/dns"
)

func main() {
	cfg := config.MustLoad()

	log := mustMakeLogger(cfg.LogLevel)

	log.Info("start dns initialization server")
	if err := run(cfg, log); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	log.Info("starting server")
	log.Debug("debug messages are enabled")

	// adapter
	dns := dns.New(log, cfg.Path)

	// service
	dnsHandler := usercase.NewService(log, dns)

	// grpc server
	listener, err := net.Listen("tcp", cfg.GRPC.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	server := grpc.NewServer()
	dnspb.RegisterDnsServiceServer(server, grpcserver.NewServer(log, dnsHandler))
	reflection.Register(server)

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Serve(listener)
	}()

	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	select {
	case <-signalCtx.Done():
	case err := <-errChan:
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(cfg.GRPC.GracefulShutdownTime))
	defer cancel()

	stopped := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-ctx.Done():
		log.Error("time is up")
		server.Stop()
	}

	if err := <-errChan; !errors.Is(err, grpc.ErrServerStopped) {
		return err
	}
	return nil
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
