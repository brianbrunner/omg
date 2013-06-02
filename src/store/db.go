package store 

import (
    "errors"
    "data"
    "fmt"
    "time"
    "encoding/base64"
    "persist"
    "strings"
    "bytes"
    "encoding/gob"
    "log"
    "os"
    "store/com"
    "path/filepath"
    "runtime"
    "io"
)

var usedTypes map[uint8]uint8

type DB struct {
    Store map[string]*data.Entry
    AlreadyDumped map[string]uint8
    InSaveMode bool
    Cache map[string]uint64 // this should be a b tree or something of the like
    ExpiryRecord map[string]uint64
    dumpKeyChan chan string
    lastDump uint8
}

func Base64Encode(input []byte) []byte {
    enc := base64.StdEncoding
    encLength := enc.EncodedLen(len(input))
    output := make([]byte, encLength)
    enc.Encode(output, input)
    return output
}

func Base64Decode(input []byte) []byte {
    dec := base64.StdEncoding
    decLength := dec.DecodedLen(len(input))
    output := make([]byte, decLength)
    n, err := dec.Decode(output, input)
    if err != nil {
        panic(err)
    }
    if n < decLength {
       output = output[:n] 
    }
    return output
}

/*
 * Command Handler
 */

type DBFunc struct {
    function func(*DB,[]string)(string, error)
    sideEffects bool
}

type DBManager struct {
    db *DB
    ComChan chan com.Command
    PersistChan chan string
    PersistStateChan chan uint8
    funcs map[string]DBFunc
}

