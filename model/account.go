// Package model defines the models used for the web application.
package model

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"io"

	"appengine"
	"appengine/datastore"
)

type Provider struct {
	Name string
	Id   string
}

// Account stores an account for an user of the application, login-wise.
type Account struct {
	key       *datastore.Key `datastore:"-"`
	Name      string         `datastore:"name"`
	Email     string         `datastore:"email"`
	Active    bool           `datastore:"active"`
	Hash      []byte         `datastore:"hash,noindex"`
	Salt      []byte         `datastore:"salt,noindex"`
	Providers []Provider
}

const AccountModelName = "account"

const saltSize = 16

var InvalidEmailErr = errors.New("invalid email address")

// NewAccount creates new user with given name and email.
// Returns an error if the name and email are empty.
func NewAccount(name, email string) (*Account, error) {
	if name == "" {
		return nil, errors.New("must provide valid name for user.")
	}

	return &Account{
		Name:   name,
		Email:  email,
		Active: false,
	}, nil
}

// SetPassword sets a salt and the Sha-256 password for the
// provided password.
func (p *Account) SetPassword(pass string) error {
	salt := make([]byte, saltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return err
	}

	h := sha256.New()
	h.Write(salt)
	h.Write([]byte(pass))

	p.Hash = h.Sum(nil)
	p.Salt = salt

	return nil
}

// Check password returns whether the provided argument
// matches the set password for the user.
func (p *Account) CheckPassword(pass string) bool {
	if p.Salt == nil || p.Hash == nil {
		return false
	}

	h := sha256.New()
	h.Write(p.Salt)
	h.Write([]byte(pass))

	return bytes.Equal(h.Sum(nil), p.Hash)
}

// HasProvider checks whether the user is associated
// with the provider given as an argument.
func (p *Account) HasProvider(providerName string) bool {
	if p.Providers == nil {
		return false
	}

	for _, prov := range p.Providers {
		if prov.Name == providerName {
			return true
		}
	}
	return false
}

// AddProvider adds a new provider for the user.
// It takes the provider name and the user id for that provider.
// Returns an error if the user is already associated with the provider.
func (p *Account) AddProvider(providerName, providerUserId string) error {
	if providerName == "" {
		return errors.New("provider name must be valid")
	}
	if providerUserId == "" {
		return errors.New("user id for provider must not be empty")
	}
	if p.HasProvider(providerName) {
		return errors.New("user is already registered with this provider")
	}

	if p.Providers == nil {
		p.Providers = make([]Provider, 1)
	}
	p.Providers[0] = Provider{Name: providerName, Id: providerUserId}

	return nil
}

func convertToAccount(t interface{}) (*Account, error) {
	if account, ok := t.(*Account); !ok {
		return nil, errors.New("unable to convert to Account type")
	} else {
		return account, nil
	}
}

func (p *Account) Add(c appengine.Context) (*datastore.Key, error) {
	if p.key != nil {
		return nil, errors.New("account already in datastore")
	}
	key := datastore.NewIncompleteKey(c, AccountModelName, nil)
	p.key = key
	return datastore.Put(c, key, p)
}

func (p *Account) GetById(c appengine.Context, encodedId string) (*datastore.Key, *Account, error) {
	if key, model, err := getById(c, encodedId, AccountModelName); model != nil {
		account, err := convertToAccount(model)
		account.key = key
		return key, account, err
	} else {
		return key, nil, err
	}
}

func (p *Account) GetByEmail(c appengine.Context, email string) (*datastore.Key, *Account, error) {
	if key, model, err := getByEmail(c, email, AccountModelName); model != nil {
		account, err := convertToAccount(model)
		account.key = key
		return key, account, err
	} else {
		return key, nil, err
	}
}

func (p *Account) GetByAncestorKey(c appengine.Context, ancestorKey *datastore.Key) (*datastore.Key, *Account, error) {
	if key, model, err := getByAncestorKey(c, ancestorKey, AccountModelName); model != nil {
		account, err := convertToAccount(model)
		account.key = key
		return key, account, err
	} else {
		return key, nil, err
	}
}

func (p *Account) Update(c appengine.Context) (*datastore.Key, error) {
	key, err := datastore.Put(c, p.key, p)
	if key != nil {
		p.key = key
	}
	return key, err
}

func (p *Account) Delete(c appengine.Context) error {
	return datastore.Delete(c, p.key)
}

func (p *Account) Key() *datastore.Key {
	return p.key
}
