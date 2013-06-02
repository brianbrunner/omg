package feed 

import (
    "data"
    "fmt"
    "strconv"
    "store"
    "store/reply"
    "encoding/gob"
)

/*
 * FEED Commands
 */ 

type Post struct {
    Author string
    Body string
    Timestamp uint64
    Expires uint64
    Likers map[string]string
}

type Feed struct {
    Name string
    Capacity int
    LastTimestamp uint64
    Posts []*Post
}

var FeedType uint8 = 21

func init() {

    gob.Register(&Post{})
    gob.Register(&Feed{})

    store.RegisterStoreType(FeedType)

    store.DefaultDBManager.AddFuncWithSideEffects("feedcap", func (db *store.DB, args []string) string {
        s := args[0]
        cap, err := strconv.Atoi(args[1])
        if err != nil {
            return reply.ErrorReply(err)
        }
        e, ok, _ := db.StoreGet(s,FeedType)
        if ok {
            if f, ok := e.Value.(*Feed); ok {
                f.Capacity = cap
                return reply.OKReply
            } else {
                return reply.ErrorReply("Type Mismatch")
            }
        } else {
            f := &Feed{s,cap,0,make([]*Post,0)}
            db.StoreSet(s,&data.Entry{f, FeedType})
        }
        return reply.OKReply
    })

    store.DefaultDBManager.AddFunc("feedget", func (db *store.DB, args []string) string {
        if e, ok, _ := db.StoreGet(args[0], FeedType); ok {
            if f, ok := e.Value.(*Feed); ok {
                cap_str := fmt.Sprintf("%d",f.Capacity)
                var rep reply.MultiBulkWriter
                rep.WriteCount(2)
                rep.WriteString(f.Name)
                rep.WriteString(cap_str)
                return rep.String()
            } else {
                return "-ERR Type mismatch\r\n"
            }
        } else {
            return reply.NilReply
        }
    })

    store.DefaultDBManager.AddFuncWithSideEffects("feedpost", func (db *store.DB, args []string) string {
        e, ok, _ := db.StoreGet(args[0], FeedType);
        if !ok {
            s := args[0]
            f := &Feed{s,-1,0,make([]*Post,0)}
            e = &data.Entry{f, FeedType}
            db.StoreSet(s, e)
        }
        if f, ok := e.Value.(*Feed); ok {
            timestamp := store.Milliseconds()
            if timestamp == f.LastTimestamp {
                timestamp += 1
            }
            f.LastTimestamp = timestamp
            p := &Post{args[1],args[2],timestamp,0,make(map[string]string)}
            f.Posts = append(f.Posts, p)
            return reply.OKReply
        } else {
            return reply.ErrorReply("Type mismatch")
        }
    })

    store.DefaultDBManager.AddFunc("feedposts", func (db *store.DB, args []string) string {
        if e, ok, _ := db.StoreGet(args[0],FeedType); ok {
            if f, ok := e.Value.(*Feed); ok {
                post_count := len(f.Posts)
                offset := 0
                count := post_count
                if len(args) == 3 {

                  var err error
                  offset, err = strconv.Atoi(args[1])
                  if err != nil {
                    panic(err)
                  }

                  count, err = strconv.Atoi(args[2])
                  if err != nil {
                    panic(err)
                  }

                }

                if offset >= post_count {
                  return reply.NilReply
                }

                if offset + count > post_count {
                  count = post_count-offset
                }

                var rep reply.MultiBulkWriter
                rep.WriteCount(3*post_count)
                for i := post_count-offset; i >= post_count-offset-count; i-- { 
                    p := f.Posts[i]
                    timestamp_str := fmt.Sprintf("%d",p.Timestamp)
                    rep.WriteString(p.Author)
                    rep.WriteString(p.Body)
                    rep.WriteString(timestamp_str)
                }
                return rep.String()
            } else {
                return reply.ErrorReply("Type mismatch")
            }
        } else {
            return reply.NilReply
        }
    })

    store.DefaultDBManager.AddFunc("feedlen", func (db *store.DB, args []string) string {
        if e, ok, _ := db.StoreGet(args[0],FeedType); ok {
            if f, ok := e.Value.(*Feed); ok {
                return reply.IntReply(len(f.Posts))
            }
        }
        return reply.NilReply
    })

    //store.DefaultDBManager.AddFuncWithSideEffects("feedlike", func (db *store.DB, args []string) string {
    //})
}
