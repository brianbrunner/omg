package persist

import (
    "os"
    "bytes"
    "time"
    "store/com"
    "bufio"
    "parser"
    "fmt"
    //"config"
    //"strconv"
)

var persistDisabled bool = true

func LoadAppendOnlyFile(comChan chan com.Command) {
  f, err := os.Open("./db/store.oaf")
  if err != nil {
    return
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
      fmt.Println(err)
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

func StartPersist(comChan chan com.Command) (chan string, chan uint8) {

    persistChan := make(chan string)
    stateChan := make(chan uint8)

    go func() {
        AOF, err := os.OpenFile("./db/store.oaf", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
        if err != nil {
            panic(err)
        }

        var saveAOF *os.File

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

                if _, err = AOF.WriteString(comBuffer.String()); err != nil {
                    AOF.Close()
                    saveAOF.Close()
                    panic(err)
                }
                if err = AOF.Sync(); err != nil {
                    AOF.Close()
                    saveAOF.Close()
                    panic(err)
                }
           
                if saveAOF != nil {

                  if _, err = saveAOF.WriteString(comBuffer.String()); err != nil {
                      AOF.Close()
                      saveAOF.Close()
                      panic(err)
                  }
                  if err = saveAOF.Sync(); err != nil {
                      AOF.Close()
                      saveAOF.Close()
                      panic(err)
                  }

                }
           
                comBuffer.Reset()

            case <-odb_tick:

                continue
                if persistDisabled {
                    continue
                }

                comChan <- com.Command{[]string{"bgsave"},"",reply}
                <-reply

            case command := <-persistChan:
                if persistDisabled {
                    continue
                }
                comBuffer.WriteString(command)
            case state := <- stateChan:

                if state == 0 {

                  saveAOF, err = os.Create("/tmp/store.oaf")
                  if err != nil {
                    stateChan <- 0
                  } else {
                    stateChan <- 1
                  }

                } else if state == 1 {

                  saveAOF.Close()
                  stateChan <- 1

                } else if state == 2 {

                  if saveAOF != nil {

                    if err = saveAOF.Sync(); err != nil {
                      stateChan <- 0
                    }

                    if err := os.Rename("/tmp/store.oaf","./db/store.oaf"); err != nil {
                      stateChan <- 0
                    }

                    saveAOF = nil

                    AOF, err = os.OpenFile("./db/store.oaf", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
                    if err != nil {
                        panic(err)
                    }

                    stateChan <- 1
      
                  } else {

                    stateChan <- 0

                  }

                }
            }
        }
    }()
    return persistChan, stateChan
}
