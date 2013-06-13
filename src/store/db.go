// Package store implements nearly all of the core database and 
// database management methods
package store

import (
	"bytes"
	"data"
	"encoding/gob"
	"errors"
	"log"
	"path/filepath"
)

// DB handles object storage as well as overall database state. It
// also manages database dumps via snapshotting.
type DB struct {

  // Store is the map that holds all user data
	Store         map[string]*data.Entry

  // InSaveMode determines whether the database is currently being dumped
	InSaveMode    bool

  // ArchiveRecord keeps track of any keys that are currently archived to disk
	ArchiveRecord map[string]uint8

  // ExpiryRecord keeps track of any keys with an expiration time set
	ExpiryRecord  map[string]uint64

  // dumpKeyChan is used internally to force the dump of a particular key while in
  // save mode
	dumpKeyChan   chan string

  // lastDump Keeps track of the signature that was left by the last dump
  // of the database
	lastDump      uint8

}

// LastDump returns the signature of the last DB dump
func (db *DB) LastDump() uint8 {
	return db.lastDump
}

// SetLastDump sets the signature for the last DB dump
func (db *DB) SetLastDump(lastDump uint8) {
  db.lastDump = lastDump
}

//StoreSet sets a key-value pair in the databse
func (db *DB) StoreSet(key string, entry *data.Entry) {

  // Make sure the entry coming in has the correct dump signature
	entry.SetLastDump(db.lastDump)

	if db.InSaveMode {

    // if the key is already in the database, make sure we dump
    // it before we commit the modified version to the DB
		_, ok := db.Store[key]
		if ok && entry.LastDump() != db.lastDump {
			db.dumpKeyChan <- key
			<-db.dumpKeyChan
		}

    // Block the loop iterating over the store map so we can
    // modify the store map
    db.dumpKeyChan <- " BLOCK "
    db.Store[key] = entry
    <-db.dumpKeyChan

	} else {

    // set the key in the DB
		db.Store[key] = entry

	}

}

// StoreGet retrieves a key from the database, making sure that the key
// isn't archived or expired
func (db *DB) StoreGet(key string, entryType uint8) (*data.Entry, bool, error) {

  // check to see if the key we're retrieving has been archived
	_, ok := db.ArchiveRecord[key]
	if ok {
		return nil, false, errors.New("ARCHIVED")
	}

  // get the value for the key
	elem, ok := db.Store[key]
	if ok {

    // if the db is in save mode, dump the key before we return it
		if db.InSaveMode {
			if elem.LastDump() != db.lastDump {
				db.dumpKeyChan <- key
				<-db.dumpKeyChan
			}
		}

    // make sure the key we've retreived is of the correct type
		if elem.IsType(entryType) || entryType == data.Any {

      // make sure the key isn't expired. delete it if it is
			if expires, ok := db.ExpiryRecord[key]; !ok || expires > Milliseconds() {
				return elem, true, nil
			} else {
				delete(db.Store, key)
				delete(db.ExpiryRecord, key)
				return &data.Entry{}, false, nil
			}

		} else {

			return nil, false, errors.New("Type mismatch")

		}

	}
	return nil, false, nil
}

// StoreDel deletes a key from the db
func (db *DB) StoreDel(key string) bool {

  // check if the key is actually in the database
	elem, ok := db.Store[key]
	if ok {

    // if we're in save mode, dump the key before we delete it
		if db.InSaveMode {

			if elem.LastDump() != db.lastDump {

        // dump the key
				db.dumpKeyChan <- key
				<-db.dumpKeyChan

        // block the execution of our dump loop while we modify
        // the store map
				db.dumpKeyChan <- " BLOCK "
				delete(db.ExpiryRecord, key)
				delete(db.Store, key)
				<-db.dumpKeyChan

			}

		} else {

			delete(db.ExpiryRecord, key)
			delete(db.Store, key)

		}

	}
	return ok
}

// StoreKeysMatch find all keys that match the specified pattern
func (db *DB) StoreKeysMatch(pattern string) []string {
	keys := []string{}
	for key, _ := range db.Store {
		if ok, _ := filepath.Match(pattern, key); ok {
			keys = append(keys, key)
		}
	}
	return keys
}

// StoreSize returns the number of keys in the DB
func (db *DB) StoreSize() int {
	return len(db.Store)
}

// Flush empties all keys from the DB
func (db *DB) Flush() {
	db.Store = make(map[string]*data.Entry)
  // TODO: delete the store files, maybe?
}

var gobBuf bytes.Buffer
var enc *gob.Encoder
var dec *gob.Decoder

// StoreDump dumps the given key to a byte string
func (db *DB) StoreDump(key string) ([]byte, bool) {
	elem, ok, _ := db.StoreGet(key, data.Any)
	if ok {
		gobBuf.Reset()
		err := enc.Encode(elem)
		if err != nil {
			log.Fatal("encode error:", err)
			return []byte{}, false
		} else {
			return gobBuf.Bytes(), true
		}
	} else {
		return []byte{}, false
	}
}

// StoreLoad loads the given key from the supplied byte string. Currently
// does not work correctly.
func (db *DB) StoreLoad(key string, bytes_rep []byte) bool {
	gobBuf.Reset()
	gobBuf.Write(bytes_rep)
	var elem data.Entry
	err := dec.Decode(&elem)
	if err != nil {
		log.Fatal("decode error:", err)
		return false
	} else {
		db.StoreSet(key, &elem)
		return true
	}
}

func init() {
	enc = gob.NewEncoder(&gobBuf)
	dec = gob.NewDecoder(&gobBuf)
}
