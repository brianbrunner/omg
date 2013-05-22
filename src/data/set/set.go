package set

import (
    "store"
    "store/reply"
    "data"
    "encoding/gob"
)

var SetType int = 3

func init() {

    gob.Register(make(map[string]bool))

    store.RegisterStoreType(SetType)

    store.DefaultDBManager.AddFuncWithSideEffects("sadd", func (db *store.DB, args []string) string {
        e, ok, err := db.StoreGet(args[0], SetType)
        if err != nil {
            return reply.ErrorReply("Type Mismatch")
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
        return reply.IntReply(count)
    })

    store.DefaultDBManager.AddFuncWithSideEffects("srem", func (db *store.DB, args []string) string {

        e, ok, err := db.StoreGet(args[0], SetType)
        if err != nil {
            return reply.ErrorReply("Type Mismatch")
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
            if ok {
                delete(setData, member)
                count += 1
            }
        }

        return reply.IntReply(count)
    })


    store.DefaultDBManager.AddFuncWithSideEffects("srem", func (db *store.DB, args []string) string {
        e, ok, err := db.StoreGet(args[0], SetType)
        if err != nil {
            return reply.ErrorReply("Type Mismatch")
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
        return reply.IntReply(count)
    })

    store.DefaultDBManager.AddFunc("sismember", func (db *store.DB, args []string) string {
        e, ok, err := db.StoreGet(args[0], SetType)
        if err != nil || !ok {
            return "-ERR Type Mismatch\r\n"
        }
        setData, ok := e.Value.(map[string]bool)
        if !ok {
            return reply.IntReply(0)
        }
        _, ok = setData[args[1]]
        if ok {
            return reply.IntReply(1)
        }
        return reply.IntReply(0)
    })

    store.DefaultDBManager.AddFunc("scard", func (db *store.DB, args []string) string {
        e, ok, err := db.StoreGet(args[0], SetType)
        if err != nil || !ok {
            return reply.ErrorReply("Type Mismatch")
        }
        setData, ok := e.Value.(map[string]bool)
        if ok {
            return reply.IntReply(len(setData))
        }
        return reply.IntReply(0)
    })

    store.DefaultDBManager.AddFuncWithSideEffects("spop", func (db *store.DB, args []string) string {
        e, ok, err := db.StoreGet(args[0], SetType)
        if err != nil || !ok {
            return "-ERR Type Mismatch\r\n"
        }
        setData, ok := e.Value.(map[string]bool)
        if !ok {
            return reply.NilReply
        }
        for k, _ := range setData {
            delete(setData, k)
            return reply.BulkReply(k)
        }
        return reply.NilReply
    })

}
