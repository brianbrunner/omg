package archive 

import (
    "store"
    "fmt"
    "data"
    "os"
    "io/ioutil"
    "store/reply"
)

type ArchiveJob struct {
    key string
    data []byte 
}

var ArchiveType uint8 = 19 
var archive_chan chan ArchiveJob = make(chan ArchiveJob, 100000)

func init() {

    go func() {

        for {
            archiveJob := <- archive_chan
            filename := fmt.Sprintf("./db/archive/%s.omg",archiveJob.key)
            f, err := os.Create(filename)
            if err != nil {
                panic(err)
            }

            if _, err = f.Write(archiveJob.data); err != nil {
                panic(err)
            }
            f.Close()
        }
    }()

    store.RegisterStoreType(ArchiveType)    
    store.DefaultDBManager.AddFuncWithSideEffects("archive", func(db *store.DB, args []string) string {
        elem, ok, _ := db.StoreGet(args[0],data.Any)
        if ok {
            if elem.EntryType != ArchiveType {
                str_rep, ok := db.StoreDump(args[0])
                if ok {
                    archive_chan <- ArchiveJob{args[0],str_rep}
                    db.StoreSet(args[0],&data.Entry{0,ArchiveType})
                    return reply.IntReply(1)
                } else {
                    return reply.NilReply
                }
            } else {
                return reply.IntReply(0)
            }
        } else {
            return reply.NilReply
        }
    })
    store.DefaultDBManager.AddFuncWithSideEffects("unarchive", func(db *store.DB, args []string) string {
        _, ok, _ := db.StoreGet(args[0],ArchiveType)
        if ok {
            filename := fmt.Sprintf("./db/archive/%s.omg",args[0])
            data, err := ioutil.ReadFile(filename)
            if err != nil {
                panic(err)
            }
            db.StoreLoad(args[0],data)
            return reply.OKReply
        } else {
            return reply.NilReply
        }
    })
}
