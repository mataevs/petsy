// +build appengine

// Package hashstore implements a hash store over Appengine's datastore.
// The hashstore holds expirable (key, value, scope) string tuples.
package hashstore

import (
	"errors"
	"time"

	"appengine"
	"appengine/datastore"
)

const HashKind = "hashstore"

// Error returned if no entry contains the specified key.
var NoSuchKeyErr = errors.New("no entry with this key found")

// Error returned if there is another entry with the specified key.
var DuplicateKeyErr = errors.New("key already exists")

// Entry represents an expirable (key, value, scope) tuple stored in the hashstore.
type Entry struct {
	// The key of the entry. Each key has to be unique across the store.
	Key string `datastore:"key"`
	// The value associated with the key.
	Value string `datastore:"value"`
	// The scope can be used for user-defined logic.
	Scope string `datastore:"scope"`
	// Generated represents the time when the entry was inserted. Its value represents
	// the starting time of the validity period.
	Generated time.Time `datastore:"generated"`
	// Valid defines the validity duration of the entry.
	Valid time.Duration `datastore:"valid"`
}

// AddEntry adds a new entry to the datastore. Takes a key, a value, a scope and a validity duration.
// Builds a new Entry with the current time of the system as the value for Generated.
// Returns DuplicateKeyErr if the datastore already contains the key.
func AddEntry(c appengine.Context, key string, value string, scope string, valid time.Duration) (*datastore.Key, error) {
	if key == "" {
		return nil, errors.New("key can't be empty")
	}
	if value == "" {
		return nil, errors.New("value can't be empty")
	}
	if scope == "" {
		return nil, errors.New("scope can't be empty")
	}
	if valid.Nanoseconds() <= 0 {
		return nil, errors.New("duration must be positive")
	}

	// Check if the key is unique.
	if _, _, err := GetValue(c, key); err != NoSuchKeyErr {
		return nil, DuplicateKeyErr
	}

	// Construct the entry and add it to the datastore.
	entry := &Entry{
		Value:     value,
		Key:       key,
		Scope:     scope,
		Generated: time.Now(),
		Valid:     valid,
	}

	dsKey := datastore.NewIncompleteKey(c, HashKind, nil)

	return datastore.Put(c, dsKey, entry)
}

// GetValue interrogates the datastore for the Entry with the specified key.
// Returns the entry with the associated key and the entry's datastore key.
// Returns NoSuchKeyErr if the key is not found.
func GetValue(c appengine.Context, key string) (*datastore.Key, *Entry, error) {
	if key == "" {
		return nil, nil, errors.New("key can't be empty")
	}

	query := datastore.NewQuery(HashKind).Filter("key =", string(key))

	for t := query.Run(c); ; {
		var entry Entry
		key, err := t.Next(&entry)
		if err == datastore.Done {
			return nil, nil, NoSuchKeyErr
		}
		if err != nil {
			return nil, nil, err
		}
		return key, &entry, nil
	}

	return nil, nil, nil
}

func GetEntriesSameValueScope(c appengine.Context, value string, scope string) ([]*datastore.Key, []*Entry, error) {
	if value == "" {
		return nil, nil, errors.New("value can't be empty")
	}
	if scope == "" {
		return nil, nil, errors.New("scope can't be empty")
	}

	query := datastore.NewQuery(HashKind).
		Filter("value =", string(value)).
		Filter("scope =", string(scope))

	keys := make([]*datastore.Key, 0)
	entries := make([]*Entry, 0)

	for t := query.Run(c); ; {
		var entry Entry
		key, err := t.Next(&entry)
		if err == datastore.Done {
			return keys, entries, nil
		}
		if err != nil {
			return nil, nil, err
		}

		keys = append(keys, key)
		entries = append(entries, &entry)
	}
}

// IsValidEntry checks whether the parameters are found in a valid stored entry.
// If no entry is found, NoSuchKeyErr is returned as an error.
// If the entry is valid, the boolean value returned is true. If the entry is expired, the returned
// value will be false.
func IsValidEntry(c appengine.Context, key string, value string, scope string) (bool, error) {
	if key == "" {
		return false, errors.New("key can't be empty")
	}

	query := datastore.NewQuery(HashKind).
		Filter("key =", string(key)).
		Filter("value =", string(value)).
		Filter("scope =", string(scope))

	for t := query.Run(c); ; {
		var entry Entry
		_, err := t.Next(&entry)
		if err == datastore.Done {
			return false, NoSuchKeyErr
		}
		if err != nil {
			return false, err
		}

		if entry.Generated.Add(entry.Valid).After(time.Now()) {
			return true, nil
		}
		return false, nil
	}

	return false, nil
}

// DeleteEntry deletes the entry associated with the specified key.
// Returns NoSuchKeyErr if there is no entry with the specified key.
func DeleteEntry(c appengine.Context, key string) error {
	dsKey, _, err := GetValue(c, key)
	if err != nil {
		return err
	}

	return datastore.Delete(c, dsKey)
}

// PurgeExpiredEntries performs a cleanup on the Hashstore, by deleting
// the entries expired a while ago (duration defined by threshold).
func PurgeExpiredEntries(c appengine.Context, threshold time.Duration) error {
	if threshold <= 0 {
		return errors.New("threshold must be positive")
	}

	keysToDelete := make([]*datastore.Key, 0)

	query := datastore.NewQuery(HashKind)

	for t := query.Run(c); ; {
		var entry Entry
		key, err := t.Next(&entry)
		if err == datastore.Done {
			break
		}
		if err != nil {
			break
		}
		if entry.Generated.Add(entry.Valid).Add(threshold).Before(time.Now()) {
			keysToDelete = append(keysToDelete, key)
		}
	}

	return datastore.DeleteMulti(c, keysToDelete)
}
