package main

import (
	"bytes"
	"encoding/binary"
	"net"
	"os"
	"path/filepath"
)

func send_file(conn net.Conn, file string) bool {
	if len(file) > 1024 {
		return false
	}
	fd, err := os.Open(file)
	if !check(err, true) {
		if os.IsNotExist(err) {
			return send_delete(conn, file)
		} else {
			return false
		}
	}
	stat, err := fd.Stat()
	if !check(err, true) {
		return false
	}
	defer fd.Close()

	//send name
	namebuf := []byte(file)
	namebuf = append(namebuf, make([]byte, 1024-len(namebuf))...)
	n, err := conn.Write(namebuf)
	if n != len(namebuf) {
		return false
	}
	if !check(err, true) {
		return false
	}

	//send size
	err = binary.Write(conn, binary.LittleEndian, stat.Size())
	if !check(err, true) {
		return false
	}

	//send permission bits
	perms := int32(stat.Mode() & os.ModePerm)
	err = binary.Write(conn, binary.LittleEndian, perms)
	if !check(err, true) {
		return false
	}

	//send the file
	for {
		next := int64(0)
		binary.Read(conn, binary.LittleEndian, &next)
		if next == -1 {
			break
		}
		buf := make([]byte, 4096)
		amt, _ := fd.ReadAt(buf, next)
		_, err = conn.Write(buf[0:amt])
		if !check(err, true) {
			return false
		}
	}

	return true
}

func send_file_chunks(conn net.Conn, file string, chunks []FileChunk) bool {
	if len(file) > 1024 {
		return false
	}
	fd, err := os.Open(file)
	if !check(err, true) {
		if os.IsNotExist(err) {
			return send_delete(conn, file)
		} else {
			return false
		}
	}
	stat, err := fd.Stat()
	if !check(err, true) {
		return false
	}
	defer fd.Close()

	//send name
	DPrintf("send name")
	namebuf := []byte(file)
	namebuf = append(namebuf, make([]byte, 1024-len(namebuf))...)
	n, err := conn.Write(namebuf)
	if n != len(namebuf) {
		return false
	}
	if !check(err, true) {
		return false
	}

	//send size
	DPrintf("send size")
	err = binary.Write(conn, binary.LittleEndian, stat.Size())
	if !check(err, true) {
		return false
	}

	//send permission bits
	DPrintf("send permissions")
	perms := int32(stat.Mode() & os.ModePerm)
	err = binary.Write(conn, binary.LittleEndian, perms)
	if !check(err, true) {
		return false
	}

	//send number of chunks
	DPrintf("send nchunks")
	err = binary.Write(conn, binary.LittleEndian, int64(len(chunks)))
	if !check(err, true) {
		return false
	}

	//send chunks
	DPrintf("send chunks")
	for i := 0; i < len(chunks); i++ {
		err = binary.Write(conn, binary.LittleEndian, chunks[i])
		if !check(err, true) {
			return false
		}
	}

	//send the file
	DPrintf("send file")
	for {
		next := int64(0)
		binary.Read(conn, binary.LittleEndian, &next)
		if next == -1 {
			break
		}
		if next == -2 {
			return false
		}
		buf := make([]byte, 4096)
		_, err = fd.Stat()
		if !check(err, true) {
			return false
		}
		amt, err := fd.ReadAt(buf, next)
		_, err = conn.Write(buf[0:amt])
		if !check(err, true) {
			return false
		}
	}

	return true
}

func send_delete(conn net.Conn, file string) bool {
	//send name
	namebuf := []byte(file)
	namebuf = append(namebuf, make([]byte, 1024-len(namebuf))...)
	n, err := conn.Write(namebuf)
	if n != len(namebuf) {
		return false
	}
	if !check(err, true) {
		return false
	}

	//send size
	err = binary.Write(conn, binary.LittleEndian, int64(-1))
	if !check(err, true) {
		return false
	}
	return true
}

func receive_file(conn net.Conn) (string, bool) {

	//get the name
	DPrintf("get new file name")
	filenamebuf := make([]byte, 1024)
	n, err := conn.Read(filenamebuf)
	DPrintf("read filename %s, len %d", string(filenamebuf), n)
	if n != len(filenamebuf) {
		return "", false
	}
	if !check(err, true) {
		return "", false
	}
	filename := string(bytes.Trim(filenamebuf, "\x00"))

	//get the size
	DPrintf("get new file size")
	filesz := int64(0)
	err = binary.Read(conn, binary.LittleEndian, &filesz)
	DPrintf("file size %v", filesz)
	if filesz == -1 {
		return delete_file(filename)
	}
	if !check(err, true) {
		return "", false
	}

	//get the permission bits
	perms := int32(0)
	err = binary.Read(conn, binary.LittleEndian, &perms)
	DPrintf("permissions %v", perms)
	if !check(err, true) {
		return "", false
	}

	//make the directory path
	dir := filepath.Dir(filename)
	DPrintf("make dir %s", dir)
	if !check(os.MkdirAll(dir, 0777), true) {
		return "", false
	}

	//make the file
	DPrintf("open file %s, len s %d", filename, len(filename))
	//fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(perms))
	if !check(err, true) {
		return "", false
	}
	defer fd.Close()

	//receive the file
	buf := make([]byte, 4096)
	for want := int64(0); want < filesz; {
		binary.Write(conn, binary.LittleEndian, want)
		n, err = conn.Read(buf)
		if !check(err, true) {
			return filename, false
		}
		want += int64(n)
		amt, err := fd.Write(buf[0:n])
		if amt < n {
			DPrintf("amount written is not enough %v vs %v", amt, n)
			return filename, false
		}
		if !check(err, true) {
			return filename, false
		}
	}
	done := int64(-1)
	binary.Write(conn, binary.LittleEndian, done)
	return filename, true
}

