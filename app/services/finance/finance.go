package finance

import (
	"fmt"
	"strconv"
	"wonk/app/database"
)

type Finance interface {
	UserBuckets(int) ([]database.Bucket, error)
	SubmitNewTransaction(TransactionFormInput) (map[string]string, error)
}

type FinanceLogic struct {
	DB database.Database
}

func InitFinance(db database.Database) Finance {
	return &FinanceLogic{
		DB: db,
	}
}

func (f *FinanceLogic) UserBuckets(userId int) ([]database.Bucket, error) {
	buckets, err := f.DB.UserBuckets(userId)
	if err != nil {
		return nil, fmt.Errorf("UserBuckets: %w", err)
	}
	return buckets, nil
}

func (f *FinanceLogic) SubmitNewTransaction(inputForm TransactionFormInput) (map[string]string, error) {
	// Convert form (multiple strings) to valid db type
	conversionProblems := make(map[string]string)
	month, err := strconv.Atoi(inputForm.Month)
	if err != nil {
		conversionProblems["Month"] = "Invalid Month: Not a number"
	}
	year, err := strconv.Atoi(inputForm.Year)
	if err != nil {
		conversionProblems["Year"] = "Invalid Year: Not a number"
	}
	price, err := strconv.ParseFloat(inputForm.Price, 64)
	if err != nil {
		conversionProblems["Price"] = "Invalid Price: Not a decimal"
	}
	bucketId, err := strconv.Atoi(inputForm.BucketId)
	if err != nil {
		conversionProblems["BucketId"] = "Invalid BucketId: Not a number"
	}
	if len(conversionProblems) > 0 {
		return conversionProblems, nil
	}
	// Validate DB input
	transactionInput := database.TransactionItemInput{
		Name:     inputForm.Name,
		Month:    month,
		Year:     year,
		Price:    price,
		UserId:   inputForm.UserId,
		BucketId: bucketId,
	}
	problems := transactionInput.Valid()
	if len(problems) > 0 {
		return problems, nil
	}

	// Save to DB
	_, err = f.DB.CreateItemTransaction(transactionInput)
	if err != nil {
		return nil, fmt.Errorf("SubmitNewTransaction: db: %w", err)
	}

	return nil, nil
}

type TransactionFormInput struct {
	Name     string
	Month    string
	Year     string
	Price    string
	UserId   int
	BucketId string
}
