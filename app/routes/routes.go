package routes

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strconv"
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
	mux.Handle("/finance", s.Auth.AuthMiddleware(handleFinance(l)))
	mux.Handle("/finance/transaction", s.Auth.AuthMiddleware(handleFinanceTransactions(l, s.Finance)))
	mux.Handle("/finance/bucket/form", s.Auth.AuthMiddleware(handleFinanceSubmitBucket(l, s.Finance)))
	mux.Handle("/finance/bucket/search", s.Auth.AuthMiddleware(handleFinanceBucket(l, s.Finance)))
	mux.Handle("/finance/bucket/list", s.Auth.AuthMiddleware(handleFinanceBucketList(l, s.Finance)))
	mux.Handle("/finance/bucket/list/{id}/edit", s.Auth.AuthMiddleware(handleFinanceBucketListEdit(l, s.Finance)))
	mux.Handle("/finance/bucket/list/{id}", s.Auth.AuthMiddleware(handleFinanceBucketListRow(l, s.Finance)))
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

func handleFinance(l *slog.Logger) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			switch r.Method {
			case "GET":
				htmxReqHeader := r.Header.Get("hx-request")
				isHtmxRequest := htmxReqHeader == "true"
				if isHtmxRequest {
					tmplFinanceDiv := views.Finance()
					err := tmplFinanceDiv.Render(context.TODO(), w)
					if err != nil {
						l.Error("/finance: error", slog.String("Error", err.Error()))
					}
					return
				} else {
					tmplFinanceDiv := views.FinancePage()
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

func handleFinanceTransactions(l *slog.Logger, f finance.Finance) http.Handler {
	funcName := "handleFinanceTransactions"
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

func handleFinanceBucket(l *slog.Logger, f finance.Finance) http.Handler {
	funcName := "handleFinanceBucket"
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
				summary, err := f.MonthlySummary(curUser.UserId, int(curTime.Month()), curTime.Year())
				if err != nil {
					l.Error(funcName, slog.String("Error", err.Error()), slog.String("DevNote", "Issue with user buckets"))
					http.Error(w, "Internal Error, try logging in again", 500)
					return
				}
				filteredBuckets := []finance.BucketSummary{}
				for _, bucket := range summary.BucketsSummary {
					if bucket.Price != 0 {
						filteredBuckets = append(filteredBuckets, bucket)
					}
				}
				summary.BucketsSummary = filteredBuckets
				htmxReqHeader := r.Header.Get("hx-request")
				isHtmxRequest := htmxReqHeader == "true"
				if isHtmxRequest {
					tmplFinanceDiv := views.MonthlySummary(*summary)
					err := tmplFinanceDiv.Render(context.TODO(), w)
					if err != nil {
						l.Error("/finance: error", slog.String("Error", err.Error()))
					}
					return
				} else {
					// Build entire page or redirect to finance
					w.WriteHeader(404)
					return
				}
			case "POST": // This should return the month data
				err := r.ParseForm()
				if err != nil {
					l.Error("handleFinance", slog.String("HttpMethod", "POST"), slog.String("Error", err.Error()))
					http.Error(w, "Internal Error: Parsing Form", 500)
					return
				}
				month := r.FormValue("month")
				year := r.FormValue("year")
				monthInt, err := strconv.Atoi(month)
				if err != nil {
					http.Error(w, "Bad Request: Month Isn't a int", 400)
					return
				}
				yearInt, err := strconv.Atoi(year)
				if err != nil {
					http.Error(w, "Bad Request: Year Isn't a int", 400)
					return
				}
				summary, err := f.MonthlySummary(curUser.UserId, monthInt, yearInt)
				if err != nil {
					l.Error("handleFinance", slog.String("HttpMethod", "POST"), slog.String("Error", err.Error()))
					http.Error(w, "Internal Error", 500)
					return
				}
				filteredBuckets := []finance.BucketSummary{}
				for _, bucket := range summary.BucketsSummary {
					if bucket.Price != 0 {
						filteredBuckets = append(filteredBuckets, bucket)
					}
				}
				summary.BucketsSummary = filteredBuckets
				tmplFinanceDiv := views.MonthlyTable(*summary)
				err = tmplFinanceDiv.Render(context.TODO(), w)
				if err != nil {
					l.Error("handleFinance", slog.String("HttpMethod", "POST"), slog.String("Error", err.Error()), slog.String("DevNote", "templ"))
				}
				return
			default:
				http.Error(w, "Not valid method", 404)
			}
		},
	)
}

