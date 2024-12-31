package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
	"wonk/app/config"
	"wonk/app/database"
	"wonk/app/routes"
	"wonk/app/secret"
	"wonk/app/services"
	"wonk/logger"

	"github.com/joho/godotenv"
)

const (
	DEFAULT_HOST = "localhost"
	DEFAULT_PORT = "8070"
)

func main() {
	ctx := context.Background()
	if err := run(ctx, os.Getenv, os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, getEnv func(string) string, _ io.Writer, args []string) error {
	err := godotenv.Load()
	if err != nil {
		return err
	}
	// Init Flags
	f := config.InitFlags()

	// Init Logger
	l := logger.InitLogger(f.LogHandler)

	l.Info("Running Server with args:", slog.Any("args", args[1:]))

	// Init Secrets
	secrets, err := secret.InitSecret(getEnv)
	if err != nil {
		return err
	}

	// Init Db
	db, err := database.InitDb()
	if err != nil {
		return err
	}

	// Init Services
	service, err := services.InitServices(secrets, l, db)
	if err != nil {
		return err
	}

	// Create Main Context
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// Create Http Server
	srv := NewServer(l, db, service)
	httpServer := &http.Server{
		Addr:    net.JoinHostPort(DEFAULT_HOST, DEFAULT_PORT),
		Handler: srv,
	}

	// Start Http Server
	go func() {
		l.Info("listening on " + httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("error listening and serving", slog.Any("error", err))
		}
	}()

	// Wait until canceled
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-ctx.Done()
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			l.Error("error shutting down http server", slog.Any("error", err))
		}
	}()
	wg.Wait()

	return nil
}

func NewServer(l *slog.Logger, db database.Database, s *services.Services) http.Handler {
	mux := http.NewServeMux()
	routes.AddRoutes(mux, l, db, s)
	var handler http.Handler = mux
	return handler
}
