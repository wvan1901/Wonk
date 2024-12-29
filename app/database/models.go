package database

import (
	"fmt"
	"math"
)

type TransactionItemInput struct {
	Name     string
	Month    int
	Year     int
	Price    float64
	UserId   int
	BucketId int
}

type Bucket struct {
	Id     int
	Name   string
	UserId int
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
		fmt.Println("Wicho: DEBUG:", math.Floor(t.Price*100), "==", 100*t.Price, twoOrLessDecimalPlaces)
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
