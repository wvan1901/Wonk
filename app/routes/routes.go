package routes

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"wonk/app/views"
)

func AddRoutes(
	mux *http.ServeMux,
	l *slog.Logger,
) {
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", http.NotFoundHandler())
	mux.Handle("/health", handleHealth(l))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.Handle("/home", handleHome(l))
}

func handleHealth(l *slog.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, err := io.WriteString(w, "Healthy!")
			if err != nil {
				l.Error("handleHealth: io write", slog.Any("error", err))
				http.Error(w, "internal error: template error", 500)
				return
			}
			w.WriteHeader(http.StatusOK)
		},
	)
}

func handleHome(l *slog.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			tmplHomePage := views.Page()
			err := tmplHomePage.Render(context.TODO(), w)
			if err != nil {
				l.Error("/home: error", err)
				return
			}
		},
	)
}
