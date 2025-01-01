package user

import (
	"errors"
	"fmt"
	"wonk/app/cuserr"
	"wonk/app/database"

	"golang.org/x/crypto/bcrypt"
)

type User interface {
	Login(string, string) (int, error)
	CreateUser(string, string) (int, error)
}

type UserLogic struct {
	DB database.Database
}

func InitUserService(db database.Database) User {
	return &UserLogic{
		DB: db,
	}
}

func (u *UserLogic) Login(userName, password string) (int, error) {
	curUser, err := u.DB.UserByUserName(userName)
	if err != nil {
		return -1, fmt.Errorf("Login: UserLogic: %w", err)
	}

	// Compare the input password to hashed password in DB
	err = bcrypt.CompareHashAndPassword([]byte(curUser.Password), []byte(password))
	if err != nil {
		return -1, fmt.Errorf("Login: password: %w", &cuserr.InvalidCred{})
	}

	return curUser.Id, nil
}

func (u *UserLogic) CreateUser(userName, password string) (int, error) {
	// Check If username exist, if so then return err
	_, err := u.DB.UserByUserName(userName)
	if err != nil {
		if errors.Is(err, &cuserr.NotFound{}) {
		} else {
			return -1, fmt.Errorf("CreateUser: %w", err)
		}
	} else {
		return -1, fmt.Errorf("CreateUser: User Found: %w", err)
	}

	// Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return -1, fmt.Errorf("CreateUser: %w", err)
	}

	// Save new User to DB
	userId, err := u.DB.CreateUser(userName, string(hashedPassword))
	if err != nil {
		return -1, fmt.Errorf("CreateUser: db: %w", err)
	}
	return userId, nil
}
