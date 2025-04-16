package database

import (
	"database/sql"
	"fmt"
	"wonk/app/cuserr"

	_ "github.com/mattn/go-sqlite3"
)

const (
	FILE_NAME                    = "wonk.db"
	USER_TABLE_NAME              = "user"
	BUCKETS_TABLE_NAME           = "bucket"
	TRANSACTION_ITEMS_TABLE_NAME = "transaction_item"
)

type Database interface {
	CreateUser(string, string) (int, error)
	CreateBucket(int, string) (int, error)
	CreateItemTransaction(TransactionItemInput) (int, error)
	UserBuckets(int) ([]Bucket, error)
	UserByUserName(string) (*User, error)
	NumBuckets(int) (int, error)
	TransactionsInBucket(int, int, int) ([]TransactionItem, error)
	BucketById(int) (*Bucket, error)
	BucketUpdateName(int, string) (int64, error)
	TransactionsPagination(int, int, string, bool, TransactionFilters) ([]TransactionItem, error)
	TransactionById(int) (*TransactionItem, error)
	TransactionUpdate(string, int, int, int, int, float64) (int64, error)
	TransactionDelete(int) (int64, error)
}

type SqliteDb struct {
	Db *sql.DB
}

func InitDb() (Database, error) {
	db, err := sql.Open("sqlite3", FILE_NAME)
	if err != nil {
		return nil, fmt.Errorf("InitDb: %w", err)
	}

	return &SqliteDb{Db: db}, nil
}

func (s *SqliteDb) UserByUserName(username string) (*User, error) {
	// User table has a unique constraint on username column
	query := "SELECT * FROM " + USER_TABLE_NAME + " WHERE username=?"
	row := s.Db.QueryRow(query, username)
	curUser := User{}
	err := row.Scan(&curUser.Id, &curUser.UserName, &curUser.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("UserByUserName: %w", &cuserr.NotFound{})
		}
		return nil, fmt.Errorf("UserByUserName: %w", err)
	}
	return &curUser, nil

}

