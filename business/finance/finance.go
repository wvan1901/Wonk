package finance

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"wonk/storage"
)

const (
	MAX_BUCKETS = 40
)

type Finance interface {
	UserBuckets(int) ([]database.Bucket, error)
	SubmitNewTransaction(database.TransactionItemInput) (map[string]string, error)
	CreateBucket(int, string) (map[string]string, error)
	MonthlySummary(int, int, int) (*MonthSummary, error)
	GetBucket(string) (*database.Bucket, error)
	UpdateBucket(int, string) error
	GetTransactions(int, int, int, string, bool) ([]database.TransactionItem, error)
	GetTransaction(string) (*database.TransactionItem, error)
	UpdateTransaction(TransactionEdit) error
	DeleteTransaction(int) error
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

func (f *FinanceLogic) SubmitNewTransaction(inputForm database.TransactionItemInput) (map[string]string, error) {
	// Validate input values
	problems := inputForm.Valid()
	if len(problems) > 0 {
		return problems, nil
	}

	// Save to DB
	_, err := f.DB.CreateItemTransaction(inputForm)
	if err != nil {
		return nil, fmt.Errorf("SubmitNewTransaction: db: %w", err)
	}

	return nil, nil
}

func (f *FinanceLogic) CreateBucket(userId int, newName string) (map[string]string, error) {
	numBuckets, err := f.DB.NumBuckets(userId)
	if err != nil {
		return nil, fmt.Errorf("CreateBucket: num: %w", err)
	}
	if numBuckets >= MAX_BUCKETS {
		return nil, errors.New("CreateBucket: user can't have more buckets")
	}

	problems := make(map[string]string)
	if len(newName) == 0 {
		problems["Name"] = "Name value must not be empty"
	}
	if len(newName) > 20 {
		problems["Name"] = "Name value must not be greater than 20 characters"
	}
	if len(problems) > 0 {
		return problems, nil
	}
	_, err = f.DB.CreateBucket(userId, newName)
	if err != nil {
		return nil, fmt.Errorf("CreateBucket: db: %w", err)
	}
	return nil, nil
}

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

func (f *FinanceLogic) MonthlySummary(userId, month, year int) (*MonthSummary, error) {
	buckets, err := f.DB.UserBuckets(userId)
	if err != nil {
		return nil, fmt.Errorf("BucketsMonthlySummary: %w", err)
	}

	totalIncome := 0.0
	totalExpense := 0.0

	newBuckets := []BucketSummary{}
	for _, b := range buckets {
		totalPrice, err := f.bucketMonthPrice(b.Id, month, year)
		if err != nil {
			return nil, fmt.Errorf("BucketsMonthlySummary: %w", err)
		}
		newB := BucketSummary{
			Reference: b,
			Price:     totalPrice,
		}
		newBuckets = append(newBuckets, newB)
		if totalPrice < 0 {
			totalExpense += totalPrice
		} else {
			totalIncome += totalPrice
		}
	}

	summary := &MonthSummary{
		BucketsSummary: newBuckets,
		TotalIncome:    totalIncome,
		TotalExpense:   totalExpense,
	}

	return summary, nil
}

func (f *FinanceLogic) bucketMonthPrice(bucketId int, month int, year int) (float64, error) {
	transactions, err := f.DB.TransactionsInBucket(bucketId, month, year)
	if err != nil {
		return 0, fmt.Errorf("bucketMonthPrice: db: %w", err)
	}

	// Get the price of all
	totalPrice := 0.0
	for _, t := range transactions {
		factor := 1
		if t.IsExpense {
			factor = factor * -1
		}
		totalPrice += t.Price * float64(factor)
	}

	return totalPrice, nil
}

func (f *FinanceLogic) GetBucket(bucketId string) (*database.Bucket, error) {
	id, err := strconv.Atoi(bucketId)
	if err != nil {
		return nil, fmt.Errorf("GetBucket: invalid id: %w", err)
	}
	bucket, err := f.DB.BucketById(id)
	if err != nil {
		return nil, fmt.Errorf("GetBucket: %w", err)
	}
	return bucket, nil
}

func (f *FinanceLogic) UpdateBucket(bucketId int, newName string) error {
	rowsChanged, err := f.DB.BucketUpdateName(bucketId, newName)
	if err != nil {
		return fmt.Errorf("UpdateBucket: %w", err)
	}
	if rowsChanged == 0 {
		return errors.New("UpdateBucket: no data changed")
	}
	return nil
}
func (f *FinanceLogic) GetTransactions(page, pagesize, userId int, sortBy string, isAscending bool) ([]database.TransactionItem, error) {
	transactions, err := f.DB.TransactionsPagination(page, pagesize, userId, sortBy, isAscending)
	if err != nil {
		return nil, fmt.Errorf("GetTransactions: %w", err)
	}
	return transactions, nil
}
func (f *FinanceLogic) GetTransaction(transactionId string) (*database.TransactionItem, error) {
	id, err := strconv.Atoi(transactionId)
	if err != nil {
		return nil, fmt.Errorf("GetTransaction: invalid id: %w", err)
	}
	transaction, err := f.DB.TransactionById(id)
	if err != nil {
		return nil, fmt.Errorf("GetTransaction: %w", err)
	}
	return transaction, nil
}
func (f *FinanceLogic) UpdateTransaction(input TransactionEdit) error {
	// Validate fields
	problems := input.Valid()
	if len(problems) > 0 {
		return fmt.Errorf("UpdateTransaction: input problems: %v", problems)
	}
	// Update transaction in db
	rowsChanged, err := f.DB.TransactionUpdate(input.Name, input.TransactionId, input.BucketId, input.Month, input.Year, input.Price)
	if err != nil {
		return fmt.Errorf("UpdateTransaction: db: %w", err)
	}
	if rowsChanged == 0 {
		return errors.New("UpdateTransaction: db: no data changed")
	}
	return nil
}

func (f *FinanceLogic) DeleteTransaction(transactionId int) error {
	rowsChanged, err := f.DB.TransactionDelete(transactionId)
	if err != nil {
		return fmt.Errorf("DeleteTransaction: db: %w", err)
	}
	if rowsChanged == 0 {
		return errors.New("DeleteTransaction: db: no data changed")
	}

	return nil
}
