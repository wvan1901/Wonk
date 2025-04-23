package routes

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"
	"wonk/app/service"
	"wonk/app/templates/views"
	"wonk/storage"
)

func AddRoutes(
	mux *http.ServeMux,
	l *slog.Logger,
	db database.Database,
	a *application.Service,
) {
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", http.NotFoundHandler())
	mux.Handle("/health", handleHealth(l))
	mux.Handle("/login", a.Auth.HandleLogin())
	mux.Handle("/signup", a.Auth.HandleSignUp())
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.Handle("/home", a.Auth.AuthMiddleware(handleHome(l)))
	mux.Handle("/finance", a.Auth.AuthMiddleware(a.Finance.Home.Home()))
	mux.Handle("/finance/transaction", a.Auth.AuthMiddleware(a.Finance.Transaction.Transaction()))
	mux.Handle("/finance/bucket/form", a.Auth.AuthMiddleware(a.Finance.Bucket.BucketForm()))
	mux.Handle("/finance/transactions/month", a.Auth.AuthMiddleware(a.Finance.Transaction.TransactionMonth()))
	mux.Handle("/finance/transactions/month/form", a.Auth.AuthMiddleware(a.Finance.Transaction.TransactionMonthForm()))
	mux.Handle("/finance/buckets", a.Auth.AuthMiddleware(a.Finance.Bucket.Buckets()))
	mux.Handle("/finance/buckets/{id}/edit", a.Auth.AuthMiddleware(a.Finance.Bucket.BucketEdit()))
	mux.Handle("/finance/buckets/{id}", a.Auth.AuthMiddleware(a.Finance.Bucket.BucketById()))
	mux.Handle("/finance/transactions", a.Auth.AuthMiddleware(a.Finance.Transaction.Transactions()))
	mux.Handle("/finance/transactions/{id}/edit", a.Auth.AuthMiddleware(a.Finance.Transaction.TransactionsEdit()))
	mux.Handle("/finance/transactions/{id}", a.Auth.AuthMiddleware(a.Finance.Transaction.TransactionsById()))
}

func handleHealth(l *slog.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err := io.WriteString(w, "Healthy!")
			if err != nil {
				l.Error("handleHealth: io write", slog.Any("error", err))
				return
			}
		},
	)
}

func handleHome(l *slog.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), time.Second*20)
			defer cancel()
			tmplHomePage := views.Page()
			err := tmplHomePage.Render(ctx, w)
			if err != nil {
				l.Error("handleHome", slog.String("path", "/home"), slog.String("Error", err.Error()))
				return
			}
		},
	)
}
