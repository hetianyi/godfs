#!/bin/bash
nginx && crond -L /cronjob/cron.log && tail -f /usr/local/nginx/logs/access.log