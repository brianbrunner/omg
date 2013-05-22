package persist

import (
    "os"
    "bytes"
    "time"
    "store/com"
    "bufio"
    "parser"
)

var persistDisabled bool = true

func LoadAppendOnlyFile(comChan chan com.Command) {
    f, err := os.Open("./db/store.oaf")
    if err != nil {
        panic(err)
    }

	b := bufio.NewReader(f)
	c := parser.CommandParser{}
    c.Reset()
    c.Restart()
	reply := make(chan string)
	line := make([]byte, 1024)
	var done bool
    var n int
	var commands []com.Command
	for {
		n, err = b.Read(line[:])
		if err != nil { 
			break
		}
        done, commands, err = c.ParseBytes(line[:n])
        if err != nil {
            break
        }
        if done {
            for _, command := range commands {
                command.ReplyChan = reply
                comChan <- command
                <-reply
            }
            c.Reset()
        }
	}

    persistDisabled = false
    
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
                if persistDisabled {
                    continue
                }
                if _, err = f.WriteString(comBuffer.String()); err != nil {
                    panic(err)
                }
                if err = f.Sync(); err != nil {
                    panic(err)
                }
                comBuffer.Reset()
            case <-odb_tick:
                if persistDisabled {
                    continue
                }
                comChan <- com.Command{[]string{"SAVE"},"",reply}
                <-reply
                f, err = os.Create("./db/store.oaf")
                if err != nil {
                    panic(err)
                }
            case command := <-persistChan:
                if persistDisabled {
                    continue
                }
                command = command
                comBuffer.WriteString(command)
            }
        }
    }()
    return persistChan
}
