package main

import (
	"fmt"
	"testing"
)

func TestAverage(t *testing.T) {
	them := Rollhash("../tests/hashA.txt")
	us := Rollhash("../tests/hashB.txt")
	fmt.Printf("hash1: %v\n", them)
	fmt.Printf("hash2: %v\n", us)
	we_need := CompareChunks(them, us)
	fmt.Printf("we need: %v\n", we_need)
}
