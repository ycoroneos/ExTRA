package main

import (
	"fmt"
	"os"
	"testing"
)

func TestAverage(t *testing.T) {
	//them := Rollhash("../tests/hashA.txt")
	fast_them := RollhashSha("../tests/hashA.txt")
	//us := Rollhash("../tests/hashB.txt")
	fast_us := RollhashSha("../tests/hashB.txt")
	//fmt.Printf("hash1: %v\n\n", them)
	fmt.Printf("fast hash1: %v\n\n", fast_them)
	//fmt.Printf("hash2: %v\n\n", us)
	fmt.Printf("fast hash2: %v\n\n", fast_us)
	we_need, chunk_deltas := Diff(fast_them, fast_us)
	fmt.Printf("\n\n\nwe need: %v\n\n", we_need)
	fmt.Printf("\n\n\n we have %v chunks\n", len(chunk_deltas))
	for _, c := range chunk_deltas {
		fmt.Printf("chunk %v -> %v\n", c.Chunk, c.Moveto)
	}

	filename := "../tests/hashB.txt"
	their_filename := "../tests/hashA.txt"
	//move around the chunks we have
	theirfd, _ := os.Open(their_filename)
	oldfd, _ := os.Open(filename)
	fd, _ := os.OpenFile(filename+"_new", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	defer oldfd.Close()
	defer fd.Close()
	defer theirfd.Close()
	for _, chunk := range chunk_deltas {
		//DPrintf("moving chunk %v -> %v", chunk.Chunk, chunk.Moveto)
		buf := make([]byte, chunk.Chunk.Size)
		oldfd.ReadAt(buf, chunk.Chunk.Offset)
		fd.WriteAt(buf, chunk.Moveto)
	}
	for _, chunk := range we_need {
		buf := make([]byte, chunk.Size)
		theirfd.ReadAt(buf, chunk.Offset)
		fd.WriteAt(buf, chunk.Offset)
	}

}
