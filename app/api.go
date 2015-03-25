// +build appengine

package petsy

import (
	"encoding/json"
	"net/http"

	"petsy/user/role"

	"appengine/datastore"

	"github.com/gorilla/mux"
)

func init() {
	api := mux.NewRouter().PathPrefix("/api/").Subrouter()

	api.Handle("/profile/{profile}", PetsyJsonHandler(getProfile)).Methods("GET")
	api.Handle("/profile/{profile}", PetsyAuthJsonHandler(updateProfile)).Methods("POST")

	api.Handle("/sitter", PetsyAuthHandler(addSitter)).Methods("POST")
	api.Handle("/sitter/{userId}", PetsyJsonHandler(getSitter)).Methods("GET")
	api.Handle("/sitter/{userId}", PetsyAuthHandler(updateSitter)).Methods("POST")
	api.Handle("/sitter/{userId}/comment", PetsyAuthHandler(addSitterComment)).Methods("POST")
	api.Handle("/sitter/{userId}/comments", PetsyJsonHandler(getSitterComments)).Methods("GET")
	api.Handle("/sitters", PetsyJsonHandler(getSitters)).Methods("GET")

	api.Handle("/owner", PetsyAuthHandler(addOwner)).Methods("POST")
	api.Handle("/owner/{userId}", PetsyJsonHandler(getOwner)).Methods("GET")
	api.Handle("/owner/{userId}", PetsyAuthHandler(updateOwner)).Methods("POST")
	api.Handle("/owner/{userId}/comment", PetsyAuthHandler(addOwnerComment)).Methods("POST")
	api.Handle("/owner/{userId}/comments", PetsyJsonHandler(getOwnerComments)).Methods("GET")
	api.Handle("/owners", PetsyJsonHandler(getOwners)).Methods("GET")

	api.Handle("/pet", PetsyAuthHandler(addPet)).Methods("POST")
	api.Handle("/pet/{pet}", PetsyJsonHandler(getPet)).Methods("GET")
	api.Handle("/pet/{pet}", PetsyAuthHandler(updatePet)).Methods("POST")
	api.Handle("/pet/{pet}/comment", PetsyAuthHandler(addPetComment)).Methods("POST")
	api.Handle("/pet/{pet}/comments", PetsyJsonHandler(getPetComments)).Methods("GET")
	api.Handle("/owner/{userId}/pets", PetsyJsonHandler(getPets)).Methods("GET")

	http.Handle("/api/", api)
}

func getProfile(c *Context, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	JsonError(c, 101, "update profile - not implemented")
}

func updateProfile(c *Context, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	JsonError(c, 101, "update profile - not implemented")
}

func getSitterFromRequest(c *Context, w http.ResponseWriter, r *http.Request) *role.Sitter {
	user, _ := c.GetUser()

	var sitter role.Sitter

	// Get sitter struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&sitter); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "error decoding data: "+err.Error())
		return nil
	}

	// Validate sitter struct fields.
	if err := sitter.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "invalid sitter data: "+err.Error())
		return nil
	}

	sitter = sitter.AddCommonData(user)

	return &sitter
}

