package routes

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
	"wonk/app/auth"
	"wonk/app/database"
	"wonk/app/services"
	"wonk/app/services/finance"
	"wonk/app/views"
)

func AddRoutes(
	mux *http.ServeMux,
	l *slog.Logger,
	db database.Database,
	s *services.Services,
) {
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", http.NotFoundHandler())
	mux.Handle("/health", handleHealth(l))
	mux.Handle("/login", s.Auth.HandleLogin())
	mux.Handle("/signup", s.Auth.HandleSignUp())
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.Handle("/home", s.Auth.AuthMiddleware(handleHome(l)))
	mux.Handle("/finance", s.Auth.AuthMiddleware(handleFinance(l, s.Finance)))
	mux.Handle("/finance/submit", s.Auth.AuthMiddleware(handleFinanceSubmit(l, s.Finance)))
	mux.Handle("/finance/submit/bucket", s.Auth.AuthMiddleware(handleFinanceSubmitBucket(l, s.Finance)))
}

func handleHealth(l *slog.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			_, err := io.WriteString(w, "Healthy!")
			if err != nil {
				l.Error("handleHealth: io write", slog.Any("error", err))
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
				l.Error("/home: error", slog.String("Error", err.Error()))
				return
			}
		},
	)
}

func handleFinance(l *slog.Logger, f finance.Finance) http.Handler {
	funcName := "handleFinance"
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			curUser, err := auth.UserCtx(r.Context())
			if err != nil {
				l.Error(funcName, slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
				http.Error(w, "Internal Error, try logging in again", 500)
				return
			}
			switch r.Method {
			case "GET":
				curTime := time.Now()
				buckets, err := f.BucketsMonthlySummary(curUser.UserId, int(curTime.Month()), curTime.Year())
				if err != nil {
					l.Error(funcName, slog.String("Error", err.Error()), slog.String("DevNote", "Issue with user buckets"))
					http.Error(w, "Internal Error, try logging in again", 500)
					return
				}
				htmxReqHeader := r.Header.Get("hx-request")
				isHtmxRequest := htmxReqHeader == "true"
				if isHtmxRequest {
					tmplFinanceDiv := views.Finance(buckets)
					err := tmplFinanceDiv.Render(context.TODO(), w)
					if err != nil {
						l.Error("/finance: error", slog.String("Error", err.Error()))
					}
					return
				} else {
					tmplFinanceDiv := views.FinancePage(buckets)
					err := tmplFinanceDiv.Render(context.TODO(), w)
					if err != nil {
						l.Error("/finance: error", slog.String("Error", err.Error()))
					}
					return
				}
			default:
				http.Error(w, "Not valid method", 404)
			}
		},
	)
}

