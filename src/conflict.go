package main

import (
	"io"
	"os"
	"time"
)

const (
	TAKE_THEIRS = 0
	KEEP_OURS   = 1
	MERGE       = 2
	NONE        = 3
)

type ConflictResolution struct {
	Filename   string
	Resolution int
}

type conflictF func(string) (bool, ConflictResolution)

//these are some dummy conflict resolvers for illustrating the behavior
//of the file synchronizer

func resolution_theirs(filename string) (bool, ConflictResolution) {
	DPrintf("resolve conflict %v by taking theirs", filename)
	return true, ConflictResolution{filename, TAKE_THEIRS}
}

func resolution_ours(filename string) (bool, ConflictResolution) {
	DPrintf("resolve conflict %v by keeping ours", filename)
	return false, ConflictResolution{filename, KEEP_OURS}
}

func resolution_merge(filename string) (bool, ConflictResolution) {
	DPrintf("resolve conflict %v by merging theirs", filename)
	return true, ConflictResolution{filename, MERGE}
}

func resolution_complain(filename string) (bool, ConflictResolution) {
	DPrintf("resolve conflict %v by COMPLAINING", filename)
	return false, ConflictResolution{filename, NONE}
}

//fake merge function
//just appends a MERGE message to the top
func merge(filename, newfilename string) bool {

	merged := false

	//defer the move from new to original
	defer func() {
		if merged {
			os.Rename(newfilename, filename)
		}
	}()

	//open the original
	fd, err := os.Open(filename)
	if !check(err, true) {
		return false
	}
	defer fd.Close()
	stat, err := fd.Stat()
	if !check(err, true) {
		return false
	}

	//open the new one
	fd2, err := os.OpenFile(newfilename, os.O_RDWR|os.O_APPEND, stat.Mode()&os.ModePerm)
	if !check(err, true) {
		return false
	}
	defer fd2.Close()
	//	stat2, err := fd2.Stat()
	//	if !check(err, true) {
	//		return false
	//	}

	fd2.Write([]byte("+++++MERGE+++++"))
	_, err = io.Copy(fd2, fd)
	if !check(err, true) {
		return false
	}
	merged = true
	return merged

	//	//make the copy
	//	perms := int32(stat.Mode() & os.ModePerm)
	//	nfile, err := os.OpenFile(filename+"_merging", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(perms))
	//	if !check(err, true) {
	//		return false
	//	}
	//	defer nfile.Close()
	//	nfile.Write([]byte("MERGED\n"))
	//	_, err = io.Copy(nfile, fd)
	//	if !check(err, true) {
	//		return false
	//	}
	//	merged = true
	//	return merged
}

func getmodtime(filename string) (time.Time, bool) {
	fd, err := os.Open(filename)
	if !check(err, true) {
		return time.Time{}, false
	}
	defer fd.Close()
	stat, err := fd.Stat()
	if !check(err, true) {
		return time.Time{}, false
	}
	return stat.ModTime(), true
}
