# build nginx docker images with stream module
FROM hehety/nginx
ADD crontab.sh /cronjob/crontab.sh
ADD nginx-structed-template.conf /cronjob/nginx-structed-template.conf
RUN echo "*	*	*	*	*	/cronjob/crontab.sh" > /tmp/cronjobs && crontab /tmp/cronjobs
WORKDIR /usr/local/nginx
ENV PATH /usr/local/nginx/sbin:$PATH
CMD ["nginx", "&&", "crond", "-f"]