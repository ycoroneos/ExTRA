#!/bin/bash

#run fswatch and only report updates and deletions
#./fswatch/src/fswatch -rx --event Updated --event Removed testdir/ | src/src
#./fswatch/src/fswatch -rx --event Updated --event Removed --event Created \
#  --event Renamed --event IsFile --event IsDir testdir/ | src/src
#./fswatch/src/fswatch -arx testdir/
./fswatch/src/fswatch -m poll_monitor -arx testdir/ | src/src birdman
