// +build appengine

package profile

import (
	"errors"
	"time"

	"appengine"
	"appengine/datastore"
)

type Pet struct {
	OwnerProfileId string `json:"-"`
	Name           string
	Species        string
	Breed          string
	Description    string
	Birthdate      time.Time
}

const (
	PetKind = "pets"
)

func (p Pet) Validate() error {
	return nil
}

func AddPet(c appengine.Context, pet *Pet, profile *UserProfile) (*datastore.Key, error) {
	profileKey, profile, err := GetProfileForEmail(c, profile.Email)
	if err != nil {
		return nil, err
	}
	if profileKey == nil {
		return nil, errors.New("Cannot find owner with specified email.")
	}

	return AddPetForProfileKey(c, pet, profileKey)
}

func AddPetForProfileKey(c appengine.Context, pet *Pet, profileKey *datastore.Key) (*datastore.Key, error) {
	pet.OwnerProfileId = profileKey.Encode()

	petKey := datastore.NewIncompleteKey(c, PetKind, profileKey)

	return datastore.Put(c, petKey, pet)
}

func UpdatePet(c appengine.Context, petKey *datastore.Key, pet *Pet) (*datastore.Key, error) {
	return datastore.Put(c, petKey, pet)
}

func GetPet(c appengine.Context, encodedId string) (*datastore.Key, *Pet, error) {
	var pet Pet

	key, err := datastore.DecodeKey(encodedId)
	if err != nil {
		return nil, nil, err
	}

	if err := datastore.Get(c, key, &pet); err != nil {
		if err == datastore.ErrNoSuchEntity {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	return key, &pet, nil
}

func GetPetForNameEmail(c appengine.Context, userEmail string, petName string) (*datastore.Key, *Pet, error) {
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

func GetPetsForUser(c appengine.Context, userId string) (keys []*datastore.Key, pets []*Pet, err error) {
	if userId == "" {
		return nil, nil, errors.New("user id cannot be nil.")
	}

	profileKey, _, err := GetProfile(c, userId)
	if err != nil {
		return nil, nil, err
	}
	if profileKey == nil {
		return nil, nil, errors.New("no owner profile defined for this email.")
	}

	query := datastore.NewQuery(PetKind).Ancestor(profileKey)

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

func GetPetsForEmail(c appengine.Context, userEmail string) (keys []*datastore.Key, pets []*Pet, err error) {
	if userEmail == "" {
		return nil, nil, errors.New("user email cannot be nil.")
	}

	profileKey, _, err := GetProfileForEmail(c, userEmail)
	if err != nil {
		return nil, nil, err
	}
	if profileKey == nil {
		return nil, nil, errors.New("no owner profile defined for this email.")
	}

	query := datastore.NewQuery(PetKind).Ancestor(profileKey)

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
