package main

import (
	"os"
	"time"
)

//ExTRA starts with input from fswatch piped in
func main() {
	config_path := string(os.Args[1])
	cfg, _ := read_config(config_path)
	DPrintf("started")
	file_table = make(map[string]File)
	ev_chan := make(chan Event)
	//go event_loop(ev_chan)
	//go fs_monitor(ev_chan, 1, cfg.Path)
	//go input_parser(ev_chan)
	go syncmaker(ev_chan, 5*time.Second, cfg.Hosts)
	go syncreceiver(ev_chan, cfg.Listen)
	dirtree := MakeWatcher(cfg.Path)
	event_loop(ev_chan, dirtree)
	//for {
	//	}
}

func event_loop(events chan Event, dirtree *Watcher) {
	for event := range events {
		switch event.Type {
		case EVENT_SYNCTO:
			file_table = syncto(event.Host, dirtree, file_table)
		case EVENT_SYNCFROM:
			file_table = syncfrom(event.Wire, dirtree, file_table)
		}
	}
}
