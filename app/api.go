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

	api.Handle("/sitter", PetsyAuthHandler(addSitter)).Methods("POST")
	api.Handle("/sitter/{user}", PetsyJsonHandler(getSitter)).Methods("GET")
	api.Handle("/sitter/{user}", PetsyAuthHandler(updateSitter)).Methods("POST")
	api.Handle("/sitters", PetsyJsonHandler(getSitters)).Methods("GET")

	api.Handle("/owner/{user}", PetsyJsonHandler(getOwner)).Methods("GET")
	api.Handle("/owner/{user}", PetsyAuthHandler(updateOwner)).Methods("POST")
	api.Handle("/owner", PetsyAuthHandler(addOwner)).Methods("POST")
	api.Handle("/owners", PetsyJsonHandler(getOwners)).Methods("GET")

	api.Handle("/pet", PetsyAuthHandler(addPet)).Methods("POST")
	api.Handle("/owner/{user}/pet/{pet}", PetsyJsonHandler(getPet)).Methods("GET")
	api.Handle("/owner/{user}/pet/{pet}", PetsyAuthHandler(updatePet)).Methods("POST")
	api.Handle("/owner/{user}/pets", PetsyJsonHandler(getPets)).Methods("GET")

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
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "error decoding data: "+err.Error())
		return
	}

	// Validate sitter struct fields.
	if err := sitter.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "invalid sitter data: "+err.Error())
		return
	}

	sitter = sitter.AddCommonData(user)

	// Check if there is another sitter profile for this user.
	_, oldSitter, err := role.GetSitter(ctx, userKey)
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
	if _, err := role.AddSitterForUser(ctx, &sitter, userKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error adding sitter profile: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getSitter(c *Context, w http.ResponseWriter, r *http.Request) {
	ctx, _ := c.GetAppengineContext()

	// Get user email from request url.
	vars := mux.Vars(r)
	userEmail := vars["user"]

	// Get sitter from datastore.
	_, sitter, err := role.GetSitterFromEmail(ctx, userEmail)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting sitter profile: "+err.Error())
		return
	}
	if sitter == nil {
		w.WriteHeader(http.StatusNotFound)
		JsonError(c, 101, "sitter does not exist")
		return
	}

	JsonResponse(c, sitter)
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
	user, _ := c.GetUser()

	// Get user email from request url.
	vars := mux.Vars(r)
	userEmail := vars["user"]

	// Allow update only on the logged user.
	if userEmail != user.Email {
		w.WriteHeader(http.StatusForbidden)
		JsonError(c, 101, "Not allowed to update another user.")
		return
	}

	// Get sitter from datastore.
	sitterKey, sitter, err := role.GetSitterFromEmail(ctx, userEmail)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "Error getting sitter profile: "+err.Error())
		return
	}
	if sitter == nil {
		w.WriteHeader(http.StatusNotFound)
		JsonError(c, 101, "Sitter does not exist.")
		return
	}

	var newSitter role.Sitter

	// Get sitter struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&newSitter); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "error decoding sitter: "+err.Error())
		return
	}

	// Validate sitter struct fields.
	if err := newSitter.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "Invalid sitter data: "+err.Error())
		return
	}

	// Add data from user (email, name).
	newSitter.AddCommonData(user)

	// Update sitter.
	if _, err := role.UpdateSitter(ctx, sitterKey, &newSitter); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error saving sitter: "+err.Error())
		return
	}
}

