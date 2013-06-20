// Package store implements nearly all of the core database and 
// database management methods
package store

import (
	"encoding/gob"
	"log"
	"os"
  "murmur"
)

// SaveToDiskAsync asynchronously save the current database to disk. It
// use gob encoding to sequentially dump each key. It listens to an internal
// dump channel that allows other DB methods to either force the dump of a key
// so that it can be modified during a save or to block execution of the loop
// so that the underlying map can be modified via store set.
func (db *DB) SaveToDiskAsync(db_enc *gob.Encoder, bucketsToDump map[uint16]uint8) error {

	if !db.InSaveMode {
		return nil
	}

  var err error
  for key, elem := range db.Store {
    if elem.LastDump() == db.lastDump {
      continue
    }
    value := db.Store[key]
    for {
      escape := false
      select {
      case dump_key := <-db.dumpKeyChan:

        if bucketsToDump != nil {
          bucket := murmur.HashBucket(dump_key,16384)
          if _, ok := bucketsToDump[uint16(bucket)]; !ok {
            db.dumpKeyChan <- ""
            continue
          }
        }

        dump_value, ok := db.Store[dump_key]
        if ok && dump_value.LastDump() != db.lastDump {
          dump_value.SetLastDump(db.lastDump)
          err = db_enc.Encode(dump_key)
          if err != nil {
            log.Fatal("encode error:", err)
            return err
          }
          err = db_enc.Encode(dump_value)
          if err != nil {
            log.Fatal("encode error:", err)
            return err
          }
        }
        db.dumpKeyChan <- ""
      default:

        if bucketsToDump != nil {
          bucket := murmur.HashBucket(key,16384)
          if _, ok := bucketsToDump[uint16(bucket)]; !ok {
            escape = true
            break
          }
        }

        var ok bool
        value, ok = db.Store[key]
        if ok && value.LastDump() != db.lastDump {
          value.SetLastDump(db.lastDump)
          err = db_enc.Encode(key)
          if err != nil {
            log.Fatal("encode error:", err)
            return err
          }
          err = db_enc.Encode(value)
          if err != nil {
            log.Fatal("encode error:", err)
            return err
          }
        }
        escape = true
      }
      if escape {
        break
      }
    }
  }

  return nil

}

// SaveToDiskSync saves the current database to disk and blocks
// execution of all other commands. It's output is equivalent to
// the output of an asynchronous save.
func (db *DB) SaveToDiskSync() error {
  f, err := os.Create("./db/_store.odb")
  if err != nil {
    log.Fatal("file create error:", err)
    return err
  }
  defer f.Close()
  db_enc := gob.NewEncoder(f)
  db_enc.Encode(db.StoreSize())
  for key, value := range db.Store {
    err = db_enc.Encode(key)
    if err != nil {
      log.Fatal("encode error:", err)
      return err
    }
    err = db_enc.Encode(value)
    if err != nil {
      log.Fatal("encode error:", err)
      return err
    }
  }
  err = f.Sync()
  if err != nil {
    log.Fatal("encode error:", err)
    return err
  }
  os.Rename("./db/_store.odb", "./db/store.odb")
  return nil
}

// StartSaveMode puts the DB into a state in which it can be saved
func (db *DB) StartSaveMode() {
	if !db.InSaveMode {

    // Turn save mode on, swap the value of our last dump
		db.InSaveMode = true
		if db.lastDump == 0 {
			db.lastDump = 1
		} else {
			db.lastDump = 0
		}

	}
}

// EndSaveMode ends save mode, cleaning up any state that was
// created by save mode
func (db *DB) EndSaveMode() {
	if db.InSaveMode {

		db.InSaveMode = false

    // If any clients tried to force a dump after we finished dumping
    // but before save mode was turned off, return control to those clients
		escape := false
		for {
			select {
			case <-db.dumpKeyChan:
				db.dumpKeyChan <- ""
			default:
				escape = true
			}
			if escape {
				break
			}
		}

	}
}

