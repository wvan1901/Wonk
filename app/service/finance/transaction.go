package finance

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"time"
	"wonk/app/auth"
	"wonk/app/templates/views"
	"wonk/business/finance"
)

type Transaction interface {
	Transaction() http.HandlerFunc
	TransactionMonth() http.HandlerFunc
	TransactionMonthForm() http.HandlerFunc
	Transactions() http.HandlerFunc
	TransactionsEdit() http.HandlerFunc
	TransactionsById() http.HandlerFunc
}

type TransactionHandler struct {
	Logger       *slog.Logger
	FinanceLogic finance.Finance
}

func initTransactionHandler(l *slog.Logger, f finance.Finance) Transaction {
	return &TransactionHandler{
		Logger:       l,
		FinanceLogic: f,
	}
}

func (t *TransactionHandler) Transaction() http.HandlerFunc {
	funcName := "handleFinanceTransactions"
	path := "/finance/transaction"
	return func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, cancel := context.WithTimeout(reqCtx, time.Second*20)
		defer cancel()
		curUser, err := auth.UserCtx(reqCtx)
		if err != nil {
			t.Logger.Error(funcName, slog.String("path", path), slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
			http.Error(w, "Internal Error, try logging in again", 500)
			return
		}
		buckets, err := t.FinanceLogic.UserBuckets(curUser.UserId)
		if err != nil {
			http.Error(w, "Internal error", 500)
			return
		}
		months := views.GetMonths()
		switch r.Method {
		case "GET":
			htmxReqHeader := r.Header.Get("hx-request")
			isHtmxRequest := htmxReqHeader == "true"
			if !isHtmxRequest { // Build entire page or redirect to finance
				w.WriteHeader(404)
				return
			}
			formData := views.TransactionFormData{}
			tmplFinanceDiv := views.FinanceSubmit(buckets, formData, months)
			err = tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				t.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("path", path), slog.String("Error", err.Error()))
			}
			return
		case "POST":
			err := r.ParseForm()
			if err != nil {
				t.Logger.Error(funcName, slog.String("httpMethod", "POST"), slog.String("path", path), slog.String("error", err.Error()))
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
			problems, err := t.FinanceLogic.SubmitNewTransaction(formData)
			if err != nil {
				t.Logger.Error(funcName, slog.String("HttpMethod", "POST"), slog.String("path", path), slog.String("error", err.Error()))
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
				err = tmplFinanceDiv.Render(ctx, w)
				if err != nil {
					t.Logger.Error(funcName, slog.String("httpMethod", "POST"), slog.String("path", path), slog.String("Error", err.Error()))
				}
				return
			}

			successMessage := views.SuccessfulTransaction()
			err = successMessage.Render(ctx, w)
			if err != nil {
				t.Logger.Error(funcName, slog.String("httpMethod", "POST"), slog.String("path", path), slog.String("Error", err.Error()))
			}
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}

func (t *TransactionHandler) TransactionMonth() http.HandlerFunc {
	funcName := "handleFinanceMontlySummary"
	path := "/finance/transactions/month"
	return func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, cancel := context.WithTimeout(reqCtx, time.Second*20)
		defer cancel()
		curUser, err := auth.UserCtx(reqCtx)
		if err != nil {
			t.Logger.Error(funcName, slog.String("path", path), slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
			http.Error(w, "Internal Error, try logging in again", 500)
			return
		}
		switch r.Method {
		case "GET":
			err := r.ParseForm()
			if err != nil {
				t.Logger.Error(funcName, slog.String("HttpMethod", "POST"), slog.String("path", path), slog.String("Error", err.Error()))
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
			summary, err := t.FinanceLogic.MonthlySummary(curUser.UserId, monthInt, yearInt)
			if err != nil {
				t.Logger.Error(funcName, slog.String("HttpMethod", "POST"), slog.String("path", path), slog.String("Error", err.Error()))
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
			err = tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				t.Logger.Error(funcName, slog.String("HttpMethod", "POST"), slog.String("path", path), slog.String("Error", err.Error()), slog.String("DevNote", "templ"))
			}
			return
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}

func (t *TransactionHandler) TransactionMonthForm() http.HandlerFunc {
	funcName := "handleFinanceBucket"
	path := "/finance/transactions/month/form"
	return func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, cancel := context.WithTimeout(reqCtx, time.Second*20)
		defer cancel()
		curUser, err := auth.UserCtx(reqCtx)
		if err != nil {
			t.Logger.Error(funcName, slog.String("Error", err.Error()), slog.String("path", path), slog.String("DevNote", "Issue getting user info from middleware ctx"))
			http.Error(w, "Internal Error, try logging in again", 500)
			return
		}
		switch r.Method {
		case "GET":
			curTime := time.Now()
			summary, err := t.FinanceLogic.MonthlySummary(curUser.UserId, int(curTime.Month()), curTime.Year())
			if err != nil {
				t.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("path", path), slog.String("Error", err.Error()), slog.String("DevNote", "Issue with user buckets"))
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
			if !isHtmxRequest { // Build entire page or redirect to finance
				w.WriteHeader(404)
				return
			}
			tmplFinanceDiv := views.MonthlySummary(*summary)
			err = tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				t.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("path", path), slog.String("Error", err.Error()))
			}
			return
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}

func (t *TransactionHandler) Transactions() http.HandlerFunc {
	funcName := "handleTransactionTable"
	path := "/finance/transactions"
	return func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, cancel := context.WithTimeout(reqCtx, time.Second*20)
		defer cancel()
		curUser, err := auth.UserCtx(reqCtx)
		if err != nil {
			t.Logger.Error(funcName, slog.String("path", path), slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
			http.Error(w, "Internal Error, try logging in again", 500)
			return
		}
		page := 1
		pageSize := 10
		pageParam := r.URL.Query().Get("page")
		if pageParam != "" {
			pageConv, err := strconv.Atoi(pageParam)
			if err == nil {
				page = pageConv
			}
		}
		pageSizeParam := r.URL.Query().Get("pagesize")
		if pageSizeParam != "" {
			sizeConv, err := strconv.Atoi(pageSizeParam)
			if err == nil {
				pageSize = sizeConv
			}
		}
		switch r.Method {
		case "GET":
			htmxReqHeader := r.Header.Get("hx-request")
			isHtmxRequest := htmxReqHeader == "true"
			if !isHtmxRequest { // Build entire page or redirect to finance
				w.WriteHeader(404)
				return
			}
			transactions, err := t.FinanceLogic.GetTransactions(page, pageSize, curUser.UserId)
			if err != nil {
				w.WriteHeader(500)
				t.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("path", path), slog.String("Error", err.Error()))
				return
			}
			pageData := views.TransactionTableInfo{
				Pagination: views.Pagination{
					Page:     page,
					PageSize: pageSize,
				},
				Transactions: transactions,
			}
			tmplFinanceDiv := views.TransactionTable(pageData)
			err = tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				t.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("path", path), slog.String("Error", err.Error()))
			}
			return
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}

func (t *TransactionHandler) TransactionsEdit() http.HandlerFunc {
	funcName := "handleTransactionEdit"
	path := "/finance/transactions/{id}/edit"
	return func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, cancel := context.WithTimeout(reqCtx, time.Second*20)
		defer cancel()
		curUser, err := auth.UserCtx(reqCtx)
		if err != nil {
			t.Logger.Error(funcName, slog.String("path", path), slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
			http.Error(w, "Internal Error, try logging in again", 500)
			return
		}
		switch r.Method {
		case "GET":
			transactionId := r.PathValue("id")
			htmxReqHeader := r.Header.Get("hx-request")
			isHtmxRequest := htmxReqHeader == "true"
			if !isHtmxRequest { // Build entire page or redirect to finance
				w.WriteHeader(404)
				return
			}
			transaction, err := t.FinanceLogic.GetTransaction(transactionId)
			if err != nil {
				t.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("path", path), slog.String("Error", err.Error()))
				w.WriteHeader(500)
				return
			}
			if curUser.UserId != transaction.UserId {
				w.WriteHeader(403)
				return
			}
			userBuckets, err := t.FinanceLogic.UserBuckets(curUser.UserId)
			if err != nil {
				w.WriteHeader(500)
				return
			}
			tmplFinanceDiv := views.EditTransactionRow(*transaction, userBuckets)
			err = tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				t.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("path", path), slog.String("Error", err.Error()))
			}
			return
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}

func (t *TransactionHandler) TransactionsById() http.HandlerFunc {
	funcName := "handleFinanceTransactionListRow"
	path := "/finance/transactions/{id}"
	return func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, cancel := context.WithTimeout(reqCtx, time.Second*20)
		defer cancel()
		curUser, err := auth.UserCtx(reqCtx)
		if err != nil {
			t.Logger.Error(funcName, slog.String("path", path), slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
			http.Error(w, "Internal Error, try logging in again", 500)
			return
		}
		transactionId := r.PathValue("id")
		transaction, err := t.FinanceLogic.GetTransaction(transactionId)
		if err != nil {
			t.Logger.Error(funcName, slog.String("path", path), slog.String("Error", err.Error()))
			w.WriteHeader(500)
			return
		}
		if curUser.UserId != transaction.UserId {
			w.WriteHeader(403)
			return
		}
		htmxReqHeader := r.Header.Get("hx-request")
		isHtmxRequest := htmxReqHeader == "true"
		if !isHtmxRequest { // Build entire page or redirect to finance
			w.WriteHeader(404)
			return
		}
		// TODO: Implement the method to submit the edit
		switch r.Method {
		case "GET":
			tmplFinanceDiv := views.GetTransactionRow(*transaction)
			err = tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				t.Logger.Error(funcName, slog.String("path", path), slog.String("Error", err.Error()))
			}
			return
		case "PUT":
			err := r.ParseForm()
			if err != nil {
				t.Logger.Error(funcName, slog.String("HttpMethod", "PUT"), slog.String("path", path), slog.String("Error", err.Error()))
				http.Error(w, "Internal Error", 500)
				return
			}
			formData := finance.TransactionRowFormInput{
				TransactionId: transaction.Id,
				Name:          r.FormValue("name"),
				Month:         r.FormValue("month"),
				Year:          r.FormValue("year"),
				Price:         r.FormValue("price"),
				BucketId:      r.FormValue("bucketId"),
			}
			err = t.FinanceLogic.UpdateTransaction(formData)
			if err != nil {
				t.Logger.Error(funcName, slog.String("HttpMethod", "PUT"), slog.String("path", path), slog.String("Error", err.Error()))
				http.Error(w, "Internal Error", 500)
				return
			}
			transaction, err := t.FinanceLogic.GetTransaction(transactionId)
			if err != nil {
				t.Logger.Error(funcName, slog.String("path", path), slog.String("Error", err.Error()))
				w.WriteHeader(500)
				return
			}
			tmplFinanceDiv := views.GetTransactionRow(*transaction)
			err = tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				t.Logger.Error(funcName, slog.String("path", path), slog.String("Error", err.Error()))
			}
			return
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}
