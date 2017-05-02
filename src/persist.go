package main

import (
	//"../golang-set"
	"encoding/gob"
	"os"
)

type Persister struct {
	Filetable       map[string]File
	Filetree        SerialWatcher
	Filters         []Sfile
	Deleted_filters []Sfile
}

type persistfunc func(map[string]File, SerialWatcher, []Sfile, []Sfile) bool

var persister Persister

func load_persist(file string) (Persister, bool) {

	gob.Register(Sfile{})
	fd, err := os.Open(file)
	if !check(err, true) {
		return Persister{}, false
	}
	defer fd.Close()
	dec := gob.NewDecoder(fd)
	var persist Persister
	if !check(dec.Decode(&persist), true) {
		return Persister{}, false
	}
	return persist, true
}

func Persist(table map[string]File, tree SerialWatcher, filter, deleted_filters []Sfile, persistfile string) bool {
	fd, err := os.OpenFile(persistfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if !check(err, true) {
		return false
	}
	defer fd.Close()
	enc := gob.NewEncoder(fd) // Will write to file
	p := Persister{table, tree, filter, deleted_filters}
	if !check(enc.Encode(p), true) {
		return false
	}
	return true
}
