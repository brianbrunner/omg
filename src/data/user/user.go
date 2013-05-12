package user

import (
    "store"
    "fmt"
    "data"
    "crypto/rand"
    "code.google.com/p/go.crypto/scrypt"
    "io"
    "encoding/gob"
)

type User struct {
    username string
    hash string
    token string
    score int
}

var UserType int = 20

func init() {
    gob.Register(*User{})
    data.RegisterStoreType(UserType)    
    store.DefaultDBManager.AddFunc("useradd", func (db *store.DB, args []string) string {
        s := args[0]
        _, ok, _ := db.StoreGet(s,UserType)
        if ok {
            return "-ERR Key already exists\r\n"
        } else {
            raw_hash, _ := scrypt.Key([]byte(args[1]),[]byte(args[0]),16384,8,1,32)
            b64_hash := store.Base64Encode(raw_hash)
            final_hash := fmt.Sprintf("$%d$%d$%d$%s",16384,8,1,string(b64_hash))
            user := &User{args[0],final_hash,"",0}
            db.StoreSet(args[0], &data.Entry{user, UserType, 0})
            return "+OK\r\n"
        }
        return "+OK\r\b"
    })
    store.DefaultDBManager.AddFunc("userget", func (db *store.DB, args []string) string {
        s := args[0]
        e, ok, _ := db.StoreGet(s,UserType)
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
    store.DefaultDBManager.AddFunc("userauth", func (db *store.DB, args []string) string {
        username := args[0]
        e, ok, _ := db.StoreGet(username,UserType)
        if ok {
            if user, ok := e.Value.(*User); ok {
                password := args[1]
                raw_hash, _ := scrypt.Key([]byte(password),[]byte(username),16384,8,1,32)
                b64_hash := store.Base64Encode(raw_hash)
                final_hash := fmt.Sprintf("$%d$%d$%d$%s",16384,8,1,string(b64_hash))
                if final_hash == user.hash {
                    c := 64
                    b := make([]byte, c)
                    n, err := io.ReadFull(rand.Reader, b)
                    if n != len(b) || err != nil {
                        return "-ERR Unable to generate a secure random token"
                    }
                    token := store.Base64Encode(b)
                    token_str := string(token)
                    db.StoreSet(token_str,&data.Entry{username,0,0})
                    if user.token != "" {
                        db.StoreDel(token_str)
                    }
                    user.token = token_str
                    return fmt.Sprintf("$%d\r\n%s\r\n",len(token_str),token_str)
                } else {
                    return "-ERR Password is incorrect\r\n"
                }
            } else {
                return "-ERR How did you manage to mess things up this good?\r\n"
            }
        } else {
            return "-ERR No user exists with that username\r\n"
        }
        return "-ERR No user exists with that username\r\n"
    })
}
