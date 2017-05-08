package main

import (
	"encoding/gob"
	"net"
	"os"
	"strconv"
	"time"
)

//generates sync events
func syncmaker(events chan Event, timeout time.Duration, hosts map[string]string) {
	ticker := time.NewTicker(timeout)
	for _ = range ticker.C {
		for k, v := range hosts {
			events <- Event{Type: EVENT_SYNCTO, Host: v, Username: k, From: ID}
		}
		//	for i := 0; i < len(hosts); i++ {
		//		//try to sync to to host
		//		events <- Event{Type: EVENT_SYNCTO, Host: hosts[i], From: ID}
		//	}
	}
}

//sync an entire path to a remote
func syncto(host string, username string, dirtree *Watcher, state map[string]File, filters, deleted_filters []Sfile, persist persistfunc) map[string]File {
	DPrintf("syncto : try and connect")
	conn, err := net.Dial("tcp", host)
	if !check(err, true) {
		return state
	}
	defer conn.Close()
	//DPrintf("syncto : poll dirtree, filters are %v", filters)
	DPrintf("syncto : poll dirtree")
	modified, deleted := dirtree.Poll(filters, deleted_filters)
	//DPrintf("deleted : %v", deleted)
	DPrintf("syncto : calculate deltas")
	versions := delta(modified, deleted, state)
	//DPrintf("syncto : send version vectors, %v", versions)

	//chunks := makechunks(modified)
	//DPrintf("time-vectors: %v", sync_version)
	wants := send_versions(conn, versions)
	count := 0
	for k, v := range wants {
		if v.Send {
			//send and sync
			DPrintf("sending file %v     -> %v%%", k, float32(count)/float32(len(wants)))
			if send_file_chunks(conn, k, Rollhash(k)) {
				//if send_file_chunks(conn, k, v.Chunks) {
				//update the file's synchronization vector on success
				file := versions[k]
				versions[k] = file.SyncModify().BackSync(username)
			} else {
				DPrintf("the file did not send")
				break
			}
		} else if v.Sync {
			//just sync but dont send
			file := versions[k]
			versions[k] = file.SyncModify().BackSync(username)
		}
		count += 1
		//		if v {
		//			DPrintf("sending file %v", k)
		//			if send_file(conn, k) {
		//				//update the file's synchronization vector on success
		//				file := versions[k]
		//				versions[k] = file.SyncModify()
		//			} else {
		//				DPrintf("the file did not send")
		//				break
		//			}
		//			//	}
		//		}
	}
	DPrintf("syncto : done sending files")
	//send_done(conn)
	return versions
}

//receive an entire path
func syncfrom(from net.Conn, dirtree *Watcher, state map[string]File, filters, deleted_filters []Sfile, persist persistfunc) (map[string]File, []Sfile, []Sfile) {
	defer from.Close()
	//DPrintf("syncfrom : poll dirtree, filters are %v", filters)
	DPrintf("syncfrom : poll dirtree")
	modified, deleted := dirtree.Poll(filters, deleted_filters)
	//DPrintf("syncfrom : calculate deltas, current state is %v", state)
	DPrintf("syncfrom : calculate deltas")
	versions := delta(modified, deleted, state)
	DPrintf("syncfrom : received their versions")
	proposed_versions, them_id := receive_versions(from)
	//	DPrintf("syncfrom : resolve differences: \n\tus %v\n\tthem %v", versions, proposed_versions)
	//want, resolutions := resolve_tvpair_with_delete(proposed_versions, versions, resolution_ours)
	files_wanted, resolutions := resolve_tvpair_with_delete(proposed_versions, versions, resolution_complain)
	//DPrintf("syncfrom : request files %v", want)
	getfiles(from, files_wanted)
	DPrintf("syncfrom : done")

	filters = make([]Sfile, 0)
	deleted_filters = make([]Sfile, 0)
	//receive new files and merged files
	for {
		newfile, good := receive_file_chunks(from)
		if !good {
			if newfile != "" {
				os.Remove(newfile)
			}
			break
		}
		DPrintf("got file %v", newfile)
		file := proposed_versions[newfile]
		nfo, err := os.Stat(newfile)
		if err == nil {
			//fix up time in the new file
			file.Time = nfo.ModTime()

			//update the synchronization vector in the file

			//stick it in the map
			versions[newfile] = file.BackSync(them_id).SyncModify()
			filters = append(filters, Sfile{file.Path, file.Time, false})
		} else if os.IsNotExist(err) {
			//fix up time in the new file
			//file.Time = nfo.ModTime()

			//update the synchronization vector in the file

			//stick it in the map
			oldtime := versions[newfile].Time
			versions[newfile] = file.BackSync(them_id).SyncModify()
			deleted_filters = append(deleted_filters, Sfile{file.Path, oldtime, false})
			DPrintf("deleted a file %v", file.Path)
		} else {
			check(err, false)
		}
	}

	for _, c := range resolutions {
		if c.Resolution == MERGE {
			//update version vector for merges
			merged := merge(c.Filename)
			if !merged {
				//dont merge if something went wrong
				continue
			}
			modTime, good := getmodtime(c.Filename)
			if !good {
				continue
			}
			file := versions[c.Filename]
			file.Time = modTime
			file = file.Modify()
			versions[c.Filename] = file
			filters = append(filters, Sfile{file.Path, file.Time, false})
		} else if c.Resolution == KEEP_OURS {
			//update sync vector for conflicts we keep
			file := versions[c.Filename]
			versions[c.Filename] = file.BackSync(them_id).SyncModify()
		}

	}

	//Cleanup(".")
	return versions, filters, deleted_filters
}

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

