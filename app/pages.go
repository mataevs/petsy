package petsy

import (
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

func init() {

	pages := mux.NewRouter()

	pages.Handle("/register", PetsyHandler(showRegisterPage)).Methods("GET")
	pages.Handle("/login", PetsyHandler(showLoginPage)).Methods("GET")
	pages.Handle("/logout", PetsyAuthHandler(showLogoutPage)).Methods("GET")

	pages.Handle("/sitter", PetsyHandler(showAddSitter)).Methods("GET")
	pages.Handle("/sitter/{user}", PetsyHandler(showSitter)).Methods("GET")
	pages.Handle("/sitter/{user}/update", PetsyAuthHandler(showUpdateSitter)).Methods("GET")
	pages.Handle("/sitters", PetsyHandler(showSitters)).Methods("GET")

	pages.Handle("/owner", PetsyHandler(showAddOwner)).Methods("GET")
	pages.Handle("/owner/{user}", PetsyHandler(showOwner)).Methods("GET")
	pages.Handle("/owner/{user}/update", PetsyAuthHandler(showUpdateOwner)).Methods("GET")
	pages.Handle("/owners", PetsyHandler(showOwners)).Methods("GET")

	pages.Handle("/pet", PetsyHandler(showAddPet)).Methods("GET")
	pages.Handle("/owner/{user}/pet/{pet}", PetsyHandler(showPet)).Methods("GET")
	pages.Handle("/owner/{user}/pet/{pet}/update", PetsyAuthHandler(showUpdatePet)).Methods("GET")
	pages.Handle("/owner/{user}/pets", PetsyHandler(showPets)).Methods("GET")

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

}

func showSitter(c *Context, w http.ResponseWriter, r *http.Request) {

}

func showUpdateSitter(c *Context, w http.ResponseWriter, r *http.Request) {

}

func showSitters(c *Context, w http.ResponseWriter, r *http.Request) {

}

func showAddOwner(c *Context, w http.ResponseWriter, r *http.Request) {

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
