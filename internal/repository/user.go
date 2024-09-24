package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/kbannyi/gophermart/internal/domain"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) UserRepository {
	return UserRepository{db: db}
}

func (r UserRepository) Save(ctx context.Context, u domain.User) error {
	_, err := r.db.NamedExecContext(ctx, "INSERT INTO users (id, login, password) VALUES (:id, :login, :password);", u)
	if err != nil {
		return convertErr(err)
	}

	return nil
}

func (r UserRepository) Get(ctx context.Context, login string, pass string) (*domain.User, error) {
	user := domain.User{}
	err := r.db.GetContext(ctx, &user, `SELECT * FROM users WHERE login = $1 and password = $2`, login, pass)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}
