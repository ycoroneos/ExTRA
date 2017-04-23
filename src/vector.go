package main

import (
	"strconv"
)

//contains definitions and methods for version vectors

//Can be used to either represent version or sync pair
type Pair struct {
	Id      string
	Counter int
}

//either version vector or time synchronization vector
//basically just a set
type PairVec struct {
	Pairs map[string]Pair
}

//checks if b is a superset of a
func LEQ(a, b PairVec) bool {
	for k, v := range a.Pairs {
		val, exists := b.Pairs[k]
		if exists && val.Counter < v.Counter {
			return false
		}
	}
	return true
}

//add a new element to the set iff it is newer or does not exist
func (pv PairVec) Add(e Pair) PairVec {
	_, exists := pv.Pairs[e.Id]
	if !exists {
		pv.Pairs[e.Id] = e
	} else {
		if pv.Pairs[e.Id].Counter < e.Counter {
			pv.Pairs[e.Id] = e
		}
	}
	return pv
}

//for making a new one
func MakePairVec(initial []Pair) PairVec {
	out := PairVec{make(map[string]Pair)}
	for _, i := range initial {
		out = out.Add(i)
	}
	return out
}

//access function
func (pv PairVec) GetPair(key string) (Pair, bool) {
	val, exists := pv.Pairs[key]
	return val, exists
}

func (pv PairVec) Show() string {
	out := "<"
	for _, v := range pv.Pairs {
		out += v.Id + "_" + strconv.Itoa(v.Counter) + ", "
	}
	out += ">"
	return out
}