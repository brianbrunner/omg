package keys

import (
    "fmt"
    "store"
    "strconv"
    "data"
)

func init() {
    store.DefaultDBManager.AddFunc("del", func (db *store.DB, args []string) string {
        del_count := 0
        for _, v := range args {
            ok := db.StoreDel(v)
            if ok {
                del_count += 1
            }
        }
        return fmt.Sprintf(":%d\r\n",del_count)
    })
    store.DefaultDBManager.AddFunc("expire", func(db *store.DB, args []string) string {
        elem, ok, _ := db.StoreGet(args[0], data.Any)
        millis := store.Milliseconds()
        if ok && (elem.Expires == 0 || elem.Expires > millis) {
            expires, err := strconv.Atoi(args[1])
            if err != nil {
                return ":0\r\n"
            } else {
                elem.Expires = millis + uint64(expires*1000)
                return ":1\r\n"
            }
        }
        return ":0\r\n"
    })
    store.DefaultDBManager.AddFunc("pexpire", func(db *store.DB, args []string) string {
        elem, ok, _ := db.StoreGet(args[0], data.Any)
        millis := store.Milliseconds()
        if ok && (elem.Expires == 0 || elem.Expires > millis) {
            expires, err := strconv.Atoi(args[1])
            if err != nil {
                return ":0\r\n"
            } else {
                elem.Expires = millis + uint64(expires)
                return ":1\r\n"
            }
        }
        return ":0\r\n"
    })
    store.DefaultDBManager.AddFunc("exists", func(db *store.DB, args []string) string {
        _, ok, _ := db.StoreGet(args[0], data.Any)
        if ok {
            return "$1\r\n1\r\n"
        }
        return "$1\r\n0\r\n"
    })
}
