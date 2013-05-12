package persist

import (
    "bufio"
    "os"
    "bytes"
    "time"
    "parser"
)

func LoadAppendOnlyFile() {
    return // TODO fix this shit
    f, err := os.OpenFile("dump.omg", os.O_RDONLY|os.O_CREATE, 0600)
    if err != nil {
        panic(err)
    }    

    b := bufio.NewReader(f)
    c := parser.CommandParser{}
    //reply := make(chan string)
    for {
        line, err := b.ReadString('\n')
        if err != nil { // EOF, or worse
            break
        }
        done, _, _, err := c.AddArg(line)
        if err != nil {
            continue
        }
        if done {
            //comChan <- store.Command{command,commandRaw,reply}
            //b := []byte(<-reply)
        }
    }    

}

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
            case <-clock:
                if _, err = f.WriteString(comBuffer.String()); err != nil {
                    panic(err)
                }
                comBuffer.Reset()
            case command := <-persistChan:
                command = command
                comBuffer.WriteString(command)
            }
        }
    }()
    return persistChan
}
