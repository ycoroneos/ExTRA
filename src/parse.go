package main

import (
	"bufio"
	"os"
	"strings"
)

//takes input from fswatch and creates FS_OP events for the event loop
func input_parser(events chan Event) {
	reader := bufio.NewReader(os.Stdin)
	for {
		command, _ := reader.ReadString('\n')
		fields := strings.Fields(command)
		switch fields[1] {
		case "Removed":
			events <- Event{Type: EVENT_FSOP, File: fields[0], Action: FSOP_DELETE}
		case "Created":
			events <- Event{Type: EVENT_FSOP, File: fields[0], Action: FSOP_MODIFY}
		case "Updated":
			events <- Event{Type: EVENT_FSOP, File: fields[0], Action: FSOP_MODIFY}
		}
	}
}
