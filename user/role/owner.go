package role

import (
	"errors"

	petsyuser "petsy/user"

	"appengine"
	"appengine/datastore"
)

type Owner struct {
	commonInfo
	Description string
	Rating      string
	Pets        []Pet
}

const (
	OwnerKind = "owners"
)

func AddOwner(c appengine.Context, owner *Owner) (*datastore.Key, error) {
	userKey, _, err := petsyuser.GetUserByEmail(c, owner.Email)
	if err != nil {
		return nil, err
	}
	if userKey == nil {
		return nil, errors.New("Cannot find user with specified email.")
	}

	owner.userid = userKey.Encode()

	ownerKey := datastore.NewIncompleteKey(c, OwnerKind, userKey)
	return datastore.Put(c, ownerKey, owner)
}

func GetOwner(c appengine.Context, email string) (*datastore.Key, *Owner, error) {
	if email == "" {
		return nil, nil, errors.New("email is null.")
	}

	query := datastore.NewQuery(OwnerKind).Filter("email =", string(email))

	for t := query.Run(c); ; {
		var owner Owner
		key, err := t.Next(&owner)
		if err == datastore.Done {
			return nil, nil, nil
		}
		if err != nil {
			return nil, nil, err
		}
		return key, &owner, nil
	}

	return nil, nil, nil
}
