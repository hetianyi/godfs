### TODO list



2019/01/25

------

- Implement a message flow mechanism, which can push new file info to tracker server quickly rather than push files by timer.

- Create a file stream tunnel at low level for file synchronization  between group servers, this is in case other storage servers cannot catch up with servers who is much ahead of them. This way can synchronize files without query db.

- If there is too many concurrent upload at a time, try to auto slow down the file synchronization or even stop it in case competition in db transaction.