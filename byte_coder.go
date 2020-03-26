package tiptop

import (
	"encoding/binary"
)

const (
	timestampSizeInBytes = 8                                                         // Number of bytes used for timestamp
	hashSizeInBytes      = 8                                                         // Number of bytes used for sum64
	crc32SizeInBytes     = 4                                                         // Number of bytes used for CRC32
	headersSizeInBytes   = timestampSizeInBytes + hashSizeInBytes + crc32SizeInBytes // Number of bytes used for all headers
)

// wrapEntry pack the []byte with expiration, the sha1 and crc32 of the key.
func wrapEntry(timestamp int64, hash uint64, crc32 uint32, entry []byte, buffer *[]byte) []byte {
	blobLength := len(entry) + headersSizeInBytes

	if blobLength > len(*buffer) {
		*buffer = make([]byte, blobLength)
	}
	blob := *buffer

	binary.LittleEndian.PutUint64(blob, uint64(timestamp))
	binary.LittleEndian.PutUint64(blob[timestampSizeInBytes:], hash)
	binary.LittleEndian.PutUint32(blob[timestampSizeInBytes+hashSizeInBytes:], crc32)
	copy(blob[headersSizeInBytes:], entry)

	return blob[:blobLength]
}

// readEntry read the value from the package of []byte
func readEntry(data []byte) []byte {
	// copy on read
	dst := make([]byte, len(data)-headersSizeInBytes)
	copy(dst, data[headersSizeInBytes:])

	return dst
}

// readTimestampFromEntry read the expiration from the package of []byte
func readTimestampFromEntry(data []byte) int64 {
	return int64(binary.LittleEndian.Uint64(data))
}

// readHashFromEntry read the sha1 from the package of []byte
func readHashFromEntry(data []byte) uint64 {
	return binary.LittleEndian.Uint64(data[timestampSizeInBytes:])
}

// readCRC32FromEntry read the crc32 from the package of []byte
func readCRC32FromEntry(data []byte) uint32 {
	return binary.LittleEndian.Uint32(data[timestampSizeInBytes+hashSizeInBytes:])
}

// resetKeyFromEntry reset the hash of the package of []byte
func resetKeyFromEntry(data []byte) {
	binary.LittleEndian.PutUint64(data[timestampSizeInBytes:], 0)
}