// API endpoint for associating a sitter profile to a user.
// Request - JSON format. Response - JSON.
// TODO - return JSON responses.
func addSitter(c *Context, w http.ResponseWriter, r *http.Request) {
	ctx, _ := c.GetAppengineContext()
	userKey, _ := c.GetUserKey()

	sitter := getSitterFromRequest(c, w, r)
	if sitter == nil {
		return
	}

	// Check if there is another sitter profile for this user.
	_, oldSitter, err := role.GetSitterForUser(ctx, userKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error decoding data: "+err.Error())
		return
	}
	if oldSitter != nil {
		w.WriteHeader(http.StatusForbidden)
		JsonError(c, 101, "User already has a sitter profile associated.")
		return
	}

	// Add the sitter profile.
	if _, err := role.AddSitterForUser(ctx, sitter, userKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error adding sitter profile: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func returnSitter(c *Context, w http.ResponseWriter, userId string) (*datastore.Key, *role.Sitter) {
	ctx, _ := c.GetAppengineContext()

	userKey, err := datastore.DecodeKey(userId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "error decoding user id: "+err.Error())
	}

	// Get sitter from datastore.
	sitterKey, sitter, err := role.GetSitterForUser(ctx, userKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting sitter profile: "+err.Error())
		return nil, nil
	}
	if sitter == nil {
		w.WriteHeader(http.StatusNotFound)
		JsonError(c, 101, "sitter does not exist")
		return nil, nil
	}

	return sitterKey, sitter
}

func getSitter(c *Context, w http.ResponseWriter, r *http.Request) {
	// Get user email from request url.
	vars := mux.Vars(r)
	userId := vars["userId"]

	if _, sitter := returnSitter(c, w, userId); sitter != nil {
		JsonResponse(c, sitter)
	}
}

func getSitters(c *Context, w http.ResponseWriter, r *http.Request) {
	ctx, _ := c.GetAppengineContext()

	_, sitters, err := role.GetSitters(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting sitter profiles: "+err.Error())
		return
	}

	JsonResponse(c, sitters)
}

func updateSitter(c *Context, w http.ResponseWriter, r *http.Request) {
	ctx, _ := c.GetAppengineContext()
	userKey, _ := c.GetUserKey()

	// Get user email from request url.
	vars := mux.Vars(r)
	userId := vars["userId"]

	// Allow update only on the logged user.
	if userId != userKey.Encode() {
		w.WriteHeader(http.StatusForbidden)
		JsonError(c, 101, "Not allowed to update another user.")
		return
	}

	sitterKey, sitter := returnSitter(c, w, userId)
	if sitter == nil {
		return
	}

	newSitter := getSitterFromRequest(c, w, r)
	if newSitter == nil {
		return
	}

	// Update sitter.
	if _, err := role.UpdateSitter(ctx, sitterKey, newSitter); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error saving sitter: "+err.Error())
		return
	}
}

func getOwnerFromRequest(c *Context, w http.ResponseWriter, r *http.Request) *role.Owner {
	user, _ := c.GetUser()

	var owner role.Owner

	// Get owner struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&owner); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "error decoding data: "+err.Error())
		return nil
	}

	// Validate owner struct fields.
	if err := owner.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "invalid sitter data: "+err.Error())
		return nil
	}

	owner = owner.AddCommonData(user)

	return &owner
}

func addOwner(c *Context, w http.ResponseWriter, r *http.Request) {
	ctx, _ := c.GetAppengineContext()
	userKey, _ := c.GetUserKey()

	owner := getOwnerFromRequest(c, w, r)
	if owner == nil {
		return
	}

	// Check if there is another owner profile for this user.
	_, oldOwner, err := role.GetOwnerForUser(ctx, userKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error checking existing owner profile: "+err.Error())
		return
	}
	if oldOwner != nil {
		w.WriteHeader(http.StatusForbidden)
		JsonError(c, 101, "User already has an owner profile associated.")
		return
	}

	// Add the owner profile.
	if _, err := role.AddOwnerForUser(ctx, owner, userKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "Error adding the owner profile: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func returnOwner(c *Context, w http.ResponseWriter, userId string) (*datastore.Key, *role.Owner) {
	ctx, _ := c.GetAppengineContext()

	// Get owner from datastore.
	ownerKey, owner, err := role.GetOwner(ctx, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting owner profile: "+err.Error())
		return nil, nil
	}
	if owner == nil {
		w.WriteHeader(http.StatusNotFound)
		JsonError(c, 101, "owner does not exist")
		return nil, nil
	}

	return ownerKey, owner
}

func getOwner(c *Context, w http.ResponseWriter, r *http.Request) {
	// Get user email from request url.
	vars := mux.Vars(r)
	userId := vars["userId"]

	if _, owner := returnOwner(c, w, userId); owner != nil {
		JsonResponse(c, owner)
	}
}

func getOwners(c *Context, w http.ResponseWriter, r *http.Request) {
	ctx, _ := c.GetAppengineContext()

	_, owners, err := role.GetOwners(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting owners: %v"+err.Error())
		return
	}

	JsonResponse(c, owners)
}

func updateOwner(c *Context, w http.ResponseWriter, r *http.Request) {
	// Get user email from request url.
	vars := mux.Vars(r)
	userId := vars["userId"]

	ctx, _ := c.GetAppengineContext()
	user, _ := c.GetUser()

	// Allow update only on the logged user.
	if userId != user.Id {
		w.WriteHeader(http.StatusForbidden)
		JsonError(c, 101, "Not allowed to update another user.")
		return
	}

	ownerKey, owner := returnOwner(c, w, userId)
	if owner == nil {
		return
	}

	newOwner := getOwnerFromRequest(c, w, r)
	if newOwner == nil {
		return
	}

	// Update owner.
	if _, err := role.UpdateOwner(ctx, ownerKey, newOwner); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error saving owner: "+err.Error())
		return
	}
}

