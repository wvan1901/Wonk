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

type Bucket interface {
	BucketForm() http.HandlerFunc
	Buckets() http.HandlerFunc
	BucketEdit() http.HandlerFunc
	BucketById() http.HandlerFunc
}

type BucketHandler struct {
	Logger       *slog.Logger
	FinanceLogic finance.Finance
}

func initBucketHandler(l *slog.Logger, f finance.Finance) Bucket {
	return &BucketHandler{
		Logger:       l,
		FinanceLogic: f,
	}
}

func (b *BucketHandler) BucketForm() http.HandlerFunc {
	funcName := "BucketForm"
	return func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, cancel := context.WithTimeout(reqCtx, time.Second*20)
		defer cancel()
		curUser, err := auth.UserCtx(reqCtx)
		if err != nil {
			b.Logger.Error(funcName, slog.String("Error", err.Error()), slog.String("Note", "Issue getting user info from middleware ctx"))
			http.Error(w, "Internal Error, try logging in again", 500)
			return
		}
		htmxReqHeader := r.Header.Get("hx-request")
		isHtmxRequest := htmxReqHeader == "true"
		if !isHtmxRequest {
			http.Error(w, "misssing header 'hx-request'", 400)
			return
		}
		switch r.Method {
		case "GET":
			formData := views.BucketFormData{}
			tmplFinanceDiv := views.BucketForm(formData)
			err := tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				b.Logger.Error(funcName, slog.String("HttpMethod", "GET"), slog.String("Error", err.Error()))
			}
			return
		case "POST":
			err := r.ParseForm()
			if err != nil {
				b.Logger.Error(funcName, slog.String("HttpMethod", "POST"), slog.String("error", err.Error()), slog.String("Note", "Parse form err"))
				http.Error(w, "Internal Error: Parsing Form", 500)
				return
			}
			newName := r.FormValue("name")
			problems, err := b.FinanceLogic.CreateBucket(curUser.UserId, newName)
			if err != nil {
				b.Logger.Error(funcName, slog.String("HttpMethod", "POST"), slog.String("error", err.Error()))
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
				err = bucketForm.Render(ctx, w)
				if err != nil {
					b.Logger.Error(funcName, slog.String("HttpMethod", "POST"), slog.String("Error", err.Error()), slog.String("Note", "Invalid: Templ err"))
				}
				return
			}
			successMessage := views.SuccessfulBucket()
			err = successMessage.Render(ctx, w)
			if err != nil {
				b.Logger.Error(funcName, slog.String("HttpMethod", "POST"), slog.String("Error", err.Error()), slog.String("Note", "Success: Templ err"))
			}
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}

func (b *BucketHandler) Buckets() http.HandlerFunc {
	funcName := "Buckets"
	return func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, cancel := context.WithTimeout(reqCtx, time.Second*20)
		defer cancel()
		curUser, err := auth.UserCtx(reqCtx)
		if err != nil {
			b.Logger.Error(funcName, slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
			http.Error(w, "Internal Error, try logging in again", 500)
			return
		}
		htmxReqHeader := r.Header.Get("hx-request")
		isHtmxRequest := htmxReqHeader == "true"
		if !isHtmxRequest {
			http.Error(w, "misssing header 'hx-request'", 400)
			return
		}
		switch r.Method {
		case "GET":
			buckets, err := b.FinanceLogic.UserBuckets(curUser.UserId)
			if err != nil {
				b.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("Error", err.Error()))
				http.Error(w, "Internal error", 500)
				return
			}
			bucketRows := []views.BucketRow{}
			for _, bucket := range buckets {
				newRow := views.BucketRow{BucketId: strconv.Itoa(bucket.Id), BucketName: bucket.Name}
				bucketRows = append(bucketRows, newRow)
			}
			tmplFinanceDiv := views.ViewBuckets(bucketRows)
			err = tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				b.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("Error", err.Error()))
			}
			return
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}

func (b *BucketHandler) BucketEdit() http.HandlerFunc {
	funcName := "BucketEdit"
	return func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, cancel := context.WithTimeout(reqCtx, time.Second*20)
		defer cancel()
		curUser, err := auth.UserCtx(reqCtx)
		if err != nil {
			b.Logger.Error(funcName, slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
			http.Error(w, "Internal Error, try logging in again", 500)
			return
		}
		htmxReqHeader := r.Header.Get("hx-request")
		isHtmxRequest := htmxReqHeader == "true"
		if !isHtmxRequest {
			http.Error(w, "misssing header 'hx-request'", 400)
			return
		}
		switch r.Method {
		case "GET":
			bucketId := r.PathValue("id")
			bucket, err := b.FinanceLogic.GetBucket(bucketId)
			if err != nil {
				b.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("Error", err.Error()))
				w.WriteHeader(500)
				return
			}
			if curUser.UserId != bucket.UserId {
				w.WriteHeader(403)
				return
			}
			row := views.BucketRow{BucketId: bucketId, BucketName: bucket.Name}
			tmplFinanceDiv := views.EditBucketRow(row)
			err = tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				b.Logger.Error(funcName, slog.String("httpMethod", "GET"), slog.String("Error", err.Error()))
			}
			return
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}

func (b *BucketHandler) BucketById() http.HandlerFunc {
	funcName := "BucketById"
	return func(w http.ResponseWriter, r *http.Request) {
		reqCtx := r.Context()
		ctx, cancel := context.WithTimeout(reqCtx, time.Second*20)
		defer cancel()
		curUser, err := auth.UserCtx(reqCtx)
		if err != nil {
			b.Logger.Error(funcName, slog.String("Error", err.Error()), slog.String("DevNote", "Issue getting user info from middleware ctx"))
			http.Error(w, "Internal Error, try logging in again", 500)
			return
		}
		bucketId := r.PathValue("id")
		bucket, err := b.FinanceLogic.GetBucket(bucketId)
		if err != nil {
			b.Logger.Error(funcName, slog.String("Error", err.Error()))
			w.WriteHeader(500)
			return
		}
		if curUser.UserId != bucket.UserId {
			w.WriteHeader(403)
			return
		}
		htmxReqHeader := r.Header.Get("hx-request")
		isHtmxRequest := htmxReqHeader == "true"
		if !isHtmxRequest {
			http.Error(w, "misssing header 'hx-request'", 400)
			return
		}
		switch r.Method {
		case "GET":
			row := views.BucketRow{BucketId: bucketId, BucketName: bucket.Name}
			tmplFinanceDiv := views.GetBucketRow(row)
			err = tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				b.Logger.Error(funcName, slog.String("Error", err.Error()))
			}
			return
		case "PUT":
			err := r.ParseForm()
			if err != nil {
				b.Logger.Error(funcName, slog.String("HttpMethod", "PUT"), slog.String("Error", err.Error()))
				http.Error(w, "Internal Error", 500)
				return
			}
			newName := r.FormValue("name")
			err = b.FinanceLogic.UpdateBucket(bucket.Id, newName)
			if err != nil {
				b.Logger.Error(funcName, slog.String("HttpMethod", "PUT"), slog.String("Error", err.Error()))
				http.Error(w, "Internal Error", 500)
				return
			}
			mockRow := views.BucketRow{BucketId: bucketId, BucketName: newName}
			tmplFinanceDiv := views.GetBucketRow(mockRow)
			err = tmplFinanceDiv.Render(ctx, w)
			if err != nil {
				b.Logger.Error(funcName, slog.String("HttpMethod", "PUT"), slog.String("Error", err.Error()))
			}
			return
		default:
			http.Error(w, "Not valid method", 404)
		}
	}
}
