package role

import (
	"time"
)

type commonInfo struct {
	userid    string
	Name      string
	Email     string
	Page      string
	Pictures  []string
	AvatarURL string
	Bio       string
	Birthdate time.Time
}
