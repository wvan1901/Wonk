package business

import (
	"log/slog"
	"wonk/app/secret"
	"wonk/business/finance"
	"wonk/business/user"
	"wonk/storage"
)

type Services struct {
	Finance finance.Finance
	User    user.User
}

func InitServices(secrets *secret.Secret, l *slog.Logger, db database.Database) (*Services, error) {
	u := user.InitUserService(db)
	f := finance.InitFinance(db)

	s := Services{
		Finance: f,
		User:    u,
	}
	return &s, nil
}
