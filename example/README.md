#### godfs docker stack example with 2 trackers and 4 storages.

deploy script:
```shell
docker node update --label-add "role=tracker" <tracker nodes>
docker node update --label-add "role=storage" <storage nodes>
docker stack deploy -c docker-compose.yml --prune --resolve-image changed godfs
```




