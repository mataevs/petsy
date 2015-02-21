package petsy

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

func init() {
	store = sessions.NewCookieStore(securecookie.GenerateRandomKey(64), securecookie.GenerateRandomKey(32))
	store.Options = &sessions.Options{
		Domain:   "",
		MaxAge:   3600 * 4,
		HttpOnly: true,
		Path:     "/",
	}

	api := mux.NewRouter().PathPrefix("/api/").Subrouter()

	api.Handle("/profile/{profile}", authReq(updateProfile)).Methods("POST")
	api.Handle("/profile/{profile}", appHandler(getProfile)).Methods("GET")

	api.Handle("/userpage/{user}", appHandler(getUserPage)).Methods("GET")
	api.Handle("/userpage/{user}", authReq(updateUserPage)).Methods("POST")

	api.Handle("/profile/pet/{pet}", authReq(updatePetProfile)).Methods("POST")
	api.Handle("/profile/pet/{pet}", appHandler(getPetProfile)).Methods("GET")

	api.Handle("/find", appHandler(findSitters)).Methods("POST")

	api.Handle("/verification", appHandler(verify)).Methods("GET")

	http.Handle("/api/", api)
}

func updateProfile(c *Context, w io.Writer, r *http.Request) (error, bool) {
	return appErrorf(http.StatusNotFound, "not implemented"), false
}

func getProfile(c *Context, w io.Writer, r *http.Request) error {
	w.Write([]byte("User profile"))
	return nil
}

func getUserPage(c *Context, w io.Writer, r *http.Request) error {
	return appErrorf(http.StatusNotFound, "not implemented")
}

func updateUserPage(c *Context, w io.Writer, r *http.Request) (error, bool) {
	return appErrorf(http.StatusNotFound, "not implemented"), false
}

func updatePetProfile(c *Context, w io.Writer, r *http.Request) (error, bool) {
	return appErrorf(http.StatusNotFound, "not implemented"), false
}

func getPetProfile(c *Context, w io.Writer, r *http.Request) error {
	return appErrorf(http.StatusNotFound, "not implemented")
}

func findSitters(c *Context, w io.Writer, r *http.Request) error {
	return appErrorf(http.StatusNotFound, "not implemented")
}

func verify(c *Context, w io.Writer, r *http.Request) error {
	queryValues := r.URL.Query()

	hash := queryValues.Get("hash")
	scope := queryValues.Get("scope")
	email := queryValues.Get("email")

	if hash == "" || scope == "" || email == "" {
		return appErrorf(http.StatusNotFound, "Link does not exist.")
	}

	responseString := fmt.Sprintf("Hash=%s Scope=%s Email=%s", hash, scope, email)

	w.Write([]byte(responseString))

	return nil
}
