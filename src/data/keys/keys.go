package keys

import (
  "store"
  "store/reply"
  "strconv"
  "data"
  "fmt"
)

func init() {

  store.DefaultDBManager.AddFuncWithSideEffects("del", func (db *store.DB, args []string) string {
    del_count := 0
    for _, v := range args {
      ok := db.StoreDel(v)
      if ok {
        del_count += 1
      }
    }
    return reply.IntReply(del_count)
  })

  store.DefaultDBManager.AddFuncWithSideEffects("expire", func(db *store.DB, args []string) string {
    _, ok, _ := db.StoreGet(args[0], data.Any)
    if ok {
      millis := store.Milliseconds()
      expires, err := strconv.Atoi(args[1])
      if err != nil {
        return reply.IntReply(0)
      } else {
        db.ExpiryRecord[args[0]] = millis + uint64(expires*1000)
        return reply.IntReply(1)
      }
    }
    return reply.IntReply(0)
  })

  store.DefaultDBManager.AddFuncWithSideEffects("pexpire", func(db *store.DB, args []string) string {
    _, ok, _ := db.StoreGet(args[0], data.Any)
    if ok {
      millis := store.Milliseconds()
      expires, err := strconv.Atoi(args[1])
      if err != nil {
        return reply.IntReply(0)
      } else {
        db.ExpiryRecord[args[0]] = millis + uint64(expires)
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
    fmt.Println(bytes_rep)
    str_rep := string(bytes_rep)
    if ok {
      return reply.BulkReply(str_rep)
    } else {
      return reply.NilReply
    }
  })

  store.DefaultDBManager.AddFuncWithSideEffects("restore", func(db *store.DB, args []string) string {
    unquoted_data, _ := strconv.Unquote(args[1])
    fmt.Println(unquoted_data)
    bytes_rep := []byte(unquoted_data)
    ok := db.StoreLoad(args[0],bytes_rep)
    if ok {
      return reply.OKReply
    } else {
      return reply.NilReply
    }
  })

  store.DefaultDBManager.AddFunc("type", func(db *store.DB, args []string) string {
    elem, ok, _ := db.StoreGet(args[0], data.Any)
    if ok {
      return reply.IntReply(int(elem.EntryType))
    } else {
      return reply.NilReply
    }
  })

  store.DefaultDBManager.AddFunc("keys", func(db *store.DB, args []string) string {
    keys := db.StoreKeysMatch(args[0])
    var writer reply.MultiBulkWriter
    writer.WriteCount(len(keys))
    for _, key := range keys {
      writer.WriteString(key)
    }
    return writer.String()
  })

  store.DefaultDBManager.AddFunc("randomkey", func(db *store.DB, args []string) string {
    for key, _ := range db.Store {
      return reply.BulkReply(key)
    }
    return reply.NilReply
  })

}
