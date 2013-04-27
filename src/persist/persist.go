package persist

import (
    "os"
    "bytes"
    "time"
)

func StartPersist() (chan string) {
    persistChan := make(chan string,1000000)
    go func() {
        f, err := os.OpenFile("dump.omg", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
        if err != nil {
            panic(err)
        }

        defer f.Close()
        
        clock := time.Tick(1 * time.Second)

        var comBuffer bytes.Buffer

        for {
            select {
            case command := <-persistChan:
                comBuffer.WriteString(command)
            case <-clock:
                go func() {
                    if _, err = f.WriteString(comBuffer.String()); err != nil {
                        panic(err)
                    }
                    comBuffer.Reset()
                }()
            }
        }
    }()
    return persistChan
}
