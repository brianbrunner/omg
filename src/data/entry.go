package data 

import (
    "fmt"
)

var Any int = -1

type Entry struct { 
    Value interface{}
    EntryType int
    Expires uint64
}

func (e Entry) GetString() string {
    if str, ok := e.Value.(string); ok {
        return str
    } else {
        return ""
    }
    return ""
}

var usedTypes map[int]bool
func init() {
    usedTypes = make(map[int]bool)
}

func RegisterStoreType(entryNum int) {
    _, ok := usedTypes[entryNum]
    if !ok {
        usedTypes[entryNum] = true
    } else {
        panic(fmt.Sprintf("You've used the same identifier, \"%i\", for multiple types",entryNum))
    } 
}
