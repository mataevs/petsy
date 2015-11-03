package model

import (
	"errors"

	"appengine"
	"appengine/datastore"
)

type Model interface {
	Add(appengine.Context) (*datastore.Key, error)
	GetById(appengine.Context, string) (*datastore.Key, Model, error)
	GetByEmail(appengine.Context, string) (*datastore.Key, Model, error)
	GetByAncestorKey(appengine.Context, *datastore.Key) (*datastore.Key, Model, error)
	Key() *datastore.Key
	Update(appengine.Context, *datastore.Key) (*datastore.Key, error)
	Delete(appengine.Context, *datastore.Key) error
	// todo - add GetAll
}

// createModel creates a new instance of the type specified by modelName.
// i.e. for modelName = 'account', the method creates an Account instance.
func createModel(modelName string) interface{} {
	switch {
	case modelName == AccountModelName:
		return new(Account)
	case modelName == PetModelName:
		return new(Pet)
	default:
		return nil
	}
}

// getById performs a search in the datastore for the entry with the encodedId specified by
// the parameter. modelName specifies the type of the entity searched for.
func getById(c appengine.Context, encodedId string, modelName string) (*datastore.Key, interface{}, error) {
	entity := createModel(modelName)

	key, err := datastore.DecodeKey(encodedId)
	if err != nil {
		return nil, nil, err
	}

	if err := datastore.Get(c, key, entity); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	return key, entity, nil
}

// getByEmail performs a search in the datastore for the entry with the email field specified by
// the parameter. modelName specifies the type of the entity searched for.
func getByEmail(c appengine.Context, email string, modelName string) (*datastore.Key, interface{}, error) {
	if email == "" {
		return nil, nil, errors.New("user email cannot be nil.")
	}

	entity := createModel(modelName)

	query := datastore.NewQuery(modelName).Filter("email =", email)

	for t := query.Run(c); ; {
		key, err := t.Next(entity)
		if err == datastore.Done {
			return nil, nil, nil
		}
		if err != nil {
			return nil, nil, err
		}
		return key, entity, nil
	}

	return nil, nil, nil
}

// getByAncestorKey performs a search in the datastore for the entry with the ancestor specified by
// the parameter. modelName specifies the type of the entity searched for.
func getByAncestorKey(c appengine.Context, ancestorKey *datastore.Key, modelName string) (*datastore.Key, interface{}, error) {
	if ancestorKey == nil {
		return nil, nil, errors.New("ancestor key cannot be nil.")
	}

	entity := createModel(modelName)

	query := datastore.NewQuery(modelName).Ancestor(ancestorKey)

	for t := query.Run(c); ; {
		key, err := t.Next(entity)
		if err == datastore.Done {
			return nil, nil, nil
		}
		if err != nil {
			return nil, nil, err
		}
		return key, entity, nil
	}

	return nil, nil, nil
}
