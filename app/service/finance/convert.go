package finance

import (
	"strconv"
	"wonk/app/templates/views"
	"wonk/business/finance"
	database "wonk/storage"
)

func parseNewTransaction(t TransactionNewInput) (database.TransactionItemInput, map[string]string) {
	dbModel := database.TransactionItemInput{}
	parseProblems := make(map[string]string)

	month, err := strconv.Atoi(t.Month)
	if err != nil {
		parseProblems["Month"] = "Not a number"
	}
	year, err := strconv.Atoi(t.Year)
	if err != nil {
		parseProblems["Year"] = "Not a number"
	}
	price, err := strconv.ParseFloat(t.Price, 64)
	if err != nil {
		parseProblems["Price"] = "Not a decimal"
	}
	isExpense := false
	switch t.IsExpense {
	case "on":
		isExpense = true
	case "":
		isExpense = false
	default:
		parseProblems["IsExpense"] = "Not valid"
	}
	bucketId, err := strconv.Atoi(t.BucketId)
	if err != nil {
		parseProblems["BucketId"] = "Invalid Id"
	}
	if len(parseProblems) > 0 {
		return dbModel, parseProblems
	}
	// Validate DB input
	dbModel = database.TransactionItemInput{
		Name:      t.Name,
		Month:     month,
		Year:      year,
		Price:     price,
		IsExpense: isExpense,
		UserId:    t.UserId,
		BucketId:  bucketId,
	}
	return dbModel, nil
}

func parseEditTransaction(input TransactionEditInput) (finance.TransactionEdit, map[string]string) {
	businessModel := finance.TransactionEdit{}
	parseProblems := make(map[string]string)
	month, err := strconv.Atoi(input.Month)
	if err != nil {
		parseProblems["Month"] = "Not a number"
	}
	year, err := strconv.Atoi(input.Year)
	if err != nil {
		parseProblems["Year"] = "Not a number"
	}
	price, err := strconv.ParseFloat(input.Price, 64)
	if err != nil {
		parseProblems["Price"] = "Not a decimal"
	}
	bucketId, err := strconv.Atoi(input.BucketId)
	if err != nil {
		parseProblems["BucketId"] = "Invalid Id"
	}
	if len(parseProblems) > 0 {
		return businessModel, parseProblems
	}
	businessModel = finance.TransactionEdit{
		TransactionId: input.TransactionId,
		Name:          input.Name,
		Month:         month,
		Year:          year,
		Price:         price,
		BucketId:      bucketId,
	}
	return businessModel, nil
}

func convertFilters(input TransactionFilter) finance.TransactionFilters {
	var name *string
	var month *int
	var year *int
	var bucketId *int
	var price *float64

	if input.Name != "" {
		name = &input.Name
	}
	parsedMonth, err := strconv.Atoi(input.Month)
	if err == nil {
		month = &parsedMonth
	}
	parsedYear, err := strconv.Atoi(input.Year)
	if err == nil {
		year = &parsedYear
	}
	parsedPrice, err := strconv.ParseFloat(input.Price, 64)
	if err == nil {
		price = &parsedPrice
	}
	parsedBucketId, err := strconv.Atoi(input.BucketId)
	if err == nil {
		bucketId = &parsedBucketId
	}
	return finance.TransactionFilters{
		Name:     name,
		Month:    month,
		Year:     year,
		Price:    price,
		BucketId: bucketId,
	}
}

func convertToFilters(f finance.TransactionFilters) []views.Filter {
	newFilters := []views.Filter{}
	if f.Name != nil {
		nameFilter := views.Filter{
			ColumnName:  "name",
			FilterValue: *f.Name,
		}
		newFilters = append(newFilters, nameFilter)
	}
	if f.Month != nil {
		monthFilter := views.Filter{
			ColumnName:  "month",
			FilterValue: strconv.Itoa(*f.Month),
		}
		newFilters = append(newFilters, monthFilter)
	}
	if f.Year != nil {
		yearFilter := views.Filter{
			ColumnName:  "year",
			FilterValue: strconv.Itoa(*f.Year),
		}
		newFilters = append(newFilters, yearFilter)
	}
	if f.Price != nil {
		priceFilter := views.Filter{
			ColumnName:  "price",
			FilterValue: strconv.FormatFloat(*f.Price, 'f', -1, 64),
		}
		newFilters = append(newFilters, priceFilter)
	}
	if f.BucketId != nil {
		bucketIdFilter := views.Filter{
			ColumnName:  "bucket_id",
			FilterValue: strconv.Itoa(*f.BucketId),
		}
		newFilters = append(newFilters, bucketIdFilter)
	}
	return newFilters
}
