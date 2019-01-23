# build nginx docker images with stream module
FROM hehety/nginx
ADD entry.sh /usr/local/nginx/entry.sh
ADD crontab.sh /cronjob/crontab.sh
ADD nginx-structed-template.conf /cronjob/nginx-structed-template.conf
RUN chmod +x /cronjob/crontab.sh && chmod +x /usr/local/nginx/entry.sh && apk add bash curl && \
    echo "*	*	*	*	*	bash /cronjob/crontab.sh >> /cronjob/sync.log" > /tmp/cronjobs && crontab /tmp/cronjobs
WORKDIR /usr/local/nginx
ENV PATH /usr/local/nginx/sbin:$PATH
CMD ["./entry.sh"]