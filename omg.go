package main

import (
  "bufio"
  "fmt"
  "net"
  "parser"
  "store"
  //"tracker"
  //"time"
  "flag"
  "log"
  "os"
  "os/signal"
  "runtime/pprof"
  "store/com"
    "persist"

  // import all of our db functions
  _ "funcs"
)

func handleConn(client net.Conn, comChan chan com.Command) {

  // Set up the operating state for the current connection

  b := bufio.NewReader(client)
  c := parser.CommandParser{}
  c.Reset()
  c.Restart()
  reply := make(chan string)
  line := make([]byte, 1024)
  var err error
  var done bool
  var n int
  var commands []com.Command

  for {

    // wait for commands to come down the pipe

    n, err = b.Read(line[:])
    if err != nil { 
      fmt.Println("Connection closed")
      break
    }
  
    // process bytes received from the client

    done, commands, err = c.ParseBytes(line[:n])
    if err != nil {
      break
    }

    // if we've processed any commands, run them

    if done {

      for _, command := range commands {
        command.ReplyChan = reply
        comChan <- command
        client.Write([]byte(<-reply))
      }
      c.Reset()

    }
  }
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write memory profile to this file")

func main() {

  //
  // CPU and Memory Profiler Code
  //

  flag.Parse()

  if *cpuprofile != "" {
    f, err := os.Create(*cpuprofile)
    if err != nil {
      log.Fatal(err)
    }
    pprof.StartCPUProfile(f)
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func() {
      for _ = range c {
        pprof.StopCPUProfile()
        if *memprofile != "" {
          f, err := os.Create(*memprofile)
          if err != nil {
            log.Fatal(err)
          }
          pprof.WriteHeapProfile(f)
          f.Close()
        }
        os.Exit(1)
      }
    }()
  }

  // Create a server to listen for connections

  ln, err := net.Listen("tcp", ":6379")
  if err != nil {
    fmt.Println("%s", err)
  } else {

    // get a reference to the global database manager

    dbm := store.DefaultDBManager

    // load data from the odb and oaf files

    dbm.LoadFromDiskSync()
    persist.LoadAppendOnlyFile(dbm.ComChan)

    for {

      // wait for connections to come in

      conn, err := ln.Accept()
      if err != nil {
        fmt.Println("%s", err)
      } else {
        fmt.Println("New connection")
        go handleConn(conn, dbm.ComChan)
      }
    }

  }
}
