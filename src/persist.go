package main

import (
	//"../golang-set"
	"encoding/gob"
	"os"
)

type Persister struct {
	Filetable map[string]File
	Filetree  SerialWatcher
}

type persistfunc func(map[string]File, SerialWatcher) bool

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

func Persist(table map[string]File, tree SerialWatcher, persistfile string) bool {
	fd, err := os.OpenFile(persistfile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if !check(err, true) {
		return false
	}
	defer fd.Close()
	enc := gob.NewEncoder(fd) // Will write to file
	p := Persister{table, tree}
	if !check(enc.Encode(p), true) {
		return false
	}
	return true
}
