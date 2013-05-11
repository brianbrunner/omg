package feed 

import (
    "bytes"
    "data"
    "fmt"
    "strconv"
    "store"
    "encoding/gob"
)

/*
 * FEED Commands
 */ 

type post struct {
    author string
    body string
    timestamp uint64
    likers map[string]string
}

type feed struct {
    name string
    capacity int
    lastTimestamp uint64
    posts []*post
}

var FeedType int = 21

func init() {
    data.RegisterStoreType(FeedType)
    gob.Register(post{})
    gob.Register(feed{})
    store.DefaultDBManager.AddFunc("feedcap", func (db *store.DB, args []string) string {
        s := args[0]
        cap, err := strconv.Atoi(args[1])
        if err != nil {
            return fmt.Sprintf("-ERR %s\r\n",err)
        }
        e, ok, _ := db.StoreGet(s,FeedType)
        if ok {
            if f, ok := e.Value.(*feed); ok {
                f.capacity = cap
                return "+OK\r\n"
            } else {
                return "-ERR Type Mismatch\r\n"
            }
        } else {
            f := &feed{s,cap,0,make([]*post,0)}
            db.StoreSet(s,&data.Entry{f, FeedType, 0})
        }
        return "+OK\r\n"
    })
    store.DefaultDBManager.AddFunc("feedget", func (db *store.DB, args []string) string {
        if e, ok, _ := db.StoreGet(args[0], FeedType); ok {
            if f, ok := e.Value.(*feed); ok {
                cap_str := fmt.Sprintf("%d",f.capacity)
                return fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",len(f.name),f.name,len(cap_str),cap_str)
            } else {
                return "-ERR Type mismatch\r\n"
            }
        } else {
            return "$-1\r\n"
        }
        return "$-1\r\n"
    })
    store.DefaultDBManager.AddFunc("feedpost", func (db *store.DB, args []string) string {
        e, ok, _ := db.StoreGet(args[0], FeedType);
        if !ok {
            s := args[0]
            f := &feed{s,-1,0,make([]*post,0)}
            e = &data.Entry{f, FeedType, 0}
            db.StoreSet(s, e)
        }
        if f, ok := e.Value.(*feed); ok {
            timestamp := store.Milliseconds()
            if timestamp == f.lastTimestamp {
                timestamp += 1
            }
            f.lastTimestamp = timestamp
            p := &post{args[1],args[2],timestamp,make(map[string]string)}
            f.posts = append(f.posts, p)
            return "+OK\r\n"
        } else {
            return "-ERR Type mismatch\r\n"
        }
        return "$-1\r\n"
    })
    store.DefaultDBManager.AddFunc("feedposts", func (db *store.DB, args []string) string {
        if e, ok, _ := db.StoreGet(args[0],FeedType); ok {
            if f, ok := e.Value.(*feed); ok {
                var buffer bytes.Buffer
                post_count := len(f.posts)
                buffer.WriteString(fmt.Sprintf("*%d\r\n",3*post_count))
                for i := post_count-1; i >= 0; i-- { 
                    p := f.posts[i]
                    timestamp_str := fmt.Sprintf("%d",p.timestamp)
                    buffer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(p.author), p.author, len(p.body), p.body, len(timestamp_str), timestamp_str))
                }
                return buffer.String()
            } else {
                return "-ERR Type mismatch\r\n"
            }
        } else {
            return "$-1\r\n"
        }
        return "$-1\r\n"
    })
    store.DefaultDBManager.AddFunc("feedlen", func (db *store.DB, args []string) string {
        if e, ok, _ := db.StoreGet(args[0],FeedType); ok {
            if f, ok := e.Value.(*feed); ok {
                return fmt.Sprintf(":%d\r\n",len(f.posts))
            }
        } else {
            return "$-1\r\n"
        }
        return "$-1\r\n"
    })
    //store.DefaultDBManager.AddFunc("feedlike", func (db *store.DB, args []string) string {
    //})
}
