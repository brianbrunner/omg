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
)

var usedTypes map[int]string

type DB struct {
    Store map[string]*data.Entry
    Cache map[string]uint64 // this should be a b tree or something of the like
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

type DBManager struct {
    db *DB
    ComChan chan com.Command
    PersistChan chan string
    funcs map[string]func(*DB,[]string)(string)
}

func (dbm *DBManager) AddFunc(key string, function func(*DB,[]string)(string)) {
    dbm.funcs[key] = function
}

func (dbm *DBManager) run() {
    fmt.Println("DB is open for business")
    var l int
    var args []string
    var command com.Command
    for {
        command = <- dbm.ComChan
        l = len(command.Args)
        args = command.Args[1:l]
        function, ok := dbm.funcs[strings.ToLower(command.Args[0])]
        if ok {
            command.ReplyChan <- function(dbm.db,args)
            if dbm.PersistChan != nil {
                dbm.PersistChan <- command.CommandRaw
            }
        } else {
            command.ReplyChan <- "-ERR Unknown command\r\n"
        }
    }
}


func (dbm *DBManager) Init() {
    dbm.ComChan = make(chan com.Command,10000)
    dbm.db = new(DB)
    dbm.db.Store = make(map[string]*data.Entry,100000)
    dbm.funcs = make(map[string]func(*DB,[]string)(string))
    go dbm.run()
    persist.LoadAppendOnlyFile()

    dbm.PersistChan = persist.StartPersist(dbm.ComChan)
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

func (db *DB) CacheSet(key string, cacheable bool) {
    if cacheable {
        db.Cache[key] = Milliseconds()
    } else {
        delete(db.Cache, key)
    }
}

func (db *DB) StoreSet(key string, entry *data.Entry) {
    db.Store[key] = entry
}

func (db *DB) StoreGet(key string, entryType int) (*data.Entry, bool, error) {
    elem, ok := db.Store[key]
    if ok {
        if elem.EntryType == entryType || entryType == data.Any {
            if elem.Expires == 0 || elem.Expires > Milliseconds() {
                return elem, true, nil
            } else {
                delete(db.Store, key)
                return &data.Entry{}, false, nil
            }
        } else {
            return &data.Entry{}, false, errors.New("Type mismatch")
        }
    }
    return &data.Entry{}, false, nil
}

func (db *DB) StoreDel(key string) (bool) {
    _, ok := db.Store[key]
    if ok {
        delete(db.Store, key)
    }
    return ok
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

func (db *DB) SaveToDiskSync() (error) {
    f, err := os.Create("./db/store.odb")
    defer f.Close()
    if err != nil {
        log.Fatal("encode error:", err)
        return err
    }
    db_enc := gob.NewEncoder(f)
    err = db_enc.Encode(db)
    if err != nil {
        log.Fatal("encode error:", err)
        return err
    }
    return nil
}

func (db *DB) LoadFromDiskSync() (error) {
    f, err := os.Open("./db/store.odb")
    if err != nil {
        return err
    }
    defer f.Close()
    db_dec := gob.NewDecoder(f)
    err = db_dec.Decode(&db)
    if err != nil {
        log.Fatal("decode error:", err)
        return err
    } else {
        return nil
    }
}

func RegisterPrefixedStoreType(entryNum int, prefix string) {
    _, ok := usedTypes[entryNum]
    if !ok {
        usedTypes[entryNum] = prefix
    } else {
        panic(fmt.Sprintf("You've used the same identifier, \"%i\", for multiple types",entryNum))
    } 
}

func RegisterStoreType(entryNum int) {
    RegisterPrefixedStoreType(entryNum, "")
}

func init() {
    usedTypes = make(map[int]string)
    enc = gob.NewEncoder(&gobBuf)
    dec = gob.NewDecoder(&gobBuf)
    DefaultDBManager.Init()
}
