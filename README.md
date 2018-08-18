godfs
==========

# Description

### ```godfs``` is a simple fast, easy use distributed file system written by golang.

```godfs``` provides out-of-the-box usage and friendly support for docker，
You can pull the image on dockerhub:
[https://hub.docker.com/r/hehety/godfs/](https://hub.docker.com/r/hehety/godfs/)

## Features

- Fast, lightweight, stable, out-of-the-box, friendly api.
- Easy to expand.
- Native client api and java client api.
- API for http upload and download.
- Clear logs help troubleshoot errors.
- Cross-platform support.

## Install

> Please install golang first!

Take CentOS 7 as an example.

### build from latest source code:
```javascript
yum install golang -y
git clone https://github.com/hetianyi/godfs.git
cd godfs
# vi docker_build.sh and change env params below:
# GOROOT="/usr/local/go"
# GOPATH="$PWD:/go:/go/src";
./docker_build.sh
```
After the build is successful, three files will be generated under the ./bin directory:
```
./bin/client
./bin/storage
./bin/tracker
```


### build ```docker``` from latest source code:
```
cd godfs/docker
docker build -t godfs .
```

You can also pull docker image from docker hub:
```javascript
docker pull hehety/godfs
```

start tracker using docker:
```javascript
docker run -d -p 1022:1022 --name tracker --restart always -v /godfs/data:/godfs/data --privileged -e log_level="info" hehety/godfs:latest tracker
```
start storage using docker:
```javascript
docker run -d -p 1024:1024 -p 80:8001 --name storage -v /godfs/data:/godfs/data --privileged -e trackers=192.168.1.172:1022 -e bind_address=192.168.1.187  -e instance_id="01" hehety/godfs storage
```

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
--skip-check bool
    whether check file md5 before upload, true|false
```

for example, it's cool that you can upload all files in a directory by:
```javascript
cd /f/foo
echo \"$(ls -m /f/foo)\" |xargs /e/foo/bin/client -u
```


