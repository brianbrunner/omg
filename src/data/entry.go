package data 

type EntryType uint8
const (
    Any = iota
    Str
    Hash
    Set
    SortedSet
    User
    Feed
) 

type Entry struct { 
    Value interface{}
    EType EntryType
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
