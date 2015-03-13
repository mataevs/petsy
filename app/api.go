// +build appengine

package petsy

import (
	"encoding/json"
	"net/http"

	"petsy/user/role"

	"github.com/gorilla/mux"
)

func init() {
	api := mux.NewRouter().PathPrefix("/api/").Subrouter()

	api.Handle("/profile/{profile}", PetsyJsonHandler(getProfile)).Methods("GET")
	api.Handle("/profile/{profile}", PetsyAuthJsonHandler(updateProfile)).Methods("POST")

	api.Handle("/sitter", PetsyAuthJsonHandler(addSitter)).Methods("POST")
	api.Handle("/sitter/{user}", PetsyJsonHandler(getSitter)).Methods("GET")
	api.Handle("/sitter/{user}", PetsyAuthJsonHandler(updateSitter)).Methods("POST")
	api.Handle("/sitters", PetsyJsonHandler(getSitters)).Methods("GET")

	api.Handle("/owner/{user}", PetsyJsonHandler(getOwner)).Methods("GET")
	api.Handle("/owner/{user}", PetsyAuthJsonHandler(updateOwner)).Methods("POST")
	api.Handle("/owner", PetsyAuthJsonHandler(addOwner)).Methods("POST")
	api.Handle("/owners", PetsyJsonHandler(getOwners)).Methods("GET")

	api.Handle("/pet", PetsyAuthJsonHandler(addPet)).Methods("POST")
	api.Handle("/owner/{user}/pet/{pet}", PetsyJsonHandler(getPet)).Methods("GET")
	api.Handle("/owner/{user}/pet/{pet}", PetsyAuthJsonHandler(updatePet)).Methods("POST")
	api.Handle("/owner/{user}/pets", PetsyJsonHandler(getPets)).Methods("GET")

	http.Handle("/api/", api)
}

func updateProfile(c *Context, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	JsonError(c, 101, "update profile - not implemented")
}

func getProfile(c *Context, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	JsonError(c, 101, "update profile - not implemented")
}

// API endpoint for associating a sitter profile to a user.
// Request - JSON format. Response - JSON.
// TODO - return JSON responses.
func addSitter(c *Context, w http.ResponseWriter, r *http.Request) {
	ctx, _ := c.GetAppengineContext()
	user, _ := c.GetUser()
	userKey, _ := c.GetUserKey()

	var sitter role.Sitter

	// Get sitter struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&sitter); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error decoding data: "+err.Error())
	}

	// Validate sitter struct fields.
	if err := sitter.Validate(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "invalid sitter data: "+err.Error())
	}

	sitter = sitter.AddCommonData(user)

	// Check if there is another sitter profile for this user.
	_, oldSitter, err := role.GetSitter(ctx, userKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error decoding data: "+err.Error())
	}
	if oldSitter != nil {
		w.WriteHeader(http.StatusForbidden)
		JsonError(c, 101, "User already has a sitter profile associated.")
	}

	// Add the sitter profile.
	if _, err := role.AddSitterForUser(ctx, &sitter, userKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error adding sitter profile: "+err.Error())
	}

	w.WriteHeader(http.StatusCreated)
}

func getSitter(c *Context, w http.ResponseWriter, r *http.Request) {
	// Get user email from request url.
	vars := mux.Vars(r)
	userEmail := vars["user"]

	// Get sitter from datastore.
	_, sitter, err := role.GetSitterFromEmail(c.ctx, userEmail)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}
	if sitter == nil {
		return appErrorf(http.StatusNotFound, "This sitter does not exist.")
	}

	// Encode sitter to json.
	enc := json.NewEncoder(w)
	if err := enc.Encode(sitter); err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}

	return nil
}

func getSitters(c *Context, w http.ResponseWriter, r *http.Request) {
	_, sitters, err := role.GetSitters(c.ctx)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(sitters); err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}

	return nil
}

func updateSitter(c *Context, w http.ResponseWriter, r *http.Request) {
	// Get user email from request url.
	vars := mux.Vars(r)
	userEmail := vars["user"]

	// Allow update only on the logged user.
	if userEmail != c.user.Email {
		return appErrorf(http.StatusForbidden, "Not allowed to update another user."), false
	}

	// Get sitter from datastore.
	sitterKey, sitter, err := role.GetSitterFromEmail(c.ctx, userEmail)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err), false
	}
	if sitter == nil {
		return appErrorf(http.StatusNotFound, "This sitter does not exist."), false
	}

	var newSitter role.Sitter

	// Get sitter struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&newSitter); err != nil {
		return appErrorf(http.StatusInternalServerError, "error decoding: %v", err), false
	}

	// Validate sitter struct fields.
	if err := newSitter.Validate(); err != nil {
		return appErrorf(http.StatusInternalServerError, "Invalid sitter data: %v", err), false
	}

	// Add data from user (email, name).
	newSitter.AddCommonData(c.user)

	// Update sitter.
	if _, err := role.UpdateSitter(c.ctx, sitterKey, &newSitter); err != nil {
		return appErrorf(http.StatusInternalServerError, "error saving sitter: %v", err), false
	}

	return nil, false
}

