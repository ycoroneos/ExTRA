# ExTRA
A re-write of the TRA file synchronizer in go

## How It works
In order to better track file history, Tra uses synchonization vectors
in addition to version vectors. Version vectors track file modification
history while synchronization vectors track the
synchronization history of a file. With both version vectors and
synchronization vectors, Tra can reduce the amount of false positive
conflicts and reduce the size of file metadata (especially for deleted
files).
