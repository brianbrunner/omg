package strings

import (
	"data"
	"store"
	"store/reply"
  "strconv"
  "fmt"
)

var StringType uint8 = 0

func init() {

	store.RegisterStoreType(StringType)

	store.DefaultDBManager.AddFuncWithSideEffects("append", func(db *store.DB, args []string) string {
		elem, ok, _ := db.StoreGet(args[0], StringType)
		if ok {
      str, _ := elem.Value.(string)
			str += args[1]
      elem.Value = str
      return reply.IntReply(len(str))
		} else {
			db.StoreSet(args[0], &data.Entry{args[1], StringType})
		  return reply.IntReply(len(args[1]))
		}
	})

  bitCount := []uint8{0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4, 
                      1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5, 
                      1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5, 
                      2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 
                      1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5, 
                      2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 
                      2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 
                      3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7, 
                      1, 2, 2, 3, 2, 3, 3, 4, 2, 3, 3, 4, 3, 4, 4, 5, 
                      2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 
                      2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 
                      3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7, 
                      2, 3, 3, 4, 3, 4, 4, 5, 3, 4, 4, 5, 4, 5, 5, 6, 
                      3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7, 
                      3, 4, 4, 5, 4, 5, 5, 6, 4, 5, 5, 6, 5, 6, 6, 7, 
                      4, 5, 5, 6, 5, 6, 6, 7, 5, 6, 6, 7, 6, 7, 7, 8}

	store.DefaultDBManager.AddFuncWithSideEffects("bitcount", func(db *store.DB, args []string) string {
		elem, ok, _ := db.StoreGet(args[0], StringType)
		if ok {
      byte_arr, _ := elem.Value.([]byte)
      c := 0
      for _, c := range byte_arr {
        c += bitCount[c]
		  }
      return reply.IntReply(c)
		} else {
		  return reply.IntReply(0)
		}
	})

	store.DefaultDBManager.AddFuncWithSideEffects("decr", func(db *store.DB, args []string) string {
		elem, ok, _ := db.StoreGet(args[0], StringType)
    var num int
		if ok {
      str, _ := elem.Value.(string)
      num, err := strconv.Atoi(str)
      if err != nil {
        panic(err)
      }
      num -= 1
			elem.Value = fmt.Sprintf("%d",num)
		} else {
      num = -1
			db.StoreSet(args[0], &data.Entry{"-1",StringType})
		}
		return reply.IntReply(num)
	})

	store.DefaultDBManager.AddFuncWithSideEffects("decrby", func(db *store.DB, args []string) string {
		elem, ok, _ := db.StoreGet(args[0], StringType)
    var num int

    decrby, err := strconv.Atoi(args[1])
    if err != nil {
      panic(err)
    }

		if ok {
      str, _ := elem.Value.(string)

      num, err := strconv.Atoi(str)
      if err != nil {
        panic(err)
      }

      num -= decrby
			elem.Value = fmt.Sprintf("%d",num)
		} else {
      str := fmt.Sprintf("-%d",decrby)
			db.StoreSet(args[0], &data.Entry{str,StringType})
		}
    println(num)
		return reply.IntReply(num)
	})

	store.DefaultDBManager.AddFuncWithSideEffects("incr", func(db *store.DB, args []string) string {
		elem, ok, _ := db.StoreGet(args[0], StringType)
    var num int
		if ok {
      str, _ := elem.Value.(string)
      num, err := strconv.Atoi(str)
      if err != nil {
        panic(err)
      }
      num += 1
			elem.Value = fmt.Sprintf("%d",num)
		} else {
      num = 1
			db.StoreSet(args[0], &data.Entry{"-1",StringType})
		}
		return reply.IntReply(num)
	})

	store.DefaultDBManager.AddFuncWithSideEffects("incrby", func(db *store.DB, args []string) string {
		elem, ok, _ := db.StoreGet(args[0], StringType)
    var num int

    decrby, err := strconv.Atoi(args[1])
    if err != nil {
      panic(err)
    }

		if ok {
      str, _ := elem.Value.(string)

      num, err := strconv.Atoi(str)
      if err != nil {
        panic(err)
      }

      num += decrby
			elem.Value = fmt.Sprintf("%d",num)
		} else {
      str := fmt.Sprintf("%d",decrby)
			db.StoreSet(args[0], &data.Entry{str,StringType})
		}
    println(num)
		return reply.IntReply(num)
	})

	store.DefaultDBManager.AddFunc("get", func(db *store.DB, args []string) string {
		elem, ok, err := db.StoreGet(args[0], StringType)
		if err != nil {
			return reply.ErrorReply(err)
		}
		if ok {
			str := elem.GetString()
			return reply.BulkReply(str)
		} else {
		  return reply.NilReply
    }
	})

	store.DefaultDBManager.AddFunc("getset", func(db *store.DB, args []string) string {
		elem, ok, err := db.StoreGet(args[0], StringType)
		if err != nil {
			return reply.ErrorReply(err)
		}
		if ok {
			str := elem.GetString()
      elem.Value = args[1]
			return reply.BulkReply(str)
		} else {
      db.StoreSet(args[0],&data.Entry{args[1],StringType})
		  return reply.NilReply
    }
	})

	store.DefaultDBManager.AddFuncWithSideEffects("psetex", func(db *store.DB, args []string) string {
		elem, ok, _ := db.StoreGet(args[0], StringType)

    millis, err := strconv.Atoi(args[1])
    if err != nil {
      panic(err)
    }

    db.ExpiryRecord[args[0]] = store.Milliseconds()+uint64(millis)

		if ok {
			elem.Value = args[2]
		} else {
			db.StoreSet(args[0], &data.Entry{args[2], StringType})
		}

		return reply.OKReply
	})

	store.DefaultDBManager.AddFuncWithSideEffects("setex", func(db *store.DB, args []string) string {
		elem, ok, _ := db.StoreGet(args[0], StringType)

    seconds, err := strconv.Atoi(args[1])
    if err != nil {
      panic(err)
    }

    db.ExpiryRecord[args[0]] = store.Milliseconds()+uint64(seconds)*1000

		if ok {
			elem.Value = args[2]
		} else {
			db.StoreSet(args[0], &data.Entry{args[2], StringType})
		}

		return reply.OKReply
	})

	store.DefaultDBManager.AddFuncWithSideEffects("set", func(db *store.DB, args []string) string {
		elem, ok, _ := db.StoreGet(args[0], StringType)
		if ok {
			elem.Value = args[1]
		} else {
			db.StoreSet(args[0], &data.Entry{args[1], StringType})
		}
		return reply.OKReply
	})

	store.DefaultDBManager.AddFuncWithSideEffects("setnx", func(db *store.DB, args []string) string {
		s := args[0]
		_, ok, _ := db.StoreGet(s, StringType)
		if ok {
			return reply.IntReply(0)
		}
		db.StoreSet(s, &data.Entry{args[1], StringType})
		return reply.IntReply(1)
	})

	store.DefaultDBManager.AddFunc("strlen", func(db *store.DB, args []string) string {
		elem, ok, err := db.StoreGet(args[0], StringType)
		if err != nil {
			return reply.ErrorReply(err)
		}
		if ok {
			str := elem.GetString()
			return reply.IntReply(len(str))
		} else {
		  return reply.IntReply(0)
    }
	})

  
}
