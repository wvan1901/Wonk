package routes

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"wonk/app/database"
	"wonk/app/views"
)

func AddRoutes(
	mux *http.ServeMux,
	l *slog.Logger,
	db database.Database,
) {
	fs := http.FileServer(http.Dir("static"))
	mux.Handle("/", http.NotFoundHandler())
	mux.Handle("/health", handleHealth(l))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))
	mux.Handle("/home", handleHome(l))
	mux.Handle("/finance", handleFinance(l))
	mux.Handle("/finance/submit", handleFinanceSubmit(l, db))
	mux.Handle("/finance/submit/bucket", handleFinanceSubmitBucket(l, db))
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

func handleFinanceSubmit(l *slog.Logger, db database.Database) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// TODO: Refactor: Abstract logic & db
			// TODO: Get userId from middleware
			userId := 1
			buckets, err := db.UserBuckets(userId)
			if err != nil {
				http.Error(w, "Internal error", 500)
				return
			}
			months := []views.Month{
				{Name: "Jan", Value: "1", IsCurrent: false},
				{Name: "Feb", Value: "2", IsCurrent: false},
				{Name: "Mar", Value: "3", IsCurrent: false},
				{Name: "Apr", Value: "4", IsCurrent: false},
				{Name: "May", Value: "5", IsCurrent: false},
				{Name: "June", Value: "6", IsCurrent: false},
				{Name: "July", Value: "7", IsCurrent: false},
				{Name: "Aug", Value: "8", IsCurrent: false},
				{Name: "Sep", Value: "9", IsCurrent: false},
				{Name: "Oct", Value: "10", IsCurrent: false},
				{Name: "Nov", Value: "11", IsCurrent: false},
				{Name: "Dec", Value: "12", IsCurrent: false},
			}
			curMonth := int(time.Now().Month())
			months[curMonth-1].IsCurrent = true
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
				conversionProblems := make(map[string]string)
				month, err := strconv.Atoi(r.FormValue("month"))
				if err != nil {
					conversionProblems["Month"] = "Invalid Month: Not a number"
				}
				year, err := strconv.Atoi(r.FormValue("year"))
				if err != nil {
					conversionProblems["Year"] = "Invalid Year: Not a number"
				}
				price, err := strconv.ParseFloat(r.FormValue("price"), 64)
				if err != nil {
					conversionProblems["Price"] = "Invalid Price: Not a decimal"
				}
				bucketId, err := strconv.Atoi(r.FormValue("bucket"))
				if err != nil {
					conversionProblems["BucketId"] = "Invalid BucketId: Not a number"
				}
				if len(conversionProblems) > 0 {
					l.Error("handleFinanceSubmit", slog.String("Route", "/finance/submit"), slog.String("HttpMethod", "POST"), slog.Any("error", conversionProblems), slog.String("DevNote", "There was a conversion error with form values"))
					http.Error(w, "Internal Error: Conversion Err", 500)
					return
				}
				transactionInput := database.TransactionItemInput{
					Name:     r.FormValue("name"),
					Month:    month,
					Year:     year,
					Price:    price,
					UserId:   userId,
					BucketId: bucketId,
				}
				problems := transactionInput.Valid()
				if len(problems) > 0 {
					// If there is a problem return form with errs
					w.WriteHeader(422)
					formData := views.TransactionFormData{
						NameValue:   r.FormValue("name"),
						MonthValue:  r.FormValue("month"),
						YearValue:   r.FormValue("year"),
						PriceValue:  r.FormValue("price"),
						BucketValue: r.FormValue("bucket"),
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
				_, err = db.CreateItemTransaction(transactionInput)
				if err != nil {
					l.Error("handleFinanceSubmit", slog.String("Route", "/finance/submit"), slog.String("HttpMethod", "POST"), slog.String("error", err.Error()), slog.String("DevNote", "DB error"))
					http.Error(w, "Internal Error", 500)
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

func handleFinanceSubmitBucket(l *slog.Logger, db database.Database) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// TODO: Use middleware to get user info
			userId := 1
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
				if len(newName) > 20 || len(newName) == 0 {
					nameErr := "Name value must not be empty or greater than 20 characters"
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
				_, err = db.CreateBucket(userId, newName)
				if err != nil {
					l.Error("handleFinanceSubmitBucket", slog.String("HttpMethod", "POST"), slog.String("error", err.Error()), slog.String("DevNote", "DB error"))
					http.Error(w, "Internal Error", 500)
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
