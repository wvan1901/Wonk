package database

import (
	"math"
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
	twoOrLessDecimalPlaces := math.Floor(t.Price*100) == 100*t.Price
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
