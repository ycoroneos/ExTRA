package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	config_path := string(os.Args[1])
	cfg, _ := read_config(config_path)
	DPrintf("started")
	file_table := make(map[string]File)
	ev_chan := make(chan Event)
	filters := make([]Sfile, 0)
	deleted_filters := make([]Sfile, 0)
	startstop_chan := make(chan string)
	dirtree := MakeWatcher(cfg.Path)
	crashstate, good := load_persist(cfg.Persistfile)
	if good {
		dirtree = MakeFromSerial(crashstate.Filetree)
		file_table = crashstate.Filetable

		filters = crashstate.Filters
		deleted_filters = crashstate.Deleted_filters

		//poll to detect if user has made changes
		modified, deleted := dirtree.Poll(crashstate.Filters, crashstate.Deleted_filters)
		if len(modified) > 0 || len(deleted) > 0 {
			fmt.Printf("contents changed since crash:\n")
			fmt.Printf("modified:\n")
			for _, k := range modified {
				fmt.Printf("\t %v\n", k)
			}
			fmt.Printf("deleted:\n")
			for _, k := range deleted {
				fmt.Printf("\t %v\n", k)
			}
			fmt.Printf("accept modifications?")
			file_table = delta(modified, deleted, file_table)

			//panic("changed contents since crash")
		}
	}
	pfunc := func(table map[string]File, tree SerialWatcher, filters, deleted_filters []Sfile) bool {
		return Persist(table, tree, filters, deleted_filters, cfg.Persistfile)
	}
	go syncmaker(ev_chan, 10*time.Second, cfg.Hosts)
	go syncreceiver(ev_chan, cfg.Listen, startstop_chan)
	event_loop(ev_chan, startstop_chan, dirtree, file_table, filters, deleted_filters, pfunc)
}

func event_loop(events chan Event, startstop chan string, dirtree *Watcher, file_table map[string]File, filters, deleted_filters []Sfile, pfunc persistfunc) {
	//	var filters []Sfile
	//	var deleted_filters []Sfile
	for event := range events {
		switch event.Type {
		case EVENT_SYNCTO:
			file_table = syncto(event.Host, event.Username, dirtree, file_table, filters, deleted_filters, pfunc)
			if !pfunc(file_table, dirtree.Serialize()) {
				DPrintf("could not persist after syncto")
			}
		case EVENT_SYNCFROM:
			file_table, filters, deleted_filters = syncfrom(event.Wire, dirtree, file_table, filters, deleted_filters, pfunc)
			if !pfunc(file_table, dirtree.Serialize(), filters, deleted_filters) {
				DPrintf("could not persist after syncto")
			}
		}
	}
}