func handleFinanceSubmit(l *slog.Logger, f finance.Finance) http.Handler {
	funcName := "handleFinanceSubmit"
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			curUser, err := auth.UserCtx(r.Context())
			if err != nil {
				l.Error(funcName, slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
				http.Error(w, "Internal Error, try logging in again", 500)
				return
			}
			buckets, err := f.UserBuckets(curUser.UserId)
			if err != nil {
				http.Error(w, "Internal error", 500)
				return
			}
			months := views.GetMonths()
			switch r.Method {
			case "GET":
				htmxReqHeader := r.Header.Get("hx-request")
				isHtmxRequest := htmxReqHeader == "true"
				if isHtmxRequest {
					formData := views.TransactionFormData{}
					tmplFinanceDiv := views.FinanceSubmit(buckets, formData, months)
					err = tmplFinanceDiv.Render(context.TODO(), w)
					if err != nil {
						l.Error("handleFinanceSubmit: GET:", slog.String("Error", err.Error()))
					}
					return
				} else {
					// Build entire page or redirect to finance
					w.WriteHeader(404)
					return
				}
			case "POST":
				err := r.ParseForm()
				if err != nil {
					l.Error("handleFinanceSubmit: /finance/submit: POST: parseForm:", slog.String("error", err.Error()))
					http.Error(w, "Internal Error: Parsing Form", 500)
					return
				}
				formData := finance.TransactionFormInput{
					Name:      r.FormValue("name"),
					Month:     r.FormValue("month"),
					Year:      r.FormValue("year"),
					Price:     r.FormValue("price"),
					IsExpense: r.FormValue("isExpense"),
					UserId:    curUser.UserId,
					BucketId:  r.FormValue("bucket"),
				}
				problems, err := f.SubmitNewTransaction(formData)
				if err != nil {
					l.Error("handleFinanceSubmit", slog.String("Route", "/finance/submit"), slog.String("HttpMethod", "POST"), slog.String("error", err.Error()))
					http.Error(w, "Internal Error", 500)
					return
				}

				if len(problems) > 0 {
					// If there is a problem return form with errs
					w.WriteHeader(422)
					formData := views.TransactionFormData{
						NameValue:   formData.Name,
						MonthValue:  formData.Month,
						YearValue:   formData.Year,
						PriceValue:  formData.Price,
						BucketValue: formData.BucketId,
					}
					if val, ok := problems["Name"]; ok {
						formData.NameErr = &val
					}
					if val, ok := problems["Month"]; ok {
						formData.MonthErr = &val
					}
					if val, ok := problems["Year"]; ok {
						formData.YearErr = &val
					}
					if val, ok := problems["Price"]; ok {
						formData.PriceErr = &val
					}
					if val, ok := problems["BucketId"]; ok {
						formData.BucketErr = &val
					}
					tmplFinanceDiv := views.TransactionForm(buckets, formData, months)
					err = tmplFinanceDiv.Render(context.TODO(), w)
					if err != nil {
						l.Error("handleFinanceSubmit: GET:", slog.String("Error", err.Error()))
					}
					return
				}

				successMessage := views.SuccessfulTransaction()
				err = successMessage.Render(context.TODO(), w)
				if err != nil {
					l.Error("handleFinanceSubmit: GET:", slog.String("Error", err.Error()))
				}
			default:
				http.Error(w, "Not valid method", 404)
			}
		},
	)
}

func handleFinanceSubmitBucket(l *slog.Logger, f finance.Finance) http.Handler {
	funcName := "handleFinanceSubmitBucket"
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			curUser, err := auth.UserCtx(r.Context())
			if err != nil {
				l.Error(funcName, slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
				http.Error(w, "Internal Error, try logging in again", 500)
				return
			}
			switch r.Method {
			case "GET":
				htmxReqHeader := r.Header.Get("hx-request")
				isHtmxRequest := htmxReqHeader == "true"
				if isHtmxRequest {
					formData := views.BucketFormData{}
					tmplFinanceDiv := views.BucketForm(formData)
					err := tmplFinanceDiv.Render(context.TODO(), w)
					if err != nil {
						l.Error("handleFinanceSubmitBucket", slog.String("HttpMethod", "GET"), slog.String("Error", err.Error()))
					}
					return
				} else {
					// Build entire page or redirect to finance
					w.WriteHeader(404)
					return
				}
			case "POST":
				err := r.ParseForm()
				if err != nil {
					l.Error("handleFinanceSubmitBucket", slog.String("HttpMethod", "POST"), slog.String("error", err.Error()), slog.String("DevNote", "Parse form err"))
					http.Error(w, "Internal Error: Parsing Form", 500)
					return
				}
				newName := r.FormValue("name")
				problems, err := f.CreateBucket(curUser.UserId, newName)
				if err != nil {
					l.Error("handleFinanceSubmitBucket", slog.String("HttpMethod", "POST"), slog.String("error", err.Error()))
					http.Error(w, "Internal Error", 500)
					return
				}
				if len(problems) > 0 {
					nameErr := "Internal Error"
					if val, ok := problems["Name"]; ok {
						nameErr = val
					}
					formData := views.BucketFormData{
						NameValue: newName,
						NameErr:   &nameErr,
					}
					w.WriteHeader(422)
					bucketForm := views.BucketForm(formData)
					err = bucketForm.Render(context.TODO(), w)
					if err != nil {
						l.Error("handleFinanceSubmitBucket", slog.String("HttpMethod", "POST"), slog.String("Error", err.Error()), slog.String("DevNote", "Invalid: Templ err"))
					}
					return
				}
				successMessage := views.SuccessfulBucket()
				err = successMessage.Render(context.TODO(), w)
				if err != nil {
					l.Error("handleFinanceSubmitBucket", slog.String("HttpMethod", "POST"), slog.String("Error", err.Error()), slog.String("DevNote", "Success: Templ err"))
				}
			default:
				http.Error(w, "Not valid method", 404)
			}
		},
	)
}
