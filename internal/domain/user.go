package domain

import "github.com/google/uuid"

type User struct {
	ID       string `db:"id"`
	Login    string `db:"login"`
	Password string `db:"password"`
}

func NewUser(login string, pass string) User {
	return User{
		ID:       uuid.New().String(),
		Login:    login,
		Password: pass,
	}
}
