### godfs docker stack example 



Here we use nginx as http proxy and tcp proxy for tracker and storage.

official nginx do not has stream module for tcp proxy, so you can use docker image:

```hehety/nginx``` as proxy or build it from nginx source file by yourself.

for starting an example godfs stack quickly, you just need one single host with latest dokcer installed.

and before that, you should also edit the nginx.conf and docker compose file

deploy script:

```shell
docker node update --label-add "tracker=true" <tracker nodes>
docker node update --label-add "storage=true" <storage nodes>
docker node update --label-add "gateway=true" <nginx gateway nodes>
docker stack deploy -c docker-compose.yml --prune --resolve-image changed godfs
```










