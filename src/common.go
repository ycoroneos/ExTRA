package main

import (
	"net"
)

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

func check(e error, soft bool) bool {
	if e != nil {
		if soft {
			DPrintf("%v", e)
			return false
		}
		panic(e)
	}
	return true
}

func gomin(a, b int64) int64 {
	if a < b {
		return a
	} else {
		return b
	}
}

type Event struct {
	Type int
	//Data interface{}

	//FSOP
	File   string
	Action int

	//SYNCOP
	Host  string
	From  string
	Files map[string]File
	//Resp  chan SyncReplyMsg
	Resp chan bool
	Wire net.Conn
	//Tx   gob.Encoder
	//Rx   gob.Decoder
}

//received in response to a sync attempt
type SyncReplyMsg struct {
	Files map[string]bool
}
