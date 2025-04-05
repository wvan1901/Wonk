package finance

type TransactionFormInput struct {
	Name      string
	Month     string
	Year      string
	Price     string
	IsExpense string
	BucketId  string
	UserId    int
}
