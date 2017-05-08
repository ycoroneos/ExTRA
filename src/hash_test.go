package main

import (
	"fmt"
	"testing"
)

func TestAverage(t *testing.T) {
	them := Rollhash("../tests/hashA.txt")
	us := Rollhash("../tests/hashB.txt")
	fmt.Printf("hash1: %v\n\n", them)
	fmt.Printf("hash2: %v\n\n", us)
	we_need, chunk_deltas := ChompAlgo(them, us)
	fmt.Printf("we need: %v\n\n", we_need)
	for _, c := range chunk_deltas {
		fmt.Printf("chunk %v -> %v\n", c.Chunk, c.Moveto)
	}
}
