#### godfs docker stack example with 2 trackers and 4 storages.

register relationship:

|tracker1|group|tracker2|group|
|:---|:---|:---|:---|
|storage1|G01|storage2|G02|
|storage3|G01|storage4|G02|


deploy script:
```shell
docker stack deploy -c docker-compose.yml --prune --resolve-image changed godfs
```




