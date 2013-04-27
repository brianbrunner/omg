package store 

import (
    "errors"
    "data"
    "fmt"
    "time"
    "encoding/base64"
    "persist"
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
        function, ok := dbm.funcs[command.Args[0]]
        if ok {
            command.ReplyChan <- function(dbm.db,args)
            dbm.PersistChan <- command.CommandRaw
        } else {
            command.ReplyChan <- "-ERR Unknown command\r\n"
        }
    }
}


func (dbm *DBManager) Init() {
    dbm.ComChan = make(chan Command,10000)
    dbm.db = new(DB)
    dbm.db.store = make(map[string]*data.Entry)
    dbm.PersistChan = persist.StartPersist()
    dbm.funcs = make(map[string]func(*DB,[]string)(string))
    go dbm.run()
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

func (db *DB) StoreGet(key string, entryType data.EntryType) (*data.Entry, bool, error) {
    elem, ok := db.store[key]
    if ok {
        if elem.EType == entryType || entryType == data.Any {
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

func init() {
    DefaultDBManager.Init()
}
