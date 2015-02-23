package petsy

import (
	"fmt"
	"html/template"
	"io"
	"net/http"

	"petsy/hashstore"
	petsyuser "petsy/user"

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

	api.Handle("/verification", appHandler(verifyLink)).Methods("GET")

	api.Handle("/resend-activation-link", appHandler(showResendActivationLink)).Methods("GET")
	api.Handle("/resend-activation-link", appHandler(resendActivationLink)).Methods("POST")

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

func verifyLink(c *Context, w io.Writer, r *http.Request) error {
	queryValues := r.URL.Query()

	hash := queryValues.Get("hash")
	scope := queryValues.Get("scope")
	email := queryValues.Get("email")

	w.Write([]byte(fmt.Sprintf("hash=%s scope=%s email=%s", hash, scope, email)))

	// Check query parameters values.
	if hash == "" || scope == "" || email == "" {
		return appErrorf(http.StatusNotFound, "Link does not exist.")
	}

	// Check validation link validity.
	if valid, err := hashstore.IsValidEntry(c.ctx, hash, email, scope); err == hashstore.NoSuchKeyErr {
		return appErrorf(http.StatusNotFound, "Link does not exist.")
	} else if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	} else if !valid {
		w.Write([]byte("Confirmation link has expired."))

		// todo - add logic for sending a new activation link.

		return nil
	}

	// Get user from datastore.
	_, user, err := petsyuser.GetUserByEmail(c.ctx, email)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}

	// Perform action depending on scope.
	switch scope {
	case "register":
		user.Active = true
		// Mark the user as active.
		if _, err := petsyuser.UpdateUser(c.ctx, user.Email, user); err != nil {
			return appErrorf(http.StatusInternalServerError, "%v", err)
		}

		// Delete hashstore entry.
		if hashstore.DeleteEntry(c.ctx, hash); err != nil {
			return appErrorf(http.StatusInternalServerError, "%v", err)
		}

		w.Write([]byte("User account activated. You can now login."))
		return nil
	default:
		return appErrorf(http.StatusUnauthorized, "Unknown scope.")
	}

	return nil
}

func showResendActivationLink(c *Context, w io.Writer, r *http.Request) error {
	t, _ := template.ParseFiles("templates/resend-activation-link.html")
	t.Execute(w, nil)
	return nil
}

func resendActivationLink(c *Context, w io.Writer, r *http.Request) error {
	email := r.PostFormValue("email")
	pass := r.PostFormValue("password")

	if email == "" {
		return appErrorf(http.StatusForbidden, "Email address cannot be empty.")
	}
	if pass == "" {
		return appErrorf(http.StatusForbidden, "Password cannot be empty.")
	}

	// Get user by email.
	_, user, err := petsyuser.GetUserByEmail(c.ctx, email)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}

	if user == nil {
		return appErrorf(http.StatusForbidden, "Non-existent user or bad password.")
	}

	if !user.CheckPassword(pass) {
		return appErrorf(http.StatusForbidden, "Non-existent user or bad password.")
	}

	if user.Active {
		return appErrorf(http.StatusForbidden, "User is already activated.")
	}

	name := user.Name

	// Send the new activation link.
	if generateActivationLink(c, name, email); err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}

	// Delete previous activation links.
	if _, entries, err := hashstore.GetEntriesSameValueScope(c.ctx, email, REGISTER_SCOPE); err == nil {
		for _, entry := range entries {
			hashstore.DeleteEntry(c.ctx, entry.Key)
		}
	}

	w.Write([]byte("Activation link was resent."))
	return nil
}
