# ExTRA
A re-implementation of the [TRA file synchronizer](http://publications.csail.mit.edu/lcs/pubs/pdf/MIT-LCS-TM-650.pdf) in go

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

