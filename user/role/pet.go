package role

import (
	"errors"
	"time"

	"appengine"
	"appengine/datastore"
)

type Pet struct {
	ownerid     string
	Name        string
	Species     string
	Breed       string
	Description string
	Birthdate   time.Time
}

const (
	PetKind = "pets"
)

func AddPet(c appengine.Context, pet *Pet, owner *Owner) (*datastore.Key, error) {
	ownerKey, owner, err := GetOwnerFromEmail(c, owner.Email)
	if err != nil {
		return nil, err
	}
	if ownerKey == nil {
		return nil, errors.New("Cannot find owner with specified email.")
	}

	pet.ownerid = ownerKey.Encode()

	petKey := datastore.NewIncompleteKey(c, PetKind, ownerKey)
	return datastore.Put(c, petKey, pet)
}
