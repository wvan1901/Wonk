package services

import (
	"log/slog"
	"wonk/app/auth"
	"wonk/app/database"
	"wonk/app/secret"
	"wonk/app/services/user"
)

type Services struct {
	Auth auth.AuthService
	User user.User
}

func InitServices(secrets *secret.Secret, l *slog.Logger, db database.Database) (*Services, error) {
	u := user.InitUserService(db)
	a := auth.InitAuthService(secrets, l, u)

	s := Services{
		Auth: a,
		User: u,
	}
	return &s, nil
}
