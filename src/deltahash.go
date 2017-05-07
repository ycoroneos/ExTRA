package main

import (
	"os"
)

type FileChunk struct {
	Offset   int64
	Checksum uint32
	Size     int64
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
	data := make([]byte, 8192)
	offset := int64(0)
	for i := int64(0); i < size; {
		n, err := fd.Read(data)
		if !check(err, true) || n == 0 {
			return make([]FileChunk, 0)
		}
		for j := int64(0); j < int64(n); j++ {
			window, checksum = adler32(window, uint32(data[j]))
			if checksum%4096 == 0 && (i+j) > 4096 {
				chunks = append(chunks, FileChunk{offset, checksum, (i + j) - offset})
				offset = i + j
			}
		}
		i += int64(n)
	}
	//trailing checksum
	if len(chunks) == 0 || checksum != chunks[len(chunks)-1].Checksum {
		chunks = append(chunks, FileChunk{offset, checksum, size - offset})
	}
	return chunks
}

type ChunkDelta struct {
	Chunk  FileChunk
	Moveto int64
}

func CompareChunks(them, ours []FileChunk) ([]FileChunk, []ChunkDelta) {
	need := make([]FileChunk, 0)
	have := make([]ChunkDelta, 0)
	index := 0
	for i := 0; i < len(ours); {
		if ours[i].Checksum == them[index].Checksum {
			have = append(have, ChunkDelta{ours[i], them[index].Offset})
			index += 1
			i += 1
		} else {
			need = append(need, them[index])
			index += 1
		}
	}
	return need, have
}

type DataChunk struct {
	Chunk FileChunk
	Data  []byte
}

func Readchunks(filename string, chunks []FileChunk) chan DataChunk {
	output := make(chan DataChunk, 0)
	go func() {
		defer close(output)
		fd, err := os.Open(filename)
		if !check(err, true) {
			return
		}
		defer fd.Close()
		for _, c := range chunks {
			data := make([]byte, c.Size)
			fd.ReadAt(data, c.Offset)
			output <- DataChunk{c, data}
		}
	}()
	return output
}
