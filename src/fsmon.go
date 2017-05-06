package main

import (
	"../golang-set"
	"os"
	"path/filepath"
	"time"
)

type Watcher struct {
	Oldmap  mapset.Set
	Newmap  mapset.Set
	Path    string
	Filters map[string]Sfile
}

type SerialWatcher struct {
	Oldmap  []Sfile
	Newmap  []Sfile
	Path    string
	Filters map[string]Sfile
}

func MakeFromSerial(s SerialWatcher) *Watcher {
	oldmap := mapset.NewSet()
	for _, k := range s.Oldmap {
		oldmap.Add(k)
	}
	newmap := mapset.NewSet()
	for _, k := range s.Newmap {
		newmap.Add(k)
	}
	return &Watcher{oldmap, newmap, s.Path, s.Filters}
}

func MakeWatcher(path string) *Watcher {
	out := &Watcher{mapset.NewSet(), mapset.NewSet(), path, make(map[string]Sfile)}
	//out.Poll()
	return out
}

func (w *Watcher) Serialize() SerialWatcher {
	oldset := make([]Sfile, w.Oldmap.Cardinality())
	for i, arg := range w.Oldmap.ToSlice() {
		oldset[i] = arg.(Sfile)
	}
	newset := make([]Sfile, w.Newmap.Cardinality())
	for i, arg := range w.Newmap.ToSlice() {
		newset[i] = arg.(Sfile)
	}
	return SerialWatcher{oldset, newset, w.Path, w.Filters}
}

func (w *Watcher) Poll(filter, deleted_filters []Sfile) ([]Sfile, map[string]bool) {
	//first add filters to the old map so they dont show up
	//in the output of the high pass filter
	for _, f := range filter {
		w.Oldmap.Add(f)
	}

	//remove deleted things from the old map so they dont show up as deleted
	for _, f := range deleted_filters {
		//		DPrintf("oldmap has %v", w.oldmap)
		w.Oldmap.Remove(f)
		//	DPrintf("oldmap has %v", w.oldmap)
	}

	//generate the current directory listing
	w.Newmap = Getdirmap(w.Path)
	//	DPrintf("new map is %v", w.newmap)
	//	DPrintf("old map is %v", w.oldmap)
	modified, deleted := CompareMaps(w.Oldmap, w.Newmap)
	//	DPrintf("deleted is %v", deleted)
	w.Oldmap = w.Newmap
	for i := 0; i < len(modified); i++ {
		val, exists := w.Filters[modified[i].Name]
		if exists && val.Time == modified[i].Time {
			modified = append(modified[:i], modified[i+1:]...)
			delete(w.Filters, modified[i].Name)
		}
	}

	//	faketime := time.Time{}
	delete_map := make(map[string]bool)
	for _, d := range deleted {
		//	val, exists := w.filters[d]
		//	if exists && val.Time == faketime {
		//		delete(w.filters, d)
		//	} else {
		delete_map[d] = true
		//	}
	}

	return modified, delete_map
}

func (w *Watcher) HasChanged(path string) bool {
	DPrintf("checking if %s has changed", path)
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			DPrintf("path does not exist")
			return true
			//stat = os.FileInfo{}
		} else {
			check(err, false)
		}
	}
	return !w.Newmap.Contains(Sfile{path, stat.ModTime(), stat.IsDir()})
}

//func (w *Watcher) Addfilter(path string) bool {
//	stat, err := os.Stat(path)
//	if !check(err, false) {
//		return false
//	}
//	w.Filters[path] = Sfile{path, stat.ModTime(), stat.IsDir()}
//	return true
//}

type Sfile struct {
	Name  string
	Time  time.Time
	Isdir bool
}

func Getdirmap(path string) mapset.Set {
	output := mapset.NewSet()
	walkfunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
			} else {
				panic(err)
			}
		}
		mode := info.Mode()

		//ignore non-files
		if (mode&os.ModeSymlink) > 0 || (mode&os.ModeSocket) > 0 || (mode&os.ModeDevice) > 0 || (mode&os.ModeNamedPipe) > 0 || info.IsDir() {
		} else {
			output.Add(Sfile{path, info.ModTime(), info.IsDir()})
		}
		return nil
	}
	filepath.Walk(path, walkfunc)
	return output
}

func Cleanup(dpath string) int {
	//	if isdir
	//	return 1
	//
	//	else
	//	sum=0
	//	for d in range readdirnames:
	//		count=Cleanup(d)
	//		sum+=count
	//		if count==0
	//			delete d
	//	return sum

	stat, err := os.Stat(dpath)
	if !check(err, true) {
		return 1
	}
	if !stat.IsDir() {
		//DPrintf("not a dir")
		return 1
	}
	sum := 0
	fd, err := os.Open(dpath)
	if !check(err, true) {
		return 1
	}
	defer fd.Close()
	subdirs, err := fd.Readdirnames(0)
	for _, d := range subdirs {
		if d == "." || d == ".." {
			continue
		}
		newpath := filepath.Join(dpath, d)
		count := Cleanup(newpath)
		if count == 0 {
			err = os.Remove(newpath)
			if !check(err, true) {
				return 1
			}
		}
		sum += count
	}
	return sum
}

func CompareMaps(old, new mapset.Set) ([]Sfile, []string) {
	modified := make([]Sfile, 0)
	for _, file := range new.Difference(old).ToSlice() {
		modified = append(modified, file.(Sfile))
	}
	elem := old.ToSlice()
	old = mapset.NewSet()
	for _, name := range elem {
		old.Add(name.(Sfile).Name)
	}
	elem = new.ToSlice()
	new = mapset.NewSet()
	for _, name := range elem {
		new.Add(name.(Sfile).Name)
	}
	deleted := make([]string, 0)
	for _, file := range old.Difference(new).ToSlice() {
		deleted = append(deleted, file.(string))
	}
	return modified, deleted
}
