package main

import (
	"encoding/gob"
	"net"
	"strconv"
	"time"
)

//generates sync events
func syncmaker(events chan Event, timeout time.Duration, hosts []string) {
	ticker := time.NewTicker(timeout)
	for _ = range ticker.C {
		for i := 0; i < len(hosts); i++ {
			//try to sync to to host
			events <- Event{Type: EVENT_SYNCTO, Host: hosts[i], From: ID, Files: file_table}
		}
	}
}

//dial the remote and send the stuff
func do_sync(syncinfo Event) bool {
	conn, err := net.Dial("tcp", syncinfo.Host)
	if !check(err, true) {
		return false
	}

	defer conn.Close()
	DPrintf("opened connection to %v", syncinfo.Host)
	enc := gob.NewEncoder(conn) // Will write to network.
	dec := gob.NewDecoder(conn) // Will read from network.
	DPrintf("sending file list")
	if !check(enc.Encode(syncinfo.Files), true) {
		return false
	}
	var reply SyncReplyMsg
	DPrintf("waiting for reply")
	if !check(dec.Decode(&reply), true) {
		return false
	}
	DPrintf("checking their reply")
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
		events <- Event{Type: EVENT_SYNCFROM, Wire: conn}
	}
}

//compare version vectors
func do_receive_sync(msg Event) {

	defer msg.Wire.Close()
	their_table := make(map[string]File)
	dec := gob.NewDecoder(msg.Wire) // Will read from network.
	enc := gob.NewEncoder(msg.Wire) // Will write to network.
	DPrintf("decoding their table")
	check(dec.Decode(&their_table), true)

	//resp := msg.Resp
	syncmap := make(map[string]bool)
	DPrintf("looping over their table")
	for k, v := range their_table {
		if LEQ(file_table[k].Version, v.Version) {
			syncmap[k] = true
		} else {
			syncmap[k] = false
		}
	}
	DPrintf("sending our response")
	check(enc.Encode(SyncReplyMsg{syncmap}), true)
	//close(resp)
}
