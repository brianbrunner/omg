package user

import (
    "store"
    "fmt"
    "data"
    "os"
    "io/ioutil"
)

type ArchiveJob struct {
    key string
    data string
}

var ArchiveType int = 19 
var archive_chan chan ArchiveJob = make(chan ArchiveJob, 100000)

func init() {

    go func() {

        for {
            archiveJob := <- archive_chan
            filename := fmt.Sprintf("./db/archive/%s.omg",archiveJob.key)
            f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0600)
            if err != nil {
                panic(err)
            }

            if _, err = f.WriteString(archiveJob.data); err != nil {
                panic(err)
            }
            f.Close()
        }
    }()

    data.RegisterStoreType(ArchiveType)    
    store.DefaultDBManager.AddFunc("archive", func(db *store.DB, args []string) string {
        elem, ok, _ := db.StoreGet(args[0],data.Any)
        if ok {
            if elem.EntryType != ArchiveType {
                str_rep, ok := db.StoreBase64Dump(args[0])
                if ok {
                    archive_chan <- ArchiveJob{args[0],str_rep}
                    db.StoreSet(args[0],&data.Entry{0,19,0})
                    return ":1\r\n"
                } else {
                    return "$-1\r\n"
                }
            } else {
                return ":0\r\n"
            }
        } else {
            return "$-1\r\n"
        }
    })
    store.DefaultDBManager.AddFunc("unarchive", func(db *store.DB, args []string) string {
        _, ok, _ := db.StoreGet(args[0],ArchiveType)
        if ok {
            filename := fmt.Sprintf("./db/archive/%s.omg",args[0])
            data, err := ioutil.ReadFile(filename)
            if err != nil {
                panic(err)
            }
            db.StoreBase64Load(args[0],string(data))
            return "+OK\r\n"
        } else {
            return "$-1\r\n"
        }
    })
}