func handleFinanceBucketList(l *slog.Logger, f finance.Finance) http.Handler {
	funcName := "handleFinanceBucketList"
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
					buckets, err := f.UserBuckets(curUser.UserId)
					if err != nil {
						http.Error(w, "Internal error", 500)
						return
					}
					bucketRows := []views.BucketRow{}
					for _, bucket := range buckets {
						newRow := views.BucketRow{BucketId: strconv.Itoa(bucket.Id), BucketName: bucket.Name}
						bucketRows = append(bucketRows, newRow)
					}
					tmplFinanceDiv := views.ViewBuckets(bucketRows)
					err = tmplFinanceDiv.Render(context.TODO(), w)
					if err != nil {
						l.Error(funcName, slog.String("Error", err.Error()))
					}
					return
				} else {
					// Build entire page or redirect to finance
					w.WriteHeader(404)
					return
				}
			default:
				http.Error(w, "Not valid method", 404)
			}
		},
	)
}

func handleFinanceBucketListEdit(l *slog.Logger, f finance.Finance) http.Handler {
	funcName := "handleFinanceBucketListEdit"
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
				bucketId := r.PathValue("id")
				htmxReqHeader := r.Header.Get("hx-request")
				isHtmxRequest := htmxReqHeader == "true"
				if isHtmxRequest {
					bucket, err := f.GetBucket(bucketId)
					if err != nil {
						w.WriteHeader(500)
						return
					}
					if curUser.UserId != bucket.UserId {
						w.WriteHeader(403)
						return
					}
					row := views.BucketRow{BucketId: bucketId, BucketName: bucket.Name}
					tmplFinanceDiv := views.EditBucketRow(row)
					err = tmplFinanceDiv.Render(context.TODO(), w)
					if err != nil {
						l.Error(funcName, slog.String("Error", err.Error()))
					}
					return
				} else {
					// Build entire page or redirect to finance
					w.WriteHeader(404)
					return
				}
			default:
				http.Error(w, "Not valid method", 404)
			}
		},
	)
}

func handleFinanceBucketListRow(l *slog.Logger, f finance.Finance) http.Handler {
	funcName := "handleFinanceBucketListRow"
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			bucketId := r.PathValue("id")
			curUser, err := auth.UserCtx(r.Context())
			if err != nil {
				l.Error(funcName, slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
				http.Error(w, "Internal Error, try logging in again", 500)
				return
			}
			bucket, err := f.GetBucket(bucketId)
			if err != nil {
				w.WriteHeader(500)
				return
			}
			if curUser.UserId != bucket.UserId {
				w.WriteHeader(403)
				return
			}
			switch r.Method {
			case "GET":
				htmxReqHeader := r.Header.Get("hx-request")
				isHtmxRequest := htmxReqHeader == "true"
				if isHtmxRequest {
					row := views.BucketRow{BucketId: bucketId, BucketName: bucket.Name}
					tmplFinanceDiv := views.GetBucketRow(row)
					err = tmplFinanceDiv.Render(context.TODO(), w)
					if err != nil {
						l.Error(funcName, slog.String("Error", err.Error()))
					}
					return
				} else {
					// Build entire page or redirect to finance
					w.WriteHeader(404)
					return
				}
			case "PUT":
				htmxReqHeader := r.Header.Get("hx-request")
				isHtmxRequest := htmxReqHeader == "true"
				if isHtmxRequest {
					err := r.ParseForm()
					if err != nil {
						l.Error(funcName, slog.String("HttpMethod", "PUT"), slog.String("Error", err.Error()))
						http.Error(w, "Internal Error", 500)
						return
					}
					newName := r.FormValue("name")
					err = f.UpdateBucket(bucket.Id, newName)
					if err != nil {
						l.Error(funcName, slog.String("HttpMethod", "PUT"), slog.String("Error", err.Error()))
						http.Error(w, "Internal Error", 500)
						return
					}
					mockRow := views.BucketRow{BucketId: bucketId, BucketName: newName}
					tmplFinanceDiv := views.GetBucketRow(mockRow)
					err = tmplFinanceDiv.Render(context.TODO(), w)
					if err != nil {
						l.Error(funcName, slog.String("Error", err.Error()))
					}
					return
				} else {
					// Build entire page or redirect to finance
					w.WriteHeader(404)
					return
				}
			default:
				http.Error(w, "Not valid method", 404)
			}
		},
	)
}
