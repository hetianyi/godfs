### TODO list



2019/01/25

------

- Implement a message flow mechanism, which can push new file info to tracker server quickly rather than push files by timer.

- Create a file stream tunnel at low level for file synchronization  between group servers, this is in case of other storage servers cannot catch up with servers who is much ahead of them. This way can synchronize files without query db.