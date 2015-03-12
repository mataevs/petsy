package handler

import (
	"net/http"
)

type Context interface{}

type HandlerFunc func(c *Context, rw http.ResponseWriter, r *http.Request, next HandlerFunc)

func (hf HandlerFunc) Serve(c *Context, rw http.ResponseWriter, r *http.Request, next HandlerFunc) {
	hf(c, rw, r, next)
}

type Handler interface {
	Serve(c *Context, rw http.ResponseWriter, r *http.Request, next HandlerFunc)
}

type StackElem struct {
	Handler
	next *StackElem
}

func (se *StackElem) Serve(c *Context, rw http.ResponseWriter, r *http.Request, next HandlerFunc) {
	se.Handler.Serve(c, rw, r, se.next.Serve)
}

func (se StackElem) ServeContextHTTP(c *Context, rw http.ResponseWriter, r *http.Request) {
	se.Handler.Serve(c, rw, r, se.next.Serve)
}

type Stack struct {
	first StackElem
	last  *StackElem
}

func NewStack(first Handler, rest ...Handler) *Stack {
	firstElem := StackElem{
		Handler: first,
		next:    nil,
	}

	stack := &Stack{
		first: firstElem,
		last:  &firstElem,
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
	// c, e := NewContext(r)
	// if e != nil {
	// 	http.Error(w, e.Error(), http.StatusInternalServerError)
	// }

	// todo - catch and log error

	// Create the buffer in which the response is buffered.
	// buf := &bytes.Buffer{}

	// Call the handler.
	// s.first.ServeContextHTTP(nil, buf, r)
	s.first.ServeContextHTTP(nil, w, r)

	// var code int
	// var err error

	// // Transform the error to appResult and fetch the code.
	// if result == nil {
	// 	code = http.StatusOK
	// 	err = nil
	// } else if res, ok := result.(*appResult); !ok {
	// 	code = http.StatusInternalServerError
	// 	err = errors.New("unable to cast error to appResult.")
	// } else {
	// 	code = res.Code
	// 	err = res.error
	// }

	// w.WriteHeader(code)

	// if err != nil {
	// 	fmt.Fprint(w, err)
	// 	c.ctx.Errorf(err.Error())
	// } else {
	// 	io.Copy(w, buf)
	// }
}
