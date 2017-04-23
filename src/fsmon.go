package main

import (
	"../golang-set"
	"os"
	"path/filepath"
	"time"
)

type Watcher struct {
	oldmap  mapset.Set
	newmap  mapset.Set
	path    string
	filters map[string]Sfile
}

func MakeWatcher(path string) *Watcher {
	out := &Watcher{mapset.NewSet(), mapset.NewSet(), path, make(map[string]Sfile)}
	//out.Poll()
	return out
}

func (w *Watcher) Poll() ([]Sfile, []string) {
	w.newmap = Getdirmap(w.path)
	modified, deleted := CompareMaps(w.oldmap, w.newmap)
	w.oldmap = w.newmap
	for i := 0; i < len(modified); i++ {
		val, exists := w.filters[modified[i].Name]
		if exists && val.Time == modified[i].Time {
			modified = append(modified[:i], modified[i+1:]...)
			delete(w.filters, modified[i].Name)
		}
	}

	//we dont support deletes yet
	return modified, deleted
}

func (w *Watcher) HasChanged(path string) bool {
	found := Sfile{}
	for _, file := range w.newmap.ToSlice() {
		if file.(Sfile).Name == path {
			found = file.(Sfile)
		}
	}
	stat, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			//stat = os.FileInfo{}
		} else {
			check(err, false)
		}
	}
	return (stat.ModTime() != found.Time)
}

func (w *Watcher) Addfilter(path string) bool {
	stat, err := os.Stat(path)
	if !check(err, false) {
		return false
	}
	w.filters[path] = Sfile{path, stat.ModTime(), stat.IsDir()}
	return true
}

type Sfile struct {
	Name  string
	Time  time.Time
	Isdir bool
}

func Getdirmap(path string) mapset.Set {
	output := mapset.NewSet()
	walkfunc := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err)
		}
		output.Add(Sfile{path, info.ModTime(), info.IsDir()})
		return nil
	}
	filepath.Walk(path, walkfunc)
	return output
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
