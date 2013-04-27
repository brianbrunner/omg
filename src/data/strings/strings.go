package strings

import (
    "data"
    "fmt"
    "store"
)

func init() {
    store.DefaultDBManager.AddFunc("get", func (db *store.DB, args []string) string {
        elem, ok, err := db.StoreGet(args[0], data.Str)
        if err != nil {
            return fmt.Sprintf("-ERR %s\r\n",err)
        }
        if ok {
            s := elem.GetString()
            return fmt.Sprintf("$%d\r\n%s\r\n",len(s),s)
        }
        return "$-1\r\n"
    })
    store.DefaultDBManager.AddFunc("set", func (db *store.DB, args []string) string {
        db.StoreSet(args[0], &data.Entry{args[1], data.Str, 0})
        return "+OK\r\n"
    })
    store.DefaultDBManager.AddFunc("setnx", func (db *store.DB, args []string) string {
        s := args[0]
        _, ok, _ := db.StoreGet(s, data.Str)
        if ok {
            return "$1\r\n0\r\n"
        }
        db.StoreSet(s, &data.Entry{args[1], data.Str, 0})
        return "$1\r\n1\r\n"
    })
}
