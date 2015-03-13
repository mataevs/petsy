package handler

import (
	"encoding/json"
	"net/http"

	"petsy/handler"
)

const (
	JsonResponseContextName = "JsonReturnObject"
)

type Context struct {
	handler.Context
}

func (c *Context) SetResponseObject(object interface{}) {
	c.Set(JsonResponseContextName, object)
}

func (c *Context) GetResponseObject() (interface{}, bool) {
	return c.Get(JsonResponseContextName)
}

func JsonResponse(c handler.Context, rw http.ResponseWriter, r *http.Request, next handler.ContextHandler) {
	next.ServeContext(c, rw, r)

	ctx := &Context{c}

	obj, ok := ctx.GetResponseObject()
	if !ok {
		return
	}

	enc := json.NewEncoder(rw)
	enc.Encode(obj)
}
