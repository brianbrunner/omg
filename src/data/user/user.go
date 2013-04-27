package user

import (
    "store"
    "fmt"
    "data"
    "code.google.com/p/go.crypto/scrypt"
)

type User struct {
    username string
    hash string
    token string
    score int
}

func init() {
    store.DefaultDBManager.AddFunc("useradd", func (db *store.DB, args []string) string {
        s := args[0]
        _, ok, _ := db.StoreGet(s,data.User)
        if ok {
            return "-ERR Key already exists"
        } else {
            raw_hash, _ := scrypt.Key([]byte(args[1]),[]byte(args[0]),16384,8,1,32)
            b64_hash := store.Base64Encode(raw_hash)
            final_hash := fmt.Sprintf("$%d$%d$%d$%s",16384,8,1,string(b64_hash))
            user := &User{args[0],final_hash,"",0}
            db.StoreSet(args[0], &data.Entry{user, data.User, 0})
            return "+OK\r\n"
        }
        return "+OK\r\b"
    })
    store.DefaultDBManager.AddFunc("userget", func (db *store.DB, args []string) string {
        s := args[0]
        e, ok, _ := db.StoreGet(s,data.User)
        if ok {
            if user, ok := e.Value.(*User); ok {
                score_str := fmt.Sprintf("%d",user.score)
                return fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",len(user.username),user.username,len(score_str),score_str)
            } else {
                return "-ERR Type Mismatch\r\n"
            }
        } else {
            return "-ERR No user with that name"
        }
        return "OK\r\n"
    })
}
