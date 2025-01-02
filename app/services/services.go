package services

import (
	"log/slog"
	"wonk/app/auth"
	"wonk/app/database"
	"wonk/app/secret"
	"wonk/app/services/finance"
	"wonk/app/services/user"
)

type Services struct {
	Auth    auth.AuthService
	Finance finance.Finance
	User    user.User
}

func InitServices(secrets *secret.Secret, l *slog.Logger, db database.Database) (*Services, error) {
	u := user.InitUserService(db)
	a := auth.InitAuthService(secrets, l, u)
	f := finance.InitFinance(db)

	s := Services{
		Auth:    a,
		Finance: f,
		User:    u,
	}
	return &s, nil
}
