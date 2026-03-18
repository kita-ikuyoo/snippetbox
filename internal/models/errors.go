package models

import (
	"errors"
)

var ErrNoRecord = errors.New("models: no matching record found")

// Add a new ErrInvalidCredentials error. We'll use this later if a user
// tries to log in with an incorrect email address or password.
var ErrInvalidCredentials = errors.New("models: invalid credentials")

// Add a new ErrDuplicateEmail error. We'll use this later if a user
// tries to sign up with an email address that's already in use.
var ErrDuplicateEmail = errors.New("models: duplicate email")
