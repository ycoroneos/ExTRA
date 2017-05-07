package main

import (
	"bytes"
	"encoding/binary"
	//	"encoding/gob"
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

	//send the chunks
	//enc := gob.NewEncoder(conn)
	//for chunk := range Readchunks(file, chunks) {
	//}

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
