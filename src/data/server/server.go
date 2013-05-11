package server

import (
    "store"
    "runtime"
)

func init() {
    store.DefaultDBManager.AddFunc("select", func(db *store.DB, args []string) string {
        return "+OK\r\n"
    })
    store.DefaultDBManager.AddFunc("info", func(db *store.DB, args []string) string {
        return "$31\r\n# Server\r\nredis_version:2.6.7\r\n\r\n"
    })
    store.DefaultDBManager.AddFunc("rungc", func(db *store.DB, args []string) string {
        runtime.GC()
        return "+OK\r\n"
    })
}
