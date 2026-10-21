// +build appengine

package petsy

import (
	"encoding/json"
	"log"
	"net/http"

	"petsy/user/profile"

	"appengine/datastore"

	"github.com/gorilla/mux"
)

func init() {
	api := mux.NewRouter().PathPrefix("/api/").Subrouter()

	// api.Handle("/user/{profile}", PetsyJsonHandler(getUser)).Methods("GET")
	// api.Handle("/user/{profile}", PetsyAuthJsonHandler(updateProfile)).Methods("POST")

	api.Handle("/sitter", PetsyAuthHandler(getOwnSitter)).Methods("GET")
	api.Handle("/sitter", PetsyAuthHandler(addSitter)).Methods("POST")
	api.Handle("/sitter/{userId}", PetsyJsonHandler(getSitter)).Methods("GET")
	api.Handle("/sitter/{userId}", PetsyAuthHandler(updateSitter)).Methods("POST")
	api.Handle("/sitter/{userId}/comment", PetsyAuthHandler(addSitterComment)).Methods("POST")
	api.Handle("/sitter/{userId}/comments", PetsyJsonHandler(getSitterComments)).Methods("GET")
	api.Handle("/sitters", PetsyJsonHandler(getSitters)).Methods("GET")

	// todo - get own profile
	api.Handle("/profile", PetsyAuthHandler(addProfile)).Methods("POST")
	api.Handle("/profile/{userId}", PetsyJsonHandler(getProfile)).Methods("GET")
	api.Handle("/profile/{userId}", PetsyAuthHandler(updateProfile)).Methods("POST")
	api.Handle("/profile/{userId}/comment", PetsyAuthHandler(addProfileComment)).Methods("POST")
	api.Handle("/profile/{userId}/comments", PetsyJsonHandler(getProfileComments)).Methods("GET")
	api.Handle("/profiles", PetsyJsonHandler(getProfiles)).Methods("GET")

	api.Handle("/pet", PetsyAuthHandler(addPet)).Methods("POST")
	api.Handle("/pet/{pet}", PetsyJsonHandler(getPet)).Methods("GET")
	api.Handle("/pet/{pet}", PetsyAuthHandler(updatePet)).Methods("POST")
	api.Handle("/pet/{pet}/comment", PetsyAuthHandler(addPetComment)).Methods("POST")
	api.Handle("/pet/{pet}/comments", PetsyJsonHandler(getPetComments)).Methods("GET")
	api.Handle("/owner/{userId}/pets", PetsyJsonHandler(getPets)).Methods("GET")

	http.Handle("/api/", api)
}

// func getProfile(c *Context, w http.ResponseWriter, r *http.Request) {
// 	w.WriteHeader(http.StatusNotImplemented)
// 	JsonError(c, 101, "update profile - not implemented")
// }

// func updateProfile(c *Context, w http.ResponseWriter, r *http.Request) {
// 	w.WriteHeader(http.StatusNotImplemented)
// 	JsonError(c, 101, "update profile - not implemented")
// }

func getSitterFromRequest(c *Context, w http.ResponseWriter, r *http.Request) *profile.Sitter {
	user, _ := c.GetUser()

	var sitter profile.Sitter

	log.Println(r.Body)

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
	_, oldSitter, err := profile.GetSitterForUserKey(ctx, userKey)
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
	if _, err := profile.AddSitterForUserKey(ctx, sitter, userKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error adding sitter profile: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func returnSitter(c *Context, w http.ResponseWriter, userId string) (*datastore.Key, *profile.Sitter) {
	ctx, _ := c.GetAppengineContext()

	userKey, err := datastore.DecodeKey(userId)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "error decoding user id: "+err.Error())
	}

	// Get sitter from datastore.
	sitterKey, sitter, err := profile.GetSitterForUserKey(ctx, userKey)
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

func getOwnSitter(c *Context, w http.ResponseWriter, r *http.Request) {
	ctx, _ := c.GetAppengineContext()
	userKey, _ := c.GetUserKey()

	_, sitter, err := profile.GetSitterForUserKey(ctx, userKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting sitter profile: "+err.Error())
	}
	if sitter == nil {
		w.WriteHeader(http.StatusNotFound)
		JsonError(c, 101, "sitter does not exist")
	}

	log.Println("getOwnSitter", sitter)

	JsonResponse(c, sitter)
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

	_, sitters, err := profile.GetSitters(ctx)
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
	if _, err := profile.UpdateSitter(ctx, sitterKey, newSitter); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error saving sitter: "+err.Error())
		return
	}
}

