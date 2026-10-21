// +build appengine

package model

import (
	"errors"
	"time"

	"appengine"
	"appengine/datastore"
)

type Pet struct {
	key            *datastore.Key `datastore:"-"`
	OwnerProfileId string         `json:"-"`
	Name           string         `datastore:"name"`
	Species        string         `datastore:"species"`
	Breed          string         `datastore:"breed"`
	Description    string         `datastore:"description"`
	Pictures       []string       `datastore:"pictures"`
	Birthdate      time.Time      `datastore:"birthdate"`
}

const (
	PetModelName = "pet"
)

func (p *Pet) Validate() error {
	return nil
}

func convertToPet(t interface{}) (*Pet, error) {
	if pet, ok := t.(*Pet); !ok {
		return nil, errors.New("unable to convert to Pet type")
	} else {
		return pet, nil
	}
}

func (p *Pet) Add(c appengine.Context) (*datastore.Key, error) {
	if p.key != nil {
		return nil, errors.New("pet already in datastore")
	}
	key := datastore.NewIncompleteKey(c, PetModelName, nil)
	p.key = key
	return datastore.Put(c, key, p)
}

func (p *Pet) GetById(c appengine.Context, encodedId string) (*datastore.Key, *Pet, error) {
	if key, model, err := getById(c, encodedId, PetModelName); model != nil {
		pet, err := convertToPet(model)
		pet.key = key
		return key, pet, err
	} else {
		return key, nil, err
	}
}

func (p *Pet) GetByEmail(c appengine.Context, email string) (*datastore.Key, *Pet, error) {
	if key, model, err := getByEmail(c, email, PetModelName); model != nil {
		pet, err := convertToPet(model)
		pet.key = key
		return key, pet, err
	} else {
		return key, nil, err
	}
}

func (p *Pet) GetByAncestorKey(c appengine.Context, ancestorKey *datastore.Key) (*datastore.Key, *Pet, error) {
	if key, model, err := getByAncestorKey(c, ancestorKey, AccountModelName); model != nil {
		pet, err := convertToPet(model)
		pet.key = key
		return key, pet, err
	} else {
		return key, nil, err
	}
}

func (p *Pet) Update(c appengine.Context) (*datastore.Key, error) {
	key, err := datastore.Put(c, p.key, p)
	if key != nil {
		p.key = key
	}
	return key, err
}

func (p *Pet) Delete(c appengine.Context) error {
	return datastore.Delete(c, p.key)
}

func (p *Pet) Key() *datastore.Key {
	return p.key
}

// func AddPet(c appengine.Context, pet *Pet, profile *UserProfile) (*datastore.Key, error) {
// 	profileKey, profile, err := GetProfileForEmail(c, profile.Email)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if profileKey == nil {
// 		return nil, errors.New("Cannot find owner with specified email.")
// 	}

// 	return AddPetForProfileKey(c, pet, profileKey)
// }

// func AddPetForProfileKey(c appengine.Context, pet *Pet, profileKey *datastore.Key) (*datastore.Key, error) {
// 	pet.OwnerProfileId = profileKey.Encode()

// 	petKey := datastore.NewIncompleteKey(c, PetKind, profileKey)

// 	return datastore.Put(c, petKey, pet)
// }

// func UpdatePet(c appengine.Context, petKey *datastore.Key, pet *Pet) (*datastore.Key, error) {
// 	return datastore.Put(c, petKey, pet)
// }

// func GetPet(c appengine.Context, encodedId string) (*datastore.Key, *Pet, error) {
// 	var pet Pet

// 	key, err := datastore.DecodeKey(encodedId)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	if err := datastore.Get(c, key, &pet); err != nil {
// 		if err == datastore.ErrNoSuchEntity {
// 			return nil, nil, nil
// 		}
// 		return nil, nil, err
// 	}

// 	return key, &pet, nil
// }

// func GetPetForNameEmail(c appengine.Context, userEmail string, petName string) (*datastore.Key, *Pet, error) {
// 	if userEmail == "" {
// 		return nil, nil, errors.New("user email cannot be nil.")
// 	}

// 	query := datastore.NewQuery(PetKind).Filter("email =", userEmail).Filter("name =", petName)

// 	for t := query.Run(c); ; {
// 		var pet Pet
// 		key, err := t.Next(&pet)
// 		if err == datastore.Done {
// 			return nil, nil, nil
// 		}
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		return key, &pet, nil
// 	}

// 	return nil, nil, nil
// }

// func GetPetForUser(c appengine.Context, userId string, petName string) (*datastore.Key, *Pet, error) {
// 	if userId == "" {
// 		return nil, nil, errors.New("user id cannot be nil.")
// 	}
// 	if petName == "" {
// 		return nil, nil, errors.New("pet name cannot be empty")
// 	}

// 	profileKey, _, err := GetProfile(c, userId)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	if profileKey == nil {
// 		return nil, nil, errors.New("no owner profile defined for this email.")
// 	}

// 	query := datastore.NewQuery(PetKind).Filter("name =", petName).Ancestor(profileKey)

// 	for t := query.Run(c); ; {
// 		var pet Pet
// 		key, err := t.Next(&pet)
// 		if err == datastore.Done {
// 			return nil, nil, nil
// 		}
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		return key, &pet, nil
// 	}

// 	return nil, nil, nil
// }

// func GetPetsForUser(c appengine.Context, userId string) (keys []*datastore.Key, pets []*Pet, err error) {
// 	if userId == "" {
// 		return nil, nil, errors.New("user id cannot be nil.")
// 	}

// 	profileKey, _, err := GetProfile(c, userId)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	if profileKey == nil {
// 		return nil, nil, errors.New("no owner profile defined for this email.")
// 	}

// 	query := datastore.NewQuery(PetKind).Ancestor(profileKey)

// 	for t := query.Run(c); ; {
// 		var pet Pet
// 		key, err := t.Next(&pet)
// 		if err == datastore.Done {
// 			return keys, pets, nil
// 		}
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		keys = append(keys, key)
// 		pets = append(pets, &pet)
// 	}

// 	return
// }

// func GetPetsForEmail(c appengine.Context, userEmail string) (keys []*datastore.Key, pets []*Pet, err error) {
// 	if userEmail == "" {
// 		return nil, nil, errors.New("user email cannot be nil.")
// 	}

// 	profileKey, _, err := GetProfileForEmail(c, userEmail)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	if profileKey == nil {
// 		return nil, nil, errors.New("no owner profile defined for this email.")
// 	}

// 	query := datastore.NewQuery(PetKind).Ancestor(profileKey)

// 	for t := query.Run(c); ; {
// 		var pet Pet
// 		key, err := t.Next(&pet)
// 		if err == datastore.Done {
// 			return keys, pets, nil
// 		}
// 		if err != nil {
// 			return nil, nil, err
// 		}
// 		keys = append(keys, key)
// 		pets = append(pets, &pet)
// 	}

// 	return
// }
