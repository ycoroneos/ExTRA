package main

import (
	"fmt"
	"testing"
)

func TestAverage(t *testing.T) {
	//them := Rollhash("../tests/hashA.txt")
	fast_them := FastRollhash("../tests/hashA.txt")
	//us := Rollhash("../tests/hashB.txt")
	fast_us := FastRollhash("../tests/hashB.txt")
	//fmt.Printf("hash1: %v\n\n", them)
	fmt.Printf("fast hash1: %v\n\n", fast_them)
	//fmt.Printf("hash2: %v\n\n", us)
	fmt.Printf("fast hash2: %v\n\n", fast_us)
	we_need, chunk_deltas := Diff(fast_them, fast_us)
	fmt.Printf("\n\n\nwe need: %v\n\n", we_need)
	for _, c := range chunk_deltas {
		fmt.Printf("chunk %v -> %v\n", c.Chunk, c.Moveto)
	}
}
