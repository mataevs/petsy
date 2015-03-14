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

func GetCommentsForUser(c appengine.Context, email string) (keys []*datastore.Key, comments []*Comment, err error) {
	query := datastore.NewQuery(CommentKind)

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

	// todo - fill in the author and parent pointers
	return
}

func GetCommentsForEntity(c appengine.Context, commentedEntityKey *datastore.Key) (keys []*datastore.Key, comments []*Comment, err error) {
	query := datastore.NewQuery(CommentKind).Ancestor(commentedEntityKey)

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

	// todo - fill in the author and parent pointers
	return
}
