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

type Post struct {
    Author string
    Body string
    Timestamp uint64
    Likers map[string]string
}

type Feed struct {
    Name string
    Capacity int
    LastTimestamp uint64
    Posts []*Post
}

var FeedType int = 21

func init() {
    data.RegisterStoreType(FeedType)
    gob.Register(&Post{})
    gob.Register(&Feed{})
    store.DefaultDBManager.AddFunc("feedcap", func (db *store.DB, args []string) string {
        s := args[0]
        cap, err := strconv.Atoi(args[1])
        if err != nil {
            return fmt.Sprintf("-ERR %s\r\n",err)
        }
        e, ok, _ := db.StoreGet(s,FeedType)
        if ok {
            if f, ok := e.Value.(*Feed); ok {
                f.Capacity = cap
                return "+OK\r\n"
            } else {
                return "-ERR Type Mismatch\r\n"
            }
        } else {
            f := &Feed{s,cap,0,make([]*Post,0)}
            db.StoreSet(s,&data.Entry{f, FeedType, 0})
        }
        return "+OK\r\n"
    })
    store.DefaultDBManager.AddFunc("feedget", func (db *store.DB, args []string) string {
        if e, ok, _ := db.StoreGet(args[0], FeedType); ok {
            if f, ok := e.Value.(*Feed); ok {
                cap_str := fmt.Sprintf("%d",f.Capacity)
                return fmt.Sprintf("*2\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n",len(f.Name),f.Name,len(cap_str),cap_str)
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
            f := &Feed{s,-1,0,make([]*Post,0)}
            e = &data.Entry{f, FeedType, 0}
            db.StoreSet(s, e)
        }
        if f, ok := e.Value.(*Feed); ok {
            timestamp := store.Milliseconds()
            if timestamp == f.LastTimestamp {
                timestamp += 1
            }
            f.LastTimestamp = timestamp
            p := &Post{args[1],args[2],timestamp,make(map[string]string)}
            f.Posts = append(f.Posts, p)
            return "+OK\r\n"
        } else {
            return "-ERR Type mismatch\r\n"
        }
        return "$-1\r\n"
    })
    store.DefaultDBManager.AddFunc("feedposts", func (db *store.DB, args []string) string {
        if e, ok, _ := db.StoreGet(args[0],FeedType); ok {
            if f, ok := e.Value.(*Feed); ok {
                var buffer bytes.Buffer
                post_count := len(f.Posts)
                buffer.WriteString(fmt.Sprintf("*%d\r\n",3*post_count))
                for i := post_count-1; i >= 0; i-- { 
                    p := f.Posts[i]
                    timestamp_str := fmt.Sprintf("%d",p.Timestamp)
                    buffer.WriteString(fmt.Sprintf("$%d\r\n%s\r\n$%d\r\n%s\r\n$%d\r\n%s\r\n", len(p.Author), p.Author, len(p.Body), p.Body, len(timestamp_str), timestamp_str))
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
            if f, ok := e.Value.(*Feed); ok {
                return fmt.Sprintf(":%d\r\n",len(f.Posts))
            }
        } else {
            return "$-1\r\n"
        }
        return "$-1\r\n"
    })
    //store.DefaultDBManager.AddFunc("feedlike", func (db *store.DB, args []string) string {
    //})
}
