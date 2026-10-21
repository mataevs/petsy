// +build appengine

package petsy

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"

	"petsy/handler/json"
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

func JsonError(c *Context, code int, message string) {
	log.Println("JsonError: code=", code, "message=", message)

	err := &(struct {
		code    int
		message string
	}{
		code,
		message,
	})

	JsonResponse(c, err)
}

func JsonResponse(c *Context, object interface{}) {
	(*json.Context)(c).SetResponseObject(object)
}
