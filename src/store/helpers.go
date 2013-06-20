// Package store implements nearly all of the core database and 
// database management methods
package store

import (
	"encoding/base64"
	"fmt"
	"runtime"
	"time"
)

// Base64Encode is a helper function that encodes a byte array
// to base64 
func Base64Encode(input []byte) []byte {
	enc := base64.StdEncoding
	encLength := enc.EncodedLen(len(input))
	output := make([]byte, encLength)
	enc.Encode(output, input)
	return output
}

// Base64Decode is a helper function that decodes a base64
// encoded byte array
func Base64Decode(input []byte) []byte {
	dec := base64.StdEncoding
	decLength := dec.DecodedLen(len(input))
	output := make([]byte, decLength)
	n, err := dec.Decode(output, input)
	if err != nil {
		panic(err)
	}
	if n < decLength {
		output = output[:n]
	}
	return output
}

var lastMillis uint64

func Milliseconds() uint64 {
	millis := uint64(time.Now().UnixNano() / 1000000)
	if millis != lastMillis {
		DefaultDBManager.PersistChan <- fmt.Sprintf("&%d\r\n", millis)
		lastMillis = millis
	}
	return millis
}

func MemInUse() uint64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats.HeapAlloc
}

var usedTypes map[uint8]uint8 = make(map[uint8]uint8)

func RegisterStoreType(entryNum uint8) {
	_, ok := usedTypes[entryNum]
	if !ok {
		usedTypes[entryNum] = 1
	} else {
		panic(fmt.Sprintf("You've used the same identifier, \"%i\", for multiple types", entryNum))
	}
}

