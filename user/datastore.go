// +build appengine

// Part of user package. Implements operations with user
// structures on the Appengine's datastore.
package user

import (
	"errors"

	"appengine"
	"appengine/datastore"
)

const UserKind = "user"

// AddUser adds a new user to the datastore. Returns the key of the new
// entry and possibly an error.
func AddUser(c appengine.Context, user *User) (*datastore.Key, error) {
	if user == nil {
		return nil, errors.New("user to add to datastore can't be empty")
	}
	userKey := datastore.NewIncompleteKey(c, UserKind, nil)

	return datastore.Put(c, userKey, user)
}

// GetUserByEmail returns from the datastorethe user associated with the provided email.
// Returns the key of the entry, the user structure and a possible error.
// The key and the user are nil if there is no user stored with the provided email.
func GetUserByEmail(c appengine.Context, email string) (*datastore.Key, *User, error) {
	if email == "" {
		return nil, nil, InvalidEmailErr
	}

	query := datastore.NewQuery(UserKind).Filter("email =", string(email))

	for t := query.Run(c); ; {
		var user User
		key, err := t.Next(&user)
		if err == datastore.Done {
			return nil, nil, nil
		}
		if err != nil {
			return nil, nil, err
		}
		return key, &user, nil
	}

	return nil, nil, nil
}

// UpdateUser takes the email of an existing user and the new user structure which
// must replace the old entry in the datastore.
// Returns the key of the entry. Returns an error if there is no user with the
// provided email or if there is an error returned by the datastore.
func UpdateUser(c appengine.Context, prevEmail string, user *User) (*datastore.Key, error) {
	if prevEmail == "" {
		return nil, InvalidEmailErr
	}

	if user == nil {
		return nil, errors.New("user to update to datastore can't be empty")
	}

	key, _, err := GetUserByEmail(c, prevEmail)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, errors.New("no user with email " + prevEmail + " found for updating.")
	}

	return datastore.Put(c, key, user)
}

// DeleteUser deletes the user entry having the provided email from the datastore.
// Returns an error if there is no user with the provided email or if there is an
// error returned by the datastore.
func DeleteUSer(c appengine.Context, email string) error {
	if email == "" {
		return InvalidEmailErr
	}

	key, _, err := GetUserByEmail(c, email)
	if err != nil {
		return err
	}
	if key == nil {
		return errors.New("no user with this email found")
	}

	return datastore.Delete(c, key)
}
