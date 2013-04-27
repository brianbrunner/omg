package server

import (
    "store"
)

func init() {
    store.DefaultDBManager.AddFunc("info", func(db *store.DB, args []string) string {
        return "$31\r\n# Server\r\nredis_version:2.6.7\r\n\r\n"
    })
}
