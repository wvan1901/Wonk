package server

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
	"wonk/app/config"
	"wonk/app/routes"
	"wonk/app/secret"
	"wonk/app/service"
	"wonk/business"
	"wonk/logger"
	"wonk/storage"

	"github.com/joho/godotenv"
)

const (
	DEFAULT_HOST = "localhost"
	DEFAULT_PORT = "8070"
	FILE_NAME    = "wonk.db"
)

func Run(ctx context.Context, getEnv func(string) string, _ io.Writer, args []string) error {
	// Init Flags
	f := config.InitFlags(args)

	// Load env file
	if !f.ExcluedEnvFile {
		err := godotenv.Load()
		if err != nil {
			return err
		}
	}

	// Init Logger
	l := logger.InitLogger(f.LogHandler)

	l.Info("Running Server with args:", slog.Any("args", args))

	// Init Secrets
	secrets, err := secret.InitSecret(getEnv)
	if err != nil {
		return err
	}

	// Init Db
	db, err := database.InitDb(FILE_NAME, f.EnableTestDb)
	if err != nil {
		return err
	}

	// Init Business Services
	businessService, err := business.InitServices(secrets, l, db)
	if err != nil {
		return err
	}

	// Init App Services
	appServices, err := application.InitServices(secrets, l, businessService)
	if err != nil {
		return err
	}

	// Create Main Context
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	// Create Http Server
	srv := NewServer(l, db, appServices)
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

func NewServer(l *slog.Logger, db database.Database, a *application.Service) http.Handler {
	mux := http.NewServeMux()
	routes.AddRoutes(mux, l, db, a)
	var handler http.Handler = mux
	return handler
}
