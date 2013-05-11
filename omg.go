package main

import (
    "fmt"
    "net"
    "bufio"
    "parser"
    "store"
    //"tracker"
    //"time"
    _ "funcs"
    "log"
    "os"
    "flag"
    "runtime/pprof"
    "os/signal"
)

func handleConn(client net.Conn, comChan chan store.Command) {
    b := bufio.NewReader(client)
    c := parser.CommandParser{}
    reply := make(chan string)
    var line string
    var err error
    var done bool
    var command []string
    var commandRaw string
    //var lap = tracker.Lapper{time.Now()}
    for {
        line, err = b.ReadString('\n')
        //lap.Lap("Read")
        if err != nil { // EOF, or worse
            fmt.Println("Connection closed")
            break
        }
        done, command, commandRaw, err = c.AddArg(line)
        if err != nil {
            continue
        }
        if done {
            //lap.Lap("Command parsed")
            comChan <- store.Command{command,commandRaw,reply}
            b := []byte(<-reply)
            //lap.Lap("Command executed")
            client.Write(b)
            //lap.Lap("Response written")
        }
    }
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var memprofile = flag.String("memprofile", "", "write memory profile to this file")

func main() {

    flag.Parse()

    if *cpuprofile != "" {
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
        c := make(chan os.Signal, 1)
        signal.Notify(c, os.Interrupt)
        go func(){
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

    

    ln, err := net.Listen("tcp", ":6379")
    if err != nil {
        fmt.Println("%s",err)   
    } else {
        dbm := store.DefaultDBManager
        for {
            conn, err := ln.Accept()
            if err != nil {
                fmt.Println("%s",err)
            } else {
                fmt.Println("New connection")
                go handleConn(conn,dbm.ComChan)
            }
        }
        close(dbm.ComChan)
    }
}
