package petsy

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"petsy/user"

	"appengine"

	"github.com/gorilla/sessions"
)

var UnauthorizedError = errors.New("unauthorized operation")

// appError is an error with a HTTP response code.
type appError struct {
	error
	Code int
}

type Context struct {
	ctx     appengine.Context
	session *sessions.Session
	user    *user.User
}

func NewContext(r *http.Request) (*Context, error) {
	sess, err := store.Get(r, "petsy")
	if err != nil {
		log.Println(err)
	}

	c := appengine.NewContext(r)

	ctx := &Context{
		ctx:     c,
		session: sess,
	}
	if err != nil {
		return ctx, err
	}

	log.Println(sess.Values["user"])
	log.Println(sess.IsNew)

	if sess.Values["user"] == nil {
		return ctx, nil
	}
	if email, ok := sess.Values["user"].(string); ok {
		_, user, err := user.GetUserByEmail(c, email)

		ctx.user = user
		return ctx, err
	}

	return ctx, errors.New("unexpected value in user session")
}

// appErrorf creates a new appError given a response code and a message.
func appErrorf(code int, format string, args ...interface{}) *appError {
	return &appError{fmt.Errorf(format, args...), code}
}

type appHandler func(c *Context, w io.Writer, r *http.Request) error

func (h appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, _ := NewContext(r)
	// todo - catch and log error

	buf := &bytes.Buffer{}
	err := h(c, buf, r)
	if err == nil {
		io.Copy(w, buf)
		return
	}

	code := http.StatusInternalServerError
	logf := c.ctx.Errorf
	if err, ok := err.(*appError); ok {
		code = err.Code
		logf = c.ctx.Infof
	}

	w.WriteHeader(code)
	logf(err.Error())
	fmt.Fprint(w, err)
}

// authReq checks that a user is logged in before executing the appHandler.
// Returns true if the session must be saved.
type authReq func(c *Context, w io.Writer, r *http.Request) (error, bool)

// authReq implements http.Handler.
func (h authReq) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := NewContext(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	if c.user == nil {
		http.Error(w, UnauthorizedError.Error(), http.StatusUnauthorized)
		return
	}

	buf := &bytes.Buffer{}
	err, saveSession := h(c, buf, r)
	if err == nil {
		if saveSession {
			c.session.Save(r, w)
		}
		io.Copy(w, buf)
		return
	}

	code := http.StatusInternalServerError
	logf := c.ctx.Errorf
	if err, ok := err.(*appError); ok {
		code = err.Code
		logf = c.ctx.Infof
	}

	w.WriteHeader(code)
	logf(err.Error())
	fmt.Fprint(w, err)
}

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
