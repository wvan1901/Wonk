package server

import (
	"log/slog"
	"net/http"
	"wonk/app/routes"
	"wonk/app/service"
	"wonk/storage"
)

func NewServer(l *slog.Logger, db database.Database, a *application.Service) http.Handler {
	mux := http.NewServeMux()
	routes.AddRoutes(mux, l, db, a)
	var handler http.Handler = mux
	return handler
}
