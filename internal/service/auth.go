package service

import (
	"context"
	"crypto/sha256"
	"encoding/base64"

	"github.com/kbannyi/gophermart/internal/domain"
)

type UserRepository interface {
	Save(ctx context.Context, u domain.User) error
	Get(ctx context.Context, login string, pass string) (*domain.User, error)
}

type AuthService struct {
	repository UserRepository
}

func NewAuthService(r UserRepository) AuthService {
	return AuthService{repository: r}
}

func (s AuthService) Register(ctx context.Context, login string, pass string) (*domain.User, error) {
	hpass := hashPassword(pass)
	user := domain.NewUser(login, hpass)
	err := s.repository.Save(ctx, user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (s AuthService) Login(ctx context.Context, login string, pass string) (*domain.User, error) {
	hpass := hashPassword(pass)
	user, err := s.repository.Get(ctx, login, hpass)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func hashPassword(pass string) string {
	bpass := []byte(pass)
	hasher := sha256.New()
	hasher.Write(bpass)

	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}
