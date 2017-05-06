package main

import (
	"time"
)

type File struct {
	Path     string
	Time     time.Time
	Deleted  bool
	Counter  int64
	Creation PairVec //holds a single Pair
	Version  PairVec
	Sync     PairVec
}

//why do i need this?
func (f File) Copy() File {
	out := File{f.Path, f.Time, f.Deleted, f.Counter, MakePairVec(f.Creation.GetSlice()), MakePairVec(f.Version.GetSlice()), MakePairVec(f.Sync.GetSlice())}
	return out
}

//symbolically deletes a file
func (f File) Delete() File {
	//DPrintf("delete %v from versions %v", f.Path, f)
	if f.Deleted {
		DPrintf("WARNING: deleting an already deleted file")
	}
	f.Deleted = true
	return f
}

//symbolically re-creates an independant file with the same name
func (f File) Baptize() File {
	DPrintf("baptize %v", f.Path)
	if !f.Deleted {
		DPrintf("WARNING: baptizing an un-deleted file")
	}
	f.Deleted = false
	firstvec := f.Version.GetSlice()[0]
	f.Creation = MakePairVec([]Pair{firstvec})
	//f.Creation.Add(Pair{ID, f.Vcounter})
	return f
}

//symbolically modifies a file with our ID
func (f File) Modify() File {
	val, exists := f.Version.GetPair(ID)
	f.Counter += 1
	if exists {
		val.Counter = f.Counter
		f.Version.Add(val)
	} else {
		f.Version.Add(Pair{ID, f.Counter})
	}
	//interesting optimization
	f.Version = f.Version.Trim()
	return f
}

//symbolically synchronizes a file
func (f File) SyncModify() File {
	DPrintf("syncmodify %v", f)
	val, exists := f.Sync.GetPair(ID)
	//f.Scounter += 1
	//DPrintf("syncmodify %v to %v", f.Path, f.Scounter)
	if exists {
		val.Counter = f.Counter
		f.Sync.Add(val)
	} else {
		f.Sync.Add(Pair{ID, f.Counter})
	}
	DPrintf("syncmodify new %v", f)
	return f
}

func (f File) BackSync(them_id string) File {
	val, exists := f.Sync.GetPair(them_id)
	DPrintf("backsyncmodify %v", f)
	//f.Scounter += 1
	if exists {
		val.Counter = f.Counter
		f.Sync.Add(val)
	} else {
		f.Sync.Add(Pair{them_id, f.Counter})
	}
	DPrintf("backsyncmodify new %v", f)
	return f
}

func (f File) Show() string {
	out := f.Path
	out += " -> " + f.Version.Show()
	return out
}

//do a pure delta of input to output
func delta(modified []Sfile, deleted map[string]bool, oldstate map[string]File) map[string]File {
	for _, mod := range modified {
		//skip directories
		if mod.Isdir {
		} else {
			val, exists := oldstate[mod.Name]
			if !exists {
				oldstate[mod.Name] = File{mod.Name, mod.Time, false, int64(0), MakePairVec([]Pair{Pair{ID, 0}}), MakePairVec([]Pair{Pair{ID, 0}}), MakePairVec([]Pair{})}
			} else {
				if val.Deleted {
					oldstate[mod.Name] = val.Modify().Baptize()
				} else {
					oldstate[mod.Name] = val.Modify()
				}
			}
		}
	}

	for del, _ := range deleted {
		//deletes are treated as another modification
		//DPrintf("fsmon deleting %s", del)
		if oldstate[del].Deleted {
			DPrintf("WARNING: fsmon tried to delete a deleted file from a sync")
			continue
		}
		file := oldstate[del]
		file = file.Modify().Delete()
		oldstate[del] = file
		//DPrintf("%v", del)
		//panic("we dont support deletes yet")
	}
	return oldstate
}

//modify the synchronization vector
func syncmodify(input map[string]File) map[string]File {
	output := make(map[string]File)
	for k, v := range input {
		output[k] = v.Copy().SyncModify()
	}
	return output
}
