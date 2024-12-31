package services

import (
	"log/slog"
	"wonk/app/auth"
	"wonk/app/database"
	"wonk/app/secret"
)

type Services struct {
	Auth auth.AuthService
}

func InitServices(secrets *secret.Secret, l *slog.Logger, db database.Database) (*Services, error) {
	a := auth.InitAuthService(secrets, l, db)

	s := Services{
		Auth: a,
	}
	return &s, nil
}
