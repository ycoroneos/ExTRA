package main

import (
	"fmt"
	"testing"
)

func TestAverage(t *testing.T) {
	h1 := Rollhash("../tests/hashA.txt")
	h2 := Rollhash("../tests/hashB.txt")
	fmt.Printf("hash1: %v\n", h1)
	fmt.Printf("hash2: %v\n", h2)
}
