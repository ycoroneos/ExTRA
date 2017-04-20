package main

import (
	"bufio"
	"os"
	"strings"
	"time"
)

var ID string

//ExTRA starts with input from fswatch piped in
func main() {
	ID = string(os.Args[1])
	file_table = make(map[string]File)
	ev_chan := make(chan Event, 100)
	go event_loop(ev_chan)
	go input_parser(ev_chan)
	go syncmaker(ev_chan, 5*time.Second)
	for {
	}
}

//takes input from fswatch and creates FS_OP events for the event loop
func input_parser(events chan Event) {
	reader := bufio.NewReader(os.Stdin)
	for {
		command, _ := reader.ReadString('\n')
		fields := strings.Fields(command)
		switch fields[1] {
		case "Removed":
			events <- Event{EVENT_FSOP, FS_OP{fields[0], FSOP_DELETE}}
		case "Created":
			events <- Event{EVENT_FSOP, FS_OP{fields[0], FSOP_MODIFY}}
		case "Updated":
			events <- Event{EVENT_FSOP, FS_OP{fields[0], FSOP_MODIFY}}
			//default:
			//		DPrintf("input_parser doesnt know about %v", fields)
			//		panic("unhandled case")
		}
	}
}

func event_loop(events chan Event) {
	for event := range events {
		//DPrintf("event : %v", event)
		switch event.Type {
		case EVENT_FSOP:
			do_fsop(event.Data.(FS_OP))
		case EVENT_SYNC:
			DPrintf("should sync now")
		}
	}
}

func do_fsop(op FS_OP) {
	val, exists := file_table[op.File]
	switch op.Action {
	case FSOP_DELETE:
		//we better have it
		if !exists {
			panic("trying to delete a file we dont have")
		} else {
			delete(file_table, op.File)
		}

	case FSOP_MODIFY:
		//we track this file now
		if !exists {
			newfile := File{op.File, MakePairVec([]Pair{Pair{ID, 1}}), MakePairVec([]Pair{Pair{ID, 1}})}
			file_table[op.File] = newfile
		} else {
			newfile := val.Modify()
			file_table[op.File] = newfile
		}
		DPrintf("%s", file_table[op.File].Show())
	default:
		panic("unimplemented fsop")
	}
}

func syncmaker(events chan Event, timeout time.Duration) {
	ticker := time.NewTicker(timeout)
	for _ = range ticker.C {
		events <- Event{EVENT_SYNC, SYNC_OP{}}
	}
}
