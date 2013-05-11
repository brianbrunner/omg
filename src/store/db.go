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
)

type DB struct {
    store map[string]*data.Entry
    cache map[string]uint64 // this should be a b tree or something of the like
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

type Command struct {
    Args []string
    CommandRaw string
    ReplyChan chan string
}

type DBManager struct {
    db *DB
    ComChan chan Command
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
    var command Command
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
    dbm.ComChan = make(chan Command,10000)
    dbm.db = new(DB)
    dbm.db.store = make(map[string]*data.Entry,100000)
    dbm.funcs = make(map[string]func(*DB,[]string)(string))
    go dbm.run()
    persist.LoadAppendOnlyFile()
    dbm.PersistChan = persist.StartPersist()
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
        db.cache[key] = Milliseconds()
    } else {
        delete(db.cache, key)
    }
}

func (db *DB) StoreSet(key string, entry *data.Entry) {
    db.store[key] = entry
}

func (db *DB) StoreGet(key string, entryType int) (*data.Entry, bool, error) {
    elem, ok := db.store[key]
    if ok {
        if elem.EntryType == entryType || entryType == data.Any {
            if elem.Expires == 0 || elem.Expires > Milliseconds() {
                return elem, true, nil
            } else {
                delete(db.store, key)
                return &data.Entry{}, false, nil
            }
        } else {
            return &data.Entry{}, false, errors.New("Type mismatch")
        }
    }
    return &data.Entry{}, false, nil
}

func (db *DB) StoreDel(key string) (bool) {
    _, ok := db.store[key]
    if ok {
        delete(db.store, key)
    }
    return ok
}

var gobBuf bytes.Buffer
var enc *gob.Encoder
var dec *gob.Decoder

func (db *DB) StoreBase64Dump(key string) (string, bool) {
    elem, ok, _ := db.StoreGet(key, data.Any)
    if ok {
        gobBuf.Reset()
        err := enc.Encode(elem)
        if err != nil {
            log.Fatal("encode error:", err)
            return "", false
        } else {
            str_rep := string(Base64Encode(gobBuf.Bytes()))
            return str_rep, true
        }
    } else {
        return "", false
    }
}

func (db *DB) StoreBase64Load(key string, str_rep string) (bool) {
    bytes_rep := Base64Decode([]byte(str_rep))
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

func init() {
    enc = gob.NewEncoder(&gobBuf)
    dec = gob.NewDecoder(&gobBuf)
    DefaultDBManager.Init()
}
