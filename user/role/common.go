// +build appengine
package role

import (
	"time"

	"appengine/datastore"
)

type commonInfo struct {
	UserKey   *datastore.Key
	Name      string
	Email     string
	Page      string
	Pictures  []string
	AvatarURL string
	Bio       string
	Birthdate time.Time
}
