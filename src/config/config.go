package config

import (
  "fmt"
  "os"
  "bufio"
  "strings"
)

var Config = map[string]string {
  "maxmemory": "inf",
  "port": "6379",
  "dbfilename": "dump.odb",
  "dir": "./",
}

func ParseConfigFile(filepath string) {

  f, err := os.Open(filepath)
  if err != nil {
    fmt.Println("Error opening config file: %s", err)
    return
  }
  defer f.Close()

  b := bufio.NewReader(f)
  for {

    str, err := b.ReadString('\n')
    if err != nil {
      break;
    }

    str = strings.TrimSpace(str)
    if len(str) > 0 && str[0] != '#' {
      config_options := strings.SplitAfterN(str," ",2)
      Config[strings.TrimSpace(config_options[0])] = strings.TrimSpace(config_options[1])
    }
  
  }

}
