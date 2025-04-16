package finance

import (
	"strconv"
	"strings"
	database "wonk/storage"
)

type BucketSummary struct {
	Reference database.Bucket
	Price     float64
}

type MonthSummary struct {
	BucketsSummary []BucketSummary
	TotalIncome    float64
	TotalExpense   float64
}

type TransactionEdit struct {
	TransactionId int
	Name          string
	Month         int
	Year          int
	Price         float64
	BucketId      int
}

func (t *TransactionEdit) Valid() map[string]string {
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

	if t.BucketId < 0 {
		problems["BucketId"] = "Invalid BucketId"
	}

	return problems
}

type TransactionFilters struct {
	Name     *string
	Price    *float64
	Month    *int
	Year     *int
	BucketId *int
}
