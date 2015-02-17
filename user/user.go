// Package user defines the user structure and basic
// operations for user administration.
package user

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"
)

type Provider struct {
	Name string
	Id   string
}

// User stores a user of the application, login-wise.
type User struct {
	Name      string `datastore:"name"`
	Email     string `datastore:"email"`
	AvatarURL string `datastore:"avatar,noindex"`
	Active    bool
	Hash      []byte `datastore:"hash,noindex"`
	Salt      []byte `datastore:"salt,noindex"`
	Providers []Provider
}

const saltSize = 16

var InvalidEmailErr = errors.New("invalid email address")

// NewUser creates new user with given name and email.
// Returns an error if the name and email are empty.
func NewUser(name, email string) (*User, error) {
	if name == "" {
		return nil, errors.New("must provide valid name for user.")
	}
	if email == "" {
		// todo - check valid email
		return nil, InvalidEmailErr
	}

	return &User{
		Name:   name,
		Email:  email,
		Active: false,
	}, nil
}

// SetPassword sets a salt and the Sha-256 password for the
// provided password.
func (u *User) SetPassword(pass string) error {
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}

	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(pass))

	u.Hash = h.Sum(nil)
	u.Salt = salt

	return nil
}

// Check password returns whether the provided argument
// matches the set password for the user.
func (u *User) CheckPassword(pass string) bool {
	if u.Salt == nil || u.Hash == nil {
		return false
	}

	h := sha256.New()
	h.Write(u.Salt)
	h.Write([]byte(pass))

	return bytes.Equal(h.Sum(nil), u.Hash)
}

// HasProvider checks whether the user is associated
// with the provider given as an argument.
func (u *User) HasProvider(providerName string) bool {
	if u.Providers == nil {
		return false
	}

	for _, prov := range u.Providers {
		if prov.Name == providerName {
			return true
		}
	}
	return false
}

// AddProvider adds a new provider for the user.
// It takes the provider name and the user id for that provider.
// Returns an error if the user is already associated with the provider.
func (u *User) AddProvider(providerName, providerUserId string) error {
	if providerName == "" {
		return errors.New("provider name must be valid")
	}
	if providerUserId == "" {
		return errors.New("user id for provider must not be empty")
	}
	if u.HasProvider(providerName) {
		return errors.New("user is already registered with this provider")
	}

	if u.Providers == nil {
		u.Providers = make([]Provider, 1)
	}
	u.Providers[0] = Provider{Name: providerName, Id: providerUserId}

	return nil
}
