godfs
==========
[![Build Status](https://travis-ci.org/hetianyi/godfs.svg?branch=master)](https://travis-ci.org/hetianyi/godfs)
[![go report card](https://goreportcard.com/badge/github.com/hetianyi/godfs "go report card")](https://goreportcard.com/report/github.com/hetianyi/godfs)

[README](README.md) | [中文文档](README_zh.md)
### ```godfs``` 是一个用go实现的轻量，快速，简单易用的分布式文件存储服务器。

```godfs``` 开箱即用，并对docker支持友好。

你可以在docker hub上下载最新镜像:
[https://hub.docker.com/r/hehety/godfs/](https://hub.docker.com/r/hehety/godfs/)

![architecture](/doc/20180830151005.png)

## 特性

- 快速, 轻量, 开箱即用, 友好的API
- 易于扩展，运行稳定
- 非常低的资源开销
- 提供原生的客户端和Java客户端(还未开始)
- 提供HTTP方式的下载和上传API
- 支持文件断点下载
- 支持http下载时基本验证
- 清晰的日志帮助排查运行异常
- 支持不同平台下的编译运行: Linux, Windows, Mac
- 更好地支持docker容器
- 文件分片保存
- 完美的文件迁移解决方案
- 支持读写和只读文件节点
- 文件组内自动同步

## 安装

> 请先安装golang1.8+

以CentOS7为例.

### 从最新的源码构建：
```javascript
yum install golang -y
git clone https://github.com/hetianyi/godfs.git
cd godfs
./make.sh
# Windows下直接点击 make.cmd 开始构建。
```
构建成功后, 三个文件会生成在`````./bin````` 目录下，分别是:
```
./bin/client
./bin/storage
./bin/tracker
```

将构建成功的二进制文件安装到目录 ```/usr/local/godfs```:
```javascript
./install.sh /usr/local/godfs
```

启动tracker服务:
```javascript
/usr/local/godfs/bin/tracker [-c /your/tracker/config/path]
```
启动storage服务:
```javascript
/usr/local/godfs/bin/storage [-c /your/storage/config/path]
```
然后你就可以在命令行直接使用 ```client``` 来上传和下载文件了。
> 当然要先设置trackers服务器设置
```javascript
# 例如，为客户端设置tracker服务器
client --set "trackers=host1:port1[,host2:port2]"
```

举个栗子:

上传一个文件:
```javascript
 client -u /you/upload/file
```
![architecture](/doc/20180828095840.png)

或者可以用一个更酷的命令来上传一个文件夹下所有的文件:
```javascript
echo \"$(ls -m /f/foo)\" |xargs client -u
```
![architecture](/doc/20180828100341.png)

如果你没有现成的godfs客户端，你可以使用 ```curl``` 来上传文件:
```javascript
curl -F "file=@/your/file" "http://your.host:http_port/upload"
```
上传成功之后，服务器会返回一个json字符串:
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

> 其中， ```formData``` 是post表单中的所有字段的name-value信息，文件已被替换为上传之后的路径地址。
如果你想上传文件到指定的group，可以在路径上加参数```?group=<groupID>```

```javascript
# 下载文件
client -d G01/10/M/2c9da7ea280c020db7f4879f8180dfd6 -n 123.zip
```

### 从最新源代码构建docker镜像：
```
cd godfs/docker
docker build -t godfs .
```
强烈推荐使用docker来运行godfs.
最新的godfs的docker镜像可以在 [docker hub](https://hub.docker.com/r/hehety/godfs/) 获取:
```javascript
docker pull hehety/godfs
```

启动tracker服务器:
```javascript
docker run -d -p 1022:1022 --name tracker --restart always -v /godfs/data:/godfs/data --privileged -e log_level="info" hehety/godfs:latest tracker
```

启动storage服务器:
```javascript
docker run -d -p 1024:1024 -p 80:8001 --name storage -v /godfs/data:/godfs/data --privileged -e trackers=192.168.1.172:1022 -e bind_address=192.168.1.187 -e port=1024  -e instance_id="01" hehety/godfs storage
# 单机上部署多个storage最好加上命令： '-e port=1024'
```
这里，我们使用宿主机上的目录 ```/godfs/data``` 来存放上传的文件，你可以使用docker的命令```-e```来覆盖配置文件中的相应配置。

客户端命令:
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




