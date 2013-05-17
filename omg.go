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

	// import all of our db functions
	_ "funcs"
)

func handleConn(client net.Conn, comChan chan com.Command) {
	b := bufio.NewReader(client)
	c := parser.CommandParser{}
    c.Reset()
	reply := make(chan string)
	line := make([]byte, 1024)
	var err error
	var done bool
    var n int
	var commands []com.Command
	for {
		n, err = b.Read(line[:])
		if err != nil { 
			fmt.Println("Connection closed")
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
                client.Write([]byte(<-reply))
            }
            c.Reset()
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

	ln, err := net.Listen("tcp", ":6379")
	if err != nil {
		fmt.Println("%s", err)
	} else {
		dbm := store.DefaultDBManager

		reply := make(chan string)
		dbm.ComChan <- com.Command{[]string{"load"}, "", reply}
		<-reply

		for {
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
