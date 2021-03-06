package main

import (
	//	"fmt"
	"crypto/sha256"
	"os"
)

type FileChunk struct {
	Offset   int64
	Checksum uint64
	Size     int64
	Sha      [32]byte
}

func calculate_sha1(data []byte) [32]byte {
	h := sha256.Sum256(data)

	out := [32]byte{}
	for i, c := range h {
		out[i] = c
	}
	return out
}

func adler32(window []uint32, num uint32) ([]uint32, uint32) {
	output := append(window[1:], num)
	sum := uint32(0)
	for _, v := range output {
		sum += v
	}
	return output, sum
}

//computes the rolling checksum
//func Rollhash(filename string) []FileChunk {
//	chunks := make([]FileChunk, 0)
//	fd, err := os.Open(filename)
//	if !check(err, true) {
//		return chunks
//	}
//	defer fd.Close()
//	stat, err := fd.Stat()
//	if !check(err, true) {
//		return chunks
//	}
//	size := stat.Size()
//	checksum := uint32(0)
//	window := make([]uint32, 8192)
//	data := make([]byte, 8192)
//	offset := int64(0)
//	for i := int64(0); i < size; {
//		n, err := fd.Read(data)
//		if !check(err, true) || n == 0 {
//			return make([]FileChunk, 0)
//		}
//		for j := int64(0); j < int64(n); j++ {
//			window, checksum = adler32(window, uint32(data[j]))
//			if checksum%4096 == 0 && (i+j) > 4096 {
//				chunks = append(chunks, FileChunk{offset, checksum, (i + j) - offset})
//				offset = i + j
//			}
//		}
//		i += int64(n)
//	}
//	//trailing checksum
//	if len(chunks) == 0 || checksum != chunks[len(chunks)-1].Checksum {
//		chunks = append(chunks, FileChunk{offset, checksum, size - offset})
//	}
//	return chunks
//}

//func FastRollhash(filename string) []FileChunk {
//	chunks := make([]FileChunk, 0)
//	fd, err := os.Open(filename)
//	if !check(err, true) {
//		return chunks
//	}
//	defer fd.Close()
//	stat, err := fd.Stat()
//	if !check(err, true) {
//		return chunks
//	}
//	size := stat.Size()
//	checksum := uint32(0)
//	offset := int64(0)
//	wsize := int64(8192 * 5)
//	window := make([]uint32, wsize)
//	data := make([]byte, 10*10*1024)
//	for i := int64(0); i < size; {
//		n, err := fd.Read(data)
//		if !check(err, true) || n == 0 {
//			return make([]FileChunk, 0)
//		}
//		for j := int64(0); j < int64(n); j++ {
//			if i+j >= wsize {
//				checksum -= window[(j+i)%wsize]
//			}
//			window[(j+i)%wsize] = uint32(data[j])
//			checksum += uint32(data[j])
//			//if checksum%(uint32(wsize)/2) == 0 && (i+j) > (int64(wsize)/2) && checksum != 0 {
//			//if checksum%(uint32(wsize)/2) == 0 && (i+j) > (int64(wsize)/2) {
//			if checksum%(uint32(65521)) == 0 && (i+j) > (int64(wsize)/2) {
//				chunks = append(chunks, FileChunk{offset, checksum, (i + j) - offset})
//				offset = i + j
//			}
//		}
//		i += int64(n)
//	}
//	if len(chunks) == 0 || checksum != chunks[len(chunks)-1].Checksum {
//		chunks = append(chunks, FileChunk{offset, checksum, size - offset})
//	}
//	return chunks
//}

func RollhashSha(filename string) []FileChunk {
	chunks := make([]FileChunk, 0)
	fd, err := os.Open(filename)
	if !check(err, true) {
		return chunks
	}
	defer fd.Close()
	stat, err := fd.Stat()
	if !check(err, true) {
		return chunks
	}
	size := stat.Size()
	checksum := uint64(0)
	offset := int64(0)
	wsize := int64(8192 * 2)
	window := make([]byte, wsize)
	data := make([]byte, 10*10*1024)
	hashbytes := make([]byte, 0)
	for i := int64(0); i < size; {
		n, err := fd.Read(data)
		if !check(err, true) || n == 0 {
			return make([]FileChunk, 0)
		}
		for j := int64(0); j < int64(n); j++ {
			if i+j >= wsize {
				checksum -= uint64(window[(j+i)%wsize])
			}
			window[(j+i)%wsize] = data[j]
			checksum += uint64(data[j])
			hashbytes = append(hashbytes, data[j])
			//if checksum%(uint32(wsize)/2) == 0 && (i+j) > (int64(wsize)/2) && checksum != 0 {
			//if checksum%(uint32(wsize)/2) == 0 && (i+j) > (int64(wsize)/2) {

			//if checksum%(uint32(65521)) == 0 && (i+j) > (int64(wsize)/2) {
			if checksum%(uint64(65521)) == 0 && (i+j) > (int64(wsize)/2) && checksum != 0 {
				chunks = append(chunks, FileChunk{offset, checksum, (i + j) - offset, calculate_sha1(hashbytes)})
				offset = i + j
				hashbytes = make([]byte, 0)
			}
		}
		i += int64(n)
	}
	if len(chunks) == 0 || checksum != chunks[len(chunks)-1].Checksum {
		chunks = append(chunks, FileChunk{offset, checksum, size - offset, calculate_sha1(hashbytes)})
	}
	return chunks
}

