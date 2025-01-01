package database

import (
	"database/sql"
	"errors"
	"fmt"
	"wonk/app/cuserr"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const (
	FILE_NAME                    = "wonk.db"
	USER_TABLE_NAME              = "user"
	BUCKETS_TABLE_NAME           = "bucket"
	TRANSACTION_ITEMS_TABLE_NAME = "transaction_item"
)

type Database interface {
	Login(string, string) (int, error)
	CreateUser(string, string) (int, error)
	CreateBucket(int, string) (int, error)
	CreateItemTransaction(TransactionItemInput) (int, error)
	UserBuckets(int) ([]Bucket, error)
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

func (s *SqliteDb) Login(username, password string) (int, error) {
	query := "SELECT * FROM " + USER_TABLE_NAME + " WHERE username=?"
	rows, err := s.Db.Query(query, username)
	if err != nil {
		return -1, fmt.Errorf("Login: Exec: %w", err)
	}
	defer rows.Close()

	var data []User
	for rows.Next() {
		b := User{}
		err := rows.Scan(&b.Id, &b.UserName, &b.Password)
		if err != nil {
			return -1, fmt.Errorf("Login: rows next: %w", err)
		}
		data = append(data, b)
	}
	if len(data) > 1 {
		return -1, errors.New("Login: More than 2 users found! should not be possible")
	}
	if len(data) == 0 {
		return -1, fmt.Errorf("Login: %w", &cuserr.NotFound{})
	}
	curUser := data[0]

	err = bcrypt.CompareHashAndPassword([]byte(curUser.Password), []byte(password))
	if err != nil {
		return -1, fmt.Errorf("Login: pass: %w", &cuserr.InvalidCred{})
	}

	return curUser.Id, nil
}

func (s *SqliteDb) CreateUser(username, password string) (int, error) {
	// TODO: This logic to check for existing username should exist in service layer
	tempColName := "num_users"
	lookupQuery := "SELECT COUNT(*) AS " + tempColName + " FROM " + USER_TABLE_NAME + " WHERE username=?"
	row := s.Db.QueryRow(lookupQuery, username)
	curUser := struct{ NumUsers int }{}
	err := row.Scan(&curUser.NumUsers)
	if err != nil {
		return -1, fmt.Errorf("CreateUser: numUser: %w", err)
	}
	if curUser.NumUsers > 0 {
		return -1, fmt.Errorf("CreateUser: username already exists")
	}
	// TODO: Add salt to password

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return -1, fmt.Errorf("CreateUser: %w", err)
	}

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
	// TODO: Limit number of buckets a user can have, this logic should live in service layer!
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
	query := "INSERT INTO " + TRANSACTION_ITEMS_TABLE_NAME + " (name, month, year, price, user_id, bucket_id) VALUES (?, ?, ?, ?, ?, ?);"
	res, err := s.Db.Exec(query, input.Name, input.Month, input.Year, input.Price, input.UserId, input.BucketId)
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
