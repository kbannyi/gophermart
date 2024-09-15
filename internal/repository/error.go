package repository

import "errors"

var ErrAlreadyExists = errors.New("already exists")
var ErrNotFound = errors.New("values not found")
