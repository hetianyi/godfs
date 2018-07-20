
#-----------------------------------------------
CREATE TABLE files (
    id        INTEGER      PRIMARY KEY AUTOINCREMENT,
    md5       VARCHAR (32) UNIQUE,
    parts_num INTEGER      DEFAULT (1)
);
CREATE UNIQUE INDEX index_file_md5 ON files (
    md5
);
#-----------------------------------------------
CREATE TABLE parts (
    id  INTEGER      PRIMARY KEY AUTOINCREMENT,
    md5 VARCHAR (32) UNIQUE
);
#-----------------------------------------------
CREATE TABLE parts_relation (
    id  INTEGER PRIMARY KEY,
    fid INTEGER,
    pid INTEGER
);
CREATE INDEX index_q1 ON parts_relation (
    fid
);
#-----------------------------------------------
CREATE TABLE task (
    id     INTEGER PRIMARY KEY AUTOINCREMENT
                   NOT NULL,
    type   INTEGER NOT NULL,
    fid    INTEGER,
    status INTEGER DEFAULT (1)
                   NOT NULL
);
#-----------------------------------------------
# 同步状态表，记录当前同步tracker服务器的最大file id
CREATE TABLE sys (
    master_sync_id INTEGER
);

