package main

//import (
//	"fmt"
//	"log"
//)

const (
	EVENT_FSOP     = 1
	EVENT_SYNC     = 2
	EVENT_SYNCTO   = 3
	EVENT_SYNCFROM = 4
	FSOP_MODIFY    = 5
	FSOP_DELETE    = 6
)

const Debug = 1

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		Log.Printf("ExTRA: "+format+"\n", a...)
	}
	return
}

var file_table map[string]File

type FS_OP struct {
	File   string
	Action int
}

type SYNC_OP struct {
	Host string
	Data SyncMsg
}

type SyncMsg struct {
	From  string
	Files map[string]File
	Resp  chan SyncReplyMsg
}

type SyncReplyMsg struct {
	Files map[string]bool
}

type Event struct {
	Type int
	Data interface{}
}

type File struct {
	Path    string
	Version PairVec
	Sync    PairVec
}

//symbolically modifies a file with our ID
func (f File) Modify() File {
	val, exists := f.Version.GetPair(ID)
	if !exists {
		panic("we dont exist in a file we have")
	}
	val.Counter += 1
	f.Version.Add(val)
	return f
}

func (f File) Show() string {
	out := f.Path
	out += " -> " + f.Version.Show()
	return out
}
