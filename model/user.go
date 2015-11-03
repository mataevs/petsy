// +build appengine
package model

import (
	"errors"
	"time"

	"appengine"
	"appengine/datastore"
)

type User struct {
	key         *datastore.Key `datastore:"-"`
	UserKey     *datastore.Key `json:"-"`
	Name        string
	Email       string
	Page        string
	Pictures    []string
	AvatarURL   string
	Bio         string
	Description string
	Pets        []Pet
	Birthdate   time.Time
	Rating      string
}

const (
	UserModelName = "user"
)

func (p *User) Validate() error {
	return nil
}

func (p *User) AddAccountData(account *Account) {
	p.Name = account.Name
	p.Email = account.Email
}

func convertToUser(t interface{}) (*User, error) {
	if user, ok := t.(*User); !ok {
		return nil, errors.New("unable to convert to User type")
	} else {
		return user, nil
	}
}

func (p *User) Add(c appengine.Context) (*datastore.Key, error) {
	if p.key != nil {
		return nil, errors.New("user already in datastore")
	}
	key := datastore.NewIncompleteKey(c, UserModelName, nil)
	p.key = key
	return datastore.Put(c, key, p)
}

func (p *User) GetById(c appengine.Context, encodedId string) (*datastore.Key, *User, error) {
	if key, model, err := getById(c, encodedId, UserModelName); model != nil {
		user, err := convertToUser(model)
		user.key = key
		return key, user, err
	} else {
		return key, nil, err
	}
}

func (p *User) GetByEmail(c appengine.Context, email string) (*datastore.Key, *User, error) {
	if key, model, err := getByEmail(c, email, UserModelName); model != nil {
		user, err := convertToUser(model)
		user.key = key
		return key, user, err
	} else {
		return key, nil, err
	}
}

func (p *User) GetByAncestorKey(c appengine.Context, ancestorKey *datastore.Key) (*datastore.Key, *User, error) {
	if key, model, err := getByAncestorKey(c, ancestorKey, UserModelName); model != nil {
		user, err := convertToUser(model)
		user.key = key
		return key, user, err
	} else {
		return key, nil, err
	}
}

func (p *User) Update(c appengine.Context) (*datastore.Key, error) {
	key, err := datastore.Put(c, p.key, p)
	if key != nil {
		p.key = key
	}
	return key, err
}

func (p *User) Delete(c appengine.Context) error {
	return datastore.Delete(c, p.key)
}

func (p *User) Key() *datastore.Key {
	return p.key
}

// func AddProfile(c appengine.Context, p *UserProfile) (*datastore.Key, error) {
// 	userKey, _, err := user.GetUserByEmail(c, p.Email)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if userKey == nil {
// 		return nil, errors.New("Cannot find user with specified email.")
// 	}

// 	return AddProfileForUserKey(c, p, userKey)
// }

// func AddProfileForUserKey(c appengine.Context, p *UserProfile, userKey *datastore.Key) (*datastore.Key, error) {
// 	p.UserKey = userKey

// 	profileKey := datastore.NewIncompleteKey(c, ProfileKind, userKey)

// 	return datastore.Put(c, profileKey, p)
// }

// func UpdateProfile(c appengine.Context, profileKey *datastore.Key, p *UserProfile) (*datastore.Key, error) {
// 	return datastore.Put(c, profileKey, p)
// }

// func GetProfile(c appengine.Context, encodedId string) (*datastore.Key, *UserProfile, error) {
// 	var profile UserProfile

// 	key, err := datastore.DecodeKey(encodedId)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	if err := datastore.Get(c, key, &profile); err != nil {
// 		if err == datastore.ErrNoSuchEntity {
// 			return nil, nil, nil
// 		}
// 		return nil, nil, err
// 	}

// 	return key, &profile, nil
// }

// func GetProfileForUserKey(c appengine.Context, userKey *datastore.Key) (*datastore.Key, *UserProfile, error) {
// 	if userKey == nil {
// 		return nil, nil, errors.New("user key cannot be nil.")
// 	}

// 	query := datastore.NewQuery(ProfileKind).Ancestor(userKey)

// 	for t := query.Run(c); ; {
// 		var userProfile UserProfile
// 		key, err := t.Next(&userProfile)
// 		if err == datastore.Done {
// 			return nil, nil, nil
// 		}
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		return key, &userProfile, nil
// 	}

// 	return nil, nil, nil
// }

// func GetProfileForEmail(c appengine.Context, userEmail string) (*datastore.Key, *UserProfile, error) {
// 	if userEmail == "" {
// 		return nil, nil, errors.New("user email cannot be nil.")
// 	}

// 	query := datastore.NewQuery(ProfileKind).Filter("email =", userEmail)

// 	for t := query.Run(c); ; {
// 		var userProfile UserProfile
// 		key, err := t.Next(&userProfile)
// 		if err == datastore.Done {
// 			return nil, nil, nil
// 		}
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		return key, &userProfile, nil
// 	}

// 	return nil, nil, nil
// }

// func GetProfiles(c appengine.Context) (keys []*datastore.Key, profiles []*UserProfile, err error) {
// 	query := datastore.NewQuery(ProfileKind)

// 	for t := query.Run(c); ; {
// 		var userProfile UserProfile
// 		key, err := t.Next(&userProfile)
// 		if err == datastore.Done {
// 			return keys, profiles, nil
// 		}
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		keys = append(keys, key)
// 		profiles = append(profiles, &userProfile)
// 	}

// 	return
// }
