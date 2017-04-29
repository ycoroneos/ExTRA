package main

import (
	"strconv"
)

//contains definitions and methods for version vectors

//Can be used to either represent version or sync pair
type Pair struct {
	Id      string
	Counter int64
}

//either version vector or time synchronization vector
//basically just a set
type PairVec struct {
	Pairs map[string]Pair
}

func (p PairVec) GetSlice() []Pair {
	out := make([]Pair, 0)
	for _, v := range p.Pairs {
		out = append(out, v)
	}
	return out
}

////checks if b is a superset of a
//func LEQ(a, b PairVec) bool {
//	for k, v := range a.Pairs {
//		val, exists := b.Pairs[k]
//		if exists && val.Counter < v.Counter {
//			return false
//		}
//	}
//	return true
//}

//a and b must be same length
//all elements in a must be in b
func EQ(a, b PairVec) bool {
	for k, v := range a.Pairs {
		val, exists := b.Pairs[k]
		if !exists || val != v {
			return false
		}
	}

	for k, v := range b.Pairs {
		val, exists := a.Pairs[k]
		if !exists || val != v {
			return false
		}
	}
	return true
}

//checks if a < b
//every element of a must be <= to its correspoding one in b
//at least one element in a must be <

func LE(a, b PairVec) bool {
	//all elements of a are contained in b and are <=
	for k, v := range a.Pairs {
		val, exists := b.Pairs[k]
		if !exists || v.Counter > val.Counter {
			return false
		}
	}

	//check that b has at least one elment that is greater
	for k, v := range b.Pairs {
		val, exists := a.Pairs[k]
		if !exists || v.Counter > val.Counter {
			return true
		}
	}

	return false
}

func LEQ(a, b PairVec) bool {
	result := LE(a, b) || EQ(a, b)
	//DPrintf("%v <= %v %v", a, b, result)
	return result
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
		out += v.Id + "_" + strconv.FormatInt(v.Counter, 10) + ", "
	}
	out += ">"
	return out
}
