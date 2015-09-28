package petsy

import (
	"encoding/json"
	"html/template"
	"net/http"
	"petsy/user/profile"

	"github.com/gorilla/mux"

	"appengine/urlfetch"
)

func init() {

	pages := mux.NewRouter()

	pages.Handle("/register", PetsyHandler(showRegisterPage)).Methods("GET")
	pages.Handle("/login", PetsyHandler(showLoginPage)).Methods("GET")
	pages.Handle("/logout", PetsyAuthHandler(showLogoutPage)).Methods("GET")

	pages.Handle("/sitter", PetsyAuthHandler(showAddSitter)).Methods("GET")
	pages.Handle("/sitter/{userId}", PetsyHandler(showSitter)).Methods("GET")
	pages.Handle("/sitter/update", PetsyAuthHandler(showUpdateSitter)).Methods("GET")
	pages.Handle("/sitters", PetsyHandler(showSitters)).Methods("GET")

	pages.Handle("/owner", PetsyAuthHandler(showAddOwner)).Methods("GET")
	pages.Handle("/owner/{userId}", PetsyHandler(showOwner)).Methods("GET")
	pages.Handle("/owner/update", PetsyAuthHandler(showUpdateOwner)).Methods("GET")
	pages.Handle("/owners", PetsyHandler(showOwners)).Methods("GET")

	pages.Handle("/pet", PetsyAuthHandler(showAddPet)).Methods("GET")
	pages.Handle("/pet/{pet}", PetsyHandler(showPet)).Methods("GET")
	pages.Handle("/pet/{pet}/update", PetsyAuthHandler(showUpdatePet)).Methods("GET")
	pages.Handle("/owner/{userId}/pets", PetsyHandler(showPets)).Methods("GET")

	http.Handle("/", pages)
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

func showAddSitter(c *Context, w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("templates/sitter.html")
	t.Execute(w, nil)
}

func showSitter(c *Context, w http.ResponseWriter, r *http.Request) {
	ctx, _ := c.GetAppengineContext()

	vars := mux.Vars(r)
	userId := vars["userId"]

	client := urlfetch.Client(ctx)
	resp, err := client.Get("http://localhost:8080/api/sitter/" + userId)
	if err != nil {
		w.Write([]byte("Error"))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		w.Write([]byte("Error getting sitter profile"))
	}

	var sitter profile.Sitter
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&sitter); err != nil {
		w.Write([]byte("Error retrieving sitter profile."))
	}

	t, _ := template.ParseFiles("templates/sitter_template.html")
	t.Execute(w, sitter)
}

func showUpdateSitter(c *Context, w http.ResponseWriter, r *http.Request) {

}

func showSitters(c *Context, w http.ResponseWriter, r *http.Request) {

}

func showAddOwner(c *Context, w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("show add owner page"))
}

func showOwner(c *Context, w http.ResponseWriter, r *http.Request) {

}

func showUpdateOwner(c *Context, w http.ResponseWriter, r *http.Request) {

}

func showOwners(c *Context, w http.ResponseWriter, r *http.Request) {

}
func showAddPet(c *Context, w http.ResponseWriter, r *http.Request) {

}

func showPet(c *Context, w http.ResponseWriter, r *http.Request) {

}

func showUpdatePet(c *Context, w http.ResponseWriter, r *http.Request) {

}

func showPets(c *Context, w http.ResponseWriter, r *http.Request) {

}
