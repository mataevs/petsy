package handler

import (
	"net/http"
)

type ContextMap map[string]interface{}

type Context interface {
	Set(name string, v interface{})
	Get(name string) (interface{}, bool)
}

func (c *ContextMap) Set(name string, v interface{}) {
	(*c)[name] = v
}

func (c *ContextMap) Get(name string) (interface{}, bool) {
	v, ok := (*c)[name]
	return v, ok
}

func NewContext() Context {
	ctx := ContextMap(make(map[string]interface{}))
	return &ctx
}

type Handler interface {
	Serve(c Context, rw http.ResponseWriter, r *http.Request, next ContextHandler)
}

type ContextHandler func(c Context, rw http.ResponseWriter, r *http.Request)

func (h ContextHandler) ServeContext(c Context, rw http.ResponseWriter, r *http.Request) {
	h(c, rw, r)
}

func (h ContextHandler) Serve(c Context, rw http.ResponseWriter, r *http.Request, next ContextHandler) {
	h.ServeContext(c, rw, r)

	if next != nil {
		next(c, rw, r)
	}
}

// HandlerFunc adapts a function to a Handler.
type HandlerFunc func(c Context, rw http.ResponseWriter, r *http.Request, next ContextHandler)

func (hf HandlerFunc) Serve(c Context, rw http.ResponseWriter, r *http.Request, next ContextHandler) {
	hf(c, rw, r, next)
}

type StackElem struct {
	Handler
	next *StackElem
}

func (se *StackElem) Serve(c Context, rw http.ResponseWriter, r *http.Request) {
	se.Handler.Serve(c, rw, r, se.next.Serve)
}

type Stack struct {
	first *StackElem
	last  *StackElem
}

func NewStack(first Handler, rest ...Handler) *Stack {
	firstElem := &StackElem{
		Handler: first,
		next:    nil,
	}

	stack := &Stack{
		first: firstElem,
		last:  firstElem,
	}

	for _, handler := range rest {
		stack.Add(handler)
	}

	return stack
}

func (s *Stack) Add(handler Handler) *Stack {
	elem := &StackElem{
		Handler: handler,
		next:    nil,
	}

	s.last.next = elem
	s.last = elem

	return s
}

func (s *Stack) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext()

	bufferedWriter := NewBufferedResponseWriter(w)

	// Call the handler.
	s.first.Serve(ctx, bufferedWriter, r)

	bufferedWriter.Send()
}
