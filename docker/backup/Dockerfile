FROM postgres:16-alpine

COPY backup.sh /usr/local/bin/backup
RUN chmod +x /usr/local/bin/backup

ENTRYPOINT ["/usr/local/bin/backup"]
