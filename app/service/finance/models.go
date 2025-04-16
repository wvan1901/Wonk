package finance

type TransactionNewInput struct {
	Name      string
	Month     string
	Year      string
	Price     string
	IsExpense string
	BucketId  string
	UserId    int
}

type TransactionEditInput struct {
	TransactionId int
	Name          string
	Month         string
	Year          string
	Price         string
	BucketId      string
}

type TransactionFilter struct {
	Name     string
	Price    string
	Month    string
	Year     string
	BucketId string
}
