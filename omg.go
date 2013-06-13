package main

import (
  "bufio"
  "fmt"
  "net"
  "parser"
  "store"
  "flag"
  "log"
  "os"
  "os/signal"
  "runtime/pprof"
  "store/com"
  "persist"
  "config"
  "runtime"
  "strings"

  // import all of our db functions
  _ "funcs"
)

func handleConn(client net.Conn, comChan chan com.Command) {

  // Buffer the current client

  b := bufio.NewReader(client)

  // Set up and initialize the command parser for this connection

  c := parser.CommandParser{}
  c.Reset()
  c.Restart()

  // Initialize varrious channels, buffers, and state variables

  reply := make(chan string)
  line := make([]byte, 1024)
  checkSync := true

  // Define variables we're going to set on every single loop

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

        // If the first command we ever see is a "SYNC" command, bypass 
        // the command processor and set this connection up as a replica
        if checkSync {

          command_key := strings.ToLower(command.Args[0])
          switch command_key {
          case "sync":
            persist.StartSync(client)
            return
          }
          checkSync = false

        }

        // make sure we get replies for all commands we send to the db

        command.ReplyChan = reply

        //send our command and wait for a repsonse

        comChan <- command
        client.Write([]byte(<-reply))

      }
      c.Reset()

    }
  }
}

var cpuprofile = flag.String("cpu", "", "write cpu profile to file")
var memprofile = flag.String("mem", "", "write memory profile to this file")
var configfile = flag.String("config", "./omg.conf", "read configuration options from this file")

func main() {

  // Parse command line flags

  flag.Parse()

  // Set the go runtime up to use all available CPUs

  num_cpus := runtime.NumCPU()
  fmt.Println("Running with",num_cpus,"procs")
  runtime.GOMAXPROCS(num_cpus)

  // Parse config file

  config.ParseConfigFile(*configfile)

  // Start the persistance manager for our DB

  store.DefaultDBManager.StartPersist()

  // CPU and Memory Profiler Code

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

  port := fmt.Sprintf(":%s",config.Config["port"])
  ln, err := net.Listen("tcp", port)
  if err != nil {
    fmt.Println("%s", err)
  } else {

    // get a reference to the global database manager

    dbm := store.DefaultDBManager

    // load data from the odb and oaf files

    fmt.Println("Loading saved database...")

    dbm.LoadFromDiskSync()
    persist.LoadAppendOnlyFile(dbm.ComChan)

    fmt.Println("DB is open for business")

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
