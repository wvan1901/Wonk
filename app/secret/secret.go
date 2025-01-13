package secret

import (
	"errors"
	"fmt"
)

type Secret struct {
	CookieKey string // Hex string
	JwtKey    string // Random string
}

func (s *Secret) Valid() error {
	if s == nil {
		return errors.New("Secret is nil")
	}
	if s.CookieKey == "" {
		return errors.New("Secret: cookie key is empty")
	}
	if s.JwtKey == "" {
		return errors.New("Secret: jwt key is empty")
	}
	return nil
}

func InitSecret(getEnv func(string) string) (*Secret, error) {
	s := Secret{
		CookieKey: getEnv("COOKIE_SECRET_KEY"),
		JwtKey:    getEnv("JWT_SECRET_KEY"),
	}
	err := s.Valid()
	if err != nil {
		return nil, fmt.Errorf("InitSecret: %w", err)
	}
	return &s, nil
}
