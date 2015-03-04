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

func (p Pet) Validate() error {
	return nil
}

func AddPet(c appengine.Context, pet *Pet, owner *Owner) (*datastore.Key, error) {
	ownerKey, owner, err := GetOwnerFromEmail(c, owner.Email)
	if err != nil {
		return nil, err
	}
	if ownerKey == nil {
		return nil, errors.New("Cannot find owner with specified email.")
	}

	return AddPetForOwner(c, pet, ownerKey)
}

func AddPetForOwner(c appengine.Context, pet *Pet, ownerKey *datastore.Key) (*datastore.Key, error) {
	pet.ownerid = ownerKey.Encode()

	petKey := datastore.NewIncompleteKey(c, PetKind, ownerKey)
	return datastore.Put(c, petKey, pet)
}

func UpdatePet(c appengine.Context, petKey *datastore.Key, pet *Pet) (*datastore.Key, error) {
	return datastore.Put(c, petKey, pet)
}

func GetPetFromEmail(c appengine.Context, userEmail string, petName string) (*datastore.Key, *Pet, error) {
	if userEmail == "" {
		return nil, nil, errors.New("user email cannot be nil.")
	}

	query := datastore.NewQuery(PetKind).Filter("email =", userEmail).Filter("name =", petName)

	for t := query.Run(c); ; {
		var pet Pet
		key, err := t.Next(&pet)
		if err == datastore.Done {
			return nil, nil, nil
		}
		if err != nil {
			return nil, nil, err
		}
		return key, &pet, nil
	}

	return nil, nil, nil
}

func GetPetsFromEmail(c appengine.Context, userEmail string) (keys []*datastore.Key, pets []*Pet, err error) {
	if userEmail == "" {
		return nil, nil, errors.New("user email cannot be nil.")
	}

	ownerKey, _, err := GetOwnerFromEmail(c, userEmail)
	if err != nil {
		return nil, nil, err
	}
	if ownerKey == nil {
		return nil, nil, errors.New("no owner profile defined for this email.")
	}

	query := datastore.NewQuery(PetKind).Ancestor(ownerKey)

	for t := query.Run(c); ; {
		var pet Pet
		key, err := t.Next(&pet)
		if err == datastore.Done {
			return keys, pets, nil
		}
		if err != nil {
			return nil, nil, err
		}
		keys = append(keys, key)
		pets = append(pets, &pet)
	}

	return
}
