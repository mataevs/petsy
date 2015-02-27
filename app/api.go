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

	return appErrorf(http.StatusNotFound, "getSitter - not implemented")
}

func updateSitter(c *Context, w io.Writer, r *http.Request) (error, bool) {
	return appErrorf(http.StatusNotFound, "updateSitter - not implemented"), false
}

func addOwner(c *Context, w io.Writer, r *http.Request) (error, bool) {
	return appErrorf(http.StatusNotFound, "add owner - not implemented"), false
}

func getOwner(c *Context, w io.Writer, r *http.Request) error {
	return appErrorf(http.StatusNotFound, "get owner - not implemented")
}

func updateOwner(c *Context, w io.Writer, r *http.Request) (error, bool) {
	return appErrorf(http.StatusNotFound, "update owner - not implemented"), false
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
