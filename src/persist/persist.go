package persist

import (
    "os"
    "bytes"
    "time"
    "store/com"
)

func LoadAppendOnlyFile() {
    return // TODO fix this shit
}

func StartPersist(comChan chan com.Command) (chan string) {
    persistChan := make(chan string,1000000)

    go func() {
        f, err := os.OpenFile("./db/store.oaf", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
        if err != nil {
            panic(err)
        }

        defer f.Close()
        
        aof_tick := time.Tick(1 * time.Second)
        odb_tick := time.Tick(60 * time.Second)

        reply := make(chan string)

        var comBuffer bytes.Buffer

        for {
            select {
            case <-aof_tick:
                if _, err = f.WriteString(comBuffer.String()); err != nil {
                    panic(err)
                }
                comBuffer.Reset()
            case <-odb_tick:
                comChan <- com.Command{[]string{"SAVE"},"",reply}
                <-reply
                f, err = os.Create("./db/store.oaf")
                if err != nil {
                    panic(err)
                }
            case command := <-persistChan:
                command = command
                comBuffer.WriteString(command)
            }
        }
    }()
    return persistChan
}
