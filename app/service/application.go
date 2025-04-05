package application

import (
	"log/slog"
	"wonk/app/auth"
	"wonk/app/secret"
	"wonk/business"
)

type Service struct {
	Auth auth.AuthService
}

func InitServices(secrets *secret.Secret, l *slog.Logger, b *business.Services) (*Service, error) {
	a := auth.InitAuthService(secrets, l, b.User)

	s := Service{
		Auth: a,
	}

	return &s, nil
}
