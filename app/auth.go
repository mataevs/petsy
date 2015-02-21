package petsy

import (
	"html/template"
	"io"
	_ "log"
	"net/http"
	"time"

	"petsy/hashstore"
	petsyuser "petsy/user"

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
	ActivationLimit, _ = Duration.ParseDuration("168h")
)

func init() {
	gomniauth.SetSecurityKey("TestSecurityKey")
	gomniauth.WithProviders(
		facebook.New(
			"605904936182713",
			"5fd71dbe58865e18ffc3f916a685b41c",
			"http://localhost:8080/auth/facebook/callback"),
		google.New(
			"494043376895-hl0dvi5jmhkprfpa354nelr77afk2546.apps.googleusercontent.com",
			"tdwi4BcpVfyq9AXwox8EQLQ5",
			"http://localhost:8080/auth/google/callback"),
	)

	auth := mux.NewRouter().PathPrefix("/auth/").Subrouter()
	auth.Handle("/facebook/login", loginHandler("facebook"))
	auth.Handle("/facebook/callback", callbackHandler("facebook"))
	auth.Handle("/google/login", loginHandler("google"))
	auth.Handle("/google/callback", callbackHandler("google"))

	auth.Handle("/register", appHandler(showRegisterPage)).Methods("GET")
	auth.Handle("/register", appHandler(register)).Methods("POST")

	auth.Handle("/login", appHandler(showLoginPage)).Methods("GET")
	auth.Handle("/login", http.HandlerFunc(login)).Methods("POST")

	auth.Handle("/logout", authReq(showLogoutPage)).Methods("GET")
	auth.Handle("/logout", authReq(logout)).Methods("POST")

	http.Handle("/auth/", auth)
}

func showRegisterPage(c *Context, w io.Writer, r *http.Request) error {
	t, _ := template.ParseFiles("templates/register.html")
	t.Execute(w, nil)

	return nil
}

func showLoginPage(c *Context, w io.Writer, r *http.Request) error {
	t, _ := template.ParseFiles("templates/login.html")
	t.Execute(w, nil)

	return nil
}

func showLogoutPage(c *Context, w io.Writer, r *http.Request) (err error, saveCookie bool) {
	t, _ := template.ParseFiles("templates/logout.html")
	t.Execute(w, nil)
	return
}

func register(c *Context, w io.Writer, r *http.Request) error {
	name := r.PostFormValue("name")
	email := r.PostFormValue("email")
	pass := r.PostFormValue("password")

	// Check if this username is already taken.
	_, user, err := petsyuser.GetUserByEmail(c.ctx, email)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}
	if user != nil {
		// todo - user email collision
		w.Write([]byte("user already exists"))
		return nil
	}

	// Create the user.
	u, err := petsyuser.NewUser(name, email)
	if err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}
	u.SetPassword(pass)

	// Add the user to the datastore.
	if _, err := petsyuser.AddUser(c.ctx, u); err != nil {
		return appErrorf(http.StatusInternalServerError, "%v", err)
	}

	// Add a validation key and send a confirmation email.
	hashstore.AddEntry(c, key, email, REGISTER_SCOPE, ActivationLimit)

	w.Write([]byte("user created"))
	return nil
}

func login(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)

	email := r.PostFormValue("email")
	pass := r.PostFormValue("password")

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

	if err = createUserSession(user, w, r); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func logout(c *Context, w io.Writer, r *http.Request) (err error, saveSession bool) {
	c.session.Options.MaxAge = -1

	w.Write([]byte("You have been logged out."))

	saveSession = true
	return
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

		if err = createUserSession(user, w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func addOrUpdateUser(c appengine.Context, name, email string, provider, providerId string) (*petsyuser.User, error) {
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

func createUserSession(user *petsyuser.User, w http.ResponseWriter, r *http.Request) error {
	c, err := NewContext(r)
	if err != nil {
		return err
	}

	c.session.Values["user"] = user.Email
	c.session.Values["login"] = time.Now().Unix()
	c.user = user

	return c.session.Save(r, w)
}