func receive_file_chunks(conn net.Conn) (string, bool) {

	success := false

	//get the name
	DPrintf("get the filename")
	filenamebuf := make([]byte, 1024)
	n, err := conn.Read(filenamebuf)
	DPrintf("read filename %s, len %d", string(filenamebuf), n)
	if n != len(filenamebuf) {
		return "", false
	}
	if !check(err, true) {
		return "", false
	}
	filename := string(bytes.Trim(filenamebuf, "\x00"))

	//get the size
	filesz := int64(0)
	err = binary.Read(conn, binary.LittleEndian, &filesz)
	DPrintf("file size %v", filesz)
	if filesz == -1 {
		return delete_file(filename)
	}
	if !check(err, true) {
		return "", false
	}

	//get the permission bits
	perms := int32(0)
	err = binary.Read(conn, binary.LittleEndian, &perms)
	DPrintf("permissions %v", perms)
	if !check(err, true) {
		return "", false
	}

	//get num chunks
	chunksz := int64(0)
	err = binary.Read(conn, binary.LittleEndian, &chunksz)
	if !check(err, true) {
		return "", false
	}

	//get the chunks
	chunks := make([]FileChunk, chunksz)
	for i := int64(0); i < chunksz; i++ {
		err = binary.Read(conn, binary.LittleEndian, &chunks[i])
		if !check(err, true) {
			return "", false
		}
	}
	//DPrintf("received chunks %v", chunks)

	//make the directory path
	dir := filepath.Dir(filename)
	DPrintf("make dir %s", dir)
	if !check(os.MkdirAll(dir, 0777), true) {
		return "", false
	}

	//make the file
	DPrintf("open file %s, len s %d", filename, len(filename))
	//fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	fd, err := os.OpenFile(filename+"~", os.O_RDWR|os.O_CREATE|os.O_TRUNC, os.FileMode(perms))
	if !check(err, true) {
		return "", false
	}

	defer func() {
		if success {
			DPrintf("Rename %v to %v", filename+"~", filename)
			os.Rename(filename+"~", filename)
		}
	}()

	defer fd.Close()

	chunks_wanted := chunks
	recipe := make([]ChunkDelta, 0)
	if filesz > (1024 * 10 * 10) {
		DPrintf("compute rolling hash")
		hash := FastRollhash(filename)
		//DPrintf("done rolling hash, start chompalgo with %v %v", chunks, hash)
		chunks_wanted, recipe = Diff(chunks, hash)
		DPrintf("done chompalgo")
	}

	//DPrintf("chunks we want: %v\nchunks we have %v", chunks_wanted, recipe)
	//move around the chunks we have
	if len(recipe) > 0 {
		DPrintf("move the old chunks")
		oldfd, err := os.Open(filename)
		if !check(err, true) {
			return filename + "~", false
		}
		defer oldfd.Close()
		for _, chunk := range recipe {
			//DPrintf("moving chunk %v -> %v", chunk.Chunk, chunk.Moveto)
			buf := make([]byte, chunk.Chunk.Size)
			for i := int64(0); i < chunk.Chunk.Size; {
				n, err = oldfd.ReadAt(buf, chunk.Chunk.Offset+i)
				if !check(err, true) {
					return filename + "~", false
				}
				_, err = fd.WriteAt(buf[0:n], chunk.Moveto+i)
				if !check(err, true) {
					return filename + "~", false
				}
				i += int64(n)
			}
		}
	}

	//receive chunks we want
	DPrintf("receive the new chunks")
	buf := make([]byte, 4096)
	for _, chunk := range chunks_wanted {
		//DPrintf("getting chunk %v", chunk)
		//buf := make([]byte, chunk.Size)
		for i := int64(0); i < chunk.Size; {
			binary.Write(conn, binary.LittleEndian, chunk.Offset+i)
			n, err = conn.Read(buf)
			if !check(err, true) {
				return filename + "~", false
			}
			_, err := fd.WriteAt(buf[0:n], chunk.Offset+i)
			if !check(err, true) {
				return filename + "~", false
			}
			i += int64(n)
		}
	}

	done := int64(-1)
	binary.Write(conn, binary.LittleEndian, done)
	success = true
	return filename, true
}

func delete_file(filename string) (string, bool) {
	//first check if it exists
	//dir := filepath.Dir(filename)
	err := os.Remove(filename)
	if !check(err, true) {
		if os.IsNotExist(err) {
			return filename, true
		}
		return "", false
	}
	return filename, true
}
