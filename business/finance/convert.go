package finance

import database "wonk/storage"

func convertTransactionFilters(input TransactionFilters) database.TransactionFilters {
	return database.TransactionFilters{
		Name:     input.Name,
		Price:    input.Price,
		Month:    input.Month,
		Year:     input.Year,
		BucketId: input.BucketId,
	}
}
