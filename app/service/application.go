package application

import (
	"log/slog"
	"wonk/app/auth"
	"wonk/app/secret"
	"wonk/app/service/finance"
	"wonk/business"
)

type Service struct {
	Auth    auth.AuthService
	Finance *finance.FinanceService
}

func InitServices(secrets *secret.Secret, l *slog.Logger, b *business.Services) (*Service, error) {
	a := auth.InitAuthService(secrets, l, b.User)
	f := finance.InitFinanceService(l, b.Finance)

	s := Service{
		Auth:    a,
		Finance: f,
	}

	return &s, nil
}