func getProfileFromRequest(c *Context, w http.ResponseWriter, r *http.Request) *profile.UserProfile {
	user, _ := c.GetUser()

	var profile profile.UserProfile

	// Get profile struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&profile); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "error decoding data: "+err.Error())
		return nil
	}

	// Validate profile struct fields.
	if err := profile.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "invalid sitter data: "+err.Error())
		return nil
	}

	profile = profile.AddCommonData(user)

	return &profile
}

func addProfile(c *Context, w http.ResponseWriter, r *http.Request) {
	ctx, _ := c.GetAppengineContext()
	userKey, _ := c.GetUserKey()

	p := getProfileFromRequest(c, w, r)
	if p == nil {
		return
	}

	// Check if there is another profile for this user.
	_, oldProfile, err := profile.GetProfileForUserKey(ctx, userKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error checking existing profile: "+err.Error())
		return
	}
	if oldProfile != nil {
		w.WriteHeader(http.StatusForbidden)
		JsonError(c, 101, "User already has a profile associated.")
		return
	}

	// Add the profile.
	if _, err := profile.AddProfileForUserKey(ctx, p, userKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "Error adding the profile: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func returnProfile(c *Context, w http.ResponseWriter, userId string) (*datastore.Key, *profile.UserProfile) {
	ctx, _ := c.GetAppengineContext()

	// Get profile from datastore.
	profileKey, p, err := profile.GetProfile(ctx, userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting profile: "+err.Error())
		return nil, nil
	}
	if p == nil {
		w.WriteHeader(http.StatusNotFound)
		JsonError(c, 101, "profile does not exist")
		return nil, nil
	}

	return profileKey, p
}

func getProfile(c *Context, w http.ResponseWriter, r *http.Request) {
	// Get user email from request url.
	vars := mux.Vars(r)
	userId := vars["userId"]

	if _, p := returnProfile(c, w, userId); p != nil {
		JsonResponse(c, p)
	}
}

func getProfiles(c *Context, w http.ResponseWriter, r *http.Request) {
	ctx, _ := c.GetAppengineContext()

	_, profiles, err := profile.GetProfiles(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting owners: %v"+err.Error())
		return
	}

	JsonResponse(c, profiles)
}

func updateProfile(c *Context, w http.ResponseWriter, r *http.Request) {
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

	profileKey, p := returnProfile(c, w, userId)
	if p == nil {
		return
	}

	newProfile := getProfileFromRequest(c, w, r)
	if newProfile == nil {
		return
	}

	// Update profile.
	if _, err := profile.UpdateProfile(ctx, profileKey, newProfile); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error saving owner: "+err.Error())
		return
	}
}

func getPetFromRequest(c *Context, w http.ResponseWriter, r *http.Request) *profile.Pet {
	var pet profile.Pet

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

	pet := getPetFromRequest(c, w, r)
	if pet == nil {
		return
	}

	profileKey, _ := returnProfile(c, w, user.Email)
	if profileKey == nil {
		return
	}

	// Check if there exists the pet in the datastore.
	_, oldPet, err := profile.GetPetForNameEmail(ctx, user.Email, pet.Name)
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
	if _, err := profile.AddPetForProfileKey(ctx, pet, profileKey); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error storing pet profile: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func returnPet(c *Context, w http.ResponseWriter, petId string) (*datastore.Key, *profile.Pet) {
	ctx, _ := c.GetAppengineContext()

	// Get pet from datastore.
	petKey, pet, err := profile.GetPet(ctx, petId)
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

	_, pets, err := profile.GetPetsForUser(ctx, userId)
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

	// Get the owner's user profile.
	ownerProfileId := pet.OwnerProfileId
	_, ownerProfile, err := profile.GetProfile(ctx, ownerProfileId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting owner profile: "+err.Error())
		return
	}
	if ownerProfile == nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "owner not found")
		return
	}
	// Check if the logged user updates its own pet.
	if ownerProfile.UserKey != userKey {
		w.WriteHeader(http.StatusUnauthorized)
		JsonError(c, 101, "not allowed to modify pet not belonging to another user.")
		return
	}

	newPet := getPetFromRequest(c, w, r)
	if newPet == nil {
		return
	}

	// Update pet.
	if _, err := profile.UpdatePet(ctx, petKey, newPet); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error updating pet profile: "+err.Error())
		return
	}
}
