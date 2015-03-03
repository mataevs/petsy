package petsy

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"petsy/user"

	"appengine"
	"appengine/datastore"

	"github.com/gorilla/sessions"
)

var UnauthorizedError = errors.New("unauthorized operation")

// appResult is a response with a HTTP response code.
type appResult struct {
	error
	Code int
}

type Context struct {
	ctx     appengine.Context
	session *sessions.Session
	user    *user.User
	userKey *datastore.Key
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
		ctx.userKey, ctx.user, err = user.GetUserByEmail(c, email)
		return ctx, err
	}

	return ctx, errors.New("unexpected value in user session")
}

// appErrorf creates a new appResult encoding an error, given a response code and a message.
func appErrorf(code int, format string, args ...interface{}) *appResult {
	return &appResult{fmt.Errorf(format, args...), code}
}

// appReturn creates a new appResult storing an HTTP return code.
func appReturn(code int) *appResult {
	return &appResult{nil, code}
}

type appHandler func(c *Context, w io.Writer, r *http.Request) error

func (h appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, e := NewContext(r)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}

	// todo - catch and log error

	// Create the buffer in which the response is buffered.
	buf := &bytes.Buffer{}

	// Call the handler.
	result := h(c, buf, r)

	var code int
	var err error

	// Transform the error to appResult and fetch the code.
	if result == nil {
		code = http.StatusOK
		err = nil
	} else if res, ok := result.(*appResult); !ok {
		code = http.StatusInternalServerError
		err = errors.New("unable to cast error to appResult.")
	} else {
		code = res.Code
		err = res.error
	}

	w.WriteHeader(code)

	if err != nil {
		fmt.Fprint(w, err)
		c.ctx.Errorf(err.Error())
	} else {
		io.Copy(w, buf)
	}
}

// authReq checks that a user is logged in before executing the appHandler.
// Returns true if the session must be saved.
type authReq func(c *Context, w io.Writer, r *http.Request) (error, bool)

// authReq implements http.Handler.
func (h authReq) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, e := NewContext(r)
	if e != nil {
		http.Error(w, e.Error(), http.StatusInternalServerError)
	}

	if c.user == nil {
		http.Error(w, UnauthorizedError.Error(), http.StatusUnauthorized)
		return
	}

	// Create the buffer in which the response is buffered.
	buf := &bytes.Buffer{}

	result, saveSession := h(c, buf, r)

	var code int
	var err error

	// Transform the error to appResult and fetch the code.
	if result == nil {
		code = http.StatusOK
		err = nil
	} else if res, ok := result.(*appResult); !ok {
		code = http.StatusInternalServerError
		err = errors.New("unable to cast error to appResult.")
	} else {
		code = res.Code
		err = res.error
	}

	w.WriteHeader(code)

	if err != nil {
		fmt.Fprint(w, err)
		c.ctx.Errorf(err.Error())
	} else {
		if saveSession {
			c.session.Save(r, w)
		}
		io.Copy(w, buf)
	}
}
