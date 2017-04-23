package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Cfg struct {
	Id      string
	Path    string
	Logfile string
	Listen  int
	Hosts   []string
	Cmd     int
}

var ID string
var Log *log.Logger

//read the json config
func read_config(path string) (Cfg, bool) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	out := Cfg{}
	err = json.Unmarshal(dat, &out)
	if err != nil {
		panic(err)
	}
	ID = out.Id
	f, err := os.Create(out.Logfile)
	if err != nil {
		panic(err)
	}
	Log = log.New(f, "", log.Lshortfile)
	return out, true
}
