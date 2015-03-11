// +build appengine

package petsy

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
)

func randomString(size int) (string, error) {
	if size <= 0 {
		return "", errors.New("size cannot be less than 1.")
	}

	buffer := make([]byte, size)
	_, err := rand.Read(buffer)

	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(buffer), nil
}
