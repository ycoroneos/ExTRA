Version Vectors
----------------

Extra uses version vectors to keep track of the modification history of
a file. Extra represents each file with a monotonically increasing
counter so that when the file system monitor detects a change in a file,
Extra appends an *<ID, counter>* pair to the file's version vector and
then increments the count.

Here is an example of 3 computers A, B, and C sending a single file to
eachother.

<!-- language: lang-none -->

      1   2       3   4
     +------------------
    A|■     ▲
     | \   / \
    B|  ■ ▲   \     ◆ ●
     |         \   /
    C|          ▲ ◆

At each spots 1, 2, and 3 the file has been modified.
In spot 1 the version vector is <A<sub>1</sub>>. In spot
2 the version vector is <A<sub>1</sub>, B<sub>2</sub>>. In spot
3 the version vector is <A<sub>1</sub>, B<sub>2</sub>, C<sub>3</sub>>.
In spot 4 the version vector is <A<sub>1</sub>, B<sub>4</sub>, C<sub>3</sub>>

Synchronization Vectors
-----------------------

Unlike version vectors, synchronization vectors track the
synchronization history of the file. Extra implements these vectors
using the same monotonically increasing counter as the version vectors.
After a successful synchronization, each participant in the sync,
updates its synchronization vector with the IDs of the participants and
the current value of the file's counter. The counter does not increment
after synchronizations, only modifications.

Determining The Most Recent File During a Sync
-----------------------------------------------

Both version vectors and synchronization vectors form a partially
ordered set of (ID, count) pairs, so less than and equality can be defined for this set:
A == B iff (all elements of A are in B) AND (all elements of B are in A)

A < B  iff (
    (all elements of A are in B)
    AND
    (the count of every element in A is less than the corresponding count in B)
  )
AND
(B has at least one element whose counter is greater then the counter of the corresponding element in A)

A <= B iff (A < B) || (A == B)

Now pretend A and B are version vectors of a file. If A <= B, then
version A is a superset of version B. Simply put, A is a newer file. If
A /<= B and B /<= A, then that means that the two files have diverging
histories and are now in a conflict. When the conflict is resolved, the
resolver can append its (ID, count) pair onto the current version to
mark the resolution.

The basic algorithm for file synchronization with version vectors is
thus:
<!-- language: lang-none -->

    sync(A -> B, file):
      if (A.modification == B.modification)
        //do nothing
      else if (A.modification < B.modification)
        //do nothing
      else if (B.modification < A.modification)
        //accept the new file
      else
        //report a conflict

Going back to the file in the example above:
<!-- language: lang-none -->

      1   2       3   4
     +------------------
    A|■     ▲
     | \   / \
    B|  ■ ▲   \     ◆ ●
     |         \   /
    C|          ▲ ◆


The sync that occurs with B and C right after C modifies the file
will compare the following version vectors:

B : <A<sub>1</sub>, B<sub>2</sub>>

C : <A<sub>1</sub>, B<sub>2</sub>, C<sub>3</sub>>

#How Synchronization Vectors Can Help
Even though version vectors contain enough information to determine
which file is newer, they can still lead to false conflicts, which are
situations where file histories diverge and it is OK. The simplest
scenario to imagine is where a user resolves a conflict by doing nothing

