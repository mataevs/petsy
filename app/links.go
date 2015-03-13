// +build appengine

package petsy

import (
	"html/template"
	"net/http"

	"petsy/hashstore"
	"petsy/mailer"
	petsyuser "petsy/user"
	. "petsy/utils"

	"github.com/gorilla/mux"
)

func init() {
	links := mux.NewRouter().PathPrefix("/links/").Subrouter()

	links.Handle("/verification", PetsyHandler(verifyLink)).Methods("GET")

	links.Handle("/resend-activation-link", PetsyHandler(showResendActivationLink)).Methods("GET")
	links.Handle("/resend-activation-link", PetsyHandler(resendActivationLink)).Methods("POST")

	http.Handle("/links/", links)
}

func verifyLink(c *Context, w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()

	hash := queryValues.Get("hash")
	scope := queryValues.Get("scope")
	email := queryValues.Get("email")

	ctx, _ := c.GetAppengineContext()

	// Check query parameters values.
	if IsEmpty(hash) || IsEmpty(scope) || IsEmpty(email) {
		http.Error(w, "Link does not exist.", http.StatusNotFound)
		return
	}

	// Check validation link validity.
	if valid, err := hashstore.IsValidEntry(ctx, hash, email, scope); err == hashstore.NoSuchKeyErr {
		http.Error(w, "Link does not exist.", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if !valid {
		w.Write([]byte("Confirmation link has expired.\n"))
		w.Write([]byte("Click <a href=\"/links/resend-activation-link.html\">here</a> for a new activation link.\n"))
		return
	}

	// Get user from datastore.
	_, user, err := petsyuser.GetUserByEmail(ctx, email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if IsEmpty(user) {
		http.Error(w, "No user found.", http.StatusNotFound)
		return
	}

	// Perform action depending on scope.
	switch scope {
	case "register":
		user.Active = true
		// Mark the user as active.
		if _, err := petsyuser.UpdateUser(ctx, user.Email, user); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Delete hashstore entry.
		if hashstore.DeleteEntry(ctx, hash); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte("User account activated. You can now login."))
		return
	default:
		http.Error(w, "Unknown scope.", http.StatusUnauthorized)
		return
	}
}

func showResendActivationLink(c *Context, w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/resend-activation-link.html")
	t.Execute(w, nil)
}

func resendActivationLink(c *Context, w http.ResponseWriter, r *http.Request) {
	email := r.PostFormValue("email")
	pass := r.PostFormValue("password")

	if IsEmpty(email) {
		http.Error(w, "Email address cannot be empty.", http.StatusForbidden)
		return
	}
	if IsEmpty(pass) {
		http.Error(w, "Password cannot be empty.", http.StatusForbidden)
		return
	}

	ctx, _ := c.GetAppengineContext()

	// Get user by email.
	_, user, err := petsyuser.GetUserByEmail(ctx, email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "Non-existent user or bad password.", http.StatusForbidden)
		return
	}
	if !user.CheckPassword(pass) {
		http.Error(w, "Non-existent user or bad password.", http.StatusForbidden)
		return
	}

	if user.Active {
		http.Error(w, "User is already activated.", http.StatusForbidden)
		return
	}

	name := user.Name

	// Send the new activation link.
	if generateActivationLink(c, name, email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete previous activation links.
	if _, entries, err := hashstore.GetEntriesSameValueScope(ctx, email, REGISTER_SCOPE); err == nil {
		for _, entry := range entries {
			hashstore.DeleteEntry(ctx, entry.Key)
		}
	}

	w.Write([]byte("Activation link was resent."))
}

func generateActivationLink(c *Context, name, email string) error {
	ctx, _ := c.GetAppengineContext()

	// Add a validation key and send a confirmation email.
	key, err := randomString(32)
	if err != nil {
		return err
	}
	hashstore.AddEntry(ctx, key, email, REGISTER_SCOPE, ActivationLimit)

	// Send confirmation email
	message := "http://petsy-ro.appspot.com" +
		"/links/verification?" + "hash=" + key +
		"&scope=" + REGISTER_SCOPE +
		"&email=" + email

	if err := mailer.SendEmail(ctx,
		[]string{email},
		"noreply@petsy-ro.appspotmail.com",
		"Petsy.ro - Account Details for "+name,
		message); err != nil {
		return err
	}

	return nil
}
