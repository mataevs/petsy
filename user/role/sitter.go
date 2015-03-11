// +build appengine

package role

import (
	"errors"

	petsyuser "petsy/user"

	"appengine"
	"appengine/datastore"
)

const (
	SitterKind = "sitters"
)

type Sitter struct {
	commonInfo
	Description  string
	HousingType  string
	Space        string
	Prices       string
	OwnsPets     bool
	HasCar       bool
	ResponseRate float32
	ResponseTime float32
	Rating       string
}

func (s Sitter) Validate() error {
	return nil
}

func (s Sitter) AddCommonData(user *petsyuser.User) Sitter {
	s.Name = user.Name
	s.Email = user.Email
	return s
}

func AddSitter(c appengine.Context, sitter *Sitter) (*datastore.Key, error) {
	userKey, _, err := petsyuser.GetUserByEmail(c, sitter.Email)
	if err != nil {
		return nil, err
	}
	if userKey == nil {
		return nil, errors.New("Cannot find user with specified email.")
	}

	return AddSitterForUser(c, sitter, userKey)
}

func AddSitterForUser(c appengine.Context, sitter *Sitter, userKey *datastore.Key) (*datastore.Key, error) {
	sitter.userid = userKey.Encode()

	sitterKey := datastore.NewIncompleteKey(c, SitterKind, userKey)
	return datastore.Put(c, sitterKey, sitter)
}

func UpdateSitter(c appengine.Context, sitterKey *datastore.Key, sitter *Sitter) (*datastore.Key, error) {
	return datastore.Put(c, sitterKey, sitter)
}

func GetSitter(c appengine.Context, userKey *datastore.Key) (*datastore.Key, *Sitter, error) {
	if userKey == nil {
		return nil, nil, errors.New("user key cannot be nil.")
	}

	query := datastore.NewQuery(SitterKind).Ancestor(userKey)

	for t := query.Run(c); ; {
		var sitter Sitter
		key, err := t.Next(&sitter)
		if err == datastore.Done {
			return nil, nil, nil
		}
		if err != nil {
			return nil, nil, err
		}
		return key, &sitter, nil
	}

	return nil, nil, nil
}

func GetSitterFromEmail(c appengine.Context, userEmail string) (*datastore.Key, *Sitter, error) {
	if userEmail == "" {
		return nil, nil, errors.New("user email cannot be nil.")
	}

	query := datastore.NewQuery(SitterKind).Filter("email =", userEmail)

	for t := query.Run(c); ; {
		var sitter Sitter
		key, err := t.Next(&sitter)
		if err == datastore.Done {
			return nil, nil, nil
		}
		if err != nil {
			return nil, nil, err
		}
		return key, &sitter, nil
	}

	return nil, nil, nil
}

func GetSitters(c appengine.Context) (keys []*datastore.Key, sitters []*Sitter, err error) {
	query := datastore.NewQuery(SitterKind)

	for t := query.Run(c); ; {
		var sitter Sitter
		key, err := t.Next(&sitter)
		if err == datastore.Done {
			return keys, sitters, nil
		}
		if err != nil {
			return nil, nil, err
		}
		keys = append(keys, key)
		sitters = append(sitters, &sitter)
	}

	return
}
