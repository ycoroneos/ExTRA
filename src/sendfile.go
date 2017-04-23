package main

import (
	"encoding/gob"
	"net"
	"os"
)

type FileChunk struct {
	Path   string
	Offset int64
	Data   []byte
}

func send_done(conn net.Conn) {
	done := "done"
	enc := gob.NewEncoder(conn)
	enc.Encode(done)
}

func wait_done(from net.Conn) {
	done := ""
	dec := gob.NewDecoder(from)
	for {
		dec.Decode(&done)
		if done == "done" {
			return
		}
	}
}

func send_file(conn net.Conn, file string) bool {
	info, err := os.Stat(file)
	size := info.Size()
	ten_mb := int64(10485760)
	//chunks := (size + ten_mb - 1) / ten_mb
	fd, err := os.Open(file) // For read access.
	check(err, false)
	defer fd.Close()            // make sure to close the file even if we panic.
	enc := gob.NewEncoder(conn) // Will write to network.
	for offset := int64(0); offset < size; offset += ten_mb {
		data := make([]byte, ten_mb)
		fd.ReadAt(data, offset)
		enc.Encode(FileChunk{file, offset, data})
	}
	enc.Encode(FileChunk{file, -1, nil})
	return true
}

func receive_file(conn net.Conn) string {
	dec := gob.NewDecoder(conn)
	var nextchunk FileChunk
	dec.Decode(&nextchunk)
	if nextchunk.Offset != 0 {
		panic("out of order reception with TCP?")
	}
	fd, err := os.Create(nextchunk.Path)
	check(err, false)
	for {
		fd.WriteAt(nextchunk.Data, nextchunk.Offset)
		if !check(dec.Decode(&nextchunk), true) {
			return ""
		}
		if nextchunk.Offset == -1 {
			return nextchunk.Path
		}
	}
}