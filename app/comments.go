// +build appengine

package petsy

import (
	"encoding/json"
	"net/http"

	"petsy/comments"

	"github.com/gorilla/mux"
)

func getCommentFromRequest(c *Context, w http.ResponseWriter, r *http.Request) *comments.Comment {
	user, _ := c.GetUser()
	userKey, _ := c.GetUserKey()

	var comment comments.Comment

	// Get comment struct from JSON request.
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&comment); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "error decoding input json: "+err.Error())
		return nil
	}

	// Validate comment struct fields.
	if err := comment.Validate(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		JsonError(c, 101, "invalid comment data"+err.Error())
		return nil
	}

	comment.Author = user
	comment.AuthorKey = userKey

	return &comment
}

func addSitterComment(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]

	ctx, _ := c.GetAppengineContext()

	comment := getCommentFromRequest(c, w, r)
	if comment == nil {
		return
	}

	key, _ := returnSitter(c, w, userId)
	if key == nil {
		return
	}

	if _, err := comments.AddComment(ctx, comment, key); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "Error adding the comment: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getSitterComments(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]

	ctx, _ := c.GetAppengineContext()

	sitterKey, _ := returnSitter(c, w, userId)
	if sitterKey == nil {
		return
	}

	_, comments, err := comments.GetCommentsTreeForEntity(ctx, sitterKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting comments: "+err.Error())
	}

	JsonResponse(c, comments)
}

func addOwnerComment(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userEmail := vars["user"]

	ctx, _ := c.GetAppengineContext()

	comment := getCommentFromRequest(c, w, r)
	if comment == nil {
		return
	}

	key, _ := returnOwner(c, w, userEmail)
	if key == nil {
		return
	}

	if _, err := comments.AddComment(ctx, comment, key); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "Error adding the comment: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getOwnerComments(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userEmail := vars["user"]

	ctx, _ := c.GetAppengineContext()

	ownerKey, _ := returnOwner(c, w, userEmail)
	if ownerKey == nil {
		return
	}

	_, comments, err := comments.GetCommentsTreeForEntity(ctx, ownerKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting comments: "+err.Error())
	}

	JsonResponse(c, comments)
}

func addPetComment(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userEmail := vars["user"]
	petName := vars["pet"]

	ctx, _ := c.GetAppengineContext()

	comment := getCommentFromRequest(c, w, r)
	if comment == nil {
		return
	}

	key, _ := returnPet(c, w, userEmail, petName)
	if key == nil {
		return
	}

	if _, err := comments.AddComment(ctx, comment, key); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "Error adding the comment: "+err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func getPetComments(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userEmail := vars["user"]
	petName := vars["pet"]

	ctx, _ := c.GetAppengineContext()

	petKey, _ := returnPet(c, w, userEmail, petName)
	if petKey == nil {
		return
	}

	_, comments, err := comments.GetCommentsTreeForEntity(ctx, petKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		JsonError(c, 101, "error getting comments: "+err.Error())
	}

	JsonResponse(c, comments)
}
