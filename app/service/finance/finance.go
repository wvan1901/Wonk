package finance

import (
	"context"
	"log/slog"
	"net/http"
	"time"
	"wonk/app/templates/views"
	"wonk/business/finance"
)

type FinanceService struct {
	Home        Finance
	Transaction Transaction
	Bucket      Bucket
}

type Finance interface {
	Home() http.HandlerFunc
}

type FinaceHandler struct {
	Logger *slog.Logger
}

func InitFinanceService(l *slog.Logger, f finance.Finance) *FinanceService {
	return &FinanceService{
		Home:        initFinanceHandler(l),
		Transaction: initTransactionHandler(l, f),
		Bucket:      initBucketHandler(l, f),
	}

}

func initFinanceHandler(l *slog.Logger) Finance {
	return &FinaceHandler{
		Logger: l,
	}
}

func (f *FinaceHandler) Home() http.HandlerFunc {
	funcName := "Home"
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			ctx, cancel := context.WithTimeout(r.Context(), time.Second*20)
			defer cancel()
			htmxReqHeader := r.Header.Get("hx-request")
			isHtmxRequest := htmxReqHeader == "true"
			if isHtmxRequest {
				tmplFinanceDiv := views.Finance()
				err := tmplFinanceDiv.Render(ctx, w)
				if err != nil {
					f.Logger.Error(funcName, slog.String("Error", err.Error()))
				}
				return
			} else {
				tmplFinanceDiv := views.FinancePage()
				err := tmplFinanceDiv.Render(ctx, w)
				if err != nil {
					f.Logger.Error(funcName, slog.String("Error", err.Error()))
				}
				return
			}
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}
