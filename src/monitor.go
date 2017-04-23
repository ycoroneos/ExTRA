package main

import (
	"github.com/andreaskoch/go-fswatch"
	"path/filepath"
	"strings"
)

func fs_monitor(events chan Event, polltime int, watchpath string) {
	recurse := true
	skipDotFilesAndFolders := func(path string) bool {
		return strings.HasPrefix(filepath.Base(path), ".")
	}

	checkIntervalInSeconds := polltime
	folderWatcher := fswatch.NewFolderWatcher(watchpath, recurse, skipDotFilesAndFolders, checkIntervalInSeconds)

	folderWatcher.Start()
	DPrintf("fs monitor is running on path %v", watchpath)
	for folderWatcher.IsRunning() {

		select {

		case <-folderWatcher.Modified():
			DPrintf("New or modified items detected")

		case <-folderWatcher.Moved():
			DPrintf("Items have been moved")

		case changes := <-folderWatcher.ChangeDetails():

			//DPrintf("%s\n", changes.String())
			//for _, file := range changes.String() {
			//	events <- Event{Type: EVENT_FSOP, File: file, Action: FSOP_MODIFY}
			//}/

			DPrintf("New: %#v\n", changes.New())
			for _, file := range changes.New() {
				events <- Event{Type: EVENT_FSOP, File: file, Action: FSOP_MODIFY}
			}

			DPrintf("Modified: %#v\n", changes.Modified())
			for _, file := range changes.Modified() {
				events <- Event{Type: EVENT_FSOP, File: file, Action: FSOP_MODIFY}
			}

			DPrintf("Moved: %#v\n", changes.Moved())
			for _, file := range changes.Moved() {
				events <- Event{Type: EVENT_FSOP, File: file, Action: FSOP_DELETE}
			}

		}
	}
}
