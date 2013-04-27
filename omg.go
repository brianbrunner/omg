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

func main() {
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
