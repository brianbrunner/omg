package zset

import (
    "tree"
    "store"
)

func init() {
    store.DefaultDBManager.AddFunc("zadd", func (db *store.DB, args []string) string {
        l := len(args)
        if l-1 % 2 == 0 {
            return "-ERR Incorrect number of arguments"
        }
        e, ok, err := db.StoreGet(args[0], data.ZSet)
        if err != nil {
            return err
        }
        
        zdata := args[1:l]
        for i := 0; i < l-1; i += 2 {
            
        }
    })
}
