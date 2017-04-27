package main

import (
	"encoding/gob"
	"net"
	"os"
	"strconv"
	"time"
)

//generates sync events
func syncmaker(events chan Event, timeout time.Duration, hosts []string) {
	ticker := time.NewTicker(timeout)
	for _ = range ticker.C {
		for i := 0; i < len(hosts); i++ {
			//try to sync to to host
			events <- Event{Type: EVENT_SYNCTO, Host: hosts[i], From: ID}
		}
	}
}

//sync an entire path to a remote
func syncto(host string, dirtree *Watcher, state map[string]File, filters []Sfile) map[string]File {
	DPrintf("syncto : try and connect")
	conn, err := net.Dial("tcp", host)
	if !check(err, true) {
		return state
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	DPrintf("syncto : poll dirtree, filters are %v", filters)
	modified, deleted := dirtree.Poll(filters)
	DPrintf("syncto : calculate deltas")
	versions := delta(modified, deleted, state)
	//DPrintf("syncto : send version vectors, %v", versions)
	wants := send_versions(conn, versions)
	for k, v := range wants {
		if v {
			if !dirtree.HasChanged(k) {
				DPrintf("sending file %v", k)
				if send_file(conn, k) {
				}
			}
		}
	}
	DPrintf("syncto : done sending files")
	//send_done(conn)
	return versions
}

//receive an entire path
func syncfrom(from net.Conn, dirtree *Watcher, state map[string]File, filters []Sfile) (map[string]File, []Sfile) {
	defer from.Close()
	DPrintf("syncfrom : poll dirtree, filters are %v", filters)
	modified, deleted := dirtree.Poll(filters)
	DPrintf("syncfrom : calculate deltas, current state is %v", state)
	versions := delta(modified, deleted, state)
	DPrintf("syncfrom : received their versions")
	proposed_versions := receive_versions(from)
	//DPrintf("syncfrom : resolve differences: \n\tus %v\n\tthem %v", versions, proposed_versions)
	want := resolve(versions, proposed_versions)
	//DPrintf("syncfrom : request files %v", want)
	getfiles(from, want)
	DPrintf("syncfrom : done")

	filters = make([]Sfile, 0)
	for {
		newfile := receive_file(from)
		if newfile == "" {
			break
		}
		DPrintf("got file %v", newfile)
		file := proposed_versions[newfile]
		nfo, _ := os.Stat(newfile)
		//fix up time in the new file
		file.Time = nfo.ModTime()
		versions[newfile] = file
		filters = append(filters, Sfile{file.Path, file.Time, false})
	}

	DPrintf("filters after sync %v", filters)
	//	for {
	//		filepath := ""
	//		dec := gob.NewDecoder(from)
	//		if !check(dec.Decode(&filepath), true) {
	//			break
	//		}
	//		if !dirtree.HasChanged(filepath) {
	//			fd, err := os.Create(filepath)
	//		}
	//	}
	//	for _, newfile := range getfiles(want) {
	//		if !dirtree.HasChanged(newfile) {
	//			write(newfile)
	//			dirtree.Addfilter(newfile)
	//		}
	//	}
	//wait_done(from)
	return versions, filters
}

//func challenge_response(conn net.Conn, challenger bool) bool {
//	enc := gob.NewEncoder(conn)
//	dec := gob.NewDecoder(conn)
//	if challenger {
//		ch := "1+1"
//		enc.Encode(ch)
//		dec.Decode(&ch)
//		return ch == "2"
//	} else {
//		ch := ""
//		dec.Decode(&ch)
//		if ch == "1+1" {
//			ch = "2"
//			enc.Encode(ch)
//			return true
//		}
//		return false
//	}
//}

func resolve(us, theirs map[string]File) map[string]bool {
	syncmap := make(map[string]bool)
	for k, v := range theirs {
		if LE(us[k].Version, v.Version) {
			syncmap[k] = true
		} else {
			syncmap[k] = false
		}
	}
	return syncmap
}

func send_versions(conn net.Conn, versions map[string]File) map[string]bool {
	wants := make(map[string]bool)
	enc := gob.NewEncoder(conn) // Will write to network.
	if !check(enc.Encode(versions), true) {
		return make(map[string]bool)
	}
	dec := gob.NewDecoder(conn)
	if !check(dec.Decode(&wants), true) {
		return make(map[string]bool)
	}
	return wants
}

func receive_versions(conn net.Conn) map[string]File {
	versions := make(map[string]File)
	dec := gob.NewDecoder(conn) // Will read from network.
	if !check(dec.Decode(&versions), true) {
		return make(map[string]File)
	}
	return versions
}

func getfiles(conn net.Conn, want map[string]bool) {
	enc := gob.NewEncoder(conn)
	if !check(enc.Encode(want), true) {
	}
	return
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
		//DPrintf("accepted connection")
		events <- Event{Type: EVENT_SYNCFROM, Wire: conn}
	}
}

//compare version vectors
//func do_receive_sync(msg Event) {
//
//	defer msg.Wire.Close()
//	their_table := make(map[string]File)
//	dec := gob.NewDecoder(msg.Wire) // Will read from network.
//	enc := gob.NewEncoder(msg.Wire) // Will write to network.
//	DPrintf("decoding their table")
//	check(dec.Decode(&their_table), true)
//
//	//resp := msg.Resp
//	syncmap := make(map[string]bool)
//	DPrintf("looping over their table")
//	for k, v := range their_table {
//		if LEQ(file_table[k].Version, v.Version) {
//			syncmap[k] = true
//		} else {
//			syncmap[k] = false
//		}
//	}
//	DPrintf("sending our response")
//	check(enc.Encode(SyncReplyMsg{syncmap}), true)
//	//close(resp)
//}