func resolve_tvpair(them, us map[string]File) map[string]bool {
	output := make(map[string]bool)
	for k, v := range them {
		if LEQ(v.Version, us[k].Sync) {
			output[k] = false
		} else if LEQ(us[k].Version, v.Sync) {
			output[k] = true
		} else {
			panic("conflict detected!")
		}
	}
	return output
}

func resolve_tvpair_with_delete(them, us map[string]File, cf conflictF) (map[string]Receive_data, []ConflictResolution) {
	output := make(map[string]Receive_data)
	conflict_decisions := make([]ConflictResolution, 0)
	//chunks_wanted := make(map[string][]FileChunk, 0)
	//chunk_recipes := make(map[string][]ChunkDelta, 0)
	for k, v := range them {
		_, err := os.Stat(k)
		_, exists := us[k]
		if exists && err == nil {
			//we know of this file, and it definitely was not deleted
			//			DPrintf("them version: %v", v.Version)
			//			DPrintf("them creation: %v", v.Creation)
			//			DPrintf("them sync: %v", v.Sync)
			//			DPrintf("us version: %v", us[k].Version)
			//			DPrintf("us creation: %v", us[k].Creation)
			//			DPrintf("us sync: %v", us[k].Sync)
			if LEQ(v.Version, us[k].Sync) {
				//output[k] = false
				output[k] = Receive_data{false, false}
			} else if LEQ(us[k].Version, v.Sync) {
				//output[k] = true
				//our_chunks := Rollhash(v.Path)
				//	needed_chunks, chunk_delta := ChompAlgo(their_chunks[v.Path], our_chunks)
				//		chunks_wanted[v.Path] = needed_chunks
				//		chunk_recipes[v.Path] = chunk_delta
				output[k] = Receive_data{true, true}
			} else {
				DPrintf("----%v-------", v.Path)
				DPrintf("them version: %v", v.Version)
				DPrintf("them creation: %v", v.Creation)
				DPrintf("them sync: %v", v.Sync)
				DPrintf("us version: %v", us[k].Version)
				DPrintf("us creation: %v", us[k].Creation)
				DPrintf("us sync: %v", us[k].Sync)
				DPrintf("CONFLICT : %s", v.Path)
				take, decision := cf(v.Path)
				//our_chunks := Rollhash(v.Path)
				//needed_chunks, chunk_delta := ChompAlgo(their_chunks[v.Path], our_chunks)
				//chunks_wanted[v.Path] = needed_chunks
				//chunk_recipes[v.Path] = chunk_delta
				output[k] = Receive_data{take, true}
				conflict_decisions = append(conflict_decisions, decision)
				//panic("conflict detected!")
			}
		} else {
			//we have never seen this file before, or it was deleted
			//			DPrintf("them version: %v", v.Version)
			//			DPrintf("them creation: %v", v.Creation)
			//			DPrintf("us sync: %v", us[k].Sync)
			if LEQ(v.Version, us[k].Sync) {
				//output[k] = false
				output[k] = Receive_data{false, false}
			} else if !LEQ(v.Creation, us[k].Sync) {
				//output[k] = true
				output[k] = Receive_data{true, true}
				//chunks_wanted[v.Path] = their_chunks[v.Path]
			} else {
				DPrintf("them version: %v", v.Version)
				DPrintf("them creation: %v", v.Creation)
				DPrintf("us sync: %v", us[k].Sync)
				DPrintf("CONFLICT : %s", v.Path)
				//output[k] = false
				take, decision := cf(v.Path)
				//output[k] = take
				output[k] = Receive_data{take, true}
				//chunks_wanted[v.Path] = their_chunks[v.Path]
				conflict_decisions = append(conflict_decisions, decision)
				//panic("conflict detected")
			}
		}
	}
	return output, conflict_decisions
}

type Send_data struct {
	From     string
	Versions map[string]File
	//Chunks   map[string][]FileChunk
}

type Receive_data struct {
	Send bool
	Sync bool
}

func send_versions(conn net.Conn, versions map[string]File) map[string]Receive_data {
	wants := make(map[string]Receive_data)
	enc := gob.NewEncoder(conn) // Will write to network.
	send := Send_data{ID, versions}
	if !check(enc.Encode(send), true) {
		return make(map[string]Receive_data)
	}
	dec := gob.NewDecoder(conn)
	if !check(dec.Decode(&wants), true) {
		return make(map[string]Receive_data)
	}
	return wants
}

func receive_versions(conn net.Conn) (map[string]File, string) {
	//versions := make(map[string]File)
	receive := Send_data{"", make(map[string]File)}
	dec := gob.NewDecoder(conn) // Will read from network.
	if !check(dec.Decode(&receive), true) {
		return make(map[string]File), ""
	}
	return receive.Versions, receive.From
}

func getfiles(conn net.Conn, want map[string]Receive_data) {
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

func syncreceiver(events chan Event, port int, startstop chan string) {

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		panic(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		select {
		case events <- Event{Type: EVENT_SYNCFROM, Wire: conn}:
		default:
			conn.Close()
		}
	}
}

func StopListening(cmd chan string) {
	cmd <- "stop"
	<-cmd
	return
}

func StartListening(cmd chan string) {
	cmd <- "start"
	<-cmd
	return
}