func getPetFromRequest(c *Context, w http.ResponseWriter, r *http.Request) *role.Pet {
	var pet role.Pet

	// Get pet struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&pet); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "error decoding data: "+err.Error())
		return nil
	}

	// Validate pet struct fields.
	if err := pet.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "invalid sitter data: "+err.Error())
		return nil
	}

	return &pet
}

func addPet(c *Context, w http.ResponseWriter, r *http.Request) {
	ctx, _ := c.GetAppengineContext()
	user, _ := c.GetUser()
	userKey, _ := c.GetUserKey()

	pet := getPetFromRequest(c, w, r)
	if pet == nil {
		return
	}

	ownerKey, _ := returnOwner(c, w, user.Email)
	if ownerKey == nil {
		return
	}

	// Check if there exists the pet in the datastore.
	_, oldPet, err := role.GetPetFromEmailName(ctx, user.Email, pet.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error checking pet profile: "+err.Error())
		return
	}
	if oldPet != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "a pet with the same name already exists.")
		return
	}

	// Add the pet profile.
	if _, err := role.AddPetForOwner(ctx, pet, userKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error storing pet profile: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func returnPet(c *Context, w http.ResponseWriter, petId string) (*datastore.Key, *role.Pet) {
	ctx, _ := c.GetAppengineContext()

	// Get pet from datastore.
	petKey, pet, err := role.GetPet(ctx, petId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting pet: "+err.Error())
		return nil, nil
	}
	if pet == nil {
		w.WriteHeader(http.StatusNotFound)
		JsonError(c, 101, "pet profile does not exist.")
		return nil, nil
	}

	return petKey, pet
}

func getPet(c *Context, w http.ResponseWriter, r *http.Request) {
	// Get user email from request url.
	vars := mux.Vars(r)
	petId := vars["petId"]

	if _, pet := returnPet(c, w, petId); pet != nil {
		JsonResponse(c, pet)
	}
}

func getPets(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]

	ctx, _ := c.GetAppengineContext()

	_, pets, err := role.GetPetsForUser(ctx, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting pet profiles: "+err.Error())
		return
	}

	JsonResponse(c, pets)
}

func updatePet(c *Context, w http.ResponseWriter, r *http.Request) {
	// Get user email from request url.
	vars := mux.Vars(r)
	petId := vars["petId"]

	ctx, _ := c.GetAppengineContext()
	userKey, _ := c.GetUserKey()

	// Get the pet profile.
	petKey, pet := returnPet(c, w, petId)
	if petKey == nil {
		return
	}

	// Get the owner profile.
	ownerId := pet.OwnerId
	_, owner, err := role.GetOwner(ctx, ownerId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting owner profile: "+err.Error())
		return
	}
	if owner == nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "owner not found")
		return
	}
	// Check if the logged user updates its own pet.
	if owner.UserKey != userKey {
		w.WriteHeader(http.StatusUnauthorized)
		JsonError(c, 101, "not allowed to modify pet not belonging to another user.")
		return
	}

	newPet := getPetFromRequest(c, w, r)
	if newPet == nil {
		return
	}

	// Update pet.
	if _, err := role.UpdatePet(ctx, petKey, newPet); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error updating pet profile: "+err.Error())
		return
	}
}
