package murmur

import (
	"bytes"
	"encoding/binary"
)

func HashBucket(key string, numBuckets int) int {
	hash := HashString(key)
	return int(hash) % numBuckets
}

func HashString(data string) uint32 {
	return HashBytes([]byte(data))
}

var c1, c2 uint32 = 0xcc9e2d51, 0x1b873593
var r1, r2 uint32 = 15, 13
var m, n uint32 = 5, 0xe6546b64

func HashBytes(data []byte) uint32 {

	key_length := len(data)
	if key_length == 0 {
		return 0
	}

	count := key_length / 4

	var h, k uint32

	data_buf := bytes.NewBuffer(data)

	for i := 0; i < count; i++ {

		binary.Read(data_buf, binary.LittleEndian, &k)

		k *= c1
		k = (k << r1) | (k >> (32 - r1))
		k *= c2

		h ^= k
		h = (h << r2) | (h >> (32 - r2))
		h = h*m + n

	}

	k = 0
	remaining_index := count * 4

	switch key_length & 3 {
	case 3:
		k ^= uint32(data[remaining_index+2]) << 16
		fallthrough
	case 2:
		k ^= uint32(data[remaining_index+1]) << 8
		fallthrough
	case 1:
		k ^= uint32(data[remaining_index])
		k *= c1
		k = (k << r1) | (k >> (32 - r1))
		k *= c2
		h ^= k
	}

	h ^= uint32(key_length)
	h ^= h >> 16
	h *= 0x85ebca6b
	h ^= h >> 13
	h *= 0xc2b2ae35
	h ^= h >> 16
	return h

}
