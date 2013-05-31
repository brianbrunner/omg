package strings

import (
    "data"
    "store"
    "store/reply"
)

var StringType uint8 = 0

func init() {
    store.RegisterStoreType(StringType)
    store.DefaultDBManager.AddFunc("get", func (db *store.DB, args []string) string {
        elem, ok, err := db.StoreGet(args[0], StringType)
        if err != nil {
            return reply.ErrorReply(err)
        }
        if ok {
            s := elem.GetString()
            return reply.BulkReply(s)
        }
        return reply.NilReply
    })
    store.DefaultDBManager.AddFuncWithSideEffects("set", func (db *store.DB, args []string) string {
        elem, ok, _ := db.StoreGet(args[0], StringType)
        if ok {
            elem.Value = args[1]
        } else {
            db.StoreSet(args[0], &data.Entry{args[1], StringType})
        }
        return reply.OKReply
    })
    store.DefaultDBManager.AddFuncWithSideEffects("setnx", func (db *store.DB, args []string) string {
        s := args[0]
        _, ok, _ := db.StoreGet(s, StringType)
        if ok {
            return reply.IntReply(0)
        }
        db.StoreSet(s, &data.Entry{args[1], StringType})
        return reply.IntReply(1)
    })
}
