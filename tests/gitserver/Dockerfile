FROM europe-docker.pkg.dev/kyma-project/prod/external/library/alpine:3.23.2

RUN apk add --no-cache git-daemon && \
    # apache2 and apache2-utils need to be in version smaller or equal to 2.4.58-r0
    # issue: https://github.com/kyma-project/serverless/issues/998
    apk add --no-cache apache2=2.4.58-r0 apache2-utils=2.4.58-r0 --repository http://dl-cdn.alpinelinux.org/alpine/v3.15/main && \
    apk upgrade --no-cache --repository=http://dl-cdn.alpinelinux.org/alpine/edge/community \
    expat

COPY ./etc/git.conf /etc/apache2/conf.d/git.conf
COPY ./etc/httpd.conf /etc/apache2/httpd.conf
RUN git config --system http.receivepack true &&\
  git config --system http.uploadpack true &&\
  git config --global user.email "gitserver@kyma-project.io" &&\
  git config --global user.name "Git Server"
ENV APACHE_LOG_DIR /var/log/apache2
ENV APACHE_LOCK_DIR /var/lock/apache2
ENV APACHE_PID_FILE /var/run/apache2.pid

COPY ./etc/init_repos.sh /tmp/init_repos.sh
COPY ./repos /tmp/repos
RUN /tmp/init_repos.sh

CMD ["/usr/sbin/httpd", "-D", "FOREGROUND", "-f", "/etc/apache2/httpd.conf"]

EXPOSE 80/tcp

