godfs
==========

# Description

### ```godfs``` is a simple fast, easy use distributed file system written by golang.

```godfs``` provides out-of-the-box usage and friendly support for docker，

You can pull the image on docker hub:
[https://hub.docker.com/r/hehety/godfs/](https://hub.docker.com/r/hehety/godfs/)

## Features

- Fast, lightweight, stable, out-of-the-box, friendly api.
- Easy to expand, Stable to RUN.
- Low resource overhead.
- Native client api and java client api(not started).
- API for http upload and download.
- Clear logs help troubleshoot errors.
- Support different platforms: Linux, Windows, Mac
- Better support docker.
- Perfect data migration solution.
- Support readonly node.
- File synchronization in same group.

## Install

> Please install golang first!

Take CentOS 7 as example.

### build from latest source code:
```javascript
yum install golang -y
git clone https://github.com/hetianyi/godfs.git
cd godfs
./make.sh
```
After the build is successful, three files will be generated under the ./bin directory:
```
./bin/client
./bin/storage
./bin/tracker
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

also, it's cool that you can upload all files in a directory by:
```javascript
echo \"$(ls -m /f/foo)\" |xargs client -u
```


