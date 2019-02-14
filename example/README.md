### godfs docker stack example 



You can start a stack with docker quickly!!!

Here we use nginx as http proxy and tcp proxy for tracker and storage.

Official nginx do not has stream module for tcp proxy, so you can use docker image:

```hehety/nginx:latest``` as proxy or build it from nginx source file by yourself.

Here is an example for a full stack of 2 groups(G01, G02 and each group has 2 replicas) with only 2 nodes:

| docker host             | node labels              | data volumes              |
| ----------------------- | ------------------------ | ------------------------- |
| 192.168.1.100 (gateway) | tracker,tracker1,storage | tracker,storage1,storage2 |
| 192.168.1.101           | tracker,tracker2,storage | tracker,storage1,storage2 |

Add labels for each node:

```shell
docker node update --label-add "tracker=true" <node ID>
docker node update --label-add "tracker1=true" <node ID>
docker node update --label-add "tracker2=true" <node ID>
docker node update --label-add "storage=true" <node ID>
```

Then start the stack:

```
./start-example.sh 192.168.1.100
```

Begin to use:

```shell
client config set secret="OASAD834jA97AAQE761=="
client config set trackers="192.168.1.100:1022,192.168.1.101:1023"
```


