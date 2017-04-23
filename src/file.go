package main

import (
	"time"
)

type File struct {
	Path    string
	Time    time.Time
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

//do a pure delta of input to output
func delta(modified []Sfile, deleted []string, oldstate map[string]File) map[string]File {
	for _, mod := range modified {
		//skip directories
		if mod.Isdir {
		} else {
			val, exists := oldstate[mod.Name]
			if !exists {
				oldstate[mod.Name] = File{mod.Name, mod.Time, MakePairVec([]Pair{Pair{ID, 1}}), MakePairVec([]Pair{Pair{ID, 1}})}
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

//func do_fsop(op Event) {
//	val, exists := file_table[op.File]
//	switch op.Action {
//	case FSOP_DELETE:
//		//we better have it
//		if !exists {
//			panic("trying to delete a file we dont have")
//		} else {
//			delete(file_table, op.File)
//		}
//
//	case FSOP_MODIFY:
//		//we track this file now
//		if !exists {
//			newfile := File{op.File, MakePairVec([]Pair{Pair{ID, 1}}), MakePairVec([]Pair{Pair{ID, 1}})}
//			file_table[op.File] = newfile
//		} else {
//			newfile := val.Modify()
//			file_table[op.File] = newfile
//		}
//		DPrintf("%s", file_table[op.File].Show())
//	default:
//		panic("unimplemented fsop")
//	}
//}
