package user

import (
    "store"
    "store/reply"
    "fmt"
    "data"
    "crypto/rand"
    "code.google.com/p/go.crypto/scrypt"
    "io"
    "encoding/gob"
)

type User struct {
    Username string
    Hash string
    Token string
    Score int
}

var UserType uint8 = 20

func init() {

    gob.Register(&User{})

    store.RegisterPrefixedStoreType(UserType,"user")    

    store.DefaultDBManager.AddFuncWithSideEffects("useradd", func (db *store.DB, args []string) string {
        s := args[0]
        _, ok, _ := db.StoreGet(s,UserType)
        if ok {
            return reply.ErrorReply("Key already exists")
        } else {
            raw_hash, _ := scrypt.Key([]byte(args[1]),[]byte(args[0]),16384,8,1,32)
            b64_hash := store.Base64Encode(raw_hash)
            final_hash := fmt.Sprintf("$%d$%d$%d$%s",16384,8,1,string(b64_hash))
            user := &User{args[0],final_hash,"",0}
            db.StoreSet(args[0], &data.Entry{user, UserType})
            return reply.OKReply
        }
        return reply.OKReply
    })

    store.DefaultDBManager.AddFunc("userget", func (db *store.DB, args []string) string {
        s := args[0]
        e, ok, _ := db.StoreGet(s,UserType)
        if ok {
            if user, ok := e.Value.(*User); ok {
                score_str := fmt.Sprintf("%d",user.Score)
                var rep reply.MultiBulkWriter
                rep.WriteCount(2)
                rep.WriteString(user.Username)
                rep.WriteString(score_str)
                return rep.String()
            } else {
                return reply.ErrorReply("Type Mismatch")
            }
        } else {
            return reply.NilReply
        }
    })

    store.DefaultDBManager.AddFuncWithSideEffects("userauth", func (db *store.DB, args []string) string {
        username := args[0]
        e, ok, _ := db.StoreGet(username,UserType)
        if ok {
            if user, ok := e.Value.(*User); ok {
                password := args[1]
                raw_hash, _ := scrypt.Key([]byte(password),[]byte(username),16384,8,1,32)
                b64_hash := store.Base64Encode(raw_hash)
                final_hash := fmt.Sprintf("$%d$%d$%d$%s",16384,8,1,string(b64_hash))
                if final_hash == user.Hash {
                    c := 64
                    b := make([]byte, c)
                    n, err := io.ReadFull(rand.Reader, b)
                    if n != len(b) || err != nil {
                        return "-ERR Unable to generate a secure random token"
                    }
                    token := store.Base64Encode(b)
                    token_str := string(token)
                    db.StoreSet(token_str,&data.Entry{username,0})
                    if user.Token != "" {
                        db.StoreDel(token_str)
                    }
                    user.Token = token_str
                    return reply.BulkReply(token_str)
                } else {
                    return reply.ErrorReply("Password is incorrect")
                }
            } else {
                return reply.ErrorReply("How did you manage to mess things up this good?")
            }
        } else {
            return reply.ErrorReply("No user exists with that username")
        }
    })

}
