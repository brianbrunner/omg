package keys

import (
    "store"
    "store/reply"
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
        return reply.IntReply(del_count)
    })

    store.DefaultDBManager.AddFunc("expire", func(db *store.DB, args []string) string {
        elem, ok, _ := db.StoreGet(args[0], data.Any)
        millis := store.Milliseconds()
        if ok && (elem.Expires == 0 || elem.Expires > millis) {
            expires, err := strconv.Atoi(args[1])
            if err != nil {
                return reply.IntReply(0)
            } else {
                elem.Expires = millis + uint64(expires*1000)
                return reply.IntReply(1)
            }
        }
        return reply.IntReply(0)
    })

    store.DefaultDBManager.AddFunc("pexpire", func(db *store.DB, args []string) string {
        elem, ok, _ := db.StoreGet(args[0], data.Any)
        millis := store.Milliseconds()
        if ok && (elem.Expires == 0 || elem.Expires > millis) {
            expires, err := strconv.Atoi(args[1])
            if err != nil {
                return reply.IntReply(0)
            } else {
                elem.Expires = millis + uint64(expires)
                return reply.IntReply(1)
            }
        }
        return reply.IntReply(0)
    })

    store.DefaultDBManager.AddFunc("exists", func(db *store.DB, args []string) string {
        _, ok, _ := db.StoreGet(args[0], data.Any)
        if ok {
            return reply.IntReply(1)
        }
        return reply.IntReply(0)
    })

    store.DefaultDBManager.AddFunc("dump", func(db *store.DB, args []string) string {
        bytes_rep, ok := db.StoreDump(args[0])
        str_rep := string(bytes_rep)
        if ok {
            return reply.BulkReply(str_rep)
        } else {
            return reply.NilReply
        }
    })

    store.DefaultDBManager.AddFunc("restore", func(db *store.DB, args []string) string {
        ok := db.StoreLoad(args[0],[]byte(args[1]))
        if ok {
            return reply.OKReply
        } else {
            return reply.NilReply
        }
    })

    store.DefaultDBManager.AddFunc("type", func(db *store.DB, args []string) string {
        elem, ok, _ := db.StoreGet(args[0], data.Any)
        if ok {
            return reply.IntReply(elem.EntryType)
        } else {
            return reply.NilReply
        }
    })
}
