package main

import (
	"time"
)

//flags for fswatch
const (
	Platformspecific = 0x1 << 0
	Created          = 0x1 << 1
	Updated          = 0x1 << 2
	Removed          = 0x1 << 3
	Renamed          = 0x1 << 4
	Isfile           = 0x1 << 9
	Isdir            = 0x1 << 10
)

//Can be used to either represent version or sync pair
type TraPair struct {
	Id   int
	Time time.Time
}

//either version vector or time synchronization vector
type TraVec struct {
	pairs []TraPair
}

//each existing file in tra gets its own data structure
type TraFile struct {
	Name            string
	Created         time.Time
	Modification    TraVec
	Synchronization TraVec
}

//every Id is guaranteed to only appear once
//find it and check if we should change it
//also preserve ordering of times
func (t TraVec) Append(p TraPair) TraVec {
	for i := 0; i < len(t.pairs)-1; i++ {
		if t.pairs[i].Id == p.Id && t.pairs[i].Time < p.Time {
			temp := append(t.pairs[:i], t.pairs[i+1:]...)
			temp = append(temp, p)
			return TraVec{temp}
		}
	}
	if t.pairs[len(t.pairs)].Id == p.Id && t.pairs[len(t.pairs)].Time < p.Time {
		t.pairs[len(t.pairs)] = p
		return t
	}
	return TraVec{append(t.pairs, p)}
}

func (A TraVec) LEQ(B TraVec) bool {
}

func MakeTraVec(pairs []TraPair) TraVec {
	result := make(TraVec)
	for _, pair := range pairs {
		result = result.Append(pair)
	}
	return result
}

//Figure 9
func ResolveFile(A, B TraFile) string {
	if A.Modification.LEQ(B.Synchronization) {
		return "nothing"
	} else if B.Modification.LEQ(A.Synchronization) {
		return "copy A to B"
	} else {
		return "conflict"
	}
}