func (s *SqliteDb) CreateUser(username, hashedPassword string) (int, error) {
	query := "INSERT INTO " + USER_TABLE_NAME + " (username, password) VALUES (?, ?);"
	res, err := s.Db.Exec(query, username, hashedPassword)
	if err != nil {
		return 0, fmt.Errorf("CreateUser: Exec: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		fmt.Println("CreateUser: insert Id: %w")
	}
	return int(id), nil
}

func (s *SqliteDb) CreateBucket(userId int, bucketName string) (int, error) {
	query := "INSERT INTO " + BUCKETS_TABLE_NAME + " (name, user_id) VALUES (?, ?);"
	res, err := s.Db.Exec(query, bucketName, userId)
	if err != nil {
		return 0, fmt.Errorf("CreateBucket: Exec: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		fmt.Println("CreateBucket: insert Id: %w")
	}
	return int(id), nil
}

func (s *SqliteDb) CreateItemTransaction(input TransactionItemInput) (int, error) {
	query := "INSERT INTO " + TRANSACTION_ITEMS_TABLE_NAME + " (name, month, year, price, is_expense, user_id, bucket_id) VALUES (?, ?, ?, ?, ?, ?, ?);"
	res, err := s.Db.Exec(query, input.Name, input.Month, input.Year, input.Price, input.IsExpense, input.UserId, input.BucketId)
	if err != nil {
		return 0, fmt.Errorf("CreateItemTransaction: Exec: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		fmt.Println("CreateItemTransaction: insert Id: %w")
	}
	return int(id), nil
}
func (s *SqliteDb) UserBuckets(userId int) ([]Bucket, error) {
	query := "SELECT * FROM " + BUCKETS_TABLE_NAME + " WHERE user_id=?"
	rows, err := s.Db.Query(query, userId)
	if err != nil {
		return nil, fmt.Errorf("UserBuckets: Exec: %w", err)
	}
	defer rows.Close()

	var data []Bucket
	for rows.Next() {
		b := Bucket{}
		err := rows.Scan(&b.Id, &b.Name, &b.UserId)
		if err != nil {
			return nil, fmt.Errorf("UserBuckets: rows next: %w", err)
		}
		data = append(data, b)
	}

	return data, nil
}

func (s *SqliteDb) NumBuckets(userId int) (int, error) {
	tempColName := "num"
	lookupQuery := "SELECT COUNT(*) AS " + tempColName + " FROM " + BUCKETS_TABLE_NAME + " WHERE user_id=?"
	row := s.Db.QueryRow(lookupQuery, userId)
	numBuckets := struct{ Num int }{}
	err := row.Scan(&numBuckets.Num)
	if err != nil {
		return -1, fmt.Errorf("NumBuckets: %w", err)
	}

	return numBuckets.Num, nil
}

func (s *SqliteDb) TransactionsInBucket(bucketId, month, year int) ([]TransactionItem, error) {
	query := "SELECT * FROM " + TRANSACTION_ITEMS_TABLE_NAME + " WHERE bucket_id=? AND month=? AND year=?"
	rows, err := s.Db.Query(query, bucketId, month, year)
	if err != nil {
		return nil, fmt.Errorf("TransactionsInBucket: Exec: %w", err)
	}
	defer rows.Close()

	var data []TransactionItem
	for rows.Next() {
		b := TransactionItem{}
		err := rows.Scan(&b.Id, &b.Name, &b.Month, &b.Year, &b.Price, &b.IsExpense, &b.UserId, &b.BucketId)
		if err != nil {
			return nil, fmt.Errorf("TransactionsInBucket: rows next: %w", err)
		}
		data = append(data, b)
	}

	return data, nil
}
func (s *SqliteDb) BucketById(bucketId int) (*Bucket, error) {
	query := "SELECT * FROM " + BUCKETS_TABLE_NAME + " WHERE id=?"
	row := s.Db.QueryRow(query, bucketId)
	curBucket := Bucket{}
	err := row.Scan(&curBucket.Id, &curBucket.Name, &curBucket.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("UserByUserName: %w", &cuserr.NotFound{})
		}
		return nil, fmt.Errorf("UserByUserName: %w", err)
	}

	return &curBucket, nil
}

func (s *SqliteDb) BucketUpdateName(bucketId int, newName string) (int64, error) {
	query := "UPDATE " + BUCKETS_TABLE_NAME + " SET name=? WHERE id=?"
	result, err := s.Db.Exec(query, newName, bucketId)
	if err != nil {
		return 0, fmt.Errorf("BucketUpdateName: %w", err)
	}

	return result.RowsAffected()
}

func (s *SqliteDb) TransactionsPagination(page, pagesize int, sortBy string, isAscending bool, filters TransactionFilters) ([]TransactionItem, error) {
	// Pagination
	if page < 1 {
		page = 1
	}
	offset := max(0, pagesize*(page-1))
	// Sorting
	orderByQuery := "ORDER BY id"
	if sortBy != "" {
		orderDirection := " ASC"
		if !isAscending {
			orderDirection = " DESC"
		}
		orderByQuery = "ORDER BY " + sortBy + orderDirection
	}
	// Filtering
	filter, values := filters.FilterQueryAndValues()

	// Query
	queryValues := values
	queryValues = append(queryValues, pagesize, offset)
	query := "SELECT * FROM " + TRANSACTION_ITEMS_TABLE_NAME + " " + filter + " " + orderByQuery + " LIMIT ? OFFSET ?"
	rows, err := s.Db.Query(query, queryValues...)
	if err != nil {
		return nil, fmt.Errorf("TransactionsPagination: Exec: %w", err)
	}
	defer rows.Close()

	var data []TransactionItem
	for rows.Next() {
		b := TransactionItem{}
		err := rows.Scan(&b.Id, &b.Name, &b.Month, &b.Year, &b.Price, &b.IsExpense, &b.UserId, &b.BucketId)
		if err != nil {
			return nil, fmt.Errorf("TransactionsPagination: rows next: %w", err)
		}
		data = append(data, b)
	}

	return data, nil
}

func (s *SqliteDb) TransactionById(transactionId int) (*TransactionItem, error) {
	query := "SELECT * FROM " + TRANSACTION_ITEMS_TABLE_NAME + " WHERE id=?"
	row := s.Db.QueryRow(query, transactionId)
	t := TransactionItem{}
	err := row.Scan(&t.Id, &t.Name, &t.Month, &t.Year, &t.Price, &t.IsExpense, &t.UserId, &t.BucketId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("TransactionById: %w", &cuserr.NotFound{})
		}
		return nil, fmt.Errorf("TransactionById: %w", err)
	}

	return &t, nil
}

func (s *SqliteDb) TransactionUpdate(name string, transactionId int, bucketId int, month int, year int, price float64) (int64, error) {
	query := "UPDATE " + TRANSACTION_ITEMS_TABLE_NAME + " SET name=?, month=?, year=?, price=?, bucket_id=? WHERE id=?"
	result, err := s.Db.Exec(query, name, month, year, price, bucketId, transactionId)
	if err != nil {
		return 0, fmt.Errorf("TransactionUpdate: %w", err)
	}

	return result.RowsAffected()
}

func (s *SqliteDb) TransactionDelete(transactionId int) (int64, error) {
	query := "DELETE FROM " + TRANSACTION_ITEMS_TABLE_NAME + " WHERE id=?"
	result, err := s.Db.Exec(query, transactionId)
	if err != nil {
		return 0, fmt.Errorf("TransactionDelete: %w", err)
	}

	return result.RowsAffected()
}
