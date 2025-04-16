package database

import (
	"strconv"
	"strings"
)

type User struct {
	Id       int
	UserName string
	Password string
}

type Bucket struct {
	Id     int
	Name   string
	UserId int
}

type TransactionItem struct {
	Id        int
	Name      string
	Month     int
	Year      int
	Price     float64
	IsExpense bool
	UserId    int
	BucketId  int
}

type TransactionItemInput struct {
	Name      string
	Month     int
	Year      int
	Price     float64
	IsExpense bool
	UserId    int
	BucketId  int
}

func (t *TransactionItemInput) Valid() map[string]string {
	problems := make(map[string]string)
	maxNameLen := 50
	if len(t.Name) > maxNameLen {
		problems["Name"] = "Name length can't be greater than 50"
	}
	if len(t.Name) == 0 {
		problems["Name"] = "Name length can't be 0"
	}

	if t.Month > 12 || t.Month < 1 {
		problems["Month"] = "Month value isn't between 1-12"
	}

	if t.Year < 2000 || t.Year > 3000 {
		problems["Year"] = "Invalid Year"
	}

	if t.Price <= 0 {
		problems["Price"] = "Invalid Price"
	}

	floatStr := strconv.FormatFloat(t.Price, 'f', -1, 64)
	parts := strings.Split(floatStr, ".")

	twoOrLessDecimalPlaces := false
	if len(parts) < 2 {
		twoOrLessDecimalPlaces = true
	} else {
		twoOrLessDecimalPlaces = len(parts[1]) <= 2
	}

	if !twoOrLessDecimalPlaces {
		problems["Price"] = "Invalid Price: has more than 2 decimal places"
	}

	if t.UserId < 0 {
		problems["UserId"] = "Invalid UserId"
	}

	if t.BucketId < 0 {
		problems["BucketId"] = "Invalid BucketId"
	}

	return problems
}

type TransactionFilters struct {
	Id       int
	Name     *string
	Price    *float64
	Month    *int
	Year     *int
	BucketId *int
}

func (t *TransactionFilters) FilterQueryAndValues() (string, []any) {
	values := []any{}
	query := "WHERE user_id=?"
	values = append(values, t.Id)

	// NOTE: To use the like operator we need to have the value wrapped with wildcards
	if t.Name != nil {
		query += " AND name LIKE ?"
		values = append(values, "%"+*t.Name+"%")
	}

	if t.Price != nil {
		query += " AND price LIKE ?"
		p := strconv.FormatFloat(*t.Price, 'f', -1, 64)
		values = append(values, "%"+p+"%")
	}

	if t.Month != nil {
		query += " AND month LIKE ?"
		values = append(values, "%"+strconv.Itoa(*t.Month)+"%")
	}

	if t.Year != nil {
		query += " AND year LIKE ?"
		values = append(values, "%"+strconv.Itoa(*t.Year)+"%")
	}

	if t.BucketId != nil {
		query += " AND bucket_id LIKE ?"
		values = append(values, "%"+strconv.Itoa(*t.BucketId)+"%")
	}

	return query, values
}
