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

func (o Owner) Validate() error {
	return nil
}

func (o Owner) AddCommonData(user *petsyuser.User) Owner {
	o.Name = user.Name
	o.Email = user.Email
	return o
}

func AddOwner(c appengine.Context, owner *Owner) (*datastore.Key, error) {
	userKey, _, err := petsyuser.GetUserByEmail(c, owner.Email)
	if err != nil {
		return nil, err
	}
	if userKey == nil {
		return nil, errors.New("Cannot find user with specified email.")
	}

	return AddOwnerForUser(c, owner, userKey)
}

func AddOwnerForUser(c appengine.Context, owner *Owner, userKey *datastore.Key) (*datastore.Key, error) {
	owner.userid = userKey.Encode()

	ownerKey := datastore.NewIncompleteKey(c, OwnerKind, userKey)
	return datastore.Put(c, ownerKey, owner)
}

func UpdateOwner(c appengine.Context, ownerKey *datastore.Key, owner *Owner) (*datastore.Key, error) {
	return datastore.Put(c, ownerKey, owner)
}

func GetOwner(c appengine.Context, userKey *datastore.Key) (*datastore.Key, *Owner, error) {
	if userKey == nil {
		return nil, nil, errors.New("user key cannot be nil.")
	}

	query := datastore.NewQuery(OwnerKind).Ancestor(userKey)

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

func GetOwnerFromEmail(c appengine.Context, userEmail string) (*datastore.Key, *Owner, error) {
	if userEmail == "" {
		return nil, nil, errors.New("user email cannot be nil.")
	}

	query := datastore.NewQuery(OwnerKind).Filter("email =", userEmail)

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

func GetOwners(c appengine.Context) (keys []*datastore.Key, owners []*Owner, err error) {
	query := datastore.NewQuery(OwnerKind)

	for t := query.Run(c); ; {
		var owner Owner
		key, err := t.Next(&owner)
		if err == datastore.Done {
			return keys, owners, nil
		}
		if err != nil {
			return nil, nil, err
		}
		keys = append(keys, key)
		owners = append(owners, &owner)
	}

	return
}
