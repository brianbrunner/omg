package server

import (
    "store"
    "store/reply"
    "runtime"
)

func init() {

    store.DefaultDBManager.AddFunc("select", func(db *store.DB, args []string) string {
        return reply.OKReply
    })

    store.DefaultDBManager.AddFunc("info", func(db *store.DB, args []string) string {
        return reply.BulkReply("# Server\r\nredis_version:2.6.7\r\n")
    })

    store.DefaultDBManager.AddFunc("rungc", func(db *store.DB, args []string) string {
        runtime.GC()
        return reply.OKReply
    })

    store.DefaultDBManager.AddFunc("save", func(db *store.DB, args []string) string {
        err := db.SaveToDiskSync()
        if err != nil {
            return reply.ErrorReply(err)
        } else {
            return reply.OKReply
        }
    })

    store.DefaultDBManager.AddFunc("load", func(db *store.DB, args []string) string {
        err := db.LoadFromDiskSync()
        if err != nil {
            return reply.ErrorReply(err)
        } else {
            return reply.OKReply
        }
    })

}
