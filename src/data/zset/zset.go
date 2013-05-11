package zset

import (
    "zset"
    "store"
    "data"
    "strconv"
    "fmt"
)

var ZSetType int = 3

func init() {
    data.RegisterStoreType(ZSetType)
    store.DefaultDBManager.AddFunc("zadd", func (db *store.DB, args []string) string {
        i, err := strconv.Atoi(args[1])
        if err != nil {
            return "-ERR Invalid score\r\n"
        }
        elem, ok, err := db.StoreGet(args[0],ZSetType)
        if !ok {
            zs := zset.New()
            elem = &data.Entry{zs,ZSetType,0}
            db.StoreSet(args[0],elem)
        }
        zs, _ := elem.Value.(zset.ZSet)
        zs.Insert(i,args[2])
        return ":1\r\n"
    })
    store.DefaultDBManager.AddFunc("zrem", func (db *store.DB, args []string) string {
        elem, ok, _ := db.StoreGet(args[0],ZSetType)
        if !ok {
            return "$-1\r\n"
        }
        zs, _ := elem.Value.(zset.ZSet)
        zs.Delete(args[1])
        return ":1\r\n"
    })
    store.DefaultDBManager.AddFunc("zcard", func (db *store.DB, args []string) string {
        elem, ok, _ := db.StoreGet(args[0],ZSetType)
        if !ok {
            return "$-1\r\n"
        }
        zs, ok := elem.Value.(zset.ZSet)
        fmt.Println(elem)
        fmt.Println(zs)
        return fmt.Sprintf(":%d\r\n",zs.Card())
    })
    store.DefaultDBManager.AddFunc("zscore", func (db *store.DB, args []string) string {
        elem, ok, _ := db.StoreGet(args[0],ZSetType)
        if !ok {
            return "$-1\r\n"
        }
        zs, _ := elem.Value.(zset.ZSet)
        score, ok := zs.Score(args[1])
        if ok {
            return fmt.Sprintf(":%d\r\n",score)
        } else {
            return "$-1\r\n"
        }
        return "$-1\r\n"
    })
    store.DefaultDBManager.AddFunc("zrank", func (db *store.DB, args []string) string {
        elem, ok, _ := db.StoreGet(args[0],ZSetType)
        if !ok {
            return "$-1\r\n"
        }
        zs, _ := elem.Value.(zset.ZSet)
        rank, ok := zs.Rank(args[1])
        if ok {
            return fmt.Sprintf(":%d\r\n",rank)
        } else {
            return "$-1\r\n"
        }
        return "$-1\r\n"
    })
}
