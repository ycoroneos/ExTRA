package main

import (
	"os"
)

type FileChunk struct {
	Offset   int64
	Checksum uint32
}

func adler32(window []uint32, num uint32) ([]uint32, uint32) {
	output := append(window[1:], num)
	sum := uint32(0)
	for _, v := range output {
		sum += v
	}
	return output, sum
}

//computes the rolling checksum
func Rollhash(filename string) []FileChunk {
	chunks := make([]FileChunk, 0)
	fd, err := os.Open(filename)
	if !check(err, true) {
		return chunks
	}
	defer fd.Close()
	stat, err := fd.Stat()
	if !check(err, true) {
		return chunks
	}
	size := stat.Size()
	checksum := uint32(0)
	window := make([]uint32, 8192)
	//thebyte := make([]byte, 1)
	data := make([]byte, 8192)
	offset := int64(0)
	for i := int64(0); i < size; {
		//if i%100 == 0 {
		//	fmt.Printf("%d / %d\n", i, size)
		//	}
		n, err := fd.Read(data)
		if !check(err, true) || n == 0 {
			return make([]FileChunk, 0)
		}
		for j := int64(0); j < int64(n); j++ {
			window, checksum = adler32(window, uint32(data[j]))
			//			if i > 8192 {
			//				checksum -= window[windowspot]
			//			}
			//			//		n, err := fd.Read(thebyte)
			//			//		if !check(err, true) || n == 0 {
			//			//			return make([]FileChunk, 0)
			//			//		}
			//			checksum += uint32(data[j])
			//			window[windowspot] = uint32(data[j])
			//			windowspot = (windowspot + 1) % 8192
			if checksum%4096 == 0 && (i+j) > 4096 {
				chunks = append(chunks, FileChunk{offset, checksum})
				offset = i + j
			}
		}
		i += int64(n)
	}
	//trailing checksum
	if len(chunks) == 0 || checksum != chunks[len(chunks)-1].Checksum {
		chunks = append(chunks, FileChunk{offset, checksum})
	}
	return chunks
}