//func makechunks(filenames []Sfile) map[string][]FileChunk {
//	output := make(map[string][]FileChunk)
//	for _, f := range filenames {
//		output[f.Name] = Rollhash(f.Name)
//	}
//	return output
//}

type ChunkDelta struct {
	Chunk  FileChunk
	Moveto int64
}

func shamatch(a, b [32]byte) bool {
	for i := 0; i < 32; i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func Diff(them, ours []FileChunk) ([]FileChunk, []ChunkDelta) {
	need := make([]FileChunk, 0)
	have := make([]ChunkDelta, 0)
	table := make(map[uint64][]FileChunk, 0)
	for _, c := range ours {
		v, exists := table[c.Checksum]
		if exists {
			table[c.Checksum] = append(v, c)
		} else {
			table[c.Checksum] = []FileChunk{c}
		}
	}

	//fmt.Printf("%v\n", table)

	for _, c := range them {
		v, exists := table[c.Checksum]
		//fmt.Printf("%v -> ", c)
		if exists {
			found := false
			for _, cc := range v {
				if c.Checksum == cc.Checksum && c.Size == cc.Size && shamatch(c.Sha, cc.Sha) {
					have = append(have, ChunkDelta{cc, c.Offset})
					found = true
				}
			}
			if !found {
				//fmt.Printf("no\n")
				need = append(need, c)
			}
		} else {
			//fmt.Printf("no\n")
			need = append(need, c)
		}
	}
	return need, have
}

func ChompAlgo(them, ours []FileChunk) ([]FileChunk, []ChunkDelta) {
	have := make([]ChunkDelta, 0)
	need := make([]FileChunk, 0)
	for {
		if len(ours) == 0 || len(them) == 0 {
			break
		} else if (them[0].Checksum == ours[0].Checksum) && (them[0].Size == ours[0].Size) {
			have = append(have, ChunkDelta{ours[0], them[0].Offset})
			them = them[1:]
			ours = ours[1:]
		} else if len(them) > len(ours) {
			need = append(need, them[0])
			them = them[1:]
		} else if len(ours) > len(them) {
			ours = ours[1:]
		} else if len(ours) == len(them) {

			//DPrintf("inception\n")
			//explore both sides of the fork
			_, havea := ChompAlgo(them[1:], ours)
			_, haveb := ChompAlgo(them, ours[1:])

			if len(havea) > len(haveb) {
				//need = append(need, needa[0:]...)
				//	have = append(have, havea[0:]...)
				them = them[1:]
			} else {
				//		need = append(need, needb[0:]...)
				//			have = append(have, haveb[0:]...)
				ours = ours[1:]
			}

		}
	}
	need = append(need, them[0:]...)
	return need, have
}

func CompareChunks(them, ours []FileChunk) ([]FileChunk, []ChunkDelta) {
	need := make([]FileChunk, 0)
	have := make([]ChunkDelta, 0)
	index := 0
	for i := 0; i < len(ours); {
		if ours[i].Checksum == them[index].Checksum {
			have = append(have, ChunkDelta{ours[i], them[index].Offset})
			index += 1
			i += 1
		} else {
			need = append(need, them[index])
			index += 1
		}
	}
	return need, have
}

type DataChunk struct {
	Chunk FileChunk
	Data  []byte
}

func Readchunks(filename string, chunks []FileChunk) chan DataChunk {
	output := make(chan DataChunk, 0)
	go func() {
		defer close(output)
		fd, err := os.Open(filename)
		if !check(err, true) {
			return
		}
		defer fd.Close()
		for _, c := range chunks {
			data := make([]byte, c.Size)
			fd.ReadAt(data, c.Offset)
			output <- DataChunk{c, data}
		}
	}()
	return output
}
