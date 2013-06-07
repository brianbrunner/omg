package data

import (
	"encoding/gob"
)

var Any uint8 = uint8(127)

type Entry struct {
	Value     interface{}
	EntryType uint8
}

func (e Entry) GetString() string {
	if str, ok := e.Value.(string); ok {
		return str
	} else {
		return ""
	}
	return ""
}

func (e Entry) Type() uint8 {
	return (e.EntryType & 0x7F)
}

func (e Entry) IsType(testType uint8) bool {
	return e.Type() == testType || testType == Any
}

func (e *Entry) LastDump() uint8 {
	return (e.EntryType & 0x80) >> 7
}

func (e *Entry) SetLastDump(lastDump uint8) {
	if lastDump == 1 {
		e.EntryType = e.EntryType | 0x80
	} else {
		e.EntryType = e.EntryType & 0x7F
	}
}

func init() {
	gob.Register(&Entry{})
}
