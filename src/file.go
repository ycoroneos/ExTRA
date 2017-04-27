package main

import (
	"time"
)

type File struct {
	Path     string
	Time     time.Time
	Vcounter int64
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
				oldstate[mod.Name] = File{mod.Name, mod.Time, int64(1), MakePairVec([]Pair{Pair{ID, 1}}), MakePairVec([]Pair{Pair{ID, 1}})}
			} else {
				oldstate[mod.Name] = val.Modify()
				//				time, ignore := modfilters[mod.Name]
				//				if ignore && mod.Time == time {
				//					DPrintf("ignoring delta on %s just this once because it was recently synced and left unmodified")
				//				} else {
				//					oldstate[mod.Name] = val.Modify()
				//				}
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
