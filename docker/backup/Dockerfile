FROM postgres:16-alpine

RUN apk add --no-cache gzip tar

COPY backup.sh /usr/local/bin/backup
RUN chmod +x /usr/local/bin/backup

ENTRYPOINT ["/usr/local/bin/backup"]
