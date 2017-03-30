# ExTRA
A re-write of the [TRA file synchronizer](http://publications.csail.mit.edu/lcs/pubs/pdf/MIT-LCS-TM-650.pdf) in go

## How It works
In order to better track file history, Tra uses synchonization vectors
in addition to version vectors. Version vectors track file modification
history while synchronization vectors track the synchronization history
of a file. With both version vectors and synchronization vectors
(the time-vector pair), Tra can reduce the amount of false positive
conflicts and reduce the size of file metadata (especially for deleted
files).

## ExTRA Plan of Attack
Maintain time-vector pairs for every file in a directory tree. FSwatch
will communicate file modifications to ExTRA. Users can sync to other
ExTRA replicas either manually or automatically every few seconds. When
a sync happens, the algorithms described in the paper (Figure 9) will
decide what to do.
