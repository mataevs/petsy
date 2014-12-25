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

type provider struct {
	name string
	id   string
}

// User stores a user of the application, login-wise.
type User struct {
	Name      string `datastore:"name"`
	Email     string `datastore:"email"`
	AvatarURL string `datastore:"avatar,noindex"`
	Active    bool
	hash      []byte `datastore:"hash,noindex"`
	salt      []byte `datastore:"salt,noindex"`
	providers []provider
}

const saltSize = 16

// NewUser creates new user with given name and email.
// Returns an error if the name and email are empty.
func NewUser(name, email string) (*User, error) {
	if name == "" {
		return nil, errors.New("must provide valid name for user.")
	}
	if email == "" {
		// todo - check valid email
		return nil, errors.New("must provide valid email for user.")
	}

	return &User{
		Name:  name,
		Email: email,
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

	u.hash = h.Sum(nil)
	u.salt = salt

	return nil
}

// Check password returns whether the provided argument
// matches the set password for the user.
func (u *User) CheckPassword(pass string) bool {
	if u.salt == nil || u.hash == nil {
		return false
	}

	h := sha256.New()
	h.Write(u.salt)
	h.Write([]byte(pass))

	return bytes.Equal(h.Sum(nil), u.hash)
}

// HasProvider checks whether the user is associated
// with the provider given as an argument.
func (u *User) HasProvider(providerName string) bool {
	if u.providers == nil {
		return false
	}

	for _, prov := range u.providers {
		if prov.name == providerName {
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

	if u.providers == nil {
		u.providers = make([]provider, 1)
	}
	u.providers[0] = provider{name: providerName, id: providerUserId}

	return nil
}

// Merge combines 2 user structures - the old one (the reference) and
// the new one (provided as an argument). The values provided by the
// new user take precedence over the old values.
func (u *User) Merge(newUser *User) *User {
	mergedUser := new(User)

	*mergedUser = *u

	if newUser.Name != "" {
		mergedUser.Name = newUser.Name
	}
	if newUser.Email != "" {
		mergedUser.Email = newUser.Email
	}
	if newUser.AvatarURL != "" {
		mergedUser.AvatarURL = newUser.AvatarURL
	}
	if newUser.providers != nil {
		for _, prov := range newUser.providers {
			mergedUser.AddProvider(prov.name, prov.id)
		}
	}
	return mergedUser
}
