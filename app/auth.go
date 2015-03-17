// +build appengine

package petsy

import (
	"html/template"
	_ "log"
	"net/http"
	"time"

	petsyuser "petsy/user"
	. "petsy/utils"

	"github.com/gorilla/mux"
	"github.com/stretchr/gomniauth"
	"github.com/stretchr/gomniauth/common"
	"github.com/stretchr/gomniauth/providers/facebook"
	"github.com/stretchr/gomniauth/providers/google"
	"github.com/stretchr/objx"

	"appengine"
	"appengine/urlfetch"
)

const (
	REGISTER_SCOPE = "register"
)

var (
	// Activation link validity period.
	ActivationLimit, _ = time.ParseDuration("168h")
)

func init() {
	gomniauth.SetSecurityKey("TestSecurityKey")
	gomniauth.WithProviders(
		facebook.New(
			"605904936182713",
			"5fd71dbe58865e18ffc3f916a685b41c",
			"http://ro-petsy.appspot.com/auth/facebook/callback"),
		google.New(
			"494043376895-hl0dvi5jmhkprfpa354nelr77afk2546.apps.googleusercontent.com",
			"tdwi4BcpVfyq9AXwox8EQLQ5",
			"http://ro-petsy.appspot.com/auth/google/callback"),
	)

	auth := mux.NewRouter().PathPrefix("/auth/").Subrouter()

	auth.Handle("/facebook/login", loginHandler("facebook"))
	auth.Handle("/facebook/callback", callbackHandler("facebook"))
	auth.Handle("/google/login", loginHandler("google"))
	auth.Handle("/google/callback", callbackHandler("google"))

	auth.Handle("/register", PetsyHandler(showRegisterPage)).Methods("GET")
	auth.Handle("/register", PetsyHandler(register)).Methods("POST")

	auth.Handle("/login", PetsyHandler(showLoginPage)).Methods("GET")
	auth.Handle("/login", http.HandlerFunc(login)).Methods("POST")

	auth.Handle("/logout", PetsyAuthHandler(showLogoutPage)).Methods("GET")
	auth.Handle("/logout", PetsyAuthHandler(logout)).Methods("POST")

	http.Handle("/auth/", auth)
}

func showRegisterPage(c *Context, w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/register.html")
	t.Execute(w, nil)
}

func showLoginPage(c *Context, w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/login.html")
	t.Execute(w, nil)
}

func showLogoutPage(c *Context, w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/logout.html")
	t.Execute(w, nil)
}

func register(c *Context, w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	email := r.PostFormValue("email")
	pass := r.PostFormValue("password")

	if IsEmpty(name) {
		http.Error(w, "Name cannot be empty.", http.StatusForbidden)
		return
	}
	if IsEmpty(email) {
		http.Error(w, "Email cannot be empty.", http.StatusForbidden)
		return
	}
	if IsEmpty(pass) {
		http.Error(w, "Password cannot be empty.", http.StatusForbidden)
		return
	}

	ctx, _ := c.GetAppengineContext()

	// Check if this username is already taken.
	_, user, err := petsyuser.GetUserByEmail(ctx, email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user != nil {
		http.Error(w, "There is another user registered with this email.", http.StatusInternalServerError)
		return
	}

	// Create the user.
	u, err := petsyuser.NewUser(name, email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	u.SetPassword(pass)

	// Add the user to the datastore.
	if _, err := petsyuser.AddUser(ctx, u); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := generateActivationLink(c, name, email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte("user created"))

	w.WriteHeader(http.StatusCreated)
}

func login(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	email := r.PostFormValue("email")
	pass := r.PostFormValue("password")

	if IsEmpty(email) {
		http.Error(w, "Email cannot be empty.", http.StatusForbidden)
	}
	if IsEmpty(pass) {
		http.Error(w, "Password cannot be empty.", http.StatusForbidden)
	}

	_, user, err := petsyuser.GetUserByEmail(c, email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if user == nil {
		http.Error(w, "user does not exist", http.StatusForbidden)
		return
	}
	if !user.CheckPassword(pass) {
		http.Error(w, "bad password", http.StatusForbidden)
		return
	}
	if !user.Active {
		http.Error(w, "User is not activated. Please check your e-mail for the activation link.", http.StatusUnauthorized)
		return
	}

	if err = createUserSession(user.Email, w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func logout(c *Context, w http.ResponseWriter, r *http.Request) {
	session, _ := c.GetSession()
	session.Options.MaxAge = -1

	w.Write([]byte("You have been logged out."))

	// todo
	c.SetUpdateSession(true)
}

func loginHandler(providerName string) http.HandlerFunc {
	provider, err := gomniauth.Provider(providerName)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Set the urlfetch mechanism used in AppEngine
		c := appengine.NewContext(r)
		t := new(urlfetch.Transport)
		t.Context = c
		common.SetRoundTripper(t)

		state := gomniauth.NewState("after", "success")

		authUrl, err := provider.GetBeginAuthURL(state, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		http.Redirect(w, r, authUrl, http.StatusFound)
	}
}

func callbackHandler(providerName string) http.HandlerFunc {
	provider, err := gomniauth.Provider(providerName)
	if err != nil {
		panic(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Set the urlfetch mechanism used in AppEngine
		c := appengine.NewContext(r)
		t := new(urlfetch.Transport)
		t.Context = c
		common.SetRoundTripper(t)

		omap, err := objx.FromURLQuery(r.URL.RawQuery)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		creds, err := provider.CompleteAuth(omap)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		u, err := provider.GetUser(creds)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		user, err := addOrUpdateUser(c, u.Name(), u.Email(), providerName, u.IDForProvider(providerName))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err = createUserSession(user.Email, w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func addOrUpdateUser(c appengine.Context, name, email, provider, providerId string) (*petsyuser.User, error) {
	_, user, err := petsyuser.GetUserByEmail(c, email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		// User does not exist, create it.
		user, err = petsyuser.NewUser(name, email)
		if err != nil {
			return nil, err
		}
		user.Active = true
		user.AddProvider(provider, providerId)

		if _, err := petsyuser.AddUser(c, user); err != nil {
			return nil, err
		}
	} else {
		if !user.HasProvider(provider) {
			user.AddProvider(provider, providerId)
			petsyuser.UpdateUser(c, email, user)
		}
	}

	return user, nil
}

func createUserSession(email string, w http.ResponseWriter, r *http.Request) error {
	session, err := store.Get(r, "petsy")
	if err != nil {
		return err
	}

	session.Values["user"] = email
	session.Values["login"] = time.Now().Unix()

	return session.Save(r, w)
}
