// +build appengine

package model

import (
	"errors"

	"appengine"
	"appengine/datastore"
)

const (
	SitterModelName = "sitter"
)

type Sitter struct {
	key               *datastore.Key `datastore:"-"`
	UserKey           *datastore.Key `json:"-"`
	SitterDescription string
	HousingType       string `json:"housing_type"`
	Space             string
	Prices            string
	OwnedPets         string  `json:"owned_pets"`
	OwnedCar          string  `json:"owned_car"`
	ResponseRate      float32 `json:"omitempty"`
	ResponseTime      float32 `json:"omitempty"`
	Rating            string  `json:"omitempty"`
}

func (s *Sitter) Validate() error {
	return nil
}

func convertToSitter(t interface{}) (*Sitter, error) {
	if sitter, ok := t.(*Sitter); !ok {
		return nil, errors.New("unable to convert to sitter type")
	} else {
		return sitter, nil
	}
}

func (p *Sitter) Add(c appengine.Context) (*datastore.Key, error) {
	if p.key != nil {
		return nil, errors.New("sitter already in datastore")
	}
	key := datastore.NewIncompleteKey(c, SitterModelName, nil)
	p.key = key
	return datastore.Put(c, key, p)
}

func (p *Sitter) GetById(c appengine.Context, encodedId string) (*datastore.Key, *Sitter, error) {
	if key, model, err := getById(c, encodedId, SitterModelName); model != nil {
		sitter, err := convertToSitter(model)
		sitter.key = key
		return key, sitter, err
	} else {
		return key, nil, err
	}
}

func (p *Sitter) GetByEmail(c appengine.Context, email string) (*datastore.Key, *Sitter, error) {
	if key, model, err := getByEmail(c, email, SitterModelName); model != nil {
		sitter, err := convertToSitter(model)
		sitter.key = key
		return key, sitter, err
	} else {
		return key, nil, err
	}
}

func (p *Sitter) GetByAncestorKey(c appengine.Context, ancestorKey *datastore.Key) (*datastore.Key, *Sitter, error) {
	if key, model, err := getByAncestorKey(c, ancestorKey, SitterModelName); model != nil {
		sitter, err := convertToSitter(model)
		sitter.key = key
		return key, sitter, err
	} else {
		return key, nil, err
	}
}

func (p *Sitter) Update(c appengine.Context) (*datastore.Key, error) {
	key, err := datastore.Put(c, p.key, p)
	if key != nil {
		p.key = key
	}
	return key, err
}

func (p *Sitter) Delete(c appengine.Context) error {
	return datastore.Delete(c, p.key)
}

func (p *Sitter) Key() *datastore.Key {
	return p.key
}

// func AddSitter(c appengine.Context, sitter *Sitter) (*datastore.Key, error) {
// 	userKey, _, err := petsyuser.GetUserByEmail(c, sitter.Email)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if userKey == nil {
// 		return nil, errors.New("Cannot find user with specified email.")
// 	}

// 	return AddSitterForUserKey(c, sitter, userKey)
// }

// func AddSitterForUserKey(c appengine.Context, sitter *Sitter, userKey *datastore.Key) (*datastore.Key, error) {
// 	sitter.UserKey = userKey

// 	sitterKey := datastore.NewIncompleteKey(c, SitterKind, userKey)

// 	return datastore.Put(c, sitterKey, sitter)
// }

// func UpdateSitter(c appengine.Context, sitterKey *datastore.Key, sitter *Sitter) (*datastore.Key, error) {

// 	return datastore.Put(c, sitterKey, sitter)
// }

// func GetSitter(c appengine.Context, encodedId string) (*datastore.Key, *Sitter, error) {
// 	var sitter Sitter

// 	key, err := datastore.DecodeKey(encodedId)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	if err := datastore.Get(c, key, &sitter); err != nil {
// 		if err == datastore.ErrNoSuchEntity {
// 			return nil, nil, nil
// 		}
// 		return nil, nil, err
// 	}

// 	return key, &sitter, nil
// }

// func GetSitterForUserKey(c appengine.Context, userKey *datastore.Key) (*datastore.Key, *Sitter, error) {
// 	if userKey == nil {
// 		return nil, nil, errors.New("user key cannot be nil.")
// 	}

// 	query := datastore.NewQuery(SitterKind).Ancestor(userKey)

// 	for t := query.Run(c); ; {
// 		var sitter Sitter
// 		key, err := t.Next(&sitter)
// 		if err == datastore.Done {
// 			return nil, nil, nil
// 		}
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		return key, &sitter, nil
// 	}

// 	return nil, nil, nil
// }

// func GetSitterForEmail(c appengine.Context, userEmail string) (*datastore.Key, *Sitter, error) {
// 	if userEmail == "" {
// 		return nil, nil, errors.New("user email cannot be nil.")
// 	}

// 	query := datastore.NewQuery(SitterKind).Filter("email =", userEmail)

// 	for t := query.Run(c); ; {
// 		var sitter Sitter
// 		key, err := t.Next(&sitter)
// 		if err == datastore.Done {
// 			return nil, nil, nil
// 		}
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		return key, &sitter, nil
// 	}

// 	return nil, nil, nil
// }

// func GetSitters(c appengine.Context) (keys []*datastore.Key, sitters []*Sitter, err error) {
// 	query := datastore.NewQuery(SitterKind)

// 	for t := query.Run(c); ; {
// 		var sitter Sitter
// 		key, err := t.Next(&sitter)
// 		if err == datastore.Done {
// 			return keys, sitters, nil
// 		}
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		keys = append(keys, key)
// 		sitters = append(sitters, &sitter)
// 	}

// 	return
// }
