// +build appengine

package petsy

import (
	"errors"
	"net/http"

	"petsy/handler"
	"petsy/handler/json"
	"petsy/user"

	"appengine"
	"appengine/datastore"

	"github.com/gorilla/sessions"
)

var UnauthorizedError = errors.New("unauthorized operation")

const (
	AppengineContextName = "AppengineContext"
	SessionName          = "Session"
	UserName             = "User"
	UserKeyName          = "UserKey"
	UpdateSessionName    = "UpdateSession"
)

type Context struct {
	handler.Context
}

func (c *Context) SetAppengineContext(ctx appengine.Context) {
	c.Set(AppengineContextName, ctx)
}

func (c *Context) GetAppengineContext() (appengine.Context, bool) {
	obj, ok := c.Get(AppengineContextName)
	if !ok {
		return nil, false
	}

	ctx, ok := obj.(appengine.Context)

	return ctx, ok
}

func (c *Context) SetSession(sess *sessions.Session) {
	c.Set(SessionName, sess)
}

func (c *Context) GetSession() (*sessions.Session, bool) {
	obj, ok := c.Get(SessionName)
	if !ok {
		return nil, false
	}

	sess, ok := obj.(*sessions.Session)
	return sess, ok
}

func (c *Context) SetUser(user *user.User) {
	c.Set(UserName, user)
}

func (c *Context) GetUser() (*user.User, bool) {
	obj, ok := c.Get(UserName)
	if !ok {
		return nil, false
	}

	user, ok := obj.(*user.User)
	return user, ok
}

func (c *Context) SetUserKey(key *datastore.Key) {
	c.Set(UserKeyName, key)
}

func (c *Context) GetUserKey() (*datastore.Key, bool) {
	obj, ok := c.Get(UserKeyName)
	if !ok {
		return nil, false
	}

	key, ok := obj.(*datastore.Key)
	return key, ok
}

func (c *Context) SetUpdateSession(update bool) {
	c.Set(UpdateSessionName, update)
}

func (c *Context) GetUpdateSession() (bool, bool) {
	obj, ok := c.Get(UpdateSessionName)
	if !ok {
		return false, false
	}

	update, ok := obj.(bool)
	return update, ok
}

func NewContext(c handler.Context, r *http.Request) (*Context, error) {
	ctx := &Context{c}

	sess, err := store.Get(r, "petsy")
	if err != nil {
		return nil, err
	}

	ctx.SetSession(sess)

	appengineCtx := appengine.NewContext(r)
	ctx.SetAppengineContext(appengineCtx)

	if sess.Values["user"] == nil {
		return ctx, nil
	}
	if email, ok := sess.Values["user"].(string); ok {
		key, u, err := user.GetUserByEmail(appengineCtx, email)
		if err != nil {
			return ctx, err
		}

		ctx.SetUser(u)
		ctx.SetUserKey(key)
	}

	return ctx, nil
}

func BaseHandler(c handler.Context, rw http.ResponseWriter, r *http.Request, next handler.ContextHandler) {
	ctx, err := NewContext(c, r)

	if err != nil {
		http.Error(rw, err.Error(), http.StatusInternalServerError)
		return
	}

	next(ctx, rw, r)
}

func AuthHandler(c handler.Context, rw http.ResponseWriter, r *http.Request, next handler.ContextHandler) {
	ctx := c.(*Context)

	if user, ok := ctx.GetUser(); !ok || user == nil {
		http.Error(rw, UnauthorizedError.Error(), http.StatusUnauthorized)
		return
	}

	next(ctx, rw, r)

	if update, ok := ctx.GetUpdateSession(); update && ok {
		if session, ok := ctx.GetSession(); ok {
			session.Save(r, rw)
		}
	}
}

func PetsyHandler(handlerFunc func(*Context, http.ResponseWriter, *http.Request)) *handler.Stack {
	return handler.NewStack(
		handler.HandlerFunc(BaseHandler),
		handler.ContextHandler(
			func(c handler.Context, rw http.ResponseWriter, r *http.Request) {
				ctx := c.(*Context)
				handlerFunc(ctx, rw, r)
			}))
}

func PetsyAuthHandler(handlerFunc func(*Context, http.ResponseWriter, *http.Request)) *handler.Stack {
	return handler.NewStack(
		handler.HandlerFunc(BaseHandler),
		handler.HandlerFunc(AuthHandler),
		handler.ContextHandler(
			func(c handler.Context, rw http.ResponseWriter, r *http.Request) {
				ctx := c.(*Context)
				handlerFunc(ctx, rw, r)
			}))
}

func PetsyJsonHandler(handlerFunc func(*Context, http.ResponseWriter, *http.Request)) *handler.Stack {
	return handler.NewStack(
		handler.HandlerFunc(json.JsonResponse),
		handler.HandlerFunc(BaseHandler),
		handler.ContextHandler(
			func(c handler.Context, rw http.ResponseWriter, r *http.Request) {
				ctx := c.(*Context)
				handlerFunc(ctx, rw, r)
			}))
}

func PetsyAuthJsonHandler(handlerFunc func(*Context, http.ResponseWriter, *http.Request)) *handler.Stack {
	return handler.NewStack(
		handler.HandlerFunc(json.JsonResponse),
		handler.HandlerFunc(BaseHandler),
		handler.HandlerFunc(AuthHandler),
		handler.ContextHandler(
			func(c handler.Context, rw http.ResponseWriter, r *http.Request) {
				ctx := c.(*Context)
				handlerFunc(ctx, rw, r)
			}))
}
