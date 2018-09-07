godfs
==========
[![Build Status](https://travis-ci.org/hetianyi/godfs.svg?branch=master)](https://travis-ci.org/hetianyi/godfs)

[README](README.md) | [中文文档](README_zh.md)
# Description

### ```godfs``` is a simple fast, easy use distributed file system written by golang.

```godfs``` provides out-of-the-box usage and friendly support for docker，

You can pull the image on docker hub:
[https://hub.docker.com/r/hehety/godfs/](https://hub.docker.com/r/hehety/godfs/)

![architecture](/doc/20180830151005.png)

## Features

- Fast, lightweight, stable, out-of-the-box, friendly api.
- Easy to expand, Stable to RUN.
- Low resource overhead.
- Native client api and java client api(not started).
- API for http upload and download.
- Clear logs help troubleshoot errors.
- Support different platforms: Linux, Windows, Mac
- Better support docker.
- File fragmentation storage.
- Perfect data migration solution.
- Support readonly node.
- File synchronization in same group.

## Install

> Please install golang1.9+ first!

Take CentOS 7 as example.

### build from latest source code:
```javascript
yum install golang -y
git clone https://github.com/hetianyi/godfs.git
cd godfs
./make.sh
# on windows you just need to click 'make.cmd'
```
After the build is successful, three files will be generated under the `````./bin````` directory:
```
./bin/client
./bin/storage
./bin/tracker
```

Install godfs binary files to ```/usr/local/godfs```:
```javascript
./install.sh /usr/local/godfs
```
You can start tracker server by:
```javascript
/usr/local/godfs/bin/tracker [-c /your/tracker/config/path]
```
and start storage node by:
```javascript
/usr/local/godfs/bin/storage [-c /your/storage/config/path]
```
then you can using command ```client``` directly in command line to upload and download file.
> Of course, you must first set up the tracker server.
```javascript
# set up tracker servers for client
client --set "trackers=host1:port1[,host2:port2]"
```

For example:
```javascript
# upload a file
client -u /you/upload/file
```

You can upload file by:
```javascript
client -u /f/project.rar
```
![architecture](/doc/20180828095840.png)

also, it's cool that you can upload all files in a directory by:
```javascript
echo \"$(ls -m /f/foo)\" |xargs client -u
```
![architecture](/doc/20180828100341.png)

if you don't has a godfs client, you can use ```curl``` to upload files by:
```javascript
curl -F "file=@/your/file" "http://your.host:81/upload"
```
if upload success, server will return a json string like this:
```json
{
    "status":"success",
    "formData":{
        "data":[
            "G01/01/M/826d552525bceec5b8e9709efaf481ec"
        ],
        "name":[
            "mike"
        ]
    },
    "fileInfo":[
        {
            "index":0,
            "fileName":"mysql-cluster-community-7.6.7-1.sles12.x86_64.rpm-bundle.tar",
            "path":"G01/01/M/826d552525bceec5b8e9709efaf481ec"
        }
    ]
}
```

> The ```formData``` contains all parameters of your posted form, the file will be replaced by a remote path.

```javascript
# download a file as 123.zip
client -d G01/10/M/2c9da7ea280c020db7f4879f8180dfd6 -n 123.zip
```

### build docker image from latest source code:
```
cd godfs/docker
docker build -t godfs .
```
It is highly recommended to use docker to run godfs.
You can pull the docker image from [docker hub](https://hub.docker.com/r/hehety/godfs/):
```javascript
docker pull hehety/godfs
```

start tracker using docker:
```javascript
docker run -d -p 1022:1022 --name tracker --restart always -v /godfs/data:/godfs/data --privileged -e log_level="info" hehety/godfs:latest tracker
```

start storage using docker:
```javascript
docker run -d -p 1024:1024 -p 80:8001 --name storage -v /godfs/data:/godfs/data --privileged -e trackers=192.168.1.172:1022 -e bind_address=192.168.1.187 -e port=1024  -e instance_id="01" hehety/godfs storage
# you'd better add docker command '-e port=1024' on single machine.  
```
we're here using directory ```/godfs/data``` to persist data.

client usage:
```javascript
-u string 
    the file to be upload, if you want upload many file once, quote file paths using """ and split with ","
    example:
    client -u "/home/foo/bar1.tar.gz, /home/foo/bar1.tar.gz"
-d string 
    the file to be download
-l string 
    custom logging level: trace, debug, info, warning, error, and fatal
-n string 
    custom download file name
--set string
    set client config, for example: 
    client --set "tracker=127.0.0.1:1022"
    client --set "log_level=info"
```



