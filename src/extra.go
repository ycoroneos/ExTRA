package main

import (
	"bufio"
	"encoding/gob"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

//ExTRA starts with input from fswatch piped in
func main() {
	config_path := string(os.Args[1])
	cfg, _ := read_config(config_path)
	DPrintf("started")
	file_table = make(map[string]File)
	ev_chan := make(chan Event, 100)
	go event_loop(ev_chan)
	go input_parser(ev_chan)
	go syncmaker(ev_chan, 5*time.Second, cfg.Hosts)
	go syncreceiver(ev_chan, cfg.Listen)
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
		}
	}
}

func event_loop(events chan Event) {
	for event := range events {
		//DPrintf("event : %v", event)
		switch event.Type {
		case EVENT_FSOP:
			do_fsop(event.Data.(FS_OP))
		case EVENT_SYNCTO:
			do_sync(event.Data.(SYNC_OP))
		case EVENT_SYNCFROM:
			do_receive_sync(event.Data.(SyncMsg))
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

//generates sync events
func syncmaker(events chan Event, timeout time.Duration, hosts []string) {
	ticker := time.NewTicker(timeout)
	for _ = range ticker.C {
		msg := SyncMsg{ID, file_table, nil}
		for i := 0; i < len(hosts); i++ {
			events <- Event{EVENT_SYNCTO, SYNC_OP{hosts[i], msg}}
		}
	}
}

//dial the remote and send the stuff
func do_sync(syncinfo SYNC_OP) bool {
	conn, err := net.Dial("tcp", syncinfo.Host)
	if err != nil {
		DPrintf("%v", err)
		return false
	}
	defer conn.Close()
	DPrintf("opened connection to %v", syncinfo.Host)
	enc := gob.NewEncoder(conn) // Will write to network.
	dec := gob.NewDecoder(conn) // Will read from network.
	DPrintf("sending file list")
	err = enc.Encode(syncinfo.Data)
	if err != nil {
		DPrintf("%v", err)
		return false
	}
	var reply SyncReplyMsg
	DPrintf("waiting for reply")
	err = dec.Decode(&reply)
	if err != nil {
		DPrintf("%v", err)
		return false
	}
	for k, v := range reply.Files {
		if v {
			DPrintf("%v wants file %v", syncinfo.Host, k)
		}
	}
	return true
}

func syncreceiver(events chan Event, port int) {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		DPrintf("accepted connection")
		var syncmsg SyncMsg
		dec := gob.NewDecoder(conn) // Will read from network.
		enc := gob.NewEncoder(conn) // Will write to network.
		DPrintf("waiting for file list")
		err = dec.Decode(&syncmsg)
		if err != nil {
			panic(err)
		}
		respchan := make(chan SyncReplyMsg)
		syncmsg.Resp = respchan
		event := Event{EVENT_SYNCFROM, syncmsg}
		DPrintf("loading event")
		events <- event
		DPrintf("waiting for event reply")
		reply := <-respchan
		DPrintf("sending reply")
		err = enc.Encode(&reply)
		if err != nil {
			panic(err)
		}
		conn.Close()
	}
}

//compare version vectors
func do_receive_sync(msg SyncMsg) {
	resp := msg.Resp
	syncmap := make(map[string]bool)
	for k, v := range msg.Files {
		if LEQ(file_table[k].Version, v.Version) {
			syncmap[k] = true
		} else {
			syncmap[k] = false
		}
	}
	resp <- SyncReplyMsg{syncmap}
}
