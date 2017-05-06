package main

import (
	"os"
	"time"
)

func main() {
	config_path := string(os.Args[1])
	cfg, _ := read_config(config_path)
	DPrintf("started")
	file_table := make(map[string]File)
	ev_chan := make(chan Event)
	startstop_chan := make(chan string)
	dirtree := MakeWatcher(cfg.Path)
	crashstate, good := load_persist(cfg.Persistfile)
	if good {
		dirtree = MakeFromSerial(crashstate.Filetree)
		file_table = crashstate.Filetable
	}
	pfunc := func(table map[string]File, tree SerialWatcher) bool {
		return Persist(table, tree, cfg.Persistfile)
	}
	go syncmaker(ev_chan, 10*time.Second, cfg.Hosts)
	go syncreceiver(ev_chan, cfg.Listen, startstop_chan)
	event_loop(ev_chan, startstop_chan, dirtree, file_table, pfunc)
}

func event_loop(events chan Event, startstop chan string, dirtree *Watcher, file_table map[string]File, pfunc persistfunc) {
	var filters []Sfile
	var deleted_filters []Sfile
	for event := range events {
		switch event.Type {
		case EVENT_SYNCTO:
			file_table = syncto(event.Host, event.Username, dirtree, file_table, filters, deleted_filters, pfunc)
			if !pfunc(file_table, dirtree.Serialize()) {
				DPrintf("could not persist after syncto")
			}
		case EVENT_SYNCFROM:
			file_table, filters, deleted_filters = syncfrom(event.Wire, dirtree, file_table, filters, deleted_filters, pfunc)
			if !pfunc(file_table, dirtree.Serialize()) {
				DPrintf("could not persist after syncto")
			}
		}
	}
}
