package main

import (
	"time"
)

type File struct {
	Path     string
	Time     time.Time
	Vcounter int64
	Scounter int64
	Version  PairVec
	Sync     PairVec
}

//symbolically modifies a file with our ID
func (f File) Modify() File {
	val, exists := f.Version.GetPair(ID)
	f.Vcounter += 1
	if exists {
		val.Counter = f.Vcounter
		f.Version.Add(val)
	} else {
		f.Version.Add(Pair{ID, f.Vcounter})
	}
	return f
}

//symbolically synchronizes a file
func (f File) SyncModify() File {
	val, exists := f.Sync.GetPair(ID)
	f.Scounter += 1
	if exists {
		val.Counter = f.Scounter
		f.Sync.Add(val)
	} else {
		f.Sync.Add(Pair{ID, f.Scounter})
	}
	return f
}

func (f File) Show() string {
	out := f.Path
	out += " -> " + f.Version.Show()
	return out
}

//do a pure delta of input to output
//modification filters are exceptions to file modifications that are set for a file when it
//is received from a remote and detected as a local modification on the file system
func delta(modified []Sfile, deleted []string, oldstate map[string]File) map[string]File {
	for _, mod := range modified {
		//skip directories
		if mod.Isdir {
		} else {
			val, exists := oldstate[mod.Name]
			if !exists {
				oldstate[mod.Name] = File{mod.Name, mod.Time, int64(1), int64(1), MakePairVec([]Pair{Pair{ID, 1}}), MakePairVec([]Pair{Pair{ID, 1}})}
			} else {
				oldstate[mod.Name] = val.Modify()
			}
		}
	}

	//we dont support deletes yet
	for _, del := range deleted {
		DPrintf("%v", del)
		panic("we dont support deletes yet")
	}
	return oldstate
}

//modify the synchronization vector
func syncmodify(input map[string]File) map[string]File {
	output := input
	for k, v := range output {
		output[k] = v.SyncModify()
	}
	return output
}
