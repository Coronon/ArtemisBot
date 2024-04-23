package git

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/coronon/artemisbot/internal/util"
)

// Create a git commit object and return its compressed data and decompressed size
func CreateCommitObject(tree, parent, author, committer, message string) ([]byte, int) {
	now := time.Now()
	_, offset := now.Zone()
	var sign string
	if offset < 0 {
		sign = "-"
		offset = -offset
	} else {
		sign = "+"
	}

	offsetDuration := time.Duration(offset) * time.Second
	offsetHours := offsetDuration / time.Hour
	offsetDuration -= offsetHours * time.Hour
	offsetMinutes := offsetDuration / time.Minute

	timeSuffix := fmt.Sprintf(" %d %s%02d%02d", now.Unix(), sign, offsetHours, offsetMinutes)
	author += timeSuffix
	committer += timeSuffix

	content := fmt.Sprintf("tree %s\nparent %s\nauthor %s\ncommitter %s\n\n%s", tree, parent, author, committer, message)
	header := fmt.Sprintf("commit %d\x00", len(content))
	obj := header + content

	var buff bytes.Buffer
	zw := zlib.NewWriter(&buff)
	zw.Write([]byte(obj))
	zw.Close()

	return buff.Bytes(), len(obj)
}

// Create a git packed object
func CreatePackedObject(obj []byte, decompressedSize int) []byte {
	pack := append([]byte("PACK"), []byte{0, 0, 0, 2, 0, 0, 0, 1}...)

	// Meta data and variable length integers
	metaBytes := encodeAsVariableLengthInt(decompressedSize)
	metaBytes[0] |= 0x90
	pack = append(pack, metaBytes...)

	// Object data
	pack = append(pack, obj...)

	// Hash
	hasher := sha1.New()
	hasher.Write(pack)
	hash := hasher.Sum(nil)

	pack = append(pack, hash...)

	return pack
}

// Encode an integer as a variable length integer in git's packfile format
//
// Editors note: This is infuriatingly complicated and weird and has cost me
// hours to figure out because the documentation is really lacking!
//
// I finally found a wonderful diagram here:
// https://github.com/robisonsantos/packfile_reader
func encodeAsVariableLengthInt(decompressedSize int) []byte {
	leftmostSetBit := util.FindLeftmostSetBit(decompressedSize)

	bytes := []byte{}

	currentByte := 0x00
	currentByteIndex := 0
	isFirst := true
	for i := 0; i < leftmostSetBit; i++ {
		currentByte |= ((decompressedSize >> i) & 1) << currentByteIndex

		currentByteIndex++

		goToNextByte := currentByteIndex > 6 || isFirst && currentByteIndex > 3
		if goToNextByte {
			if (decompressedSize >> i) > 0 {
				currentByte |= 0x80
			}

			bytes = append(bytes, byte(currentByte))
			currentByte = 0x00
			currentByteIndex = 0
			isFirst = false
		}
	}
	if currentByteIndex > 0 {
		bytes = append(bytes, byte(currentByte))
	}

	return bytes
}
