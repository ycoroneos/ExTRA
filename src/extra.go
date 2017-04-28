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
	file_table := make(map[string]File)
	ev_chan := make(chan Event)
	startstop_chan := make(chan string)
	go syncmaker(ev_chan, 10*time.Second, cfg.Hosts)
	go syncreceiver(ev_chan, cfg.Listen, startstop_chan)
	dirtree := MakeWatcher(cfg.Path)
	event_loop(ev_chan, startstop_chan, dirtree, file_table)
	//for {
	//	}
}

func event_loop(events chan Event, startstop chan string, dirtree *Watcher, file_table map[string]File) {
	var filters []Sfile
	for event := range events {
		switch event.Type {
		case EVENT_SYNCTO:
			StopListening(startstop)
			file_table = syncto(event.Host, dirtree, file_table, filters)
			StartListening(startstop)
		case EVENT_SYNCFROM:
			file_table, filters = syncfrom(event.Wire, dirtree, file_table, filters)
		}
	}
}