func (dbm *DBManager) AddFuncWithSideEffects(key string, function func(*DB,[]string)(string)) {
    rec_function := func(db *DB,args []string) (str string, err error) {
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

func (dbm *DBManager) AddFunc(key string, function func(*DB,[]string)(string)) {
    rec_function := func(db *DB,args []string) (str string, err error) {
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
    var l int
    var args []string
    var command com.Command
    for {
        command = <- dbm.ComChan
        l = len(command.Args)
        args = command.Args[1:l]
        db_func, ok := dbm.funcs[strings.ToLower(command.Args[0])]
        if ok {
            if dbm.PersistChan != nil && db_func.sideEffects {
                dbm.PersistChan <- command.CommandRaw
            }
            if str, err := db_func.function(dbm.db,args); err != nil {
                command.ReplyChan <- fmt.Sprintf("-ERR %s\r\n",err)
            } else {
                command.ReplyChan <- str
            }
        } else {
            command.ReplyChan <- "-ERR Unknown command\r\n"
        }
    }
}


func (dbm *DBManager) Init() {
    dbm.ComChan = make(chan com.Command)
    dbm.db = new(DB)
    dbm.db.InSaveMode = false
    dbm.db.lastDump = 0
    dbm.db.Store = make(map[string]*data.Entry,100000)
    dbm.db.dumpKeyChan = make(chan string)
    dbm.funcs = make(map[string]DBFunc)
    go dbm.run()
    dbm.PersistChan, dbm.PersistStateChan = persist.StartPersist(dbm.ComChan)
}

var DefaultDBManager *DBManager = new(DBManager)
var lastMillis uint64

func Milliseconds() uint64 {
    millis := uint64(time.Now().UnixNano()/1000000)
    if millis != lastMillis {
        DefaultDBManager.PersistChan <- fmt.Sprintf("&%d\r\n",millis)
        lastMillis = millis
    }
    return millis
}

func (db *DB) LastDump() uint8 {
    return db.lastDump
}

func (db *DB) CacheSet(key string, cacheable bool) {
    if cacheable {
        db.Cache[key] = Milliseconds()
    } else {
        delete(db.Cache, key)
    }
}

func (db *DB) StartSaveMode() {
  if !db.InSaveMode {
    DefaultDBManager.PersistStateChan <- 0
    if response := <-DefaultDBManager.PersistStateChan; response == 0 {
      panic("Couldn't start save mode")
    }
    db.AlreadyDumped = make(map[string]uint8)
    db.InSaveMode = true
    if db.lastDump == 0 {
      db.lastDump = 1
    } else {
      db.lastDump = 0
    }
    fmt.Println("Starting background save")
  }
}

func (db *DB) EndSaveMode() {

  if db.InSaveMode {

    DefaultDBManager.PersistStateChan <- 2
    if response := <-DefaultDBManager.PersistStateChan; response == 0 {
      panic("Couldn't end save mode")
    }

    db.InSaveMode = false

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

    fmt.Println("Ending background save")

  }
}


func (db *DB) StoreSet(key string, entry *data.Entry) {
    entry.SetLastDump(db.lastDump)
    if db.InSaveMode {
        _, ok := db.Store[key]
        if ok && entry.LastDump() != db.lastDump {
          db.dumpKeyChan <- key
          <- db.dumpKeyChan
          db.dumpKeyChan <- " BLOCK "
          db.Store[key] = entry
          <- db.dumpKeyChan
        }
    } else {
      db.Store[key] = entry
    }
}

func (db *DB) StoreGet(key string, entryType uint8) (*data.Entry, bool, error) {
    elem, ok := db.Store[key]
    if ok {

        if db.InSaveMode {
          if elem.LastDump() != db.lastDump {
              db.dumpKeyChan <- key
              <- db.dumpKeyChan
          }
        }

        if elem.IsType(entryType) || entryType == data.Any {
            if expires, ok := db.ExpiryRecord[key]; !ok || expires > Milliseconds() {
                return elem, true, nil
            } else {
                delete(db.Store, key)
                delete(db.ExpiryRecord, key)
                return &data.Entry{}, false, nil
            }
        } else {
            return &data.Entry{}, false, errors.New("Type mismatch")
        }
    }
    return nil, false, nil
}

func (db *DB) StoreDel(key string) (bool) {
    elem, ok := db.Store[key]
    if ok {

        if db.InSaveMode {
          if elem.LastDump() != db.lastDump {
              db.dumpKeyChan <- key
              <- db.dumpKeyChan
              db.dumpKeyChan <- " BLOCK "
              delete(db.ExpiryRecord, key)
              delete(db.Store, key)
              <- db.dumpKeyChan
          }
        } else {
          delete(db.ExpiryRecord, key)
          delete(db.Store, key)
        }

    }
    return ok
}

func (db *DB) StoreKeysMatch(pattern string) ([]string) {
  keys := []string{}
  for key, _ := range db.Store {
    if ok, _ := filepath.Match(pattern, key); ok {
      keys = append(keys,key)
    }
  }
  return keys
}

func (db *DB) StoreSize() (int) {
  return len(db.Store)
}
var gobBuf bytes.Buffer
var enc *gob.Encoder
var dec *gob.Decoder

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

func (db *DB) StoreLoad(key string, bytes_rep []byte) (bool) {
    gobBuf.Reset()
    gobBuf.Write(bytes_rep)
    var elem data.Entry
    err := dec.Decode(&elem)
    if err != nil {
        log.Fatal("decode error:", err)
        return false
    } else {
        db.StoreSet(key,&elem)
        return true
    }
}

func (db *DB) SaveToDiskAsync() (error) {

    if (!db.InSaveMode) {
        return nil
    }

    go func(){

      f, err := os.Create("./db/_store.odb")
      if err != nil {
          log.Fatal("file create error:", err)
          return 
      }
      defer f.Close()

      db_enc := gob.NewEncoder(f)
      for key, elem := range db.Store {
        if elem.LastDump() == db.lastDump {
          continue
        }
        value := db.Store[key]
        for {
          escape := false
          select {
          case dump_key := <- db.dumpKeyChan:

            dump_value, ok := db.Store[dump_key]
            if ok && dump_value.LastDump() != db.lastDump {
              dump_value.SetLastDump(db.lastDump)
              err = db_enc.Encode(dump_key)
              if err != nil {
                log.Fatal("encode error:", err)
                return 
              }
              err = db_enc.Encode(dump_value)
              if err != nil {
                log.Fatal("encode error:", err)
                return 
              }
            }
            db.dumpKeyChan <- ""
          default:
            var ok bool
            value, ok = db.Store[key]
            if ok && value.LastDump() != db.lastDump {
              value.SetLastDump(db.lastDump)
              err = db_enc.Encode(key)
              if err != nil {
                log.Fatal("encode error:", err)
                return
              }
              err = db_enc.Encode(value)
              if err != nil {
                log.Fatal("encode error:", err)
                return
              }
            }
            escape = true
          }
          if escape {
            break
          }
        }
      }

      err = f.Sync()
      if err != nil {
          log.Fatal("encode error:", err)
          return
      }

      os.Rename("./db/_store.odb","./db/store.odb")

      reply := make(chan string)
      DefaultDBManager.ComChan <- com.Command{[]string{"__end_save_mode__"},"",reply}
      <-reply

    }()

    return nil
}

func (db *DB) SaveToDiskSync() (error) {
    f, err := os.Create("./db/_store.odb")
    if err != nil {
        log.Fatal("file create error:", err)
        return err
    }
    defer f.Close()
    db_enc := gob.NewEncoder(f)
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
    os.Rename("./db/_store.odb","./db/store.odb")
    return nil
}

func (db *DB) LoadFromDiskSync() (error) {
    f, err := os.Open("./db/store.odb")
    if err != nil {
        return err
    }
    defer f.Close()
    db_dec := gob.NewDecoder(f)
    for {
      var key string
      var elem *data.Entry
      err = db_dec.Decode(&key)
      if err != nil {
          if err == io.EOF {
            for _, value := range db.Store {
              db.lastDump = value.LastDump()
            }
            return nil
          } else {
            fmt.Println("decode error:", err)
          }
      }
      err = db_dec.Decode(&elem)
      if err != nil {
        panic("decode error!")
      }
      db.StoreSet(key, elem)
    }
}

func (dbm *DBManager) LoadFromDiskSync() (error) {
  return dbm.db.LoadFromDiskSync()
}

func MemInUse() (uint64) {
  var memStats runtime.MemStats
  runtime.ReadMemStats(&memStats)
  return memStats.HeapAlloc
}

func RegisterStoreType(entryNum uint8) {
    _, ok := usedTypes[entryNum]
    if !ok {
        usedTypes[entryNum] = 1
    } else {
        panic(fmt.Sprintf("You've used the same identifier, \"%i\", for multiple types",entryNum))
    } 
}

func init() {
    usedTypes = make(map[uint8]uint8)
    enc = gob.NewEncoder(&gobBuf)
    dec = gob.NewDecoder(&gobBuf)
    DefaultDBManager.Init()
}
