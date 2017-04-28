package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"net"
	"os"
	"path/filepath"
	//"strings"
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

func fillright(str, pad string, length int) string {
	for {
		str += pad
		if len(str) > length {
			return str[0:length]
		}
	}
}

type Filechunk struct {
	Name   string
	Size   int64
	Data   []byte
	Offset int64
}

func send_file(conn net.Conn, file string) bool {
	if len(file) > 1024 {
		return false
	}
	fd, err := os.Open(file)
	if !check(err, true) {
		return false
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

	//	//send the file
	//	copied, err := io.Copy(conn, fd)
	//	if copied != stat.Size() {
	//		DPrintf("did not send the whole file")
	//		return false
	//	}
	//	return true

	//	bufsize := 1024
	//	fd, err := os.Open(file)
	//	if !check(err, true) {
	//		return false
	//	}
	//	stat, err := fd.Stat()
	//	if !check(err, true) {
	//		return false
	//	}
	//	defer fd.Close()
	//
	//	enc := gob.NewEncoder(conn)
	//	dec := gob.NewDecoder(conn)
	//	firstchunk := Filechunk{file, stat.Size(), make([]byte, 0), 0}
	//	if !check(enc.Encode(firstchunk), true) {
	//		return false
	//	}
	//
	//	//	if !check(binary.Write(conn, binary.LittleEndian, stat.Size()), true) {
	//	//		return false
	//	//	}
	//	//
	//	//	_, err = conn.Write([]byte(fillright(file, "/", bufsize)))
	//	//	if !check(err, true) {
	//	//		return false
	//	//	}
	//
	//	//fstats := filestats{file, stat.Size()}
	//	//enc := gob.NewEncoder(conn)
	//	//if !check(enc.Encode(fstats), true) {
	//	//	return false
	//	//}
	//	offset := int64(0)
	//	sendbuffer := make([]byte, bufsize)
	//	for {
	//		n, err := fd.Read(sendbuffer)
	//		if err == io.EOF && n == 0 {
	//			break
	//		}
	//		nextchunk := Filechunk{file, stat.Size(), sendbuffer, offset}
	//		if !check(enc.Encode(nextchunk), true) {
	//			return false
	//		}
	//		offset += int64(n)
	//		//conn.Write(sendbuffer)
	//	}
	//
	//	var ack Filechunk
	//	if !check(dec.Decode(&ack), true) {
	//		return false
	//	}
	//	return ack.Offset == -1

	//wait for ack
	//ack := make([]byte, 3)
	//conn.Read(ack)
	//return string(ack[:3]) == "ack"
	//	ack := 0
	//	if !check(binary.Read(conn, binary.LittleEndian, &ack), true) {
	//		return false
	//	}
	//	return ack == 0xcafe

	//info, err := os.Stat(file)
	//size := info.Size()
	//ten_mb := int64(10485760)
	//chunks := (size + ten_mb - 1) / ten_mb
	//fd, err := os.Open(file) // For read access.
	//check(err, false)
	//defer fd.Close()            // make sure to close the file even if we panic.
	//	enc := gob.NewEncoder(conn) // Will write to network.
	//	data, err := ioutil.ReadFile(file)
	//	if !check(err, true) {
	//	}
	//	//DPrintf("read file as %v", data)
	//	enc.Encode(FileChunk{file, 0, data})
	//	//	offset := int64(0)
	//	//	for {
	//	//		//data := make([]byte, gomin(ten_mb, size-offset))
	//	//		data := make([]byte, 1)
	//	//		n, err := fd.Read(data)
	//	//		DPrintf("sent bytes %v, %v out of %v", n, data, size)
	//	//		enc.Encode(FileChunk{file, offset, data[0:n]})
	//	//		if err == io.EOF {
	//	//			break
	//	//		}
	//	//		offset += int64(n)
	//	//	}
	//	enc.Encode(FileChunk{file, -1, nil})
	//	return true
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
	//	if filesz == 0 {
	//		return "", false
	//	}
	if !check(err, true) {
		return "", false
	}

	//make the directory path
	dir := filepath.Dir(filename)
	DPrintf("make dir %s", dir)
	if !check(os.MkdirAll(dir, 0744), true) {
		return "", false
	}

	//make the file
	DPrintf("open file %s, len s %d", filename, len(filename))
	fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
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

	//	copied, err := io.Copy(fd, conn)
	//	if copied != filesz {
	//		DPrintf("did not receive the whole file")
	//		return ""
	//	}
	//	if !check(err, true) {
	//		return ""
	//	}
	//	return filename

	//	var firstchunk Filechunk
	//	//bufsize := int64(1024)
	//	enc := gob.NewEncoder(conn)
	//	dec := gob.NewDecoder(conn)
	//	if !check(dec.Decode(&firstchunk), true) {
	//		return ""
	//	}
	//	filename := firstchunk.Name
	//	//filesz := firstchunk.Size
	//	//	var filesz int64
	//	//	if !check(binary.Read(conn, binary.LittleEndian, &filesz), true) {
	//	//		return ""
	//	//	}
	//	//
	//	//	filebytes := make([]byte, bufsize)
	//	//	_, err := conn.Read(filebytes)
	//	//	if !check(err, true) {
	//	//		return ""
	//	//	}
	//	//	filename := string(filebytes[:len(filebytes)])
	//	//	filename = strings.TrimRight(filename, "/")
	//
	//	dir := filepath.Dir(filename)
	//	if !check(os.MkdirAll(dir, 0744), true) {
	//		return ""
	//	}
	//
	//	//next make the file
	//	fd, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	//	if !check(err, true) {
	//		return ""
	//	}
	//	defer fd.Close()
	//
	//	received := int64(0)
	//	var nextchunk Filechunk
	//	for {
	//		if !check(dec.Decode(&nextchunk), true) {
	//			return ""
	//		}
	//		if nextchunk.Offset < received {
	//			panic("out of order reception in TCP")
	//		}
	//		n, _ := fd.Write(nextchunk.Data)
	//		if n != len(nextchunk.Data) {
	//			panic("wrote less than data size")
	//		}
	//		received += int64(n)
	//		DPrintf("received %d out of %d bytes", received, nextchunk.Size)
	//		if received >= nextchunk.Size {
	//			break
	//		}
	//	}
	//
	//	ackchunk := Filechunk{filename, 0, make([]byte, 0), -1}
	//	if !check(enc.Encode(ackchunk), true) {
	//		return ""
	//	}
	//
	//	//	received := int64(0)
	//	//	for {
	//	//		if (filesz - received) < bufsize {
	//	//			io.CopyN(fd, conn, filesz-received)
	//	//			conn.Read(make([]byte, (received+bufsize)-filesz))
	//	//			break
	//	//		}
	//	//		io.CopyN(fd, conn, bufsize)
	//	//		received += bufsize
	//	//	}
	//	//	ack := "ack"
	//	//	conn.Write([]byte(ack))
	//	//ack := 0xcafe
	//	//if !check(binary.Write(conn, binary.LittleEndian, ack), true) {
	//	//		return ""
	//	//	}
	//	return filename

	//	dec := gob.NewDecoder(conn)
	//	var nextchunk FileChunk
	//	if !check(dec.Decode(&nextchunk), true) {
	//		DPrintf("bad reception")
	//		return ""
	//	}
	//	if nextchunk.Offset != 0 {
	//		panic("out of order reception with TCP?")
	//	}
	//	//first make the path
	//	dir := filepath.Dir(nextchunk.Path)
	//	if !check(os.MkdirAll(dir, 0744), true) {
	//		DPrintf("bad reception")
	//		return ""
	//	}
	//	//fd, err := os.Create(nextchunk.Path)
	//	fd, err := os.OpenFile(nextchunk.Path, os.O_RDWR|os.O_CREATE, 0644)
	//	if !check(err, true) {
	//		DPrintf("bad reception")
	//		return ""
	//	}
	//	for {
	//		fd.WriteAt(nextchunk.Data, nextchunk.Offset)
	//		if !check(dec.Decode(&nextchunk), true) {
	//			return ""
	//		}
	//		if nextchunk.Offset == -1 {
	//			return nextchunk.Path
	//		}
	//	}
}
