package replication

import (
  "net"
  "store"
  "fmt"
  "store/reply"
  "parser"
  "store/com"
  "encoding/gob"
)

func init() {

  slaveChan := make(chan string)

  var slaveConn net.Conn

  go func() {

    c := parser.CommandParser{}
    c.Reset()
    c.Restart()
    replyChan := make(chan string)
    line := make([]byte, 1024)
    var done bool
    var err error
    var commands []com.Command
    comChan := store.DefaultDBManager.ComChan

    for {

      master := <-slaveChan

      if slaveConn != nil {
        slaveConn.Close()
        slaveConn = nil
      }

      if master != "NO:ONE" {

        slaveConn, err = net.Dial("tcp", master)
        if err == nil {

          slaveConn.Write([]byte("*1\r\n$4\r\nSYNC\r\n"))

          go func() {

            db_dec := gob.NewDecoder(slaveConn)
            store.DefaultDBManager.LoadFromGobDecoder(db_dec)

            for {
              n, err := slaveConn.Read(line[:])
              if err != nil {
                return
              }

              done, commands, err = c.ParseBytes(line[:n])
              if err != nil {
                break
              }

              // if we've processed any commands, run them

              if done {

                for _, command := range commands {
                  command.ReplyChan = replyChan
                  comChan <- command
                  <-replyChan
                }
                c.Reset()

              }

            }

          }()

        } else {
          slaveChan <- reply.ErrorReply(fmt.Sprintf("Unable to connect to server at %s",master))
          continue
        }
      }

      slaveChan <- reply.OKReply

    }

  }()

	store.DefaultDBManager.AddFunc("slaveof", func(db *store.DB, args []string) string {
    ip := args[0]
    port := args[1]
    slaveChan <- fmt.Sprintf("%s:%s",ip,port)
		return <-slaveChan
	})

  store.DefaultDBManager.AddFunc("movebucket", func(db *store.DB, args []string) string {
    ip := args[0]
    port := args[1]
    bucket, err := strconv.Atoi(args[2])
    if err != nil {
      return reply.ErrorReply("Invalid bucket")
    }
    slaveChan <- fmt.Sprintf("%s:%s",ip,port)
		return <-slaveChan
  })

}
