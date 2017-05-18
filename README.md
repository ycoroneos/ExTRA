# ExTRA
A re-implementation of the [TRA file synchronizer](http://publications.csail.mit.edu/lcs/pubs/pdf/MIT-LCS-TM-650.pdf) in go
with some extra tweaks

## How To Use
First write a JSON config file like the samples in config/
Then run extra out of the sync directory and pass in the config path:


user@host:~/some/path/to/sync$ path/to/extra/extra.elf path/to/config.cfg


## Features
### False conflict detection:
Independently creating files with the same
name, when one of them has already been deleted, will not report a conflict.
Synchronization of two deleted files will never report a conflict.
Conflicts that are resolved by keeping an old version will never be
reported again unless the conflicting files change again.

### No Lost Updates
The updates to a file will not be lost because another participant
in the network has rejected a conflict.

### Incremental change propogation:
Only new parts of a file are sent over
the wire.

### Persistence:
Synchronization state is saved and loaded from disk on
startup.

### Efficient File Representation:
The size of the persisted state scales
linearly with the number of files that are being synchronized. A tree
with 22000 files must only maintain about 20 MB of state.

### Time Invariance
Computer wall clock time is not involved in the synchronization at all.

### Distributed
There is no central coordinator. All synchronizations are pairwise and
state is replicated across all participants.

### Portability:
ExTRA is 100% written in Go and the only non-standard package is golang-set


## How It works
In order to better track file history, Tra uses synchonization vectors
in addition to version vectors. Version vectors track file modification
history while synchronization vectors track the synchronization history
of a file. Each file is associated with a monotonically increasing counter for
tracking the causal order of modifications and syncs. With both version
vectors and synchronization vectors (the time-vector pair), Tra can
reduce the amount of false positive conflicts and reduce the size of
file metadata (especially for deleted files).

## ExTRA Plan of Attack
Maintain time-vector pairs for every file in a directory tree. When a
sync happens, the shared directory will be scanned and modifications
from the last scan will be incorporated into the current time-vectors.
The actual version resolution will proceed as described in the paper
(Figure 9).

