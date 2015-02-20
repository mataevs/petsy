package role

import (
	"petsy/user"
)

type Role interface {
	Bio() string
	Page() string
	Pictures() []string
	User() *user.User
}

type baseRole struct {
	*user.User
	page     string
	pictures []string
	bio      string
}

func newBaseRole() baseRole {
	return baseRole{}
}

func (r baseRole) Bio() string {
	return r.bio
}

func (r baseRole) Page() string {
	return r.page
}

func (r baseRole) Pictures() []string {
	return r.pictures
}
