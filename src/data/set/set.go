package set

import (
    "store"
    "data"
    "fmt"
    "encoding/gob"
)

var SetType int = 2

func init() {
    gob.Register(make(map[string]bool))
    data.RegisterStoreType(SetType)
    store.DefaultDBManager.AddFunc("sadd", func (db *store.DB, args []string) string {
        e, ok, err := db.StoreGet(args[0], SetType)
        if err != nil {
            return "-ERR Type Mismatch\r\n"
        }
        setData, ok := e.Value.(map[string]bool)
        if !ok {
            setData = make(map[string]bool)
            db.StoreSet(args[0],&data.Entry{setData,SetType,0})
        }
        l := len(args)
        sdata := args[1:l]
        count := 0
        for _, member := range sdata {
            _, ok = setData[member]
            if !ok {
                setData[member] = true 
                count += 1
            }
        }
        return fmt.Sprintf(":%d\r\n",count)
    })
    store.DefaultDBManager.AddFunc("srem", func (db *store.DB, args []string) string {
        e, ok, err := db.StoreGet(args[0], SetType)
        if err != nil {
            return "-ERR Type Mismatch\r\n"
        }
        setData, ok := e.Value.(map[string]bool)
        if !ok {
            setData = make(map[string]bool)
        }
        l := len(args)
        sdata := args[1:l]
        count := 0
        for _, member := range sdata {
            _, ok = setData[member]
            if ok {
                delete(setData, member)
                count += 1
            }
        }
        return fmt.Sprintf(":%d\r\n",count)
    })
    store.DefaultDBManager.AddFunc("sismember", func (db *store.DB, args []string) string {
        e, ok, err := db.StoreGet(args[0], SetType)
        if err != nil || !ok {
            return "-ERR Type Mismatch\r\n"
        }
        setData, ok := e.Value.(map[string]bool)
        if !ok {
            return ":0\r\n"
        }
        _, ok = setData[args[1]]
        if ok {
            return ":1\r\n"
        }
        return ":0\r\n"
    })
    store.DefaultDBManager.AddFunc("scard", func (db *store.DB, args []string) string {
        e, ok, err := db.StoreGet(args[0], SetType)
        if err != nil || !ok {
            return "-ERR Type Mismatch\r\n"
        }
        setData, ok := e.Value.(map[string]bool)
        if ok {
            return fmt.Sprintf(":%d\r\n",len(setData))
        }
        return fmt.Sprintf(":%d\r\n",0)
    })
    store.DefaultDBManager.AddFunc("spop", func (db *store.DB, args []string) string {
        e, ok, err := db.StoreGet(args[0], SetType)
        if err != nil || !ok {
            return "-ERR Type Mismatch\r\n"
        }
        setData, ok := e.Value.(map[string]bool)
        if !ok {
            return "$-1\r\n"
        }
        for k, _ := range setData {
            delete(setData, k)
            return fmt.Sprintf("$%d\r\n%s\r\n",len(k), k)
        }
        return "$-1\r\n"
    })
}
