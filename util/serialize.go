package util

import "encoding/binary"

func SerializeInt64(value int64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, uint64(value))
	return buf
}

func DeserializeInt64(b []byte) int64 {
	return int64(binary.LittleEndian.Uint64(b))
}
