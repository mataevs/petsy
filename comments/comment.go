// +build appengine

package comments

import (
	"time"

	"appengine"
	"appengine/datastore"

	"petsy/user"
)

const (
	CommentKind = "comments"
)

// A Comment is a descendant of another entity. Thus, a comment has an ancestor key of another entity.
// There can be a hierarchy of comments (a comment can have a parent).
type Comment struct {
	Author    *user.User `datastore:"-"`
	Email     string
	AuthorKey *datastore.Key

	Title string
	Body  string
	Date  time.Time

	Parent    *Comment `datastore:"-"`
	ParentKey *datastore.Key

	Visible bool
}

func (c Comment) Validate() error {
	return nil
}

func AddComment(c appengine.Context, comment *Comment, commentedEntityKey *datastore.Key) (*datastore.Key, error) {
	key := datastore.NewIncompleteKey(c, CommentKind, commentedEntityKey)
	return datastore.Put(c, key, comment)
}

func UpdateComment(c appengine.Context, comment *Comment, commentKey *datastore.Key) (*datastore.Key, error) {
	return datastore.Put(c, commentKey, comment)
}

func getAuthorsForComments(c appengine.Context, comments []*Comment) ([]*Comment, error) {
	authors := make(map[*datastore.Key]bool)

	// Mark the unique author keys.
	for _, comment := range comments {
		authors[comment.AuthorKey] = true
	}

	authorsUsers := make(map[*datastore.Key]*user.User)

	// Fetch the user profile for each author key.
	for authorKey, _ := range authors {
		var u user.User
		err := datastore.Get(c, authorKey, &u)
		if err != nil {
			return nil, err
		}
		authorsUsers[authorKey] = &u
	}

	// Fill in the user profile to each comment.
	for _, comment := range comments {
		comment.Author = authorsUsers[comment.AuthorKey]
	}

	return comments, nil
}

func runGetCommentsQuery(c appengine.Context, query *datastore.Query) (keys []*datastore.Key, comments []*Comment, err error) {
	for t := query.Run(c); ; {
		var comment Comment
		key, err := t.Next(&comment)
		if err == datastore.Done {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		keys = append(keys, key)
		comments = append(comments, &comment)
	}

	comments, err = getAuthorsForComments(c, comments)
	if err != nil {
		return nil, nil, err
	}
	return
}

func GetCommentsForEmail(c appengine.Context, email string) (keys []*datastore.Key, comments []*Comment, err error) {
	query := datastore.NewQuery(CommentKind).Order("-Date")
	return runGetCommentsQuery(c, query)
}

func GetCommentsForEmailPaged(c appengine.Context, email string, offset, limit int) (keys []*datastore.Key, comments []*Comment, err error) {
	query := datastore.NewQuery(CommentKind).Order("-Date").Offset(offset).Limit(limit)
	return runGetCommentsQuery(c, query)
}

func GetCommentsForEntity(c appengine.Context, commentedEntityKey *datastore.Key) (keys []*datastore.Key, comments []*Comment, err error) {
	query := datastore.NewQuery(CommentKind).Ancestor(commentedEntityKey).Order("Date")
	return runGetCommentsQuery(c, query)
}

func GetCommentsForEntityPaged(c appengine.Context, commentedEntityKey *datastore.Key, offset, limit int) (keys []*datastore.Key, comments []*Comment, err error) {
	query := datastore.NewQuery(CommentKind).Ancestor(commentedEntityKey).Order("Date").Offset(offset).Limit(limit)
	return runGetCommentsQuery(c, query)
}

func GetCommentsTreeForEntity(c appengine.Context, commentedEntityKey *datastore.Key) ([]*datastore.Key, []*Comment, error) {

	keys, comments, err := GetCommentsForEntity(c, commentedEntityKey)
	if err != nil {
		return keys, comments, err
	}

	// Create a map (key, Comment)
	commentsKeysMap := make(map[*datastore.Key]*Comment)
	for i, key := range keys {
		commentsKeysMap[key] = comments[i]
	}

	for _, comment := range comments {
		if comment.ParentKey != nil {
			comment.Parent = commentsKeysMap[comment.ParentKey]
		}
	}

	return keys, comments, err
}
