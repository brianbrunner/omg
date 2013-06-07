package persist

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"parser"
	"store/com"
	"time"
	"config"
	"strconv"
  "io"
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

var replicaChan chan io.Writer

// replicas that are actively receiving commands 
// from the server 
var activeReplicas []io.Writer = make([]io.Writer,0)

// Replicas that will receive a copy of the dump 
// file once the current background save finishes
var dumpToReplicas []io.Writer = make([]io.Writer,0)

func dumpFileToReplicas(filename string) {

  dump, err := os.Open(filename)
  if err != nil {
    panic(err)
  }

  line := make([]byte,1024)
  for {
    n, err := dump.Read(line[:])
    if err == io.EOF {
      break
    }

    for _, replica := range dumpToReplicas {
      replica.Write(line[:n])
    }

  }
}

func StartPersist(comChan chan com.Command) (chan string, chan uint8) {

	persistChan := make(chan string)
	stateChan := make(chan uint8)
  
  replicaChan = make(chan io.Writer,100)

	go func() {
		AOF, err := os.OpenFile("./db/store.oaf", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			panic(err)
		}

		var saveAOF *os.File

    aof_tick := make(<-chan time.Time)
    aof_interval, err := strconv.Atoi(config.Config["aof_interval"])
    if err != nil {
      panic("Invalid integer for aof_interval")
    }
    if aof_interval > 0 {
		  aof_tick = time.Tick(time.Duration(aof_interval) * time.Second)
    }
		odb_tick := time.Tick(120 * time.Second)

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

				comChan <- com.Command{[]string{"bgsave"}, "bgsave", reply}
				<-reply

      case replica := <-replicaChan:

        dumpToReplicas = append(dumpToReplicas,replica)
				comChan <- com.Command{[]string{"bgsave"}, "bgsave", reply}
        go func() {
				  <-reply
        }()

			case command := <-persistChan:

				if persistDisabled {
					continue
				}

				comBuffer.WriteString(command)

        for _, replica := range activeReplicas {
          replica.Write([]byte(command))
        }

        if aof_interval == 0 {

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

        }

			case state := <-stateChan:

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

          dumpFileToReplicas("./db/store.odb")
          for _, replica := range dumpToReplicas {
            replica.Write([]byte(""))
          }

					if saveAOF != nil {

						if err = saveAOF.Sync(); err != nil {
							stateChan <- 0
						}

						if err := os.Rename("/tmp/store.oaf", "./db/store.oaf"); err != nil {
							stateChan <- 0
						}

						saveAOF = nil

            if len(dumpToReplicas) > 0 {

              AOFDump, err := os.Open("./db/store.oaf")
              if err != nil {
                panic(err)
              }

              aof_line := make([]byte,1024)
              for {

                n, err := AOFDump.Read(aof_line[:])
                if err != nil {
                  break
                }

                for _, replica := range dumpToReplicas {
                  replica.Write(aof_line[:n])
                }

              }
              AOFDump.Close()

            }

            for _, replica := range dumpToReplicas {
              activeReplicas = append(activeReplicas, replica)
            }
            dumpToReplicas = make([]io.Writer,0)

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

func StartSync(client io.Writer) {
   replicaChan <- client
}
