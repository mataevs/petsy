package user

import (
	"testing"
)

const (
	name    = "test"
	email   = "test@petsy.ro"
	pass    = "password"
	badPass = "pwned"
)

var (
	provider = Provider{
		Name: "provider",
		Id:   "1234",
	}
)

func TestNewUser(t *testing.T) {
	user, err := NewUser(name, email)
	if err != nil {
		t.Errorf("unexpected error creating user: %v", err)
	}
	user.AddProvider(provider.Name, provider.Id)

	if user.Name != name {
		t.Errorf("error creating user: name: got %s; want %s", user.Name, name)
	}
	if user.Email != email {
		t.Errorf("error creating user: email: got %s; want %s", user.Email, email)
	}
	if !user.HasProvider(provider.Name) {
		t.Errorf("error creating user: does not have provider %s", provider.Name)
	}
}

func TestUserPassword(t *testing.T) {
	user, _ := NewUser(name, email)
	user.SetPassword(pass)

	if user.CheckPassword(badPass) {
		t.Errorf("error testing user: bad password works")
	}
	if !user.CheckPassword(pass) {
		t.Errorf("error testing user: good password does not work")
	}
}
