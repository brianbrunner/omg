package data 

import (
    "encoding/gob"
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

func init() {
    gob.Register(&Entry{})
}
