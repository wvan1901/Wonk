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
	funcName := "Transaction"
	return func(w http.ResponseWriter, r *http.Request) {
		htmxReqHeader := r.Header.Get("hx-request")
		isHtmxRequest := htmxReqHeader == "true"
		if !isHtmxRequest {
			http.Error(w, "misssing header 'hx-request'", 400)
			return
		}
		reqCtx := r.Context()
		ctx, cancel := context.WithTimeout(reqCtx, time.Second*20)
		defer cancel()
		curUser, err := auth.UserCtx(reqCtx)
		if err != nil {
			t.Logger.Error(funcName, slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
			http.Error(w, "Internal Error, try logging in again", 500)
			return
		}
		buckets, err := t.FinanceLogic.UserBuckets(curUser.UserId)
		if err != nil {
			http.Error(w, "Internal error", 500)
			return
		}
		switch r.Method {
		case "GET":
			formData := views.TransactionFormData{}
			tmplFinanceDiv := views.FinanceSubmit(buckets, formData)
			err = tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				t.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("Error", err.Error()))
			}
			return
		case "POST":
			err := r.ParseForm()
			if err != nil {
				t.Logger.Error(funcName, slog.String("httpMethod", "POST"), slog.String("error", err.Error()))
				http.Error(w, "Internal Error: Parsing Form", 500)
				return
			}
			formData := TransactionNewInput{
				Name:      r.FormValue("name"),
				Month:     r.FormValue("month"),
				Year:      r.FormValue("year"),
				Price:     r.FormValue("price"),
				IsExpense: r.FormValue("isExpense"),
				UserId:    curUser.UserId,
				BucketId:  r.FormValue("bucket"),
			}
			dbTranaction, problems := parseNewTransaction(formData)
			if len(problems) == 0 {
				problems, err = t.FinanceLogic.SubmitNewTransaction(dbTranaction)
				if err != nil {
					t.Logger.Error(funcName, slog.String("HttpMethod", "POST"), slog.String("error", err.Error()))
					http.Error(w, "Internal Error", 500)
					return
				}
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
				tmplFinanceDiv := views.TransactionForm(buckets, formData)
				err = tmplFinanceDiv.Render(ctx, w)
				if err != nil {
					t.Logger.Error(funcName, slog.String("httpMethod", "POST"), slog.String("Error", err.Error()))
				}
				return
			}

			successMessage := views.SuccessfulTransaction()
			err = successMessage.Render(ctx, w)
			if err != nil {
				t.Logger.Error(funcName, slog.String("httpMethod", "POST"), slog.String("Error", err.Error()))
			}
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}

