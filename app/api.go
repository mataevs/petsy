package petsy

import (
	"encoding/json"
	"io"
	"net/http"

	"petsy/user/role"

	"github.com/gorilla/mux"
)

func init() {
	api := mux.NewRouter().PathPrefix("/api/").Subrouter()

	api.Handle("/profile/{profile}", appHandler(getProfile)).Methods("GET")
	api.Handle("/profile/{profile}", authReq(updateProfile)).Methods("POST")

	api.Handle("/sitter", authReq(addSitter)).Methods("POST")
	api.Handle("/sitter/{user}", appHandler(getSitter)).Methods("GET")
	api.Handle("/sitter/{user}", authReq(updateSitter)).Methods("POST")

	api.Handle("/owner/{user}", appHandler(getOwner)).Methods("GET")
	api.Handle("/owner/{user}", authReq(updateOwner)).Methods("POST")
	api.Handle("/owner", authReq(addOwner)).Methods("POST")

	api.Handle("/pet/{pet}", appHandler(getPet)).Methods("GET")
	api.Handle("/pet/{pet}", authReq(updatePet)).Methods("POST")
	api.Handle("/pet", authReq(addPet)).Methods("POST")

	http.Handle("/api/", api)
}

func updateProfile(c *Context, w io.Writer, r *http.Request) (error, bool) {
	return appErrorf(http.StatusNotFound, "update profile - not implemented"), false
}

func getProfile(c *Context, w io.Writer, r *http.Request) error {
	w.Write([]byte("User profile"))
	return nil
}

// API endpoint for associating a sitter profile to a user.
// Request - JSON format. Response - JSON.
// TODO - return 201 Created in case of success.
// TODO - return JSON responses.
func addSitter(c *Context, w io.Writer, r *http.Request) (error, bool) {
	var sitter role.Sitter

	// Get sitter struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&sitter); err != nil {
		return appErrorf(http.StatusInternalServerError, "error decoding: %v", err), false
	}

	// Validate sitter struct fields.
	if err := sitter.Validate(); err != nil {
		return appErrorf(http.StatusInternalServerError, "Invalid sitter data: %v", err), false
	}

	// todo - add data from user (email, name)

	// Check if there is another sitter profile for this user.
	_, oldSitter, err := role.GetSitter(c.ctx, c.userKey)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err), false
	}
	if oldSitter != nil {
		return appErrorf(http.StatusForbidden, "User already has a sitter profile associated."), false
	}

	// Add the sitter profile.
	if _, err := role.AddSitterForUser(c.ctx, &sitter, c.userKey); err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err), false
	}

	return nil, false
}

func getSitter(c *Context, w io.Writer, r *http.Request) error {
	// Get user email from request url.
	vars := mux.Vars(r)
	userEmail := vars["user"]

	// Get sitter from datastore.
	_, sitter, err := role.GetSitterFromEmail(c.ctx, userEmail)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}
	if sitter != nil {
		return appErrorf(http.StatusNotFound, "This sitter does not exist.")
	}

	// Encode sitter to json.
	enc := json.NewEncoder(w)
	if err := enc.Encode(sitter); err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}

	return nil
}

func updateSitter(c *Context, w io.Writer, r *http.Request) (error, bool) {
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
	if sitter != nil {
		return appErrorf(http.StatusNotFound, "This sitter does not exist."), false
	}

	// Get sitter struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&sitter); err != nil {
		return appErrorf(http.StatusInternalServerError, "error decoding: %v", err), false
	}

	// Validate sitter struct fields.
	if err := sitter.Validate(); err != nil {
		return appErrorf(http.StatusInternalServerError, "Invalid sitter data: %v", err), false
	}

	// todo - add data from user (email, name)

	// Update sitter.
	if _, err := role.UpdateSitter(c.ctx, sitterKey, sitter); err != nil {
		return appErrorf(http.StatusInternalServerError, "error saving sitter: %v", err), false
	}

	return nil, false
}

func addOwner(c *Context, w io.Writer, r *http.Request) (error, bool) {
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

	// todo - add data from user (email, name)

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

	return nil, false
}

func getOwner(c *Context, w io.Writer, r *http.Request) error {
	// Get user email from request url.
	vars := mux.Vars(r)
	userEmail := vars["user"]

	// Get owner from datastore.
	_, owner, err := role.GetOwnerFromEmail(c.ctx, userEmail)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}
	if owner != nil {
		return appErrorf(http.StatusNotFound, "This owner does not exist.")
	}

	// Encode owner to json.
	enc := json.NewEncoder(w)
	if err := enc.Encode(owner); err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}

	return nil
}

func updateOwner(c *Context, w io.Writer, r *http.Request) (error, bool) {
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
	if owner != nil {
		return appErrorf(http.StatusNotFound, "This owner does not exist."), false
	}

	// Get owner struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&owner); err != nil {
		return appErrorf(http.StatusInternalServerError, "error decoding: %v", err), false
	}

	// Validate owner struct fields.
	if err := owner.Validate(); err != nil {
		return appErrorf(http.StatusInternalServerError, "Invalid owner data: %v", err), false
	}

	// todo - add data from user (email, name)

	// Update owner.
	if _, err := role.UpdateOwner(c.ctx, ownerKey, owner); err != nil {
		return appErrorf(http.StatusInternalServerError, "error saving owner: %v", err), false
	}

	return nil, false
}

func addPet(c *Context, w io.Writer, r *http.Request) (error, bool) {
	var pet role.Pet

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&pet); err != nil {
		return appErrorf(http.StatusInternalServerError, "error decoding: %v", err), false
	}

	enc := json.NewEncoder(w)
	if err := enc.Encode(&pet); err != nil {
		return appErrorf(http.StatusInternalServerError, "error encoding: %v", err), false
	}

	return nil, false
	return appErrorf(http.StatusNotFound, "add pet - not implemented"), false
}

func getPet(c *Context, w io.Writer, r *http.Request) error {
	return appErrorf(http.StatusNotFound, "get pet - not implemented")
}

func updatePet(c *Context, w io.Writer, r *http.Request) (error, bool) {
	return appErrorf(http.StatusNotFound, "update pet - not implemented"), false
}
