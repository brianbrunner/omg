# omg

Data persistance layer that is **O**bject-oriented, in-**M**emory, and written in **G**o.

**Note: This project is currently on hiatus, but will likely come back once I've got. some time to get around to it**

# The Premise

omg is an in-memory data structure server, similar to redis. 

* It is fairly fast for operations that have redis equivalents, roughly 80-90% the speed of redis. 
* It consumes a fairly small amount of memory, although again, not as little as redis. 
* It persists to disk, similarly to redis, using both an append only file as well as full backups
* Its backup method is arguably better than redis' as it doesn't do any silly forking. Instead, it makes sure that if a write operation is performed on a key during backup, that key is immediately dumped to the backup file before the modification is performed.
* And of course, the biggest feature, it is extensible, allowing you to create new data structures and commands to operate on them.