func addOwner(c *Context, w http.ResponseWriter, r *http.Request) {
	var owner role.Owner

	// Get owner struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&owner); err != nil {
		return appErrorf(http.StatusInternalServerError, "error decoding: %v", err), false
	}

	// Validate owner struct fields.
	if err := owner.Validate(); err != nil {
		return appErrorf(http.StatusInternalServerError, "Invalid sitter data: %v", err), false
	}

	// Add data from user (email, name).
	owner = owner.AddCommonData(c.user)

	// Check if there is another sitter profile for this user.
	_, oldOwner, err := role.GetOwner(c.ctx, c.userKey)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err), false
	}
	if oldOwner != nil {
		return appErrorf(http.StatusForbidden, "User already has an owner profile associated."), false
	}

	// Add the owner profile.
	if _, err := role.AddOwnerForUser(c.ctx, &owner, c.userKey); err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err), false
	}

	return appReturn(http.StatusCreated), false
}

func getOwner(c *Context, w http.ResponseWriter, r *http.Request) {
	// Get user email from request url.
	vars := mux.Vars(r)
	userEmail := vars["user"]

	// Get owner from datastore.
	_, owner, err := role.GetOwnerFromEmail(c.ctx, userEmail)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}
	if owner == nil {
		return appErrorf(http.StatusNotFound, "This owner does not exist.")
	}

	// Encode owner to json.
	enc := json.NewEncoder(w)
	if err := enc.Encode(owner); err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}

	return nil
}

func getOwners(c *Context, w http.ResponseWriter, r *http.Request) {
	_, owners, err := role.GetOwners(c.ctx)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(owners); err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}

	return nil
}

func updateOwner(c *Context, w http.ResponseWriter, r *http.Request) {
	// Get user email from request url.
	vars := mux.Vars(r)
	userEmail := vars["user"]

	// Allow update only on the logged user.
	if userEmail != c.user.Email {
		return appErrorf(http.StatusForbidden, "Not allowed to update another user."), false
	}

	// Get owner from datastore.
	ownerKey, owner, err := role.GetOwnerFromEmail(c.ctx, userEmail)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err), false
	}
	if owner == nil {
		return appErrorf(http.StatusNotFound, "This owner does not exist."), false
	}

	var newOwner role.Owner

	// Get owner struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&newOwner); err != nil {
		return appErrorf(http.StatusInternalServerError, "error decoding: %v", err), false
	}

	// Validate owner struct fields.
	if err := newOwner.Validate(); err != nil {
		return appErrorf(http.StatusInternalServerError, "Invalid owner data: %v", err), false
	}

	newOwner.AddCommonData(c.user)

	// Update owner.
	if _, err := role.UpdateOwner(c.ctx, ownerKey, &newOwner); err != nil {
		return appErrorf(http.StatusInternalServerError, "error saving owner: %v", err), false
	}

	return nil, false
}

func addPet(c *Context, w http.ResponseWriter, r *http.Request) {
	var pet role.Pet

	// Get pet struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&pet); err != nil {
		return appErrorf(http.StatusInternalServerError, "error decoding: %v", err), false
	}

	// Validate pet struct fields.
	if err := pet.Validate(); err != nil {
		return appErrorf(http.StatusInternalServerError, "Invalid pet data: %v", err), false
	}

	ownerKey, _, err := role.GetOwner(c.ctx, c.userKey)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "error getting owner entry: %v", err), false
	}
	if ownerKey == nil {
		return appErrorf(http.StatusNotFound, "no owner profile found for user."), false
	}

	// Check if there exists the pet in the datastore.
	_, oldPet, err := role.GetPetFromEmail(c.ctx, c.user.Email, pet.Name)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err), false
	}
	if oldPet != nil {
		return appErrorf(http.StatusNotFound, "A pet with the same name already exists."), false
	}

	// Add the pet profile.
	if _, err := role.AddPetForOwner(c.ctx, &pet, c.userKey); err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err), false
	}

	return appReturn(http.StatusCreated), false
}

func getPet(c *Context, w http.ResponseWriter, r *http.Request) {
	// Get user email from request url.
	vars := mux.Vars(r)
	userEmail := vars["user"]
	petName := vars["pet"]

	// Get pet from datastore.
	_, pet, err := role.GetPetFromEmail(c.ctx, userEmail, petName)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}
	if pet == nil {
		return appErrorf(http.StatusNotFound, "This pet does not exist.")
	}

	// Encode pet to json.
	enc := json.NewEncoder(w)
	if err := enc.Encode(pet); err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}

	return nil
}

func getPets(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userEmail := vars["user"]

	_, pets, err := role.GetPetsFromEmail(c.ctx, userEmail)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(pets); err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}

	return nil
}

func updatePet(c *Context, w http.ResponseWriter, r *http.Request) {
	// Get user email from request url.
	vars := mux.Vars(r)
	userEmail := vars["user"]
	petName := vars["pet"]

	// Allow update only on the logged user.
	if userEmail != c.user.Email {
		return appErrorf(http.StatusForbidden, "Not allowed to update another user."), false
	}

	// Get pet from datastore.
	petKey, pet, err := role.GetPetFromEmail(c.ctx, userEmail, petName)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err), false
	}
	if pet == nil {
		return appErrorf(http.StatusNotFound, "This pet does not exist."), false
	}

	var newPet role.Pet

	// Get pet struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&newPet); err != nil {
		return appErrorf(http.StatusInternalServerError, "error decoding: %v", err), false
	}

	// Validate pet struct fields.
	if err := newPet.Validate(); err != nil {
		return appErrorf(http.StatusInternalServerError, "Invalid owner data: %v", err), false
	}

	// Update pet.
	if _, err := role.UpdatePet(c.ctx, petKey, &newPet); err != nil {
		return appErrorf(http.StatusInternalServerError, "error saving pet: %v", err), false
	}

	return nil, false
}
