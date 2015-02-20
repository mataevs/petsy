package user

import (
	"testing"

	"appengine/aetest"
)

func TestDatastoreUsers(t *testing.T) {
	user, _ := NewUser(name, email)

	c, err := aetest.NewContext(nil)
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	// Add user
	key, err := AddUser(c, user)
	if err != nil {
		t.Errorf("AddUser: unexpected error: %v", err)
	}

	if key == nil {
		t.Errorf("AddUser: nil key")
	}

	// // Get user
	// _, gotUser, err := GetUserByEmail(c, email)
	// if err != nil {
	// 	t.Errorf("GetUserByEmail: unexpected error: %v", err)
	// }
	// if gotUser == nil {
	// 	t.Errorf("GetUserByEmail: want one user, got none")
	// 	return
	// }
	// if gotUser.Name != user.Name {
	// 	t.Errorf("GetUserByEmail: want user %s, got user %s.", user.Name, gotUser.Name)
	// }
	// if gotUser.Email != user.Email {
	// 	t.Errorf("GetUserByEmail: want user %s, got user %s.", user.Email, gotUser.Email)
	// }
}
