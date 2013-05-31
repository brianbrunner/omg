package server

import (
    "store"
    "store/reply"
    "runtime"
    "data"
)

func init() {

    store.DefaultDBManager.AddFunc("select", func(db *store.DB, args []string) string {
        return reply.OKReply
    })

    store.DefaultDBManager.AddFunc("info", func(db *store.DB, args []string) string {
        return reply.BulkReply("# Server\r\nredis_version:2.6.7\r\n")
    })

    store.DefaultDBManager.AddFunc("meminuse", func(db *store.DB, args []string) string {
        return reply.IntReply(int(store.MemInUse()))
    })

    store.DefaultDBManager.AddFunc("objectsinuse", func(db *store.DB, args []string) string {
        var memProfiles []runtime.MemProfileRecord
        n, _ := runtime.MemProfile(memProfiles, false)
        memProfiles = make([]runtime.MemProfileRecord,n)
        n, _ = runtime.MemProfile(memProfiles, false)

        objects := int64(0)
        for _, prof := range memProfiles {
          objects += prof.InUseObjects()
        }

        return reply.IntReply(int(objects))
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

    store.DefaultDBManager.AddFunc("bgsave", func(db *store.DB, args []string) string {
        db.StartSaveMode()
        db.SaveToDiskAsync()
        return reply.OKReply
    })

    store.DefaultDBManager.AddFunc("__end_save_mode__", func(db *store.DB, args []string) string {
        db.EndSaveMode()
        return reply.OKReply
    })


    store.DefaultDBManager.AddFunc("load", func(db *store.DB, args []string) string {
        err := db.LoadFromDiskSync()
        if err != nil {
            return reply.ErrorReply(err)
        } else {
            return reply.OKReply
        }
    })

    store.DefaultDBManager.AddFunc("dbsize", func(db *store.DB, args []string) string {
        return reply.IntReply(db.StoreSize())
    })

    store.DefaultDBManager.AddFunc("lastdump", func(db *store.DB, args []string) string {
        elem, ok, _ := db.StoreGet(args[0],data.Any)
        if ok {
          return reply.IntReply(int(elem.LastDump()))
        } else {
          return reply.NilReply
        }
    })

    store.DefaultDBManager.AddFunc("lastdumpdb", func(db *store.DB, args []string) string {
        return reply.IntReply(int(db.LastDump()))
    })

}