func (t *TransactionHandler) TransactionMonth() http.HandlerFunc {
	funcName := "TransactionMonth"
	return func(w http.ResponseWriter, r *http.Request) {
		htmxReqHeader := r.Header.Get("hx-request")
		isHtmxRequest := htmxReqHeader == "true"
		if !isHtmxRequest {
			http.Error(w, "misssing header 'hx-request'", 400)
			return
		}
		reqCtx := r.Context()
		ctx, cancel := context.WithTimeout(reqCtx, time.Second*20)
		defer cancel()
		curUser, err := auth.UserCtx(reqCtx)
		if err != nil {
			t.Logger.Error(funcName, slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
			http.Error(w, "Internal Error, try logging in again", 500)
			return
		}
		switch r.Method {
		case "GET":
			err := r.ParseForm()
			if err != nil {
				t.Logger.Error(funcName, slog.String("HttpMethod", "POST"), slog.String("Error", err.Error()))
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
				t.Logger.Error(funcName, slog.String("HttpMethod", "POST"), slog.String("Error", err.Error()))
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
				t.Logger.Error(funcName, slog.String("HttpMethod", "POST"), slog.String("Error", err.Error()), slog.String("DevNote", "templ"))
			}
			return
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}

func (t *TransactionHandler) TransactionMonthForm() http.HandlerFunc {
	funcName := "TransactionMonthForm"
	return func(w http.ResponseWriter, r *http.Request) {
		htmxReqHeader := r.Header.Get("hx-request")
		isHtmxRequest := htmxReqHeader == "true"
		if !isHtmxRequest {
			http.Error(w, "misssing header 'hx-request'", 400)
			return
		}
		reqCtx := r.Context()
		ctx, cancel := context.WithTimeout(reqCtx, time.Second*20)
		defer cancel()
		curUser, err := auth.UserCtx(reqCtx)
		if err != nil {
			t.Logger.Error(funcName, slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
			http.Error(w, "Internal Error, try logging in again", 500)
			return
		}
		switch r.Method {
		case "GET":
			curTime := time.Now()
			summary, err := t.FinanceLogic.MonthlySummary(curUser.UserId, int(curTime.Month()), curTime.Year())
			if err != nil {
				t.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("Error", err.Error()), slog.String("DevNote", "Issue with user buckets"))
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
			tmplFinanceDiv := views.MonthlySummary(*summary)
			err = tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				t.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("Error", err.Error()))
			}
			return
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}

func (t *TransactionHandler) Transactions() http.HandlerFunc {
	funcName := "Transactions"
	return func(w http.ResponseWriter, r *http.Request) {
		htmxReqHeader := r.Header.Get("hx-request")
		isHtmxRequest := htmxReqHeader == "true"
		if !isHtmxRequest {
			http.Error(w, "misssing header 'hx-request'", 400)
			return
		}
		reqCtx := r.Context()
		ctx, cancel := context.WithTimeout(reqCtx, time.Second*20)
		defer cancel()
		curUser, err := auth.UserCtx(reqCtx)
		if err != nil {
			t.Logger.Error(funcName, slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
			http.Error(w, "Internal Error, try logging in again", 500)
			return
		}
		// Getting pagination Info
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

		// Getting Sorting info
		sortColumn := r.URL.Query().Get("sortcolumn")
		sortDirection := r.URL.Query().Get("sortdirection")
		if sortDirection == "" {
			sortColumn = ""
		}
		isAscending := true
		if sortDirection != "ascending" {
			isAscending = false
		}
		switch r.Method {
		case "GET":
			transactions, err := t.FinanceLogic.GetTransactions(page, pageSize, curUser.UserId, sortColumn, isAscending)
			if err != nil {
				w.WriteHeader(500)
				t.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("Error", err.Error()))
				return
			}
			pageData := views.TransactionTableInfo{
				Pagination: views.Pagination{
					Page:     page,
					PageSize: pageSize,
				},
				Sorting: views.Sorting{
					CurrentColumn: sortColumn,
					Direction:     sortDirection,
				},
				Transactions: transactions,
			}
			tmplFinanceDiv := views.TransactionTable(pageData)
			err = tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				t.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("Error", err.Error()))
			}
			return
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}

func (t *TransactionHandler) TransactionsEdit() http.HandlerFunc {
	funcName := "TransactionsEdit"
	return func(w http.ResponseWriter, r *http.Request) {
		htmxReqHeader := r.Header.Get("hx-request")
		isHtmxRequest := htmxReqHeader == "true"
		if !isHtmxRequest {
			http.Error(w, "misssing header 'hx-request'", 400)
			return
		}
		reqCtx := r.Context()
		ctx, cancel := context.WithTimeout(reqCtx, time.Second*20)
		defer cancel()
		curUser, err := auth.UserCtx(reqCtx)
		if err != nil {
			t.Logger.Error(funcName, slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
			http.Error(w, "Internal Error, try logging in again", 500)
			return
		}
		switch r.Method {
		case "GET":
			transactionId := r.PathValue("id")
			transaction, err := t.FinanceLogic.GetTransaction(transactionId)
			if err != nil {
				t.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("Error", err.Error()))
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
				t.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("Error", err.Error()))
			}
			return
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}

func (t *TransactionHandler) TransactionsById() http.HandlerFunc {
	funcName := "TransactionsById"
	return func(w http.ResponseWriter, r *http.Request) {
		htmxReqHeader := r.Header.Get("hx-request")
		isHtmxRequest := htmxReqHeader == "true"
		if !isHtmxRequest {
			http.Error(w, "misssing header 'hx-request'", 400)
			return
		}
		reqCtx := r.Context()
		ctx, cancel := context.WithTimeout(reqCtx, time.Second*20)
		defer cancel()
		curUser, err := auth.UserCtx(reqCtx)
		if err != nil {
			t.Logger.Error(funcName, slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
			http.Error(w, "Internal Error, try logging in again", 500)
			return
		}
		transactionId := r.PathValue("id")
		transaction, err := t.FinanceLogic.GetTransaction(transactionId)
		if err != nil {
			t.Logger.Error(funcName, slog.String("Error", err.Error()))
			w.WriteHeader(500)
			return
		}
		if curUser.UserId != transaction.UserId {
			w.WriteHeader(403)
			return
		}
		switch r.Method {
		case "GET":
			tmplFinanceDiv := views.GetTransactionRow(*transaction)
			err = tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				t.Logger.Error(funcName, slog.String("Error", err.Error()))
			}
			return
		case "PUT":
			err := r.ParseForm()
			if err != nil {
				t.Logger.Error(funcName, slog.String("HttpMethod", "PUT"), slog.String("Error", err.Error()))
				http.Error(w, "Internal Error", 500)
				return
			}
			formData := TransactionEditInput{
				TransactionId: transaction.Id,
				Name:          r.FormValue("name"),
				Month:         r.FormValue("month"),
				Year:          r.FormValue("year"),
				Price:         r.FormValue("price"),
				BucketId:      r.FormValue("bucketId"),
			}
			validTransaction, problems := parseEditTransaction(formData)
			if len(problems) > 0 {
				http.Error(w, "Invalid inputs", 400)
				return
			}
			err = t.FinanceLogic.UpdateTransaction(validTransaction)
			if err != nil {
				t.Logger.Error(funcName, slog.String("HttpMethod", "PUT"), slog.String("Error", err.Error()))
				http.Error(w, "Internal Error", 500)
				return
			}
			transaction, err := t.FinanceLogic.GetTransaction(transactionId)
			if err != nil {
				t.Logger.Error(funcName, slog.String("Error", err.Error()))
				w.WriteHeader(500)
				return
			}
			tmplFinanceDiv := views.GetTransactionRow(*transaction)
			err = tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				t.Logger.Error(funcName, slog.String("Error", err.Error()))
			}
			return
		case "DELETE":
			err := t.FinanceLogic.DeleteTransaction(transaction.Id)
			if err != nil {
				t.Logger.Error(funcName, slog.String("HttpMethod", "DELETE"), slog.String("Error", err.Error()))
				http.Error(w, "Internal Error", 500)
				return
			}
			rowtTmpl := views.GetTransactionDeletedRow()
			err = rowtTmpl.Render(ctx, w)
			if err != nil {
				t.Logger.Error(funcName, slog.String("Error", err.Error()))
			}
			return
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}
