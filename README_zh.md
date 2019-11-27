godfs
==========
[![Build Status](https://travis-ci.org/hetianyi/godfs.svg?branch=master)](https://travis-ci.org/hetianyi/godfs)
[![go report card](https://goreportcard.com/badge/github.com/hetianyi/godfs "go report card")](https://goreportcard.com/report/github.com/hetianyi/godfs)

[README](README.md) | [中文文档](README_zh.md)

### ```godfs``` 是一个用go实现的轻量，快速，简单易用的分布式文件存储服务器。

```godfs``` 开箱即用，并对docker支持友好。

你可以在docker hub上下载最新镜像:
[https://hub.docker.com/r/hehety/godfs/](https://hub.docker.com/r/hehety/godfs/)



版本兼容表



## 特性

- 快速, 轻量, 开箱即用, 友好的API
- 易于扩展，运行稳定
- 低开销
- 提供HTTP方式的下载和上传API
- 支持文件断点下载
- 跨站点资源保护
- 支持不同平台下的编译运行: Linux, Windows, Mac
- 更好地支持docker容器
- 更好的文件迁移解决方案
- 支持读写和只读文件节点
- 文件组内自动同步
- 支持访问令牌

## 安装

> 请先安装golang1.11+

以CentOS7为例.

### 从最新的源码构建：
```shell
yum install golang -y
git clone https://github.com/hetianyi/godfs.git
cd godfs
./make.sh
# Windows下直接点击 make.cmd 开始构建。
```
构建成功后, 会在`````bin````` 目录下生成一个可执行文件:
```shell
./bin/godfs
```

将构建成功的二进制文件安装到目录 ```/usr/bin```:
```shell
./install.sh
```

启动tracker服务:
```shell
godfs tracker [options]
```
启动storage服务:
```shell
godfs storage [options]
```


上传文件:

```shell
# 已有tracker
godfs client upload /you/upload/file [--trackers <pass>@<host>:<port>[,<pass>@<host>:<port>]]
# 没有tracker服务器，也可以直接上传到指定的storage服务器
godfs client upload /you/upload/file [--storages <pass>@<host>:<port>[,<pass>@<host>:<port>]]
```
![architecture](/doc/20180828095840.png)

如果你想上传文件到指定的group，可以在命令行加参数```-g <groupID>```

你还可以用一个更酷的命令来上传一个文件夹下所有的文件:
```shell
godfs client upload *
```
![architecture](/doc/20180828100341.png)

也可以使用 ```curl``` 来上传文件:
```shell
curl -F "file=@/your/file" "http://your.host:http_port/upload"
```
上传成功之后，服务器会返回一个json字符串:
```json
{
    "accessMode":"public",
    "form":[
        {
            "index":1,
            "type":"file",
            "name":"f",
            "value":"fme_eval.msi",
            "size":1309757440,
            "group":"G01",
            "instanceId":"9de71d97",
            "md5":"71e55b30e36b1b8c3343270f86bb6234",
            "fileId":"CfzJHbO1MS84thD13PWEsLIURCw_ZZ7bIqPgpWFJxZ3Ad1cZFzTSL9AMP1CnCChK3Au9dqQ0ciAmdQ5Oaxgj0g"
        }
    ]
}
```

> 其中， ```form``` 是post表单中的所有字段的name-value信息，文件已被替换为上传之后的路径地址
> 如果你想上传文件到指定的group，可以在路径上加参数```?group=<groupID>```

```shell
# 下载文件
godfs client download CfzJHbO1MS84thD13PWEsLIURCw_ZZ7bIqPgpWFJxZ3Ad1cZFzTSL9AMP1CnCChK3Au9dqQ0ciAmdQ5Oaxgj0g --name 123.zip
```



#### Token的使用

token机制是参考FastDFS实现的，能够控制一个私有文件在一定时间内的可访问性。

token需要在后端自行生成，godfs只需要计算并匹配token，token携带的格式如下：

http://...?tk=<md5>&ts=<timestamp>

token计算：

md5(timestamp+filemd5+secret) ==> token



### 从最新源代码构建docker镜像：
```shell
cd godfs/docker
docker build -t godfs:latest .
```
推荐使用docker来运行godfs，最新的godfs的docker镜像可以在 [docker hub](https://hub.docker.com/r/hehety/godfs/) 获取:
```shell
docker pull hehety/godfs
```

启动tracker服务器:
```shell
docker run -d --net host --name tracker hehety/godfs:2.0.0-dev \
godfs tracker \
--bind-address 0.0.0.0 \
--secret 123123 \
--data-dir /godfs/data
```

启动storage服务器:
```shell
docker run -d --net host --name s1  hehety/godfs:2.0.0-dev \
godfs storage \
--bind-address 0.0.0.0 \
--secret 123123 \
--data-dir /godfs/data \
--log-dir /godfs/logs \
--trackers 123123@149.28.82.229:11706 \
--preferred-network eth0
```
你可以使用docker的命令```-e```来覆盖配置文件中的相应配置。

客户端命令:
```shell
Usage: godfs [global options] command [command options] [arguments...]

Commands:

     tracker  start as tracker server
     storage  start as storage server
     client   godfs client cli
     help, h  Shows a list of commands or help for one command

Options:

   --version, -v  show version
   --help, -h     show help
```





##### 说明：

godfs 2.0版本移除了sqlite数据库，使用boltdb来存储配置数据，使用gox包的set来存储文件id，简化的数据存储策略能够大幅提升随机写入性能。