func addOwner(c *Context, w http.ResponseWriter, r *http.Request) {
	ctx, _ := c.GetAppengineContext()
	user, _ := c.GetUser()
	userKey, _ := c.GetUserKey()

	var owner role.Owner

	// Get owner struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&owner); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "error decoding input json: "+err.Error())
		return
	}

	// Validate owner struct fields.
	if err := owner.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "invalid owner data"+err.Error())
		return
	}

	// Add data from user (email, name).
	owner = owner.AddCommonData(user)

	// Check if there is another owner profile for this user.
	_, oldOwner, err := role.GetOwner(ctx, userKey)
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
	if _, err := role.AddOwnerForUser(ctx, &owner, userKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "Error adding the owner profile: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getOwner(c *Context, w http.ResponseWriter, r *http.Request) {
	// Get user email from request url.
	vars := mux.Vars(r)
	userEmail := vars["user"]

	ctx, _ := c.GetAppengineContext()

	// Get owner from datastore.
	_, owner, err := role.GetOwnerFromEmail(ctx, userEmail)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting owner profile: "+err.Error())
		return
	}
	if owner == nil {
		w.WriteHeader(http.StatusNotFound)
		JsonError(c, 101, "This owner does not exist.")
		return
	}

	JsonResponse(c, owner)
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
	userEmail := vars["user"]

	ctx, _ := c.GetAppengineContext()
	user, _ := c.GetUser()

	// Allow update only on the logged user.
	if userEmail != user.Email {
		w.WriteHeader(http.StatusForbidden)
		JsonError(c, 101, "Not allowed to update another user.")
		return
	}

	// Get owner from datastore.
	ownerKey, owner, err := role.GetOwnerFromEmail(ctx, userEmail)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting owner data: "+err.Error())
		return
	}
	if owner == nil {
		w.WriteHeader(http.StatusNotFound)
		JsonError(c, 101, "Owner profile does not exist.")
		return
	}

	var newOwner role.Owner

	// Get owner struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&newOwner); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "error decoding owner data: "+err.Error())
		return
	}

	// Validate owner struct fields.
	if err := newOwner.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "invalid owner data: "+err.Error())
		return
	}

	newOwner.AddCommonData(user)

	// Update owner.
	if _, err := role.UpdateOwner(ctx, ownerKey, &newOwner); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error saving owner: "+err.Error())
		return
	}
}

func addPet(c *Context, w http.ResponseWriter, r *http.Request) {
	ctx, _ := c.GetAppengineContext()
	user, _ := c.GetUser()
	userKey, _ := c.GetUserKey()

	var pet role.Pet

	// Get pet struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&pet); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "error decoding pet data:"+err.Error())
		return
	}

	// Validate pet struct fields.
	if err := pet.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "invalid pet data: "+err.Error())
		return
	}

	// Check the owner profile of the pet owner
	ownerKey, _, err := role.GetOwner(ctx, userKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting owner entry: "+err.Error())
		return
	}
	if ownerKey == nil {
		w.WriteHeader(http.StatusNotFound)
		JsonError(c, 101, "no owner profile found for user.")
		return
	}

	// Check if there exists the pet in the datastore.
	_, oldPet, err := role.GetPetFromEmail(ctx, user.Email, pet.Name)
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
	if _, err := role.AddPetForOwner(ctx, &pet, userKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error storing pet profile: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getPet(c *Context, w http.ResponseWriter, r *http.Request) {
	// Get user email from request url.
	vars := mux.Vars(r)
	userEmail := vars["user"]
	petName := vars["pet"]

	ctx, _ := c.GetAppengineContext()

	// Get pet from datastore.
	_, pet, err := role.GetPetFromEmail(ctx, userEmail, petName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting pet: "+err.Error())
		return
	}
	if pet == nil {
		w.WriteHeader(http.StatusNotFound)
		JsonError(c, 101, "pet profile does not exist.")
		return
	}

	JsonResponse(c, pet)
}

func getPets(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userEmail := vars["user"]

	ctx, _ := c.GetAppengineContext()

	_, pets, err := role.GetPetsFromEmail(ctx, userEmail)
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
	userEmail := vars["user"]
	petName := vars["pet"]

	ctx, _ := c.GetAppengineContext()
	user, _ := c.GetUser()

	// Allow update only on the logged user.
	if userEmail != user.Email {
		w.WriteHeader(http.StatusForbidden)
		JsonError(c, 101, "Not allowed to update another user.")
		return
	}

	// Get pet from datastore.
	petKey, pet, err := role.GetPetFromEmail(ctx, userEmail, petName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting pet profile: "+err.Error())
		return
	}
	if pet == nil {
		w.WriteHeader(http.StatusNotFound)
		JsonError(c, 101, "this pet profile does not exist")
		return
	}

	var newPet role.Pet

	// Get pet struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&newPet); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "error decoding sent pet profile")
		return
	}

	// Validate pet struct fields.
	if err := newPet.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "invalid pet data"+err.Error())
		return
	}

	// Update pet.
	if _, err := role.UpdatePet(ctx, petKey, &newPet); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error updating pet profile: "+err.Error())
		return
	}
}
