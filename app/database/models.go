package database

type TransactionItemInput struct {
	Name     string
	Month    int
	Year     int
	Price    float64
	UserId   int
	BucketId int
}
