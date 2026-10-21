// +build appengine

package petsy

import (
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
}
