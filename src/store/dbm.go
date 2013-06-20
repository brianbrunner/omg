// Package store implements nearly all of the core database and 
// database management methods
package store

import (
  "murmur"
	"store/com"
  "strings"
  "fmt"
  "data"
  "persist"
  "config"
  "strconv"
  "os"
  "encoding/gob"
  "io"
)

// DBFunc is a function that can be added to the database by
// a user
type DBFunc struct {
	function    func(*DB, []string) (string, error)
	sideEffects bool
}

// DBManager manages application state around the DB
type DBManager struct {
	dbs               []*DB
  dbChans           []chan com.Command
	ComChan           chan com.Command
	PersistChan       chan string
	PersistStateChan  chan uint8
	funcs             map[string]DBFunc
  NumDBs            int
  InSaveMode        bool
}

// Getter function for the DB
func (dbm *DBManager) GetDBs() []*DB {
  return dbm.dbs
}

func (dbm *DBManager) AddFuncWithSideEffects(key string, function func(*DB, []string) string) {
	rec_function := func(db *DB, args []string) (str string, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = r.(error)
				str = ""
			}
		}()

		str = function(db, args)
		return str, nil
	}
	dbm.funcs[key] = DBFunc{rec_function, true}
}

func (dbm *DBManager) AddFunc(key string, function func(*DB, []string) string) {
	rec_function := func(db *DB, args []string) (str string, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = r.(error)
				str = ""
			}
		}()

		str = function(db, args)
		return str, nil
	}
	dbm.funcs[key] = DBFunc{rec_function, false}
}

func (dbm *DBManager) run() {

  for i, db := range dbm.dbs {

    go func(db *DB, comChan chan com.Command) {

	    var l int
	    var args []string
	    var command com.Command

      for {

        command = <-comChan
		    l = len(command.Args)

        args = command.Args[1:l]

        db_func, ok := dbm.funcs[strings.ToLower(command.Args[0])]
        if ok {
          if str, err := db_func.function(db, args); err != nil {
            command.ReplyChan <- fmt.Sprintf("-ERR %s\r\n", err)
          } else {
            command.ReplyChan <- str
          }
          if dbm.PersistChan != nil && db_func.sideEffects {
            dbm.PersistChan <- command.CommandRaw
          }
        } else {
          command.ReplyChan <- "-ERR Unknown command\r\n"
        }

      }

    }(db, dbm.dbChans[i])

  }

  var l int
	var command com.Command

	for {
		command = <-dbm.ComChan
		l = len(command.Args)

    bucket := 0
    if l > 1 {
      bucket = murmur.HashBucket(command.Args[1],dbm.NumDBs)
    }

    dbm.dbChans[bucket] <-command
	}
}

func (dbm *DBManager) Init() {

  numDBs, err := strconv.Atoi(config.Config["dbs"])
  if err != nil {
    panic(err)
  }

  dbm.NumDBs = numDBs
  for i := 0; i < dbm.NumDBs; i += 1 {
    db := new(DB)
    db.InSaveMode = false
    db.lastDump = 0
    db.Store = make(map[string]*data.Entry)
    db.ExpiryRecord = make(map[string]uint64)
    db.ArchiveRecord = make(map[string]uint8)
    db.dumpKeyChan = make(chan string)
    dbm.dbs = append(dbm.dbs, db)
    dbm.dbChans = append(dbm.dbChans, make(chan com.Command))
  }

	dbm.ComChan = make(chan com.Command,dbm.NumDBs*4)
	dbm.funcs = make(map[string]DBFunc)
  dbm.InSaveMode = false
	go dbm.run()

}

func (dbm *DBManager) StartPersist() {
	dbm.PersistChan, dbm.PersistStateChan = persist.StartPersist(dbm.ComChan)
}

func (dbm *DBManager) SaveToDiskAsync(bucketsToDump map[uint16]uint8) error {

  if !dbm.InSaveMode {

    f, err := os.Create("/tmp/store.odb")
    if err != nil {
      return err
    }

    println("Starting save mode...")
    dbm.InSaveMode = true

    // Tell the persister to enter it's save mode state
		DefaultDBManager.PersistStateChan <- persist.PersistStateStarted
		if response := <-DefaultDBManager.PersistStateChan; response == persist.PersistStateFail {
			panic("Couldn't start save mode")
		}

    db_enc := gob.NewEncoder(f)

    db_enc.Encode(dbm.DBSize())

    go func() {
      for i, db := range dbm.dbs {
        db.StartSaveMode()
        db.SaveToDiskAsync(db_enc, bucketsToDump)
        db.EndSaveMode()
        fmt.Println("Finished dumping db",(i+1),"/",dbm.NumDBs)
      }

      err := f.Sync()
      if err == nil {

        err := os.Rename("/tmp/store.odb",config.Config["dump_file"])
        if err != nil {

          // Tell the persister that save mode has finished
          DefaultDBManager.PersistStateChan <- persist.PersistStateFailed
          if response := <-DefaultDBManager.PersistStateChan; response == persist.PersistStateFail {
            panic("Couldn't end save mode")
          }

        } else {

          // Tell the persister that save mode has finished
          DefaultDBManager.PersistStateChan <- persist.PersistStateFinished
          if response := <-DefaultDBManager.PersistStateChan; response == persist.PersistStateFail {
            panic("Couldn't end save mode")
          }

        }

      } else {

        // Tell the persister that save mode has finished
        DefaultDBManager.PersistStateChan <- persist.PersistStateFailed
        if response := <-DefaultDBManager.PersistStateChan; response == persist.PersistStateFail {
          panic("Couldn't end save mode")
        }

      }
      f.Close()


      dbm.InSaveMode = false
      println("Save mode complete...")

    }()

  }

  return nil
}

func (dbm *DBManager) SaveToDiskSync() error {
  // TODO: move save to disk code here
  return nil
}

// LoadFromDiskSync blocks execution of any commands and
// loads the database stored in the current dump file
func (dbm *DBManager) LoadFromDiskSync() error {

  f, err := os.Open("./db/store.odb")
  if err != nil {
    return err
  }
  defer f.Close()
  db_dec := gob.NewDecoder(f)

  return dbm.LoadFromGobDecoder(db_dec)

}

func (dbm *DBManager) LoadFromGobDecoder(db_dec *gob.Decoder)  error {

  var storeLen int
  err := db_dec.Decode(&storeLen)
  if err != nil {
    panic(err)
  }

  for i := 0; i < storeLen; i++ {

    var key string
    var elem *data.Entry
    err = db_dec.Decode(&key)

    if err != nil {

      if err == io.EOF {

        for _, value := range dbm.dbs[0].Store {

          for _, db := range dbm.dbs {
            db.lastDump = value.LastDump()
          }
          break

        }
        break

      } else {
        panic(err)
      }

    }

    err = db_dec.Decode(&elem)
    if err != nil {
      panic(err)
    }

    bucket := murmur.HashBucket(key,dbm.NumDBs)
    dbm.dbs[bucket].StoreSet(key, elem)

  }
  return nil
}


func (dbm *DBManager) DBSize() int {
  size := 0
  for _, db := range dbm.dbs {
    size += len(db.Store)
  }
  return size
}

var DefaultDBManager *DBManager = new(DBManager)

func init() {
  DefaultDBManager.Init()
}
