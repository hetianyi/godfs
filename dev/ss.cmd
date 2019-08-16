mkdir storage
mkdir storage\\logs
"../bin/godfs.exe" --log-level=debug storage storage -g G01 --secret 123456 --bind-address=0.0.0.0 --data-dir ./storage --log-dir ./storage/logs